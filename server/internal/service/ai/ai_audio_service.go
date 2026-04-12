package ai

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"your-project/config"
	"your-project/pkg/asr"
)

func (s *AIService) TranscribeAudio(audioData string) (string, error) {
	mimeType, base64Payload := parseAudioPayload(audioData)
	decodedAudio, err := decodeAudioBase64Payload(base64Payload)
	if err != nil {
		return "", fmt.Errorf("failed to decode audio data: %w", err)
	}

	asrConfig := config.GetConfig().ASR
	if asrConfig.Provider == "whisper" || asrConfig.Provider == "openai" || asrConfig.Provider == "" {
		return s.transcribeWithWhisper(decodedAudio, mimeType)
	}
	return "", fmt.Errorf("unsupported ASR provider: %s", asrConfig.Provider)
}

func parseAudioPayload(audioData string) (mimeType string, base64Payload string) {
	trimmed := strings.TrimSpace(audioData)
	if strings.HasPrefix(trimmed, "data:") {
		if comma := strings.Index(trimmed, ","); comma > 0 && comma < len(trimmed)-1 {
			header := trimmed[:comma]
			base64Payload = strings.TrimSpace(trimmed[comma+1:])
			meta := strings.TrimPrefix(header, "data:")
			if semi := strings.Index(meta, ";"); semi > 0 {
				mimeType = strings.TrimSpace(meta[:semi])
			} else {
				mimeType = strings.TrimSpace(meta)
			}
			return mimeType, base64Payload
		}
	}
	return "", trimmed
}

func decodeAudioBase64Payload(payload string) ([]byte, error) {
	trimmed := strings.TrimSpace(payload)
	if trimmed == "" {
		return nil, fmt.Errorf("audio payload is empty")
	}

	decodeOrders := []func(string) ([]byte, error){
		base64.StdEncoding.DecodeString,
		base64.RawStdEncoding.DecodeString,
		base64.URLEncoding.DecodeString,
		base64.RawURLEncoding.DecodeString,
	}
	for _, decode := range decodeOrders {
		if data, err := decode(trimmed); err == nil {
			return data, nil
		}
	}

	normalized := strings.ReplaceAll(trimmed, " ", "+")
	if rem := len(normalized) % 4; rem != 0 {
		normalized += strings.Repeat("=", 4-rem)
	}
	for _, decode := range decodeOrders {
		if data, err := decode(normalized); err == nil {
			return data, nil
		}
	}
	return nil, fmt.Errorf("invalid base64 payload")
}

func (s *AIService) SynthesizeSpeech(text string) ([]byte, error) {
	ttsConfig := config.GetConfig().TTS
	if !ttsConfig.Enabled {
		return nil, fmt.Errorf("tts is disabled")
	}

	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil, fmt.Errorf("text is empty")
	}
	baseURL := strings.TrimSpace(ttsConfig.BaseURL)
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	url := strings.TrimSuffix(baseURL, "/") + "/audio/speech"

	model := strings.TrimSpace(ttsConfig.Model)
	if model == "" {
		model = "tts-1-1106"
	}
	voice := strings.TrimSpace(ttsConfig.Voice)
	if voice == "" {
		voice = "alloy"
	}

	bodyMap := map[string]interface{}{
		"model":           model,
		"input":           trimmed,
		"voice":           voice,
		"response_format": "mp3",
	}
	body, err := json.Marshal(bodyMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tts request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create tts request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(ttsConfig.APIKey) != "" {
		req.Header.Set("Authorization", "Bearer "+ttsConfig.APIKey)
	}

	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call tts api: %w", err)
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tts api returned status: %d, body: %s", resp.StatusCode, string(respBytes))
	}
	return respBytes, nil
}

func (s *AIService) transcribeWithWhisper(audioData []byte, mimeType string) (string, error) {
	asrConfig := config.GetConfig().ASR
	configuredModel := strings.TrimSpace(asrConfig.Model)
	primaryModel := resolveASRModel(configuredModel)
	if configuredModel != "" && !strings.EqualFold(configuredModel, primaryModel) {
		log.Printf("ASR model '%s' is not a transcription model, fallback to '%s'", configuredModel, primaryModel)
	}
	client := asr.NewWhisperClientWithModel(config.GetConfig(), primaryModel)
	language := "zh"
	log.Printf("ASR request start: model=%s mime=%s bytes=%d", primaryModel, strings.TrimSpace(mimeType), len(audioData))

	prompt := ""
	text, err := client.TranscribeAudioWithOptions(audioData, language, mimeType, prompt)
	primaryErr := err
	if err == nil {
		text = strings.TrimSpace(text)
		if isASRPromptEcho(text) {
			primaryErr = fmt.Errorf("asr returned instruction text, possible model/provider mismatch")
		} else if text == "" {
			primaryErr = fmt.Errorf("empty transcription result")
		}
	}
	if primaryErr != nil {
		noPromptText, noPromptErr := client.TranscribeAudioWithOptions(audioData, language, mimeType, "")
		if noPromptErr == nil {
			noPromptText = strings.TrimSpace(noPromptText)
			if noPromptText != "" && !isASRPromptEcho(noPromptText) {
				if shouldRetryASR(noPromptText, len(audioData)) {
					retried, retryErr := client.TranscribeAudioWithOptions(audioData, language, mimeType, "")
					if retryErr == nil {
						retried = strings.TrimSpace(retried)
						if retried != "" && !isASRPromptEcho(retried) && !shouldRetryASR(retried, len(audioData)) {
							noPromptText = retried
						}
					}
				}
				log.Printf("ASR prompt-less retry activated: model=%s", primaryModel)
				return noPromptText, nil
			}
		}
	}

	if primaryErr == nil {
		if shouldRetryASR(text, len(audioData)) {
			retried, retryErr := client.TranscribeAudioWithOptions(audioData, language, mimeType, prompt)
			if retryErr == nil {
				retried = strings.TrimSpace(retried)
				if retried != "" && !isASRPromptEcho(retried) && !shouldRetryASR(retried, len(audioData)) {
					text = retried
				}
			}
		}
		log.Printf("ASR request success: model=%s bytes=%d chars=%d", primaryModel, len(audioData), len([]rune(text)))
		return text, nil
	}

	if strings.EqualFold(primaryModel, config.DefaultWhisperModel) {
		return "", fmt.Errorf("whisper transcription failed (model=%s): %w", primaryModel, primaryErr)
	}

	fallbackModel := config.DefaultWhisperModel
	fallbackClient := asr.NewWhisperClientWithModel(config.GetConfig(), fallbackModel)
	fallbackText, fallbackErr := fallbackClient.TranscribeAudioWithOptions(audioData, language, mimeType, prompt)
	if fallbackErr == nil {
		fallbackText = strings.TrimSpace(fallbackText)
		if fallbackText != "" && !isASRPromptEcho(fallbackText) {
			if shouldRetryASR(fallbackText, len(audioData)) {
				retried, retryErr := fallbackClient.TranscribeAudioWithOptions(audioData, language, mimeType, prompt)
				if retryErr == nil {
					retried = strings.TrimSpace(retried)
					if retried != "" && !isASRPromptEcho(retried) && !shouldRetryASR(retried, len(audioData)) {
						fallbackText = retried
					}
				}
			}
			log.Printf("ASR fallback activated: model=%s -> %s", primaryModel, fallbackModel)
			log.Printf("ASR request success: model=%s bytes=%d chars=%d", fallbackModel, len(audioData), len([]rune(fallbackText)))
			return fallbackText, nil
		}
		fallbackErr = fmt.Errorf("empty or prompt-echo transcription")
	}
	if fallbackErr != nil {
		fallbackNoPromptText, fallbackNoPromptErr := fallbackClient.TranscribeAudioWithOptions(audioData, language, mimeType, "")
		if fallbackNoPromptErr == nil {
			fallbackNoPromptText = strings.TrimSpace(fallbackNoPromptText)
			if fallbackNoPromptText != "" && !isASRPromptEcho(fallbackNoPromptText) {
				if shouldRetryASR(fallbackNoPromptText, len(audioData)) {
					retried, retryErr := fallbackClient.TranscribeAudioWithOptions(audioData, language, mimeType, "")
					if retryErr == nil {
						retried = strings.TrimSpace(retried)
						if retried != "" && !isASRPromptEcho(retried) && !shouldRetryASR(retried, len(audioData)) {
							fallbackNoPromptText = retried
						}
					}
				}
				log.Printf("ASR fallback prompt-less retry activated: model=%s -> %s", primaryModel, fallbackModel)
				log.Printf("ASR request success: model=%s bytes=%d chars=%d", fallbackModel, len(audioData), len([]rune(fallbackNoPromptText)))
				return fallbackNoPromptText, nil
			}
		}
	}

	return "", fmt.Errorf("whisper transcription failed (model=%s): %v; fallback %s failed: %v", primaryModel, primaryErr, fallbackModel, fallbackErr)
}

func isASRPromptEcho(text string) bool {
	_ = text
	return false
}

func resolveASRModel(configured string) string {
	model := strings.TrimSpace(configured)
	if model == "" {
		return config.DefaultWhisperModel
	}
	if looksLikeNonASRModel(model) {
		return config.DefaultWhisperModel
	}
	return model
}

func looksLikeNonASRModel(model string) bool {
	lower := strings.ToLower(strings.TrimSpace(model))
	if lower == "" {
		return true
	}
	if strings.Contains(lower, "transcribe") || strings.Contains(lower, "whisper") || strings.Contains(lower, "asr") {
		return false
	}
	if strings.HasPrefix(lower, "gpt-") {
		return true
	}
	nonASRTokens := []string{"gemini", "deepseek", "qwen", "glm", "claude", "grok", "llama", "ernie", "doubao"}
	for _, token := range nonASRTokens {
		if strings.Contains(lower, token) {
			return true
		}
	}
	return false
}

func isLowInformationTranscription(text string) bool {
	normalized := strings.ToLower(strings.TrimSpace(text))
	normalized = strings.Trim(normalized, " \t\r\n.,!?;:'\"`~()[]{}<>")
	if normalized == "" {
		return true
	}

	lowInfo := map[string]struct{}{
		"you":          {},
		"uh":           {},
		"um":           {},
		"hmm":          {},
		"i don't know": {},
	}
	_, exists := lowInfo[normalized]
	return exists
}

func shouldRetryASR(text string, audioBytes int) bool {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return true
	}
	if audioBytes >= 30000 && isLowInformationTranscription(trimmed) {
		return true
	}
	runes := []rune(trimmed)
	if audioBytes >= 120000 && len(runes) <= 8 {
		return true
	}
	if audioBytes >= 120000 {
		badShort := []string{"i don't know", "unknown", "can't hear", "uh", "um"}
		for _, b := range badShort {
			if strings.Contains(trimmed, b) && len(runes) <= 10 {
				return true
			}
		}
	}
	return false
}
