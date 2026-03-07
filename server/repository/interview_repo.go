package repository

import (
	"time"

	"your-project/model"

	"gorm.io/gorm"
)

type InterviewRepository struct {
	db *gorm.DB
}

func NewInterviewRepository() *InterviewRepository {
	return &InterviewRepository{
		db: GetDB(),
	}
}

func (r *InterviewRepository) Create(interview *model.Interview) error {
	return r.db.Create(interview).Error
}

func (r *InterviewRepository) GetByID(id uint) (*model.Interview, error) {
	var interview model.Interview
	err := r.db.Preload("InterviewQuestions", func(db *gorm.DB) *gorm.DB {
		return db.Order("order_index ASC")
	}).Preload("InterviewQuestions.Question").
		Preload("AnswerResults.Question").
		Preload("User").
		First(&interview, id).Error
	if err != nil {
		return nil, err
	}
	return &interview, nil
}

func (r *InterviewRepository) GetByUserID(userID uint, page, pageSize int) ([]*model.Interview, int64, error) {
	var interviews []*model.Interview
	var total int64

	offset := (page - 1) * pageSize

	err := r.db.Model(&model.Interview{}).Where("user_id = ?", userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Preload("InterviewQuestions.Question").
		Preload("InterviewQuestions", func(db *gorm.DB) *gorm.DB {
			return db.Order("order_index ASC")
		}).
		Preload("Report").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&interviews).Error
	if err != nil {
		return nil, 0, err
	}

	return interviews, total, nil
}

func (r *InterviewRepository) Update(interview *model.Interview) error {
	return r.db.Save(interview).Error
}

func (r *InterviewRepository) Delete(id uint) error {
	return r.db.Delete(&model.Interview{}, id).Error
}

func (r *InterviewRepository) SaveAnswer(answer *model.AnswerResult) error {
	return r.db.Create(answer).Error
}

func (r *InterviewRepository) GetAnswersByInterviewID(interviewID uint) ([]model.AnswerResult, error) {
	var answers []model.AnswerResult
	err := r.db.Preload("Question").
		Where("interview_id = ?", interviewID).
		Order("created_at ASC").
		Find(&answers).Error
	return answers, err
}

func (r *InterviewRepository) GetActiveInterview(userID uint) (*model.Interview, error) {
	var interview model.Interview
	err := r.db.Where("user_id = ? AND status = ?", userID, "in_progress").
		Order("created_at DESC").
		First(&interview).Error
	if err != nil {
		return nil, err
	}
	return &interview, nil
}

func (r *InterviewRepository) CreateInterviewQuestion(iq *model.InterviewQuestion) error {
	return r.db.Create(iq).Error
}

func (r *InterviewRepository) GetInterviewQuestions(interviewID uint) ([]model.InterviewQuestion, error) {
	var questions []model.InterviewQuestion
	err := r.db.Preload("Question").
		Where("interview_id = ?", interviewID).
		Order("order_index ASC").
		Find(&questions).Error
	return questions, err
}

func (r *InterviewRepository) UpdateInterviewQuestion(iq *model.InterviewQuestion) error {
	return r.db.Save(iq).Error
}

func (r *InterviewRepository) GetInterviewStatistics(userID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalInterviews int64
	err := r.db.Model(&model.Interview{}).Where("user_id = ?", userID).Count(&totalInterviews).Error
	if err != nil {
		return nil, err
	}
	stats["total_interviews"] = totalInterviews

	var completedInterviews int64
	err = r.db.Model(&model.Interview{}).
		Where("user_id = ? AND status = ?", userID, "completed").
		Count(&completedInterviews).Error
	if err != nil {
		return nil, err
	}
	stats["completed_interviews"] = completedInterviews

	var avgScore float64
	err = r.db.Model(&model.AnswerResult{}).
		Joins("JOIN interviews ON answer_results.interview_id = interviews.id").
		Where("interviews.user_id = ?", userID).
		Select("AVG(score)").
		Scan(&avgScore).Error
	if err != nil {
		avgScore = 0
	}
	stats["average_score"] = avgScore

	var lastInterview struct {
		Position string
		EndTime  time.Time
	}
	err = r.db.Model(&model.Interview{}).
		Where("user_id = ? AND status = ?", userID, "completed").
		Order("end_time DESC").
		Limit(1).
		Select("position", "end_time").
		Scan(&lastInterview).Error
	if err == nil {
		stats["last_interview_position"] = lastInterview.Position
		stats["last_interview_time"] = lastInterview.EndTime
	}

	return stats, nil
}

// ========== Human Interviewer Repository Methods ==========

func (r *InterviewRepository) GetInterviewers(interviewerType string, page, pageSize int) ([]model.HumanInterviewer, int64, error) {
	var interviewers []model.HumanInterviewer
	var total int64

	query := r.db.Model(&model.HumanInterviewer{}).Where("available = ?", true)
	if interviewerType != "" {
		query = query.Where("type = ?", interviewerType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("rating DESC").Limit(pageSize).Offset(offset).Find(&interviewers).Error; err != nil {
		return nil, 0, err
	}

	return interviewers, total, nil
}

func (r *InterviewRepository) GetInterviewerByID(id uint) (*model.HumanInterviewer, error) {
	var interviewer model.HumanInterviewer
	if err := r.db.First(&interviewer, id).Error; err != nil {
		return nil, err
	}
	return &interviewer, nil
}

func (r *InterviewRepository) CreateBooking(booking *model.InterviewBooking) error {
	return r.db.Create(booking).Error
}

func (r *InterviewRepository) GetBookingsByUserID(userID uint) ([]model.InterviewBooking, error) {
	var bookings []model.InterviewBooking
	err := r.db.Preload("Interviewer").
		Where("user_id = ?", userID).
		Order("scheduled_at DESC").
		Find(&bookings).Error
	return bookings, err
}

func (r *InterviewRepository) UpdateBooking(booking *model.InterviewBooking) error {
	return r.db.Save(booking).Error
}

// InsertQuestionAt inserts a question at a specific index and shifts subsequent questions
func (r *InterviewRepository) InsertQuestionAt(interviewID uint, iq *model.InterviewQuestion, index int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Shift existing questions down
		if err := tx.Model(&model.InterviewQuestion{}).
			Where("interview_id = ? AND order_index >= ?", interviewID, index).
			Update("order_index", gorm.Expr("order_index + 1")).Error; err != nil {
			return err
		}

		// 2. Insert new question
		if err := tx.Create(iq).Error; err != nil {
			return err
		}
		return nil
	})
}
