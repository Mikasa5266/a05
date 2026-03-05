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
	ragService    *RAGService // Add RAG service
}

func NewInterviewService() *InterviewService {
	return &InterviewService{
		interviewRepo: repository.NewInterviewRepository(),
		questionRepo:  repository.NewQuestionRepository(),
		aiService:     NewAIService(),
		ragService:    GetRAGService(), // Init RAG service
	}
}

// StartInterview now accepts mode, style, company, and interviewMode. It uses AI to generate questions based on these parameters.
func (s *InterviewService) StartInterview(userID uint, position, difficulty, mode, style, company, interviewMode string) (*model.Interview, error) {
	var questions []*model.Question
	var scenarioJSON string
	var revealedStyle string
	var capabilityGraph *model.JobCapabilityDimension

	// ==== Dynamic Adapter: Load Capability Graph ====
	// Try to find capability graph from RAG/KnowledgeBase or Enterprise settings
	// For now, we simulate loading it based on position
	// In a real scenario, we might query the enterprise_jobs table or RAG
	capabilityGraph = s.loadJobCapabilityGraph(position)

	// ==== Random Mode: pick a random style/company, don't tell the user ====
	if interviewMode == "random" {
		randomStyle, randomCompany := GenerateRandomStyleForInterview()
		style = randomStyle
		company = randomCompany
		revealedStyle = style // will be stored but not shown until end
	}

	// Blindbox mode: draw a random scenario and generate tailored questions
	if mode == "blindbox" {
		bbService := NewBlindBoxService()
		scenario := bbService.DrawScenario()
		scenarioJSON = ScenarioToJSON(scenario)

		// Override style with scenario's style
		style = scenario.Style

		// Generate questions tailored to this scenario
		generated, err := bbService.GenerateBlindBoxQuestions(scenario, position, difficulty, 5)
		if err != nil {
			return nil, fmt.Errorf("blindbox question generation failed: %w", err)
		}
		for _, q := range generated {
			q.Position = position
			q.Difficulty = difficulty
			if err := s.questionRepo.Create(q); err != nil {
				return nil, fmt.Errorf("failed to save blindbox question: %w", err)
			}
			questions = append(questions, q)
		}
	} else {
		// Standard mode: fetch from DB or generate via AI
		var err error
		questions, err = s.questionRepo.GetQuestionsByPositionAndDifficulty(position, difficulty)
		if err != nil {
			return nil, fmt.Errorf("failed to get questions: %w", err)
		}

		// Use Dynamic Adapter to generate tailored questions if needed
		if len(questions) < 5 {
			dummyInterview := &model.Interview{
				Position:   position,
				Difficulty: difficulty,
				Mode:       mode,
				Style:      style,
				Company:    company,
			}
			needed := 5 - len(questions)

			// Generate questions with weights awareness
			for i := 0; i < needed; i++ {
				// We pass nil for previous answers as this is initial generation
				q, err := s.aiService.GenerateNextQuestionWithWeights(dummyInterview, nil, capabilityGraph)
				if err != nil {
					// Fallback to standard generation
					q, err = s.aiService.GenerateNextQuestion(dummyInterview, nil)
					if err != nil {
						continue // Skip if both fail
					}
				}

				// Ensure consistency
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
		UserID:        userID,
		Position:      position,
		Difficulty:    difficulty,
		Mode:          mode,
		Style:         style,
		Company:       company,
		InterviewMode: interviewMode,
		RevealedStyle: revealedStyle,
		Scenario:      scenarioJSON,
		Status:        "in_progress",
		StartTime:     time.Now(),
		CurrentIndex:  0,
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
func StartInterview(userID uint, position, difficulty, mode, style, company, interviewMode string) (*model.Interview, error) {
	svc := NewInterviewService()
	return svc.StartInterview(userID, position, difficulty, mode, style, company, interviewMode)
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

// loadJobCapabilityGraph simulates loading job capability graph
// In production, this would fetch from DB or RAG
func (s *InterviewService) loadJobCapabilityGraph(position string) *model.JobCapabilityDimension {
	switch position {
	case "后端开发", "backend":
		return &model.JobCapabilityDimension{
			Name:   "后端开发",
			Weight: 100,
			SubDimensions: []model.JobCapabilitySubDimension{
				{Name: "JVM原理", Weight: 30, Tags: []string{"GC", "Memory Model", "Classloader"}},
				{Name: "分布式系统", Weight: 25, Tags: []string{"CAP", "Microservices", "RPC"}},
				{Name: "数据库优化", Weight: 20, Tags: []string{"Index", "SQL Tuning", "Locking"}},
				{Name: "网络协议", Weight: 15, Tags: []string{"TCP/IP", "HTTP", "Socket"}},
				{Name: "并发编程", Weight: 10, Tags: []string{"Go Routine", "Channel", "Sync"}},
			},
		}
	case "前端开发", "frontend":
		return &model.JobCapabilityDimension{
			Name:   "前端开发",
			Weight: 100,
			SubDimensions: []model.JobCapabilitySubDimension{
				{Name: "React/Vue框架深度", Weight: 30, Tags: []string{"Virtual DOM", "Hooks", "Reactivity"}},
				{Name: "JavaScript核心", Weight: 25, Tags: []string{"ES6+", "Closure", "Prototype"}},
				{Name: "工程化", Weight: 20, Tags: []string{"Webpack", "Vite", "CI/CD"}},
				{Name: "性能优化", Weight: 15, Tags: []string{"Rendering", "Network", "Cache"}},
				{Name: "CSS/布局", Weight: 10, Tags: []string{"Flexbox", "Grid", "Animation"}},
			},
		}
	default:
		// Generic default graph
		return &model.JobCapabilityDimension{
			Name:   position,
			Weight: 100,
			SubDimensions: []model.JobCapabilitySubDimension{
				{Name: "核心技术", Weight: 40, Tags: []string{"Core Concepts", "Architecture"}},
				{Name: "实战经验", Weight: 30, Tags: []string{"Project", "Problem Solving"}},
				{Name: "基础知识", Weight: 30, Tags: []string{"Algorithms", "Data Structures"}},
			},
		}
	}
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

// ========== Human Interviewer Functions ==========

func GetHumanInterviewers(interviewerType string, page, pageSize int) ([]model.HumanInterviewer, int64, error) {
	svc := NewInterviewService()
	return svc.interviewRepo.GetInterviewers(interviewerType, page, pageSize)
}

func GetHumanInterviewerByID(id uint) (*model.HumanInterviewer, error) {
	svc := NewInterviewService()
	return svc.interviewRepo.GetInterviewerByID(id)
}

func BookHumanInterview(userID, interviewerID uint, scheduledAt time.Time, position, difficulty, notes string) (*model.InterviewBooking, error) {
	svc := NewInterviewService()

	interviewer, err := svc.interviewRepo.GetInterviewerByID(interviewerID)
	if err != nil {
		return nil, fmt.Errorf("interviewer not found: %w", err)
	}
	if !interviewer.Available {
		return nil, fmt.Errorf("该面试官当前不可预约")
	}

	booking := &model.InterviewBooking{
		UserID:        userID,
		InterviewerID: interviewerID,
		ScheduledAt:   scheduledAt,
		Position:      position,
		Difficulty:    difficulty,
		Status:        "pending",
		Notes:         notes,
	}

	if err := svc.interviewRepo.CreateBooking(booking); err != nil {
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}

	return booking, nil
}

func GetUserBookings(userID uint) ([]model.InterviewBooking, error) {
	svc := NewInterviewService()
	return svc.interviewRepo.GetBookingsByUserID(userID)
}

// SubmitHumanFeedback allows a human interviewer to submit feedback after an interview
func SubmitHumanFeedback(interviewID uint, feedback string, score int) error {
	svc := NewInterviewService()
	interview, err := svc.interviewRepo.GetByID(interviewID)
	if err != nil {
		return fmt.Errorf("interview not found: %w", err)
	}

	interview.HumanFeedback = feedback
	interview.HumanScore = &score

	return svc.interviewRepo.Update(interview)
}

// RevealRandomStyle returns the hidden style for a random-mode interview (after completion)
func RevealRandomStyle(userID, interviewID uint) (string, string, error) {
	svc := NewInterviewService()
	interview, err := svc.GetInterviewByID(userID, interviewID)
	if err != nil {
		return "", "", err
	}
	if interview.InterviewMode != "random" {
		return interview.Style, interview.Company, nil
	}
	if interview.Status != "completed" {
		return "", "", fmt.Errorf("面试尚未结束，无法揭晓风格")
	}
	return interview.RevealedStyle, interview.Company, nil
}
