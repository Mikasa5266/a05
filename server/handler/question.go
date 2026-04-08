package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

var validQuestionBankPositions = map[string]struct{}{
	"backend":   {},
	"frontend":  {},
	"algorithm": {},
	"ai":        {},
}

func GetQuestionBankPositions(c *gin.Context) {
	positions, err := mustQuestionBankService().ListPositions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"positions": positions,
		"count":     len(positions),
	})
}

func GetPositionQuestionList(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	positionCode := strings.ToLower(strings.TrimSpace(c.Param("positionCode")))
	if _, ok := validQuestionBankPositions[positionCode]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "positionCode must be one of: backend, frontend, algorithm, ai"})
		return
	}

	difficulty := strings.TrimSpace(c.Query("difficulty"))
	limit := parsePositiveInt(c.Query("limit"), 12)

	questions, err := mustQuestionBankService().GeneratePositionQuestionList(c.Request.Context(), userID, positionCode, difficulty, limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"position_code": positionCode,
		"difficulty":    difficulty,
		"count":         len(questions),
		"questions":     questions,
	})
}

func GenerateResumeQuestionList(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	resumeResultID, err := strconv.ParseUint(c.Param("resumeResultID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid resumeResultID"})
		return
	}

	difficulty := strings.TrimSpace(c.Query("difficulty"))
	limit := parsePositiveInt(c.Query("limit"), 12)

	questions, svcErr := mustQuestionBankService().GenerateResumeQuestionList(c.Request.Context(), userID, uint(resumeResultID), difficulty, limit)
	if svcErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": svcErr.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"resume_result_id": resumeResultID,
		"difficulty":       difficulty,
		"count":            len(questions),
		"questions":        questions,
	})
}

func GetQuestion(c *gin.Context) {
	questionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question id"})
		return
	}

	question, svcErr := mustQuestionBankService().GetQuestion(c.Request.Context(), uint(questionID))
	if svcErr != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"question": question})
}

func EvaluateQuestion(c *gin.Context) {
	questionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question id"})
		return
	}

	var req struct {
		Answer string `json:"answer" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	question, svcErr := mustQuestionBankService().GetQuestion(c.Request.Context(), uint(questionID))
	if svcErr != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
		return
	}

	evaluation, evalErr := mustAIService().EvaluateAnswer(c.Request.Context(), question, strings.TrimSpace(req.Answer))
	if evalErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": evalErr.Error()})
		return
	}

	feedback := gin.H{"raw": evaluation.Feedback}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(evaluation.Feedback), &parsed); err == nil {
		feedback = parsed
	}

	c.JSON(http.StatusOK, gin.H{
		"question_id": question.ID,
		"score":       evaluation.Score,
		"feedback":    feedback,
	})
}

func SetQuestionFavorite(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	questionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question id"})
		return
	}

	var req struct {
		IsFavorite bool `json:"is_favorite"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := mustQuestionBankService().SetFavorite(c.Request.Context(), userID, uint(questionID), req.IsFavorite); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "favorite state updated"})
}

func MarkQuestionWrong(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	questionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question id"})
		return
	}

	var req struct {
		Note string `json:"note"`
	}
	_ = c.ShouldBindJSON(&req)

	if err := mustQuestionBankService().MarkWrongQuestion(c.Request.Context(), userID, uint(questionID), req.Note); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "wrong question marked"})
}

func ClearQuestionWrong(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	questionID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question id"})
		return
	}

	if err := mustQuestionBankService().ClearWrongQuestion(c.Request.Context(), userID, uint(questionID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "wrong question cleared"})
}

func ListFavoriteQuestions(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	page := parsePositiveInt(c.Query("page"), 1)
	pageSize := parsePositiveInt(c.Query("page_size"), 20)

	questions, total, err := mustQuestionBankService().ListFavorites(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"questions": questions,
	})
}

func ListWrongQuestions(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	page := parsePositiveInt(c.Query("page"), 1)
	pageSize := parsePositiveInt(c.Query("page_size"), 20)

	questions, total, err := mustQuestionBankService().ListWrongQuestions(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"questions": questions,
	})
}

func parsePositiveInt(raw string, fallback int) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}
