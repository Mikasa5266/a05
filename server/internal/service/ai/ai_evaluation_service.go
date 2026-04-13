package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"strings"

	"your-project/internal/model"
)

const (
	evaluationGroundTruthTopK    = 5
	evaluationGroundTruthMaxRune = 3600
)

type MockGroupAssessmentTemplate struct {
	TemplateVersion string   `json:"template_version"`
	AssessmentType  string   `json:"assessment_type"`
	Leadership      int      `json:"leadership"`
	Expression      int      `json:"expression"`
	Teamwork        int      `json:"teamwork"`
	Logic           int      `json:"logic"`
	TotalScore      int      `json:"total_score"`
	KeySignals      []string `json:"key_signals"`
	PreliminaryNote string   `json:"preliminary_note"`
}

// FUTURE-LINK: RAG_EVAL_INTEGRATION
type RealRAGAnalysis interface {
	AnalyzeGroupAssessment(ctx context.Context, question *model.Question, answer string, participantResumeContext string) (*MockGroupAssessmentTemplate, error)
}

type groundedEvalLLMResponse struct {
	Score           int    `json:"score"`
	Reasoning       string `json:"reasoning"`
	ShouldFollowUp  bool   `json:"should_follow_up"`
	FollowUpContext string `json:"follow_up_context"`
}

func (s *AIService) EvaluateAnswer(ctx context.Context, question *model.Question, answer string) (*EvaluationResult, error) {
	if isGroupAssessmentQuestion(question) {
		return s.evaluateGroupAnswerWithMockTemplate(ctx, question, answer), nil
	}

	if question == nil {
		return s.evaluateAnswerLocal(nil, answer), nil
	}

	groundTruth := s.buildEvaluationGroundTruth(ctx, question)
	review, err := s.evaluateAnswerWithGroundTruth(ctx, question, answer, groundTruth)
	if err != nil {
		log.Printf("AI grounded review failed, fallback to local heuristic: %v", err)
		return s.evaluateAnswerLocal(question, answer), nil
	}

	review.Score = clampScore(review.Score)
	review.FinalScore = clampScore(review.FinalScore)
	if review.FinalScore == 0 {
		review.FinalScore = review.Score
	}

	evaluation := s.EnsureChineseOutput(ctx, firstNonEmpty(strings.TrimSpace(review.Comment), strings.TrimSpace(review.Reasoning)), "回答已接收，当前质量仍有提升空间。")
	reasoning := s.EnsureChineseOutput(ctx, firstNonEmpty(strings.TrimSpace(review.Reasoning), strings.TrimSpace(review.Comment)), "回答与参考知识存在差异，请补充关键原理与证据。")
	suggestion := s.EnsureChineseOutput(ctx, review.Suggestion, "请补充核心原理、实现细节和边界条件。")

	feedback := map[string]interface{}{
		"evaluation":           evaluation,
		"reasoning":            reasoning,
		"suggestions":          splitSuggestionText(suggestion),
		"dimensions":           review.Dimensions,
		"knowledge_retrieval":  review.KnowledgeRetrieval,
		"scores":               review.Scores,
		"final_score":          review.FinalScore,
		"should_follow_up":     review.ShouldFollowUp,
		"follow_up_context":    strings.TrimSpace(review.FollowUpContext),
		"ground_truth":         groundTruth,
		"highlights":           review.Highlights,
		"gaps":                 review.Gaps,
		"model_answer_outline": review.ModelAnswerOutline,
		"follow_up":            review.FollowUp,
	}
	feedbackJSON, _ := json.Marshal(feedback)
	return &EvaluationResult{Score: review.Score, Feedback: string(feedbackJSON)}, nil
}

func isGroupAssessmentQuestion(question *model.Question) bool {
	if question == nil {
		return false
	}
	joined := strings.ToLower(strings.TrimSpace(question.Category + " " + question.Title + " " + question.Content))
	if strings.Contains(joined, "group_") {
		return true
	}
	keywords := []string{"群面", "协同", "冲突", "方案评审", "多人", "team", "collaboration"}
	for _, keyword := range keywords {
		if strings.Contains(joined, keyword) {
			return true
		}
	}
	return false
}

func (s *AIService) evaluateGroupAnswerWithMockTemplate(ctx context.Context, question *model.Question, answer string) *EvaluationResult {
	template := buildMockGroupAssessmentTemplate(question, answer)
	shouldFollowUp := template.TotalScore < 78
	followUpContext := ""
	if shouldFollowUp {
		followUpContext = "请补充你在小组分歧中如何推进共识，并给出可执行落地步骤。"
	}

	evaluation := s.EnsureChineseOutput(ctx, template.PreliminaryNote, "群面回答已接收，请进一步补充协作细节。")
	reasoning := s.EnsureChineseOutput(ctx, fmt.Sprintf("本次采用 MockGroupAssessmentTemplate 进行初步评分，维度结果：领导力%d、表达%d、协作%d、逻辑%d。", template.Leadership, template.Expression, template.Teamwork, template.Logic), "已完成群面初步评分。")

	dimensions := &ReviewDimensions{
		TechnicalDepth: clampScore((template.Logic*60 + template.Teamwork*40 + 50) / 100),
		Expression:     clampScore(template.Expression),
		Logic:          clampScore(template.Logic),
		Completeness:   clampScore((template.Teamwork + template.Leadership + 1) / 2),
	}

	feedback := map[string]interface{}{
		"evaluation":            evaluation,
		"reasoning":             reasoning,
		"suggestions":           []string{"增加角色分工说明", "补充冲突收敛步骤", "给出最终决策依据"},
		"dimensions":            dimensions,
		"knowledge_retrieval":   []KnowledgeCheck{},
		"scores":                &RubricScores{TechnicalAccuracy: dimensions.TechnicalDepth, LogicalClarity: dimensions.Logic, Completeness: dimensions.Completeness, Groundedness: template.Teamwork},
		"final_score":           template.TotalScore,
		"should_follow_up":      shouldFollowUp,
		"follow_up_context":     followUpContext,
		"group_assessment_mock": template,
	}

	payload, _ := json.Marshal(feedback)
	return &EvaluationResult{Score: template.TotalScore, Feedback: string(payload)}
}

func buildMockGroupAssessmentTemplate(question *model.Question, answer string) *MockGroupAssessmentTemplate {
	assessmentType := "协同冲突解决"
	if question != nil {
		joined := strings.ToLower(question.Category + " " + question.Title + " " + question.Content)
		if strings.Contains(joined, "tech") || strings.Contains(joined, "方案评审") || strings.Contains(joined, "架构") {
			assessmentType = "技术方案评审"
		}
	}

	leadership := mockGroupDimensionScore(answer, []string{"推动", "分工", "协调", "主导", "负责人", "共识"}, 66)
	expression := mockGroupDimensionScore(answer, []string{"首先", "其次", "最后", "结论", "因为", "所以"}, 64)
	teamwork := mockGroupDimensionScore(answer, []string{"我们", "团队", "协作", "配合", "反馈", "对齐"}, 65)
	logic := mockGroupDimensionScore(answer, []string{"方案", "评审", "风险", "取舍", "优先级", "落地"}, 67)

	total := clampScore((leadership*22 + expression*24 + teamwork*28 + logic*26 + 50) / 100)
	note := "群面表达具备基础协作意识，建议补充冲突闭环与决策量化依据。"
	if total >= 85 {
		note = "群面回答结构完整，体现了协作推进与方案评审能力。"
	} else if total < 70 {
		note = "群面回答偏概述，建议补充分工细节、争议处理和落地计划。"
	}

	return &MockGroupAssessmentTemplate{
		TemplateVersion: "mock-group-assessment-v1",
		AssessmentType:  assessmentType,
		Leadership:      leadership,
		Expression:      expression,
		Teamwork:        teamwork,
		Logic:           logic,
		TotalScore:      total,
		KeySignals:      extractGroupSignals(answer),
		PreliminaryNote: note,
	}
}

func mockGroupDimensionScore(answer string, keywords []string, base int) int {
	trimmed := strings.TrimSpace(answer)
	if trimmed == "" {
		return clampScore(base - 24)
	}

	runes := []rune(trimmed)
	lengthBonus := len(runes) / 28
	if lengthBonus > 12 {
		lengthBonus = 12
	}

	hit := 0
	lower := strings.ToLower(trimmed)
	for _, keyword := range keywords {
		if strings.Contains(lower, strings.ToLower(keyword)) {
			hit++
		}
	}
	keywordBonus := hit * 3

	h := fnv.New32a()
	_, _ = h.Write([]byte(trimmed))
	jitter := int(h.Sum32()%7) - 3

	return clampScore(base + lengthBonus + keywordBonus + jitter)
}

func extractGroupSignals(answer string) []string {
	lower := strings.ToLower(strings.TrimSpace(answer))
	if lower == "" {
		return []string{"回答为空，缺少可评估信号"}
	}

	mapping := []struct {
		keyword string
		signal  string
	}{
		{keyword: "分工", signal: "提到了团队分工"},
		{keyword: "共识", signal: "提到了共识机制"},
		{keyword: "风险", signal: "提到了风险识别"},
		{keyword: "方案", signal: "提到了方案比较"},
		{keyword: "复盘", signal: "提到了复盘闭环"},
	}

	signals := make([]string, 0, len(mapping))
	for _, item := range mapping {
		if strings.Contains(lower, item.keyword) {
			signals = append(signals, item.signal)
		}
	}
	if len(signals) == 0 {
		signals = append(signals, "回答包含基础观点，但协作信号较弱")
	}
	return signals
}

func (s *AIService) evaluateAnswerWithGroundTruth(ctx context.Context, question *model.Question, answer, groundTruth string) (*ReviewResult, error) {
	if question == nil {
		return nil, fmt.Errorf("question is nil")
	}

	content := strings.TrimSpace(answer)
	if IsInvalidAnswer(content) {
		return &ReviewResult{
			Score:              0,
			FinalScore:         0,
			Comment:            "回答无效或内容过少，无法完成可靠评估。",
			Reasoning:          "回答长度不足或语义无效，无法与参考知识进行比对。",
			Suggestion:         "请先给出核心定义、关键原理与一个具体示例。",
			Dimensions:         &ReviewDimensions{},
			KnowledgeRetrieval: []KnowledgeCheck{{Point: "回答有效性", Verdict: "missing", Evidence: "回答长度过短或无有效语义"}},
			Scores:             &RubricScores{},
			ShouldFollowUp:     true,
			FollowUpContext:    "请先按定义-原理-示例-边界给出可评估的基础回答。",
			ModelAnswerOutline: defaultModelAnswerOutline(strings.TrimSpace(groundTruth)),
			FollowUp:           "请基于参考知识补充一个结构化回答。",
		}, nil
	}

	prompt, err := s.renderPrompt("evaluate_answer_with_ground_truth.tmpl", map[string]interface{}{
		"QuestionTitle":   strings.TrimSpace(question.Title),
		"Question":        strings.TrimSpace(question.Content),
		"ExpectedAnswer":  strings.TrimSpace(question.ExpectedAnswer),
		"GroundTruth":     strings.TrimSpace(groundTruth),
		"CandidateAnswer": content,
	})
	if err != nil {
		prompt = s.buildGroundedEvalPromptFallback(question, content, groundTruth)
	}

	raw, err := s.chat(ctx, prompt, "evaluation", jsonObjectResponseFormat())
	if err != nil {
		return nil, fmt.Errorf("llm evaluation call failed: %w", err)
	}

	var parsed groundedEvalLLMResponse
	if err := parseGroundedEvalResponse(raw, &parsed); err != nil {
		return nil, fmt.Errorf("invalid grounded evaluation json: %w", err)
	}

	score := clampScore(parsed.Score)
	reasoning := strings.TrimSpace(parsed.Reasoning)
	if reasoning == "" {
		reasoning = "回答与参考答案存在差异，请补充关键机制与证据。"
	}

	followUpContext := strings.TrimSpace(parsed.FollowUpContext)
	if parsed.ShouldFollowUp && followUpContext == "" {
		followUpContext = "请围绕未覆盖的关键机制继续补充实现细节与边界条件。"
	}

	scores := RubricScores{
		TechnicalAccuracy: score,
		LogicalClarity:    clampScore(score - 5),
		Completeness:      clampScore(score - 3),
		Groundedness:      clampScore(score + 2),
	}
	dims := &ReviewDimensions{
		TechnicalDepth: scores.TechnicalAccuracy,
		Expression:     clampScore((scores.LogicalClarity + scores.Groundedness) / 2),
		Logic:          scores.LogicalClarity,
		Completeness:   scores.Completeness,
	}

	highlights := []string{}
	gaps := []string{}
	if parsed.ShouldFollowUp {
		gaps = append(gaps, followUpContext)
	} else {
		highlights = append(highlights, "回答在题目要求范围内整体覆盖较完整。")
	}

	suggestion := "在保持正确性的前提下，继续补充实现细节、边界条件与技术取舍。"
	if parsed.ShouldFollowUp {
		suggestion = fmt.Sprintf("请重点补充：%s", followUpContext)
	}

	return &ReviewResult{
		Score:              score,
		FinalScore:         score,
		Comment:            reasoning,
		Reasoning:          reasoning,
		Suggestion:         suggestion,
		Dimensions:         dims,
		KnowledgeRetrieval: []KnowledgeCheck{},
		Scores:             &scores,
		ShouldFollowUp:     parsed.ShouldFollowUp,
		FollowUpContext:    followUpContext,
		Highlights:         highlights,
		Gaps:               gaps,
		ModelAnswerOutline: defaultModelAnswerOutline(strings.TrimSpace(groundTruth)),
		FollowUp:           defaultFollowUpQuestion(strings.TrimSpace(question.Content)),
	}, nil
}

func parseGroundedEvalResponse(raw string, out *groundedEvalLLMResponse) error {
	if out == nil {
		return fmt.Errorf("grounded evaluation output receiver is nil")
	}

	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return fmt.Errorf("grounded evaluation response is empty")
	}

	candidates := []string{trimmed}
	if unfenced := stripOptionalCodeFence(trimmed); unfenced != "" && unfenced != trimmed {
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

		var parsed groundedEvalLLMResponse
		if err := json.Unmarshal([]byte(candidate), &parsed); err == nil {
			normalized := normalizeGroundedEvalResponse(parsed)
			*out = normalized
			return nil
		}
	}

	return fmt.Errorf("cannot decode grounded evaluation response: %s", truncateEvalRunes(trimmed, 220))
}

func stripOptionalCodeFence(text string) string {
	trimmed := strings.TrimSpace(text)
	if !strings.HasPrefix(trimmed, "```") {
		return trimmed
	}
	lines := strings.Split(trimmed, "\n")
	if len(lines) < 2 {
		return strings.Trim(trimmed, "`")
	}
	if strings.HasPrefix(strings.TrimSpace(lines[len(lines)-1]), "```") {
		return strings.TrimSpace(strings.Join(lines[1:len(lines)-1], "\n"))
	}
	return strings.TrimSpace(strings.Join(lines[1:], "\n"))
}

func normalizeGroundedEvalResponse(parsed groundedEvalLLMResponse) groundedEvalLLMResponse {
	parsed.Score = clampScore(parsed.Score)
	parsed.Reasoning = strings.TrimSpace(parsed.Reasoning)
	parsed.FollowUpContext = strings.TrimSpace(parsed.FollowUpContext)
	if !parsed.ShouldFollowUp {
		parsed.FollowUpContext = ""
	}
	return parsed
}

func (s *AIService) buildEvaluationGroundTruth(ctx context.Context, question *model.Question) string {
	_ = ctx
	if question == nil {
		return ""
	}

	segments := make([]string, 0, evaluationGroundTruthTopK+1)
	if strings.TrimSpace(question.ExpectedAnswer) != "" {
		segments = append(segments, fmt.Sprintf("题库参考答案：\n%s", strings.TrimSpace(question.ExpectedAnswer)))
	}

	if s.groundTruthRetriever != nil {
		query := strings.TrimSpace(strings.Join([]string{question.Title, question.Content, question.ExpectedAnswer}, "\n"))
		if query != "" {
			refs, err := s.groundTruthRetriever(context.Background(), query, evaluationGroundTruthTopK)
			if err != nil {
				log.Printf("ground truth retrieval failed: %v", err)
			} else {
				for i, ref := range refs {
					trimmed := strings.TrimSpace(ref)
					if trimmed == "" {
						continue
					}
					segments = append(segments, fmt.Sprintf("RAG参考知识片段%d：\n%s", i+1, trimmed))
				}
			}
		}
	}

	if len(segments) == 0 {
		return "(No grounded reference available)"
	}
	return truncateEvalRunes(strings.Join(segments, "\n\n---\n\n"), evaluationGroundTruthMaxRune)
}

func (s *AIService) evaluateAnswerLocal(question *model.Question, answer string) *EvaluationResult {
	content := strings.TrimSpace(answer)
	if IsInvalidAnswer(content) {
		feedback, _ := json.Marshal(map[string]interface{}{
			"evaluation":        "回答无效或内容过少，当前无法有效评估。",
			"suggestions":       []string{"请先给出核心定义", "再补充关键原理与一个具体示例"},
			"should_follow_up":  true,
			"follow_up_context": "请先给出可评估的结构化回答。",
		})
		return &EvaluationResult{Score: 0, Feedback: string(feedback)}
	}

	score := 30 + len([]rune(content))/8
	if score > 85 {
		score = 85
	}
	if question != nil && strings.TrimSpace(question.ExpectedAnswer) != "" {
		if strings.Contains(strings.ToLower(content), strings.ToLower(strings.TrimSpace(question.Position))) {
			score += 5
		}
	}
	score = clampScore(score)

	evaluation := "回答与题目相关，但在机制深度和证据支撑上仍有提升空间。"
	if score >= 80 {
		evaluation = "核心要点覆盖较完整，结构较清晰。"
	}

	feedback, _ := json.Marshal(map[string]interface{}{
		"evaluation":        evaluation,
		"suggestions":       []string{"补充底层原理", "加入边界条件", "给出可验证结果"},
		"should_follow_up":  score < 85,
		"follow_up_context": "请补充最薄弱的一点并给出实现细节。",
	})
	return &EvaluationResult{Score: score, Feedback: string(feedback)}
}

func (s *AIService) buildGroundedEvalPromptFallback(question *model.Question, candidateAnswer, groundTruth string) string {
	questionTitle, questionContent, expectedAnswer := "", "", ""
	if question != nil {
		questionTitle = strings.TrimSpace(question.Title)
		questionContent = strings.TrimSpace(question.Content)
		expectedAnswer = strings.TrimSpace(question.ExpectedAnswer)
	}

	prompt, err := s.renderDynamicPrompt("ai/evaluation/grounded_fallback", map[string]interface{}{
		"QuestionTitle":   questionTitle,
		"Question":        questionContent,
		"ExpectedAnswer":  expectedAnswer,
		"GroundTruth":     strings.TrimSpace(groundTruth),
		"CandidateAnswer": strings.TrimSpace(candidateAnswer),
	})
	if err == nil && strings.TrimSpace(prompt) != "" {
		return prompt
	}

	return fmt.Sprintf(`你是严格阅卷系统。请仅返回 JSON，字段为 score, reasoning, should_follow_up, follow_up_context。
题目标题：%s
题目内容：%s
题库参考答案：%s
GROUND_TRUTH：%s
用户回答：%s`, questionTitle, questionContent, expectedAnswer, strings.TrimSpace(groundTruth), strings.TrimSpace(candidateAnswer))
}

func truncateEvalRunes(text string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(strings.TrimSpace(text))
	if len(runes) <= max {
		return string(runes)
	}
	return strings.TrimSpace(string(runes[:max]))
}

func splitSuggestionText(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return []string{"请补充核心原理、实现细节与可验证结果。"}
	}
	parts := strings.FieldsFunc(text, func(r rune) bool {
		return r == ';' || r == '；' || r == '\n'
	})
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	if len(result) == 0 {
		return []string{text}
	}
	return result
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
