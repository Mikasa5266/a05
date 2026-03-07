package asr

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type WhisperClient struct {
	apiKey  string
	baseURL string
}

func NewWhisperClient(apiKey, baseURL string) *WhisperClient {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://api.openai.com/v1"
		if apiKey == "" {
			baseURL = "http://localhost:9000/v1"
		}
	}

	return &WhisperClient{
		apiKey:  apiKey,
		baseURL: baseURL,
	}
}

type TranscriptionRequest struct {
	File           []byte
	Language       string
	Prompt         string
	ResponseFormat string
}

type TranscriptionResponse struct {
	Text     string  `json:"text"`
	Task     string  `json:"task"`
	Language string  `json:"language"`
	Duration float64 `json:"duration"`
	Segments []struct {
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
	if len(audioData) == 0 {
		return "", fmt.Errorf("audio data is empty")
	}

	url := fmt.Sprintf("%s/audio/transcriptions", c.baseURL)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "audio.webm")
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

	err = writer.WriteField("response_format", "json")
	if err != nil {
		return "", fmt.Errorf("failed to write response format field: %w", err)
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
		Timeout: 60 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status: %d", resp.StatusCode)
	}

	var result TranscriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Text, nil
}

func (c *WhisperClient) TranscribeBase64Audio(base64Audio string, language string) (string, error) {
	audioData, err := base64.StdEncoding.DecodeString(base64Audio)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 audio: %w", err)
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
