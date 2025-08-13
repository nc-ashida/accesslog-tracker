package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"accesslog-tracker/internal/utils/logger"
)

// Logging はログミドルウェアを設定します
func Logging(log logger.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// リクエスト情報をログに記録
		log.Info("HTTP Request",
			"method", param.Method,
			"path", param.Path,
			"status", param.StatusCode,
			"latency", param.Latency,
			"client_ip", param.ClientIP,
			"user_agent", param.Request.UserAgent(),
			"error", param.ErrorMessage,
		)
		
		return ""
	})
}

// RequestLogging は詳細なリクエストログを記録します
func RequestLogging(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// リクエスト開始時のログ
		log.Debug("Request started",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"query", c.Request.URL.RawQuery,
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)
		
		// リクエスト処理
		c.Next()
		
		// レスポンス完了時のログ
		duration := time.Since(start)
		status := c.Writer.Status()
		
		log.Debug("Request completed",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", status,
			"duration", duration,
			"size", c.Writer.Size(),
		)
		
		// エラーの場合は警告ログ
		if status >= 400 {
			log.Warn("Request error",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"status", status,
				"duration", duration,
				"client_ip", c.ClientIP(),
			)
		}
	}
}
