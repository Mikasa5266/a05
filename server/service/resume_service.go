package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
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
	log.Printf("Starting resume parsing, content length: %d characters", len(fileContent))

	const MaxContentLength = 12000
	if len(fileContent) > MaxContentLength {
		log.Printf("Content too long, truncating from %d to %d characters", len(fileContent), MaxContentLength)
		fileContent = fileContent[:MaxContentLength] + "\n...(content truncated)..."
	}

	prompt := fmt.Sprintf(`
你是一位专业的人力资源和技术招聘专家。请将以下简历内容解析为结构化的 JSON 格式。
如果内容被截断，请尽力根据现有信息进行解析。

【重要要求】
1. 所有提取的内容（职位、描述、技能等）必须使用**简体中文**输出。如果原文是英文，请翻译成中文。
2. 不要输出任何 Markdown 标记（如 `+"`"+"```json"+"`"+`），只返回纯 JSON 字符串。
3. **严格根据简历内容分析**，不要使用示例中的硬编码内容。

简历内容:
%s

输出格式示例 (JSON):
{
  "techStack": ["Java", "Spring Boot", "MySQL"],
  "experience": [
    { "title": "高级后端工程师", "description": "负责核心支付系统开发...", "highlights": ["优化了数据库性能", "重构了遗留代码"] }
  ],
  "intent": "Java 后端开发专家",
  "softSkills": ["团队协作", "沟通能力", "抗压能力"]
}
`, fileContent)

	log.Printf("Sending request to AI service for resume parsing")

	resp, err := s.aiService.Chat(context.Background(), prompt)
	if err != nil {
		log.Printf("AI parsing failed: %v", err)
		return nil, fmt.Errorf("AI parsing failed: %w", err)
	}

	log.Printf("AI response received, length: %d characters", len(resp))

	jsonStr := CleanJSON(resp)

	var data model.ResumeData
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		log.Printf("Failed to parse AI response: %v, response: %s", err, jsonStr)
		return nil, fmt.Errorf("failed to parse AI response: %w, response: %s", err, jsonStr)
	}

	log.Printf("Resume parsed successfully: techStack=%v, experienceCount=%d", data.TechStack, len(data.Experience))

	return &data, nil
}

// MatchJobs generates job recommendations based on resume data
func (s *ResumeService) MatchJobs(resumeData *model.ResumeData) ([]*model.JobMatch, error) {
	log.Printf("Starting job matching for resume: techStack=%v", resumeData.TechStack)

	resumeJson, _ := json.Marshal(resumeData)
	prompt := fmt.Sprintf(`
根据以下简历数据，推荐 3 个最适合的职位。

【重要要求】
1. **必须使用简体中文**输出所有内容。
2. 职位名称可以是中英文（如 "Go 后端开发" 或 "Backend Engineer"），但描述和理由必须是中文。
3. **严格根据简历的技术栈和经验推荐职位**，不要推荐与简历内容不符的职位。
4. 不要输出 Markdown 标记。

简历数据:
%s

输出格式 (JSON 数组):
[
  {
    "jobTitle": "推荐职位名称",
    "matchScore": 90, // 0-100 的整数
    "reason": "详细的推荐理由...",
    "requirements": ["该职位的核心要求1", "要求2"]
  }
]
`, string(resumeJson))

	log.Printf("Sending request to AI service for job matching")

	resp, err := s.aiService.Chat(context.Background(), prompt)
	if err != nil {
		log.Printf("AI matching failed: %v", err)
		return nil, fmt.Errorf("AI matching failed: %w", err)
	}

	log.Printf("AI response received, length: %d characters", len(resp))

	jsonStr := CleanJSON(resp)

	var matches []*model.JobMatch
	if err := json.Unmarshal([]byte(jsonStr), &matches); err != nil {
		log.Printf("Failed to parse AI response: %v, response: %s", err, jsonStr)
		return nil, fmt.Errorf("failed to parse AI response: %w, response: %s", err, jsonStr)
	}

	log.Printf("Job matching completed: %d matches generated", len(matches))
	for i, match := range matches {
		log.Printf("Match %d: %s (score: %d)", i+1, match.JobTitle, match.MatchScore)
	}

	return matches, nil
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
