package service

import (
	"context"
	"encoding/json"
	"fmt"
	"your-project/model"
)

type ResumeService struct {
	aiService *AIService
}

func NewResumeService() *ResumeService {
	return &ResumeService{
		aiService: NewAIService(),
	}
}

// ParseResume calls the AI service to parse the resume text
func (s *ResumeService) ParseResume(fileContent string) (*model.ResumeData, error) {
	// 1. Construct the prompt
	prompt := fmt.Sprintf(`
You are an expert HR and Tech Recruiter. Please parse the following resume content into a structured JSON format.
Content:
%s

Output Format (JSON only, no markdown):
{
  "techStack": ["string"],
  "experience": [
    { "title": "string", "description": "string", "highlights": ["string"] }
  ],
  "intent": "string",
  "softSkills": ["string"]
}
`, fileContent)

	// 2. Call AI Service
	// We reuse the CallLLM method from AIService, but we need to ensure it supports this generic call.
	// Currently AIService has specific methods like EvaluateAnswer.
	// We should add a generic Chat or Generate method to AIService or reuse existing internal logic.
	// For now, let's assume we can call aiService.Chat (we need to implement/expose it).

	resp, err := s.aiService.Chat(context.Background(), prompt)
	if err != nil {
		return nil, fmt.Errorf("AI parsing failed: %w", err)
	}

	// 3. Unmarshal JSON
	// Clean up potential markdown code blocks
	jsonStr := CleanJSON(resp)

	var data model.ResumeData
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w, response: %s", err, jsonStr)
	}

	return &data, nil
}

// MatchJobs generates job recommendations based on resume data
func (s *ResumeService) MatchJobs(resumeData *model.ResumeData) ([]*model.JobMatch, error) {
	resumeJson, _ := json.Marshal(resumeData)
	prompt := fmt.Sprintf(`
Based on the following resume data, recommend 3 suitable job positions.
Resume Data:
%s

Output Format (JSON Array only, no markdown):
[
  {
    "jobTitle": "string",
    "matchScore": 85, // integer 0-100
    "reason": "string",
    "requirements": ["string"]
  }
]
`, string(resumeJson))

	resp, err := s.aiService.Chat(context.Background(), prompt)
	if err != nil {
		return nil, fmt.Errorf("AI matching failed: %w", err)
	}

	jsonStr := CleanJSON(resp)

	var matches []*model.JobMatch
	if err := json.Unmarshal([]byte(jsonStr), &matches); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w, response: %s", err, jsonStr)
	}

	return matches, nil
}

// Helper to clean markdown code blocks if AI returns them
func CleanJSON(s string) string {
	// Simple removal of ```json and ```
	if len(s) > 7 && s[:7] == "```json" {
		s = s[7:]
	} else if len(s) > 3 && s[:3] == "```" {
		s = s[3:]
	}
	if len(s) > 3 && s[len(s)-3:] == "```" {
		s = s[:len(s)-3]
	}
	return s
}
