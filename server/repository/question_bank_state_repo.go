package repository

import (
	"errors"
	"strings"
	"time"

	"your-project/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UserQuestionStateRepository provides persistence access for favorites and wrong-question sets.
type UserQuestionStateRepository interface {
	SetFavorite(userID, questionID uint, isFavorite bool) error
	MarkWrong(userID, questionID uint, note string) error
	ClearWrong(userID, questionID uint) error
	GetState(userID, questionID uint) (*model.UserQuestionState, error)
	ListFavorites(userID uint, page, pageSize int) ([]*model.Question, int64, error)
	ListWrongQuestions(userID uint, page, pageSize int) ([]*model.Question, int64, error)
}

type GormUserQuestionStateRepository struct {
	db *gorm.DB
}

var _ UserQuestionStateRepository = (*GormUserQuestionStateRepository)(nil)

func NewUserQuestionStateRepository() UserQuestionStateRepository {
	return &GormUserQuestionStateRepository{db: GetDB()}
}

func NewUserQuestionStateRepositoryWithDB(db *gorm.DB) UserQuestionStateRepository {
	if db == nil {
		db = GetDB()
	}
	return &GormUserQuestionStateRepository{db: db}
}

func (r *GormUserQuestionStateRepository) SetFavorite(userID, questionID uint, isFavorite bool) error {
	now := time.Now()
	state := &model.UserQuestionState{
		UserID:         userID,
		QuestionID:     questionID,
		IsFavorite:     isFavorite,
		LastAnsweredAt: &now,
	}

	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "question_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"is_favorite":      isFavorite,
			"last_answered_at": now,
			"updated_at":       now,
		}),
	}).Create(state).Error
}

func (r *GormUserQuestionStateRepository) MarkWrong(userID, questionID uint, note string) error {
	now := time.Now()
	normalizedNote := strings.TrimSpace(note)

	return r.db.Transaction(func(tx *gorm.DB) error {
		var existing model.UserQuestionState
		err := tx.Where("user_id = ? AND question_id = ?", userID, questionID).First(&existing).Error
		if err == nil {
			existing.IsWrong = true
			existing.WrongCount = existing.WrongCount + 1
			existing.LastWrongAt = &now
			existing.LastAnsweredAt = &now
			if normalizedNote != "" {
				existing.Note = normalizedNote
			}
			return tx.Save(&existing).Error
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		state := &model.UserQuestionState{
			UserID:         userID,
			QuestionID:     questionID,
			IsWrong:        true,
			WrongCount:     1,
			LastWrongAt:    &now,
			LastAnsweredAt: &now,
			Note:           normalizedNote,
		}
		return tx.Create(state).Error
	})
}

func (r *GormUserQuestionStateRepository) ClearWrong(userID, questionID uint) error {
	now := time.Now()
	return r.db.Model(&model.UserQuestionState{}).
		Where("user_id = ? AND question_id = ?", userID, questionID).
		Updates(map[string]interface{}{
			"is_wrong":         false,
			"last_answered_at": now,
			"updated_at":       now,
		}).Error
}

func (r *GormUserQuestionStateRepository) GetState(userID, questionID uint) (*model.UserQuestionState, error) {
	var state model.UserQuestionState
	err := r.db.Preload("Question").Where("user_id = ? AND question_id = ?", userID, questionID).First(&state).Error
	if err != nil {
		return nil, err
	}
	return &state, nil
}

func (r *GormUserQuestionStateRepository) ListFavorites(userID uint, page, pageSize int) ([]*model.Question, int64, error) {
	return r.listQuestionsByFlag(userID, page, pageSize, "is_favorite = ?", true)
}

func (r *GormUserQuestionStateRepository) ListWrongQuestions(userID uint, page, pageSize int) ([]*model.Question, int64, error) {
	return r.listQuestionsByFlag(userID, page, pageSize, "is_wrong = ?", true)
}

func (r *GormUserQuestionStateRepository) listQuestionsByFlag(userID uint, page, pageSize int, condition string, value bool) ([]*model.Question, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	query := r.db.Model(&model.UserQuestionState{}).
		Where("user_id = ?", userID).
		Where(condition, value)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var states []model.UserQuestionState
	err := query.Preload("Question").Order("updated_at DESC").Limit(pageSize).Offset(offset).Find(&states).Error
	if err != nil {
		return nil, 0, err
	}

	questions := make([]*model.Question, 0, len(states))
	for i := range states {
		if states[i].Question.ID == 0 {
			continue
		}
		q := states[i].Question
		questions = append(questions, &q)
	}

	return questions, total, nil
}
