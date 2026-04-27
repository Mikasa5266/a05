package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"your-project/internal/model"
	"your-project/internal/repository"
	"your-project/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	adminPlatformDefaultPageSize = 20
	adminPlatformMaxPageSize     = 100
)

type adminUserDisposeResult struct {
	UserID          uint
	Username        string
	Email           string
	DeletedPosts    int64
	DeletedComments int64
	DeletedLikes    int64
}

func AdminPlatformListModerationPosts(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("keyword"))
	page, pageSize := parseAdminPlatformPagination(c)

	db := repository.GetDB()
	query := db.Model(&model.CommunityPost{})
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("title LIKE ? OR content LIKE ? OR author LIKE ?", like, like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query posts"})
		return
	}

	var posts []model.CommunityPost
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query posts"})
		return
	}

	items := make([]gin.H, 0, len(posts))
	for _, post := range posts {
		items = append(items, gin.H{
			"id":           post.ID,
			"user_id":      post.UserID,
			"author":       post.Author,
			"title":        post.Title,
			"company":      post.Company,
			"position":     post.Position,
			"content":      excerptText(post.Content, 200),
			"likes":        post.Likes,
			"comments":     post.Comments,
			"views":        post.Views,
			"created_at":   post.CreatedAt,
			"updated_at":   post.UpdatedAt,
			"deleted_at":   post.DeletedAt,
			"offer_status": post.OfferStatus,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"posts":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func AdminPlatformDeleteModerationPost(c *gin.Context) {
	adminUsername := strings.TrimSpace(c.GetString("admin_platform_user"))
	postID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	db := repository.GetDB()
	var deletedComments int64
	err := db.Transaction(func(tx *gorm.DB) error {
		var txErr error
		deletedComments, txErr = adminDeletePostTx(tx, postID)
		return txErr
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete post"})
		return
	}

	service.RecordSecurityAudit(c, 0, "admin_platform_delete_post", "success", http.StatusOK, map[string]interface{}{
		"post_id":           postID,
		"deleted_comments":  deletedComments,
		"admin_platform":    adminUsername,
		"dispose_operation": true,
	})
	service.RecordAuditLog(c, "admin_platform", adminUsername, "delete_post", "success", "post", postID, map[string]interface{}{
		"deleted_comments": deletedComments,
	})

	c.JSON(http.StatusOK, gin.H{
		"message":           "post deleted",
		"post_id":           postID,
		"deleted_comments":  deletedComments,
		"dispose_operation": true,
	})
}

func AdminPlatformListModerationComments(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("keyword"))
	postID := strings.TrimSpace(c.Query("post_id"))
	page, pageSize := parseAdminPlatformPagination(c)

	db := repository.GetDB()
	query := db.Model(&model.PostComment{})
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("content LIKE ? OR author LIKE ?", like, like)
	}
	if postID != "" {
		idValue, err := strconv.ParseUint(postID, 10, 64)
		if err != nil || idValue == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post_id"})
			return
		}
		query = query.Where("post_id = ?", uint(idValue))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query comments"})
		return
	}

	var comments []model.PostComment
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query comments"})
		return
	}

	postTitleMap := loadPostTitleMap(db, comments)
	items := make([]gin.H, 0, len(comments))
	for _, comment := range comments {
		items = append(items, gin.H{
			"id":         comment.ID,
			"post_id":    comment.PostID,
			"post_title": postTitleMap[comment.PostID],
			"user_id":    comment.UserID,
			"author":     comment.Author,
			"content":    comment.Content,
			"created_at": comment.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"comments":  items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func AdminPlatformDeleteModerationComment(c *gin.Context) {
	adminUsername := strings.TrimSpace(c.GetString("admin_platform_user"))
	commentID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	db := repository.GetDB()
	var postID uint
	err := db.Transaction(func(tx *gorm.DB) error {
		var txErr error
		postID, txErr = adminDeleteCommentTx(tx, commentID)
		return txErr
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete comment"})
		return
	}

	service.RecordSecurityAudit(c, 0, "admin_platform_delete_comment", "success", http.StatusOK, map[string]interface{}{
		"comment_id":        commentID,
		"post_id":           postID,
		"admin_platform":    adminUsername,
		"dispose_operation": true,
	})
	service.RecordAuditLog(c, "admin_platform", adminUsername, "delete_comment", "success", "comment", commentID, map[string]interface{}{
		"post_id": postID,
	})

	c.JSON(http.StatusOK, gin.H{
		"message":           "comment deleted",
		"comment_id":        commentID,
		"post_id":           postID,
		"dispose_operation": true,
	})
}

func AdminPlatformListModerationUsers(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("keyword"))
	role := strings.ToLower(strings.TrimSpace(c.Query("role")))
	page, pageSize := parseAdminPlatformPagination(c)

	db := repository.GetDB()
	query := db.Model(&model.User{})
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("username LIKE ? OR email LIKE ? OR real_name LIKE ? OR phone LIKE ?", like, like, like, like)
	}
	if role != "" {
		query = query.Where("role = ?", role)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query users"})
		return
	}

	var users []model.User
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query users"})
		return
	}

	items := make([]gin.H, 0, len(users))
	for _, user := range users {
		items = append(items, gin.H{
			"id":                    user.ID,
			"username":              user.Username,
			"email":                 user.Email,
			"real_name":             user.RealName,
			"phone":                 user.Phone,
			"role":                  user.Role,
			"real_name_verified":    user.RealNameVerified,
			"real_name_verified_at": user.RealNameVerifiedAt,
			"created_at":            user.CreatedAt,
			"updated_at":            user.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"users":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func AdminPlatformDeleteModerationUser(c *gin.Context) {
	adminUsername := strings.TrimSpace(c.GetString("admin_platform_user"))
	userID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	db := repository.GetDB()
	result := adminUserDisposeResult{}
	err := db.Transaction(func(tx *gorm.DB) error {
		var txErr error
		result, txErr = adminDeleteUserTx(tx, userID)
		return txErr
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}

	service.RecordSecurityAudit(c, 0, "admin_platform_delete_user", "success", http.StatusOK, map[string]interface{}{
		"target_user_id":    result.UserID,
		"target_username":   result.Username,
		"deleted_posts":     result.DeletedPosts,
		"deleted_comments":  result.DeletedComments,
		"deleted_likes":     result.DeletedLikes,
		"admin_platform":    adminUsername,
		"dispose_operation": true,
	})
	service.RecordAuditLog(c, "admin_platform", adminUsername, "delete_user", "success", "user", result.UserID, map[string]interface{}{
		"target_username":  result.Username,
		"deleted_posts":    result.DeletedPosts,
		"deleted_comments": result.DeletedComments,
		"deleted_likes":    result.DeletedLikes,
	})

	c.JSON(http.StatusOK, gin.H{
		"message":          "user deleted",
		"user_id":          result.UserID,
		"username":         result.Username,
		"deleted_posts":    result.DeletedPosts,
		"deleted_comments": result.DeletedComments,
		"deleted_likes":    result.DeletedLikes,
	})
}

func AdminPlatformDisposeSecurityReport(c *gin.Context) {
	adminUsername := strings.TrimSpace(c.GetString("admin_platform_user"))
	reportID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req struct {
		Action     string `json:"action" binding:"required"`
		HandleNote string `json:"handle_note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	action := strings.ToLower(strings.TrimSpace(req.Action))
	switch action {
	case "delete_post", "delete_comment", "delete_user":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "action only supports delete_post/delete_comment/delete_user"})
		return
	}

	note, err := normalizeAdminHandleNote(req.HandleNote)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := repository.GetDB()
	var report model.SecurityReport
	var resultMessage string
	var disposeDetail map[string]interface{}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&report, reportID).Error; err != nil {
			return err
		}

		switch action {
		case "delete_post":
			if report.TargetType != "post" || report.TargetID == 0 {
				return errors.New("report target is not a post")
			}
			deletedComments, err := adminDeletePostTx(tx, report.TargetID)
			if err != nil {
				return err
			}
			disposeDetail = map[string]interface{}{
				"post_id":          report.TargetID,
				"deleted_comments": deletedComments,
			}
			resultMessage = "post deleted and report resolved"
		case "delete_comment":
			if report.TargetType != "comment" || report.TargetID == 0 {
				return errors.New("report target is not a comment")
			}
			postID, err := adminDeleteCommentTx(tx, report.TargetID)
			if err != nil {
				return err
			}
			disposeDetail = map[string]interface{}{
				"comment_id": report.TargetID,
				"post_id":    postID,
			}
			resultMessage = "comment deleted and report resolved"
		case "delete_user":
			targetUserID := report.TargetID
			if report.TargetType != "user" || targetUserID == 0 {
				targetUserID = extractUserIDFromEvidence(parseReportEvidence(report.EvidenceJSON))
			}
			if targetUserID == 0 {
				return errors.New("report target is not a user")
			}
			disposeResult, err := adminDeleteUserTx(tx, targetUserID)
			if err != nil {
				return err
			}
			disposeDetail = map[string]interface{}{
				"user_id":          disposeResult.UserID,
				"username":         disposeResult.Username,
				"deleted_posts":    disposeResult.DeletedPosts,
				"deleted_comments": disposeResult.DeletedComments,
				"deleted_likes":    disposeResult.DeletedLikes,
			}
			resultMessage = "user deleted and report resolved"
		}

		now := time.Now()
		report.Status = "resolved"
		report.HandleNote = buildDisposeHandleNote(note, action)
		report.HandledBy = nil
		report.HandledAt = &now
		return tx.Save(&report).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "report or target not found"})
			return
		}
		if strings.Contains(err.Error(), "report target is not") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to dispose report target"})
		return
	}

	service.RecordSecurityAudit(c, 0, "security_report_dispose_admin_platform", "success", http.StatusOK, map[string]interface{}{
		"report_id":      report.ID,
		"action":         action,
		"target_type":    report.TargetType,
		"target_id":      report.TargetID,
		"admin_platform": adminUsername,
		"detail":         disposeDetail,
	})
	service.RecordAuditLog(c, "admin_platform", adminUsername, "dispose_report_target", "success", report.TargetType, report.TargetID, map[string]interface{}{
		"report_id": report.ID,
		"action":    action,
		"detail":    disposeDetail,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": resultMessage,
		"report": gin.H{
			"id":          report.ID,
			"status":      report.Status,
			"handle_note": report.HandleNote,
			"handled_at":  report.HandledAt,
		},
		"dispose_detail": disposeDetail,
	})
}

func AdminPlatformRunSecurityPatrol(c *gin.Context) {
	adminUsername := strings.TrimSpace(c.GetString("admin_platform_user"))

	var req struct {
		Limit int `json:"limit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	limit := req.Limit
	if limit <= 0 {
		limit, _ = strconv.Atoi(strings.TrimSpace(c.Query("limit")))
	}

	result, err := service.RunIllegalContentPatrol(limit)
	if err != nil {
		service.RecordSecurityAudit(c, 0, "security_patrol_run_admin_platform", "failed", http.StatusInternalServerError, map[string]interface{}{
			"admin_platform": adminUsername,
		})
		service.RecordAuditLog(c, "admin_platform", adminUsername, "run_patrol", "failed", "system", 0, map[string]interface{}{
			"reason": "run_patrol_failed",
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to run patrol"})
		return
	}

	service.RecordSecurityAudit(c, 0, "security_patrol_run_admin_platform", "success", http.StatusOK, map[string]interface{}{
		"admin_platform":   adminUsername,
		"scanned_posts":    result.ScannedPosts,
		"scanned_comments": result.ScannedComments,
		"hit_posts":        result.HitPosts,
		"hit_comments":     result.HitComments,
		"created_reports":  result.CreatedReports,
	})
	service.RecordAuditLog(c, "admin_platform", adminUsername, "run_patrol", "success", "system", 0, map[string]interface{}{
		"scanned_posts":    result.ScannedPosts,
		"scanned_comments": result.ScannedComments,
		"hit_posts":        result.HitPosts,
		"hit_comments":     result.HitComments,
		"created_reports":  result.CreatedReports,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "patrol run finished",
		"result":  result,
	})
}

func AdminPlatformListSecurityAuditLogs(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("keyword"))
	action := strings.TrimSpace(c.Query("action"))
	outcome := strings.TrimSpace(c.Query("outcome"))
	page, pageSize := parseAdminPlatformPagination(c)

	db := repository.GetDB()
	query := db.Model(&model.AuditLog{})
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if outcome != "" {
		query = query.Where("outcome = ?", outcome)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("path LIKE ? OR detail_json LIKE ? OR actor_name LIKE ? OR target_type LIKE ?", like, like, like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query audit logs"})
		return
	}

	var logs []model.AuditLog
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query audit logs"})
		return
	}

	items := make([]gin.H, 0, len(logs))
	for _, logItem := range logs {
		items = append(items, gin.H{
			"id":          logItem.ID,
			"actor_type":  logItem.ActorType,
			"actor_name":  logItem.ActorName,
			"action":      logItem.Action,
			"outcome":     logItem.Outcome,
			"method":      logItem.Method,
			"path":        logItem.Path,
			"target_type": logItem.TargetType,
			"target_id":   logItem.TargetID,
			"source_ip":   service.RevealSecurityLogField(logItem.SourceIP),
			"detail":      parseAuditLogDetail(logItem.DetailJSON),
			"created_at":  logItem.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"audit_logs": items,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
	})
}

func parseAdminPlatformPagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("page", "1")))
	pageSize, _ := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("page_size", strconv.Itoa(adminPlatformDefaultPageSize))))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > adminPlatformMaxPageSize {
		pageSize = adminPlatformDefaultPageSize
	}
	return page, pageSize
}

func parseUintParam(c *gin.Context, key string) (uint, bool) {
	raw := strings.TrimSpace(c.Param(key))
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return uint(id), true
}

func normalizeAdminHandleNote(note string) (string, error) {
	trimmed := strings.TrimSpace(note)
	if err := service.ValidateSafeTextField("handle_note", trimmed); err != nil {
		return "", err
	}
	if len([]rune(trimmed)) > 1000 {
		return "", errors.New("handle note is too long")
	}
	return trimmed, nil
}

func buildDisposeHandleNote(note, action string) string {
	actionText := "处置动作：" + action
	if note == "" {
		return actionText
	}
	return actionText + "；备注：" + note
}

func excerptText(input string, max int) string {
	trimmed := strings.TrimSpace(input)
	if max <= 0 {
		return ""
	}
	runes := []rune(trimmed)
	if len(runes) <= max {
		return trimmed
	}
	return string(runes[:max]) + "..."
}

func loadPostTitleMap(db *gorm.DB, comments []model.PostComment) map[uint]string {
	postIDs := make([]uint, 0, len(comments))
	seen := make(map[uint]struct{}, len(comments))
	for _, comment := range comments {
		if comment.PostID == 0 {
			continue
		}
		if _, ok := seen[comment.PostID]; ok {
			continue
		}
		seen[comment.PostID] = struct{}{}
		postIDs = append(postIDs, comment.PostID)
	}
	if len(postIDs) == 0 {
		return map[uint]string{}
	}

	var posts []model.CommunityPost
	if err := db.Select("id", "title").Where("id IN ?", postIDs).Find(&posts).Error; err != nil {
		return map[uint]string{}
	}

	result := make(map[uint]string, len(posts))
	for _, post := range posts {
		result[post.ID] = post.Title
	}
	return result
}

func adminDeletePostTx(tx *gorm.DB, postID uint) (int64, error) {
	var post model.CommunityPost
	if err := tx.First(&post, postID).Error; err != nil {
		return 0, err
	}

	if err := tx.Where("post_id = ?", post.ID).Delete(&model.PostLike{}).Error; err != nil {
		return 0, err
	}
	commentsResult := tx.Where("post_id = ?", post.ID).Delete(&model.PostComment{})
	if commentsResult.Error != nil {
		return 0, commentsResult.Error
	}

	if err := tx.Delete(&post).Error; err != nil {
		return 0, err
	}
	return commentsResult.RowsAffected, nil
}

func adminDeleteCommentTx(tx *gorm.DB, commentID uint) (uint, error) {
	var comment model.PostComment
	if err := tx.First(&comment, commentID).Error; err != nil {
		return 0, err
	}

	if err := tx.Delete(&comment).Error; err != nil {
		return 0, err
	}

	if comment.PostID > 0 {
		if err := tx.Model(&model.CommunityPost{}).
			Where("id = ? AND comments > 0", comment.PostID).
			UpdateColumn("comments", gorm.Expr("comments - ?", 1)).Error; err != nil {
			return 0, err
		}
	}

	return comment.PostID, nil
}

func adminDeleteUserTx(tx *gorm.DB, userID uint) (adminUserDisposeResult, error) {
	result := adminUserDisposeResult{}

	var user model.User
	if err := tx.First(&user, userID).Error; err != nil {
		return result, err
	}
	result.UserID = user.ID
	result.Username = user.Username
	result.Email = user.Email

	type commentAgg struct {
		PostID uint
		Count  int64
	}
	aggregates := make([]commentAgg, 0)
	if err := tx.Model(&model.PostComment{}).
		Select("post_id, COUNT(*) AS count").
		Where("user_id = ?", user.ID).
		Group("post_id").
		Scan(&aggregates).Error; err != nil {
		return result, err
	}

	deleteCommentResult := tx.Where("user_id = ?", user.ID).Delete(&model.PostComment{})
	if deleteCommentResult.Error != nil {
		return result, deleteCommentResult.Error
	}
	result.DeletedComments += deleteCommentResult.RowsAffected

	for _, item := range aggregates {
		if item.PostID == 0 || item.Count <= 0 {
			continue
		}
		if err := tx.Model(&model.CommunityPost{}).
			Where("id = ? AND comments > 0", item.PostID).
			UpdateColumn("comments", gorm.Expr("CASE WHEN comments >= ? THEN comments - ? ELSE 0 END", item.Count, item.Count)).Error; err != nil {
			return result, err
		}
	}

	postIDs := make([]uint, 0)
	if err := tx.Model(&model.CommunityPost{}).Where("user_id = ?", user.ID).Pluck("id", &postIDs).Error; err != nil {
		return result, err
	}
	if len(postIDs) > 0 {
		deletePostCommentResult := tx.Where("post_id IN ?", postIDs).Delete(&model.PostComment{})
		if deletePostCommentResult.Error != nil {
			return result, deletePostCommentResult.Error
		}
		result.DeletedComments += deletePostCommentResult.RowsAffected

		deletePostLikeResult := tx.Where("post_id IN ?", postIDs).Delete(&model.PostLike{})
		if deletePostLikeResult.Error != nil {
			return result, deletePostLikeResult.Error
		}
		result.DeletedLikes += deletePostLikeResult.RowsAffected

		deletePostResult := tx.Where("id IN ?", postIDs).Delete(&model.CommunityPost{})
		if deletePostResult.Error != nil {
			return result, deletePostResult.Error
		}
		result.DeletedPosts += deletePostResult.RowsAffected
	}

	deleteUserLikeResult := tx.Where("user_id = ?", user.ID).Delete(&model.PostLike{})
	if deleteUserLikeResult.Error != nil {
		return result, deleteUserLikeResult.Error
	}
	result.DeletedLikes += deleteUserLikeResult.RowsAffected

	if err := tx.Delete(&user).Error; err != nil {
		return result, err
	}

	return result, nil
}

func parseAuditLogDetail(raw string) map[string]interface{} {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || trimmed == "{}" {
		return map[string]interface{}{}
	}

	payload := map[string]interface{}{}
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return map[string]interface{}{"raw": trimmed}
	}
	return payload
}

func extractUserIDFromEvidence(evidence map[string]interface{}) uint {
	if len(evidence) == 0 {
		return 0
	}
	keys := []string{"user_id", "comment_user_id", "post_user_id"}
	for _, key := range keys {
		id := normalizeToUint(evidence[key])
		if id > 0 {
			return id
		}
	}
	return 0
}

func normalizeToUint(value interface{}) uint {
	switch v := value.(type) {
	case float64:
		if v > 0 {
			return uint(v)
		}
	case float32:
		if v > 0 {
			return uint(v)
		}
	case int:
		if v > 0 {
			return uint(v)
		}
	case int64:
		if v > 0 {
			return uint(v)
		}
	case uint:
		return v
	case uint64:
		return uint(v)
	case string:
		i, err := strconv.ParseUint(strings.TrimSpace(v), 10, 64)
		if err == nil && i > 0 {
			return uint(i)
		}
	}
	return 0
}
