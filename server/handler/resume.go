package handler

import (
	"log"
	"net/http"
	"strings"
	"your-project/service"
	"your-project/utils"

	"github.com/gin-gonic/gin"
)

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

	svc := service.NewResumeService()
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
