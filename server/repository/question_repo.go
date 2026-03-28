package repository

import (
	"fmt"
	"strings"

	"your-project/model"

	"gorm.io/gorm"
)

type QuestionRepository struct {
	db *gorm.DB
}

func NewQuestionRepository() *QuestionRepository {
	return &QuestionRepository{
		db: GetDB(),
	}
}

func (r *QuestionRepository) Create(question *model.Question) error {
	return r.db.Create(question).Error
}

func (r *QuestionRepository) GetByID(id uint) (*model.Question, error) {
	var question model.Question
	err := r.db.First(&question, id).Error
	if err != nil {
		return nil, err
	}
	return &question, nil
}

func (r *QuestionRepository) GetQuestions(position, difficulty, category string) ([]*model.Question, error) {
	query := r.db.Model(&model.Question{}).
		Where("(source IS NULL OR source <> ?) AND (rag_eligible IS NULL OR rag_eligible = ?)", "follow_up", true)

	if position != "" {
		candidates := buildPositionCandidates(position)
		if len(candidates) > 0 {
			query = query.Where("position IN ?", candidates)
		} else {
			query = query.Where("position = ?", position)
		}
	}
	if difficulty != "" {
		candidates := buildDifficultyCandidates(difficulty)
		if len(candidates) > 0 {
			query = query.Where("difficulty IN ?", candidates)
		} else {
			query = query.Where("difficulty = ?", difficulty)
		}
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}

	var questions []*model.Question
	err := query.Find(&questions).Error
	if err != nil {
		return nil, err
	}

	return questions, nil
}

func (r *QuestionRepository) GetQuestionsByPositionAndDifficulty(position, difficulty string) ([]*model.Question, error) {
	var questions []*model.Question
	query := r.db.Model(&model.Question{}).
		Where("(source IS NULL OR source <> ?) AND (rag_eligible IS NULL OR rag_eligible = ?)", "follow_up", true)

	if position != "" {
		candidates := buildPositionCandidates(position)
		if len(candidates) > 0 {
			query = query.Where("position IN ?", candidates)
		} else {
			query = query.Where("position = ?", position)
		}
	}
	if difficulty != "" {
		candidates := buildDifficultyCandidates(difficulty)
		if len(candidates) > 0 {
			query = query.Where("difficulty IN ?", candidates)
		} else {
			query = query.Where("difficulty = ?", difficulty)
		}
	}

	err := query.Order("RAND()").Limit(10).Find(&questions).Error
	if err != nil {
		return nil, err
	}

	return questions, nil
}

// GetQuestionsForInterviewInit fetches a bounded, deterministic set of questions for fast startup.
func (r *QuestionRepository) GetQuestionsForInterviewInit(position, difficulty, category string, limit int) ([]*model.Question, error) {
	if limit <= 0 {
		limit = 12
	}

	query := r.db.Model(&model.Question{}).
		Where("(source IS NULL OR source <> ?) AND (rag_eligible IS NULL OR rag_eligible = ?)", "follow_up", true)

	if position != "" {
		candidates := buildPositionCandidates(position)
		if len(candidates) > 0 {
			query = query.Where("position IN ?", candidates)
		} else {
			query = query.Where("position = ?", position)
		}
	}
	if difficulty != "" {
		candidates := buildDifficultyCandidates(difficulty)
		if len(candidates) > 0 {
			query = query.Where("difficulty IN ?", candidates)
		} else {
			query = query.Where("difficulty = ?", difficulty)
		}
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}

	var questions []*model.Question
	err := query.Order("id DESC").Limit(limit).Find(&questions).Error
	if err != nil {
		return nil, err
	}

	return questions, nil
}

func (r *QuestionRepository) Update(question *model.Question) error {
	return r.db.Save(question).Error
}

func (r *QuestionRepository) Delete(id uint) error {
	return r.db.Delete(&model.Question{}, id).Error
}

func (r *QuestionRepository) List(page, pageSize int) ([]*model.Question, int64, error) {
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

func (r *QuestionRepository) SearchByKeyword(keyword string, page, pageSize int) ([]*model.Question, int64, error) {
	var questions []*model.Question
	var total int64

	offset := (page - 1) * pageSize

	searchPattern := fmt.Sprintf("%%%s%%", keyword)

	err := r.db.Model(&model.Question{}).
		Where("title LIKE ? OR content LIKE ? OR tags LIKE ?", searchPattern, searchPattern, searchPattern).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("title LIKE ? OR content LIKE ? OR tags LIKE ?", searchPattern, searchPattern, searchPattern).
		Limit(pageSize).
		Offset(offset).
		Find(&questions).Error
	if err != nil {
		return nil, 0, err
	}

	return questions, total, nil
}

func (r *QuestionRepository) GetQuestionStats() (map[string]interface{}, error) {
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
