package model

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

type Report struct {
	ID              uint   `gorm:"primaryKey" json:"id"`
	UserID          uint   `gorm:"index;not null" json:"user_id"`
	InterviewID     uint   `gorm:"uniqueIndex;not null" json:"interview_id"`
	Position        string `gorm:"not null" json:"position"`
	Difficulty      string `gorm:"not null" json:"difficulty"`
	TotalQuestions  int    `gorm:"not null" json:"total_questions"`
	AverageScore    int    `gorm:"not null" json:"average_score"`
	Strengths       string `gorm:"type:text" json:"-"`
	Weaknesses      string `gorm:"type:text" json:"-"`
	Suggestions     string `gorm:"type:text" json:"-"`
	OverallAnalysis string `gorm:"type:text" json:"overall_analysis"`

	// New fields for Radar Chart
	TechnicalScore  int `gorm:"default:0" json:"technical_score"`
	ExpressionScore int `gorm:"default:0" json:"expression_score"`
	LogicScore      int `gorm:"default:0" json:"logic_score"`
	MatchingScore   int `gorm:"default:0" json:"matching_score"`
	BehaviorScore   int `gorm:"default:0" json:"behavior_score"`

	StartTime time.Time      `json:"start_time"`
	EndTime   time.Time      `json:"end_time"`
	Duration  int            `gorm:"not null" json:"duration"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Interview Interview `gorm:"foreignKey:InterviewID" json:"interview,omitempty"`
}

func (r *Report) GetStrengths() []string {
	if r.Strengths == "" {
		return []string{}
	}
	return strings.Split(r.Strengths, "|")
}

func (r *Report) SetStrengths(strengths []string) {
	r.Strengths = strings.Join(strengths, "|")
}

func (r *Report) GetWeaknesses() []string {
	if r.Weaknesses == "" {
		return []string{}
	}
	return strings.Split(r.Weaknesses, "|")
}

func (r *Report) SetWeaknesses(weaknesses []string) {
	r.Weaknesses = strings.Join(weaknesses, "|")
}

func (r *Report) GetSuggestions() []string {
	if r.Suggestions == "" {
		return []string{}
	}
	return strings.Split(r.Suggestions, "|")
}

func (r *Report) SetSuggestions(suggestions []string) {
	r.Suggestions = strings.Join(suggestions, "|")
}
