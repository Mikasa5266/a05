package middleware

import (
	"net/http"
	"strings"

	"your-project/internal/service"

	"github.com/gin-gonic/gin"
)

const adminPlatformTokenHeader = "X-Admin-Token"

func AdminPlatformAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(c.GetHeader(adminPlatformTokenHeader))
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "admin token required"})
			c.Abort()
			return
		}

		username, ok := service.ValidateAdminPlatformSession(token)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "admin token invalid or expired"})
			c.Abort()
			return
		}

		c.Set("admin_platform_user", username)
		c.Set("admin_platform_token", token)
		c.Next()
	}
}
