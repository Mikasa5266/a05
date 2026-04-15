package initializer

import (
	"fmt"
	"os"
	"strings"

	"your-project/internal/repository"

	"gorm.io/gorm"
)

type BootstrapOptions struct {
	EnsureDefaultPositions bool
	SeedSampleQuestions    bool
	SeedDemoAccounts       bool
}

func BootstrapOptionsFromEnv() BootstrapOptions {
	return BootstrapOptions{
		EnsureDefaultPositions: parseBoolEnv("INIT_DEFAULT_POSITIONS"),
		SeedSampleQuestions:    parseBoolEnv("INIT_SAMPLE_QUESTIONS"),
		SeedDemoAccounts:       parseBoolEnv("INIT_DEMO_ACCOUNTS"),
	}
}

func RunBootstrap(db *gorm.DB, opts BootstrapOptions) error {
	if opts.SeedDemoAccounts {
		if err := EnsureDemoAccounts(db); err != nil {
			return fmt.Errorf("failed to initialize demo accounts: %w", err)
		}
	}

	if opts.EnsureDefaultPositions {
		if err := repository.NewPositionRepository().EnsureDefaults(); err != nil {
			return fmt.Errorf("failed to ensure default job positions: %w", err)
		}
	}

	if opts.SeedSampleQuestions {
		if err := InitSampleQuestions(db); err != nil {
			return fmt.Errorf("failed to initialize sample questions: %w", err)
		}
	}

	return nil
}

func parseBoolEnv(key string) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	switch v {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
