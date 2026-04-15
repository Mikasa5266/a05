package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"your-project/internal/model"
	internalruntime "your-project/internal/runtime"
	aidomain "your-project/internal/service/ai"
)

func buildGeneratingReportRecord(existing *model.Report, interview *model.Interview, userID uint, answers []model.AnswerResult) *model.Report {
	now := time.Now()
	end := now
	if interview != nil && interview.EndTime != nil {
		end = *interview.EndTime
	}

	coverage := buildReportCoverageStats(interview, answers)
	averageScore := computePenalizedAverageScore(answers, coverage.ExpectedCount)

	report := existing
	if report == nil {
		report = &model.Report{
			UserID:      userID,
			InterviewID: interview.ID,
			CreatedAt:   now,
		}
		report.SetStrengths([]string{})
		report.SetWeaknesses([]string{})
		report.SetSuggestions([]string{})
		report.SetQADetails([]model.ReportQADetail{})
	}

	report.UserID = userID
	report.InterviewID = interview.ID
	report.Position = interview.Position
	report.Difficulty = interview.Difficulty
	report.TotalQuestions = coverage.ExpectedCount
	report.AverageScore = averageScore
	report.OverallAnalysis = "报告生成中，请稍后刷新。"
	report.Status = reportStatusGenerating
	report.StartTime = interview.StartTime
	report.EndTime = end
	duration := int(end.Sub(interview.StartTime).Minutes())
	if duration < 0 {
		duration = 0
	}
	report.Duration = duration
	report.UpdatedAt = now

	recordingURL := strings.TrimSpace(interview.RecordingURL)
	if interview.IsGroup {
		report.SinglePlayback = ""
		report.MultiPlayback = recordingURL
	} else {
		report.SinglePlayback = recordingURL
		report.MultiPlayback = ""
	}

	return report
}

func (s *ReportService) generateInterviewReportAsync(userID, interviewID uint) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("[report] panic while generating report interview=%d: %v", interviewID, recovered)
				s.markReportFailed(interviewID, "后台生成过程发生异常")
			}
		}()

		if err := s.generateAndPersistCompletedReport(ctx, userID, interviewID); err != nil {
			log.Printf("[report] generate failed interview=%d: %v", interviewID, err)
			s.markReportFailed(interviewID, err.Error())
		}
	}()
}

func (s *ReportService) generateAndPersistCompletedReport(ctx context.Context, userID, interviewID uint) error {
	existing, err := s.reportRepo.GetByInterviewID(interviewID)
	if err != nil {
		return fmt.Errorf("load queued report failed: %w", err)
	}

	interview, err := s.interviewRepo.GetByID(interviewID)
	if err != nil {
		return fmt.Errorf("load interview failed: %w", err)
	}
	if interview.UserID != userID {
		return fmt.Errorf("unauthorized report generation")
	}

	answers, err := s.interviewRepo.GetAnswersByInterviewID(interviewID)
	if err != nil {
		return fmt.Errorf("load answers failed: %w", err)
	}

	coverage := buildReportCoverageStats(interview, answers)
	averageScore := computePenalizedAverageScore(answers, coverage.ExpectedCount)
	aggregated := aggregateReportDimensionScores(answers, averageScore, coverage)
	qaDetails := buildReportQADetailsFromAnswers(answers)
	strengths, weaknesses, suggestions := s.analyzePerformance(answers)
	overallAnalysis := "基于面试表现，建议继续提升技术能力。"
	technicalScore := aggregated.Technical
	expressionScore := aggregated.Expression
	logicScore := aggregated.Logic
	matchingScore := aggregated.Matching
	behaviorScore := aggregated.Behavior

	policy := aidomain.ReportScoringPolicy{
		ExpectedCount: coverage.ExpectedCount,
		ActualCount:   coverage.ActualCount,
		MissingCount:  coverage.MissingCount,
		MissingAsZero: true,
		EarlyExit:     normalizeInterviewExitType(interview.ExitType) == interviewExitTypeEarlyExit,
	}

	if insights, aiErr := s.aiService.GenerateReportInsights(ctx, interview, answers, policy); aiErr == nil && insights != nil {
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
		if mappedDetails := mapInsightsQADetails(insights.QADetails); len(mappedDetails) > 0 {
			qaDetails = mappedDetails
		}
	} else {
		if analysis, analysisErr := s.aiService.GenerateOverallAnalysis(ctx, interview, answers); analysisErr == nil && analysis != "" {
			overallAnalysis = analysis
		}
	}

	if coverage.MissingCount > 0 {
		technicalScore = minScore(technicalScore, aggregated.Technical)
		expressionScore = minScore(expressionScore, aggregated.Expression)
		logicScore = minScore(logicScore, aggregated.Logic)
		matchingScore = minScore(matchingScore, aggregated.Matching)
		behaviorScore = minScore(behaviorScore, aggregated.Behavior)
		settlementNote := fmt.Sprintf("本场面试按提前结算规则处理：计划 %d 题，实际作答 %d 题，缺失 %d 题按 0 分计入。", coverage.ExpectedCount, coverage.ActualCount, coverage.MissingCount)
		overallAnalysis = strings.TrimSpace(settlementNote + " " + strings.TrimSpace(overallAnalysis))
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
	report.TotalQuestions = coverage.ExpectedCount
	report.AverageScore = averageScore
	report.OverallAnalysis = overallAnalysis
	report.Status = reportStatusCompleted
	report.TechnicalScore = technicalScore
	report.ExpressionScore = expressionScore
	report.LogicScore = logicScore
	report.MatchingScore = matchingScore
	report.BehaviorScore = behaviorScore
	report.StartTime = interview.StartTime
	report.EndTime = end
	duration := int(end.Sub(interview.StartTime).Minutes())
	if duration < 0 {
		duration = 0
	}
	report.Duration = duration
	report.UpdatedAt = time.Now()
	report.SetStrengths(strengths)
	report.SetWeaknesses(weaknesses)
	report.SetSuggestions(suggestions)
	report.SetQADetails(qaDetails)

	recordingURL := strings.TrimSpace(interview.RecordingURL)
	if interview.IsGroup {
		report.SinglePlayback = ""
		report.MultiPlayback = recordingURL
	} else {
		report.SinglePlayback = recordingURL
		report.MultiPlayback = ""
	}

	if cacheKey := internalruntime.InterviewRoomCacheKey(interviewID); cacheKey != "" {
		audioTranscripts, chatByReceiver := internalruntime.GetLiveRoomStore().Snapshot(cacheKey)
		if len(audioTranscripts) > 0 {
			report.SetAudioTranscripts(audioTranscripts)
		}
		if chatMessages := chatByReceiver[userID]; len(chatMessages) > 0 {
			report.SetChatMessages(chatMessages)
		}
	}

	if err := s.reportRepo.UpsertByInterview(report); err != nil {
		return fmt.Errorf("persist completed report failed: %w", err)
	}
	return nil
}

func minScore(left, right int) int {
	if left < right {
		return clampScore(left)
	}
	return clampScore(right)
}

func (s *ReportService) markReportFailed(interviewID uint, reason string) {
	report, err := s.reportRepo.GetByInterviewID(interviewID)
	if err != nil {
		log.Printf("[report] mark failed skipped, report not found interview=%d: %v", interviewID, err)
		return
	}

	report.Status = reportStatusFailed
	report.UpdatedAt = time.Now()
	base := "报告生成失败，请稍后重试。"
	if stringsTrimSpaceFast(report.OverallAnalysis) == "" || report.OverallAnalysis == "报告生成中，请稍后刷新。" {
		report.OverallAnalysis = base
	}
	if msg := stringsTrimSpaceFast(reason); msg != "" {
		report.OverallAnalysis = fmt.Sprintf("%s 错误信息：%s", base, truncateRunesForReport(msg, 160))
	}

	if err := s.reportRepo.Update(report); err != nil {
		log.Printf("[report] failed to persist failed status interview=%d: %v", interviewID, err)
	}
}
