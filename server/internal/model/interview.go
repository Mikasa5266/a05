package model

import (
	"encoding/json"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	GroupInterviewRoomStatusWaiting    = "waiting"
	GroupInterviewRoomStatusInProgress = "in_progress"
	GroupInterviewRoomStatusFinished   = "finished"
)

type GroupInterviewRoom struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	RoomID    string `gorm:"size:64;uniqueIndex;not null" json:"room_id"`
	CreatorID uint   `gorm:"index;not null" json:"creator_id"`
	// TODO: [LEGACY-TRASH] - Group Interview Refactor
	ParticipantIDs string         `gorm:"column:participant_ids;type:json" json:"participant_ids"`
	Status         string         `gorm:"size:20;default:'waiting';index" json:"status"` // waiting, in_progress, finished
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	Creator    User        `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
	Interviews []Interview `gorm:"foreignKey:GroupInterviewRoomID" json:"interviews,omitempty"`
}

func (r *GroupInterviewRoom) GetParticipantIDs() []uint {
	if strings.TrimSpace(r.ParticipantIDs) == "" {
		return []uint{}
	}

	var ids []uint
	if err := json.Unmarshal([]byte(r.ParticipantIDs), &ids); err != nil {
		return []uint{}
	}

	return normalizeParticipantIDs(ids)
}

func (r *GroupInterviewRoom) SetParticipantIDs(ids []uint) {
	normalized := normalizeParticipantIDs(ids)
	if len(normalized) == 0 {
		r.ParticipantIDs = "[]"
		return
	}

	raw, err := json.Marshal(normalized)
	if err != nil {
		r.ParticipantIDs = "[]"
		return
	}
	r.ParticipantIDs = string(raw)
}

func normalizeParticipantIDs(ids []uint) []uint {
	if len(ids) == 0 {
		return []uint{}
	}

	normalized := make([]uint, 0, len(ids))
	seen := make(map[uint]struct{}, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		normalized = append(normalized, id)
		if len(normalized) >= 5 {
			break
		}
	}

	return normalized
}

type Interview struct {
	ID uint `gorm:"primaryKey" json:"id"`
	// TODO: [LEGACY-TRASH] - Group Interview Refactor
	UserID                 uint           `gorm:"index;not null" json:"user_id"`
	IsGroup                bool           `gorm:"default:false;index" json:"is_group"`
	GroupInterviewRoomID   *uint          `gorm:"index" json:"group_interview_room_id,omitempty"`
	Position               string         `gorm:"not null" json:"position"`
	Difficulty             string         `gorm:"not null" json:"difficulty"`
	Mode                   string         `gorm:"default:'technical'" json:"mode"`    // technical, hr, comprehensive
	Style                  string         `gorm:"default:'gentle'" json:"style"`      // gentle, stress, deep, practical, algorithm
	Company                string         `gorm:"default:''" json:"company"`          // ali, bytedance, tencent, meituan, baidu, or empty
	InterviewMode          string         `gorm:"default:'ai'" json:"interview_mode"` // ai, human, random
	InvitationCode         *string        `gorm:"size:64;uniqueIndex" json:"invitation_code,omitempty"`
	Role                   string         `gorm:"size:20;default:'candidate'" json:"role"`     // candidate, interviewer
	Scenario               string         `gorm:"type:text" json:"scenario,omitempty"`         // blindbox scenario JSON
	RevealedStyle          string         `gorm:"default:''" json:"revealed_style,omitempty"`  // For random mode: the actual style used (revealed after interview)
	HumanInterviewerID     *uint          `gorm:"index" json:"human_interviewer_id,omitempty"` // For human interview mode
	HumanInterviewerUserID *uint          `gorm:"index" json:"human_interviewer_user_id,omitempty"`
	HumanInterviewerName   string         `gorm:"size:120" json:"human_interviewer_name,omitempty"`
	HumanInterviewerRole   string         `gorm:"size:50" json:"human_interviewer_role,omitempty"`
	HumanFeedback          string         `gorm:"type:text" json:"human_feedback,omitempty"` // Human interviewer notes
	HumanScore             *int           `json:"human_score,omitempty"`                     // Human interviewer score
	RecordingURL           string         `gorm:"size:500" json:"recording_url,omitempty"`
	RecordingStatus        string         `gorm:"default:'none';size:20" json:"recording_status,omitempty"`
	ASRCallCount           int            `gorm:"default:0" json:"asr_call_count,omitempty"`
	TTSCharCount           int            `gorm:"default:0" json:"tts_char_count,omitempty"`
	Status                 string         `gorm:"default:'pending';size:20" json:"status"` // pending, in_progress, completed
	StartTime              time.Time      `json:"start_time"`
	EndTime                *time.Time     `json:"end_time,omitempty"`
	CurrentIndex           int            `gorm:"default:0" json:"current_index"`
	AskedQuestionIDs       string         `gorm:"type:text" json:"asked_question_ids,omitempty"`
	CurrentTopic           string         `gorm:"default:''" json:"current_topic,omitempty"` // Current interview topic (e.g. "Project Experience")
	FollowUpCount          int            `gorm:"default:0" json:"follow_up_count"`          // Number of follow-ups asked for current topic
	MaxFollowUps           int            `gorm:"default:3" json:"max_follow_ups"`           // Max follow-ups per topic
	TopicIndex             int            `gorm:"default:0" json:"topic_index"`              // Current topic index (0-based)
	TopicCountTarget       int            `gorm:"default:0" json:"topic_count_target"`       // Planned number of topics
	TopicQuestionMin       int            `gorm:"default:2" json:"topic_question_min"`       // Min questions per topic
	TopicQuestionMax       int            `gorm:"default:4" json:"topic_question_max"`       // Max questions per topic
	TotalQuestionTarget    int            `gorm:"default:0" json:"total_question_target"`    // Planned total questions for interview
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
	DeletedAt              gorm.DeletedAt `gorm:"index" json:"-"`

	User               User                `gorm:"foreignKey:UserID" json:"user,omitempty"`
	GroupInterviewRoom *GroupInterviewRoom `gorm:"foreignKey:GroupInterviewRoomID" json:"group_interview_room,omitempty"`
	InterviewQuestions []InterviewQuestion `gorm:"foreignKey:InterviewID" json:"questions,omitempty"`
	AnswerResults      []AnswerResult      `gorm:"foreignKey:InterviewID" json:"answers,omitempty"`
	Report             *Report             `gorm:"foreignKey:InterviewID" json:"report,omitempty"`
}

// HumanInterviewInvitation is a student-created invitation sent to university/enterprise users.
type HumanInterviewInvitation struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	InvitationCode string     `gorm:"size:64;uniqueIndex;not null" json:"invitation_code"`
	StudentID      uint       `gorm:"index;not null" json:"student_id"`
	StudentUUID    string     `gorm:"type:char(36);index" json:"student_uuid"`
	InviteeUserID  uint       `gorm:"index;not null" json:"invitee_user_id"`
	InviteeUUID    string     `gorm:"type:char(36);index" json:"invitee_uuid"`
	InviteeRole    string     `gorm:"size:50;not null" json:"invitee_role"` // university, enterprise
	Position       string     `gorm:"size:120;not null" json:"position"`
	Difficulty     string     `gorm:"size:50;not null" json:"difficulty"`
	Mode           string     `gorm:"size:50;not null" json:"mode"`
	Style          string     `gorm:"size:50;not null" json:"style"`
	Company        string     `gorm:"size:50" json:"company,omitempty"`
	Status         string     `gorm:"size:20;default:'pending'" json:"status"` // pending, accepted, rejected, in_progress, completed, cancelled
	ScheduledAt    *time.Time `json:"scheduled_at,omitempty"`
	Notes          string     `gorm:"type:text" json:"notes,omitempty"`
	InterviewID    *uint      `gorm:"index" json:"interview_id,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	Student User `gorm:"foreignKey:StudentID" json:"student,omitempty"`
	Invitee User `gorm:"foreignKey:InviteeUserID" json:"invitee,omitempty"`
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

	NextQuestion       *Question `gorm:"-" json:"next_question,omitempty"`
	InterviewCompleted bool      `gorm:"-" json:"interview_completed,omitempty"`
}
