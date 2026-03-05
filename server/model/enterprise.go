package model

import (
	"time"

	"gorm.io/gorm"
)

// ===== Enterprise Models =====

// Job represents a job posting by an enterprise
type Job struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	EnterpriseID      uint           `gorm:"index;not null" json:"enterprise_id"`
	Title             string         `gorm:"not null;size:200" json:"title"`
	Department        string         `gorm:"size:100" json:"department"`
	Location          string         `gorm:"size:100" json:"location"`
	SalaryRange       string         `gorm:"size:100" json:"salary_range"`
	Description       string         `gorm:"type:text" json:"description"`
	Requirements      string         `gorm:"type:text" json:"requirements"`
	Dimensions        int            `gorm:"default:4" json:"dimensions"`
	Status            string         `gorm:"default:'active';size:20" json:"status"` // active, paused, closed
	TechnicalWeight   int            `gorm:"default:30" json:"technical_weight"`
	CommunicateWeight int            `gorm:"default:25" json:"communicate_weight"`
	LogicWeight       int            `gorm:"default:25" json:"logic_weight"`
	MatchWeight       int            `gorm:"default:20" json:"match_weight"`
	CapabilityGraph   string         `gorm:"type:text" json:"capability_graph"` // JSON string for capability graph structure
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

// JobCapabilityDimension represents a dimension in the capability graph
type JobCapabilityDimension struct {
	Name        string `json:"name"`        // e.g. "后端开发"
	Weight      int    `json:"weight"`      // e.g. 30
	Description string `json:"description"` // e.g. "JVM原理、分布式系统"
	SubDimensions []JobCapabilitySubDimension `json:"sub_dimensions"`
}

type JobCapabilitySubDimension struct {
	Name   string `json:"name"`   // e.g. "JVM原理"
	Weight int    `json:"weight"` // e.g. 30 (relative to parent)
	Tags   []string `json:"tags"` // Keywords for question generation
}

// TalentRecord represents talent in the enterprise pool
type TalentRecord struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	EnterpriseID     uint           `gorm:"index;not null" json:"enterprise_id"`
	UserID           uint           `gorm:"index" json:"user_id"`
	Name             string         `gorm:"not null;size:100" json:"name"`
	Position         string         `gorm:"size:100" json:"position"`
	School           string         `gorm:"size:200" json:"school"`
	MatchScore       int            `gorm:"default:0" json:"match_score"`
	TechnicalScore   int            `gorm:"default:0" json:"technical_score"`
	CommunicateScore int            `gorm:"default:0" json:"communicate_score"`
	LogicScore       int            `gorm:"default:0" json:"logic_score"`
	Tags             string         `gorm:"type:text" json:"tags"`
	Status           string         `gorm:"default:'available';size:20" json:"status"` // available, invited, hired
	IsSaved          bool           `gorm:"default:false" json:"is_saved"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

// InterviewSession represents an HR interview session
type InterviewSession struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	EnterpriseID  uint           `gorm:"index;not null" json:"enterprise_id"`
	CandidateName string         `gorm:"size:100" json:"candidate_name"`
	Position      string         `gorm:"size:200" json:"position"`
	Status        string         `gorm:"default:'pending';size:20" json:"status"` // pending, live, completed
	ScheduledAt   *time.Time     `json:"scheduled_at"`
	ScenarioType  string         `gorm:"size:50" json:"scenario_type"` // written, behavioral, comprehensive
	Score         int            `gorm:"default:0" json:"score"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// CapabilityStandard represents co-created capability standards
type CapabilityStandard struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	EnterpriseID       uint           `gorm:"index" json:"enterprise_id"`
	Title              string         `gorm:"not null;size:200" json:"title"`
	Industry           string         `gorm:"size:100" json:"industry"`
	TechnicalWeight    int            `gorm:"default:30" json:"technical_weight"`
	CommunicateWeight  int            `gorm:"default:25" json:"communicate_weight"`
	LogicWeight        int            `gorm:"default:25" json:"logic_weight"`
	ProfessionalWeight int            `gorm:"default:20" json:"professional_weight"`
	SyncedUniversities int            `gorm:"default:0" json:"synced_universities"`
	Version            string         `gorm:"size:20" json:"version"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

// Referral represents internal referral channel
type Referral struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	EnterpriseID  uint           `gorm:"index;not null" json:"enterprise_id"`
	ReferrerName  string         `gorm:"size:100" json:"referrer_name"`
	CandidateName string         `gorm:"size:100" json:"candidate_name"`
	Position      string         `gorm:"size:200" json:"position"`
	Status        string         `gorm:"default:'pending';size:20" json:"status"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}
