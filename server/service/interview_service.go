package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"your-project/model"
	"your-project/repository"
)

type InterviewService struct {
	interviewRepo *repository.InterviewRepository
	questionRepo  *repository.QuestionRepository
	userRepo      *repository.UserRepository
	aiService     *AIService
	ragService    *RAGService // Add RAG service
}

func NewInterviewService() *InterviewService {
	return &InterviewService{
		interviewRepo: repository.NewInterviewRepository(),
		questionRepo:  repository.NewQuestionRepository(),
		userRepo:      repository.NewUserRepository(),
		aiService:     MustGetAIService(),
		ragService:    GetRAGService(), // Init RAG service
	}
}

func normalizeInterviewPosition(position string) string {
	p := strings.ToLower(strings.TrimSpace(position))
	switch {
	case p == "", strings.Contains(p, "java"), strings.Contains(p, "后端"), p == "backend":
		return "Java后端工程师"
	case strings.Contains(p, "前端"), strings.Contains(p, "frontend"):
		return "前端工程师"
	case strings.Contains(p, "算法"), p == "algorithm":
		return "算法工程师"
	case strings.Contains(p, "ai"), strings.Contains(p, "llm"), strings.Contains(p, "模型"), p == "ai_engineer":
		return "AI工程师"
	default:
		return "Java后端工程师"
	}
}

// StartInterview now accepts mode, style, company, and interviewMode. It uses AI to generate questions based on these parameters.
func (s *InterviewService) StartInterview(userID uint, position, difficulty, mode, style, company, interviewMode string, invitationID *uint) (*model.Interview, error) {
	ctx := context.Background()
	var questions []*model.Question
	var scenarioJSON string
	var revealedStyle string
	var capabilityGraph *model.JobCapabilityDimension
	var invitation *model.HumanInterviewInvitation

	topicCount, totalTarget := buildInterviewPlan(difficulty)
	topicQuestionMin, topicQuestionMax, maxFollowUps := 2, 4, 3
	position = normalizeInterviewPosition(position)

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

	if interviewMode == "human" {
		if invitationID == nil || *invitationID == 0 {
			return nil, fmt.Errorf("请选择已邀请的真人面试官")
		}
		loaded, err := s.interviewRepo.GetInvitationByID(*invitationID)
		if err != nil {
			return nil, fmt.Errorf("邀请记录不存在")
		}
		if loaded.StudentID != userID {
			return nil, fmt.Errorf("无权使用该邀请")
		}
		if loaded.Status == "cancelled" {
			return nil, fmt.Errorf("该邀请已取消")
		}
		if loaded.Status == "rejected" {
			return nil, fmt.Errorf("该邀请已被对方拒绝")
		}
		if loaded.Status != "accepted" && loaded.Status != "in_progress" {
			return nil, fmt.Errorf("请等待对方接受邀请后再开始真人面试")
		}
		invitation = loaded
		if invitation.Position != "" {
			position = normalizeInterviewPosition(invitation.Position)
		}
		if invitation.Difficulty != "" {
			difficulty = invitation.Difficulty
		}
		if invitation.Mode != "" {
			mode = invitation.Mode
		}
		if invitation.Style != "" {
			style = invitation.Style
		}
		if invitation.Company != "" {
			company = invitation.Company
		}
	}

	// Blindbox mode: draw a random scenario and generate tailored questions
	if mode == "blindbox" {
		bbService := NewBlindBoxService()
		scenario := bbService.DrawScenario()
		scenarioJSON = ScenarioToJSON(scenario)

		// Override style with scenario's style
		style = scenario.Style

		// Generate questions tailored to this scenario
		generated, err := bbService.GenerateBlindBoxQuestions(ctx, scenario, position, difficulty, topicCount)
		if err != nil {
			return nil, fmt.Errorf("blindbox question generation failed: %w", err)
		}
		for _, q := range generated {
			q.Position = position
			q.Difficulty = difficulty
			q.Source = "ai_opening"
			q.RAGEligible = true
			s.normalizeOpeningQuestion(ctx, q)
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
					q, qErr := s.aiService.GenerateTopicQuestionFromContext(ctx, dummyInterview, chunk.Content, chunk.Category)
					if qErr == nil && q != nil {
						q.Source = "ai_opening"
						q.RAGEligible = true
						if normalized := s.normalizeOpeningQuestion(ctx, q); normalized != nil {
							if err := s.questionRepo.Create(normalized); err != nil {
								continue
							}
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
				if s.aiService.IsContextDependentOpeningQuestion(q) {
					s.quarantineQuestionAsFollowUp(q)
					continue
				}
				if normalized := s.normalizeOpeningQuestion(ctx, q); normalized != nil {
					questions = append(questions, normalized)
				}
			}
		}

		// Final fallback: generate topic questions if still short
		if len(questions) < topicCount {
			needed := topicCount - len(questions)
			maxAttempts := needed * 3
			for i := 0; i < maxAttempts && len(questions) < topicCount; i++ {
				q, err := s.aiService.GenerateNextQuestionWithWeights(ctx, dummyInterview, nil, capabilityGraph)
				if err != nil {
					q, err = s.aiService.GenerateNextQuestion(ctx, dummyInterview, nil)
					if err != nil {
						continue
					}
				}
				q.Position = position
				q.Difficulty = difficulty
				q.Source = "ai_opening"
				q.RAGEligible = true
				if normalized := s.normalizeOpeningQuestion(ctx, q); normalized == nil {
					continue
				}
				if err := s.questionRepo.Create(q); err == nil {
					questions = append(questions, q)
				}
			}
		}
	}

	questions = s.prioritizeTopicVariety(questions, topicCount)

	if len(questions) < topicCount {
		dummyInterview := &model.Interview{
			Position:   position,
			Difficulty: difficulty,
			Mode:       mode,
			Style:      style,
			Company:    company,
		}
		maxAttempts := (topicCount - len(questions)) * 4
		for i := 0; i < maxAttempts && len(questions) < topicCount; i++ {
			q, genErr := s.aiService.GenerateNextQuestionWithWeights(ctx, dummyInterview, nil, capabilityGraph)
			if genErr != nil {
				q, genErr = s.aiService.GenerateNextQuestion(ctx, dummyInterview, nil)
				if genErr != nil {
					continue
				}
			}
			q.Position = position
			q.Difficulty = difficulty
			q.Source = "ai_opening"
			q.RAGEligible = true
			normalized := s.normalizeOpeningQuestion(ctx, q)
			if normalized == nil {
				continue
			}
			if s.isDuplicateOpeningQuestion(questions, normalized) {
				continue
			}
			if err := s.questionRepo.Create(normalized); err != nil {
				continue
			}
			questions = append(questions, normalized)
		}
	}

	if len(questions) > topicCount {
		questions = questions[:topicCount]
	}

	for _, q := range questions {
		s.normalizeOpeningQuestion(ctx, q)
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

	if invitation != nil {
		interview.HumanInterviewerUserID = &invitation.InviteeUserID
		interview.HumanInterviewerName = invitation.Invitee.Username
		interview.HumanInterviewerRole = invitation.InviteeRole
	}

	if err := s.interviewRepo.Create(interview); err != nil {
		return nil, fmt.Errorf("failed to create interview: %w", err)
	}

	if invitation != nil {
		invitation.Status = "in_progress"
		invitation.InterviewID = &interview.ID
		if err := s.interviewRepo.UpdateInvitation(invitation); err != nil {
			return nil, fmt.Errorf("failed to update invitation status: %w", err)
		}
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

func (s *InterviewService) normalizeOpeningQuestion(ctx context.Context, q *model.Question) *model.Question {
	if q == nil {
		return nil
	}
	if strings.TrimSpace(q.Source) == "" {
		q.Source = "standard"
	}
	if q.Source == "follow_up" {
		return nil
	}

	s.aiService.EnsureQuestionChinese(ctx, q)
	if s.aiService.IsContextDependentOpeningQuestion(q) {
		s.aiService.NormalizeToSelfContainedOpening(q)
	}
	return q
}

func (s *InterviewService) quarantineQuestionAsFollowUp(q *model.Question) {
	if q == nil || q.ID == 0 {
		return
	}
	if q.Source == "follow_up" && !q.RAGEligible {
		return
	}
	q.Source = "follow_up"
	q.RAGEligible = false
	if err := s.questionRepo.Update(q); err != nil {
		fmt.Printf("failed to quarantine question %d: %v\n", q.ID, err)
	}
}

// Package-level wrapper
func StartInterview(userID uint, position, difficulty, mode, style, company, interviewMode string, invitationID *uint) (*model.Interview, error) {
	svc := NewInterviewService()
	return svc.StartInterview(userID, position, difficulty, mode, style, company, interviewMode, invitationID)
}

func GetInterviewByID(userID, interviewID uint) (*model.Interview, error) {
	svc := NewInterviewService()
	return svc.GetInterviewByID(userID, interviewID)
}

func SubmitAnswer(userID, interviewID, questionID uint, answer, audioData, audioMime, questionTitle, questionContent string) (*model.AnswerResult, error) {
	svc := NewInterviewService()
	return svc.SubmitAnswer(userID, interviewID, questionID, answer, audioData, audioMime, questionTitle, questionContent)
}

func EndInterview(userID, interviewID uint) (*model.Interview, error) {
	svc := NewInterviewService()
	return svc.EndInterview(userID, interviewID)
}

func ListInviteCandidates(role, keyword string, page, pageSize int) ([]model.User, int64, error) {
	svc := NewInterviewService()
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return svc.userRepo.ListInviteCandidates(role, keyword, page, pageSize)
}

func CreateHumanInvitation(studentID, inviteeUserID uint, scheduledAt *time.Time, position, difficulty, mode, style, company, notes string) (*model.HumanInterviewInvitation, error) {
	svc := NewInterviewService()
	student, err := svc.userRepo.GetByID(studentID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}
	if student.Role != "student" {
		return nil, fmt.Errorf("仅学生用户可以发起真人面试邀请")
	}
	invitee, err := svc.userRepo.GetByID(inviteeUserID)
	if err != nil {
		return nil, fmt.Errorf("被邀请用户不存在")
	}
	if invitee.Role != "enterprise" && invitee.Role != "university" {
		return nil, fmt.Errorf("只能邀请学校端或企业端用户")
	}

	inv := &model.HumanInterviewInvitation{
		StudentID:     studentID,
		InviteeUserID: inviteeUserID,
		InviteeRole:   invitee.Role,
		Position:      strings.TrimSpace(position),
		Difficulty:    strings.TrimSpace(difficulty),
		Mode:          strings.TrimSpace(mode),
		Style:         strings.TrimSpace(style),
		Company:       strings.TrimSpace(company),
		Status:        "pending",
		ScheduledAt:   scheduledAt,
		Notes:         strings.TrimSpace(notes),
	}

	if err := svc.interviewRepo.CreateInvitation(inv); err != nil {
		return nil, fmt.Errorf("创建邀请失败: %w", err)
	}

	inv.Invitee = *invitee
	return inv, nil
}

func ListHumanInvitations(studentID uint) ([]model.HumanInterviewInvitation, error) {
	svc := NewInterviewService()
	return svc.interviewRepo.GetInvitationsByStudentID(studentID)
}

func ListReceivedHumanInvitations(inviteeUserID uint) ([]model.HumanInterviewInvitation, error) {
	svc := NewInterviewService()
	user, err := svc.userRepo.GetByID(inviteeUserID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}
	if user.Role != "enterprise" && user.Role != "university" {
		return nil, fmt.Errorf("仅企业端或学校端可以查看收到的邀请")
	}
	return svc.interviewRepo.GetInvitationsByInviteeID(inviteeUserID)
}

func GetInvitationByID(invitationID uint) (*model.HumanInterviewInvitation, error) {
	svc := NewInterviewService()
	return svc.interviewRepo.GetInvitationByID(invitationID)
}

func RespondHumanInvitation(inviteeUserID, invitationID uint, action string) (*model.HumanInterviewInvitation, error) {
	svc := NewInterviewService()
	user, err := svc.userRepo.GetByID(inviteeUserID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}
	if user.Role != "enterprise" && user.Role != "university" {
		return nil, fmt.Errorf("仅企业端或学校端可以处理邀请")
	}

	invitation, err := svc.interviewRepo.GetInvitationByIDForInvitee(invitationID, inviteeUserID)
	if err != nil {
		return nil, fmt.Errorf("邀请不存在")
	}

	normalizedAction := strings.ToLower(strings.TrimSpace(action))
	if normalizedAction != "accept" && normalizedAction != "reject" {
		return nil, fmt.Errorf("action 仅支持 accept 或 reject")
	}

	if invitation.Status != "pending" {
		return nil, fmt.Errorf("当前邀请状态为 %s，无法重复处理", invitation.Status)
	}

	if normalizedAction == "accept" {
		invitation.Status = "accepted"
	} else {
		invitation.Status = "rejected"
	}

	if err := svc.interviewRepo.UpdateInvitation(invitation); err != nil {
		return nil, fmt.Errorf("更新邀请状态失败: %w", err)
	}

	return invitation, nil
}

func GenerateShadowHint(userID, interviewID uint, question, transcript string, silenceSeconds int) (string, error) {
	svc := NewInterviewService()
	return svc.GenerateShadowHint(userID, interviewID, question, transcript, silenceSeconds)
}

func GenerateShadowHintPack(userID, interviewID uint, question, transcript, expectedAnswer string) ([]string, error) {
	svc := NewInterviewService()
	return svc.GenerateShadowHintPack(userID, interviewID, question, transcript, expectedAnswer)
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

func (s *InterviewService) SubmitAnswer(userID, interviewID, questionID uint, answer, audioData, audioMime, questionTitle, questionContent string) (*model.AnswerResult, error) {
	_ = audioMime

	ctx := context.Background()

	interview, err := s.GetInterviewByID(userID, interviewID)
	if err != nil {
		return nil, err
	}

	if interview.Status != "in_progress" {
		return nil, fmt.Errorf("interview is not in progress")
	}

	baseQuestion, err := s.questionRepo.GetByID(questionID)
	if err != nil {
		return nil, fmt.Errorf("question not found")
	}

	evalQuestion := baseQuestion
	if strings.TrimSpace(questionContent) != "" {
		tempQ := *baseQuestion
		tempQ.Content = strings.TrimSpace(questionContent)
		if strings.TrimSpace(questionTitle) != "" {
			tempQ.Title = strings.TrimSpace(questionTitle)
		}
		evalQuestion = &tempQ
	}

	var finalAnswer string
	if audioData != "" {
		transcribedText, err := s.aiService.TranscribeAudio(strings.TrimSpace(audioData))
		if err != nil {
			return nil, fmt.Errorf("failed to transcribe audio: %w", err)
		}
		finalAnswer = transcribedText
		interview.ASRCallCount++
	} else {
		finalAnswer = answer
	}

	evaluation, err := s.aiService.EvaluateAnswer(ctx, evalQuestion, finalAnswer)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate answer: %w", err)
	}

	result := &model.AnswerResult{
		InterviewID: interviewID,
		QuestionID:  baseQuestion.ID,
		Answer:      finalAnswer,
		Score:       evaluation.Score,
		Feedback:    evaluation.Feedback,
		CreatedAt:   time.Now(),
	}

	if err := s.interviewRepo.SaveAnswer(result); err != nil {
		return nil, fmt.Errorf("failed to save answer: %w", err)
	}

	answers, _ := s.interviewRepo.GetAnswersByInterviewID(interviewID)

	// ==== Dynamic Follow-Up Logic ====
	// Follow-up questions are generated in real time and kept ephemeral (no DB write).
	askedQuestionText := strings.TrimSpace(evalQuestion.Content)
	if askedQuestionText == "" {
		askedQuestionText = strings.TrimSpace(evalQuestion.Title)
	}
	shouldFollowUp, nextQuestion, err := s.decideNextQuestion(ctx, interview, baseQuestion, askedQuestionText, finalAnswer, evaluation.Score)
	if err != nil {
		fmt.Printf("Dynamic question generation failed: %v\n", err)
	}

	if shouldFollowUp && nextQuestion != nil {
		nextQuestion.Source = "follow_up"
		nextQuestion.RAGEligible = false
		nextQuestion.ID = baseQuestion.ID
		nextQuestion.Position = baseQuestion.Position
		nextQuestion.Difficulty = baseQuestion.Difficulty
		nextQuestion.Category = baseQuestion.Category
		result.NextQuestion = nextQuestion
		interview.FollowUpCount++
	} else {
		interview.CurrentIndex++
		interview.FollowUpCount = 0
		interview.TopicIndex++

		allQuestions, _ := s.interviewRepo.GetInterviewQuestions(interviewID)
		if interview.CurrentIndex < len(allQuestions) {
			nextQ, _ := s.questionRepo.GetByID(allQuestions[interview.CurrentIndex].QuestionID)
			if nextQ != nil {
				s.normalizeOpeningQuestion(ctx, nextQ)
				interview.CurrentTopic = nextQ.Category
				result.NextQuestion = nextQ
			}
		}
	}

	allQuestions, _ := s.interviewRepo.GetInterviewQuestions(interviewID)
	if interview.TotalQuestionTarget > 0 && len(answers) >= interview.TotalQuestionTarget {
		interview.Status = "completed"
		t := time.Now()
		interview.EndTime = &t
		result.InterviewCompleted = true
	} else if interview.CurrentIndex >= len(allQuestions) && result.NextQuestion == nil {
		interview.Status = "completed"
		t := time.Now()
		interview.EndTime = &t
		result.InterviewCompleted = true
	}

	if err := s.interviewRepo.Update(interview); err != nil {
		return nil, fmt.Errorf("failed to update interview: %w", err)
	}

	return result, nil
}

// decideNextQuestion determines if a follow-up is needed and generates it
func (s *InterviewService) decideNextQuestion(ctx context.Context, interview *model.Interview, topicRootQ *model.Question, askedQuestionText, answer string, score int) (bool, *model.Question, error) {
	if interview == nil {
		return false, nil, nil
	}
	if score <= 0 {
		// Hard stop current topic when answer quality is 0.
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

	nextQ, reason, err := s.aiService.GenerateFollowUpQuestion(ctx, interview, topicRootQ, answer, ragContext, interview.FollowUpCount)
	if err != nil {
		return false, nil, err
	}

	if nextQ == nil {
		if forceFollowUp {
			forced, err := s.aiService.GenerateClarifyingFollowUpQuestion(ctx, topicRootQ, answer, interview.FollowUpCount)
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
	if s.isDuplicateQuestionInSession(interview.ID, askedQuestionText, nextQ) {
		if forceFollowUp {
			forced, forceErr := s.aiService.GenerateClarifyingFollowUpQuestion(ctx, topicRootQ, answer, interview.FollowUpCount)
			if forceErr != nil || isMeaninglessFollowUpQuestion(forced) || s.isDuplicateQuestionInSession(interview.ID, askedQuestionText, forced) {
				return false, nil, nil
			}
			return true, forced, nil
		}
		return false, nil, nil
	}

	// Add metadata for frontend (e.g., reason for follow-up)
	// We could store this in a new field if needed.
	fmt.Printf("Generated Follow-Up: %s (Reason: %s)\n", nextQ.Title, reason)

	return true, nextQ, nil
}

func (s *InterviewService) isDuplicateQuestionInSession(interviewID uint, askedQuestionText string, candidate *model.Question) bool {
	if candidate == nil {
		return true
	}
	if isNearDuplicateQuestionText(askedQuestionText, candidate) {
		return true
	}

	candidateKey := normalizeQuestionKey(candidate.Title, candidate.Content)
	if candidateKey == "" {
		return true
	}

	answers, err := s.interviewRepo.GetAnswersByInterviewID(interviewID)
	if err != nil {
		return false
	}

	for _, ans := range answers {
		text := strings.TrimSpace(ans.Question.Content)
		if text == "" {
			text = strings.TrimSpace(ans.Question.Title)
		}
		if text == "" && ans.QuestionID != 0 {
			storedQ, qErr := s.questionRepo.GetByID(ans.QuestionID)
			if qErr == nil && storedQ != nil {
				text = strings.TrimSpace(storedQ.Content)
				if text == "" {
					text = strings.TrimSpace(storedQ.Title)
				}
			}
		}
		if text == "" {
			continue
		}

		askedKey := normalizeQuestionKey("", text)
		if askedKey == "" {
			continue
		}
		if askedKey == candidateKey || strings.Contains(askedKey, candidateKey) || strings.Contains(candidateKey, askedKey) {
			return true
		}
	}

	return false
}

func buildInterviewPlan(difficulty string) (int, int) {
	var topicCount int
	var totalTarget int
	switch difficulty {
	case "campus_intern":
		topicCount, totalTarget = 5, 15
	case "campus_graduate":
		topicCount, totalTarget = 4, 12
	case "social_junior":
		topicCount, totalTarget = 3, 9
	default:
		topicCount, totalTarget = 4, 12
	}
	if topicCount > 5 {
		topicCount = 5
	}
	return topicCount, totalTarget
}

func buildStyleQuestionPlan(style string) (topicQuestionMin, topicQuestionMax, maxFollowUps int) {
	switch style {
	case "gentle":
		return 3, 4, 3
	case "stress":
		return 3, 4, 3
	case "deep":
		return 3, 4, 3
	case "practical":
		return 3, 4, 3
	case "algorithm":
		return 3, 4, 3
	default:
		return 3, 4, 3
	}
}

func normalizeQuestionKey(title, content string) string {
	joined := strings.ToLower(strings.TrimSpace(title + " " + content))
	joined = strings.ReplaceAll(joined, " ", "")
	joined = strings.ReplaceAll(joined, "\n", "")
	joined = strings.ReplaceAll(joined, "\t", "")
	joined = strings.ReplaceAll(joined, "，", "")
	joined = strings.ReplaceAll(joined, "。", "")
	joined = strings.ReplaceAll(joined, "？", "")
	joined = strings.ReplaceAll(joined, "!", "")
	joined = strings.ReplaceAll(joined, "?", "")
	joined = strings.ReplaceAll(joined, ":", "")
	joined = strings.ReplaceAll(joined, "：", "")
	return joined
}

func (s *InterviewService) isDuplicateOpeningQuestion(existing []*model.Question, candidate *model.Question) bool {
	if candidate == nil {
		return true
	}
	key := normalizeQuestionKey(candidate.Title, candidate.Content)
	if key == "" {
		return true
	}
	for _, q := range existing {
		if q == nil {
			continue
		}
		if normalizeQuestionKey(q.Title, q.Content) == key {
			return true
		}
	}
	return false
}

func (s *InterviewService) prioritizeTopicVariety(questions []*model.Question, target int) []*model.Question {
	if len(questions) == 0 || target <= 0 {
		return questions
	}

	unique := make([]*model.Question, 0, len(questions))
	seenKey := map[string]bool{}
	for _, q := range questions {
		if q == nil {
			continue
		}
		key := normalizeQuestionKey(q.Title, q.Content)
		if key == "" || seenKey[key] {
			continue
		}
		seenKey[key] = true
		unique = append(unique, q)
	}

	if len(unique) <= 1 {
		if len(unique) > target {
			return unique[:target]
		}
		return unique
	}

	result := make([]*model.Question, 0, len(unique))
	usedCategory := map[string]bool{}

	for _, q := range unique {
		cat := strings.TrimSpace(q.Category)
		if cat == "" || usedCategory[cat] {
			continue
		}
		usedCategory[cat] = true
		result = append(result, q)
		if len(result) >= target {
			return result
		}
	}

	for _, q := range unique {
		if len(result) >= target {
			break
		}
		already := false
		for _, picked := range result {
			if picked == q {
				already = true
				break
			}
		}
		if !already {
			result = append(result, q)
		}
	}

	return result
}

func isNearDuplicateQuestionText(lastAsked string, candidate *model.Question) bool {
	if candidate == nil {
		return true
	}
	left := normalizeQuestionKey("", lastAsked)
	right := normalizeQuestionKey(candidate.Title, candidate.Content)
	if left == "" || right == "" {
		return false
	}
	if left == right {
		return true
	}
	if strings.Contains(left, right) || strings.Contains(right, left) {
		return true
	}
	return false
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

	if interview.InterviewMode == "human" {
		invitations, err := s.interviewRepo.GetInvitationsByStudentID(userID)
		if err == nil {
			for i := range invitations {
				inv := invitations[i]
				if inv.InterviewID != nil && *inv.InterviewID == interviewID && inv.Status == "in_progress" {
					inv.Status = "completed"
					_ = s.interviewRepo.UpdateInvitation(&inv)
					break
				}
			}
		}
	}

	return interview, nil
}

func (s *InterviewService) GenerateShadowHint(userID, interviewID uint, question, transcript string, silenceSeconds int) (string, error) {
	interview, err := s.GetInterviewByID(userID, interviewID)
	if err != nil {
		return "", err
	}
	if interview.Status != "in_progress" {
		return "", fmt.Errorf("interview is not in progress")
	}

	ctx := context.Background()

	hint, err := s.aiService.GenerateShadowCoachHint(
		ctx,
		interview.Position,
		question,
		transcript,
		interview.Style,
		silenceSeconds,
	)
	if err != nil {
		return "", err
	}

	return hint, nil
}

func (s *InterviewService) GenerateShadowHintPack(userID, interviewID uint, question, transcript, expectedAnswer string) ([]string, error) {
	interview, err := s.GetInterviewByID(userID, interviewID)
	if err != nil {
		return nil, err
	}
	if interview.Status != "in_progress" {
		return nil, fmt.Errorf("interview is not in progress")
	}

	ctx := context.Background()

	knowledgeContext := ""
	if s.ragService != nil {
		queryParts := make([]string, 0, 3)
		if q := strings.TrimSpace(question); q != "" {
			queryParts = append(queryParts, q)
		}
		if ea := strings.TrimSpace(expectedAnswer); ea != "" {
			runes := []rune(ea)
			if len(runes) > 120 {
				ea = string(runes[:120])
			}
			queryParts = append(queryParts, ea)
		}
		if tr := strings.TrimSpace(transcript); tr != "" {
			runes := []rune(tr)
			if len(runes) > 80 {
				tr = string(runes[:80])
			}
			queryParts = append(queryParts, tr)
		}
		query := strings.TrimSpace(strings.Join(queryParts, "\n"))
		if query != "" {
			chunks, ragErr := s.ragService.SearchKnowledgeChunksWithLimit(query, 5)
			if ragErr == nil && len(chunks) > 0 {
				parts := make([]string, 0, len(chunks))
				for _, chunk := range chunks {
					text := strings.TrimSpace(chunk.Content)
					if text == "" {
						continue
					}
					runes := []rune(text)
					if len(runes) > 260 {
						text = string(runes[:260])
					}
					parts = append(parts, text)
				}
				knowledgeContext = strings.Join(parts, "\n---\n")
			}
		}
	}

	hints, err := s.aiService.GenerateShadowCoachHintLevels(
		ctx,
		interview.Position,
		question,
		transcript,
		interview.Style,
		expectedAnswer,
		knowledgeContext,
	)
	if err != nil {
		return nil, err
	}

	return hints, nil
}

// loadJobCapabilityGraph simulates loading job capability graph
// In production, this would fetch from DB or RAG
func (s *InterviewService) loadJobCapabilityGraph(position string) *model.JobCapabilityDimension {
	position = normalizeInterviewPosition(position)
	switch position {
	case "Java后端工程师":
		return &model.JobCapabilityDimension{
			Name:   "Java后端工程师",
			Weight: 100,
			SubDimensions: []model.JobCapabilitySubDimension{
				{Name: "JVM原理", Weight: 30, Tags: []string{"GC", "Memory Model", "Classloader"}},
				{Name: "分布式系统", Weight: 25, Tags: []string{"CAP", "Microservices", "RPC"}},
				{Name: "数据库优化", Weight: 20, Tags: []string{"Index", "SQL Tuning", "Locking"}},
				{Name: "网络协议", Weight: 15, Tags: []string{"TCP/IP", "HTTP", "Socket"}},
				{Name: "并发编程", Weight: 10, Tags: []string{"Go Routine", "Channel", "Sync"}},
			},
		}
	case "前端工程师":
		return &model.JobCapabilityDimension{
			Name:   "前端工程师",
			Weight: 100,
			SubDimensions: []model.JobCapabilitySubDimension{
				{Name: "React/Vue框架深度", Weight: 30, Tags: []string{"Virtual DOM", "Hooks", "Reactivity"}},
				{Name: "JavaScript核心", Weight: 25, Tags: []string{"ES6+", "Closure", "Prototype"}},
				{Name: "工程化", Weight: 20, Tags: []string{"Webpack", "Vite", "CI/CD"}},
				{Name: "性能优化", Weight: 15, Tags: []string{"Rendering", "Network", "Cache"}},
				{Name: "CSS/布局", Weight: 10, Tags: []string{"Flexbox", "Grid", "Animation"}},
			},
		}
	case "算法工程师":
		return &model.JobCapabilityDimension{
			Name:   "算法工程师",
			Weight: 100,
			SubDimensions: []model.JobCapabilitySubDimension{
				{Name: "数据结构与算法", Weight: 35, Tags: []string{"DP", "Graph", "Complexity"}},
				{Name: "机器学习基础", Weight: 30, Tags: []string{"Supervised Learning", "Evaluation", "Feature Engineering"}},
				{Name: "数学基础", Weight: 20, Tags: []string{"Probability", "Linear Algebra", "Optimization"}},
				{Name: "工程实现", Weight: 15, Tags: []string{"Python", "Model Serving", "Debugging"}},
			},
		}
	case "AI工程师":
		return &model.JobCapabilityDimension{
			Name:   "AI工程师",
			Weight: 100,
			SubDimensions: []model.JobCapabilitySubDimension{
				{Name: "大模型基础", Weight: 30, Tags: []string{"Transformer", "Prompt", "Fine-tuning"}},
				{Name: "RAG与Agent", Weight: 30, Tags: []string{"RAG", "Tool Calling", "Workflow"}},
				{Name: "多模态能力", Weight: 20, Tags: []string{"Vision", "Audio", "Fusion"}},
				{Name: "部署与优化", Weight: 20, Tags: []string{"Inference", "Latency", "Cost"}},
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
