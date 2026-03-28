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
	KnowledgeRetrieval []KnowledgeCheck `json:"knowledge_retrieval"`
	Reasoning          string           `json:"reasoning"`
	Scores             RubricScores     `json:"scores"`
	FinalScore         *int             `json:"final_score"`
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

	evaluationText := s.EnsureChineseOutput(ctx, firstNonEmpty(strings.TrimSpace(reviewResult.Comment), strings.TrimSpace(reviewResult.Reasoning)), "Answer received. Quality needs improvement.")
	reasoningText := s.EnsureChineseOutput(ctx, firstNonEmpty(strings.TrimSpace(reviewResult.Reasoning), strings.TrimSpace(reviewResult.Comment)), "回答与参考知识存在明显差异，请补充关键原理并纠正事实错误。")
	suggestionText := s.EnsureChineseOutput(ctx, reviewResult.Suggestion, "Please add core principles and implementation details.")
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
			Comment:            "Invalid or empty answer.",
			Reasoning:          "回答长度不足或无有效语义，无法与参考知识进行可靠比对。",
			Suggestion:         "请先给出核心定义、关键原理与一个具体实例，再说明边界条件。",
			Dimensions:         &ReviewDimensions{},
			KnowledgeRetrieval: checks,
			Scores:             &RubricScores{},
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
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &parsed); err != nil {
		return nil, fmt.Errorf("invalid grounded evaluation json: %w", err)
	}

	return buildGroundedReviewResult(question, groundTruth, parsed), nil
}

func buildGroundedReviewResult(question *model.Question, groundTruth string, parsed groundedEvalLLMResponse) *ReviewResult {
	normalizedChecks := normalizeKnowledgeChecks(parsed.KnowledgeRetrieval)
	scores := normalizeRubricScores(parsed.Scores)
	finalScore := computeFinalScore(scores, parsed.FinalScore)

	highlights, gaps := summarizeKnowledgeChecks(normalizedChecks)
	if len(normalizedChecks) == 0 {
		normalizedChecks = []KnowledgeCheck{{Point: "参考知识比对", Verdict: "unknown", Evidence: "模型未返回可解析的知识点对照结果"}}
	}

	reasoning := strings.TrimSpace(parsed.Reasoning)
	if reasoning == "" {
		reasoning = "回答与参考知识比对信息不足，需重点核查关键概念是否完整与准确。"
	}

	comment := reasoning
	suggestion := buildSuggestionFromGaps(gaps)
	outline := truncateRunes(strings.TrimSpace(groundTruth), evaluationOutlineMaxRune)
	if outline == "" {
		outline = defaultModelAnswerOutline(strings.TrimSpace(question.ExpectedAnswer))
	}

	followUp := defaultFollowUpQuestion(strings.TrimSpace(question.Content))
	if len(gaps) > 0 {
		followUp = fmt.Sprintf("你刚才遗漏或说错了“%s”，请基于参考知识补充完整。", gaps[0])
	}

	dims := &ReviewDimensions{
		TechnicalDepth: clampScore(scores.TechnicalAccuracy),
		Expression:     clampScore((scores.LogicalClarity + scores.Groundedness) / 2),
		Logic:          clampScore(scores.LogicalClarity),
		Completeness:   clampScore(scores.Completeness),
	}
	alignDimensionsWithScoreSimple(dims, finalScore)

	return &ReviewResult{
		Score:              finalScore,
		Comment:            comment,
		Suggestion:         suggestion,
		Dimensions:         dims,
		KnowledgeRetrieval: normalizedChecks,
		Reasoning:          reasoning,
		Scores:             &scores,
		FinalScore:         finalScore,
		Highlights:         highlights,
		Gaps:               gaps,
		ModelAnswerOutline: outline,
		FollowUp:           followUp,
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
规则：
1. 严禁自行脑补知识。若回答在参考知识中找不到支撑，必须扣分。
2. 若出现事实性错误或与参考知识矛盾，必须在 reasoning 中指出并扣分。
3. 仅返回一个 JSON 对象，不要输出额外文本。
4. JSON 必须包含字段：knowledge_retrieval, reasoning, scores, final_score。
5. scores 维度：technical_accuracy, logical_clarity, completeness, groundedness（均为 0-100 整数）。
6. final_score 按权重计算：0.50*technical_accuracy + 0.20*logical_clarity + 0.20*completeness + 0.10*groundedness。

题目标题：%s
题目：%s
题库参考答案：%s
GROUND_TRUTH：%s
用户回答：%s`, questionTitle, questionContent, expectedAnswer, strings.TrimSpace(groundTruth), strings.TrimSpace(candidateAnswer))
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
		return []string{"Add core principles and practical details."}
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
			"Answer is invalid or explicitly gives up.",
			[]string{"Start from key definition, then principles and examples", "Avoid empty responses like I don't know"},
			&ReviewDimensions{},
			nil,
			[]string{"No valid technical signal for evaluation"},
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
		evaluation = "Good coverage of key points with clear structure."
		highlights = []string{"Core points were covered", "Structured and reasonably deep"}
		gaps = []string{"Could add more low-level implementation details"}
	} else if score >= 60 {
		evaluation = "Answer is relevant but lacks depth in mechanism and evidence."
		if signals.hasStructure {
			highlights = []string{"Basic structure is present"}
		}
		gaps = []string{"Mechanism explanation is shallow", "Missing trade-offs and outcomes"}
	} else {
		evaluation = "Answer lacks relevance and completeness for a high score."
		gaps = []string{"Core points missing", "Causal chain is incomplete", "No verifiable practical evidence"}
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
		suggestions = append(suggestions, "Cover core keywords first, then details")
	}
	if signals.technicalHits < 2 {
		suggestions = append(suggestions, "Add low-level principles, complexity, and trade-offs")
	}
	if !signals.hasStructure {
		suggestions = append(suggestions, "Use a conclusion -> principle -> example -> boundary structure")
	}
	suggestions = append(suggestions, "Add measurable outcomes (latency, throughput, error rate)")
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
	richFeedback := map[string]interface{}{
		"evaluation":           evaluation,
		"suggestions":          suggestions,
		"dimensions":           dims,
		"highlights":           highlights,
		"gaps":                 gaps,
		"model_answer_outline": "Start with definition, then mechanism, scenario and caveats.",
		"follow_up":            "Can you describe how this was used in your project with concrete details?",
	}
	feedbackJSON, _ := json.Marshal(richFeedback)
	return &EvaluationResult{Score: score, Feedback: string(feedbackJSON)}
}
