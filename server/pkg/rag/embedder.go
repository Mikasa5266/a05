package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultEmbeddingBaseURL        = "https://api.openai.com/v1"
	defaultEmbeddingModel          = "text-embedding-3-small"
	defaultEmbeddingTimeout        = 20 * time.Second
	defaultEmbeddingMaxRetries     = 3
	defaultEmbeddingInitialBackoff = 400 * time.Millisecond
	defaultEmbeddingMaxBackoff     = 4 * time.Second
	maxResponseBodyBytes           = 2 << 20
)

type OpenAIEmbedderConfig struct {
	APIKey         string
	BaseURL        string
	Model          string
	HTTPClient     *http.Client
	Timeout        time.Duration
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
}

type OpenAIEmbedder struct {
	apiKey         string
	baseURL        string
	model          string
	httpClient     *http.Client
	maxRetries     int
	initialBackoff time.Duration
	maxBackoff     time.Duration
}

type embeddingRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Error *openAIErrorPayload `json:"error,omitempty"`
}

type openAIErrorPayload struct {
	Message string `json:"message"`
	Type    string `json:"type,omitempty"`
	Code    any    `json:"code,omitempty"`
}

type apiStatusError struct {
	StatusCode int
	Message    string
	Retryable  bool
}

func (e *apiStatusError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("embedding API returned status %d", e.StatusCode)
	}
	return fmt.Sprintf("embedding API returned status %d: %s", e.StatusCode, e.Message)
}

func NewOpenAIEmbedderFromEnv() (*OpenAIEmbedder, error) {
	apiKey := strings.TrimSpace(getEnvFirst("EMBEDDING_API_KEY", "OPENAI_API_KEY"))
	baseURL := strings.TrimSpace(getEnvFirst("EMBEDDING_API_BASE_URL", "OPENAI_BASE_URL"))
	model := strings.TrimSpace(os.Getenv("EMBEDDING_MODEL"))

	timeout := parseDurationEnv("EMBEDDING_TIMEOUT", defaultEmbeddingTimeout)
	maxRetries := parseIntEnv("EMBEDDING_MAX_RETRIES", defaultEmbeddingMaxRetries)
	initialBackoff := parseDurationEnv("EMBEDDING_RETRY_INITIAL_BACKOFF", defaultEmbeddingInitialBackoff)
	maxBackoff := parseDurationEnv("EMBEDDING_RETRY_MAX_BACKOFF", defaultEmbeddingMaxBackoff)

	return NewOpenAIEmbedder(OpenAIEmbedderConfig{
		APIKey:         apiKey,
		BaseURL:        baseURL,
		Model:          model,
		Timeout:        timeout,
		MaxRetries:     maxRetries,
		InitialBackoff: initialBackoff,
		MaxBackoff:     maxBackoff,
	})
}

func NewOpenAIEmbedder(cfg OpenAIEmbedderConfig) (*OpenAIEmbedder, error) {
	apiKey := strings.TrimSpace(cfg.APIKey)
	if apiKey == "" {
		return nil, fmt.Errorf("embedding API key is required")
	}

	baseURL := strings.TrimSpace(cfg.BaseURL)
	if baseURL == "" {
		baseURL = defaultEmbeddingBaseURL
	}
	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = defaultEmbeddingModel
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultEmbeddingTimeout
	}

	maxRetries := cfg.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}

	initialBackoff := cfg.InitialBackoff
	if initialBackoff <= 0 {
		initialBackoff = defaultEmbeddingInitialBackoff
	}
	maxBackoff := cfg.MaxBackoff
	if maxBackoff <= 0 {
		maxBackoff = defaultEmbeddingMaxBackoff
	}
	if maxBackoff < initialBackoff {
		maxBackoff = initialBackoff
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				Proxy:               http.ProxyFromEnvironment,
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 20,
				IdleConnTimeout:     90 * time.Second,
			},
		}
	} else if httpClient.Timeout <= 0 {
		httpClient.Timeout = timeout
	}

	return &OpenAIEmbedder{
		apiKey:         apiKey,
		baseURL:        strings.TrimRight(baseURL, "/"),
		model:          model,
		httpClient:     httpClient,
		maxRetries:     maxRetries,
		initialBackoff: initialBackoff,
		maxBackoff:     maxBackoff,
	}, nil
}

func (e *OpenAIEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	if e == nil {
		return nil, fmt.Errorf("embedder is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	input := strings.TrimSpace(text)
	if input == "" {
		return nil, fmt.Errorf("embedding input text is empty")
	}

	payload := embeddingRequest{
		Input: input,
		Model: e.model,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= e.maxRetries; attempt++ {
		if attempt > 0 {
			if sleepErr := sleepWithBackoff(ctx, attempt, e.initialBackoff, e.maxBackoff); sleepErr != nil {
				return nil, sleepErr
			}
		}

		vector, retryable, reqErr := e.embedOnce(ctx, body)
		if reqErr == nil {
			return vector, nil
		}

		lastErr = reqErr
		if !retryable || attempt == e.maxRetries {
			break
		}
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("unknown embedding error")
	}
	return nil, fmt.Errorf("embedding request failed after %d attempt(s): %w", e.maxRetries+1, lastErr)
}

func (e *OpenAIEmbedder) embedOnce(ctx context.Context, payload []byte) ([]float32, bool, error) {
	endpoint := e.baseURL + "/embeddings"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, false, fmt.Errorf("failed to create embedding request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+e.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, false, err
		}
		return nil, isRetryableTransportError(err), fmt.Errorf("failed to call embedding API: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodyBytes))
	if err != nil {
		return nil, true, fmt.Errorf("failed to read embedding API response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		statusErr := buildAPIStatusError(resp.StatusCode, bodyBytes)
		return nil, statusErr.Retryable, statusErr
	}

	var parsed embeddingResponse
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return nil, false, fmt.Errorf("failed to decode embedding response: %w", err)
	}
	if parsed.Error != nil {
		retryable := false
		errType := strings.TrimSpace(strings.ToLower(parsed.Error.Type))
		switch errType {
		case "rate_limit_error", "server_error", "api_error", "overloaded_error":
			retryable = true
		}
		if parsed.Error.Type == "invalid_request_error" || parsed.Error.Type == "authentication_error" {
			retryable = false
		}
		return nil, retryable, fmt.Errorf("embedding API error: %s", strings.TrimSpace(parsed.Error.Message))
	}
	if len(parsed.Data) == 0 {
		return nil, false, fmt.Errorf("embedding API response has no data")
	}
	vector := parsed.Data[0].Embedding
	if len(vector) == 0 {
		return nil, false, fmt.Errorf("embedding API returned empty vector")
	}

	return vector, false, nil
}

func buildAPIStatusError(statusCode int, body []byte) *apiStatusError {
	message := strings.TrimSpace(string(body))
	var payload embeddingResponse
	if err := json.Unmarshal(body, &payload); err == nil && payload.Error != nil {
		if strings.TrimSpace(payload.Error.Message) != "" {
			message = strings.TrimSpace(payload.Error.Message)
		}
	}
	if message == "" {
		message = http.StatusText(statusCode)
	}

	return &apiStatusError{
		StatusCode: statusCode,
		Message:    message,
		Retryable:  isRetryableStatus(statusCode),
	}
}

func isRetryableStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusRequestTimeout,
		http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func isRetryableTransportError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return true
	}

	return false
}

func sleepWithBackoff(ctx context.Context, attempt int, initial, max time.Duration) error {
	if attempt <= 0 {
		return nil
	}
	delay := initial
	for i := 1; i < attempt; i++ {
		delay *= 2
		if delay >= max {
			delay = max
			break
		}
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func getEnvFirst(keys ...string) string {
	for _, key := range keys {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
	}
	return ""
}

func parseDurationEnv(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	if d, err := time.ParseDuration(raw); err == nil && d > 0 {
		return d
	}
	if seconds, err := strconv.Atoi(raw); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	return fallback
}

func parseIntEnv(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		return fallback
	}
	return v
}
