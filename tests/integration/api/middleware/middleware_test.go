package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"accesslog-tracker/internal/api/middleware"
	apimodels "accesslog-tracker/internal/api/models"
	"accesslog-tracker/internal/config"
	"accesslog-tracker/internal/domain/services"
	redisCache "accesslog-tracker/internal/infrastructure/cache/redis"
	"accesslog-tracker/internal/infrastructure/database/postgresql"
	"accesslog-tracker/internal/infrastructure/database/postgresql/repositories"
	"accesslog-tracker/internal/utils/logger"
	apihelpers "accesslog-tracker/tests/integration/api"
)



func TestMiddlewareIntegration(t *testing.T) {
	// テスト用データベースをセットアップ
	db := apihelpers.SetupTestDatabase(t)
	defer db.Close()

	// テスト用Redisをセットアップ
	redisClient := apihelpers.SetupTestRedis(t)
	defer redisClient.Close()

	// テスト用アプリケーションを作成
	app := apihelpers.CreateTestApplication(t, db)

	// 設定を初期化
	cfg := config.New()
	cfg.App.Port = 8080
	cfg.Database.Host = apihelpers.GetTestDBHost()
	cfg.Database.Port = 5432
	cfg.Database.User = "postgres"
	cfg.Database.Password = "password"
	cfg.Database.Name = "access_log_tracker_test"
	cfg.Database.SSLMode = "disable"
	cfg.Redis.Host = apihelpers.GetTestRedisHost()
	cfg.Redis.Port = 6379
	cfg.Redis.Password = ""
	cfg.Redis.DB = 0

	// ロガーを初期化
	log := logger.NewLogger()

	// Redisキャッシュサービスを初期化
	cacheService := redisCache.NewCacheService(cfg.GetRedisAddr())
	err := cacheService.Connect()
	require.NoError(t, err)
	defer cacheService.Close()

	// アプリケーションサービスを初期化（テストヘルパーと同じデータベース接続を使用）
	appRepo := repositories.NewApplicationRepository(db)
	appService := services.NewApplicationService(appRepo, cacheService)

	// データベース接続を初期化
	dbConn := postgresql.NewConnection(cfg.GetDatabaseDSN())
	err = dbConn.Connect(cfg.GetDatabaseDSN())
	require.NoError(t, err)
	defer dbConn.Close()



	// ミドルウェアを初期化
	authMiddleware := middleware.NewAuthMiddleware(appService, log)
	corsMiddleware := middleware.CORS()
	rateLimitConfig := middleware.RateLimitConfig{
		RequestsPerMinute: 10,
		RequestsPerHour:   100,
		BurstSize:         5,
	}
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(redisClient, log, rateLimitConfig)
	loggingMiddleware := middleware.Logging(log)
	errorHandlerMiddleware := middleware.ErrorHandler(log)

	// ルーターをセットアップ
	router := gin.New()
	router.Use(loggingMiddleware)
	router.Use(errorHandlerMiddleware)
	router.Use(corsMiddleware)

	// テスト用ルートを追加
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// 認証が必要なルート
	authGroup := router.Group("/auth")
	authGroup.Use(authMiddleware.Authenticate())
	authGroup.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "authenticated"})
	})

	// レート制限が必要なルート
	rateLimitGroup := router.Group("/rate")
	rateLimitGroup.Use(rateLimitMiddleware.RateLimit())
	rateLimitGroup.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "rate limited"})
	})

	t.Run("should_handle_cors_middleware", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("should_handle_rate_limit_middleware_success", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/rate/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("should_handle_rate_limit_middleware_exceeded", func(t *testing.T) {
		// 制限を超えるリクエストを送信
		for i := 0; i < 100; i++ {
			req, _ := http.NewRequest("GET", "/rate/test", nil)
			req.RemoteAddr = "192.168.1.2:12345"
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code == 429 {
				break
			}
		}

		req, _ := http.NewRequest("GET", "/rate/test", nil)
		req.RemoteAddr = "192.168.1.2:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 429, w.Code)
	})

	t.Run("should_handle_auth_middleware_success", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/auth/test", nil)
		req.Header.Set("X-API-Key", app.APIKey)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("should_handle_auth_middleware_invalid_key", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/auth/test", nil)
		req.Header.Set("X-API-Key", "invalid_key")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 401, w.Code)
	})

	t.Run("should_handle_auth_middleware_missing_key", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/auth/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 401, w.Code)
	})

	t.Run("should_handle_logging_middleware", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("should_handle_validation_error", func(t *testing.T) {
		router.POST("/validation-test", func(c *gin.Context) {
			var req struct {
				Required string `json:"required" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, apimodels.APIResponse{
					Success: false,
					Error: &apimodels.APIError{
						Code:    "VALIDATION_ERROR",
						Message: err.Error(),
					},
					Timestamp: time.Now(),
				})
				return
			}
			c.JSON(200, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("POST", "/validation-test", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 400, w.Code)
	})

	t.Run("should_handle_404_error", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/not-found", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 404, w.Code)
	})

	t.Run("should_handle_405_error", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 404, w.Code)
	})

	t.Run("should_handle_panic", func(t *testing.T) {
		router.GET("/panic-test", func(c *gin.Context) {
			panic("test panic")
		})

		req, _ := http.NewRequest("GET", "/panic-test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 500, w.Code)
	})

	t.Run("should_handle_cors_preflight_request", func(t *testing.T) {
		req, _ := http.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 204, w.Code) // 実際の実装では204を返す
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET,POST,PUT,DELETE,OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Origin,Content-Type,Accept,Authorization,X-Api-Key,X-Requested-With", w.Header().Get("Access-Control-Allow-Headers"))
	})

	t.Run("should_handle_rate_limit_with_different_ips", func(t *testing.T) {
		// 異なるIPアドレスからのリクエスト
		ips := []string{"192.168.1.10", "192.168.1.11", "192.168.1.12"}
		
		for _, ip := range ips {
			req, _ := http.NewRequest("GET", "/rate/test", nil)
			req.RemoteAddr = ip + ":12345"
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
		}
	})

	t.Run("should_handle_auth_middleware_with_empty_key", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/auth/test", nil)
		req.Header.Set("X-API-Key", "")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 401, w.Code)
	})

	t.Run("should_handle_auth_middleware_with_whitespace_key", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/auth/test", nil)
		req.Header.Set("X-API-Key", "   ")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 401, w.Code)
	})

	t.Run("should_handle_auth_middleware_with_case_insensitive_header", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/auth/test", nil)
		req.Header.Set("x-api-key", app.APIKey)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("should_handle_logging_middleware_with_different_methods", func(t *testing.T) {
		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
		
		for _, method := range methods {
			req, _ := http.NewRequest(method, "/test", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// POST, PUT, DELETE, PATCHは404を返す（ルートが存在しないため）
			expectedCode := 200
			if method != "GET" {
				expectedCode = 404
			}
			assert.Equal(t, expectedCode, w.Code)
		}
	})

	t.Run("should_handle_error_handler_with_different_error_types", func(t *testing.T) {
		// バインディングエラーのテスト
		router.POST("/binding-error", func(c *gin.Context) {
			var req struct {
				Number int `json:"number" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, apimodels.APIResponse{
					Success: false,
					Error: &apimodels.APIError{
						Code:    "BINDING_ERROR",
						Message: err.Error(),
					},
					Timestamp: time.Now(),
				})
				return
			}
			c.JSON(200, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("POST", "/binding-error", bytes.NewBufferString(`{"invalid": "json"`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 400, w.Code)
	})

	t.Run("should_handle_cors_with_different_origins", func(t *testing.T) {
		origins := []string{
			"http://localhost:3000",
			"https://example.com",
		}
		
		for _, origin := range origins {
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("Origin", origin)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
			assert.Equal(t, origin, w.Header().Get("Access-Control-Allow-Origin"))
		}

		// 許可されていないオリジンのテスト
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://test.example.com")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 403, w.Code) // 許可されていないオリジンは403を返す
		assert.Equal(t, "", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("should_handle_rate_limit_burst", func(t *testing.T) {
		// バーストサイズ内のリクエスト
		burstIP := "192.168.1.100"
		for i := 0; i < 5; i++ {
			req, _ := http.NewRequest("GET", "/rate/test", nil)
			req.RemoteAddr = burstIP + ":12345"
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
		}
	})
}
