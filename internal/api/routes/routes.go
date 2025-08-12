package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/your-username/accesslog-tracker/internal/api/middleware"
	"github.com/your-username/accesslog-tracker/internal/infrastructure/cache/redis"
	"gorm.io/gorm"
)

// ServerConfig サーバー設定
type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	MaxHeaderBytes  int
	TrustedProxies  []string
}

// DefaultServerConfig デフォルトサーバー設定
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Port:           ":8080",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
		TrustedProxies: []string{"127.0.0.1", "::1"},
	}
}

// SetupRoutes ルートを設定
func SetupRoutes(
	db *gorm.DB,
	redisClient *redis.Connection,
	authConfig middleware.AuthConfig,
	rateLimitConfig middleware.RateLimitConfig,
	serverConfig ServerConfig,
) *gin.Engine {
	// 本番環境ではリリースモードに設定
	if gin.Mode() == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	// ルーターを作成
	router := gin.New()

	// 信頼できるプロキシを設定
	router.SetTrustedProxies(serverConfig.TrustedProxies)

	// グローバルミドルウェアを設定
	router.Use(middleware.RequestLogging())
	router.Use(middleware.ErrorLogging())
	router.Use(middleware.PerformanceLogging(5 * time.Second))

	// 404ハンドラー
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"error": "Endpoint not found",
			"path":  c.Request.URL.Path,
		})
	})

	// 405ハンドラー
	router.NoMethod(func(c *gin.Context) {
		c.JSON(405, gin.H{
			"error": "Method not allowed",
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
		})
	})

	// パニックリカバリー
	router.Use(gin.Recovery())

	// ルートを設定
	setupAPIRoutes(router, db, redisClient, authConfig, rateLimitConfig)

	return router
}

// setupAPIRoutes APIルートを設定
func setupAPIRoutes(
	router *gin.Engine,
	db *gorm.DB,
	redisClient *redis.Connection,
	authConfig middleware.AuthConfig,
	rateLimitConfig middleware.RateLimitConfig,
) {
	// ルートエンドポイント
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"name":        "Access Log Tracker API",
			"version":     "1.0.0",
			"description": "Web analytics and tracking API",
			"endpoints": gin.H{
				"health":     "/api/v1/health",
				"tracking":   "/api/v1/tracking",
				"docs":       "/docs",
				"swagger":    "/swagger",
			},
		})
	})

	// APIドキュメント（将来の実装）
	router.GET("/docs", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "API documentation - to be implemented",
		})
	})

	// Swagger UI（将来の実装）
	router.GET("/swagger", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Swagger UI - to be implemented",
		})
	})

	// v1 APIルートを設定
	V1Routes(router, db, redisClient, authConfig, rateLimitConfig)
}

// SetupDevelopmentRoutes 開発用ルートを設定
func SetupDevelopmentRoutes(
	router *gin.Engine,
	db *gorm.DB,
	redisClient *redis.Connection,
	authConfig middleware.AuthConfig,
	rateLimitConfig middleware.RateLimitConfig,
) {
	// 開発環境でのみ有効なルート
	if gin.Mode() == gin.DebugMode {
		dev := router.Group("/dev")
		{
			// 開発用ヘルスチェック
			dev.GET("/health", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"status": "development mode",
					"time":   time.Now(),
				})
			})

			// 開発用設定確認
			dev.GET("/config", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"auth_config": gin.H{
						"token_duration": authConfig.TokenDuration,
					},
					"rate_limit_config": gin.H{
						"requests_per_minute": rateLimitConfig.RequestsPerMinute,
						"requests_per_hour":   rateLimitConfig.RequestsPerHour,
					},
				})
			})

			// 開発用データベース接続テスト
			dev.GET("/db-test", func(c *gin.Context) {
				sqlDB, err := db.DB()
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}

				err = sqlDB.Ping()
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}

				c.JSON(200, gin.H{"status": "database connected"})
			})

			// 開発用Redis接続テスト
			dev.GET("/redis-test", func(c *gin.Context) {
				err := redisClient.Ping()
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}

				c.JSON(200, gin.H{"status": "redis connected"})
			})
		}
	}
}
