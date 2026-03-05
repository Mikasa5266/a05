package service

import (
	"math"
	"regexp"
	"strings"
	"unicode/utf8"
)

// SpeechMetrics holds real-time speech analysis results for a single audio chunk
type SpeechMetrics struct {
	// Speech rate in characters per minute
	SpeechRate float64 `json:"speech_rate"`
	// Speech rate level: "slow", "normal", "fast"
	SpeechRateLevel string `json:"speech_rate_level"`
	// Detected filler words in this chunk
	FillerWords []string `json:"filler_words"`
	// Total filler word count in this chunk
	FillerWordCount int `json:"filler_word_count"`
	// Whether fluency alert triggered (too many fillers)
	FluencyAlert bool `json:"fluency_alert"`
	// Transcribed text from this chunk
	TranscribedText string `json:"transcribed_text"`
	// Audio duration in seconds
	Duration float64 `json:"duration"`
	// Character count
	CharCount int `json:"char_count"`
}

// SpeechAnalysisService provides real-time speech analysis capabilities
type SpeechAnalysisService struct {
	aiService *AIService
}

func NewSpeechAnalysisService() *SpeechAnalysisService {
	return &SpeechAnalysisService{
		aiService: NewAIService(),
	}
}

// fillerWordPatterns defines common Chinese filler words/phrases to detect
var fillerWordPatterns = regexp.MustCompile(`嗯+|啊+|那个|然后|呢|额+|哦+|就是说|就是|这个|怎么说|对吧|嘛|吧|呃+`)

// AnalyzeAudioChunk transcribes an audio chunk and computes speech metrics
func (s *SpeechAnalysisService) AnalyzeAudioChunk(audioBase64 string, chunkDurationSec float64) (*SpeechMetrics, error) {
	// Transcribe the audio chunk
	transcribedText, err := s.aiService.TranscribeAudio(audioBase64)
	if err != nil {
		return nil, err
	}

	return s.AnalyzeText(transcribedText, chunkDurationSec), nil
}

// AnalyzeText computes speech metrics from already-transcribed text and known duration
func (s *SpeechAnalysisService) AnalyzeText(text string, durationSec float64) *SpeechMetrics {
	text = strings.TrimSpace(text)
	charCount := utf8.RuneCountInString(text)

	// Calculate speech rate (chars per minute)
	var speechRate float64
	if durationSec > 0 {
		speechRate = float64(charCount) / (durationSec / 60.0)
	}

	// Determine speech rate level
	speechRateLevel := classifySpeechRate(speechRate)

	// Detect filler words
	fillerMatches := fillerWordPatterns.FindAllString(text, -1)
	if fillerMatches == nil {
		fillerMatches = []string{}
	}

	// Fluency alert: if filler words per minute > 5, flag it
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

// classifySpeechRate maps chars/min to a level
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

// AccumulatedSpeechStats tracks running totals for an entire interview session
type AccumulatedSpeechStats struct {
	TotalChars       int     `json:"total_chars"`
	TotalDuration    float64 `json:"total_duration_sec"`
	TotalFillerWords int     `json:"total_filler_words"`
	AvgSpeechRate    float64 `json:"avg_speech_rate"`
	AvgRateLevel     string  `json:"avg_rate_level"`
}

// ComputeAccumulatedStats computes overall stats from a sequence of chunk metrics
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
