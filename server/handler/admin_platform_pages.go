package handler

import (
	"embed"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed admin_platform_ui/*
var adminPlatformUIFS embed.FS

func AdminPlatformRoot(c *gin.Context) {
	setAdminPlatformHeaders(c)
	c.Redirect(http.StatusTemporaryRedirect, "/admin-platform/login")
}

func AdminPlatformLoginPage(c *gin.Context) {
	serveAdminPlatformFile(c, "login.html")
}

func AdminPlatformDashboardPage(c *gin.Context) {
	serveAdminPlatformFile(c, "dashboard.html")
}

func AdminPlatformStaticAsset(c *gin.Context) {
	relative := strings.TrimPrefix(c.Param("filepath"), "/")
	if relative == "" {
		c.Status(http.StatusNotFound)
		return
	}

	cleaned := path.Clean("/" + relative)
	if strings.Contains(cleaned, "..") {
		c.Status(http.StatusBadRequest)
		return
	}
	cleaned = strings.TrimPrefix(cleaned, "/")

	content, err := adminPlatformUIFS.ReadFile(path.Join("admin_platform_ui", cleaned))
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	setAdminPlatformHeaders(c)
	if contentType := mime.TypeByExtension(path.Ext(cleaned)); contentType != "" {
		c.Header("Content-Type", contentType)
	}
	c.Data(http.StatusOK, c.Writer.Header().Get("Content-Type"), content)
}

func serveAdminPlatformFile(c *gin.Context, fileName string) {
	content, err := adminPlatformUIFS.ReadFile(path.Join("admin_platform_ui", fileName))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin platform page not found"})
		return
	}
	setAdminPlatformHeaders(c)
	c.Data(http.StatusOK, "text/html; charset=utf-8", content)
}

func setAdminPlatformHeaders(c *gin.Context) {
	headers := c.Writer.Header()
	headers.Set("X-Content-Type-Options", "nosniff")
	headers.Set("X-Frame-Options", "DENY")
	headers.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	headers.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
	headers.Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
}
