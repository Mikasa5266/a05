package repository

import (
	"strings"

	"your-project/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PositionRepository defines persistence operations for interview job positions.
type PositionRepository interface {
	Create(position *model.JobPosition) error
	GetByCode(code string) (*model.JobPosition, error)
	ListActive() ([]model.JobPosition, error)
	UpsertDefaults(defaults []model.JobPosition) error
	EnsureDefaults() error
	Update(position *model.JobPosition) error
	DeleteByCode(code string) error
}

type GormPositionRepository struct {
	db *gorm.DB
}

var _ PositionRepository = (*GormPositionRepository)(nil)

func NewPositionRepository() PositionRepository {
	return &GormPositionRepository{db: GetDB()}
}

func NewPositionRepositoryWithDB(db *gorm.DB) PositionRepository {
	if db == nil {
		db = GetDB()
	}
	return &GormPositionRepository{db: db}
}

func (r *GormPositionRepository) Create(position *model.JobPosition) error {
	return r.db.Create(position).Error
}

func (r *GormPositionRepository) GetByCode(code string) (*model.JobPosition, error) {
	var position model.JobPosition
	err := r.db.Where("code = ?", strings.TrimSpace(code)).First(&position).Error
	if err != nil {
		return nil, err
	}
	return &position, nil
}

func (r *GormPositionRepository) ListActive() ([]model.JobPosition, error) {
	var positions []model.JobPosition
	err := r.db.Where("is_active = ?", true).Order("code ASC").Find(&positions).Error
	if err != nil {
		return nil, err
	}
	return positions, nil
}

func (r *GormPositionRepository) UpsertDefaults(defaults []model.JobPosition) error {
	if len(defaults) == 0 {
		return nil
	}
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "code"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"name",
			"domain",
			"description",
			"is_active",
			"updated_at",
		}),
	}).Create(&defaults).Error
}

func (r *GormPositionRepository) EnsureDefaults() error {
	return r.UpsertDefaults(model.DefaultJobPositions)
}

func (r *GormPositionRepository) Update(position *model.JobPosition) error {
	return r.db.Save(position).Error
}

func (r *GormPositionRepository) DeleteByCode(code string) error {
	return r.db.Where("code = ?", strings.TrimSpace(code)).Delete(&model.JobPosition{}).Error
}
