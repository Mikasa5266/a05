package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"

	"golang.org/x/sync/errgroup"

	"your-project/model"
	"your-project/repository"
)

const (
	fastInitAIFillMaxAttempts     = 1
	fastInitAIFillConcurrency     = 1
	fastInitAITaskTimeout         = 1200 * time.Millisecond
	fastInitChineseRewriteTimeout = 1200 * time.Millisecond
	fastInitOpeningRewriteBudget  = 2
)

var openingQuestionEnglishTokenPattern = regexp.MustCompile(`[A-Za-z]{8,}`)

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
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	return s.StartInterviewWithContext(ctx, userID, position, difficulty, mode, style, company, interviewMode, invitationID)
}

func (s *InterviewService) StartInterviewWithContext(ctx context.Context, userID uint, position, difficulty, mode, style, company, interviewMode string, invitationID *uint) (*model.Interview, error) {
	var questions []*model.Question
	var scenarioJSON string
	var revealedStyle string
	var capabilityGraph *model.JobCapabilityDimension
	var invitation *model.HumanInterviewInvitation

	topicCount, totalTarget := buildInterviewPlan(difficulty)
	openingQuestionChineseRewriteBudget := int32(fastInitOpeningRewriteBudget)
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

	// 首题硬约束：优先从题库按岗位+级别随机直取，绕过 LLM 生成。
	firstQuestion, firstErr := s.questionRepo.GetRandomQuestionForInterview(position, difficulty, nil)
	if firstErr == nil && firstQuestion != nil {
		if s.aiService.IsContextDependentOpeningQuestion(firstQuestion) {
			s.quarantineQuestionAsFollowUp(firstQuestion)
		} else if normalized := s.normalizeOpeningQuestionForInit(ctx, firstQuestion, &openingQuestionChineseRewriteBudget); normalized != nil {
			questions = append(questions, normalized)
		}
	}

	isAlgorithmStyle := style == "algorithm" && mode != "blindbox"
	if isAlgorithmStyle {
		// 算法开场题固定 3 题，避免沿用难度档位导致额外补题并触发超时。
		topicCount = 3
		seededAlgorithmQuestions := s.buildAlgorithmOpeningQuestions(position, difficulty)
		for _, q := range seededAlgorithmQuestions {
			if len(questions) >= topicCount {
				break
			}
			normalized := s.normalizeOpeningQuestionForInit(ctx, q, &openingQuestionChineseRewriteBudget)
			if normalized == nil || s.isDuplicateOpeningQuestion(questions, normalized) {
				continue
			}
			questions = append(questions, normalized)
		}

		// 预设题不足时，优先从本地算法题库补齐，避免触发耗时 AI 生成。
		if len(questions) < topicCount {
			localAlgorithmBank, err := s.questionRepo.GetQuestionsForInterviewInitWithExclude(position, difficulty, "algorithm", topicCount*4, collectQuestionIDs(questions))
			if err == nil {
				for _, q := range localAlgorithmBank {
					if len(questions) >= topicCount {
						break
					}
					normalized := s.normalizeOpeningQuestionForInit(ctx, q, &openingQuestionChineseRewriteBudget)
					if normalized == nil || s.isDuplicateOpeningQuestion(questions, normalized) {
						continue
					}
					questions = append(questions, normalized)
				}
			}
		}

		// 若算法分类题仍不足，再使用本地通用题库兜底，不走 AI 强制补题。
		if len(questions) < topicCount {
			localFallback, err := s.questionRepo.GetQuestionsForInterviewInitWithExclude(position, difficulty, "", topicCount*4, collectQuestionIDs(questions))
			if err == nil {
				for _, q := range localFallback {
					if len(questions) >= topicCount {
						break
					}
					normalized := s.normalizeOpeningQuestionForInit(ctx, q, &openingQuestionChineseRewriteBudget)
					if normalized == nil || s.isDuplicateOpeningQuestion(questions, normalized) {
						continue
					}
					questions = append(questions, normalized)
				}
			}
		}
	} else if mode == "blindbox" {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("面试初始化超时，请稍后重试")
		}
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
			s.normalizeOpeningQuestionForInit(ctx, q, &openingQuestionChineseRewriteBudget)
			if err := s.questionRepo.Create(q); err != nil {
				return nil, fmt.Errorf("failed to save blindbox question: %w", err)
			}
			questions = append(questions, q)
		}
	} else {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("面试初始化超时，请稍后重试")
		}

		// 速度优先：启动阶段不做 RAG->AI 开场题生成，优先题库与本地兜底，确保 10 秒内可进入。
		dummyInterview := &model.Interview{
			Position:   position,
			Difficulty: difficulty,
			Mode:       mode,
			Style:      style,
			Company:    company,
		}

		// Fill from question bank if needed
		if len(questions) < topicCount {
			fallback, err := s.questionRepo.GetQuestionsForInterviewInitWithExclude(position, difficulty, "", topicCount*4, collectQuestionIDs(questions))
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
				if normalized := s.normalizeOpeningQuestionForInit(ctx, q, &openingQuestionChineseRewriteBudget); normalized != nil {
					questions = append(questions, normalized)
				}
			}
		}

		// 先走本地模板兜底，保证优先可进入。
		if len(questions) < topicCount {
			questions = s.fillWithLocalFallbackQuestions(ctx, questions, topicCount, position, difficulty, mode, style, &openingQuestionChineseRewriteBudget)
		}

		// 本地仍不足时再做极少量 AI 补题，避免启动阶段长时间阻塞。
		if len(questions) < topicCount {
			needed := topicCount - len(questions)
			maxAttempts := needed
			if maxAttempts > fastInitAIFillMaxAttempts {
				maxAttempts = fastInitAIFillMaxAttempts
			}
			questions = s.fillOpeningQuestionsConcurrently(ctx, questions, topicCount, dummyInterview, capabilityGraph, position, difficulty, maxAttempts, &openingQuestionChineseRewriteBudget)
		}
	}

	questions = s.prioritizeTopicVariety(questions, topicCount)

	if len(questions) < topicCount && !isAlgorithmStyle {
		questions = s.fillWithLocalFallbackQuestions(ctx, questions, topicCount, position, difficulty, mode, style, &openingQuestionChineseRewriteBudget)
		questions = s.prioritizeTopicVariety(questions, topicCount)
	}

	// 最后兜底：即使 AI 和题库都不足，也保证至少能拿到可用开场题，避免前端无法进入面试。
	if len(questions) < topicCount {
		questions = s.fillWithLocalFallbackQuestions(ctx, questions, topicCount, position, difficulty, mode, style, &openingQuestionChineseRewriteBudget)
		questions = s.prioritizeTopicVariety(questions, topicCount)
	}

	if len(questions) == 0 {
		return nil, fmt.Errorf("面试初始化失败：未获取到可用题目，请稍后重试")
	}

	if len(questions) > topicCount {
		questions = questions[:topicCount]
	}

	for _, q := range questions {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("面试初始化超时，请稍后重试")
		}
		s.normalizeOpeningQuestionForInit(ctx, q, &openingQuestionChineseRewriteBudget)
	}

	initialAskedQuestionIDs := []uint{}
	if len(questions) > 0 && questions[0] != nil && questions[0].ID > 0 {
		initialAskedQuestionIDs = append(initialAskedQuestionIDs, questions[0].ID)
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
		Role:                "candidate",
		Status:              "in_progress",
		StartTime:           time.Now(),
		CurrentIndex:        0,
		AskedQuestionIDs:    encodeAskedQuestionIDs(initialAskedQuestionIDs),
		MaxFollowUps:        maxFollowUps,
		TopicIndex:          0,
		TopicCountTarget:    topicCount,
		TopicQuestionMin:    topicQuestionMin,
		TopicQuestionMax:    topicQuestionMax,
		TotalQuestionTarget: totalTarget,
	}

	if invitation != nil {
		interview.Status = "pending"
		interview.HumanInterviewerUserID = &invitation.InviteeUserID
		interview.HumanInterviewerName = invitation.Invitee.Username
		interview.HumanInterviewerRole = invitation.InviteeRole
		code := strings.TrimSpace(invitation.InvitationCode)
		if code != "" {
			interview.InvitationCode = &code
		}
	}

	if err := s.interviewRepo.Create(interview); err != nil {
		return nil, fmt.Errorf("failed to create interview: %w", err)
	}

	if invitation != nil {
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

func (s *InterviewService) fillOpeningQuestionsConcurrently(
	ctx context.Context,
	questions []*model.Question,
	target int,
	dummyInterview *model.Interview,
	capabilityGraph *model.JobCapabilityDimension,
	position, difficulty string,
	maxAttempts int,
	rewriteBudget *int32,
) []*model.Question {
	if target <= len(questions) || maxAttempts <= 0 {
		return questions
	}
	if ctx.Err() != nil {
		return questions
	}
	needed := target - len(questions)
	if needed <= 0 {
		return questions
	}
	if maxAttempts > needed {
		maxAttempts = needed
	}

	fillCtx, fillCancel := context.WithCancel(ctx)
	defer fillCancel()

	eg, egCtx := errgroup.WithContext(fillCtx)
	eg.SetLimit(fastInitAIFillConcurrency)

	generated := make([]*model.Question, 0, maxAttempts)
	var mu sync.Mutex

	for i := 0; i < maxAttempts; i++ {
		eg.Go(func() error {
			if egCtx.Err() != nil {
				return nil
			}
			taskCtx, taskCancel := context.WithTimeout(egCtx, fastInitAITaskTimeout)
			defer taskCancel()

			q, err := s.aiService.GenerateNextQuestionWithWeights(taskCtx, dummyInterview, nil, capabilityGraph)
			if err != nil {
				q, err = s.aiService.GenerateNextQuestion(taskCtx, dummyInterview, nil)
				if err != nil || q == nil {
					// 单次生成失败容错，不中断整体面试创建流程。
					return nil
				}
			}

			q.Position = position
			q.Difficulty = difficulty
			q.Source = "ai_opening"
			q.RAGEligible = true

			normalized := s.normalizeOpeningQuestionForInit(ctx, q, rewriteBudget)
			if normalized == nil {
				return nil
			}

			mu.Lock()
			defer mu.Unlock()
			if len(generated) >= needed {
				fillCancel()
				return nil
			}
			if s.isDuplicateOpeningQuestion(questions, normalized) || s.isDuplicateOpeningQuestion(generated, normalized) {
				return nil
			}
			generated = append(generated, normalized)
			if len(generated) >= needed {
				fillCancel()
			}
			return nil
		})
	}

	_ = eg.Wait()

	generated = s.prioritizeTopicVariety(generated, needed)
	for _, q := range generated {
		if len(questions) >= target {
			break
		}
		if s.isDuplicateOpeningQuestion(questions, q) {
			continue
		}
		if err := s.questionRepo.Create(q); err != nil {
			continue
		}
		questions = append(questions, q)
	}

	return questions
}

func (s *InterviewService) fillWithLocalFallbackQuestions(
	ctx context.Context,
	questions []*model.Question,
	target int,
	position, difficulty, mode, style string,
	rewriteBudget *int32,
) []*model.Question {
	if target <= 0 || len(questions) >= target {
		return questions
	}
	if ctx.Err() != nil {
		return questions
	}

	candidates := s.buildLocalFallbackOpeningQuestions(position, difficulty, mode, style, target*2)
	for _, candidate := range candidates {
		if len(questions) >= target {
			break
		}
		if ctx.Err() != nil {
			break
		}
		normalized := s.normalizeOpeningQuestionForInit(ctx, candidate, rewriteBudget)
		if normalized == nil || s.isDuplicateOpeningQuestion(questions, normalized) {
			continue
		}
		if err := s.questionRepo.Create(normalized); err != nil {
			continue
		}
		questions = append(questions, normalized)
	}

	return questions
}

func (s *InterviewService) buildLocalFallbackOpeningQuestions(position, difficulty, mode, style string, maxCount int) []*model.Question {
	type tpl struct {
		Title          string
		Content        string
		Category       string
		ExpectedAnswer string
	}

	technicalTemplates := []tpl{
		{
			Title:          "项目开场：核心挑战与价值",
			Content:        fmt.Sprintf("请结合你应聘的%s岗位，选一个你参与最深的项目，说明业务目标、你的职责、关键技术方案与最终效果。", position),
			Category:       "technical",
			ExpectedAnswer: "应包含项目背景、本人职责、方案选型依据、性能或稳定性收益及复盘。",
		},
		{
			Title:          "工程能力：定位与修复线上问题",
			Content:        "线上出现间歇性超时且日志信息不足时，你会如何进行排查、复现、止损与长期治理？请给出分阶段方案。",
			Category:       "technical",
			ExpectedAnswer: "应包含监控告警、日志补齐、链路追踪、灰度止损、根因分析与治理闭环。",
		},
		{
			Title:          "系统设计：高并发场景下的稳定性",
			Content:        "如果需要设计一个支持高并发访问的核心服务，你会如何从限流、缓存、降级、幂等和可观测性角度进行方案设计？",
			Category:       "system_design",
			ExpectedAnswer: "应包含架构分层、容量评估、流量治理、故障演练与监控指标。",
		},
		{
			Title:          "基础能力：复杂度与数据结构选择",
			Content:        "请举一个你在项目中通过替换数据结构或算法将性能显著提升的案例，并说明优化前后复杂度变化。",
			Category:       "algorithm",
			ExpectedAnswer: "应包含瓶颈定位、复杂度分析、实现细节及性能对比数据。",
		},
	}

	hrTemplates := []tpl{
		{
			Title:          "自我介绍与岗位匹配",
			Content:        fmt.Sprintf("请做一个 2 分钟自我介绍，并重点说明你与%s岗位最匹配的三项能力。", position),
			Category:       "hr",
			ExpectedAnswer: "应包含经历亮点、岗位匹配点与可量化成果。",
		},
		{
			Title:          "沟通协作：冲突处理",
			Content:        "当你与同事在技术方案上出现明显分歧时，你会如何推动达成共识并保证项目进度？",
			Category:       "hr",
			ExpectedAnswer: "应包含事实对齐、方案比较、决策机制与复盘。",
		},
		{
			Title:          "职业规划：短中期成长",
			Content:        "请说明你未来 1-3 年的职业目标，以及你计划通过哪些具体行动达成这些目标。",
			Category:       "hr",
			ExpectedAnswer: "应包含阶段目标、能力补齐计划与可验证里程碑。",
		},
	}

	algorithmTemplates := []tpl{
		{
			Title:          "算法兜底：数组与哈希",
			Content:        "给定一个整数数组，请设计 O(n) 复杂度算法找出目标和对应的两个下标，并说明边界处理。",
			Category:       "algorithm",
			ExpectedAnswer: "应说明哈希表思路、重复值处理与复杂度。",
		},
		{
			Title:          "算法兜底：滑动窗口",
			Content:        "请实现无重复字符最长子串问题，并解释窗口扩张、收缩与索引更新策略。",
			Category:       "algorithm",
			ExpectedAnswer: "应说明双指针与哈希映射维护，时间复杂度 O(n)。",
		},
		{
			Title:          "算法兜底：区间合并",
			Content:        "给定一组区间，输出合并后的结果，并解释排序后线性扫描的核心判断逻辑。",
			Category:       "algorithm",
			ExpectedAnswer: "应说明排序依据、重叠判定、边界更新与复杂度。",
		},
	}

	pick := technicalTemplates
	if style == "algorithm" {
		pick = algorithmTemplates
	} else if mode == "hr" {
		pick = hrTemplates
	} else if mode == "comprehensive" {
		pick = append(append([]tpl{}, technicalTemplates[:2]...), hrTemplates...)
	}

	if maxCount <= 0 {
		maxCount = len(pick)
	}

	questions := make([]*model.Question, 0, len(pick))
	for i := 0; i < len(pick) && len(questions) < maxCount; i++ {
		item := pick[i]
		questions = append(questions, &model.Question{
			Title:          item.Title,
			Content:        item.Content,
			Position:       position,
			Difficulty:     difficulty,
			Category:       item.Category,
			Source:         "local_opening",
			RAGEligible:    true,
			ExpectedAnswer: item.ExpectedAnswer,
		})
	}

	return questions
}

func (s *InterviewService) normalizeOpeningQuestionForInit(ctx context.Context, q *model.Question, rewriteBudget *int32) *model.Question {
	if q == nil {
		return nil
	}
	if strings.TrimSpace(q.Source) == "" {
		q.Source = "standard"
	}
	if q.Source == "follow_up" {
		return nil
	}

	q.Title = strings.TrimSpace(q.Title)
	q.Content = strings.TrimSpace(q.Content)
	q.ExpectedAnswer = strings.TrimSpace(q.ExpectedAnswer)

	if q.Title == "" {
		q.Title = "技术问题"
	}
	if q.Content == "" {
		q.Content = "请结合岗位要求，说明你的思路、关键实现和工程取舍。"
	}
	if q.ExpectedAnswer == "" {
		q.ExpectedAnswer = "回答应包含核心原理、实现步骤、关键细节与风险边界。"
	}

	if s.aiService.IsContextDependentOpeningQuestion(q) {
		s.aiService.NormalizeToSelfContainedOpening(q)
	}

	if shouldRewriteOpeningQuestionForInit(q) && consumeOpeningQuestionRewriteBudget(rewriteBudget) {
		baseCtx := ctx
		if baseCtx == nil {
			baseCtx = context.Background()
		}
		rewriteCtx, cancel := context.WithTimeout(baseCtx, fastInitChineseRewriteTimeout)
		s.aiService.EnsureQuestionChinese(rewriteCtx, q)
		cancel()
	}

	q.Title = strings.TrimSpace(q.Title)
	q.Content = strings.TrimSpace(q.Content)
	q.ExpectedAnswer = strings.TrimSpace(q.ExpectedAnswer)
	return q
}

func consumeOpeningQuestionRewriteBudget(budget *int32) bool {
	if budget == nil {
		return false
	}

	for {
		current := atomic.LoadInt32(budget)
		if current <= 0 {
			return false
		}
		if atomic.CompareAndSwapInt32(budget, current, current-1) {
			return true
		}
	}
}

func shouldRewriteOpeningQuestionForInit(q *model.Question) bool {
	if q == nil {
		return false
	}

	fields := []string{q.Title, q.Content, q.ExpectedAnswer}
	for _, field := range fields {
		text := strings.TrimSpace(field)
		if text == "" {
			continue
		}
		if strings.ContainsRune(text, '\ufffd') {
			return true
		}
		if openingQuestionEnglishTokenPattern.MatchString(text) {
			return true
		}
		if !isMostlyChineseForInit(text, 0.35) {
			return true
		}
	}

	return false
}

func isMostlyChineseForInit(text string, ratio float64) bool {
	content := strings.TrimSpace(text)
	if content == "" {
		return true
	}

	hanCount := 0
	letterCount := 0
	for _, r := range content {
		if unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsDigit(r) {
			continue
		}
		if unicode.IsLetter(r) {
			letterCount++
		}
		if r >= 0x4E00 && r <= 0x9FFF {
			hanCount++
		}
	}

	if letterCount == 0 {
		return true
	}
	return float64(hanCount)/float64(letterCount) >= ratio
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

func (s *InterviewService) buildAlgorithmOpeningQuestions(position, difficulty string) []*model.Question {
	problems := []struct {
		Title          string
		Content        string
		ExpectedAnswer string
	}{
		{
			Title:          "算法热身：双指针基础",
			Content:        "给定有序数组 nums 和目标值 target，找出两数之和等于 target 的下标组合，并说明你的时间复杂度。",
			ExpectedAnswer: "使用双指针从两端收敛，时间复杂度 O(n)，空间复杂度 O(1)。",
		},
		{
			Title:          "算法核心：滑动窗口",
			Content:        "给定字符串 s，求不含重复字符的最长子串长度，要求说明窗口更新策略和边界处理。",
			ExpectedAnswer: "使用哈希表记录字符索引，维护左边界，线性扫描，时间复杂度 O(n)。",
		},
		{
			Title:          "算法进阶：区间处理",
			Content:        "给定若干区间，输出合并后的区间集合，并解释排序后如何进行线性合并。",
			ExpectedAnswer: "先按起点排序，再比较当前区间与结果尾区间，重叠则扩展右边界，否则追加新区间。",
		},
	}

	level := strings.TrimSpace(difficulty)
	questions := make([]*model.Question, 0, len(problems))
	for _, item := range problems {
		q := &model.Question{
			Title:          item.Title,
			Content:        item.Content,
			Position:       position,
			Difficulty:     level,
			Category:       "algorithm",
			Source:         "algorithm_opening",
			RAGEligible:    true,
			ExpectedAnswer: item.ExpectedAnswer,
		}
		if err := s.questionRepo.Create(q); err != nil {
			continue
		}
		questions = append(questions, q)
	}
	return questions
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

func StartInterviewWithContext(ctx context.Context, userID uint, position, difficulty, mode, style, company, interviewMode string, invitationID *uint) (*model.Interview, error) {
	svc := NewInterviewService()
	return svc.StartInterviewWithContext(ctx, userID, position, difficulty, mode, style, company, interviewMode, invitationID)
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
	studentUUID, err := ensureUserUUID(svc.userRepo, student)
	if err != nil {
		return nil, fmt.Errorf("学生身份标识初始化失败: %w", err)
	}
	if student.Role != "student" {
		return nil, fmt.Errorf("仅学生用户可以发起真人面试邀请")
	}
	invitee, err := svc.userRepo.GetByID(inviteeUserID)
	if err != nil {
		return nil, fmt.Errorf("被邀请用户不存在")
	}
	inviteeUUID, err := ensureUserUUID(svc.userRepo, invitee)
	if err != nil {
		return nil, fmt.Errorf("被邀请用户身份标识初始化失败: %w", err)
	}
	if invitee.Role != "enterprise" && invitee.Role != "university" {
		return nil, fmt.Errorf("只能邀请学校端或企业端用户")
	}

	var invitationCode string
	for i := 0; i < 5; i++ {
		code, codeErr := generateInvitationCode()
		if codeErr != nil {
			return nil, fmt.Errorf("生成邀请码失败: %w", codeErr)
		}
		invitationCode = code
		if invitationCode != "" {
			break
		}
	}
	if invitationCode == "" {
		return nil, fmt.Errorf("生成邀请码失败")
	}

	inv := &model.HumanInterviewInvitation{
		InvitationCode: invitationCode,
		StudentID:      studentID,
		StudentUUID:    studentUUID,
		InviteeUserID:  inviteeUserID,
		InviteeUUID:    inviteeUUID,
		InviteeRole:    invitee.Role,
		Position:       strings.TrimSpace(position),
		Difficulty:     strings.TrimSpace(difficulty),
		Mode:           strings.TrimSpace(mode),
		Style:          strings.TrimSpace(style),
		Company:        strings.TrimSpace(company),
		Status:         "pending",
		ScheduledAt:    scheduledAt,
		Notes:          strings.TrimSpace(notes),
	}

	if err := svc.interviewRepo.CreateInvitation(inv); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(strings.ToLower(err.Error()), "unique") {
			for i := 0; i < 3; i++ {
				code, codeErr := generateInvitationCode()
				if codeErr != nil {
					break
				}
				inv.InvitationCode = code
				if retryErr := svc.interviewRepo.CreateInvitation(inv); retryErr == nil {
					err = nil
					break
				}
			}
		}
		if err != nil {
			return nil, fmt.Errorf("创建邀请失败: %w", err)
		}
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

	if strings.TrimSpace(invitation.InvitationCode) == "" {
		code, codeErr := generateInvitationCode()
		if codeErr != nil {
			return nil, fmt.Errorf("生成邀请码失败: %w", codeErr)
		}
		invitation.InvitationCode = code
	}

	if strings.TrimSpace(invitation.StudentUUID) == "" && strings.TrimSpace(invitation.Student.UUID) != "" {
		invitation.StudentUUID = strings.TrimSpace(invitation.Student.UUID)
	}
	if strings.TrimSpace(invitation.InviteeUUID) == "" {
		invitation.InviteeUUID = strings.TrimSpace(user.UUID)
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
	interview.AskedQuestionIDs = mergeAskedQuestionIDs(interview.AskedQuestionIDs, baseQuestion.ID)

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
	shouldFollowUpHint, followUpContext := parseEvaluationFollowUpHint(evaluation.Feedback)

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
	shouldFollowUp, nextQuestion, err := s.decideNextQuestion(ctx, interview, baseQuestion, askedQuestionText, finalAnswer, evaluation.Score, shouldFollowUpHint, followUpContext)
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
		excludeIDs := decodeAskedQuestionIDs(interview.AskedQuestionIDs)
		if interview.CurrentIndex < len(allQuestions) {
			nextQ, _ := s.questionRepo.GetByID(allQuestions[interview.CurrentIndex].QuestionID)
			if nextQ != nil {
				s.normalizeOpeningQuestion(ctx, nextQ)
				interview.CurrentTopic = nextQ.Category
				interview.AskedQuestionIDs = mergeAskedQuestionIDs(interview.AskedQuestionIDs, nextQ.ID)
				result.NextQuestion = nextQ
			}
		}

		if result.NextQuestion == nil {
			fallbackQ, pickErr := s.questionRepo.GetRandomQuestionForInterview(interview.Position, interview.Difficulty, excludeIDs)
			if pickErr == nil && fallbackQ != nil {
				s.normalizeOpeningQuestion(ctx, fallbackQ)
				interview.CurrentTopic = fallbackQ.Category
				interview.AskedQuestionIDs = mergeAskedQuestionIDs(interview.AskedQuestionIDs, fallbackQ.ID)
				result.NextQuestion = fallbackQ

				_ = s.interviewRepo.CreateInterviewQuestion(&model.InterviewQuestion{
					InterviewID: interviewID,
					QuestionID:  fallbackQ.ID,
					OrderIndex:  interview.CurrentIndex,
					IsAnswered:  false,
				})
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

func parseEvaluationFollowUpHint(feedback string) (bool, string) {
	trimmed := strings.TrimSpace(feedback)
	if trimmed == "" {
		return false, ""
	}

	var payload struct {
		ShouldFollowUp  bool     `json:"should_follow_up"`
		FollowUpContext string   `json:"follow_up_context"`
		Gaps            []string `json:"gaps"`
	}
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return false, ""
	}

	context := strings.TrimSpace(payload.FollowUpContext)
	if context == "" && len(payload.Gaps) > 0 {
		context = strings.TrimSpace(payload.Gaps[0])
	}

	shouldFollowUp := payload.ShouldFollowUp
	if !shouldFollowUp && strings.TrimSpace(payload.FollowUpContext) != "" {
		shouldFollowUp = true
	}
	if !shouldFollowUp && len(payload.Gaps) > 0 {
		shouldFollowUp = true
	}

	return shouldFollowUp, context
}

func collectQuestionIDs(questions []*model.Question) []uint {
	if len(questions) == 0 {
		return nil
	}
	ids := make([]uint, 0, len(questions))
	seen := make(map[uint]struct{}, len(questions))
	for _, q := range questions {
		if q == nil || q.ID == 0 {
			continue
		}
		if _, ok := seen[q.ID]; ok {
			continue
		}
		seen[q.ID] = struct{}{}
		ids = append(ids, q.ID)
	}
	return ids
}

func decodeAskedQuestionIDs(raw string) []uint {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return []uint{}
	}

	var parsed []uint
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return []uint{}
	}

	result := make([]uint, 0, len(parsed))
	seen := make(map[uint]struct{}, len(parsed))
	for _, id := range parsed {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func encodeAskedQuestionIDs(ids []uint) string {
	if len(ids) == 0 {
		return "[]"
	}
	normalized := make([]uint, 0, len(ids))
	seen := make(map[uint]struct{}, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		normalized = append(normalized, id)
	}
	if len(normalized) == 0 {
		return "[]"
	}

	encoded, err := json.Marshal(normalized)
	if err != nil {
		return "[]"
	}
	return string(encoded)
}

func mergeAskedQuestionIDs(raw string, extraIDs ...uint) string {
	merged := decodeAskedQuestionIDs(raw)
	seen := make(map[uint]struct{}, len(merged)+len(extraIDs))
	for _, id := range merged {
		seen[id] = struct{}{}
	}
	for _, id := range extraIDs {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		merged = append(merged, id)
	}
	return encodeAskedQuestionIDs(merged)
}

// decideNextQuestion determines if a follow-up is needed and generates it
func (s *InterviewService) decideNextQuestion(ctx context.Context, interview *model.Interview, topicRootQ *model.Question, askedQuestionText, answer string, score int, shouldFollowUpHint bool, followUpContext string) (bool, *model.Question, error) {
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

	resolvedFollowUpContext := strings.TrimSpace(followUpContext)
	if resolvedFollowUpContext == "" && shouldFollowUpHint {
		resolvedFollowUpContext = "请围绕候选人回答中未充分展开的关键机制继续追问其实现细节与边界条件。"
	}
	if resolvedFollowUpContext == "" && ragContext != "" {
		resolvedFollowUpContext = "请结合候选人回答与知识上下文中最关键的机制差异，继续追问其原理与取舍。"
	}

	nextQ, reason, err := s.aiService.GenerateFollowUpQuestion(ctx, interview, topicRootQ, answer, ragContext, resolvedFollowUpContext, interview.FollowUpCount)
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

	// 保留首题顺序，避免首题被后续按类别重排覆盖。
	anchor := unique[0]
	result = append(result, anchor)
	anchorCategory := strings.TrimSpace(anchor.Category)
	if anchorCategory != "" {
		usedCategory[anchorCategory] = true
	}
	if len(result) >= target {
		return result[:target]
	}

	for _, q := range unique[1:] {
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

	for _, q := range unique[1:] {
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
		inv, invErr := s.interviewRepo.GetInvitationByInterviewID(interviewID)
		if invErr == nil && inv != nil {
			inv.Status = "completed"
			_ = s.interviewRepo.UpdateInvitation(inv)
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
