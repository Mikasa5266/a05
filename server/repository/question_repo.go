package repository

import (
	"fmt"
	"strings"

	"your-project/model"

	"gorm.io/gorm"
)

// QuestionRepository defines question aggregate persistence contracts for service-layer DI.
type QuestionRepository interface {
	Create(question *model.Question) error
	BatchCreate(questions []*model.Question) error
	GetByID(id uint) (*model.Question, error)
	GetQuestions(position, difficulty, category string) ([]*model.Question, error)
	GetQuestionsByPositionAndDifficulty(position, difficulty string) ([]*model.Question, error)
	GetQuestionsByPositionAndDifficultyWithExclude(position, difficulty string, excludeIDs []uint) ([]*model.Question, error)
	GetQuestionsForInterviewInit(position, difficulty, category string, limit int) ([]*model.Question, error)
	GetQuestionsForInterviewInitWithExclude(position, difficulty, category string, limit int, excludeIDs []uint) ([]*model.Question, error)
	GetRandomQuestionForInterview(position, difficulty string, excludeIDs []uint) (*model.Question, error)
	FindByKnowledgePoint(position, difficulty, knowledgePoint string, limit int) ([]*model.Question, error)
	ListRandomByPosition(position string, limit int) ([]*model.Question, error)
	Update(question *model.Question) error
	Delete(id uint) error
	List(page, pageSize int) ([]*model.Question, int64, error)
	SearchByKeyword(keyword string, page, pageSize int) ([]*model.Question, int64, error)
	GetQuestionStats() (map[string]interface{}, error)
}

type GormQuestionRepository struct {
	db *gorm.DB
}

var _ QuestionRepository = (*GormQuestionRepository)(nil)

func NewQuestionRepository() QuestionRepository {
	return &GormQuestionRepository{db: GetDB()}
}

func NewQuestionRepositoryWithDB(db *gorm.DB) QuestionRepository {
	if db == nil {
		db = GetDB()
	}
	return &GormQuestionRepository{db: db}
}

func (r *GormQuestionRepository) Create(question *model.Question) error {
	return r.db.Create(question).Error
}

func (r *GormQuestionRepository) BatchCreate(questions []*model.Question) error {
	if len(questions) == 0 {
		return nil
	}
	return r.db.Create(&questions).Error
}

func (r *GormQuestionRepository) GetByID(id uint) (*model.Question, error) {
	var question model.Question
	err := r.db.First(&question, id).Error
	if err != nil {
		return nil, err
	}
	return &question, nil
}

func (r *GormQuestionRepository) GetQuestions(position, difficulty, category string) ([]*model.Question, error) {
	query := applyQuestionBaseFilters(r.db.Model(&model.Question{}), position, difficulty, category, nil)
	var questions []*model.Question
	err := query.Find(&questions).Error
	if err != nil {
		return nil, err
	}
	return questions, nil
}

func (r *GormQuestionRepository) GetQuestionsByPositionAndDifficulty(position, difficulty string) ([]*model.Question, error) {
	return r.GetQuestionsByPositionAndDifficultyWithExclude(position, difficulty, nil)
}

func (r *GormQuestionRepository) GetQuestionsByPositionAndDifficultyWithExclude(position, difficulty string, excludeIDs []uint) ([]*model.Question, error) {
	var questions []*model.Question
	query := applyQuestionBaseFilters(r.db.Model(&model.Question{}), position, difficulty, "", excludeIDs)
	err := query.Order(randomOrderExpression(r.db)).Limit(10).Find(&questions).Error
	if err != nil {
		return nil, err
	}
	return questions, nil
}

func (r *GormQuestionRepository) GetQuestionsForInterviewInit(position, difficulty, category string, limit int) ([]*model.Question, error) {
	return r.GetQuestionsForInterviewInitWithExclude(position, difficulty, category, limit, nil)
}

func (r *GormQuestionRepository) GetQuestionsForInterviewInitWithExclude(position, difficulty, category string, limit int, excludeIDs []uint) ([]*model.Question, error) {
	if limit <= 0 {
		limit = 12
	}

	query := applyQuestionBaseFilters(r.db.Model(&model.Question{}), position, difficulty, category, excludeIDs)
	var questions []*model.Question
	err := query.Order("id DESC").Limit(limit).Find(&questions).Error
	if err != nil {
		return nil, err
	}
	return questions, nil
}

func (r *GormQuestionRepository) GetRandomQuestionForInterview(position, difficulty string, excludeIDs []uint) (*model.Question, error) {
	query := applyQuestionBaseFilters(r.db.Model(&model.Question{}), position, difficulty, "", excludeIDs)
	var question model.Question
	err := query.Order(randomOrderExpression(r.db)).Limit(1).Take(&question).Error
	if err != nil {
		return nil, err
	}
	return &question, nil
}

// FindByKnowledgePoint maps app.py/SQLite style "按考点查题" into GORM queries.
func (r *GormQuestionRepository) FindByKnowledgePoint(position, difficulty, knowledgePoint string, limit int) ([]*model.Question, error) {
	if limit <= 0 {
		limit = 20
	}

	kp := strings.TrimSpace(knowledgePoint)
	query := applyQuestionBaseFilters(r.db.Model(&model.Question{}), position, difficulty, "", nil)
	if kp != "" {
		pattern := fmt.Sprintf("%%%s%%", kp)
		query = query.Where("knowledge_point = ? OR knowledge_area = ? OR category = ? OR tags LIKE ?", kp, kp, kp, pattern)
	}

	var questions []*model.Question
	err := query.Order("difficulty_rank DESC, id DESC").Limit(limit).Find(&questions).Error
	if err != nil {
		return nil, err
	}
	return questions, nil
}

// ListRandomByPosition maps app.py/SQLite style "按岗位随机抽题" list query.
func (r *GormQuestionRepository) ListRandomByPosition(position string, limit int) ([]*model.Question, error) {
	if limit <= 0 {
		limit = 10
	}

	query := applyQuestionBaseFilters(r.db.Model(&model.Question{}), position, "", "", nil)
	var questions []*model.Question
	err := query.Order(randomOrderExpression(r.db)).Limit(limit).Find(&questions).Error
	if err != nil {
		return nil, err
	}
	return questions, nil
}

func (r *GormQuestionRepository) Update(question *model.Question) error {
	return r.db.Save(question).Error
}

func (r *GormQuestionRepository) Delete(id uint) error {
	return r.db.Delete(&model.Question{}, id).Error
}

func (r *GormQuestionRepository) List(page, pageSize int) ([]*model.Question, int64, error) {
	var questions []*model.Question
	var total int64

	offset := (page - 1) * pageSize

	err := r.db.Model(&model.Question{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Limit(pageSize).Offset(offset).Find(&questions).Error
	if err != nil {
		return nil, 0, err
	}

	return questions, total, nil
}

func (r *GormQuestionRepository) SearchByKeyword(keyword string, page, pageSize int) ([]*model.Question, int64, error) {
	var questions []*model.Question
	var total int64

	offset := (page - 1) * pageSize
	searchPattern := fmt.Sprintf("%%%s%%", keyword)

	err := r.db.Model(&model.Question{}).
		Where("title LIKE ? OR content LIKE ? OR tags LIKE ? OR knowledge_point LIKE ? OR category LIKE ?", searchPattern, searchPattern, searchPattern, searchPattern, searchPattern).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("title LIKE ? OR content LIKE ? OR tags LIKE ? OR knowledge_point LIKE ? OR category LIKE ?", searchPattern, searchPattern, searchPattern, searchPattern, searchPattern).
		Limit(pageSize).
		Offset(offset).
		Find(&questions).Error
	if err != nil {
		return nil, 0, err
	}

	return questions, total, nil
}

func (r *GormQuestionRepository) GetQuestionStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalQuestions int64
	err := r.db.Model(&model.Question{}).Count(&totalQuestions).Error
	if err != nil {
		return nil, err
	}
	stats["total_questions"] = totalQuestions

	var positionStats []struct {
		Position string
		Count    int64
	}
	err = r.db.Model(&model.Question{}).
		Select("position, COUNT(*) as count").
		Group("position").
		Scan(&positionStats).Error
	if err != nil {
		return nil, err
	}
	stats["by_position"] = positionStats

	var knowledgePointStats []struct {
		KnowledgePoint string
		Count          int64
	}
	err = r.db.Model(&model.Question{}).
		Select("knowledge_point, COUNT(*) as count").
		Where("knowledge_point IS NOT NULL AND knowledge_point <> ''").
		Group("knowledge_point").
		Scan(&knowledgePointStats).Error
	if err != nil {
		return nil, err
	}
	stats["by_knowledge_point"] = knowledgePointStats

	var difficultyStats []struct {
		Difficulty string
		Count      int64
	}
	err = r.db.Model(&model.Question{}).
		Select("difficulty, COUNT(*) as count").
		Group("difficulty").
		Scan(&difficultyStats).Error
	if err != nil {
		return nil, err
	}
	stats["by_difficulty"] = difficultyStats

	return stats, nil
}

func buildPositionCandidates(position string) []string {
	p := strings.ToLower(strings.TrimSpace(position))
	switch {
	case p == "", strings.Contains(p, "java"), strings.Contains(p, "后端"), p == "backend":
		return []string{"Java后端工程师", "后端工程师", "后端开发", "Java", "Go", "Python"}
	case strings.Contains(p, "前端"), strings.Contains(p, "frontend"):
		return []string{"前端工程师", "前端开发工程师", "前端开发", "Frontend"}
	case strings.Contains(p, "算法"), p == "algorithm":
		return []string{"算法工程师", "Algorithm", "算法"}
	case strings.Contains(p, "ai"), strings.Contains(p, "llm"), strings.Contains(p, "模型"), p == "ai_engineer":
		return []string{"AI工程师", "ai_engineer", "AI", "机器学习", "Python"}
	default:
		return []string{position}
	}
}

func applyQuestionBaseFilters(query *gorm.DB, position, difficulty, category string, excludeIDs []uint) *gorm.DB {
	filtered := query.Where("is_active = ?", true).
		Where("(source IS NULL OR source <> ?) AND (rag_eligible IS NULL OR rag_eligible = ?)", "follow_up", true)

	if len(excludeIDs) > 0 {
		filtered = filtered.Where("id NOT IN ?", excludeIDs)
	}

	if position != "" {
		candidates := buildPositionCandidates(position)
		if len(candidates) > 0 {
			filtered = filtered.Where("position IN ?", candidates)
		} else {
			filtered = filtered.Where("position = ?", position)
		}
	}

	if difficulty != "" {
		candidates := buildDifficultyCandidates(difficulty)
		if len(candidates) > 0 {
			filtered = filtered.Where("difficulty IN ?", candidates)
		} else {
			filtered = filtered.Where("difficulty = ?", difficulty)
		}
	}

	if category != "" {
		filtered = filtered.Where("category = ?", category)
	}

	return filtered
}

func randomOrderExpression(db *gorm.DB) string {
	if db != nil && db.Dialector != nil {
		switch strings.ToLower(strings.TrimSpace(db.Dialector.Name())) {
		case "postgres", "postgresql":
			return "RANDOM()"
		}
	}
	return "RAND()"
}

func buildDifficultyCandidates(difficulty string) []string {
	d := strings.ToLower(strings.TrimSpace(difficulty))
	base := []string{difficulty}

	var aliases []string
	switch d {
	case "campus_intern", "easy":
		aliases = []string{"campus_intern", "easy", "junior", "Junior", "初级", "基础"}
	case "campus_graduate", "medium":
		aliases = []string{"campus_graduate", "medium", "intermediate", "Intermediate", "junior", "Junior", "中级"}
	case "social_junior", "hard":
		aliases = []string{"social_junior", "hard", "senior", "Senior", "intermediate", "Intermediate", "中高级"}
	case "junior":
		aliases = []string{"junior", "Junior", "campus_intern", "easy", "初级"}
	case "intermediate":
		aliases = []string{"intermediate", "Intermediate", "campus_graduate", "medium", "中级"}
	case "senior":
		aliases = []string{"senior", "Senior", "social_junior", "hard", "中高级"}
	default:
		aliases = []string{difficulty}
	}

	seen := map[string]bool{}
	result := make([]string, 0, len(base)+len(aliases))
	for _, item := range append(base, aliases...) {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		result = append(result, trimmed)
	}
	return result
}
