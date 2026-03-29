package repository

import (
	"strings"

	"your-project/model"

	"gorm.io/gorm"
)

type EnterpriseRepository interface {
	CreateEnterprise(item *model.Enterprise) error
	GetEnterpriseByID(id uint) (*model.Enterprise, error)
	GetEnterpriseByUserID(userID uint) (*model.Enterprise, error)
	UpdateEnterprise(item *model.Enterprise) error
	DeleteEnterprise(id uint) error

	CreateJob(job *model.Job) error
	GetJobByID(id uint) (*model.Job, error)
	ListJobs(status string) ([]model.Job, error)
	UpdateJob(job *model.Job) error
	DeleteJob(id uint) error

	CreateTalentRecord(record *model.TalentRecord) error
	GetTalentRecordByID(id uint) (*model.TalentRecord, error)
	ListTalentRecords(search string, limit int) ([]model.TalentRecord, error)
	UpdateTalentRecord(record *model.TalentRecord) error
	DeleteTalentRecord(id uint) error
}

type GormEnterpriseRepository struct {
	db *gorm.DB
}

var _ EnterpriseRepository = (*GormEnterpriseRepository)(nil)

func NewEnterpriseRepository() EnterpriseRepository {
	return &GormEnterpriseRepository{db: GetDB()}
}

func NewEnterpriseRepositoryWithDB(db *gorm.DB) EnterpriseRepository {
	if db == nil {
		db = GetDB()
	}
	return &GormEnterpriseRepository{db: db}
}

func (r *GormEnterpriseRepository) CreateEnterprise(item *model.Enterprise) error {
	return r.db.Create(item).Error
}

func (r *GormEnterpriseRepository) GetEnterpriseByID(id uint) (*model.Enterprise, error) {
	var item model.Enterprise
	if err := r.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormEnterpriseRepository) GetEnterpriseByUserID(userID uint) (*model.Enterprise, error) {
	var item model.Enterprise
	if err := r.db.Where("user_id = ?", userID).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormEnterpriseRepository) UpdateEnterprise(item *model.Enterprise) error {
	return r.db.Save(item).Error
}

func (r *GormEnterpriseRepository) DeleteEnterprise(id uint) error {
	return r.db.Delete(&model.Enterprise{}, id).Error
}

func (r *GormEnterpriseRepository) CreateJob(job *model.Job) error {
	return r.db.Create(job).Error
}

func (r *GormEnterpriseRepository) GetJobByID(id uint) (*model.Job, error) {
	var job model.Job
	if err := r.db.First(&job, id).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *GormEnterpriseRepository) ListJobs(status string) ([]model.Job, error) {
	jobs := make([]model.Job, 0)
	query := r.db.Model(&model.Job{})
	if s := strings.TrimSpace(status); s != "" {
		query = query.Where("status = ?", s)
	}
	if err := query.Order("created_at DESC").Find(&jobs).Error; err != nil {
		return nil, err
	}
	return jobs, nil
}

func (r *GormEnterpriseRepository) UpdateJob(job *model.Job) error {
	return r.db.Save(job).Error
}

func (r *GormEnterpriseRepository) DeleteJob(id uint) error {
	return r.db.Delete(&model.Job{}, id).Error
}

func (r *GormEnterpriseRepository) CreateTalentRecord(record *model.TalentRecord) error {
	return r.db.Create(record).Error
}

func (r *GormEnterpriseRepository) GetTalentRecordByID(id uint) (*model.TalentRecord, error) {
	var record model.TalentRecord
	if err := r.db.First(&record, id).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *GormEnterpriseRepository) ListTalentRecords(search string, limit int) ([]model.TalentRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	records := make([]model.TalentRecord, 0, limit)
	query := r.db.Model(&model.TalentRecord{})
	if s := strings.TrimSpace(search); s != "" {
		like := "%" + s + "%"
		query = query.Where("name LIKE ? OR position LIKE ?", like, like)
	}
	if err := query.Order("match_score DESC").Limit(limit).Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func (r *GormEnterpriseRepository) UpdateTalentRecord(record *model.TalentRecord) error {
	return r.db.Save(record).Error
}

func (r *GormEnterpriseRepository) DeleteTalentRecord(id uint) error {
	return r.db.Delete(&model.TalentRecord{}, id).Error
}
