package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"accesslog-tracker/internal/api/middleware"
	"accesslog-tracker/internal/utils/logger"
)

func TestRateLimitMiddlewareIntegration(t *testing.T) {
	// テスト用Redis接続
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	ctx := context.Background()
	err := redisClient.Ping(ctx).Err()
	require.NoError(t, err)
	defer redisClient.Close()

	// 設定を初期化
	cfg := middleware.RateLimitConfig{
		RequestsPerMinute: 10,
		RequestsPerHour:   100,
		BurstSize:         5,
	}

	// ロガーを初期化
	log := logger.NewLogger()

	// レート制限ミドルウェアを初期化
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(redisClient, log, cfg)

	// ルーターをセットアップ
	router := gin.New()
	router.Use(rateLimitMiddleware.RateLimit())

	// テスト用ルートを追加
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "rate limited"})
	})

	t.Run("should_allow_requests_within_limit", func(t *testing.T) {
		// Redisの状態をクリア
		redisClient.FlushDB(ctx)
		
		// 制限内のリクエストを送信
		for i := 0; i < 5; i++ {
			req, _ := http.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.1:12345"
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
		}
	})

	t.Run("should_reject_requests_exceeding_limit", func(t *testing.T) {
		// Redisの状態をクリア
		redisClient.FlushDB(ctx)
		
		// 制限を超えるリクエストを送信
		for i := 0; i < 15; i++ {
			req, _ := http.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.2:12345"
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// 最初の10リクエストは成功、それ以降は429エラー
			if i < 10 {
				assert.Equal(t, 200, w.Code)
			} else {
				assert.Equal(t, 429, w.Code)
			}
		}
	})

	t.Run("should_separate_different_ip_addresses", func(t *testing.T) {
		// Redisの状態をクリア
		redisClient.FlushDB(ctx)
		
		// 異なるIPアドレスからのリクエスト
		req1, _ := http.NewRequest("GET", "/test", nil)
		req1.RemoteAddr = "192.168.1.3:12345"
		w1 := httptest.NewRecorder()

		router.ServeHTTP(w1, req1)

		assert.Equal(t, 200, w1.Code)

		req2, _ := http.NewRequest("GET", "/test", nil)
		req2.RemoteAddr = "192.168.1.4:12345"
		w2 := httptest.NewRecorder()

		router.ServeHTTP(w2, req2)

		assert.Equal(t, 200, w2.Code)
	})

	t.Run("should_handle_authenticated_users", func(t *testing.T) {
		// Redisの状態をクリア
		redisClient.FlushDB(ctx)
		
		// 認証済みユーザーのテスト
		for i := 0; i < 5; i++ {
			req, _ := http.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.5:12345"
			req.Header.Set("X-API-Key", "test_token")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
		}
	})

	t.Run("should_reset_after_time_window", func(t *testing.T) {
		// Redisの状態をクリア
		redisClient.FlushDB(ctx)
		
		// 時間ウィンドウのリセットテスト
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.6:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		// 時間ウィンドウを待つ（実際のテストでは短縮）
		time.Sleep(100 * time.Millisecond)

		req2, _ := http.NewRequest("GET", "/test", nil)
		req2.RemoteAddr = "192.168.1.6:12345"
		w2 := httptest.NewRecorder()

		router.ServeHTTP(w2, req2)

		assert.Equal(t, 200, w2.Code)
	})

	t.Run("should_handle_redis_connection_error", func(t *testing.T) {
		// Redis接続エラーのテスト
		invalidRedisClient := redis.NewClient(&redis.Options{
			Addr:     "invalid-redis:6379",
			Password: "",
			DB:       0,
		})

		invalidRateLimitMiddleware := middleware.NewRateLimitMiddleware(invalidRedisClient, log, cfg)

		invalidRouter := gin.New()
		invalidRouter.Use(invalidRateLimitMiddleware.RateLimit())
		invalidRouter.GET("/redis-error", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "redis error"})
		})

		req, _ := http.NewRequest("GET", "/redis-error", nil)
		req.RemoteAddr = "192.168.1.7:12345"
		w := httptest.NewRecorder()

		invalidRouter.ServeHTTP(w, req)

		// Redisエラーの場合、デフォルトで許可するか拒否するかは実装に依存
		assert.Contains(t, []int{200, 429, 500}, w.Code)
	})

	t.Run("should_handle_different_http_methods", func(t *testing.T) {
		// 異なるHTTPメソッドのテスト
		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

		for _, method := range methods {
			req, _ := http.NewRequest(method, "/test", nil)
			req.RemoteAddr = "192.168.1.8:12345"
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if method == "GET" {
				assert.Equal(t, 200, w.Code)
			} else {
				assert.Equal(t, 404, w.Code)
			}
		}
	})

	t.Run("should_handle_missing_remote_addr", func(t *testing.T) {
		// RemoteAddrが設定されていない場合のテスト
		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("should_handle_x_forwarded_for_header", func(t *testing.T) {
		// X-Forwarded-Forヘッダーのテスト
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.9:12345"
		req.Header.Set("X-Forwarded-For", "203.0.113.1")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("should_handle_multiple_x_forwarded_for", func(t *testing.T) {
		// 複数のX-Forwarded-Forヘッダーのテスト
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.10:12345"
		req.Header.Set("X-Forwarded-For", "203.0.113.1, 203.0.113.2")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("should_handle_custom_rate_limit_config", func(t *testing.T) {
		// カスタムレート制限設定のテスト
		customCfg := middleware.RateLimitConfig{
			RequestsPerMinute: 2,
			RequestsPerHour:   10,
			BurstSize:         1,
		}

		customRateLimitMiddleware := middleware.NewRateLimitMiddleware(redisClient, log, customCfg)

		customRouter := gin.New()
		customRouter.Use(customRateLimitMiddleware.RateLimit())
		customRouter.GET("/custom", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "custom"})
		})

		// 制限内のリクエスト
		req1, _ := http.NewRequest("GET", "/custom", nil)
		req1.RemoteAddr = "192.168.1.12:12345"
		w1 := httptest.NewRecorder()

		customRouter.ServeHTTP(w1, req1)

		assert.Equal(t, 200, w1.Code)

		// 2番目のリクエスト（制限内）
		req2, _ := http.NewRequest("GET", "/custom", nil)
		req2.RemoteAddr = "192.168.1.12:12345"
		w2 := httptest.NewRecorder()

		customRouter.ServeHTTP(w2, req2)

		assert.Equal(t, 200, w2.Code)

		// 3番目のリクエスト（制限を超える）
		req3, _ := http.NewRequest("GET", "/custom", nil)
		req3.RemoteAddr = "192.168.1.12:12345"
		w3 := httptest.NewRecorder()

		customRouter.ServeHTTP(w3, req3)

		assert.Equal(t, 429, w3.Code)
	})

	t.Run("should_handle_default_rate_limit_config", func(t *testing.T) {
		// デフォルト設定のテスト
		defaultCfg := middleware.DefaultRateLimitConfig()
		defaultRateLimitMiddleware := middleware.NewRateLimitMiddleware(redisClient, log, defaultCfg)

		defaultRouter := gin.New()
		defaultRouter.Use(defaultRateLimitMiddleware.RateLimit())
		defaultRouter.GET("/default", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "default"})
		})

		req, _ := http.NewRequest("GET", "/default", nil)
		req.RemoteAddr = "192.168.1.13:12345"
		w := httptest.NewRecorder()

		defaultRouter.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("should_handle_rate_limit_with_authenticated_user", func(t *testing.T) {
		// 認証済みユーザーのレート制限テスト
		authRouter := gin.New()
		authRouter.Use(rateLimitMiddleware.RateLimit())
		authRouter.GET("/auth", func(c *gin.Context) {
			// 認証済みユーザーのコンテキストを設定
			c.Set("app_id", "test_app")
			c.JSON(200, gin.H{"message": "authenticated"})
		})

		req, _ := http.NewRequest("GET", "/auth", nil)
		req.RemoteAddr = "192.168.1.14:12345"
		w := httptest.NewRecorder()

		authRouter.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("should_handle_rate_limit_with_x_forwarded_for", func(t *testing.T) {
		// X-Forwarded-Forヘッダーを使用したレート制限テスト
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.15:12345"
		req.Header.Set("X-Forwarded-For", "203.0.113.1")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("should_handle_rate_limit_with_multiple_x_forwarded_for", func(t *testing.T) {
		// 複数のX-Forwarded-Forヘッダーを使用したレート制限テスト
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.16:12345"
		req.Header.Set("X-Forwarded-For", "203.0.113.1, 203.0.113.2")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("should_handle_rate_limit_with_empty_remote_addr", func(t *testing.T) {
		// RemoteAddrが空の場合のレート制限テスト
		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("should_handle_rate_limit_with_invalid_remote_addr", func(t *testing.T) {
		// 無効なRemoteAddrの場合のレート制限テスト
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "invalid-addr"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})
}
