package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
	"unicode"

	"your-project/model"
	"your-project/repository"

	"gorm.io/gorm"
)

type QuestionBankService interface {
	GetMeta(ctx context.Context) (*QuestionBankMetaResponse, error)
	GetFilterOptions(ctx context.Context, positionCode string) (*QuestionBankFilterOptions, error)
	ListQuestions(ctx context.Context, userID uint, filters QuestionBankListFilters) (*QuestionBankListResponse, error)
	GetQuestionLists(ctx context.Context, userID uint, positionCode string) ([]QuestionBankListDefinition, error)
	DrawQuestion(ctx context.Context, userID uint, req QuestionBankDrawRequest) (*QuestionBankQuestionEnvelope, error)
	SubmitAnswer(ctx context.Context, userID uint, req QuestionBankAnswerRequest) (*QuestionBankFeedback, error)
	GetSolution(ctx context.Context, questionID uint) (*QuestionBankSolutionResponse, error)
	GetPointSummary(ctx context.Context, userID uint, positionCode, point string) (*QuestionBankPointSummary, error)
	StartAssessment(ctx context.Context, userID uint, req QuestionBankAssessmentStartRequest) (*QuestionBankAssessmentSession, error)
	SubmitAssessmentAnswer(ctx context.Context, userID uint, req QuestionBankAssessmentAnswerRequest) (*QuestionBankFeedback, error)
	CompleteAssessment(ctx context.Context, userID uint, assessmentID uint) (*QuestionBankAssessmentSummary, error)
}

type QuestionBankListFilters struct {
	PositionCode string
	Level        string
	QuestionType string
	Specialty    string
	Point        string
	CompanyType  string
	Status       string
	Keyword      string
	ListKey      string
	Page         int
	PageSize     int
}

type QuestionBankDrawRequest struct {
	PositionCode string
	Level        string
	QuestionType string
	Specialty    string
	Point        string
	CompanyType  string
	ListKey      string
}

type QuestionBankAnswerRequest struct {
	QuestionID     uint
	UserAnswer     string
	ErrorReason    string
	ElapsedSeconds *int
	TimedMode      bool
	IsTimeout      bool
	AssessmentID   *uint
	SourceKind     string
}

type QuestionBankAssessmentStartRequest struct {
	PositionCode string
	Difficulty   string
	TotalCount   int
}

type QuestionBankAssessmentAnswerRequest struct {
	AssessmentID   uint
	QuestionID     uint
	UserAnswer     string
	ElapsedSeconds *int
	IsTimeout      bool
}

type QuestionBankMetaResponse struct {
	Positions      []QuestionBankPositionOption `json:"positions"`
	Levels         []QuestionBankNamedOption    `json:"levels"`
	QuestionTypes  []string                     `json:"question_types"`
	CompanyTypes   []string                     `json:"company_types"`
	StatusOptions  []string                     `json:"status_options"`
	QuestionCounts map[string]int64             `json:"question_counts"`
}

type QuestionBankPositionOption struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type QuestionBankNamedOption struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

type QuestionBankFilterOptions struct {
	Points        []string                  `json:"points"`
	Specialties   []string                  `json:"specialties"`
	Levels        []QuestionBankNamedOption `json:"levels"`
	QuestionTypes []string                  `json:"question_types"`
	CompanyTypes  []string                  `json:"company_types"`
	StatusOptions []string                  `json:"status_options"`
}

type QuestionBankQuestion struct {
	ID              uint                   `json:"id"`
	PositionCode    string                 `json:"position_code"`
	Position        string                 `json:"position"`
	Difficulty      string                 `json:"difficulty"`
	Level           string                 `json:"level"`
	DifficultyScore int                    `json:"difficulty_score"`
	Category        string                 `json:"category"`
	Specialty       string                 `json:"specialty"`
	KnowledgePoint  string                 `json:"knowledge_point"`
	KnowledgeArea   string                 `json:"knowledge_area"`
	Points          string                 `json:"points"`
	CompanyType     string                 `json:"company_type"`
	QuestionType    string                 `json:"question_type"`
	Title           string                 `json:"title"`
	Stem            string                 `json:"stem"`
	Options         []model.QuestionOption `json:"options,omitempty"`
	HasOptions      bool                   `json:"has_options"`
	KnowledgeTags   []string               `json:"knowledge_tags,omitempty"`
	Status          string                 `json:"status,omitempty"`
}

type QuestionBankPagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type QuestionBankStatusStats struct {
	Todo   int `json:"todo"`
	Solved int `json:"solved"`
	Wrong  int `json:"wrong"`
}

type QuestionBankListResponse struct {
	Items       []QuestionBankQuestion  `json:"items"`
	Pagination  QuestionBankPagination  `json:"pagination"`
	StatusStats QuestionBankStatusStats `json:"status_stats"`
}

type QuestionBankListDefinition struct {
	Key         string   `json:"key"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	TotalCount  int      `json:"total_count"`
	SolvedCount int      `json:"solved_count"`
	Progress    float64  `json:"progress"`
}

type QuestionBankQuestionEnvelope struct {
	Question QuestionBankQuestion `json:"question"`
}

type QuestionBankPointPacket struct {
	PositionCode        string   `json:"position_code"`
	Point               string   `json:"point"`
	Memo                string   `json:"memo"`
	InterviewExtensions []string `json:"interview_extensions"`
}

type QuestionBankPointProgress struct {
	Total      int     `json:"total"`
	Solved     int     `json:"solved"`
	Completion float64 `json:"completion"`
}

type QuestionBankFeedback struct {
	IsCorrect       bool                      `json:"is_correct"`
	AnswerMode      string                    `json:"answer_mode"`
	StandardAnswer  string                    `json:"standard_answer"`
	Analysis        string                    `json:"analysis"`
	Tips            string                    `json:"tips,omitempty"`
	Exemplar        string                    `json:"exemplar,omitempty"`
	MatchedKeywords []string                  `json:"matched_keywords,omitempty"`
	MissingKeywords []string                  `json:"missing_keywords,omitempty"`
	PointPacket     QuestionBankPointPacket   `json:"point_packet"`
	PointState      QuestionBankPointProgress `json:"point_state"`
}

type QuestionBankSolutionResponse struct {
	StandardAnswer string `json:"standard_answer"`
	Analysis       string `json:"analysis"`
	Tips           string `json:"tips,omitempty"`
	Exemplar       string `json:"exemplar,omitempty"`
}

type QuestionBankPointSummary struct {
	QuestionBankPointPacket
	Progress         QuestionBankPointProgress `json:"progress"`
	IsPointCompleted bool                      `json:"is_point_completed"`
}

type QuestionBankAssessmentSession struct {
	AssessmentID uint                   `json:"assessment_id"`
	Questions    []QuestionBankQuestion `json:"questions"`
}

type QuestionBankPointMastery struct {
	Point   string  `json:"point"`
	Mastery float64 `json:"mastery"`
}

type QuestionBankAssessmentSummary struct {
	AssessmentID      uint                       `json:"assessment_id"`
	PositionCode      string                     `json:"position_code"`
	Score             float64                    `json:"score"`
	CorrectCount      int                        `json:"correct_count"`
	TotalCount        int                        `json:"total_count"`
	TargetCompanyType string                     `json:"target_company_type"`
	PointReport       []QuestionBankPointMastery `json:"point_report"`
	NeedImprovePoints []string                   `json:"need_improve_points"`
}

type questionBankService struct {
	db *gorm.DB
}

type questionCollectionDescriptor struct {
	Key         string
	Title       string
	Description string
	Tags        []string
	Level       string
	Types       []string
}

type questionEvalResult struct {
	AnswerMode       string
	NormalizedAnswer string
	IsCorrect        bool
	MatchedKeywords  []string
	MissingKeywords  []string
}

var questionBankLevels = []QuestionBankNamedOption{
	{Key: model.QuestionLevelBase, Label: "基础层"},
	{Key: model.QuestionLevelAdvanced, Label: "提升层"},
	{Key: model.QuestionLevelSprint, Label: "冲刺层"},
}

var questionBankQuestionTypes = []string{
	model.QuestionTypeTechnicalKnowledge,
	model.QuestionTypeProjectDeepDive,
	model.QuestionTypeScenario,
	model.QuestionTypeBehavioral,
}

var questionBankCompanyTypes = []string{
	model.QuestionCompanyTypeBigTech,
	model.QuestionCompanyTypeMidInternet,
	model.QuestionCompanyTypeTraditional,
	model.QuestionCompanyTypeGeneral,
}

var questionBankPointKnowledge = map[string]string{
	"RAG":     "速记：RAG 的核心在于检索质量、重排策略、上下文组织和可追溯引用，答题时要说明召回、过滤、生成三段链路。",
	"MySQL优化": "速记：优先看慢查询与执行计划，再讲索引设计、SQL重写、连接池和冷热分层，回答里要带定位方法。",
	"动态规划":    "速记：先明确状态定义、转移方程和边界条件，再补空间优化与复杂度分析。",
	"性能优化":    "速记：从瓶颈定位、指标拆解、实验验证和收益复盘四步展开，不要只讲口号。",
}

var questionBankPointExtensions = map[string][]string{
	"RAG": {
		"如果召回率偏低但生成质量还可以，你会优先优化哪个环节，怎么验证？",
		"如何区分检索问题和生成问题，并建立线上评估指标？",
	},
	"MySQL优化": {
		"如何用慢日志、EXPLAIN 和压测闭环定位数据库瓶颈？",
		"联合索引的最左前缀原则在复杂查询里如何真正落地？",
	},
	"动态规划": {
		"你会如何把记忆化搜索平滑改写成自底向上的 DP？",
		"什么时候可以做空间压缩，什么时候会破坏状态依赖？",
	},
}

func NewQuestionBankService() QuestionBankService {
	return &questionBankService{db: repository.GetDB()}
}

func (s *questionBankService) GetMeta(ctx context.Context) (*QuestionBankMetaResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	type countRow struct {
		PositionCode string
		Count        int64
	}

	var rows []countRow
	if err := s.baseQuestionQuery().WithContext(ctx).
		Model(&model.Question{}).
		Select("position_code, COUNT(*) AS count").
		Group("position_code").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	counts := make(map[string]int64, len(rows))
	for _, row := range rows {
		counts[normalizeQuestionBankPositionCode(row.PositionCode)] = row.Count
	}

	positions := make([]QuestionBankPositionOption, 0, len(model.DefaultJobPositions))
	for _, item := range model.DefaultJobPositions {
		positions = append(positions, QuestionBankPositionOption{
			Code: item.Code,
			Name: cleanQuestionPayloadText(item.Name),
		})
	}

	return &QuestionBankMetaResponse{
		Positions:      positions,
		Levels:         questionBankLevels,
		QuestionTypes:  append([]string(nil), questionBankQuestionTypes...),
		CompanyTypes:   append([]string(nil), questionBankCompanyTypes...),
		StatusOptions:  []string{"todo", "solved", "wrong"},
		QuestionCounts: counts,
	}, nil
}

func (s *questionBankService) GetFilterOptions(ctx context.Context, positionCode string) (*QuestionBankFilterOptions, error) {
	query := s.baseQuestionQuery().WithContext(ctx)
	positionCode = normalizeQuestionBankPositionCode(positionCode)
	if positionCode != "" {
		query = query.Where("position_code = ?", positionCode)
	}

	var questions []model.Question
	if err := query.Find(&questions).Error; err != nil {
		return nil, err
	}

	pointsSet := make(map[string]struct{})
	specialtySet := make(map[string]struct{})
	for _, q := range questions {
		point := cleanQuestionPayloadText(q.Points)
		if point != "" {
			pointsSet[point] = struct{}{}
		}
		specialty := cleanQuestionPayloadText(q.Specialty)
		if specialty != "" {
			specialtySet[specialty] = struct{}{}
		}
	}

	return &QuestionBankFilterOptions{
		Points:        sortedMapKeys(pointsSet),
		Specialties:   sortedMapKeys(specialtySet),
		Levels:        questionBankLevels,
		QuestionTypes: append([]string(nil), questionBankQuestionTypes...),
		CompanyTypes:  append([]string(nil), questionBankCompanyTypes...),
		StatusOptions: []string{"todo", "solved", "wrong"},
	}, nil
}

func (s *questionBankService) ListQuestions(ctx context.Context, userID uint, filters QuestionBankListFilters) (*QuestionBankListResponse, error) {
	normalized := normalizeQuestionBankFilters(filters)

	questions, err := s.queryQuestions(ctx, normalized)
	if err != nil {
		return nil, err
	}

	statusMap, err := s.latestAttemptStatusMap(ctx, userID, collectQuestionBankQuestionIDs(questions), nil)
	if err != nil {
		return nil, err
	}

	items := make([]QuestionBankQuestion, 0, len(questions))
	stats := QuestionBankStatusStats{}
	for _, q := range questions {
		status := statusMap[q.ID]
		if status == "" {
			status = "todo"
		}

		switch status {
		case "solved":
			stats.Solved++
		case "wrong":
			stats.Wrong++
		default:
			stats.Todo++
		}

		if normalized.Status != "" && normalized.Status != status {
			continue
		}

		items = append(items, buildQuestionBankQuestion(q, status))
	}

	total := len(items)
	start := (normalized.Page - 1) * normalized.PageSize
	if start > total {
		start = total
	}
	end := start + normalized.PageSize
	if end > total {
		end = total
	}

	return &QuestionBankListResponse{
		Items: items[start:end],
		Pagination: QuestionBankPagination{
			Page:       normalized.Page,
			PageSize:   normalized.PageSize,
			Total:      total,
			TotalPages: calcTotalPages(total, normalized.PageSize),
		},
		StatusStats: stats,
	}, nil
}

func (s *questionBankService) GetQuestionLists(ctx context.Context, userID uint, positionCode string) ([]QuestionBankListDefinition, error) {
	positionCode = normalizeQuestionBankPositionCode(positionCode)
	if positionCode == "" {
		return nil, fmt.Errorf("position_code is required")
	}

	questions, err := s.queryQuestions(ctx, QuestionBankListFilters{
		PositionCode: positionCode,
		Page:         1,
		PageSize:     5000,
	})
	if err != nil {
		return nil, err
	}

	statusMap, err := s.latestAttemptStatusMap(ctx, userID, collectQuestionBankQuestionIDs(questions), nil)
	if err != nil {
		return nil, err
	}

	descriptors := collectionDescriptorsForPosition(positionDisplayName(positionCode))
	items := make([]QuestionBankListDefinition, 0, len(descriptors))
	for _, desc := range descriptors {
		total := 0
		solved := 0
		for _, q := range questions {
			if !matchesCollectionDescriptor(q, desc) {
				continue
			}
			total++
			if statusMap[q.ID] == "solved" {
				solved++
			}
		}
		progress := 0.0
		if total > 0 {
			progress = roundFloat(float64(solved*100)/float64(total), 2)
		}
		items = append(items, QuestionBankListDefinition{
			Key:         desc.Key,
			Title:       desc.Title,
			Description: desc.Description,
			Tags:        append([]string(nil), desc.Tags...),
			TotalCount:  total,
			SolvedCount: solved,
			Progress:    progress,
		})
	}

	return items, nil
}

func (s *questionBankService) DrawQuestion(ctx context.Context, userID uint, req QuestionBankDrawRequest) (*QuestionBankQuestionEnvelope, error) {
	filters := normalizeQuestionBankFilters(QuestionBankListFilters{
		PositionCode: req.PositionCode,
		Level:        req.Level,
		QuestionType: req.QuestionType,
		Specialty:    req.Specialty,
		Point:        req.Point,
		CompanyType:  req.CompanyType,
		ListKey:      req.ListKey,
		Page:         1,
		PageSize:     1,
	})

	query := s.applyQuestionFilters(s.baseQuestionQuery().WithContext(ctx), filters)
	var question model.Question
	if err := query.Order(questionBankRandomOrderExpression(s.db)).Limit(1).Take(&question).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("当前筛选条件下暂无题目")
		}
		return nil, err
	}

	statusMap, err := s.latestAttemptStatusMap(ctx, userID, []uint{question.ID}, nil)
	if err != nil {
		return nil, err
	}

	return &QuestionBankQuestionEnvelope{
		Question: buildQuestionBankQuestion(question, statusMap[question.ID]),
	}, nil
}

func (s *questionBankService) SubmitAnswer(ctx context.Context, userID uint, req QuestionBankAnswerRequest) (*QuestionBankFeedback, error) {
	if userID == 0 {
		return nil, fmt.Errorf("unauthorized")
	}
	return s.submitAnswerInternal(ctx, userID, req)
}

func (s *questionBankService) GetSolution(ctx context.Context, questionID uint) (*QuestionBankSolutionResponse, error) {
	question, err := s.findQuestionByID(ctx, questionID)
	if err != nil {
		return nil, err
	}

	return &QuestionBankSolutionResponse{
		StandardAnswer: cleanQuestionPayloadText(question.EffectiveStandardAnswer()),
		Analysis:       cleanQuestionPayloadText(question.EffectiveAnalysis()),
		Tips:           cleanQuestionPayloadText(question.Tips),
		Exemplar:       cleanQuestionPayloadText(question.Exemplar),
	}, nil
}

func (s *questionBankService) GetPointSummary(ctx context.Context, userID uint, positionCode, point string) (*QuestionBankPointSummary, error) {
	positionCode = normalizeQuestionBankPositionCode(positionCode)
	point = cleanQuestionPayloadText(point)
	if positionCode == "" || point == "" {
		return nil, fmt.Errorf("position_code and point are required")
	}

	progress, err := s.pointProgress(ctx, userID, positionCode, point)
	if err != nil {
		return nil, err
	}

	packet := buildPointPacket(positionCode, point)
	return &QuestionBankPointSummary{
		QuestionBankPointPacket: packet,
		Progress:                progress,
		IsPointCompleted:        progress.Completion >= 70,
	}, nil
}

func (s *questionBankService) StartAssessment(ctx context.Context, userID uint, req QuestionBankAssessmentStartRequest) (*QuestionBankAssessmentSession, error) {
	if userID == 0 {
		return nil, fmt.Errorf("unauthorized")
	}

	positionCode := normalizeQuestionBankPositionCode(req.PositionCode)
	if positionCode == "" {
		return nil, fmt.Errorf("position_code is required")
	}

	totalCount := req.TotalCount
	if totalCount <= 0 {
		totalCount = 12
	}
	if totalCount > 30 {
		totalCount = 30
	}

	questions, err := s.queryQuestions(ctx, QuestionBankListFilters{
		PositionCode: positionCode,
		Page:         1,
		PageSize:     5000,
	})
	if err != nil {
		return nil, err
	}
	if len(questions) == 0 {
		return nil, fmt.Errorf("该岗位暂无可用题目")
	}

	picked := pickAssessmentQuestions(questions, totalCount)
	if len(picked) == 0 {
		return nil, fmt.Errorf("未能生成测评题单")
	}

	assessment := &model.QuestionAssessment{
		UserID:       userID,
		PositionCode: positionCode,
		Difficulty:   strings.TrimSpace(req.Difficulty),
		TotalCount:   len(picked),
		Status:       "in_progress",
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(assessment).Error; err != nil {
			return err
		}
		items := make([]model.QuestionAssessmentItem, 0, len(picked))
		for idx, question := range picked {
			items = append(items, model.QuestionAssessmentItem{
				AssessmentID: assessment.ID,
				QuestionID:   question.ID,
				OrderNo:      idx + 1,
			})
		}
		return tx.Create(&items).Error
	}); err != nil {
		return nil, err
	}

	responseQuestions := make([]QuestionBankQuestion, 0, len(picked))
	for _, question := range picked {
		responseQuestions = append(responseQuestions, buildQuestionBankQuestion(question, "todo"))
	}

	return &QuestionBankAssessmentSession{
		AssessmentID: assessment.ID,
		Questions:    responseQuestions,
	}, nil
}

func (s *questionBankService) SubmitAssessmentAnswer(ctx context.Context, userID uint, req QuestionBankAssessmentAnswerRequest) (*QuestionBankFeedback, error) {
	if userID == 0 {
		return nil, fmt.Errorf("unauthorized")
	}

	var assessment model.QuestionAssessment
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", req.AssessmentID, userID).First(&assessment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("assessment not found")
		}
		return nil, err
	}
	if assessment.Status == "completed" {
		return nil, fmt.Errorf("assessment already completed")
	}

	var item model.QuestionAssessmentItem
	if err := s.db.WithContext(ctx).
		Where("assessment_id = ? AND question_id = ?", req.AssessmentID, req.QuestionID).
		First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("question is not part of this assessment")
		}
		return nil, err
	}

	return s.submitAnswerInternal(ctx, userID, QuestionBankAnswerRequest{
		QuestionID:     req.QuestionID,
		UserAnswer:     req.UserAnswer,
		ElapsedSeconds: req.ElapsedSeconds,
		IsTimeout:      req.IsTimeout,
		TimedMode:      true,
		AssessmentID:   &req.AssessmentID,
		SourceKind:     model.QuestionAttemptSourceAssessment,
	})
}

func (s *questionBankService) CompleteAssessment(ctx context.Context, userID uint, assessmentID uint) (*QuestionBankAssessmentSummary, error) {
	if userID == 0 {
		return nil, fmt.Errorf("unauthorized")
	}

	var assessment model.QuestionAssessment
	if err := s.db.WithContext(ctx).
		Preload("Items.Question").
		Where("id = ? AND user_id = ?", assessmentID, userID).
		First(&assessment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("assessment not found")
		}
		return nil, err
	}

	questionIDs := make([]uint, 0, len(assessment.Items))
	for _, item := range assessment.Items {
		questionIDs = append(questionIDs, item.QuestionID)
	}

	recordMap, err := s.latestAttemptRecords(ctx, userID, questionIDs, &assessmentID)
	if err != nil {
		return nil, err
	}

	totalCount := len(assessment.Items)
	correctCount := 0
	pointBuckets := make(map[string][]bool)
	for _, item := range assessment.Items {
		record, ok := recordMap[item.QuestionID]
		isCorrect := ok && record.IsCorrect
		if isCorrect {
			correctCount++
		}
		point := cleanQuestionPayloadText(item.Question.Points)
		if point == "" {
			point = cleanQuestionPayloadText(item.Question.KnowledgePoint)
		}
		if point == "" {
			point = "综合"
		}
		pointBuckets[point] = append(pointBuckets[point], isCorrect)
	}

	score := 0.0
	if totalCount > 0 {
		score = roundFloat(float64(correctCount*100)/float64(totalCount), 2)
	}

	pointNames := make([]string, 0, len(pointBuckets))
	for point := range pointBuckets {
		pointNames = append(pointNames, point)
	}
	sort.Strings(pointNames)

	pointReport := make([]QuestionBankPointMastery, 0, len(pointNames))
	needImprove := make([]string, 0, len(pointNames))
	for _, point := range pointNames {
		bucket := pointBuckets[point]
		hit := 0
		for _, item := range bucket {
			if item {
				hit++
			}
		}
		mastery := roundFloat(float64(hit*100)/float64(len(bucket)), 2)
		pointReport = append(pointReport, QuestionBankPointMastery{
			Point:   point,
			Mastery: mastery,
		})
		if mastery < 60 {
			needImprove = append(needImprove, point)
		}
	}

	targetCompanyType := model.QuestionCompanyTypeTraditional
	if score >= 80 {
		targetCompanyType = model.QuestionCompanyTypeBigTech
	} else if score >= 60 {
		targetCompanyType = model.QuestionCompanyTypeMidInternet
	}

	now := time.Now()
	assessment.Status = "completed"
	assessment.CorrectCount = correctCount
	assessment.Score = score
	assessment.CompletedAt = &now
	if err := s.db.WithContext(ctx).Save(&assessment).Error; err != nil {
		return nil, err
	}

	return &QuestionBankAssessmentSummary{
		AssessmentID:      assessment.ID,
		PositionCode:      assessment.PositionCode,
		Score:             score,
		CorrectCount:      correctCount,
		TotalCount:        totalCount,
		TargetCompanyType: targetCompanyType,
		PointReport:       pointReport,
		NeedImprovePoints: needImprove,
	}, nil
}

func (s *questionBankService) submitAnswerInternal(ctx context.Context, userID uint, req QuestionBankAnswerRequest) (*QuestionBankFeedback, error) {
	question, err := s.findQuestionByID(ctx, req.QuestionID)
	if err != nil {
		return nil, err
	}

	eval := evaluateQuestionAnswer(question, req.UserAnswer)
	sourceKind := strings.TrimSpace(req.SourceKind)
	if sourceKind == "" {
		if req.AssessmentID != nil && *req.AssessmentID > 0 {
			sourceKind = model.QuestionAttemptSourceAssessment
		} else {
			sourceKind = model.QuestionAttemptSourcePractice
		}
	}

	record := &model.QuestionPracticeRecord{
		UserID:           userID,
		QuestionID:       question.ID,
		AssessmentID:     req.AssessmentID,
		SourceKind:       sourceKind,
		UserAnswer:       cleanQuestionPayloadText(req.UserAnswer),
		NormalizedAnswer: eval.NormalizedAnswer,
		AnswerMode:       eval.AnswerMode,
		IsCorrect:        eval.IsCorrect,
		ErrorReason:      cleanQuestionPayloadText(req.ErrorReason),
		ElapsedSeconds:   req.ElapsedSeconds,
		TimedMode:        req.TimedMode,
		IsTimeout:        req.IsTimeout,
	}
	record.SetMatchedKeywords(eval.MatchedKeywords)
	record.SetMissingKeywords(eval.MissingKeywords)

	if err := s.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, err
	}

	point := cleanQuestionPayloadText(question.Points)
	if point == "" {
		point = cleanQuestionPayloadText(question.KnowledgePoint)
	}
	pointPacket := buildPointPacket(question.PositionCode, point)
	pointState, err := s.pointProgress(ctx, userID, question.PositionCode, point)
	if err != nil {
		return nil, err
	}

	return &QuestionBankFeedback{
		IsCorrect:       eval.IsCorrect,
		AnswerMode:      eval.AnswerMode,
		StandardAnswer:  cleanQuestionPayloadText(question.EffectiveStandardAnswer()),
		Analysis:        cleanQuestionPayloadText(question.EffectiveAnalysis()),
		Tips:            cleanQuestionPayloadText(question.Tips),
		Exemplar:        cleanQuestionPayloadText(question.Exemplar),
		MatchedKeywords: append([]string(nil), eval.MatchedKeywords...),
		MissingKeywords: append([]string(nil), eval.MissingKeywords...),
		PointPacket:     pointPacket,
		PointState:      pointState,
	}, nil
}

func (s *questionBankService) queryQuestions(ctx context.Context, filters QuestionBankListFilters) ([]model.Question, error) {
	query := s.applyQuestionFilters(s.baseQuestionQuery().WithContext(ctx), normalizeQuestionBankFilters(filters))
	var questions []model.Question
	if err := query.Order("difficulty_score DESC, id ASC").Find(&questions).Error; err != nil {
		return nil, err
	}
	return questions, nil
}

func (s *questionBankService) baseQuestionQuery() *gorm.DB {
	return s.db.Model(&model.Question{}).
		Where("is_active = ?", true).
		Where("source IS NULL OR source <> ?", "follow_up")
}

func (s *questionBankService) applyQuestionFilters(query *gorm.DB, filters QuestionBankListFilters) *gorm.DB {
	if filters.PositionCode != "" {
		query = query.Where("position_code = ?", filters.PositionCode)
	}
	if filters.Level != "" {
		query = query.Where("level = ?", filters.Level)
	}
	if filters.QuestionType != "" {
		query = query.Where("question_type = ?", filters.QuestionType)
	}
	if filters.Specialty != "" {
		query = query.Where("specialty = ?", filters.Specialty)
	}
	if filters.Point != "" {
		query = query.Where("points = ? OR knowledge_point = ?", filters.Point, filters.Point)
	}
	if filters.CompanyType != "" {
		query = query.Where("company_type = ?", filters.CompanyType)
	}
	if filters.Keyword != "" {
		keyword := "%" + filters.Keyword + "%"
		query = query.Where(
			"title LIKE ? OR stem LIKE ? OR content LIKE ? OR analysis LIKE ? OR points LIKE ? OR knowledge_point LIKE ?",
			keyword, keyword, keyword, keyword, keyword, keyword,
		)
	}
	if filters.ListKey != "" {
		if descriptor, ok := descriptorByKey(filters.ListKey); ok {
			if descriptor.Level != "" {
				query = query.Where("level = ?", descriptor.Level)
			}
			if len(descriptor.Types) > 0 {
				query = query.Where("question_type IN ?", descriptor.Types)
			}
		}
	}
	return query
}

func (s *questionBankService) latestAttemptStatusMap(ctx context.Context, userID uint, questionIDs []uint, assessmentID *uint) (map[uint]string, error) {
	records, err := s.latestAttemptRecords(ctx, userID, questionIDs, assessmentID)
	if err != nil {
		return nil, err
	}
	result := make(map[uint]string, len(records))
	for questionID, record := range records {
		if record.IsCorrect {
			result[questionID] = "solved"
		} else {
			result[questionID] = "wrong"
		}
	}
	return result, nil
}

func (s *questionBankService) latestAttemptRecords(ctx context.Context, userID uint, questionIDs []uint, assessmentID *uint) (map[uint]model.QuestionPracticeRecord, error) {
	result := make(map[uint]model.QuestionPracticeRecord)
	if userID == 0 || len(questionIDs) == 0 {
		return result, nil
	}

	sub := s.db.WithContext(ctx).
		Model(&model.QuestionPracticeRecord{}).
		Select("MAX(id)").
		Where("user_id = ?", userID).
		Where("question_id IN ?", questionIDs).
		Group("question_id")
	if assessmentID != nil && *assessmentID > 0 {
		sub = sub.Where("assessment_id = ?", *assessmentID)
	}

	var records []model.QuestionPracticeRecord
	if err := s.db.WithContext(ctx).Where("id IN (?)", sub).Find(&records).Error; err != nil {
		return nil, err
	}

	for _, record := range records {
		result[record.QuestionID] = record
	}
	return result, nil
}

func (s *questionBankService) findQuestionByID(ctx context.Context, questionID uint) (*model.Question, error) {
	var question model.Question
	if err := s.baseQuestionQuery().WithContext(ctx).Where("id = ?", questionID).First(&question).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("question not found")
		}
		return nil, err
	}
	return &question, nil
}

func (s *questionBankService) pointProgress(ctx context.Context, userID uint, positionCode, point string) (QuestionBankPointProgress, error) {
	positionCode = normalizeQuestionBankPositionCode(positionCode)
	point = cleanQuestionPayloadText(point)
	if positionCode == "" || point == "" {
		return QuestionBankPointProgress{}, nil
	}

	var total int64
	if err := s.baseQuestionQuery().WithContext(ctx).
		Where("position_code = ?", positionCode).
		Where("points = ? OR knowledge_point = ?", point, point).
		Count(&total).Error; err != nil {
		return QuestionBankPointProgress{}, err
	}
	if total == 0 {
		return QuestionBankPointProgress{}, nil
	}

	var solvedIDs []uint
	if err := s.db.WithContext(ctx).
		Model(&model.QuestionPracticeRecord{}).
		Distinct("question_practice_records.question_id").
		Joins("JOIN questions q ON q.id = question_practice_records.question_id").
		Where("question_practice_records.user_id = ? AND question_practice_records.is_correct = ?", userID, true).
		Where("q.position_code = ?", positionCode).
		Where("q.points = ? OR q.knowledge_point = ?", point, point).
		Pluck("question_practice_records.question_id", &solvedIDs).Error; err != nil {
		return QuestionBankPointProgress{}, err
	}

	completion := roundFloat(float64(len(solvedIDs)*100)/float64(total), 2)
	return QuestionBankPointProgress{
		Total:      int(total),
		Solved:     len(solvedIDs),
		Completion: completion,
	}, nil
}

func buildQuestionBankQuestion(question model.Question, status string) QuestionBankQuestion {
	options := question.GetOptions()
	return QuestionBankQuestion{
		ID:              question.ID,
		PositionCode:    normalizeQuestionBankPositionCode(question.PositionCode),
		Position:        cleanQuestionPayloadText(question.Position),
		Difficulty:      cleanQuestionPayloadText(question.Difficulty),
		Level:           cleanQuestionPayloadText(question.Level),
		DifficultyScore: question.DifficultyScore,
		Category:        cleanQuestionPayloadText(question.Category),
		Specialty:       cleanQuestionPayloadText(question.Specialty),
		KnowledgePoint:  cleanQuestionPayloadText(question.KnowledgePoint),
		KnowledgeArea:   cleanQuestionPayloadText(question.KnowledgeArea),
		Points:          cleanQuestionPayloadText(question.Points),
		CompanyType:     cleanQuestionPayloadText(question.CompanyType),
		QuestionType:    cleanQuestionPayloadText(question.QuestionType),
		Title:           cleanQuestionPayloadText(question.Title),
		Stem:            cleanQuestionPayloadText(question.EffectiveStem()),
		Options:         sanitizeQuestionOptions(options),
		HasOptions:      len(options) > 0,
		KnowledgeTags:   sanitizeStringSlice(question.GetKnowledgeTags()),
		Status:          cleanQuestionPayloadText(status),
	}
}

func evaluateQuestionAnswer(question *model.Question, userAnswer string) questionEvalResult {
	answer := cleanQuestionPayloadText(userAnswer)
	options := question.GetOptions()
	if len(options) > 0 {
		expected := normalizeOptionAnswer(question.EffectiveStandardAnswer())
		actual := normalizeOptionAnswer(answer)
		result := questionEvalResult{
			AnswerMode:       "choice",
			NormalizedAnswer: actual,
			IsCorrect:        actual != "" && actual == expected,
		}
		if result.IsCorrect {
			result.MatchedKeywords = []string{expected}
		} else if expected != "" {
			result.MissingKeywords = []string{expected}
		}
		return result
	}

	keywords := extractExpectedKeywords(question)
	normalized := normalizeNaturalAnswer(answer)
	matched := make([]string, 0, len(keywords))
	missing := make([]string, 0, len(keywords))
	for _, keyword := range keywords {
		if keyword == "" {
			continue
		}
		if strings.Contains(normalized, normalizeNaturalAnswer(keyword)) {
			matched = append(matched, keyword)
		} else {
			missing = append(missing, keyword)
		}
	}

	threshold := 1
	switch question.QuestionType {
	case model.QuestionTypeTechnicalKnowledge:
		if len(keywords) > 1 {
			threshold = len(keywords) - 1
		}
	default:
		if len(keywords) > 2 {
			threshold = len(keywords) / 2
		}
	}
	if threshold < 1 {
		threshold = 1
	}

	isCorrect := len(answer) >= 8 && len(matched) >= threshold
	return questionEvalResult{
		AnswerMode:       "keyword",
		NormalizedAnswer: normalized,
		IsCorrect:        isCorrect,
		MatchedKeywords:  matched,
		MissingKeywords:  missing,
	}
}

func extractExpectedKeywords(question *model.Question) []string {
	if question == nil {
		return []string{}
	}

	candidates := make([]string, 0, 12)
	answer := cleanQuestionPayloadText(question.EffectiveStandardAnswer())
	if strings.Contains(answer, "|") {
		candidates = append(candidates, strings.Split(answer, "|")...)
	} else {
		answer = strings.NewReplacer("；", "\n", ";", "\n", "，", "\n", ",", "\n", "、", "\n").Replace(answer)
		candidates = append(candidates, strings.Split(answer, "\n")...)
	}

	candidates = append(candidates, question.KnowledgePoint, question.Points, question.Specialty, question.Category)
	normalized := sanitizeStringSlice(candidates)
	if len(normalized) > 6 {
		return normalized[:6]
	}
	return normalized
}

func buildPointPacket(positionCode, point string) QuestionBankPointPacket {
	point = cleanQuestionPayloadText(point)
	memo := questionBankPointKnowledge[point]
	if memo == "" {
		memo = fmt.Sprintf("速记：围绕 %s 回答时，建议按“问题背景-核心原理-工程落地-风险边界-复盘收益”五段来组织。", point)
	}

	extensions := questionBankPointExtensions[point]
	if len(extensions) == 0 {
		extensions = []string{
			fmt.Sprintf("如果让你围绕 %s 做一轮技术方案选型，你会怎么权衡？", point),
			fmt.Sprintf("%s 在真实项目里最常见的失败点是什么，如何预防？", point),
		}
	}

	return QuestionBankPointPacket{
		PositionCode:        normalizeQuestionBankPositionCode(positionCode),
		Point:               point,
		Memo:                cleanQuestionPayloadText(memo),
		InterviewExtensions: sanitizeStringSlice(extensions),
	}
}

func normalizeQuestionBankFilters(filters QuestionBankListFilters) QuestionBankListFilters {
	filters.PositionCode = normalizeQuestionBankPositionCode(filters.PositionCode)
	filters.Level = strings.TrimSpace(strings.ToLower(filters.Level))
	filters.QuestionType = cleanQuestionPayloadText(filters.QuestionType)
	filters.Specialty = cleanQuestionPayloadText(filters.Specialty)
	filters.Point = cleanQuestionPayloadText(filters.Point)
	filters.CompanyType = cleanQuestionPayloadText(filters.CompanyType)
	filters.Status = strings.TrimSpace(strings.ToLower(filters.Status))
	filters.Keyword = cleanQuestionPayloadText(filters.Keyword)
	filters.ListKey = strings.TrimSpace(strings.ToLower(filters.ListKey))
	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.PageSize <= 0 {
		filters.PageSize = 15
	}
	if filters.PageSize > 50 {
		filters.PageSize = 50
	}
	if filters.Status != "solved" && filters.Status != "wrong" && filters.Status != "todo" {
		filters.Status = ""
	}
	return filters
}

func normalizeQuestionBankPositionCode(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch {
	case value == "", value == string(model.PositionBackend), value == "java", strings.Contains(value, "backend"), strings.Contains(value, "后端"):
		if value == "" {
			return ""
		}
		return string(model.PositionBackend)
	case value == string(model.PositionFrontend), strings.Contains(value, "frontend"), strings.Contains(value, "前端"):
		return string(model.PositionFrontend)
	case value == string(model.PositionAlgorithm), strings.Contains(value, "algorithm"), strings.Contains(value, "算法"):
		return string(model.PositionAlgorithm)
	case value == string(model.PositionAI), strings.Contains(value, "ai"), strings.Contains(value, "llm"), strings.Contains(value, "模型"):
		return string(model.PositionAI)
	default:
		return value
	}
}

func collectionDescriptorsForPosition(positionName string) []questionCollectionDescriptor {
	titlePrefix := cleanQuestionPayloadText(positionName)
	if titlePrefix == "" {
		titlePrefix = "题库"
	}

	return []questionCollectionDescriptor{
		{
			Key:         "base-100",
			Title:       fmt.Sprintf("%s 入门题单", titlePrefix),
			Description: "适合打基础，覆盖高频核心概念，先把题感和回答结构建立起来。",
			Tags:        []string{"入门", "高频", "基础层"},
			Level:       model.QuestionLevelBase,
		},
		{
			Key:         "advanced-core",
			Title:       fmt.Sprintf("%s 核心进阶题单", titlePrefix),
			Description: "围绕岗位技术栈做专项强化，适合刷深度和工程表达。",
			Tags:        []string{"进阶", "专项", "提升层"},
			Level:       model.QuestionLevelAdvanced,
		},
		{
			Key:         "sprint-hot",
			Title:       fmt.Sprintf("%s 冲刺高频题单", titlePrefix),
			Description: "模拟高压面试节奏，适合冲刺阶段做综合复盘。",
			Tags:        []string{"冲刺", "高频", "大厂"},
			Level:       model.QuestionLevelSprint,
		},
		{
			Key:         "project-behavior",
			Title:       fmt.Sprintf("%s 项目与行为表达", titlePrefix),
			Description: "重点训练项目深挖和行为表达，不再只会背知识点。",
			Tags:        []string{"项目", "行为", "表达"},
			Types:       []string{model.QuestionTypeProjectDeepDive, model.QuestionTypeBehavioral},
		},
		{
			Key:         "mock-mixed",
			Title:       fmt.Sprintf("%s 综合模拟", titlePrefix),
			Description: "混合题型完整走一轮，更接近真实面试链路。",
			Tags:        []string{"模拟", "混合", "综合"},
		},
	}
}

func descriptorByKey(key string) (questionCollectionDescriptor, bool) {
	for _, desc := range collectionDescriptorsForPosition("题库") {
		if desc.Key == key {
			return desc, true
		}
	}
	return questionCollectionDescriptor{}, false
}

func matchesCollectionDescriptor(question model.Question, desc questionCollectionDescriptor) bool {
	if desc.Level != "" && !strings.EqualFold(strings.TrimSpace(question.Level), desc.Level) {
		return false
	}
	if len(desc.Types) > 0 {
		for _, item := range desc.Types {
			if strings.TrimSpace(question.QuestionType) == item {
				return true
			}
		}
		return false
	}
	return true
}

func pickAssessmentQuestions(questions []model.Question, totalCount int) []model.Question {
	if len(questions) == 0 {
		return nil
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	working := append([]model.Question(nil), questions...)
	rng.Shuffle(len(working), func(i, j int) {
		working[i], working[j] = working[j], working[i]
	})

	byPoint := make(map[string][]model.Question)
	pointOrder := make([]string, 0)
	for _, question := range working {
		point := cleanQuestionPayloadText(question.Points)
		if point == "" {
			point = cleanQuestionPayloadText(question.KnowledgePoint)
		}
		if point == "" {
			point = "综合"
		}
		if _, ok := byPoint[point]; !ok {
			pointOrder = append(pointOrder, point)
		}
		byPoint[point] = append(byPoint[point], question)
	}

	picked := make([]model.Question, 0, totalCount)
	used := make(map[uint]struct{})
	for _, point := range pointOrder {
		bucket := byPoint[point]
		limit := 2
		if len(bucket) < limit {
			limit = len(bucket)
		}
		for i := 0; i < limit && len(picked) < totalCount; i++ {
			if _, ok := used[bucket[i].ID]; ok {
				continue
			}
			used[bucket[i].ID] = struct{}{}
			picked = append(picked, bucket[i])
		}
		if len(picked) >= totalCount {
			return picked
		}
	}

	for _, question := range working {
		if len(picked) >= totalCount {
			break
		}
		if _, ok := used[question.ID]; ok {
			continue
		}
		used[question.ID] = struct{}{}
		picked = append(picked, question)
	}

	return picked
}

func collectQuestionBankQuestionIDs(items []model.Question) []uint {
	out := make([]uint, 0, len(items))
	for _, item := range items {
		if item.ID == 0 {
			continue
		}
		out = append(out, item.ID)
	}
	return out
}

func positionDisplayName(code string) string {
	code = normalizeQuestionBankPositionCode(code)
	for _, item := range model.DefaultJobPositions {
		if item.Code == code {
			return item.Name
		}
	}
	return code
}

func sanitizeQuestionOptions(items []model.QuestionOption) []model.QuestionOption {
	if len(items) == 0 {
		return nil
	}
	out := make([]model.QuestionOption, 0, len(items))
	for _, item := range items {
		key := strings.TrimSpace(strings.ToUpper(item.Key))
		text := cleanQuestionPayloadText(item.Text)
		if key == "" || text == "" {
			continue
		}
		out = append(out, model.QuestionOption{Key: key, Text: text})
	}
	return out
}

func sanitizeStringSlice(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		cleaned := cleanQuestionPayloadText(item)
		if cleaned == "" {
			continue
		}
		if _, ok := seen[cleaned]; ok {
			continue
		}
		seen[cleaned] = struct{}{}
		out = append(out, cleaned)
	}
	return out
}

func cleanQuestionPayloadText(value string) string {
	value = strings.ToValidUTF8(value, "")
	value = strings.ReplaceAll(value, "\uFEFF", "")
	value = strings.ReplaceAll(value, "\ufffd", "")
	value = strings.ReplaceAll(value, "```json", "")
	value = strings.ReplaceAll(value, "```JSON", "")
	value = strings.ReplaceAll(value, "```markdown", "")
	value = strings.ReplaceAll(value, "```md", "")
	value = strings.ReplaceAll(value, "```text", "")
	value = strings.ReplaceAll(value, "```", "")

	var builder strings.Builder
	builder.Grow(len(value))
	lastSpace := false
	lastLineBreak := false
	for _, r := range value {
		switch {
		case r == '\r':
			continue
		case r == '\n':
			if !lastLineBreak {
				builder.WriteRune('\n')
			}
			lastLineBreak = true
			lastSpace = false
		case r == '\t' || unicode.IsSpace(r):
			if !lastSpace && !lastLineBreak {
				builder.WriteRune(' ')
				lastSpace = true
			}
		case unicode.IsControl(r):
			continue
		default:
			builder.WriteRune(r)
			lastSpace = false
			lastLineBreak = false
		}
	}
	return strings.TrimSpace(builder.String())
}

func normalizeNaturalAnswer(value string) string {
	value = cleanQuestionPayloadText(strings.ToLower(value))
	replacer := strings.NewReplacer(" ", "", "\n", "", "\t", "", "，", "", ",", "", "。", "", ".", "", "；", "", ";", "", "：", "", ":", "")
	return replacer.Replace(value)
}

func normalizeOptionAnswer(value string) string {
	value = strings.ToUpper(cleanQuestionPayloadText(value))
	value = strings.NewReplacer(" ", "", "\n", "", "\t", "", "，", ",", "、", ",", ";", ",", "|", ",").Replace(value)
	parts := strings.Split(value, ",")
	normalized := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		normalized = append(normalized, part)
	}
	sort.Strings(normalized)
	return strings.Join(normalized, ",")
}

func sortedMapKeys(items map[string]struct{}) []string {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func calcTotalPages(total, pageSize int) int {
	if total == 0 || pageSize <= 0 {
		return 0
	}
	return (total + pageSize - 1) / pageSize
}

func roundFloat(value float64, precision int) float64 {
	if precision < 0 {
		return value
	}
	pow := 1.0
	for i := 0; i < precision; i++ {
		pow *= 10
	}
	return float64(int(value*pow+0.5)) / pow
}

func questionBankRandomOrderExpression(db *gorm.DB) string {
	if db != nil && db.Dialector != nil && strings.EqualFold(db.Dialector.Name(), "postgres") {
		return "RANDOM()"
	}
	return "RAND()"
}
