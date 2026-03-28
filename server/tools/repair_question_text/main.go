package main

import (
	"flag"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	"your-project/config"
	"your-project/model"
	"your-project/repository"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type repairCandidate struct {
	QuestionID     uint
	Reasons        []string
	BeforeTitle    string
	AfterTitle     string
	BeforeContent  string
	AfterContent   string
	BeforeExpected string
	AfterExpected  string
	Source         string
	Position       string
	Difficulty     string
}

var englishTokenPattern = regexp.MustCompile(`[A-Za-z]{8,}`)
var hanGapPattern = regexp.MustCompile(`([\p{Han}])\s+([\p{Han}])`)

func main() {
	apply := flag.Bool("apply", false, "apply text repair changes; default is dry-run")
	limit := flag.Int("limit", 30, "max candidate samples to print")
	scanLimit := flag.Int("scan-limit", 0, "max questions to scan; 0 means all")
	flag.Parse()

	if err := config.LoadConfig("config.yaml"); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := initDatabase()
	if err != nil {
		log.Fatalf("failed to init database: %v", err)
	}
	repository.SetDB(db)

	candidates, err := findRepairCandidates(db, *scanLimit)
	if err != nil {
		log.Fatalf("failed to scan questions: %v", err)
	}

	if len(candidates) == 0 {
		fmt.Println("No repair candidates found.")
		return
	}

	sort.Slice(candidates, func(i, j int) bool { return candidates[i].QuestionID < candidates[j].QuestionID })
	fmt.Printf("Found %d question repair candidates.\n", len(candidates))
	fmt.Printf("Mode: %s\n", ternary(*apply, "APPLY", "DRY-RUN"))

	for i, c := range candidates {
		if i >= *limit {
			fmt.Printf("... %d more omitted\n", len(candidates)-*limit)
			break
		}
		fmt.Printf("QID=%d source=%s reasons=%s\n", c.QuestionID, c.Source, strings.Join(c.Reasons, ","))
		fmt.Printf("  title:   %q -> %q\n", shorten(c.BeforeTitle, 80), shorten(c.AfterTitle, 80))
		fmt.Printf("  content: %q -> %q\n", shorten(c.BeforeContent, 80), shorten(c.AfterContent, 80))
		fmt.Printf("  answer:  %q -> %q\n", shorten(c.BeforeExpected, 80), shorten(c.AfterExpected, 80))
	}

	if !*apply {
		fmt.Println("Dry-run complete. Re-run with -apply to execute updates.")
		return
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		for _, c := range candidates {
			updates := map[string]interface{}{
				"title":           c.AfterTitle,
				"content":         c.AfterContent,
				"expected_answer": c.AfterExpected,
				"updated_at":      time.Now(),
			}
			if err := tx.Model(&model.Question{}).Where("id = ?", c.QuestionID).Updates(updates).Error; err != nil {
				return fmt.Errorf("update question %d failed: %w", c.QuestionID, err)
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalf("repair apply failed: %v", err)
	}

	fmt.Printf("Repair applied. Updated %d questions.\n", len(candidates))
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

func findRepairCandidates(db *gorm.DB, scanLimit int) ([]repairCandidate, error) {
	query := db.Model(&model.Question{}).Order("id DESC")
	if scanLimit > 0 {
		query = query.Limit(scanLimit)
	}

	var questions []model.Question
	if err := query.Find(&questions).Error; err != nil {
		return nil, err
	}

	candidates := make([]repairCandidate, 0, len(questions))
	for _, q := range questions {
		afterTitle, afterContent, afterExpected, reasons := repairQuestionText(q)
		if len(reasons) == 0 {
			continue
		}
		candidates = append(candidates, repairCandidate{
			QuestionID:     q.ID,
			Reasons:        reasons,
			BeforeTitle:    q.Title,
			AfterTitle:     afterTitle,
			BeforeContent:  q.Content,
			AfterContent:   afterContent,
			BeforeExpected: q.ExpectedAnswer,
			AfterExpected:  afterExpected,
			Source:         q.Source,
			Position:       q.Position,
			Difficulty:     q.Difficulty,
		})
	}

	return candidates, nil
}

func repairQuestionText(q model.Question) (title, content, expected string, reasons []string) {
	title = sanitizeText(q.Title)
	content = sanitizeText(q.Content)
	expected = sanitizeText(q.ExpectedAnswer)

	reasonSet := map[string]struct{}{}
	addReason := func(reason string) {
		if strings.TrimSpace(reason) == "" {
			return
		}
		if _, ok := reasonSet[reason]; ok {
			return
		}
		reasonSet[reason] = struct{}{}
		reasons = append(reasons, reason)
	}

	if title != strings.TrimSpace(q.Title) || content != strings.TrimSpace(q.Content) || expected != strings.TrimSpace(q.ExpectedAnswer) {
		addReason("sanitize_invalid_chars")
	}

	topic := deriveTopic(q, title)
	if looksLikeOpeningEnglishTemplate(title) {
		title = fmt.Sprintf("%s：核心原理与实践应用", topic)
		addReason("english_opening_title_template")
	}

	if strings.Contains(strings.ToLower(content), "please explain") && strings.Contains(strings.ToLower(content), "concept") {
		content = fmt.Sprintf("请系统说明%s的概念、运行机制及典型应用场景。", topic)
		addReason("english_opening_content_template")
	}
	if strings.Contains(strings.ToLower(content), "thread safety") {
		content = fmt.Sprintf("请系统说明%s的概念、运行机制、线程安全与性能取舍。", topic)
		addReason("english_thread_safety_template")
	}
	if strings.Contains(strings.ToLower(content), "level question") {
		content = fmt.Sprintf("请结合%s岗位要求，系统说明你的思路、关键实现与工程取舍。", topic)
		addReason("english_level_content_template")
	}

	if looksLikeExpectedEnglishTemplate(expected) {
		expected = fmt.Sprintf("回答应覆盖%s的定义、实现机制、边界条件与技术取舍。", topic)
		addReason("english_expected_template")
	}

	if strings.Contains(strings.ToLower(title), "level question") {
		title = fmt.Sprintf("%s岗位技术问题", topic)
		addReason("english_level_title_template")
	}

	if shouldRewriteToChinese(title, 0.35) {
		title = fallbackTitle(topic)
		addReason("title_non_chinese")
	}
	if shouldRewriteToChinese(content, 0.4) {
		content = fallbackContent(topic)
		addReason("content_non_chinese")
	}
	if shouldRewriteToChinese(expected, 0.35) {
		expected = fallbackExpected(topic)
		addReason("expected_non_chinese")
	}

	if title == "" {
		title = fallbackTitle(topic)
		addReason("empty_title")
	}
	if content == "" {
		content = fallbackContent(topic)
		addReason("empty_content")
	}
	if expected == "" {
		expected = fallbackExpected(topic)
		addReason("empty_expected")
	}

	if title == strings.TrimSpace(q.Title) && content == strings.TrimSpace(q.Content) && expected == strings.TrimSpace(q.ExpectedAnswer) {
		return title, content, expected, []string{}
	}

	return title, content, expected, reasons
}

func sanitizeText(text string) string {
	if strings.TrimSpace(text) == "" {
		return ""
	}
	cleaned := strings.ToValidUTF8(text, "")
	cleaned = strings.ReplaceAll(cleaned, "\uFFFD", "")
	cleaned = strings.ReplaceAll(cleaned, "\x00", "")
	cleaned = strings.ReplaceAll(cleaned, "\uFEFF", "")
	cleaned = strings.ReplaceAll(cleaned, "\u200B", "")
	cleaned = strings.ReplaceAll(cleaned, "\r\n", "\n")
	cleaned = strings.ReplaceAll(cleaned, "\r", "\n")
	cleaned = strings.TrimSpace(cleaned)
	cleaned = hanGapPattern.ReplaceAllString(cleaned, "$1$2")
	return strings.TrimSpace(cleaned)
}

func looksLikeOpeningEnglishTemplate(title string) bool {
	lower := strings.ToLower(title)
	return strings.Contains(lower, "core principles and practice") || strings.Contains(lower, "topic:")
}

func looksLikeExpectedEnglishTemplate(expected string) bool {
	lower := strings.ToLower(expected)
	return strings.Contains(lower, "should cover definition") || strings.Contains(lower, "cover definition")
}

func deriveTopic(q model.Question, title string) string {
	candidate := strings.TrimSpace(q.Category)
	if candidate == "" {
		candidate = strings.TrimSpace(q.Position)
	}
	if candidate == "" {
		candidate = strings.TrimSpace(title)
	}
	if candidate == "" {
		candidate = "通用技术"
	}
	if !isMostlyChinese(candidate, 0.55) {
		if isMostlyChinese(strings.TrimSpace(q.Position), 0.45) {
			candidate = strings.TrimSpace(q.Position)
		} else {
			candidate = "该技术方向"
		}
	}
	candidate = strings.Trim(candidate, ":：")
	if strings.Contains(strings.ToLower(candidate), "core principles and practice") {
		candidate = strings.TrimSpace(strings.Split(candidate, ":")[0])
	}
	if candidate == "" {
		candidate = "通用技术"
	}
	return candidate
}

func fallbackTitle(topic string) string {
	t := strings.TrimSpace(topic)
	if t == "" {
		t = "通用技术"
	}
	return fmt.Sprintf("%s：核心原理与实践应用", t)
}

func fallbackContent(topic string) string {
	t := strings.TrimSpace(topic)
	if t == "" {
		t = "该岗位"
	}
	if strings.Contains(t, "岗位") || strings.Contains(t, "方向") {
		return fmt.Sprintf("请结合%s，系统说明你的思路、关键实现与工程取舍。", t)
	}
	return fmt.Sprintf("请结合%s岗位要求，系统说明你的思路、关键实现与工程取舍。", t)
}

func fallbackExpected(topic string) string {
	t := strings.TrimSpace(topic)
	if t == "" {
		t = "该技术方向"
	}
	return fmt.Sprintf("回答应覆盖%s的核心原理、实现步骤、关键细节与风险边界。", t)
}

func shouldRewriteToChinese(text string, ratio float64) bool {
	content := strings.TrimSpace(text)
	if content == "" {
		return false
	}
	if strings.ContainsRune(content, '\ufffd') {
		return true
	}
	if !englishTokenPattern.MatchString(content) {
		return false
	}
	return !isMostlyChinese(content, ratio)
}

func isMostlyChinese(text string, ratio float64) bool {
	content := strings.TrimSpace(text)
	if content == "" {
		return false
	}
	hanCount := 0
	letterCount := 0
	for _, r := range content {
		if unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsDigit(r) {
			continue
		}
		if unicode.IsLetter(r) {
			letterCount++
		}
		if r >= 0x4E00 && r <= 0x9FFF {
			hanCount++
		}
	}
	if letterCount == 0 {
		return false
	}
	return float64(hanCount)/float64(letterCount) >= ratio
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
