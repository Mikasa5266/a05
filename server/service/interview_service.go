package service

import (
	"fmt"
	"strings"
	"time"

	"your-project/model"
	"your-project/pkg/websocket"
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

	topicCount, totalTarget := buildInterviewPlan(difficulty)
	topicQuestionMin, topicQuestionMax, maxFollowUps := 2, 4, 3

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
		generated, err := bbService.GenerateBlindBoxQuestions(scenario, position, difficulty, topicCount)
		if err != nil {
			return nil, fmt.Errorf("blindbox question generation failed: %w", err)
		}
		for _, q := range generated {
			q.Position = position
			q.Difficulty = difficulty
			q.Source = "standard"
			q.RAGEligible = true
			s.normalizeOpeningQuestion(q)
			if err := s.questionRepo.Create(q); err != nil {
				return nil, fmt.Errorf("failed to save blindbox question: %w", err)
			}
			questions = append(questions, q)
		}
	} else {
		// Standard mode: build topic-opening questions via RAG
		dummyInterview := &model.Interview{
			Position:   position,
			Difficulty: difficulty,
			Mode:       mode,
			Style:      style,
			Company:    company,
		}

		if s.ragService != nil {
			query := fmt.Sprintf("%s %s 面试 题目", position, difficulty)
			chunks, err := s.ragService.SearchKnowledgeChunksWithLimit(query, topicCount)
			if err == nil {
				for _, chunk := range chunks {
					q, qErr := s.aiService.GenerateTopicQuestionFromContext(dummyInterview, chunk.Content, chunk.Category)
					if qErr == nil && q != nil {
						if normalized := s.normalizeOpeningQuestion(q); normalized != nil {
							questions = append(questions, q)
						}
					}
				}
			}
		}

		// Fill from question bank if needed
		if len(questions) < topicCount {
			fallback, err := s.questionRepo.GetQuestionsByPositionAndDifficulty(position, difficulty)
			if err != nil {
				return nil, fmt.Errorf("failed to get questions: %w", err)
			}
			for _, q := range fallback {
				if len(questions) >= topicCount {
					break
				}
				if normalized := s.normalizeOpeningQuestion(q); normalized != nil {
					questions = append(questions, normalized)
				}
			}
		}

		// Final fallback: generate topic questions if still short
		if len(questions) < topicCount {
			needed := topicCount - len(questions)
			maxAttempts := needed * 3
			for i := 0; i < maxAttempts && len(questions) < topicCount; i++ {
				q, err := s.aiService.GenerateNextQuestionWithWeights(dummyInterview, nil, capabilityGraph)
				if err != nil {
					q, err = s.aiService.GenerateNextQuestion(dummyInterview, nil)
					if err != nil {
						continue
					}
				}
				q.Position = position
				q.Difficulty = difficulty
				q.Source = "standard"
				q.RAGEligible = true
				if normalized := s.normalizeOpeningQuestion(q); normalized == nil {
					continue
				}
				if err := s.questionRepo.Create(q); err == nil {
					questions = append(questions, q)
				}
			}
		}
	}

	if len(questions) > topicCount {
		questions = questions[:topicCount]
	}

	for _, q := range questions {
		s.normalizeOpeningQuestion(q)
	}

	// Style could be overridden by random/blindbox mode, so compute strategy at the end.
	topicQuestionMin, topicQuestionMax, maxFollowUps = buildStyleQuestionPlan(style)

	interview := &model.Interview{
		UserID:              userID,
		Position:            position,
		Difficulty:          difficulty,
		Mode:                mode,
		Style:               style,
		Company:             company,
		InterviewMode:       interviewMode,
		RevealedStyle:       revealedStyle,
		Scenario:            scenarioJSON,
		Status:              "in_progress",
		StartTime:           time.Now(),
		CurrentIndex:        0,
		MaxFollowUps:        maxFollowUps,
		TopicIndex:          0,
		TopicCountTarget:    topicCount,
		TopicQuestionMin:    topicQuestionMin,
		TopicQuestionMax:    topicQuestionMax,
		TotalQuestionTarget: totalTarget,
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

func (s *InterviewService) normalizeOpeningQuestion(q *model.Question) *model.Question {
	if q == nil {
		return nil
	}
	if strings.TrimSpace(q.Source) == "" {
		q.Source = "standard"
	}
	if q.Source == "follow_up" {
		return nil
	}

	s.aiService.EnsureQuestionChinese(q)
	if s.aiService.IsContextDependentOpeningQuestion(q) {
		s.aiService.NormalizeToSelfContainedOpening(q)
	}
	return q
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

func SaveInterviewRecording(userID, interviewID uint, recordingURL string) (*model.Interview, error) {
	svc := NewInterviewService()
	return svc.SaveInterviewRecording(userID, interviewID, recordingURL)
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
	userStr := fmt.Sprintf("%d", userID)

	if audioData != "" {
		websocket.GetHub().SendToUser(userStr, "interviewer.transcribing", nil)
		transcribedText, err := s.aiService.TranscribeAudio(audioData)
		if err != nil {
			return nil, fmt.Errorf("failed to transcribe audio: %w", err)
		}
		finalAnswer = transcribedText
	} else {
		finalAnswer = answer
	}

	websocket.GetHub().SendToUser(userStr, "interviewer.thinking", nil)
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

	answers, _ := s.interviewRepo.GetAnswersByInterviewID(interviewID)
	if shouldEndEarlyForZeroScores(answers, 3) {
		interview.Status = "completed"
		now := time.Now()
		interview.EndTime = &now
		interview.CurrentIndex = len(answers)
		if err := s.interviewRepo.Update(interview); err != nil {
			return nil, fmt.Errorf("failed to update interview: %w", err)
		}
		return result, nil
	}

	// ==== Dynamic Follow-Up Logic ====
	// Instead of just incrementing index, we check if we should ask a follow-up
	shouldFollowUp, nextQuestion, err := s.decideNextQuestion(interview, question, finalAnswer, evaluation.Score)
	if err != nil {
		// Log error but continue with standard flow if dynamic fails
		fmt.Printf("Dynamic question generation failed: %v\n", err)
	}

	if shouldFollowUp && nextQuestion != nil {
		nextQuestion.Source = "follow_up"
		nextQuestion.RAGEligible = false

		// Insert follow-up question
		// 1. Save question to DB
		if err := s.questionRepo.Create(nextQuestion); err != nil {
			return nil, fmt.Errorf("failed to create follow-up question: %w", err)
		}

		// 2. Insert into interview_questions at current_index + 1
		// We need to shift existing questions if any (though usually we generate on the fly)
		// For simplicity in this architecture, we append it as the next question
		// and ensure the frontend fetches the updated list or next question.

		iq := &model.InterviewQuestion{
			InterviewID: interviewID,
			QuestionID:  nextQuestion.ID,
			OrderIndex:  interview.CurrentIndex + 1, // Next one
			IsAnswered:  false,
		}

		// We need to shift subsequent questions down?
		// Actually, if we are in a dynamic flow, we might not have pre-generated all questions.
		// If we did, we insert.
		if err := s.interviewRepo.InsertQuestionAt(interviewID, iq, interview.CurrentIndex+1); err != nil {
			return nil, fmt.Errorf("failed to insert follow-up: %w", err)
		}

		interview.FollowUpCount++
		interview.CurrentIndex++
	} else {
		// No follow-up, move to next topic or question
		interview.CurrentIndex++
		interview.FollowUpCount = 0 // Reset for next topic
		interview.TopicIndex++
	}

	allQuestions, _ := s.interviewRepo.GetInterviewQuestions(interviewID)

	// Update CurrentTopic based on the next question if available
	if interview.CurrentIndex < len(allQuestions) {
		nextQ, _ := s.questionRepo.GetByID(allQuestions[interview.CurrentIndex].QuestionID)
		if nextQ != nil {
			// If it's a new topic (not a follow-up), update topic
			// For simplicity, we assume standard questions start new topics
			// Follow-ups inherit topic or have specific prefix
			if interview.FollowUpCount == 0 {
				interview.CurrentTopic = nextQ.Category // Or derive from title
			}
		}
	}

	// Debug logging
	fmt.Printf("Interview %d: CurrentIndex %d, TotalQuestions %d\n", interviewID, interview.CurrentIndex, len(allQuestions))

	if interview.TotalQuestionTarget > 0 {
		if interview.CurrentIndex >= interview.TotalQuestionTarget {
			interview.Status = "completed"
			t := time.Now()
			interview.EndTime = &t
		}
	} else if interview.CurrentIndex >= len(allQuestions) {
		interview.Status = "completed"
		t := time.Now()
		interview.EndTime = &t
	}

	if err := s.interviewRepo.Update(interview); err != nil {
		return nil, fmt.Errorf("failed to update interview: %w", err)
	}

	return result, nil
}

// decideNextQuestion determines if a follow-up is needed and generates it
func (s *InterviewService) decideNextQuestion(interview *model.Interview, currentQ *model.Question, answer string, score int) (bool, *model.Question, error) {
	if interview == nil {
		return false, nil, nil
	}
	if score <= 0 {
		return false, nil, nil
	}
	// 1. Check constraints
	if interview.FollowUpCount >= interview.MaxFollowUps {
		return false, nil, nil
	}

	currentTopicCount := interview.FollowUpCount + 1
	if interview.TopicQuestionMax > 0 && currentTopicCount >= interview.TopicQuestionMax {
		return false, nil, nil
	}
	forceFollowUp := interview.TopicQuestionMin > 0 && currentTopicCount < interview.TopicQuestionMin

	if interview.TotalQuestionTarget > 0 {
		answeredSoFar := interview.CurrentIndex + 1
		remaining := interview.TotalQuestionTarget - answeredSoFar
		if remaining <= 0 {
			return false, nil, nil
		}

		remainingTopics := interview.TopicCountTarget - (interview.TopicIndex + 1)
		if remainingTopics < 0 {
			remainingTopics = 0
		}
		minNeededForCurrent := 0
		if interview.TopicQuestionMin > currentTopicCount {
			minNeededForCurrent = interview.TopicQuestionMin - currentTopicCount
		}
		requiredForFuture := remainingTopics * interview.TopicQuestionMin
		if remaining < requiredForFuture+minNeededForCurrent {
			// Not enough slots to satisfy minimums; avoid extra follow-ups.
			if !forceFollowUp {
				return false, nil, nil
			}
		}
	}
	if isLowSignalAnswer(answer) && !forceFollowUp {
		// Aggressive styles still force a clarifying follow-up on vague answers.
		switch interview.Style {
		case "stress", "deep", "algorithm", "practical":
			forceFollowUp = true
		default:
			return false, nil, nil
		}
	}

	// 2. Use AI to analyze answer and decide
	// We use RAG context if available to make the follow-up more specific
	ragContext := ""
	if s.ragService != nil {
		// Search RAG for concepts mentioned in the answer
		chunks, _ := s.ragService.SearchKnowledgeChunks(answer)
		if len(chunks) > 0 {
			ragContext = chunks[0].Content // Use top match
		}
	}

	nextQ, reason, err := s.aiService.GenerateFollowUpQuestion(interview, currentQ, answer, ragContext, interview.FollowUpCount)
	if err != nil {
		return false, nil, err
	}

	if nextQ == nil {
		if forceFollowUp {
			forced, err := s.aiService.GenerateClarifyingFollowUpQuestion(currentQ, answer, interview.FollowUpCount)
			if err != nil {
				return false, nil, nil
			}
			if isMeaninglessFollowUpQuestion(forced) {
				return false, nil, nil
			}
			return true, forced, nil
		}
		return false, nil, nil // AI decided not to follow up
	}

	if isMeaninglessFollowUpQuestion(nextQ) {
		return false, nil, nil
	}

	// Add metadata for frontend (e.g., reason for follow-up)
	// We could store this in a new field if needed.
	fmt.Printf("Generated Follow-Up: %s (Reason: %s)\n", nextQ.Title, reason)

	return true, nextQ, nil
}

func shouldEndEarlyForZeroScores(answers []model.AnswerResult, minCount int) bool {
	if len(answers) < minCount {
		return false
	}
	for _, a := range answers {
		if a.Score > 0 {
			return false
		}
	}
	return true
}

func buildInterviewPlan(difficulty string) (int, int) {
	switch difficulty {
	case "campus_intern":
		return 5, 11
	case "campus_graduate":
		return 4, 11
	case "social_junior":
		return 3, 11
	default:
		return 4, 11
	}
}

func buildStyleQuestionPlan(style string) (topicQuestionMin, topicQuestionMax, maxFollowUps int) {
	switch style {
	case "gentle":
		return 1, 2, 1
	case "stress":
		return 2, 5, 4
	case "deep":
		return 2, 4, 3
	case "practical":
		return 2, 3, 2
	case "algorithm":
		return 2, 4, 3
	default:
		return 2, 4, 3
	}
}

func isLowSignalAnswer(answer string) bool {
	text := strings.TrimSpace(answer)
	if text == "" {
		return true
	}
	if IsInvalidAnswer(text) {
		return true
	}
	if len([]rune(text)) < 8 {
		letterCount := 0
		digitCount := 0
		for _, r := range text {
			switch {
			case r >= '0' && r <= '9':
				digitCount++
			case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z'):
				letterCount++
			}
		}
		if digitCount+letterCount == len(text) {
			return true
		}
	}
	return false
}

func isMeaninglessFollowUpQuestion(q *model.Question) bool {
	if q == nil {
		return true
	}
	content := strings.TrimSpace(q.Content)
	title := strings.TrimSpace(q.Title)
	all := strings.TrimSpace(title + " " + content)
	if all == "" || IsInvalidAnswer(all) {
		return true
	}
	if len([]rune(content)) < 8 {
		return true
	}
	lower := strings.ToLower(all)
	meaninglessPatterns := []string{
		"继续说", "再说一下", "展开说说", "详细说说", "补充一下", "还有吗", "还有没有", "嗯", "啊", "哈",
	}
	for _, p := range meaninglessPatterns {
		if strings.Contains(lower, p) && len([]rune(content)) < 16 {
			return true
		}
	}
	return false
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

func (s *InterviewService) SaveInterviewRecording(userID, interviewID uint, recordingURL string) (*model.Interview, error) {
	interview, err := s.GetInterviewByID(userID, interviewID)
	if err != nil {
		return nil, err
	}
	interview.RecordingURL = recordingURL
	interview.RecordingStatus = "ready"
	if err := s.interviewRepo.Update(interview); err != nil {
		return nil, fmt.Errorf("failed to save recording url: %w", err)
	}
	return interview, nil
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
