package ai

import "your-project/pkg/llm"

// AIFacade is the only AI capability entry for upper service layer.
// External services should depend on this interface instead of concrete AIService.
type AIFacade interface {
	AICoreService
	AIChatService
	AIQuestionService
	AIEvaluationService
	AIReportService
	AISpeechService
	AITextService
	AICoachService
	AIUtilityService
}

// DefaultAIFacade only orchestrates subservice wiring and dispatching.
// It intentionally contains no business logic.
type DefaultAIFacade struct {
	AICoreService
	AIChatService
	AIQuestionService
	AIEvaluationService
	AIReportService
	AISpeechService
	AITextService
	AICoachService
	AIUtilityService
}

var _ AIFacade = (*DefaultAIFacade)(nil)

func NewAIFacade(llmClient llm.LLMClient, retriever GroundTruthRetriever, promptManager ...DynamicPromptManager) AIFacade {
	svc := NewAIService(llmClient, promptManager...)
	if retriever != nil {
		svc.SetGroundTruthRetriever(retriever)
	}

	questionAdapter := &aiQuestionServiceAdapter{service: svc}
	coachAdapter := &aiCoachServiceAdapter{service: svc}

	return &DefaultAIFacade{
		AICoreService:       svc,
		AIChatService:       svc,
		AIQuestionService:   questionAdapter,
		AIEvaluationService: svc,
		AIReportService:     svc,
		AISpeechService:     svc,
		AITextService:       svc,
		AICoachService:      coachAdapter,
		AIUtilityService:    defaultAIUtilityService{},
	}
}
