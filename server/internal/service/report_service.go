package service

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"your-project/internal/model"
	"your-project/internal/repository"
	aidomain "your-project/internal/service/ai"
)

const (
	reportStatusGenerating = "generating"
	reportStatusCompleted  = "completed"
	reportStatusFailed     = "failed"
)

type ReportService struct {
	reportRepo    repository.ReportRepository
	interviewRepo *repository.InterviewRepository
	aiService     aidomain.AIFacade
}

func NewReportService() *ReportService {
	return &ReportService{
		reportRepo:    repository.NewReportRepository(),
		interviewRepo: repository.NewInterviewRepository(),
		aiService:     MustGetAIService(),
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
	if existing != nil && existing.UserID != userID {
		return nil, fmt.Errorf("unauthorized access")
	}

	answers, err := s.interviewRepo.GetAnswersByInterviewID(interviewID)
	if err != nil {
		return nil, fmt.Errorf("failed to get answers: %w", err)
	}

	if existing != nil {
		if stringsTrimSpaceFast(existing.Status) == reportStatusGenerating {
			return existing, nil
		}
		if !shouldRegenerateReport(existing, answers) {
			return existing, nil
		}
	}

	queued := buildGeneratingReportRecord(existing, interview, userID, answers)
	if err := s.reportRepo.UpsertByInterview(queued); err != nil {
		return nil, fmt.Errorf("failed to queue report generation: %w", err)
	}

	persisted, err := s.reportRepo.GetByInterviewID(interviewID)
	if err != nil {
		return nil, fmt.Errorf("failed to load queued report: %w", err)
	}

	s.generateInterviewReportAsync(userID, interviewID)
	return persisted, nil
}

func shouldRegenerateReport(report *model.Report, answers []model.AnswerResult) bool {
	if report == nil {
		return true
	}

	status := strings.ToLower(stringsTrimSpaceFast(report.Status))
	switch status {
	case reportStatusGenerating:
		return false
	case reportStatusFailed:
		return true
	case "", reportStatusCompleted:
		// 兼容历史数据：空状态视为已完成
	default:
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
	if len(answers) > 0 && len(report.GetQADetails()) == 0 {
		return true
	}
	// 兼容旧逻辑生成的“全维度=平均分”报告，发现后自动重算。
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

type reportCoverageStats struct {
	ExpectedCount   int
	ActualCount     int
	MissingCount    int
	CompletionRatio float64
}

func buildReportCoverageStats(interview *model.Interview, answers []model.AnswerResult) reportCoverageStats {
	actualCount := len(answers)
	expectedCount := 0
	if interview != nil {
		expectedCount = interview.TotalQuestionTarget
		if arranged := len(interview.InterviewQuestions); arranged > expectedCount {
			expectedCount = arranged
		}
	}
	if expectedCount <= 0 {
		expectedCount = actualCount
	}
	if expectedCount < actualCount {
		expectedCount = actualCount
	}

	missingCount := expectedCount - actualCount
	if missingCount < 0 {
		missingCount = 0
	}

	completionRatio := 1.0
	if expectedCount > 0 {
		completionRatio = float64(actualCount) / float64(expectedCount)
	}
	if completionRatio < 0 {
		completionRatio = 0
	}
	if completionRatio > 1 {
		completionRatio = 1
	}

	return reportCoverageStats{
		ExpectedCount:   expectedCount,
		ActualCount:     actualCount,
		MissingCount:    missingCount,
		CompletionRatio: completionRatio,
	}
}

func computePenalizedAverageScore(answers []model.AnswerResult, expectedCount int) int {
	totalScore := 0
	for _, answer := range answers {
		totalScore += answer.Score
	}

	denominator := expectedCount
	if denominator <= 0 {
		denominator = len(answers)
	}
	if denominator <= 0 {
		return 0
	}
	return clampScore(totalScore / denominator)
}

func applyCoveragePenalty(score int, coverage reportCoverageStats) int {
	base := clampScore(score)
	if coverage.MissingCount <= 0 {
		return base
	}
	penalized := int(math.Round(float64(base) * coverage.CompletionRatio))
	return clampScore(penalized)
}

func aggregateReportDimensionScores(answers []model.AnswerResult, fallback int, coverage reportCoverageStats) reportDimensionScores {
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
			Dimensions *aidomain.ReviewDimensions `json:"dimensions"`
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
		scores := reportDimensionScores{
			Technical:  clampScore(fallback),
			Expression: clampScore(fallback + 3),
			Logic:      clampScore(fallback),
			Matching:   clampScore(fallback - 2),
			Behavior:   clampScore(fallback + 2),
		}
		scores.Technical = applyCoveragePenalty(scores.Technical, coverage)
		scores.Expression = applyCoveragePenalty(scores.Expression, coverage)
		scores.Logic = applyCoveragePenalty(scores.Logic, coverage)
		scores.Matching = applyCoveragePenalty(scores.Matching, coverage)
		scores.Behavior = applyCoveragePenalty(scores.Behavior, coverage)
		return scores
	}

	avgTech := totalTech / count
	avgExpr := totalExpr / count
	avgLogic := totalLogic / count
	avgComp := totalComp / count

	matching := (avgTech*45 + avgComp*35 + avgLogic*20 + 50) / 100
	behavior := (avgExpr*60 + avgLogic*40 + 50) / 100

	scores := reportDimensionScores{
		Technical:  clampScore(avgTech),
		Expression: clampScore(avgExpr),
		Logic:      clampScore(avgLogic),
		Matching:   clampScore(matching),
		Behavior:   clampScore(behavior),
	}
	scores.Technical = applyCoveragePenalty(scores.Technical, coverage)
	scores.Expression = applyCoveragePenalty(scores.Expression, coverage)
	scores.Logic = applyCoveragePenalty(scores.Logic, coverage)
	scores.Matching = applyCoveragePenalty(scores.Matching, coverage)
	scores.Behavior = applyCoveragePenalty(scores.Behavior, coverage)
	return scores
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

func mapInsightsQADetails(items []aidomain.ReportQADetail) []model.ReportQADetail {
	if len(items) == 0 {
		return []model.ReportQADetail{}
	}

	mapped := make([]model.ReportQADetail, 0, len(items))
	for _, item := range items {
		question := stringsTrimSpaceFast(item.Question)
		userAnswer := stringsTrimSpaceFast(item.UserAnswer)
		optimizedAnswer := stringsTrimSpaceFast(item.OptimizedAnswer)
		if question == "" || (userAnswer == "" && optimizedAnswer == "") {
			continue
		}

		if userAnswer == "" {
			userAnswer = "候选人回答摘要暂缺。"
		}
		if optimizedAnswer == "" {
			optimizedAnswer = "建议按“结论-原理-实践-边界”结构组织回答，并补充工程细节。"
		}

		mapped = append(mapped, model.ReportQADetail{
			Question:        truncateRunesForReport(question, 240),
			UserAnswer:      truncateRunesForReport(userAnswer, 520),
			OptimizedAnswer: truncateRunesForReport(optimizedAnswer, 900),
			KeyImprovements: normalizeReportImprovementList(item.KeyImprovements),
		})

		if len(mapped) >= 12 {
			break
		}
	}

	return mapped
}

func buildReportQADetailsFromAnswers(answers []model.AnswerResult) []model.ReportQADetail {
	if len(answers) == 0 {
		return []model.ReportQADetail{}
	}

	result := make([]model.ReportQADetail, 0, len(answers))
	for _, ans := range answers {
		question := stringsTrimSpaceFast(ans.Question.Content)
		if question == "" {
			question = stringsTrimSpaceFast(ans.Question.Title)
		}
		if question == "" {
			continue
		}

		userAnswer := stringsTrimSpaceFast(ans.Answer)
		if userAnswer == "" {
			userAnswer = "候选人回答内容较少，建议补充关键技术点。"
		}

		optimizedAnswer, improvements := extractReportQAHintsFromFeedback(ans.Feedback)
		if optimizedAnswer == "" {
			optimizedAnswer = stringsTrimSpaceFast(ans.Question.ExpectedAnswer)
		}
		if optimizedAnswer == "" {
			optimizedAnswer = "建议按“结论-原理-实践-边界”结构补充回答，并增加可验证的工程细节。"
		}

		result = append(result, model.ReportQADetail{
			Question:        truncateRunesForReport(question, 240),
			UserAnswer:      truncateRunesForReport(userAnswer, 520),
			OptimizedAnswer: truncateRunesForReport(optimizedAnswer, 900),
			KeyImprovements: improvements,
		})

		if len(result) >= 12 {
			break
		}
	}

	return result
}

func extractReportQAHintsFromFeedback(feedback string) (string, []string) {
	raw := stringsTrimSpaceFast(feedback)
	if raw == "" {
		return "", []string{"补充关键机制说明", "增加边界条件与异常处理"}
	}

	var payload struct {
		ModelAnswerOutline string   `json:"model_answer_outline"`
		Suggestions        []string `json:"suggestions"`
		Gaps               []string `json:"gaps"`
		FollowUpContext    string   `json:"follow_up_context"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return "", []string{"补充关键机制说明", "增加边界条件与异常处理"}
	}

	improvements := make([]string, 0, len(payload.Suggestions)+len(payload.Gaps)+1)
	improvements = append(improvements, payload.Suggestions...)
	improvements = append(improvements, payload.Gaps...)
	if stringsTrimSpaceFast(payload.FollowUpContext) != "" {
		improvements = append(improvements, payload.FollowUpContext)
	}

	return stringsTrimSpaceFast(payload.ModelAnswerOutline), normalizeReportImprovementList(improvements)
}

func normalizeReportImprovementList(items []string) []string {
	clean := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		line := stringsTrimSpaceFast(item)
		if line == "" {
			continue
		}
		if _, ok := seen[line]; ok {
			continue
		}
		seen[line] = struct{}{}
		clean = append(clean, truncateRunesForReport(line, 120))
		if len(clean) >= 4 {
			break
		}
	}

	if len(clean) == 0 {
		return []string{"补充关键机制说明", "增加边界条件与异常处理"}
	}
	return clean
}

func truncateRunesForReport(text string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(stringsTrimSpaceFast(text))
	if len(runes) <= max {
		return string(runes)
	}
	return stringsTrimSpaceFast(string(runes[:max])) + "..."
}

func (s *ReportService) GetUserReports(userID uint, page, pageSize int) ([]*model.Report, int64, error) {
	return s.reportRepo.ListByUserPaged(userID, page, pageSize)
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
