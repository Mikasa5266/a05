package service

import (
	"context"
	"fmt"
	"testing"

	"your-project/model"
	"your-project/pkg/llm"
	"your-project/repository"
)

type stubResumeLLMClient struct {
	responses []string
	index     int
}

func (s *stubResumeLLMClient) Chat(ctx context.Context, req llm.ChatRequest) (string, error) {
	_ = ctx
	_ = req
	if s.index >= len(s.responses) {
		return "", fmt.Errorf("no more stub responses")
	}
	out := s.responses[s.index]
	s.index++
	return out, nil
}

type stubPositionRepo struct {
	positions []model.JobPosition
}

func (s *stubPositionRepo) Create(position *model.JobPosition) error { return nil }
func (s *stubPositionRepo) GetByCode(code string) (*model.JobPosition, error) {
	for _, item := range s.positions {
		if item.Code == code {
			copy := item
			return &copy, nil
		}
	}
	return nil, fmt.Errorf("not found")
}
func (s *stubPositionRepo) ListActive() ([]model.JobPosition, error)          { return s.positions, nil }
func (s *stubPositionRepo) UpsertDefaults(defaults []model.JobPosition) error { return nil }
func (s *stubPositionRepo) EnsureDefaults() error                             { return nil }
func (s *stubPositionRepo) Update(position *model.JobPosition) error          { return nil }
func (s *stubPositionRepo) DeleteByCode(code string) error                    { return nil }

var _ repository.PositionRepository = (*stubPositionRepo)(nil)

type stubResumeRepo struct {
	last *model.ResumeParseResult
}

func (s *stubResumeRepo) Create(result *model.ResumeParseResult) error {
	s.last = result
	return nil
}

func (s *stubResumeRepo) GetByID(id uint) (*model.ResumeParseResult, error) {
	if s.last == nil {
		return nil, fmt.Errorf("not found")
	}
	return s.last, nil
}

func (s *stubResumeRepo) GetLatestByUser(userID uint) (*model.ResumeParseResult, error) {
	if s.last == nil {
		return nil, fmt.Errorf("not found")
	}
	return s.last, nil
}

var _ repository.ResumeParseResultRepository = (*stubResumeRepo)(nil)

func TestResumeServiceAnalyzeOnlyParsesFencedJSON(t *testing.T) {
	client := &stubResumeLLMClient{
		responses: []string{
			"```json\n" + `{
  "structured_resume": {
    "personal_info": {
      "name": "Alex",
      "email": "alex@example.com",
      "phone": "13800138000",
      "location": "Shanghai",
      "github": "",
      "portfolio": "",
      "linkedin": ""
    },
    "professional_summary": "Backend engineer with Java and Redis experience across transaction systems.",
    "career_intent": {
      "target_roles": ["Backend Engineer"],
      "target_industries": ["Internet"],
      "target_cities": ["Shanghai"],
      "seniority": "Junior"
    },
    "education": [],
    "work_experience": [],
    "project_experience": [],
    "skill_graph": {
      "programming_languages": [{"name":"Java","level":"advanced","evidence":"Resume mentions Java projects","last_used":"current"}],
      "frameworks": [],
      "databases": [],
      "cloud_devops": [],
      "ai_data": [],
      "tooling": [],
      "product_business": [],
      "others": []
    },
    "certifications": [],
    "awards": [],
    "languages": [],
    "highlights": ["Strong transaction system context"],
    "concerns": [],
    "raw_preview": ""
  },
  "match_results": [
    {
      "position_code": "backend",
      "position_name": "Backend Engineer",
      "role_key": "java_backend",
      "score": 88,
      "score_breakdown": {
        "skill_depth": 90,
        "project_relevance": 84,
        "domain_alignment": 82,
        "delivery_impact": 80
      },
      "hit_skills": ["Java"],
      "hit_keywords": ["High Concurrency"],
      "evidence": ["Summary shows backend and cache optimization experience"],
      "gap_skills": ["Service Governance"],
      "requirements": ["Master mainstream backend frameworks"],
      "analysis": "Backend is the strongest match."
    }
  ],
  "interview_questions": [],
  "optimization": [],
  "risk_report": [],
  "confidence_score": 86
}` + "\n```",
		},
	}

	svc := NewResumeServiceWithDeps(client, &stubPositionRepo{positions: model.DefaultJobPositions}, &stubResumeRepo{})
	result, err := svc.AnalyzeOnly(context.Background(), ResumeAnalysisInput{
		FileName: "resume.pdf",
		RawText:  "Alex has 3 years of Java backend experience, worked on transaction systems, cache optimization, database tuning, and multiple production releases with Redis and MySQL in high traffic scenarios.",
	})
	if err != nil {
		t.Fatalf("expected analyze success, got error: %v", err)
	}
	if result.BestMatch == nil || result.BestMatch.PositionCode != "backend" {
		t.Fatalf("expected backend best match, got %+v", result.BestMatch)
	}
	if len(result.InterviewQuestions) == 0 {
		t.Fatalf("expected synthesized interview questions when llm output is empty")
	}
	if result.StructuredResume.RawPreview == "" {
		t.Fatalf("expected raw preview to be normalized")
	}
}

func TestResumeServiceAnalyzeOnlyRetriesOnBrokenJSON(t *testing.T) {
	client := &stubResumeLLMClient{
		responses: []string{
			"not-json-at-all",
			`{
  "structured_resume": {
    "personal_info": {
      "name": "Robin",
      "email": "",
      "phone": "",
      "location": "",
      "github": "",
      "portfolio": "",
      "linkedin": ""
    },
    "professional_summary": "Hands-on RAG and LLM application experience.",
    "career_intent": {
      "target_roles": ["AI Engineer"],
      "target_industries": [],
      "target_cities": [],
      "seniority": "Graduate"
    },
    "education": [],
    "work_experience": [],
    "project_experience": [],
    "skill_graph": {
      "programming_languages": [],
      "frameworks": [],
      "databases": [],
      "cloud_devops": [],
      "ai_data": [{"name":"RAG","level":"intermediate","evidence":"Summary mentions RAG and LLM work","last_used":"recent"}],
      "tooling": [],
      "product_business": [],
      "others": []
    },
    "certifications": [],
    "awards": [],
    "languages": [],
    "highlights": [],
    "concerns": [],
    "raw_preview": ""
  },
  "match_results": [
    {
      "position_code": "ai",
      "position_name": "AI Engineer",
      "role_key": "ai_engineer",
      "score": 91,
      "score_breakdown": {
        "skill_depth": 90,
        "project_relevance": 88,
        "domain_alignment": 92,
        "delivery_impact": 80
      },
      "hit_skills": ["RAG"],
      "hit_keywords": ["LLM"],
      "evidence": ["Summary explicitly mentions RAG and LLM application work"],
      "gap_skills": ["Model Evaluation"],
      "requirements": [],
      "analysis": "AI is the strongest match."
    }
  ],
  "interview_questions": [],
  "optimization": [],
  "risk_report": [],
  "confidence_score": 90
}`,
		},
	}

	svc := NewResumeServiceWithDeps(client, &stubPositionRepo{positions: model.DefaultJobPositions}, &stubResumeRepo{})
	result, err := svc.AnalyzeOnly(context.Background(), ResumeAnalysisInput{
		FileName: "resume.txt",
		RawText:  "Robin built a RAG question answering system, owned vector retrieval, prompt design, chunking strategy, answer evaluation, and participated in iterative LLM application delivery.",
	})
	if err != nil {
		t.Fatalf("expected analyze success after retry, got error: %v", err)
	}
	if client.index != 2 {
		t.Fatalf("expected retry to consume second llm response, got index %d", client.index)
	}
	if result.BestMatch == nil || result.BestMatch.PositionCode != "ai" {
		t.Fatalf("expected ai best match after retry, got %+v", result.BestMatch)
	}
}
