package handler

import (
	"net/http"
	"strconv"

	"your-project/service"

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

	// 使用AI服务进行对话
	response, err := service.AIChat(userID, req.Message, req.Context)
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

	// 使用AI服务进行对话，包含面试上下文
	response, err := service.AIChatWithInterviewContext(userID, uint(interviewID), req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "AI response generated successfully",
		"response": response,
	})
}
