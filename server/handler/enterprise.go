package handler

import (
	"net/http"
	"strconv"

	"your-project/model"
	"your-project/repository"

	"github.com/gin-gonic/gin"
)

// ===== Dashboard =====

func GetEnterpriseDashboard(c *gin.Context) {
	db := repository.GetDB()
	var jobCount, talentCount, sessionCount int64
	db.Model(&model.Job{}).Count(&jobCount)
	db.Model(&model.TalentRecord{}).Count(&talentCount)
	db.Model(&model.InterviewSession{}).Count(&sessionCount)

	c.JSON(http.StatusOK, gin.H{
		"total_jobs":     jobCount,
		"total_talent":   talentCount,
		"total_sessions": sessionCount,
		"match_rate":     87,
	})
}

// ===== Talent Pool =====

func GetTalentPool(c *gin.Context) {
	db := repository.GetDB()
	var records []model.TalentRecord
	query := db.Model(&model.TalentRecord{})

	if search := c.Query("search"); search != "" {
		query = query.Where("name LIKE ? OR position LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	query.Order("match_score DESC").Limit(50).Find(&records)
	c.JSON(http.StatusOK, gin.H{"candidates": records})
}

func InviteTalent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	db := repository.GetDB()
	db.Model(&model.TalentRecord{}).Where("id = ?", id).Update("status", "invited")
	c.JSON(http.StatusOK, gin.H{"message": "邀请已发送"})
}

func SaveTalent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	db := repository.GetDB()
	db.Model(&model.TalentRecord{}).Where("id = ?", id).Update("is_saved", true)
	c.JSON(http.StatusOK, gin.H{"message": "已收藏"})
}

// ===== Jobs =====

func GetJobs(c *gin.Context) {
	db := repository.GetDB()
	var jobs []model.Job
	db.Where("status != 'closed'").Order("created_at DESC").Find(&jobs)
	c.JSON(http.StatusOK, gin.H{"jobs": jobs})
}

func CreateJob(c *gin.Context) {
	var job model.Job
	if err := c.ShouldBindJSON(&job); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db := repository.GetDB()
	if err := db.Create(&job).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"job": job})
}

func UpdateJob(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	db := repository.GetDB()
	var job model.Job
	if err := db.First(&job, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "岗位不存在"})
		return
	}
	if err := c.ShouldBindJSON(&job); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.Save(&job)
	c.JSON(http.StatusOK, gin.H{"job": job})
}

func DeleteJob(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	db := repository.GetDB()
	db.Delete(&model.Job{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

func GetAbilityAtlas(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	db := repository.GetDB()
	var job model.Job
	if err := db.First(&job, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "岗位不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"job_title":          job.Title,
		"technical_weight":   job.TechnicalWeight,
		"communicate_weight": job.CommunicateWeight,
		"logic_weight":       job.LogicWeight,
		"match_weight":       job.MatchWeight,
	})
}

// ===== Interview Sessions =====

func GetInterviewSessions(c *gin.Context) {
	db := repository.GetDB()
	var sessions []model.InterviewSession
	status := c.Query("status")
	query := db.Model(&model.InterviewSession{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Order("created_at DESC").Limit(50).Find(&sessions)
	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

func CreateCustomScenario(c *gin.Context) {
	var session model.InterviewSession
	if err := c.ShouldBindJSON(&session); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db := repository.GetDB()
	if err := db.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"session": session})
}

func GetCustomScenarios(c *gin.Context) {
	db := repository.GetDB()
	var sessions []model.InterviewSession
	db.Where("scenario_type IS NOT NULL AND scenario_type != ''").Find(&sessions)
	c.JSON(http.StatusOK, gin.H{"scenarios": sessions})
}

// ===== Analytics =====

func GetRecruitmentAnalytics(c *gin.Context) {
	db := repository.GetDB()
	var totalCandidates, totalHired int64
	db.Model(&model.TalentRecord{}).Count(&totalCandidates)
	db.Model(&model.TalentRecord{}).Where("status = 'hired'").Count(&totalHired)

	c.JSON(http.StatusOK, gin.H{
		"total_candidates":  totalCandidates,
		"total_hired":       totalHired,
		"conversion_rate":   float64(totalHired) / float64(max(totalCandidates, 1)) * 100,
		"avg_time_to_hire":  14,
		"avg_quality_score": 78,
	})
}

func GetRecruitmentFunnel(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"funnel": []gin.H{
			{"stage": "简历筛选", "count": 1200},
			{"stage": "AI初面", "count": 580},
			{"stage": "HR复面", "count": 210},
			{"stage": "终面", "count": 95},
			{"stage": "录用", "count": 42},
		},
	})
}

func GetCandidateQualityDistribution(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"distribution": []gin.H{
			{"range": "90-100", "count": 15},
			{"range": "80-89", "count": 38},
			{"range": "70-79", "count": 52},
			{"range": "60-69", "count": 28},
			{"range": "0-59", "count": 12},
		},
	})
}

// ===== Capability Standards =====

func GetCapabilityStandards(c *gin.Context) {
	db := repository.GetDB()
	var standards []model.CapabilityStandard
	db.Order("created_at DESC").Find(&standards)
	c.JSON(http.StatusOK, gin.H{"standards": standards})
}

func CreateCapabilityStandard(c *gin.Context) {
	var standard model.CapabilityStandard
	if err := c.ShouldBindJSON(&standard); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db := repository.GetDB()
	if err := db.Create(&standard).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"standard": standard})
}

func UpdateCapabilityStandard(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	db := repository.GetDB()
	var standard model.CapabilityStandard
	if err := db.First(&standard, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "标准不存在"})
		return
	}
	if err := c.ShouldBindJSON(&standard); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.Save(&standard)
	c.JSON(http.StatusOK, gin.H{"standard": standard})
}

// ===== Certified & Referrals =====

func GetCertifiedCandidates(c *gin.Context) {
	db := repository.GetDB()
	var records []model.TalentRecord
	db.Where("match_score >= 80").Order("match_score DESC").Limit(20).Find(&records)
	c.JSON(http.StatusOK, gin.H{"certified": records})
}

func GetReferralChannels(c *gin.Context) {
	db := repository.GetDB()
	var referrals []model.Referral
	db.Order("created_at DESC").Find(&referrals)
	c.JSON(http.StatusOK, gin.H{"referrals": referrals})
}

func CreateReferral(c *gin.Context) {
	var referral model.Referral
	if err := c.ShouldBindJSON(&referral); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db := repository.GetDB()
	if err := db.Create(&referral).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"referral": referral})
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
