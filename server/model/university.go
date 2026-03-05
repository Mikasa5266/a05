package model

import (
	"time"

	"gorm.io/gorm"
)

// ===== University Models =====

// StudentRecord represents a student tracked by the university
type StudentRecord struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	UniversityID     uint           `gorm:"index;not null" json:"university_id"`
	UserID           uint           `gorm:"index" json:"user_id"`
	Name             string         `gorm:"not null;size:100" json:"name"`
	StudentNo        string         `gorm:"size:50" json:"student_no"`
	Major            string         `gorm:"size:100" json:"major"`
	Grade            string         `gorm:"size:20" json:"grade"`
	RiskLevel        string         `gorm:"default:'low';size:20" json:"risk_level"` // high, medium, low
	InterviewCount   int            `gorm:"default:0" json:"interview_count"`
	AverageScore     int            `gorm:"default:0" json:"average_score"`
	EmploymentStatus string         `gorm:"default:'seeking';size:20" json:"employment_status"` // seeking, interviewing, offered, employed
	MentorID         *uint          `gorm:"index" json:"mentor_id"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

// Course represents an employment guidance course
type Course struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UniversityID uint           `gorm:"index" json:"university_id"`
	Title        string         `gorm:"not null;size:200" json:"title"`
	Category     string         `gorm:"size:50" json:"category"` // interview, resume, career
	StudentCount int            `gorm:"default:0" json:"student_count"`
	Duration     string         `gorm:"size:50" json:"duration"`
	Instructor   string         `gorm:"size:100" json:"instructor"`
	Description  string         `gorm:"type:text" json:"description"`
	CoverColor   string         `gorm:"size:50;default:'from-indigo-500 to-purple-500'" json:"cover_color"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TalentPush represents a student pushed to enterprise
type TalentPush struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	UniversityID   uint           `gorm:"index;not null" json:"university_id"`
	StudentID      uint           `gorm:"index;not null" json:"student_id"`
	EnterpriseName string         `gorm:"size:200" json:"enterprise_name"`
	Position       string         `gorm:"size:200" json:"position"`
	MatchScore     int            `gorm:"default:0" json:"match_score"`
	Status         string         `gorm:"default:'pending';size:20" json:"status"` // pending, accepted, rejected
	PushedAt       time.Time      `json:"pushed_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}
