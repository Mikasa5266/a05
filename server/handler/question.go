package handler

import (
	"net/http"
	"strconv"

	"your-project/service"

	"github.com/gin-gonic/gin"
)

func GetQuestions(c *gin.Context) {
	position := c.Query("position")
	difficulty := c.Query("difficulty")
	category := c.Query("category")

	questions, err := service.GetQuestions(position, difficulty, category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"questions": questions,
		"count":     len(questions),
	})
}

func GetQuestion(c *gin.Context) {
	questionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question ID"})
		return
	}

	question, err := service.GetQuestionByID(uint(questionID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"question": question})
}

func CreateQuestion(c *gin.Context) {
	var req struct {
		Title          string   `json:"title" binding:"required"`
		Content        string   `json:"content" binding:"required"`
		Position       string   `json:"position" binding:"required"`
		Difficulty     string   `json:"difficulty" binding:"required"`
		Category       string   `json:"category" binding:"required"`
		Tags           []string `json:"tags"`
		ExpectedAnswer string   `json:"expected_answer"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	question, err := service.CreateQuestion(req.Title, req.Content, req.Position, req.Difficulty, req.Category, req.Tags, req.ExpectedAnswer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Question created successfully",
		"question": question,
	})
}
