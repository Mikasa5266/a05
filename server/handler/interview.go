package handler

import (
	"net/http"
	"strconv"

	"your-project/service"

	"github.com/gin-gonic/gin"
)

func StartInterview(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		Position   string `json:"position" binding:"required"`
		Difficulty string `json:"difficulty" binding:"required"`
		Mode       string `json:"mode"`  // Optional, default handled by service/model
		Style      string `json:"style"` // Optional
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults if empty
	if req.Mode == "" {
		req.Mode = "technical"
	}
	if req.Style == "" {
		req.Style = "gentle"
	}

	interview, err := service.StartInterview(userID, req.Position, req.Difficulty, req.Mode, req.Style)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Interview started successfully",
		"interview": interview,
	})
}

func GetInterview(c *gin.Context) {
	userID := c.GetUint("user_id")
	interviewID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interview ID"})
		return
	}

	interview, err := service.GetInterviewByID(userID, uint(interviewID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Interview not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"interview": interview})
}

func GetInterviews(c *gin.Context) {
	userID := c.GetUint("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	interviews, total, err := service.GetUserInterviews(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"interviews": interviews,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
	})
}

func SubmitAnswer(c *gin.Context) {
	userID := c.GetUint("user_id")
	interviewID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interview ID"})
		return
	}

	var req struct {
		QuestionID uint   `json:"question_id" binding:"required"`
		Answer     string `json:"answer" binding:"required"`
		AudioData  string `json:"audio_data,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := service.SubmitAnswer(userID, uint(interviewID), req.QuestionID, req.Answer, req.AudioData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Answer submitted successfully",
		"result":  result,
	})
}

func EndInterview(c *gin.Context) {
	userID := c.GetUint("user_id")
	interviewID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interview ID"})
		return
	}

	interview, err := service.EndInterview(userID, uint(interviewID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Interview ended successfully",
		"interview": interview,
	})
}
