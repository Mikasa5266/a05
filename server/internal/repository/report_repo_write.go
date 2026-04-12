package repository

import (
	"your-project/internal/model"

	"gorm.io/gorm/clause"
)

func (r *GormReportRepository) UpsertByInterview(report *model.Report) error {
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
			"qa_details",
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

func (r *GormReportRepository) Create(report *model.Report) error {
	return r.UpsertByInterview(report)
}

func (r *GormReportRepository) Update(report *model.Report) error {
	return r.db.Save(report).Error
}

func (r *GormReportRepository) Delete(id uint) error {
	return r.db.Delete(&model.Report{}, id).Error
}
