package service

import (
	"context"

	"your-project/pkg/llm"
	aidomain "your-project/service/ai"
)

// AIService is a facade over the AI domain package to keep existing call sites stable.
type AIService struct {
	*aidomain.AIService
}

type EvaluationResult = aidomain.EvaluationResult

type ReportInsights = aidomain.ReportInsights

type ReviewResult = aidomain.ReviewResult

type ReviewDimensions = aidomain.ReviewDimensions

func NewAIService(llmClient llm.LLMClient) *AIService {
	return &AIService{AIService: aidomain.NewAIService(llmClient)}
}

func EvaluateCandidateAnswer(question, expectedAnswer, answer string, llmCallFunc func(prompt string) (string, error)) (*ReviewResult, error) {
	return aidomain.EvaluateCandidateAnswer(question, expectedAnswer, answer, llmCallFunc)
}

func IsInvalidAnswer(answer string) bool {
	return aidomain.IsInvalidAnswer(answer)
}

func GenerateRandomStyleForInterview() (style string, company string) {
	return aidomain.GenerateRandomStyleForInterview()
}

func (s *AIService) ChatWithFormat(ctx context.Context, prompt string, taskType string, responseFormat *llm.ResponseFormat) (string, error) {
	return s.AIService.ChatWithFormat(ctx, prompt, taskType, responseFormat)
}

func (s *AIService) RenderPrompt(templateName string, data interface{}) (string, error) {
	return s.AIService.RenderPrompt(templateName, data)
}
