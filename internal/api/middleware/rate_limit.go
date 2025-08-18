package middleware

import (
	"fmt"
	"net/http"
	"time"

	"accesslog-tracker/internal/api/models"
	"accesslog-tracker/internal/utils/iputil"
	"accesslog-tracker/internal/utils/logger"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

// RateLimitConfig はレート制限の設定です
type RateLimitConfig struct {
	RequestsPerMinute int
	RequestsPerHour   int
	BurstSize         int
}

// RateLimitMiddleware はレート制限ミドルウェアの構造体です
type RateLimitMiddleware struct {
	redisClient *redis.Client
	logger      logger.Logger
	config      RateLimitConfig
}

// NewRateLimitMiddleware は新しいレート制限ミドルウェアを作成します
func NewRateLimitMiddleware(redisClient *redis.Client, logger logger.Logger, config RateLimitConfig) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		redisClient: redisClient,
		logger:      logger,
		config:      config,
	}
}

// RateLimit はレート制限を適用します
func (m *RateLimitMiddleware) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// ヘッダーをマップに変換
		headers := make(map[string]string)
		for key, values := range c.Request.Header {
			if len(values) > 0 {
				headers[key] = values[0]
			}
		}
		clientIP := iputil.GetClientIP(headers, c.Request.RemoteAddr)
		appID, exists := c.Get("app_id")
		if !exists {
			appID = "anonymous"
		}

		minuteKey := fmt.Sprintf("rate_limit:%s:%s:minute", appID, clientIP)
		ctx := context.Background()

		if err := m.checkRateLimit(ctx, minuteKey, m.config.RequestsPerMinute, time.Minute); err != nil {
			m.logger.Warnf("Rate limit exceeded for app_id: %s, ip: %s", appID, clientIP)
			c.JSON(http.StatusTooManyRequests, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "RATE_LIMIT_EXCEEDED",
					Message: "Rate limit exceeded. Please try again later.",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (m *RateLimitMiddleware) checkRateLimit(ctx context.Context, key string, limit int, window time.Duration) error {
	current, err := m.redisClient.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		return err
	}

	if current >= limit {
		return fmt.Errorf("rate limit exceeded")
	}

	pipe := m.redisClient.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)

	_, err = pipe.Exec(ctx)
	return err
}

func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerMinute: 2000,  // 増加
		RequestsPerHour:   20000, // 増加
		BurstSize:         200,   // 増加
	}
}
