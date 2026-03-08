package handler

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"your-project/service"

	"github.com/gin-gonic/gin"
)

const (
	SessionStateInit      = "INIT"
	SessionStateReady     = "READY"
	SessionStateAsking    = "ASKING"
	SessionStateListening = "LISTENING"
	SessionStateThinking  = "THINKING"
	SessionStateEnd       = "END"
)

func normalizeSessionState(state string) string {
	normalized := strings.ToUpper(strings.TrimSpace(state))
	switch normalized {
	case SessionStateInit, SessionStateReady, SessionStateAsking, SessionStateListening, SessionStateThinking, SessionStateEnd:
		return normalized
	default:
		return ""
	}
}

func deriveInterviewSessionState(status string, currentIndex int) string {
	if status == "completed" {
		return SessionStateEnd
	}
	if currentIndex <= 0 {
		return SessionStateReady
	}
	return SessionStateAsking
}

func sessionEvent(eventType, from, to string) gin.H {
	return gin.H{
		"type":      eventType,
		"from":      from,
		"to":        to,
		"timestamp": time.Now().UnixMilli(),
	}
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
		SessionState  string `json:"session_state"`
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

	interview, err := service.StartInterview(userID, req.Position, req.Difficulty, req.Mode, req.Style, req.Company, req.InterviewMode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":       "Interview started successfully",
		"interview":     interview,
		"session_state": SessionStateReady,
		"event":         sessionEvent("session.ready", normalizeSessionState(req.SessionState), SessionStateReady),
	})
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

	state := deriveInterviewSessionState(interview.Status, interview.CurrentIndex)
	c.JSON(http.StatusOK, gin.H{
		"interview":      interview,
		"session_state":  state,
		"connection_ok":  true,
		"recoverable":    true,
		"recovery_event": sessionEvent("session.synced", "", state),
	})
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
		QuestionID uint   `json:"question_id" binding:"required"`
		Answer     string `json:"answer"`
		AudioData  string `json:"audio_data,omitempty"`
		SessionState string `json:"session_state"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if strings.TrimSpace(req.Answer) == "" && strings.TrimSpace(req.AudioData) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "您似乎没有做出任何回答"})
		return
	}

	result, err := service.SubmitAnswer(userID, uint(interviewID), req.QuestionID, req.Answer, req.AudioData)
	if err != nil {
		errMsg := err.Error()
		log.Printf("SubmitAnswer failed: user=%d interview=%d question=%d hasAudio=%t err=%s", userID, interviewID, req.QuestionID, strings.TrimSpace(req.AudioData) != "", errMsg)

		switch {
		case strings.Contains(errMsg, "unauthorized access"):
			c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
		case strings.Contains(errMsg, "question not found"):
			c.JSON(http.StatusNotFound, gin.H{"error": errMsg})
		case strings.Contains(errMsg, "interview is not in progress"):
			c.JSON(http.StatusBadRequest, gin.H{"error": "面试已结束或未开始，无法继续作答"})
		case strings.Contains(errMsg, "failed to transcribe audio") || strings.Contains(errMsg, "whisper transcription failed"):
			c.JSON(http.StatusBadGateway, gin.H{"error": errMsg})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		}
		return
	}

	updatedInterview, _ := service.GetInterviewByID(userID, uint(interviewID))
	nextState := SessionStateAsking
	eventType := "interviewer.asking"
	if updatedInterview != nil && updatedInterview.Status == "completed" {
		nextState = SessionStateEnd
		eventType = "session.ended"
	}
	fromState := normalizeSessionState(req.SessionState)
	if fromState == "" {
		fromState = SessionStateListening
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Answer submitted successfully",
		"result":        result,
		"session_state": nextState,
		"event":         sessionEvent(eventType, fromState, nextState),
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
		"message":       "Interview ended successfully",
		"interview":     interview,
		"session_state": SessionStateEnd,
		"event":         sessionEvent("session.ended", SessionStateAsking, SessionStateEnd),
	})
}

func GetInterviewSession(c *gin.Context) {
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

	state := deriveInterviewSessionState(interview.Status, interview.CurrentIndex)
	c.JSON(http.StatusOK, gin.H{
		"interview_id":   interview.ID,
		"status":         interview.Status,
		"session_state":  state,
		"connection_ok":  true,
		"recoverable":    interview.Status != "completed",
		"recovery_event": sessionEvent("session.ping", "", state),
	})
}

// AnalyzeSpeechChunk receives a short audio chunk and returns real-time speech metrics
func AnalyzeSpeechChunk(c *gin.Context) {
	var req struct {
		AudioData string  `json:"audio_data" binding:"required"`
		Duration  float64 `json:"duration" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := service.NewSpeechAnalysisService()
	metrics, err := svc.AnalyzeAudioChunk(req.AudioData, req.Duration)
	if err != nil {
		log.Printf("AnalyzeSpeechChunk failed: duration=%.2f err=%s", req.Duration, err.Error())
		// Real-time chunk analysis should be best-effort. If ASR fails for a chunk,
		// return empty metrics instead of surfacing 500 to avoid noisy frontend errors.
		fallback := svc.AnalyzeText("", req.Duration)
		c.JSON(http.StatusOK, gin.H{
			"metrics": fallback,
			"warning": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
	})
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
			{"key": "human", "label": "真人面试", "icon": "👤", "description": "预约真人面试官进行面试"},
			{"key": "random", "label": "随机模式", "icon": "🎲", "description": "系统随机分配面试风格，不提前告知"},
		},
	})
}
