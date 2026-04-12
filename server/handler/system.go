package handler

import (
	"net/http"
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
