package ai

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

type ReviewResult struct {
	Score              int               `json:"score"`
	Comment            string            `json:"comment"`
	Suggestion         string            `json:"suggestion"`
	Dimensions         *ReviewDimensions `json:"dimensions,omitempty"`
	KnowledgeRetrieval []KnowledgeCheck  `json:"knowledge_retrieval,omitempty"`
	Reasoning          string            `json:"reasoning,omitempty"`
	Scores             *RubricScores     `json:"scores,omitempty"`
	FinalScore         int               `json:"final_score,omitempty"`
	ShouldFollowUp     bool              `json:"should_follow_up,omitempty"`
	FollowUpContext    string            `json:"follow_up_context,omitempty"`
	Highlights         []string          `json:"highlights,omitempty"`
	Gaps               []string          `json:"gaps,omitempty"`
	ModelAnswerOutline string            `json:"model_answer_outline,omitempty"`
	FollowUp           string            `json:"follow_up,omitempty"`
}

type KnowledgeCheck struct {
	Point    string `json:"point"`
	Verdict  string `json:"verdict"`
	Evidence string `json:"evidence"`
}

type RubricScores struct {
	TechnicalAccuracy int `json:"technical_accuracy"`
	LogicalClarity    int `json:"logical_clarity"`
	Completeness      int `json:"completeness"`
	Groundedness      int `json:"groundedness"`
}

type ReviewDimensions struct {
	TechnicalDepth int `json:"technical_depth"`
	Expression     int `json:"expression"`
	Logic          int `json:"logic"`
	Completeness   int `json:"completeness"`
}

func IsInvalidAnswer(answer string) bool {
	normalized := normalizeAnswerForValidation(answer)
	if len([]rune(normalized)) < 5 {
		return true
	}
	if isMeaninglessGibberish(normalized) {
		return true
	}
	return false
}

func EvaluateCandidateAnswer(question, expectedAnswer, answer string, llmCallFunc func(prompt string) (string, error)) (*ReviewResult, error) {
	if IsInvalidAnswer(answer) {
		return &ReviewResult{
			Score:      0,
			Comment:    "Invalid or empty answer.",
			Suggestion: "Provide key concept, mechanism, and one practical example.",
			Dimensions: &ReviewDimensions{},
		}, nil
	}

	prompt := BuildStrictEvalPrompt(question, expectedAnswer, answer)
	raw, err := llmCallFunc(prompt)
	if err != nil {
		return nil, fmt.Errorf("llm review failed: %w", err)
	}

	var result ReviewResult
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &result); err != nil {
		return nil, fmt.Errorf("invalid review json: %w", err)
	}

	if result.Dimensions == nil {
		result.Dimensions = &ReviewDimensions{TechnicalDepth: result.Score, Expression: result.Score, Logic: result.Score, Completeness: result.Score}
	}
	result.Score = clampReviewScore(result.Score)
	result.Dimensions.TechnicalDepth = clampReviewScore(result.Dimensions.TechnicalDepth)
	result.Dimensions.Expression = clampReviewScore(result.Dimensions.Expression)
	result.Dimensions.Logic = clampReviewScore(result.Dimensions.Logic)
	result.Dimensions.Completeness = clampReviewScore(result.Dimensions.Completeness)

	if strings.TrimSpace(result.Comment) == "" {
		result.Comment = "回答已覆盖部分要点，但在深度和清晰度上仍有提升空间。"
	}
	if strings.TrimSpace(result.Suggestion) == "" {
		result.Suggestion = "建议按“结论 -> 原理 -> 示例 -> 边界”结构作答。"
	}
	if strings.TrimSpace(result.ModelAnswerOutline) == "" {
		result.ModelAnswerOutline = defaultModelAnswerOutline(expectedAnswer)
	}
	if strings.TrimSpace(result.FollowUp) == "" {
		result.FollowUp = defaultFollowUpQuestion(question)
	}
	if result.ShouldFollowUp && strings.TrimSpace(result.FollowUpContext) == "" {
		result.FollowUpContext = "请围绕题目要求中的关键机制继续展开，补充实现细节与边界条件。"
	}

	return &result, nil
}

func BuildStrictEvalPrompt(question, expectedAnswer, answer string) string {
	expected := strings.TrimSpace(expectedAnswer)
	if expected == "" {
		expected = "(No expected answer provided)"
	}
	if prompt, ok := tryRenderDynamicPrompt("ai/review/strict_eval", map[string]interface{}{
		"Question": question,
		"Expected": expected,
		"Answer":   answer,
	}); ok {
		return prompt
	}
	fallback := getPromptOrFallback("ai/review/strict_eval_fallback", "")
	if prompt, ok := tryRenderDynamicPrompt("ai/review/strict_eval_fallback", map[string]interface{}{
		"Question": question,
		"Expected": expected,
		"Answer":   answer,
	}); ok {
		return prompt
	}
	if strings.TrimSpace(fallback) != "" {
		return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(fallback, "{{.Question}}", question), "{{.Expected}}", expected), "{{.Answer}}", answer)
	}
	return fmt.Sprintf("Question: %s\nExpected: %s\nAnswer: %s", question, expected, answer)
}

func clampReviewScore(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func defaultModelAnswerOutline(expected string) string {
	if strings.TrimSpace(expected) != "" {
		return strings.TrimSpace(expected)
	}
	return "先给出定义，再说明机制，随后结合一个真实场景，最后补充边界与取舍。"
}

func defaultFollowUpQuestion(question string) string {
	q := strings.TrimSpace(question)
	if q == "" {
		return "请补充一个关键实现细节，并说明对应取舍。"
	}
	return "请继续说明一个关键实现细节，以及你为何做出该技术选择。"
}

func normalizeAnswerForValidation(answer string) string {
	lower := strings.ToLower(strings.TrimSpace(answer))
	if lower == "" {
		return ""
	}
	var builder strings.Builder
	builder.Grow(len(lower))
	for _, r := range lower {
		if unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r) {
			continue
		}
		builder.WriteRune(r)
	}
	return builder.String()
}

func isMeaninglessGibberish(text string) bool {
	if text == "" {
		return true
	}
	onlyDigits := true
	semanticCharCount := 0
	unique := make(map[rune]struct{})
	for _, r := range text {
		unique[r] = struct{}{}
		if !unicode.IsDigit(r) {
			onlyDigits = false
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			semanticCharCount++
		}
	}
	if onlyDigits {
		return true
	}
	if semanticCharCount == 0 {
		return true
	}
	if len(unique) <= 1 {
		return true
	}
	return false
}
