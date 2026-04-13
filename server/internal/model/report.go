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

type ReportAudioTranscript struct {
	SpeakerID uint   `json:"speaker_id,omitempty"`
	Content   string `json:"content"`
	StartMS   int64  `json:"start_ms,omitempty"`
	EndMS     int64  `json:"end_ms,omitempty"`
}

type ReportChatMessage struct {
	SenderID uint       `json:"sender_id"`
	Content  string     `json:"content"`
	SentAt   *time.Time `json:"sent_at,omitempty"`
}

type Report struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	UserID uint `gorm:"index;not null" json:"user_id"`
	// TODO: [LEGACY-TRASH] - Group Interview Refactor
	InterviewID      uint   `gorm:"uniqueIndex;not null" json:"interview_id"`
	Position         string `gorm:"not null" json:"position"`
	Difficulty       string `gorm:"not null" json:"difficulty"`
	TotalQuestions   int    `gorm:"not null" json:"total_questions"`
	AverageScore     int    `gorm:"not null" json:"average_score"`
	Strengths        string `gorm:"type:text" json:"-"`
	Weaknesses       string `gorm:"type:text" json:"-"`
	Suggestions      string `gorm:"type:text" json:"-"`
	QADetails        string `gorm:"column:qa_details;type:json" json:"-"`
	AudioTranscripts string `gorm:"column:audio_transcripts;type:json" json:"-"`
	ChatMessages     string `gorm:"column:chat_messages;type:json" json:"-"`
	SinglePlayback   string `gorm:"column:single_playback;size:500" json:"single_playback,omitempty"`
	MultiPlayback    string `gorm:"column:multi_playback;size:500" json:"multi_playback,omitempty"`
	OverallAnalysis  string `gorm:"type:text" json:"overall_analysis"`

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

func (r *Report) GetAudioTranscripts() []ReportAudioTranscript {
	if strings.TrimSpace(r.AudioTranscripts) == "" {
		return []ReportAudioTranscript{}
	}

	var transcripts []ReportAudioTranscript
	if err := json.Unmarshal([]byte(r.AudioTranscripts), &transcripts); err != nil {
		return []ReportAudioTranscript{}
	}

	return normalizeReportAudioTranscripts(transcripts)
}

func (r *Report) SetAudioTranscripts(transcripts []ReportAudioTranscript) {
	normalized := normalizeReportAudioTranscripts(transcripts)
	if len(normalized) == 0 {
		r.AudioTranscripts = "[]"
		return
	}

	raw, err := json.Marshal(normalized)
	if err != nil {
		r.AudioTranscripts = "[]"
		return
	}
	r.AudioTranscripts = string(raw)
}

func (r *Report) GetChatMessages() []ReportChatMessage {
	if strings.TrimSpace(r.ChatMessages) == "" {
		return []ReportChatMessage{}
	}

	var messages []ReportChatMessage
	if err := json.Unmarshal([]byte(r.ChatMessages), &messages); err != nil {
		return []ReportChatMessage{}
	}

	return normalizeReportChatMessages(messages)
}

func (r *Report) SetChatMessages(messages []ReportChatMessage) {
	normalized := normalizeReportChatMessages(messages)
	if len(normalized) == 0 {
		r.ChatMessages = "[]"
		return
	}

	raw, err := json.Marshal(normalized)
	if err != nil {
		r.ChatMessages = "[]"
		return
	}
	r.ChatMessages = string(raw)
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

func normalizeReportAudioTranscripts(transcripts []ReportAudioTranscript) []ReportAudioTranscript {
	normalized := make([]ReportAudioTranscript, 0, len(transcripts))
	for _, transcript := range transcripts {
		content := strings.TrimSpace(transcript.Content)
		if content == "" {
			continue
		}

		startMS := transcript.StartMS
		endMS := transcript.EndMS
		if startMS < 0 {
			startMS = 0
		}
		if endMS < 0 {
			endMS = 0
		}
		if endMS > 0 && endMS < startMS {
			endMS = startMS
		}

		normalized = append(normalized, ReportAudioTranscript{
			SpeakerID: transcript.SpeakerID,
			Content:   content,
			StartMS:   startMS,
			EndMS:     endMS,
		})

		if len(normalized) >= 2000 {
			break
		}
	}

	return normalized
}

func normalizeReportChatMessages(messages []ReportChatMessage) []ReportChatMessage {
	normalized := make([]ReportChatMessage, 0, len(messages))
	for _, message := range messages {
		content := strings.TrimSpace(message.Content)
		if message.SenderID == 0 || content == "" {
			continue
		}

		normalized = append(normalized, ReportChatMessage{
			SenderID: message.SenderID,
			Content:  content,
			SentAt:   message.SentAt,
		})

		if len(normalized) >= 5000 {
			break
		}
	}

	return normalized
}
