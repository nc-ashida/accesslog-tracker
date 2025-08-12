package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/your-username/accesslog-tracker/internal/api/handlers"
	"github.com/your-username/accesslog-tracker/internal/api/middleware"
	"github.com/your-username/accesslog-tracker/internal/beacon/generator"
	"github.com/your-username/accesslog-tracker/internal/domain/services"
	"github.com/your-username/accesslog-tracker/internal/infrastructure/cache/redis"
	"github.com/your-username/accesslog-tracker/internal/infrastructure/database/postgresql"
	"gorm.io/gorm"
)

// V1Routes v1 APIルートを設定
func V1Routes(
	router *gin.Engine,
	db *gorm.DB,
	redisClient *redis.Connection,
	authConfig middleware.AuthConfig,
	rateLimitConfig middleware.RateLimitConfig,
) {
	// リポジトリを作成
	trackingRepo := postgresql.NewTrackingRepository(db)
	applicationRepo := postgresql.NewApplicationRepository(db)
	sessionRepo := postgresql.NewSessionRepository(db)
	cacheService := redis.NewCacheService(redisClient.GetClient())

	// サービスを作成
	trackingService := services.NewTrackingService(trackingRepo, sessionRepo, cacheService)
	applicationService := services.NewApplicationService(applicationRepo, cacheService)
	statisticsService := services.NewStatisticsService(trackingRepo, cacheService)
	webhookService := services.NewWebhookService(cacheService)

	// ビーコン生成器を作成
	beaconGenerator := generator.NewBeaconGenerator()

	// ハンドラーを作成
	trackingHandler := handlers.NewTrackingHandler(trackingService, applicationService)
	beaconHandler := handlers.NewBeaconHandler(beaconGenerator)
	healthHandler := handlers.NewHealthHandler(db, redisClient)
	statisticsHandler := handlers.NewStatisticsHandler(statisticsService, applicationService)
	applicationHandler := handlers.NewApplicationHandler(applicationService)
	webhookHandler := handlers.NewWebhookHandler(webhookService, applicationService)

	// v1 APIグループ
	v1 := router.Group("/api/v1")
	{
		// ヘルスチェック（認証不要）
		v1.GET("/health", healthHandler.Health)
		v1.GET("/health/live", healthHandler.Liveness)
		v1.GET("/health/ready", healthHandler.Readiness)
		v1.GET("/metrics", healthHandler.Metrics)

		// トラッキングエンドポイント（API Key認証、レート制限あり）
		v1.Use(middleware.TrackingCORS())
		v1.Use(middleware.TrackingRateLimit(rateLimitConfig))
		v1.Use(middleware.TrackingTimeout())
		
		// 仕様書準拠のトラッキングエンドポイント（API Key認証）
		tracking := v1.Group("")
		tracking.Use(middleware.APIKeyAuthMiddleware())
		{
			tracking.POST("/track", trackingHandler.Track)
		}

		// ビーコン関連エンドポイント（認証不要）
		v1.GET("/beacon", beaconHandler.GetBeaconFile)
		v1.POST("/beacon/generate", beaconHandler.GenerateBeacon)

		// 認証が必要なエンドポイント
		authenticated := v1.Group("")
		authenticated.Use(middleware.AuthMiddleware(authConfig))
		authenticated.Use(middleware.CORS(middleware.DefaultCORSConfig()))
		authenticated.Use(middleware.RateLimit(rateLimitConfig))
		authenticated.Use(middleware.Timeout(middleware.DefaultTimeoutConfig()))
		{
			// アプリケーション管理
			applications := authenticated.Group("/applications")
			{
				applications.POST("", applicationHandler.CreateApplication)
				applications.GET("", applicationHandler.ListApplications)
				applications.GET("/:application_id", applicationHandler.GetApplication)
				applications.PUT("/:application_id", applicationHandler.UpdateApplication)
				applications.DELETE("/:application_id", applicationHandler.DeleteApplication)
				applications.GET("/:application_id/api-key", applicationHandler.GetApplicationAPIKey)
				applications.POST("/:application_id/api-key/regenerate", applicationHandler.RegenerateAPIKey)
			}

			// トラッキングデータ管理
			trackingData := authenticated.Group("/applications/:application_id/tracking")
			{
				trackingData.GET("/stats", trackingHandler.GetTrackingStats)
				trackingData.GET("/sessions/:session_id", trackingHandler.GetSessionData)
			}

			// 統計データ
			statistics := authenticated.Group("/applications/:application_id/statistics")
			{
				statistics.GET("/overview", statisticsHandler.GetApplicationStats)
				statistics.GET("/page-views", statisticsHandler.GetPageViews)
				statistics.GET("/referrers", statisticsHandler.GetReferrers)
				statistics.GET("/user-agents", statisticsHandler.GetUserAgents)
				statistics.GET("/geographic", statisticsHandler.GetGeographicStats)
				statistics.GET("/time-series", statisticsHandler.GetTimeSeriesData)
				statistics.GET("/custom-params/:param_name", statisticsHandler.GetCustomParamStats)
				statistics.GET("/real-time", statisticsHandler.GetRealTimeStats)
			}

			// Webhook管理
			webhooks := authenticated.Group("/applications/:application_id/webhooks")
			{
				webhooks.POST("", webhookHandler.CreateWebhook)
				webhooks.GET("", webhookHandler.ListWebhooks)
				webhooks.GET("/:webhook_id", webhookHandler.GetWebhook)
				webhooks.PUT("/:webhook_id", webhookHandler.UpdateWebhook)
				webhooks.DELETE("/:webhook_id", webhookHandler.DeleteWebhook)
				webhooks.POST("/:webhook_id/test", webhookHandler.TestWebhook)
				webhooks.GET("/:webhook_id/logs", webhookHandler.GetWebhookLogs)
			}
		}

		// オプショナル認証エンドポイント（認証があればユーザー情報を取得、なければ匿名）
		optionalAuth := v1.Group("")
		optionalAuth.Use(middleware.OptionalAuthMiddleware(authConfig))
		optionalAuth.Use(middleware.CORS(middleware.DefaultCORSConfig()))
		optionalAuth.Use(middleware.RateLimit(rateLimitConfig))
		{
			// 公開統計データ（認証があれば詳細情報、なければ基本情報のみ）
			publicStats := optionalAuth.Group("/public/statistics")
			{
				publicStats.GET("/:application_id/overview", statisticsHandler.GetApplicationStats)
				publicStats.GET("/:application_id/page-views", statisticsHandler.GetPageViews)
			}
		}
	}
}
