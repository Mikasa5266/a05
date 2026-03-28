package ai

import (
	"context"
	"fmt"
	"log"
	"strings"

	"your-project/config"
	"your-project/pkg/llm"
	promptpkg "your-project/pkg/prompt"
)

type AIService struct {
	config               *config.LLMConfig
	llmClient            llm.LLMClient
	promptManager        *promptpkg.PromptManager
	groundTruthRetriever GroundTruthRetriever
}

// GroundTruthRetriever fetches reference snippets for evaluation grounding.
type GroundTruthRetriever func(ctx context.Context, query string, limit int) ([]string, error)

func NewAIService(llmClient llm.LLMClient) *AIService {
	pm, err := promptpkg.NewPromptManager()
	if err != nil {
		log.Printf("failed to initialize prompt manager: %v", err)
	}

	return &AIService{
		config:        &config.GetConfig().LLM,
		llmClient:     llmClient,
		promptManager: pm,
	}
}

func (s *AIService) SetGroundTruthRetriever(retriever GroundTruthRetriever) {
	s.groundTruthRetriever = retriever
}

type EvaluationResult struct {
	Score    int    `json:"score"`
	Feedback string `json:"feedback"`
}

type ReportQADetail struct {
	Question        string   `json:"question"`
	UserAnswer      string   `json:"user_answer"`
	OptimizedAnswer string   `json:"optimized_answer"`
	KeyImprovements []string `json:"key_improvements"`
}

type ReportInsights struct {
	OverallAnalysis string           `json:"overall_analysis"`
	Strengths       []string         `json:"strengths"`
	Weaknesses      []string         `json:"weaknesses"`
	Suggestions     []string         `json:"suggestions"`
	TechnicalScore  int              `json:"technical_score"`
	ExpressionScore int              `json:"expression_score"`
	LogicScore      int              `json:"logic_score"`
	MatchingScore   int              `json:"matching_score"`
	BehaviorScore   int              `json:"behavior_score"`
	QADetails       []ReportQADetail `json:"qa_details"`
}

// Chat exposes the raw LLM chat capability.
func (s *AIService) Chat(ctx context.Context, prompt string) (string, error) {
	return s.chat(ctx, prompt, "chat", nil)
}

// ChatWithTask exposes the raw LLM chat capability with a specific task type.
func (s *AIService) ChatWithTask(ctx context.Context, prompt string, taskType string) (string, error) {
	return s.chat(ctx, prompt, taskType, nil)
}

func (s *AIService) ChatWithFormat(ctx context.Context, prompt string, taskType string, responseFormat *llm.ResponseFormat) (string, error) {
	return s.chat(ctx, prompt, taskType, responseFormat)
}

func (s *AIService) RenderPrompt(templateName string, data interface{}) (string, error) {
	return s.renderPrompt(templateName, data)
}

func jsonObjectResponseFormat() *llm.ResponseFormat {
	return &llm.ResponseFormat{Type: llm.ResponseFormatJSON}
}

func (s *AIService) renderPrompt(templateName string, data interface{}) (string, error) {
	if s.promptManager == nil {
		return "", fmt.Errorf("prompt manager is not initialized")
	}
	return s.promptManager.Render(templateName, data)
}

func (s *AIService) chat(ctx context.Context, prompt string, taskType string, responseFormat *llm.ResponseFormat) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if s.llmClient == nil {
		return "", fmt.Errorf("llm client is not configured")
	}

	model := ""
	if s.config != nil {
		model = s.config.Model
		if specificModel, ok := s.config.Models[taskType]; ok && strings.TrimSpace(specificModel) != "" {
			model = specificModel
		}
	}
	model = strings.TrimSpace(model)
	if model == "" {
		model = "deepseek-chat"
	}

	req := llm.ChatRequest{
		Model:          model,
		Messages:       []llm.ChatMessage{{Role: "user", Content: prompt}},
		Temperature:    s.temperatureForTask(taskType),
		ResponseFormat: responseFormat,
	}

	return s.llmClient.Chat(ctx, req)
}

func (s *AIService) temperatureForTask(taskType string) float64 {
	switch taskType {
	case "evaluation":
		return 0.05
	case "report", "resume":
		return 0.2
	case "shadow_hint":
		return 0.5
	default:
		return 0.7
	}
}
