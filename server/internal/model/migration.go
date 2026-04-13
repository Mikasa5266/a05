package model

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// NormalizeLegacyMigrationData runs all data normalization steps needed before schema migration.
func NormalizeLegacyMigrationData(db *gorm.DB) error {
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
	if err := normalizeLegacyHumanInvitationActorFields(db); err != nil {
		return err
	}
	return nil
}

func normalizeLegacyQuestionPositionCode(db *gorm.DB) error {
	if err := db.AutoMigrate(&JobPosition{}); err != nil {
		return fmt.Errorf("failed to prepare job_positions table before question FK migration: %w", err)
	}

	defaults := append([]JobPosition{}, DefaultJobPositions...)
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

	if !db.Migrator().HasTable(&Question{}) {
		return nil
	}
	if !db.Migrator().HasColumn(&Question{}, "PositionCode") {
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

	if db.Migrator().HasColumn(&Question{}, "Position") {
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
	if !db.Migrator().HasTable(&User{}) {
		return nil
	}
	if !db.Migrator().HasColumn(&User{}, "UUID") {
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
	if !db.Migrator().HasTable(&Interview{}) {
		return nil
	}

	if !db.Migrator().HasColumn(&Interview{}, "InvitationCode") {
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
	if !db.Migrator().HasTable(&HumanInterviewInvitation{}) {
		return nil
	}

	if !db.Migrator().HasColumn(&HumanInterviewInvitation{}, "InvitationCode") {
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

func normalizeLegacyHumanInvitationActorFields(db *gorm.DB) error {
	if !db.Migrator().HasTable(&HumanInterviewInvitation{}) {
		return nil
	}

	columns := []string{
		"InitiatorUserID",
		"InitiatorUUID",
		"InitiatorRole",
		"TargetUserID",
		"TargetUUID",
		"TargetRole",
		"ScenarioType",
		"TargetParticipants",
		"StartThreshold",
	}
	for _, column := range columns {
		if db.Migrator().HasColumn(&HumanInterviewInvitation{}, column) {
			continue
		}
		if err := db.Migrator().AddColumn(&HumanInterviewInvitation{}, column); err != nil {
			return fmt.Errorf("failed to add human_interview_invitations.%s: %w", column, err)
		}
	}

	if err := db.Exec(`
		UPDATE human_interview_invitations
		SET initiator_user_id = student_id
		WHERE (initiator_user_id IS NULL OR initiator_user_id = 0) AND student_id IS NOT NULL AND student_id > 0
	`).Error; err != nil {
		return fmt.Errorf("failed to backfill initiator_user_id: %w", err)
	}

	if err := db.Exec(`
		UPDATE human_interview_invitations
		SET initiator_uuid = student_uuid
		WHERE (initiator_uuid IS NULL OR initiator_uuid = '') AND student_uuid IS NOT NULL AND student_uuid <> ''
	`).Error; err != nil {
		return fmt.Errorf("failed to backfill initiator_uuid: %w", err)
	}

	if err := db.Exec(`
		UPDATE human_interview_invitations
		SET initiator_role = 'student'
		WHERE initiator_role IS NULL OR initiator_role = ''
	`).Error; err != nil {
		return fmt.Errorf("failed to backfill initiator_role: %w", err)
	}

	if err := db.Exec(`
		UPDATE human_interview_invitations
		SET target_user_id = invitee_user_id
		WHERE (target_user_id IS NULL OR target_user_id = 0) AND invitee_user_id IS NOT NULL AND invitee_user_id > 0
	`).Error; err != nil {
		return fmt.Errorf("failed to backfill target_user_id: %w", err)
	}

	if err := db.Exec(`
		UPDATE human_interview_invitations
		SET target_uuid = invitee_uuid
		WHERE (target_uuid IS NULL OR target_uuid = '') AND invitee_uuid IS NOT NULL AND invitee_uuid <> ''
	`).Error; err != nil {
		return fmt.Errorf("failed to backfill target_uuid: %w", err)
	}

	if err := db.Exec(`
		UPDATE human_interview_invitations
		SET target_role = invitee_role
		WHERE (target_role IS NULL OR target_role = '') AND invitee_role IS NOT NULL AND invitee_role <> ''
	`).Error; err != nil {
		return fmt.Errorf("failed to backfill target_role: %w", err)
	}

	if err := db.Exec(`
		UPDATE human_interview_invitations
		SET scenario_type = 'single'
		WHERE scenario_type IS NULL OR scenario_type = ''
	`).Error; err != nil {
		return fmt.Errorf("failed to backfill scenario_type: %w", err)
	}

	if err := db.Exec(`
		UPDATE human_interview_invitations
		SET target_participants = 2
		WHERE target_participants IS NULL OR target_participants < 2
	`).Error; err != nil {
		return fmt.Errorf("failed to backfill target_participants: %w", err)
	}

	if err := db.Exec(`
		UPDATE human_interview_invitations
		SET start_threshold = 2
		WHERE start_threshold IS NULL OR start_threshold < 1
	`).Error; err != nil {
		return fmt.Errorf("failed to backfill start_threshold: %w", err)
	}

	return nil
}
