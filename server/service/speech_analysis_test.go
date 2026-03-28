package service

import (
	"testing"

	"your-project/config"
)

func bootstrapSpeechAnalysisTestConfig() {
	config.GlobalConfig = &config.Config{
		LLM: config.LLMConfig{},
		ASR: config.ASRConfig{},
	}
	SetAIService(NewAIService(nil))
}

func TestAnalyzeAudioChunkSkipsASROnLowEnergy(t *testing.T) {
	bootstrapSpeechAnalysisTestConfig()
	svc := NewSpeechAnalysisService()

	metrics, err := svc.AnalyzeAudioChunk("invalid_base64_payload", 4, 0.0)
	if err != nil {
		t.Fatalf("expected no error when low-energy gate skips ASR, got: %v", err)
	}
	if metrics == nil {
		t.Fatalf("expected metrics, got nil")
	}
	if !metrics.ASRSkipped {
		t.Fatalf("expected ASRSkipped=true for low-energy chunk")
	}
	if metrics.AudioDetected {
		t.Fatalf("expected AudioDetected=false for low-energy chunk")
	}
	if metrics.TranscribedText != "" {
		t.Fatalf("expected empty transcription for low-energy chunk, got: %q", metrics.TranscribedText)
	}
}

func TestAnalyzeAudioChunkRequiresValidAudioOnHighEnergy(t *testing.T) {
	bootstrapSpeechAnalysisTestConfig()
	svc := NewSpeechAnalysisService()

	_, err := svc.AnalyzeAudioChunk("invalid_base64_payload", 4, 0.5)
	if err == nil {
		t.Fatalf("expected ASR path to fail on invalid base64 when energy is high")
	}
}
