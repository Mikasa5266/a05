package handler

import (
	"net/http"

	"your-project/internal/model"
	"your-project/internal/repository"

	"github.com/gin-gonic/gin"
)

func GetTopAlumni(c *gin.Context) {
	db := repository.GetDB()
	var results []struct {
		Name    string `json:"name"`
		Company string `json:"company"`
		Posts   int64  `json:"posts"`
		Avatar  string `json:"avatar"`
	}

	err := db.Model(&model.CommunityPost{}).
		Select("author as name, company, count(*) as posts, max(avatar) as avatar").
		Where("author != '' AND author IS NOT NULL").
		Group("author, company").
		Order("posts DESC").
		Limit(5).
		Scan(&results).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch top alumni"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"alumni": results})
}

func GetHotCompanies(c *gin.Context) {
	db := repository.GetDB()
	var results []struct {
		Name  string `json:"name"`
		Posts int64  `json:"posts"`
	}

	err := db.Model(&model.CommunityPost{}).
		Select("company as name, count(*) as posts").
		Where("company != '' AND company IS NOT NULL").
		Group("company").
		Order("posts DESC").
		Limit(8).
		Scan(&results).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch hot companies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"companies": results})
}
