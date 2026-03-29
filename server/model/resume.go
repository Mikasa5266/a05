package model

import (
	"time"

	"gorm.io/gorm"
)

type ResumeData struct {
	TechStack  []string `json:"techStack"`
	Experience []struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Highlights  []string `json:"highlights"`
	} `json:"experience"`
	Intent     string   `json:"intent"`
	SoftSkills []string `json:"softSkills"`
}

type JobMatch struct {
	JobTitle     string   `json:"jobTitle"`
	MatchScore   int      `json:"matchScore"`
	Reason       string   `json:"reason"`
	Requirements []string `json:"requirements"`
}

type ResumeAuthenticityIssue struct {
	Claim           string `json:"claim"`
	RiskLevel       string `json:"riskLevel"`
	Reason          string `json:"reason"`
	VerificationTip string `json:"verificationTip"`
}

type ResumeAuthenticityReport struct {
	OverallRiskScore   int                       `json:"overallRiskScore"`
	Summary            string                    `json:"summary"`
	VerifiableItems    []string                  `json:"verifiableItems"`
	PotentialRiskItems []ResumeAuthenticityIssue `json:"potentialRiskItems"`
	InterviewChecks    []string                  `json:"interviewChecks"`
	Disclaimer         string                    `json:"disclaimer"`
}

type ResumeOptimizationReport struct {
	OverallScore int      `json:"overallScore"`
	Strengths    []string `json:"strengths"`
	Weaknesses   []string `json:"weaknesses"`
	Suggestions  []string `json:"suggestions"`
	RewriteDemo  []string `json:"rewriteDemo"`
	Keywords     []string `json:"keywords"`
}

type ResumeTemplate struct {
	TargetRole       string   `json:"targetRole"`
	TemplateMarkdown string   `json:"templateMarkdown"`
	WritingGuides    []string `json:"writingGuides"`
	CommonMistakes   []string `json:"commonMistakes"`
}

type ResumeValidationResult struct {
	IsResume        bool   `json:"is_resume"`
	ConfidenceScore int    `json:"confidence_score"`
	RejectReason    string `json:"reject_reason"`
}

type ResumeBasicInfo struct {
	Name              string `json:"name"`
	Education         string `json:"education"`
	YearsOfExperience string `json:"years_of_experience"`
	TargetDirection   string `json:"target_direction"`
}

type ResumeProjectHighlight struct {
	Name             string   `json:"name"`
	TechStack        []string `json:"tech_stack"`
	CoreContribution string   `json:"core_contribution"`
}

type ResumeExtractedData struct {
	BasicInfo         ResumeBasicInfo          `json:"basic_info"`
	CoreSkills        []string                 `json:"core_skills"`
	ProjectHighlights []ResumeProjectHighlight `json:"project_highlights"`
}

type ResumeRoleMatch struct {
	RoleName    string `json:"role_name"`
	MatchDegree int    `json:"match_degree"`
	Reason      string `json:"reason"`
}

type ResumeMatchResult struct {
	MatchedRoles        []ResumeRoleMatch `json:"matched_roles"`
	TargetQuestionBanks []string          `json:"target_question_banks"`
}

// ResumeRecord stores resume parsing snapshots for traceability and reuse.
type ResumeRecord struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	UserID     uint           `gorm:"index;not null" json:"user_id"`
	FileName   string         `gorm:"size:255" json:"file_name"`
	RawText    string         `gorm:"type:longtext" json:"raw_text"`
	ParsedData string         `gorm:"type:longtext" json:"parsed_data"`
	MatchData  string         `gorm:"type:longtext" json:"match_data"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}
