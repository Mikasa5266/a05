package handler

import (
	"net/http"
	"strings"

	"your-project/service"
	"your-project/utils"

	"github.com/gin-gonic/gin"
)

// ParseResume handles resume upload, extraction, structured analysis and persistence.
func ParseResume(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open uploaded file"})
		return
	}
	defer src.Close()

	rawText, err := utils.ExtractTextFromFile(src, file.Filename)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to extract text: " + err.Error()})
		return
	}

	analysis, record, err := mustResumeService().AnalyzeAndPersist(c.Request.Context(), service.ResumeAnalysisInput{
		UserID:   userID,
		FileName: file.Filename,
		RawText:  rawText,
		Source:   strings.TrimSpace(c.PostForm("source")),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "resume analyzed",
		"analysis": analysis,
		"record":   record,
	})
}

// GetLatestResumeAnalysis returns latest stored analysis snapshot for current user.
func GetLatestResumeAnalysis(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	analysis, record, err := mustResumeService().GetLatestAnalysis(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "resume analysis not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"analysis": analysis,
		"record":   record,
	})
}
