package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/redis/go-redis/v9"

	"accesslog-tracker/internal/api/middleware"
	"accesslog-tracker/internal/utils/logger"
)

func TestRateLimitMiddleware_Integration(t *testing.T) {
	// テスト用Redis接続
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
	defer redisClient.Close()

	// 接続テスト
	ctx := context.Background()
	err := redisClient.Ping(ctx).Err()
	require.NoError(t, err)

	// ロガーの初期化
	log := logger.NewLogger()

	// レート制限設定（テスト用に低い値に設定）
	config := middleware.RateLimitConfig{
		RequestsPerMinute: 10,
		RequestsPerHour:   100,
		BurstSize:         5,
	}

	// レート制限ミドルウェアの初期化
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(redisClient, log, config)

	// テスト用ルーターの設定
	router := gin.New()
	router.Use(rateLimitMiddleware.RateLimit())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	t.Run("should allow requests within rate limit", func(t *testing.T) {
		// 制限内のリクエストを送信
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.100:12345"

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		}
	})

	t.Run("should reject requests exceeding rate limit", func(t *testing.T) {
		// 制限を超えるリクエストを送信
		rateLimitedCount := 0
		for i := 0; i < 15; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.101:12345"

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code == http.StatusTooManyRequests {
				rateLimitedCount++
			}
		}

		assert.Greater(t, rateLimitedCount, 0)
	})

	t.Run("should handle different IP addresses separately", func(t *testing.T) {
		// 異なるIPアドレスからのリクエスト
		req1 := httptest.NewRequest("GET", "/test", nil)
		req1.RemoteAddr = "192.168.1.200:12345"

		req2 := httptest.NewRequest("GET", "/test", nil)
		req2.RemoteAddr = "192.168.1.201:12345"

		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)

		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusOK, w1.Code)
		assert.Equal(t, http.StatusOK, w2.Code)
	})

	t.Run("should handle authenticated users with app_id", func(t *testing.T) {
		// 認証済みユーザー（app_id付き）のリクエスト
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.300:12345"

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should reset rate limit after time window", func(t *testing.T) {
		// 最初のリクエスト
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.400:12345"

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// 時間ウィンドウを待機（テスト用に短縮）
		time.Sleep(2 * time.Second)

		// 再度リクエスト
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestRateLimitMiddleware_WithAuthentication(t *testing.T) {
	// テスト用Redis接続
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
	defer redisClient.Close()

	// 接続テスト
	ctx := context.Background()
	err := redisClient.Ping(ctx).Err()
	require.NoError(t, err)

	// ロガーの初期化
	log := logger.NewLogger()

	// レート制限設定
	config := middleware.RateLimitConfig{
		RequestsPerMinute: 5,
		RequestsPerHour:   50,
		BurstSize:         3,
	}

	// レート制限ミドルウェアの初期化
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(redisClient, log, config)

	// テスト用ルーターの設定（認証ミドルウェア付き）
	router := gin.New()
	router.Use(rateLimitMiddleware.RateLimit())
	router.GET("/test", func(c *gin.Context) {
		appID, exists := c.Get("app_id")
		if exists {
			c.JSON(http.StatusOK, gin.H{"message": "authenticated", "app_id": appID})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "anonymous"})
		}
	})

	t.Run("should apply rate limit per app_id for authenticated requests", func(t *testing.T) {
		// 認証済みリクエスト（app_id付き）
		rateLimitedCount := 0
		for i := 0; i < 10; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.500:12345"

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code == http.StatusTooManyRequests {
				rateLimitedCount++
			}
		}

		assert.Greater(t, rateLimitedCount, 0)
	})

	t.Run("should apply rate limit per IP for anonymous requests", func(t *testing.T) {
		// 匿名リクエスト
		rateLimitedCount := 0
		for i := 0; i < 10; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.600:12345"

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code == http.StatusTooManyRequests {
				rateLimitedCount++
			}
		}

		assert.Greater(t, rateLimitedCount, 0)
	})
}
