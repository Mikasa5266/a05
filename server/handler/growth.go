package handler

import (
	"net/http"
	"your-project/service"

	"github.com/gin-gonic/gin"
)

func GetGrowthStats(c *gin.Context) {
	userID := c.GetUint("user_id")

	svc := service.NewGrowthService()
	stats, err := svc.GetGrowthStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}
