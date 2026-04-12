package model

import (
	"time"

	"gorm.io/gorm"
)

// ===== Community Models =====

// CommunityPost represents a post in the alumni community
type CommunityPost struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"index;not null" json:"user_id"`
	Author    string         `gorm:"size:100" json:"author"`
	Avatar    string         `gorm:"size:255" json:"avatar"`
	Company   string         `gorm:"size:200" json:"company"`
	Position  string         `gorm:"size:200" json:"position"`
	Title     string         `gorm:"not null;size:300" json:"title"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	Process   string         `gorm:"type:text" json:"process"`
	Questions string         `gorm:"type:text" json:"questions"`
	Review    string         `gorm:"type:text" json:"review"`
	Difficulty int           `gorm:"default:3" json:"difficulty"` // 1-5
	OfferStatus string       `gorm:"size:50" json:"offer_status"` // Pending, Received, Rejected
	Rounds    int            `gorm:"default:1" json:"rounds"`
	InterviewDate *time.Time `json:"interview_date"`
	Tags      string         `gorm:"type:text" json:"tags"`
	IsIndexed bool           `gorm:"default:false" json:"is_indexed"`
	Likes     int            `gorm:"default:0" json:"likes"`
	Comments  int            `gorm:"default:0" json:"comments"`
	Views     int            `gorm:"default:0" json:"views"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// PostComment represents a comment on a post
type PostComment struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	PostID    uint           `gorm:"index;not null" json:"post_id"`
	UserID    uint           `gorm:"index;not null" json:"user_id"`
	Author    string         `gorm:"size:100" json:"author"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// MentorBooking represents a 1-on-1 booking with a mentor
type MentorBooking struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	MentorID  uint           `gorm:"index;not null" json:"mentor_id"`
	UserID    uint           `gorm:"index;not null" json:"user_id"`
	Topic     string         `gorm:"size:300" json:"topic"`
	Message   string         `gorm:"type:text" json:"message"`
	Status    string         `gorm:"default:'pending';size:20" json:"status"` // pending, confirmed, completed, cancelled
	BookedAt  *time.Time     `json:"booked_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// PostLike tracks which users liked which posts
type PostLike struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PostID    uint      `gorm:"index;not null" json:"post_id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}
