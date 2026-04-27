package handler

import (
	"net/http"
	"strings"

	"your-project/internal/service"

	"github.com/gin-gonic/gin"
)

func AdminPlatformLogin(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid login payload"})
		return
	}

	username := strings.TrimSpace(req.Username)
	if !service.VerifyAdminPlatformCredentials(username, req.Password) {
		service.RecordSecurityAudit(c, 0, "admin_platform_login", "failed", http.StatusUnauthorized, map[string]interface{}{
			"username": username,
			"reason":   "invalid_admin_credentials",
		})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid admin credentials"})
		return
	}

	token, expiresAt, err := service.CreateAdminPlatformSession(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create admin session"})
		return
	}

	service.RecordSecurityAudit(c, 0, "admin_platform_login", "success", http.StatusOK, map[string]interface{}{
		"username": username,
	})

	c.JSON(http.StatusOK, gin.H{
		"token":      token,
		"expires_at": expiresAt,
		"username":   username,
	})
}

func AdminPlatformLogout(c *gin.Context) {
	token := strings.TrimSpace(c.GetString("admin_platform_token"))
	username := strings.TrimSpace(c.GetString("admin_platform_user"))
	if token != "" {
		service.RevokeAdminPlatformSession(token)
	}
	service.RecordSecurityAudit(c, 0, "admin_platform_logout", "success", http.StatusOK, map[string]interface{}{
		"username": username,
	})
	c.JSON(http.StatusOK, gin.H{"message": "logout success"})
}

func AdminPlatformMe(c *gin.Context) {
	username := strings.TrimSpace(c.GetString("admin_platform_user"))
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "admin session invalid"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"username": username,
	})
}
