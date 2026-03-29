package ai

import (
	"context"

	"your-project/model"
	"your-project/pkg/llm"
)

// AIFacade is the only AI capability entry for upper service layer.
// External services should depend on this interface instead of concrete AIService.
type AIFacade interface {
	// LLM and prompt capabilities
	Chat(ctx context.Context, prompt string) (string, error)
	ChatWithTask(ctx context.Context, prompt string, taskType string) (string, error)
	ChatWithFormat(ctx context.Context, prompt string, taskType string, responseFormat *llm.ResponseFormat) (string, error)
	RenderPrompt(templateName string, data interface{}) (string, error)

	// Conversational and question-generation capabilities
	AIChat(ctx context.Context, userID uint, message, convoContext string) (*AIChatResponse, error)
	AIChatWithInterviewContext(ctx context.Context, userID uint, interviewID uint, message string) (*AIChatResponse, error)
	GenerateQuestionFromMessage(ctx context.Context, position, difficulty, message string) (*model.Question, error)
	GenerateQuestions(ctx context.Context, interview *model.Interview, count int) ([]*model.Question, error)
	GenerateNextQuestionWithWeights(ctx context.Context, interview *model.Interview, previousAnswers []model.AnswerResult, capabilityGraph *model.JobCapabilityDimension) (*model.Question, error)
	GenerateNextQuestion(ctx context.Context, interview *model.Interview, previousAnswers []model.AnswerResult) (*model.Question, error)
	GenerateFollowUpQuestion(ctx context.Context, interview *model.Interview, currentQ *model.Question, answer string, ragContext string, followUpContext string, followUpIndex int) (*model.Question, string, error)
	GenerateClarifyingFollowUpQuestion(ctx context.Context, currentQ *model.Question, answer string, followUpIndex int) (*model.Question, error)

	// Speech and output quality capabilities
	TranscribeAudio(audioData string) (string, error)
	SynthesizeSpeech(text string) ([]byte, error)
	EnsureChineseOutput(ctx context.Context, text, fallback string) string
	EnsureQuestionChinese(ctx context.Context, question *model.Question)
	IsContextDependentOpeningQuestion(question *model.Question) bool
	NormalizeToSelfContainedOpening(question *model.Question)

	// Evaluation and reporting capabilities
	EvaluateAnswer(ctx context.Context, question *model.Question, answer string) (*EvaluationResult, error)
	GenerateOverallAnalysis(ctx context.Context, interview *model.Interview, answers []model.AnswerResult) (string, error)
	GenerateReportInsights(ctx context.Context, interview *model.Interview, answers []model.AnswerResult) (*ReportInsights, error)

	// Coach capabilities
	GenerateShadowCoachHint(ctx context.Context, position, question, transcript, style string, silenceSeconds int) (string, error)
	GenerateShadowCoachHintLevels(ctx context.Context, position, question, transcript, style, referenceAnswer, knowledgeContext string) ([]string, error)

	// Utility capabilities used by upper service layer
	GenerateRandomStyleForInterview() (style string, company string)
	IsInvalidAnswer(answer string) bool
}

type DefaultAIFacade struct {
	service *AIService
}

var _ AIFacade = (*DefaultAIFacade)(nil)

func NewAIFacade(llmClient llm.LLMClient, retriever GroundTruthRetriever) AIFacade {
	svc := NewAIService(llmClient)
	if retriever != nil {
		svc.SetGroundTruthRetriever(retriever)
	}
	return &DefaultAIFacade{service: svc}
}

func (f *DefaultAIFacade) Chat(ctx context.Context, prompt string) (string, error) {
	return f.service.Chat(ctx, prompt)
}

func (f *DefaultAIFacade) ChatWithTask(ctx context.Context, prompt string, taskType string) (string, error) {
	return f.service.ChatWithTask(ctx, prompt, taskType)
}

func (f *DefaultAIFacade) ChatWithFormat(ctx context.Context, prompt string, taskType string, responseFormat *llm.ResponseFormat) (string, error) {
	return f.service.ChatWithFormat(ctx, prompt, taskType, responseFormat)
}

func (f *DefaultAIFacade) RenderPrompt(templateName string, data interface{}) (string, error) {
	return f.service.RenderPrompt(templateName, data)
}

func (f *DefaultAIFacade) AIChat(ctx context.Context, userID uint, message, convoContext string) (*AIChatResponse, error) {
	return f.service.AIChat(ctx, userID, message, convoContext)
}

func (f *DefaultAIFacade) AIChatWithInterviewContext(ctx context.Context, userID uint, interviewID uint, message string) (*AIChatResponse, error) {
	return f.service.AIChatWithInterviewContext(ctx, userID, interviewID, message)
}

func (f *DefaultAIFacade) GenerateQuestionFromMessage(ctx context.Context, position, difficulty, message string) (*model.Question, error) {
	return f.service.GenerateQuestionFromMessage(ctx, position, difficulty, message)
}

func (f *DefaultAIFacade) GenerateQuestions(ctx context.Context, interview *model.Interview, count int) ([]*model.Question, error) {
	return f.service.GenerateQuestions(ctx, interview, count)
}

func (f *DefaultAIFacade) GenerateNextQuestionWithWeights(ctx context.Context, interview *model.Interview, previousAnswers []model.AnswerResult, capabilityGraph *model.JobCapabilityDimension) (*model.Question, error) {
	return f.service.GenerateNextQuestionWithWeights(ctx, interview, previousAnswers, capabilityGraph)
}

func (f *DefaultAIFacade) GenerateNextQuestion(ctx context.Context, interview *model.Interview, previousAnswers []model.AnswerResult) (*model.Question, error) {
	return f.service.GenerateNextQuestion(ctx, interview, previousAnswers)
}

func (f *DefaultAIFacade) GenerateFollowUpQuestion(ctx context.Context, interview *model.Interview, currentQ *model.Question, answer string, ragContext string, followUpContext string, followUpIndex int) (*model.Question, string, error) {
	return f.service.GenerateFollowUpQuestion(ctx, interview, currentQ, answer, ragContext, followUpContext, followUpIndex)
}

func (f *DefaultAIFacade) GenerateClarifyingFollowUpQuestion(ctx context.Context, currentQ *model.Question, answer string, followUpIndex int) (*model.Question, error) {
	return f.service.GenerateClarifyingFollowUpQuestion(ctx, currentQ, answer, followUpIndex)
}

func (f *DefaultAIFacade) TranscribeAudio(audioData string) (string, error) {
	return f.service.TranscribeAudio(audioData)
}

func (f *DefaultAIFacade) SynthesizeSpeech(text string) ([]byte, error) {
	return f.service.SynthesizeSpeech(text)
}

func (f *DefaultAIFacade) EnsureChineseOutput(ctx context.Context, text, fallback string) string {
	return f.service.EnsureChineseOutput(ctx, text, fallback)
}

func (f *DefaultAIFacade) EnsureQuestionChinese(ctx context.Context, question *model.Question) {
	f.service.EnsureQuestionChinese(ctx, question)
}

func (f *DefaultAIFacade) IsContextDependentOpeningQuestion(question *model.Question) bool {
	return f.service.IsContextDependentOpeningQuestion(question)
}

func (f *DefaultAIFacade) NormalizeToSelfContainedOpening(question *model.Question) {
	f.service.NormalizeToSelfContainedOpening(question)
}

func (f *DefaultAIFacade) EvaluateAnswer(ctx context.Context, question *model.Question, answer string) (*EvaluationResult, error) {
	return f.service.EvaluateAnswer(ctx, question, answer)
}

func (f *DefaultAIFacade) GenerateOverallAnalysis(ctx context.Context, interview *model.Interview, answers []model.AnswerResult) (string, error) {
	return f.service.GenerateOverallAnalysis(ctx, interview, answers)
}

func (f *DefaultAIFacade) GenerateReportInsights(ctx context.Context, interview *model.Interview, answers []model.AnswerResult) (*ReportInsights, error) {
	return f.service.GenerateReportInsights(ctx, interview, answers)
}

func (f *DefaultAIFacade) GenerateShadowCoachHint(ctx context.Context, position, question, transcript, style string, silenceSeconds int) (string, error) {
	return f.service.GenerateShadowCoachHint(ctx, position, question, transcript, style, silenceSeconds)
}

func (f *DefaultAIFacade) GenerateShadowCoachHintLevels(ctx context.Context, position, question, transcript, style, referenceAnswer, knowledgeContext string) ([]string, error) {
	return f.service.GenerateShadowCoachHintLevels(ctx, position, question, transcript, style, referenceAnswer, knowledgeContext)
}

func (f *DefaultAIFacade) GenerateRandomStyleForInterview() (style string, company string) {
	return GenerateRandomStyleForInterview()
}

func (f *DefaultAIFacade) IsInvalidAnswer(answer string) bool {
	return IsInvalidAnswer(answer)
}
