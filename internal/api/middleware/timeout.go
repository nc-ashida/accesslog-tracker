package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// TimeoutConfig はタイムアウトの設定
type TimeoutConfig struct {
	Timeout time.Duration
}

// DefaultTimeoutConfig はデフォルトのタイムアウト設定
func DefaultTimeoutConfig() TimeoutConfig {
	return TimeoutConfig{
		Timeout: 30 * time.Second,
	}
}

// Timeout はタイムアウトミドルウェア
func Timeout(config TimeoutConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), config.Timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		done := make(chan bool, 1)
		go func() {
			c.Next()
			done <- true
		}()

		select {
		case <-done:
			return
		case <-ctx.Done():
			c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{
				"success":   false,
				"message":   "Request timeout",
				"timestamp": time.Now().Format(time.RFC3339),
			})
			return
		}
	}
}

// CustomTimeout カスタムタイムアウトミドルウェア
func CustomTimeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		done := make(chan bool, 1)
		go func() {
			c.Next()
			done <- true
		}()

		select {
		case <-done:
			return
		case <-ctx.Done():
			c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{
				"success":   false,
				"message":   "Request timeout",
				"timestamp": time.Now().Format(time.RFC3339),
			})
			return
		}
	}
}

// TrackingTimeout トラッキング用タイムアウト（短いタイムアウト）
func TrackingTimeout() gin.HandlerFunc {
	return CustomTimeout(5 * time.Second)
}

// LongRunningTimeout 長時間実行用タイムアウト（長いタイムアウト）
func LongRunningTimeout() gin.HandlerFunc {
	return CustomTimeout(2 * time.Minute)
}

// ContextTimeout コンテキストタイムアウトミドルウェア
func ContextTimeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		// タイムアウトを監視
		go func() {
			select {
			case <-ctx.Done():
				if ctx.Err() == context.DeadlineExceeded {
					// logrus.WithFields(logrus.Fields{
					// 	"method":  c.Request.Method,
					// 	"path":    c.Request.URL.Path,
					// 	"timeout": timeout,
					// }).Warn("Context timeout")
				}
			}
		}()

		c.Next()
	}
}

// DatabaseTimeout データベース操作用タイムアウト
func DatabaseTimeout() gin.HandlerFunc {
	return ContextTimeout(10 * time.Second)
}

// ExternalAPITimeout 外部API呼び出し用タイムアウト
func ExternalAPITimeout() gin.HandlerFunc {
	return ContextTimeout(15 * time.Second)
}
