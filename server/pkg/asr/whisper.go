package asr

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type WhisperClient struct {
	apiKey  string
	baseURL string
	model   string
}

func NewWhisperClient(apiKey, baseURL, model string) *WhisperClient {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://api.openai.com/v1"
		if apiKey == "" {
			baseURL = "http://localhost:9000/v1"
		}
	}
	if strings.TrimSpace(model) == "" {
		model = "whisper-1"
	}

	return &WhisperClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
	}
}

type TranscriptionRequest struct {
	File           []byte
	Language       string
	Prompt         string
	ResponseFormat string
}

type TranscriptionResponse struct {
	Text       string  `json:"text"`
	Result     string  `json:"result"`
	Transcript string  `json:"transcript"`
	Task       string  `json:"task"`
	Language   string  `json:"language"`
	Duration   float64 `json:"duration"`
	Segments   []struct {
		ID               int     `json:"id"`
		Start            float64 `json:"start"`
		End              float64 `json:"end"`
		Text             string  `json:"text"`
		AvgLogprob       float64 `json:"avg_logprob"`
		CompressionRatio float64 `json:"compression_ratio"`
		NoSpeechProb     float64 `json:"no_speech_prob"`
	} `json:"segments"`
}

func (c *WhisperClient) TranscribeAudio(audioData []byte, language string) (string, error) {
	return c.TranscribeAudioWithOptions(audioData, language, "", "")
}

func (c *WhisperClient) TranscribeAudioWithOptions(audioData []byte, language, mimeType, prompt string) (string, error) {
	if len(audioData) == 0 {
		return "", fmt.Errorf("audio data is empty")
	}

	url := fmt.Sprintf("%s/audio/transcriptions", c.baseURL)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	filename := "audio" + extByMimeType(mimeType)
	fileHeader := make(textproto.MIMEHeader)
	fileHeader.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filename))
	if strings.TrimSpace(mimeType) != "" {
		fileHeader.Set("Content-Type", mimeType)
	}

	part, err := writer.CreatePart(fileHeader)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = part.Write(audioData)
	if err != nil {
		return "", fmt.Errorf("failed to write audio data: %w", err)
	}

	if language != "" {
		err = writer.WriteField("language", language)
		if err != nil {
			return "", fmt.Errorf("failed to write language field: %w", err)
		}
	}

	if strings.TrimSpace(prompt) != "" {
		err = writer.WriteField("prompt", prompt)
		if err != nil {
			return "", fmt.Errorf("failed to write prompt field: %w", err)
		}
	}

	err = writer.WriteField("response_format", "json")
	if err != nil {
		return "", fmt.Errorf("failed to write response format field: %w", err)
	}

	err = writer.WriteField("model", c.model)
	if err != nil {
		return "", fmt.Errorf("failed to write model field: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	if c.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	}

	client := &http.Client{
		Timeout: 180 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned status: %d, body: %s", resp.StatusCode, strings.TrimSpace(string(respBytes)))
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	text, err := extractTranscriptionText(respBytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse transcription response: %w, body: %s", err, strings.TrimSpace(string(respBytes)))
	}

	return text, nil
}

func extractTranscriptionText(respBytes []byte) (string, error) {
	trimmed := strings.TrimSpace(string(respBytes))
	if trimmed == "" {
		return "", nil
	}

	// Some gateways may return plain text directly.
	if !strings.HasPrefix(trimmed, "{") && !strings.HasPrefix(trimmed, "[") {
		var quoted string
		if err := json.Unmarshal(respBytes, &quoted); err == nil {
			return strings.TrimSpace(quoted), nil
		}
		return trimmed, nil
	}

	var result TranscriptionResponse
	if err := json.Unmarshal(respBytes, &result); err == nil {
		if text := strings.TrimSpace(firstNonEmpty(result.Text, result.Transcript, result.Result)); text != "" {
			return text, nil
		}
		if text := flattenStructuredSegments(result.Segments); text != "" {
			return text, nil
		}
	}

	var generic map[string]interface{}
	if err := json.Unmarshal(respBytes, &generic); err == nil {
		if text := extractTextFromGenericPayload(generic); text != "" {
			return text, nil
		}
		return "", nil
	}

	return "", fmt.Errorf("unsupported response format")
}

func firstNonEmpty(values ...string) string {
	for _, item := range values {
		if strings.TrimSpace(item) != "" {
			return item
		}
	}
	return ""
}

func flattenStructuredSegments(segments []struct {
	ID               int     `json:"id"`
	Start            float64 `json:"start"`
	End              float64 `json:"end"`
	Text             string  `json:"text"`
	AvgLogprob       float64 `json:"avg_logprob"`
	CompressionRatio float64 `json:"compression_ratio"`
	NoSpeechProb     float64 `json:"no_speech_prob"`
}) string {
	var builder strings.Builder
	for _, seg := range segments {
		if strings.TrimSpace(seg.Text) == "" {
			continue
		}
		builder.WriteString(seg.Text)
	}
	return strings.TrimSpace(builder.String())
}

func extractTextFromGenericPayload(payload map[string]interface{}) string {
	for _, key := range []string{"text", "transcript", "result", "output_text"} {
		if text := asNonEmptyString(payload[key]); text != "" {
			return text
		}
	}

	if dataNode, ok := payload["data"].(map[string]interface{}); ok {
		for _, key := range []string{"text", "transcript", "result"} {
			if text := asNonEmptyString(dataNode[key]); text != "" {
				return text
			}
		}
	}

	if text := flattenGenericSegments(payload["segments"]); text != "" {
		return text
	}

	if choices, ok := payload["choices"].([]interface{}); ok && len(choices) > 0 {
		if firstChoice, ok := choices[0].(map[string]interface{}); ok {
			if text := asNonEmptyString(firstChoice["text"]); text != "" {
				return text
			}
			if msg, ok := firstChoice["message"].(map[string]interface{}); ok {
				if text := asNonEmptyString(msg["content"]); text != "" {
					return text
				}
			}
		}
	}

	return ""
}

func flattenGenericSegments(raw interface{}) string {
	segments, ok := raw.([]interface{})
	if !ok || len(segments) == 0 {
		return ""
	}

	var builder strings.Builder
	for _, seg := range segments {
		segMap, ok := seg.(map[string]interface{})
		if !ok {
			continue
		}
		text := asNonEmptyString(segMap["text"])
		if text == "" {
			continue
		}
		builder.WriteString(text)
	}
	return strings.TrimSpace(builder.String())
}

func asNonEmptyString(raw interface{}) string {
	if value, ok := raw.(string); ok {
		return strings.TrimSpace(value)
	}
	return ""
}

func extByMimeType(mimeType string) string {
	mime := strings.ToLower(strings.TrimSpace(mimeType))
	switch {
	case strings.Contains(mime, "webm"):
		return ".webm"
	case strings.Contains(mime, "mp4"):
		return ".mp4"
	case strings.Contains(mime, "mpeg"), strings.Contains(mime, "mp3"):
		return ".mp3"
	case strings.Contains(mime, "wav"):
		return ".wav"
	case strings.Contains(mime, "ogg"):
		return ".ogg"
	case strings.Contains(mime, "m4a"), strings.Contains(mime, "aac"):
		return ".m4a"
	default:
		return ".webm"
	}
}

func (c *WhisperClient) TranscribeBase64Audio(base64Audio string, language string) (string, error) {
	cleanBase64 := strings.TrimSpace(base64Audio)
	if comma := strings.Index(cleanBase64, ","); comma >= 0 && comma < len(cleanBase64)-1 {
		cleanBase64 = strings.TrimSpace(cleanBase64[comma+1:])
	}

	audioData, err := base64.StdEncoding.DecodeString(cleanBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 audio: %w", err)
	}

	if len(audioData) < 100 {
		return "", nil
	}

	return c.TranscribeAudio(audioData, language)
}

func (c *WhisperClient) TranscribeAudioFile(filePath string, language string) (string, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found: %s", filePath)
	}

	audioData, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read audio file: %w", err)
	}

	return c.TranscribeAudio(audioData, language)
}

func (c *WhisperClient) GetSupportedLanguages() []string {
	return []string{
		"zh", "en", "es", "fr", "de", "it", "pt", "ru", "ja", "ko",
		"ar", "hi", "th", "vi", "tr", "pl", "nl", "sv", "da", "no",
	}
}

func IsAudioFileSupported(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	supported := []string{".mp3", ".mp4", ".wav", ".m4a", ".webm", ".mpeg", ".mpga"}

	for _, supportedExt := range supported {
		if ext == supportedExt {
			return true
		}
	}

	return false
}
