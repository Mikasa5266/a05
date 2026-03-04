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
