package service

import (
	"context"
	"strings"

	"your-project/pkg/llm"
	aidomain "your-project/service/ai"
)

var defaultAIService aidomain.AIFacade

// SetAIService registers the shared AIFacade instance for packages that need AI capabilities.
func SetAIService(ai aidomain.AIFacade) {
	defaultAIService = ai
}

// MustGetAIService returns the shared AIFacade or panics if it is not initialized.
func MustGetAIService() aidomain.AIFacade {
	if defaultAIService == nil {
		panic("AIFacade is not initialized; call SetAIService first")
	}
	return defaultAIService
}

// MustNewAIService initializes AIFacade with a provided LLM client and sets it as default.
func MustNewAIService(client llm.LLMClient) aidomain.AIFacade {
	facade := aidomain.NewAIFacade(client, func(ctx context.Context, query string, limit int) ([]string, error) {
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

	SetAIService(facade)
	return facade
}
