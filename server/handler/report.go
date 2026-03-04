package handler

import (
	"net/http"
	"strconv"

	"your-project/service"

	"github.com/gin-gonic/gin"
)

func buildReportResponse(report interface {
	GetStrengths() []string
	GetWeaknesses() []string
	GetSuggestions() []string
}) gin.H {
	return gin.H{
		"strengths":   report.GetStrengths(),
		"weaknesses":  report.GetWeaknesses(),
		"suggestions": report.GetSuggestions(),
	}
}

func GetReports(c *gin.Context) {
	userID := c.GetUint("user_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	reports, total, err := service.GetUserReports(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reports":    reports,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_page": (int(total) + pageSize - 1) / pageSize,
	})
}

func GetReport(c *gin.Context) {
	userID := c.GetUint("user_id")
	reportID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}

	report, err := service.GetReportByID(userID, uint(reportID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	resp := gin.H{
		"id":               report.ID,
		"user_id":          report.UserID,
		"interview_id":     report.InterviewID,
		"position":         report.Position,
		"difficulty":       report.Difficulty,
		"total_questions":  report.TotalQuestions,
		"average_score":    report.AverageScore,
		"overall_analysis": report.OverallAnalysis,
		"technical_score":  report.TechnicalScore,
		"expression_score": report.ExpressionScore,
		"logic_score":      report.LogicScore,
		"matching_score":   report.MatchingScore,
		"behavior_score":   report.BehaviorScore,
		"start_time":       report.StartTime,
		"end_time":         report.EndTime,
		"duration":         report.Duration,
		"created_at":       report.CreatedAt,
		"updated_at":       report.UpdatedAt,
	}
	for k, v := range buildReportResponse(report) {
		resp[k] = v
	}

	c.JSON(http.StatusOK, gin.H{"report": resp})
}

func GenerateReport(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		InterviewID uint `json:"interview_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report, err := service.GenerateInterviewReport(userID, req.InterviewID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Report generated successfully",
		"report": gin.H{
			"id":               report.ID,
			"user_id":          report.UserID,
			"interview_id":     report.InterviewID,
			"position":         report.Position,
			"difficulty":       report.Difficulty,
			"total_questions":  report.TotalQuestions,
			"average_score":    report.AverageScore,
			"overall_analysis": report.OverallAnalysis,
			"technical_score":  report.TechnicalScore,
			"expression_score": report.ExpressionScore,
			"logic_score":      report.LogicScore,
			"matching_score":   report.MatchingScore,
			"behavior_score":   report.BehaviorScore,
			"start_time":       report.StartTime,
			"end_time":         report.EndTime,
			"duration":         report.Duration,
			"created_at":       report.CreatedAt,
			"updated_at":       report.UpdatedAt,
			"strengths":        report.GetStrengths(),
			"weaknesses":       report.GetWeaknesses(),
			"suggestions":      report.GetSuggestions(),
		},
	})
}
