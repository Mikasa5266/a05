package service

import (
	"fmt"
	"time"

	"your-project/model"
	"your-project/repository"
)

type InterviewService struct {
	interviewRepo *repository.InterviewRepository
	questionRepo  *repository.QuestionRepository
	aiService     *AIService
}

func NewInterviewService() *InterviewService {
	return &InterviewService{
		interviewRepo: repository.NewInterviewRepository(),
		questionRepo:  repository.NewQuestionRepository(),
		aiService:     NewAIService(),
	}
}

// StartInterview now accepts mode and style. It uses AI to generate questions if not found in DB or based on mode.
func (s *InterviewService) StartInterview(userID uint, position, difficulty, mode, style string) (*model.Interview, error) {
	// Try to get questions from DB first (legacy behavior or mixed)
	// Or should we fully switch to AI generation?
	// The prompt implies "AI Driver", so maybe we should generate at least some if DB is empty.

	questions, err := s.questionRepo.GetQuestionsByPositionAndDifficulty(position, difficulty)
	if err != nil {
		// Log error but continue? No, fail.
		return nil, fmt.Errorf("failed to get questions: %w", err)
	}

	// If no questions in DB, or we want to force AI generation (optional),
	// For now, let's keep the existing logic: if DB has questions, use them.
	// If DB is empty, we MUST generate them using AI to support the "AI Interview" feature fully.
	if len(questions) == 0 {
		// Generate 1 initial question to start
		// We need a dummy interview object to pass context
		dummyInterview := &model.Interview{
			Position:   position,
			Difficulty: difficulty,
			Mode:       mode,
			Style:      style,
		}

		// Generate 5 questions (standard length) using ONE call
		generatedQuestions, err := s.aiService.GenerateQuestions(dummyInterview, 5)
		if err != nil {
			return nil, fmt.Errorf("no questions available and AI generation failed: %w", err)
		}

		for _, q := range generatedQuestions {
			// Ensure fields are set
			q.Position = position
			q.Difficulty = difficulty
			s.aiService.EnsureQuestionChinese(q)

			if err := s.questionRepo.Create(q); err != nil {
				return nil, fmt.Errorf("failed to save generated question: %w", err)
			}
			// Use the created question which now has an ID
			questions = append(questions, q)
		}
	} else if len(questions) < 5 {
		// Supplement if not enough
		dummyInterview := &model.Interview{
			Position:   position,
			Difficulty: difficulty,
			Mode:       mode,
			Style:      style,
		}
		needed := 5 - len(questions)
		generatedQuestions, err := s.aiService.GenerateQuestions(dummyInterview, needed)
		if err == nil {
			for _, q := range generatedQuestions {
				q.Position = position
				q.Difficulty = difficulty
				s.aiService.EnsureQuestionChinese(q)
				if err := s.questionRepo.Create(q); err == nil {
					questions = append(questions, q)
				}
			}
		}
	}

	for _, q := range questions {
		s.aiService.EnsureQuestionChinese(q)
	}

	interview := &model.Interview{
		UserID:       userID,
		Position:     position,
		Difficulty:   difficulty,
		Mode:         mode,
		Style:        style,
		Status:       "in_progress",
		StartTime:    time.Now(),
		CurrentIndex: 0,
	}

	if err := s.interviewRepo.Create(interview); err != nil {
		return nil, fmt.Errorf("failed to create interview: %w", err)
	}

	for i, q := range questions {
		iq := &model.InterviewQuestion{
			InterviewID: interview.ID,
			QuestionID:  q.ID,
			OrderIndex:  i,
			IsAnswered:  false,
		}
		if err := s.interviewRepo.CreateInterviewQuestion(iq); err != nil {
			return nil, fmt.Errorf("failed to create interview question: %w", err)
		}
	}

	// Load interview with associated questions
	interviewWithQuestions, err := s.interviewRepo.GetByID(interview.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load interview questions: %w", err)
	}

	return interviewWithQuestions, nil
}

// Package-level wrapper
func StartInterview(userID uint, position, difficulty, mode, style string) (*model.Interview, error) {
	svc := NewInterviewService()
	return svc.StartInterview(userID, position, difficulty, mode, style)
}

func GetInterviewByID(userID, interviewID uint) (*model.Interview, error) {
	svc := NewInterviewService()
	return svc.GetInterviewByID(userID, interviewID)
}

func SubmitAnswer(userID, interviewID, questionID uint, answer, audioData string) (*model.AnswerResult, error) {
	svc := NewInterviewService()
	return svc.SubmitAnswer(userID, interviewID, questionID, answer, audioData)
}

func EndInterview(userID, interviewID uint) (*model.Interview, error) {
	svc := NewInterviewService()
	return svc.EndInterview(userID, interviewID)
}

func GetUserInterviews(userID uint, page, pageSize int) ([]*model.Interview, int64, error) {
	svc := NewInterviewService()
	return svc.GetUserInterviews(userID, page, pageSize)
}

func (s *InterviewService) GetInterviewByID(userID, interviewID uint) (*model.Interview, error) {
	interview, err := s.interviewRepo.GetByID(interviewID)
	if err != nil {
		return nil, err
	}

	if interview.UserID != userID {
		return nil, fmt.Errorf("unauthorized access")
	}

	return interview, nil
}

func (s *InterviewService) SubmitAnswer(userID, interviewID, questionID uint, answer, audioData string) (*model.AnswerResult, error) {
	interview, err := s.GetInterviewByID(userID, interviewID)
	if err != nil {
		return nil, err
	}

	if interview.Status != "in_progress" {
		return nil, fmt.Errorf("interview is not in progress")
	}

	question, err := s.questionRepo.GetByID(questionID)
	if err != nil {
		return nil, fmt.Errorf("question not found")
	}

	var finalAnswer string
	if audioData != "" {
		transcribedText, err := s.aiService.TranscribeAudio(audioData)
		if err != nil {
			return nil, fmt.Errorf("failed to transcribe audio: %w", err)
		}
		finalAnswer = transcribedText
	} else {
		finalAnswer = answer
	}

	evaluation, err := s.aiService.EvaluateAnswer(question, finalAnswer)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate answer: %w", err)
	}

	result := &model.AnswerResult{
		InterviewID: interviewID,
		QuestionID:  questionID,
		Answer:      finalAnswer,
		Score:       evaluation.Score,
		Feedback:    evaluation.Feedback,
		CreatedAt:   time.Now(),
	}

	if err := s.interviewRepo.SaveAnswer(result); err != nil {
		return nil, fmt.Errorf("failed to save answer: %w", err)
	}

	// Mark question as answered in InterviewQuestion mapping
	// This is important if we track progress via InterviewQuestion status
	// But currently we use interview.CurrentIndex.
	// However, let's make sure we don't have logic errors.

	interview.CurrentIndex++
	allQuestions, _ := s.interviewRepo.GetInterviewQuestions(interviewID)

	// Debug logging
	fmt.Printf("Interview %d: CurrentIndex %d, TotalQuestions %d\n", interviewID, interview.CurrentIndex, len(allQuestions))

	if interview.CurrentIndex >= len(allQuestions) {
		interview.Status = "completed"
		t := time.Now()
		interview.EndTime = &t
	}

	if err := s.interviewRepo.Update(interview); err != nil {
		return nil, fmt.Errorf("failed to update interview: %w", err)
	}

	return result, nil
}

func (s *InterviewService) EndInterview(userID, interviewID uint) (*model.Interview, error) {
	interview, err := s.GetInterviewByID(userID, interviewID)
	if err != nil {
		return nil, err
	}

	if interview.Status == "completed" {
		return interview, nil
	}

	interview.Status = "completed"
	t := time.Now()
	interview.EndTime = &t

	if err := s.interviewRepo.Update(interview); err != nil {
		return nil, fmt.Errorf("failed to update interview: %w", err)
	}

	return interview, nil
}

func (s *InterviewService) GetUserInterviews(userID uint, page, pageSize int) ([]*model.Interview, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.interviewRepo.GetByUserID(userID, page, pageSize)
}
