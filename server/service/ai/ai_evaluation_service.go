package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"your-project/model"
)

const (
	evaluationGroundTruthTopK    = 5
	evaluationGroundTruthMaxRune = 3600
	evaluationOutlineMaxRune     = 1200
)

type groundedEvalLLMResponse struct {
	Score           int    `json:"score"`
	Reasoning       string `json:"reasoning"`
	ShouldFollowUp  bool   `json:"should_follow_up"`
	FollowUpContext string `json:"follow_up_context"`
}

func (s *AIService) EvaluateAnswer(ctx context.Context, question *model.Question, answer string) (*EvaluationResult, error) {
	if question == nil {
		log.Printf("AI review skipped: nil question, fallback to local heuristic")
		return s.evaluateAnswerLocal(nil, answer), nil
	}

	groundTruth := s.buildEvaluationGroundTruth(ctx, question)
	reviewResult, err := s.evaluateAnswerWithGroundTruth(ctx, question, answer, groundTruth)
	if err != nil {
		log.Printf("AI grounded review failed, fallback to local heuristic: %v", err)
		return s.evaluateAnswerLocal(question, answer), nil
	}

	if reviewResult.Score == 0 && reviewResult.FinalScore > 0 {
		reviewResult.Score = reviewResult.FinalScore
	}
	if reviewResult.FinalScore == 0 && reviewResult.Score > 0 {
		reviewResult.FinalScore = reviewResult.Score
	}
	reviewResult.Score = clampScore(reviewResult.Score)
	reviewResult.FinalScore = clampScore(reviewResult.FinalScore)

	evaluationText := s.EnsureChineseOutput(ctx, firstNonEmpty(strings.TrimSpace(reviewResult.Comment), strings.TrimSpace(reviewResult.Reasoning)), "回答已接收，当前质量仍有提升空间。")
	reasoningText := s.EnsureChineseOutput(ctx, firstNonEmpty(strings.TrimSpace(reviewResult.Reasoning), strings.TrimSpace(reviewResult.Comment)), "回答与参考知识存在明显差异，请补充关键原理并纠正事实错误。")
	suggestionText := s.EnsureChineseOutput(ctx, reviewResult.Suggestion, "请补充核心原理、实现细节和边界条件。")
	suggestionItems := splitSuggestionText(suggestionText)

	dims := reviewResult.Dimensions
	if dims == nil {
		dims = estimateDimensions(reviewResult.Score)
	}

	if reviewResult.Scores == nil {
		reviewResult.Scores = &RubricScores{
			TechnicalAccuracy: clampScore(dims.TechnicalDepth),
			LogicalClarity:    clampScore(dims.Logic),
			Completeness:      clampScore(dims.Completeness),
			Groundedness:      clampScore((dims.TechnicalDepth + dims.Logic) / 2),
		}
	}

	richFeedback := map[string]interface{}{
		"evaluation":           evaluationText,
		"reasoning":            reasoningText,
		"suggestions":          suggestionItems,
		"dimensions":           dims,
		"knowledge_retrieval":  reviewResult.KnowledgeRetrieval,
		"scores":               reviewResult.Scores,
		"final_score":          reviewResult.FinalScore,
		"should_follow_up":     reviewResult.ShouldFollowUp,
		"follow_up_context":    strings.TrimSpace(reviewResult.FollowUpContext),
		"ground_truth":         groundTruth,
		"highlights":           reviewResult.Highlights,
		"gaps":                 reviewResult.Gaps,
		"model_answer_outline": reviewResult.ModelAnswerOutline,
		"follow_up":            reviewResult.FollowUp,
	}
	feedbackJSON, _ := json.Marshal(richFeedback)
	return &EvaluationResult{Score: reviewResult.Score, Feedback: string(feedbackJSON)}, nil
}

func (s *AIService) evaluateAnswerWithGroundTruth(ctx context.Context, question *model.Question, answer, groundTruth string) (*ReviewResult, error) {
	if question == nil {
		return nil, fmt.Errorf("question is nil")
	}

	content := strings.TrimSpace(answer)
	if IsInvalidAnswer(content) {
		checks := []KnowledgeCheck{}
		if strings.TrimSpace(groundTruth) != "" {
			checks = append(checks, KnowledgeCheck{Point: "回答有效性", Verdict: "missing", Evidence: "回答长度过短或缺乏有效语义"})
		}
		return &ReviewResult{
			Score:              0,
			FinalScore:         0,
			Comment:            "回答无效或内容过少，无法完成可靠评估。",
			Reasoning:          "回答长度不足或无有效语义，无法与参考知识进行可靠比对。",
			Suggestion:         "请先给出核心定义、关键原理与一个具体实例，再说明边界条件。",
			Dimensions:         &ReviewDimensions{},
			KnowledgeRetrieval: checks,
			Scores:             &RubricScores{},
			ShouldFollowUp:     true,
			FollowUpContext:    "请先按“定义-原理-示例-边界”补充一个可评估的基础回答。",
			ModelAnswerOutline: truncateRunes(strings.TrimSpace(groundTruth), evaluationOutlineMaxRune),
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
		prompt = buildGroundedEvalPromptFallback(question, content, groundTruth)
	}

	raw, err := s.chat(ctx, prompt, "evaluation", jsonObjectResponseFormat())
	if err != nil {
		return nil, fmt.Errorf("llm evaluation call failed: %w", err)
	}

	var parsed groundedEvalLLMResponse
	if err := parseGroundedEvalResponse(raw, &parsed); err != nil {
		return nil, fmt.Errorf("invalid grounded evaluation json: %w", err)
	}

	return buildGroundedReviewResult(question, groundTruth, parsed), nil
}

func buildGroundedReviewResult(question *model.Question, groundTruth string, parsed groundedEvalLLMResponse) *ReviewResult {
	finalScore := clampScore(parsed.Score)
	reasoning := strings.TrimSpace(parsed.Reasoning)
	if reasoning == "" {
		reasoning = "回答与标准参考答案的比对信息不足，请补充关键知识点后再评估。"
	}

	shouldFollowUp := parsed.ShouldFollowUp
	followUpContext := strings.TrimSpace(parsed.FollowUpContext)
	if !shouldFollowUp {
		followUpContext = ""
	}
	if shouldFollowUp && followUpContext == "" {
		followUpContext = "请围绕标准答案范围内尚未展开的关键机制继续追问其实现细节与边界条件。"
	}

	highlights := []string{}
	gaps := []string{}
	if shouldFollowUp {
		gaps = append(gaps, followUpContext)
	} else {
		highlights = append(highlights, "回答在题目要求范围内整体覆盖较完整。")
	}

	knowledgeChecks := make([]KnowledgeCheck, 0, 1)
	if shouldFollowUp {
		knowledgeChecks = append(knowledgeChecks, KnowledgeCheck{
			Point:    followUpContext,
			Verdict:  "missing",
			Evidence: "评估结果建议在该点继续深挖。",
		})
	} else {
		knowledgeChecks = append(knowledgeChecks, KnowledgeCheck{
			Point:    "标准答案范围覆盖",
			Verdict:  "supported",
			Evidence: "评估结果显示当前回答可在题目要求范围内通过。",
		})
	}

	scores := deriveRubricScoresFromOverall(finalScore, shouldFollowUp)
	dims := &ReviewDimensions{
		TechnicalDepth: clampScore(scores.TechnicalAccuracy),
		Expression:     clampScore((scores.LogicalClarity + scores.Groundedness) / 2),
		Logic:          clampScore(scores.LogicalClarity),
		Completeness:   clampScore(scores.Completeness),
	}
	alignDimensionsWithScoreSimple(dims, finalScore)

	suggestion := "在保持正确性的前提下，继续补充实现细节、边界条件与技术取舍。"
	if shouldFollowUp {
		suggestion = fmt.Sprintf("请重点补充：%s", followUpContext)
	}

	outline := truncateRunes(strings.TrimSpace(groundTruth), evaluationOutlineMaxRune)
	if outline == "" {
		outline = defaultModelAnswerOutline(strings.TrimSpace(question.ExpectedAnswer))
	}

	followUp := defaultFollowUpQuestion(strings.TrimSpace(question.Content))
	if shouldFollowUp {
		followUp = fmt.Sprintf("请围绕“%s”继续补充说明。", followUpContext)
	}

	return &ReviewResult{
		Score:              finalScore,
		Comment:            reasoning,
		Suggestion:         suggestion,
		Dimensions:         dims,
		KnowledgeRetrieval: knowledgeChecks,
		Reasoning:          reasoning,
		Scores:             &scores,
		FinalScore:         finalScore,
		ShouldFollowUp:     shouldFollowUp,
		FollowUpContext:    followUpContext,
		Highlights:         highlights,
		Gaps:               gaps,
		ModelAnswerOutline: outline,
		FollowUp:           followUp,
	}
}

func parseGroundedEvalResponse(raw string, out *groundedEvalLLMResponse) error {
	if out == nil {
		return fmt.Errorf("grounded evaluation output receiver is nil")
	}

	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return fmt.Errorf("grounded evaluation response is empty")
	}

	candidates := make([]string, 0, 3)
	candidates = append(candidates, trimmed)

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

	return fmt.Errorf("cannot decode grounded evaluation response: %s", truncateRunes(trimmed, 220))
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

func deriveRubricScoresFromOverall(score int, shouldFollowUp bool) RubricScores {
	base := clampScore(score)
	technical := base
	logical := clampScore(base - 2)
	completeness := base
	groundedness := clampScore(base + 2)

	if shouldFollowUp {
		completeness = clampScore(completeness - 5)
		groundedness = clampScore(groundedness - 3)
	}

	return RubricScores{
		TechnicalAccuracy: technical,
		LogicalClarity:    logical,
		Completeness:      completeness,
		Groundedness:      groundedness,
	}
}

func (s *AIService) buildEvaluationGroundTruth(ctx context.Context, question *model.Question) string {
	if question == nil {
		return ""
	}

	segments := make([]string, 0, evaluationGroundTruthTopK+1)
	seen := make(map[string]struct{}, evaluationGroundTruthTopK+1)

	addSegment := func(label, value string) {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		if strings.TrimSpace(label) == "" {
			segments = append(segments, trimmed)
			return
		}
		segments = append(segments, fmt.Sprintf("%s\n%s", strings.TrimSpace(label), trimmed))
	}

	addSegment("题库参考答案：", question.ExpectedAnswer)

	query := strings.TrimSpace(strings.Join([]string{
		strings.TrimSpace(question.Title),
		strings.TrimSpace(question.Content),
		strings.TrimSpace(question.ExpectedAnswer),
	}, "\n"))

	if s.groundTruthRetriever != nil && query != "" {
		references, err := s.groundTruthRetriever(ctx, query, evaluationGroundTruthTopK)
		if err != nil {
			log.Printf("ground truth retrieval failed: %v", err)
		} else {
			for i, ref := range references {
				addSegment(fmt.Sprintf("RAG参考知识片段 %d：", i+1), ref)
			}
		}
	}

	if len(segments) == 0 {
		return "(No grounded reference available)"
	}

	return truncateRunes(strings.Join(segments, "\n\n---\n\n"), evaluationGroundTruthMaxRune)
}

func normalizeKnowledgeChecks(items []KnowledgeCheck) []KnowledgeCheck {
	normalized := make([]KnowledgeCheck, 0, len(items))
	for _, item := range items {
		point := strings.TrimSpace(item.Point)
		evidence := strings.TrimSpace(item.Evidence)
		if point == "" && evidence == "" {
			continue
		}
		verdict := normalizeKnowledgeVerdict(item.Verdict, evidence)
		normalized = append(normalized, KnowledgeCheck{
			Point:    point,
			Verdict:  verdict,
			Evidence: evidence,
		})
	}
	return normalized
}

func normalizeKnowledgeVerdict(verdict, evidence string) string {
	v := strings.ToLower(strings.TrimSpace(verdict))
	switch v {
	case "supported", "match", "matched", "correct", "正确", "匹配":
		return "supported"
	case "missing", "lack", "缺失", "遗漏", "不足":
		return "missing"
	case "contradicted", "wrong", "error", "错误", "事实性错误", "矛盾":
		return "contradicted"
	}

	hint := strings.ToLower(strings.TrimSpace(evidence))
	switch {
	case strings.Contains(hint, "缺") || strings.Contains(hint, "未") || strings.Contains(hint, "漏"):
		return "missing"
	case strings.Contains(hint, "错") || strings.Contains(hint, "矛盾") || strings.Contains(hint, "contradict") || strings.Contains(hint, "incorrect"):
		return "contradicted"
	case strings.Contains(hint, "匹配") || strings.Contains(hint, "一致") || strings.Contains(hint, "support"):
		return "supported"
	default:
		return "unknown"
	}
}

func normalizeRubricScores(scores RubricScores) RubricScores {
	return RubricScores{
		TechnicalAccuracy: clampScore(scores.TechnicalAccuracy),
		LogicalClarity:    clampScore(scores.LogicalClarity),
		Completeness:      clampScore(scores.Completeness),
		Groundedness:      clampScore(scores.Groundedness),
	}
}

func computeFinalScore(scores RubricScores, rawFinal *int) int {
	weighted := int(
		float64(scores.TechnicalAccuracy)*0.50 +
			float64(scores.LogicalClarity)*0.20 +
			float64(scores.Completeness)*0.20 +
			float64(scores.Groundedness)*0.10 +
			0.5,
	)
	weighted = clampScore(weighted)

	if rawFinal == nil {
		return weighted
	}
	parsed := clampScore(*rawFinal)
	if parsed == 0 && weighted > 0 {
		return weighted
	}
	return parsed
}

func summarizeKnowledgeChecks(checks []KnowledgeCheck) ([]string, []string) {
	highlights := make([]string, 0, len(checks))
	gaps := make([]string, 0, len(checks))
	for _, check := range checks {
		point := strings.TrimSpace(check.Point)
		if point == "" {
			point = strings.TrimSpace(check.Evidence)
		}
		if point == "" {
			continue
		}

		switch check.Verdict {
		case "supported":
			highlights = append(highlights, point)
		case "missing", "contradicted":
			gaps = append(gaps, point)
		}
	}

	if len(highlights) > 4 {
		highlights = highlights[:4]
	}
	if len(gaps) > 4 {
		gaps = gaps[:4]
	}
	return highlights, gaps
}

func buildSuggestionFromGaps(gaps []string) string {
	if len(gaps) == 0 {
		return "在保持准确性的前提下，补充实现细节、边界条件与指标数据。"
	}
	if len(gaps) == 1 {
		return fmt.Sprintf("请优先补齐并核对关键点：%s。", gaps[0])
	}
	return fmt.Sprintf("请优先补齐并纠正以下关键点：%s；%s。", gaps[0], gaps[1])
}

func buildGroundedEvalPromptFallback(question *model.Question, candidateAnswer, groundTruth string) string {
	questionTitle := ""
	questionContent := ""
	expectedAnswer := ""
	if question != nil {
		questionTitle = strings.TrimSpace(question.Title)
		questionContent = strings.TrimSpace(question.Content)
		expectedAnswer = strings.TrimSpace(question.ExpectedAnswer)
	}

	return fmt.Sprintf(`你是无情的阅卷系统。你必须严格以“GROUND_TRUTH”为唯一准绳。
强制约束：
1. 语言强制：你所有输出必须 100%% 使用中文（zh-CN）。
2. 边界限制：你只能基于【原题目】与【标准参考答案】中明确要求的知识点评分，不得因未提及题目未要求的扩展知识扣分。
3. 严禁脑补：若无证据，不得推断。
4. 仅返回一个 JSON 对象，不要输出额外文本。
5. JSON 只允许字段：score, reasoning, should_follow_up, follow_up_context。
6. score 必须是 0-100 的整数。

题目标题：%s
题目：%s
题库参考答案：%s
GROUND_TRUTH：%s
用户回答：%s

输出格式示例：
{
  "score": 85,
  "reasoning": "答对了XXX，但在XXX方面有所缺失（仅限标准答案范围内的缺失）",
  "should_follow_up": true,
  "follow_up_context": "可围绕XXX的底层机制继续追问"
}`,
		questionTitle,
		questionContent,
		expectedAnswer,
		strings.TrimSpace(groundTruth),
		strings.TrimSpace(candidateAnswer),
	)
}

func truncateRunes(text string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(strings.TrimSpace(text))
	if len(runes) <= max {
		return string(runes)
	}
	return strings.TrimSpace(string(runes[:max]))
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func splitSuggestionText(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return []string{"请补充核心原理、实现细节与可验证的结果数据。"}
	}
	parts := strings.FieldsFunc(text, func(r rune) bool {
		return r == ';' || r == '；'
	})
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.TrimLeft(p, "0123456789.、) ")
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

func estimateDimensions(totalScore int) *ReviewDimensions {
	base := totalScore
	return &ReviewDimensions{
		TechnicalDepth: clampScore(base - 5 + (base%7 - 3)),
		Expression:     clampScore(base + 3 + (base%5 - 2)),
		Logic:          clampScore(base + (base%6 - 3)),
		Completeness:   clampScore(base - 3 + (base%4 - 2)),
	}
}

func (s *AIService) evaluateAnswerLocal(question *model.Question, answer string) *EvaluationResult {
	content := strings.TrimSpace(answer)
	questionContent := ""
	questionTitle := ""
	expectedAnswer := ""
	if question != nil {
		questionContent = strings.TrimSpace(question.Content)
		questionTitle = strings.TrimSpace(question.Title)
		expectedAnswer = strings.TrimSpace(question.ExpectedAnswer)
	}
	promptContext := strings.TrimSpace(questionTitle + " " + questionContent)

	if IsInvalidAnswer(content) {
		return s.buildRichLocalFeedback(0, questionContent,
			"回答无效或内容过少，当前无法有效评估。",
			[]string{"请先给出核心定义，再补充关键原理与一个具体示例", "避免“不会”“不知道”等无信息量回答"},
			&ReviewDimensions{},
			nil,
			[]string{"当前回答缺乏可用于评分的技术信号"},
		)
	}

	signals := simpleAnswerSignals(promptContext, expectedAnswer, content)

	score := 18
	switch {
	case signals.runeLen >= 220:
		score += 22
	case signals.runeLen >= 140:
		score += 18
	case signals.runeLen >= 80:
		score += 14
	case signals.runeLen >= 40:
		score += 8
	case signals.runeLen >= 20:
		score += 4
	}
	if signals.hasStructure {
		score += 8
	}
	score += int(signals.keywordCoverage*45 + 0.5)

	techBonus := signals.technicalHits * 3
	if techBonus > 15 {
		techBonus = 15
	}
	score += techBonus
	if signals.keywordCoverage < 0.05 && signals.runeLen < 40 && score > 45 {
		score = 45
	}
	score = clampScore(score)

	var evaluation string
	var highlights []string
	var gaps []string
	if score >= 80 {
		evaluation = "核心要点覆盖较完整，结构清晰。"
		highlights = []string{"已覆盖主要知识点", "表达结构清晰且有一定深度"}
		gaps = []string{"可继续补充更底层的实现细节"}
	} else if score >= 60 {
		evaluation = "回答与题目相关，但在机制深度和证据支撑上仍有不足。"
		if signals.hasStructure {
			highlights = []string{"具备基本回答结构"}
		}
		gaps = []string{"机制说明偏浅", "缺少技术取舍与结果指标"}
	} else {
		evaluation = "回答的相关性和完整性不足，难以支撑较高评分。"
		gaps = []string{"核心点缺失", "因果链路不完整", "缺少可验证的实践证据"}
	}

	dims := &ReviewDimensions{
		TechnicalDepth: clampScore(score - 12 + signals.technicalHits*2),
		Expression:     clampScore(score - 4),
		Logic:          clampScore(score - 5),
		Completeness:   clampScore(int(float64(score) * (0.65 + 0.5*signals.keywordCoverage))),
	}
	if signals.hasStructure {
		dims.Expression = clampScore(dims.Expression + 5)
		dims.Logic = clampScore(dims.Logic + 6)
	}
	if signals.keywordCoverage >= 0.25 {
		dims.Completeness = clampScore(dims.Completeness + 8)
	}
	alignDimensionsWithScoreSimple(dims, score)

	suggestions := make([]string, 0, 4)
	if signals.keywordCoverage < 0.2 {
		suggestions = append(suggestions, "先覆盖题目中的核心关键词，再展开实现细节")
	}
	if signals.technicalHits < 2 {
		suggestions = append(suggestions, "补充底层原理、复杂度分析与技术取舍")
	}
	if !signals.hasStructure {
		suggestions = append(suggestions, "建议按“结论 -> 原理 -> 示例 -> 边界”结构组织回答")
	}
	suggestions = append(suggestions, "补充可量化结果（如时延、吞吐、错误率）")
	if len(suggestions) > 3 {
		suggestions = suggestions[:3]
	}

	return s.buildRichLocalFeedback(score, questionContent, evaluation, suggestions, dims, highlights, gaps)
}

type simpleSignals struct {
	runeLen         int
	keywordCoverage float64
	technicalHits   int
	hasStructure    bool
}

func simpleAnswerSignals(promptContext, expectedAnswer, content string) simpleSignals {
	merged := strings.ToLower(strings.TrimSpace(promptContext + " " + expectedAnswer))
	answer := strings.ToLower(strings.TrimSpace(content))
	keywords := strings.Fields(merged)
	hit := 0
	seen := map[string]bool{}
	for _, k := range keywords {
		k = strings.TrimSpace(k)
		if len([]rune(k)) < 2 || seen[k] {
			continue
		}
		seen[k] = true
		if strings.Contains(answer, k) {
			hit++
		}
	}
	total := len(seen)
	coverage := 0.0
	if total > 0 {
		coverage = float64(hit) / float64(total)
	}

	techTerms := []string{"complexity", "cache", "index", "transaction", "thread", "lock", "性能", "复杂度", "索引", "并发"}
	techHits := 0
	for _, t := range techTerms {
		if strings.Contains(answer, strings.ToLower(t)) {
			techHits++
		}
	}

	hasStructure := strings.Contains(answer, "first") || strings.Contains(answer, "then") || strings.Contains(answer, "最后") || strings.Contains(answer, "首先")

	return simpleSignals{
		runeLen:         len([]rune(strings.TrimSpace(content))),
		keywordCoverage: coverage,
		technicalHits:   techHits,
		hasStructure:    hasStructure,
	}
}

func alignDimensionsWithScoreSimple(dims *ReviewDimensions, score int) {
	if dims == nil {
		return
	}
	avg := (dims.TechnicalDepth + dims.Expression + dims.Logic + dims.Completeness) / 4
	delta := score - avg
	if delta == 0 {
		return
	}
	dims.TechnicalDepth = clampScore(dims.TechnicalDepth + delta)
	dims.Expression = clampScore(dims.Expression + delta)
	dims.Logic = clampScore(dims.Logic + delta)
	dims.Completeness = clampScore(dims.Completeness + delta)
}

func (s *AIService) buildRichLocalFeedback(score int, questionContent, evaluation string, suggestions []string, dims *ReviewDimensions, highlights, gaps []string) *EvaluationResult {
	shouldFollowUp := score < 85
	followUpContext := ""
	if shouldFollowUp {
		if len(gaps) > 0 {
			followUpContext = strings.TrimSpace(gaps[0])
		} else {
			followUpContext = "请继续补充题目要求范围内的关键机制与边界条件。"
		}
	}

	richFeedback := map[string]interface{}{
		"evaluation":           evaluation,
		"suggestions":          suggestions,
		"dimensions":           dims,
		"highlights":           highlights,
		"gaps":                 gaps,
		"should_follow_up":     shouldFollowUp,
		"follow_up_context":    followUpContext,
		"model_answer_outline": "先给出定义，再说明机制，随后结合场景，最后补充边界与取舍。",
		"follow_up":            "请围绕当前回答中最薄弱的一点，补充底层原理和实现细节。",
	}
	feedbackJSON, _ := json.Marshal(richFeedback)
	return &EvaluationResult{Score: score, Feedback: string(feedbackJSON)}
}
