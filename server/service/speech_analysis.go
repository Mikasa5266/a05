package service

import (
	"math"
	"regexp"
	"strings"
	"unicode"
)

// SpeechMetrics holds real-time speech analysis results for a single audio chunk.
type SpeechMetrics struct {
	// Speech rate in characters per minute.
	SpeechRate float64 `json:"speech_rate"`
	// Speech rate level: "slow", "normal", "fast".
	SpeechRateLevel string `json:"speech_rate_level"`
	// Detected filler words in this chunk.
	FillerWords []string `json:"filler_words"`
	// Total filler word count in this chunk.
	FillerWordCount int `json:"filler_word_count"`
	// Whether fluency alert triggered (too many fillers).
	FluencyAlert bool `json:"fluency_alert"`
	// Transcribed text from this chunk.
	TranscribedText string `json:"transcribed_text"`
	// Audio duration in seconds.
	Duration float64 `json:"duration"`
	// Character count.
	CharCount int `json:"char_count"`
	// Frontend-reported chunk energy level (0-1).
	EnergyLevel float64 `json:"energy_level"`
	// Whether energy implies meaningful speech is detected.
	AudioDetected bool `json:"audio_detected"`
	// Whether ASR was skipped due low-energy/no-speech gate.
	ASRSkipped bool `json:"asr_skipped"`
}

// SpeechAnalysisService provides real-time speech analysis capabilities.
type SpeechAnalysisService struct {
	aiService *AIService
}

func NewSpeechAnalysisService() *SpeechAnalysisService {
	return &SpeechAnalysisService{
		aiService: NewAIService(),
	}
}

// Common Chinese filler words/phrases.
var fillerWordPatterns = regexp.MustCompile(`嗯|啊|呃|那个|然后|就是|这个|怎么说|对吧`)

const (
	minAudioDetectedEnergy = 0.02
	minASREnergyGate       = 0.03
)

// AnalyzeAudioChunk transcribes an audio chunk and computes speech metrics.
func (s *SpeechAnalysisService) AnalyzeAudioChunk(audioBase64 string, chunkDurationSec float64, energyLevel float64) (*SpeechMetrics, error) {
	normalizedEnergy := normalizeEnergy(energyLevel)
	audioDetected := normalizedEnergy >= minAudioDetectedEnergy

	// Silence/very-low-energy chunks should skip ASR to avoid hallucinated text.
	if normalizedEnergy < minASREnergyGate {
		metrics := s.AnalyzeText("", chunkDurationSec)
		metrics.EnergyLevel = normalizedEnergy
		metrics.AudioDetected = audioDetected
		metrics.ASRSkipped = true
		return metrics, nil
	}

	transcribedText, err := s.aiService.TranscribeAudio(audioBase64)
	if err != nil {
		return nil, err
	}

	transcribedText = strings.TrimSpace(transcribedText)
	if !isUsableChunkTranscription(transcribedText, normalizedEnergy) {
		transcribedText = ""
	}

	metrics := s.AnalyzeText(transcribedText, chunkDurationSec)
	metrics.EnergyLevel = normalizedEnergy
	metrics.AudioDetected = audioDetected
	metrics.ASRSkipped = false
	return metrics, nil
}

// AnalyzeText computes speech metrics from already-transcribed text and known duration.
func (s *SpeechAnalysisService) AnalyzeText(text string, durationSec float64) *SpeechMetrics {
	text = strings.TrimSpace(text)
	charCount := countMeaningfulChars(text)

	var speechRate float64
	if durationSec > 0 {
		effectiveDuration := durationSec
		if effectiveDuration < 5 {
			effectiveDuration = 5
		}
		speechRate = float64(charCount) / (effectiveDuration / 60.0)
		if speechRate > 280 {
			speechRate = 280
		}
	}

	speechRateLevel := classifySpeechRate(speechRate)

	fillerMatches := fillerWordPatterns.FindAllString(text, -1)
	if fillerMatches == nil {
		fillerMatches = []string{}
	}

	fillerPerMinute := float64(len(fillerMatches)) / math.Max(durationSec/60.0, 0.1)
	fluencyAlert := fillerPerMinute > 5

	return &SpeechMetrics{
		SpeechRate:      math.Round(speechRate*10) / 10,
		SpeechRateLevel: speechRateLevel,
		FillerWords:     fillerMatches,
		FillerWordCount: len(fillerMatches),
		FluencyAlert:    fluencyAlert,
		TranscribedText: text,
		Duration:        durationSec,
		CharCount:       charCount,
	}
}

func countMeaningfulChars(text string) int {
	count := 0
	for _, r := range text {
		if unicode.IsSpace(r) || unicode.IsPunct(r) {
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.In(r, unicode.Han) {
			count++
		}
	}
	return count
}

func normalizeEnergy(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return math.Round(value*1000) / 1000
}

func isUsableChunkTranscription(text string, energyLevel float64) bool {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return false
	}

	normalized := strings.ToLower(trimmed)
	normalized = strings.Trim(normalized, " \t\r\n.,!?;:'\"`~()[]{}<>，。！？；：")
	if normalized == "" {
		return false
	}

	lowInfo := map[string]struct{}{
		"you":          {},
		"uh":           {},
		"um":           {},
		"hmm":          {},
		"嗯":            {},
		"啊":            {},
		"呃":            {},
		"哦":            {},
		"不知道":          {},
		"听不清":          {},
		"i don't know": {},
	}
	if _, ok := lowInfo[normalized]; ok {
		return false
	}

	// Under low energy, short text is likely hallucinated.
	if energyLevel < 0.04 && len([]rune(trimmed)) <= 10 {
		return false
	}

	return true
}

// classifySpeechRate maps chars/min to a level.
func classifySpeechRate(rate float64) string {
	switch {
	case rate < 120:
		return "slow"
	case rate <= 240:
		return "normal"
	default:
		return "fast"
	}
}

// AccumulatedSpeechStats tracks running totals for an entire interview session.
type AccumulatedSpeechStats struct {
	TotalChars       int     `json:"total_chars"`
	TotalDuration    float64 `json:"total_duration_sec"`
	TotalFillerWords int     `json:"total_filler_words"`
	AvgSpeechRate    float64 `json:"avg_speech_rate"`
	AvgRateLevel     string  `json:"avg_rate_level"`
}

// ComputeAccumulatedStats computes overall stats from a sequence of chunk metrics.
func ComputeAccumulatedStats(chunks []*SpeechMetrics) *AccumulatedSpeechStats {
	var totalChars int
	var totalDuration float64
	var totalFillers int

	for _, c := range chunks {
		totalChars += c.CharCount
		totalDuration += c.Duration
		totalFillers += c.FillerWordCount
	}

	var avgRate float64
	if totalDuration > 0 {
		avgRate = float64(totalChars) / (totalDuration / 60.0)
		avgRate = math.Round(avgRate*10) / 10
	}

	return &AccumulatedSpeechStats{
		TotalChars:       totalChars,
		TotalDuration:    math.Round(totalDuration*10) / 10,
		TotalFillerWords: totalFillers,
		AvgSpeechRate:    avgRate,
		AvgRateLevel:     classifySpeechRate(avgRate),
	}
}
