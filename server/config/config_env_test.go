package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandConfigEnvPlaceholders(t *testing.T) {
	t.Setenv("CFG_TEST_TOKEN", "abc123")

	input := []byte("api_key: ${CFG_TEST_TOKEN}\nport: ${CFG_TEST_PORT:-8080}\n")
	resolved, err := expandConfigEnvPlaceholders(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := string(resolved)
	if !strings.Contains(text, "api_key: abc123") {
		t.Fatalf("expected token to be expanded, got: %s", text)
	}
	if !strings.Contains(text, "port: 8080") {
		t.Fatalf("expected default value to be applied, got: %s", text)
	}
}

func TestExpandConfigEnvPlaceholdersMissingRequired(t *testing.T) {
	_, err := expandConfigEnvPlaceholders([]byte("jwt:\n  secret: ${CFG_TEST_MISSING_SECRET}\n"))
	if err == nil {
		t.Fatalf("expected missing env var error, got nil")
	}
	if !strings.Contains(err.Error(), "CFG_TEST_MISSING_SECRET") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExpandConfigEnvPlaceholdersEmptyUsesDefault(t *testing.T) {
	t.Setenv("CFG_TEST_EMPTY", "")

	resolved, err := expandConfigEnvPlaceholders([]byte("x: ${CFG_TEST_EMPTY:-fallback}\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(resolved), "x: fallback") {
		t.Fatalf("expected fallback to be used, got: %s", string(resolved))
	}
}

func TestLoadDotEnvIfPresent(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	envPath := filepath.Join(dir, ".env")

	if err := os.WriteFile(envPath, []byte("CFG_TEST_FROM_DOTENV=from_file\nCFG_TEST_PRIORITY=from_file\n"), 0644); err != nil {
		t.Fatalf("write .env failed: %v", err)
	}

	if err := os.Unsetenv("CFG_TEST_FROM_DOTENV"); err != nil {
		t.Fatalf("unset env failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("CFG_TEST_FROM_DOTENV") })

	t.Setenv("CFG_TEST_PRIORITY", "from_process")

	if err := loadDotEnvIfPresent(configPath); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := os.Getenv("CFG_TEST_FROM_DOTENV"); got != "from_file" {
		t.Fatalf("expected dotenv value, got: %s", got)
	}
	if got := os.Getenv("CFG_TEST_PRIORITY"); got != "from_process" {
		t.Fatalf("expected process env to keep precedence, got: %s", got)
	}
}
