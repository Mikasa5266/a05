package model

import (
	"time"

	"gorm.io/gorm"
)

// ResumeParseResult stores the persisted snapshot of a structured resume analysis.
type ResumeParseResult struct {
	ID                  uint           `gorm:"primaryKey" json:"id"`
	UserID              uint           `gorm:"index;not null" json:"user_id"`
	FileName            string         `gorm:"size:255;not null" json:"file_name"`
	FileHash            string         `gorm:"size:64;index" json:"file_hash,omitempty"`
	RawText             string         `gorm:"type:longtext" json:"raw_text,omitempty"`
	StructuredJSON      string         `gorm:"type:longtext;not null" json:"structured_json"`
	PrimaryPositionCode string         `gorm:"size:32;index" json:"primary_position_code,omitempty"`
	PrimaryPositionName string         `gorm:"size:64" json:"primary_position_name,omitempty"`
	ConfidenceScore     int            `gorm:"default:0" json:"confidence_score"`
	ParserMode          string         `gorm:"size:32;default:'text'" json:"parser_mode"`
	ParserVersion       string         `gorm:"size:32;default:'resume-structured-v1'" json:"parser_version"`
	Source              string         `gorm:"size:32;default:'upload'" json:"source"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ResumeParseResult) TableName() string {
	return "resume_parse_results"
}

type ResumeAnalysisResult struct {
	ParserStatus       string                         `json:"parser_status"`
	ParserMode         string                         `json:"parser_mode"`
	Source             string                         `json:"source,omitempty"`
	Architecture       ResumeAnalysisArchitecture     `json:"architecture"`
	AgenticRAGTrace    []ResumeAnalysisTraceStep      `json:"agentic_rag_trace"`
	MoERouting         ResumeMoERouting               `json:"moe_routing"`
	StructuredResume   ResumeStructuredResume         `json:"structured_resume"`
	MatchResults       []ResumePositionMatch          `json:"match_results"`
	BestMatch          *ResumePositionMatch           `json:"best_match,omitempty"`
	InterviewQuestions []ResumeSuggestedQuestion      `json:"interview_questions"`
	Optimization       []ResumeOptimizationSuggestion `json:"optimization"`
	RiskReport         []ResumeRiskItem               `json:"risk_report"`
	Integration        ResumeIntegrationPayload       `json:"integration"`
	ConfidenceScore    int                            `json:"confidence_score"`
	ModelVersion       string                         `json:"model_version"`
}

// ResumeStructuredPayload is the exact JSON contract expected from the LLM.
type ResumeStructuredPayload struct {
	StructuredResume   ResumeStructuredResume         `json:"structured_resume"`
	MatchResults       []ResumePositionMatch          `json:"match_results"`
	InterviewQuestions []ResumeSuggestedQuestion      `json:"interview_questions"`
	Optimization       []ResumeOptimizationSuggestion `json:"optimization"`
	RiskReport         []ResumeRiskItem               `json:"risk_report"`
	ConfidenceScore    int                            `json:"confidence_score"`
}

type ResumeStructuredResume struct {
	PersonalInfo        ResumePersonalInfo          `json:"personal_info"`
	ProfessionalSummary string                      `json:"professional_summary"`
	CareerIntent        ResumeCareerIntent          `json:"career_intent"`
	Education           []ResumeEducationExperience `json:"education"`
	WorkExperience      []ResumeWorkExperience      `json:"work_experience"`
	ProjectExperience   []ResumeProjectExperience   `json:"project_experience"`
	SkillGraph          ResumeSkillGraph            `json:"skill_graph"`
	Certifications      []ResumeCredential          `json:"certifications"`
	Awards              []ResumeHonor               `json:"awards"`
	Languages           []ResumeLanguageProficiency `json:"languages"`
	Highlights          []string                    `json:"highlights"`
	Concerns            []string                    `json:"concerns"`
	RawPreview          string                      `json:"raw_preview"`
}

type ResumePersonalInfo struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Location  string `json:"location"`
	Github    string `json:"github"`
	Portfolio string `json:"portfolio"`
	LinkedIn  string `json:"linkedin"`
}

type ResumeCareerIntent struct {
	TargetRoles      []string `json:"target_roles"`
	TargetIndustries []string `json:"target_industries"`
	TargetCities     []string `json:"target_cities"`
	Seniority        string   `json:"seniority"`
}

type ResumeEducationExperience struct {
	School     string   `json:"school"`
	Degree     string   `json:"degree"`
	Major      string   `json:"major"`
	StartDate  string   `json:"start_date"`
	EndDate    string   `json:"end_date"`
	GPA        string   `json:"gpa"`
	Ranking    string   `json:"ranking"`
	Highlights []string `json:"highlights"`
}

type ResumeWorkExperience struct {
	Company          string   `json:"company"`
	Role             string   `json:"role"`
	StartDate        string   `json:"start_date"`
	EndDate          string   `json:"end_date"`
	Duration         string   `json:"duration"`
	Summary          string   `json:"summary"`
	Responsibilities []string `json:"responsibilities"`
	Achievements     []string `json:"achievements"`
	TechStack        []string `json:"tech_stack"`
}

type ResumeProjectExperience struct {
	Name       string   `json:"name"`
	Role       string   `json:"role"`
	StartDate  string   `json:"start_date"`
	EndDate    string   `json:"end_date"`
	Background string   `json:"background"`
	Summary    string   `json:"summary"`
	TechStack  []string `json:"tech_stack"`
	Highlights []string `json:"highlights"`
	Impact     []string `json:"impact"`
}

type ResumeSkillGraph struct {
	ProgrammingLanguages []ResumeSkillEvidence `json:"programming_languages"`
	Frameworks           []ResumeSkillEvidence `json:"frameworks"`
	Databases            []ResumeSkillEvidence `json:"databases"`
	CloudDevOps          []ResumeSkillEvidence `json:"cloud_devops"`
	AIData               []ResumeSkillEvidence `json:"ai_data"`
	Tooling              []ResumeSkillEvidence `json:"tooling"`
	ProductBusiness      []ResumeSkillEvidence `json:"product_business"`
	Others               []ResumeSkillEvidence `json:"others"`
}

type ResumeSkillEvidence struct {
	Name     string `json:"name"`
	Level    string `json:"level"`
	Evidence string `json:"evidence"`
	LastUsed string `json:"last_used"`
}

type ResumeCredential struct {
	Name        string `json:"name"`
	Issuer      string `json:"issuer"`
	AwardedDate string `json:"awarded_date"`
}

type ResumeHonor struct {
	Name        string `json:"name"`
	AwardedBy   string `json:"awarded_by"`
	AwardedDate string `json:"awarded_date"`
	Detail      string `json:"detail"`
}

type ResumeLanguageProficiency struct {
	Language    string `json:"language"`
	Proficiency string `json:"proficiency"`
	Evidence    string `json:"evidence"`
}

type ResumePositionMatch struct {
	PositionCode   string                    `json:"position_code"`
	PositionName   string                    `json:"position_name"`
	RoleKey        string                    `json:"role_key"`
	Score          int                       `json:"score"`
	ScoreBreakdown ResumeMatchScoreBreakdown `json:"score_breakdown"`
	HitSkills      []string                  `json:"hit_skills"`
	HitKeywords    []string                  `json:"hit_keywords"`
	Evidence       []string                  `json:"evidence"`
	GapSkills      []string                  `json:"gap_skills"`
	Requirements   []string                  `json:"requirements"`
	Analysis       string                    `json:"analysis"`
}

type ResumeMatchScoreBreakdown struct {
	SkillDepth       int `json:"skill_depth"`
	ProjectRelevance int `json:"project_relevance"`
	DomainAlignment  int `json:"domain_alignment"`
	DeliveryImpact   int `json:"delivery_impact"`
}

type ResumeSuggestedQuestion struct {
	Question    string   `json:"question"`
	Intent      string   `json:"intent"`
	FocusSkills []string `json:"focus_skills"`
}

type ResumeOptimizationSuggestion struct {
	Title     string `json:"title"`
	Action    string `json:"action"`
	Rationale string `json:"rationale"`
	Priority  string `json:"priority"`
}

type ResumeRiskItem struct {
	Level    string   `json:"level"`
	Item     string   `json:"item"`
	Detail   string   `json:"detail"`
	Evidence []string `json:"evidence"`
}

type ResumeAnalysisArchitecture struct {
	DecisionStack   string                    `json:"decision_stack"`
	PainPointsFixed []string                  `json:"pain_points_fixed"`
	Modules         ResumeArchitectureModules `json:"modules"`
}

type ResumeArchitectureModules struct {
	FeatureExtractor string `json:"feature_extractor"`
	CoreProcessor    string `json:"core_processor"`
	OutputGenerator  string `json:"output_generator"`
}

type ResumeAnalysisTraceStep struct {
	Agent    string `json:"agent"`
	Decision string `json:"decision"`
}

type ResumeMoERouting struct {
	Router         string            `json:"router"`
	Experts        []ResumeMoEExpert `json:"experts"`
	FusionStrategy string            `json:"fusion_strategy"`
}

type ResumeMoEExpert struct {
	Expert string  `json:"expert"`
	Weight float64 `json:"weight"`
	Reason string  `json:"reason"`
}

type ResumeIntegrationPayload struct {
	TargetRole              string                 `json:"target_role"`
	TargetPosition          string                 `json:"target_position"`
	WeakPoints              []string               `json:"weak_points"`
	QuestionRecommendations []string               `json:"question_recommendations"`
	DrillPlan               ResumeDrillPlan        `json:"drill_plan"`
	InterviewPayload        ResumeInterviewPayload `json:"interview_payload"`
}

type ResumeDrillPlan struct {
	Phase1 string `json:"phase_1"`
	Phase2 string `json:"phase_2"`
	Phase3 string `json:"phase_3"`
}

type ResumeInterviewPayload struct {
	CandidateContact ResumePersonalInfo `json:"candidate_contact"`
	FocusTopics      []string           `json:"focus_topics"`
	GeneratedAt      string             `json:"generated_at"`
}
