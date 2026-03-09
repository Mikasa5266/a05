package service

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"
)

// ReviewResult 定义了严谨的评分返回结构
type ReviewResult struct {
	Score      int    `json:"score"`
	Comment    string `json:"comment"`
	Suggestion string `json:"suggestion"`
	// 多维度评分
	Dimensions *ReviewDimensions `json:"dimensions,omitempty"`
	// 亮点（候选人做得好的具体方面）
	Highlights []string `json:"highlights,omitempty"`
	// 差距（对比期望答案缺少的核心点）
	Gaps []string `json:"gaps,omitempty"`
	// 参考答案大纲
	ModelAnswerOutline string `json:"model_answer_outline,omitempty"`
	// 追问方向
	FollowUp string `json:"follow_up,omitempty"`
}

// ReviewDimensions 多维度评分
type ReviewDimensions struct {
	TechnicalDepth int `json:"technical_depth"` // 技术深度 0-100
	Expression     int `json:"expression"`      // 表达清晰度 0-100
	Logic          int `json:"logic"`           // 逻辑严谨性 0-100
	Completeness   int `json:"completeness"`    // 完整度 0-100
}

type answerSignals struct {
	runeLen         int
	keywordHits     int
	keywordTotal    int
	keywordCoverage float64
	technicalHits   int
	hasStructure    bool
	genericFiller   bool
	actionVerbHits  int
}

const (
	dimWeightTech  = 0.40
	dimWeightExpr  = 0.20
	dimWeightLogic = 0.20
	dimWeightComp  = 0.20
)

var keywordTokenPattern = regexp.MustCompile(`[\p{Han}]{2,}|[A-Za-z]{3,}`)

// IsInvalidAnswer 预处理拦截：直接在代码层物理拦截“乱回”和“不会”
// 确保一字不答或废话直接0分，解决 AI 随机给同情分的问题
func IsInvalidAnswer(answer string) bool {
	ans := strings.TrimSpace(answer)
	// 完全没回答
	if len(ans) == 0 {
		return true
	}

	if hasStrongGiveUpIntent(ans) {
		return true
	}

	// 常见放弃与敷衍词汇全集
	giveUpWords := []string{
		"不会", "不知道", "不清楚", "没学过", "不懂", "忘了", "忘记了",
		"啊", "嗯", "哈", "略", "什么", "没听过", "没了解过", "不会答",
	}

	// 清除常见标点后比对
	cleanAns := normalizeAnswerText(ans)
	for _, w := range giveUpWords {
		if cleanAns == w {
			return true
		}
	}

	// 纯乱码或极短无意义回答拦截 (例如用户乱敲 "asd" 或 "123")
	// 正常的技术名词解释就算再短也不会少于3个汉字（不全为英文时）
	if len([]rune(ans)) <= 3 && !isEnglishOnly(ans) {
		return true
	}

	if isPureNoise(ans) {
		return true
	}

	return false
}

func normalizeAnswerText(s string) string {
	replacer := strings.NewReplacer(
		"，", "", "。", "", "！", "", "？", "", "、", "", ",", "", ".", "", "!", "", "?", "", " ", "", "\t", "", "\n", "", "\r", "",
	)
	return strings.ToLower(replacer.Replace(strings.TrimSpace(s)))
}

func hasStrongGiveUpIntent(ans string) bool {
	normalized := normalizeAnswerText(ans)
	if normalized == "" {
		return true
	}

	giveUpPhrases := []string{
		"我不会", "真的不会", "完全不会", "我不知道", "真不知道", "不清楚", "不太清楚", "我不懂", "答不上来", "回答不出来", "我回答不出来", "没法回答", "想不起来",
		"idontknow", "donotknow", "dontknow", "noidea", "cannotanswer", "cantanswer",
	}
	for _, phrase := range giveUpPhrases {
		if strings.Contains(normalized, phrase) {
			return true
		}
	}

	patterns := []*regexp.Regexp{
		regexp.MustCompile(`我.{0,4}(不会|不懂|不知道|不清楚|答不出来|回答不出来|没法回答|想不起来)`),
		regexp.MustCompile(`(不会|不知道|答不出来|回答不出来).{0,6}(这题|这个题|这个问题|这个)`),
		regexp.MustCompile(`(i\s*don'?t\s*know|no\s*idea|can'?t\s*answer)`),
	}
	for _, p := range patterns {
		if p.MatchString(strings.ToLower(ans)) {
			return true
		}
	}

	return false
}

func isPureNoise(ans string) bool {
	trimmed := strings.TrimSpace(ans)
	if trimmed == "" {
		return true
	}
	noiseRe := regexp.MustCompile(`^[0-9a-zA-Z\p{P}\p{S}\s]+$`)
	if noiseRe.MatchString(trimmed) && len([]rune(trimmed)) <= 12 {
		return true
	}
	return false
}

// isEnglishOnly 判断是否全是英文字符（容忍空格）
func isEnglishOnly(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && r != ' ' {
			return false
		}
	}
	return true
}

// BuildStrictEvalPrompt 构建多维度、深度分析的评估 Prompt
func BuildStrictEvalPrompt(question, expectedAnswer, answer string) string {
	expected := strings.TrimSpace(expectedAnswer)
	if expected == "" {
		expected = "（未提供标准答案，请基于题目核心考点自行构建评分锚点）"
	}

	return fmt.Sprintf(`你是一位来自顶级互联网公司（字节跳动/腾讯/阿里级别）的资深技术面试官。
你必须严格按评分规则打分，不允许同情分或保底分。

【面试问题】
"%s"

【标准答案锚点（用于比对覆盖度）】
"%s"

【候选人回答】
"%s"

【评分维度】（每个维度 0-100 分）：
1. 技术深度 (technical_depth)：是否触及底层原理、设计权衡、关键机制
2. 表达清晰度 (expression)：语言组织是否有条理，是否便于面试官理解
3. 逻辑严谨性 (logic)：推理链是否完整，有无自相矛盾或跳跃
4. 完整度 (completeness)：是否覆盖标准答案锚点中的核心点

【综合评分计算公式】
score = technical_depth*0.40 + expression*0.20 + logic*0.20 + completeness*0.20
最终 score 必须与维度分数一致（允许误差不超过 ±5）。

【综合评分标准（0-100）】
- 0分：完全未作答 / 明确放弃 / 答非所问 / 乱码敷衍
- 1-30分：存在严重事实性错误或完全偏离核心
- 31-50分：仅答出皮毛，缺乏深度，有明显知识漏洞
- 51-70分：基本答出主干但深度不足，缺少原理或实践延伸
- 71-85分：回答准确完整，逻辑清晰，有一定深度
- 86-100分：深入底层原理，结合实践案例，展现极强技术功底

【强制红线（必须执行）】
出现以下任一情况，score 必须是 0：
1. 明确放弃作答（例如“我不会”“我不知道怎么答”）。
2. 无意义内容、灌水、口头禅堆砌、明显敷衍（例如“123”“asd”“随便说说”）。
3. 与题目无关且没有技术信息。
4. 仅表达情绪/态度，不给出任何有效技术点。

【输出要求】
返回纯 JSON 对象（不要 markdown 代码块），格式严格如下：
{
  "score": 综合评分(整数),
  "dimensions": {
    "technical_depth": 技术深度分(整数),
    "expression": 表达清晰度分(整数),
    "logic": 逻辑严谨性分(整数),
    "completeness": 完整度分(整数)
  },
  "highlights": ["候选人做得好的第1个具体方面", "做得好的第2个方面"],
  "gaps": ["对比标准答案缺失的第1个核心点", "缺失的第2个核心点"],
  "comment": "2-4句话的整体点评，先肯定亮点再指出不足，语气专业客观",
  "suggestion": "2-3条有针对性的改进建议，用分号分隔",
  "model_answer_outline": "用3-5个要点概括这道题的优秀回答应包含哪些核心内容",
  "follow_up": "基于候选人回答，你会进一步追问什么（1句话）"
}`, question, expected, answer)
}

// EvaluateCandidateAnswer 核心流程：模拟调用大模型并执行三道防线校验
// 参数 llmCallFunc 是你的大模型调用包装函数，负责传入 prompt 返回纯文本字符串
func EvaluateCandidateAnswer(question, expectedAnswer, answer string, llmCallFunc func(prompt string) (string, error)) (*ReviewResult, error) {
	trimmedAnswer := strings.TrimSpace(answer)

	// 【第一道防线】：代码层面拦截废话
	if IsInvalidAnswer(trimmedAnswer) {
		return &ReviewResult{
			Score:      0,
			Comment:    "候选人未作答、表示不会或提供了无意义的敷衍回答。作为面试官，我判定该题得 0 分。",
			Suggestion: "遇到完全不会的问题，可以坦诚表示没有接触过，但切忌乱敲乱答。建议针对此问题核心概念进行系统性学习和补充。",
			Dimensions: &ReviewDimensions{},
		}, nil
	}

	// 【第二道防线】：构建严苛的 Prompt 给 AI
	prompt := BuildStrictEvalPrompt(question, expectedAnswer, trimmedAnswer)

	// 调用底层 LLM (DeepSeek)
	llmResp, err := llmCallFunc(prompt)
	if err != nil {
		return nil, fmt.Errorf("AI评估请求失败: %v", err)
	}

	// 解析 AI 返回的 JSON
	result, err := parseReviewResult(llmResp)
	if err != nil {
		return nil, fmt.Errorf("AI返回格式无法解析: %v, 原始返回: %s", err, llmResp)
	}

	// 【第三道防线】：维度一致性 + 内容信号强校验
	return normalizeReviewResult(question, expectedAnswer, trimmedAnswer, result), nil
}

func parseReviewResult(raw string) (*ReviewResult, error) {
	text := strings.TrimSpace(raw)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	if extracted, ok := extractFirstJSONObject(text); ok {
		text = extracted
	}

	var result ReviewResult
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func extractFirstJSONObject(text string) (string, bool) {
	runes := []rune(text)
	start := -1
	depth := 0
	inString := false
	escaped := false

	for i, r := range runes {
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if r == '\\' {
				escaped = true
				continue
			}
			if r == '"' {
				inString = false
			}
			continue
		}

		if r == '"' {
			inString = true
			continue
		}
		if r == '{' {
			if depth == 0 {
				start = i
			}
			depth++
			continue
		}
		if r == '}' {
			if depth == 0 {
				continue
			}
			depth--
			if depth == 0 && start >= 0 {
				return strings.TrimSpace(string(runes[start : i+1])), true
			}
		}
	}
	return "", false
}

func normalizeReviewResult(question, expectedAnswer, answer string, result *ReviewResult) *ReviewResult {
	if result == nil {
		result = &ReviewResult{}
	}
	if hasStrongGiveUpIntent(answer) || IsInvalidAnswer(answer) {
		result.Score = 0
		result.Comment = "候选人明确表示无法作答或回答无效，按严格评分规则判定该题 0 分。"
		result.Suggestion = "建议先回答该题的核心定义，再补充原理、案例和边界条件。"
		result.Dimensions = &ReviewDimensions{}
		return result
	}

	if result.Dimensions == nil {
		result.Dimensions = estimateDimensionsFromScore(result.Score)
	}
	result.Score = clampReviewScore(result.Score)
	result.Dimensions.TechnicalDepth = clampReviewScore(result.Dimensions.TechnicalDepth)
	result.Dimensions.Expression = clampReviewScore(result.Dimensions.Expression)
	result.Dimensions.Logic = clampReviewScore(result.Dimensions.Logic)
	result.Dimensions.Completeness = clampReviewScore(result.Dimensions.Completeness)

	signals := collectAnswerSignals(question, expectedAnswer, answer)
	if hardFail, reason := isSeverelyInsufficientAnswer(question, expectedAnswer, answer, signals); hardFail {
		result.Score = 0
		result.Dimensions = &ReviewDimensions{}
		result.Comment = reason
		result.Suggestion = "请先覆盖题目关键条件，再说明实现步骤、数据结构/复杂度和去重策略。"
		result.Highlights = nil
		result.Gaps = []string{"回答未形成有效解题思路", "缺少关键实现步骤与技术信息"}
		result.ModelAnswerOutline = defaultModelAnswerOutline(expectedAnswer)
		result.FollowUp = defaultFollowUpQuestion(question)
		return result
	}

	dimScore := weightedDimensionScore(result.Dimensions)
	if math.Abs(float64(result.Score-dimScore)) > 15 {
		result.Score = int(math.Round((float64(result.Score) + float64(dimScore)) / 2))
	}
	result.Score = applyStrictCaps(result.Score, signals)
	result.Score = clampReviewScore(result.Score)

	if result.Score == 0 {
		result.Dimensions = &ReviewDimensions{}
	} else {
		alignDimensionsWithScore(result.Dimensions, result.Score)
	}

	if strings.TrimSpace(result.Comment) == "" {
		result.Comment = defaultCommentByScore(result.Score, signals)
	}
	if strings.TrimSpace(result.Suggestion) == "" {
		result.Suggestion = defaultSuggestionBySignals(signals, result.Score)
	}
	if len(result.Highlights) == 0 && result.Score >= 60 {
		result.Highlights = []string{"回答具备一定结构，能够围绕问题展开。"}
	}
	if len(result.Gaps) == 0 {
		result.Gaps = defaultGapsBySignals(signals)
	}
	if strings.TrimSpace(result.ModelAnswerOutline) == "" {
		result.ModelAnswerOutline = defaultModelAnswerOutline(expectedAnswer)
	}
	if strings.TrimSpace(result.FollowUp) == "" {
		result.FollowUp = defaultFollowUpQuestion(question)
	}

	return result
}

func collectAnswerSignals(question, expectedAnswer, answer string) answerSignals {
	signals := answerSignals{
		runeLen: len([]rune(strings.TrimSpace(answer))),
	}

	keywords := buildScoringKeywords(question, expectedAnswer)
	signals.keywordTotal = len(keywords)
	if signals.keywordTotal > 0 {
		lowerAnswer := strings.ToLower(answer)
		for _, kw := range keywords {
			if strings.Contains(lowerAnswer, kw) {
				signals.keywordHits++
			}
		}
		signals.keywordCoverage = float64(signals.keywordHits) / float64(signals.keywordTotal)
	}

	lowerAnswer := strings.ToLower(answer)
	technicalTerms := []string{
		"复杂度", "时间复杂度", "空间复杂度", "并发", "锁", "事务", "索引", "缓存", "分布式",
		"一致性", "幂等", "吞吐", "延迟", "可用性", "隔离级别", "回滚", "架构", "高可用", "限流", "降级",
		"http", "tcp", "sql", "redis", "mysql", "jvm", "gc", "线程", "消息队列", "mq", "微服务",
		"algorithm", "complexity", "latency", "throughput", "consistency",
	}
	for _, term := range technicalTerms {
		if strings.Contains(lowerAnswer, strings.ToLower(term)) {
			signals.technicalHits++
		}
	}

	structureMarkers := []string{"首先", "其次", "然后", "最后", "总结", "第一", "第二", "第三", "first", "second", "finally"}
	for _, marker := range structureMarkers {
		if strings.Contains(lowerAnswer, strings.ToLower(marker)) {
			signals.hasStructure = true
			break
		}
	}

	actionVerbs := []string{
		"先", "再", "然后", "最后", "遍历", "过滤", "去重", "转换", "分组", "统计", "排序", "比较", "判断", "返回",
		"实现", "处理", "优化", "设计", "选择", "使用", "stream", "filter", "map", "collect", "distinct", "set",
		"hashset", "hashmap", "linkedhashset", "for", "while",
	}
	for _, verb := range actionVerbs {
		if strings.Contains(lowerAnswer, strings.ToLower(verb)) {
			signals.actionVerbHits++
		}
	}

	fillerPhrases := []string{"我觉得", "大概", "差不多", "就是", "看情况", "可能", "不好说", "随便说说"}
	for _, filler := range fillerPhrases {
		if strings.Contains(answer, filler) {
			signals.genericFiller = true
			break
		}
	}
	if signals.technicalHits == 0 && signals.keywordCoverage < 0.08 && signals.runeLen < 40 {
		signals.genericFiller = true
	}

	return signals
}

func buildScoringKeywords(question, expectedAnswer string) []string {
	source := strings.TrimSpace(expectedAnswer)
	if source == "" {
		source = strings.TrimSpace(question)
	}
	matches := keywordTokenPattern.FindAllString(strings.ToLower(source), -1)
	seen := make(map[string]struct{}, len(matches))
	keywords := make([]string, 0, len(matches))

	for _, token := range matches {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		if len([]rune(token)) < 2 {
			continue
		}
		if _, exists := seen[token]; exists {
			continue
		}
		seen[token] = struct{}{}
		keywords = append(keywords, token)
		if len(keywords) >= 24 {
			break
		}
	}
	return keywords
}

func weightedDimensionScore(d *ReviewDimensions) int {
	if d == nil {
		return 0
	}
	score := float64(d.TechnicalDepth)*dimWeightTech +
		float64(d.Expression)*dimWeightExpr +
		float64(d.Logic)*dimWeightLogic +
		float64(d.Completeness)*dimWeightComp
	return int(math.Round(score))
}

func applyStrictCaps(score int, signals answerSignals) int {
	capped := score

	if signals.runeLen < 8 && capped > 15 {
		capped = 15
	}
	if signals.runeLen < 12 && capped > 20 {
		capped = 20
	}
	if signals.runeLen < 20 && capped > 30 {
		capped = 30
	}
	if signals.keywordTotal >= 6 {
		if signals.keywordCoverage < 0.08 && capped > 35 {
			capped = 35
		} else if signals.keywordCoverage < 0.15 && capped > 50 {
			capped = 50
		}
	}
	if signals.technicalHits == 0 && signals.runeLen < 80 && capped > 45 {
		capped = 45
	}
	if signals.genericFiller && capped > 40 {
		capped = 40
	}
	return clampReviewScore(capped)
}

func questionRequiresReasoningAnswer(question, expectedAnswer string) bool {
	text := strings.ToLower(strings.TrimSpace(question + " " + expectedAnswer))
	if text == "" {
		return false
	}
	indicators := []string{
		"如何", "怎么", "思路", "实现", "设计", "优化", "分析", "步骤", "算法", "代码", "去重", "复杂度",
		"请结合", "架构", "tradeoff", "implement", "design", "optimize", "approach", "complexity",
	}
	for _, item := range indicators {
		if strings.Contains(text, item) {
			return true
		}
	}
	return false
}

func answerContainsActionOrReasoning(answer string) bool {
	text := strings.ToLower(strings.TrimSpace(answer))
	if text == "" {
		return false
	}
	patterns := []string{
		"先", "再", "然后", "最后", "因为", "所以", "因此", "通过", "使用", "遍历", "过滤", "去重",
		"实现", "处理", "优化", "判断", "返回", "stream", "filter", "distinct", "set", "map", "for", "while",
	}
	for _, p := range patterns {
		if strings.Contains(text, p) {
			return true
		}
	}
	return false
}

func isSeverelyInsufficientAnswer(question, expectedAnswer, answer string, signals answerSignals) (bool, string) {
	trimmed := strings.TrimSpace(answer)
	if trimmed == "" {
		return true, "未检测到有效回答内容，按规则判定 0 分。"
	}

	// 极短 + 无技术信号：直接判为无效作答
	if signals.runeLen <= 6 && signals.technicalHits == 0 && signals.actionVerbHits == 0 {
		return true, "回答过短且缺少技术信息，无法形成有效解题思路，按规则判定 0 分。"
	}

	// 开放技术题必须体现“步骤/机制/实现”之一
	if questionRequiresReasoningAnswer(question, expectedAnswer) {
		if signals.runeLen < 14 && signals.actionVerbHits == 0 {
			return true, "该题要求说明实现思路，但回答未体现任何步骤或机制，按规则判定 0 分。"
		}
		if signals.keywordTotal >= 6 && signals.keywordCoverage < 0.10 && signals.technicalHits == 0 {
			return true, "回答与题目关键条件覆盖严重不足，且没有技术实现信息，按规则判定 0 分。"
		}
		if !answerContainsActionOrReasoning(trimmed) && signals.runeLen < 24 {
			return true, "回答未体现可执行的解题动作或推理链，按规则判定 0 分。"
		}
	}

	// 泛化灌水短答：判为无效
	if signals.genericFiller && signals.runeLen < 24 && signals.technicalHits == 0 {
		return true, "回答偏泛化且缺少技术细节，按规则判定 0 分。"
	}

	return false, ""
}

func alignDimensionsWithScore(d *ReviewDimensions, score int) {
	if d == nil {
		return
	}
	minAllowed := clampReviewScore(score - 35)
	maxAllowed := clampReviewScore(score + 35)
	d.TechnicalDepth = clampBetween(d.TechnicalDepth, minAllowed, maxAllowed)
	d.Expression = clampBetween(d.Expression, minAllowed, maxAllowed)
	d.Logic = clampBetween(d.Logic, minAllowed, maxAllowed)
	d.Completeness = clampBetween(d.Completeness, minAllowed, maxAllowed)
}

func clampBetween(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func clampReviewScore(score int) int {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

func estimateDimensionsFromScore(score int) *ReviewDimensions {
	base := clampReviewScore(score)
	return &ReviewDimensions{
		TechnicalDepth: clampReviewScore(base - 8),
		Expression:     clampReviewScore(base + 3),
		Logic:          clampReviewScore(base),
		Completeness:   clampReviewScore(base - 5),
	}
}

func defaultCommentByScore(score int, signals answerSignals) string {
	switch {
	case score >= 80:
		return "回答覆盖较完整，逻辑清晰，核心技术点表达到位，整体表现较强。"
	case score >= 60:
		return "回答基本围绕题目展开，但在关键机制和细节深度上还有提升空间。"
	case score > 0:
		return "回答有一定相关性，但覆盖度和技术深度不足，难以支撑更高评分。"
	default:
		return "回答未形成有效技术论证，按规则判定为低分或零分。"
	}
}

func defaultSuggestionBySignals(signals answerSignals, score int) string {
	suggestions := make([]string, 0, 3)
	if signals.keywordCoverage < 0.15 {
		suggestions = append(suggestions, "先覆盖题目核心关键词，再补充关键机制")
	}
	if signals.technicalHits < 2 {
		suggestions = append(suggestions, "补充底层原理、复杂度或工程权衡，避免空泛描述")
	}
	if !signals.hasStructure {
		suggestions = append(suggestions, "使用“结论-原理-案例-边界”结构组织回答")
	}
	if len(suggestions) == 0 {
		if score >= 80 {
			return "可以进一步加入量化结果与异常场景处理，提升说服力。"
		}
		return "建议先给结论，再补充原理、项目案例和边界条件。"
	}
	return strings.Join(suggestions, "；")
}

func defaultGapsBySignals(signals answerSignals) []string {
	gaps := make([]string, 0, 3)
	if signals.keywordCoverage < 0.15 {
		gaps = append(gaps, "对题目核心考点覆盖不足")
	}
	if signals.technicalHits < 2 {
		gaps = append(gaps, "缺少关键技术机制或工程权衡分析")
	}
	if !signals.hasStructure {
		gaps = append(gaps, "回答结构不够清晰，论证链条不完整")
	}
	if len(gaps) == 0 {
		gaps = append(gaps, "可继续补充更细粒度的实现细节")
	}
	return gaps
}

func defaultModelAnswerOutline(expectedAnswer string) string {
	trimmed := strings.TrimSpace(expectedAnswer)
	if trimmed != "" {
		runes := []rune(trimmed)
		if len(runes) > 180 {
			trimmed = string(runes[:180]) + "..."
		}
		return trimmed
	}
	return "建议按“定义与目标 -> 核心原理 -> 实现步骤 -> 方案权衡 -> 边界与风险”五步展开。"
}

func defaultFollowUpQuestion(question string) string {
	trimmed := strings.TrimSpace(question)
	if trimmed == "" {
		return "如果业务量提升 10 倍，你会如何调整当前方案并说明取舍？"
	}
	runes := []rune(trimmed)
	if len(runes) > 20 {
		trimmed = string(runes[:20])
	}
	return fmt.Sprintf("围绕“%s”，请补充一个实际项目中的落地案例和关键权衡。", trimmed)
}
