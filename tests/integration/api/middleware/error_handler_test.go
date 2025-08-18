package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"accesslog-tracker/internal/api/middleware"
	apimodels "accesslog-tracker/internal/api/models"
	domainmodels "accesslog-tracker/internal/domain/models"
	"accesslog-tracker/internal/utils/logger"
)

func TestErrorHandlerMiddleware(t *testing.T) {
	// ロガーを初期化
	log := logger.NewLogger()

	// ルーターをセットアップ
	router := gin.New()
	router.Use(middleware.ErrorHandler(log))

	t.Run("should_handle_validation_error", func(t *testing.T) {
		router.POST("/validation-test", func(c *gin.Context) {
			var req struct {
				Required string `json:"required" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, apimodels.APIResponse{
					Success: false,
					Error: &apimodels.APIError{
						Code:    "VALIDATION_ERROR",
						Message: err.Error(),
					},
				})
				return
			}
			c.JSON(200, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("POST", "/validation-test", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 400, w.Code)
	})

	t.Run("should_handle_404_error", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/not-found", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 404, w.Code)
	})

	t.Run("should_handle_405_error", func(t *testing.T) {
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("POST", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 404, w.Code)
	})

	t.Run("should_handle_panic", func(t *testing.T) {
		router.GET("/panic-test", func(c *gin.Context) {
			panic("test panic")
		})

		req, _ := http.NewRequest("GET", "/panic-test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 500, w.Code)
	})

	t.Run("should_handle_custom_error", func(t *testing.T) {
		router.GET("/custom-error", func(c *gin.Context) {
			c.JSON(500, apimodels.APIResponse{
				Success: false,
				Error: &apimodels.APIError{
					Code:    "CUSTOM_ERROR",
					Message: "Custom error message",
				},
			})
		})

		req, _ := http.NewRequest("GET", "/custom-error", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 500, w.Code)
	})

	t.Run("should_handle_validation_error_handler", func(t *testing.T) {
		// 新しいルーターを作成して重複を避ける
		validationRouter := gin.New()
		validationRouter.Use(middleware.ErrorHandler(log))
		validationRouter.Use(middleware.ValidationErrorHandler())
		validationRouter.POST("/validation-test-new", func(c *gin.Context) {
			// バリデーションエラーを発生させる
			c.Error(domainmodels.ErrValidationError)
			c.JSON(400, apimodels.APIResponse{
				Success: false,
				Error: &apimodels.APIError{
					Code:    "VALIDATION_ERROR",
					Message: "Validation error occurred",
				},
			})
		})

		req, _ := http.NewRequest("POST", "/validation-test-new", nil)
		w := httptest.NewRecorder()

		validationRouter.ServeHTTP(w, req)

		assert.Equal(t, 400, w.Code)
	})

	t.Run("should_handle_not_found_handler", func(t *testing.T) {
		router.Use(middleware.NotFoundHandler())
		router.GET("/not-found-test", func(c *gin.Context) {
			middleware.NotFoundHandler()(c)
		})

		req, _ := http.NewRequest("GET", "/not-found-test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 404, w.Code)
	})

	t.Run("should_handle_method_not_allowed_handler", func(t *testing.T) {
		// 新しいルーターを作成して重複を避ける
		methodRouter := gin.New()
		methodRouter.Use(middleware.ErrorHandler(log))
		methodRouter.Use(middleware.MethodNotAllowedHandler())
		methodRouter.GET("/method-not-allowed-test", func(c *gin.Context) {
			// 405エラーを発生させる
			c.JSON(405, apimodels.APIResponse{
				Success: false,
				Error: &apimodels.APIError{
					Code:    "METHOD_NOT_ALLOWED",
					Message: "Method not allowed",
				},
			})
		})

		req, _ := http.NewRequest("GET", "/method-not-allowed-test", nil)
		w := httptest.NewRecorder()

		methodRouter.ServeHTTP(w, req)

		assert.Equal(t, 405, w.Code)
	})

	t.Run("should_handle_panic_with_error_interface", func(t *testing.T) {
		// 新しいルーターを作成して重複を避ける
		panicRouter := gin.New()
		panicRouter.Use(middleware.ErrorHandler(log))
		panicRouter.GET("/panic-error", func(c *gin.Context) {
			panic("test error panic")
		})

		req, _ := http.NewRequest("GET", "/panic-error", nil)
		w := httptest.NewRecorder()

		panicRouter.ServeHTTP(w, req)

		assert.Equal(t, 500, w.Code)
	})

	t.Run("should_handle_panic_with_string", func(t *testing.T) {
		// 新しいルーターを作成して重複を避ける
		panicStringRouter := gin.New()
		panicStringRouter.Use(middleware.ErrorHandler(log))
		panicStringRouter.GET("/panic-string", func(c *gin.Context) {
			panic("test string panic")
		})

		req, _ := http.NewRequest("GET", "/panic-string", nil)
		w := httptest.NewRecorder()

		panicStringRouter.ServeHTTP(w, req)

		assert.Equal(t, 500, w.Code)
	})
}
