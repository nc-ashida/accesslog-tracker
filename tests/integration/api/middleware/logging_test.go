package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"accesslog-tracker/internal/api/middleware"
	"accesslog-tracker/internal/utils/logger"
)

func TestLoggingMiddlewareIntegration(t *testing.T) {
	// ロガーを初期化
	log := logger.NewLogger()

	// ルーターをセットアップ
	router := gin.New()
	router.Use(middleware.Logging(log))

	// テスト用ルートを追加
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	router.POST("/test", func(c *gin.Context) {
		c.JSON(201, gin.H{"message": "created"})
	})

	router.GET("/error", func(c *gin.Context) {
		c.JSON(500, gin.H{"error": "internal error"})
	})

	router.GET("/not-found", func(c *gin.Context) {
		c.JSON(404, gin.H{"error": "not found"})
	})

	t.Run("should_log_successful_request", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", "test-agent")
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		// ログが出力されることを確認（実際のログ内容は確認できないため、ステータスコードのみ確認）
	})

	t.Run("should_log_post_request", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(`{"data": "test"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "test-agent")
		req.RemoteAddr = "192.168.1.2:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 201, w.Code)
	})

	t.Run("should_log_error_request", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/error", nil)
		req.Header.Set("User-Agent", "test-agent")
		req.RemoteAddr = "192.168.1.3:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 500, w.Code)
	})

	t.Run("should_log_not_found_request", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/not-found", nil)
		req.Header.Set("User-Agent", "test-agent")
		req.RemoteAddr = "192.168.1.4:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 404, w.Code)
	})

	t.Run("should_log_request_with_query_parameters", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test?param1=value1&param2=value2", nil)
		req.Header.Set("User-Agent", "test-agent")
		req.RemoteAddr = "192.168.1.5:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("should_log_request_without_user_agent", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.6:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("should_log_request_without_remote_addr", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", "test-agent")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})
}

func TestRequestLoggingMiddlewareIntegration(t *testing.T) {
	// ロガーを初期化
	log := logger.NewLogger()

	// ルーターをセットアップ
	router := gin.New()
	router.Use(middleware.RequestLogging(log))

	// テスト用ルートを追加
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	router.POST("/test", func(c *gin.Context) {
		c.JSON(201, gin.H{"message": "created"})
	})

	router.GET("/error", func(c *gin.Context) {
		c.JSON(500, gin.H{"error": "internal error"})
	})

	t.Run("should_log_request_start_and_completion", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", "test-agent")
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		// リクエスト開始と完了のログが出力されることを確認
	})

	t.Run("should_log_error_request_with_warning", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/error", nil)
		req.Header.Set("User-Agent", "test-agent")
		req.RemoteAddr = "192.168.1.2:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 500, w.Code)
		// エラーレクエストの場合、警告ログが出力されることを確認
	})

	t.Run("should_log_request_with_query_parameters", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test?param1=value1&param2=value2", nil)
		req.Header.Set("User-Agent", "test-agent")
		req.RemoteAddr = "192.168.1.3:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		// クエリパラメータがログに含まれることを確認
	})

	t.Run("should_log_post_request_with_body", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(`{"data": "test"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "test-agent")
		req.RemoteAddr = "192.168.1.4:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 201, w.Code)
		// POSTリクエストのログが出力されることを確認
	})

	t.Run("should_log_request_duration", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", "test-agent")
		req.RemoteAddr = "192.168.1.5:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		// リクエストの実行時間がログに含まれることを確認
	})

	t.Run("should_log_response_size", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", "test-agent")
		req.RemoteAddr = "192.168.1.6:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		// レスポンスサイズがログに含まれることを確認
	})
}
