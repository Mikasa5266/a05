package model

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
