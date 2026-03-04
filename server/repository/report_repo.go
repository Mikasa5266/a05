package repository

import (
	"your-project/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReportRepository struct {
	db *gorm.DB
}

func NewReportRepository() *ReportRepository {
	return &ReportRepository{
		db: GetDB(),
	}
}

func (r *ReportRepository) Create(report *model.Report) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "interview_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"user_id",
			"position",
			"difficulty",
			"total_questions",
			"average_score",
			"strengths",
			"weaknesses",
			"suggestions",
			"overall_analysis",
			"technical_score",
			"expression_score",
			"logic_score",
			"matching_score",
			"behavior_score",
			"start_time",
			"end_time",
			"duration",
			"updated_at",
		}),
	}).Create(report).Error
}

func (r *ReportRepository) GetByID(id uint) (*model.Report, error) {
	var report model.Report
	err := r.db.Preload("User").Preload("Interview").First(&report, id).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *ReportRepository) GetByUserID(userID uint, page, pageSize int) ([]*model.Report, int64, error) {
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

func (r *ReportRepository) GetByInterviewID(interviewID uint) (*model.Report, error) {
	var report model.Report
	err := r.db.Where("interview_id = ?", interviewID).First(&report).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *ReportRepository) Update(report *model.Report) error {
	return r.db.Save(report).Error
}

func (r *ReportRepository) Delete(id uint) error {
	return r.db.Delete(&model.Report{}, id).Error
}

func (r *ReportRepository) GetReportStats(userID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalReports int64
	err := r.db.Model(&model.Report{}).Where("user_id = ?", userID).Count(&totalReports).Error
	if err != nil {
		return nil, err
	}
	stats["total_reports"] = totalReports

	var avgScore float64
	err = r.db.Model(&model.Report{}).
		Where("user_id = ?", userID).
		Select("AVG(average_score)").
		Scan(&avgScore).Error
	if err != nil {
		avgScore = 0
	}
	stats["average_score"] = avgScore

	var positionStats []struct {
		Position string
		Count    int64
		AvgScore float64
	}
	err = r.db.Model(&model.Report{}).
		Where("user_id = ?", userID).
		Select("position, COUNT(*) as count, AVG(average_score) as avg_score").
		Group("position").
		Scan(&positionStats).Error
	if err != nil {
		return nil, err
	}
	stats["by_position"] = positionStats

	var difficultyStats []struct {
		Difficulty string
		Count      int64
		AvgScore   float64
	}
	err = r.db.Model(&model.Report{}).
		Where("user_id = ?", userID).
		Select("difficulty, COUNT(*) as count, AVG(average_score) as avg_score").
		Group("difficulty").
		Scan(&difficultyStats).Error
	if err != nil {
		return nil, err
	}
	stats["by_difficulty"] = difficultyStats

	return stats, nil
}
