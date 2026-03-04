package repository

import (
	"your-project/model"
)

// Add GetByUserID without pagination or with large limit for growth stats
func (r *ReportRepository) GetAllByUserID(userID uint) ([]*model.Report, error) {
	var reports []*model.Report
	err := r.db.Where("user_id = ?", userID).Order("created_at ASC").Find(&reports).Error
	if err != nil {
		return nil, err
	}
	return reports, nil
}
