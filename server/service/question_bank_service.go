package service

import (
	"context"
	"fmt"
	"strings"

	"your-project/model"
	"your-project/repository"
)

type QuestionBankService interface {
	ListPositions(ctx context.Context) ([]model.JobPosition, error)
	GeneratePositionQuestionList(ctx context.Context, userID uint, positionCode, difficulty string, limit int) ([]*model.Question, error)
	GenerateResumeQuestionList(ctx context.Context, userID, resumeResultID uint, difficulty string, limit int) ([]*model.Question, error)
	GetQuestion(ctx context.Context, questionID uint) (*model.Question, error)
	SetFavorite(ctx context.Context, userID, questionID uint, isFavorite bool) error
	MarkWrongQuestion(ctx context.Context, userID, questionID uint, note string) error
	ClearWrongQuestion(ctx context.Context, userID, questionID uint) error
	ListFavorites(ctx context.Context, userID uint, page, pageSize int) ([]*model.Question, int64, error)
	ListWrongQuestions(ctx context.Context, userID uint, page, pageSize int) ([]*model.Question, int64, error)
}

type DefaultQuestionBankService struct {
	questionRepo repository.QuestionRepository
	positionRepo repository.PositionRepository
	resumeRepo   repository.ResumeParseResultRepository
	stateRepo    repository.UserQuestionStateRepository
}

var _ QuestionBankService = (*DefaultQuestionBankService)(nil)

func NewQuestionBankService() QuestionBankService {
	return NewQuestionBankServiceWithDeps(
		repository.NewQuestionRepository(),
		repository.NewPositionRepository(),
		repository.NewResumeParseResultRepository(),
		repository.NewUserQuestionStateRepository(),
	)
}

func NewQuestionBankServiceWithDeps(
	questionRepo repository.QuestionRepository,
	positionRepo repository.PositionRepository,
	resumeRepo repository.ResumeParseResultRepository,
	stateRepo repository.UserQuestionStateRepository,
) QuestionBankService {
	if questionRepo == nil {
		questionRepo = repository.NewQuestionRepository()
	}
	if positionRepo == nil {
		positionRepo = repository.NewPositionRepository()
	}
	if resumeRepo == nil {
		resumeRepo = repository.NewResumeParseResultRepository()
	}
	if stateRepo == nil {
		stateRepo = repository.NewUserQuestionStateRepository()
	}

	return &DefaultQuestionBankService{
		questionRepo: questionRepo,
		positionRepo: positionRepo,
		resumeRepo:   resumeRepo,
		stateRepo:    stateRepo,
	}
}

func (s *DefaultQuestionBankService) ListPositions(ctx context.Context) ([]model.JobPosition, error) {
	_ = ctx
	positions, err := s.positionRepo.ListActive()
	if err != nil || len(positions) == 0 {
		fallback := make([]model.JobPosition, 0, len(model.DefaultJobPositions))
		fallback = append(fallback, model.DefaultJobPositions...)
		return fallback, nil
	}
	return positions, nil
}

func (s *DefaultQuestionBankService) GeneratePositionQuestionList(ctx context.Context, userID uint, positionCode, difficulty string, limit int) ([]*model.Question, error) {
	_ = ctx
	_ = userID

	if limit <= 0 {
		limit = 12
	}

	positionName := s.resolvePositionName(positionCode)
	if strings.TrimSpace(positionName) == "" {
		positionName = "Java后端工程师"
	}

	collected := make([]*model.Question, 0, limit)
	seen := map[uint]struct{}{}

	appendQuestions := func(list []*model.Question) {
		for _, q := range list {
			if q == nil || q.ID == 0 {
				continue
			}
			if _, ok := seen[q.ID]; ok {
				continue
			}
			seen[q.ID] = struct{}{}
			collected = append(collected, q)
			if len(collected) >= limit {
				return
			}
		}
	}

	seed, err := s.questionRepo.GetQuestionsForInterviewInit(positionName, difficulty, "", limit*3)
	if err == nil {
		appendQuestions(seed)
	}

	if len(collected) < limit {
		randomList, randomErr := s.questionRepo.ListRandomByPosition(positionName, limit*2)
		if randomErr == nil {
			appendQuestions(randomList)
		}
	}

	if len(collected) == 0 {
		return nil, fmt.Errorf("no questions found for position %s", positionName)
	}

	if len(collected) > limit {
		collected = collected[:limit]
	}
	return collected, nil
}

func (s *DefaultQuestionBankService) GenerateResumeQuestionList(ctx context.Context, userID, resumeResultID uint, difficulty string, limit int) ([]*model.Question, error) {
	_ = ctx
	if limit <= 0 {
		limit = 12
	}

	record, err := s.resumeRepo.GetByID(resumeResultID)
	if err != nil {
		return nil, err
	}
	if record.UserID != userID {
		return nil, fmt.Errorf("unauthorized access to resume result")
	}

	analysis := &ResumeAnalysisResult{}
	if err := parseStrictJSON(record.StructuredJSON, analysis); err != nil {
		analysis = &ResumeAnalysisResult{}
	}

	targetCode := strings.TrimSpace(record.MatchedPositionCode)
	if targetCode == "" && len(analysis.SuggestedPositions) > 0 {
		targetCode = strings.TrimSpace(analysis.SuggestedPositions[0].PositionCode)
	}
	positionName := s.resolvePositionName(targetCode)
	if positionName == "" {
		positionName = strings.TrimSpace(record.MatchedPositionName)
	}
	if positionName == "" {
		positionName = "Java后端工程师"
	}

	collected := make([]*model.Question, 0, limit)
	seen := map[uint]struct{}{}
	appendQuestions := func(list []*model.Question) {
		for _, q := range list {
			if q == nil || q.ID == 0 {
				continue
			}
			if _, ok := seen[q.ID]; ok {
				continue
			}
			seen[q.ID] = struct{}{}
			collected = append(collected, q)
			if len(collected) >= limit {
				return
			}
		}
	}

	wrongList, _, wrongErr := s.stateRepo.ListWrongQuestions(userID, 1, limit)
	if wrongErr == nil {
		appendQuestions(wrongList)
	}

	knowledgeSeeds := make([]string, 0, len(analysis.MissingSkills)+len(analysis.Skills))
	knowledgeSeeds = append(knowledgeSeeds, analysis.MissingSkills...)
	for _, skill := range analysis.Skills {
		if len(knowledgeSeeds) >= 8 {
			break
		}
		knowledgeSeeds = append(knowledgeSeeds, skill.Name)
	}

	for _, seed := range uniqueNonEmpty(knowledgeSeeds) {
		if len(collected) >= limit {
			break
		}
		list, findErr := s.questionRepo.FindByKnowledgePoint(positionName, difficulty, seed, 4)
		if findErr != nil {
			continue
		}
		appendQuestions(list)
	}

	if len(collected) < limit {
		positionQuestions, listErr := s.GeneratePositionQuestionList(ctx, userID, targetCode, difficulty, limit)
		if listErr == nil {
			appendQuestions(positionQuestions)
		}
	}

	if len(collected) == 0 {
		return nil, fmt.Errorf("no personalized questions generated")
	}
	if len(collected) > limit {
		collected = collected[:limit]
	}
	return collected, nil
}

func (s *DefaultQuestionBankService) GetQuestion(ctx context.Context, questionID uint) (*model.Question, error) {
	_ = ctx
	if questionID == 0 {
		return nil, fmt.Errorf("question id is required")
	}
	return s.questionRepo.GetByID(questionID)
}

func (s *DefaultQuestionBankService) SetFavorite(ctx context.Context, userID, questionID uint, isFavorite bool) error {
	_ = ctx
	if questionID == 0 {
		return fmt.Errorf("question id is required")
	}
	if _, err := s.questionRepo.GetByID(questionID); err != nil {
		return fmt.Errorf("question not found: %w", err)
	}
	return s.stateRepo.SetFavorite(userID, questionID, isFavorite)
}

func (s *DefaultQuestionBankService) MarkWrongQuestion(ctx context.Context, userID, questionID uint, note string) error {
	_ = ctx
	if questionID == 0 {
		return fmt.Errorf("question id is required")
	}
	if _, err := s.questionRepo.GetByID(questionID); err != nil {
		return fmt.Errorf("question not found: %w", err)
	}
	return s.stateRepo.MarkWrong(userID, questionID, note)
}

func (s *DefaultQuestionBankService) ClearWrongQuestion(ctx context.Context, userID, questionID uint) error {
	_ = ctx
	if questionID == 0 {
		return fmt.Errorf("question id is required")
	}
	return s.stateRepo.ClearWrong(userID, questionID)
}

func (s *DefaultQuestionBankService) ListFavorites(ctx context.Context, userID uint, page, pageSize int) ([]*model.Question, int64, error) {
	_ = ctx
	return s.stateRepo.ListFavorites(userID, page, pageSize)
}

func (s *DefaultQuestionBankService) ListWrongQuestions(ctx context.Context, userID uint, page, pageSize int) ([]*model.Question, int64, error) {
	_ = ctx
	return s.stateRepo.ListWrongQuestions(userID, page, pageSize)
}

func (s *DefaultQuestionBankService) resolvePositionName(positionCode string) string {
	code := strings.TrimSpace(positionCode)
	if code == "" {
		return ""
	}
	position, err := s.positionRepo.GetByCode(code)
	if err == nil && position != nil {
		return strings.TrimSpace(position.Name)
	}
	for _, p := range model.DefaultJobPositions {
		if strings.EqualFold(strings.TrimSpace(p.Code), code) {
			return strings.TrimSpace(p.Name)
		}
	}
	return ""
}
