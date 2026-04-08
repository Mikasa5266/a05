package model

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

type PositionCode string

const (
	PositionBackend   PositionCode = "backend"
	PositionFrontend  PositionCode = "frontend"
	PositionAlgorithm PositionCode = "algorithm"
	PositionAI        PositionCode = "ai"
)

type JobPosition struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Code        string         `gorm:"size:32;uniqueIndex;not null" json:"code"`
	Name        string         `gorm:"size:64;uniqueIndex;not null" json:"name"`
	Domain      string         `gorm:"size:64" json:"domain,omitempty"`
	Description string         `gorm:"type:text" json:"description,omitempty"`
	IsActive    bool           `gorm:"index;default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Questions []Question `gorm:"foreignKey:PositionCode;references:Code" json:"-"`
}

func (JobPosition) TableName() string {
	return "job_positions"
}

var DefaultJobPositions = []JobPosition{
	{Code: string(PositionBackend), Name: "Java后端工程师", Domain: "backend", IsActive: true},
	{Code: string(PositionFrontend), Name: "前端工程师", Domain: "frontend", IsActive: true},
	{Code: string(PositionAlgorithm), Name: "算法工程师", Domain: "algorithm", IsActive: true},
	{Code: string(PositionAI), Name: "AI工程师", Domain: "ai", IsActive: true},
}

type Question struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	PositionCode   string         `gorm:"size:32;index:idx_question_core,priority:1;not null;default:'backend'" json:"position_code"`
	Position       string         `gorm:"size:64;index:idx_question_core,priority:2;not null" json:"position"`
	Difficulty     string         `gorm:"size:32;index:idx_question_core,priority:3;not null" json:"difficulty"`
	DifficultyRank int            `gorm:"index;default:1" json:"difficulty_rank"`
	Category       string         `gorm:"size:64;index" json:"category"`
	KnowledgePoint string         `gorm:"size:128;index" json:"knowledge_point,omitempty"`
	KnowledgeArea  string         `gorm:"size:128;index" json:"knowledge_area,omitempty"`
	QuestionType   string         `gorm:"size:32;index;default:'technical'" json:"question_type"`
	Title          string         `gorm:"size:255;not null" json:"title"`
	Content        string         `gorm:"type:text;not null" json:"content"`
	ExpectedAnswer string         `gorm:"type:text" json:"expected_answer,omitempty"`
	Tags           string         `gorm:"type:text" json:"tags,omitempty"`
	MetadataJSON   string         `gorm:"type:longtext" json:"metadata_json,omitempty"`
	Source         string         `gorm:"size:32;index;default:'question_bank'" json:"source"`
	RAGEligible    bool           `gorm:"index;default:true" json:"rag_eligible"`
	IsActive       bool           `gorm:"index;default:true" json:"is_active"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	InterviewQuestions []InterviewQuestion `gorm:"foreignKey:QuestionID" json:"-"`
	PositionRef        *JobPosition        `gorm:"foreignKey:PositionCode;references:Code" json:"position_ref,omitempty"`
}

func (q *Question) GetTags() []string {
	if strings.TrimSpace(q.Tags) == "" {
		return []string{}
	}
	raw := strings.Split(q.Tags, ",")
	out := make([]string, 0, len(raw))
	seen := make(map[string]struct{}, len(raw))
	for _, item := range raw {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func (q *Question) SetTags(tags []string) {
	if len(tags) == 0 {
		q.Tags = ""
		return
	}
	normalized := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	q.Tags = strings.Join(normalized, ",")
}

func (q *Question) BeforeCreate(tx *gorm.DB) error {
	q.normalizeDerivedFields()
	return nil
}

func (q *Question) BeforeSave(tx *gorm.DB) error {
	q.normalizeDerivedFields()
	return nil
}

func (q *Question) normalizeDerivedFields() {
	if strings.TrimSpace(q.PositionCode) == "" {
		q.PositionCode = inferPositionCode(q.Position)
	}
	if strings.TrimSpace(q.PositionCode) == "" {
		q.PositionCode = string(PositionBackend)
	}

	if q.DifficultyRank <= 0 {
		q.DifficultyRank = inferDifficultyRank(q.Difficulty)
	}
	if q.DifficultyRank <= 0 {
		q.DifficultyRank = 1
	}
}

func inferPositionCode(position string) string {
	p := strings.ToLower(strings.TrimSpace(position))
	switch {
	case p == "", strings.Contains(p, "java"), strings.Contains(p, "后端"), p == "backend":
		return string(PositionBackend)
	case strings.Contains(p, "前端"), strings.Contains(p, "frontend"):
		return string(PositionFrontend)
	case strings.Contains(p, "算法"), p == "algorithm":
		return string(PositionAlgorithm)
	case strings.Contains(p, "ai"), strings.Contains(p, "llm"), strings.Contains(p, "模型"), p == "ai_engineer":
		return string(PositionAI)
	default:
		return string(PositionBackend)
	}
}

func inferDifficultyRank(difficulty string) int {
	d := strings.ToLower(strings.TrimSpace(difficulty))
	switch d {
	case "campus_intern", "easy", "junior", "初级", "基础":
		return 1
	case "campus_graduate", "medium", "intermediate", "中级":
		return 2
	case "social_junior", "hard", "senior", "中高级":
		return 3
	default:
		return 1
	}
}
