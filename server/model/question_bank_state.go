package model

import (
	"time"

	"gorm.io/gorm"
)

// UserQuestionState stores per-user question interaction state for favorites and mistakes.
type UserQuestionState struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	UserID         uint           `gorm:"not null;uniqueIndex:idx_user_question_state,priority:1;index" json:"user_id"`
	QuestionID     uint           `gorm:"not null;uniqueIndex:idx_user_question_state,priority:2;index" json:"question_id"`
	IsFavorite     bool           `gorm:"default:false;index" json:"is_favorite"`
	IsWrong        bool           `gorm:"default:false;index" json:"is_wrong"`
	WrongCount     int            `gorm:"default:0" json:"wrong_count"`
	LastWrongAt    *time.Time     `json:"last_wrong_at,omitempty"`
	LastAnsweredAt *time.Time     `json:"last_answered_at,omitempty"`
	Note           string         `gorm:"type:text" json:"note,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	Question Question `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
}

func (UserQuestionState) TableName() string {
	return "user_question_states"
}
