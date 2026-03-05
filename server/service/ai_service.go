package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"your-project/config"
	"your-project/model"
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
	prompt := fmt.Sprintf(`
		请为以下面试场景生成 %d 个面试问题：
		
		职位：%s
		难度级别：%s
		面试模式：%s
		面试风格：%s
		
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
	`, count, interview.Position, interview.Difficulty, interview.Mode, interview.Style)

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

func (s *AIService) GenerateNextQuestion(interview *model.Interview, previousAnswers []model.AnswerResult) (*model.Question, error) {
	prompt := fmt.Sprintf(`
		基于以下面试信息，生成下一个合适的面试问题：
		
		职位：%s
		难度级别：%s
		面试模式：%s
		面试风格：%s
		已回答问题数量：%d
		
		请生成一个合适的后续问题，以深入了解候选人的技术能力。
		只使用简体中文。
		返回格式：{"title": "问题标题", "content": "具体问题内容", "expected_answer": "简要的期望回答"}
	`, interview.Position, interview.Difficulty, interview.Mode, interview.Style, len(previousAnswers))

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

func (s *AIService) TranscribeAudio(audioData string) (string, error) {
	decodedAudio, err := base64.StdEncoding.DecodeString(audioData)
	if err != nil {
		return "", fmt.Errorf("failed to decode audio data: %w", err)
	}

	asrConfig := config.GetConfig().ASR

	if asrConfig.Provider == "whisper" {
		return s.transcribeWithWhisper(decodedAudio)
	}

	return "", fmt.Errorf("unsupported ASR provider: %s", asrConfig.Provider)
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
	return "音频转录功能待实现", nil
}
