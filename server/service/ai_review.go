package service

import (
	"encoding/json"
	"fmt"
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

// IsInvalidAnswer 预处理拦截：直接在代码层物理拦截“乱回”和“不会”
// 确保一字不答或废话直接0分，解决 AI 随机给同情分的问题
func IsInvalidAnswer(answer string) bool {
	ans := strings.TrimSpace(answer)
	// 完全没回答
	if len(ans) == 0 {
		return true
	}

	// 常见放弃与敷衍词汇全集
	giveUpWords := []string{
		"不会", "不知道", "不清楚", "没学过", "不懂", "忘了", "忘记了",
		"啊", "嗯", "哈", "略", "什么", "没听过", "没了解过", "不会答",
	}

	// 清除常见标点后比对
	cleanAns := strings.ReplaceAll(strings.ReplaceAll(ans, "，", ""), "。", "")
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
func BuildStrictEvalPrompt(question, answer string) string {
	return fmt.Sprintf(`你是一位来自顶级互联网公司（字节跳动/腾讯/阿里级别）的资深技术面试官。
请对候选人的回答进行多维度、深度的专业评估。

【面试问题】
"%s"

【候选人回答】
"%s"

【评分维度】（每个维度 0-100 分）：
1. 技术深度 (technical_depth)：是否触及底层原理、源码级理解、设计权衡
2. 表达清晰度 (expression)：语言组织是否有条理，是否便于面试官理解
3. 逻辑严谨性 (logic)：推理链是否完整，有无自相矛盾或跳跃
4. 完整度 (completeness)：是否覆盖了核心考点，有无遗漏关键面

【综合评分标准】（0-100）：
- 0分：完全未作答 / 答非所问 / 乱码敷衍
- 1-30分：存在严重事实性错误或完全偏离核心
- 31-50分：仅答出皮毛，缺乏深度，有明显知识漏洞
- 51-70分：基本答出主干但深度不足，缺少原理或实践延伸
- 71-85分：回答准确完整，逻辑清晰，有一定深度
- 86-100分：深入底层原理，结合实践案例，展现极强技术功底

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
}`, question, answer)
}

// EvaluateCandidateAnswer 核心流程：模拟调用大模型并执行三道防线校验
// 参数 llmCallFunc 是你的大模型调用包装函数，负责传入 prompt 返回纯文本字符串
func EvaluateCandidateAnswer(question, answer string, llmCallFunc func(prompt string) (string, error)) (*ReviewResult, error) {
	// 【第一道防线】：代码层面拦截废话
	if IsInvalidAnswer(answer) {
		return &ReviewResult{
			Score:      0,
			Comment:    "候选人未作答、表示不会或提供了无意义的敷衍回答。作为面试官，我判定该题得 0 分。",
			Suggestion: "遇到完全不会的问题，可以坦诚表示没有接触过，但切忌乱敲乱答。建议针对此问题核心概念进行系统性学习和补充。",
		}, nil
	}

	// 【第二道防线】：构建严苛的 Prompt 给 AI
	prompt := BuildStrictEvalPrompt(question, answer)

	// 调用底层 LLM (DeepSeek)
	llmResp, err := llmCallFunc(prompt)
	if err != nil {
		return nil, fmt.Errorf("AI评估请求失败: %v", err)
	}

	// 容错处理：清理大模型可能返回的 Markdown 代码块残留
	llmResp = strings.TrimSpace(llmResp)
	llmResp = strings.TrimPrefix(llmResp, "```json")
	llmResp = strings.TrimPrefix(llmResp, "```")
	llmResp = strings.TrimSuffix(llmResp, "```")
	llmResp = strings.TrimSpace(llmResp)

	// 解析 AI 返回的 JSON
	var result ReviewResult
	if err := json.Unmarshal([]byte(llmResp), &result); err != nil {
		return nil, fmt.Errorf("AI返回格式无法解析: %v, 原始返回: %s", err, llmResp)
	}

	// 【第三道防线】：AI 幻觉兜底逻辑
	// 如果回答长度极其简短（不足10个字符），但 AI 却给出了超过 30 分的成绩，必定是 AI 在自行脑补发挥
	if result.Score > 30 && len([]rune(strings.TrimSpace(answer))) < 10 {
		result.Score = 0
		result.Comment = "系统检测到异常评分。修正为 0 分：回答内容极度匮乏，不足以构成有效的技术解答。"
	}

	return &result, nil
}
