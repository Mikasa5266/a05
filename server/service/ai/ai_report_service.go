package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"your-project/model"
)

func (s *AIService) GenerateOverallAnalysis(ctx context.Context, interview *model.Interview, answers []model.AnswerResult) (string, error) {
	summary := struct {
		Position   string               `json:"position"`
		Difficulty string               `json:"difficulty"`
		Answers    []model.AnswerResult `json:"answers"`
	}{Position: interview.Position, Difficulty: interview.Difficulty, Answers: answers}

	payload, err := json.Marshal(summary)
	if err != nil {
		return "", fmt.Errorf("failed to marshal summary: %w", err)
	}

	prompt, err := s.renderPrompt("generate_overall_analysis.tmpl", map[string]interface{}{"PayloadJSON": string(payload)})
	if err != nil {
		return "", fmt.Errorf("failed to render overall analysis prompt: %w", err)
	}

	response, err := s.chat(ctx, prompt, "report", jsonObjectResponseFormat())
	if err != nil {
		return "", err
	}
	return s.EnsureChineseOutput(ctx, response, "Overall performance is moderate. Improve depth and structure."), nil
}

func (s *AIService) GenerateReportInsights(ctx context.Context, interview *model.Interview, answers []model.AnswerResult) (*ReportInsights, error) {
	payload := struct {
		Position   string               `json:"position"`
		Difficulty string               `json:"difficulty"`
		Mode       string               `json:"mode"`
		Style      string               `json:"style"`
		Answers    []model.AnswerResult `json:"answers"`
	}{Position: interview.Position, Difficulty: interview.Difficulty, Mode: interview.Mode, Style: interview.Style, Answers: answers}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal report payload: %w", err)
	}

	prompt, err := s.renderPrompt("generate_report_insights.tmpl", map[string]interface{}{"PayloadJSON": string(body)})
	if err != nil {
		return nil, fmt.Errorf("failed to render report insights prompt: %w", err)
	}

	response, err := s.chat(ctx, prompt, "report", nil)
	if err != nil {
		return nil, err
	}

	var insights ReportInsights
	if err := json.Unmarshal([]byte(response), &insights); err != nil {
		return nil, fmt.Errorf("failed to parse report insights: %w", err)
	}

	insights.OverallAnalysis = s.EnsureChineseOutput(ctx, insights.OverallAnalysis, "Performance is moderate with room for improvement.")
	insights.Strengths = ensureChineseList(ctx, s, insights.Strengths, []string{"Basic knowledge exists", "Positive communication"})
	insights.Weaknesses = ensureChineseList(ctx, s, insights.Weaknesses, []string{"Insufficient depth", "Missing key details"})
	insights.Suggestions = ensureChineseList(ctx, s, insights.Suggestions, []string{"Use conclusion-principle-practice structure", "Add project examples and metrics"})

	insights.TechnicalScore = clampScore(insights.TechnicalScore)
	insights.ExpressionScore = clampScore(insights.ExpressionScore)
	insights.LogicScore = clampScore(insights.LogicScore)
	insights.MatchingScore = clampScore(insights.MatchingScore)
	insights.BehaviorScore = clampScore(insights.BehaviorScore)

	return &insights, nil
}

func ensureChineseList(ctx context.Context, s *AIService, items []string, fallback []string) []string {
	clean := make([]string, 0, len(items))
	for _, item := range items {
		line := strings.TrimSpace(item)
		if line == "" {
			continue
		}
		clean = append(clean, s.EnsureChineseOutput(ctx, line, "Please provide more specific analysis."))
	}
	if len(clean) == 0 {
		return fallback
	}
	if len(clean) > 4 {
		return clean[:4]
	}
	return clean
}

func clampScore(value int) int {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}
