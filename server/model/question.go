package model

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

type Question struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Title          string         `gorm:"not null" json:"title"`
	Content        string         `gorm:"type:text;not null" json:"content"`
	Position       string         `gorm:"index;not null" json:"position"`
	Difficulty     string         `gorm:"index;not null" json:"difficulty"`
	Category       string         `gorm:"index" json:"category"`
	Tags           string         `gorm:"type:text" json:"-"`
	ExpectedAnswer string         `gorm:"type:text" json:"expected_answer,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	InterviewQuestions []InterviewQuestion `gorm:"foreignKey:QuestionID" json:"-"`
}

func (q *Question) GetTags() []string {
	if q.Tags == "" {
		return []string{}
	}
	return strings.Split(q.Tags, ",")
}

func (q *Question) SetTags(tags []string) {
	q.Tags = strings.Join(tags, ",")
}
