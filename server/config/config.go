package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	LLM      LLMConfig      `yaml:"llm"`
	ASR      ASRConfig      `yaml:"asr"`
	TTS      TTSConfig      `yaml:"tts"`
	OCR      OCRConfig      `yaml:"ocr"`
}

type OCRConfig struct {
	TesseractPath string `yaml:"tesseract_path"`
	PdftoppmPath  string `yaml:"pdftoppm_path"`
	TessdataPath  string `yaml:"tessdata_path"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	SSLMode  string `yaml:"sslmode"`
}

type JWTConfig struct {
	Secret     string `yaml:"secret"`
	ExpireTime int    `yaml:"expire_time"`
}

type LLMConfig struct {
	Provider string            `yaml:"provider"`
	APIKey   string            `yaml:"api_key"`
	BaseURL  string            `yaml:"base_url"`
	Model    string            `yaml:"model"`  // Default model
	Models   map[string]string `yaml:"models"` // Task-specific models: resume, chat, evaluation
}

type ASRConfig struct {
	Provider                string `yaml:"provider"`
	APIKey                  string `yaml:"api_key"`
	BaseURL                 string `yaml:"base_url"`
	Model                   string `yaml:"model"`
	MaxAudioBytes           int    `yaml:"max_audio_bytes"`
	ChunkMinIntervalSeconds int    `yaml:"chunk_min_interval_seconds"`
	MaxCallsPerInterview    int    `yaml:"max_calls_per_interview"`
}

type TTSConfig struct {
	Provider             string `yaml:"provider"`
	APIKey               string `yaml:"api_key"`
	BaseURL              string `yaml:"base_url"`
	Model                string `yaml:"model"`
	Voice                string `yaml:"voice"`
	Enabled              bool   `yaml:"enabled"`
	MaxCharsPerRequest   int    `yaml:"max_chars_per_request"`
	MaxCharsPerInterview int    `yaml:"max_chars_per_interview"`
}

var GlobalConfig *Config

func LoadConfig(configPath string) error {
	file, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}
	if err := yaml.Unmarshal(file, config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	GlobalConfig = config
	return nil
}

func GetConfig() *Config {
	return GlobalConfig
}
