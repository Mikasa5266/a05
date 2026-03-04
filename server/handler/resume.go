package handler

import (
	"io"
	"net/http"
	"your-project/service"

	"github.com/gin-gonic/gin"
)

func ParseResume(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer src.Close()

	// Read file content
	// Warning: This reads the whole file into memory. For large files, use streaming or limit size.
	// Resumes are usually small (< 5MB).
	content, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	// Basic text extraction (mock for binary formats, assumes text for now)
	// In production, use a library to extract text from PDF/DOCX
	textContent := string(content)

	svc := service.NewResumeService()
	resumeData, err := svc.ParseResume(textContent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Automatically match jobs after parsing
	matches, err := svc.MatchJobs(resumeData)
	if err != nil {
		// Log error but return resume data at least?
		// Or just fail. Let's return what we have.
		c.JSON(http.StatusOK, gin.H{
			"resume":  resumeData,
			"matches": []*string{}, // empty
			"warning": "Failed to match jobs: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"resume":  resumeData,
		"matches": matches,
	})
}
