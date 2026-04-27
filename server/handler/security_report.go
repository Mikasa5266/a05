package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"your-project/internal/model"
	"your-project/internal/repository"
	"your-project/internal/service"

	"github.com/gin-gonic/gin"
)

var allowedReportTargets = map[string]struct{}{
	"post":    {},
	"comment": {},
	"user":    {},
	"other":   {},
}

var allowedReportStatuses = map[string]struct{}{
	"pending":    {},
	"processing": {},
	"resolved":   {},
	"rejected":   {},
}

func CreateSecurityReport(c *gin.Context) {
	userID := c.GetUint("user_id")
	role := strings.TrimSpace(c.GetString("role"))

	if !service.AllowRateLimitedAction(userID, "security_report_submit", 5, 5*time.Minute) {
		service.RecordSecurityAudit(c, userID, "security_report_submit", "blocked", http.StatusTooManyRequests, map[string]interface{}{
			"reason": "rate_limit",
		})
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "提交过于频繁，请稍后再试"})
		return
	}

	var req struct {
		TargetType  string `json:"target_type" binding:"required"`
		TargetID    uint   `json:"target_id"`
		Reason      string `json:"reason" binding:"required"`
		Description string `json:"description"`
		Contact     string `json:"contact"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		service.RecordSecurityAudit(c, userID, "security_report_submit", "failed", http.StatusBadRequest, map[string]interface{}{
			"reason": "invalid_payload",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	targetType := strings.ToLower(strings.TrimSpace(req.TargetType))
	if _, ok := allowedReportTargets[targetType]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_type 仅支持 post/comment/user/other"})
		return
	}
	reason := strings.TrimSpace(req.Reason)
	description := strings.TrimSpace(req.Description)
	contact := strings.TrimSpace(req.Contact)
	if err := service.ValidateSafeTextField("reason", reason); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := service.ValidateSafeTextField("description", description); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := service.ValidateSafeTextField("contact", contact); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "举报原因不能为空"})
		return
	}
	if len([]rune(reason)) > 120 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "举报原因过长"})
		return
	}
	if len([]rune(description)) > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "补充说明过长"})
		return
	}
	if len([]rune(contact)) > 120 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "联系方式过长"})
		return
	}

	reason = service.SanitizeUserInput(reason)
	description = service.SanitizeUserInput(description)
	contact = service.SanitizeUserInput(contact)

	db := repository.GetDB()
	evidence := buildSecurityReportEvidence(targetType, req.TargetID)
	evidenceJSON := "{}"
	if payload, err := json.Marshal(evidence); err == nil {
		evidenceJSON = string(payload)
	}

	report := model.SecurityReport{
		ReporterUserID: userID,
		ReporterRole:   role,
		TargetType:     targetType,
		TargetID:       req.TargetID,
		Reason:         reason,
		Description:    description,
		Contact:        contact,
		EvidenceJSON:   evidenceJSON,
		Status:         "pending",
	}

	if err := db.Create(&report).Error; err != nil {
		service.RecordSecurityAudit(c, userID, "security_report_submit", "failed", http.StatusInternalServerError, map[string]interface{}{
			"reason":      "db_error",
			"target_type": targetType,
			"target_id":   req.TargetID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "举报提交失败，请稍后重试"})
		return
	}

	service.RecordSecurityAudit(c, userID, "security_report_submit", "success", http.StatusCreated, map[string]interface{}{
		"report_id":   report.ID,
		"target_type": targetType,
		"target_id":   req.TargetID,
	})

	c.JSON(http.StatusCreated, gin.H{
		"message": "举报已提交，我们将在 24 小时内审核处理",
		"report": gin.H{
			"id":          report.ID,
			"target_type": report.TargetType,
			"target_id":   report.TargetID,
			"reason":      report.Reason,
			"description": report.Description,
			"contact":     report.Contact,
			"status":      report.Status,
			"created_at":  report.CreatedAt,
		},
	})
}

func GetMySecurityReports(c *gin.Context) {
	userID := c.GetUint("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	db := repository.GetDB()
	query := db.Model(&model.SecurityReport{}).Where("reporter_user_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	var reports []model.SecurityReport
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&reports).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reports":     reports,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"has_next":    int64(page*pageSize) < total,
		"has_prev":    page > 1,
		"reporter_id": userID,
	})
}

func GetSecurityReportsForAdmin(c *gin.Context) {
	status := strings.ToLower(strings.TrimSpace(c.Query("status")))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	db := repository.GetDB()
	query := db.Model(&model.SecurityReport{})
	if status != "" {
		if _, ok := allowedReportStatuses[status]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "状态值无效"})
			return
		}
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	var reports []model.SecurityReport
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&reports).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reports":   reports,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func HandleSecurityReportForAdmin(c *gin.Context) {
	adminID := c.GetUint("user_id")
	reportID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || reportID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的举报 ID"})
		return
	}

	var req struct {
		Status     string `json:"status" binding:"required"`
		HandleNote string `json:"handle_note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := strings.ToLower(strings.TrimSpace(req.Status))
	if _, ok := allowedReportStatuses[status]; !ok || status == "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "状态仅支持 processing/resolved/rejected"})
		return
	}

	note := strings.TrimSpace(req.HandleNote)
	if len([]rune(note)) > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "处理备注过长"})
		return
	}

	db := repository.GetDB()
	var report model.SecurityReport
	if err := db.First(&report, uint(reportID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "举报记录不存在"})
		return
	}

	now := time.Now()
	report.Status = status
	report.HandleNote = note
	report.HandledBy = &adminID
	report.HandledAt = &now

	if err := db.Save(&report).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "处理失败"})
		return
	}

	service.RecordSecurityAudit(c, adminID, "security_report_handle", "success", http.StatusOK, map[string]interface{}{
		"report_id": report.ID,
		"status":    status,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "处理成功",
		"report":  report,
	})
}

func RunSecurityPatrolForAdmin(c *gin.Context) {
	adminID := c.GetUint("user_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "0"))

	result, err := service.RunIllegalContentPatrol(limit)
	if err != nil {
		service.RecordSecurityAudit(c, adminID, "security_patrol_run", "failed", http.StatusInternalServerError, map[string]interface{}{
			"reason": "patrol_error",
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "巡查执行失败"})
		return
	}

	service.RecordSecurityAudit(c, adminID, "security_patrol_run", "success", http.StatusOK, map[string]interface{}{
		"scanned_posts":    result.ScannedPosts,
		"scanned_comments": result.ScannedComments,
		"hit_posts":        result.HitPosts,
		"hit_comments":     result.HitComments,
		"created_reports":  result.CreatedReports,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "巡查执行完成",
		"result":  result,
	})
}

func buildSecurityReportEvidence(targetType string, targetID uint) map[string]interface{} {
	evidence := map[string]interface{}{
		"target_type": targetType,
		"target_id":   targetID,
	}
	if targetID == 0 {
		return evidence
	}

	switch targetType {
	case "post":
		var post model.CommunityPost
		if err := repository.GetDB().Select("id", "title", "user_id", "content").First(&post, targetID).Error; err == nil {
			evidence["post_title"] = post.Title
			evidence["post_user_id"] = post.UserID
			evidence["post_excerpt"] = serviceExcerpt(post.Content, 160)
		}
	case "comment":
		var comment model.PostComment
		if err := repository.GetDB().Select("id", "post_id", "user_id", "content").First(&comment, targetID).Error; err == nil {
			evidence["post_id"] = comment.PostID
			evidence["comment_user_id"] = comment.UserID
			evidence["comment_excerpt"] = serviceExcerpt(comment.Content, 160)
		}
	}

	return evidence
}

func serviceExcerpt(input string, max int) string {
	trimmed := strings.TrimSpace(input)
	if max <= 0 {
		return ""
	}
	runes := []rune(trimmed)
	if len(runes) <= max {
		return trimmed
	}
	return string(runes[:max])
}
