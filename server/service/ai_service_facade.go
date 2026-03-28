package service

import (
	"context"
	"strings"

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
	aiDomainSvc := aidomain.NewAIService(llmClient)
	aiDomainSvc.SetGroundTruthRetriever(func(ctx context.Context, query string, limit int) ([]string, error) {
		ragSvc := GetRAGService()
		if ragSvc == nil {
			return nil, nil
		}
		chunks, err := ragSvc.SearchKnowledgeChunksWithLimit(query, limit)
		if err != nil {
			return nil, err
		}

		seen := make(map[string]struct{}, len(chunks))
		references := make([]string, 0, len(chunks))
		for _, chunk := range chunks {
			content := strings.TrimSpace(chunk.Content)
			if content == "" {
				continue
			}
			if _, ok := seen[content]; ok {
				continue
			}
			seen[content] = struct{}{}
			references = append(references, content)
		}
		return references, nil
	})

	return &AIService{AIService: aiDomainSvc}
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
