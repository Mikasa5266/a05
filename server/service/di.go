package service

import "your-project/pkg/llm"

var defaultAIService *AIService

// SetAIService registers the shared AIService instance for packages that need LLM access.
func SetAIService(ai *AIService) {
	defaultAIService = ai
}

// MustGetAIService returns the shared AIService or panics if it is not initialized.
func MustGetAIService() *AIService {
	if defaultAIService == nil {
		panic("AIService is not initialized; call SetAIService first")
	}
	return defaultAIService
}

// MustNewAIService initializes AIService with a provided LLM client and sets it as default.
func MustNewAIService(client llm.LLMClient) *AIService {
	ai := NewAIService(client)
	SetAIService(ai)
	return ai
}
