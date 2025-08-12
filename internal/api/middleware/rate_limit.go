package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimitConfig はレート制限の設定
type RateLimitConfig struct {
	RequestsPerMinute int
	BurstSize         int
	RedisClient       *redis.Client
}

// RateLimitInfo レート制限情報
type RateLimitInfo struct {
	Remaining int
	Reset     time.Time
	Limit     int
}

// RateLimit はレート制限ミドルウェア
func RateLimit(config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// API Keyを取得
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success":   false,
				"message":   "X-API-Key header is required",
				"timestamp": time.Now().Format(time.RFC3339),
			})
			c.Abort()
			return
		}

		// Redisキーを作成
		key := fmt.Sprintf("rate_limit:%s", apiKey)

		// 現在のリクエスト数を取得
		ctx := context.Background()
		current, err := config.RedisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success":   false,
				"message":   "Rate limit check failed",
				"timestamp": time.Now().Format(time.RFC3339),
			})
			c.Abort()
			return
		}

		// 制限をチェック
		if current >= config.RequestsPerMinute {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success":   false,
				"message":   "Rate limit exceeded",
				"timestamp": time.Now().Format(time.RFC3339),
			})
			c.Abort()
			return
		}

		// カウンターをインクリメント
		pipe := config.RedisClient.Pipeline()
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, time.Minute)
		_, err = pipe.Exec(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success":   false,
				"message":   "Rate limit update failed",
				"timestamp": time.Now().Format(time.RFC3339),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// TrackingRateLimit はトラッキング用のレート制限ミドルウェア
func TrackingRateLimit(config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// API Keyを取得
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success":   false,
				"message":   "X-API-Key header is required",
				"timestamp": time.Now().Format(time.RFC3339),
			})
			c.Abort()
			return
		}

		// Redisキーを作成
		key := fmt.Sprintf("tracking_rate_limit:%s", apiKey)

		// 現在のリクエスト数を取得
		ctx := context.Background()
		current, err := config.RedisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success":   false,
				"message":   "Rate limit check failed",
				"timestamp": time.Now().Format(time.RFC3339),
			})
			c.Abort()
			return
		}

		// 制限をチェック（トラッキングはより緩い制限）
		if current >= config.RequestsPerMinute*2 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success":   false,
				"message":   "Tracking rate limit exceeded",
				"timestamp": time.Now().Format(time.RFC3339),
			})
			c.Abort()
			return
		}

		// カウンターをインクリメント
		pipe := config.RedisClient.Pipeline()
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, time.Minute)
		_, err = pipe.Exec(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success":   false,
				"message":   "Rate limit update failed",
				"timestamp": time.Now().Format(time.RFC3339),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkRateLimit レート制限をチェック
func checkRateLimit(c *gin.Context, key string, limit int, window time.Duration, redisClient *redis.Client) bool {
	ctx := context.Background()
	
	// 現在のリクエスト数を取得
	current, err := redisClient.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		return false
	}
	
	// 制限をチェック
	if current >= limit {
		return false
	}
	
	// カウンターをインクリメント
	pipe := redisClient.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return false
	}
	
	return true
}

// getClientID はクライアント識別子を取得
func getClientID(c *gin.Context) string {
	// API Keyを優先
	if apiKey := c.GetHeader("X-API-Key"); apiKey != "" {
		return apiKey
	}
	
	// IPアドレスをフォールバック
	return c.ClientIP()
}

// GetRateLimitInfo レート制限情報を取得
func GetRateLimitInfo(c *gin.Context, config RateLimitConfig) (*RateLimitInfo, error) {
	clientID := getClientID(c)
	minuteKey := fmt.Sprintf("rate_limit:minute:%s", clientID)
	
	ctx := context.Background()
	current, err := config.RedisClient.Get(ctx, minuteKey).Int()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	// TTLを取得してリセット時間を計算
	ttl, err := config.RedisClient.TTL(ctx, minuteKey).Result()
	if err != nil {
		return nil, err
	}

	resetTime := time.Now().Add(ttl)
	remaining := config.RequestsPerMinute - current

	return &RateLimitInfo{
		Remaining: remaining,
		Reset:     resetTime,
		Limit:     config.RequestsPerMinute,
	}, nil
}
