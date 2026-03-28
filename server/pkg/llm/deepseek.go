package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultDeepSeekBaseURL = "https://api.deepseek.com/v1"

var sharedHTTPClient = &http.Client{
	Timeout: 60 * time.Second,
	Transport: &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	},
}

type DeepSeekClient struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

func NewDeepSeekClient(apiKey, model, baseURL string) *DeepSeekClient {
	trimmedModel := strings.TrimSpace(model)
	if trimmedModel == "" {
		trimmedModel = "deepseek-chat"
	}
	trimmedBaseURL := strings.TrimSpace(baseURL)
	if trimmedBaseURL == "" {
		trimmedBaseURL = defaultDeepSeekBaseURL
	}

	return &DeepSeekClient{
		apiKey:     strings.TrimSpace(apiKey),
		baseURL:    trimmedBaseURL,
		model:      trimmedModel,
		httpClient: sharedHTTPClient,
	}
}

type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int         `json:"index"`
		Message      ChatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func (c *DeepSeekClient) Chat(ctx context.Context, req ChatRequest) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if len(req.Messages) == 0 {
		return "", fmt.Errorf("messages cannot be empty")
	}
	if req.Stream {
		return "", fmt.Errorf("stream mode is not supported for DeepSeekClient")
	}

	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = c.model
	}
	if model == "" {
		return "", fmt.Errorf("model is required")
	}

	payload := ChatRequest{
		Model:          model,
		Messages:       req.Messages,
		Stream:         req.Stream,
		Temperature:    req.Temperature,
		ResponseFormat: req.ResponseFormat,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := strings.TrimSuffix(c.baseURL, "/") + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var result ChatResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w, body: %s", err, string(respBody))
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	return result.Choices[0].Message.Content, nil
}

func (c *DeepSeekClient) GenerateText(ctx context.Context, prompt string) (string, error) {
	message := ChatMessage{
		Role:    "user",
		Content: prompt,
	}

	return c.Chat(ctx, ChatRequest{
		Messages: []ChatMessage{message},
	})
}

func (c *DeepSeekClient) GenerateStructuredResponse(ctx context.Context, prompt string, format string) (string, error) {
	message := ChatMessage{
		Role:    "user",
		Content: fmt.Sprintf("%s\n\nPlease strictly follow this format:\n%s", prompt, format),
	}

	return c.Chat(ctx, ChatRequest{
		Messages:       []ChatMessage{message},
		ResponseFormat: &ResponseFormat{Type: ResponseFormatJSON},
	})
}
