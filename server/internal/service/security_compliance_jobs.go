package service

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"your-project/config"
	"your-project/internal/model"
	"your-project/internal/repository"

	"gorm.io/gorm"
)

var complianceWorkersOnce sync.Once

const systemPatrolRole = "system_patrol"

var unresolvedSecurityReportStatuses = []string{"pending", "processing"}

type IllegalContentPatrolResult struct {
	ScannedPosts    int `json:"scanned_posts"`
	ScannedComments int `json:"scanned_comments"`
	HitPosts        int `json:"hit_posts"`
	HitComments     int `json:"hit_comments"`
	CreatedReports  int `json:"created_reports"`
}

func StartSecurityComplianceWorkers() {
	complianceWorkersOnce.Do(func() {
		startSecurityLogRetentionWorker()
		startIllegalContentPatrolWorker()
	})
}

func CleanupExpiredSecurityAuditLogs(now time.Time) (int64, error) {
	if now.IsZero() {
		now = time.Now()
	}

	cutoff := now.AddDate(0, 0, -effectiveSecurityLogRetentionDays())
	result := repository.GetDB().Where("created_at < ?", cutoff).Delete(&model.SecurityAuditLog{})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func RunIllegalContentPatrol(limit int) (IllegalContentPatrolResult, error) {
	if limit <= 0 {
		limit = resolvePatrolScanLimit()
	}

	result := IllegalContentPatrolResult{}
	db := repository.GetDB()

	var posts []model.CommunityPost
	if err := db.Order("updated_at DESC").Limit(limit).Find(&posts).Error; err != nil {
		return result, fmt.Errorf("query posts failed: %w", err)
	}
	result.ScannedPosts = len(posts)

	for _, post := range posts {
		reason := detectIllegalContentReason(strings.Join([]string{post.Title, post.Content, post.Process, post.Questions, post.Review}, "\n"))
		if reason == "" {
			continue
		}
		result.HitPosts++

		created, err := createSystemPatrolReport("post", post.ID, reason, patrolExcerpt(post.Content, 180))
		if err != nil {
			return result, err
		}
		if created {
			result.CreatedReports++
		}
	}

	var comments []model.PostComment
	if err := db.Order("created_at DESC").Limit(limit).Find(&comments).Error; err != nil {
		return result, fmt.Errorf("query comments failed: %w", err)
	}
	result.ScannedComments = len(comments)

	for _, comment := range comments {
		reason := detectIllegalContentReason(comment.Content)
		if reason == "" {
			continue
		}
		result.HitComments++

		created, err := createSystemPatrolReport("comment", comment.ID, reason, patrolExcerpt(comment.Content, 180))
		if err != nil {
			return result, err
		}
		if created {
			result.CreatedReports++
		}
	}

	return result, nil
}

func startSecurityLogRetentionWorker() {
	go func() {
		runSecurityLogRetentionJob()

		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			runSecurityLogRetentionJob()
		}
	}()
}

func startIllegalContentPatrolWorker() {
	go func() {
		runIllegalContentPatrolJob()

		ticker := time.NewTicker(resolvePatrolInterval())
		defer ticker.Stop()
		for range ticker.C {
			runIllegalContentPatrolJob()
		}
	}()
}

func runSecurityLogRetentionJob() {
	rows, err := CleanupExpiredSecurityAuditLogs(time.Now())
	if err != nil {
		log.Printf("[security] cleanup expired security logs failed: %v", err)
		return
	}
	if rows > 0 {
		log.Printf("[security] cleanup expired security logs done, removed=%d", rows)
	}
}

func runIllegalContentPatrolJob() {
	result, err := RunIllegalContentPatrol(resolvePatrolScanLimit())
	if err != nil {
		log.Printf("[security] illegal content patrol failed: %v", err)
		return
	}
	if result.CreatedReports > 0 {
		log.Printf("[security] illegal content patrol created reports=%d (posts=%d, comments=%d)", result.CreatedReports, result.HitPosts, result.HitComments)
	}
}

func createSystemPatrolReport(targetType string, targetID uint, reason, description string) (bool, error) {
	db := repository.GetDB()

	var existing model.SecurityReport
	err := db.Where("reporter_role = ? AND target_type = ? AND target_id = ? AND status IN ? AND reason = ?",
		systemPatrolRole,
		targetType,
		targetID,
		unresolvedSecurityReportStatuses,
		reason,
	).First(&existing).Error
	if err == nil {
		return false, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, fmt.Errorf("query existing patrol report failed: %w", err)
	}

	report := model.SecurityReport{
		ReporterUserID: 0,
		ReporterRole:   systemPatrolRole,
		TargetType:     targetType,
		TargetID:       targetID,
		Reason:         trimReason(reason),
		Description:    patrolExcerpt(description, 1000),
		Status:         "pending",
	}
	if err := db.Create(&report).Error; err != nil {
		return false, fmt.Errorf("create patrol report failed: %w", err)
	}

	return true, nil
}

func detectIllegalContentReason(content string) string {
	if attackType, found := DetectSecurityThreat(content); found {
		return "系统巡查命中" + attackType + "风险特征"
	}

	hits := MatchBlockedWords(content)
	if len(hits) == 0 {
		return ""
	}
	if len(hits) > 3 {
		hits = hits[:3]
	}
	return "系统巡查命中违规关键词：" + strings.Join(hits, "、")
}

func trimReason(reason string) string {
	runes := []rune(strings.TrimSpace(reason))
	if len(runes) <= 120 {
		return string(runes)
	}
	return string(runes[:120])
}

func patrolExcerpt(text string, max int) string {
	trimmed := strings.TrimSpace(text)
	if max <= 0 {
		return ""
	}
	runes := []rune(trimmed)
	if len(runes) <= max {
		return trimmed
	}
	return string(runes[:max])
}

func effectiveSecurityLogRetentionDays() int {
	cfg := config.GetConfig()
	if cfg == nil {
		return config.DefaultSecurityLogRetentionDays
	}
	if cfg.Security.LogRetentionDays < config.DefaultSecurityLogRetentionDays {
		return config.DefaultSecurityLogRetentionDays
	}
	return cfg.Security.LogRetentionDays
}

func resolvePatrolInterval() time.Duration {
	cfg := config.GetConfig()
	if cfg == nil {
		return time.Duration(config.DefaultSecurityPatrolIntervalMinute) * time.Minute
	}
	minutes := cfg.Security.PatrolIntervalMinutes
	if minutes <= 0 {
		minutes = config.DefaultSecurityPatrolIntervalMinute
	}
	if minutes < 5 {
		minutes = 5
	}
	return time.Duration(minutes) * time.Minute
}

func resolvePatrolScanLimit() int {
	cfg := config.GetConfig()
	if cfg == nil {
		return config.DefaultSecurityPatrolScanLimit
	}
	limit := cfg.Security.PatrolScanLimit
	if limit <= 0 {
		limit = config.DefaultSecurityPatrolScanLimit
	}
	if limit > 1000 {
		limit = 1000
	}
	return limit
}
