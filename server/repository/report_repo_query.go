package repository

import "your-project/model"

func (r *GormReportRepository) GetByID(id uint) (*model.Report, error) {
	var report model.Report
	err := r.db.Preload("User").Preload("Interview").First(&report, id).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *GormReportRepository) GetByInterviewID(interviewID uint) (*model.Report, error) {
	var report model.Report
	err := r.db.Where("interview_id = ?", interviewID).First(&report).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *GormReportRepository) ListByUserPaged(userID uint, page, pageSize int) ([]*model.Report, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	var reports []*model.Report
	var total int64
	offset := (page - 1) * pageSize

	err := r.db.Model(&model.Report{}).Where("user_id = ?", userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Preload("Interview").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&reports).Error
	if err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

func (r *GormReportRepository) GetByUserID(userID uint, page, pageSize int) ([]*model.Report, int64, error) {
	return r.ListByUserPaged(userID, page, pageSize)
}

func (r *GormReportRepository) ListByUserTimeline(userID uint) ([]*model.Report, error) {
	var reports []*model.Report
	err := r.db.Where("user_id = ?", userID).Order("created_at ASC").Find(&reports).Error
	if err != nil {
		return nil, err
	}
	return reports, nil
}

func (r *GormReportRepository) GetAllByUserID(userID uint) ([]*model.Report, error) {
	return r.ListByUserTimeline(userID)
}
