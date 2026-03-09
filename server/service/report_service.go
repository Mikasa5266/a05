package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"your-project/model"
	"your-project/repository"
)

type ReportService struct {
	reportRepo    *repository.ReportRepository
	interviewRepo *repository.InterviewRepository
	aiService     *AIService
}

func NewReportService() *ReportService {
	return &ReportService{
		reportRepo:    repository.NewReportRepository(),
		interviewRepo: repository.NewInterviewRepository(),
		aiService:     NewAIService(),
	}
}

func (s *ReportService) GenerateInterviewReport(userID, interviewID uint) (*model.Report, error) {
	existing, _ := s.reportRepo.GetByInterviewID(interviewID)

	interview, err := s.interviewRepo.GetByID(interviewID)
	if err != nil {
		return nil, fmt.Errorf("interview not found: %w", err)
	}

	if interview.UserID != userID {
		return nil, fmt.Errorf("unauthorized access")
	}

	if interview.Status != "completed" {
		return nil, fmt.Errorf("interview is not completed")
	}

	answers, err := s.interviewRepo.GetAnswersByInterviewID(interviewID)
	if err != nil {
		return nil, fmt.Errorf("failed to get answers: %w", err)
	}

	if existing != nil {
		if existing.UserID != userID {
			return nil, fmt.Errorf("unauthorized access")
		}
		if !shouldRegenerateReport(existing, answers) {
			return existing, nil
		}
	}

	totalScore := 0
	for _, answer := range answers {
		totalScore += answer.Score
	}

	averageScore := 0
	if len(answers) > 0 {
		averageScore = totalScore / len(answers)
	}
	aggregated := aggregateReportDimensionScores(answers, averageScore)

	strengths, weaknesses, suggestions := s.analyzePerformance(answers)
	overallAnalysis := "基于面试表现，建议继续提升技术能力。"
	technicalScore := aggregated.Technical
	expressionScore := aggregated.Expression
	logicScore := aggregated.Logic
	matchingScore := aggregated.Matching
	behaviorScore := aggregated.Behavior

	if insights, aiErr := s.aiService.GenerateReportInsights(interview, answers); aiErr == nil && insights != nil {
		overallAnalysis = insights.OverallAnalysis
		if len(insights.Strengths) > 0 {
			strengths = insights.Strengths
		}
		if len(insights.Weaknesses) > 0 {
			weaknesses = insights.Weaknesses
		}
		if len(insights.Suggestions) > 0 {
			suggestions = insights.Suggestions
		}
		technicalScore = insights.TechnicalScore
		expressionScore = insights.ExpressionScore
		logicScore = insights.LogicScore
		matchingScore = insights.MatchingScore
		behaviorScore = insights.BehaviorScore
	} else {
		if analysis, analysisErr := s.aiService.GenerateOverallAnalysis(interview, answers); analysisErr == nil && analysis != "" {
			overallAnalysis = analysis
		}
	}

	end := time.Now()
	if interview.EndTime != nil {
		end = *interview.EndTime
	}
	report := existing
	if report == nil {
		report = &model.Report{
			UserID:      userID,
			InterviewID: interviewID,
			CreatedAt:   time.Now(),
		}
	}
	report.UserID = userID
	report.InterviewID = interviewID
	report.Position = interview.Position
	report.Difficulty = interview.Difficulty
	report.TotalQuestions = len(answers)
	report.AverageScore = averageScore
	report.OverallAnalysis = overallAnalysis
	report.TechnicalScore = technicalScore
	report.ExpressionScore = expressionScore
	report.LogicScore = logicScore
	report.MatchingScore = matchingScore
	report.BehaviorScore = behaviorScore
	report.StartTime = interview.StartTime
	report.EndTime = end
	report.Duration = int(end.Sub(interview.StartTime).Minutes())
	report.UpdatedAt = time.Now()
	report.SetStrengths(strengths)
	report.SetWeaknesses(weaknesses)
	report.SetSuggestions(suggestions)

	if existing != nil {
		if err := s.reportRepo.Update(report); err != nil {
			return nil, fmt.Errorf("failed to refresh report: %w", err)
		}
		return report, nil
	}

	if err := s.reportRepo.Create(report); err != nil {
		return nil, fmt.Errorf("failed to create report: %w", err)
	}

	persisted, err := s.reportRepo.GetByInterviewID(interviewID)
	if err != nil {
		return nil, fmt.Errorf("failed to load report after save: %w", err)
	}

	return persisted, nil
}

func shouldRegenerateReport(report *model.Report, answers []model.AnswerResult) bool {
	if report == nil {
		return true
	}
	if report.TotalQuestions == 0 || report.AverageScore == 0 {
		if len(answers) > 0 {
			return true
		}
	}
	if report.OverallAnalysis == "" {
		return true
	}
	if len(report.GetStrengths()) == 0 || len(report.GetWeaknesses()) == 0 || len(report.GetSuggestions()) == 0 {
		return true
	}
	// 兼容旧逻辑生成的“全维度=平均分”报告，发现后自动重算
	if len(answers) > 0 && isFlatDimensionReport(report) {
		return true
	}
	return false
}

type reportDimensionScores struct {
	Technical  int
	Expression int
	Logic      int
	Matching   int
	Behavior   int
}

func aggregateReportDimensionScores(answers []model.AnswerResult, fallback int) reportDimensionScores {
	totalTech := 0
	totalExpr := 0
	totalLogic := 0
	totalComp := 0
	count := 0

	for _, ans := range answers {
		feedback := stringsTrimSpaceFast(ans.Feedback)
		if feedback == "" {
			continue
		}

		var payload struct {
			Dimensions *ReviewDimensions `json:"dimensions"`
		}
		if err := json.Unmarshal([]byte(feedback), &payload); err != nil {
			continue
		}
		if payload.Dimensions == nil {
			continue
		}

		totalTech += clampScore(payload.Dimensions.TechnicalDepth)
		totalExpr += clampScore(payload.Dimensions.Expression)
		totalLogic += clampScore(payload.Dimensions.Logic)
		totalComp += clampScore(payload.Dimensions.Completeness)
		count++
	}

	if count == 0 {
		return reportDimensionScores{
			Technical:  clampScore(fallback),
			Expression: clampScore(fallback + 3),
			Logic:      clampScore(fallback),
			Matching:   clampScore(fallback - 2),
			Behavior:   clampScore(fallback + 2),
		}
	}

	avgTech := totalTech / count
	avgExpr := totalExpr / count
	avgLogic := totalLogic / count
	avgComp := totalComp / count

	matching := (avgTech*45 + avgComp*35 + avgLogic*20 + 50) / 100
	behavior := (avgExpr*60 + avgLogic*40 + 50) / 100

	return reportDimensionScores{
		Technical:  clampScore(avgTech),
		Expression: clampScore(avgExpr),
		Logic:      clampScore(avgLogic),
		Matching:   clampScore(matching),
		Behavior:   clampScore(behavior),
	}
}

func isFlatDimensionReport(report *model.Report) bool {
	if report == nil {
		return false
	}
	flat := report.TechnicalScore == report.ExpressionScore &&
		report.ExpressionScore == report.LogicScore &&
		report.LogicScore == report.MatchingScore &&
		report.MatchingScore == report.BehaviorScore

	if !flat {
		return false
	}
	return report.TechnicalScore == report.AverageScore
}

func stringsTrimSpaceFast(s string) string {
	if s == "" {
		return ""
	}
	return strings.TrimSpace(s)
}

func (s *ReportService) GetUserReports(userID uint, page, pageSize int) ([]*model.Report, int64, error) {
	return s.reportRepo.GetByUserID(userID, page, pageSize)
}

func (s *ReportService) GetReportByID(userID, reportID uint) (*model.Report, error) {
	report, err := s.reportRepo.GetByID(reportID)
	if err != nil {
		return nil, err
	}

	if report.UserID != userID {
		return nil, fmt.Errorf("unauthorized access")
	}

	return report, nil
}

func (s *ReportService) analyzePerformance(answers []model.AnswerResult) (strengths, weaknesses, suggestions []string) {
	scoreDistribution := make(map[string]int)
	for _, answer := range answers {
		if answer.Score >= 80 {
			scoreDistribution["excellent"]++
		} else if answer.Score >= 60 {
			scoreDistribution["good"]++
		} else if answer.Score >= 40 {
			scoreDistribution["average"]++
		} else {
			scoreDistribution["poor"]++
		}
	}

	if scoreDistribution["excellent"] > len(answers)/2 {
		strengths = append(strengths, "技术基础扎实，回答准确率高")
	}

	if scoreDistribution["poor"] > len(answers)/4 {
		weaknesses = append(weaknesses, "部分基础知识掌握不够牢固")
		suggestions = append(suggestions, "建议加强基础知识的系统学习")
	}

	if len(answers) > 0 {
		firstAnswer := answers[0]
		lastAnswer := answers[len(answers)-1]

		if lastAnswer.Score > firstAnswer.Score {
			strengths = append(strengths, "面试过程中表现逐渐改善，适应能力强")
		} else if lastAnswer.Score < firstAnswer.Score {
			weaknesses = append(weaknesses, "面试后期表现有所下降，可能是紧张或疲劳")
			suggestions = append(suggestions, "建议加强面试技巧训练，提升稳定性")
		}
	}

	if len(strengths) == 0 {
		strengths = append(strengths, "积极参与面试，态度认真")
	}

	if len(weaknesses) == 0 {
		weaknesses = append(weaknesses, "整体表现良好，仍有提升空间")
	}

	if len(suggestions) == 0 {
		suggestions = append(suggestions, "继续保持学习态度，关注行业最新动态")
	}

	return strengths, weaknesses, suggestions
}
