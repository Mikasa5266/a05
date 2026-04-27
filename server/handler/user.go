package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"your-project/internal/service"
	"your-project/middleware"

	"github.com/gin-gonic/gin"
)

func Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		Role     string `json:"role"`
		RealName string `json:"real_name" binding:"required"`
		Phone    string `json:"phone" binding:"required"`
		IDCardNo string `json:"id_card_no" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		service.RecordSecurityAudit(c, 0, "user_register", "failed", http.StatusBadRequest, map[string]interface{}{
			"reason": "invalid_payload",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := req.Role
	if role == "" {
		role = "student"
	}
	if role != "student" {
		service.RecordSecurityAudit(c, 0, "user_register", "blocked", http.StatusBadRequest, map[string]interface{}{
			"reason": "invalid_role",
			"role":   role,
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "企业/高校账号请使用入驻申请接口"})
		return
	}

	if err := service.ValidateSafeTextField("username", req.Username); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := service.ValidateSafeTextField("email", req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := service.ValidateSafeTextField("real_name", req.RealName); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Username = service.SanitizeUserInput(req.Username)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.RealName = service.SanitizeUserInput(req.RealName)

	user, err := service.CreateStudentUser(req.Username, req.Email, req.Password, req.RealName, req.Phone, req.IDCardNo)
	if err != nil {
		statusCode := http.StatusInternalServerError
		reason := "create_user_failed"
		errMessage := strings.ToLower(strings.TrimSpace(err.Error()))
		if strings.Contains(errMessage, "invalid") || strings.Contains(errMessage, "requires") || strings.Contains(errMessage, "exists") {
			statusCode = http.StatusBadRequest
			reason = "register_validation_failed"
		}

		service.RecordSecurityAudit(c, 0, "user_register", "failed", statusCode, map[string]interface{}{
			"reason": reason,
		})
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	service.RecordSecurityAudit(c, user.ID, "user_register", "success", http.StatusCreated, map[string]interface{}{
		"role": user.Role,
	})

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user":    user,
	})
}

func ApplyEnterprise(c *gin.Context) {
	var req struct {
		Username      string `json:"username" binding:"required"`
		Email         string `json:"email" binding:"required,email"`
		Password      string `json:"password" binding:"required,min=6"`
		CompanyName   string `json:"company_name" binding:"required"`
		ContactName   string `json:"contact_name"`
		ContactPhone  string `json:"contact_phone"`
		BusinessScope string `json:"business_scope"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		service.RecordSecurityAudit(c, 0, "user_login", "failed", http.StatusBadRequest, map[string]interface{}{
			"reason": "invalid_payload",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, enterprise, err := service.ApplyEnterprise(
		req.Username,
		req.Email,
		req.Password,
		req.CompanyName,
		req.ContactName,
		req.ContactPhone,
		req.BusinessScope,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "企业入驻申请已提交，等待审核",
		"user":        user,
		"application": enterprise,
	})
}

func ApplyUniversity(c *gin.Context) {
	var req struct {
		Username       string `json:"username" binding:"required"`
		Email          string `json:"email" binding:"required,email"`
		Password       string `json:"password" binding:"required,min=6"`
		UniversityName string `json:"university_name" binding:"required"`
		ContactName    string `json:"contact_name"`
		ContactPhone   string `json:"contact_phone"`
		Department     string `json:"department"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, university, err := service.ApplyUniversity(
		req.Username,
		req.Email,
		req.Password,
		req.UniversityName,
		req.ContactName,
		req.ContactPhone,
		req.Department,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "高校入驻申请已提交，等待审核",
		"user":        user,
		"application": university,
	})
}

func AuditApplication(c *gin.Context) {
	adminID := c.GetUint("user_id")
	role := c.Param("role")

	var req struct {
		ApplicationID uint   `json:"application_id" binding:"required"`
		Status        string `json:"status" binding:"required"`
		Remark        string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if role != "enterprise" && role != "university" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role must be enterprise or university"})
		return
	}

	if err := service.AuditApplication(role, req.ApplicationID, req.Status, req.Remark, adminID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "审核处理成功"})
}

func Login(c *gin.Context) {
	var req struct {
		Account  string `json:"account"`
		Email    string `json:"email"`
		Password string `json:"password" binding:"required"`
		Role     string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account := strings.TrimSpace(req.Account)
	if account == "" {
		account = strings.TrimSpace(req.Email)
	}
	if account == "" {
		service.RecordSecurityAudit(c, 0, "user_login", "failed", http.StatusBadRequest, map[string]interface{}{
			"reason": "missing_account",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "account is required"})
		return
	}

	user, err := service.AuthenticateUser(account, req.Password)
	if err != nil {
		service.RecordSecurityAudit(c, 0, "user_login", "failed", http.StatusUnauthorized, map[string]interface{}{
			"reason":  "invalid_credentials",
			"account": account,
		})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Validate role if provided
	if req.Role != "" && user.Role != req.Role {
		service.RecordSecurityAudit(c, user.ID, "user_login", "blocked", http.StatusForbidden, map[string]interface{}{
			"reason":       "role_mismatch",
			"request_role": req.Role,
			"account_role": user.Role,
		})
		c.JSON(http.StatusForbidden, gin.H{"error": "该账号不属于此端，请切换到正确的登录入口"})
		return
	}

	auditStatus, _, err := service.GetAuditStatusForUser(user)
	if err != nil {
		service.RecordSecurityAudit(c, user.ID, "user_login", "failed", http.StatusUnauthorized, map[string]interface{}{
			"reason": "audit_status_error",
		})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "账号状态异常，请联系管理员"})
		return
	}

	if user.Role == "enterprise" || user.Role == "university" {
		if auditStatus != "approved" {
			message := "您的账号资质审核未通过"
			if auditStatus == "pending" {
				if user.Role == "enterprise" {
					message = "您的企业资质正在审核中"
				} else {
					message = "您的高校资质正在审核中"
				}
			} else if auditStatus == "rejected" {
				if user.Role == "enterprise" {
					message = "您的企业资质审核未通过，请联系管理员"
				} else {
					message = "您的高校资质审核未通过，请联系管理员"
				}
			}
			service.RecordSecurityAudit(c, user.ID, "user_login", "blocked", http.StatusForbidden, map[string]interface{}{
				"reason":       "audit_not_approved",
				"audit_status": auditStatus,
			})
			c.JSON(http.StatusForbidden, gin.H{"error": message})
			return
		}
	}

	token, err := middleware.GenerateToken(user.ID, user.Role, user.UUID)
	if err != nil {
		service.RecordSecurityAudit(c, user.ID, "user_login", "failed", http.StatusInternalServerError, map[string]interface{}{
			"reason": "token_generate_failed",
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	service.RecordSecurityAudit(c, user.ID, "user_login", "success", http.StatusOK, map[string]interface{}{
		"role": user.Role,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user":    user,
	})
}

func GetUserProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	user, err := service.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func UpdateUserProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := service.UpdateUserProfile(userID, req.Username, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user":    user,
	})
}

func UpdateAvatar(c *gin.Context) {
	userID := c.GetUint("user_id")
	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	uploadDir := "./uploads/avatars"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, 0755)
	}

	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d_%d%s", userID, time.Now().UnixNano(), ext)
	dst := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	avatarURL := "/uploads/avatars/" + filename
	user, err := service.UpdateUserAvatar(userID, avatarURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Avatar updated successfully",
		"user":    user,
	})
}

func UpdatePassword(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.UpdateUserPassword(userID, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}
