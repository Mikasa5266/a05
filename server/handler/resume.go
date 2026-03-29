package handler

import (
	"log"
	"net/http"
	"strings"
	"your-project/model"
	"your-project/utils"

	"github.com/gin-gonic/gin"
)

// ParseResume handles resume upload and parsing
func ParseResume(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	log.Printf("Received resume file: %s, size: %d bytes", file.Filename, file.Size)

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer src.Close()

	textContent, err := utils.ExtractTextFromFile(src, file.Filename)
	if err != nil {
		log.Printf("Failed to extract text from file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extract text from file: " + err.Error()})
		return
	}

	log.Printf("Extracted text length: %d characters", len(textContent))
	// 若文本过少则警告但不阻断，交由更强的 DeepSeek 解析与容错
	if len(strings.TrimSpace(textContent)) < 50 {
		log.Printf("Warning: Extracted text is very short; the PDF might be image-based or encoded fonts")
	}

	svc := mustResumeService()
	resumeData, err := svc.ParseResume(textContent)
	if err != nil {
		log.Printf("Failed to parse resume: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse resume: " + err.Error()})
		return
	}

	log.Printf("Resume parsed successfully: %+v", resumeData)

	matches, err := svc.MatchJobs(resumeData)
	if err != nil {
		log.Printf("Failed to match jobs: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"resume":  resumeData,
			"matches": []*string{},
			"warning": "Failed to match jobs: " + err.Error(),
		})
		return
	}

	log.Printf("Job matches generated: %d matches", len(matches))

	c.JSON(http.StatusOK, gin.H{
		"resume":  resumeData,
		"matches": matches,
	})
}

// GenerateQuestions generates personalized interview questions based on resume and selected job
func GenerateQuestions(c *gin.Context) {
	var req struct {
		ResumeData *model.ResumeData `json:"resumeData" binding:"required"`
		JobTitle   string            `json:"jobTitle" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := mustResumeService()
	questions, err := svc.GenerateInterviewQuestions(req.ResumeData, req.JobTitle)
	if err != nil {
		log.Printf("Failed to generate interview questions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate questions: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"questions": questions,
	})
}

// AnalyzeResumeAuthenticity checks potential credibility risks and returns interview-verification suggestions.
func AnalyzeResumeAuthenticity(c *gin.Context) {
	var req struct {
		ResumeData *model.ResumeData `json:"resumeData" binding:"required"`
		RawText    string            `json:"rawText"`
		TargetRole string            `json:"targetRole"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := mustResumeService()
	report, err := svc.AnalyzeAuthenticity(req.ResumeData, req.RawText, req.TargetRole)
	if err != nil {
		log.Printf("Failed to analyze resume authenticity: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze authenticity: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"report": report})
}

// GetResumeOptimizationSuggestions returns actionable suggestions and rewrite demos.
func GetResumeOptimizationSuggestions(c *gin.Context) {
	var req struct {
		ResumeData *model.ResumeData `json:"resumeData" binding:"required"`
		TargetRole string            `json:"targetRole"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := mustResumeService()
	report, err := svc.GenerateOptimizationSuggestions(req.ResumeData, req.TargetRole)
	if err != nil {
		log.Printf("Failed to generate resume optimization report: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate optimization suggestions: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"report": report})
}

// GenerateResumeTemplate generates a role-specific resume template.
func GenerateResumeTemplate(c *gin.Context) {
	var req struct {
		TargetRole string `json:"targetRole" binding:"required"`
		Seniority  string `json:"seniority"`
		Language   string `json:"language"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := mustResumeService()
	tpl, err := svc.GenerateResumeTemplate(req.TargetRole, req.Seniority, req.Language)
	if err != nil {
		log.Printf("Failed to generate resume template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate resume template: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"template": tpl})
}
