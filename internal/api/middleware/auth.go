package middleware

import (
	"net/http"

	"accesslog-tracker/internal/api/models"
	"accesslog-tracker/internal/domain/services"
	"accesslog-tracker/internal/utils/logger"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware は認証ミドルウェアの構造体です
type AuthMiddleware struct {
	applicationService services.ApplicationServiceInterface
	logger             logger.Logger
}

// NewAuthMiddleware は新しい認証ミドルウェアを作成します
func NewAuthMiddleware(applicationService services.ApplicationServiceInterface, logger logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		applicationService: applicationService,
		logger:             logger,
	}
}

// Authenticate はAPIキーによる認証を行います
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// APIキーをヘッダーから取得
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			m.logger.Warn("API key not provided", "path", c.Request.URL.Path, "ip", c.ClientIP())
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "AUTHENTICATION_ERROR",
					Message: "API key is required",
				},
			})
			c.Abort()
			return
		}

		// 仕様準拠: APIキーのプレフィックス制約は課さない

		// アプリケーションの存在確認
		app, err := m.applicationService.GetByAPIKey(c.Request.Context(), apiKey)
		if err != nil {
			m.logger.Warn("Invalid API key", "path", c.Request.URL.Path, "ip", c.ClientIP(), "error", err.Error())
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "INVALID_API_KEY",
					Message: "Invalid API key",
				},
			})
			c.Abort()
			return
		}

		// アプリケーションがアクティブかチェック
		if !app.Active {
			m.logger.Warn("Inactive application", "app_id", app.AppID, "path", c.Request.URL.Path, "ip", c.ClientIP())
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "APPLICATION_INACTIVE",
					Message: "Application is inactive",
				},
			})
			c.Abort()
			return
		}

		// コンテキストにアプリケーション情報を設定
		c.Set("app_id", app.AppID)
		c.Set("application", app)

		m.logger.Debug("Authentication successful", "app_id", app.AppID, "path", c.Request.URL.Path, "ip", c.ClientIP())
		c.Next()
	}
}

// OptionalAuth はオプショナルな認証を行います（認証が失敗しても処理を続行）
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.Next()
			return
		}

		app, err := m.applicationService.GetByAPIKey(c.Request.Context(), apiKey)
		if err != nil {
			c.Next()
			return
		}

		if app.Active {
			c.Set("app_id", app.AppID)
			c.Set("application", app)
		}

		c.Next()
	}
}
