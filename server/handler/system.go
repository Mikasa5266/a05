package handler

import (
	"net/http"
	"your-project/config"
	"your-project/internal/repository"
	"your-project/utils"

	"github.com/gin-gonic/gin"
)

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func Ready(c *gin.Context) {
	db := repository.GetDB()
	sqlDB, err := db.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "degraded",
			"error":  "database handle unavailable",
		})
		return
	}

	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "degraded",
			"error":  "database unreachable",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

func OCRStatus(c *gin.Context) {
	exe, tess, langs, hasPPM := utils.OCRStatus()
	c.JSON(http.StatusOK, gin.H{
		"tesseract_path":  exe,
		"tessdata_prefix": tess,
		"languages":       langs,
		"has_pdftoppm":    hasPPM,
	})
}

func GetSecurityContacts(c *gin.Context) {
	cfg := config.GetConfig()
	if cfg == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "security config unavailable"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"responsible_person": gin.H{
			"name":  cfg.Security.ResponsiblePerson.Name,
			"phone": cfg.Security.ResponsiblePerson.Phone,
			"email": cfg.Security.ResponsiblePerson.Email,
		},
		"emergency_contact": gin.H{
			"name":  cfg.Security.EmergencyContact.Name,
			"phone": cfg.Security.EmergencyContact.Phone,
			"email": cfg.Security.EmergencyContact.Email,
		},
		"log_retention_days":      cfg.Security.LogRetentionDays,
		"patrol_interval_minutes": cfg.Security.PatrolIntervalMinutes,
	})
}
