package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"your-project/config"
	"your-project/model"
	"your-project/pkg/asr"
)

type AIService struct {
	config *config.LLMConfig
}

func NewAIService() *AIService {
	return &AIService{
		config: &config.GetConfig().LLM,
	}
}

type EvaluationResult struct {
	Score    int    `json:"score"`
	Feedback string `json:"feedback"`
}

type ReportInsights struct {
	OverallAnalysis string   `json:"overall_analysis"`
	Strengths       []string `json:"strengths"`
	Weaknesses      []string `json:"weaknesses"`
	Suggestions     []string `json:"suggestions"`
	TechnicalScore  int      `json:"technical_score"`
	ExpressionScore int      `json:"expression_score"`
	LogicScore      int      `json:"logic_score"`
	MatchingScore   int      `json:"matching_score"`
	BehaviorScore   int      `json:"behavior_score"`
}

// Chat exposes the raw LLM chat capability
func (s *AIService) Chat(ctx context.Context, prompt string) (string, error) {
	// Default model or generic chat
	return s.callLLM(prompt, "chat")
}

// ChatWithTask exposes the raw LLM chat capability with a specific task type
func (s *AIService) ChatWithTask(ctx context.Context, prompt string, taskType string) (string, error) {
	return s.callLLM(prompt, taskType)
}

func normalizeShadowHintText(text string, fallback string) string {
	clean := strings.TrimSpace(strings.ReplaceAll(text, "\n", " "))
	clean = strings.Join(strings.Fields(clean), " ")
	if clean == "" {
		clean = fallback
	}
	runes := []rune(clean)
	if len(runes) > 72 {
		clean = strings.TrimSpace(string(runes[:72]))
	}
	return clean
}

func extractShadowHintFocus(question string) string {
	trimmed := strings.TrimSpace(question)
	if trimmed == "" {
		return "这道题"
	}
	replacer := strings.NewReplacer(
		"请问", "",
		"你觉得", "",
		"你如何", "",
		"你怎么", "",
		"是什么", "",
		"为什么", "",
		"如何", "",
		"怎么", "",
		"？", "",
		"?", "",
	)
	focus := strings.TrimSpace(replacer.Replace(trimmed))
	if focus == "" {
		focus = trimmed
	}
	runes := []rune(focus)
	if len(runes) > 18 {
		focus = string(runes[:18])
	}
	if strings.TrimSpace(focus) == "" {
		return "这道题"
	}
	return focus
}

func buildShadowHintFallbacks(question string) []string {
	focus := extractShadowHintFocus(question)
	return []string{
		fmt.Sprintf("先别求完整，围绕“%s”先抛一个判断。", focus),
		fmt.Sprintf("把“%s”拆成两段讲：先原理，再落一个业务场景。", focus),
		fmt.Sprintf("直接开口：观点 -> 机制 -> 结果，围绕“%s”连成30秒回答。", focus),
	}
}

func looksTemplateLikeHint(text string) bool {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return true
	}
	templateMarkers := []string{
		"结论-依据",
		"结论 - 依据",
		"三步回答",
		"四步",
		"讲完整",
		"按“",
		"按\"",
		"套模板",
	}
	for _, marker := range templateMarkers {
		if strings.Contains(trimmed, marker) {
			return true
		}
	}
	return false
}

func extractShadowHintAnchors(referenceAnswer, knowledgeContext string) []string {
	merged := strings.TrimSpace(referenceAnswer + "\n" + knowledgeContext)
	if merged == "" {
		return nil
	}
	segments := strings.FieldsFunc(merged, func(r rune) bool {
		switch r {
		case '\n', '。', '；', ';', '，', ',', '、', '：', ':', '（', '）', '(', ')':
			return true
		default:
			return false
		}
	})

	anchors := make([]string, 0, 6)
	seen := map[string]bool{}
	for _, seg := range segments {
		item := strings.TrimSpace(seg)
		if item == "" {
			continue
		}
		runes := []rune(item)
		if len(runes) < 4 {
			continue
		}
		if len(runes) > 20 {
			item = string(runes[:20])
		}
		if seen[item] {
			continue
		}
		seen[item] = true
		anchors = append(anchors, item)
		if len(anchors) >= 6 {
			break
		}
	}
	return anchors
}

func containsAnyShadowAnchor(hint string, anchors []string) bool {
	if strings.TrimSpace(hint) == "" || len(anchors) == 0 {
		return false
	}
	for _, anchor := range anchors {
		if strings.TrimSpace(anchor) == "" {
			continue
		}
		if strings.Contains(hint, anchor) {
			return true
		}
	}
	return false
}

// GenerateShadowCoachHintLevels generates three escalating hints in one LLM call.
func (s *AIService) GenerateShadowCoachHintLevels(position, question, transcript, style, referenceAnswer, knowledgeContext string) ([]string, error) {
	trimmedQuestion := strings.TrimSpace(question)
	fallbacks := buildShadowHintFallbacks(trimmedQuestion)
	anchors := extractShadowHintAnchors(referenceAnswer, knowledgeContext)
	anchorText := "无"
	if len(anchors) > 0 {
		anchorText = strings.Join(anchors, "、")
	}
	if trimmedQuestion == "" {
		return fallbacks, nil
	}

	prompt := fmt.Sprintf(`你是“影子教练”，要给面试者三层递进提示。
请严格输出 JSON：
{
  "level_1": "...",
  "level_2": "...",
  "level_3": "..."
}

要求：
1) 全部简体中文，每层1-2句，单层不超过45字。
2) 三层强度递进：L1轻提醒，L2更明确结构，L3接近作答框架。
3) 可以参考“题目参考答案/知识库片段”提炼逻辑提示，但禁止逐句复述原答案。
4) L3 要足够强，能让“学过这题的人”立刻想起答题路径。
5) 禁止直接给出最终完整答案段落。
6) 每层句式要明显不同，禁止反复使用“按xxx三步/四步”这类模板句。
7) 每层尽量带上题目关键词，语气像耳返提醒，不是讲义。
8) 若“可用锚点关键词”不为“无”，L2 和 L3 至少提到一个锚点关键词。
9) 不要输出除 JSON 外的任何内容。

岗位：%s
面试官风格：%s
当前问题：%s
候选人已转写内容（可能为空）：%s
题目参考答案（可能为空）：%s
知识库相关片段（可能为空）：%s
可用锚点关键词：%s
`, position, style, trimmedQuestion, strings.TrimSpace(transcript), strings.TrimSpace(referenceAnswer), strings.TrimSpace(knowledgeContext), anchorText)

	raw, err := s.callLLM(prompt, "shadow_hint")
	if err != nil {
		return fallbacks, nil
	}

	var parsed struct {
		Level1 string `json:"level_1"`
		Level2 string `json:"level_2"`
		Level3 string `json:"level_3"`
	}
	if unmarshalErr := json.Unmarshal([]byte(extractJSONContent(raw)), &parsed); unmarshalErr != nil {
		return fallbacks, nil
	}

	hints := []string{
		normalizeShadowHintText(parsed.Level1, fallbacks[0]),
		normalizeShadowHintText(parsed.Level2, fallbacks[1]),
		normalizeShadowHintText(parsed.Level3, fallbacks[2]),
	}

	for idx := range hints {
		if looksTemplateLikeHint(hints[idx]) {
			hints[idx] = fallbacks[idx]
		}
	}

	if hints[1] == hints[0] {
		hints[1] = fallbacks[1]
	}
	if hints[2] == hints[1] || hints[2] == hints[0] {
		hints[2] = fallbacks[2]
	}

	if len(anchors) > 0 {
		if !containsAnyShadowAnchor(hints[1], anchors) {
			hints[1] = normalizeShadowHintText(
				fmt.Sprintf("围绕“%s”补一句关键机制，再落一个业务动作。", anchors[0]),
				fallbacks[1],
			)
		}
		if !containsAnyShadowAnchor(hints[2], anchors) {
			hints[2] = normalizeShadowHintText(
				fmt.Sprintf("直接按“观点 -> 机制 -> 结果”开口，并点到“%s”。", anchors[0]),
				fallbacks[2],
			)
		}
	}

	return hints, nil
}

// GenerateShadowCoachHint returns a short nudge when the candidate is stuck.
// It must avoid giving direct answers and keep hints actionable.
func (s *AIService) GenerateShadowCoachHint(position, question, transcript, style string, silenceSeconds int) (string, error) {
	hints, err := s.GenerateShadowCoachHintLevels(position, question, transcript, style, "", "")
	if err != nil {
		return buildShadowHintFallbacks(question)[0], nil
	}
	if len(hints) == 0 {
		return buildShadowHintFallbacks(question)[0], nil
	}
	if silenceSeconds >= 60 && len(hints) >= 3 {
		return hints[2], nil
	}
	if silenceSeconds >= 40 && len(hints) >= 2 {
		return hints[1], nil
	}
	return hints[0], nil
}

func (s *AIService) EvaluateAnswer(question *model.Question, answer string) (*EvaluationResult, error) {
	// 【已替换】接入 AIReview 强校验流程
	// 调用底层 LLM 的闭包函数
	llmFunc := func(p string) (string, error) {
		return s.callLLM(p, "evaluation")
	}

	// 使用 EvaluateCandidateAnswer 进行严苛评估
	reviewResult, err := EvaluateCandidateAnswer(question.Content, answer, llmFunc)
	if err != nil {
		log.Printf("AIReview failed, using local heuristic fallback: %v", err)
		return s.evaluateAnswerLocal(question, answer), nil
	}

	// 构造多维度结构化反馈 JSON
	evaluationText := s.EnsureChineseOutput(reviewResult.Comment, "回答已收到，但内容质量有待提升。")
	suggestionText := s.EnsureChineseOutput(reviewResult.Suggestion, "建议补充核心原理与实践细节。")

	// 将 suggestion 按分号拆分为多条建议
	suggestionItems := splitSuggestionText(suggestionText)

	// 构建维度数据（如果 LLM 没返回维度则根据总分推算）
	dims := reviewResult.Dimensions
	if dims == nil {
		dims = estimateDimensions(reviewResult.Score)
	}

	richFeedback := map[string]interface{}{
		"evaluation":           evaluationText,
		"suggestions":          suggestionItems,
		"dimensions":           dims,
		"highlights":           reviewResult.Highlights,
		"gaps":                 reviewResult.Gaps,
		"model_answer_outline": reviewResult.ModelAnswerOutline,
		"follow_up":            reviewResult.FollowUp,
	}
	feedbackJSON, _ := json.Marshal(richFeedback)

	return &EvaluationResult{
		Score:    reviewResult.Score,
		Feedback: string(feedbackJSON),
	}, nil
}

// splitSuggestionText 将分号/句号分隔的建议拆为独立条目
func splitSuggestionText(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return []string{"建议补充核心原理与实践细节。"}
	}
	// 先尝试分号拆分
	parts := strings.FieldsFunc(text, func(r rune) bool {
		return r == ';' || r == '；'
	})
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		// 去除前导数字编号
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

// estimateDimensions 当 LLM 未返回维度分时，根据总分推算合理的各维度分数
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
	if content == "" {
		return s.buildRichLocalFeedback(35, question.Content,
			"回答内容过短，尚未覆盖题目核心点。",
			[]string{"先给出结论，再解释原理", "补充具体项目经历与结果", "至少给出1个边界情况或异常处理思路"},
			&ReviewDimensions{TechnicalDepth: 15, Expression: 30, Logic: 25, Completeness: 10},
			nil,
			[]string{"未能针对问题给出任何实质性内容"},
		)
	}

	lengthScore := 35
	runeLen := len([]rune(content))
	switch {
	case runeLen >= 280:
		lengthScore = 70
	case runeLen >= 180:
		lengthScore = 62
	case runeLen >= 120:
		lengthScore = 54
	case runeLen >= 80:
		lengthScore = 46
	}

	structureBonus := 0
	hasStructure := false
	if strings.Contains(content, "首先") || strings.Contains(content, "第一") {
		structureBonus += 6
		hasStructure = true
	}
	if strings.Contains(content, "其次") || strings.Contains(content, "然后") {
		structureBonus += 5
		hasStructure = true
	}
	if strings.Contains(content, "最后") || strings.Contains(content, "总结") {
		structureBonus += 5
		hasStructure = true
	}

	questionText := strings.TrimSpace(question.Content + " " + question.Title)
	keywordBonus := 0
	matchedKeywords := 0
	for _, token := range strings.Fields(questionText) {
		t := strings.TrimSpace(token)
		if len([]rune(t)) < 2 {
			continue
		}
		if strings.Contains(content, t) {
			keywordBonus += 2
			matchedKeywords++
		}
		if keywordBonus >= 12 {
			break
		}
	}

	score := clampScore(lengthScore + structureBonus + keywordBonus)

	var evaluation string
	var highlights []string
	var gaps []string

	if score >= 80 {
		evaluation = "回答结构完整，覆盖了核心要点，表达较清晰，整体表现良好。"
		highlights = []string{"答案覆盖了主要考点", "表达有条理"}
		gaps = []string{"可进一步补充底层原理分析"}
	} else if score >= 60 {
		evaluation = "回答思路基本清晰，能够围绕题目展开，但在细节深度和案例支撑方面仍有提升空间。"
		if hasStructure {
			highlights = []string{"使用了结构化表达方式"}
		}
		gaps = []string{"缺少底层原理或源码层面的深入分析", "未结合实际项目案例论证"}
	} else {
		evaluation = "回答覆盖面偏窄，关键点阐述不够充分，建议进一步补充技术细节与实际场景。"
		gaps = []string{"核心概念阐述不充分", "缺少实际案例支撑", "表达深度有待提升"}
	}

	dims := &ReviewDimensions{
		TechnicalDepth: clampScore(score - 8),
		Expression:     clampScore(score + 5),
		Logic:          score,
		Completeness:   clampScore(score - 5),
	}
	if hasStructure {
		dims.Expression = clampScore(dims.Expression + 5)
		dims.Logic = clampScore(dims.Logic + 3)
	}
	if matchedKeywords >= 3 {
		dims.Completeness = clampScore(dims.Completeness + 8)
	}

	suggestions := []string{
		"按结论、原理、实践案例、风险与优化的顺序组织回答",
		"补充可量化结果（性能提升、耗时降低、错误率变化等）",
		"增加边界条件与异常处理说明，体现工程化能力",
	}

	return s.buildRichLocalFeedback(score, question.Content, evaluation, suggestions, dims, highlights, gaps)
}

// buildRichLocalFeedback 构建本地评估的丰富 JSON 反馈
func (s *AIService) buildRichLocalFeedback(score int, questionContent, evaluation string, suggestions []string, dims *ReviewDimensions, highlights, gaps []string) *EvaluationResult {
	richFeedback := map[string]interface{}{
		"evaluation":           evaluation,
		"suggestions":          suggestions,
		"dimensions":           dims,
		"highlights":           highlights,
		"gaps":                 gaps,
		"model_answer_outline": "建议从核心概念定义出发，结合底层原理、典型使用场景和注意事项进行系统阐述。",
		"follow_up":            "能否结合你的项目经历，更具体地描述你是如何应用这个技术的？",
	}
	feedbackJSON, _ := json.Marshal(richFeedback)
	return &EvaluationResult{
		Score:    score,
		Feedback: string(feedbackJSON),
	}
}

func (s *AIService) ensureChineseFeedback(feedback string) string {
	text := strings.TrimSpace(feedback)
	if text == "" {
		return "回答内容已收到，建议补充更具体的技术细节与实践案例。"
	}

	hanCount := 0
	totalCount := 0
	for _, r := range text {
		if r == '\n' || r == '\r' || r == '\t' || r == ' ' {
			continue
		}
		totalCount++
		if r >= 0x4E00 && r <= 0x9FFF {
			hanCount++
		}
	}
	if totalCount == 0 {
		return "回答内容已收到，建议补充更具体的技术细节与实践案例。"
	}
	if float64(hanCount)/float64(totalCount) >= 0.2 {
		return text
	}

	prompt := fmt.Sprintf("请将下面的面试反馈改写为简洁、专业、完全中文的两段文本：第一段是评价，第二段是建议。不要输出JSON，不要输出英文。\n\n%s", text)
	translated, err := s.callLLM(prompt, "chat")
	if err != nil {
		return "你的回答信息量偏少，尚未完整覆盖问题核心。建议补充关键原理、实际例子和边界情况，以提升答案完整度。"
	}
	translated = normalizeFeedbackText(translated)
	han2 := 0
	total2 := 0
	for _, r := range translated {
		if r == '\n' || r == '\r' || r == '\t' || r == ' ' {
			continue
		}
		total2++
		if r >= 0x4E00 && r <= 0x9FFF {
			han2++
		}
	}
	if total2 == 0 || float64(han2)/float64(total2) < 0.2 {
		return "你的回答信息量偏少，尚未完整覆盖问题核心。建议补充关键原理、实际例子和边界情况，以提升答案完整度。"
	}
	return translated
}

func (s *AIService) EnsureChineseOutput(text, fallback string) string {
	normalized := normalizeFeedbackText(text)
	if normalized == "" {
		return fallback
	}
	if isMostlyChinese(normalized, 0.45) {
		return normalized
	}

	prompt := fmt.Sprintf("请将以下内容改写为自然、专业、简体中文。要求只输出中文内容，不要输出JSON或英文：\n\n%s", normalized)
	rewritten, err := s.callLLM(prompt, "chat")
	if err != nil {
		return fallback
	}
	rewritten = normalizeFeedbackText(rewritten)
	if rewritten == "" || !isMostlyChinese(rewritten, 0.45) {
		return fallback
	}
	return rewritten
}

func (s *AIService) EnsureQuestionChinese(question *model.Question) {
	if question == nil {
		return
	}

	needRewrite := !isMostlyChinese(question.Title, 0.3) || !isMostlyChinese(question.Content, 0.35) || !isMostlyChinese(question.ExpectedAnswer, 0.3)
	if !needRewrite {
		question.Title = strings.TrimSpace(question.Title)
		question.Content = strings.TrimSpace(question.Content)
		question.ExpectedAnswer = strings.TrimSpace(question.ExpectedAnswer)
		return
	}

	prompt := fmt.Sprintf(`
请将下面的面试题改写为简体中文，保持语义一致且表达专业。
只返回 JSON 对象，不要输出其它内容：
{
  "title": "中文标题",
  "content": "中文问题内容",
  "expected_answer": "中文期望要点"
}

原始标题：%s
原始内容：%s
原始期望答案：%s
`, question.Title, question.Content, question.ExpectedAnswer)

	response, err := s.callLLM(prompt, "chat")
	if err == nil {
		var localized struct {
			Title          string `json:"title"`
			Content        string `json:"content"`
			ExpectedAnswer string `json:"expected_answer"`
		}
		if unmarshalErr := json.Unmarshal([]byte(extractJSONContent(response)), &localized); unmarshalErr == nil {
			if strings.TrimSpace(localized.Title) != "" {
				question.Title = localized.Title
			}
			if strings.TrimSpace(localized.Content) != "" {
				question.Content = localized.Content
			}
			if strings.TrimSpace(localized.ExpectedAnswer) != "" {
				question.ExpectedAnswer = localized.ExpectedAnswer
			}
		}
	}

	if !isMostlyChinese(question.Title, 0.3) {
		question.Title = "技术问题"
	}
	if !isMostlyChinese(question.Content, 0.35) {
		question.Content = "请结合实际项目经验，系统说明你的思路、关键实现和取舍。"
	}
	if !isMostlyChinese(question.ExpectedAnswer, 0.3) {
		question.ExpectedAnswer = "回答应包含核心原理、实现步骤、关键细节与风险边界。"
	}

	question.Title = strings.TrimSpace(question.Title)
	question.Content = strings.TrimSpace(question.Content)
	question.ExpectedAnswer = strings.TrimSpace(question.ExpectedAnswer)
}

func parseLooseEvaluation(raw string) *EvaluationResult {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil
	}

	score := 60
	scoreKeys := []string{"score", "overall_score", "rating", "final_score"}
	for _, key := range scoreKeys {
		if v, ok := data[key]; ok {
			switch t := v.(type) {
			case float64:
				score = int(t)
			case string:
				if n, err := strconv.Atoi(strings.TrimSpace(t)); err == nil {
					score = n
				}
			}
			break
		}
	}
	if score < 0 || score > 100 {
		score = 60
	}

	textKeys := []string{"feedback", "analysis", "comment", "summary", "advice", "suggestion"}
	parts := make([]string, 0, len(textKeys))
	for _, key := range textKeys {
		if v, ok := data[key]; ok {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				parts = append(parts, strings.TrimSpace(s))
				continue
			}
			if key == "feedback" {
				if m, ok := v.(map[string]interface{}); ok {
					if c, ok := m["content"].(string); ok && strings.TrimSpace(c) != "" {
						parts = append(parts, strings.TrimSpace(c))
					}
					if arr, ok := m["suggestions"].([]interface{}); ok {
						for _, item := range arr {
							if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
								parts = append(parts, "建议："+strings.TrimSpace(s))
							}
						}
					}
				}
			}
		}
	}
	if len(parts) == 0 {
		return nil
	}

	return &EvaluationResult{
		Score:    score,
		Feedback: normalizeFeedbackText(strings.Join(parts, "\n")),
	}
}

func normalizeFeedbackText(s string) string {
	text := strings.TrimSpace(s)
	if text == "" {
		return "回答内容已收到，建议补充更具体的技术细节与实践案例。"
	}

	if strings.HasPrefix(text, "{") || strings.HasPrefix(text, "[") {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(text), &obj); err == nil {
			textKeys := []string{"feedback", "analysis", "comment", "summary", "advice", "suggestion"}
			parts := make([]string, 0, len(textKeys))
			for _, key := range textKeys {
				if v, ok := obj[key]; ok {
					if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
						parts = append(parts, strings.TrimSpace(s))
						continue
					}
					if key == "feedback" {
						if m, ok := v.(map[string]interface{}); ok {
							if c, ok := m["content"].(string); ok && strings.TrimSpace(c) != "" {
								parts = append(parts, strings.TrimSpace(c))
							}
							if arr, ok := m["suggestions"].([]interface{}); ok {
								for _, item := range arr {
									if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
										parts = append(parts, "建议："+strings.TrimSpace(s))
									}
								}
							}
						}
					}
				}
			}
			if len(parts) > 0 {
				text = strings.Join(parts, "\n")
			} else {
				text = "回答内容已收到，建议补充更具体的技术细节与实践案例。"
			}
		}
	}

	re := regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")
	text = strings.TrimSpace(strings.Trim(text, "`"))
	return text
}

func extractJSONContent(raw string) string {
	text := strings.TrimSpace(raw)
	if strings.HasPrefix(text, "```json") {
		text = strings.TrimPrefix(text, "```json")
		text = strings.TrimSuffix(text, "```")
		return strings.TrimSpace(text)
	}
	if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```")
		text = strings.TrimSuffix(text, "```")
		return strings.TrimSpace(text)
	}
	return text
}

func isMostlyChinese(text string, ratio float64) bool {
	content := strings.TrimSpace(text)
	if content == "" {
		return false
	}
	hanCount := 0
	letterCount := 0
	for _, r := range content {
		if unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsDigit(r) {
			continue
		}
		if unicode.IsLetter(r) {
			letterCount++
		}
		if r >= 0x4E00 && r <= 0x9FFF {
			hanCount++
		}
	}
	if letterCount == 0 {
		return false
	}
	return float64(hanCount)/float64(letterCount) >= ratio
}

func buildStructuredFeedback(evaluation string, suggestions []string) string {
	parts := []string{fmt.Sprintf("【评价】%s", strings.TrimSpace(evaluation))}
	if len(suggestions) > 0 {
		parts = append(parts, "【建议】")
		for i, item := range suggestions {
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}
			parts = append(parts, fmt.Sprintf("%d. %s", i+1, item))
		}
	}
	return strings.Join(parts, "\n")
}

func (s *AIService) defaultSuggestionsByScore(score int) []string {
	if score >= 80 {
		return []string{
			"继续保持结构化表达，先结论后细节。",
			"补充一到两个真实项目数据来增强说服力。",
		}
	}
	if score >= 60 {
		return []string{
			"先给出核心结论，再按原理、实现、风险三个层次展开。",
			"补充关键技术细节和边界条件，避免泛泛而谈。",
		}
	}
	return []string{
		"先明确问题核心，再按步骤组织答案。",
		"至少补充一个项目实例，说明你的实际做法和结果。",
		"回答中加入关键原理和取舍依据，提升完整度。",
	}
}

func (s *AIService) adjustScoreByAnswerQuality(baseScore int, question *model.Question, answer string) int {
	score := baseScore
	if score < 0 || score > 100 {
		score = 60
	}
	content := strings.TrimSpace(answer)
	lower := strings.ToLower(content)

	if content == "" {
		return 20
	}
	if isNonsenseAnswer(lower) {
		if score > 35 {
			score = 35
		}
	}
	if strings.Contains(lower, "不知道") || strings.Contains(lower, "不会") || strings.Contains(lower, "不太清楚") {
		if score > 40 {
			score = 40
		}
	}
	runes := []rune(content)
	if len(runes) < 15 {
		if score > 45 {
			score = 45
		}
	} else if len(runes) > 150 && score < 70 {
		score = 70
	}

	if question != nil {
		overlap := keywordOverlapRatio(question, content)
		if overlap == 0 {
			if len(runes) < 80 && score > 50 {
				score = 50
			}
		} else if overlap < 0.12 {
			if score > 58 {
				score = 58
			}
		}
	}
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return score
}

func isNonsenseAnswer(lower string) bool {
	trimmed := strings.TrimSpace(lower)
	if trimmed == "" {
		return true
	}
	badSamples := []string{"1", "2", "3", "asd", "aaaa", "111", "测试", "随便", "不知道", "不会"}
	for _, sample := range badSamples {
		if trimmed == sample {
			return true
		}
	}
	if len([]rune(trimmed)) <= 3 {
		return true
	}
	return false
}

func keywordOverlapRatio(question *model.Question, answer string) float64 {
	if question == nil {
		return 0
	}
	reference := strings.TrimSpace(question.ExpectedAnswer + " " + question.Content)
	if reference == "" {
		return 0
	}
	refTokens := extractKeywords(reference)
	ansTokens := extractKeywords(answer)
	if len(refTokens) == 0 || len(ansTokens) == 0 {
		return 0
	}
	hit := 0
	for token := range refTokens {
		if _, ok := ansTokens[token]; ok {
			hit++
		}
	}
	return float64(hit) / float64(len(refTokens))
}

func extractKeywords(text string) map[string]struct{} {
	re := regexp.MustCompile(`[\p{Han}]{2,}|[A-Za-z]{3,}`)
	tokens := re.FindAllString(strings.ToLower(text), -1)
	result := make(map[string]struct{}, len(tokens))
	for _, token := range tokens {
		result[token] = struct{}{}
	}
	return result
}

func (s *AIService) normalizeStructuredFeedback(feedback string, score int) string {
	raw := strings.TrimSpace(feedback)
	if raw == "" {
		raw = "你的回答覆盖了部分要点，但深度和细节还可以继续加强。"
	}
	evaluation := s.EnsureChineseOutput(raw, "你的回答覆盖了部分要点，但深度和细节还可以继续加强。")
	suggestions := s.defaultSuggestionsByScore(score)
	return buildStructuredFeedback(evaluation, suggestions)
}

func (s *AIService) GenerateQuestions(interview *model.Interview, count int) ([]*model.Question, error) {
	modeInstruction := buildModePrompt(interview.Mode)
	styleInstruction := buildStylePrompt(interview.Style, interview.Company)
	difficultyInstruction := buildDifficultyPrompt(interview.Difficulty)

	prompt := fmt.Sprintf(`
		请为以下面试场景生成 %d 个面试问题：
		
		职位：%s
		难度级别：%s
		面试模式：%s
		面试风格：%s
		
		%s

		%s

		%s

		要求：
		1. 问题应循序渐进，涵盖该职位的核心技能点。
		2. 问题应具有针对性，考察候选人的实际能力。
		3. 所有字段必须使用简体中文。
		4. 返回格式必须为 JSON 数组，每个对象包含 "title", "content", "expected_answer"。
		
		示例格式：
		[
			{"title": "问题1标题", "content": "问题1内容", "expected_answer": "期望回答1"},
			{"title": "问题2标题", "content": "问题2内容", "expected_answer": "期望回答2"}
		]
	`, count, interview.Position, interview.Difficulty, interview.Mode, interview.Style,
		modeInstruction, styleInstruction, difficultyInstruction)

	response, err := s.callLLM(prompt, "chat")
	if err != nil {
		return nil, fmt.Errorf("failed to generate questions: %w", err)
	}

	var questionsData []struct {
		Title          string `json:"title"`
		Content        string `json:"content"`
		ExpectedAnswer string `json:"expected_answer"`
	}

	cleanResponse := extractJSONContent(response)

	if err := json.Unmarshal([]byte(cleanResponse), &questionsData); err != nil {
		return nil, fmt.Errorf("failed to parse questions response: %w, body: %s", err, response)
	}

	var questions []*model.Question
	for _, qd := range questionsData {
		item := &model.Question{
			Title:          qd.Title,
			Content:        qd.Content,
			ExpectedAnswer: qd.ExpectedAnswer,
			Position:       interview.Position,
			Difficulty:     interview.Difficulty,
		}
		s.EnsureQuestionChinese(item)
		questions = append(questions, item)
	}

	return questions, nil
}

// GenerateTopicQuestionFromContext builds a topic-opening question using RAG context.
func (s *AIService) GenerateTopicQuestionFromContext(interview *model.Interview, ragContext string, category string) (*model.Question, error) {
	if interview == nil {
		return nil, fmt.Errorf("interview is nil")
	}
	modeInstruction := buildModePrompt(interview.Mode)
	styleInstruction := buildStylePrompt(interview.Style, interview.Company)
	difficultyInstruction := buildDifficultyPrompt(interview.Difficulty)

	prompt := fmt.Sprintf(`
		你是一位资深技术面试官。请基于以下检索到的知识片段，生成一个“话题引入题”。
		要求：
		1. 必须围绕知识片段的核心内容提问，作为话题的第一题。
		2. 题目清晰、可回答，不要过于宽泛。
		3. 语气必须是“开场提问”，禁止任何追问口吻或上下文引用，例如“你提到/你刚才/继续/进一步/补充说明/候选人提到/基于上一题”。
		4. 题面必须自包含，不能依赖前文，不得出现“继续、再次、补充、进一步”等承接词。
		5. 只使用简体中文。
		6. 返回格式：{"title": "问题标题", "content": "具体问题内容", "expected_answer": "简要的期望回答"}

		职位：%s
		难度级别：%s
		面试模式：%s
		面试风格：%s

		%s
		%s
		%s

		【知识片段】
		%s
	`, interview.Position, interview.Difficulty, interview.Mode, interview.Style,
		modeInstruction, styleInstruction, difficultyInstruction, ragContext)

	response, err := s.callLLM(prompt, "chat")
	if err != nil {
		return nil, fmt.Errorf("failed to generate topic question: %w", err)
	}

	var q struct {
		Title          string `json:"title"`
		Content        string `json:"content"`
		ExpectedAnswer string `json:"expected_answer"`
	}

	cleanResponse := extractJSONContent(response)
	if err := json.Unmarshal([]byte(cleanResponse), &q); err != nil {
		return nil, fmt.Errorf("failed to parse topic question response: %w, body: %s", err, response)
	}

	result := &model.Question{
		Title:          q.Title,
		Content:        q.Content,
		ExpectedAnswer: q.ExpectedAnswer,
		Position:       interview.Position,
		Difficulty:     interview.Difficulty,
		Category:       category,
	}
	s.EnsureQuestionChinese(result)
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
		topic = "该技术主题"
	}

	q.Title = fmt.Sprintf("%s核心原理与实践", topic)
	q.Content = fmt.Sprintf("请你系统讲解%s的核心原理、关键设计取舍，以及在实际项目中的应用方式。", topic)
	if strings.TrimSpace(q.ExpectedAnswer) == "" {
		q.ExpectedAnswer = fmt.Sprintf("应覆盖%s的定义、核心机制、适用场景、常见问题与优化思路。", topic)
	}
}

func isFollowUpWording(text string) bool {
	t := strings.TrimSpace(text)
	if t == "" {
		return false
	}
	patterns := []string{
		"你提到", "你刚才", "你刚刚", "继续", "进一步", "补充", "候选人提到", "基于上一题", "上一题", "再追问", "追问",
	}
	for _, p := range patterns {
		if strings.Contains(t, p) {
			return true
		}
	}
	return false
}

var openingQuestionContextPatterns = []string{
	"你刚才", "你刚刚", "你提到", "继续", "进一步", "补充", "上一题", "上一个问题", "前文", "上文", "上述", "如上", "再说",
	"这三个类", "这三类", "这几类", "这几个", "这些类", "这些对象", "它们", "分别说下", "分别解释一下",
}

var openingQuestionQuantifierRef = regexp.MustCompile(`这[一二两三四五六七八九十\d]+(个|类|种|点|方面|模块|对象|方法|步骤)`)

// IsContextDependentOpeningQuestion checks whether a supposed opening question depends on missing prior context.
func (s *AIService) IsContextDependentOpeningQuestion(question *model.Question) bool {
	if question == nil {
		return true
	}
	text := strings.TrimSpace(question.Title + " " + question.Content)
	if text == "" {
		return true
	}
	if isFollowUpWording(text) {
		return true
	}
	if openingQuestionQuantifierRef.MatchString(text) {
		return true
	}
	for _, p := range openingQuestionContextPatterns {
		if strings.Contains(text, p) {
			return true
		}
	}
	return false
}

// NormalizeToSelfContainedOpening rewrites a problematic opening question into a self-contained one.
func (s *AIService) NormalizeToSelfContainedOpening(question *model.Question) {
	if question == nil {
		return
	}
	topic := strings.TrimSpace(question.Category)
	if topic == "" {
		topic = strings.TrimSpace(question.Title)
	}
	if topic == "" {
		topic = "该技术主题"
	}

	question.Title = fmt.Sprintf("%s：核心原理与实践", topic)
	question.Content = fmt.Sprintf("请你围绕%s，完整说明核心概念、关键实现机制、线程安全性与性能特点，并给出典型使用场景。请不要省略被比较对象的名称。", topic)
	if strings.TrimSpace(question.ExpectedAnswer) == "" {
		question.ExpectedAnswer = fmt.Sprintf("应覆盖%s的定义、实现原理、线程安全边界、性能取舍与项目实践。", topic)
	}
	question.Title = strings.TrimSpace(question.Title)
	question.Content = strings.TrimSpace(question.Content)
	question.ExpectedAnswer = strings.TrimSpace(question.ExpectedAnswer)
}

// GenerateClarifyingFollowUpQuestion forces a follow-up based on the answer content (no hallucination).
func (s *AIService) GenerateClarifyingFollowUpQuestion(currentQ *model.Question, answer string, followUpIndex int) (*model.Question, error) {
	if currentQ == nil {
		return nil, fmt.Errorf("current question is nil")
	}
	prompt := fmt.Sprintf(`
		你是一位资深技术面试官。候选人的回答信息不足，但仍需要继续追问。
		要求：
		1. 追问必须基于候选人回答的内容或其“信息不足”这一事实，禁止编造候选人未提到的信息。
		2. 可以要求补充理由、细节、边界条件或实现步骤。
		3. 只使用简体中文。
		4. 返回格式：{"title": "追问标题", "content": "追问具体内容", "expected_answer": "期望回答要点"}

		【当前问题】
		标题：%s
		内容：%s

		【候选人回答】
		%s

		【当前追问次数】%d
	`, currentQ.Title, currentQ.Content, answer, followUpIndex)

	response, err := s.callLLM(prompt, "chat")
	if err != nil {
		return nil, fmt.Errorf("failed to generate clarifying follow-up: %w", err)
	}

	var q struct {
		Title          string `json:"title"`
		Content        string `json:"content"`
		ExpectedAnswer string `json:"expected_answer"`
	}

	cleanResponse := extractJSONContent(response)
	if err := json.Unmarshal([]byte(cleanResponse), &q); err != nil {
		return nil, fmt.Errorf("failed to parse clarifying follow-up: %w, body: %s", err, response)
	}

	result := &model.Question{
		Title:          q.Title,
		Content:        q.Content,
		ExpectedAnswer: q.ExpectedAnswer,
		Position:       currentQ.Position,
		Difficulty:     currentQ.Difficulty,
		Category:       currentQ.Category,
	}
	s.EnsureQuestionChinese(result)
	return result, nil
}

// GenerateNextQuestionWithWeights generates next question considering capability weights
func (s *AIService) GenerateNextQuestionWithWeights(interview *model.Interview, previousAnswers []model.AnswerResult, capabilityGraph *model.JobCapabilityDimension) (*model.Question, error) {
	// If no capability graph provided, fallback to standard generation
	if capabilityGraph == nil {
		return s.GenerateNextQuestion(interview, previousAnswers)
	}

	modeInstruction := buildModePrompt(interview.Mode)
	styleInstruction := buildStylePrompt(interview.Style, interview.Company)
	difficultyInstruction := buildDifficultyPrompt(interview.Difficulty)

	// Build weights instruction
	var weightsBuilder strings.Builder
	weightsBuilder.WriteString("【岗位能力侧重】\n")
	weightsBuilder.WriteString(fmt.Sprintf("- %s: 权重 %d%%\n", capabilityGraph.Name, capabilityGraph.Weight))
	for _, sub := range capabilityGraph.SubDimensions {
		weightsBuilder.WriteString(fmt.Sprintf("  - %s (权重 %d%%): %s\n", sub.Name, sub.Weight, strings.Join(sub.Tags, ", ")))
	}

	// Calculate which dimension needs more coverage based on previous questions
	// This is a simplified logic; in a real system, we'd track coverage per dimension
	nextFocus := "请根据上述权重分布，选择一个尚未充分考察或权重较高的维度进行提问。"

	prompt := fmt.Sprintf(`
		基于以下面试信息和岗位能力图谱，生成下一个合适的面试问题：
		
		职位：%s
		难度级别：%s
		面试模式：%s
		面试风格：%s
		已回答问题数量：%d
		
		%s
		%s
		%s
		
		%s

		%s

		请生成一个合适的后续问题，以深入了解候选人的技术能力。
		只使用简体中文。
		返回格式：{"title": "问题标题", "content": "具体问题内容", "expected_answer": "简要的期望回答"}
	`, interview.Position, interview.Difficulty, interview.Mode, interview.Style, len(previousAnswers),
		modeInstruction, styleInstruction, difficultyInstruction, weightsBuilder.String(), nextFocus)

	response, err := s.callLLM(prompt, "chat")
	if err != nil {
		return nil, fmt.Errorf("failed to generate question: %w", err)
	}

	var question struct {
		Title          string `json:"title"`
		Content        string `json:"content"`
		ExpectedAnswer string `json:"expected_answer"`
	}

	cleanResponse := extractJSONContent(response)

	if err := json.Unmarshal([]byte(cleanResponse), &question); err != nil {
		return nil, fmt.Errorf("failed to parse question response: %w, body: %s", err, response)
	}

	result := &model.Question{
		Title:          question.Title,
		Content:        question.Content,
		ExpectedAnswer: question.ExpectedAnswer,
		Position:       interview.Position,
		Difficulty:     interview.Difficulty,
	}
	s.EnsureQuestionChinese(result)
	return result, nil
}

func (s *AIService) GenerateNextQuestion(interview *model.Interview, previousAnswers []model.AnswerResult) (*model.Question, error) {
	modeInstruction := buildModePrompt(interview.Mode)
	styleInstruction := buildStylePrompt(interview.Style, interview.Company)
	difficultyInstruction := buildDifficultyPrompt(interview.Difficulty)

	prompt := fmt.Sprintf(`
		基于以下面试信息，生成下一个合适的面试问题：
		
		职位：%s
		难度级别：%s
		面试模式：%s
		面试风格：%s
		已回答问题数量：%d
		
		%s
		%s
		%s

		请生成一个合适的后续问题，以深入了解候选人的技术能力。
		只使用简体中文。
		返回格式：{"title": "问题标题", "content": "具体问题内容", "expected_answer": "简要的期望回答"}
	`, interview.Position, interview.Difficulty, interview.Mode, interview.Style, len(previousAnswers),
		modeInstruction, styleInstruction, difficultyInstruction)

	response, err := s.callLLM(prompt, "chat")
	if err != nil {
		return nil, fmt.Errorf("failed to generate question: %w", err)
	}

	var question struct {
		Title          string `json:"title"`
		Content        string `json:"content"`
		ExpectedAnswer string `json:"expected_answer"`
	}

	cleanResponse := extractJSONContent(response)

	if err := json.Unmarshal([]byte(cleanResponse), &question); err != nil {
		return nil, fmt.Errorf("failed to parse question response: %w, body: %s", err, response)
	}

	result := &model.Question{
		Title:          question.Title,
		Content:        question.Content,
		ExpectedAnswer: question.ExpectedAnswer,
		Position:       interview.Position,
		Difficulty:     interview.Difficulty,
	}
	s.EnsureQuestionChinese(result)
	return result, nil
}

func (s *AIService) GenerateFollowUpQuestion(interview *model.Interview, currentQ *model.Question, answer string, ragContext string, followUpIndex int) (*model.Question, string, error) {
	// Decide if follow-up is needed based on answer quality and depth
	// This is a simplified prompt. In production, we might want a two-step process: Analyze -> Generate
	mode := "technical"
	style := "gentle"
	difficulty := "campus_intern"
	company := ""
	modeInstruction := buildModePrompt(mode)
	styleInstruction := buildStylePrompt(style, company)
	difficultyInstruction := buildDifficultyPrompt(difficulty)
	if interview != nil {
		mode = interview.Mode
		style = interview.Style
		difficulty = interview.Difficulty
		company = interview.Company
		modeInstruction = buildModePrompt(mode)
		styleInstruction = buildStylePrompt(style, company)
		difficultyInstruction = buildDifficultyPrompt(difficulty)
	}

	prompt := fmt.Sprintf(`
		你是一位资深技术面试官。候选人刚刚回答了你的问题。

		【面试上下文】
		面试类型：%s
		面试风格：%s
		难度等级：%s

		%s
		%s
		%s
		
		【当前问题】
		标题：%s
		内容：%s
		
		【候选人回答】
		%s
		
		【相关知识库上下文】
		%s
		
		【当前追问次数】
		%d
		
		请分析候选人的回答：
		1. 如果回答非常完美、全面，且没有明显的漏洞或值得深挖的点，则不需要追问，返回 "NO_FOLLOWUP"。
		2. 如果回答存在模糊不清、逻辑漏洞、或者提到了值得深入的技术点（特别是项目经验或底层原理），请生成一个追问问题。
		3. 追问必须与候选人的回答强关联，循序渐进地深入同一话题，禁止突然切换到无关主题。
		4. 追问只能基于候选人已回答的内容，不得假设或编造其未提到的信息。
		5. 追问措辞避免使用“你提到了X”这类句式，除非 X 在回答中明确出现。
		6. 如果这是第3次追问，除非非常必要，否则尽量结束该话题，返回 "NO_FOLLOWUP"。
		
		如果需要追问，请返回 JSON 格式：
		{
			"follow_up_needed": true,
			"reason": "追问理由（简短中文）",
			"question": {
				"title": "追问标题",
				"content": "追问具体内容",
				"expected_answer": "期望回答要点"
			}
		}
		
		如果不需要追问，请返回 JSON 格式：
		{
			"follow_up_needed": false,
			"reason": "回答已足够完整/话题已结束"
		}
	`, mode, style, difficulty, modeInstruction, styleInstruction, difficultyInstruction, currentQ.Title, currentQ.Content, answer, ragContext, followUpIndex)

	response, err := s.callLLM(prompt, "chat")
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

	cleanResponse := extractJSONContent(response)
	if err := json.Unmarshal([]byte(cleanResponse), &result); err != nil {
		// If parsing fails, assume no follow-up to be safe
		log.Printf("Failed to parse follow-up response: %v, body: %s", err, response)
		return nil, "", nil
	}

	if !result.FollowUpNeeded {
		return nil, result.Reason, nil
	}

	q := &model.Question{
		Title:          result.Question.Title,
		Content:        result.Question.Content,
		ExpectedAnswer: result.Question.ExpectedAnswer,
		Position:       currentQ.Position,
		Difficulty:     currentQ.Difficulty,
		Category:       currentQ.Category, // Inherit category
	}
	s.EnsureQuestionChinese(q)

	return q, result.Reason, nil
}

func (s *AIService) TranscribeAudio(audioData string) (string, error) {
	decodedAudio, err := base64.StdEncoding.DecodeString(audioData)
	if err != nil {
		return "", fmt.Errorf("failed to decode audio data: %w", err)
	}

	asrConfig := config.GetConfig().ASR
	if asrConfig.MaxAudioBytes > 0 && len(decodedAudio) > asrConfig.MaxAudioBytes {
		return "", fmt.Errorf("audio too large: %d bytes (max %d)", len(decodedAudio), asrConfig.MaxAudioBytes)
	}

	if asrConfig.Provider == "whisper" || asrConfig.Provider == "openai" || asrConfig.Provider == "" {
		return s.transcribeWithWhisper(decodedAudio)
	}

	return "", fmt.Errorf("unsupported ASR provider: %s", asrConfig.Provider)
}

func (s *AIService) SynthesizeSpeech(text string) ([]byte, error) {
	ttsConfig := config.GetConfig().TTS
	if !ttsConfig.Enabled {
		return nil, fmt.Errorf("tts is disabled")
	}

	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil, fmt.Errorf("text is empty")
	}
	if ttsConfig.MaxCharsPerRequest > 0 && len([]rune(trimmed)) > ttsConfig.MaxCharsPerRequest {
		trimmed = string([]rune(trimmed)[:ttsConfig.MaxCharsPerRequest])
	}

	baseURL := strings.TrimSpace(ttsConfig.BaseURL)
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	url := strings.TrimSuffix(baseURL, "/") + "/audio/speech"

	model := strings.TrimSpace(ttsConfig.Model)
	if model == "" {
		model = "tts-1-1106"
	}
	voice := strings.TrimSpace(ttsConfig.Voice)
	if voice == "" {
		voice = "alloy"
	}

	bodyMap := map[string]interface{}{
		"model":           model,
		"input":           trimmed,
		"voice":           voice,
		"response_format": "mp3",
	}
	body, err := json.Marshal(bodyMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tts request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create tts request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(ttsConfig.APIKey) != "" {
		req.Header.Set("Authorization", "Bearer "+ttsConfig.APIKey)
	}

	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call tts api: %w", err)
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tts api returned status: %d, body: %s", resp.StatusCode, string(respBytes))
	}

	return respBytes, nil
}

func (s *AIService) callLLM(prompt string, taskType string) (string, error) {
	baseURL := s.config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.deepseek.com/v1"
	}

	// 自动拼接 /chat/completions
	url := baseURL
	if !strings.HasSuffix(url, "/chat/completions") {
		url = strings.TrimSuffix(url, "/") + "/chat/completions"
	}

	if strings.HasSuffix(url, "/chat/completions/chat/completions") {
		url = strings.Replace(url, "/chat/completions/chat/completions", "/chat/completions", 1)
	}

	// Select model based on taskType
	model := s.config.Model // Default model
	if specificModel, ok := s.config.Models[taskType]; ok && specificModel != "" {
		model = specificModel
	}

	log.Printf("Calling LLM API: %s, Model: %s, Task: %s", url, model, taskType)

	// Ensure max_tokens is within reasonable limits (some models reject >4096 or have strict limits)
	// maxTokens := 2000
	if model == "glm-4-flash" {
		// Specific adjustment if needed
	}

	requestBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature": 0.7,
		// "max_tokens":  maxTokens, // Comment out max_tokens if it causes issues with some providers
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LLM API returned status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Check if response is just a string (some proxies do this?) no, standard is JSON.
	// But let's handle the case where provider returns error in body even with 200 OK?
	// The standard OpenAI format:
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string      `json:"message"`
			Code    interface{} `json:"code"`
		} `json:"error"`
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w, body: %s", err, string(bodyBytes))
	}

	if result.Error.Message != "" {
		return "", fmt.Errorf("LLM API error: %s", result.Error.Message)
	}

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no response from LLM, body: %s", string(bodyBytes))
}

func (s *AIService) GenerateOverallAnalysis(interview *model.Interview, answers []model.AnswerResult) (string, error) {
	summary := struct {
		Position   string               `json:"position"`
		Difficulty string               `json:"difficulty"`
		Answers    []model.AnswerResult `json:"answers"`
	}{
		Position:   interview.Position,
		Difficulty: interview.Difficulty,
		Answers:    answers,
	}

	payload, err := json.Marshal(summary)
	if err != nil {
		return "", fmt.Errorf("failed to marshal summary: %w", err)
	}

	prompt := fmt.Sprintf(`
你是资深技术面试官。请基于以下面试整体数据，输出一段中文的综合分析，给出候选人的优势、薄弱点以及改进建议，长度不超过400字：

数据：
%s
`, string(payload))

	response, err := s.callLLM(prompt, "report")
	if err != nil {
		return "", err
	}
	return s.EnsureChineseOutput(response, "候选人整体表现中等，建议补充关键原理、项目细节与结构化表达，以提升面试竞争力。"), nil
}

func (s *AIService) GenerateReportInsights(interview *model.Interview, answers []model.AnswerResult) (*ReportInsights, error) {
	payload := struct {
		Position   string               `json:"position"`
		Difficulty string               `json:"difficulty"`
		Mode       string               `json:"mode"`
		Style      string               `json:"style"`
		Answers    []model.AnswerResult `json:"answers"`
	}{
		Position:   interview.Position,
		Difficulty: interview.Difficulty,
		Mode:       interview.Mode,
		Style:      interview.Style,
		Answers:    answers,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal report payload: %w", err)
	}

	prompt := fmt.Sprintf(`
你是资深中文面试官，请根据下面面试数据输出报告 JSON。

要求：
1. 所有文本必须为简体中文。
2. 内容必须基于给定答案，不能空泛。
3. 打分范围 0-100。
4. strengths / weaknesses / suggestions 各输出 2-4 条。
5. 只输出 JSON，不要其它解释。

返回格式：
{
  "overall_analysis": "综合分析",
  "strengths": ["优势1", "优势2"],
  "weaknesses": ["不足1", "不足2"],
  "suggestions": ["建议1", "建议2"],
  "technical_score": 70,
  "expression_score": 68,
  "logic_score": 72,
  "matching_score": 66,
  "behavior_score": 74
}

面试数据：
%s
`, string(body))

	response, err := s.callLLM(prompt, "report")
	if err != nil {
		return nil, err
	}

	var insights ReportInsights
	if err := json.Unmarshal([]byte(extractJSONContent(response)), &insights); err != nil {
		return nil, fmt.Errorf("failed to parse report insights: %w", err)
	}

	insights.OverallAnalysis = s.EnsureChineseOutput(insights.OverallAnalysis, "本次面试表现中等，基础能力具备，但在回答深度和结构化表达方面仍有提升空间。")
	insights.Strengths = ensureChineseList(s, insights.Strengths, []string{"具备一定基础知识储备", "回答态度积极"})
	insights.Weaknesses = ensureChineseList(s, insights.Weaknesses, []string{"部分回答不够深入", "关键细节覆盖不足"})
	insights.Suggestions = ensureChineseList(s, insights.Suggestions, []string{"按结论-原理-实践结构组织回答", "补充项目案例和量化结果"})

	insights.TechnicalScore = clampScore(insights.TechnicalScore)
	insights.ExpressionScore = clampScore(insights.ExpressionScore)
	insights.LogicScore = clampScore(insights.LogicScore)
	insights.MatchingScore = clampScore(insights.MatchingScore)
	insights.BehaviorScore = clampScore(insights.BehaviorScore)

	return &insights, nil
}

func ensureChineseList(s *AIService, items []string, fallback []string) []string {
	clean := make([]string, 0, len(items))
	for _, item := range items {
		line := strings.TrimSpace(item)
		if line == "" {
			continue
		}
		clean = append(clean, s.EnsureChineseOutput(line, "请补充更具体的面试表现分析。"))
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

func (s *AIService) transcribeWithWhisper(audioData []byte) (string, error) {
	asrConfig := config.GetConfig().ASR
	client := asr.NewWhisperClient(asrConfig.APIKey, asrConfig.BaseURL, asrConfig.Model)

	text, err := client.TranscribeAudio(audioData, "zh")
	if err != nil {
		return "", fmt.Errorf("whisper transcription failed: %w", err)
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return "", fmt.Errorf("empty transcription result")
	}

	return text, nil
}

// ========== Interview Mode / Style / Difficulty Prompt Builders ==========

// buildModePrompt generates prompt instructions based on interview type
func buildModePrompt(mode string) string {
	switch mode {
	case "technical":
		return `【面试类型：技术面】
- 所有题目必须聚焦技术能力考察。
- 涵盖：编程基础、数据结构与算法、系统设计、框架原理、性能优化等。
- 可以包含代码题、设计题和原理解释题。
- 不涉及行为面或软技能相关问题。`

	case "hr":
		return `【面试类型：HR面】
- 所有题目聚焦软实力与职业素养考察。
- 涵盖：自我介绍、职业规划、团队协作、抗压能力、价值观与文化匹配。
- 使用STAR法则场景题（情境+任务+行动+结果）。
- 包含：优势劣势分析、离职原因、薪资期望引导、冲突处理等经典HR问题。
- 不涉及具体技术实现细节。`

	case "comprehensive":
		return `【面试类型：综合面（技术+HR联合面试）】
- 前3题为技术面，考察核心技术能力。
- 后2题为HR面，考察软实力与职业匹配度。
- 技术题考察深度与广度，HR题需要有STAR法则场景。
- 模拟企业终面场景，同时由技术主管和HR共同评估。`

	default:
		return ""
	}
}

// companyStyleProfiles maps company names to their interview characteristics
var companyStyleProfiles = map[string]string{
	"ali": `【阿里面试风格】
- 重视系统设计与「大局观」，偏好追问"如果量级增大10倍怎么办"。
- 提问逻辑：先考基础 → 追问原理 → 延伸到实际业务场景。
- 注重候选人的"思考过程"而非只看最终答案。
- 常见追问："这个方案能否支撑双11级别的流量？瓶颈在哪？"
- 风格关键词：务实、重业务、追求ROI。`,

	"bytedance": `【字节跳动面试风格】
- 面试节奏非常快，讲究"逻辑倒推"和"5-why 追问"。
- 习惯连续深挖3-5层，每一层都要求候选人给出更底层的原理。
- 偏好算法思维与系统设计并重，代码能力要求严格。
- 典型追问链：概念 → 原理 → 源码 → 性能 → 边界case。
- 风格关键词：极致深挖、高效、不留情面。`,

	"tencent": `【腾讯面试风格】
- 面试氛围相对轻松，但技术深度不减。
- 偏好从项目经验出发，延伸到技术细节和方案权衡。
- 重视候选人的"沟通表达"能力和"技术视野"。
- 喜欢考察设计模式、架构演进、技术选型理由。
- 风格关键词：聊项目、讲思路、重视技术品味。`,

	"meituan": `【美团面试风格】
- 偏务实，题目贴近实际业务场景（高并发、分布式事务、数据一致性）。
- 重视候选人解决实际问题的能力而非纯理论。
- 常见场景：外卖系统设计、库存扣减、订单拆分、配送路径。
- 追问方向偏重"在实际项目中你是怎么做的？遇到什么坑？"
- 风格关键词：实战、业务导向、场景驱动。`,

	"baidu": `【百度面试风格】
- 技术基础考察非常扎实，重视数据结构和算法功底。
- 偏好考察候选人的工程素养和代码规范。
- 喜欢出设计类问题，考察搜索、推荐、NLP相关系统设计。
- 风格关键词：基础扎实、算法能力、工程规范。`,
}

// buildStylePrompt generates prompt instructions based on interviewer style and optional company
func buildStylePrompt(style, company string) string {
	var parts []string

	switch style {
	case "gentle":
		parts = append(parts, `【面试官风格：温和型】
- 语气友好、有引导性，像一场技术交流。
- 遇到候选人回答不上来时，给予适当提示而非施压。
- 提问方式："你觉得……"、"可以聊聊你的想法吗？"
- 关注候选人潜力和学习能力，不过分追求标准答案。`)

	case "stress":
		parts = append(parts, `【面试官风格：压力型】
- 模拟高压面试环境，面试官语气严肃、节奏紧凑。
- 不论候选人回答是否正确，都会质疑："你确定吗？"、"还有其他方案吗？"
- 连续追问，不给喘息空间，测试抗压和临场反应。
- 遇到回答含糊的地方立即要求澄清，不接受模棱两可。
- 偶尔故意提出反驳意见，测试候选人是否能坚持正确观点。`)

	case "deep":
		parts = append(parts, `【面试官风格：技术深挖型】
- 每个问题追问到底层原理（源码级别）。
- 追问链示例：概念 → 实现原理 → 数据结构 → 时间复杂度 → JVM/操作系统层面 → 优化空间。
- 期望候选人能画出内存模型、讲清字节码或汇编层面的行为。
- 不满足于"知道怎么用"，必须"知道为什么这样设计"。`)

	case "practical":
		parts = append(parts, `【面试官风格：项目实战型】
- 围绕候选人简历中的项目经历深入提问。
- 考察"你在项目中具体负责什么？遇到什么技术难题？如何解决的？"
- 追问项目的架构、技术选型、可扩展性、实际效果。
- 关注候选人在团队中的角色和贡献。`)

	case "algorithm":
		parts = append(parts, `【面试官风格：算法考察型】
- 每题都要求候选人写出思路或手撕代码。
- 追问时间/空间复杂度，以及是否有更优解法。
- 考察数据结构选择、边界条件处理、代码鲁棒性。`)

	default:
		parts = append(parts, `【面试官风格：标准型】
- 按照标准面试流程进行，节奏适中。
- 考察全面，深度适中。`)
	}

	// Add company-specific instructions if available
	if company != "" {
		if profile, ok := companyStyleProfiles[company]; ok {
			parts = append(parts, profile)
		}
	}

	return strings.Join(parts, "\n\n")
}

// buildDifficultyPrompt generates difficulty-specific instructions
func buildDifficultyPrompt(difficulty string) string {
	switch difficulty {
	case "campus_intern":
		return `【难度等级：校招实习】
- 主要考察基础知识掌握程度，不要求太深的原理。
- 题目偏向：语言基础、常见数据结构、基本算法、简单项目理解。
- 允许候选人在某些方面有知识空白，重点考察学习能力和逻辑思维。
- 评分标准适度宽松，展现出学习意愿和基本功即可获得及格分。`

	case "campus_graduate":
		return `【难度等级：校招应届】
- 考察扎实的计算机基础和一定的项目经验。
- 题目涵盖：核心框架原理、中等难度算法、基础系统设计、项目深度复盘。
- 要求候选人能够解释"为什么"而非仅仅"是什么"。
- 评分标准正常，需要展现一定的技术深度和独立思考能力。`

	case "social_junior":
		return `【难度等级：社招初级（1-3年经验）】
- 考察实际工作经验和解决问题的能力。
- 题目偏向：生产环境问题排查、性能优化经验、架构设计思路、中高难度算法。
- 要求候选人能结合真实项目案例说明技术选型和权衡。
- 评分标准严格，要求有实战经验支撑回答，不接受纯理论背诵。`

	default:
		return ""
	}
}

// GenerateRandomStyleForInterview picks a random style for "random" interview mode
// The style is not revealed to the user until after the interview
func GenerateRandomStyleForInterview() (style string, company string) {
	styles := []string{"gentle", "stress", "deep", "practical"}
	companies := []string{"", "ali", "bytedance", "tencent", "meituan", ""}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	style = styles[rng.Intn(len(styles))]
	company = companies[rng.Intn(len(companies))]
	return
}
