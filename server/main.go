package main

import (
	"fmt"
	"log"
	"os"

	"your-project/config"
	"your-project/initializer"
	"your-project/model"
	"your-project/pkg/llm"
	"your-project/repository"
	"your-project/router"
	"your-project/service"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

func main() {
	if err := config.LoadConfig("config.yaml"); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	cfg := config.GetConfig()
	llmClient := llm.NewDeepSeekClient(cfg)
	aiService := service.MustNewAIService(llmClient)

	db, err := initDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	repository.SetDB(db)

	if err := autoMigrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	if err := repository.NewPositionRepository().EnsureDefaults(); err != nil {
		log.Printf("Warning: Failed to ensure default job positions: %v", err)
	}

	// Initialize sample questions
	if err := initializer.InitSampleQuestions(db); err != nil {
		log.Printf("Warning: Failed to initialize sample questions: %v", err)
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
	if err := normalizeLegacyMigrationData(db); err != nil {
		return err
	}

	return db.AutoMigrate(
		&model.User{},
		&model.JobPosition{},
		&model.Question{},
		&model.UserQuestionState{},
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

func normalizeLegacyMigrationData(db *gorm.DB) error {
	if err := normalizeLegacyQuestionPositionCode(db); err != nil {
		return err
	}
	if err := normalizeLegacyUserUUID(db); err != nil {
		return err
	}
	if err := normalizeLegacyInterviewInvitationCode(db); err != nil {
		return err
	}
	if err := normalizeLegacyHumanInvitationCode(db); err != nil {
		return err
	}
	return nil
}

func normalizeLegacyQuestionPositionCode(db *gorm.DB) error {
	if err := db.AutoMigrate(&model.JobPosition{}); err != nil {
		return fmt.Errorf("failed to prepare job_positions table before question FK migration: %w", err)
	}

	defaults := append([]model.JobPosition{}, model.DefaultJobPositions...)
	if err := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "code"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"name",
			"domain",
			"description",
			"is_active",
			"updated_at",
		}),
	}).Create(&defaults).Error; err != nil {
		return fmt.Errorf("failed to upsert default job positions before question FK migration: %w", err)
	}

	if !db.Migrator().HasTable(&model.Question{}) {
		return nil
	}
	if !db.Migrator().HasColumn(&model.Question{}, "PositionCode") {
		return nil
	}

	if err := db.Exec(`
		UPDATE questions
		SET position_code = LOWER(TRIM(position_code))
		WHERE position_code IS NOT NULL
	`).Error; err != nil {
		return fmt.Errorf("failed to normalize case of questions.position_code: %w", err)
	}

	if err := db.Exec(`
		UPDATE questions
		SET position_code = CASE
			WHEN position_code IN ('backend', 'java', 'go', 'python_backend', 'java_backend', '后端', '后端工程师', 'java后端工程师') THEN 'backend'
			WHEN position_code IN ('frontend', 'fe', 'web', '前端', '前端工程师', '前端开发工程师') THEN 'frontend'
			WHEN position_code IN ('algorithm', 'algo', '算法', '算法工程师') THEN 'algorithm'
			WHEN position_code IN ('ai', 'ml', 'llm', 'ai_engineer', 'machine_learning', '深度学习', 'ai工程师') THEN 'ai'
			ELSE position_code
		END
	`).Error; err != nil {
		return fmt.Errorf("failed to map legacy questions.position_code aliases: %w", err)
	}

	if db.Migrator().HasColumn(&model.Question{}, "Position") {
		if err := db.Exec(`
			UPDATE questions q
			LEFT JOIN job_positions jp ON jp.code = q.position_code
			SET q.position_code = CASE
				WHEN LOWER(COALESCE(q.position, '')) LIKE '%frontend%' OR COALESCE(q.position, '') LIKE '%前端%' THEN 'frontend'
				WHEN LOWER(COALESCE(q.position, '')) LIKE '%algorithm%' OR COALESCE(q.position, '') LIKE '%算法%' THEN 'algorithm'
				WHEN LOWER(COALESCE(q.position, '')) LIKE '%ai%' OR LOWER(COALESCE(q.position, '')) LIKE '%llm%' OR LOWER(COALESCE(q.position, '')) LIKE '%machine learning%' OR COALESCE(q.position, '') LIKE '%模型%' THEN 'ai'
				ELSE 'backend'
			END
			WHERE q.position_code IS NULL OR TRIM(q.position_code) = '' OR jp.code IS NULL
		`).Error; err != nil {
			return fmt.Errorf("failed to backfill invalid questions.position_code by position text: %w", err)
		}
	}

	if err := db.Exec(`
		UPDATE questions q
		LEFT JOIN job_positions jp ON jp.code = q.position_code
		SET q.position_code = 'backend'
		WHERE q.position_code IS NULL OR TRIM(q.position_code) = '' OR jp.code IS NULL
	`).Error; err != nil {
		return fmt.Errorf("failed to fallback questions.position_code to backend: %w", err)
	}

	return nil
}

func normalizeLegacyUserUUID(db *gorm.DB) error {
	if !db.Migrator().HasTable(&model.User{}) {
		return nil
	}
	if !db.Migrator().HasColumn(&model.User{}, "UUID") {
		return nil
	}

	if err := db.Exec(`
		UPDATE users
		SET uuid = UUID()
		WHERE uuid IS NULL OR uuid = ''
	`).Error; err != nil {
		return fmt.Errorf("failed to normalize users.uuid: %w", err)
	}

	return nil
}

func normalizeLegacyInterviewInvitationCode(db *gorm.DB) error {
	if !db.Migrator().HasTable(&model.Interview{}) {
		return nil
	}

	if !db.Migrator().HasColumn(&model.Interview{}, "InvitationCode") {
		if err := db.Exec(`
			ALTER TABLE interviews
			ADD COLUMN invitation_code VARCHAR(64) NULL AFTER interview_mode
		`).Error; err != nil {
			return fmt.Errorf("failed to add interviews.invitation_code for legacy migration: %w", err)
		}
	}

	if err := db.Exec(`
		UPDATE interviews
		SET invitation_code = NULL
		WHERE invitation_code = ''
	`).Error; err != nil {
		return fmt.Errorf("failed to normalize interviews.invitation_code: %w", err)
	}

	return nil
}

func normalizeLegacyHumanInvitationCode(db *gorm.DB) error {
	if !db.Migrator().HasTable(&model.HumanInterviewInvitation{}) {
		return nil
	}

	if !db.Migrator().HasColumn(&model.HumanInterviewInvitation{}, "InvitationCode") {
		if err := db.Exec(`
			ALTER TABLE human_interview_invitations
			ADD COLUMN invitation_code VARCHAR(64) NULL AFTER id
		`).Error; err != nil {
			return fmt.Errorf("failed to add human_interview_invitations.invitation_code for legacy migration: %w", err)
		}
	}

	if err := db.Exec(`
		UPDATE human_interview_invitations
		SET invitation_code = UPPER(REPLACE(UUID(), '-', ''))
		WHERE invitation_code IS NULL OR invitation_code = ''
	`).Error; err != nil {
		return fmt.Errorf("failed to backfill empty invitation_code in human_interview_invitations: %w", err)
	}

	if err := db.Exec(`
		UPDATE human_interview_invitations hi
		JOIN (
			SELECT invitation_code
			FROM human_interview_invitations
			WHERE invitation_code IS NOT NULL AND invitation_code <> ''
			GROUP BY invitation_code
			HAVING COUNT(*) > 1
		) dup ON dup.invitation_code = hi.invitation_code
		SET hi.invitation_code = UPPER(REPLACE(UUID(), '-', ''))
	`).Error; err != nil {
		return fmt.Errorf("failed to deduplicate invitation_code in human_interview_invitations: %w", err)
	}

	return nil
}
