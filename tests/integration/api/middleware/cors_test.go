package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"accesslog-tracker/internal/api/middleware"
)

func TestCORSMiddlewareIntegration(t *testing.T) {
	// ルーターをセットアップ
	router := gin.New()
	router.Use(middleware.CORS())

	// テスト用ルートを追加
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	router.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	t.Run("should_handle_cors_preflight_request", func(t *testing.T) {
		req, _ := http.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type, X-API-Key")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 204, w.Code)
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("should_handle_cors_request_with_allowed_origin", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("should_handle_cors_request_with_another_allowed_origin", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://example.com")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("should_handle_cors_request_with_wildcard_subdomain", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://sub.example.com")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 403, w.Code)
		assert.Equal(t, "", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("should_handle_cors_request_without_origin", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		// Originヘッダーがない場合、CORSヘッダーは設定されない
		assert.Equal(t, "", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("should_handle_post_request_with_cors", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/test", nil)
		req.Header.Set("Origin", "http://localhost:8080")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "http://localhost:8080", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("should_handle_cors_with_custom_headers", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("X-API-Key", "test-key")
		req.Header.Set("Authorization", "Bearer token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("should_handle_cors_preflight_with_multiple_methods", func(t *testing.T) {
		req, _ := http.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "PUT")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 204, w.Code)
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
		// PUTメソッドは許可されていないため、Allow-Methodsヘッダーには含まれない
		assert.NotEqual(t, "PUT", w.Header().Get("Access-Control-Allow-Methods"))
	})

	t.Run("should_handle_cors_preflight_with_allowed_method", func(t *testing.T) {
		req, _ := http.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "DELETE")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 204, w.Code)
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
		// DELETEメソッドは許可されている
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "DELETE")
	})
}
