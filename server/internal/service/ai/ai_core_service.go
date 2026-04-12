package ai

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"text/template"

	"your-project/config"
	"your-project/pkg/llm"
	promptpkg "your-project/pkg/prompt"
)

type AIService struct {
	config               *config.LLMConfig
	llmClient            llm.LLMClient
	templateManager      *promptpkg.PromptManager
	promptManager        DynamicPromptManager
	groundTruthRetriever GroundTruthRetriever
}

// DynamicPromptManager provides hot-reload prompt access by key/path/version.
type DynamicPromptManager interface {
	GetPrompt(key string) (string, error)
	GetPromptByPath(promptPath string) (string, error)
	GetPromptVersion(key, version string) (string, error)
	RenderPrompt(key string, data interface{}) (string, error)
}

// GroundTruthRetriever fetches reference snippets for evaluation grounding.
type GroundTruthRetriever func(ctx context.Context, query string, limit int) ([]string, error)

func NewAIService(llmClient llm.LLMClient, dynamicPromptManager ...DynamicPromptManager) *AIService {
	tplManager, err := promptpkg.NewPromptManager()
	if err != nil {
		log.Printf("failed to initialize legacy template manager: %v", err)
	}

	var promptManager DynamicPromptManager
	if len(dynamicPromptManager) > 0 {
		promptManager = dynamicPromptManager[0]
	}
	if promptManager != nil {
		registerDynamicPromptManager(promptManager)
	}

	return &AIService{
		config:          &config.GetConfig().LLM,
		llmClient:       llmClient,
		templateManager: tplManager,
		promptManager:   promptManager,
	}
}

func (s *AIService) SetPromptManager(pm DynamicPromptManager) {
	s.promptManager = pm
	if pm != nil {
		registerDynamicPromptManager(pm)
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
	if s.promptManager != nil {
		for _, key := range dynamicPromptKeysForTemplate(templateName) {
			rendered, err := s.renderDynamicPrompt(key, data)
			if err == nil && strings.TrimSpace(rendered) != "" {
				return rendered, nil
			}
		}
	}

	if s.templateManager == nil {
		return "", fmt.Errorf("legacy template manager is not initialized")
	}
	return s.templateManager.Render(templateName, data)
}

func dynamicPromptKeysForTemplate(templateName string) []string {
	trimmed := strings.TrimSpace(templateName)
	if trimmed == "" {
		return nil
	}
	base := strings.TrimSuffix(trimmed, ".tmpl")
	base = strings.TrimSpace(strings.ReplaceAll(base, "\\", "/"))
	if base == "" {
		return nil
	}
	return []string{
		"ai/template/" + base,
		"ai/templates/" + base,
		base,
	}
}

func (s *AIService) getPrompt(key string) (string, error) {
	if s.promptManager == nil {
		return "", fmt.Errorf("dynamic prompt manager is not initialized")
	}
	return s.promptManager.GetPrompt(key)
}

func (s *AIService) getPromptOrDefault(key, fallback string) string {
	prompt, err := s.getPrompt(key)
	if err != nil {
		return fallback
	}
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return fallback
	}
	return prompt
}

func (s *AIService) renderDynamicPrompt(key string, data interface{}) (string, error) {
	if s.promptManager == nil {
		return "", fmt.Errorf("dynamic prompt manager is not initialized")
	}
	if rendered, err := s.promptManager.RenderPrompt(key, data); err == nil && strings.TrimSpace(rendered) != "" {
		return rendered, nil
	}

	raw, err := s.promptManager.GetPrompt(key)
	if err != nil {
		return "", err
	}

	tpl, err := template.New(key).Option("missingkey=error").Parse(raw)
	if err != nil {
		return "", fmt.Errorf("failed to parse dynamic prompt %q: %w", key, err)
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render dynamic prompt %q: %w", key, err)
	}
	return strings.TrimSpace(buf.String()), nil
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
