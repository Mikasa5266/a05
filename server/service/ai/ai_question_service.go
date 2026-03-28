package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"your-project/model"
)

func (s *AIService) GenerateQuestions(ctx context.Context, interview *model.Interview, count int) ([]*model.Question, error) {
	prompt, err := s.renderPrompt("generate_questions.tmpl", map[string]interface{}{
		"Count":                 count,
		"Position":              interview.Position,
		"Difficulty":            interview.Difficulty,
		"Mode":                  interview.Mode,
		"Style":                 interview.Style,
		"ModeInstruction":       buildModePrompt(interview.Mode),
		"StyleInstruction":      buildStylePrompt(interview.Style, interview.Company),
		"DifficultyInstruction": buildDifficultyPrompt(interview.Difficulty),
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
		"ModeInstruction":       buildModePrompt(interview.Mode),
		"StyleInstruction":      buildStylePrompt(interview.Style, interview.Company),
		"DifficultyInstruction": buildDifficultyPrompt(interview.Difficulty),
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
		topic = "topic"
	}
	q.Title = fmt.Sprintf("%s: core principles and practice", topic)
	q.Content = fmt.Sprintf("Please explain %s including concepts, mechanisms and practical scenarios.", topic)
	if strings.TrimSpace(q.ExpectedAnswer) == "" {
		q.ExpectedAnswer = fmt.Sprintf("Cover definition, mechanism, constraints and practical trade-offs of %s.", topic)
	}
}

func isFollowUpWording(text string) bool {
	t := strings.TrimSpace(strings.ToLower(text))
	if t == "" {
		return false
	}
	patterns := []string{"you mentioned", "continue", "further", "follow-up", "based on previous"}
	for _, p := range patterns {
		if strings.Contains(t, p) {
			return true
		}
	}
	return false
}

var openingQuestionContextPatterns = []string{"previous", "above", "continue", "further", "those", "these"}
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
		topic = "topic"
	}
	question.Title = fmt.Sprintf("%s: core principles and practice", topic)
	question.Content = fmt.Sprintf("Please explain %s from concept, mechanism, thread safety and performance.", topic)
	if strings.TrimSpace(question.ExpectedAnswer) == "" {
		question.ExpectedAnswer = fmt.Sprintf("Should cover definition, implementation, boundaries and trade-offs of %s.", topic)
	}
	question.Title = strings.TrimSpace(question.Title)
	question.Content = strings.TrimSpace(question.Content)
	question.ExpectedAnswer = strings.TrimSpace(question.ExpectedAnswer)
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
		"Position":              interview.Position,
		"Difficulty":            interview.Difficulty,
		"Mode":                  interview.Mode,
		"Style":                 interview.Style,
		"AnsweredCount":         len(previousAnswers),
		"ModeInstruction":       buildModePrompt(interview.Mode),
		"StyleInstruction":      buildStylePrompt(interview.Style, interview.Company),
		"DifficultyInstruction": buildDifficultyPrompt(interview.Difficulty),
		"WeightsInstruction":    weightsBuilder.String(),
		"NextFocus":             "Choose a high-weight dimension not fully covered yet.",
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
		"Position":              interview.Position,
		"Difficulty":            interview.Difficulty,
		"Mode":                  interview.Mode,
		"Style":                 interview.Style,
		"AnsweredCount":         len(previousAnswers),
		"ModeInstruction":       buildModePrompt(interview.Mode),
		"StyleInstruction":      buildStylePrompt(interview.Style, interview.Company),
		"DifficultyInstruction": buildDifficultyPrompt(interview.Difficulty),
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

func (s *AIService) GenerateFollowUpQuestion(ctx context.Context, interview *model.Interview, currentQ *model.Question, answer string, ragContext string, followUpIndex int) (*model.Question, string, error) {
	mode, style, difficulty, company := "technical", "gentle", "campus_intern", ""
	if interview != nil {
		mode, style, difficulty, company = interview.Mode, interview.Style, interview.Difficulty, interview.Company
	}
	prompt, err := s.renderPrompt("generate_followup_question.tmpl", map[string]interface{}{
		"Mode":                  mode,
		"Style":                 style,
		"Difficulty":            difficulty,
		"ModeInstruction":       buildModePrompt(mode),
		"StyleInstruction":      buildStylePrompt(style, company),
		"DifficultyInstruction": buildDifficultyPrompt(difficulty),
		"CurrentTitle":          currentQ.Title,
		"CurrentContent":        currentQ.Content,
		"Answer":                answer,
		"RAGContext":            ragContext,
		"FollowUpIndex":         followUpIndex,
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to render follow-up prompt: %w", err)
	}
	response, err := s.chat(ctx, prompt, "chat", jsonObjectResponseFormat())
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate follow-up: %w", err)
	}
	var result struct {
		FollowUpNeeded bool   `json:"follow_up_needed"`
		Reason         string `json:"reason"`
		Question       struct {
			Title          string `json:"title"`
			Content        string `json:"content"`
			ExpectedAnswer string `json:"expected_answer"`
		} `json:"question"`
	}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		log.Printf("failed to parse follow-up response: %v, body: %s", err, response)
		return nil, "", nil
	}
	if !result.FollowUpNeeded {
		return nil, result.Reason, nil
	}
	q := &model.Question{Title: result.Question.Title, Content: result.Question.Content, ExpectedAnswer: result.Question.ExpectedAnswer, Position: currentQ.Position, Difficulty: currentQ.Difficulty, Category: currentQ.Category}
	s.EnsureQuestionChinese(ctx, q)
	return q, result.Reason, nil
}

func buildModePrompt(mode string) string {
	switch mode {
	case "technical":
		return "Technical interview only."
	case "hr":
		return "HR interview only."
	case "comprehensive":
		return "Mix technical and HR questions."
	default:
		return ""
	}
}

func buildStylePrompt(style, company string) string {
	base := "Standard interview style."
	switch style {
	case "gentle":
		base = "Friendly and guiding style."
	case "stress":
		base = "High-pressure style with strict follow-ups."
	case "deep":
		base = "Deep technical probing style."
	case "practical":
		base = "Project and practical problem style."
	case "algorithm":
		base = "Algorithm-focused style."
	}
	if company != "" {
		return base + " Company profile: " + company
	}
	return base
}

func buildDifficultyPrompt(difficulty string) string {
	switch difficulty {
	case "campus_intern":
		return "Intern-level questions with fundamentals."
	case "campus_graduate":
		return "Graduate-level with moderate depth."
	case "social_junior":
		return "Junior social hiring with practical depth."
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
