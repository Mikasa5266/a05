package repository

import "your-project/internal/model"

type PositionReportStat struct {
	Position string  `json:"position"`
	Count    int64   `json:"count"`
	AvgScore float64 `json:"avg_score"`
}

type DifficultyReportStat struct {
	Difficulty string  `json:"difficulty"`
	Count      int64   `json:"count"`
	AvgScore   float64 `json:"avg_score"`
}

type ReportStats struct {
	TotalReports int64                  `json:"total_reports"`
	AverageScore float64                `json:"average_score"`
	ByPosition   []PositionReportStat   `json:"by_position"`
	ByDifficulty []DifficultyReportStat `json:"by_difficulty"`
}

func (r *GormReportRepository) GetUserReportStats(userID uint) (*ReportStats, error) {
	stats := &ReportStats{
		ByPosition:   make([]PositionReportStat, 0),
		ByDifficulty: make([]DifficultyReportStat, 0),
	}

	err := r.db.Model(&model.Report{}).Where("user_id = ?", userID).Count(&stats.TotalReports).Error
	if err != nil {
		return nil, err
	}

	err = r.db.Model(&model.Report{}).
		Where("user_id = ?", userID).
		Select("AVG(average_score)").
		Scan(&stats.AverageScore).Error
	if err != nil {
		stats.AverageScore = 0
	}

	err = r.db.Model(&model.Report{}).
		Where("user_id = ?", userID).
		Select("position, COUNT(*) as count, AVG(average_score) as avg_score").
		Group("position").
		Scan(&stats.ByPosition).Error
	if err != nil {
		return nil, err
	}

	err = r.db.Model(&model.Report{}).
		Where("user_id = ?", userID).
		Select("difficulty, COUNT(*) as count, AVG(average_score) as avg_score").
		Group("difficulty").
		Scan(&stats.ByDifficulty).Error
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (r *GormReportRepository) GetReportStats(userID uint) (map[string]interface{}, error) {
	stats, err := r.GetUserReportStats(userID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_reports": stats.TotalReports,
		"average_score": stats.AverageScore,
		"by_position":   stats.ByPosition,
		"by_difficulty": stats.ByDifficulty,
	}, nil
}
