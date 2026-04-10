package repository

import (
	"your-project/model"

	"gorm.io/gorm"
)

type ResumeParseResultRepository interface {
	Create(result *model.ResumeParseResult) error
	GetByID(id uint) (*model.ResumeParseResult, error)
	GetLatestByUser(userID uint) (*model.ResumeParseResult, error)
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
	if err := r.db.First(&result, id).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *GormResumeParseResultRepository) GetLatestByUser(userID uint) (*model.ResumeParseResult, error) {
	var result model.ResumeParseResult
	if err := r.db.Where("user_id = ?", userID).Order("id DESC").First(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}
