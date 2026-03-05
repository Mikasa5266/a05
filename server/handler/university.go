package handler

import (
	"net/http"
	"strconv"

	"your-project/model"
	"your-project/repository"

	"github.com/gin-gonic/gin"
)

// ===== Dashboard =====

func GetUniversityDashboard(c *gin.Context) {
	db := repository.GetDB()
	var studentCount, courseCount int64
	var highRiskCount int64
	db.Model(&model.StudentRecord{}).Count(&studentCount)
	db.Model(&model.Course{}).Count(&courseCount)
	db.Model(&model.StudentRecord{}).Where("risk_level = 'high'").Count(&highRiskCount)

	c.JSON(http.StatusOK, gin.H{
		"total_students":  studentCount,
		"total_courses":   courseCount,
		"high_risk_count": highRiskCount,
		"employment_rate": 72,
		"interview_rate":  85,
	})
}

// ===== Student Tracking =====

func GetStudentTracking(c *gin.Context) {
	db := repository.GetDB()
	var students []model.StudentRecord
	query := db.Model(&model.StudentRecord{})

	if search := c.Query("search"); search != "" {
		query = query.Where("name LIKE ? OR student_no LIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if major := c.Query("major"); major != "" {
		query = query.Where("major = ?", major)
	}
	if risk := c.Query("risk_level"); risk != "" {
		query = query.Where("risk_level = ?", risk)
	}

	query.Order("risk_level DESC, average_score ASC").Limit(100).Find(&students)
	c.JSON(http.StatusOK, gin.H{"students": students})
}

func GetStudentDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	db := repository.GetDB()
	var student model.StudentRecord
	if err := db.First(&student, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "学生不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"student": student})
}

func UpdateStudentRisk(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		RiskLevel string `json:"risk_level"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db := repository.GetDB()
	db.Model(&model.StudentRecord{}).Where("id = ?", id).Update("risk_level", req.RiskLevel)
	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// ===== Risk Groups & Support =====

func GetRiskGroups(c *gin.Context) {
	db := repository.GetDB()
	var highCount, mediumCount, lowCount int64
	db.Model(&model.StudentRecord{}).Where("risk_level = 'high'").Count(&highCount)
	db.Model(&model.StudentRecord{}).Where("risk_level = 'medium'").Count(&mediumCount)
	db.Model(&model.StudentRecord{}).Where("risk_level = 'low'").Count(&lowCount)

	c.JSON(http.StatusOK, gin.H{
		"groups": []gin.H{
			{"level": "high", "label": "高风险", "count": highCount},
			{"level": "medium", "label": "中风险", "count": mediumCount},
			{"level": "low", "label": "低风险", "count": lowCount},
		},
	})
}

func AssignMentor(c *gin.Context) {
	var req struct {
		StudentID uint `json:"student_id"`
		MentorID  uint `json:"mentor_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db := repository.GetDB()
	db.Model(&model.StudentRecord{}).Where("id = ?", req.StudentID).Update("mentor_id", req.MentorID)
	c.JSON(http.StatusOK, gin.H{"message": "导师分配成功"})
}

func BatchSupport(c *gin.Context) {
	var req struct {
		StudentIDs []uint `json:"student_ids"`
		Action     string `json:"action"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "批量帮扶操作已执行", "affected": len(req.StudentIDs)})
}

func RecommendCourse(c *gin.Context) {
	var req struct {
		StudentID uint `json:"student_id"`
		CourseID  uint `json:"course_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "课程已推荐"})
}

// ===== Courses =====

func GetCourses(c *gin.Context) {
	db := repository.GetDB()
	var courses []model.Course
	db.Order("created_at DESC").Find(&courses)
	c.JSON(http.StatusOK, gin.H{"courses": courses})
}

func CreateCourse(c *gin.Context) {
	var course model.Course
	if err := c.ShouldBindJSON(&course); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db := repository.GetDB()
	if err := db.Create(&course).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"course": course})
}

func GetResources(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"resources": []gin.H{
			{"type": "video", "count": 128, "label": "视频课程"},
			{"type": "document", "count": 256, "label": "文档资料"},
			{"type": "exercise", "count": 512, "label": "模拟练习"},
		},
	})
}

// ===== Employment Stats =====

func GetEmploymentStats(c *gin.Context) {
	db := repository.GetDB()
	var totalStudents, employedCount int64
	db.Model(&model.StudentRecord{}).Count(&totalStudents)
	db.Model(&model.StudentRecord{}).Where("employment_status = 'employed'").Count(&employedCount)

	rate := float64(0)
	if totalStudents > 0 {
		rate = float64(employedCount) / float64(totalStudents) * 100
	}
	c.JSON(http.StatusOK, gin.H{
		"total_students":  totalStudents,
		"employed":        employedCount,
		"employment_rate": rate,
	})
}

func GetMajorEmployment(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"majors": []gin.H{
			{"name": "计算机科学", "rate": 94, "count": 280},
			{"name": "软件工程", "rate": 92, "count": 240},
			{"name": "人工智能", "rate": 90, "count": 120},
			{"name": "数据科学", "rate": 88, "count": 95},
			{"name": "信息安全", "rate": 85, "count": 78},
		},
	})
}

func GetSalaryDistribution(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"distribution": []gin.H{
			{"range": "5-8k", "percent": 15},
			{"range": "8-12k", "percent": 35},
			{"range": "12-18k", "percent": 30},
			{"range": "18-25k", "percent": 15},
			{"range": "25k+", "percent": 5},
		},
	})
}

func GetCityDistribution(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"cities": []gin.H{
			{"name": "北京", "percent": 25},
			{"name": "上海", "percent": 22},
			{"name": "深圳", "percent": 18},
			{"name": "杭州", "percent": 12},
			{"name": "广州", "percent": 10},
			{"name": "其他", "percent": 13},
		},
	})
}

func GetIndustryDistribution(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"industries": []gin.H{
			{"name": "互联网", "percent": 42},
			{"name": "金融科技", "percent": 18},
			{"name": "人工智能", "percent": 15},
			{"name": "通信电子", "percent": 12},
			{"name": "其他", "percent": 13},
		},
	})
}

// ===== Talent Push =====

func GetRecommendedStudents(c *gin.Context) {
	db := repository.GetDB()
	var students []model.StudentRecord
	db.Where("average_score >= 70 AND employment_status = 'seeking'").
		Order("average_score DESC").Limit(20).Find(&students)
	c.JSON(http.StatusOK, gin.H{"students": students})
}

func PushStudentsToEnterprise(c *gin.Context) {
	var req model.TalentPush
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db := repository.GetDB()
	if err := db.Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"push": req})
}

func GetPushHistory(c *gin.Context) {
	db := repository.GetDB()
	var pushes []model.TalentPush
	db.Order("created_at DESC").Limit(50).Find(&pushes)
	c.JSON(http.StatusOK, gin.H{"history": pushes})
}
