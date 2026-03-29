package repository

import (
	"strings"

	"your-project/model"

	"gorm.io/gorm"
)

type ResumeRepository interface {
	Create(record *model.ResumeRecord) error
	GetByID(id uint) (*model.ResumeRecord, error)
	GetLatestByUserID(userID uint) (*model.ResumeRecord, error)
	Update(record *model.ResumeRecord) error
	Delete(id uint) error

	ListDistinctQuestionPositions() ([]string, error)
	ListDistinctQuestionCategories() ([]string, error)
}

type GormResumeRepository struct {
	db *gorm.DB
}

var _ ResumeRepository = (*GormResumeRepository)(nil)

func NewResumeRepository() ResumeRepository {
	return &GormResumeRepository{db: GetDB()}
}

func NewResumeRepositoryWithDB(db *gorm.DB) ResumeRepository {
	if db == nil {
		db = GetDB()
	}
	return &GormResumeRepository{db: db}
}

func (r *GormResumeRepository) Create(record *model.ResumeRecord) error {
	return r.db.Create(record).Error
}

func (r *GormResumeRepository) GetByID(id uint) (*model.ResumeRecord, error) {
	var record model.ResumeRecord
	if err := r.db.First(&record, id).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *GormResumeRepository) GetLatestByUserID(userID uint) (*model.ResumeRecord, error) {
	var record model.ResumeRecord
	if err := r.db.Where("user_id = ?", userID).Order("created_at DESC").First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *GormResumeRepository) Update(record *model.ResumeRecord) error {
	return r.db.Save(record).Error
}

func (r *GormResumeRepository) Delete(id uint) error {
	return r.db.Delete(&model.ResumeRecord{}, id).Error
}

func (r *GormResumeRepository) ListDistinctQuestionPositions() ([]string, error) {
	positions := make([]string, 0)
	if err := r.db.Model(&model.Question{}).
		Where("position IS NOT NULL AND position <> ''").
		Distinct().
		Pluck("position", &positions).Error; err != nil {
		return nil, err
	}
	return normalizeDistinctStrings(positions), nil
}

func (r *GormResumeRepository) ListDistinctQuestionCategories() ([]string, error) {
	categories := make([]string, 0)
	if err := r.db.Model(&model.Question{}).
		Where("category IS NOT NULL AND category <> ''").
		Distinct().
		Pluck("category", &categories).Error; err != nil {
		return nil, err
	}
	return normalizeDistinctStrings(categories), nil
}

func normalizeDistinctStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, v := range values {
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}
