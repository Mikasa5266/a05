package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ReviewResult struct {
	Score              int               `json:"score"`
	Comment            string            `json:"comment"`
	Suggestion         string            `json:"suggestion"`
	Dimensions         *ReviewDimensions `json:"dimensions,omitempty"`
	Highlights         []string          `json:"highlights,omitempty"`
	Gaps               []string          `json:"gaps,omitempty"`
	ModelAnswerOutline string            `json:"model_answer_outline,omitempty"`
	FollowUp           string            `json:"follow_up,omitempty"`
}

type ReviewDimensions struct {
	TechnicalDepth int `json:"technical_depth"`
	Expression     int `json:"expression"`
	Logic          int `json:"logic"`
	Completeness   int `json:"completeness"`
}

func IsInvalidAnswer(answer string) bool {
	a := strings.TrimSpace(strings.ToLower(answer))
	if a == "" {
		return true
	}
	bad := []string{"不会", "不知道", "i don't know", "dont know", "no idea", "随便", "asd", "123"}
	for _, b := range bad {
		if strings.Contains(a, b) {
			return true
		}
	}
	if len([]rune(a)) <= 3 {
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

	if len([]rune(strings.TrimSpace(answer))) < 12 && keywordOverlap(question+" "+expectedAnswer, answer) < 0.05 {
		result.Score = 0
		result.Dimensions = &ReviewDimensions{}
		result.Gaps = []string{"Answer is too short", "No substantial technical details"}
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

func keywordOverlap(reference, answer string) float64 {
	refWords := tokenize(reference)
	ansWords := tokenize(answer)
	if len(refWords) == 0 || len(ansWords) == 0 {
		return 0
	}
	hit := 0
	for w := range refWords {
		if _, ok := ansWords[w]; ok {
			hit++
		}
	}
	return float64(hit) / float64(len(refWords))
}

func tokenize(s string) map[string]struct{} {
	s = strings.ToLower(strings.TrimSpace(s))
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == '\n' || r == '\t' || r == ',' || r == '.' || r == ';' || r == ':' || r == '，' || r == '。' || r == '；' || r == '：' || r == '(' || r == ')' || r == '（' || r == '）'
	})
	out := make(map[string]struct{}, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if len([]rune(p)) < 2 {
			continue
		}
		out[p] = struct{}{}
	}
	return out
}
