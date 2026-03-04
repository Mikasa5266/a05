package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"your-project/model"
	"your-project/repository"
)

type AIChatResponse struct {
	Message string `json:"message"`
	Type    string `json:"type"` // "answer" or "question"
}

func AIChat(userID uint, message, context string) (*AIChatResponse, error) {
	aiService := NewAIService()

	// 构建对话提示词
	prompt := buildChatPrompt(userID, message, context, nil, nil)

	response, err := aiService.callLLM(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to call AI: %w", err)
	}
	response = aiService.EnsureChineseOutput(response, "我已收到你的问题。请你补充更具体的技术背景和目标，我会给出更有针对性的中文建议。")

	return &AIChatResponse{
		Message: response,
		Type:    "answer",
	}, nil
}

func AIChatWithInterviewContext(userID uint, interviewID uint, message string) (*AIChatResponse, error) {
	// 获取面试信息
	interview, err := GetInterviewByID(userID, interviewID)
	if err != nil {
		return nil, fmt.Errorf("failed to get interview: %w", err)
	}

	// 获取已回答的问题
	repo := repository.NewInterviewRepository()
	answers, err := repo.GetAnswersByInterviewID(interviewID)
	if err != nil {
		return nil, fmt.Errorf("failed to get answers: %w", err)
	}

	aiService := NewAIService()

	// 构建包含面试上下文的提示词
	prompt := buildChatPrompt(userID, message, "", interview, answers)

	response, err := aiService.callLLM(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to call AI: %w", err)
	}
	response = aiService.EnsureChineseOutput(response, "我已结合当前面试上下文进行分析。建议你先按结论、原理、实践案例的结构来组织回答。")

	return &AIChatResponse{
		Message: response,
		Type:    "answer",
	}, nil
}

func buildChatPrompt(userID uint, message, context string, interview *model.Interview, answers []model.AnswerResult) string {
	var prompt strings.Builder

	prompt.WriteString("你是一个专业的AI面试助手，请根据以下信息进行对话：\n\n")

	// 添加用户信息
	userRepo := repository.NewUserRepository()
	user, _ := userRepo.GetByID(userID)
	if user != nil {
		prompt.WriteString(fmt.Sprintf("用户：%s (ID: %d)\n", user.Username, user.ID))
	}

	// 添加上下文信息
	if context != "" {
		prompt.WriteString(fmt.Sprintf("对话上下文：%s\n", context))
	}

	// 添加面试信息
	if interview != nil {
		prompt.WriteString(fmt.Sprintf("\n面试信息：\n"))
		prompt.WriteString(fmt.Sprintf("- 职位：%s\n", interview.Position))
		prompt.WriteString(fmt.Sprintf("- 难度：%s\n", interview.Difficulty))
		prompt.WriteString(fmt.Sprintf("- 状态：%s\n", interview.Status))

		// 添加已回答问题
		if len(answers) > 0 {
			prompt.WriteString(fmt.Sprintf("\n已回答问题：\n"))
			for i, answer := range answers {
				prompt.WriteString(fmt.Sprintf("%d. %s (得分：%d)\n", i+1, answer.Question.Title, answer.Score))
			}
		}
	}

	prompt.WriteString(fmt.Sprintf("\n用户消息：%s\n\n", message))
	prompt.WriteString("请根据以上信息，给出专业、有帮助的回答。")

	return prompt.String()
}

// 检查是否需要将用户消息转换为面试问题
func CheckIfShouldGenerateQuestion(message string) bool {
	// 简单的关键词检测逻辑
	questionKeywords := []string{
		"问题", "题目", "面试", "技术", "编程", "代码",
		"如何", "什么", "为什么", "怎么", "区别", "原理",
		"解释", "说明", "实现", "优化", "设计", "架构",
	}

	for _, keyword := range questionKeywords {
		if strings.Contains(strings.ToLower(message), strings.ToLower(keyword)) {
			return true
		}
	}

	return false
}

// 将用户消息转换为面试问题
func GenerateQuestionFromMessage(position, difficulty, message string) (*model.Question, error) {
	aiService := NewAIService()

	prompt := fmt.Sprintf(`
基于用户的输入，生成一个合适的面试问题：

用户输入：%s
职位：%s
难度：%s

请生成一个与技术面试相关的问题，要求：
1. 与用户输入内容相关
2. 符合职位和难度要求
3. 专业且有深度

返回格式：{"title": "问题标题", "content": "问题内容", "expected_answer": "期望答案要点"}
`, message, position, difficulty)

	response, err := aiService.callLLM(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate question: %w", err)
	}

	// 解析AI返回的问题
	var result struct {
		Title          string `json:"title"`
		Content        string `json:"content"`
		ExpectedAnswer string `json:"expected_answer"`
	}

	cleanResponse := extractJSONContent(response)

	// 尝试解析JSON，如果失败则使用默认值
	if err := json.Unmarshal([]byte(cleanResponse), &result); err == nil && result.Title != "" {
		// 成功解析，使用AI生成的问题
		question := &model.Question{
			Title:          result.Title,
			Content:        result.Content,
			Position:       position,
			Difficulty:     difficulty,
			ExpectedAnswer: result.ExpectedAnswer,
		}
		aiService.EnsureQuestionChinese(question)
		return question, nil
	}

	// 解析失败或格式不正确，使用默认问题格式
	question := &model.Question{
		Title:          fmt.Sprintf("基于用户输入的问题: %s", message),
		Content:        aiService.EnsureChineseOutput(response, "请结合岗位要求，详细说明你的技术方案、关键实现与优化思路。"),
		Position:       position,
		Difficulty:     difficulty,
		ExpectedAnswer: "请根据具体问题提供详细回答",
	}
	aiService.EnsureQuestionChinese(question)
	return question, nil
}
