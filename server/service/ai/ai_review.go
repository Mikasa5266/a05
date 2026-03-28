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
		result.Comment = "Answer has partial coverage with room to improve depth and clarity."
	}
	if strings.TrimSpace(result.Suggestion) == "" {
		result.Suggestion = "Use conclusion -> principle -> example -> boundary structure."
	}
	if strings.TrimSpace(result.ModelAnswerOutline) == "" {
		result.ModelAnswerOutline = defaultModelAnswerOutline(expectedAnswer)
	}
	if strings.TrimSpace(result.FollowUp) == "" {
		result.FollowUp = defaultFollowUpQuestion(question)
	}

	return &result, nil
}

func BuildStrictEvalPrompt(question, expectedAnswer, answer string) string {
	expected := strings.TrimSpace(expectedAnswer)
	if expected == "" {
		expected = "(No expected answer provided)"
	}
	return fmt.Sprintf("Review the candidate answer strictly and return a JSON object with score, dimensions, comment, suggestion, highlights, gaps, model_answer_outline, follow_up.\nQuestion: %s\nExpected: %s\nAnswer: %s", question, expected, answer)
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
	return "Define concept, explain mechanism, show one real scenario, discuss boundaries and trade-offs."
}

func defaultFollowUpQuestion(question string) string {
	q := strings.TrimSpace(question)
	if q == "" {
		return "Can you provide a concrete implementation detail and trade-off?"
	}
	return "Can you explain one implementation detail and why you chose it?"
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
