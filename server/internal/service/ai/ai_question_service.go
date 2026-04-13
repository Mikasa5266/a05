package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strings"
	"time"

	"your-project/internal/model"
	"your-project/internal/repository"

	"gorm.io/gorm"
)

func (s *AIService) GenerateQuestions(ctx context.Context, interview *model.Interview, count int) ([]*model.Question, error) {
	if groupQuestion, err := s.tryGenerateGroupOpenQuestion(ctx, interview); err != nil {
		return nil, err
	} else if groupQuestion != nil {
		if count <= 1 {
			return []*model.Question{groupQuestion}, nil
		}

		more, err := s.generateQuestionsDefault(ctx, interview, count-1)
		if err != nil {
			return []*model.Question{groupQuestion}, nil
		}
		questions := make([]*model.Question, 0, len(more)+1)
		questions = append(questions, groupQuestion)
		questions = append(questions, more...)
		return questions, nil
	}

	return s.generateQuestionsDefault(ctx, interview, count)
}

func (s *AIService) generateQuestionsDefault(ctx context.Context, interview *model.Interview, count int) ([]*model.Question, error) {
	prompt, err := s.renderPrompt("generate_questions.tmpl", map[string]interface{}{
		"Count":                 count,
		"Position":              interview.Position,
		"Difficulty":            interview.Difficulty,
		"Mode":                  interview.Mode,
		"Style":                 interview.Style,
		"ModeInstruction":       s.buildModePrompt(interview.Mode),
		"StyleInstruction":      s.buildStylePrompt(interview.Style, interview.Company),
		"DifficultyInstruction": s.buildDifficultyPrompt(interview.Difficulty),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to render generate questions prompt: %w", err)
	}

	response, err := s.chat(ctx, prompt, "chat", jsonObjectResponseFormat())
	if err != nil {
		return nil, fmt.Errorf("failed to generate questions: %w", err)
	}

	var questionsData []struct {
		Title          string `json:"title"`
		Content        string `json:"content"`
		ExpectedAnswer string `json:"expected_answer"`
	}
	if err := json.Unmarshal([]byte(response), &questionsData); err != nil {
		return nil, fmt.Errorf("failed to parse questions response: %w, body: %s", err, response)
	}

	questions := make([]*model.Question, 0, len(questionsData))
	for _, qd := range questionsData {
		item := &model.Question{Title: qd.Title, Content: qd.Content, ExpectedAnswer: qd.ExpectedAnswer, Position: interview.Position, Difficulty: interview.Difficulty}
		s.EnsureQuestionChinese(ctx, item)
		questions = append(questions, item)
	}
	return questions, nil
}

type groupResumeSnapshot struct {
	UserID      uint
	DisplayName string
	Summary     string
	Skills      []string
}

func (s *AIService) tryGenerateGroupOpenQuestion(ctx context.Context, interview *model.Interview) (*model.Question, error) {
	if interview == nil || !interview.IsGroup {
		return nil, nil
	}

	snapshots := s.loadGroupResumeSnapshots(interview)
	questionType := pickGroupQuestionType(snapshots)
	resumeContext := buildGroupResumePromptContext(snapshots)
	if resumeContext == "" {
		resumeContext = "暂无可用简历摘要，请基于协作能力与技术决策能力生成通用群面开放题。"
	}

	prompt := fmt.Sprintf(`你是群面面试官。请基于候选人简历摘要生成 1 道开放题。
要求：
1) 题型必须是“协同冲突解决”或“技术方案评审”。
2) 必须要求多人协作表达，不得是单人八股题。
3) 输出严格 JSON：{"title":"...","content":"...","expected_answer":"...","question_type":"..."}

岗位：%s
难度：%s
推荐题型：%s

群面候选人简历摘要：
%s`, strings.TrimSpace(interview.Position), strings.TrimSpace(interview.Difficulty), questionType, resumeContext)

	raw, err := s.chat(ctx, prompt, "chat", jsonObjectResponseFormat())
	if err != nil {
		fallback := buildGroupQuestionFallback(interview, questionType, snapshots)
		s.EnsureQuestionChinese(ctx, fallback)
		return fallback, nil
	}

	var parsed struct {
		Title          string `json:"title"`
		Content        string `json:"content"`
		ExpectedAnswer string `json:"expected_answer"`
		QuestionType   string `json:"question_type"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		fallback := buildGroupQuestionFallback(interview, questionType, snapshots)
		s.EnsureQuestionChinese(ctx, fallback)
		return fallback, nil
	}

	resolvedType := strings.TrimSpace(parsed.QuestionType)
	if resolvedType == "" {
		resolvedType = questionType
	}
	if !isSupportedGroupQuestionType(resolvedType) {
		resolvedType = questionType
	}

	question := &model.Question{
		Title:          sanitizeGeneratedText(strings.TrimSpace(parsed.Title)),
		Content:        sanitizeGeneratedText(strings.TrimSpace(parsed.Content)),
		ExpectedAnswer: sanitizeGeneratedText(strings.TrimSpace(parsed.ExpectedAnswer)),
		Position:       interview.Position,
		Difficulty:     interview.Difficulty,
		Category:       mapGroupQuestionCategory(resolvedType),
	}

	if question.Title == "" || question.Content == "" {
		question = buildGroupQuestionFallback(interview, resolvedType, snapshots)
	}

	s.EnsureQuestionChinese(ctx, question)
	return question, nil
}

func (s *AIService) loadGroupResumeSnapshots(interview *model.Interview) []groupResumeSnapshot {
	if interview == nil || !interview.IsGroup {
		return nil
	}

	db := repository.GetDB()
	if db == nil {
		return nil
	}

	participantIDs := s.collectGroupParticipantIDs(db, interview)
	if len(participantIDs) == 0 {
		return nil
	}

	snapshots := make([]groupResumeSnapshot, 0, len(participantIDs))
	for _, userID := range participantIDs {
		if userID == 0 {
			continue
		}

		var user model.User
		if err := db.Select("id", "username").First(&user, userID).Error; err != nil {
			continue
		}

		var latest model.ResumeParseResult
		if err := db.Where("user_id = ?", userID).Order("id DESC").First(&latest).Error; err != nil {
			continue
		}

		summary, skills := buildResumeDigestForGroup(&latest)
		if strings.TrimSpace(summary) == "" {
			continue
		}

		snapshots = append(snapshots, groupResumeSnapshot{
			UserID:      userID,
			DisplayName: strings.TrimSpace(user.Username),
			Summary:     summary,
			Skills:      skills,
		})
	}

	return snapshots
}

func (s *AIService) collectGroupParticipantIDs(db *gorm.DB, interview *model.Interview) []uint {
	ids := make([]uint, 0, 6)
	if interview.UserID > 0 {
		ids = append(ids, interview.UserID)
	}

	if interview.GroupInterviewRoom != nil {
		ids = append(ids, interview.GroupInterviewRoom.GetParticipantIDs()...)
		if interview.GroupInterviewRoom.CreatorID > 0 {
			ids = append(ids, interview.GroupInterviewRoom.CreatorID)
		}
	}

	if interview.GroupInterviewRoomID != nil && *interview.GroupInterviewRoomID > 0 {
		var room model.GroupInterviewRoom
		if err := db.First(&room, *interview.GroupInterviewRoomID).Error; err == nil {
			ids = append(ids, room.GetParticipantIDs()...)
			if room.CreatorID > 0 {
				ids = append(ids, room.CreatorID)
			}
		}
	}

	seen := make(map[uint]struct{}, len(ids))
	normalized := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		normalized = append(normalized, id)
		if len(normalized) >= 5 {
			break
		}
	}
	sort.Slice(normalized, func(i, j int) bool { return normalized[i] < normalized[j] })
	return normalized
}

func buildResumeDigestForGroup(record *model.ResumeParseResult) (string, []string) {
	if record == nil {
		return "", nil
	}

	var payload model.ResumeStructuredPayload
	if err := json.Unmarshal([]byte(record.StructuredJSON), &payload); err == nil {
		summary := strings.TrimSpace(payload.StructuredResume.ProfessionalSummary)
		if summary == "" && len(payload.StructuredResume.Highlights) > 0 {
			summary = strings.TrimSpace(strings.Join(payload.StructuredResume.Highlights[:minInt(2, len(payload.StructuredResume.Highlights))], "；"))
		}
		skills := collectResumeSkillNames(payload.StructuredResume.SkillGraph)
		if summary != "" {
			return summary, skills
		}
	}

	rawText := strings.TrimSpace(record.RawText)
	if rawText == "" {
		return "", nil
	}
	runes := []rune(rawText)
	if len(runes) > 120 {
		rawText = string(runes[:120]) + "..."
	}
	return rawText, nil
}

func collectResumeSkillNames(graph model.ResumeSkillGraph) []string {
	flat := []model.ResumeSkillEvidence{}
	flat = append(flat, graph.ProgrammingLanguages...)
	flat = append(flat, graph.Frameworks...)
	flat = append(flat, graph.Databases...)
	flat = append(flat, graph.CloudDevOps...)
	flat = append(flat, graph.AIData...)
	flat = append(flat, graph.Tooling...)
	flat = append(flat, graph.ProductBusiness...)
	flat = append(flat, graph.Others...)

	seen := make(map[string]struct{}, len(flat))
	out := make([]string, 0, 6)
	for _, item := range flat {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, name)
		if len(out) >= 6 {
			break
		}
	}
	return out
}

func buildGroupResumePromptContext(snapshots []groupResumeSnapshot) string {
	if len(snapshots) == 0 {
		return ""
	}
	lines := make([]string, 0, len(snapshots))
	for _, item := range snapshots {
		name := strings.TrimSpace(item.DisplayName)
		if name == "" {
			name = fmt.Sprintf("候选人%d", item.UserID)
		}
		line := fmt.Sprintf("- %s：%s", name, strings.TrimSpace(item.Summary))
		if len(item.Skills) > 0 {
			line += fmt.Sprintf("（技能：%s）", strings.Join(item.Skills, ", "))
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func pickGroupQuestionType(snapshots []groupResumeSnapshot) string {
	techHits := 0
	for _, item := range snapshots {
		joined := strings.ToLower(strings.Join(item.Skills, " ") + " " + item.Summary)
		for _, kw := range []string{"architecture", "系统", "服务", "golang", "java", "数据库", "缓存", "微服务", "api", "deployment"} {
			if strings.Contains(joined, kw) {
				techHits++
				break
			}
		}
	}
	if techHits >= 2 {
		return "技术方案评审"
	}
	return "协同冲突解决"
}

func isSupportedGroupQuestionType(questionType string) bool {
	trimmed := strings.TrimSpace(questionType)
	return trimmed == "协同冲突解决" || trimmed == "技术方案评审"
}

func mapGroupQuestionCategory(questionType string) string {
	if strings.TrimSpace(questionType) == "技术方案评审" {
		return "group_tech_review"
	}
	return "group_conflict_resolution"
}

func buildGroupQuestionFallback(interview *model.Interview, questionType string, snapshots []groupResumeSnapshot) *model.Question {
	position := "通用岗位"
	difficulty := "campus_graduate"
	if interview != nil {
		if strings.TrimSpace(interview.Position) != "" {
			position = strings.TrimSpace(interview.Position)
		}
		if strings.TrimSpace(interview.Difficulty) != "" {
			difficulty = strings.TrimSpace(interview.Difficulty)
		}
	}

	snapshotContext := "请在观点冲突下达成可执行结论"
	if len(snapshots) > 0 {
		names := make([]string, 0, len(snapshots))
		for _, item := range snapshots {
			if name := strings.TrimSpace(item.DisplayName); name != "" {
				names = append(names, name)
			}
		}
		if len(names) > 0 {
			snapshotContext = fmt.Sprintf("请结合%s等成员的背景差异，推动共识", strings.Join(names, "、"))
		}
	}

	if strings.TrimSpace(questionType) == "技术方案评审" {
		return &model.Question{
			Title:          fmt.Sprintf("[%s] 群面开放题：技术方案评审", position),
			Content:        fmt.Sprintf("你们 4 人小组需要在 20 分钟内评审一个%s相关方案。请围绕可扩展性、稳定性、成本与上线风险给出统一结论，并明确分歧如何裁决。%s。", position, snapshotContext),
			ExpectedAnswer: "优秀回答应体现：角色分工、评审维度对齐、争议点收敛策略、最终决策依据与上线验证计划。",
			Position:       position,
			Difficulty:     difficulty,
			Category:       "group_tech_review",
		}
	}

	return &model.Question{
		Title:          fmt.Sprintf("[%s] 群面开放题：协同冲突解决", position),
		Content:        fmt.Sprintf("你们 4 人小组在推进%s项目时出现优先级冲突：一方强调交付速度，另一方强调技术重构。请现场达成一致行动方案并说明如何分工推进。%s。", position, snapshotContext),
		ExpectedAnswer: "优秀回答应体现：冲突拆解、目标重述、共识机制、可执行计划与风险兜底。",
		Position:       position,
		Difficulty:     difficulty,
		Category:       "group_conflict_resolution",
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *AIService) GenerateTopicQuestionFromContext(ctx context.Context, interview *model.Interview, ragContext string, category string) (*model.Question, error) {
	if interview == nil {
		return nil, fmt.Errorf("interview is nil")
	}

	prompt, err := s.renderPrompt("generate_topic_question_from_context.tmpl", map[string]interface{}{
		"Position":              interview.Position,
		"Difficulty":            interview.Difficulty,
		"Mode":                  interview.Mode,
		"Style":                 interview.Style,
		"ModeInstruction":       s.buildModePrompt(interview.Mode),
		"StyleInstruction":      s.buildStylePrompt(interview.Style, interview.Company),
		"DifficultyInstruction": s.buildDifficultyPrompt(interview.Difficulty),
		"RAGContext":            ragContext,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to render topic question prompt: %w", err)
	}

	response, err := s.chat(ctx, prompt, "chat", jsonObjectResponseFormat())
	if err != nil {
		return nil, fmt.Errorf("failed to generate topic question: %w", err)
	}

	var q struct {
		Title          string `json:"title"`
		Content        string `json:"content"`
		ExpectedAnswer string `json:"expected_answer"`
	}
	if err := json.Unmarshal([]byte(response), &q); err != nil {
		return nil, fmt.Errorf("failed to parse topic question response: %w, body: %s", err, response)
	}

	result := &model.Question{Title: q.Title, Content: q.Content, ExpectedAnswer: q.ExpectedAnswer, Position: interview.Position, Difficulty: interview.Difficulty, Category: category}
	s.EnsureQuestionChinese(ctx, result)
	s.ensureOpeningQuestionTone(result, category)
	return result, nil
}

func (s *AIService) ensureOpeningQuestionTone(q *model.Question, category string) {
	if q == nil {
		return
	}
	text := strings.TrimSpace(q.Title + " " + q.Content)
	if !isFollowUpWording(text) {
		return
	}
	topic := strings.TrimSpace(category)
	if topic == "" {
		topic = "通用技术主题"
	}
	q.Title = fmt.Sprintf("%s：核心原理与实践应用", topic)
	q.Content = fmt.Sprintf("请系统说明%s的概念、运行机制及典型应用场景。", topic)
	if strings.TrimSpace(q.ExpectedAnswer) == "" {
		q.ExpectedAnswer = fmt.Sprintf("回答应覆盖%s的定义、实现机制、约束条件与技术取舍。", topic)
	}
}

func isFollowUpWording(text string) bool {
	t := strings.TrimSpace(strings.ToLower(text))
	if t == "" {
		return false
	}
	patterns := []string{"you mentioned", "continue", "further", "follow-up", "based on previous", "你提到", "继续", "进一步", "基于上一轮", "接着"}
	for _, p := range patterns {
		if strings.Contains(t, p) {
			return true
		}
	}
	return false
}

var openingQuestionContextPatterns = []string{"previous", "above", "continue", "further", "those", "these", "上文", "上一轮", "之前", "继续", "进一步", "上述"}
var openingQuestionQuantifierRef = regexp.MustCompile(`this\s+\d+\s+`)

func (s *AIService) IsContextDependentOpeningQuestion(question *model.Question) bool {
	if question == nil {
		return true
	}
	text := strings.TrimSpace(question.Title + " " + question.Content)
	if text == "" {
		return true
	}
	if isFollowUpWording(text) || openingQuestionQuantifierRef.MatchString(strings.ToLower(text)) {
		return true
	}
	for _, p := range openingQuestionContextPatterns {
		if strings.Contains(strings.ToLower(text), p) {
			return true
		}
	}
	return false
}

func (s *AIService) NormalizeToSelfContainedOpening(question *model.Question) {
	if question == nil {
		return
	}
	topic := strings.TrimSpace(question.Category)
	if topic == "" {
		topic = strings.TrimSpace(question.Title)
	}
	if topic == "" {
		topic = "通用技术主题"
	}
	question.Title = fmt.Sprintf("%s：核心原理与实践应用", topic)
	question.Content = fmt.Sprintf("请系统说明%s的概念、运行机制、线程安全与性能取舍。", topic)
	if strings.TrimSpace(question.ExpectedAnswer) == "" {
		question.ExpectedAnswer = fmt.Sprintf("回答应覆盖%s的定义、实现机制、边界条件与技术取舍。", topic)
	}
	question.Title = sanitizeGeneratedText(question.Title)
	question.Content = sanitizeGeneratedText(question.Content)
	question.ExpectedAnswer = sanitizeGeneratedText(question.ExpectedAnswer)
}

func (s *AIService) GenerateClarifyingFollowUpQuestion(ctx context.Context, currentQ *model.Question, answer string, followUpIndex int) (*model.Question, error) {
	if currentQ == nil {
		return nil, fmt.Errorf("current question is nil")
	}
	prompt, err := s.renderPrompt("generate_clarifying_followup_question.tmpl", map[string]interface{}{"CurrentTitle": currentQ.Title, "CurrentContent": currentQ.Content, "Answer": answer, "FollowUpIndex": followUpIndex})
	if err != nil {
		return nil, fmt.Errorf("failed to render clarifying follow-up prompt: %w", err)
	}
	response, err := s.chat(ctx, prompt, "chat", jsonObjectResponseFormat())
	if err != nil {
		return nil, fmt.Errorf("failed to generate clarifying follow-up: %w", err)
	}
	var q struct {
		Title          string `json:"title"`
		Content        string `json:"content"`
		ExpectedAnswer string `json:"expected_answer"`
	}
	if err := json.Unmarshal([]byte(response), &q); err != nil {
		return nil, fmt.Errorf("failed to parse clarifying follow-up: %w, body: %s", err, response)
	}
	result := &model.Question{Title: q.Title, Content: q.Content, ExpectedAnswer: q.ExpectedAnswer, Position: currentQ.Position, Difficulty: currentQ.Difficulty, Category: currentQ.Category}
	s.EnsureQuestionChinese(ctx, result)
	return result, nil
}

func (s *AIService) GenerateNextQuestionWithWeights(ctx context.Context, interview *model.Interview, previousAnswers []model.AnswerResult, capabilityGraph *model.JobCapabilityDimension) (*model.Question, error) {
	if groupQuestion, err := s.tryGenerateGroupOpenQuestion(ctx, interview); err != nil {
		return nil, err
	} else if groupQuestion != nil {
		return groupQuestion, nil
	}

	if capabilityGraph == nil {
		return s.GenerateNextQuestion(ctx, interview, previousAnswers)
	}
	var weightsBuilder strings.Builder
	weightsBuilder.WriteString("[Capability Weights]\n")
	weightsBuilder.WriteString(fmt.Sprintf("- %s: %d%%\n", capabilityGraph.Name, capabilityGraph.Weight))
	for _, sub := range capabilityGraph.SubDimensions {
		weightsBuilder.WriteString(fmt.Sprintf("  - %s (%d%%): %s\n", sub.Name, sub.Weight, strings.Join(sub.Tags, ", ")))
	}

	prompt, err := s.renderPrompt("generate_next_question_with_weights.tmpl", map[string]interface{}{
		"Position":                interview.Position,
		"Difficulty":              interview.Difficulty,
		"Mode":                    interview.Mode,
		"Style":                   interview.Style,
		"AnsweredCount":           len(previousAnswers),
		"AnsweredKnowledgePoints": buildAnsweredKnowledgePoints(previousAnswers),
		"ModeInstruction":         s.buildModePrompt(interview.Mode),
		"StyleInstruction":        s.buildStylePrompt(interview.Style, interview.Company),
		"DifficultyInstruction":   s.buildDifficultyPrompt(interview.Difficulty),
		"WeightsInstruction":      weightsBuilder.String(),
		"NextFocus":               "Choose a high-weight dimension not fully covered yet.",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to render weighted next question prompt: %w", err)
	}

	response, err := s.chat(ctx, prompt, "chat", jsonObjectResponseFormat())
	if err != nil {
		return nil, fmt.Errorf("failed to generate question: %w", err)
	}
	var question struct {
		Title          string `json:"title"`
		Content        string `json:"content"`
		ExpectedAnswer string `json:"expected_answer"`
	}
	if err := json.Unmarshal([]byte(response), &question); err != nil {
		return nil, fmt.Errorf("failed to parse question response: %w, body: %s", err, response)
	}
	result := &model.Question{Title: question.Title, Content: question.Content, ExpectedAnswer: question.ExpectedAnswer, Position: interview.Position, Difficulty: interview.Difficulty}
	s.EnsureQuestionChinese(ctx, result)
	return result, nil
}

func (s *AIService) GenerateNextQuestion(ctx context.Context, interview *model.Interview, previousAnswers []model.AnswerResult) (*model.Question, error) {
	if groupQuestion, err := s.tryGenerateGroupOpenQuestion(ctx, interview); err != nil {
		return nil, err
	} else if groupQuestion != nil {
		return groupQuestion, nil
	}

	prompt, err := s.renderPrompt("generate_next_question.tmpl", map[string]interface{}{
		"Position":                interview.Position,
		"Difficulty":              interview.Difficulty,
		"Mode":                    interview.Mode,
		"Style":                   interview.Style,
		"AnsweredCount":           len(previousAnswers),
		"AnsweredKnowledgePoints": buildAnsweredKnowledgePoints(previousAnswers),
		"ModeInstruction":         s.buildModePrompt(interview.Mode),
		"StyleInstruction":        s.buildStylePrompt(interview.Style, interview.Company),
		"DifficultyInstruction":   s.buildDifficultyPrompt(interview.Difficulty),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to render next question prompt: %w", err)
	}
	response, err := s.chat(ctx, prompt, "chat", jsonObjectResponseFormat())
	if err != nil {
		return nil, fmt.Errorf("failed to generate question: %w", err)
	}
	var question struct {
		Title          string `json:"title"`
		Content        string `json:"content"`
		ExpectedAnswer string `json:"expected_answer"`
	}
	if err := json.Unmarshal([]byte(response), &question); err != nil {
		return nil, fmt.Errorf("failed to parse question response: %w, body: %s", err, response)
	}
	result := &model.Question{Title: question.Title, Content: question.Content, ExpectedAnswer: question.ExpectedAnswer, Position: interview.Position, Difficulty: interview.Difficulty}
	s.EnsureQuestionChinese(ctx, result)
	return result, nil
}

func buildAnsweredKnowledgePoints(previousAnswers []model.AnswerResult) string {
	if len(previousAnswers) == 0 {
		return "（暂无）"
	}

	points := make([]string, 0, len(previousAnswers)*2)
	seen := make(map[string]struct{}, len(previousAnswers)*2)
	for _, item := range previousAnswers {
		title := strings.TrimSpace(item.Question.Title)
		if title != "" {
			if _, ok := seen[title]; !ok {
				seen[title] = struct{}{}
				points = append(points, title)
			}
		}

		category := strings.TrimSpace(item.Question.Category)
		if category != "" {
			label := "领域:" + category
			if _, ok := seen[label]; !ok {
				seen[label] = struct{}{}
				points = append(points, label)
			}
		}

		if len(points) >= 8 {
			break
		}
	}

	if len(points) == 0 {
		return "（暂无）"
	}
	return strings.Join(points, "；")
}

func (s *AIService) GenerateFollowUpQuestion(ctx context.Context, interview *model.Interview, currentQ *model.Question, answer string, ragContext string, followUpContext string, followUpIndex int) (*model.Question, string, error) {
	if currentQ == nil {
		return nil, "", fmt.Errorf("current question is nil")
	}

	mode, style, difficulty, company := "technical", "gentle", "campus_intern", ""
	if interview != nil {
		mode, style, difficulty, company = interview.Mode, interview.Style, interview.Difficulty, interview.Company
	}
	prompt, err := s.renderPrompt("generate_followup_question.tmpl", map[string]interface{}{
		"Mode":                  mode,
		"Style":                 style,
		"Difficulty":            difficulty,
		"ModeInstruction":       s.buildModePrompt(mode),
		"StyleInstruction":      s.buildStylePrompt(style, company),
		"DifficultyInstruction": s.buildDifficultyPrompt(difficulty),
		"CurrentTitle":          currentQ.Title,
		"CurrentContent":        currentQ.Content,
		"Answer":                answer,
		"RAGContext":            ragContext,
		"FollowUpContext":       strings.TrimSpace(followUpContext),
		"FollowUpIndex":         followUpIndex,
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to render follow-up prompt: %w", err)
	}
	response, err := s.chat(ctx, prompt, "chat", nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate follow-up: %w", err)
	}

	questionText, noFollowUp, reason := parseFollowUpModelOutput(response)
	if noFollowUp {
		return nil, reason, nil
	}
	if questionText == "" {
		return nil, "追问生成结果为空", nil
	}

	q := &model.Question{
		Title:          buildFollowUpTitle(currentQ.Title),
		Content:        ensureQuestionSentence(questionText),
		ExpectedAnswer: buildFollowUpExpectedAnswer(followUpContext, ragContext),
		Position:       currentQ.Position,
		Difficulty:     currentQ.Difficulty,
		Category:       currentQ.Category,
	}
	s.EnsureQuestionChinese(ctx, q)
	return q, reason, nil
}

func parseFollowUpModelOutput(response string) (question string, noFollowUp bool, reason string) {
	trimmed := sanitizeGeneratedText(strings.TrimSpace(stripOptionalCodeFence(sanitizeGeneratedText(response))))
	if trimmed == "" {
		return "", true, "模型未返回有效追问"
	}

	normalized := sanitizeGeneratedText(strings.Trim(strings.TrimSpace(trimmed), "\"'` "))
	if strings.EqualFold(normalized, "NO_FOLLOWUP") {
		return "", true, "模型判定当前无需继续追问"
	}

	if strings.HasPrefix(normalized, "{") {
		var legacy struct {
			FollowUpNeeded bool   `json:"follow_up_needed"`
			Reason         string `json:"reason"`
			Question       struct {
				Title          string `json:"title"`
				Content        string `json:"content"`
				ExpectedAnswer string `json:"expected_answer"`
			} `json:"question"`
		}
		if err := json.Unmarshal([]byte(normalized), &legacy); err == nil {
			if !legacy.FollowUpNeeded {
				r := strings.TrimSpace(legacy.Reason)
				if r == "" {
					r = "模型判定当前无需继续追问"
				}
				return "", true, r
			}

			line := strings.TrimSpace(firstNonEmpty(legacy.Question.Content, legacy.Question.Title))
			if line != "" {
				return normalizeFollowUpQuestionLine(line), false, strings.TrimSpace(legacy.Reason)
			}
		}
	}

	for _, line := range strings.Split(normalized, "\n") {
		cleaned := normalizeFollowUpQuestionLine(line)
		if cleaned == "" {
			continue
		}
		if strings.EqualFold(cleaned, "NO_FOLLOWUP") {
			return "", true, "模型判定当前无需继续追问"
		}
		return cleaned, false, "基于评估上下文生成追问"
	}

	return "", true, "模型未返回可用追问"
}

func normalizeFollowUpQuestionLine(line string) string {
	cleaned := sanitizeGeneratedText(strings.TrimSpace(line))
	if cleaned == "" {
		return ""
	}
	cleaned = sanitizeGeneratedText(strings.Trim(cleaned, "\"'` "))
	for _, prefix := range []string{"问题：", "追问：", "Q:", "q:"} {
		if strings.HasPrefix(cleaned, prefix) {
			cleaned = strings.TrimSpace(strings.TrimPrefix(cleaned, prefix))
		}
	}
	cleaned = strings.TrimLeft(cleaned, "-、0123456789.． ")
	return sanitizeGeneratedText(strings.TrimSpace(cleaned))
}

func ensureQuestionSentence(text string) string {
	normalized := sanitizeGeneratedText(strings.TrimSpace(text))
	if normalized == "" {
		return ""
	}
	if strings.HasSuffix(normalized, "？") || strings.HasSuffix(normalized, "?") {
		return normalized
	}
	return normalized + "？"
}

func buildFollowUpTitle(currentTitle string) string {
	base := sanitizeGeneratedText(strings.TrimSpace(currentTitle))
	if base == "" {
		return "追问深入"
	}
	return "追问深入：" + base
}

func buildFollowUpExpectedAnswer(followUpContext, ragContext string) string {
	context := sanitizeGeneratedText(strings.TrimSpace(followUpContext))
	if context == "" {
		context = "候选人回答中需要进一步验证的技术点"
	}
	expected := fmt.Sprintf("回答应围绕“%s”，说明核心原理、实现步骤、边界条件与技术取舍。", context)
	if strings.TrimSpace(ragContext) != "" {
		expected += "可结合知识上下文中的依据进行论证。"
	}
	return expected
}

func (s *AIService) buildModePrompt(mode string) string {
	switch mode {
	case "technical":
		return s.getPromptOrDefault("ai/question/mode/technical", "")
	case "hr":
		return s.getPromptOrDefault("ai/question/mode/hr", "")
	case "comprehensive":
		return s.getPromptOrDefault("ai/question/mode/comprehensive", "")
	default:
		return ""
	}
}

func (s *AIService) buildStylePrompt(style, company string) string {
	base := s.getPromptOrDefault("ai/question/style/default", "")
	switch style {
	case "gentle":
		base = s.getPromptOrDefault("ai/question/style/gentle", "")
	case "stress":
		base = s.getPromptOrDefault("ai/question/style/stress", "")
	case "deep":
		base = s.getPromptOrDefault("ai/question/style/deep", "")
	case "practical":
		base = s.getPromptOrDefault("ai/question/style/practical", "")
	case "algorithm":
		base = s.getPromptOrDefault("ai/question/style/algorithm", "")
	}
	if company != "" {
		companySuffixTemplate := s.getPromptOrDefault("ai/question/style/company_suffix", "{{.Company}}")
		rendered, err := s.renderDynamicPrompt("ai/question/style/company_suffix", map[string]interface{}{"Company": company})
		if err == nil && strings.TrimSpace(rendered) != "" {
			if strings.TrimSpace(base) == "" {
				return rendered
			}
			return base + " " + rendered
		}
		resolved := strings.TrimSpace(strings.ReplaceAll(companySuffixTemplate, "{{.Company}}", company))
		if strings.TrimSpace(base) == "" {
			return resolved
		}
		if resolved == "" {
			return base
		}
		return base + " " + resolved
	}
	return base
}

func (s *AIService) buildDifficultyPrompt(difficulty string) string {
	switch difficulty {
	case "campus_intern":
		return s.getPromptOrDefault("ai/question/difficulty/campus_intern", "")
	case "campus_graduate":
		return s.getPromptOrDefault("ai/question/difficulty/campus_graduate", "")
	case "social_junior":
		return s.getPromptOrDefault("ai/question/difficulty/social_junior", "")
	default:
		return ""
	}
}

func GenerateRandomStyleForInterview() (style string, company string) {
	styles := []string{"gentle", "stress", "deep", "practical"}
	companies := []string{"", "ali", "bytedance", "tencent", "meituan", ""}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	style = styles[rng.Intn(len(styles))]
	company = companies[rng.Intn(len(companies))]
	return
}
