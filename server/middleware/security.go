package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"your-project/internal/service"

	"github.com/gin-gonic/gin"
)

const maxInspectableBodyBytes = 1024 * 1024

var operationMethods = map[string]struct{}{
	http.MethodPost:   {},
	http.MethodPut:    {},
	http.MethodPatch:  {},
	http.MethodDelete: {},
}

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		headers := c.Writer.Header()
		headers.Set("X-Content-Type-Options", "nosniff")
		headers.Set("X-Frame-Options", "DENY")
		headers.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		headers.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		headers.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none'")
		c.Next()
	}
}

func SecurityRequestGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api/v1/") {
			c.Next()
			return
		}

		inspectText := buildInspectionText(c.Request)
		if attackType, found := service.DetectSecurityThreat(inspectText); found {
			service.RecordSecurityAudit(c, c.GetUint("user_id"), "request_guard", "blocked", http.StatusBadRequest, map[string]interface{}{
				"reason": attackType,
			})
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "请求包含潜在攻击特征，已被安全策略拦截"})
			return
		}

		c.Next()
	}
}

func SecurityAccessLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()

		if !strings.HasPrefix(c.Request.URL.Path, "/api/v1/") {
			return
		}
		if strings.HasSuffix(c.Request.URL.Path, "/health") || strings.HasSuffix(c.Request.URL.Path, "/ready") {
			return
		}

		method := strings.ToUpper(c.Request.Method)
		action := "access_log"
		if _, ok := operationMethods[method]; ok {
			action = "operation_log"
		}

		statusCode := c.Writer.Status()
		outcome := "success"
		if statusCode >= http.StatusBadRequest {
			outcome = "failed"
		}

		service.RecordSecurityAudit(c, c.GetUint("user_id"), action, outcome, statusCode, map[string]interface{}{
			"latency_ms": time.Since(startedAt).Milliseconds(),
			"role":       strings.TrimSpace(c.GetString("role")),
			"query":      trimForDetail(c.Request.URL.RawQuery, 256),
		})
	}
}

func buildInspectionText(req *http.Request) string {
	if req == nil {
		return ""
	}

	var builder strings.Builder
	builder.WriteString(strings.TrimSpace(req.URL.Path))
	if rawQuery := strings.TrimSpace(req.URL.RawQuery); rawQuery != "" {
		builder.WriteByte('\n')
		builder.WriteString(rawQuery)
	}

	if !shouldInspectBody(req) {
		return builder.String()
	}

	body := readRequestBody(req)
	if len(body) == 0 {
		return builder.String()
	}

	builder.WriteByte('\n')
	builder.Write(body)
	return builder.String()
}

func shouldInspectBody(req *http.Request) bool {
	if req == nil || req.Body == nil {
		return false
	}

	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if _, ok := operationMethods[method]; !ok {
		return false
	}

	contentType := strings.ToLower(strings.TrimSpace(req.Header.Get("Content-Type")))
	return strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "text/") ||
		strings.Contains(contentType, "application/x-www-form-urlencoded")
}

func readRequestBody(req *http.Request) []byte {
	if req == nil || req.Body == nil {
		return nil
	}

	buffer, err := io.ReadAll(io.LimitReader(req.Body, maxInspectableBodyBytes+1))
	if err != nil {
		req.Body = io.NopCloser(bytes.NewBuffer(nil))
		return nil
	}
	req.Body = io.NopCloser(bytes.NewBuffer(buffer))

	if len(buffer) > maxInspectableBodyBytes {
		return nil
	}
	return buffer
}

func trimForDetail(value string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= max {
		return string(runes)
	}
	return string(runes[:max])
}
