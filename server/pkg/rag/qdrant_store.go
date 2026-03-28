package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	defaultQdrantURL            = "http://127.0.0.1:6333"
	defaultQdrantCollection     = "rag_knowledge"
	defaultQdrantDistance       = "Cosine"
	defaultQdrantTimeout        = 8 * time.Second
	defaultQdrantMaxRetries     = 3
	defaultQdrantInitialBackoff = 200 * time.Millisecond
	defaultQdrantMaxBackoff     = 2 * time.Second
	qdrantMaxResponseBytes      = 4 << 20
)

type QdrantStoreConfig struct {
	URL            string
	APIKey         string
	Collection     string
	VectorSize     int
	Distance       string
	HTTPClient     *http.Client
	Timeout        time.Duration
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
}

type QdrantStore struct {
	url            string
	apiKey         string
	collection     string
	vectorSize     int
	distance       string
	httpClient     *http.Client
	maxRetries     int
	initialBackoff time.Duration
	maxBackoff     time.Duration

	mu              sync.Mutex
	collectionReady bool
	collectionDim   int
}

type qdrantStatusError struct {
	StatusCode int
	Message    string
}

func (e *qdrantStatusError) Error() string {
	if strings.TrimSpace(e.Message) == "" {
		return fmt.Sprintf("qdrant returned status %d", e.StatusCode)
	}
	return fmt.Sprintf("qdrant returned status %d: %s", e.StatusCode, e.Message)
}

type qdrantCreateCollectionRequest struct {
	Vectors qdrantVectorConfig `json:"vectors"`
}

type qdrantVectorConfig struct {
	Size     int    `json:"size"`
	Distance string `json:"distance"`
}

type qdrantUpsertRequest struct {
	Points []qdrantPoint `json:"points"`
}

type qdrantPoint struct {
	ID      uint64         `json:"id"`
	Vector  []float32      `json:"vector"`
	Payload map[string]any `json:"payload,omitempty"`
}

type qdrantSearchRequest struct {
	Vector      []float32 `json:"vector"`
	Limit       int       `json:"limit"`
	WithPayload bool      `json:"with_payload"`
}

type qdrantSearchResponse struct {
	Result []qdrantSearchItem `json:"result"`
}

type qdrantSearchItem struct {
	ID      any            `json:"id"`
	Score   float32        `json:"score"`
	Payload map[string]any `json:"payload"`
}

func NewQdrantStoreFromEnv() (*QdrantStore, error) {
	storeURL := strings.TrimSpace(getEnvFirst("QDRANT_URL", "QDRANT_ENDPOINT"))
	if storeURL == "" {
		storeURL = defaultQdrantURL
	}

	collection := strings.TrimSpace(os.Getenv("QDRANT_COLLECTION"))
	if collection == "" {
		collection = defaultQdrantCollection
	}

	distance := strings.TrimSpace(os.Getenv("QDRANT_DISTANCE"))
	vectorSize := parseIntEnv("QDRANT_VECTOR_SIZE", 0)
	timeout := parseDurationEnv("QDRANT_TIMEOUT", defaultQdrantTimeout)
	maxRetries := parseIntEnv("QDRANT_MAX_RETRIES", defaultQdrantMaxRetries)
	initialBackoff := parseDurationEnv("QDRANT_RETRY_INITIAL_BACKOFF", defaultQdrantInitialBackoff)
	maxBackoff := parseDurationEnv("QDRANT_RETRY_MAX_BACKOFF", defaultQdrantMaxBackoff)

	return NewQdrantStore(QdrantStoreConfig{
		URL:            storeURL,
		APIKey:         strings.TrimSpace(os.Getenv("QDRANT_API_KEY")),
		Collection:     collection,
		VectorSize:     vectorSize,
		Distance:       distance,
		Timeout:        timeout,
		MaxRetries:     maxRetries,
		InitialBackoff: initialBackoff,
		MaxBackoff:     maxBackoff,
	})
}

func NewQdrantStore(cfg QdrantStoreConfig) (*QdrantStore, error) {
	storeURL := strings.TrimSpace(cfg.URL)
	if storeURL == "" {
		storeURL = defaultQdrantURL
	}
	if _, err := url.ParseRequestURI(storeURL); err != nil {
		return nil, fmt.Errorf("invalid qdrant url %q: %w", storeURL, err)
	}

	collection := strings.TrimSpace(cfg.Collection)
	if collection == "" {
		collection = defaultQdrantCollection
	}

	distance, err := normalizeQdrantDistance(cfg.Distance)
	if err != nil {
		return nil, err
	}

	if cfg.VectorSize < 0 {
		return nil, fmt.Errorf("qdrant vector size must be >= 0")
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultQdrantTimeout
	}

	maxRetries := cfg.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}

	initialBackoff := cfg.InitialBackoff
	if initialBackoff <= 0 {
		initialBackoff = defaultQdrantInitialBackoff
	}
	maxBackoff := cfg.MaxBackoff
	if maxBackoff <= 0 {
		maxBackoff = defaultQdrantMaxBackoff
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

	store := &QdrantStore{
		url:            strings.TrimRight(storeURL, "/"),
		apiKey:         strings.TrimSpace(cfg.APIKey),
		collection:     collection,
		vectorSize:     cfg.VectorSize,
		distance:       distance,
		httpClient:     httpClient,
		maxRetries:     maxRetries,
		initialBackoff: initialBackoff,
		maxBackoff:     maxBackoff,
	}

	if store.vectorSize > 0 {
		store.collectionDim = store.vectorSize
	}

	return store, nil
}

func (s *QdrantStore) Ping(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return s.doJSONWithRetry(ctx, http.MethodGet, "/collections", nil, nil, nil)
}

func (s *QdrantStore) Upsert(ctx context.Context, points []VectorPoint) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if s == nil {
		return fmt.Errorf("qdrant store is nil")
	}
	if len(points) == 0 {
		return nil
	}

	prepared := make([]qdrantPoint, 0, len(points))
	vectorDim := 0

	for i, point := range points {
		if len(point.Vector) == 0 {
			return fmt.Errorf("point[%d] has empty vector", i)
		}
		if vectorDim == 0 {
			vectorDim = len(point.Vector)
		} else if len(point.Vector) != vectorDim {
			return fmt.Errorf("inconsistent vector dimension: got %d and %d", vectorDim, len(point.Vector))
		}
		if s.vectorSize > 0 && len(point.Vector) != s.vectorSize {
			return fmt.Errorf("point[%d] vector dimension %d does not match configured size %d", i, len(point.Vector), s.vectorSize)
		}

		rawID := strings.TrimSpace(point.ID)
		if rawID == "" {
			rawID = fmt.Sprintf("point_%d", i)
		}

		payload := map[string]any{
			"raw_id":   rawID,
			"content":  strings.TrimSpace(point.Content),
			"metadata": cloneMetadata(point.Metadata),
		}

		prepared = append(prepared, qdrantPoint{
			ID:      hashPointID(rawID),
			Vector:  point.Vector,
			Payload: payload,
		})
	}

	if len(prepared) == 0 {
		return nil
	}

	if err := s.ensureCollection(ctx, vectorDim); err != nil {
		return err
	}

	request := qdrantUpsertRequest{Points: prepared}
	query := url.Values{}
	query.Set("wait", "true")

	if err := s.doJSONWithRetry(ctx, http.MethodPut, s.collectionPath()+"/points", query, request, nil); err != nil {
		return fmt.Errorf("qdrant upsert failed: %w", err)
	}

	return nil
}

func (s *QdrantStore) Search(ctx context.Context, queryVector []float32, topK int) ([]SearchResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if s == nil {
		return nil, fmt.Errorf("qdrant store is nil")
	}
	if len(queryVector) == 0 {
		return nil, fmt.Errorf("query vector is empty")
	}
	if topK <= 0 {
		topK = 1
	}

	exists, err := s.collectionExists(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return []SearchResult{}, nil
	}

	request := qdrantSearchRequest{
		Vector:      queryVector,
		Limit:       topK,
		WithPayload: true,
	}

	var response qdrantSearchResponse
	if err := s.doJSONWithRetry(ctx, http.MethodPost, s.collectionPath()+"/points/search", nil, request, &response); err != nil {
		return nil, fmt.Errorf("qdrant search failed: %w", err)
	}

	results := make([]SearchResult, 0, len(response.Result))
	for _, item := range response.Result {
		searchResult := SearchResult{
			ID:    pointIDFromPayload(item.Payload, item.ID),
			Score: item.Score,
		}

		if content, ok := item.Payload["content"].(string); ok {
			searchResult.Content = strings.TrimSpace(content)
		}
		searchResult.Metadata = metadataFromPayload(item.Payload)

		if searchResult.Content == "" {
			continue
		}
		results = append(results, searchResult)
	}

	return results, nil
}

func (s *QdrantStore) ensureCollection(ctx context.Context, vectorDim int) error {
	if vectorDim <= 0 {
		return fmt.Errorf("vector dimension must be positive")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	targetDim := s.vectorSize
	if targetDim <= 0 {
		targetDim = vectorDim
	}
	if s.collectionReady && s.collectionDim == targetDim {
		return nil
	}

	exists, err := s.collectionExists(ctx)
	if err != nil {
		return err
	}

	if !exists {
		request := qdrantCreateCollectionRequest{
			Vectors: qdrantVectorConfig{
				Size:     targetDim,
				Distance: s.distance,
			},
		}
		if err := s.doJSONWithRetry(ctx, http.MethodPut, s.collectionPath(), nil, request, nil); err != nil {
			return fmt.Errorf("create qdrant collection %q failed: %w", s.collection, err)
		}
	}

	s.collectionReady = true
	s.collectionDim = targetDim
	return nil
}

func (s *QdrantStore) collectionExists(ctx context.Context) (bool, error) {
	err := s.doJSONWithRetry(ctx, http.MethodGet, s.collectionPath(), nil, nil, nil)
	if err == nil {
		return true, nil
	}

	var statusErr *qdrantStatusError
	if errors.As(err, &statusErr) && statusErr.StatusCode == http.StatusNotFound {
		return false, nil
	}
	return false, err
}

func (s *QdrantStore) doJSONWithRetry(ctx context.Context, method, path string, query url.Values, requestBody any, responseBody any) error {
	if ctx == nil {
		ctx = context.Background()
	}

	var payload []byte
	if requestBody != nil {
		data, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("marshal qdrant request failed: %w", err)
		}
		payload = data
	}

	var lastErr error
	for attempt := 0; attempt <= s.maxRetries; attempt++ {
		if attempt > 0 {
			if err := sleepWithBackoff(ctx, attempt, s.initialBackoff, s.maxBackoff); err != nil {
				return err
			}
		}

		retryable, err := s.doJSONOnce(ctx, method, path, query, payload, responseBody)
		if err == nil {
			return nil
		}
		lastErr = err
		if !retryable || attempt == s.maxRetries {
			break
		}
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("unknown qdrant request error")
	}
	return lastErr
}

func (s *QdrantStore) doJSONOnce(ctx context.Context, method, path string, query url.Values, payload []byte, responseBody any) (bool, error) {
	reqURL := s.buildURL(path, query)
	req, err := http.NewRequestWithContext(ctx, method, reqURL, bytes.NewReader(payload))
	if err != nil {
		return false, fmt.Errorf("build qdrant request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if s.apiKey != "" {
		req.Header.Set("api-key", s.apiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return false, err
		}
		return isRetryableTransportError(err), fmt.Errorf("qdrant request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, qdrantMaxResponseBytes))
	if err != nil {
		return true, fmt.Errorf("read qdrant response failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return isRetryableStatus(resp.StatusCode), &qdrantStatusError{
			StatusCode: resp.StatusCode,
			Message:    extractQdrantErrorMessage(body),
		}
	}

	if responseBody == nil || len(body) == 0 {
		return false, nil
	}
	if err := json.Unmarshal(body, responseBody); err != nil {
		return false, fmt.Errorf("decode qdrant response failed: %w", err)
	}

	return false, nil
}

func (s *QdrantStore) collectionPath() string {
	return "/collections/" + url.PathEscape(s.collection)
}

func (s *QdrantStore) buildURL(path string, query url.Values) string {
	path = strings.TrimSpace(path)
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	finalURL := s.url + path
	if query != nil && len(query) > 0 {
		finalURL += "?" + query.Encode()
	}
	return finalURL
}

func normalizeQdrantDistance(distance string) (string, error) {
	d := strings.ToLower(strings.TrimSpace(distance))
	if d == "" {
		return defaultQdrantDistance, nil
	}
	switch d {
	case "cosine":
		return "Cosine", nil
	case "dot":
		return "Dot", nil
	case "euclid", "l2":
		return "Euclid", nil
	case "manhattan", "l1":
		return "Manhattan", nil
	default:
		return "", fmt.Errorf("unsupported qdrant distance: %s", distance)
	}
}

func hashPointID(raw string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(raw))
	return h.Sum64()
}

func pointIDFromPayload(payload map[string]any, fallback any) string {
	if raw, ok := payload["raw_id"]; ok {
		trimmed := strings.TrimSpace(fmt.Sprint(raw))
		if trimmed != "" {
			return trimmed
		}
	}
	return strings.TrimSpace(fmt.Sprint(fallback))
}

func metadataFromPayload(payload map[string]any) map[string]string {
	metadata := map[string]string{}
	if payload == nil {
		return metadata
	}

	if rawMeta, ok := payload["metadata"]; ok {
		switch m := rawMeta.(type) {
		case map[string]any:
			for k, v := range m {
				metadata[k] = strings.TrimSpace(fmt.Sprint(v))
			}
		case map[string]string:
			for k, v := range m {
				metadata[k] = strings.TrimSpace(v)
			}
		}
	}

	for k, v := range payload {
		if k == "raw_id" || k == "content" || k == "metadata" {
			continue
		}
		if _, exists := metadata[k]; exists {
			continue
		}
		metadata[k] = strings.TrimSpace(fmt.Sprint(v))
	}

	return metadata
}

func extractQdrantErrorMessage(body []byte) string {
	raw := strings.TrimSpace(string(body))
	if raw == "" {
		return ""
	}

	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return raw
	}

	if status, ok := parsed["status"]; ok {
		switch v := status.(type) {
		case string:
			if strings.TrimSpace(v) != "" && strings.ToLower(strings.TrimSpace(v)) != "ok" {
				return strings.TrimSpace(v)
			}
		case map[string]any:
			if msg, ok := v["error"]; ok {
				trimmed := strings.TrimSpace(fmt.Sprint(msg))
				if trimmed != "" {
					return trimmed
				}
			}
			if msg, ok := v["message"]; ok {
				trimmed := strings.TrimSpace(fmt.Sprint(msg))
				if trimmed != "" {
					return trimmed
				}
			}
		}
	}

	if msg, ok := parsed["message"]; ok {
		trimmed := strings.TrimSpace(fmt.Sprint(msg))
		if trimmed != "" {
			return trimmed
		}
	}

	return raw
}
