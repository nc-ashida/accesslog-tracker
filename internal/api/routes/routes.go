package routes

import (
	"github.com/gin-gonic/gin"
	"accesslog-tracker/internal/api/handlers"
	"accesslog-tracker/internal/api/middleware"
	"accesslog-tracker/internal/domain/services"
	"accesslog-tracker/internal/infrastructure/database/postgresql"
	"accesslog-tracker/internal/infrastructure/cache/redis"
	"accesslog-tracker/internal/utils/logger"
)

// Setup はAPIルートを設定します
func Setup(
	router *gin.Engine,
	trackingService *services.TrackingService,
	applicationService *services.ApplicationService,
	dbConn *postgresql.Connection,
	redisConn *redis.CacheService,
	log logger.Logger,
) {
	// ミドルウェアの設定
	authMiddleware := middleware.NewAuthMiddleware(applicationService, log)
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(redisConn.GetClient(), log, middleware.DefaultRateLimitConfig())

	// グローバルミドルウェア
	router.Use(middleware.CORS())
	router.Use(middleware.Logging(log))
	router.Use(middleware.RequestLogging(log))
	router.Use(middleware.ErrorHandler(log))
	router.Use(middleware.ValidationErrorHandler())

	// ヘルスチェックエンドポイント（認証不要）
	healthHandler := handlers.NewHealthHandler(dbConn, redisConn, log)
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Readiness)
	router.GET("/live", healthHandler.Liveness)

	// API v1 ルートグループ
	v1 := router.Group("/v1")
	{
		// トラッキングエンドポイント（認証必須）
		trackingHandler := handlers.NewTrackingHandler(trackingService, log)
		tracking := v1.Group("/tracking")
		tracking.Use(authMiddleware.Authenticate())
		tracking.Use(rateLimitMiddleware.RateLimit())
		{
			tracking.POST("/track", trackingHandler.Track)
			tracking.GET("/statistics", trackingHandler.GetStatistics)
		}

		// アプリケーション管理エンドポイント（認証不要）
		applicationHandler := handlers.NewApplicationHandler(applicationService, log)
		applications := v1.Group("/applications")
		applications.Use(rateLimitMiddleware.RateLimit())
		{
			applications.POST("", applicationHandler.Create)
			applications.GET("", applicationHandler.List)
			applications.GET("/:id", applicationHandler.Get)
			applications.PUT("/:id", applicationHandler.Update)
			applications.DELETE("/:id", applicationHandler.Delete)
		}

		// ビーコン関連エンドポイント（認証不要）
		beaconHandler := handlers.NewBeaconHandler()
		beacon := v1.Group("/beacon")
		beacon.Use(rateLimitMiddleware.RateLimit())
		{
			beacon.GET("/generate", beaconHandler.GenerateBeacon)
			beacon.POST("/generate", beaconHandler.GenerateBeaconWithConfig)
			beacon.GET("/health", beaconHandler.Health)
		}
	}

	// ビーコン配信ルート（APIバージョンなし、認証不要）
	beaconHandler := handlers.NewBeaconHandler()
	router.GET("/tracker.js", beaconHandler.Serve)
	router.GET("/tracker.min.js", beaconHandler.ServeMinified)
	router.GET("/tracker/:app_id.js", beaconHandler.ServeCustom)

	// 404ハンドラー
	router.NoRoute(middleware.NotFoundHandler())

	// 405ハンドラー
	router.NoMethod(middleware.MethodNotAllowedHandler())
}

// SetupTest はテスト用のルートを設定します
func SetupTest(
	router *gin.Engine,
	trackingService *services.TrackingService,
	applicationService *services.ApplicationService,
	dbConn *postgresql.Connection,
	redisConn *redis.CacheService,
	log logger.Logger,
) {
	// テスト用のミドルウェア設定（認証を緩和）
	authMiddleware := middleware.NewAuthMiddleware(applicationService, log)
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(redisConn.GetClient(), log, middleware.DefaultRateLimitConfig())

	// グローバルミドルウェア
	router.Use(middleware.CORS())
	router.Use(middleware.Logging(log))
	router.Use(middleware.ErrorHandler(log))

	// ヘルスチェックエンドポイント
	healthHandler := handlers.NewHealthHandler(dbConn, redisConn, log)
	router.GET("/health", healthHandler.Health)

	// API v1 ルートグループ
	v1 := router.Group("/v1")
	{
		// トラッキングエンドポイント（テスト用に認証を緩和）
		trackingHandler := handlers.NewTrackingHandler(trackingService, log)
		tracking := v1.Group("/tracking")
		tracking.Use(authMiddleware.OptionalAuth()) // オプショナル認証
		tracking.Use(rateLimitMiddleware.RateLimit())
		{
			tracking.POST("/track", trackingHandler.Track)
			tracking.GET("/statistics", trackingHandler.GetStatistics)
		}

		// アプリケーション管理エンドポイント
		applicationHandler := handlers.NewApplicationHandler(applicationService, log)
		applications := v1.Group("/applications")
		applications.Use(rateLimitMiddleware.RateLimit())
		{
			applications.POST("", applicationHandler.Create)
			applications.GET("", applicationHandler.List)
			applications.GET("/:id", applicationHandler.Get)
			applications.PUT("/:id", applicationHandler.Update)
			applications.DELETE("/:id", applicationHandler.Delete)
		}

		// ビーコン関連エンドポイント（テスト用）
		beaconHandler := handlers.NewBeaconHandler()
		beacon := v1.Group("/beacon")
		beacon.Use(rateLimitMiddleware.RateLimit())
		{
			beacon.GET("/generate", beaconHandler.GenerateBeacon)
			beacon.POST("/generate", beaconHandler.GenerateBeaconWithConfig)
			beacon.GET("/health", beaconHandler.Health)
		}
	}

	// ビーコン配信ルート（テスト用）
	beaconHandler := handlers.NewBeaconHandler()
	router.GET("/tracker.js", beaconHandler.Serve)
	router.GET("/tracker.min.js", beaconHandler.ServeMinified)
	router.GET("/tracker/:app_id.js", beaconHandler.ServeCustom)

	// 404ハンドラー
	router.NoRoute(middleware.NotFoundHandler())
}
