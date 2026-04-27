package model

import (
	"time"

	"gorm.io/gorm"
)

// SecurityReport stores user-submitted reports for illegal/harmful content or abuse risks.
type SecurityReport struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	ReporterUserID uint           `gorm:"index;not null" json:"reporter_user_id"`
	ReporterRole   string         `gorm:"size:50" json:"reporter_role"`
	TargetType     string         `gorm:"size:30;index;not null" json:"target_type"`
	TargetID       uint           `gorm:"index" json:"target_id"`
	Reason         string         `gorm:"size:120;not null" json:"reason"`
	Description    string         `gorm:"type:text" json:"description"`
	Contact        string         `gorm:"size:120" json:"contact"`
	EvidenceJSON   string         `gorm:"type:longtext" json:"-"`
	Status         string         `gorm:"size:20;default:'pending';index" json:"status"`
	HandleNote     string         `gorm:"type:text" json:"handle_note"`
	HandledBy      *uint          `gorm:"index" json:"handled_by"`
	HandledAt      *time.Time     `json:"handled_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// SecurityAuditLog records key security operations for tracing and compliance support.
type SecurityAuditLog struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	UserID            uint      `gorm:"index" json:"user_id"`
	Action            string    `gorm:"size:80;index" json:"action"`
	Outcome           string    `gorm:"size:20;index" json:"outcome"`
	Method            string    `gorm:"size:10" json:"method"`
	Path              string    `gorm:"size:255" json:"path"`
	StatusCode        int       `json:"status_code"`
	SourceIP          string    `gorm:"size:255" json:"source_ip"`
	SourcePort        string    `gorm:"size:255" json:"source_port"`
	TargetHost        string    `gorm:"size:255" json:"target_host"`
	TargetPort        string    `gorm:"size:255" json:"target_port"`
	ClientFingerprint string    `gorm:"size:255" json:"client_fingerprint"`
	DetailJSON        string    `gorm:"type:longtext" json:"detail"`
	CreatedAt         time.Time `gorm:"index" json:"created_at"`
}

// AuditLog records moderation and disposal operations as closed-loop evidence.
type AuditLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ActorType  string    `gorm:"size:30;index" json:"actor_type"`
	ActorName  string    `gorm:"size:100;index" json:"actor_name"`
	Action     string    `gorm:"size:100;index" json:"action"`
	Outcome    string    `gorm:"size:20;index" json:"outcome"`
	Method     string    `gorm:"size:10" json:"method"`
	Path       string    `gorm:"size:255" json:"path"`
	TargetType string    `gorm:"size:30;index" json:"target_type"`
	TargetID   uint      `gorm:"index" json:"target_id"`
	SourceIP   string    `gorm:"size:255" json:"source_ip"`
	DetailJSON string    `gorm:"type:longtext" json:"detail"`
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
