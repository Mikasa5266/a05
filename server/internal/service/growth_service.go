package service

import (
	"sort"

	"your-project/internal/repository"
)

type GrowthService struct {
	reportRepo repository.ReportRepository
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
	A        int    `json:"A"`
	FullMark int    `json:"fullMark"`
}

type GrowthPoint struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
}

type SkillGap struct {
	Name  string `json:"name"`
	Level string `json:"level"`
}

func (s *GrowthService) GetGrowthStats(userID uint) (*GrowthStats, error) {
	reports, err := s.reportRepo.ListByUserTimeline(userID)
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
		// 按天聚合，避免同一天多场面试导致成长曲线抖动。
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

	dayKeys := make([]string, 0, len(growthByDay))
	for k := range growthByDay {
		dayKeys = append(dayKeys, k)
	}
	sort.Strings(dayKeys)

	growthData := make([]GrowthPoint, 0, len(dayKeys))
	for _, dayKey := range dayKeys {
		agg := growthByDay[dayKey]
		score := 0
		if agg.count > 0 {
			score = agg.sum / agg.count
		}
		growthData = append(growthData, GrowthPoint{
			Name:  dayKey[5:],
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
		for i := range radarData {
			radarData[i].A = 60
		}
	}

	skillGaps := make([]SkillGap, 0, len(radarData))
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
