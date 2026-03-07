package main

import (
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"

	"your-project/config"
	"your-project/model"
	"your-project/repository"
	"your-project/service"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type candidate struct {
	QuestionID uint
	Source     string
	Title      string
	Content    string
	Reason     string
}

func main() {
	apply := flag.Bool("apply", false, "apply deletion changes; default is dry-run")
	limit := flag.Int("limit", 30, "max candidate samples to print")
	flag.Parse()

	if err := config.LoadConfig("config.yaml"); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := initDatabase()
	if err != nil {
		log.Fatalf("failed to init database: %v", err)
	}
	repository.SetDB(db)

	candidates, err := findGarbageFollowUps(db)
	if err != nil {
		log.Fatalf("failed to scan garbage data: %v", err)
	}

	if len(candidates) == 0 {
		fmt.Println("No garbage follow-up questions found.")
		return
	}

	sort.Slice(candidates, func(i, j int) bool { return candidates[i].QuestionID < candidates[j].QuestionID })

	fmt.Printf("Found %d garbage follow-up questions.\n", len(candidates))
	fmt.Printf("Mode: %s\n", ternary(*apply, "APPLY", "DRY-RUN"))

	for i, c := range candidates {
		if i >= *limit {
			fmt.Printf("... %d more omitted\n", len(candidates)-*limit)
			break
		}
		fmt.Printf("QID=%d source=%s reason=%s title=%q\n", c.QuestionID, c.Source, c.Reason, shorten(c.Title, 80))
	}

	if !*apply {
		fmt.Println("Dry-run complete. Re-run with -apply to delete these records.")
		return
	}

	ids := make([]uint, 0, len(candidates))
	for _, c := range candidates {
		ids = append(ids, c.QuestionID)
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("question_id IN ?", ids).Delete(&model.AnswerResult{}).Error; err != nil {
			return fmt.Errorf("delete answer_results failed: %w", err)
		}
		if err := tx.Where("question_id IN ?", ids).Delete(&model.InterviewQuestion{}).Error; err != nil {
			return fmt.Errorf("delete interview_questions failed: %w", err)
		}
		if err := tx.Where("id IN ?", ids).Delete(&model.Question{}).Error; err != nil {
			return fmt.Errorf("delete questions failed: %w", err)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("cleanup failed: %v", err)
	}

	fmt.Printf("Cleanup applied. Deleted %d questions and related interview data.\n", len(ids))
}

func initDatabase() (*gorm.DB, error) {
	cfg := config.GetConfig()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Database,
	)
	return gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
}

func findGarbageFollowUps(db *gorm.DB) ([]candidate, error) {
	var iqs []model.InterviewQuestion
	if err := db.Where("order_index > 0").Order("interview_id ASC, order_index ASC").Find(&iqs).Error; err != nil {
		return nil, err
	}

	byID := make(map[uint]candidate)
	for _, iq := range iqs {
		var q model.Question
		if err := db.First(&q, iq.QuestionID).Error; err != nil {
			continue
		}

		if q.Source == "follow_up" && !q.RAGEligible && likelyGarbageQuestion(q) {
			byID[q.ID] = candidate{QuestionID: q.ID, Source: q.Source, Title: q.Title, Content: q.Content, Reason: "low_quality_followup_text"}
			continue
		}

		var prevIQ model.InterviewQuestion
		err := db.Where("interview_id = ? AND order_index = ?", iq.InterviewID, iq.OrderIndex-1).First(&prevIQ).Error
		if err != nil {
			continue
		}

		var prevAnswer model.AnswerResult
		err = db.Where("interview_id = ? AND question_id = ?", iq.InterviewID, prevIQ.QuestionID).Order("created_at DESC").First(&prevAnswer).Error
		if err != nil {
			continue
		}

		if isGarbageTriggerAnswer(prevAnswer.Answer) && likelyGarbageQuestion(q) {
			reason := "generated_after_low_signal_answer"
			if q.Source == "follow_up" {
				reason = "follow_up_after_low_signal_answer"
			}
			byID[q.ID] = candidate{QuestionID: q.ID, Source: q.Source, Title: q.Title, Content: q.Content, Reason: reason}
		}
	}

	out := make([]candidate, 0, len(byID))

	// Fallback scan: questions that are only used from follow-up positions and look low quality.
	type usageInfo struct {
		QuestionID uint
		MinOrder   int
	}
	var usage []usageInfo
	if err := db.Model(&model.InterviewQuestion{}).
		Select("question_id, MIN(order_index) as min_order").
		Group("question_id").
		Scan(&usage).Error; err != nil {
		return nil, err
	}
	for _, u := range usage {
		if u.MinOrder <= 0 {
			continue
		}
		if _, exists := byID[u.QuestionID]; exists {
			continue
		}
		var q model.Question
		if err := db.First(&q, u.QuestionID).Error; err != nil {
			continue
		}
		if likelyGarbageQuestion(q) {
			byID[q.ID] = candidate{QuestionID: q.ID, Source: q.Source, Title: q.Title, Content: q.Content, Reason: "only_seen_as_followup_and_low_quality"}
		}
	}

	for _, c := range byID {
		out = append(out, c)
	}
	return out, nil
}

func isGarbageTriggerAnswer(answer string) bool {
	trimmed := strings.TrimSpace(answer)
	if service.IsInvalidAnswer(trimmed) {
		return true
	}
	if len([]rune(trimmed)) <= 8 {
		return true
	}
	lower := strings.ToLower(trimmed)
	badTokens := []string{"不知道", "不会", "答不出来", "回答不出来", "不清楚", "idk", "i don't know", "no idea"}
	for _, token := range badTokens {
		if strings.Contains(lower, token) {
			return true
		}
	}
	return false
}

func likelyGarbageQuestion(q model.Question) bool {
	all := strings.TrimSpace(q.Title + " " + q.Content)
	if all == "" {
		return true
	}
	if service.IsInvalidAnswer(all) {
		return true
	}
	if len([]rune(strings.TrimSpace(q.Content))) < 10 {
		return true
	}
	lower := strings.ToLower(all)
	generic := []string{"继续说", "再说一下", "展开说说", "补充一下", "还有吗", "请继续"}
	for _, g := range generic {
		if strings.Contains(lower, g) && len([]rune(q.Content)) < 18 {
			return true
		}
	}
	return false
}

func shorten(s string, n int) string {
	runes := []rune(strings.TrimSpace(s))
	if len(runes) <= n {
		return string(runes)
	}
	return string(runes[:n]) + "..."
}

func ternary(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}
