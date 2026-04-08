package repository

import (
	"strings"

	"your-project/model"

	"gorm.io/gorm"
)

// ResumeParseResultRepository defines persistence operations for parsed resume results.
type ResumeParseResultRepository interface {
	Create(result *model.ResumeParseResult) error
	GetByID(id uint) (*model.ResumeParseResult, error)
	GetLatestByUser(userID uint) (*model.ResumeParseResult, error)
	ListByUser(userID uint, page, pageSize int) ([]*model.ResumeParseResult, int64, error)
	ListByMatchedPosition(positionCode string, limit int) ([]*model.ResumeParseResult, error)
	Update(result *model.ResumeParseResult) error
	Delete(id uint) error
}

type GormResumeParseResultRepository struct {
	db *gorm.DB
}

var _ ResumeParseResultRepository = (*GormResumeParseResultRepository)(nil)

func NewResumeParseResultRepository() ResumeParseResultRepository {
	return &GormResumeParseResultRepository{db: GetDB()}
}

func NewResumeParseResultRepositoryWithDB(db *gorm.DB) ResumeParseResultRepository {
	if db == nil {
		db = GetDB()
	}
	return &GormResumeParseResultRepository{db: db}
}

func (r *GormResumeParseResultRepository) Create(result *model.ResumeParseResult) error {
	return r.db.Create(result).Error
}

func (r *GormResumeParseResultRepository) GetByID(id uint) (*model.ResumeParseResult, error) {
	var result model.ResumeParseResult
	err := r.db.First(&result, id).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *GormResumeParseResultRepository) GetLatestByUser(userID uint) (*model.ResumeParseResult, error) {
	var result model.ResumeParseResult
	err := r.db.Where("user_id = ?", userID).Order("id DESC").First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *GormResumeParseResultRepository) ListByUser(userID uint, page, pageSize int) ([]*model.ResumeParseResult, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	var results []*model.ResumeParseResult
	var total int64
	offset := (page - 1) * pageSize

	query := r.db.Model(&model.ResumeParseResult{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("id DESC").Limit(pageSize).Offset(offset).Find(&results).Error
	if err != nil {
		return nil, 0, err
	}
	return results, total, nil
}

func (r *GormResumeParseResultRepository) ListByMatchedPosition(positionCode string, limit int) ([]*model.ResumeParseResult, error) {
	if limit <= 0 {
		limit = 20
	}

	var results []*model.ResumeParseResult
	query := r.db.Model(&model.ResumeParseResult{})
	if trimmed := strings.TrimSpace(positionCode); trimmed != "" {
		query = query.Where("matched_position_code = ?", trimmed)
	}

	err := query.Order("id DESC").Limit(limit).Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (r *GormResumeParseResultRepository) Update(result *model.ResumeParseResult) error {
	return r.db.Save(result).Error
}

func (r *GormResumeParseResultRepository) Delete(id uint) error {
	return r.db.Delete(&model.ResumeParseResult{}, id).Error
}
