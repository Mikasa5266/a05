package model

import (
	"encoding/json"
	"strings"
	"time"

	"gorm.io/gorm"
)

type PracticeQuestionFavorite struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"not null;index;uniqueIndex:idx_question_favorite_user_question,priority:1" json:"user_id"`
	QuestionID uint      `gorm:"not null;index;uniqueIndex:idx_question_favorite_user_question,priority:2" json:"question_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	Question Question `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
}

func (PracticeQuestionFavorite) TableName() string {
	return "question_favorites"
}

type PracticeWrongBookEntry struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UserID         uint      `gorm:"not null;index;uniqueIndex:idx_question_wrong_book_user_question,priority:1" json:"user_id"`
	QuestionID     uint      `gorm:"not null;index;uniqueIndex:idx_question_wrong_book_user_question,priority:2" json:"question_id"`
	LastUserAnswer string    `gorm:"type:longtext" json:"last_user_answer,omitempty"`
	ErrorReason    string    `gorm:"type:text" json:"error_reason,omitempty"`
	IsFavorite     bool      `gorm:"not null;default:false;index" json:"is_favorite"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	Question Question `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
}

func (PracticeWrongBookEntry) TableName() string {
	return "question_wrong_books"
}

type PracticeQuestionList struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	PositionCode string    `gorm:"size:32;not null;index" json:"position_code"`
	Title        string    `gorm:"size:255;not null" json:"title"`
	Description  string    `gorm:"type:text;not null" json:"description"`
	TagsJSON     string    `gorm:"column:tags_json;type:longtext" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	Tags  []string                   `gorm:"-" json:"tags,omitempty"`
	Items []PracticeQuestionListItem `gorm:"foreignKey:ListID" json:"items,omitempty"`
}

func (PracticeQuestionList) TableName() string {
	return "question_lists"
}

func (l *PracticeQuestionList) GetTags() []string {
	if len(l.Tags) > 0 {
		return append([]string(nil), l.Tags...)
	}
	if strings.TrimSpace(l.TagsJSON) == "" {
		return []string{}
	}

	var tags []string
	if err := json.Unmarshal([]byte(l.TagsJSON), &tags); err != nil {
		return []string{}
	}
	l.Tags = normalizePracticeTags(tags)
	return append([]string(nil), l.Tags...)
}

func (l *PracticeQuestionList) SetTags(tags []string) {
	l.Tags = normalizePracticeTags(tags)
	if len(l.Tags) == 0 {
		l.TagsJSON = ""
		return
	}

	raw, err := json.Marshal(l.Tags)
	if err != nil {
		l.TagsJSON = ""
		return
	}
	l.TagsJSON = string(raw)
}

func (l *PracticeQuestionList) AfterFind(tx *gorm.DB) error {
	l.Tags = l.GetTags()
	return nil
}

func (l *PracticeQuestionList) BeforeSave(tx *gorm.DB) error {
	l.PositionCode = strings.TrimSpace(strings.ToLower(l.PositionCode))
	l.Title = strings.TrimSpace(l.Title)
	l.Description = strings.TrimSpace(l.Description)
	l.SetTags(l.Tags)
	return nil
}

type PracticeQuestionListItem struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ListID     uint      `gorm:"not null;index;uniqueIndex:idx_question_list_item_unique,priority:1" json:"list_id"`
	QuestionID uint      `gorm:"not null;index;uniqueIndex:idx_question_list_item_unique,priority:2" json:"question_id"`
	OrderNo    int       `gorm:"not null;index" json:"order_no"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	Question Question `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
}

func (PracticeQuestionListItem) TableName() string {
	return "question_list_items"
}

type PracticeAssessmentAnswer struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	AssessmentID   uint      `gorm:"not null;index;uniqueIndex:idx_question_assessment_answer_unique,priority:1" json:"assessment_id"`
	QuestionID     uint      `gorm:"not null;index;uniqueIndex:idx_question_assessment_answer_unique,priority:2" json:"question_id"`
	UserID         uint      `gorm:"not null;index" json:"user_id"`
	UserAnswer     string    `gorm:"type:longtext" json:"user_answer,omitempty"`
	IsCorrect      bool      `gorm:"not null;index" json:"is_correct"`
	ElapsedSeconds *int      `json:"elapsed_seconds,omitempty"`
	IsTimeout      bool      `gorm:"not null;default:false" json:"is_timeout"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	Assessment QuestionAssessment `gorm:"foreignKey:AssessmentID" json:"-"`
	Question   Question           `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
}

func (PracticeAssessmentAnswer) TableName() string {
	return "question_assessment_answers"
}

type PracticeInterviewSyncLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      *uint     `gorm:"index" json:"user_id,omitempty"`
	Source      string    `gorm:"size:64;not null;index" json:"source"`
	PayloadJSON string    `gorm:"column:payload_json;type:longtext;not null" json:"payload_json"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (PracticeInterviewSyncLog) TableName() string {
	return "interview_sync_logs"
}

func normalizePracticeTags(tags []string) []string {
	if len(tags) == 0 {
		return []string{}
	}

	result := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		result = append(result, tag)
	}
	return result
}
