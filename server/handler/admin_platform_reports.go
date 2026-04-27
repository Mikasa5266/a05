package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"your-project/config"
	"your-project/internal/model"
	"your-project/internal/repository"
	"your-project/internal/service"

	"github.com/gin-gonic/gin"
)

func AdminPlatformListSecurityReports(c *gin.Context) {
	status := strings.ToLower(strings.TrimSpace(c.Query("status")))
	targetType := strings.ToLower(strings.TrimSpace(c.Query("target_type")))
	keyword := strings.TrimSpace(c.Query("keyword"))

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := repository.GetDB().Model(&model.SecurityReport{})
	if status != "" {
		if _, ok := allowedReportStatuses[status]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
			return
		}
		query = query.Where("status = ?", status)
	}
	if targetType != "" {
		if _, ok := allowedReportTargets[targetType]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target_type"})
			return
		}
		query = query.Where("target_type = ?", targetType)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("reason LIKE ? OR description LIKE ? OR contact LIKE ?", like, like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query reports"})
		return
	}

	var reports []model.SecurityReport
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&reports).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query reports"})
		return
	}

	reportItems := make([]gin.H, 0, len(reports))
	for _, item := range reports {
		reportItems = append(reportItems, gin.H{
			"id":               item.ID,
			"reporter_user_id": item.ReporterUserID,
			"reporter_role":    item.ReporterRole,
			"target_type":      item.TargetType,
			"target_id":        item.TargetID,
			"reason":           item.Reason,
			"description":      item.Description,
			"contact":          item.Contact,
			"status":           item.Status,
			"handle_note":      item.HandleNote,
			"handled_by":       item.HandledBy,
			"handled_at":       item.HandledAt,
			"evidence":         parseReportEvidence(item.EvidenceJSON),
			"created_at":       item.CreatedAt,
			"updated_at":       item.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"reports":   reportItems,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func AdminPlatformHandleSecurityReport(c *gin.Context) {
	adminUsername := strings.TrimSpace(c.GetString("admin_platform_user"))
	reportID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || reportID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report id"})
		return
	}

	var req struct {
		Status     string `json:"status" binding:"required"`
		HandleNote string `json:"handle_note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	status := strings.ToLower(strings.TrimSpace(req.Status))
	if _, ok := allowedReportStatuses[status]; !ok || status == "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status only supports processing/resolved/rejected"})
		return
	}

	note := strings.TrimSpace(req.HandleNote)
	if err := service.ValidateSafeTextField("handle_note", note); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len([]rune(note)) > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "handle note is too long"})
		return
	}

	db := repository.GetDB()
	var report model.SecurityReport
	if err := db.First(&report, uint(reportID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
		return
	}

	now := time.Now()
	report.Status = status
	report.HandleNote = note
	report.HandledBy = nil
	report.HandledAt = &now

	if err := db.Save(&report).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to handle report"})
		return
	}

	service.RecordSecurityAudit(c, 0, "security_report_handle_admin_platform", "success", http.StatusOK, map[string]interface{}{
		"report_id":        report.ID,
		"status":           status,
		"admin_platform":   adminUsername,
		"reporter_role":    report.ReporterRole,
		"reporter_user_id": report.ReporterUserID,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "report handled",
		"report": gin.H{
			"id":          report.ID,
			"status":      report.Status,
			"handle_note": report.HandleNote,
			"handled_at":  report.HandledAt,
		},
	})
}

func AdminPlatformComplianceOverview(c *gin.Context) {
	cfg := config.GetConfig()

	logRetentionDays := config.DefaultSecurityLogRetentionDays
	patrolIntervalMinutes := config.DefaultSecurityPatrolIntervalMinute
	patrolScanLimit := config.DefaultSecurityPatrolScanLimit

	responsible := config.ContactConfig{
		Name:  "刘宇恒",
		Phone: "17707284972",
		Email: "1140893485@qq.com",
	}
	emergency := config.ContactConfig{
		Name:  "刘宇恒",
		Phone: "17707284972",
		Email: "1140893485@qq.com",
	}

	if cfg != nil {
		if cfg.Security.LogRetentionDays > 0 {
			logRetentionDays = cfg.Security.LogRetentionDays
		}
		if cfg.Security.PatrolIntervalMinutes > 0 {
			patrolIntervalMinutes = cfg.Security.PatrolIntervalMinutes
		}
		if cfg.Security.PatrolScanLimit > 0 {
			patrolScanLimit = cfg.Security.PatrolScanLimit
		}
		if strings.TrimSpace(cfg.Security.ResponsiblePerson.Phone) != "" {
			responsible.Phone = strings.TrimSpace(cfg.Security.ResponsiblePerson.Phone)
		}
		if strings.TrimSpace(cfg.Security.ResponsiblePerson.Email) != "" {
			responsible.Email = strings.TrimSpace(cfg.Security.ResponsiblePerson.Email)
		}
		if strings.TrimSpace(cfg.Security.EmergencyContact.Phone) != "" {
			emergency.Phone = strings.TrimSpace(cfg.Security.EmergencyContact.Phone)
		}
		if strings.TrimSpace(cfg.Security.EmergencyContact.Email) != "" {
			emergency.Email = strings.TrimSpace(cfg.Security.EmergencyContact.Email)
		}
	}

	db := repository.GetDB()
	var total, pending, processing, resolved, rejected int64
	_ = db.Model(&model.SecurityReport{}).Count(&total).Error
	_ = db.Model(&model.SecurityReport{}).Where("status = ?", "pending").Count(&pending).Error
	_ = db.Model(&model.SecurityReport{}).Where("status = ?", "processing").Count(&processing).Error
	_ = db.Model(&model.SecurityReport{}).Where("status = ?", "resolved").Count(&resolved).Error
	_ = db.Model(&model.SecurityReport{}).Where("status = ?", "rejected").Count(&rejected).Error

	c.JSON(http.StatusOK, gin.H{
		"report_pipeline": gin.H{
			"main_program_submit_endpoint":    "/api/v1/community/reports",
			"admin_platform_list_endpoint":    "/admin-platform/api/reports",
			"admin_platform_dispose_endpoint": "/admin-platform/api/reports/:id/dispose",
			"shared_storage_table":            "security_reports",
			"total_reports":                   total,
			"pending_reports":                 pending,
			"processing_reports":              processing,
			"resolved_reports":                resolved,
			"rejected_reports":                rejected,
		},
		"realname_verification": gin.H{
			"enabled":                 true,
			"register_requires_phone": true,
			"register_requires_id_no": true,
			"register_requires_name":  true,
		},
		"log_retention": gin.H{
			"days":                 logRetentionDays,
			"meets_six_months_min": logRetentionDays >= 180,
		},
		"security_protection": gin.H{
			"sql_injection_guard": true,
			"xss_guard":           true,
		},
		"illegal_content_governance": gin.H{
			"keyword_filter":          true,
			"patrol_enabled":          true,
			"patrol_interval_minutes": patrolIntervalMinutes,
			"patrol_scan_limit":       patrolScanLimit,
			"manual_disposal_enabled": true,
		},
		"security_contacts": gin.H{
			"responsible_person": responsible,
			"emergency_contact":  emergency,
			"complaint_phone":    fallbackString(responsible.Phone, emergency.Phone),
			"complaint_email":    fallbackString(responsible.Email, emergency.Email),
		},
		"moderation_capability": gin.H{
			"list_posts_endpoint":     "/admin-platform/api/moderation/posts",
			"list_comments_endpoint":  "/admin-platform/api/moderation/comments",
			"list_users_endpoint":     "/admin-platform/api/moderation/users",
			"delete_post_endpoint":    "/admin-platform/api/moderation/posts/:id",
			"delete_comment_endpoint": "/admin-platform/api/moderation/comments/:id",
			"delete_user_endpoint":    "/admin-platform/api/moderation/users/:id",
			"audit_logs_endpoint":     "/admin-platform/api/audit-logs",
			"audit_log_table":         "audit_logs",
		},
	})
}

func parseReportEvidence(raw string) map[string]interface{} {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" {
		return map[string]interface{}{}
	}

	result := map[string]interface{}{}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return map[string]interface{}{}
	}
	return result
}

func fallbackString(primary, backup string) string {
	p := strings.TrimSpace(primary)
	if p != "" {
		return p
	}
	return strings.TrimSpace(backup)
}
