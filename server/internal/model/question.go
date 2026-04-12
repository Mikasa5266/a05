package model

import (
	"encoding/json"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

type PositionCode string

const (
	PositionBackend   PositionCode = "backend"
	PositionFrontend  PositionCode = "frontend"
	PositionAlgorithm PositionCode = "algorithm"
	PositionAI        PositionCode = "ai"
)

const (
	QuestionLevelBase     = "base"
	QuestionLevelAdvanced = "advanced"
	QuestionLevelSprint   = "sprint"

	QuestionTypeTechnicalKnowledge = "技术知识题"
	QuestionTypeProjectDeepDive    = "项目经历深挖题"
	QuestionTypeScenario           = "场景应用题"
	QuestionTypeBehavioral         = "行为题"

	QuestionCompanyTypeGeneral     = "通用"
	QuestionCompanyTypeBigTech     = "大厂"
	QuestionCompanyTypeMidInternet = "中型互联网"
	QuestionCompanyTypeTraditional = "传统企业数字化"

	QuestionAttemptSourcePractice   = "practice"
	QuestionAttemptSourceAssessment = "assessment"
)

type QuestionOption struct {
	Key  string `json:"key"`
	Text string `json:"text"`
}

type JobPosition struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Code        string         `gorm:"size:32;uniqueIndex;not null" json:"code"`
	Name        string         `gorm:"size:64;uniqueIndex;not null" json:"name"`
	Domain      string         `gorm:"size:64" json:"domain,omitempty"`
	Description string         `gorm:"type:text" json:"description,omitempty"`
	IsActive    bool           `gorm:"index;default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Questions []Question `gorm:"foreignKey:PositionCode;references:Code" json:"-"`
}

func (JobPosition) TableName() string {
	return "job_positions"
}

var DefaultJobPositions = []JobPosition{
	{Code: string(PositionBackend), Name: "Java后端工程师", Domain: "backend", IsActive: true},
	{Code: string(PositionFrontend), Name: "前端工程师", Domain: "frontend", IsActive: true},
	{Code: string(PositionAlgorithm), Name: "算法工程师", Domain: "algorithm", IsActive: true},
	{Code: string(PositionAI), Name: "AI工程师", Domain: "ai", IsActive: true},
}

type Question struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	PositionCode      string         `gorm:"size:32;index:idx_question_core,priority:1;not null;default:'backend'" json:"position_code"`
	Position          string         `gorm:"size:64;index:idx_question_core,priority:2;not null" json:"position"`
	Difficulty        string         `gorm:"size:32;index:idx_question_core,priority:3;not null" json:"difficulty"`
	Level             string         `gorm:"size:24;index" json:"level"`
	DifficultyRank    int            `gorm:"index;default:1" json:"difficulty_rank"`
	DifficultyScore   int            `gorm:"index;default:3" json:"difficulty_score"`
	Category          string         `gorm:"size:64;index" json:"category"`
	Specialty         string         `gorm:"size:128;index" json:"specialty,omitempty"`
	KnowledgePoint    string         `gorm:"size:128;index" json:"knowledge_point,omitempty"`
	KnowledgeArea     string         `gorm:"size:128;index" json:"knowledge_area,omitempty"`
	Points            string         `gorm:"size:128;index" json:"points,omitempty"`
	CompanyType       string         `gorm:"size:64;index;default:'通用'" json:"company_type"`
	QuestionType      string         `gorm:"size:32;index;default:'技术知识题'" json:"question_type"`
	Title             string         `gorm:"size:255;not null" json:"title"`
	Stem              string         `gorm:"type:text" json:"stem"`
	Content           string         `gorm:"type:text;not null" json:"content"`
	OptionsJSON       string         `gorm:"column:options_json;type:longtext" json:"-"`
	StandardAnswer    string         `gorm:"column:standard_answer;type:longtext" json:"standard_answer,omitempty"`
	ExpectedAnswer    string         `gorm:"column:expected_answer;type:longtext" json:"expected_answer,omitempty"`
	Analysis          string         `gorm:"type:longtext" json:"analysis,omitempty"`
	Tips              string         `gorm:"type:text" json:"tips,omitempty"`
	Exemplar          string         `gorm:"type:text" json:"exemplar,omitempty"`
	Tags              string         `gorm:"type:text" json:"tags,omitempty"`
	KnowledgeTagsJSON string         `gorm:"column:knowledge_tags_json;type:longtext" json:"-"`
	MetadataJSON      string         `gorm:"type:longtext" json:"metadata_json,omitempty"`
	Source            string         `gorm:"size:32;index;default:'question_bank'" json:"source"`
	RAGEligible       bool           `gorm:"index;default:true" json:"rag_eligible"`
	IsActive          bool           `gorm:"index;default:true" json:"is_active"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`

	Options            []QuestionOption    `gorm:"-" json:"options,omitempty"`
	KnowledgeTags      []string            `gorm:"-" json:"knowledge_tags,omitempty"`
	InterviewQuestions []InterviewQuestion `gorm:"foreignKey:QuestionID" json:"-"`
	PositionRef        *JobPosition        `gorm:"foreignKey:PositionCode;references:Code" json:"position_ref,omitempty"`
}

type QuestionPracticeRecord struct {
	ID                  uint           `gorm:"primaryKey" json:"id"`
	UserID              uint           `gorm:"index:idx_q_attempt_user_question,priority:1;index;not null" json:"user_id"`
	QuestionID          uint           `gorm:"index:idx_q_attempt_user_question,priority:2;index;not null" json:"question_id"`
	AssessmentID        *uint          `gorm:"index" json:"assessment_id,omitempty"`
	SourceKind          string         `gorm:"size:24;index;default:'practice'" json:"source_kind"`
	UserAnswer          string         `gorm:"type:longtext" json:"user_answer"`
	NormalizedAnswer    string         `gorm:"type:text" json:"normalized_answer,omitempty"`
	AnswerMode          string         `gorm:"size:32" json:"answer_mode,omitempty"`
	IsCorrect           bool           `gorm:"index" json:"is_correct"`
	ErrorReason         string         `gorm:"type:text" json:"error_reason,omitempty"`
	MatchedKeywordsJSON string         `gorm:"column:matched_keywords_json;type:longtext" json:"-"`
	MissingKeywordsJSON string         `gorm:"column:missing_keywords_json;type:longtext" json:"-"`
	ElapsedSeconds      *int           `json:"elapsed_seconds,omitempty"`
	TimedMode           bool           `gorm:"default:false" json:"timed_mode"`
	IsTimeout           bool           `gorm:"default:false" json:"is_timeout"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`

	MatchedKeywords []string `gorm:"-" json:"matched_keywords,omitempty"`
	MissingKeywords []string `gorm:"-" json:"missing_keywords,omitempty"`
	Question        Question `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
}

type QuestionAssessment struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"index;not null" json:"user_id"`
	PositionCode string         `gorm:"size:32;index;not null" json:"position_code"`
	Difficulty   string         `gorm:"size:32;index" json:"difficulty,omitempty"`
	TotalCount   int            `gorm:"default:12" json:"total_count"`
	CorrectCount int            `gorm:"default:0" json:"correct_count"`
	Score        float64        `gorm:"default:0" json:"score"`
	Status       string         `gorm:"size:24;index;default:'in_progress'" json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	CompletedAt  *time.Time     `json:"completed_at,omitempty"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	Items []QuestionAssessmentItem `gorm:"foreignKey:AssessmentID" json:"items,omitempty"`
}

type QuestionAssessmentItem struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	AssessmentID uint      `gorm:"uniqueIndex:idx_q_assessment_item_unique,priority:1;not null" json:"assessment_id"`
	QuestionID   uint      `gorm:"uniqueIndex:idx_q_assessment_item_unique,priority:2;not null" json:"question_id"`
	OrderNo      int       `gorm:"index;not null" json:"order_no"`
	CreatedAt    time.Time `json:"created_at"`

	Assessment QuestionAssessment `gorm:"foreignKey:AssessmentID" json:"-"`
	Question   Question           `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
}

func (QuestionPracticeRecord) TableName() string {
	return "question_practice_records"
}

func (QuestionAssessment) TableName() string {
	return "question_assessments"
}

func (QuestionAssessmentItem) TableName() string {
	return "question_assessment_items"
}

func (q *Question) GetTags() []string {
	if strings.TrimSpace(q.Tags) == "" {
		return []string{}
	}
	raw := strings.Split(q.Tags, ",")
	out := make([]string, 0, len(raw))
	seen := make(map[string]struct{}, len(raw))
	for _, item := range raw {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func (q *Question) SetTags(tags []string) {
	q.Tags = strings.Join(normalizeStringSlice(tags), ",")
}

func (q *Question) GetKnowledgeTags() []string {
	if len(q.KnowledgeTags) > 0 {
		return append([]string(nil), q.KnowledgeTags...)
	}
	if strings.TrimSpace(q.KnowledgeTagsJSON) != "" {
		var tags []string
		if err := json.Unmarshal([]byte(q.KnowledgeTagsJSON), &tags); err == nil {
			q.KnowledgeTags = normalizeStringSlice(tags)
			return append([]string(nil), q.KnowledgeTags...)
		}
	}
	return q.GetTags()
}

func (q *Question) SetKnowledgeTags(tags []string) {
	normalized := normalizeStringSlice(tags)
	q.KnowledgeTags = append([]string(nil), normalized...)
	if len(normalized) == 0 {
		q.KnowledgeTagsJSON = ""
		if strings.TrimSpace(q.Tags) == "" {
			q.Tags = ""
		}
		return
	}
	encoded, err := json.Marshal(normalized)
	if err != nil {
		q.KnowledgeTagsJSON = ""
	} else {
		q.KnowledgeTagsJSON = string(encoded)
	}
	if strings.TrimSpace(q.Tags) == "" {
		q.Tags = strings.Join(normalized, ",")
	}
}

func (q *Question) GetOptions() []QuestionOption {
	if len(q.Options) > 0 {
		return append([]QuestionOption(nil), q.Options...)
	}
	options := parseQuestionOptions(q.OptionsJSON)
	q.Options = append([]QuestionOption(nil), options...)
	return append([]QuestionOption(nil), options...)
}

func (q *Question) SetOptions(options []QuestionOption) {
	normalized := normalizeQuestionOptions(options)
	q.Options = append([]QuestionOption(nil), normalized...)
	if len(normalized) == 0 {
		q.OptionsJSON = ""
		return
	}
	encoded, err := json.Marshal(normalized)
	if err != nil {
		q.OptionsJSON = ""
		return
	}
	q.OptionsJSON = string(encoded)
}

func (q *Question) EffectiveStem() string {
	stem := strings.TrimSpace(q.Stem)
	if stem != "" {
		return stem
	}
	return strings.TrimSpace(q.Content)
}

func (q *Question) EffectiveStandardAnswer() string {
	answer := strings.TrimSpace(q.StandardAnswer)
	if answer != "" {
		return answer
	}
	return strings.TrimSpace(q.ExpectedAnswer)
}

func (q *Question) EffectiveAnalysis() string {
	analysis := strings.TrimSpace(q.Analysis)
	if analysis != "" {
		return analysis
	}
	return strings.TrimSpace(q.EffectiveStandardAnswer())
}

func (q *Question) HydrateTransientFields() {
	q.Options = parseQuestionOptions(q.OptionsJSON)
	q.KnowledgeTags = q.GetKnowledgeTags()
}

func (r *QuestionPracticeRecord) GetMatchedKeywords() []string {
	if len(r.MatchedKeywords) > 0 {
		return append([]string(nil), r.MatchedKeywords...)
	}
	return parseJSONStringSlice(r.MatchedKeywordsJSON)
}

func (r *QuestionPracticeRecord) SetMatchedKeywords(items []string) {
	r.MatchedKeywords = normalizeStringSlice(items)
	r.MatchedKeywordsJSON = encodeJSONStringSlice(r.MatchedKeywords)
}

func (r *QuestionPracticeRecord) GetMissingKeywords() []string {
	if len(r.MissingKeywords) > 0 {
		return append([]string(nil), r.MissingKeywords...)
	}
	return parseJSONStringSlice(r.MissingKeywordsJSON)
}

func (r *QuestionPracticeRecord) SetMissingKeywords(items []string) {
	r.MissingKeywords = normalizeStringSlice(items)
	r.MissingKeywordsJSON = encodeJSONStringSlice(r.MissingKeywords)
}

func (q *Question) BeforeCreate(tx *gorm.DB) error {
	q.normalizeDerivedFields()
	return nil
}

func (q *Question) AfterFind(tx *gorm.DB) error {
	q.HydrateTransientFields()
	return nil
}

func (q *Question) BeforeSave(tx *gorm.DB) error {
	q.normalizeDerivedFields()
	return nil
}

func (r *QuestionPracticeRecord) BeforeCreate(tx *gorm.DB) error {
	r.normalizeDerivedFields()
	return nil
}

func (r *QuestionPracticeRecord) AfterFind(tx *gorm.DB) error {
	r.MatchedKeywords = parseJSONStringSlice(r.MatchedKeywordsJSON)
	r.MissingKeywords = parseJSONStringSlice(r.MissingKeywordsJSON)
	return nil
}

func (r *QuestionPracticeRecord) BeforeSave(tx *gorm.DB) error {
	r.normalizeDerivedFields()
	return nil
}

func (q *Question) normalizeDerivedFields() {
	q.PositionCode = strings.TrimSpace(strings.ToLower(q.PositionCode))
	q.Position = sanitizeQuestionText(q.Position)
	q.Difficulty = strings.TrimSpace(strings.ToLower(q.Difficulty))
	q.Level = strings.TrimSpace(strings.ToLower(q.Level))
	q.Category = sanitizeQuestionText(q.Category)
	q.Specialty = sanitizeQuestionText(q.Specialty)
	q.KnowledgePoint = sanitizeQuestionText(q.KnowledgePoint)
	q.KnowledgeArea = sanitizeQuestionText(q.KnowledgeArea)
	q.Points = sanitizeQuestionText(q.Points)
	q.CompanyType = sanitizeQuestionText(q.CompanyType)
	q.QuestionType = sanitizeQuestionText(q.QuestionType)
	q.Title = sanitizeQuestionText(q.Title)
	q.Stem = sanitizeQuestionText(q.Stem)
	q.Content = sanitizeQuestionText(q.Content)
	q.StandardAnswer = sanitizeQuestionText(q.StandardAnswer)
	q.ExpectedAnswer = sanitizeQuestionText(q.ExpectedAnswer)
	q.Analysis = sanitizeQuestionText(q.Analysis)
	q.Tips = sanitizeQuestionText(q.Tips)
	q.Exemplar = sanitizeQuestionText(q.Exemplar)
	q.Tags = strings.Join(normalizeStringSlice(q.GetTags()), ",")
	q.MetadataJSON = strings.TrimSpace(strings.ToValidUTF8(q.MetadataJSON, ""))
	q.Source = strings.TrimSpace(q.Source)

	if q.Stem == "" {
		q.Stem = q.Content
	}
	if q.Content == "" {
		q.Content = q.Stem
	}
	if q.Title == "" {
		q.Title = buildQuestionTitle(q.EffectiveStem())
	}
	if q.StandardAnswer == "" {
		q.StandardAnswer = q.ExpectedAnswer
	}
	if q.ExpectedAnswer == "" {
		q.ExpectedAnswer = q.StandardAnswer
	}
	if q.Analysis == "" {
		q.Analysis = q.EffectiveStandardAnswer()
	}
	if q.Specialty == "" {
		q.Specialty = q.Category
	}
	if q.Category == "" {
		q.Category = q.Specialty
	}
	if q.Points == "" {
		switch {
		case q.KnowledgePoint != "":
			q.Points = q.KnowledgePoint
		case q.Specialty != "":
			q.Points = q.Specialty
		default:
			q.Points = q.Category
		}
	}
	if q.KnowledgePoint == "" {
		q.KnowledgePoint = q.Points
	}
	if q.KnowledgeArea == "" {
		switch {
		case q.Specialty != "":
			q.KnowledgeArea = q.Specialty
		default:
			q.KnowledgeArea = q.Category
		}
	}
	if q.PositionCode == "" {
		q.PositionCode = inferPositionCode(q.Position)
	}
	if q.PositionCode == "" {
		q.PositionCode = string(PositionBackend)
	}
	if q.Position == "" {
		q.Position = defaultPositionName(q.PositionCode)
	}
	if q.DifficultyRank <= 0 {
		q.DifficultyRank = inferDifficultyRank(q.Difficulty)
	}
	if q.DifficultyRank <= 0 {
		q.DifficultyRank = 1
	}
	if q.Level == "" {
		q.Level = inferPracticeLevel(q.Difficulty, q.DifficultyRank)
	}
	if q.DifficultyScore <= 0 {
		q.DifficultyScore = inferDifficultyScore(q.DifficultyRank, q.Level)
	}
	if q.QuestionType == "" {
		q.QuestionType = inferQuestionType(q.Title, q.EffectiveStem(), q.Category, q.Specialty, q.OptionsJSON)
	}
	if q.CompanyType == "" {
		q.CompanyType = QuestionCompanyTypeGeneral
	}
	if strings.TrimSpace(q.OptionsJSON) == "" && len(q.Options) > 0 {
		q.SetOptions(q.Options)
	} else {
		q.Options = parseQuestionOptions(q.OptionsJSON)
	}
	if len(q.KnowledgeTags) == 0 && strings.TrimSpace(q.KnowledgeTagsJSON) == "" {
		q.SetKnowledgeTags(buildDefaultKnowledgeTags(q))
	} else if len(q.KnowledgeTags) > 0 {
		q.SetKnowledgeTags(q.KnowledgeTags)
	} else {
		q.KnowledgeTags = normalizeStringSlice(parseJSONStringSlice(q.KnowledgeTagsJSON))
		if len(q.KnowledgeTags) == 0 {
			q.KnowledgeTags = q.GetTags()
		}
		if strings.TrimSpace(q.Tags) == "" {
			q.Tags = strings.Join(q.KnowledgeTags, ",")
		}
	}
	if q.Source == "" {
		q.Source = "question_bank"
	}
}

func (r *QuestionPracticeRecord) normalizeDerivedFields() {
	r.SourceKind = strings.TrimSpace(r.SourceKind)
	r.UserAnswer = sanitizeQuestionText(r.UserAnswer)
	r.NormalizedAnswer = sanitizeQuestionText(r.NormalizedAnswer)
	r.AnswerMode = sanitizeQuestionText(r.AnswerMode)
	r.ErrorReason = sanitizeQuestionText(r.ErrorReason)
	if r.SourceKind == "" {
		r.SourceKind = QuestionAttemptSourcePractice
	}
	r.SetMatchedKeywords(r.MatchedKeywords)
	r.SetMissingKeywords(r.MissingKeywords)
}

func sanitizeQuestionText(value string) string {
	value = strings.ToValidUTF8(value, "")
	value = strings.ReplaceAll(value, "\uFEFF", "")
	value = strings.ReplaceAll(value, "\ufffd", "")
	value = strings.ReplaceAll(value, "```json", "")
	value = strings.ReplaceAll(value, "```JSON", "")
	value = strings.ReplaceAll(value, "```markdown", "")
	value = strings.ReplaceAll(value, "```md", "")
	value = strings.ReplaceAll(value, "```text", "")
	value = strings.ReplaceAll(value, "```", "")

	var builder strings.Builder
	builder.Grow(len(value))
	lastWasBlank := false
	for _, r := range value {
		switch {
		case r == '\r':
			continue
		case r == '\n':
			if !lastWasBlank {
				builder.WriteRune('\n')
			}
			lastWasBlank = true
		case r == '\t':
			builder.WriteRune(' ')
			lastWasBlank = false
		case r < 32:
			continue
		default:
			builder.WriteRune(r)
			lastWasBlank = false
		}
	}
	return strings.TrimSpace(builder.String())
}

func normalizeStringSlice(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		trimmed := sanitizeQuestionText(item)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func normalizeQuestionOptions(items []QuestionOption) []QuestionOption {
	if len(items) == 0 {
		return []QuestionOption{}
	}
	out := make([]QuestionOption, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for idx, item := range items {
		key := strings.TrimSpace(strings.ToUpper(item.Key))
		text := sanitizeQuestionText(item.Text)
		if key == "" {
			key = optionKeyByIndex(idx)
		}
		if text == "" {
			continue
		}
		compound := key + "|" + text
		if _, ok := seen[compound]; ok {
			continue
		}
		seen[compound] = struct{}{}
		out = append(out, QuestionOption{Key: key, Text: text})
	}
	return out
}

func parseJSONStringSlice(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{}
	}
	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err == nil {
		return normalizeStringSlice(out)
	}
	return []string{}
}

func encodeJSONStringSlice(items []string) string {
	if len(items) == 0 {
		return ""
	}
	encoded, err := json.Marshal(normalizeStringSlice(items))
	if err != nil {
		return ""
	}
	return string(encoded)
}

func parseQuestionOptions(raw string) []QuestionOption {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []QuestionOption{}
	}

	var direct []QuestionOption
	if err := json.Unmarshal([]byte(raw), &direct); err == nil {
		return normalizeQuestionOptions(direct)
	}

	var textList []string
	if err := json.Unmarshal([]byte(raw), &textList); err == nil {
		options := make([]QuestionOption, 0, len(textList))
		for idx, item := range textList {
			options = append(options, QuestionOption{Key: optionKeyByIndex(idx), Text: item})
		}
		return normalizeQuestionOptions(options)
	}

	var objectMap map[string]string
	if err := json.Unmarshal([]byte(raw), &objectMap); err == nil {
		keys := make([]string, 0, len(objectMap))
		for key := range objectMap {
			keys = append(keys, strings.ToUpper(strings.TrimSpace(key)))
		}
		sort.Strings(keys)
		options := make([]QuestionOption, 0, len(keys))
		for _, key := range keys {
			options = append(options, QuestionOption{Key: key, Text: objectMap[key]})
		}
		return normalizeQuestionOptions(options)
	}

	return []QuestionOption{}
}

func optionKeyByIndex(index int) string {
	if index < 0 {
		return ""
	}
	return string(rune('A' + index))
}

func buildDefaultKnowledgeTags(q *Question) []string {
	if q == nil {
		return []string{}
	}
	return normalizeStringSlice([]string{
		q.PositionCode,
		q.Level,
		q.QuestionType,
		q.Specialty,
		q.KnowledgePoint,
		q.Category,
	})
}

func buildQuestionTitle(stem string) string {
	stem = sanitizeQuestionText(stem)
	if stem == "" {
		return "题目"
	}
	runes := []rune(stem)
	if len(runes) <= 48 {
		return stem
	}
	return strings.TrimSpace(string(runes[:48]))
}

func defaultPositionName(code string) string {
	switch strings.TrimSpace(strings.ToLower(code)) {
	case string(PositionFrontend):
		return "前端工程师"
	case string(PositionAlgorithm):
		return "算法工程师"
	case string(PositionAI):
		return "AI工程师"
	default:
		return "Java后端工程师"
	}
}

func inferPositionCode(position string) string {
	p := strings.ToLower(strings.TrimSpace(position))
	switch {
	case p == "", strings.Contains(p, "java"), strings.Contains(p, "后端"), p == "backend":
		return string(PositionBackend)
	case strings.Contains(p, "前端"), strings.Contains(p, "frontend"):
		return string(PositionFrontend)
	case strings.Contains(p, "算法"), p == "algorithm":
		return string(PositionAlgorithm)
	case strings.Contains(p, "ai"), strings.Contains(p, "llm"), strings.Contains(p, "模型"), p == "ai_engineer":
		return string(PositionAI)
	default:
		return string(PositionBackend)
	}
}

func inferDifficultyRank(difficulty string) int {
	d := strings.ToLower(strings.TrimSpace(difficulty))
	switch d {
	case "campus_intern", "easy", "junior", "初级", "基础":
		return 1
	case "campus_graduate", "medium", "intermediate", "中级":
		return 2
	case "social_junior", "hard", "senior", "中高级":
		return 3
	default:
		return 1
	}
}

func inferPracticeLevel(difficulty string, rank int) string {
	switch {
	case rank >= 3:
		return QuestionLevelSprint
	case rank == 2:
		return QuestionLevelAdvanced
	}

	switch strings.ToLower(strings.TrimSpace(difficulty)) {
	case "social_junior", "hard", "senior":
		return QuestionLevelSprint
	case "campus_graduate", "medium", "intermediate":
		return QuestionLevelAdvanced
	default:
		return QuestionLevelBase
	}
}

func inferDifficultyScore(rank int, level string) int {
	switch {
	case strings.TrimSpace(strings.ToLower(level)) == QuestionLevelSprint:
		return 8
	case strings.TrimSpace(strings.ToLower(level)) == QuestionLevelAdvanced:
		return 6
	case rank >= 3:
		return 8
	case rank == 2:
		return 6
	default:
		return 3
	}
}

func inferQuestionType(title, stem, category, specialty, optionsJSON string) string {
	text := strings.ToLower(strings.TrimSpace(title + " " + stem + " " + category + " " + specialty))
	if strings.TrimSpace(optionsJSON) != "" {
		return QuestionTypeTechnicalKnowledge
	}
	switch {
	case strings.Contains(text, "项目"), strings.Contains(text, "经历"), strings.Contains(text, "实习"):
		return QuestionTypeProjectDeepDive
	case strings.Contains(text, "场景"), strings.Contains(text, "方案"), strings.Contains(text, "设计"), strings.Contains(text, "如何"):
		return QuestionTypeScenario
	case strings.Contains(text, "行为"), strings.Contains(text, "冲突"), strings.Contains(text, "协作"), strings.Contains(text, "复盘"):
		return QuestionTypeBehavioral
	default:
		return QuestionTypeTechnicalKnowledge
	}
}
