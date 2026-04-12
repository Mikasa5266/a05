package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"your-project/internal/service"

	"github.com/gin-gonic/gin"
)

func GetPracticeMeta(c *gin.Context) {
	resp, err := mustPracticeService().GetMeta(c.Request.Context())
	if err != nil {
		respondPracticeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func GetPracticeOptions(c *gin.Context) {
	resp, err := mustPracticeService().GetFilterOptions(c.Request.Context(), queryOrAlias(c, "position_code", "role"))
	if err != nil {
		respondPracticeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func ListPracticeQuestions(c *gin.Context) {
	resp, err := mustPracticeService().ListQuestions(c.Request.Context(), c.GetUint("user_id"), service.PracticeListQuestionsRequest{
		PositionCode: queryOrAlias(c, "position_code", "role"),
		Level:        c.Query("level"),
		QuestionType: queryOrAlias(c, "question_type", "qtype"),
		Specialty:    c.Query("specialty"),
		Point:        c.Query("point"),
		CompanyType:  c.Query("company_type"),
		Status:       c.Query("status"),
		Keyword:      c.Query("keyword"),
		FavoriteOnly: parseBoolQuery(c, "favorite"),
		ListID:       parseUintQuery(c, "list_id"),
		Page:         parseIntQuery(c, "page", 1),
		PageSize:     parseIntQuery(c, "page_size", 15),
	})
	if err != nil {
		respondPracticeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func GetPracticeQuestionLists(c *gin.Context) {
	resp, err := mustPracticeService().GetQuestionLists(c.Request.Context(), c.GetUint("user_id"), queryOrAlias(c, "position_code", "role"))
	if err != nil {
		respondPracticeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": resp})
}

func GetPracticeQuestionsOfList(c *gin.Context) {
	listID, ok := parseUintPathParam(c, "id")
	if !ok {
		respondAPIError(c, http.StatusBadRequest, string(service.PracticeErrorInvalidArgument), "invalid list id")
		return
	}

	resp, err := mustPracticeService().ListQuestions(c.Request.Context(), c.GetUint("user_id"), service.PracticeListQuestionsRequest{
		PositionCode: queryOrAlias(c, "position_code", "role"),
		QuestionType: queryOrAlias(c, "question_type", "qtype"),
		Status:       c.Query("status"),
		Keyword:      c.Query("keyword"),
		ListID:       listID,
		Page:         parseIntQuery(c, "page", 1),
		PageSize:     parseIntQuery(c, "page_size", 15),
	})
	if err != nil {
		respondPracticeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func GetPracticeSpecialties(c *gin.Context) {
	resp, err := mustPracticeService().GetSpecialties(c.Request.Context(), queryOrAlias(c, "position_code", "role"))
	if err != nil {
		respondPracticeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"role":        queryOrAlias(c, "position_code", "role"),
		"specialties": resp,
	})
}

func DrawPracticeQuestion(c *gin.Context) {
	resp, err := mustPracticeService().DrawQuestion(c.Request.Context(), c.GetUint("user_id"), service.PracticeDrawRequest{
		PositionCode: queryOrAlias(c, "position_code", "role"),
		Level:        c.Query("level"),
		QuestionType: queryOrAlias(c, "question_type", "qtype"),
		Specialty:    c.Query("specialty"),
		Point:        c.Query("point"),
		CompanyType:  c.Query("company_type"),
		ListID:       parseUintQuery(c, "list_id"),
	})
	if err != nil {
		respondPracticeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func SubmitPracticeAnswer(c *gin.Context) {
	var req struct {
		QuestionID     uint   `json:"question_id" binding:"required"`
		UserAnswer     string `json:"user_answer"`
		ErrorReason    string `json:"error_reason"`
		ElapsedSeconds *int   `json:"elapsed_seconds"`
		TimedMode      bool   `json:"timed_mode"`
		IsTimeout      bool   `json:"is_timeout"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondAPIError(c, http.StatusBadRequest, string(service.PracticeErrorInvalidArgument), err.Error())
		return
	}

	resp, err := mustPracticeService().SubmitAnswer(c.Request.Context(), c.GetUint("user_id"), service.PracticeAnswerRequest{
		QuestionID:     req.QuestionID,
		UserAnswer:     req.UserAnswer,
		ErrorReason:    req.ErrorReason,
		ElapsedSeconds: req.ElapsedSeconds,
		TimedMode:      req.TimedMode,
		IsTimeout:      req.IsTimeout,
	})
	if err != nil {
		respondPracticeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func GetPracticePointSummary(c *gin.Context) {
	resp, err := mustPracticeService().GetPointSummary(
		c.Request.Context(),
		c.GetUint("user_id"),
		queryOrAlias(c, "position_code", "role"),
		c.Query("point"),
	)
	if err != nil {
		respondPracticeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func GetPracticeWrongRemedial(c *gin.Context) {
	wrongID, ok := parseUintPathParam(c, "id")
	if !ok {
		respondAPIError(c, http.StatusBadRequest, string(service.PracticeErrorInvalidArgument), "invalid wrong id")
		return
	}

	resp, err := mustPracticeService().GetWrongRemedial(c.Request.Context(), c.GetUint("user_id"), wrongID)
	if err != nil {
		respondPracticeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func GetPracticeSolution(c *gin.Context) {
	questionID, ok := parseUintPathParam(c, "id")
	if !ok {
		respondAPIError(c, http.StatusBadRequest, string(service.PracticeErrorInvalidArgument), "invalid question id")
		return
	}

	resp, err := mustPracticeService().GetSolution(c.Request.Context(), questionID)
	if err != nil {
		respondPracticeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func StartPracticeAssessment(c *gin.Context) {
	var req struct {
		PositionCode string `json:"position_code"`
		Role         string `json:"role"`
		Difficulty   string `json:"difficulty"`
		TotalCount   int    `json:"total_count"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondAPIError(c, http.StatusBadRequest, string(service.PracticeErrorInvalidArgument), err.Error())
		return
	}

	resp, err := mustPracticeService().StartAssessment(c.Request.Context(), c.GetUint("user_id"), service.PracticeAssessmentStartRequest{
		PositionCode: firstNonEmpty(req.PositionCode, req.Role),
		Difficulty:   req.Difficulty,
		TotalCount:   req.TotalCount,
	})
	if err != nil {
		respondPracticeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func SubmitPracticeAssessmentAnswer(c *gin.Context) {
	var req struct {
		AssessmentID   uint   `json:"assessment_id" binding:"required"`
		QuestionID     uint   `json:"question_id" binding:"required"`
		UserAnswer     string `json:"user_answer"`
		ElapsedSeconds *int   `json:"elapsed_seconds"`
		IsTimeout      bool   `json:"is_timeout"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondAPIError(c, http.StatusBadRequest, string(service.PracticeErrorInvalidArgument), err.Error())
		return
	}

	resp, err := mustPracticeService().SubmitAssessmentAnswer(c.Request.Context(), c.GetUint("user_id"), service.PracticeAssessmentAnswerRequest{
		AssessmentID:   req.AssessmentID,
		QuestionID:     req.QuestionID,
		UserAnswer:     req.UserAnswer,
		ElapsedSeconds: req.ElapsedSeconds,
		IsTimeout:      req.IsTimeout,
	})
	if err != nil {
		respondPracticeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func CompletePracticeAssessment(c *gin.Context) {
	assessmentID, ok := parseUintPathParam(c, "id")
	if !ok {
		respondAPIError(c, http.StatusBadRequest, string(service.PracticeErrorInvalidArgument), "invalid assessment id")
		return
	}

	resp, err := mustPracticeService().CompleteAssessment(c.Request.Context(), c.GetUint("user_id"), assessmentID)
	if err != nil {
		respondPracticeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func GetPracticeIntegrationSnapshot(c *gin.Context) {
	resp, err := mustPracticeService().GetIntegrationSnapshot(c.Request.Context(), c.GetUint("user_id"), queryOrAlias(c, "position_code", "role"))
	if err != nil {
		respondPracticeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func SubmitPracticeIntegrationFeedback(c *gin.Context) {
	var req struct {
		PositionCode string   `json:"position_code"`
		Role         string   `json:"role"`
		WeakPoints   []string `json:"weak_points"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondAPIError(c, http.StatusBadRequest, string(service.PracticeErrorInvalidArgument), err.Error())
		return
	}

	resp, err := mustPracticeService().SubmitIntegrationFeedback(c.Request.Context(), c.GetUint("user_id"), service.PracticeIntegrationFeedbackRequest{
		PositionCode: firstNonEmpty(req.PositionCode, req.Role),
		WeakPoints:   req.WeakPoints,
	})
	if err != nil {
		respondPracticeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func GetPracticeWrongs(c *gin.Context) {
	resp, err := mustPracticeService().ListWrongs(c.Request.Context(), c.GetUint("user_id"), service.PracticeWrongsRequest{
		PositionCode: queryOrAlias(c, "position_code", "role"),
		Point:        c.Query("point"),
		QuestionType: queryOrAlias(c, "question_type", "qtype"),
		FavoriteOnly: parseBoolQuery(c, "favorite"),
	})
	if err != nil {
		respondPracticeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": resp})
}

func DeletePracticeWrong(c *gin.Context) {
	wrongID, ok := parseUintPathParam(c, "id")
	if !ok {
		respondAPIError(c, http.StatusBadRequest, string(service.PracticeErrorInvalidArgument), "invalid wrong id")
		return
	}
	if err := mustPracticeService().DeleteWrong(c.Request.Context(), c.GetUint("user_id"), wrongID); err != nil {
		respondPracticeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func TogglePracticeWrongFavorite(c *gin.Context) {
	wrongID, ok := parseUintPathParam(c, "id")
	if !ok {
		respondAPIError(c, http.StatusBadRequest, string(service.PracticeErrorInvalidArgument), "invalid wrong id")
		return
	}
	state, err := mustPracticeService().ToggleWrongFavorite(c.Request.Context(), c.GetUint("user_id"), wrongID)
	if err != nil {
		respondPracticeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"is_favorite": state})
}

func TogglePracticeQuestionFavorite(c *gin.Context) {
	questionID, ok := parseUintPathParam(c, "question_id")
	if !ok {
		respondAPIError(c, http.StatusBadRequest, string(service.PracticeErrorInvalidArgument), "invalid question id")
		return
	}
	state, err := mustPracticeService().ToggleQuestionFavorite(c.Request.Context(), c.GetUint("user_id"), questionID)
	if err != nil {
		respondPracticeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"is_favorite": state})
}

func GetPracticeDashboard(c *gin.Context) {
	resp, err := mustPracticeService().GetDashboard(c.Request.Context(), c.GetUint("user_id"), queryOrAlias(c, "position_code", "role"))
	if err != nil {
		respondPracticeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func ImportPracticeQuestionBank(c *gin.Context) {
	var req service.PracticeImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondAPIError(c, http.StatusBadRequest, string(service.PracticeErrorInvalidArgument), err.Error())
		return
	}
	imported, err := mustPracticeService().ImportQuestionBank(c.Request.Context(), req)
	if err != nil {
		respondPracticeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"imported": imported})
}

func ExportPracticeRecords(c *gin.Context) {
	resp, err := mustPracticeService().ExportRecords(c.Request.Context(), c.GetUint("user_id"))
	if err != nil {
		respondPracticeError(c, err)
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", resp.Filename))
	c.Data(http.StatusOK, "text/csv; charset=utf-8", []byte(resp.Content))
}

func GetPracticeQuestionByID(c *gin.Context) {
	questionID, ok := parseUintPathParam(c, "id")
	if !ok {
		respondAPIError(c, http.StatusBadRequest, string(service.PracticeErrorInvalidArgument), "invalid question id")
		return
	}
	resp, err := mustPracticeService().GetQuestionByID(c.Request.Context(), c.GetUint("user_id"), questionID)
	if err != nil {
		respondPracticeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func queryOrAlias(c *gin.Context, primary, alias string) string {
	if value := strings.TrimSpace(c.Query(primary)); value != "" {
		return value
	}
	return strings.TrimSpace(c.Query(alias))
}

func parseBoolQuery(c *gin.Context, key string) bool {
	value := strings.TrimSpace(strings.ToLower(c.Query(key)))
	return value == "1" || value == "true" || value == "yes" || value == "on"
}

func parseIntQuery(c *gin.Context, key string, fallback int) int {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func parseUintQuery(c *gin.Context, key string) uint {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return 0
	}
	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0
	}
	return uint(value)
}

func parseUintPathParam(c *gin.Context, key string) (uint, bool) {
	value, err := strconv.ParseUint(c.Param(key), 10, 64)
	if err != nil {
		return 0, false
	}
	return uint(value), true
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
