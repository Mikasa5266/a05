package repository

import (
	"your-project/internal/model"

	"gorm.io/gorm"
)

type ReportWriteRepository interface {
	UpsertByInterview(report *model.Report) error
	Update(report *model.Report) error
	Delete(id uint) error

	// Backward compatible aliases.
	Create(report *model.Report) error
}

type ReportQueryRepository interface {
	GetByID(id uint) (*model.Report, error)
	GetByInterviewID(interviewID uint) (*model.Report, error)
	ListByUserPaged(userID uint, page, pageSize int) ([]*model.Report, int64, error)
	ListByUserTimeline(userID uint) ([]*model.Report, error)

	// Backward compatible aliases.
	GetByUserID(userID uint, page, pageSize int) ([]*model.Report, int64, error)
	GetAllByUserID(userID uint) ([]*model.Report, error)
}

type ReportStatsRepository interface {
	GetUserReportStats(userID uint) (*ReportStats, error)

	// Backward compatible alias.
	GetReportStats(userID uint) (map[string]interface{}, error)
}

type ReportRepository interface {
	ReportWriteRepository
	ReportQueryRepository
	ReportStatsRepository
}

type GormReportRepository struct {
	db *gorm.DB
}

var _ ReportRepository = (*GormReportRepository)(nil)

func NewReportRepository() ReportRepository {
	return &GormReportRepository{db: GetDB()}
}

func NewReportRepositoryWithDB(db *gorm.DB) ReportRepository {
	if db == nil {
		db = GetDB()
	}
	return &GormReportRepository{db: db}
}
