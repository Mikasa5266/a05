package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"your-project/model"
	"your-project/repository"

	"gorm.io/gorm"
)

type practiceService struct {
	repo repository.PracticeQuestionRepository
}

func NewPracticeService() PracticeService {
	return NewPracticeServiceWithRepo(MustGetPracticeQuestionRepository())
}

func NewPracticeServiceWithRepo(repo repository.PracticeQuestionRepository) PracticeService {
	if repo == nil {
		repo = MustGetPracticeQuestionRepository()
	}

	return &practiceService{
		repo: repo,
	}
}

func (s *practiceService) GetMeta(ctx context.Context) (*PracticeMetaResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, practiceInternalError("context canceled", err)
	}

	counts, err := s.repo.CountQuestionsByPosition(ctx)
	if err != nil {
		return nil, practiceInternalError("failed to load practice meta", err)
	}

	return &PracticeMetaResponse{
		Roles:         buildPracticeRoleMetaMap(),
		Levels:        buildPracticeLevelMap(),
		QuestionTypes: append([]string(nil), questionBankQuestionTypes...),
		Counts:        counts,
	}, nil
}

func (s *practiceService) GetFilterOptions(ctx context.Context, positionCode string) (*PracticeFilterOptions, error) {
	normalized, err := optionalPracticePositionCode(positionCode)
	if err != nil {
		return nil, err
	}

	points, err := s.repo.ListDistinctPoints(ctx, normalized)
	if err != nil {
		return nil, practiceInternalError("failed to load practice points", err)
	}
	specialties, err := s.repo.ListDistinctSpecialties(ctx, normalized)
	if err != nil {
		return nil, practiceInternalError("failed to load practice specialties", err)
	}

	return &PracticeFilterOptions{
		Points:        sanitizeStringSlice(points),
		Specialties:   sanitizeStringSlice(specialties),
		CompanyTypes:  append([]string(nil), questionBankCompanyTypes...),
		Levels:        buildPracticeLevelMap(),
		QuestionTypes: append([]string(nil), questionBankQuestionTypes...),
		StatusOptions: []string{"todo", "solved", "wrong"},
	}, nil
}

func (s *practiceService) ListQuestions(ctx context.Context, userID uint, req PracticeListQuestionsRequest) (*PracticeQuestionListResponse, error) {
	if userID == 0 {
		return nil, unauthorizedPracticeError("unauthorized")
	}

	query, err := normalizePracticeQuestionQuery(req)
	if err != nil {
		return nil, err
	}

	rows, total, stats, err := s.repo.ListQuestionPage(ctx, userID, query)
	if err != nil {
		return nil, practiceInternalError("failed to load questions", err)
	}

	items := make([]PracticeQuestionSummary, 0, len(rows))
	for _, row := range rows {
		items = append(items, buildPracticeQuestionSummary(row))
	}

	return &PracticeQuestionListResponse{
		Items: items,
		Pagination: PracticePagination{
			Page:       query.Page,
			PageSize:   query.PageSize,
			Total:      int(total),
			TotalPages: calcTotalPages(int(total), query.PageSize),
		},
		StatusStats: PracticeStatusStats{
			Todo:   int(stats.Todo),
			Solved: int(stats.Solved),
			Wrong:  int(stats.Wrong),
		},
	}, nil
}

func (s *practiceService) GetQuestionLists(ctx context.Context, userID uint, positionCode string) ([]PracticeQuestionListDefinition, error) {
	if userID == 0 {
		return nil, unauthorizedPracticeError("unauthorized")
	}

	normalized, err := optionalPracticePositionCode(positionCode)
	if err != nil {
		return nil, err
	}

	summaries, err := s.repo.ListQuestionListsWithProgress(ctx, userID, normalized)
	if err != nil {
		return nil, practiceInternalError("failed to load question lists", err)
	}

	items := make([]PracticeQuestionListDefinition, 0, len(summaries))
	for _, summary := range summaries {
		progress := 0.0
		if summary.TotalCount > 0 {
			progress = roundFloat(float64(summary.SolvedCount*100)/float64(summary.TotalCount), 2)
		}
		role := normalizeQuestionBankPositionCode(summary.List.PositionCode)
		items = append(items, PracticeQuestionListDefinition{
			ID:           summary.List.ID,
			Role:         role,
			PositionCode: role,
			Title:        cleanQuestionPayloadText(summary.List.Title),
			Description:  cleanQuestionPayloadText(summary.List.Description),
			Tags:         sanitizeStringSlice(summary.List.GetTags()),
			TotalCount:   int(summary.TotalCount),
			SolvedCount:  int(summary.SolvedCount),
			Progress:     progress,
		})
	}
	return items, nil
}

func (s *practiceService) GetSpecialties(ctx context.Context, positionCode string) ([]string, error) {
	normalized, err := requiredPracticePositionCode(positionCode)
	if err != nil {
		return nil, err
	}

	items, err := s.repo.ListDistinctSpecialties(ctx, normalized)
	if err != nil {
		return nil, practiceInternalError("failed to load specialties", err)
	}
	return sanitizeStringSlice(items), nil
}

func (s *practiceService) DrawQuestion(ctx context.Context, userID uint, req PracticeDrawRequest) (*PracticeQuestionEnvelope, error) {
	if userID == 0 {
		return nil, unauthorizedPracticeError("unauthorized")
	}

	positionCode, err := requiredPracticePositionCode(req.PositionCode)
	if err != nil {
		return nil, err
	}

	filters := repository.PracticeQuestionFilters{
		PositionCode: positionCode,
		Level:        strings.TrimSpace(strings.ToLower(req.Level)),
		QuestionType: cleanQuestionPayloadText(req.QuestionType),
		Specialty:    cleanQuestionPayloadText(req.Specialty),
		Point:        cleanQuestionPayloadText(req.Point),
		CompanyType:  cleanQuestionPayloadText(req.CompanyType),
	}

	var question *model.Question
	if req.ListID > 0 {
		list, err := s.repo.GetQuestionList(ctx, req.ListID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, practiceNotFound("question list not found", err)
			}
			return nil, practiceInternalError("failed to verify question list", err)
		}
		allowed := make(map[uint]struct{}, len(list.Items))
		for _, item := range list.Items {
			allowed[item.QuestionID] = struct{}{}
		}

		candidates, err := s.repo.ListQuestions(ctx, filters)
		if err != nil {
			return nil, practiceInternalError("failed to draw question", err)
		}

		matched := make([]model.Question, 0, len(candidates))
		for _, candidate := range candidates {
			if _, ok := allowed[candidate.ID]; ok {
				matched = append(matched, candidate)
			}
		}
		if len(matched) == 0 {
			return nil, practiceNotFound("no matching question found", gorm.ErrRecordNotFound)
		}
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		selected := matched[rng.Intn(len(matched))]
		question = &selected
	} else {
		var err error
		question, err = s.repo.DrawRandomQuestion(ctx, filters)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, practiceNotFound("no matching question found", err)
			}
			return nil, practiceInternalError("failed to draw question", err)
		}
	}

	row, err := s.repo.GetQuestionDetail(ctx, userID, question.ID)
	if err != nil {
		return nil, practiceInternalError("failed to load question detail", err)
	}
	return &PracticeQuestionEnvelope{Question: buildPracticeQuestionDetail(*row)}, nil
}

func (s *practiceService) SubmitAnswer(ctx context.Context, userID uint, req PracticeAnswerRequest) (*PracticeAnswerFeedback, error) {
	return s.submitAnswer(ctx, userID, req)
}

func (s *practiceService) GetPointSummary(ctx context.Context, userID uint, positionCode, point string) (*PracticePointSummary, error) {
	normalized, err := requiredPracticePositionCode(positionCode)
	if err != nil {
		return nil, err
	}

	return s.getPointSummary(ctx, userID, normalized, point)
}

func (s *practiceService) GetWrongRemedial(ctx context.Context, userID, wrongID uint) (*PracticeWrongRemedialResponse, error) {
	if userID == 0 {
		return nil, unauthorizedPracticeError("unauthorized")
	}

	wrong, err := s.repo.GetWrongBookEntry(ctx, userID, wrongID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, practiceNotFound("wrong question not found", err)
		}
		return nil, practiceInternalError("failed to load wrong question", err)
	}

	point := cleanQuestionPayloadText(wrong.Question.Points)
	if point == "" {
		point = cleanQuestionPayloadText(wrong.Question.KnowledgePoint)
	}
	positionCode := normalizeQuestionBankPositionCode(wrong.Question.PositionCode)
	questions, err := s.repo.ListRemedialQuestions(ctx, positionCode, point, wrong.QuestionID, 10)
	if err != nil {
		return nil, practiceInternalError("failed to load remedial questions", err)
	}

	baseQuestions := make([]PracticeRemedialQuestion, 0, 4)
	advancedQuestions := make([]PracticeRemedialQuestion, 0, 4)
	for _, question := range questions {
		item := PracticeRemedialQuestion{
			ID:         question.ID,
			Stem:       cleanQuestionPayloadText(question.EffectiveStem()),
			Level:      cleanQuestionPayloadText(question.Level),
			Difficulty: question.DifficultyScore,
		}
		switch strings.TrimSpace(strings.ToLower(question.Level)) {
		case model.QuestionLevelBase:
			if len(baseQuestions) < 4 {
				baseQuestions = append(baseQuestions, item)
			}
		default:
			if len(advancedQuestions) < 4 {
				advancedQuestions = append(advancedQuestions, item)
			}
		}
	}

	return &PracticeWrongRemedialResponse{
		Role:              positionCode,
		PositionCode:      positionCode,
		Point:             point,
		Packet:            buildPointPacket(positionCode, point),
		BaseQuestions:     baseQuestions,
		AdvancedQuestions: advancedQuestions,
	}, nil
}

func (s *practiceService) GetSolution(ctx context.Context, questionID uint) (*PracticeSolutionResponse, error) {
	return s.getSolution(ctx, questionID)
}

func (s *practiceService) StartAssessment(ctx context.Context, userID uint, req PracticeAssessmentStartRequest) (*PracticeAssessmentSession, error) {
	positionCode, err := requiredPracticePositionCode(req.PositionCode)
	if err != nil {
		return nil, err
	}

	return s.startAssessment(ctx, userID, PracticeAssessmentStartRequest{
		PositionCode: positionCode,
		Difficulty:   req.Difficulty,
		TotalCount:   req.TotalCount,
	})
}

func (s *practiceService) SubmitAssessmentAnswer(ctx context.Context, userID uint, req PracticeAssessmentAnswerRequest) (*PracticeAnswerFeedback, error) {
	return s.submitAssessmentAnswer(ctx, userID, req)
}

func (s *practiceService) CompleteAssessment(ctx context.Context, userID uint, assessmentID uint) (*PracticeAssessmentSummary, error) {
	return s.completeAssessment(ctx, userID, assessmentID)
}

func (s *practiceService) GetIntegrationSnapshot(ctx context.Context, userID uint, positionCode string) (*PracticeIntegrationSnapshotResponse, error) {
	if userID == 0 {
		return nil, unauthorizedPracticeError("unauthorized")
	}

	normalized, err := optionalPracticePositionCode(positionCode)
	if err != nil {
		return nil, err
	}

	roleStats, err := s.repo.ListRoleProgressStats(ctx, userID, normalized)
	if err != nil {
		return nil, practiceInternalError("failed to load role snapshot", err)
	}
	pointStats, err := s.repo.ListPointMasteryStats(ctx, userID, normalized)
	if err != nil {
		return nil, practiceInternalError("failed to load point snapshot", err)
	}

	roleItems := make([]PracticeRoleSnapshotItem, 0, len(roleStats))
	for _, item := range roleStats {
		roleItems = append(roleItems, PracticeRoleSnapshotItem{
			Role:          item.PositionCode,
			PositionCode:  item.PositionCode,
			TotalAttempts: item.AttemptCount,
			Accuracy:      calcAccuracy(item.CorrectCount, item.AttemptCount),
		})
	}

	pointItems := make([]PracticePointMasteryItem, 0, len(pointStats))
	for _, item := range pointStats {
		pointItems = append(pointItems, PracticePointMasteryItem{
			Role:         item.PositionCode,
			PositionCode: item.PositionCode,
			Point:        cleanQuestionPayloadText(item.Point),
			Mastery:      calcAccuracy(item.Correct, item.Total),
		})
	}

	resp := &PracticeIntegrationSnapshotResponse{
		GeneratedAt:  time.Now().Format(time.RFC3339),
		RoleStats:    roleItems,
		PointMastery: pointItems,
	}
	if err := s.logPracticeSync(ctx, userID, "drill_to_interview", resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *practiceService) SubmitIntegrationFeedback(ctx context.Context, userID uint, req PracticeIntegrationFeedbackRequest) (*PracticeIntegrationFeedbackResponse, error) {
	if userID == 0 {
		return nil, unauthorizedPracticeError("unauthorized")
	}

	positionCode, err := requiredPracticePositionCode(req.PositionCode)
	if err != nil {
		return nil, err
	}
	weakPoints := sanitizeStringSlice(req.WeakPoints)
	if len(weakPoints) == 0 {
		return nil, invalidPracticeArgument("weak_points is required")
	}

	if err := s.logPracticeSync(ctx, userID, "interview_to_drill", req); err != nil {
		return nil, err
	}

	sets := make([]PracticeRemedialSet, 0, len(weakPoints))
	for idx, point := range weakPoints {
		if idx >= 5 {
			break
		}

		questions, err := s.repo.ListRemedialQuestions(ctx, positionCode, point, 0, 4)
		if err != nil {
			return nil, practiceInternalError("failed to load remedial feedback", err)
		}

		items := make([]PracticeRemedialQuestion, 0, len(questions))
		for _, question := range questions {
			items = append(items, PracticeRemedialQuestion{
				ID:         question.ID,
				Stem:       cleanQuestionPayloadText(question.EffectiveStem()),
				Level:      cleanQuestionPayloadText(question.Level),
				Difficulty: question.DifficultyScore,
			})
		}

		sets = append(sets, PracticeRemedialSet{
			Point:     point,
			Questions: items,
		})
	}

	return &PracticeIntegrationFeedbackResponse{
		Role:         positionCode,
		PositionCode: positionCode,
		RemedialSets: sets,
	}, nil
}

func (s *practiceService) ListWrongs(ctx context.Context, userID uint, req PracticeWrongsRequest) ([]PracticeWrongBookItem, error) {
	if userID == 0 {
		return nil, unauthorizedPracticeError("unauthorized")
	}

	normalized, err := optionalPracticePositionCode(req.PositionCode)
	if err != nil {
		return nil, err
	}

	items, err := s.repo.ListWrongBookEntries(ctx, userID, repository.PracticeWrongBookFilter{
		PositionCode: normalized,
		Point:        cleanQuestionPayloadText(req.Point),
		QuestionType: cleanQuestionPayloadText(req.QuestionType),
		FavoriteOnly: req.FavoriteOnly,
	})
	if err != nil {
		return nil, practiceInternalError("failed to load wrong questions", err)
	}

	result := make([]PracticeWrongBookItem, 0, len(items))
	for _, item := range items {
		options := sanitizeQuestionOptions(item.Question.GetOptions())
		role := normalizeQuestionBankPositionCode(item.Question.PositionCode)
		result = append(result, PracticeWrongBookItem{
			WrongID:         item.ID,
			ID:              item.Question.ID,
			Role:            role,
			PositionCode:    role,
			Level:           cleanQuestionPayloadText(item.Question.Level),
			QuestionType:    cleanQuestionPayloadText(item.Question.QuestionType),
			Specialty:       cleanQuestionPayloadText(item.Question.Specialty),
			Stem:            cleanQuestionPayloadText(item.Question.EffectiveStem()),
			Points:          cleanQuestionPayloadText(item.Question.Points),
			CompanyType:     cleanQuestionPayloadText(item.Question.CompanyType),
			DifficultyScore: item.Question.DifficultyScore,
			HasOptions:      len(options) > 0,
			LastUserAnswer:  cleanQuestionPayloadText(item.LastUserAnswer),
			ErrorReason:     cleanQuestionPayloadText(item.ErrorReason),
			IsFavorite:      item.IsFavorite,
			UpdatedAt:       item.UpdatedAt,
		})
	}
	return result, nil
}

func (s *practiceService) DeleteWrong(ctx context.Context, userID, wrongID uint) error {
	if userID == 0 {
		return unauthorizedPracticeError("unauthorized")
	}
	if err := s.repo.DeleteWrongBookEntryByID(ctx, userID, wrongID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return practiceNotFound("wrong question not found", err)
		}
		return practiceInternalError("failed to delete wrong question", err)
	}
	return nil
}

func (s *practiceService) ToggleWrongFavorite(ctx context.Context, userID, wrongID uint) (bool, error) {
	if userID == 0 {
		return false, unauthorizedPracticeError("unauthorized")
	}
	state, err := s.repo.ToggleWrongBookFavoriteByID(ctx, userID, wrongID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, practiceNotFound("wrong question not found", err)
		}
		return false, practiceInternalError("failed to toggle wrong favorite", err)
	}
	return state, nil
}

func (s *practiceService) ToggleQuestionFavorite(ctx context.Context, userID, questionID uint) (bool, error) {
	if userID == 0 {
		return false, unauthorizedPracticeError("unauthorized")
	}
	if _, err := s.repo.GetQuestionByID(ctx, questionID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, practiceNotFound("question not found", err)
		}
		return false, practiceInternalError("failed to load question", err)
	}
	state, err := s.repo.ToggleFavorite(ctx, userID, questionID)
	if err != nil {
		return false, practiceInternalError("failed to toggle question favorite", err)
	}
	return state, nil
}

func (s *practiceService) GetDashboard(ctx context.Context, userID uint, positionCode string) (*PracticeDashboardResponse, error) {
	if userID == 0 {
		return nil, unauthorizedPracticeError("unauthorized")
	}

	radarPosition, err := optionalPracticePositionCode(positionCode)
	if err != nil {
		return nil, err
	}

	totals, err := s.repo.GetAttemptTotals(ctx, userID)
	if err != nil {
		return nil, practiceInternalError("failed to load practice totals", err)
	}
	roleStats, err := s.repo.ListRoleProgressStats(ctx, userID, "")
	if err != nil {
		return nil, practiceInternalError("failed to load role progress", err)
	}

	roleProgress := make(map[string]PracticeDashboardRoleProgress, len(roleStats))
	for _, item := range roleStats {
		progress := 0.0
		if item.TotalQuestions > 0 {
			progress = roundFloat(float64(item.SolvedUnique*100)/float64(item.TotalQuestions), 2)
		}
		roleProgress[item.PositionCode] = PracticeDashboardRoleProgress{
			AttemptCount: item.AttemptCount,
			Accuracy:     calcAccuracy(item.CorrectCount, item.AttemptCount),
			Progress:     progress,
		}
	}

	startDay := time.Now().AddDate(0, 0, -13)
	dailyStats, err := s.repo.ListDailyAttemptStats(ctx, userID, startDay)
	if err != nil {
		return nil, practiceInternalError("failed to load practice trend", err)
	}
	dailyMap := make(map[string]repository.PracticeDailyAttemptStat, len(dailyStats))
	for _, item := range dailyStats {
		dailyMap[item.Day] = item
	}

	trend := make([]PracticeDashboardTrendItem, 0, 14)
	for i := 0; i < 14; i++ {
		day := startDay.AddDate(0, 0, i).Format("2006-01-02")
		if stat, ok := dailyMap[day]; ok {
			trend = append(trend, PracticeDashboardTrendItem{
				Day:      day,
				Count:    stat.Total,
				Accuracy: calcAccuracy(stat.Correct, stat.Total),
			})
			continue
		}
		trend = append(trend, PracticeDashboardTrendItem{Day: day})
	}

	pointStats, err := s.repo.ListPointMasteryStats(ctx, userID, radarPosition)
	if err != nil {
		return nil, practiceInternalError("failed to load radar data", err)
	}

	radar := make([]PracticeDashboardRadarItem, 0, len(pointStats))
	for _, item := range pointStats {
		radar = append(radar, PracticeDashboardRadarItem{
			Dimension: cleanQuestionPayloadText(item.Point),
			Mastery:   calcAccuracy(item.Correct, item.Total),
		})
	}

	return &PracticeDashboardResponse{
		TotalAttempts: totals.TotalAttempts,
		Accuracy:      calcAccuracy(totals.CorrectAttempts, totals.TotalAttempts),
		RoleProgress:  roleProgress,
		Trend:         trend,
		Radar:         radar,
	}, nil
}

func (s *practiceService) ImportQuestionBank(ctx context.Context, req PracticeImportRequest) (int, error) {
	if len(req.Items) == 0 {
		return 0, invalidPracticeArgument("items is required")
	}

	questions := make([]*model.Question, 0, len(req.Items))
	for _, item := range req.Items {
		positionCode, ok := normalizeImportPositionCode(item.PositionCode, item.Role)
		if !ok {
			continue
		}

		level := strings.TrimSpace(strings.ToLower(item.Level))
		if _, exists := buildPracticeLevelMap()[level]; !exists {
			continue
		}

		questionType := cleanQuestionPayloadText(item.QuestionType)
		if !isAllowedPracticeQuestionType(questionType) {
			continue
		}

		stem := cleanQuestionPayloadText(item.Stem)
		answer := cleanQuestionPayloadText(item.Answer)
		analysis := cleanQuestionPayloadText(item.Analysis)
		points := cleanQuestionPayloadText(item.Points)
		if stem == "" || answer == "" || analysis == "" || points == "" {
			continue
		}

		question := &model.Question{
			PositionCode:    positionCode,
			Position:        positionDisplayName(positionCode),
			Level:           level,
			QuestionType:    questionType,
			Specialty:       cleanQuestionPayloadText(item.Specialty),
			Stem:            stem,
			Content:         stem,
			StandardAnswer:  answer,
			ExpectedAnswer:  answer,
			Analysis:        analysis,
			Tips:            cleanQuestionPayloadText(item.Tips),
			Exemplar:        cleanQuestionPayloadText(item.Exemplar),
			Points:          points,
			KnowledgePoint:  points,
			CompanyType:     cleanQuestionPayloadText(item.CompanyType),
			DifficultyScore: item.DifficultyScore,
			Source:          "question_bank",
			IsActive:        true,
		}
		question.SetOptions(item.Options)
		questions = append(questions, question)
	}

	if len(questions) == 0 {
		return 0, invalidPracticeArgument("no valid questions to import")
	}

	if err := s.repo.BatchCreateQuestions(ctx, questions); err != nil {
		return 0, practiceInternalError("failed to import question bank", err)
	}
	return len(questions), nil
}

func (s *practiceService) ExportRecords(ctx context.Context, userID uint) (*PracticeRecordExport, error) {
	if userID == 0 {
		return nil, unauthorizedPracticeError("unauthorized")
	}

	rows, err := s.repo.ListPracticeRecordExports(ctx, userID)
	if err != nil {
		return nil, practiceInternalError("failed to export practice records", err)
	}

	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	if err := writer.Write([]string{"id", "created_at", "role", "level", "question_type", "points", "user_answer", "is_correct", "error_reason"}); err != nil {
		return nil, practiceInternalError("failed to build export file", err)
	}
	for _, row := range rows {
		record := []string{
			fmt.Sprintf("%d", row.ID),
			row.CreatedAt.Format("2006-01-02 15:04:05"),
			row.PositionCode,
			row.Level,
			row.QuestionType,
			row.Point,
			row.UserAnswer,
			fmt.Sprintf("%t", row.IsCorrect),
			row.ErrorReason,
		}
		if err := writer.Write(record); err != nil {
			return nil, practiceInternalError("failed to build export file", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, practiceInternalError("failed to build export file", err)
	}

	return &PracticeRecordExport{
		Filename: fmt.Sprintf("practice_records_%s.csv", time.Now().Format("20060102_150405")),
		Content:  buffer.String(),
	}, nil
}

func (s *practiceService) GetQuestionByID(ctx context.Context, userID, questionID uint) (*PracticeQuestionEnvelope, error) {
	if userID == 0 {
		return nil, unauthorizedPracticeError("unauthorized")
	}

	row, err := s.repo.GetQuestionDetail(ctx, userID, questionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, practiceNotFound("question not found", err)
		}
		return nil, practiceInternalError("failed to load question", err)
	}
	return &PracticeQuestionEnvelope{Question: buildPracticeQuestionDetail(*row)}, nil
}

func normalizePracticeQuestionQuery(req PracticeListQuestionsRequest) (repository.PracticeQuestionQuery, error) {
	positionCode, err := optionalPracticePositionCode(req.PositionCode)
	if err != nil {
		return repository.PracticeQuestionQuery{}, err
	}

	status := strings.TrimSpace(strings.ToLower(req.Status))
	if status != "todo" && status != "solved" && status != "wrong" {
		status = ""
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 15
	}
	if pageSize > 50 {
		pageSize = 50
	}

	return repository.PracticeQuestionQuery{
		Filters: repository.PracticeQuestionFilters{
			PositionCode: positionCode,
			Level:        strings.TrimSpace(strings.ToLower(req.Level)),
			QuestionType: cleanQuestionPayloadText(req.QuestionType),
			Specialty:    cleanQuestionPayloadText(req.Specialty),
			Point:        cleanQuestionPayloadText(req.Point),
			CompanyType:  cleanQuestionPayloadText(req.CompanyType),
			Keyword:      cleanQuestionPayloadText(req.Keyword),
		},
		Status:       status,
		FavoriteOnly: req.FavoriteOnly,
		ListID:       req.ListID,
		Page:         page,
		PageSize:     pageSize,
	}, nil
}

func buildPracticeRoleMetaMap() map[string]PracticeRoleMeta {
	result := make(map[string]PracticeRoleMeta, len(model.DefaultJobPositions))
	for _, item := range model.DefaultJobPositions {
		code := normalizeQuestionBankPositionCode(item.Code)
		result[code] = PracticeRoleMeta{
			Name:        cleanQuestionPayloadText(item.Name),
			Focus:       []string{},
			Specialties: []string{},
		}
	}
	return result
}

func buildPracticeLevelMap() map[string]string {
	result := make(map[string]string, len(questionBankLevels))
	for _, item := range questionBankLevels {
		result[item.Key] = cleanQuestionPayloadText(item.Label)
	}
	return result
}

func optionalPracticePositionCode(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", nil
	}
	return requiredPracticePositionCode(value)
}

func requiredPracticePositionCode(value string) (string, error) {
	normalized := normalizeQuestionBankPositionCode(value)
	if normalized == "" {
		return "", invalidPracticeArgument("position_code is required")
	}
	if !isKnownPracticePosition(normalized) {
		return "", invalidPracticeArgument("invalid position_code")
	}
	return normalized, nil
}

func isKnownPracticePosition(value string) bool {
	for _, item := range model.DefaultJobPositions {
		if normalizeQuestionBankPositionCode(item.Code) == value {
			return true
		}
	}
	return false
}

func normalizeImportPositionCode(positionCode, role string) (string, bool) {
	raw := positionCode
	if strings.TrimSpace(raw) == "" {
		raw = role
	}
	normalized := normalizeQuestionBankPositionCode(raw)
	if normalized == "" || !isKnownPracticePosition(normalized) {
		return "", false
	}
	return normalized, true
}

func isAllowedPracticeQuestionType(value string) bool {
	for _, item := range questionBankQuestionTypes {
		if item == value {
			return true
		}
	}
	return false
}

func buildPracticeQuestionSummary(row repository.PracticeQuestionListRow) PracticeQuestionSummary {
	options := sanitizeQuestionOptions(row.Question.GetOptions())
	role := normalizeQuestionBankPositionCode(row.Question.PositionCode)
	return PracticeQuestionSummary{
		ID:              row.Question.ID,
		Role:            role,
		PositionCode:    role,
		Level:           cleanQuestionPayloadText(row.Question.Level),
		QuestionType:    cleanQuestionPayloadText(row.Question.QuestionType),
		Specialty:       cleanQuestionPayloadText(row.Question.Specialty),
		Points:          cleanQuestionPayloadText(row.Question.Points),
		CompanyType:     cleanQuestionPayloadText(row.Question.CompanyType),
		DifficultyScore: row.Question.DifficultyScore,
		Stem:            cleanQuestionPayloadText(row.Question.EffectiveStem()),
		HasOptions:      len(options) > 0,
		Status:          cleanQuestionPayloadText(row.Status),
		IsFavorite:      row.IsFavorite,
	}
}

func buildPracticeQuestionDetail(row repository.PracticeQuestionListRow) PracticeQuestionDetail {
	options := sanitizeQuestionOptions(row.Question.GetOptions())
	role := normalizeQuestionBankPositionCode(row.Question.PositionCode)
	return PracticeQuestionDetail{
		ID:              row.Question.ID,
		Role:            role,
		PositionCode:    role,
		Position:        cleanQuestionPayloadText(row.Question.Position),
		Difficulty:      cleanQuestionPayloadText(row.Question.Difficulty),
		Level:           cleanQuestionPayloadText(row.Question.Level),
		QuestionType:    cleanQuestionPayloadText(row.Question.QuestionType),
		Specialty:       cleanQuestionPayloadText(row.Question.Specialty),
		Stem:            cleanQuestionPayloadText(row.Question.EffectiveStem()),
		Options:         options,
		Points:          cleanQuestionPayloadText(row.Question.Points),
		CompanyType:     cleanQuestionPayloadText(row.Question.CompanyType),
		DifficultyScore: row.Question.DifficultyScore,
		HasOptions:      len(options) > 0,
		Status:          cleanQuestionPayloadText(row.Status),
		IsFavorite:      row.IsFavorite,
	}
}

func calcAccuracy(correct, total int64) float64 {
	if total <= 0 {
		return 0
	}
	return roundFloat(float64(correct*100)/float64(total), 2)
}

func (s *practiceService) logPracticeSync(ctx context.Context, userID uint, source string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return practiceInternalError("failed to serialize sync payload", err)
	}
	uid := userID
	if err := s.repo.CreateSyncLog(ctx, &model.PracticeInterviewSyncLog{
		UserID:      &uid,
		Source:      source,
		PayloadJSON: string(raw),
	}); err != nil {
		return practiceInternalError("failed to persist sync payload", err)
	}
	return nil
}
