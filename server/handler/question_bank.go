package handler

import (
	"net/http"
	"strconv"

	"your-project/service"

	"github.com/gin-gonic/gin"
)

func GetQuestionBankMeta(c *gin.Context) {
	resp, err := mustQuestionBankService().GetMeta(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func GetQuestionBankOptions(c *gin.Context) {
	resp, err := mustQuestionBankService().GetFilterOptions(c.Request.Context(), c.Query("position_code"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func ListQuestionBankQuestions(c *gin.Context) {
	userID := c.GetUint("user_id")
	resp, err := mustQuestionBankService().ListQuestions(c.Request.Context(), userID, service.QuestionBankListFilters{
		PositionCode: c.Query("position_code"),
		Level:        c.Query("level"),
		QuestionType: c.Query("question_type"),
		Specialty:    c.Query("specialty"),
		Point:        c.Query("point"),
		CompanyType:  c.Query("company_type"),
		Status:       c.Query("status"),
		Keyword:      c.Query("keyword"),
		ListKey:      c.Query("list_key"),
		Page:         parseIntQuery(c, "page", 1),
		PageSize:     parseIntQuery(c, "page_size", 15),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func GetQuestionBankLists(c *gin.Context) {
	userID := c.GetUint("user_id")
	resp, err := mustQuestionBankService().GetQuestionLists(c.Request.Context(), userID, c.Query("position_code"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": resp})
}

func DrawQuestionBankQuestion(c *gin.Context) {
	userID := c.GetUint("user_id")
	resp, err := mustQuestionBankService().DrawQuestion(c.Request.Context(), userID, service.QuestionBankDrawRequest{
		PositionCode: c.Query("position_code"),
		Level:        c.Query("level"),
		QuestionType: c.Query("question_type"),
		Specialty:    c.Query("specialty"),
		Point:        c.Query("point"),
		CompanyType:  c.Query("company_type"),
		ListKey:      c.Query("list_key"),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func SubmitQuestionBankAnswer(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req struct {
		QuestionID     uint   `json:"question_id" binding:"required"`
		UserAnswer     string `json:"user_answer"`
		ErrorReason    string `json:"error_reason"`
		ElapsedSeconds *int   `json:"elapsed_seconds"`
		TimedMode      bool   `json:"timed_mode"`
		IsTimeout      bool   `json:"is_timeout"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := mustQuestionBankService().SubmitAnswer(c.Request.Context(), userID, service.QuestionBankAnswerRequest{
		QuestionID:     req.QuestionID,
		UserAnswer:     req.UserAnswer,
		ErrorReason:    req.ErrorReason,
		ElapsedSeconds: req.ElapsedSeconds,
		TimedMode:      req.TimedMode,
		IsTimeout:      req.IsTimeout,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func GetQuestionBankSolution(c *gin.Context) {
	questionID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question id"})
		return
	}
	resp, err := mustQuestionBankService().GetSolution(c.Request.Context(), uint(questionID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func GetQuestionBankPointSummary(c *gin.Context) {
	userID := c.GetUint("user_id")
	resp, err := mustQuestionBankService().GetPointSummary(
		c.Request.Context(),
		userID,
		c.Query("position_code"),
		c.Param("point"),
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func StartQuestionBankAssessment(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req struct {
		PositionCode string `json:"position_code" binding:"required"`
		Difficulty   string `json:"difficulty"`
		TotalCount   int    `json:"total_count"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := mustQuestionBankService().StartAssessment(c.Request.Context(), userID, service.QuestionBankAssessmentStartRequest{
		PositionCode: req.PositionCode,
		Difficulty:   req.Difficulty,
		TotalCount:   req.TotalCount,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func SubmitQuestionBankAssessmentAnswer(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req struct {
		AssessmentID   uint   `json:"assessment_id" binding:"required"`
		QuestionID     uint   `json:"question_id" binding:"required"`
		UserAnswer     string `json:"user_answer"`
		ElapsedSeconds *int   `json:"elapsed_seconds"`
		IsTimeout      bool   `json:"is_timeout"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := mustQuestionBankService().SubmitAssessmentAnswer(c.Request.Context(), userID, service.QuestionBankAssessmentAnswerRequest{
		AssessmentID:   req.AssessmentID,
		QuestionID:     req.QuestionID,
		UserAnswer:     req.UserAnswer,
		ElapsedSeconds: req.ElapsedSeconds,
		IsTimeout:      req.IsTimeout,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func CompleteQuestionBankAssessment(c *gin.Context) {
	userID := c.GetUint("user_id")
	assessmentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assessment id"})
		return
	}

	resp, err := mustQuestionBankService().CompleteAssessment(c.Request.Context(), userID, uint(assessmentID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func parseIntQuery(c *gin.Context, key string, fallback int) int {
	raw := c.Query(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}
