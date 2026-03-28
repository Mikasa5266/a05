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
	return s.EnsureChineseOutput(ctx, response, "本场面试整体表现中等，建议继续提升回答深度与结构化表达。"), nil
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

	response, err := s.chat(ctx, prompt, "report", jsonObjectResponseFormat())
	if err != nil {
		return nil, err
	}

	var insights ReportInsights
	if err := parseReportInsightsResponse(response, &insights); err != nil {
		return nil, fmt.Errorf("failed to parse report insights: %w", err)
	}

	insights.OverallAnalysis = s.EnsureChineseOutput(ctx, insights.OverallAnalysis, "本场面试整体表现中等，建议继续提升回答深度与结构化表达。")
	insights.Strengths = ensureChineseList(ctx, s, insights.Strengths, []string{"具备基础知识框架", "表达态度积极"})
	insights.Weaknesses = ensureChineseList(ctx, s, insights.Weaknesses, []string{"回答深度不足", "关键细节覆盖不完整"})
	insights.Suggestions = ensureChineseList(ctx, s, insights.Suggestions, []string{"建议按“结论-原理-实践-边界”结构回答", "补充项目量化结果与异常场景处理"})
	insights.QADetails = normalizeReportQADetails(ctx, s, insights.QADetails)

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
		clean = append(clean, s.EnsureChineseOutput(ctx, line, "请给出更具体的分析。"))
	}
	if len(clean) == 0 {
		return fallback
	}
	if len(clean) > 4 {
		return clean[:4]
	}
	return clean
}

func parseReportInsightsResponse(raw string, out *ReportInsights) error {
	if out == nil {
		return fmt.Errorf("report insights receiver is nil")
	}

	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return fmt.Errorf("report insights response is empty")
	}

	candidates := make([]string, 0, 3)
	candidates = append(candidates, trimmed)

	if unfenced := strings.TrimSpace(stripOptionalCodeFence(trimmed)); unfenced != "" && unfenced != trimmed {
		candidates = append(candidates, unfenced)
	}

	if start := strings.Index(trimmed, "{"); start >= 0 {
		if end := strings.LastIndex(trimmed, "}"); end > start {
			candidates = append(candidates, strings.TrimSpace(trimmed[start:end+1]))
		}
	}

	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}

		var parsed ReportInsights
		if err := json.Unmarshal([]byte(candidate), &parsed); err == nil {
			*out = parsed
			return nil
		}
	}

	return fmt.Errorf("invalid report insights response")
}

func normalizeReportQADetails(ctx context.Context, s *AIService, items []ReportQADetail) []ReportQADetail {
	if len(items) == 0 {
		return []ReportQADetail{}
	}

	normalized := make([]ReportQADetail, 0, len(items))
	for _, item := range items {
		question := strings.TrimSpace(item.Question)
		userAnswer := strings.TrimSpace(item.UserAnswer)
		optimizedAnswer := strings.TrimSpace(item.OptimizedAnswer)
		if question == "" || (userAnswer == "" && optimizedAnswer == "") {
			continue
		}

		if userAnswer == "" {
			userAnswer = "候选人的作答信息不足，建议补充关键技术要点。"
		}
		if optimizedAnswer == "" {
			optimizedAnswer = "建议按“结论-原理-实践-边界”结构组织回答，并补充工程细节。"
		}

		normalized = append(normalized, ReportQADetail{
			Question:        s.EnsureChineseOutput(ctx, question, "请回顾原题目。"),
			UserAnswer:      s.EnsureChineseOutput(ctx, userAnswer, "候选人回答摘要暂缺。"),
			OptimizedAnswer: s.EnsureChineseOutput(ctx, optimizedAnswer, "建议补充高质量优化答案。"),
			KeyImprovements: normalizeImprovementList(item.KeyImprovements),
		})

		if len(normalized) >= 12 {
			break
		}
	}

	return normalized
}

func normalizeImprovementList(items []string) []string {
	clean := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		line := strings.TrimSpace(item)
		if line == "" {
			continue
		}
		if _, ok := seen[line]; ok {
			continue
		}
		seen[line] = struct{}{}
		clean = append(clean, line)
		if len(clean) >= 4 {
			break
		}
	}

	if len(clean) == 0 {
		return []string{"补充关键机制说明", "增加边界条件与异常处理"}
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
