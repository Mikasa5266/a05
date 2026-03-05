package main

import (
	"fmt"
	"log"
	"os"

	"your-project/config"
	"your-project/initializer"
	"your-project/model"
	"your-project/repository"
	"your-project/router"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	if err := config.LoadConfig("config.yaml"); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := initDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	repository.SetDB(db)

	if err := autoMigrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize sample questions
	if err := initializer.InitSampleQuestions(db); err != nil {
		log.Printf("Warning: Failed to initialize sample questions: %v", err)
	}

	r := router.SetupRouter()

	cfg := config.GetConfig()
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	if addr == ":" {
		addr = ":8080"
	}

	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
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
	return db.AutoMigrate(
		&model.User{},
		&model.Question{},
		&model.Interview{},
		&model.InterviewQuestion{},
		&model.AnswerResult{},
		&model.Report{},
		&model.HumanInterviewer{},
		&model.InterviewBooking{},
		// Enterprise
		&model.Job{},
		&model.TalentRecord{},
		&model.InterviewSession{},
		&model.CapabilityStandard{},
		&model.Referral{},
		// University
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
