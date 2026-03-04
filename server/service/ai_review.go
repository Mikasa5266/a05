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

// BuildStrictEvalPrompt 构建具有强烈约束力的 Prompt
func BuildStrictEvalPrompt(question, answer string) string {
	return fmt.Sprintf(`作为一名字节跳动/腾讯级别的顶级资深技术面试官，你需要对候选人的回答进行专业且极其严苛的评估。
你的评估关乎公司的技术标准，绝不允许对候选人有任何无意义的仁慈或过度宽容！

【当前面试问题】
"%s"

【候选人回答】
"%s"

【极其严格的打分标准】(满分100分，0分代表毫无价值，请严格对号入座)：
- [0分 - 致命/未作答]：候选人完全答非所问、闲聊、乱码、扯皮。只要没有任何实质性的正确技术内容，**必须强行打0分！绝不可给同情分！**
- [1-20分 - 极差]：提到了极个别沾边的词汇，但逻辑完全是错的，对基本概念的理解南辕北辙。
- [21-40分 - 不及格]：答对了一小部分，但存在严重的常识性事实错误，完全无法胜任实际工作。
- [41-60分 - 勉强及格]：基本答出主干内容，但完全没有深入，或者表述很不专业，勉强能算他知道。
- [61-80分 - 良好]：回答准确，逻辑清晰，覆盖了主要考点，但缺乏底层原理或实际场景的延伸。
- [81-100分 - 优秀]：完美切中要害，极其严谨，不仅包含概念，还有底层原理、优缺点分析及最佳实践，展现极强技术功底。

【输出格式要求】
你必须且只能返回一个纯合法的 JSON 对象，不要包含任何 markdown 标记（例如绝不要包裹 ` + "`" + "```json" + "`" + ` ），不要输出任何其他文本！
JSON 格式严格如下：
{
    "score": 0,
    "comment": "你的专业点评。一针见血地指出错误或缺失点，语气要严厉客观。如果是乱回，请直接严厉指出'回答与问题无关'。",
    "suggestion": "给出正确的解答方向或学习建议。"
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