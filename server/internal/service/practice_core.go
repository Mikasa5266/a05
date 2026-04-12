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

	"your-project/internal/model"
	"your-project/internal/repository"

	"gorm.io/gorm"
)

type practiceNamedOption struct {
	Key   string
	Label string
}

type practiceQuestionEvalResult struct {
	AnswerMode       string
	NormalizedAnswer string
	IsCorrect        bool
	MatchedKeywords  []string
	MissingKeywords  []string
}

var questionBankLevels = []practiceNamedOption{
	{Key: model.QuestionLevelBase, Label: "基础"},
	{Key: model.QuestionLevelAdvanced, Label: "提升"},
	{Key: model.QuestionLevelSprint, Label: "冲刺"},
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
	"MySQL优化": "速记：优先看慢查询与执行计划，再讲索引设计、SQL 重写、连接池和冷热分层，回答里要带定位方法。",
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

func (s *practiceService) submitAnswer(ctx context.Context, userID uint, req PracticeAnswerRequest) (*PracticeAnswerFeedback, error) {
	if userID == 0 {
		return nil, unauthorizedPracticeError("unauthorized")
	}
	return s.submitAnswerInternal(ctx, userID, req)
}

func (s *practiceService) getSolution(ctx context.Context, questionID uint) (*PracticeSolutionResponse, error) {
	question, err := s.findQuestionByID(ctx, questionID)
	if err != nil {
		return nil, err
	}

	return &PracticeSolutionResponse{
		StandardAnswer: cleanQuestionPayloadText(question.EffectiveStandardAnswer()),
		Analysis:       cleanQuestionPayloadText(question.EffectiveAnalysis()),
		Tips:           cleanQuestionPayloadText(question.Tips),
		Exemplar:       cleanQuestionPayloadText(question.Exemplar),
	}, nil
}

func (s *practiceService) getPointSummary(ctx context.Context, userID uint, positionCode, point string) (*PracticePointSummary, error) {
	positionCode = normalizeQuestionBankPositionCode(positionCode)
	point = cleanQuestionPayloadText(point)
	if positionCode == "" || point == "" {
		return nil, invalidPracticeArgument("position_code and point are required")
	}

	progress, err := s.pointProgress(ctx, userID, positionCode, point)
	if err != nil {
		return nil, err
	}

	packet := buildPointPacket(positionCode, point)
	return &PracticePointSummary{
		PracticePointPacket: packet,
		Progress:            progress,
		IsPointCompleted:    progress.Completion >= 70,
	}, nil
}

func (s *practiceService) startAssessment(ctx context.Context, userID uint, req PracticeAssessmentStartRequest) (*PracticeAssessmentSession, error) {
	if userID == 0 {
		return nil, unauthorizedPracticeError("unauthorized")
	}

	positionCode := normalizeQuestionBankPositionCode(req.PositionCode)
	if positionCode == "" {
		return nil, invalidPracticeArgument("position_code is required")
	}

	totalCount := req.TotalCount
	if totalCount <= 0 {
		totalCount = 12
	}
	if totalCount > 30 {
		totalCount = 30
	}

	questions, err := s.repo.ListQuestions(ctx, repository.PracticeQuestionFilters{
		PositionCode: positionCode,
	})
	if err != nil {
		return nil, practiceInternalError("failed to load questions", err)
	}
	if len(questions) == 0 {
		return nil, practiceNotFound("该岗位暂无可用题目", gorm.ErrRecordNotFound)
	}

	picked := pickAssessmentQuestions(questions, totalCount)
	if len(picked) == 0 {
		return nil, practiceInternalError("未能生成测评题单", nil)
	}

	assessment := &model.QuestionAssessment{
		UserID:       userID,
		PositionCode: positionCode,
		Difficulty:   strings.TrimSpace(req.Difficulty),
		TotalCount:   len(picked),
		Status:       "in_progress",
	}

	items := make([]model.QuestionAssessmentItem, 0, len(picked))
	for idx, question := range picked {
		items = append(items, model.QuestionAssessmentItem{
			QuestionID: question.ID,
			OrderNo:    idx + 1,
		})
	}

	if err := s.repo.CreateAssessment(ctx, assessment, items); err != nil {
		return nil, practiceInternalError("failed to create assessment", err)
	}

	responseQuestions := make([]PracticeQuestionDetail, 0, len(picked))
	for _, question := range picked {
		responseQuestions = append(responseQuestions, buildPracticeAssessmentQuestion(question, "todo"))
	}

	return &PracticeAssessmentSession{
		AssessmentID: assessment.ID,
		Questions:    responseQuestions,
	}, nil
}

func (s *practiceService) submitAssessmentAnswer(ctx context.Context, userID uint, req PracticeAssessmentAnswerRequest) (*PracticeAnswerFeedback, error) {
	if userID == 0 {
		return nil, unauthorizedPracticeError("unauthorized")
	}

	assessment, err := s.repo.GetAssessment(ctx, req.AssessmentID, userID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, practiceNotFound("assessment not found", err)
		}
		return nil, practiceInternalError("failed to load assessment", err)
	}
	if assessment.Status == "completed" {
		return nil, invalidPracticeArgument("assessment already completed")
	}

	if _, err := s.repo.GetAssessmentItem(ctx, req.AssessmentID, req.QuestionID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, invalidPracticeArgument("question is not part of this assessment")
		}
		return nil, practiceInternalError("failed to verify assessment item", err)
	}

	return s.submitAnswerInternal(ctx, userID, PracticeAnswerRequest{
		QuestionID:     req.QuestionID,
		UserAnswer:     req.UserAnswer,
		ElapsedSeconds: req.ElapsedSeconds,
		IsTimeout:      req.IsTimeout,
		TimedMode:      true,
		AssessmentID:   &req.AssessmentID,
		SourceKind:     model.QuestionAttemptSourceAssessment,
	})
}

func (s *practiceService) completeAssessment(ctx context.Context, userID uint, assessmentID uint) (*PracticeAssessmentSummary, error) {
	if userID == 0 {
		return nil, unauthorizedPracticeError("unauthorized")
	}

	assessment, err := s.repo.GetAssessment(ctx, assessmentID, userID, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, practiceNotFound("assessment not found", err)
		}
		return nil, practiceInternalError("failed to load assessment", err)
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

	pointReport := make([]PracticePointMastery, 0, len(pointNames))
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
		pointReport = append(pointReport, PracticePointMastery{
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
	if err := s.repo.UpdateAssessment(ctx, assessment); err != nil {
		return nil, practiceInternalError("failed to update assessment", err)
	}

	return &PracticeAssessmentSummary{
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

func (s *practiceService) submitAnswerInternal(ctx context.Context, userID uint, req PracticeAnswerRequest) (*PracticeAnswerFeedback, error) {
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

	if err := s.repo.CreatePracticeRecord(ctx, record); err != nil {
		return nil, practiceInternalError("failed to create practice record", err)
	}

	if req.AssessmentID != nil && *req.AssessmentID > 0 {
		assessmentAnswer := &model.PracticeAssessmentAnswer{
			AssessmentID:   *req.AssessmentID,
			QuestionID:     question.ID,
			UserID:         userID,
			UserAnswer:     cleanQuestionPayloadText(req.UserAnswer),
			IsCorrect:      eval.IsCorrect,
			ElapsedSeconds: req.ElapsedSeconds,
			IsTimeout:      req.IsTimeout,
		}
		if err := s.repo.UpsertAssessmentAnswer(ctx, assessmentAnswer); err != nil {
			return nil, practiceInternalError("failed to persist assessment answer", err)
		}
	}

	if eval.IsCorrect {
		if err := s.repo.DeleteWrongBookEntry(ctx, userID, question.ID); err != nil {
			return nil, practiceInternalError("failed to remove wrong book entry", err)
		}
	} else {
		wrongBookEntry := &model.PracticeWrongBookEntry{
			UserID:         userID,
			QuestionID:     question.ID,
			LastUserAnswer: cleanQuestionPayloadText(req.UserAnswer),
			ErrorReason:    cleanQuestionPayloadText(req.ErrorReason),
		}
		if err := s.repo.UpsertWrongBookEntry(ctx, wrongBookEntry); err != nil {
			return nil, practiceInternalError("failed to update wrong book entry", err)
		}
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

	return &PracticeAnswerFeedback{
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

func (s *practiceService) latestAttemptRecords(ctx context.Context, userID uint, questionIDs []uint, assessmentID *uint) (map[uint]model.QuestionPracticeRecord, error) {
	records, err := s.repo.LatestAttemptRecords(ctx, userID, questionIDs, assessmentID)
	if err != nil {
		return nil, practiceInternalError("failed to load latest attempts", err)
	}
	return records, nil
}

func (s *practiceService) findQuestionByID(ctx context.Context, questionID uint) (*model.Question, error) {
	question, err := s.repo.GetQuestionByID(ctx, questionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, practiceNotFound("question not found", err)
		}
		return nil, practiceInternalError("failed to load question", err)
	}
	return question, nil
}

func (s *practiceService) pointProgress(ctx context.Context, userID uint, positionCode, point string) (PracticePointProgress, error) {
	positionCode = normalizeQuestionBankPositionCode(positionCode)
	point = cleanQuestionPayloadText(point)
	if positionCode == "" || point == "" {
		return PracticePointProgress{}, nil
	}

	total, err := s.repo.CountQuestionsByPoint(ctx, positionCode, point)
	if err != nil {
		return PracticePointProgress{}, practiceInternalError("failed to count point questions", err)
	}
	if total == 0 {
		return PracticePointProgress{}, nil
	}

	solvedIDs, err := s.repo.ListSolvedQuestionIDsByPoint(ctx, userID, positionCode, point)
	if err != nil {
		return PracticePointProgress{}, practiceInternalError("failed to load solved question ids", err)
	}

	completion := roundFloat(float64(len(solvedIDs)*100)/float64(total), 2)
	return PracticePointProgress{
		Total:      int(total),
		Solved:     len(solvedIDs),
		Completion: completion,
	}, nil
}

func buildPracticeAssessmentQuestion(question model.Question, status string) PracticeQuestionDetail {
	options := sanitizeQuestionOptions(question.GetOptions())
	role := normalizeQuestionBankPositionCode(question.PositionCode)
	return PracticeQuestionDetail{
		ID:              question.ID,
		Role:            role,
		PositionCode:    role,
		Position:        cleanQuestionPayloadText(question.Position),
		Difficulty:      cleanQuestionPayloadText(question.Difficulty),
		Level:           cleanQuestionPayloadText(question.Level),
		QuestionType:    cleanQuestionPayloadText(question.QuestionType),
		Specialty:       cleanQuestionPayloadText(question.Specialty),
		Stem:            cleanQuestionPayloadText(question.EffectiveStem()),
		Options:         options,
		Points:          cleanQuestionPayloadText(question.Points),
		CompanyType:     cleanQuestionPayloadText(question.CompanyType),
		DifficultyScore: question.DifficultyScore,
		HasOptions:      len(options) > 0,
		Status:          cleanQuestionPayloadText(status),
		IsFavorite:      false,
	}
}

func evaluateQuestionAnswer(question *model.Question, userAnswer string) practiceQuestionEvalResult {
	answer := cleanQuestionPayloadText(userAnswer)
	options := question.GetOptions()
	if len(options) > 0 {
		expected := normalizeOptionAnswer(question.EffectiveStandardAnswer())
		actual := normalizeOptionAnswer(answer)
		result := practiceQuestionEvalResult{
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
	return practiceQuestionEvalResult{
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
		answer = strings.NewReplacer("。", "\n", ";", "\n", "；", "\n", ",", "\n", "，", "\n").Replace(answer)
		candidates = append(candidates, strings.Split(answer, "\n")...)
	}

	candidates = append(candidates, question.KnowledgePoint, question.Points, question.Specialty, question.Category)
	normalized := sanitizeStringSlice(candidates)
	if len(normalized) > 6 {
		return normalized[:6]
	}
	return normalized
}

func buildPointPacket(positionCode, point string) PracticePointPacket {
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

	return PracticePointPacket{
		PositionCode:        normalizeQuestionBankPositionCode(positionCode),
		Point:               point,
		Memo:                cleanQuestionPayloadText(memo),
		InterviewExtensions: sanitizeStringSlice(extensions),
	}
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
	replacer := strings.NewReplacer(" ", "", "\n", "", "\t", "", "。", "", ",", "", "，", "", ".", "", "！", "", ";", "", "；", "", ":", "", "：", "")
	return replacer.Replace(value)
}

func normalizeOptionAnswer(value string) string {
	value = strings.ToUpper(cleanQuestionPayloadText(value))
	value = strings.NewReplacer(" ", "", "\n", "", "\t", "", "，", ",", "；", ",", ";", ",", "|", ",").Replace(value)
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
