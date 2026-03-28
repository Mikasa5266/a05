package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func AIChat(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		Message string `json:"message" binding:"required"`
		Context string `json:"context,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	aiSvc := mustAIService()
	response, err := aiSvc.AIChat(c.Request.Context(), userID, req.Message, req.Context)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "AI response generated successfully",
		"response": response,
	})
}

func AIChatWithInterviewContext(c *gin.Context) {
	userID := c.GetUint("user_id")
	interviewID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interview ID"})
		return
	}

	var req struct {
		Message string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	aiSvc := mustAIService()
	response, err := aiSvc.AIChatWithInterviewContext(c.Request.Context(), userID, uint(interviewID), req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "AI response generated successfully",
		"response": response,
	})
}
