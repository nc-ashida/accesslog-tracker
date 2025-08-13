package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	apimodels "accesslog-tracker/internal/api/models"
	domainmodels "accesslog-tracker/internal/domain/models"
	"accesslog-tracker/internal/utils/logger"
)

// ErrorHandler はエラーハンドリングミドルウェアを設定します
func ErrorHandler(log logger.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			log.Error("Panic recovered", "error", err, "stack", string(debug.Stack()))
			c.JSON(http.StatusInternalServerError, apimodels.APIResponse{
				Success: false,
				Error: &apimodels.APIError{
					Code:    "INTERNAL_SERVER_ERROR",
					Message: "An unexpected error occurred",
				},
			})
		} else {
			log.Error("Panic recovered", "error", recovered, "stack", string(debug.Stack()))
			c.JSON(http.StatusInternalServerError, apimodels.APIResponse{
				Success: false,
				Error: &apimodels.APIError{
					Code:    "INTERNAL_SERVER_ERROR",
					Message: "An unexpected error occurred",
				},
			})
		}
	})
}

// ValidationErrorHandler はバリデーションエラーを処理します
func ValidationErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// バリデーションエラーが発生した場合
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				if validationErr, ok := err.Err.(error); ok && validationErr == domainmodels.ErrValidationError {
					c.JSON(http.StatusBadRequest, apimodels.APIResponse{
						Success: false,
						Error: &apimodels.APIError{
							Code:    "VALIDATION_ERROR",
							Message: validationErr.Error(),
						},
					})
					return
				}
			}
		}
	}
}

// NotFoundHandler は404エラーを処理します
func NotFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotFound, apimodels.APIResponse{
			Success: false,
			Error: &apimodels.APIError{
				Code:    "NOT_FOUND",
				Message: "The requested resource was not found",
			},
		})
	}
}

// MethodNotAllowedHandler は405エラーを処理します
func MethodNotAllowedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, apimodels.APIResponse{
			Success: false,
			Error: &apimodels.APIError{
				Code:    "METHOD_NOT_ALLOWED",
				Message: "The requested method is not allowed for this resource",
			},
		})
	}
}
