package repository

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"your-project/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PracticeQuestionFilters struct {
	PositionCode  string
	Level         string
	QuestionType  string
	QuestionTypes []string
	Specialty     string
	Point         string
	CompanyType   string
	Keyword       string
}

type PracticeQuestionQuery struct {
	Filters      PracticeQuestionFilters
	Status       string
	FavoriteOnly bool
	ListID       uint
	Page         int
	PageSize     int
}

type PracticeQuestionStatusCounts struct {
	Todo   int64
	Solved int64
	Wrong  int64
}

type PracticeQuestionListRow struct {
	Question   model.Question
	Status     string
	IsFavorite bool
}

type PracticeQuestionListSummary struct {
	List        model.PracticeQuestionList
	TotalCount  int64
	SolvedCount int64
}

type PracticeWrongBookFilter struct {
	PositionCode string
	Point        string
	QuestionType string
	FavoriteOnly bool
}

type PracticeAttemptTotals struct {
	TotalAttempts   int64
	CorrectAttempts int64
}

type PracticeRoleAttemptStat struct {
	PositionCode   string
	AttemptCount   int64
	CorrectCount   int64
	SolvedUnique   int64
	TotalQuestions int64
}

type PracticeDailyAttemptStat struct {
	Day     string
	Total   int64
	Correct int64
}

type PracticePointMasteryStat struct {
	PositionCode string
	Point        string
	Total        int64
	Correct      int64
}

type PracticeRecordExportRow struct {
	ID           uint
	CreatedAt    time.Time
	PositionCode string
	Level        string
	QuestionType string
	Point        string
	UserAnswer   string
	IsCorrect    bool
	ErrorReason  string
}

type PracticeQuestionRepository interface {
	CountQuestionsByPosition(ctx context.Context) (map[string]int64, error)
	ListQuestions(ctx context.Context, filters PracticeQuestionFilters) ([]model.Question, error)
	ListQuestionPage(ctx context.Context, userID uint, query PracticeQuestionQuery) ([]PracticeQuestionListRow, int64, PracticeQuestionStatusCounts, error)
	DrawRandomQuestion(ctx context.Context, filters PracticeQuestionFilters) (*model.Question, error)
	GetQuestionByID(ctx context.Context, questionID uint) (*model.Question, error)
	GetQuestionDetail(ctx context.Context, userID, questionID uint) (*PracticeQuestionListRow, error)
	ListDistinctPoints(ctx context.Context, positionCode string) ([]string, error)
	ListDistinctSpecialties(ctx context.Context, positionCode string) ([]string, error)

	CreatePracticeRecord(ctx context.Context, record *model.QuestionPracticeRecord) error
	LatestAttemptRecords(ctx context.Context, userID uint, questionIDs []uint, assessmentID *uint) (map[uint]model.QuestionPracticeRecord, error)
	CountQuestionsByPoint(ctx context.Context, positionCode, point string) (int64, error)
	ListSolvedQuestionIDsByPoint(ctx context.Context, userID uint, positionCode, point string) ([]uint, error)

	CreateAssessment(ctx context.Context, assessment *model.QuestionAssessment, items []model.QuestionAssessmentItem) error
	GetAssessment(ctx context.Context, assessmentID, userID uint, withItems bool) (*model.QuestionAssessment, error)
	GetAssessmentItem(ctx context.Context, assessmentID, questionID uint) (*model.QuestionAssessmentItem, error)
	UpdateAssessment(ctx context.Context, assessment *model.QuestionAssessment) error
	UpsertAssessmentAnswer(ctx context.Context, answer *model.PracticeAssessmentAnswer) error

	UpsertWrongBookEntry(ctx context.Context, entry *model.PracticeWrongBookEntry) error
	GetWrongBookEntry(ctx context.Context, userID, wrongID uint) (*model.PracticeWrongBookEntry, error)
	DeleteWrongBookEntry(ctx context.Context, userID, questionID uint) error
	DeleteWrongBookEntryByID(ctx context.Context, userID, wrongID uint) error
	ListWrongBookEntries(ctx context.Context, userID uint, filter PracticeWrongBookFilter) ([]model.PracticeWrongBookEntry, error)
	ToggleWrongBookFavoriteByID(ctx context.Context, userID, wrongID uint) (bool, error)
	ToggleFavorite(ctx context.Context, userID, questionID uint) (bool, error)
	GetQuestionFavoriteState(ctx context.Context, userID, questionID uint) (bool, error)

	CreateQuestionList(ctx context.Context, list *model.PracticeQuestionList) error
	UpdateQuestionList(ctx context.Context, list *model.PracticeQuestionList) error
	DeleteQuestionList(ctx context.Context, listID uint) error
	GetQuestionList(ctx context.Context, listID uint) (*model.PracticeQuestionList, error)
	ReplaceQuestionListItems(ctx context.Context, listID uint, items []model.PracticeQuestionListItem) error
	ListQuestionLists(ctx context.Context, positionCode string) ([]model.PracticeQuestionList, error)
	ListQuestionListsWithProgress(ctx context.Context, userID uint, positionCode string) ([]PracticeQuestionListSummary, error)

	ListRemedialQuestions(ctx context.Context, positionCode, point string, excludeQuestionID uint, limit int) ([]model.Question, error)
	BatchCreateQuestions(ctx context.Context, questions []*model.Question) error
	ListPracticeRecordExports(ctx context.Context, userID uint) ([]PracticeRecordExportRow, error)
	GetAttemptTotals(ctx context.Context, userID uint) (PracticeAttemptTotals, error)
	ListRoleProgressStats(ctx context.Context, userID uint, positionCode string) ([]PracticeRoleAttemptStat, error)
	ListDailyAttemptStats(ctx context.Context, userID uint, fromDate time.Time) ([]PracticeDailyAttemptStat, error)
	ListPointMasteryStats(ctx context.Context, userID uint, positionCode string) ([]PracticePointMasteryStat, error)

	CreateSyncLog(ctx context.Context, log *model.PracticeInterviewSyncLog) error
}

type GormPracticeQuestionRepository struct {
	db *gorm.DB
}

var _ PracticeQuestionRepository = (*GormPracticeQuestionRepository)(nil)

func NewPracticeQuestionRepository() PracticeQuestionRepository {
	return &GormPracticeQuestionRepository{db: GetDB()}
}

func NewPracticeQuestionRepositoryWithDB(db *gorm.DB) PracticeQuestionRepository {
	if db == nil {
		db = GetDB()
	}
	return &GormPracticeQuestionRepository{db: db}
}

func (r *GormPracticeQuestionRepository) CountQuestionsByPosition(ctx context.Context) (map[string]int64, error) {
	type row struct {
		PositionCode string
		Count        int64
	}

	var rows []row
	if err := r.baseQuestionQuery(ctx).
		Select("position_code, COUNT(*) AS count").
		Group("position_code").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make(map[string]int64, len(rows))
	for _, item := range rows {
		result[strings.TrimSpace(strings.ToLower(item.PositionCode))] = item.Count
	}
	return result, nil
}

func (r *GormPracticeQuestionRepository) ListQuestions(ctx context.Context, filters PracticeQuestionFilters) ([]model.Question, error) {
	var questions []model.Question
	err := r.applyQuestionFilters(r.baseQuestionQuery(ctx), filters).
		Order("difficulty_score DESC, id ASC").
		Find(&questions).Error
	return questions, err
}

func (r *GormPracticeQuestionRepository) ListQuestionPage(ctx context.Context, userID uint, query PracticeQuestionQuery) ([]PracticeQuestionListRow, int64, PracticeQuestionStatusCounts, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 15
	}
	if query.PageSize > 50 {
		query.PageSize = 50
	}

	questions, err := r.listQuestionsForQuery(ctx, query.Filters, query.ListID)
	if err != nil {
		return nil, 0, PracticeQuestionStatusCounts{}, err
	}

	questionIDs := collectPracticeQuestionIDs(questions)
	recordMap, err := r.LatestAttemptRecords(ctx, userID, questionIDs, nil)
	if err != nil {
		return nil, 0, PracticeQuestionStatusCounts{}, err
	}
	favoriteMap, err := r.favoriteQuestionIDSet(ctx, userID, questionIDs)
	if err != nil {
		return nil, 0, PracticeQuestionStatusCounts{}, err
	}

	rows := make([]PracticeQuestionListRow, 0, len(questions))
	stats := PracticeQuestionStatusCounts{}
	for _, question := range questions {
		isFavorite := favoriteMap[question.ID]
		if query.FavoriteOnly && !isFavorite {
			continue
		}

		status := resolvePracticeQuestionStatus(recordMap, question.ID)
		switch status {
		case "solved":
			stats.Solved++
		case "wrong":
			stats.Wrong++
		default:
			stats.Todo++
		}

		if query.Status != "" && query.Status != status {
			continue
		}

		rows = append(rows, PracticeQuestionListRow{
			Question:   question,
			Status:     status,
			IsFavorite: isFavorite,
		})
	}

	total := int64(len(rows))
	start := (query.Page - 1) * query.PageSize
	if start > len(rows) {
		start = len(rows)
	}
	end := start + query.PageSize
	if end > len(rows) {
		end = len(rows)
	}
	return rows[start:end], total, stats, nil
}

func (r *GormPracticeQuestionRepository) DrawRandomQuestion(ctx context.Context, filters PracticeQuestionFilters) (*model.Question, error) {
	var question model.Question
	err := r.applyQuestionFilters(r.baseQuestionQuery(ctx), filters).
		Order(r.randomOrderExpression()).
		Limit(1).
		Take(&question).Error
	if err != nil {
		return nil, err
	}
	return &question, nil
}

func (r *GormPracticeQuestionRepository) GetQuestionByID(ctx context.Context, questionID uint) (*model.Question, error) {
	var question model.Question
	if err := r.baseQuestionQuery(ctx).Where("id = ?", questionID).First(&question).Error; err != nil {
		return nil, err
	}
	return &question, nil
}

func (r *GormPracticeQuestionRepository) GetQuestionDetail(ctx context.Context, userID, questionID uint) (*PracticeQuestionListRow, error) {
	question, err := r.GetQuestionByID(ctx, questionID)
	if err != nil {
		return nil, err
	}

	recordMap, err := r.LatestAttemptRecords(ctx, userID, []uint{questionID}, nil)
	if err != nil {
		return nil, err
	}
	isFavorite, err := r.GetQuestionFavoriteState(ctx, userID, questionID)
	if err != nil {
		return nil, err
	}

	return &PracticeQuestionListRow{
		Question:   *question,
		Status:     resolvePracticeQuestionStatus(recordMap, questionID),
		IsFavorite: isFavorite,
	}, nil
}

func (r *GormPracticeQuestionRepository) ListDistinctPoints(ctx context.Context, positionCode string) ([]string, error) {
	var points []string
	query := r.baseQuestionQuery(ctx).Where("points <> ''")
	if normalized := strings.TrimSpace(strings.ToLower(positionCode)); normalized != "" {
		query = query.Where("position_code = ?", normalized)
	}
	if err := query.Distinct("points").Order("points ASC").Pluck("points", &points).Error; err != nil {
		return nil, err
	}
	return points, nil
}

func (r *GormPracticeQuestionRepository) ListDistinctSpecialties(ctx context.Context, positionCode string) ([]string, error) {
	var specialties []string
	query := r.baseQuestionQuery(ctx).Where("specialty <> ''")
	if normalized := strings.TrimSpace(strings.ToLower(positionCode)); normalized != "" {
		query = query.Where("position_code = ?", normalized)
	}
	if err := query.Distinct("specialty").Order("specialty ASC").Pluck("specialty", &specialties).Error; err != nil {
		return nil, err
	}
	return specialties, nil
}

func (r *GormPracticeQuestionRepository) CreatePracticeRecord(ctx context.Context, record *model.QuestionPracticeRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *GormPracticeQuestionRepository) LatestAttemptRecords(ctx context.Context, userID uint, questionIDs []uint, assessmentID *uint) (map[uint]model.QuestionPracticeRecord, error) {
	result := make(map[uint]model.QuestionPracticeRecord)
	if userID == 0 || len(questionIDs) == 0 {
		return result, nil
	}

	subQuery := r.db.WithContext(ctx).
		Model(&model.QuestionPracticeRecord{}).
		Select("MAX(id)").
		Where("user_id = ?", userID).
		Where("question_id IN ?", questionIDs).
		Group("question_id")
	if assessmentID != nil && *assessmentID > 0 {
		subQuery = subQuery.Where("assessment_id = ?", *assessmentID)
	}

	var records []model.QuestionPracticeRecord
	if err := r.db.WithContext(ctx).Where("id IN (?)", subQuery).Find(&records).Error; err != nil {
		return nil, err
	}

	for _, record := range records {
		result[record.QuestionID] = record
	}
	return result, nil
}

func (r *GormPracticeQuestionRepository) CountQuestionsByPoint(ctx context.Context, positionCode, point string) (int64, error) {
	var total int64
	err := r.baseQuestionQuery(ctx).
		Where("position_code = ?", strings.TrimSpace(strings.ToLower(positionCode))).
		Where("points = ? OR knowledge_point = ?", point, point).
		Count(&total).Error
	return total, err
}

func (r *GormPracticeQuestionRepository) ListSolvedQuestionIDsByPoint(ctx context.Context, userID uint, positionCode, point string) ([]uint, error) {
	var questionIDs []uint
	err := r.db.WithContext(ctx).
		Model(&model.QuestionPracticeRecord{}).
		Distinct("question_practice_records.question_id").
		Joins("JOIN questions q ON q.id = question_practice_records.question_id").
		Where("question_practice_records.user_id = ? AND question_practice_records.is_correct = ?", userID, true).
		Where("q.position_code = ?", strings.TrimSpace(strings.ToLower(positionCode))).
		Where("q.points = ? OR q.knowledge_point = ?", point, point).
		Pluck("question_practice_records.question_id", &questionIDs).Error
	return questionIDs, err
}

func (r *GormPracticeQuestionRepository) CreateAssessment(ctx context.Context, assessment *model.QuestionAssessment, items []model.QuestionAssessmentItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(assessment).Error; err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		return tx.Create(&items).Error
	})
}

func (r *GormPracticeQuestionRepository) GetAssessment(ctx context.Context, assessmentID, userID uint, withItems bool) (*model.QuestionAssessment, error) {
	var assessment model.QuestionAssessment
	query := r.db.WithContext(ctx).Model(&model.QuestionAssessment{}).Where("id = ? AND user_id = ?", assessmentID, userID)
	if withItems {
		query = query.Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("order_no ASC")
		}).Preload("Items.Question")
	}
	if err := query.First(&assessment).Error; err != nil {
		return nil, err
	}
	return &assessment, nil
}

func (r *GormPracticeQuestionRepository) GetAssessmentItem(ctx context.Context, assessmentID, questionID uint) (*model.QuestionAssessmentItem, error) {
	var item model.QuestionAssessmentItem
	if err := r.db.WithContext(ctx).
		Where("assessment_id = ? AND question_id = ?", assessmentID, questionID).
		First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormPracticeQuestionRepository) UpdateAssessment(ctx context.Context, assessment *model.QuestionAssessment) error {
	return r.db.WithContext(ctx).Save(assessment).Error
}

func (r *GormPracticeQuestionRepository) UpsertAssessmentAnswer(ctx context.Context, answer *model.PracticeAssessmentAnswer) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "assessment_id"},
			{Name: "question_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"user_id",
			"user_answer",
			"is_correct",
			"elapsed_seconds",
			"is_timeout",
			"updated_at",
		}),
	}).Create(answer).Error
}

func (r *GormPracticeQuestionRepository) UpsertWrongBookEntry(ctx context.Context, entry *model.PracticeWrongBookEntry) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "user_id"},
			{Name: "question_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"last_user_answer",
			"error_reason",
			"updated_at",
		}),
	}).Create(entry).Error
}

func (r *GormPracticeQuestionRepository) GetWrongBookEntry(ctx context.Context, userID, wrongID uint) (*model.PracticeWrongBookEntry, error) {
	var item model.PracticeWrongBookEntry
	if err := r.db.WithContext(ctx).
		Preload("Question").
		Where("id = ? AND user_id = ?", wrongID, userID).
		First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GormPracticeQuestionRepository) DeleteWrongBookEntry(ctx context.Context, userID, questionID uint) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND question_id = ?", userID, questionID).
		Delete(&model.PracticeWrongBookEntry{}).Error
}

func (r *GormPracticeQuestionRepository) DeleteWrongBookEntryByID(ctx context.Context, userID, wrongID uint) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", wrongID, userID).
		Delete(&model.PracticeWrongBookEntry{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormPracticeQuestionRepository) ListWrongBookEntries(ctx context.Context, userID uint, filter PracticeWrongBookFilter) ([]model.PracticeWrongBookEntry, error) {
	var items []model.PracticeWrongBookEntry
	query := r.db.WithContext(ctx).
		Model(&model.PracticeWrongBookEntry{}).
		Preload("Question").
		Joins("JOIN questions q ON q.id = question_wrong_books.question_id").
		Where("question_wrong_books.user_id = ?", userID)

	if normalized := strings.TrimSpace(strings.ToLower(filter.PositionCode)); normalized != "" {
		query = query.Where("q.position_code = ?", normalized)
	}
	if point := strings.TrimSpace(filter.Point); point != "" {
		like := "%" + point + "%"
		query = query.Where("q.points LIKE ? OR q.knowledge_point LIKE ?", like, like)
	}
	if questionType := strings.TrimSpace(filter.QuestionType); questionType != "" {
		query = query.Where("q.question_type = ?", questionType)
	}
	if filter.FavoriteOnly {
		query = query.Where("question_wrong_books.is_favorite = ?", true)
	}

	err := query.Order("question_wrong_books.updated_at DESC").Find(&items).Error
	return items, err
}

func (r *GormPracticeQuestionRepository) ToggleWrongBookFavoriteByID(ctx context.Context, userID, wrongID uint) (bool, error) {
	var wrong model.PracticeWrongBookEntry
	if err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", wrongID, userID).
		First(&wrong).Error; err != nil {
		return false, err
	}
	newState := !wrong.IsFavorite
	if err := r.db.WithContext(ctx).
		Model(&model.PracticeWrongBookEntry{}).
		Where("id = ?", wrong.ID).
		Update("is_favorite", newState).Error; err != nil {
		return false, err
	}
	return newState, nil
}

func (r *GormPracticeQuestionRepository) ToggleFavorite(ctx context.Context, userID, questionID uint) (bool, error) {
	var favorite model.PracticeQuestionFavorite
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND question_id = ?", userID, questionID).
		First(&favorite).Error
	switch {
	case err == nil:
		if err := r.db.WithContext(ctx).Delete(&favorite).Error; err != nil {
			return false, err
		}
		return false, nil
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return false, err
	}

	favorite = model.PracticeQuestionFavorite{
		UserID:     userID,
		QuestionID: questionID,
	}
	if err := r.db.WithContext(ctx).Create(&favorite).Error; err != nil {
		return false, err
	}
	return true, nil
}

func (r *GormPracticeQuestionRepository) GetQuestionFavoriteState(ctx context.Context, userID, questionID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.PracticeQuestionFavorite{}).
		Where("user_id = ? AND question_id = ?", userID, questionID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GormPracticeQuestionRepository) CreateQuestionList(ctx context.Context, list *model.PracticeQuestionList) error {
	return r.db.WithContext(ctx).Create(list).Error
}

func (r *GormPracticeQuestionRepository) UpdateQuestionList(ctx context.Context, list *model.PracticeQuestionList) error {
	return r.db.WithContext(ctx).Save(list).Error
}

func (r *GormPracticeQuestionRepository) DeleteQuestionList(ctx context.Context, listID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("list_id = ?", listID).Delete(&model.PracticeQuestionListItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.PracticeQuestionList{}, listID).Error
	})
}

func (r *GormPracticeQuestionRepository) GetQuestionList(ctx context.Context, listID uint) (*model.PracticeQuestionList, error) {
	var list model.PracticeQuestionList
	if err := r.db.WithContext(ctx).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("order_no ASC")
		}).
		Preload("Items.Question").
		First(&list, listID).Error; err != nil {
		return nil, err
	}
	return &list, nil
}

func (r *GormPracticeQuestionRepository) ReplaceQuestionListItems(ctx context.Context, listID uint, items []model.PracticeQuestionListItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("list_id = ?", listID).Delete(&model.PracticeQuestionListItem{}).Error; err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		return tx.Create(&items).Error
	})
}

func (r *GormPracticeQuestionRepository) ListQuestionLists(ctx context.Context, positionCode string) ([]model.PracticeQuestionList, error) {
	var lists []model.PracticeQuestionList
	query := r.db.WithContext(ctx).
		Model(&model.PracticeQuestionList{}).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("order_no ASC")
		})
	if normalized := strings.TrimSpace(strings.ToLower(positionCode)); normalized != "" {
		query = query.Where("position_code = ?", normalized)
	}
	err := query.Order("created_at ASC, id ASC").Find(&lists).Error
	return lists, err
}

func (r *GormPracticeQuestionRepository) ListQuestionListsWithProgress(ctx context.Context, userID uint, positionCode string) ([]PracticeQuestionListSummary, error) {
	lists, err := r.ListQuestionLists(ctx, positionCode)
	if err != nil {
		return nil, err
	}

	questionIDs := make([]uint, 0)
	for _, list := range lists {
		for _, item := range list.Items {
			if item.QuestionID == 0 {
				continue
			}
			questionIDs = append(questionIDs, item.QuestionID)
		}
	}

	recordMap, err := r.LatestAttemptRecords(ctx, userID, questionIDs, nil)
	if err != nil {
		return nil, err
	}

	summaries := make([]PracticeQuestionListSummary, 0, len(lists))
	for _, list := range lists {
		total := int64(len(list.Items))
		solved := int64(0)
		for _, item := range list.Items {
			record, ok := recordMap[item.QuestionID]
			if ok && record.IsCorrect {
				solved++
			}
		}
		summaries = append(summaries, PracticeQuestionListSummary{
			List:        list,
			TotalCount:  total,
			SolvedCount: solved,
		})
	}
	return summaries, nil
}

func (r *GormPracticeQuestionRepository) ListRemedialQuestions(ctx context.Context, positionCode, point string, excludeQuestionID uint, limit int) ([]model.Question, error) {
	if limit <= 0 {
		limit = 10
	}

	query := r.baseQuestionQuery(ctx).
		Where("position_code = ?", strings.TrimSpace(strings.ToLower(positionCode))).
		Where("points = ? OR knowledge_point = ?", point, point)
	if excludeQuestionID > 0 {
		query = query.Where("id <> ?", excludeQuestionID)
	}

	var questions []model.Question
	err := query.
		Order("CASE level WHEN 'base' THEN 1 WHEN 'advanced' THEN 2 ELSE 3 END ASC, difficulty_score ASC, id ASC").
		Limit(limit).
		Find(&questions).Error
	return questions, err
}

func (r *GormPracticeQuestionRepository) BatchCreateQuestions(ctx context.Context, questions []*model.Question) error {
	if len(questions) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&questions).Error
}

func (r *GormPracticeQuestionRepository) ListPracticeRecordExports(ctx context.Context, userID uint) ([]PracticeRecordExportRow, error) {
	pointExpr := practicePointExpression("q")

	var rows []PracticeRecordExportRow
	err := r.db.WithContext(ctx).
		Table("question_practice_records").
		Select(
			"question_practice_records.id, question_practice_records.created_at, q.position_code, q.level, q.question_type, "+pointExpr+" AS point, question_practice_records.user_answer, question_practice_records.is_correct, question_practice_records.error_reason",
		).
		Joins("JOIN questions q ON q.id = question_practice_records.question_id").
		Where("question_practice_records.user_id = ?", userID).
		Order("question_practice_records.created_at DESC, question_practice_records.id DESC").
		Scan(&rows).Error
	return rows, err
}

func (r *GormPracticeQuestionRepository) GetAttemptTotals(ctx context.Context, userID uint) (PracticeAttemptTotals, error) {
	type row struct {
		TotalAttempts   int64
		CorrectAttempts int64
	}

	var result row
	err := r.db.WithContext(ctx).
		Model(&model.QuestionPracticeRecord{}).
		Select("COUNT(*) AS total_attempts, COALESCE(SUM(CASE WHEN is_correct = 1 THEN 1 ELSE 0 END), 0) AS correct_attempts").
		Where("user_id = ?", userID).
		Scan(&result).Error
	if err != nil {
		return PracticeAttemptTotals{}, err
	}
	return PracticeAttemptTotals{
		TotalAttempts:   result.TotalAttempts,
		CorrectAttempts: result.CorrectAttempts,
	}, nil
}

func (r *GormPracticeQuestionRepository) ListRoleProgressStats(ctx context.Context, userID uint, positionCode string) ([]PracticeRoleAttemptStat, error) {
	counts, err := r.CountQuestionsByPosition(ctx)
	if err != nil {
		return nil, err
	}

	type row struct {
		PositionCode string
		AttemptCount int64
		CorrectCount int64
		SolvedUnique int64
	}

	query := r.db.WithContext(ctx).
		Table("questions q").
		Select("q.position_code, COUNT(r.id) AS attempt_count, COALESCE(SUM(CASE WHEN r.is_correct = 1 THEN 1 ELSE 0 END), 0) AS correct_count, COUNT(DISTINCT CASE WHEN r.is_correct = 1 THEN r.question_id ELSE NULL END) AS solved_unique").
		Joins("LEFT JOIN question_practice_records r ON r.question_id = q.id AND r.user_id = ?", userID).
		Where("q.is_active = ?", true).
		Where("q.source IS NULL OR q.source <> ?", "follow_up")
	if normalized := strings.TrimSpace(strings.ToLower(positionCode)); normalized != "" {
		query = query.Where("q.position_code = ?", normalized)
	}

	var rows []row
	if err := query.Group("q.position_code").Scan(&rows).Error; err != nil {
		return nil, err
	}

	rowMap := make(map[string]row, len(rows))
	for _, item := range rows {
		rowMap[strings.TrimSpace(strings.ToLower(item.PositionCode))] = item
	}

	normalizedFilter := strings.TrimSpace(strings.ToLower(positionCode))
	keys := make([]string, 0, len(counts))
	for key := range counts {
		if normalizedFilter != "" && key != normalizedFilter {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make([]PracticeRoleAttemptStat, 0, len(keys))
	for _, key := range keys {
		stat := rowMap[key]
		result = append(result, PracticeRoleAttemptStat{
			PositionCode:   key,
			AttemptCount:   stat.AttemptCount,
			CorrectCount:   stat.CorrectCount,
			SolvedUnique:   stat.SolvedUnique,
			TotalQuestions: counts[key],
		})
	}
	return result, nil
}

func (r *GormPracticeQuestionRepository) ListDailyAttemptStats(ctx context.Context, userID uint, fromDate time.Time) ([]PracticeDailyAttemptStat, error) {
	var rows []PracticeDailyAttemptStat
	err := r.db.WithContext(ctx).
		Model(&model.QuestionPracticeRecord{}).
		Select("DATE(created_at) AS day, COUNT(*) AS total, COALESCE(SUM(CASE WHEN is_correct = 1 THEN 1 ELSE 0 END), 0) AS correct").
		Where("user_id = ? AND DATE(created_at) >= ?", userID, fromDate.Format("2006-01-02")).
		Group("DATE(created_at)").
		Order("day ASC").
		Scan(&rows).Error
	return rows, err
}

func (r *GormPracticeQuestionRepository) ListPointMasteryStats(ctx context.Context, userID uint, positionCode string) ([]PracticePointMasteryStat, error) {
	pointExpr := practicePointExpression("q")

	query := r.db.WithContext(ctx).
		Table("questions q").
		Select(pointExpr+" AS point, q.position_code, COUNT(r.id) AS total, COALESCE(SUM(CASE WHEN r.is_correct = 1 THEN 1 ELSE 0 END), 0) AS correct").
		Joins("LEFT JOIN question_practice_records r ON r.question_id = q.id AND r.user_id = ?", userID).
		Where("q.is_active = ?", true).
		Where("q.source IS NULL OR q.source <> ?", "follow_up")
	if normalized := strings.TrimSpace(strings.ToLower(positionCode)); normalized != "" {
		query = query.Where("q.position_code = ?", normalized)
	}

	var rows []PracticePointMasteryStat
	err := query.
		Group("q.position_code, " + pointExpr).
		Order("q.position_code ASC, point ASC").
		Scan(&rows).Error
	return rows, err
}

func (r *GormPracticeQuestionRepository) CreateSyncLog(ctx context.Context, log *model.PracticeInterviewSyncLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *GormPracticeQuestionRepository) baseQuestionQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Model(&model.Question{}).
		Where("is_active = ?", true).
		Where("source IS NULL OR source <> ?", "follow_up")
}

func (r *GormPracticeQuestionRepository) applyQuestionFilters(query *gorm.DB, filters PracticeQuestionFilters) *gorm.DB {
	if positionCode := strings.TrimSpace(strings.ToLower(filters.PositionCode)); positionCode != "" {
		query = query.Where("position_code = ?", positionCode)
	}
	if level := strings.TrimSpace(strings.ToLower(filters.Level)); level != "" {
		query = query.Where("level = ?", level)
	}
	if questionType := strings.TrimSpace(filters.QuestionType); questionType != "" {
		query = query.Where("question_type = ?", questionType)
	}
	if questionTypes := normalizeQuestionTypes(filters.QuestionTypes); len(questionTypes) > 0 {
		query = query.Where("question_type IN ?", questionTypes)
	}
	if specialty := strings.TrimSpace(filters.Specialty); specialty != "" {
		query = query.Where("specialty = ?", specialty)
	}
	if point := strings.TrimSpace(filters.Point); point != "" {
		query = query.Where("points = ? OR knowledge_point = ?", point, point)
	}
	if companyType := strings.TrimSpace(filters.CompanyType); companyType != "" {
		query = query.Where("company_type = ?", companyType)
	}
	if keyword := strings.TrimSpace(filters.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where(
			"title LIKE ? OR stem LIKE ? OR content LIKE ? OR analysis LIKE ? OR points LIKE ? OR knowledge_point LIKE ?",
			like, like, like, like, like, like,
		)
	}
	return query
}

func (r *GormPracticeQuestionRepository) randomOrderExpression() string {
	if r.db != nil && r.db.Dialector != nil && strings.EqualFold(r.db.Dialector.Name(), "postgres") {
		return "RANDOM()"
	}
	return "RAND()"
}

func (r *GormPracticeQuestionRepository) listQuestionsForQuery(ctx context.Context, filters PracticeQuestionFilters, listID uint) ([]model.Question, error) {
	query := r.applyQuestionFilters(r.baseQuestionQuery(ctx), filters)
	if listID > 0 {
		query = query.Joins("JOIN question_list_items qli ON qli.question_id = questions.id AND qli.list_id = ?", listID)
	}

	var questions []model.Question
	err := query.Order("difficulty_score DESC, questions.id ASC").Find(&questions).Error
	return questions, err
}

func (r *GormPracticeQuestionRepository) favoriteQuestionIDSet(ctx context.Context, userID uint, questionIDs []uint) (map[uint]bool, error) {
	result := make(map[uint]bool)
	if userID == 0 || len(questionIDs) == 0 {
		return result, nil
	}

	var rows []model.PracticeQuestionFavorite
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND question_id IN ?", userID, questionIDs).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, item := range rows {
		result[item.QuestionID] = true
	}
	return result, nil
}

func collectPracticeQuestionIDs(items []model.Question) []uint {
	out := make([]uint, 0, len(items))
	for _, item := range items {
		if item.ID == 0 {
			continue
		}
		out = append(out, item.ID)
	}
	return out
}

func resolvePracticeQuestionStatus(recordMap map[uint]model.QuestionPracticeRecord, questionID uint) string {
	record, ok := recordMap[questionID]
	if !ok {
		return "todo"
	}
	if record.IsCorrect {
		return "solved"
	}
	return "wrong"
}

func practicePointExpression(alias string) string {
	if strings.TrimSpace(alias) == "" {
		alias = "questions"
	}
	return "CASE WHEN TRIM(COALESCE(" + alias + ".points, '')) <> '' THEN " + alias + ".points WHEN TRIM(COALESCE(" + alias + ".knowledge_point, '')) <> '' THEN " + alias + ".knowledge_point ELSE '综合' END"
}

func normalizeQuestionTypes(items []string) []string {
	if len(items) == 0 {
		return nil
	}

	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}
