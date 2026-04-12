package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"your-project/internal/model"
)

func (s *AIService) GenerateQuestions(ctx context.Context, interview *model.Interview, count int) ([]*model.Question, error) {
	prompt, err := s.renderPrompt("generate_questions.tmpl", map[string]interface{}{
		"Count":                 count,
		"Position":              interview.Position,
		"Difficulty":            interview.Difficulty,
		"Mode":                  interview.Mode,
		"Style":                 interview.Style,
		"ModeInstruction":       s.buildModePrompt(interview.Mode),
		"StyleInstruction":      s.buildStylePrompt(interview.Style, interview.Company),
		"DifficultyInstruction": s.buildDifficultyPrompt(interview.Difficulty),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to render generate questions prompt: %w", err)
	}

	response, err := s.chat(ctx, prompt, "chat", jsonObjectResponseFormat())
	if err != nil {
		return nil, fmt.Errorf("failed to generate questions: %w", err)
	}

	var questionsData []struct {
		Title          string `json:"title"`
		Content        string `json:"content"`
		ExpectedAnswer string `json:"expected_answer"`
	}
	if err := json.Unmarshal([]byte(response), &questionsData); err != nil {
		return nil, fmt.Errorf("failed to parse questions response: %w, body: %s", err, response)
	}

	questions := make([]*model.Question, 0, len(questionsData))
	for _, qd := range questionsData {
		item := &model.Question{Title: qd.Title, Content: qd.Content, ExpectedAnswer: qd.ExpectedAnswer, Position: interview.Position, Difficulty: interview.Difficulty}
		s.EnsureQuestionChinese(ctx, item)
		questions = append(questions, item)
	}
	return questions, nil
}

func (s *AIService) GenerateTopicQuestionFromContext(ctx context.Context, interview *model.Interview, ragContext string, category string) (*model.Question, error) {
	if interview == nil {
		return nil, fmt.Errorf("interview is nil")
	}

	prompt, err := s.renderPrompt("generate_topic_question_from_context.tmpl", map[string]interface{}{
		"Position":              interview.Position,
		"Difficulty":            interview.Difficulty,
		"Mode":                  interview.Mode,
		"Style":                 interview.Style,
		"ModeInstruction":       s.buildModePrompt(interview.Mode),
		"StyleInstruction":      s.buildStylePrompt(interview.Style, interview.Company),
		"DifficultyInstruction": s.buildDifficultyPrompt(interview.Difficulty),
		"RAGContext":            ragContext,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to render topic question prompt: %w", err)
	}

	response, err := s.chat(ctx, prompt, "chat", jsonObjectResponseFormat())
	if err != nil {
		return nil, fmt.Errorf("failed to generate topic question: %w", err)
	}

	var q struct {
		Title          string `json:"title"`
		Content        string `json:"content"`
		ExpectedAnswer string `json:"expected_answer"`
	}
	if err := json.Unmarshal([]byte(response), &q); err != nil {
		return nil, fmt.Errorf("failed to parse topic question response: %w, body: %s", err, response)
	}

	result := &model.Question{Title: q.Title, Content: q.Content, ExpectedAnswer: q.ExpectedAnswer, Position: interview.Position, Difficulty: interview.Difficulty, Category: category}
	s.EnsureQuestionChinese(ctx, result)
	s.ensureOpeningQuestionTone(result, category)
	return result, nil
}

func (s *AIService) ensureOpeningQuestionTone(q *model.Question, category string) {
	if q == nil {
		return
	}
	text := strings.TrimSpace(q.Title + " " + q.Content)
	if !isFollowUpWording(text) {
		return
	}
	topic := strings.TrimSpace(category)
	if topic == "" {
		topic = "通用技术主题"
	}
	q.Title = fmt.Sprintf("%s：核心原理与实践应用", topic)
	q.Content = fmt.Sprintf("请系统说明%s的概念、运行机制及典型应用场景。", topic)
	if strings.TrimSpace(q.ExpectedAnswer) == "" {
		q.ExpectedAnswer = fmt.Sprintf("回答应覆盖%s的定义、实现机制、约束条件与技术取舍。", topic)
	}
}

func isFollowUpWording(text string) bool {
	t := strings.TrimSpace(strings.ToLower(text))
	if t == "" {
		return false
	}
	patterns := []string{"you mentioned", "continue", "further", "follow-up", "based on previous", "你提到", "继续", "进一步", "基于上一轮", "接着"}
	for _, p := range patterns {
		if strings.Contains(t, p) {
			return true
		}
	}
	return false
}

var openingQuestionContextPatterns = []string{"previous", "above", "continue", "further", "those", "these", "上文", "上一轮", "之前", "继续", "进一步", "上述"}
var openingQuestionQuantifierRef = regexp.MustCompile(`this\s+\d+\s+`)

func (s *AIService) IsContextDependentOpeningQuestion(question *model.Question) bool {
	if question == nil {
		return true
	}
	text := strings.TrimSpace(question.Title + " " + question.Content)
	if text == "" {
		return true
	}
	if isFollowUpWording(text) || openingQuestionQuantifierRef.MatchString(strings.ToLower(text)) {
		return true
	}
	for _, p := range openingQuestionContextPatterns {
		if strings.Contains(strings.ToLower(text), p) {
			return true
		}
	}
	return false
}

func (s *AIService) NormalizeToSelfContainedOpening(question *model.Question) {
	if question == nil {
		return
	}
	topic := strings.TrimSpace(question.Category)
	if topic == "" {
		topic = strings.TrimSpace(question.Title)
	}
	if topic == "" {
		topic = "通用技术主题"
	}
	question.Title = fmt.Sprintf("%s：核心原理与实践应用", topic)
	question.Content = fmt.Sprintf("请系统说明%s的概念、运行机制、线程安全与性能取舍。", topic)
	if strings.TrimSpace(question.ExpectedAnswer) == "" {
		question.ExpectedAnswer = fmt.Sprintf("回答应覆盖%s的定义、实现机制、边界条件与技术取舍。", topic)
	}
	question.Title = sanitizeGeneratedText(question.Title)
	question.Content = sanitizeGeneratedText(question.Content)
	question.ExpectedAnswer = sanitizeGeneratedText(question.ExpectedAnswer)
}

func (s *AIService) GenerateClarifyingFollowUpQuestion(ctx context.Context, currentQ *model.Question, answer string, followUpIndex int) (*model.Question, error) {
	if currentQ == nil {
		return nil, fmt.Errorf("current question is nil")
	}
	prompt, err := s.renderPrompt("generate_clarifying_followup_question.tmpl", map[string]interface{}{"CurrentTitle": currentQ.Title, "CurrentContent": currentQ.Content, "Answer": answer, "FollowUpIndex": followUpIndex})
	if err != nil {
		return nil, fmt.Errorf("failed to render clarifying follow-up prompt: %w", err)
	}
	response, err := s.chat(ctx, prompt, "chat", jsonObjectResponseFormat())
	if err != nil {
		return nil, fmt.Errorf("failed to generate clarifying follow-up: %w", err)
	}
	var q struct {
		Title          string `json:"title"`
		Content        string `json:"content"`
		ExpectedAnswer string `json:"expected_answer"`
	}
	if err := json.Unmarshal([]byte(response), &q); err != nil {
		return nil, fmt.Errorf("failed to parse clarifying follow-up: %w, body: %s", err, response)
	}
	result := &model.Question{Title: q.Title, Content: q.Content, ExpectedAnswer: q.ExpectedAnswer, Position: currentQ.Position, Difficulty: currentQ.Difficulty, Category: currentQ.Category}
	s.EnsureQuestionChinese(ctx, result)
	return result, nil
}

func (s *AIService) GenerateNextQuestionWithWeights(ctx context.Context, interview *model.Interview, previousAnswers []model.AnswerResult, capabilityGraph *model.JobCapabilityDimension) (*model.Question, error) {
	if capabilityGraph == nil {
		return s.GenerateNextQuestion(ctx, interview, previousAnswers)
	}
	var weightsBuilder strings.Builder
	weightsBuilder.WriteString("[Capability Weights]\n")
	weightsBuilder.WriteString(fmt.Sprintf("- %s: %d%%\n", capabilityGraph.Name, capabilityGraph.Weight))
	for _, sub := range capabilityGraph.SubDimensions {
		weightsBuilder.WriteString(fmt.Sprintf("  - %s (%d%%): %s\n", sub.Name, sub.Weight, strings.Join(sub.Tags, ", ")))
	}

	prompt, err := s.renderPrompt("generate_next_question_with_weights.tmpl", map[string]interface{}{
		"Position":                interview.Position,
		"Difficulty":              interview.Difficulty,
		"Mode":                    interview.Mode,
		"Style":                   interview.Style,
		"AnsweredCount":           len(previousAnswers),
		"AnsweredKnowledgePoints": buildAnsweredKnowledgePoints(previousAnswers),
		"ModeInstruction":         s.buildModePrompt(interview.Mode),
		"StyleInstruction":        s.buildStylePrompt(interview.Style, interview.Company),
		"DifficultyInstruction":   s.buildDifficultyPrompt(interview.Difficulty),
		"WeightsInstruction":      weightsBuilder.String(),
		"NextFocus":               "Choose a high-weight dimension not fully covered yet.",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to render weighted next question prompt: %w", err)
	}

	response, err := s.chat(ctx, prompt, "chat", jsonObjectResponseFormat())
	if err != nil {
		return nil, fmt.Errorf("failed to generate question: %w", err)
	}
	var question struct {
		Title          string `json:"title"`
		Content        string `json:"content"`
		ExpectedAnswer string `json:"expected_answer"`
	}
	if err := json.Unmarshal([]byte(response), &question); err != nil {
		return nil, fmt.Errorf("failed to parse question response: %w, body: %s", err, response)
	}
	result := &model.Question{Title: question.Title, Content: question.Content, ExpectedAnswer: question.ExpectedAnswer, Position: interview.Position, Difficulty: interview.Difficulty}
	s.EnsureQuestionChinese(ctx, result)
	return result, nil
}

func (s *AIService) GenerateNextQuestion(ctx context.Context, interview *model.Interview, previousAnswers []model.AnswerResult) (*model.Question, error) {
	prompt, err := s.renderPrompt("generate_next_question.tmpl", map[string]interface{}{
		"Position":                interview.Position,
		"Difficulty":              interview.Difficulty,
		"Mode":                    interview.Mode,
		"Style":                   interview.Style,
		"AnsweredCount":           len(previousAnswers),
		"AnsweredKnowledgePoints": buildAnsweredKnowledgePoints(previousAnswers),
		"ModeInstruction":         s.buildModePrompt(interview.Mode),
		"StyleInstruction":        s.buildStylePrompt(interview.Style, interview.Company),
		"DifficultyInstruction":   s.buildDifficultyPrompt(interview.Difficulty),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to render next question prompt: %w", err)
	}
	response, err := s.chat(ctx, prompt, "chat", jsonObjectResponseFormat())
	if err != nil {
		return nil, fmt.Errorf("failed to generate question: %w", err)
	}
	var question struct {
		Title          string `json:"title"`
		Content        string `json:"content"`
		ExpectedAnswer string `json:"expected_answer"`
	}
	if err := json.Unmarshal([]byte(response), &question); err != nil {
		return nil, fmt.Errorf("failed to parse question response: %w, body: %s", err, response)
	}
	result := &model.Question{Title: question.Title, Content: question.Content, ExpectedAnswer: question.ExpectedAnswer, Position: interview.Position, Difficulty: interview.Difficulty}
	s.EnsureQuestionChinese(ctx, result)
	return result, nil
}

func buildAnsweredKnowledgePoints(previousAnswers []model.AnswerResult) string {
	if len(previousAnswers) == 0 {
		return "（暂无）"
	}

	points := make([]string, 0, len(previousAnswers)*2)
	seen := make(map[string]struct{}, len(previousAnswers)*2)
	for _, item := range previousAnswers {
		title := strings.TrimSpace(item.Question.Title)
		if title != "" {
			if _, ok := seen[title]; !ok {
				seen[title] = struct{}{}
				points = append(points, title)
			}
		}

		category := strings.TrimSpace(item.Question.Category)
		if category != "" {
			label := "领域:" + category
			if _, ok := seen[label]; !ok {
				seen[label] = struct{}{}
				points = append(points, label)
			}
		}

		if len(points) >= 8 {
			break
		}
	}

	if len(points) == 0 {
		return "（暂无）"
	}
	return strings.Join(points, "；")
}

func (s *AIService) GenerateFollowUpQuestion(ctx context.Context, interview *model.Interview, currentQ *model.Question, answer string, ragContext string, followUpContext string, followUpIndex int) (*model.Question, string, error) {
	if currentQ == nil {
		return nil, "", fmt.Errorf("current question is nil")
	}

	mode, style, difficulty, company := "technical", "gentle", "campus_intern", ""
	if interview != nil {
		mode, style, difficulty, company = interview.Mode, interview.Style, interview.Difficulty, interview.Company
	}
	prompt, err := s.renderPrompt("generate_followup_question.tmpl", map[string]interface{}{
		"Mode":                  mode,
		"Style":                 style,
		"Difficulty":            difficulty,
		"ModeInstruction":       s.buildModePrompt(mode),
		"StyleInstruction":      s.buildStylePrompt(style, company),
		"DifficultyInstruction": s.buildDifficultyPrompt(difficulty),
		"CurrentTitle":          currentQ.Title,
		"CurrentContent":        currentQ.Content,
		"Answer":                answer,
		"RAGContext":            ragContext,
		"FollowUpContext":       strings.TrimSpace(followUpContext),
		"FollowUpIndex":         followUpIndex,
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to render follow-up prompt: %w", err)
	}
	response, err := s.chat(ctx, prompt, "chat", nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate follow-up: %w", err)
	}

	questionText, noFollowUp, reason := parseFollowUpModelOutput(response)
	if noFollowUp {
		return nil, reason, nil
	}
	if questionText == "" {
		return nil, "追问生成结果为空", nil
	}

	q := &model.Question{
		Title:          buildFollowUpTitle(currentQ.Title),
		Content:        ensureQuestionSentence(questionText),
		ExpectedAnswer: buildFollowUpExpectedAnswer(followUpContext, ragContext),
		Position:       currentQ.Position,
		Difficulty:     currentQ.Difficulty,
		Category:       currentQ.Category,
	}
	s.EnsureQuestionChinese(ctx, q)
	return q, reason, nil
}

func parseFollowUpModelOutput(response string) (question string, noFollowUp bool, reason string) {
	trimmed := sanitizeGeneratedText(strings.TrimSpace(stripOptionalCodeFence(sanitizeGeneratedText(response))))
	if trimmed == "" {
		return "", true, "模型未返回有效追问"
	}

	normalized := sanitizeGeneratedText(strings.Trim(strings.TrimSpace(trimmed), "\"'` "))
	if strings.EqualFold(normalized, "NO_FOLLOWUP") {
		return "", true, "模型判定当前无需继续追问"
	}

	if strings.HasPrefix(normalized, "{") {
		var legacy struct {
			FollowUpNeeded bool   `json:"follow_up_needed"`
			Reason         string `json:"reason"`
			Question       struct {
				Title          string `json:"title"`
				Content        string `json:"content"`
				ExpectedAnswer string `json:"expected_answer"`
			} `json:"question"`
		}
		if err := json.Unmarshal([]byte(normalized), &legacy); err == nil {
			if !legacy.FollowUpNeeded {
				r := strings.TrimSpace(legacy.Reason)
				if r == "" {
					r = "模型判定当前无需继续追问"
				}
				return "", true, r
			}

			line := strings.TrimSpace(firstNonEmpty(legacy.Question.Content, legacy.Question.Title))
			if line != "" {
				return normalizeFollowUpQuestionLine(line), false, strings.TrimSpace(legacy.Reason)
			}
		}
	}

	for _, line := range strings.Split(normalized, "\n") {
		cleaned := normalizeFollowUpQuestionLine(line)
		if cleaned == "" {
			continue
		}
		if strings.EqualFold(cleaned, "NO_FOLLOWUP") {
			return "", true, "模型判定当前无需继续追问"
		}
		return cleaned, false, "基于评估上下文生成追问"
	}

	return "", true, "模型未返回可用追问"
}

func normalizeFollowUpQuestionLine(line string) string {
	cleaned := sanitizeGeneratedText(strings.TrimSpace(line))
	if cleaned == "" {
		return ""
	}
	cleaned = sanitizeGeneratedText(strings.Trim(cleaned, "\"'` "))
	for _, prefix := range []string{"问题：", "追问：", "Q:", "q:"} {
		if strings.HasPrefix(cleaned, prefix) {
			cleaned = strings.TrimSpace(strings.TrimPrefix(cleaned, prefix))
		}
	}
	cleaned = strings.TrimLeft(cleaned, "-、0123456789.． ")
	return sanitizeGeneratedText(strings.TrimSpace(cleaned))
}

func ensureQuestionSentence(text string) string {
	normalized := sanitizeGeneratedText(strings.TrimSpace(text))
	if normalized == "" {
		return ""
	}
	if strings.HasSuffix(normalized, "？") || strings.HasSuffix(normalized, "?") {
		return normalized
	}
	return normalized + "？"
}

func buildFollowUpTitle(currentTitle string) string {
	base := sanitizeGeneratedText(strings.TrimSpace(currentTitle))
	if base == "" {
		return "追问深入"
	}
	return "追问深入：" + base
}

func buildFollowUpExpectedAnswer(followUpContext, ragContext string) string {
	context := sanitizeGeneratedText(strings.TrimSpace(followUpContext))
	if context == "" {
		context = "候选人回答中需要进一步验证的技术点"
	}
	expected := fmt.Sprintf("回答应围绕“%s”，说明核心原理、实现步骤、边界条件与技术取舍。", context)
	if strings.TrimSpace(ragContext) != "" {
		expected += "可结合知识上下文中的依据进行论证。"
	}
	return expected
}

func (s *AIService) buildModePrompt(mode string) string {
	switch mode {
	case "technical":
		return s.getPromptOrDefault("ai/question/mode/technical", "")
	case "hr":
		return s.getPromptOrDefault("ai/question/mode/hr", "")
	case "comprehensive":
		return s.getPromptOrDefault("ai/question/mode/comprehensive", "")
	default:
		return ""
	}
}

func (s *AIService) buildStylePrompt(style, company string) string {
	base := s.getPromptOrDefault("ai/question/style/default", "")
	switch style {
	case "gentle":
		base = s.getPromptOrDefault("ai/question/style/gentle", "")
	case "stress":
		base = s.getPromptOrDefault("ai/question/style/stress", "")
	case "deep":
		base = s.getPromptOrDefault("ai/question/style/deep", "")
	case "practical":
		base = s.getPromptOrDefault("ai/question/style/practical", "")
	case "algorithm":
		base = s.getPromptOrDefault("ai/question/style/algorithm", "")
	}
	if company != "" {
		companySuffixTemplate := s.getPromptOrDefault("ai/question/style/company_suffix", "{{.Company}}")
		rendered, err := s.renderDynamicPrompt("ai/question/style/company_suffix", map[string]interface{}{"Company": company})
		if err == nil && strings.TrimSpace(rendered) != "" {
			if strings.TrimSpace(base) == "" {
				return rendered
			}
			return base + " " + rendered
		}
		resolved := strings.TrimSpace(strings.ReplaceAll(companySuffixTemplate, "{{.Company}}", company))
		if strings.TrimSpace(base) == "" {
			return resolved
		}
		if resolved == "" {
			return base
		}
		return base + " " + resolved
	}
	return base
}

func (s *AIService) buildDifficultyPrompt(difficulty string) string {
	switch difficulty {
	case "campus_intern":
		return s.getPromptOrDefault("ai/question/difficulty/campus_intern", "")
	case "campus_graduate":
		return s.getPromptOrDefault("ai/question/difficulty/campus_graduate", "")
	case "social_junior":
		return s.getPromptOrDefault("ai/question/difficulty/social_junior", "")
	default:
		return ""
	}
}

func GenerateRandomStyleForInterview() (style string, company string) {
	styles := []string{"gentle", "stress", "deep", "practical"}
	companies := []string{"", "ali", "bytedance", "tencent", "meituan", ""}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	style = styles[rng.Intn(len(styles))]
	company = companies[rng.Intn(len(companies))]
	return
}
