package ai

import (
	"context"

	"your-project/internal/model"
	"your-project/pkg/llm"
)

type FollowUpQuestionRequest struct {
	Interview       *model.Interview
	CurrentQuestion *model.Question
	Answer          string
	RAGContext      string
	FollowUpContext string
	FollowUpIndex   int
}

type ShadowCoachHintRequest struct {
	Position       string
	Question       string
	Transcript     string
	Style          string
	SilenceSeconds int
}

type ShadowCoachHintLevelsRequest struct {
	Position         string
	Question         string
	Transcript       string
	Style            string
	ReferenceAnswer  string
	KnowledgeContext string
}

type AICoreService interface {
	Chat(ctx context.Context, prompt string) (string, error)
	ChatWithTask(ctx context.Context, prompt string, taskType string) (string, error)
	ChatWithFormat(ctx context.Context, prompt string, taskType string, responseFormat *llm.ResponseFormat) (string, error)
	RenderPrompt(templateName string, data interface{}) (string, error)
}

type AIChatService interface {
	AIChat(ctx context.Context, userID uint, message, convoContext string) (*AIChatResponse, error)
	AIChatWithInterviewContext(ctx context.Context, userID uint, interviewID uint, message string) (*AIChatResponse, error)
	GenerateQuestionFromMessage(ctx context.Context, position, difficulty, message string) (*model.Question, error)
}

type AIQuestionService interface {
	GenerateQuestions(ctx context.Context, interview *model.Interview, count int) ([]*model.Question, error)
	GenerateNextQuestionWithWeights(ctx context.Context, interview *model.Interview, previousAnswers []model.AnswerResult, capabilityGraph *model.JobCapabilityDimension) (*model.Question, error)
	GenerateNextQuestion(ctx context.Context, interview *model.Interview, previousAnswers []model.AnswerResult) (*model.Question, error)
	GenerateFollowUpQuestion(ctx context.Context, req FollowUpQuestionRequest) (*model.Question, string, error)
	GenerateClarifyingFollowUpQuestion(ctx context.Context, currentQ *model.Question, answer string, followUpIndex int) (*model.Question, error)
	IsContextDependentOpeningQuestion(question *model.Question) bool
	NormalizeToSelfContainedOpening(question *model.Question)
}

type AIEvaluationService interface {
	EvaluateAnswer(ctx context.Context, question *model.Question, answer string) (*EvaluationResult, error)
}

type AIReportService interface {
	GenerateOverallAnalysis(ctx context.Context, interview *model.Interview, answers []model.AnswerResult) (string, error)
	GenerateReportInsights(ctx context.Context, interview *model.Interview, answers []model.AnswerResult) (*ReportInsights, error)
}

type AISpeechService interface {
	TranscribeAudio(audioData string) (string, error)
	SynthesizeSpeech(text string) ([]byte, error)
}

type AITextService interface {
	EnsureChineseOutput(ctx context.Context, text, fallback string) string
	EnsureQuestionChinese(ctx context.Context, question *model.Question)
}

type AICoachService interface {
	GenerateShadowCoachHint(ctx context.Context, req ShadowCoachHintRequest) (string, error)
	GenerateShadowCoachHintLevels(ctx context.Context, req ShadowCoachHintLevelsRequest) ([]string, error)
}

type AIUtilityService interface {
	GenerateRandomStyleForInterview() (style string, company string)
	IsInvalidAnswer(answer string) bool
}

type aiQuestionServiceAdapter struct {
	service *AIService
}

func (a *aiQuestionServiceAdapter) GenerateQuestions(ctx context.Context, interview *model.Interview, count int) ([]*model.Question, error) {
	return a.service.GenerateQuestions(ctx, interview, count)
}

func (a *aiQuestionServiceAdapter) GenerateNextQuestionWithWeights(ctx context.Context, interview *model.Interview, previousAnswers []model.AnswerResult, capabilityGraph *model.JobCapabilityDimension) (*model.Question, error) {
	return a.service.GenerateNextQuestionWithWeights(ctx, interview, previousAnswers, capabilityGraph)
}

func (a *aiQuestionServiceAdapter) GenerateNextQuestion(ctx context.Context, interview *model.Interview, previousAnswers []model.AnswerResult) (*model.Question, error) {
	return a.service.GenerateNextQuestion(ctx, interview, previousAnswers)
}

func (a *aiQuestionServiceAdapter) GenerateFollowUpQuestion(ctx context.Context, req FollowUpQuestionRequest) (*model.Question, string, error) {
	return a.service.GenerateFollowUpQuestion(
		ctx,
		req.Interview,
		req.CurrentQuestion,
		req.Answer,
		req.RAGContext,
		req.FollowUpContext,
		req.FollowUpIndex,
	)
}

func (a *aiQuestionServiceAdapter) GenerateClarifyingFollowUpQuestion(ctx context.Context, currentQ *model.Question, answer string, followUpIndex int) (*model.Question, error) {
	return a.service.GenerateClarifyingFollowUpQuestion(ctx, currentQ, answer, followUpIndex)
}

func (a *aiQuestionServiceAdapter) IsContextDependentOpeningQuestion(question *model.Question) bool {
	return a.service.IsContextDependentOpeningQuestion(question)
}

func (a *aiQuestionServiceAdapter) NormalizeToSelfContainedOpening(question *model.Question) {
	a.service.NormalizeToSelfContainedOpening(question)
}

type aiCoachServiceAdapter struct {
	service *AIService
}

func (a *aiCoachServiceAdapter) GenerateShadowCoachHint(ctx context.Context, req ShadowCoachHintRequest) (string, error) {
	return a.service.GenerateShadowCoachHint(ctx, req.Position, req.Question, req.Transcript, req.Style, req.SilenceSeconds)
}

func (a *aiCoachServiceAdapter) GenerateShadowCoachHintLevels(ctx context.Context, req ShadowCoachHintLevelsRequest) ([]string, error) {
	return a.service.GenerateShadowCoachHintLevels(ctx, req.Position, req.Question, req.Transcript, req.Style, req.ReferenceAnswer, req.KnowledgeContext)
}

type defaultAIUtilityService struct{}

func (defaultAIUtilityService) GenerateRandomStyleForInterview() (style string, company string) {
	return GenerateRandomStyleForInterview()
}

func (defaultAIUtilityService) IsInvalidAnswer(answer string) bool {
	return IsInvalidAnswer(answer)
}
