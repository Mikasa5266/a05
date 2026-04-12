package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"your-project/config"
	"your-project/internal/initializer"
	"your-project/internal/model"
	"your-project/internal/repository"
	"your-project/internal/service"
	"your-project/pkg/llm"
	"your-project/router"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	configPath, err := resolveConfigPath()
	if err != nil {
		log.Fatalf("Failed to locate config file: %v", err)
	}

	if err := config.LoadConfig(configPath); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	cfg := config.GetConfig()
	llmClient := llm.NewDeepSeekClient(cfg)
	aiService := service.MustNewAIService(llmClient)

	initDefaultsFlag := flag.Bool("init-defaults", false, "ensure default job positions at startup")
	initSampleQuestionsFlag := flag.Bool("init-sample-questions", false, "seed sample questions from knowledge base at startup")
	flag.Parse()

	db, err := initDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	repository.SetDB(db)
	service.SetPracticeQuestionRepository(repository.NewPracticeQuestionRepository())

	if err := autoMigrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	bootstrapOpts := initializer.BootstrapOptionsFromEnv()
	if *initDefaultsFlag {
		bootstrapOpts.EnsureDefaultPositions = true
	}
	if *initSampleQuestionsFlag {
		bootstrapOpts.SeedSampleQuestions = true
	}
	if err := initializer.RunBootstrap(db, bootstrapOpts); err != nil {
		log.Printf("Warning: Startup bootstrap failed: %v", err)
	}

	r := router.SetupRouter(aiService)
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	if addr == ":" {
		addr = ":8082"
	}

	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func resolveConfigPath() (string, error) {
	candidates := []string{
		"config.yaml",
		filepath.Join("..", "config.yaml"),
		filepath.Join("..", "..", "config.yaml"),
	}

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		candidates = append(candidates,
			filepath.Join(exeDir, "config.yaml"),
			filepath.Join(exeDir, "..", "config.yaml"),
			filepath.Join(exeDir, "..", "..", "config.yaml"),
		)
	}

	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		if strings.TrimSpace(candidate) == "" {
			continue
		}
		abs, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		if _, ok := seen[abs]; ok {
			continue
		}
		seen[abs] = struct{}{}
		if info, err := os.Stat(abs); err == nil && !info.IsDir() {
			return abs, nil
		}
	}

	return "", fmt.Errorf("config.yaml not found from current working directory")
}

func initDatabase() (*gorm.DB, error) {
	cfg := config.GetConfig()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Database,
	)

	gormConfig := &gorm.Config{}
	if os.Getenv("DEBUG") == "true" {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

func autoMigrate(db *gorm.DB) error {
	if err := model.NormalizeLegacyMigrationData(db); err != nil {
		return err
	}

	return db.AutoMigrate(
		&model.User{},
		&model.JobPosition{},
		&model.Question{},
		&model.QuestionPracticeRecord{},
		&model.QuestionAssessment{},
		&model.QuestionAssessmentItem{},
		&model.PracticeQuestionFavorite{},
		&model.PracticeWrongBookEntry{},
		&model.PracticeQuestionList{},
		&model.PracticeQuestionListItem{},
		&model.PracticeAssessmentAnswer{},
		&model.PracticeInterviewSyncLog{},
		&model.ResumeParseResult{},
		&model.Interview{},
		&model.InterviewQuestion{},
		&model.AnswerResult{},
		&model.Report{},
		&model.HumanInterviewer{},
		&model.InterviewBooking{},
		&model.HumanInterviewInvitation{},
		// Enterprise
		&model.Enterprise{},
		&model.Job{},
		&model.TalentRecord{},
		&model.InterviewSession{},
		&model.CapabilityStandard{},
		&model.Referral{},
		// University
		&model.University{},
		&model.StudentRecord{},
		&model.Course{},
		&model.TalentPush{},
		// Community
		&model.CommunityPost{},
		&model.PostComment{},
		&model.MentorBooking{},
		&model.PostLike{},
	)
}
