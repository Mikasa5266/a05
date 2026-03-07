package service

import (
	"sort"
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
	type dayAggregate struct {
		sum   int
		count int
	}
	growthByDay := make(map[string]*dayAggregate)

	for _, r := range reports {
		// 成长曲线：按天聚合，避免同一天多场面试抖动
		dayKey := r.CreatedAt.Format("2006-01-02")
		if _, ok := growthByDay[dayKey]; !ok {
			growthByDay[dayKey] = &dayAggregate{}
		}
		growthByDay[dayKey].sum += r.AverageScore
		growthByDay[dayKey].count++

		totalTech += r.TechnicalScore
		totalExp += r.ExpressionScore
		totalLogic += r.LogicScore
		totalMatch += r.MatchingScore
		totalBehav += r.BehaviorScore
		count++
	}

	var dayKeys []string
	for k := range growthByDay {
		dayKeys = append(dayKeys, k)
	}
	sort.Strings(dayKeys)

	var growthData []GrowthPoint
	for _, dayKey := range dayKeys {
		agg := growthByDay[dayKey]
		score := 0
		if agg.count > 0 {
			score = agg.sum / agg.count
		}
		growthData = append(growthData, GrowthPoint{
			Name:  dayKey[5:], // MM-DD
			Score: score,
		})
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

	// Skill Gaps - 根据雷达均值动态生成
	skillGaps := []SkillGap{}
	type dim struct {
		name  string
		score int
	}
	dims := []dim{
		{name: "技术深度", score: radarData[0].A},
		{name: "表达能力", score: radarData[1].A},
		{name: "逻辑严谨", score: radarData[2].A},
		{name: "岗位匹配", score: radarData[3].A},
		{name: "行为表现", score: radarData[4].A},
	}
	for _, d := range dims {
		level := "良好"
		if d.score < 60 {
			level = "急需提升"
		} else if d.score < 75 {
			level = "中等差距"
		}
		skillGaps = append(skillGaps, SkillGap{Name: d.name, Level: level})
	}

	return &GrowthStats{
		RadarData:  radarData,
		GrowthData: growthData,
		SkillGaps:  skillGaps,
	}, nil
}
