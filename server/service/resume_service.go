package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"your-project/config"
	"your-project/model"
	"your-project/pkg/llm"
	"your-project/repository"
)

const (
	resumeMinLength     = 80
	resumeMaxLength     = 28000
	resumeModelVersion  = "resume-structured-v3"
	defaultParserMode   = "text"
	defaultResumeSource = "upload"
)

type ResumeService interface {
	AnalyzeOnly(ctx context.Context, input ResumeAnalysisInput) (*model.ResumeAnalysisResult, error)
	AnalyzeAndPersist(ctx context.Context, input ResumeAnalysisInput) (*model.ResumeAnalysisResult, *model.ResumeParseResult, error)
	GetLatestAnalysis(ctx context.Context, userID uint) (*model.ResumeAnalysisResult, *model.ResumeParseResult, error)
}

type ResumeAnalysisInput struct {
	UserID     uint
	FileName   string
	RawText    string
	Source     string
	ParserMode string
}

type LLMResumeService struct {
	llmClient    llm.LLMClient
	positionRepo repository.PositionRepository
	resumeRepo   repository.ResumeParseResultRepository
}

type resumeRoleProfile struct {
	Code            string   `json:"code"`
	RoleKey         string   `json:"role_key"`
	Name            string   `json:"name"`
	TechStack       []string `json:"tech_stack"`
	Keywords        []string `json:"keywords"`
	Requirements    []string `json:"requirements"`
	Checkpoints     []string `json:"checkpoints"`
	SampleQuestions []string `json:"sample_questions"`
}

var defaultResumeRoleProfiles = map[string]resumeRoleProfile{
	string(model.PositionAI): {
		Code:         string(model.PositionAI),
		RoleKey:      "ai_engineer",
		Name:         "AI工程师",
		TechStack:    []string{"Python", "PyTorch", "TensorFlow", "LLM", "RAG", "NLP", "向量数据库"},
		Keywords:     []string{"Prompt Engineering", "模型评测", "多模态", "Agent", "LoRA", "推理优化"},
		Requirements: []string{"具备模型应用或训练经验", "能够设计知识增强链路", "理解模型评测与上线治理"},
		Checkpoints:  []string{"RAG召回与重排", "提示词策略", "模型对齐", "推理成本控制"},
		SampleQuestions: []string{
			"请拆解一个 RAG 系统从召回到回答生成的关键环节。",
			"如果线上 LLM 幻觉率偏高，你会如何定位并缓解？",
		},
	},
	string(model.PositionBackend): {
		Code:         string(model.PositionBackend),
		RoleKey:      "java_backend",
		Name:         "后端工程师",
		TechStack:    []string{"Java", "Go", "Spring Boot", "MySQL", "Redis", "Kafka", "微服务"},
		Keywords:     []string{"高并发", "事务", "缓存", "JVM", "分布式", "消息队列"},
		Requirements: []string{"掌握主流后端开发框架", "能处理数据库与缓存性能问题", "理解分布式系统设计"},
		Checkpoints:  []string{"事务一致性", "SQL优化", "缓存设计", "服务治理"},
		SampleQuestions: []string{
			"在高并发场景下，你会如何设计缓存与数据库的一致性方案？",
			"请结合项目解释一次你做过的性能瓶颈定位与优化。",
		},
	},
	string(model.PositionFrontend): {
		Code:         string(model.PositionFrontend),
		RoleKey:      "web_frontend",
		Name:         "前端工程师",
		TechStack:    []string{"JavaScript", "TypeScript", "Vue", "React", "HTML", "CSS", "Vite"},
		Keywords:     []string{"组件化", "工程化", "性能优化", "响应式", "状态管理", "跨端"},
		Requirements: []string{"掌握至少一种主流前端框架", "理解前端工程化与发布流程", "具备性能优化经验"},
		Checkpoints:  []string{"渲染性能", "状态管理", "构建链路", "兼容性治理"},
		SampleQuestions: []string{
			"你如何定位并解决一个复杂页面的首屏性能问题？",
			"请说明你在组件抽象或前端工程化上的一次代表性实践。",
		},
	},
	string(model.PositionAlgorithm): {
		Code:         string(model.PositionAlgorithm),
		RoleKey:      "algorithm_engineer",
		Name:         "算法工程师",
		TechStack:    []string{"算法", "数据结构", "C++", "Python", "机器学习", "LeetCode"},
		Keywords:     []string{"复杂度分析", "图论", "动态规划", "搜索", "建模", "优化"},
		Requirements: []string{"具备扎实的数据结构与算法功底", "能够把算法方案落地为工程实现", "能分析性能瓶颈并优化"},
		Checkpoints:  []string{"复杂度权衡", "边界条件", "建模能力", "工程实现"},
		SampleQuestions: []string{
			"请挑一个你熟悉的算法问题，说明建模思路和复杂度优化过程。",
			"如果一个图搜索方案在大规模数据下超时，你会从哪些方向优化？",
		},
	},
}

var _ ResumeService = (*LLMResumeService)(nil)

func NewResumeService() ResumeService {
	return NewResumeServiceWithDeps(
		llm.NewDeepSeekClient(config.GetConfig()),
		repository.NewPositionRepository(),
		repository.NewResumeParseResultRepository(),
	)
}

func NewResumeServiceWithDeps(
	llmClient llm.LLMClient,
	positionRepo repository.PositionRepository,
	resumeRepo repository.ResumeParseResultRepository,
) ResumeService {
	if llmClient == nil {
		llmClient = llm.NewDeepSeekClient(config.GetConfig())
	}
	if positionRepo == nil {
		positionRepo = repository.NewPositionRepository()
	}
	if resumeRepo == nil {
		resumeRepo = repository.NewResumeParseResultRepository()
	}

	return &LLMResumeService{
		llmClient:    llmClient,
		positionRepo: positionRepo,
		resumeRepo:   resumeRepo,
	}
}

func (s *LLMResumeService) AnalyzeOnly(ctx context.Context, input ResumeAnalysisInput) (*model.ResumeAnalysisResult, error) {
	rawText := normalizeResumeText(input.RawText)
	if len([]rune(rawText)) < resumeMinLength {
		return nil, fmt.Errorf("resume text is too short")
	}

	positionCatalog := s.loadPositionCatalog()
	userPrompt, err := buildResumeExtractionPrompt(rawText, positionCatalog)
	if err != nil {
		return nil, fmt.Errorf("failed to build resume extraction prompt: %w", err)
	}

	rawOutput, err := s.chatJSON(ctx, buildResumeSystemPrompt(), userPrompt, "resume")
	if err != nil {
		return nil, err
	}

	payload, err := s.decodeStructuredPayload(rawOutput)
	if err != nil {
		payload, err = s.retryStructuredPayload(ctx, rawText, rawOutput, positionCatalog)
		if err != nil {
			return nil, err
		}
	}

	normalizeStructuredPayload(payload, rawText, positionCatalog)
	return buildResumeAnalysisResult(payload, input), nil
}

func (s *LLMResumeService) AnalyzeAndPersist(ctx context.Context, input ResumeAnalysisInput) (*model.ResumeAnalysisResult, *model.ResumeParseResult, error) {
	result, err := s.AnalyzeOnly(ctx, input)
	if err != nil {
		return nil, nil, err
	}

	record, err := s.persistResult(input, result)
	if err != nil {
		return nil, nil, err
	}
	return result, record, nil
}

func (s *LLMResumeService) GetLatestAnalysis(ctx context.Context, userID uint) (*model.ResumeAnalysisResult, *model.ResumeParseResult, error) {
	_ = ctx
	record, err := s.resumeRepo.GetLatestByUser(userID)
	if err != nil {
		return nil, nil, err
	}

	var result model.ResumeAnalysisResult
	if err := decodeStrictJSONObject(record.StructuredJSON, &result); err != nil {
		return nil, nil, fmt.Errorf("invalid stored resume analysis: %w", err)
	}
	return &result, record, nil
}

func (s *LLMResumeService) loadPositionCatalog() []resumeRoleProfile {
	positions, err := s.positionRepo.ListActive()
	if err != nil || len(positions) == 0 {
		positions = append([]model.JobPosition{}, model.DefaultJobPositions...)
	}

	catalog := make([]resumeRoleProfile, 0, len(positions))
	seen := make(map[string]struct{}, len(positions))
	for _, pos := range positions {
		code := strings.TrimSpace(pos.Code)
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}

		profile, ok := defaultResumeRoleProfiles[code]
		if !ok {
			profile = resumeRoleProfile{
				Code:            code,
				RoleKey:         code,
				Name:            strings.TrimSpace(pos.Name),
				TechStack:       []string{},
				Keywords:        []string{},
				Requirements:    []string{},
				Checkpoints:     []string{},
				SampleQuestions: []string{},
			}
		}
		if strings.TrimSpace(pos.Name) != "" {
			profile.Name = strings.TrimSpace(pos.Name)
		}
		catalog = append(catalog, profile)
	}

	sort.SliceStable(catalog, func(i, j int) bool {
		return catalog[i].Code < catalog[j].Code
	})
	return catalog
}

func (s *LLMResumeService) chatJSON(ctx context.Context, systemPrompt, userPrompt, taskType string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if s.llmClient == nil {
		return "", fmt.Errorf("llm client is not initialized")
	}

	req := llm.ChatRequest{
		Model: resolveModelForTask(taskType),
		Messages: []llm.ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature:    0.1,
		ResponseFormat: &llm.ResponseFormat{Type: llm.ResponseFormatJSON},
	}
	raw, err := s.llmClient.Chat(ctx, req)
	if err != nil {
		return "", fmt.Errorf("resume llm request failed: %w", err)
	}
	return strings.TrimSpace(raw), nil
}

func (s *LLMResumeService) decodeStructuredPayload(raw string) (*model.ResumeStructuredPayload, error) {
	candidates := extractJSONObjectCandidates(raw)
	for _, candidate := range candidates {
		var payload model.ResumeStructuredPayload
		if err := decodeStrictJSONObject(candidate, &payload); err != nil {
			continue
		}
		if err := validateStructuredPayload(&payload); err != nil {
			continue
		}
		return &payload, nil
	}
	return nil, fmt.Errorf("failed to decode structured resume payload")
}

func (s *LLMResumeService) retryStructuredPayload(ctx context.Context, rawText, brokenOutput string, catalog []resumeRoleProfile) (*model.ResumeStructuredPayload, error) {
	repairPrompt, err := buildResumeRepairPrompt(rawText, brokenOutput, catalog)
	if err != nil {
		return nil, fmt.Errorf("failed to build resume repair prompt: %w", err)
	}

	repaired, err := s.chatJSON(ctx, buildResumeSystemPrompt(), repairPrompt, "resume")
	if err != nil {
		return nil, fmt.Errorf("resume payload repair failed: %w", err)
	}

	payload, decodeErr := s.decodeStructuredPayload(repaired)
	if decodeErr != nil {
		return nil, fmt.Errorf("resume payload remained invalid after repair: %w", decodeErr)
	}
	return payload, nil
}

func (s *LLMResumeService) persistResult(input ResumeAnalysisInput, result *model.ResumeAnalysisResult) (*model.ResumeParseResult, error) {
	body, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resume analysis: %w", err)
	}

	record := &model.ResumeParseResult{
		UserID:          input.UserID,
		FileName:        strings.TrimSpace(input.FileName),
		FileHash:        hashText(normalizeResumeText(input.RawText)),
		RawText:         normalizeResumeText(input.RawText),
		StructuredJSON:  string(body),
		ConfidenceScore: clampScore(result.ConfidenceScore),
		ParserMode:      normalizeParserMode(input.ParserMode),
		ParserVersion:   resumeModelVersion,
		Source:          normalizeResumeSource(input.Source),
	}
	if result.BestMatch != nil {
		record.PrimaryPositionCode = strings.TrimSpace(result.BestMatch.PositionCode)
		record.PrimaryPositionName = strings.TrimSpace(result.BestMatch.PositionName)
	}

	if err := s.resumeRepo.Create(record); err != nil {
		return nil, err
	}
	return record, nil
}

func buildResumeAnalysisResult(payload *model.ResumeStructuredPayload, input ResumeAnalysisInput) *model.ResumeAnalysisResult {
	bestMatch := pickBestMatch(payload.MatchResults)
	now := time.Now().Format(time.RFC3339)

	return &model.ResumeAnalysisResult{
		ParserStatus:       "success",
		ParserMode:         normalizeParserMode(input.ParserMode),
		Source:             normalizeResumeSource(input.Source),
		Architecture:       defaultResumeArchitecture(),
		AgenticRAGTrace:    buildResumeTrace(input.FileName, bestMatch),
		MoERouting:         buildResumeMoERouting(payload.MatchResults),
		StructuredResume:   payload.StructuredResume,
		MatchResults:       payload.MatchResults,
		BestMatch:          bestMatch,
		InterviewQuestions: payload.InterviewQuestions,
		Optimization:       payload.Optimization,
		RiskReport:         payload.RiskReport,
		Integration:        buildResumeIntegration(payload, bestMatch, now),
		ConfidenceScore:    clampScore(payload.ConfidenceScore),
		ModelVersion:       resumeModelVersion,
	}
}

func buildResumeSystemPrompt() string {
	return strings.TrimSpace(`
你是资深技术招聘顾问、简历结构化提取专家和岗位匹配分析师。
你的唯一任务是根据候选人简历原文与给定岗位知识目录，输出一个严格合法的 JSON 对象。

硬性规则：
1. 只能输出 JSON，不要 Markdown、代码块、注释、解释、前后缀。
2. 只能依据简历原文提取，不允许编造不存在的学校、公司、时间、技术、职责或成果。
3. 证据不足时使用空字符串、空数组或保守表述，不要臆测。
4. 所有数组字段都必须返回 JSON 数组，不能为 null。
5. structured_resume.skill_graph 中每个技能都必须给出 evidence，没有证据就不要写。
6. match_results.position_code 只能从岗位知识目录中选择，并按 score 从高到低排序。
7. match_results.score_breakdown 的四个维度和 confidence_score 都必须是 0-100 的整数。
8. risk_report 需要指出信息缺口、真实性风险或表达问题，并尽量附带 evidence。
9. optimization 必须写成可执行动作，priority 只能是 high、medium、low。
10. interview_questions 必须围绕最匹配岗位给出高质量追问，并填充 focus_skills。
`)
}

func buildResumeExtractionPrompt(rawText string, catalog []resumeRoleProfile) (string, error) {
	catalogJSON, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return "", err
	}
	schemaJSON, err := json.MarshalIndent(sampleStructuredPayload(), "", "  ")
	if err != nil {
		return "", err
	}

	var prompt strings.Builder
	prompt.WriteString("请阅读下面的岗位知识目录与候选人简历原文，输出一个严格满足 schema 的 JSON 对象。\n\n")
	prompt.WriteString("岗位知识目录（position_code 只能从这里选）：\n")
	prompt.Write(catalogJSON)
	prompt.WriteString("\n\n字段要求：\n")
	prompt.WriteString("1. structured_resume 需要覆盖：基本信息、职业摘要、求职意向、教育、工作经历、项目经历、技能图谱、证书、奖项、语言、亮点、隐患、原文摘要。\n")
	prompt.WriteString("2. skill_graph 要按分类整理，并尽量给出 level / evidence / last_used。\n")
	prompt.WriteString("3. match_results 至少给出最相关岗位的充分证据，并解释匹配度、短板与要求。\n")
	prompt.WriteString("4. interview_questions 用于后续深挖面试，不能泛泛而谈。\n")
	prompt.WriteString("5. optimization 要围绕简历改写、证据补强、叙事提升给动作建议。\n")
	prompt.WriteString("6. risk_report 要指出缺失信息、真实性风险、项目描述过浅、量化结果不足等问题。\n")
	prompt.WriteString("7. confidence_score 代表本次结构化提取整体可靠性。\n")
	prompt.WriteString("\n输出 schema 示例（字段名必须完全一致）：\n")
	prompt.Write(schemaJSON)
	prompt.WriteString("\n\n候选人简历原文如下：\n")
	prompt.WriteString(rawText)
	return prompt.String(), nil
}

func buildResumeRepairPrompt(rawText, brokenOutput string, catalog []resumeRoleProfile) (string, error) {
	catalogJSON, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return "", err
	}
	schemaJSON, err := json.MarshalIndent(sampleStructuredPayload(), "", "  ")
	if err != nil {
		return "", err
	}

	var prompt strings.Builder
	prompt.WriteString("上一次输出不是严格合法 JSON。请重新执行提取，并只输出一个合法 JSON 对象。\n\n")
	prompt.WriteString("岗位知识目录：\n")
	prompt.Write(catalogJSON)
	prompt.WriteString("\n\n必须满足的 schema：\n")
	prompt.Write(schemaJSON)
	prompt.WriteString("\n\n上一轮错误输出（仅供参考，不要原样照抄）：\n")
	prompt.WriteString(strings.TrimSpace(brokenOutput))
	prompt.WriteString("\n\n请基于下面的简历原文重新生成严格 JSON：\n")
	prompt.WriteString(rawText)
	return prompt.String(), nil
}

func sampleStructuredPayload() model.ResumeStructuredPayload {
	return model.ResumeStructuredPayload{
		StructuredResume: model.ResumeStructuredResume{
			PersonalInfo: model.ResumePersonalInfo{
				Name:      "候选人姓名",
				Email:     "candidate@example.com",
				Phone:     "13800138000",
				Location:  "上海",
				Github:    "https://github.com/example",
				Portfolio: "",
				LinkedIn:  "",
			},
			ProfessionalSummary: "3 年后端开发经验，主导过高并发业务接口与中台服务建设。",
			CareerIntent: model.ResumeCareerIntent{
				TargetRoles:      []string{"后端工程师"},
				TargetIndustries: []string{"互联网"},
				TargetCities:     []string{"上海"},
				Seniority:        "社招初级",
			},
			Education: []model.ResumeEducationExperience{
				{
					School:     "XX大学",
					Degree:     "本科",
					Major:      "计算机科学与技术",
					StartDate:  "2019-09",
					EndDate:    "2023-06",
					GPA:        "",
					Ranking:    "",
					Highlights: []string{"获校级奖学金"},
				},
			},
			WorkExperience: []model.ResumeWorkExperience{
				{
					Company:          "XX科技",
					Role:             "后端开发工程师",
					StartDate:        "2023-07",
					EndDate:          "至今",
					Duration:         "1年+",
					Summary:          "负责订单链路相关服务开发与优化。",
					Responsibilities: []string{"设计核心接口", "维护数据库与缓存方案"},
					Achievements:     []string{"接口延迟下降 30%"},
					TechStack:        []string{"Java", "Spring Boot", "MySQL", "Redis"},
				},
			},
			ProjectExperience: []model.ResumeProjectExperience{
				{
					Name:       "智能推荐平台",
					Role:       "核心开发",
					StartDate:  "2023-09",
					EndDate:    "2024-02",
					Background: "面向业务团队提供推荐能力。",
					Summary:    "负责推荐服务接口与数据链路。",
					TechStack:  []string{"Java", "Redis", "Kafka"},
					Highlights: []string{"完成高并发接口重构"},
					Impact:     []string{"峰值吞吐提升 40%"},
				},
			},
			SkillGraph: model.ResumeSkillGraph{
				ProgrammingLanguages: []model.ResumeSkillEvidence{
					{Name: "Java", Level: "advanced", Evidence: "工作经历与项目中多次出现", LastUsed: "current"},
				},
				Frameworks: []model.ResumeSkillEvidence{
					{Name: "Spring Boot", Level: "advanced", Evidence: "主导服务开发", LastUsed: "current"},
				},
				Databases: []model.ResumeSkillEvidence{
					{Name: "MySQL", Level: "intermediate", Evidence: "负责 SQL 优化", LastUsed: "current"},
				},
				CloudDevOps:     []model.ResumeSkillEvidence{},
				AIData:          []model.ResumeSkillEvidence{},
				Tooling:         []model.ResumeSkillEvidence{},
				ProductBusiness: []model.ResumeSkillEvidence{},
				Others:          []model.ResumeSkillEvidence{},
			},
			Certifications: []model.ResumeCredential{
				{Name: "英语六级", Issuer: "教育部考试中心", AwardedDate: ""},
			},
			Awards: []model.ResumeHonor{
				{Name: "校级奖学金", AwardedBy: "XX大学", AwardedDate: "2022", Detail: "学习成绩优秀"},
			},
			Languages: []model.ResumeLanguageProficiency{
				{Language: "英语", Proficiency: "CET-6", Evidence: "证书"},
			},
			Highlights: []string{"有明确性能优化成果", "项目经历与目标岗位相关"},
			Concerns:   []string{"部分成果量化维度仍可补充"},
			RawPreview: "简历原文摘要",
		},
		MatchResults: []model.ResumePositionMatch{
			{
				PositionCode: "backend",
				PositionName: "后端工程师",
				RoleKey:      "java_backend",
				Score:        86,
				ScoreBreakdown: model.ResumeMatchScoreBreakdown{
					SkillDepth:       88,
					ProjectRelevance: 84,
					DomainAlignment:  80,
					DeliveryImpact:   82,
				},
				HitSkills:    []string{"Java", "Spring Boot", "MySQL", "Redis"},
				HitKeywords:  []string{"高并发", "事务", "缓存"},
				Evidence:     []string{"工作经历有后端服务开发", "项目包含数据库和缓存优化"},
				GapSkills:    []string{"服务治理", "分布式事务"},
				Requirements: []string{"理解数据库与缓存优化", "具备分布式系统设计意识"},
				Analysis:     "后端技术栈与项目经历匹配度较高，但分布式治理证据还不够充分。",
			},
		},
		InterviewQuestions: []model.ResumeSuggestedQuestion{
			{
				Question:    "请结合项目说明你做过的一次缓存一致性设计。",
				Intent:      "验证分布式系统设计深度",
				FocusSkills: []string{"Redis", "缓存一致性", "高并发"},
			},
		},
		Optimization: []model.ResumeOptimizationSuggestion{
			{
				Title:     "补强结果量化",
				Action:    "为每段项目经历补充吞吐、延迟、成本等指标。",
				Rationale: "量化结果能更有力地证明影响力。",
				Priority:  "high",
			},
		},
		RiskReport: []model.ResumeRiskItem{
			{
				Level:    "medium",
				Item:     "分布式治理证据不足",
				Detail:   "简历提到高并发与缓存，但对服务治理、容灾、链路稳定性描述不够。",
				Evidence: []string{"项目描述中没有明确服务治理或容灾细节"},
			},
		},
		ConfidenceScore: 85,
	}
}

func defaultResumeArchitecture() model.ResumeAnalysisArchitecture {
	return model.ResumeAnalysisArchitecture{
		DecisionStack:   "Structured Extraction + Role Knowledge Grounding + Evidence First Normalization",
		PainPointsFixed: []string{"输出非结构化", "岗位匹配维度单薄", "字段不稳定", "缺少证据链"},
		Modules: model.ResumeArchitectureModules{
			FeatureExtractor: "Document text extraction + schema constrained LLM parsing",
			CoreProcessor:    "Role knowledge grounding + contract validation + JSON repair",
			OutputGenerator:  "Position match view + interview follow-ups + optimization + risk report",
		},
	}
}

func buildResumeTrace(fileName string, bestMatch *model.ResumePositionMatch) []model.ResumeAnalysisTraceStep {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(strings.TrimSpace(fileName))), ".")
	if ext == "" {
		ext = "unknown"
	}
	target := "unknown"
	if bestMatch != nil && strings.TrimSpace(bestMatch.PositionName) != "" {
		target = strings.TrimSpace(bestMatch.PositionName)
	}
	return []model.ResumeAnalysisTraceStep{
		{Agent: "ingest-agent", Decision: fmt.Sprintf("Detected source file extension %s and normalized the resume text input.", ext)},
		{Agent: "extract-agent", Decision: "Executed schema-constrained extraction for identity, education, experience, projects and skill graph."},
		{Agent: "grounding-agent", Decision: "Injected role knowledge catalog so every position match is grounded to known position codes and requirements."},
		{Agent: "match-agent", Decision: fmt.Sprintf("Ranked position matches and selected %s as the current top focus.", target)},
		{Agent: "output-agent", Decision: "Composed interview questions, optimization actions and risk report from the structured analysis."},
	}
}

func buildResumeMoERouting(matches []model.ResumePositionMatch) model.ResumeMoERouting {
	top := matches
	if len(top) > 2 {
		top = top[:2]
	}
	experts := make([]model.ResumeMoEExpert, 0, len(top))
	for _, item := range top {
		experts = append(experts, model.ResumeMoEExpert{
			Expert: item.RoleKey,
			Weight: roundWeight(float64(clampScore(item.Score)) / 100),
			Reason: fmt.Sprintf("Matched %s with score %d and prioritized focused follow-up generation.", item.PositionName, clampScore(item.Score)),
		})
	}
	return model.ResumeMoERouting{
		Router:         "Resume-MoE Router",
		Experts:        experts,
		FusionStrategy: "Top-2 weighted fusion",
	}
}

func buildResumeIntegration(payload *model.ResumeStructuredPayload, bestMatch *model.ResumePositionMatch, generatedAt string) model.ResumeIntegrationPayload {
	weakPoints := []string{}
	targetRole := ""
	targetPosition := ""
	if bestMatch != nil {
		weakPoints = firstN(uniqueStrings(bestMatch.GapSkills), 3)
		targetRole = strings.TrimSpace(bestMatch.RoleKey)
		targetPosition = strings.TrimSpace(bestMatch.PositionName)
	}

	questionRecommendations := make([]string, 0, len(payload.InterviewQuestions))
	for _, item := range payload.InterviewQuestions {
		if question := strings.TrimSpace(item.Question); question != "" {
			questionRecommendations = append(questionRecommendations, question)
		}
	}

	return model.ResumeIntegrationPayload{
		TargetRole:              targetRole,
		TargetPosition:          targetPosition,
		WeakPoints:              weakPoints,
		QuestionRecommendations: questionRecommendations,
		DrillPlan: model.ResumeDrillPlan{
			Phase1: "Fill evidence gaps and quantify project outcomes.",
			Phase2: "Drill the weak skills that block the best matched role.",
			Phase3: "Run targeted mock interviews with the generated questions.",
		},
		InterviewPayload: model.ResumeInterviewPayload{
			CandidateContact: payload.StructuredResume.PersonalInfo,
			FocusTopics:      weakPoints,
			GeneratedAt:      generatedAt,
		},
	}
}

func normalizeStructuredPayload(payload *model.ResumeStructuredPayload, rawText string, catalog []resumeRoleProfile) {
	if payload == nil {
		return
	}

	payload.StructuredResume = normalizeStructuredResume(payload.StructuredResume, rawText)
	payload.MatchResults = normalizeMatchResults(payload.MatchResults, catalog)
	payload.InterviewQuestions = normalizeInterviewQuestions(payload.InterviewQuestions)
	if len(payload.InterviewQuestions) == 0 && len(payload.MatchResults) > 0 {
		payload.InterviewQuestions = synthesizeCatalogQuestions(payload.MatchResults[0], catalog)
	}
	payload.Optimization = normalizeOptimization(payload.Optimization)
	payload.RiskReport = normalizeRiskReport(payload.RiskReport)
	payload.ConfidenceScore = clampScore(payload.ConfidenceScore)
}

func normalizeStructuredResume(in model.ResumeStructuredResume, rawText string) model.ResumeStructuredResume {
	in.PersonalInfo = normalizePersonalInfo(in.PersonalInfo)
	in.ProfessionalSummary = strings.TrimSpace(in.ProfessionalSummary)
	in.CareerIntent = normalizeCareerIntent(in.CareerIntent)
	in.Education = normalizeEducationList(in.Education)
	in.WorkExperience = normalizeWorkList(in.WorkExperience)
	in.ProjectExperience = normalizeProjectList(in.ProjectExperience)
	in.SkillGraph = normalizeSkillGraph(in.SkillGraph)
	in.Certifications = normalizeCredentialList(in.Certifications)
	in.Awards = normalizeHonorList(in.Awards)
	in.Languages = normalizeLanguageList(in.Languages)
	in.Highlights = uniqueStrings(in.Highlights)
	in.Concerns = uniqueStrings(in.Concerns)
	in.RawPreview = strings.TrimSpace(in.RawPreview)
	if in.RawPreview == "" {
		in.RawPreview = truncateRunes(rawText, 800)
	}
	return in
}

func normalizePersonalInfo(in model.ResumePersonalInfo) model.ResumePersonalInfo {
	in.Name = strings.TrimSpace(in.Name)
	in.Email = strings.TrimSpace(in.Email)
	in.Phone = strings.TrimSpace(in.Phone)
	in.Location = strings.TrimSpace(in.Location)
	in.Github = strings.TrimSpace(in.Github)
	in.Portfolio = strings.TrimSpace(in.Portfolio)
	in.LinkedIn = strings.TrimSpace(in.LinkedIn)
	return in
}

func normalizeCareerIntent(in model.ResumeCareerIntent) model.ResumeCareerIntent {
	in.TargetRoles = uniqueStrings(in.TargetRoles)
	in.TargetIndustries = uniqueStrings(in.TargetIndustries)
	in.TargetCities = uniqueStrings(in.TargetCities)
	in.Seniority = strings.TrimSpace(in.Seniority)
	return in
}

func normalizeEducationList(items []model.ResumeEducationExperience) []model.ResumeEducationExperience {
	out := make([]model.ResumeEducationExperience, 0, len(items))
	for _, item := range items {
		item.School = strings.TrimSpace(item.School)
		item.Degree = strings.TrimSpace(item.Degree)
		item.Major = strings.TrimSpace(item.Major)
		item.StartDate = strings.TrimSpace(item.StartDate)
		item.EndDate = strings.TrimSpace(item.EndDate)
		item.GPA = strings.TrimSpace(item.GPA)
		item.Ranking = strings.TrimSpace(item.Ranking)
		item.Highlights = uniqueStrings(item.Highlights)
		if item.School == "" && item.Degree == "" && item.Major == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}

func normalizeWorkList(items []model.ResumeWorkExperience) []model.ResumeWorkExperience {
	out := make([]model.ResumeWorkExperience, 0, len(items))
	for _, item := range items {
		item.Company = strings.TrimSpace(item.Company)
		item.Role = strings.TrimSpace(item.Role)
		item.StartDate = strings.TrimSpace(item.StartDate)
		item.EndDate = strings.TrimSpace(item.EndDate)
		item.Duration = strings.TrimSpace(item.Duration)
		item.Summary = strings.TrimSpace(item.Summary)
		item.Responsibilities = uniqueStrings(item.Responsibilities)
		item.Achievements = uniqueStrings(item.Achievements)
		item.TechStack = uniqueStrings(item.TechStack)
		if item.Company == "" && item.Role == "" && item.Summary == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}

func normalizeProjectList(items []model.ResumeProjectExperience) []model.ResumeProjectExperience {
	out := make([]model.ResumeProjectExperience, 0, len(items))
	for _, item := range items {
		item.Name = strings.TrimSpace(item.Name)
		item.Role = strings.TrimSpace(item.Role)
		item.StartDate = strings.TrimSpace(item.StartDate)
		item.EndDate = strings.TrimSpace(item.EndDate)
		item.Background = strings.TrimSpace(item.Background)
		item.Summary = strings.TrimSpace(item.Summary)
		item.TechStack = uniqueStrings(item.TechStack)
		item.Highlights = uniqueStrings(item.Highlights)
		item.Impact = uniqueStrings(item.Impact)
		if item.Name == "" && item.Role == "" && item.Summary == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}

func normalizeSkillGraph(in model.ResumeSkillGraph) model.ResumeSkillGraph {
	in.ProgrammingLanguages = normalizeSkillEvidenceList(in.ProgrammingLanguages)
	in.Frameworks = normalizeSkillEvidenceList(in.Frameworks)
	in.Databases = normalizeSkillEvidenceList(in.Databases)
	in.CloudDevOps = normalizeSkillEvidenceList(in.CloudDevOps)
	in.AIData = normalizeSkillEvidenceList(in.AIData)
	in.Tooling = normalizeSkillEvidenceList(in.Tooling)
	in.ProductBusiness = normalizeSkillEvidenceList(in.ProductBusiness)
	in.Others = normalizeSkillEvidenceList(in.Others)
	return in
}

func normalizeSkillEvidenceList(items []model.ResumeSkillEvidence) []model.ResumeSkillEvidence {
	out := make([]model.ResumeSkillEvidence, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		item.Name = strings.TrimSpace(item.Name)
		item.Level = strings.TrimSpace(item.Level)
		item.Evidence = strings.TrimSpace(item.Evidence)
		item.LastUsed = strings.TrimSpace(item.LastUsed)
		if item.Name == "" {
			continue
		}
		key := strings.ToLower(item.Name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}

func normalizeCredentialList(items []model.ResumeCredential) []model.ResumeCredential {
	out := make([]model.ResumeCredential, 0, len(items))
	for _, item := range items {
		item.Name = strings.TrimSpace(item.Name)
		item.Issuer = strings.TrimSpace(item.Issuer)
		item.AwardedDate = strings.TrimSpace(item.AwardedDate)
		if item.Name == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}

func normalizeHonorList(items []model.ResumeHonor) []model.ResumeHonor {
	out := make([]model.ResumeHonor, 0, len(items))
	for _, item := range items {
		item.Name = strings.TrimSpace(item.Name)
		item.AwardedBy = strings.TrimSpace(item.AwardedBy)
		item.AwardedDate = strings.TrimSpace(item.AwardedDate)
		item.Detail = strings.TrimSpace(item.Detail)
		if item.Name == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}

func normalizeLanguageList(items []model.ResumeLanguageProficiency) []model.ResumeLanguageProficiency {
	out := make([]model.ResumeLanguageProficiency, 0, len(items))
	for _, item := range items {
		item.Language = strings.TrimSpace(item.Language)
		item.Proficiency = strings.TrimSpace(item.Proficiency)
		item.Evidence = strings.TrimSpace(item.Evidence)
		if item.Language == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}

func normalizeMatchResults(items []model.ResumePositionMatch, catalog []resumeRoleProfile) []model.ResumePositionMatch {
	profileByCode := make(map[string]resumeRoleProfile, len(catalog))
	for _, item := range catalog {
		profileByCode[item.Code] = item
	}

	out := make([]model.ResumePositionMatch, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		item.PositionCode = strings.TrimSpace(item.PositionCode)
		item.PositionName = strings.TrimSpace(item.PositionName)
		item.RoleKey = strings.TrimSpace(item.RoleKey)
		item.Score = clampScore(item.Score)
		item.ScoreBreakdown = clampScoreBreakdown(item.ScoreBreakdown)
		item.HitSkills = uniqueStrings(item.HitSkills)
		item.HitKeywords = uniqueStrings(item.HitKeywords)
		item.Evidence = uniqueStrings(item.Evidence)
		item.GapSkills = uniqueStrings(item.GapSkills)
		item.Requirements = uniqueStrings(item.Requirements)
		item.Analysis = strings.TrimSpace(item.Analysis)

		if item.PositionCode == "" {
			item.PositionCode = inferPositionCodeByName(item.PositionName, catalog)
		}
		profile, ok := profileByCode[item.PositionCode]
		if !ok {
			continue
		}
		if item.PositionName == "" {
			item.PositionName = profile.Name
		}
		if item.RoleKey == "" {
			item.RoleKey = profile.RoleKey
		}
		if len(item.Requirements) == 0 {
			item.Requirements = append([]string{}, profile.Requirements...)
		}
		if _, exists := seen[item.PositionCode]; exists {
			continue
		}
		seen[item.PositionCode] = struct{}{}
		out = append(out, item)
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			return out[i].PositionCode < out[j].PositionCode
		}
		return out[i].Score > out[j].Score
	})
	return out
}

func normalizeInterviewQuestions(items []model.ResumeSuggestedQuestion) []model.ResumeSuggestedQuestion {
	out := make([]model.ResumeSuggestedQuestion, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		item.Question = strings.TrimSpace(item.Question)
		item.Intent = strings.TrimSpace(item.Intent)
		item.FocusSkills = uniqueStrings(item.FocusSkills)
		if item.Question == "" {
			continue
		}
		key := strings.ToLower(item.Question)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}

func synthesizeCatalogQuestions(bestMatch model.ResumePositionMatch, catalog []resumeRoleProfile) []model.ResumeSuggestedQuestion {
	for _, profile := range catalog {
		if profile.Code != bestMatch.PositionCode {
			continue
		}
		out := make([]model.ResumeSuggestedQuestion, 0, len(profile.SampleQuestions))
		for _, question := range profile.SampleQuestions {
			out = append(out, model.ResumeSuggestedQuestion{
				Question:    question,
				Intent:      "Validate the depth and authenticity of the candidate's matching role experience.",
				FocusSkills: firstN(uniqueStrings(append(bestMatch.GapSkills, bestMatch.HitSkills...)), 3),
			})
		}
		return out
	}
	return []model.ResumeSuggestedQuestion{}
}

func normalizeOptimization(items []model.ResumeOptimizationSuggestion) []model.ResumeOptimizationSuggestion {
	out := make([]model.ResumeOptimizationSuggestion, 0, len(items))
	for _, item := range items {
		item.Title = strings.TrimSpace(item.Title)
		item.Action = strings.TrimSpace(item.Action)
		item.Rationale = strings.TrimSpace(item.Rationale)
		item.Priority = normalizePriority(item.Priority)
		if item.Title == "" && item.Action == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}

func normalizeRiskReport(items []model.ResumeRiskItem) []model.ResumeRiskItem {
	out := make([]model.ResumeRiskItem, 0, len(items))
	for _, item := range items {
		item.Level = normalizePriority(item.Level)
		if item.Level == "" {
			item.Level = "low"
		}
		item.Item = strings.TrimSpace(item.Item)
		item.Detail = strings.TrimSpace(item.Detail)
		item.Evidence = uniqueStrings(item.Evidence)
		if item.Item == "" && item.Detail == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}

func validateStructuredPayload(payload *model.ResumeStructuredPayload) error {
	if payload == nil {
		return fmt.Errorf("payload is nil")
	}
	resume := payload.StructuredResume
	skillCount := len(resume.SkillGraph.ProgrammingLanguages) + len(resume.SkillGraph.Frameworks) + len(resume.SkillGraph.Databases) + len(resume.SkillGraph.CloudDevOps) + len(resume.SkillGraph.AIData) + len(resume.SkillGraph.Tooling) + len(resume.SkillGraph.ProductBusiness) + len(resume.SkillGraph.Others)
	if strings.TrimSpace(resume.ProfessionalSummary) == "" &&
		len(resume.Education) == 0 &&
		len(resume.WorkExperience) == 0 &&
		len(resume.ProjectExperience) == 0 &&
		skillCount == 0 {
		return fmt.Errorf("payload is not meaningful")
	}
	return nil
}

func decodeStrictJSONObject(raw string, dest interface{}) error {
	decoder := json.NewDecoder(strings.NewReader(strings.TrimSpace(raw)))
	decoder.DisallowUnknownFields()
	return decoder.Decode(dest)
}

func extractJSONObjectCandidates(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return []string{}
	}
	candidates := make([]string, 0, 6)
	seen := map[string]struct{}{}

	addCandidate := func(candidate string) {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			return
		}
		if _, ok := seen[candidate]; ok {
			return
		}
		seen[candidate] = struct{}{}
		candidates = append(candidates, candidate)
	}

	unfenced := stripCodeFence(trimmed)
	addCandidate(unfenced)
	if start := strings.Index(unfenced, "{"); start >= 0 {
		if end := strings.LastIndex(unfenced, "}"); end > start {
			addCandidate(unfenced[start : end+1])
		}
	}
	for _, item := range balancedJSONObjectSegments(unfenced) {
		addCandidate(item)
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		return len(candidates[i]) > len(candidates[j])
	})
	return candidates
}

func stripCodeFence(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```json")
		trimmed = strings.TrimPrefix(trimmed, "```JSON")
		trimmed = strings.TrimPrefix(trimmed, "```")
		trimmed = strings.TrimSuffix(trimmed, "```")
	}
	return strings.TrimSpace(trimmed)
}

func balancedJSONObjectSegments(raw string) []string {
	segments := []string{}
	inString := false
	escape := false
	depth := 0
	start := -1

	for idx, r := range raw {
		switch {
		case escape:
			escape = false
		case r == '\\':
			escape = true
		case r == '"':
			inString = !inString
		case !inString && r == '{':
			if depth == 0 {
				start = idx
			}
			depth++
		case !inString && r == '}':
			if depth == 0 {
				continue
			}
			depth--
			if depth == 0 && start >= 0 {
				segments = append(segments, raw[start:idx+1])
				start = -1
			}
		}
	}

	return segments
}

func resolveModelForTask(taskType string) string {
	cfg := config.GetConfig()
	if cfg == nil {
		return config.DefaultDeepSeekModel
	}
	if modelName := strings.TrimSpace(cfg.LLM.Models[taskType]); modelName != "" {
		return modelName
	}
	if modelName := strings.TrimSpace(cfg.LLM.Model); modelName != "" {
		return modelName
	}
	return config.DefaultDeepSeekModel
}

func pickBestMatch(matches []model.ResumePositionMatch) *model.ResumePositionMatch {
	if len(matches) == 0 {
		return nil
	}
	best := matches[0]
	return &best
}

func clampScoreBreakdown(in model.ResumeMatchScoreBreakdown) model.ResumeMatchScoreBreakdown {
	in.SkillDepth = clampScore(in.SkillDepth)
	in.ProjectRelevance = clampScore(in.ProjectRelevance)
	in.DomainAlignment = clampScore(in.DomainAlignment)
	in.DeliveryImpact = clampScore(in.DeliveryImpact)
	return in
}

func normalizeResumeText(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	runes := []rune(trimmed)
	if len(runes) > resumeMaxLength {
		runes = runes[:resumeMaxLength]
	}
	return strings.TrimSpace(string(runes))
}

func normalizeResumeSource(source string) string {
	normalized := strings.TrimSpace(source)
	if normalized == "" {
		return defaultResumeSource
	}
	return normalized
}

func normalizeParserMode(mode string) string {
	normalized := strings.TrimSpace(mode)
	if normalized == "" {
		return defaultParserMode
	}
	return normalized
}

func normalizePriority(priority string) string {
	switch strings.ToLower(strings.TrimSpace(priority)) {
	case "high", "medium", "low":
		return strings.ToLower(strings.TrimSpace(priority))
	default:
		return ""
	}
}

func inferPositionCodeByName(name string, catalog []resumeRoleProfile) string {
	normalized := strings.ToLower(strings.TrimSpace(name))
	for _, item := range catalog {
		if strings.ToLower(item.Name) == normalized {
			return item.Code
		}
	}
	for _, item := range catalog {
		if strings.Contains(normalized, strings.ToLower(item.Name)) {
			return item.Code
		}
	}
	return ""
}

func uniqueStrings(items []string) []string {
	out := make([]string, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func firstN(items []string, n int) []string {
	if n <= 0 || len(items) == 0 {
		return []string{}
	}
	if len(items) <= n {
		return append([]string{}, items...)
	}
	return append([]string{}, items[:n]...)
}

func truncateRunes(raw string, max int) string {
	runes := []rune(strings.TrimSpace(raw))
	if len(runes) <= max {
		return string(runes)
	}
	return string(runes[:max])
}

func hashText(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func roundWeight(v float64) float64 {
	if v < 0 {
		return 0
	}
	return float64(int(v*100+0.5)) / 100
}
