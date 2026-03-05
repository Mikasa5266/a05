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

	const MaxContentLength = 50000
	if len(fileContent) > MaxContentLength {
		log.Printf("Content too long, truncating from %d to %d characters", len(fileContent), MaxContentLength)
		fileContent = fileContent[:MaxContentLength] + "\n...(content truncated)..."
	}

	prompt := fmt.Sprintf(`
你是一位资深技术面试官和职业规划专家。请仔细阅读以下简历内容，并进行深度解析。
这是一份 PDF 导出的文本，可能包含排版错乱、换行符丢失或多余空格。请根据上下文智能重建语义。

【解析目标】
将简历内容转化为结构化的 JSON 数据，以便系统进行岗位匹配。

【重要规则】
1. **必须使用简体中文**输出。
2. **严格基于简历内容**，不要编造或猜测未提及的信息。如果某项信息完全缺失，请留空或返回空数组。
3. **不要输出 Markdown 代码块**，直接返回纯 JSON 字符串。
4. **技术栈提取**：请提取具体的编程语言、框架、工具（如 Java, Spring Boot, MySQL, Redis, Vue.js 等）。
5. **项目经验**：请提取项目名称、描述和关键亮点（技术难点、优化成果等）。
6. **求职意向**：如果简历未明确写明，请根据技术栈和经验推断最可能的职位（如 "后端开发工程师", "全栈工程师"）。
7. **软技能**：提取简历中体现的非技术能力（如 "团队管理", "沟通协作", "英语读写"）。

简历文本内容:
"""
%s
"""

输出 JSON 格式（请严格遵守此结构）：
{
  "techStack": ["技能1", "技能2"],
  "experience": [
    { 
      "title": "项目名称或职位", 
      "description": "项目简述", 
      "highlights": ["亮点1", "亮点2"] 
    }
  ],
  "intent": "求职意向",
  "softSkills": ["软技能1", "软技能2"]
}
`, fileContent)

	log.Printf("Sending request to AI service for resume parsing")

	resp, err := s.aiService.ChatWithTask(context.Background(), prompt, "resume")
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

	// 1. Get RAG context for the resume's tech stack and intent
	ragSvc := GetRAGService()
	var ragContext string
	if ragSvc != nil {
		query := fmt.Sprintf("%s %s", resumeData.Intent, strings.Join(resumeData.TechStack, " "))
		context, err := ragSvc.SearchKnowledgeBase(query)
		if err == nil && context != "" {
			ragContext = context
		}
	}

	prompt := fmt.Sprintf(`
根据以下简历数据和岗位知识库上下文，推荐 3 个最适合的职位。

【岗位知识库上下文】
%s

【重要要求】
1. **必须使用简体中文**输出所有内容。
2. 职位名称可以是中英文（如 "Go 后端开发" 或 "Backend Engineer"），但描述和理由必须是中文。
3. **严格根据简历的技术栈和经验推荐职位**，不要推荐与简历内容不符的职位。
4. 参考知识库中的岗位能力模型，计算匹配度。
5. 不要输出 Markdown 标记。

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
`, ragContext, string(resumeJson))

	log.Printf("Sending request to AI service for job matching")

	resp, err := s.aiService.ChatWithTask(context.Background(), prompt, "resume")
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
