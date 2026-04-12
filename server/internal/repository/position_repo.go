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
	defaults := append([]model.JobPosition{}, model.DefaultJobPositions...)
	defaults = append(defaults,
		model.JobPosition{Code: "go_backend", Name: "Go后端工程师", Domain: "backend", Description: "高并发后端服务开发", IsActive: true},
		model.JobPosition{Code: "python_backend", Name: "Python后端工程师", Domain: "backend", Description: "数据与服务端应用开发", IsActive: true},
		model.JobPosition{Code: "fullstack", Name: "全栈工程师", Domain: "frontend", Description: "前后端协同开发", IsActive: true},
		model.JobPosition{Code: "ios", Name: "iOS开发工程师", Domain: "mobile", Description: "iOS客户端开发", IsActive: true},
		model.JobPosition{Code: "android", Name: "Android开发工程师", Domain: "mobile", Description: "Android客户端开发", IsActive: true},
		model.JobPosition{Code: "devops", Name: "DevOps工程师", Domain: "infrastructure", Description: "发布与运维自动化", IsActive: true},
		model.JobPosition{Code: "data_engineer", Name: "数据工程师", Domain: "data", Description: "数据链路与数据平台建设", IsActive: true},
		model.JobPosition{Code: "test_engineer", Name: "测试开发工程师", Domain: "qa", Description: "质量保障与自动化测试", IsActive: true},
		model.JobPosition{Code: "product_manager", Name: "产品经理", Domain: "product", Description: "产品规划与需求管理", IsActive: true},
		model.JobPosition{Code: "uiux_designer", Name: "UI/UX设计师", Domain: "design", Description: "交互与视觉体验设计", IsActive: true},
	)
	return r.UpsertDefaults(defaults)
}

func (r *GormPositionRepository) Update(position *model.JobPosition) error {
	return r.db.Save(position).Error
}

func (r *GormPositionRepository) DeleteByCode(code string) error {
	return r.db.Where("code = ?", strings.TrimSpace(code)).Delete(&model.JobPosition{}).Error
}
