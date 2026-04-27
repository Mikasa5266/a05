package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	DefaultDeepSeekBaseURL              = "https://api.deepseek.com/v1"
	DefaultDeepSeekModel                = "deepseek-chat"
	DefaultWhisperBaseURL               = "https://api.openai.com/v1"
	DefaultWhisperLocalBaseURL          = "http://localhost:9000/v1"
	DefaultWhisperModel                 = "whisper-1"
	DefaultSecurityLogRetentionDays     = 180
	DefaultSecurityPatrolIntervalMinute = 30
	DefaultSecurityPatrolScanLimit      = 200
)

var envTemplatePattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)(:-([^}]*))?\}`)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	LLM      LLMConfig      `yaml:"llm"`
	ASR      ASRConfig      `yaml:"asr"`
	TTS      TTSConfig      `yaml:"tts"`
	Prompt   PromptConfig   `yaml:"prompt"`
	OCR      OCRConfig      `yaml:"ocr"`
	Security SecurityConfig `yaml:"security"`
}

type ContactConfig struct {
	Name  string `yaml:"name"`
	Phone string `yaml:"phone"`
	Email string `yaml:"email"`
}

type SecurityConfig struct {
	LogRetentionDays      int           `yaml:"log_retention_days"`
	PatrolIntervalMinutes int           `yaml:"patrol_interval_minutes"`
	PatrolScanLimit       int           `yaml:"patrol_scan_limit"`
	ResponsiblePerson     ContactConfig `yaml:"responsible_person"`
	EmergencyContact      ContactConfig `yaml:"emergency_contact"`
}

type PromptConfig struct {
	TemplatesDir string `yaml:"templates_dir"`
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
	Models   map[string]string `yaml:"models"` // Task-specific models: resume, chat, evaluation, report, resume_authenticity, resume_optimization, resume_template
	DeepSeek DeepSeekConfig    `yaml:"deepseek"`
}

type DeepSeekConfig struct {
	APIKey  string `yaml:"api_key"`
	BaseURL string `yaml:"base_url"`
	Model   string `yaml:"model"`
}

type ASRConfig struct {
	Provider                string        `yaml:"provider"`
	APIKey                  string        `yaml:"api_key"`
	BaseURL                 string        `yaml:"base_url"`
	Model                   string        `yaml:"model"`
	MaxAudioBytes           int           `yaml:"max_audio_bytes"`
	ChunkMinIntervalSeconds int           `yaml:"chunk_min_interval_seconds"`
	MaxCallsPerInterview    int           `yaml:"max_calls_per_interview"`
	Whisper                 WhisperConfig `yaml:"whisper"`
}

type WhisperConfig struct {
	APIKey       string `yaml:"api_key"`
	BaseURL      string `yaml:"base_url"`
	LocalBaseURL string `yaml:"local_base_url"`
	Model        string `yaml:"model"`
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
	if err := loadDotEnvIfPresent(configPath); err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}

	file, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	resolvedFile, err := expandConfigEnvPlaceholders(file)
	if err != nil {
		return fmt.Errorf("failed to resolve config environment placeholders: %w", err)
	}

	config := &Config{}
	if err := yaml.Unmarshal(resolvedFile, config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	applyDefaultConfigValues(config)

	GlobalConfig = config
	return nil
}

func GetConfig() *Config {
	return GlobalConfig
}

func applyDefaultConfigValues(cfg *Config) {
	if cfg == nil {
		return
	}
	applyLLMDefaults(&cfg.LLM)
	applyASRDefaults(&cfg.ASR)
	applySecurityDefaults(&cfg.Security)
	cfg.Prompt.TemplatesDir = strings.TrimSpace(cfg.Prompt.TemplatesDir)
}

func applySecurityDefaults(cfg *SecurityConfig) {
	if cfg == nil {
		return
	}

	if cfg.LogRetentionDays < DefaultSecurityLogRetentionDays {
		cfg.LogRetentionDays = DefaultSecurityLogRetentionDays
	}
	if cfg.PatrolIntervalMinutes <= 0 {
		cfg.PatrolIntervalMinutes = DefaultSecurityPatrolIntervalMinute
	}
	if cfg.PatrolScanLimit <= 0 {
		cfg.PatrolScanLimit = DefaultSecurityPatrolScanLimit
	}

	cfg.ResponsiblePerson.Name = strings.TrimSpace(cfg.ResponsiblePerson.Name)
	cfg.ResponsiblePerson.Phone = strings.TrimSpace(cfg.ResponsiblePerson.Phone)
	cfg.ResponsiblePerson.Email = strings.TrimSpace(cfg.ResponsiblePerson.Email)
	cfg.EmergencyContact.Name = strings.TrimSpace(cfg.EmergencyContact.Name)
	cfg.EmergencyContact.Phone = strings.TrimSpace(cfg.EmergencyContact.Phone)
	cfg.EmergencyContact.Email = strings.TrimSpace(cfg.EmergencyContact.Email)
}

func applyLLMDefaults(cfg *LLMConfig) {
	if cfg == nil {
		return
	}

	if strings.TrimSpace(cfg.DeepSeek.APIKey) == "" {
		cfg.DeepSeek.APIKey = strings.TrimSpace(cfg.APIKey)
	}
	if strings.TrimSpace(cfg.DeepSeek.BaseURL) == "" {
		cfg.DeepSeek.BaseURL = strings.TrimSpace(cfg.BaseURL)
	}
	if strings.TrimSpace(cfg.DeepSeek.Model) == "" {
		cfg.DeepSeek.Model = strings.TrimSpace(cfg.Model)
	}

	if strings.TrimSpace(cfg.DeepSeek.BaseURL) == "" {
		cfg.DeepSeek.BaseURL = DefaultDeepSeekBaseURL
	}
	if strings.TrimSpace(cfg.DeepSeek.Model) == "" {
		cfg.DeepSeek.Model = DefaultDeepSeekModel
	}

	if strings.TrimSpace(cfg.APIKey) == "" {
		cfg.APIKey = strings.TrimSpace(cfg.DeepSeek.APIKey)
	}
	if strings.TrimSpace(cfg.BaseURL) == "" {
		cfg.BaseURL = strings.TrimSpace(cfg.DeepSeek.BaseURL)
	}
	if strings.TrimSpace(cfg.Model) == "" {
		cfg.Model = strings.TrimSpace(cfg.DeepSeek.Model)
	}
}

func applyASRDefaults(cfg *ASRConfig) {
	if cfg == nil {
		return
	}

	if strings.TrimSpace(cfg.Whisper.APIKey) == "" {
		cfg.Whisper.APIKey = strings.TrimSpace(cfg.APIKey)
	}
	if strings.TrimSpace(cfg.Whisper.BaseURL) == "" {
		cfg.Whisper.BaseURL = strings.TrimSpace(cfg.BaseURL)
	}
	if strings.TrimSpace(cfg.Whisper.Model) == "" {
		cfg.Whisper.Model = strings.TrimSpace(cfg.Model)
	}

	if strings.TrimSpace(cfg.Whisper.BaseURL) == "" {
		cfg.Whisper.BaseURL = DefaultWhisperBaseURL
	}
	if strings.TrimSpace(cfg.Whisper.LocalBaseURL) == "" {
		cfg.Whisper.LocalBaseURL = DefaultWhisperLocalBaseURL
	}
	if strings.TrimSpace(cfg.Whisper.Model) == "" {
		cfg.Whisper.Model = DefaultWhisperModel
	}

	if strings.TrimSpace(cfg.APIKey) == "" {
		cfg.APIKey = strings.TrimSpace(cfg.Whisper.APIKey)
	}
	if strings.TrimSpace(cfg.BaseURL) == "" {
		cfg.BaseURL = strings.TrimSpace(cfg.Whisper.BaseURL)
	}
	if strings.TrimSpace(cfg.Model) == "" {
		cfg.Model = strings.TrimSpace(cfg.Whisper.Model)
	}
}

func expandConfigEnvPlaceholders(content []byte) ([]byte, error) {
	raw := string(content)
	missing := make([]string, 0)

	expanded := envTemplatePattern.ReplaceAllStringFunc(raw, func(token string) string {
		parts := envTemplatePattern.FindStringSubmatch(token)
		if len(parts) < 4 {
			return token
		}

		name := parts[1]
		hasDefault := strings.HasPrefix(parts[2], ":-")
		defaultValue := parts[3]

		if value, ok := os.LookupEnv(name); ok && value != "" {
			return value
		}

		if hasDefault {
			return defaultValue
		}

		missing = append(missing, name)
		return token
	})

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(uniqueSortedStrings(missing), ", "))
	}

	return []byte(expanded), nil
}

func uniqueSortedStrings(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	sort.Strings(result)
	return result
}

func loadDotEnvIfPresent(configPath string) error {
	envPath := filepath.Join(filepath.Dir(configPath), ".env")
	file, err := os.Open(envPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}

		value = strings.TrimSpace(value)
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("failed to set env %s: %w", key, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
