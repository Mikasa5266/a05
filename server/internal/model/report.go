package model

import (
	"encoding/json"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ReportQADetail struct {
	Question        string   `json:"question"`
	UserAnswer      string   `json:"user_answer"`
	OptimizedAnswer string   `json:"optimized_answer"`
	KeyImprovements []string `json:"key_improvements"`
}

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
	QADetails       string `gorm:"column:qa_details;type:json" json:"-"`
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

func (r *Report) GetQADetails() []ReportQADetail {
	if strings.TrimSpace(r.QADetails) == "" {
		return []ReportQADetail{}
	}

	var details []ReportQADetail
	if err := json.Unmarshal([]byte(r.QADetails), &details); err != nil {
		return []ReportQADetail{}
	}

	return normalizeReportQADetails(details)
}

func (r *Report) SetQADetails(details []ReportQADetail) {
	normalized := normalizeReportQADetails(details)
	if len(normalized) == 0 {
		r.QADetails = "[]"
		return
	}

	raw, err := json.Marshal(normalized)
	if err != nil {
		r.QADetails = "[]"
		return
	}
	r.QADetails = string(raw)
}

func normalizeReportQADetails(details []ReportQADetail) []ReportQADetail {
	normalized := make([]ReportQADetail, 0, len(details))
	for _, detail := range details {
		question := strings.TrimSpace(detail.Question)
		userAnswer := strings.TrimSpace(detail.UserAnswer)
		optimizedAnswer := strings.TrimSpace(detail.OptimizedAnswer)
		if question == "" || (userAnswer == "" && optimizedAnswer == "") {
			continue
		}

		if userAnswer == "" {
			userAnswer = "候选人回答摘要暂缺。"
		}
		if optimizedAnswer == "" {
			optimizedAnswer = "建议按“结论-原理-实践-边界”结构组织回答。"
		}

		improvements := make([]string, 0, len(detail.KeyImprovements))
		seen := make(map[string]struct{}, len(detail.KeyImprovements))
		for _, item := range detail.KeyImprovements {
			line := strings.TrimSpace(item)
			if line == "" {
				continue
			}
			if _, ok := seen[line]; ok {
				continue
			}
			seen[line] = struct{}{}
			improvements = append(improvements, line)
			if len(improvements) >= 4 {
				break
			}
		}
		if len(improvements) == 0 {
			improvements = []string{"补充关键机制说明", "增加边界条件与异常处理"}
		}

		normalized = append(normalized, ReportQADetail{
			Question:        question,
			UserAnswer:      userAnswer,
			OptimizedAnswer: optimizedAnswer,
			KeyImprovements: improvements,
		})

		if len(normalized) >= 12 {
			break
		}
	}

	return normalized
}
