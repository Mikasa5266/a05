package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"your-project/model"
	"your-project/repository"

	"gorm.io/gorm"
)

type ResumeService struct {
	aiService          *AIService
	lastPipelineResult *ResumePipelineResult
}

type ResumePipelineResult struct {
	Validation *model.ResumeValidationResult
	Extracted  *model.ResumeExtractedData
	Match      *model.ResumeMatchResult
	Resume     *model.ResumeData
	Matches    []*model.JobMatch
}

func NewResumeService() *ResumeService {
	return &ResumeService{
		aiService: MustGetAIService(),
	}
}

func (s *ResumeService) ParseResume(fileContent string) (*model.ResumeData, error) {
	pipelineResult, err := s.AnalyzeResume(fileContent)
	if err != nil {
		return nil, err
	}
	s.lastPipelineResult = pipelineResult
	return pipelineResult.Resume, nil
}

func (s *ResumeService) AnalyzeResume(fileContent string) (*ResumePipelineResult, error) {
	start := time.Now()
	log.Printf("resume pipeline started, raw content length=%d", len(fileContent))

	const maxContentLength = 50000
	if len(fileContent) > maxContentLength {
		log.Printf("resume pipeline content truncated from %d to %d", len(fileContent), maxContentLength)
		fileContent = fileContent[:maxContentLength] + "\n...(content truncated)..."
	}

	ctx := context.Background()

	validatorStart := time.Now()
	validatorPrompt, err := s.aiService.RenderPrompt("resume/01_validator.tmpl", map[string]interface{}{
		"RawText": fileContent,
	})
	if err != nil {
		return nil, fmt.Errorf("render validator prompt failed: %w", err)
	}
	validatorResp, err := s.aiService.ChatWithTask(ctx, validatorPrompt, "resume")
	if err != nil {
		return nil, fmt.Errorf("validator llm call failed: %w", err)
	}
	validatorJSON := CleanJSON(validatorResp)
	var validation model.ResumeValidationResult
	if err := json.Unmarshal([]byte(validatorJSON), &validation); err != nil {
		log.Printf("validator unmarshal failed, response=%s", validatorJSON)
		return nil, fmt.Errorf("validator json unmarshal failed: %w", err)
	}
	log.Printf("resume validator completed in %s, is_resume=%v, confidence_score=%d", time.Since(validatorStart), validation.IsResume, validation.ConfidenceScore)

	if !validation.IsResume || validation.ConfidenceScore < 60 {
		rejectReason := strings.TrimSpace(validation.RejectReason)
		if rejectReason == "" {
			if !validation.IsResume {
				rejectReason = "输入内容不是可用于岗位匹配的简历文本"
			} else {
				rejectReason = "简历置信度不足，无法继续解析"
			}
		}
		log.Printf("resume pipeline rejected, reason=%s", rejectReason)
		return nil, fmt.Errorf("%s", rejectReason)
	}

	extractorStart := time.Now()
	extractorPrompt, err := s.aiService.RenderPrompt("resume/02_extractor.tmpl", map[string]interface{}{
		"ValidatedResumeText": fileContent,
	})
	if err != nil {
		return nil, fmt.Errorf("render extractor prompt failed: %w", err)
	}
	extractorResp, err := s.aiService.ChatWithTask(ctx, extractorPrompt, "resume")
	if err != nil {
		return nil, fmt.Errorf("extractor llm call failed: %w", err)
	}
	extractorJSON := CleanJSON(extractorResp)
	var extracted model.ResumeExtractedData
	if err := json.Unmarshal([]byte(extractorJSON), &extracted); err != nil {
		log.Printf("extractor unmarshal failed, response=%s", extractorJSON)
		return nil, fmt.Errorf("extractor json unmarshal failed: %w", err)
	}
	log.Printf("resume extractor completed in %s, skills=%d, projects=%d", time.Since(extractorStart), len(extracted.CoreSkills), len(extracted.ProjectHighlights))

	matcherStart := time.Now()
	matchResult, err := s.runMatcher(ctx, &extracted)
	if err != nil {
		return nil, err
	}
	log.Printf("resume matcher completed in %s, matched_roles=%d, question_banks=%d", time.Since(matcherStart), len(matchResult.MatchedRoles), len(matchResult.TargetQuestionBanks))

	resumeData := convertExtractedToResumeData(&extracted)
	matches := convertMatchResultToJobMatches(matchResult)

	log.Printf("resume pipeline completed in %s, legacy_resume_experience=%d, legacy_matches=%d", time.Since(start), len(resumeData.Experience), len(matches))

	return &ResumePipelineResult{
		Validation: &validation,
		Extracted:  &extracted,
		Match:      matchResult,
		Resume:     resumeData,
		Matches:    matches,
	}, nil
}

func (s *ResumeService) MatchJobs(resumeData *model.ResumeData) ([]*model.JobMatch, error) {
	if resumeData == nil {
		return nil, fmt.Errorf("resume data is required")
	}

	if s.lastPipelineResult != nil && s.lastPipelineResult.Resume != nil && isSameResumeData(resumeData, s.lastPipelineResult.Resume) {
		log.Printf("resume matcher returned cached pipeline result, matches=%d", len(s.lastPipelineResult.Matches))
		return s.lastPipelineResult.Matches, nil
	}

	start := time.Now()
	log.Printf("resume matcher started from legacy resume input, techStack=%v", resumeData.TechStack)
	extracted := convertResumeDataToExtracted(resumeData)
	matchResult, err := s.runMatcher(context.Background(), extracted)
	if err != nil {
		return nil, err
	}
	matches := convertMatchResultToJobMatches(matchResult)
	log.Printf("resume matcher completed in %s from legacy resume input, matches=%d", time.Since(start), len(matches))
	return matches, nil
}

func (s *ResumeService) runMatcher(ctx context.Context, extracted *model.ResumeExtractedData) (*model.ResumeMatchResult, error) {
	extractedJSONBytes, err := json.Marshal(extracted)
	if err != nil {
		return nil, fmt.Errorf("marshal extracted resume failed: %w", err)
	}
	roles := s.loadAvailableRoles()
	questionBanks := s.loadAvailableQuestionBanks()
	rolesJSONBytes, err := json.Marshal(roles)
	if err != nil {
		return nil, fmt.Errorf("marshal available roles failed: %w", err)
	}
	questionBanksJSONBytes, err := json.Marshal(questionBanks)
	if err != nil {
		return nil, fmt.Errorf("marshal available question banks failed: %w", err)
	}

	matcherPrompt, err := s.aiService.RenderPrompt("resume/03_matcher.tmpl", map[string]interface{}{
		"ExtractedResumeJSON":    string(extractedJSONBytes),
		"AvailableRoles":         string(rolesJSONBytes),
		"AvailableQuestionBanks": string(questionBanksJSONBytes),
	})
	if err != nil {
		return nil, fmt.Errorf("render matcher prompt failed: %w", err)
	}

	matcherResp, err := s.aiService.ChatWithTask(ctx, matcherPrompt, "resume")
	if err != nil {
		return nil, fmt.Errorf("matcher llm call failed: %w", err)
	}
	matcherJSON := CleanJSON(matcherResp)
	var matchResult model.ResumeMatchResult
	if err := json.Unmarshal([]byte(matcherJSON), &matchResult); err != nil {
		log.Printf("matcher unmarshal failed, response=%s", matcherJSON)
		return nil, fmt.Errorf("matcher json unmarshal failed: %w", err)
	}
	return &matchResult, nil
}

func (s *ResumeService) loadAvailableRoles() []string {
	roles := make([]string, 0)
	db := getDBSafe()
	if db == nil {
		return roles
	}
	if err := db.Model(&model.Question{}).
		Where("position IS NOT NULL AND position <> ''").
		Distinct().
		Pluck("position", &roles).Error; err != nil {
		log.Printf("load available roles failed: %v", err)
		return []string{}
	}
	return uniqueSortedNonEmpty(roles)
}

func (s *ResumeService) loadAvailableQuestionBanks() []string {
	questionBanks := make([]string, 0)
	db := getDBSafe()
	if db == nil {
		return questionBanks
	}
	if err := db.Model(&model.Question{}).
		Where("category IS NOT NULL AND category <> ''").
		Distinct().
		Pluck("category", &questionBanks).Error; err != nil {
		log.Printf("load available question banks failed: %v", err)
		return []string{}
	}
	return uniqueSortedNonEmpty(questionBanks)
}

func getDBSafe() (db *gorm.DB) {
	defer func() {
		if recover() != nil {
			db = nil
		}
	}()
	return repository.GetDB()
}

func convertExtractedToResumeData(extracted *model.ResumeExtractedData) *model.ResumeData {
	if extracted == nil {
		return &model.ResumeData{}
	}
	experience := make([]struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Highlights  []string `json:"highlights"`
	}, 0, len(extracted.ProjectHighlights))
	for _, project := range extracted.ProjectHighlights {
		experience = append(experience, struct {
			Title       string   `json:"title"`
			Description string   `json:"description"`
			Highlights  []string `json:"highlights"`
		}{
			Title:       strings.TrimSpace(project.Name),
			Description: strings.TrimSpace(project.CoreContribution),
			Highlights:  uniqueSortedNonEmpty(project.TechStack),
		})
	}
	return &model.ResumeData{
		TechStack:  uniqueSortedNonEmpty(extracted.CoreSkills),
		Experience: experience,
		Intent:     strings.TrimSpace(extracted.BasicInfo.TargetDirection),
		SoftSkills: []string{},
	}
}

func convertResumeDataToExtracted(resumeData *model.ResumeData) *model.ResumeExtractedData {
	if resumeData == nil {
		return &model.ResumeExtractedData{}
	}
	projects := make([]model.ResumeProjectHighlight, 0, len(resumeData.Experience))
	for _, exp := range resumeData.Experience {
		projects = append(projects, model.ResumeProjectHighlight{
			Name:             strings.TrimSpace(exp.Title),
			TechStack:        uniqueSortedNonEmpty(exp.Highlights),
			CoreContribution: strings.TrimSpace(exp.Description),
		})
	}
	return &model.ResumeExtractedData{
		BasicInfo: model.ResumeBasicInfo{
			TargetDirection: strings.TrimSpace(resumeData.Intent),
		},
		CoreSkills:        uniqueSortedNonEmpty(resumeData.TechStack),
		ProjectHighlights: projects,
	}
}

func convertMatchResultToJobMatches(matchResult *model.ResumeMatchResult) []*model.JobMatch {
	if matchResult == nil {
		return []*model.JobMatch{}
	}
	matches := make([]*model.JobMatch, 0, len(matchResult.MatchedRoles))
	requirements := uniqueSortedNonEmpty(matchResult.TargetQuestionBanks)
	for _, role := range matchResult.MatchedRoles {
		matches = append(matches, &model.JobMatch{
			JobTitle:     strings.TrimSpace(role.RoleName),
			MatchScore:   role.MatchDegree,
			Reason:       strings.TrimSpace(role.Reason),
			Requirements: requirements,
		})
	}
	return matches
}

func uniqueSortedNonEmpty(values []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	sort.Strings(result)
	return result
}

func isSameResumeData(left, right *model.ResumeData) bool {
	if left == nil || right == nil {
		return false
	}
	leftJSON, err := json.Marshal(left)
	if err != nil {
		return false
	}
	rightJSON, err := json.Marshal(right)
	if err != nil {
		return false
	}
	return string(leftJSON) == string(rightJSON)
}

// GenerateInterviewQuestions generates personalized questions based on resume and job title
func (s *ResumeService) GenerateInterviewQuestions(resumeData *model.ResumeData, jobTitle string) (map[string][]string, error) {
	resumeJson, _ := json.Marshal(resumeData)

	// 1. Get RAG context for the job title
	ragSvc := GetRAGService()
	var ragContext string
	if ragSvc != nil {
		context, err := ragSvc.SearchKnowledgeBase(jobTitle)
		if err == nil && context != "" {
			ragContext = context
		}
	}

	prompt := fmt.Sprintf(`
你是一位资深技术面试官。请根据候选人的简历和目标岗位，生成一份个性化的面试题库。

【岗位知识库上下文】
%s

【候选人简历】
%s

【目标岗位】
%s

【生成要求】
1. **深挖追问题库** (3题)：针对简历中的项目经历，设计能考察深度和真实性的追问（例如：“在这个项目中你提到的高并发优化，具体是采用了什么策略？有什么数据支撑？”）。
2. **岗位高频考点题库** (3题)：基于目标岗位的能力模型，生成该岗位面试中高频出现的核心技术问题。
3. **基础补漏题库** (3题)：基于简历中技术栈的薄弱点或未提及但该岗位必备的基础知识（例如：如果简历没写并发，就问并发基础）。

【输出格式】
请直接返回 JSON 对象，不要包含 Markdown 标记：
{
  "deep_dive": ["问题1", "问题2", "问题3"],
  "high_freq": ["问题1", "问题2", "问题3"],
  "basic_check": ["问题1", "问题2", "问题3"]
}
`, ragContext, string(resumeJson), jobTitle)

	log.Printf("Generating interview questions for job: %s", jobTitle)

	resp, err := s.aiService.ChatWithTask(context.Background(), prompt, "chat") // Use chat model for generation
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	jsonStr := CleanJSON(resp)
	var result map[string][]string
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		log.Printf("Failed to parse AI response: %v, response: %s", err, jsonStr)
		return nil, fmt.Errorf("failed to parse AI response: %w, response: %s", err, jsonStr)
	}

	return result, nil
}

// AnalyzeAuthenticity evaluates claim credibility from resume content and returns a risk report.
func (s *ResumeService) AnalyzeAuthenticity(resumeData *model.ResumeData, rawText string, targetRole string) (*model.ResumeAuthenticityReport, error) {
	resumeJSON, _ := json.Marshal(resumeData)

	prompt := fmt.Sprintf(`
你是一位招聘风控顾问。请基于候选人的简历结构化信息和原文，做“真实性风险分析”。

【目标岗位】
%s

【结构化简历】
%s

【简历原文】
"""
%s
"""

【重要约束】
1. 必须使用简体中文。
2. 你不能断言候选人“造假”，只能输出“潜在风险”和“核验建议”。
3. 仅根据输入内容分析，不得编造外部信息。
4. 输出必须是 JSON，不要 Markdown。

输出 JSON 结构：
{
  "overallRiskScore": 0,
  "summary": "整体判断",
  "verifiableItems": ["可核验点1"],
  "potentialRiskItems": [
    {
      "claim": "某条经历或数据",
      "riskLevel": "low|medium|high",
      "reason": "为什么存在风险",
      "verificationTip": "如何在面试/背调中核验"
    }
  ],
  "interviewChecks": ["可直接用于面试的核验问题1"],
  "disclaimer": "免责声明"
}
`, targetRole, string(resumeJSON), rawText)

	resp, err := s.aiService.ChatWithTask(context.Background(), prompt, "resume_authenticity")
	if err != nil {
		return nil, fmt.Errorf("authenticity analysis failed: %w", err)
	}

	jsonStr := CleanJSON(resp)
	var result model.ResumeAuthenticityReport
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse authenticity report: %w", err)
	}

	if result.Disclaimer == "" {
		result.Disclaimer = "该分析仅用于面试核验参考，不构成事实认定或法律结论。"
	}

	return &result, nil
}

// GenerateOptimizationSuggestions provides actionable resume improvements and rewrite demos.
func (s *ResumeService) GenerateOptimizationSuggestions(resumeData *model.ResumeData, targetRole string) (*model.ResumeOptimizationReport, error) {
	resumeJSON, _ := json.Marshal(resumeData)

	prompt := fmt.Sprintf(`
你是一位资深简历教练。请根据候选人的简历信息输出“可执行的简历优化建议”。

【目标岗位】
%s

【简历信息】
%s

【要求】
1. 必须使用简体中文。
2. 建议必须具体、可执行，不要空话。
3. 重点关注：项目描述量化、技术深度表达、关键词覆盖、结构层次。
4. 输出必须是 JSON，不要 Markdown。

输出 JSON：
{
  "overallScore": 0,
  "strengths": ["优势1"],
  "weaknesses": ["问题1"],
  "suggestions": ["改进建议1"],
  "rewriteDemo": ["原句 -> 改写句"],
  "keywords": ["岗位关键词1"]
}
`, targetRole, string(resumeJSON))

	resp, err := s.aiService.ChatWithTask(context.Background(), prompt, "resume_optimization")
	if err != nil {
		return nil, fmt.Errorf("resume optimization failed: %w", err)
	}

	jsonStr := CleanJSON(resp)
	var result model.ResumeOptimizationReport
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse optimization report: %w", err)
	}

	return &result, nil
}

// GenerateResumeTemplate creates a role-specific resume template in markdown.
func (s *ResumeService) GenerateResumeTemplate(targetRole string, seniority string, language string) (*model.ResumeTemplate, error) {
	if strings.TrimSpace(language) == "" {
		language = "zh-CN"
	}

	prompt := fmt.Sprintf(`
你是一位招聘专家，请生成一份高质量的简历范本。

【参数】
- 目标岗位: %s
- 经验级别: %s
- 输出语言: %s

【要求】
1. 输出适合该岗位的真实可用模板，结构清晰，包含：个人信息、职业摘要、核心技能、工作经历、项目经历、教育背景、证书/加分项。
2. 工作经历与项目经历要包含“结果导向”的表达示例（含量化指标占位符）。
3. 输出必须是 JSON，不要 Markdown 代码块。

输出 JSON：
{
  "targetRole": "岗位名",
  "templateMarkdown": "完整 Markdown 模板文本",
  "writingGuides": ["撰写建议1"],
  "commonMistakes": ["常见错误1"]
}
`, targetRole, seniority, language)

	resp, err := s.aiService.ChatWithTask(context.Background(), prompt, "resume_template")
	if err != nil {
		return nil, fmt.Errorf("resume template generation failed: %w", err)
	}

	jsonStr := CleanJSON(resp)
	var result model.ResumeTemplate
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse resume template: %w", err)
	}

	if strings.TrimSpace(result.TargetRole) == "" {
		result.TargetRole = targetRole
	}

	return &result, nil
}

// Helper to clean markdown code blocks if AI returns them
func CleanJSON(s string) string {
	s = strings.TrimSpace(s)
	// Remove markdown code blocks
	if strings.HasPrefix(s, "```json") {
		s = s[7:]
	} else if strings.HasPrefix(s, "```") {
		s = s[3:]
	}
	if strings.HasSuffix(s, "```") {
		s = s[:len(s)-3]
	}
	s = strings.TrimSpace(s)

	// Ensure we only have JSON content by finding the first '{' or '[' and last '}' or ']'
	firstBrace := strings.IndexAny(s, "{[")
	lastBrace := strings.LastIndexAny(s, "}]")

	if firstBrace != -1 && lastBrace != -1 && lastBrace > firstBrace {
		s = s[firstBrace : lastBrace+1]
	}

	return s
}
