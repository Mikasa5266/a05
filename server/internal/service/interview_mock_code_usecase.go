package service

import (
	"fmt"
	"strings"
	"time"

	"your-project/internal/model"
)

func SubmitMockCode(userID, interviewID, questionID uint, questionTitle, questionContent, code, language string) (*model.AnswerResult, error) {
	svc := NewInterviewService()
	return svc.SubmitMockCode(userID, interviewID, questionID, questionTitle, questionContent, code, language)
}

func (s *InterviewService) SubmitMockCode(userID, interviewID, questionID uint, questionTitle, questionContent, code, language string) (*model.AnswerResult, error) {
	trimmedCode := strings.TrimSpace(code)
	if trimmedCode == "" {
		return nil, fmt.Errorf("code is required")
	}

	interview, err := s.GetInterviewByID(userID, interviewID)
	if err != nil {
		return nil, err
	}

	if normalizeStatusValue(interview.Status) != interviewStatusInProgress {
		return nil, fmt.Errorf("interview is not in progress")
	}

	question, err := s.resolveMockCodeQuestion(interview, questionID, questionTitle, questionContent)
	if err != nil {
		return nil, err
	}

	score, feedback := evaluateMockCode(trimmedCode, language)
	result := &model.AnswerResult{
		InterviewID: interviewID,
		QuestionID:  question.ID,
		Answer:      trimmedCode,
		Score:       score,
		Feedback:    feedback,
		CreatedAt:   time.Now(),
	}

	if err := s.interviewRepo.SaveAnswer(result); err != nil {
		return nil, fmt.Errorf("failed to save mock code answer: %w", err)
	}

	return result, nil
}

func (s *InterviewService) resolveMockCodeQuestion(interview *model.Interview, questionID uint, questionTitle, questionContent string) (*model.Question, error) {
	if questionID > 0 {
		question, err := s.questionRepo.GetByID(questionID)
		if err == nil && question != nil {
			return question, nil
		}
	}

	title := strings.TrimSpace(questionTitle)
	if title == "" {
		title = "现场编程题"
	}

	content := strings.TrimSpace(questionContent)
	if content == "" {
		content = "请实现核心逻辑，并解释复杂度与边界处理。"
	}

	position := strings.TrimSpace(interview.Position)
	if position == "" {
		position = "Java后端工程师"
	}

	difficulty := strings.TrimSpace(interview.Difficulty)
	if difficulty == "" {
		difficulty = "campus_intern"
	}

	question := &model.Question{
		Position:    position,
		Difficulty:  difficulty,
		Category:    "现场编程",
		Title:       title,
		Content:     content,
		Source:      "live_code_mock",
		RAGEligible: false,
	}

	if err := s.questionRepo.Create(question); err != nil {
		return nil, fmt.Errorf("failed to create mock code question: %w", err)
	}

	return question, nil
}

func evaluateMockCode(code, language string) (int, string) {
	normalizedCode := strings.TrimSpace(code)
	loweredCode := strings.ToLower(normalizedCode)
	loweredLanguage := strings.ToLower(strings.TrimSpace(language))

	score := 35
	highlights := make([]string, 0, 4)
	improvements := make([]string, 0, 4)

	lineCount := strings.Count(normalizedCode, "\n") + 1
	switch {
	case lineCount >= 20:
		score += 22
		highlights = append(highlights, "代码结构较完整")
	case lineCount >= 10:
		score += 16
		highlights = append(highlights, "已具备基础实现")
	case lineCount >= 5:
		score += 10
	default:
		score += 4
		improvements = append(improvements, "建议补充更完整的实现细节")
	}

	if containsAny(loweredCode, "func ", "function ", "def ", "class ", "=>") {
		score += 12
		highlights = append(highlights, "包含核心函数定义")
	} else {
		improvements = append(improvements, "建议补充明确的函数入口")
	}

	if containsAny(loweredCode, "for ", "while ", "range", "foreach") {
		score += 10
		highlights = append(highlights, "包含循环处理逻辑")
	} else {
		improvements = append(improvements, "可考虑补充迭代处理逻辑")
	}

	if containsAny(loweredCode, "if ", "else", "switch", "case") {
		score += 8
		highlights = append(highlights, "具备分支判断")
	} else {
		improvements = append(improvements, "建议补充边界判断")
	}

	if containsAny(loweredCode, "return", "yield") {
		score += 8
	} else {
		improvements = append(improvements, "建议补充返回结果")
	}

	if containsAny(loweredCode, "//", "#", "/*") {
		score += 5
		highlights = append(highlights, "有一定可读性注释")
	}

	if containsAny(loweredCode, "todo", "pass", "待实现", "stub") {
		score -= 14
		improvements = append(improvements, "仍有未完成占位逻辑")
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	if len(highlights) == 0 {
		highlights = append(highlights, "已提交代码骨架")
	}
	if len(improvements) == 0 {
		improvements = append(improvements, "可继续补充复杂度分析与测试用例")
	}

	feedback := fmt.Sprintf(
		"Mock评分（%s）：%d 分。亮点：%s。建议：%s。",
		nonEmptyOrDefault(loweredLanguage, "unknown"),
		score,
		strings.Join(uniqueMockStrings(highlights), "；"),
		strings.Join(uniqueMockStrings(improvements), "；"),
	)

	return score, feedback
}

func containsAny(content string, keys ...string) bool {
	for _, key := range keys {
		if strings.Contains(content, key) {
			return true
		}
	}
	return false
}

func uniqueMockStrings(items []string) []string {
	if len(items) == 0 {
		return items
	}
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func nonEmptyOrDefault(val, fallback string) string {
	trimmed := strings.TrimSpace(val)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}
