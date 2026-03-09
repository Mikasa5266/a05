package handler

import (
	"encoding/base64"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"your-project/config"
	ws "your-project/pkg/websocket"
	"your-project/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func parseUserIDFromToken(tokenString string) (uint, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(config.GetConfig().JWT.Secret), nil
	})
	if err != nil {
		return 0, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, jwt.ErrTokenInvalidClaims
	}
	uid, ok := claims["user_id"].(float64)
	if !ok || uid <= 0 {
		return 0, jwt.ErrTokenInvalidClaims
	}
	return uint(uid), nil
}

// InterviewSignalWS provides a websocket signaling channel for live human interview rooms.
func InterviewSignalWS(c *gin.Context) {
	tokenString := strings.TrimSpace(c.Query("token"))
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token is required"})
		return
	}

	roomID := strings.TrimSpace(c.Query("room_id"))
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room_id is required"})
		return
	}

	userID, err := parseUserIDFromToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// 1. 基础逻辑检查：如果面试已经结束，禁止进入
	if strings.HasPrefix(roomID, "invitation-") {
		invIDStr := strings.TrimPrefix(roomID, "invitation-")
		invID, _ := strconv.ParseUint(invIDStr, 10, 32)
		invitation, err := service.GetInvitationByID(uint(invID))
		if err == nil && (invitation.Status == "completed" || invitation.Status == "rejected" || invitation.Status == "cancelled") {
			c.JSON(http.StatusForbidden, gin.H{"error": "面试已结束或已失效，无法进入"})
			return
		}
	}

	// 2. 并发检查：检查该用户是否已经在其他房间（防止分身）
	// 注意：这里允许重新进入同一个房间（断线重连），但禁止跨房间
	activeClients := ws.GetHub().GetClientsByUserID(strconv.FormatUint(uint64(userID), 10))
	for _, client := range activeClients {
		if client.GetInterviewID() != roomID {
			c.JSON(http.StatusConflict, gin.H{"error": "您已在其他面试间中，请先退出"})
			return
		}
	}

	query := c.Request.URL.Query()
	query.Set("user_id", strconv.FormatUint(uint64(userID), 10))
	query.Set("interview_id", roomID)
	c.Request.URL.RawQuery = query.Encode()

	ws.GetHub().HandleWebSocket(c.Writer, c.Request)
}

func StartInterview(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		Position      string `json:"position" binding:"required"`
		Difficulty    string `json:"difficulty" binding:"required"`
		Mode          string `json:"mode"`           // technical, hr, comprehensive
		Style         string `json:"style"`          // gentle, stress, deep, practical, algorithm
		Company       string `json:"company"`        // ali, bytedance, tencent, meituan, baidu
		InterviewMode string `json:"interview_mode"` // ai, human, random
		InvitationID  *uint  `json:"invitation_id,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults if empty
	if req.Mode == "" {
		req.Mode = "technical"
	}
	if req.Style == "" {
		req.Style = "gentle"
	}
	if req.InterviewMode == "" {
		req.InterviewMode = "ai"
	}

	// Validate mode
	validModes := map[string]bool{"technical": true, "hr": true, "comprehensive": true, "blindbox": true}
	if !validModes[req.Mode] {
		req.Mode = "technical"
	}

	// Validate style
	validStyles := map[string]bool{"gentle": true, "stress": true, "deep": true, "practical": true, "algorithm": true}
	if !validStyles[req.Style] {
		req.Style = "gentle"
	}

	// Validate difficulty
	validDifficulties := map[string]bool{"campus_intern": true, "campus_graduate": true, "social_junior": true}
	if !validDifficulties[req.Difficulty] {
		req.Difficulty = "campus_intern"
	}

	// Validate interview mode
	validInterviewModes := map[string]bool{"ai": true, "human": true, "random": true}
	if !validInterviewModes[req.InterviewMode] {
		req.InterviewMode = "ai"
	}

	interview, err := service.StartInterview(userID, req.Position, req.Difficulty, req.Mode, req.Style, req.Company, req.InterviewMode, req.InvitationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Interview started successfully",
		"interview": interview,
	})
}

// ListInviteCandidates returns university/enterprise users that can be invited for a mock interview.
func ListInviteCandidates(c *gin.Context) {
	role := c.DefaultQuery("role", "")
	keyword := c.DefaultQuery("keyword", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	users, total, err := service.ListInviteCandidates(role, keyword, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users":     users,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func CreateHumanInvitation(c *gin.Context) {
	studentID := c.GetUint("user_id")

	var req struct {
		InviteeUserID uint   `json:"invitee_user_id" binding:"required"`
		ScheduledAt   string `json:"scheduled_at,omitempty"`
		Position      string `json:"position" binding:"required"`
		Difficulty    string `json:"difficulty" binding:"required"`
		Mode          string `json:"mode" binding:"required"`
		Style         string `json:"style" binding:"required"`
		Company       string `json:"company"`
		Notes         string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var scheduledAt *time.Time
	if strings.TrimSpace(req.ScheduledAt) != "" {
		parsed, err := time.Parse(time.RFC3339, req.ScheduledAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "时间格式无效，请使用 ISO 8601 格式"})
			return
		}
		scheduledAt = &parsed
	}

	invitation, err := service.CreateHumanInvitation(
		studentID,
		req.InviteeUserID,
		scheduledAt,
		req.Position,
		req.Difficulty,
		req.Mode,
		req.Style,
		req.Company,
		req.Notes,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "邀请已发送", "invitation": invitation})
}

func GetHumanInvitations(c *gin.Context) {
	studentID := c.GetUint("user_id")
	invitations, err := service.ListHumanInvitations(studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"invitations": invitations})
}

func GetReceivedHumanInvitations(c *gin.Context) {
	inviteeUserID := c.GetUint("user_id")
	invitations, err := service.ListReceivedHumanInvitations(inviteeUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"invitations": invitations})
}

func RespondHumanInvitation(c *gin.Context) {
	inviteeUserID := c.GetUint("user_id")
	invitationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invitation ID"})
		return
	}

	var req struct {
		Action string `json:"action" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invitation, err := service.RespondHumanInvitation(inviteeUserID, uint(invitationID), req.Action)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "邀请状态已更新", "invitation": invitation})
}

func GetInterview(c *gin.Context) {
	userID := c.GetUint("user_id")
	interviewID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interview ID"})
		return
	}

	interview, err := service.GetInterviewByID(userID, uint(interviewID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Interview not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"interview": interview})
}

func GetInterviews(c *gin.Context) {
	userID := c.GetUint("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	interviews, total, err := service.GetUserInterviews(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"interviews": interviews,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
	})
}

func SubmitAnswer(c *gin.Context) {
	userID := c.GetUint("user_id")
	interviewID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interview ID"})
		return
	}

	var req struct {
		QuestionID      uint   `json:"question_id" binding:"required"`
		QuestionTitle   string `json:"question_title,omitempty"`
		QuestionContent string `json:"question_content,omitempty"`
		Answer          string `json:"answer"`
		AudioData       string `json:"audio_data,omitempty"`
		AudioMime       string `json:"audio_mime,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if strings.TrimSpace(req.Answer) == "" && strings.TrimSpace(req.AudioData) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "您似乎没有做出任何回答"})
		return
	}

	result, err := service.SubmitAnswer(userID, uint(interviewID), req.QuestionID, req.Answer, req.AudioData, req.AudioMime, req.QuestionTitle, req.QuestionContent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Answer submitted successfully",
		"result":  result,
	})
}

func SynthesizeInterviewSpeech(c *gin.Context) {
	userID := c.GetUint("user_id")
	interviewID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interview ID"})
		return
	}

	var req struct {
		Text string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	interview, err := service.GetInterviewByID(userID, uint(interviewID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Interview not found"})
		return
	}

	text := strings.TrimSpace(req.Text)
	if text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text is required"})
		return
	}

	ttsCfg := service.GetTTSConfig()
	if ttsCfg.MaxCharsPerInterview > 0 && interview.TTSCharCount+len([]rune(text)) > ttsCfg.MaxCharsPerInterview {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "语音播报预算已达上限，请切换文字模式"})
		return
	}

	aiService := service.NewAIService()
	audioBytes, err := aiService.SynthesizeSpeech(text)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	interview.TTSCharCount += len([]rune(text))
	if _, updateErr := service.SaveInterviewBudgetUsage(interview); updateErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": updateErr.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"audio_base64": base64.StdEncoding.EncodeToString(audioBytes),
	})
}

func EndInterview(c *gin.Context) {
	userID := c.GetUint("user_id")
	interviewID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interview ID"})
		return
	}

	interview, err := service.EndInterview(userID, uint(interviewID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Interview ended successfully",
		"interview": interview,
	})
}

// AnalyzeSpeechChunk receives a short audio chunk and returns real-time speech metrics
func AnalyzeSpeechChunk(c *gin.Context) {
	var req struct {
		AudioData string  `json:"audio_data" binding:"required"`
		AudioMime string  `json:"audio_mime,omitempty"`
		Duration  float64 `json:"duration" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	audioPayload := strings.TrimSpace(req.AudioData)
	if !strings.HasPrefix(audioPayload, "data:") && strings.TrimSpace(req.AudioMime) != "" {
		audioPayload = "data:" + strings.TrimSpace(req.AudioMime) + ";base64," + audioPayload
	}

	svc := service.NewSpeechAnalysisService()
	metrics, err := svc.AnalyzeAudioChunk(audioPayload, req.Duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
	})
}

func GetShadowCoachHint(c *gin.Context) {
	userID := c.GetUint("user_id")
	interviewID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interview ID"})
		return
	}

	var req struct {
		Question       string `json:"question" binding:"required"`
		Transcript     string `json:"transcript"`
		ExpectedAnswer string `json:"expected_answer"`
		SilenceSeconds int    `json:"silence_seconds"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.SilenceSeconds < 0 {
		req.SilenceSeconds = 0
	}

	hints, err := service.GenerateShadowHintPack(userID, uint(interviewID), req.Question, req.Transcript, req.ExpectedAnswer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	hint := ""
	if len(hints) > 0 {
		if req.SilenceSeconds >= 60 && len(hints) >= 3 {
			hint = hints[2]
		} else if req.SilenceSeconds >= 40 && len(hints) >= 2 {
			hint = hints[1]
		} else {
			hint = hints[0]
		}
	}

	c.JSON(http.StatusOK, gin.H{"hint": hint, "hints": hints})
}

func UploadInterviewRecording(c *gin.Context) {
	userID := c.GetUint("user_id")
	interviewID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interview ID"})
		return
	}

	file, err := c.FormFile("recording")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Recording file is required"})
		return
	}

	dirPath := filepath.Join("uploads", "interviews", strconv.FormatUint(interviewID, 10))
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	filename := strconv.FormatInt(time.Now().Unix(), 10) + "_" + filepath.Base(file.Filename)
	targetPath := filepath.Join(dirPath, filename)
	if err := c.SaveUploadedFile(file, targetPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save recording file"})
		return
	}

	relativeURL := "/" + filepath.ToSlash(targetPath)
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	publicURL := scheme + "://" + c.Request.Host + relativeURL
	interview, err := service.SaveInterviewRecording(userID, uint(interviewID), publicURL)
	if err != nil {
		_ = os.Remove(targetPath)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "Recording uploaded successfully",
		"recording_url":    interview.RecordingURL,
		"recording_status": interview.RecordingStatus,
	})
}

// DrawBlindBoxScenario draws a random interview scenario for the blindbox mode
func DrawBlindBoxScenario(c *gin.Context) {
	var req struct {
		MinPressure string `json:"min_pressure"` // optional: "low", "medium", "high", "extreme"
	}
	c.ShouldBindJSON(&req)

	bbService := service.NewBlindBoxService()

	var scenario *service.BlindBoxScenario
	if req.MinPressure != "" {
		scenario = bbService.DrawScenarioByPressure(req.MinPressure)
	} else {
		scenario = bbService.DrawScenario()
	}

	c.JSON(http.StatusOK, gin.H{
		"scenario": scenario,
	})
}

// GetBlindBoxScenarios returns all available blindbox scenarios
func GetBlindBoxScenarios(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"scenarios": service.GetAllScenarios(),
	})
}

// ========== Human Interviewer Endpoints ==========

// GetHumanInterviewers lists available human interviewers
func GetHumanInterviewers(c *gin.Context) {
	interviewerType := c.Query("type") // campus, enterprise, or empty for all
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	interviewers, total, err := service.GetHumanInterviewers(interviewerType, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"interviewers": interviewers,
		"total":        total,
		"page":         page,
		"page_size":    pageSize,
	})
}

// GetHumanInterviewer returns details of a specific human interviewer
func GetHumanInterviewer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interviewer ID"})
		return
	}

	interviewer, err := service.GetHumanInterviewerByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Interviewer not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"interviewer": interviewer})
}

// BookHumanInterview creates a booking for a human interviewer
func BookHumanInterview(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		InterviewerID uint   `json:"interviewer_id" binding:"required"`
		ScheduledAt   string `json:"scheduled_at" binding:"required"` // ISO 8601 format
		Position      string `json:"position" binding:"required"`
		Difficulty    string `json:"difficulty" binding:"required"`
		Notes         string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "时间格式无效，请使用 ISO 8601 格式"})
		return
	}

	if scheduledAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "预约时间不能早于当前时间"})
		return
	}

	booking, err := service.BookHumanInterview(userID, req.InterviewerID, scheduledAt, req.Position, req.Difficulty, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "预约成功",
		"booking": booking,
	})
}

// GetUserBookings returns the user's interview bookings
func GetUserBookings(c *gin.Context) {
	userID := c.GetUint("user_id")
	bookings, err := service.GetUserBookings(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"bookings": bookings})
}

// SubmitHumanFeedback allows submitting human interviewer feedback
func SubmitHumanFeedback(c *gin.Context) {
	interviewID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interview ID"})
		return
	}

	var req struct {
		Feedback string `json:"feedback" binding:"required"`
		Score    int    `json:"score" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Score < 0 || req.Score > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "评分范围为0-100"})
		return
	}

	if err := service.SubmitHumanFeedback(uint(interviewID), req.Feedback, req.Score); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "真人反馈已提交"})
}

// RevealRandomStyle reveals the hidden style after a random-mode interview ends
func RevealRandomStyle(c *gin.Context) {
	userID := c.GetUint("user_id")
	interviewID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interview ID"})
		return
	}

	style, company, err := service.RevealRandomStyle(userID, uint(interviewID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	styleLabels := map[string]string{
		"gentle": "温和型", "stress": "压力型", "deep": "技术深挖型",
		"practical": "项目实战型", "algorithm": "算法考察型",
	}
	companyLabels := map[string]string{
		"ali": "阿里巴巴", "bytedance": "字节跳动", "tencent": "腾讯",
		"meituan": "美团", "baidu": "百度",
	}

	styleLabel := styleLabels[style]
	if styleLabel == "" {
		styleLabel = style
	}
	companyLabel := companyLabels[company]

	c.JSON(http.StatusOK, gin.H{
		"style":         style,
		"style_label":   styleLabel,
		"company":       company,
		"company_label": companyLabel,
	})
}

// GetInterviewConfig returns available interview configuration options
func GetInterviewConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"modes": []gin.H{
			{"key": "technical", "label": "技术面", "description": "纯技术能力考察，涵盖编程、算法、系统设计"},
			{"key": "hr", "label": "HR面", "description": "软实力与职业素养考察，沟通表达、职业规划、团队协作"},
			{"key": "comprehensive", "label": "综合面", "description": "技术+HR联合面试，模拟企业终面"},
		},
		"styles": []gin.H{
			{"key": "gentle", "label": "温和型", "description": "友好引导型面试官，鼓励自由表达"},
			{"key": "stress", "label": "压力型", "description": "高压追问，模拟真实压力面试场景"},
			{"key": "deep", "label": "技术深挖型", "description": "追问到底层原理，源码级考察"},
			{"key": "practical", "label": "项目实战型", "description": "围绕简历项目经历深入追问"},
			{"key": "algorithm", "label": "算法考察型", "description": "重点考察算法能力和编码实现"},
		},
		"companies": []gin.H{
			{"key": "", "label": "不限", "description": "通用面试风格"},
			{"key": "ali", "label": "阿里巴巴", "description": "重系统设计与业务思考"},
			{"key": "bytedance", "label": "字节跳动", "description": "极致深挖，快节奏面试"},
			{"key": "tencent", "label": "腾讯", "description": "轻松但有深度，重技术品味"},
			{"key": "meituan", "label": "美团", "description": "务实导向，重实战经验"},
			{"key": "baidu", "label": "百度", "description": "重基础功底，算法与工程规范"},
		},
		"difficulties": []gin.H{
			{"key": "campus_intern", "label": "校招实习", "description": "适合在校实习生，考察基础能力"},
			{"key": "campus_graduate", "label": "校招应届", "description": "适合应届毕业生，需一定项目经验"},
			{"key": "social_junior", "label": "社招初级", "description": "适合1-3年经验，要求实战能力"},
		},
		"interview_modes": []gin.H{
			{"key": "ai", "label": "AI仿真面试官", "icon": "🤖", "description": "AI模拟真实面试官进行面试"},
			{"key": "human", "label": "真人面试", "icon": "👤", "description": "邀请学校端/企业端账号参与模拟面试"},
			{"key": "random", "label": "随机模式", "icon": "🎲", "description": "系统随机分配面试风格，不提前告知"},
		},
	})
}
