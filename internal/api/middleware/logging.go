package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// responseWriter レスポンスライターのラッパー
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// Logging ログミドルウェア
func Logging() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 構造化ログとして出力
		logrus.WithFields(logrus.Fields{
			"timestamp":    param.TimeStamp.Format(time.RFC3339),
			"status":       param.StatusCode,
			"latency":      param.Latency,
			"client_ip":    param.ClientIP,
			"method":       param.Method,
			"path":         param.Path,
			"user_agent":   param.Request.UserAgent(),
			"error":        param.ErrorMessage,
		}).Info("HTTP Request")

		return ""
	})
}

// RequestLogging 詳細なリクエストログミドルウェア
func RequestLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// リクエストボディを読み取り
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// レスポンスボディをキャプチャ
		blw := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = blw

		// リクエスト処理
		c.Next()

		// レスポンス時間を計算
		duration := time.Since(start)

		// ログフィールドを準備
		fields := logrus.Fields{
			"timestamp":     start.Format(time.RFC3339),
			"method":        c.Request.Method,
			"path":          c.Request.URL.Path,
			"query":         c.Request.URL.RawQuery,
			"status":        c.Writer.Status(),
			"latency":       duration,
			"client_ip":     c.ClientIP(),
			"user_agent":    c.Request.UserAgent(),
			"content_length": c.Writer.Size(),
		}

		// 認証情報があれば追加
		if userID, exists := c.Get("user_id"); exists {
			fields["user_id"] = userID
		}
		if applicationID, exists := c.Get("application_id"); exists {
			fields["application_id"] = applicationID
		}

		// エラーがあれば追加
		if len(c.Errors) > 0 {
			fields["errors"] = c.Errors.String()
		}

		// リクエストボディがあれば追加（機密情報は除外）
		if len(requestBody) > 0 && !isSensitivePath(c.Request.URL.Path) {
			fields["request_body"] = string(requestBody)
		}

		// レスポンスボディがあれば追加（機密情報は除外）
		if blw.body.Len() > 0 && !isSensitivePath(c.Request.URL.Path) {
			fields["response_body"] = blw.body.String()
		}

		// ログレベルを決定
		logger := logrus.WithFields(fields)
		if c.Writer.Status() >= 400 {
			logger.Error("HTTP Request Error")
		} else {
			logger.Info("HTTP Request")
		}
	}
}

// isSensitivePath 機密情報を含む可能性のあるパスかチェック
func isSensitivePath(path string) bool {
	sensitivePaths := []string{
		"/auth/login",
		"/auth/register",
		"/api/v1/auth",
		"/tracking", // トラッキングデータは機密情報
	}
	
	for _, sensitivePath := range sensitivePaths {
		if path == sensitivePath || (len(sensitivePath) > 0 && path[:len(sensitivePath)] == sensitivePath) {
			return true
		}
	}
	return false
}

// ErrorLogging エラーログミドルウェア
func ErrorLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// エラーがあればログ出力
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				logrus.WithFields(logrus.Fields{
					"method":        c.Request.Method,
					"path":          c.Request.URL.Path,
					"client_ip":     c.ClientIP(),
					"user_agent":    c.Request.UserAgent(),
					"status":        c.Writer.Status(),
				}).WithError(err.Err).Error("Request Error")
			}
		}
	}
}

// PerformanceLogging パフォーマンスログミドルウェア
func PerformanceLogging(threshold time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		// 閾値を超えた場合のみログ出力
		if duration > threshold {
			logrus.WithFields(logrus.Fields{
				"method":    c.Request.Method,
				"path":      c.Request.URL.Path,
				"latency":   duration,
				"threshold": threshold,
			}).Warn("Slow Request Detected")
		}
	}
}
