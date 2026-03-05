package handler

import (
	"net/http"
	"your-project/utils"

	"github.com/gin-gonic/gin"
)

func OCRStatus(c *gin.Context) {
	exe, tess, langs, hasPPM := utils.OCRStatus()
	c.JSON(http.StatusOK, gin.H{
		"tesseract_path":   exe,
		"tessdata_prefix":  tess,
		"languages":        langs,
		"has_pdftoppm":     hasPPM,
	})
}

