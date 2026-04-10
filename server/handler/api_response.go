package handler

import (
	"net/http"

	"your-project/service"

	"github.com/gin-gonic/gin"
)

type apiErrorPayload struct {
	Error apiErrorDetail `json:"error"`
}

type apiErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func respondAPIError(c *gin.Context, status int, code, message string) {
	c.JSON(status, apiErrorPayload{
		Error: apiErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

func respondPracticeError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	if practiceErr, ok := service.AsPracticeError(err); ok {
		respondAPIError(c, practiceErrorStatus(practiceErr.Code), string(practiceErr.Code), practiceErr.Message)
		return
	}

	respondAPIError(c, http.StatusInternalServerError, string(service.PracticeErrorInternal), "internal server error")
}

func practiceErrorStatus(code service.PracticeErrorCode) int {
	switch code {
	case service.PracticeErrorInvalidArgument:
		return http.StatusBadRequest
	case service.PracticeErrorUnauthorized:
		return http.StatusUnauthorized
	case service.PracticeErrorNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
