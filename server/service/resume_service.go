package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"your-project/config"
	"your-project/model"
	"your-project/pkg/llm"
	promptpkg "your-project/pkg/prompt"
	"your-project/repository"
)

const (
	resumeMinLength = 60
	resumeMaxLength = 32000
)

type ResumeService interface {
	AnalyzeOnly(ctx context.Context, input ResumeAnalysisInput) (*ResumeAnalysisResult, error)
	AnalyzeAndPersist(ctx context.Context, input ResumeAnalysisInput) (*ResumeAnalysisResult, *model.ResumeParseResult, error)
	GetLatestAnalysis(ctx context.Context, userID uint) (*ResumeAnalysisResult, *model.ResumeParseResult, error)
}

type ResumeAnalysisInput struct {
	UserID   uint
	FileName string
	RawText  string
	Source   string
}

type ResumeProfile struct {
	Name            string `json:"name"`
	Email           string `json:"email,omitempty"`
	Phone           string `json:"phone,omitempty"`
	YearsExperience string `json:"years_experience,omitempty"`
	Education       string `json:"education,omitempty"`
	TargetPosition  string `json:"target_position,omitempty"`
	Summary         string `json:"summary,omitempty"`
}

type ResumeSkill struct {
	Name     string `json:"name"`
	Level    string `json:"level,omitempty"`
	Evidence string `json:"evidence,omitempty"`
}

type ResumeProject struct {
	Name        string   `json:"name"`
	Role        string   `json:"role,omitempty"`
	Description string   `json:"description,omitempty"`
	TechStack   []string `json:"tech_stack,omitempty"`
	Highlights  []string `json:"highlights,omitempty"`
}

type ResumePositionFit struct {
	PositionCode string   `json:"position_code"`
	PositionName string   `json:"position_name"`
	Score        int      `json:"score"`
	Reasons      []string `json:"reasons,omitempty"`
}

type ResumeAnalysisResult struct {
	Profile                ResumeProfile       `json:"profile"`
	Skills                 []ResumeSkill       `json:"skills"`
	Projects               []ResumeProject     `json:"projects"`
	StrengthHighlights     []string            `json:"strength_highlights"`
	MissingSkills          []string            `json:"missing_skills"`
	SuggestedPositions     []ResumePositionFit `json:"suggested_positions"`
	RecommendedQuestionIDs []uint              `json:"recommended_question_ids"`
	ConfidenceScore        int                 `json:"confidence_score"`
	ModelVersion           string              `json:"model_version"`
}

type resumeExtractionPayload struct {
	Profile            ResumeProfile   `json:"profile"`
	Skills             []ResumeSkill   `json:"skills"`
	Projects           []ResumeProject `json:"projects"`
	StrengthHighlights []string        `json:"strength_highlights"`
	PotentialGaps      []string        `json:"potential_gaps"`
	ConfidenceScore    int             `json:"confidence_score"`
}

type resumeFitPayload struct {
	SuggestedPositions []ResumePositionFit `json:"suggested_positions"`
	MissingSkills      []string            `json:"missing_skills"`
	StrengthHighlights []string            `json:"strength_highlights"`
	ConfidenceScore    int                 `json:"confidence_score"`
}

type LLMResumeService struct {
	llmClient    llm.LLMClient
	promptMgr    *promptpkg.PromptManager
	positionRepo repository.PositionRepository
	questionRepo repository.QuestionRepository
	resumeRepo   repository.ResumeParseResultRepository
}

var _ ResumeService = (*LLMResumeService)(nil)

func NewResumeService() ResumeService {
	pm, _ := promptpkg.NewPromptManager()
	return NewResumeServiceWithDeps(
		llm.NewDeepSeekClient(config.GetConfig()),
		pm,
		repository.NewPositionRepository(),
		repository.NewQuestionRepository(),
		repository.NewResumeParseResultRepository(),
	)
}

func NewResumeServiceWithDeps(
	llmClient llm.LLMClient,
	promptMgr *promptpkg.PromptManager,
	positionRepo repository.PositionRepository,
	questionRepo repository.QuestionRepository,
	resumeRepo repository.ResumeParseResultRepository,
) ResumeService {
	if llmClient == nil {
		llmClient = llm.NewDeepSeekClient(config.GetConfig())
	}
	if promptMgr == nil {
		promptMgr, _ = promptpkg.NewPromptManager()
	}
	if positionRepo == nil {
		positionRepo = repository.NewPositionRepository()
	}
	if questionRepo == nil {
		questionRepo = repository.NewQuestionRepository()
	}
	if resumeRepo == nil {
		resumeRepo = repository.NewResumeParseResultRepository()
	}

	return &LLMResumeService{
		llmClient:    llmClient,
		promptMgr:    promptMgr,
		positionRepo: positionRepo,
		questionRepo: questionRepo,
		resumeRepo:   resumeRepo,
	}
}

func (s *LLMResumeService) AnalyzeOnly(ctx context.Context, input ResumeAnalysisInput) (*ResumeAnalysisResult, error) {
	raw := normalizeResumeText(input.RawText)
	if len([]rune(raw)) < resumeMinLength {
		return nil, fmt.Errorf("resume text is too short")
	}

	positions, err := s.positionRepo.ListActive()
	if err != nil || len(positions) == 0 {
		positions = append([]model.JobPosition{}, model.DefaultJobPositions...)
	}

	extracted, err := s.extractResume(ctx, raw)
	if err != nil {
		extracted = fallbackResumeExtraction(raw)
	}

	fit, fitErr := s.analyzePositionFit(ctx, raw, extracted, positions)
	if fitErr != nil {
		fit = fallbackResumeFit(extracted, positions)
	}

	result := mergeResumeAnalysis(extracted, fit)
	recommended, _ := s.recommendQuestions(result, positions)
	result.RecommendedQuestionIDs = recommended
	result.ModelVersion = "resume-v2"

	return result, nil
}

func (s *LLMResumeService) AnalyzeAndPersist(ctx context.Context, input ResumeAnalysisInput) (*ResumeAnalysisResult, *model.ResumeParseResult, error) {
	result, err := s.AnalyzeOnly(ctx, input)
	if err != nil {
		return nil, nil, err
	}

	persisted, err := s.persistResult(input, result)
	if err != nil {
		return nil, nil, err
	}

	return result, persisted, nil
}

func (s *LLMResumeService) GetLatestAnalysis(ctx context.Context, userID uint) (*ResumeAnalysisResult, *model.ResumeParseResult, error) {
	_ = ctx

	record, err := s.resumeRepo.GetLatestByUser(userID)
	if err != nil {
		return nil, nil, err
	}

	parsed := &ResumeAnalysisResult{}
	if err := parseStrictJSON(record.StructuredJSON, parsed); err != nil {
		return nil, nil, fmt.Errorf("invalid structured resume payload: %w", err)
	}
	return parsed, record, nil
}

func (s *LLMResumeService) extractResume(ctx context.Context, rawText string) (*resumeExtractionPayload, error) {
	prompt := s.renderPrompt(
		"resume/extract_structured.tmpl",
		map[string]interface{}{"RawText": rawText},
		buildResumeExtractionPromptFallback(rawText),
	)

	resp, err := s.chatJSON(ctx, prompt, "resume")
	if err != nil {
		return nil, err
	}

	payload := &resumeExtractionPayload{}
	if err := parseStrictJSON(resp, payload); err != nil {
		return nil, err
	}
	payload.StrengthHighlights = uniqueNonEmpty(payload.StrengthHighlights)
	payload.PotentialGaps = uniqueNonEmpty(payload.PotentialGaps)
	payload.Skills = normalizeSkillList(payload.Skills)
	payload.Projects = normalizeProjectList(payload.Projects)
	payload.ConfidenceScore = clampResumeScore(payload.ConfidenceScore)
	return payload, nil
}

func (s *LLMResumeService) analyzePositionFit(ctx context.Context, rawText string, extracted *resumeExtractionPayload, positions []model.JobPosition) (*resumeFitPayload, error) {
	extractedJSON, _ := json.Marshal(extracted)
	positionsJSON, _ := json.Marshal(positions)

	prompt := s.renderPrompt(
		"resume/analyze_fit.tmpl",
		map[string]interface{}{
			"RawText":       rawText,
			"ExtractedJSON": string(extractedJSON),
			"PositionsJSON": string(positionsJSON),
		},
		buildResumeFitPromptFallback(rawText, string(extractedJSON), string(positionsJSON)),
	)

	resp, err := s.chatJSON(ctx, prompt, "resume")
	if err != nil {
		return nil, err
	}

	payload := &resumeFitPayload{}
	if err := parseStrictJSON(resp, payload); err != nil {
		return nil, err
	}

	payload.MissingSkills = uniqueNonEmpty(payload.MissingSkills)
	payload.StrengthHighlights = uniqueNonEmpty(payload.StrengthHighlights)
	payload.SuggestedPositions = normalizePositionFit(payload.SuggestedPositions, positions)
	payload.ConfidenceScore = clampResumeScore(payload.ConfidenceScore)

	return payload, nil
}

func (s *LLMResumeService) recommendQuestions(result *ResumeAnalysisResult, positions []model.JobPosition) ([]uint, error) {
	if result == nil {
		return []uint{}, nil
	}

	targetCode := ""
	targetName := ""
	if len(result.SuggestedPositions) > 0 {
		targetCode = strings.TrimSpace(result.SuggestedPositions[0].PositionCode)
		targetName = strings.TrimSpace(result.SuggestedPositions[0].PositionName)
	}
	if targetName == "" {
		targetName = resolvePositionName(targetCode, positions)
	}
	if targetName == "" {
		targetName = "Java后端工程师"
	}

	questionIDs := make([]uint, 0, 12)
	seen := map[uint]struct{}{}
	appendQuestions := func(list []*model.Question) {
		for _, q := range list {
			if q == nil || q.ID == 0 {
				continue
			}
			if _, ok := seen[q.ID]; ok {
				continue
			}
			seen[q.ID] = struct{}{}
			questionIDs = append(questionIDs, q.ID)
			if len(questionIDs) >= 12 {
				return
			}
		}
	}

	knowledgeSeeds := make([]string, 0, len(result.MissingSkills)+len(result.Skills))
	knowledgeSeeds = append(knowledgeSeeds, result.MissingSkills...)
	for _, skill := range result.Skills {
		if len(knowledgeSeeds) >= 8 {
			break
		}
		knowledgeSeeds = append(knowledgeSeeds, skill.Name)
	}

	for _, seed := range uniqueNonEmpty(knowledgeSeeds) {
		if len(questionIDs) >= 12 {
			break
		}
		list, err := s.questionRepo.FindByKnowledgePoint(targetName, "", seed, 4)
		if err != nil {
			continue
		}
		appendQuestions(list)
	}

	if len(questionIDs) < 12 {
		fallback, err := s.questionRepo.ListRandomByPosition(targetName, 16)
		if err == nil {
			appendQuestions(fallback)
		}
	}

	return questionIDs, nil
}

func (s *LLMResumeService) persistResult(input ResumeAnalysisInput, result *ResumeAnalysisResult) (*model.ResumeParseResult, error) {
	structuredJSONBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("marshal structured payload failed: %w", err)
	}

	targetCode := ""
	targetName := ""
	if len(result.SuggestedPositions) > 0 {
		targetCode = strings.TrimSpace(result.SuggestedPositions[0].PositionCode)
		targetName = strings.TrimSpace(result.SuggestedPositions[0].PositionName)
	}

	questionIDs := make([]string, 0, len(result.RecommendedQuestionIDs))
	for _, id := range result.RecommendedQuestionIDs {
		if id == 0 {
			continue
		}
		questionIDs = append(questionIDs, strconv.FormatUint(uint64(id), 10))
	}

	knowledgePayload, _ := json.Marshal(map[string]interface{}{
		"missing_skills":      result.MissingSkills,
		"strength_highlights": result.StrengthHighlights,
	})

	rawText := normalizeResumeText(input.RawText)
	record := &model.ResumeParseResult{
		UserID:               input.UserID,
		FileName:             strings.TrimSpace(input.FileName),
		FileHash:             hashText(rawText),
		RawText:              rawText,
		StructuredJSON:       string(structuredJSONBytes),
		MatchedPositionCode:  targetCode,
		MatchedPositionName:  targetName,
		MatchedQuestionIDs:   strings.Join(questionIDs, ","),
		MatchedKnowledgeJSON: string(knowledgePayload),
		ConfidenceScore:      clampResumeScore(result.ConfidenceScore),
		ParserVersion:        result.ModelVersion,
		Source:               strings.TrimSpace(input.Source),
	}
	if record.Source == "" {
		record.Source = "upload"
	}

	if err := s.resumeRepo.Create(record); err != nil {
		return nil, err
	}
	return record, nil
}

func (s *LLMResumeService) chatJSON(ctx context.Context, prompt string, taskType string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if s.llmClient == nil {
		return "", fmt.Errorf("llm client is not initialized")
	}

	request := llm.ChatRequest{
		Model:          resolveModelForTask(taskType),
		Messages:       []llm.ChatMessage{{Role: "user", Content: prompt}},
		Temperature:    0.15,
		ResponseFormat: &llm.ResponseFormat{Type: llm.ResponseFormatJSON},
	}
	return s.llmClient.Chat(ctx, request)
}

func (s *LLMResumeService) renderPrompt(templateName string, data interface{}, fallback string) string {
	if s.promptMgr == nil {
		return fallback
	}
	rendered, err := s.promptMgr.Render(templateName, data)
	if err != nil {
		return fallback
	}
	trimmed := strings.TrimSpace(rendered)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

func resolveModelForTask(taskType string) string {
	cfg := config.GetConfig()
	if cfg == nil {
		return config.DefaultDeepSeekModel
	}
	if model := strings.TrimSpace(cfg.LLM.Models[taskType]); model != "" {
		return model
	}
	if model := strings.TrimSpace(cfg.LLM.Model); model != "" {
		return model
	}
	return config.DefaultDeepSeekModel
}

func mergeResumeAnalysis(extracted *resumeExtractionPayload, fit *resumeFitPayload) *ResumeAnalysisResult {
	result := &ResumeAnalysisResult{
		Profile:            extracted.Profile,
		Skills:             normalizeSkillList(extracted.Skills),
		Projects:           normalizeProjectList(extracted.Projects),
		StrengthHighlights: uniqueNonEmpty(append([]string{}, extracted.StrengthHighlights...)),
		MissingSkills:      uniqueNonEmpty(append([]string{}, extracted.PotentialGaps...)),
		SuggestedPositions: append([]ResumePositionFit{}, fit.SuggestedPositions...),
		ConfidenceScore:    clampResumeScore(maxInt(extracted.ConfidenceScore, fit.ConfidenceScore)),
	}
	result.StrengthHighlights = uniqueNonEmpty(append(result.StrengthHighlights, fit.StrengthHighlights...))
	result.MissingSkills = uniqueNonEmpty(append(result.MissingSkills, fit.MissingSkills...))
	return result
}

func fallbackResumeExtraction(rawText string) *resumeExtractionPayload {
	skills := inferSkillsFromText(rawText)
	profile := ResumeProfile{
		Summary: truncateString(rawText, 240),
	}
	return &resumeExtractionPayload{
		Profile:            profile,
		Skills:             skills,
		Projects:           []ResumeProject{},
		StrengthHighlights: firstNStrings(extractSentences(rawText), 4),
		PotentialGaps:      []string{},
		ConfidenceScore:    55,
	}
}

func fallbackResumeFit(extracted *resumeExtractionPayload, positions []model.JobPosition) *resumeFitPayload {
	scoreByCode := map[string]int{}
	for _, pos := range positions {
		scoreByCode[pos.Code] = 50
	}

	for _, skill := range extracted.Skills {
		low := strings.ToLower(skill.Name)
		switch {
		case strings.Contains(low, "go"), strings.Contains(low, "java"), strings.Contains(low, "mysql"), strings.Contains(low, "redis"):
			scoreByCode[string(model.PositionBackend)] += 12
		case strings.Contains(low, "vue"), strings.Contains(low, "react"), strings.Contains(low, "typescript"), strings.Contains(low, "css"):
			scoreByCode[string(model.PositionFrontend)] += 12
		case strings.Contains(low, "algorithm"), strings.Contains(low, "leetcode"), strings.Contains(low, "复杂度"):
			scoreByCode[string(model.PositionAlgorithm)] += 14
		case strings.Contains(low, "llm"), strings.Contains(low, "pytorch"), strings.Contains(low, "tensorflow"), strings.Contains(low, "machine learning"):
			scoreByCode[string(model.PositionAI)] += 14
		}
	}

	list := make([]ResumePositionFit, 0, len(positions))
	for _, pos := range positions {
		score := clampResumeScore(scoreByCode[pos.Code])
		list = append(list, ResumePositionFit{
			PositionCode: pos.Code,
			PositionName: pos.Name,
			Score:        score,
			Reasons:      []string{"rule-based fallback scoring"},
		})
	}
	sort.SliceStable(list, func(i, j int) bool { return list[i].Score > list[j].Score })

	return &resumeFitPayload{
		SuggestedPositions: list,
		MissingSkills:      []string{},
		StrengthHighlights: uniqueNonEmpty(extracted.StrengthHighlights),
		ConfidenceScore:    60,
	}
}

func normalizePositionFit(items []ResumePositionFit, positions []model.JobPosition) []ResumePositionFit {
	nameByCode := map[string]string{}
	for _, pos := range positions {
		nameByCode[strings.TrimSpace(pos.Code)] = strings.TrimSpace(pos.Name)
	}

	out := make([]ResumePositionFit, 0, len(items))
	for _, item := range items {
		code := strings.TrimSpace(item.PositionCode)
		name := strings.TrimSpace(item.PositionName)
		if code == "" && name != "" {
			code = inferPositionCodeByName(name)
		}
		if name == "" {
			name = nameByCode[code]
		}
		if code == "" {
			continue
		}
		out = append(out, ResumePositionFit{
			PositionCode: code,
			PositionName: name,
			Score:        clampResumeScore(item.Score),
			Reasons:      uniqueNonEmpty(item.Reasons),
		})
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			return out[i].PositionCode < out[j].PositionCode
		}
		return out[i].Score > out[j].Score
	})

	return out
}

func normalizeSkillList(skills []ResumeSkill) []ResumeSkill {
	out := make([]ResumeSkill, 0, len(skills))
	seen := map[string]struct{}{}
	for _, skill := range skills {
		name := strings.TrimSpace(skill.Name)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, ResumeSkill{
			Name:     name,
			Level:    strings.TrimSpace(skill.Level),
			Evidence: strings.TrimSpace(skill.Evidence),
		})
	}
	return out
}

func normalizeProjectList(projects []ResumeProject) []ResumeProject {
	out := make([]ResumeProject, 0, len(projects))
	for _, project := range projects {
		name := strings.TrimSpace(project.Name)
		desc := strings.TrimSpace(project.Description)
		if name == "" && desc == "" {
			continue
		}
		out = append(out, ResumeProject{
			Name:        name,
			Role:        strings.TrimSpace(project.Role),
			Description: desc,
			TechStack:   uniqueNonEmpty(project.TechStack),
			Highlights:  uniqueNonEmpty(project.Highlights),
		})
	}
	return out
}

func parseStrictJSON(raw string, dest interface{}) error {
	cleaned := strings.TrimSpace(raw)
	if cleaned == "" {
		return fmt.Errorf("empty json payload")
	}
	if strings.HasPrefix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```json")
		cleaned = strings.TrimPrefix(cleaned, "```")
		cleaned = strings.TrimSuffix(cleaned, "```")
		cleaned = strings.TrimSpace(cleaned)
	}
	if start := strings.IndexAny(cleaned, "{["); start >= 0 {
		if end := strings.LastIndexAny(cleaned, "}]"); end > start {
			cleaned = cleaned[start : end+1]
		}
	}
	decoder := json.NewDecoder(strings.NewReader(cleaned))
	decoder.DisallowUnknownFields()
	return decoder.Decode(dest)
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

func hashText(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func clampResumeScore(score int) int {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

func maxInt(a, b int) int {
	if a >= b {
		return a
	}
	return b
}

func uniqueNonEmpty(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
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

func resolvePositionName(code string, positions []model.JobPosition) string {
	for _, pos := range positions {
		if strings.EqualFold(strings.TrimSpace(pos.Code), strings.TrimSpace(code)) {
			return pos.Name
		}
	}
	for _, pos := range model.DefaultJobPositions {
		if strings.EqualFold(strings.TrimSpace(pos.Code), strings.TrimSpace(code)) {
			return pos.Name
		}
	}
	return ""
}

func inferPositionCodeByName(name string) string {
	low := strings.ToLower(strings.TrimSpace(name))
	switch {
	case strings.Contains(low, "前端"), strings.Contains(low, "frontend"):
		return string(model.PositionFrontend)
	case strings.Contains(low, "算法"), strings.Contains(low, "algorithm"):
		return string(model.PositionAlgorithm)
	case strings.Contains(low, "ai"), strings.Contains(low, "模型"), strings.Contains(low, "machine"):
		return string(model.PositionAI)
	default:
		return string(model.PositionBackend)
	}
}

func inferSkillsFromText(raw string) []ResumeSkill {
	candidates := []string{
		"Go", "Java", "Python", "MySQL", "Redis", "Kafka", "Docker", "Kubernetes",
		"Vue", "React", "TypeScript", "Node.js", "CSS", "HTML",
		"Machine Learning", "Deep Learning", "PyTorch", "TensorFlow", "LLM",
	}
	low := strings.ToLower(raw)
	out := make([]ResumeSkill, 0, 8)
	for _, candidate := range candidates {
		if strings.Contains(low, strings.ToLower(candidate)) {
			out = append(out, ResumeSkill{Name: candidate, Level: "unknown", Evidence: "text matched"})
		}
	}
	return out
}

func extractSentences(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		switch r {
		case '\n', '\r', '。', '.', ';', '；', '!', '！', '?', '？':
			return true
		default:
			return false
		}
	})
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

func firstNStrings(items []string, n int) []string {
	if n <= 0 || len(items) == 0 {
		return []string{}
	}
	if len(items) <= n {
		return uniqueNonEmpty(items)
	}
	return uniqueNonEmpty(items[:n])
}

func truncateString(s string, max int) string {
	runes := []rune(strings.TrimSpace(s))
	if len(runes) <= max {
		return string(runes)
	}
	return strings.TrimSpace(string(runes[:max]))
}

func buildResumeExtractionPromptFallback(rawText string) string {
	return "你是一名资深招聘顾问。请从简历文本中提取结构化信息，只输出 JSON。\n" +
		"字段: profile{name,email,phone,years_experience,education,target_position,summary}, skills[{name,level,evidence}], projects[{name,role,description,tech_stack,highlights}], strength_highlights[], potential_gaps[], confidence_score(0-100)。\n" +
		"必须基于原文证据，不得编造。\n简历原文:\n" + rawText
}

func buildResumeFitPromptFallback(rawText, extractedJSON, positionsJSON string) string {
	return "你是一名面试题库调度专家。请根据简历与岗位列表给出岗位匹配分析，只输出 JSON。\n" +
		"字段: suggested_positions[{position_code,position_name,score,reasons}], missing_skills[], strength_highlights[], confidence_score(0-100)。\n" +
		"position_code 必须来自岗位列表。\n简历原文:\n" + rawText +
		"\n结构化简历:\n" + extractedJSON +
		"\n岗位列表:\n" + positionsJSON
}
