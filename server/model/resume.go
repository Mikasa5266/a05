package model

import (
	"time"

	"gorm.io/gorm"
)

// ResumeParseResult stores structured parsing outputs and matching metadata for resume uploads.
type ResumeParseResult struct {
	ID                   uint           `gorm:"primaryKey" json:"id"`
	UserID               uint           `gorm:"index;not null" json:"user_id"`
	FileName             string         `gorm:"size:255;not null" json:"file_name"`
	FileHash             string         `gorm:"size:64;index" json:"file_hash,omitempty"`
	RawText              string         `gorm:"type:longtext" json:"raw_text,omitempty"`
	StructuredJSON       string         `gorm:"type:longtext;not null" json:"structured_json"`
	MatchedPositionCode  string         `gorm:"size:32;index" json:"matched_position_code,omitempty"`
	MatchedPositionName  string         `gorm:"size:64" json:"matched_position_name,omitempty"`
	MatchedQuestionIDs   string         `gorm:"type:text" json:"matched_question_ids,omitempty"`
	MatchedKnowledgeJSON string         `gorm:"type:longtext" json:"matched_knowledge_json,omitempty"`
	ConfidenceScore      int            `gorm:"default:0" json:"confidence_score"`
	ParserVersion        string         `gorm:"size:32;default:'v1'" json:"parser_version"`
	Source               string         `gorm:"size:32;default:'upload'" json:"source"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ResumeParseResult) TableName() string {
	return "resume_parse_results"
}
