package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"your-project/internal/model"
	"your-project/internal/repository"
	"your-project/pkg/llm"
)

type AIChatResponse struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

func (s *AIService) AIChat(ctx context.Context, userID uint, message, convoContext string) (*AIChatResponse, error) {
	prompt := s.buildChatPrompt(userID, message, convoContext, nil, nil)

	response, err := s.ChatWithTask(ctx, prompt, "chat")
	if err != nil {
		return nil, fmt.Errorf("failed to call AI: %w", err)
	}
	response = s.EnsureChineseOutput(ctx, response, "我已收到你的问题，请补充更具体的技术背景与目标，我会给出更有针对性的建议。")

	return &AIChatResponse{
		Message: response,
		Type:    "answer",
	}, nil
}

func (s *AIService) AIChatWithInterviewContext(ctx context.Context, userID uint, interviewID uint, message string) (*AIChatResponse, error) {
	interviewRepo := repository.NewInterviewRepository()
	interview, err := interviewRepo.GetByID(interviewID)
	if err != nil {
		return nil, fmt.Errorf("failed to get interview: %w", err)
	}
	if interview == nil || interview.UserID != userID {
		return nil, fmt.Errorf("unauthorized access")
	}

	answers, err := interviewRepo.GetAnswersByInterviewID(interviewID)
	if err != nil {
		return nil, fmt.Errorf("failed to get answers: %w", err)
	}

	prompt := s.buildChatPrompt(userID, message, "", interview, answers)

	response, err := s.ChatWithTask(ctx, prompt, "chat")
	if err != nil {
		return nil, fmt.Errorf("failed to call AI: %w", err)
	}
	response = s.EnsureChineseOutput(ctx, response, "我已结合当前面试上下文进行分析。建议你先按结论、原理、实现与案例的结构来组织回答。")

	return &AIChatResponse{
		Message: response,
		Type:    "answer",
	}, nil
}

func (s *AIService) buildChatPrompt(userID uint, message, convoContext string, interview *model.Interview, answers []model.AnswerResult) string {
	var prompt strings.Builder

	systemPrompt := s.getPromptOrDefault("ai/chat/system", "")
	responsePrompt := s.getPromptOrDefault("ai/chat/response_instruction", "")
	if strings.TrimSpace(systemPrompt) != "" {
		prompt.WriteString(systemPrompt + "\n\n")
	}

	userRepo := repository.NewUserRepository()
	user, _ := userRepo.GetByID(userID)
	if user != nil {
		prompt.WriteString(fmt.Sprintf("用户: %s (ID: %d)\n", user.Username, user.ID))
	}

	if convoContext != "" {
		prompt.WriteString(fmt.Sprintf("对话上下文: %s\n", convoContext))
	}

	if interview != nil {
		prompt.WriteString("\n面试信息:\n")
		prompt.WriteString(fmt.Sprintf("- 职位: %s\n", interview.Position))
		prompt.WriteString(fmt.Sprintf("- 难度: %s\n", interview.Difficulty))
		prompt.WriteString(fmt.Sprintf("- 状态: %s\n", interview.Status))

		if len(answers) > 0 {
			prompt.WriteString("\n已回答问题:\n")
			for i, answer := range answers {
				prompt.WriteString(fmt.Sprintf("%d. %s (得分: %d)\n", i+1, answer.Question.Title, answer.Score))
			}
		}
	}

	prompt.WriteString(fmt.Sprintf("\n用户消息: %s\n\n", message))
	if strings.TrimSpace(responsePrompt) != "" {
		prompt.WriteString(responsePrompt)
	}

	return prompt.String()
}

func CheckIfShouldGenerateQuestion(message string) bool {
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

func (s *AIService) GenerateQuestionFromMessage(ctx context.Context, position, difficulty, message string) (*model.Question, error) {
	prompt, err := s.renderDynamicPrompt("ai/chat/generate_question_from_message", map[string]interface{}{
		"Message":    message,
		"Position":   position,
		"Difficulty": difficulty,
	})
	if err != nil {
		prompt = fmt.Sprintf(`
基于用户输入，生成一个合适的技术面试问题：
用户输入：%s
职位：%s
难度：%s

请返回 JSON：{"title":"问题标题","content":"问题内容","expected_answer":"期望回答要点"}
`, message, position, difficulty)
	}

	response, err := s.ChatWithFormat(ctx, prompt, "chat", &llm.ResponseFormat{Type: llm.ResponseFormatJSON})
	if err != nil {
		return nil, fmt.Errorf("failed to generate question: %w", err)
	}

	var result struct {
		Title          string `json:"title"`
		Content        string `json:"content"`
		ExpectedAnswer string `json:"expected_answer"`
	}

	if err := json.Unmarshal([]byte(response), &result); err == nil && result.Title != "" {
		question := &model.Question{
			Title:          result.Title,
			Content:        result.Content,
			Position:       position,
			Difficulty:     difficulty,
			ExpectedAnswer: result.ExpectedAnswer,
		}
		s.EnsureQuestionChinese(ctx, question)
		return question, nil
	}

	question := &model.Question{
		Title:          fmt.Sprintf("基于输入的面试问题: %s", message),
		Content:        s.EnsureChineseOutput(ctx, response, "请结合岗位要求说明你的技术方案、关键实现与优化思路。"),
		Position:       position,
		Difficulty:     difficulty,
		ExpectedAnswer: "请围绕题目给出结构化回答。",
	}
	s.EnsureQuestionChinese(ctx, question)
	return question, nil
}
