package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	UUID               string         `gorm:"type:char(36);uniqueIndex" json:"uuid,omitempty"`
	Username           string         `gorm:"uniqueIndex;not null;size:100" json:"username"`
	Email              string         `gorm:"uniqueIndex;not null;size:255" json:"email"`
	RealName           string         `gorm:"size:80" json:"real_name,omitempty"`
	Phone              string         `gorm:"size:20;index" json:"phone,omitempty"`
	IDCardHash         string         `gorm:"size:64;index" json:"-"`
	IDCardMasked       string         `gorm:"size:32" json:"id_card_masked,omitempty"`
	RealNameVerified   bool           `gorm:"default:false;index" json:"real_name_verified"`
	RealNameVerifiedAt *time.Time     `json:"real_name_verified_at,omitempty"`
	Password           string         `gorm:"not null;size:255" json:"-"`
	Role               string         `gorm:"default:student;size:50" json:"role"`
	Avatar             string         `json:"avatar,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`

	Interviews []Interview `gorm:"foreignKey:UserID" json:"interviews,omitempty"`
	Reports    []Report    `gorm:"foreignKey:UserID" json:"reports,omitempty"`
}
