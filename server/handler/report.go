package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"your-project/internal/service"

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

func normalizeReplayURL(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "/") {
		return trimmed
	}
	u, err := url.Parse(trimmed)
	if err != nil {
		return trimmed
	}
	host := strings.ToLower(u.Hostname())
	if host == "127.0.0.1" || host == "localhost" {
		path := strings.TrimSpace(u.EscapedPath())
		if path == "" {
			path = "/"
		}
		if u.RawQuery != "" {
			path += "?" + u.RawQuery
		}
		return path
	}
	return trimmed
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
	interview, _ := service.GetInterviewByID(userID, report.InterviewID)
	replayURL := ""
	if interview != nil {
		replayURL = normalizeReplayURL(interview.RecordingURL)
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
		"replay_url":       replayURL,
		"created_at":       report.CreatedAt,
		"updated_at":       report.UpdatedAt,
		"qa_details":       report.GetQADetails(),
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
	interview, _ := service.GetInterviewByID(userID, report.InterviewID)
	replayURL := ""
	if interview != nil {
		replayURL = normalizeReplayURL(interview.RecordingURL)
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
			"replay_url":       replayURL,
			"created_at":       report.CreatedAt,
			"updated_at":       report.UpdatedAt,
			"strengths":        report.GetStrengths(),
			"weaknesses":       report.GetWeaknesses(),
			"suggestions":      report.GetSuggestions(),
			"qa_details":       report.GetQADetails(),
		},
	})
}

func DownloadReport(c *gin.Context) {
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

	interview, _ := service.GetInterviewByID(userID, report.InterviewID)
	replayURL := ""
	if interview != nil {
		replayURL = normalizeReplayURL(interview.RecordingURL)
	}

	builder := &strings.Builder{}
	builder.WriteString("# 面试复盘报告\n\n")
	builder.WriteString(fmt.Sprintf("- 报告编号：%d\n", report.ID))
	builder.WriteString(fmt.Sprintf("- 岗位：%s\n", report.Position))
	builder.WriteString(fmt.Sprintf("- 难度：%s\n", report.Difficulty))
	builder.WriteString(fmt.Sprintf("- 面试时长：%d 分钟\n", report.Duration))
	builder.WriteString(fmt.Sprintf("- 报告生成时间：%s\n", report.UpdatedAt.Format(time.DateTime)))
	if replayURL != "" {
		builder.WriteString(fmt.Sprintf("- 面试回放：%s\n", replayURL))
	}
	builder.WriteString("\n")

	builder.WriteString("## 综合评分\n\n")
	builder.WriteString(fmt.Sprintf("- 综合得分：%d\n", report.AverageScore))
	builder.WriteString(fmt.Sprintf("- 技术深度：%d\n", report.TechnicalScore))
	builder.WriteString(fmt.Sprintf("- 表达沟通：%d\n", report.ExpressionScore))
	builder.WriteString(fmt.Sprintf("- 逻辑思维：%d\n", report.LogicScore))
	builder.WriteString(fmt.Sprintf("- 岗位匹配：%d\n", report.MatchingScore))
	builder.WriteString(fmt.Sprintf("- 职业素养：%d\n\n", report.BehaviorScore))

	builder.WriteString("## 总体分析\n\n")
	builder.WriteString(report.OverallAnalysis + "\n\n")

	builder.WriteString("## 优势分析\n\n")
	for _, item := range report.GetStrengths() {
		builder.WriteString("- " + item + "\n")
	}
	builder.WriteString("\n## 待改进项\n\n")
	for _, item := range report.GetWeaknesses() {
		builder.WriteString("- " + item + "\n")
	}
	builder.WriteString("\n## 优化建议\n\n")
	for _, item := range report.GetSuggestions() {
		builder.WriteString("- " + item + "\n")
	}

	qaDetails := report.GetQADetails()
	if len(qaDetails) > 0 {
		builder.WriteString("\n## 详细题目回顾\n\n")
		for idx, item := range qaDetails {
			builder.WriteString(fmt.Sprintf("### 题目 %d：%s\n\n", idx+1, item.Question))
			builder.WriteString("- 候选人回答摘要：\n")
			builder.WriteString(item.UserAnswer + "\n\n")
			builder.WriteString("- 优化参考回答：\n")
			builder.WriteString(item.OptimizedAnswer + "\n\n")
			if len(item.KeyImprovements) > 0 {
				builder.WriteString("- 关键改进点：\n")
				for _, improvement := range item.KeyImprovements {
					builder.WriteString("  - " + improvement + "\n")
				}
				builder.WriteString("\n")
			}
		}
	}

	filename := fmt.Sprintf("report_%d_%d.md", report.ID, time.Now().Unix())
	c.Header("Content-Type", "text/markdown; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.String(http.StatusOK, builder.String())
}
