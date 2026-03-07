package repository

import (
	"fmt"

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
		query = query.Where("position = ?", position)
	}
	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
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
		query = query.Where("position = ?", position)
	}
	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}

	err := query.Order("RAND()").Limit(10).Find(&questions).Error
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
