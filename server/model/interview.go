package model

import (
	"time"

	"gorm.io/gorm"
)

type Interview struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	UserID             uint           `gorm:"index;not null" json:"user_id"`
	Position           string         `gorm:"not null" json:"position"`
	Difficulty         string         `gorm:"not null" json:"difficulty"`
	Mode               string         `gorm:"default:'technical'" json:"mode"`             // technical, hr, comprehensive
	Style              string         `gorm:"default:'gentle'" json:"style"`               // gentle, stress, deep, practical, algorithm
	Company            string         `gorm:"default:''" json:"company"`                   // ali, bytedance, tencent, meituan, baidu, or empty
	InterviewMode      string         `gorm:"default:'ai'" json:"interview_mode"`          // ai, human, random
	Scenario           string         `gorm:"type:text" json:"scenario,omitempty"`         // blindbox scenario JSON
	RevealedStyle      string         `gorm:"default:''" json:"revealed_style,omitempty"`  // For random mode: the actual style used (revealed after interview)
	HumanInterviewerID *uint          `gorm:"index" json:"human_interviewer_id,omitempty"` // For human interview mode
	HumanFeedback      string         `gorm:"type:text" json:"human_feedback,omitempty"`   // Human interviewer notes
	HumanScore         *int           `json:"human_score,omitempty"`                       // Human interviewer score
	Status             string         `gorm:"default:in_progress" json:"status"`
	StartTime          time.Time      `json:"start_time"`
	EndTime            *time.Time     `json:"end_time,omitempty"`
	CurrentIndex       int            `gorm:"default:0" json:"current_index"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`

	User               User                `gorm:"foreignKey:UserID" json:"user,omitempty"`
	InterviewQuestions []InterviewQuestion `gorm:"foreignKey:InterviewID" json:"questions,omitempty"`
	AnswerResults      []AnswerResult      `gorm:"foreignKey:InterviewID" json:"answers,omitempty"`
	Report             *Report             `gorm:"foreignKey:InterviewID" json:"report,omitempty"`
}

// HumanInterviewer represents an available human interviewer (teacher/enterprise expert)
type HumanInterviewer struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Name          string         `gorm:"not null" json:"name"`
	Title         string         `json:"title"`                        // e.g., "辅导员", "Java高级工程师"
	Type          string         `gorm:"not null;index" json:"type"`   // campus, enterprise
	Company       string         `json:"company,omitempty"`            // Enterprise name if type=enterprise
	Department    string         `json:"department,omitempty"`         // Department or school
	Specialties   string         `gorm:"type:text" json:"specialties"` // Comma-separated specialties
	AvatarURL     string         `json:"avatar_url,omitempty"`
	Available     bool           `gorm:"default:true" json:"available"`
	Rating        float64        `gorm:"default:5.0" json:"rating"`
	TotalSessions int            `gorm:"default:0" json:"total_sessions"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// InterviewBooking represents a booking for human interview
type InterviewBooking struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"index;not null" json:"user_id"`
	InterviewerID uint      `gorm:"index;not null" json:"interviewer_id"`
	InterviewID   *uint     `gorm:"index" json:"interview_id,omitempty"` // Linked after interview starts
	ScheduledAt   time.Time `gorm:"not null" json:"scheduled_at"`
	Duration      int       `gorm:"default:30" json:"duration"` // minutes
	Position      string    `json:"position"`
	Difficulty    string    `json:"difficulty"`
	Status        string    `gorm:"default:'pending'" json:"status"` // pending, confirmed, completed, cancelled
	Notes         string    `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	User        User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Interviewer HumanInterviewer `gorm:"foreignKey:InterviewerID" json:"interviewer,omitempty"`
}

type InterviewQuestion struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	InterviewID uint      `gorm:"index;not null" json:"interview_id"`
	QuestionID  uint      `gorm:"index;not null" json:"question_id"`
	OrderIndex  int       `gorm:"not null" json:"order_index"`
	IsAnswered  bool      `gorm:"default:false" json:"is_answered"`
	CreatedAt   time.Time `json:"created_at"`

	Interview Interview `gorm:"foreignKey:InterviewID" json:"-"`
	Question  Question  `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
}

type AnswerResult struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	InterviewID uint      `gorm:"index;not null" json:"interview_id"`
	QuestionID  uint      `gorm:"index;not null" json:"question_id"`
	Answer      string    `gorm:"type:text;not null" json:"answer"`
	Score       int       `gorm:"not null" json:"score"`
	Feedback    string    `gorm:"type:text" json:"feedback"`
	CreatedAt   time.Time `json:"created_at"`

	Interview Interview `gorm:"foreignKey:InterviewID" json:"-"`
	Question  Question  `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
}
