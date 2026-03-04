package model

import (
	"time"

	"gorm.io/gorm"
)

type Interview struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"index;not null" json:"user_id"`
	Position     string         `gorm:"not null" json:"position"`
	Difficulty   string         `gorm:"not null" json:"difficulty"`
	Mode         string         `gorm:"default:'technical'" json:"mode"` // technical, hr, comprehensive
	Style        string         `gorm:"default:'gentle'" json:"style"`   // gentle, stress, deep
	Status       string         `gorm:"default:in_progress" json:"status"`
	StartTime    time.Time      `json:"start_time"`
	EndTime      *time.Time     `json:"end_time,omitempty"`
	CurrentIndex int            `gorm:"default:0" json:"current_index"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	User               User                `gorm:"foreignKey:UserID" json:"user,omitempty"`
	InterviewQuestions []InterviewQuestion `gorm:"foreignKey:InterviewID" json:"questions,omitempty"`
	AnswerResults      []AnswerResult      `gorm:"foreignKey:InterviewID" json:"answers,omitempty"`
	Report             *Report             `gorm:"foreignKey:InterviewID" json:"report,omitempty"`
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
