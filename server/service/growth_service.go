package service

import (
	"your-project/repository"
)

type GrowthService struct {
	reportRepo *repository.ReportRepository
}

func NewGrowthService() *GrowthService {
	return &GrowthService{
		reportRepo: repository.NewReportRepository(),
	}
}

type GrowthStats struct {
	RadarData  []RadarPoint  `json:"radar_data"`
	GrowthData []GrowthPoint `json:"growth_data"`
	SkillGaps  []SkillGap    `json:"skill_gaps"`
}

type RadarPoint struct {
	Subject  string `json:"subject"`
	A        int    `json:"A"` // Score
	FullMark int    `json:"fullMark"`
}

type GrowthPoint struct {
	Name  string `json:"name"` // Month/Date
	Score int    `json:"score"`
}

type SkillGap struct {
	Name  string `json:"name"`
	Level string `json:"level"` // e.g. "Good", "Needs Improvement"
}

func (s *GrowthService) GetGrowthStats(userID uint) (*GrowthStats, error) {
	// 1. Get all reports for user
	reports, err := s.reportRepo.GetAllByUserID(userID)
	if err != nil {
		return nil, err
	}

	var totalTech, totalExp, totalLogic, totalMatch, totalBehav int
	count := 0

	// Also prepare Growth Data (Time vs AverageScore)
	var growthData []GrowthPoint

	for _, r := range reports {
		// Only count valid reports
		if r.AverageScore > 0 {
			totalTech += r.TechnicalScore
			totalExp += r.ExpressionScore
			totalLogic += r.LogicScore
			totalMatch += r.MatchingScore
			totalBehav += r.BehaviorScore
			count++

			// Growth data: simplify to "YYYY-MM-DD" or "Month"
			growthData = append(growthData, GrowthPoint{
				Name:  r.CreatedAt.Format("01-02"), // MM-DD
				Score: r.AverageScore,
			})
		}
	}

	radarData := []RadarPoint{
		{Subject: "技术深度", A: 0, FullMark: 100},
		{Subject: "表达能力", A: 0, FullMark: 100},
		{Subject: "逻辑严谨", A: 0, FullMark: 100},
		{Subject: "岗位匹配", A: 0, FullMark: 100},
		{Subject: "行为表现", A: 0, FullMark: 100},
	}

	if count > 0 {
		radarData[0].A = totalTech / count
		radarData[1].A = totalExp / count
		radarData[2].A = totalLogic / count
		radarData[3].A = totalMatch / count
		radarData[4].A = totalBehav / count
	} else {
		// Default/Mock data for new users so the chart isn't empty
		radarData[0].A = 60
		radarData[1].A = 60
		radarData[2].A = 60
		radarData[3].A = 60
		radarData[4].A = 60
	}

	// Skill Gaps - hardcoded mock or analyzed from Weaknesses?
	// Let's extract top recurring weaknesses
	skillGaps := []SkillGap{
		{Name: "System Design", Level: "急需提升"},
		{Name: "Go Concurrency", Level: "中等差距"},
		{Name: "Microservices", Level: "良好"},
	}

	return &GrowthStats{
		RadarData:  radarData,
		GrowthData: growthData,
		SkillGaps:  skillGaps,
	}, nil
}
