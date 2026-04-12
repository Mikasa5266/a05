package initializer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"your-project/config"
	"your-project/internal/model"
	ragpkg "your-project/pkg/rag"

	"gorm.io/gorm"
)

const (
	seedSourceKnowledgeBase = "knowledge_base_seed"
	knowledgeEmbedBatchSize = 48
)

var (
	headingLevel3Pattern = regexp.MustCompile(`^###\s+(.+)$`)
	optionLinePattern    = regexp.MustCompile(`^[A-D][\.、．]\s+.+`)
	answerPattern        = regexp.MustCompile(`(?s)\*\*参考答案：\*\*\s*(.+?)(\n\*\*(答案解析|解析)：\*\*|$)`)
	analysisPattern      = regexp.MustCompile(`(?s)\*\*(答案解析|解析)：\*\*\s*(.+)`)
	numberPrefixPattern  = regexp.MustCompile(`^\d+[\.、．]\s*`)
)

// InitSampleQuestions initializes question bank from knowledge_base markdown files.
func InitSampleQuestions(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database is nil")
	}

	kbRoot, err := resolveKnowledgeBaseRoot()
	if err != nil {
		return err
	}

	questions, err := parseQuestionsFromKnowledgeBase(kbRoot)
	if err != nil {
		return err
	}
	if len(questions) == 0 {
		log.Printf("No seed question parsed from %s", kbRoot)
		return nil
	}

	persisted, err := upsertSeedQuestions(db, questions)
	if err != nil {
		return err
	}
	log.Printf("Knowledge base seed complete: parsed=%d, upserted=%d", len(questions), len(persisted))

	if err := upsertQuestionsToQdrant(context.Background(), persisted); err != nil {
		log.Printf("Warning: failed to upsert seeded questions to Qdrant: %v", err)
	}

	return nil
}

func resolveKnowledgeBaseRoot() (string, error) {
	candidates := []string{
		filepath.Join("..", "knowledge_base"),
		"knowledge_base",
	}

	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err != nil {
			continue
		}
		if info.IsDir() {
			abs, absErr := filepath.Abs(candidate)
			if absErr != nil {
				return candidate, nil
			}
			return abs, nil
		}
	}

	return "", fmt.Errorf("knowledge_base directory not found")
}

func parseQuestionsFromKnowledgeBase(root string) ([]model.Question, error) {
	allowedDomains := map[string]bool{
		"backend":     true,
		"frontend":    true,
		"algorithm":   true,
		"ai_engineer": true,
		"behavioral":  true,
	}

	result := make([]model.Question, 0, 256)
	seen := make(map[string]struct{}, 256)

	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		if strings.ToLower(filepath.Ext(path)) != ".md" {
			return nil
		}
		if strings.EqualFold(info.Name(), "README.md") {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		parts := strings.Split(rel, "/")
		if len(parts) < 2 {
			return nil
		}
		domain := strings.TrimSpace(parts[0])
		if !allowedDomains[domain] {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		meta := questionSeedMeta{
			Domain:      domain,
			RelativeSrc: rel,
			Category:    strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
			Position:    mapDomainToPosition(domain),
		}

		parsed := extractQuestionsFromMarkdown(string(content), meta)
		for _, item := range parsed {
			key := normalizeQuestionFingerprint(item.Position, item.Title, item.Content)
			if key == "" {
				continue
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			result = append(result, item)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

type questionSeedMeta struct {
	Domain      string
	RelativeSrc string
	Category    string
	Position    string
}

func extractQuestionsFromMarkdown(doc string, meta questionSeedMeta) []model.Question {
	lines := strings.Split(doc, "\n")
	questions := make([]model.Question, 0, 32)

	currentHeading := ""
	var bodyBuilder strings.Builder

	flushSection := func() {
		heading := strings.TrimSpace(currentHeading)
		body := strings.TrimSpace(bodyBuilder.String())
		bodyBuilder.Reset()
		if heading == "" {
			return
		}

		q, ok := buildQuestionFromSection(heading, body, meta)
		if !ok {
			return
		}
		questions = append(questions, q)
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if matches := headingLevel3Pattern.FindStringSubmatch(trimmed); len(matches) > 1 {
			flushSection()
			currentHeading = strings.TrimSpace(matches[1])
			continue
		}
		if currentHeading == "" {
			continue
		}
		bodyBuilder.WriteString(line)
		bodyBuilder.WriteString("\n")
	}
	flushSection()

	return questions
}

func buildQuestionFromSection(rawHeading, body string, meta questionSeedMeta) (model.Question, bool) {
	heading := cleanQuestionHeading(rawHeading)
	if !looksLikeQuestionHeading(heading, body) {
		return model.Question{}, false
	}

	questionContent := extractQuestionContent(heading, body)
	if questionContent == "" {
		questionContent = fmt.Sprintf("请围绕“%s”进行系统说明。", heading)
	}

	expectedAnswer := extractExpectedAnswer(body)
	if expectedAnswer == "" {
		expectedAnswer = truncateText(body, 1200)
	}
	if expectedAnswer == "" {
		expectedAnswer = "请结合知识点给出结构化回答，覆盖定义、原理、实现和边界。"
	}

	difficulty := inferDifficultyFromSection(heading, body)
	tags := []string{meta.Domain, meta.Category, seedSourceKnowledgeBase, difficulty}

	q := model.Question{
		Title:          truncateText(heading, 120),
		Content:        truncateText(questionContent, 500),
		Position:       meta.Position,
		Difficulty:     difficulty,
		Category:       meta.Category,
		Source:         seedSourceKnowledgeBase,
		RAGEligible:    true,
		ExpectedAnswer: truncateText(expectedAnswer, 2600),
	}
	q.SetTags(tags)
	return q, true
}

func cleanQuestionHeading(raw string) string {
	heading := strings.TrimSpace(raw)
	heading = numberPrefixPattern.ReplaceAllString(heading, "")
	heading = strings.TrimPrefix(heading, "- ")
	heading = strings.TrimSpace(heading)
	return heading
}

func looksLikeQuestionHeading(heading, body string) bool {
	if heading == "" {
		return false
	}
	headingLower := strings.ToLower(heading)
	if strings.Contains(heading, "知识点部分") || strings.Contains(heading, "题库部分") || strings.Contains(heading, "扩展") || strings.Contains(heading, "加量") {
		return false
	}
	if strings.Contains(headingLower, "knowledge points") || strings.Contains(headingLower, "appendix") {
		return false
	}
	if strings.Contains(heading, "单选题") || strings.Contains(heading, "简答题") {
		return true
	}
	if strings.ContainsAny(heading, "？?") {
		return true
	}
	if strings.Contains(body, "**解析：**") || strings.Contains(body, "**参考答案：**") {
		return true
	}
	return false
}

func extractQuestionContent(heading, body string) string {
	content := heading
	if idx := strings.IndexAny(heading, "？?"); idx >= 0 && idx+1 < len(heading) {
		after := strings.TrimSpace(heading[idx+1:])
		if after != "" {
			content = after
		}
	}

	options := collectOptionLines(body)
	if len(options) > 0 {
		content = strings.TrimSpace(content + "\n" + strings.Join(options, "\n"))
	}
	return content
}

func collectOptionLines(body string) []string {
	lines := strings.Split(body, "\n")
	options := make([]string, 0, 4)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if optionLinePattern.MatchString(trimmed) {
			options = append(options, trimmed)
		}
	}
	return options
}

func extractExpectedAnswer(body string) string {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return ""
	}

	if matches := answerPattern.FindStringSubmatch(trimmed); len(matches) > 1 {
		answer := strings.TrimSpace(matches[1])
		if answer != "" {
			return answer
		}
	}

	if matches := analysisPattern.FindStringSubmatch(trimmed); len(matches) > 2 {
		analysis := strings.TrimSpace(matches[2])
		if analysis != "" {
			return analysis
		}
	}

	return ""
}

func inferDifficultyFromSection(heading, body string) string {
	text := strings.ToLower(strings.TrimSpace(heading + " " + body))
	switch {
	case strings.Contains(text, "单选题"):
		return "campus_intern"
	case strings.Contains(text, "高并发"), strings.Contains(text, "分布式"), strings.Contains(text, "多活"), strings.Contains(text, "系统设计"), strings.Contains(text, "架构"):
		return "social_junior"
	case strings.Contains(text, "简答题"):
		return "campus_graduate"
	default:
		return "campus_graduate"
	}
}

func mapDomainToPosition(domain string) string {
	switch strings.TrimSpace(domain) {
	case "backend":
		return "Java后端工程师"
	case "frontend":
		return "前端工程师"
	case "algorithm":
		return "算法工程师"
	case "ai_engineer":
		return "AI工程师"
	case "behavioral":
		return "Java后端工程师"
	default:
		return "Java后端工程师"
	}
}

func normalizeQuestionFingerprint(position, title, content string) string {
	joined := strings.ToLower(strings.TrimSpace(position + "|" + title + "|" + content))
	joined = strings.ReplaceAll(joined, " ", "")
	joined = strings.ReplaceAll(joined, "\n", "")
	joined = strings.ReplaceAll(joined, "\t", "")
	joined = strings.ReplaceAll(joined, "。", "")
	joined = strings.ReplaceAll(joined, "，", "")
	joined = strings.ReplaceAll(joined, "：", "")
	joined = strings.ReplaceAll(joined, ":", "")
	joined = strings.ReplaceAll(joined, "？", "")
	joined = strings.ReplaceAll(joined, "?", "")
	joined = strings.ReplaceAll(joined, "（", "")
	joined = strings.ReplaceAll(joined, "）", "")
	joined = strings.ReplaceAll(joined, "(", "")
	joined = strings.ReplaceAll(joined, ")", "")
	return joined
}

func truncateText(text string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(strings.TrimSpace(text))
	if len(runes) <= max {
		return string(runes)
	}
	return strings.TrimSpace(string(runes[:max]))
}

func upsertSeedQuestions(db *gorm.DB, questions []model.Question) ([]*model.Question, error) {
	persisted := make([]*model.Question, 0, len(questions))
	for _, candidate := range questions {
		stored, err := upsertSingleQuestion(db, candidate)
		if err != nil {
			return nil, err
		}
		if stored != nil {
			persisted = append(persisted, stored)
		}
	}
	return persisted, nil
}

func upsertSingleQuestion(db *gorm.DB, candidate model.Question) (*model.Question, error) {
	var existing model.Question
	err := db.Where("title = ? AND content = ? AND position = ? AND difficulty = ?", candidate.Title, candidate.Content, candidate.Position, candidate.Difficulty).First(&existing).Error
	if err == nil {
		existing.Category = candidate.Category
		existing.Source = candidate.Source
		existing.RAGEligible = true
		existing.ExpectedAnswer = candidate.ExpectedAnswer
		existing.Tags = candidate.Tags
		if saveErr := db.Save(&existing).Error; saveErr != nil {
			return nil, fmt.Errorf("failed to update seed question: %w", saveErr)
		}
		return &existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to query seed question: %w", err)
	}

	if createErr := db.Create(&candidate).Error; createErr != nil {
		return nil, fmt.Errorf("failed to create seed question: %w", createErr)
	}
	return &candidate, nil
}

func upsertQuestionsToQdrant(ctx context.Context, questions []*model.Question) error {
	if len(questions) == 0 {
		return nil
	}

	store, err := ragpkg.NewQdrantStoreFromEnv()
	if err != nil {
		return err
	}

	pingCtx, cancelPing := context.WithTimeout(ctx, 5*time.Second)
	defer cancelPing()
	if err := store.Ping(pingCtx); err != nil {
		return err
	}

	embedder, err := buildInitializerEmbedder()
	if err != nil {
		return err
	}

	points := make([]ragpkg.VectorPoint, 0, len(questions))
	for _, q := range questions {
		if q == nil || q.ID == 0 {
			continue
		}
		text := buildQuestionEmbeddingText(q)
		if strings.TrimSpace(text) == "" {
			continue
		}

		vector, embErr := embedder.Embed(ctx, text)
		if embErr != nil {
			continue
		}

		points = append(points, ragpkg.VectorPoint{
			ID:      fmt.Sprintf("question_%d", q.ID),
			Vector:  vector,
			Content: text,
			Metadata: map[string]string{
				"kind":        "question",
				"question_id": fmt.Sprintf("%d", q.ID),
				"position":    strings.TrimSpace(q.Position),
				"difficulty":  strings.TrimSpace(q.Difficulty),
				"category":    strings.TrimSpace(q.Category),
				"source":      strings.TrimSpace(q.Source),
			},
		})
	}

	if len(points) == 0 {
		return nil
	}

	for start := 0; start < len(points); start += knowledgeEmbedBatchSize {
		end := start + knowledgeEmbedBatchSize
		if end > len(points) {
			end = len(points)
		}
		if err := store.Upsert(ctx, points[start:end]); err != nil {
			return err
		}
	}
	return nil
}

func buildInitializerEmbedder() (ragpkg.Embedder, error) {
	embedder, err := ragpkg.NewOpenAIEmbedderFromEnv()
	if err == nil {
		return embedder, nil
	}

	cfg := config.GetConfig()
	if cfg == nil {
		return nil, fmt.Errorf("embedding config not found")
	}

	modelName := strings.TrimSpace(os.Getenv("EMBEDDING_MODEL"))
	if modelName == "" {
		modelName = strings.TrimSpace(cfg.LLM.Models["embedding"])
	}

	return ragpkg.NewOpenAIEmbedder(ragpkg.OpenAIEmbedderConfig{
		APIKey:  strings.TrimSpace(cfg.LLM.APIKey),
		BaseURL: strings.TrimSpace(cfg.LLM.BaseURL),
		Model:   modelName,
	})
}

func buildQuestionEmbeddingText(q *model.Question) string {
	if q == nil {
		return ""
	}
	parts := []string{
		strings.TrimSpace(q.Title),
		strings.TrimSpace(q.Content),
	}
	if strings.TrimSpace(q.ExpectedAnswer) != "" {
		parts = append(parts, "参考答案："+strings.TrimSpace(q.ExpectedAnswer))
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}
