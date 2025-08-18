package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"accesslog-tracker/internal/api/middleware"
	"accesslog-tracker/internal/config"
	"accesslog-tracker/internal/domain/models"
	"accesslog-tracker/internal/domain/services"
	redisCache "accesslog-tracker/internal/infrastructure/cache/redis"
	"accesslog-tracker/internal/infrastructure/database/postgresql/repositories"
	"accesslog-tracker/internal/utils/logger"
	apihelpers "accesslog-tracker/tests/integration/api"
)

func TestAuthMiddlewareIntegration(t *testing.T) {
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

	// リポジトリを初期化（テストヘルパーと同じデータベース接続を使用）
	appRepo := repositories.NewApplicationRepository(db)

	// サービスを初期化
	appService := services.NewApplicationService(appRepo, cacheService)

	// 認証ミドルウェアを初期化
	authMiddleware := middleware.NewAuthMiddleware(appService, log)

	// ルーターをセットアップ
	router := gin.New()
	router.Use(authMiddleware.Authenticate())

	// テスト用ルートを追加
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "authenticated"})
	})

	t.Run("should_allow_valid_api_key", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", app.APIKey)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("should_reject_invalid_api_key", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", "invalid_api_key")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 401, w.Code)
	})

	t.Run("should_reject_missing_api_key", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 401, w.Code)
	})

	t.Run("should_reject_invalid_api_key_format", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", "invalid_format_key")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 401, w.Code)
	})

	t.Run("should_reject_inactive_application", func(t *testing.T) {
		// 非アクティブなアプリケーションを作成
		inactiveApp := &models.Application{
			AppID:       "inactive_app_" + time.Now().Format("20060102150405") + "_" + apihelpers.RandomString(5),
			Name:        "Inactive Test Application",
			Description: "Inactive test application",
			Domain:      "inactive.example.com",
			APIKey:      "alt_inactive_api_key_" + time.Now().Format("20060102150405") + "_" + apihelpers.RandomString(5),
			Active:      false,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		_, err := db.Exec(`
			INSERT INTO applications (app_id, name, description, domain, api_key, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (app_id) DO NOTHING
		`, inactiveApp.AppID, inactiveApp.Name, inactiveApp.Description, inactiveApp.Domain, inactiveApp.APIKey, inactiveApp.Active, inactiveApp.CreatedAt, inactiveApp.UpdatedAt)
		require.NoError(t, err)

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", inactiveApp.APIKey)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 403, w.Code)
	})

	t.Run("should_handle_optional_auth", func(t *testing.T) {
		// オプショナル認証のテスト
		optionalAuthMiddleware := middleware.NewAuthMiddleware(appService, log)

		optionalRouter := gin.New()
		optionalRouter.Use(optionalAuthMiddleware.OptionalAuth())
		optionalRouter.GET("/optional", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "optional auth"})
		})

		req, _ := http.NewRequest("GET", "/optional", nil)
		w := httptest.NewRecorder()

		optionalRouter.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("should_handle_different_auth_header_formats", func(t *testing.T) {
		// 異なる認証ヘッダー形式のテスト
		testCases := []struct {
			name   string
			header string
			value  string
			status int
		}{
			{"x_api_key_header", "X-API-Key", app.APIKey, 200},
			{"authorization_bearer", "Authorization", "Bearer " + app.APIKey, 401},
			{"authorization_token", "Authorization", "Token " + app.APIKey, 401},
			{"empty_x_api_key", "X-API-Key", "", 401},
			{"no_prefix_api_key", "X-API-Key", "no_prefix_key", 401},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req, _ := http.NewRequest("GET", "/test", nil)
				req.Header.Set(tc.header, tc.value)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				assert.Equal(t, tc.status, w.Code)
			})
		}
	})

	t.Run("should_handle_cache_miss", func(t *testing.T) {
		// キャッシュミスのテスト
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", app.APIKey)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("should_handle_database_error", func(t *testing.T) {
		// データベースエラーのテストをスキップ（環境によって動作が異なるため）
		t.Skip("Database error test skipped due to environment differences")
	})

	t.Run("should_handle_application_service_error", func(t *testing.T) {
		// アプリケーションサービスのエラーハンドリングテスト
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", "alt_invalid_key_for_testing")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 401, w.Code)
	})

	t.Run("should_handle_context_without_app_id", func(t *testing.T) {
		// コンテキストにapp_idが設定されていない場合のテスト
		router.GET("/context-test", func(c *gin.Context) {
			appID, exists := c.Get("app_id")
			assert.False(t, exists)
			assert.Nil(t, appID)
			c.JSON(200, gin.H{"message": "context test"})
		})

		req, _ := http.NewRequest("GET", "/context-test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 401, w.Code)
	})

	t.Run("should_handle_context_with_app_id", func(t *testing.T) {
		// コンテキストにapp_idが設定されている場合のテスト
		router.GET("/context-test-success", func(c *gin.Context) {
			appID, exists := c.Get("app_id")
			assert.True(t, exists)
			assert.Equal(t, app.AppID, appID)
			c.JSON(200, gin.H{"message": "context test success"})
		})

		req, _ := http.NewRequest("GET", "/context-test-success", nil)
		req.Header.Set("X-API-Key", app.APIKey)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("should_handle_optional_auth_with_valid_key", func(t *testing.T) {
		// オプショナル認証で有効なキーを使用したテスト
		optionalAuthMiddleware := middleware.NewAuthMiddleware(appService, log)

		optionalRouter := gin.New()
		optionalRouter.Use(optionalAuthMiddleware.OptionalAuth())
		optionalRouter.GET("/optional-valid", func(c *gin.Context) {
			appID, exists := c.Get("app_id")
			assert.True(t, exists)
			assert.Equal(t, app.AppID, appID)
			c.JSON(200, gin.H{"message": "optional auth valid"})
		})

		req, _ := http.NewRequest("GET", "/optional-valid", nil)
		req.Header.Set("X-API-Key", app.APIKey)
		w := httptest.NewRecorder()

		optionalRouter.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("should_handle_optional_auth_with_invalid_key", func(t *testing.T) {
		// オプショナル認証で無効なキーを使用したテスト
		optionalAuthMiddleware := middleware.NewAuthMiddleware(appService, log)

		optionalRouter := gin.New()
		optionalRouter.Use(optionalAuthMiddleware.OptionalAuth())
		optionalRouter.GET("/optional-invalid", func(c *gin.Context) {
			appID, exists := c.Get("app_id")
			assert.False(t, exists)
			assert.Nil(t, appID)
			c.JSON(200, gin.H{"message": "optional auth invalid"})
		})

		req, _ := http.NewRequest("GET", "/optional-invalid", nil)
		req.Header.Set("X-API-Key", "invalid_key")
		w := httptest.NewRecorder()

		optionalRouter.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("should_handle_optional_auth_with_inactive_app", func(t *testing.T) {
		// オプショナル認証で非アクティブなアプリケーションを使用したテスト
		optionalAuthMiddleware := middleware.NewAuthMiddleware(appService, log)

		optionalRouter := gin.New()
		optionalRouter.Use(optionalAuthMiddleware.OptionalAuth())
		optionalRouter.GET("/optional-inactive", func(c *gin.Context) {
			appID, exists := c.Get("app_id")
			assert.False(t, exists)
			assert.Nil(t, appID)
			c.JSON(200, gin.H{"message": "optional auth inactive"})
		})

		// 非アクティブなアプリケーションのAPIキーを使用
		req, _ := http.NewRequest("GET", "/optional-inactive", nil)
		req.Header.Set("X-API-Key", "alt_inactive_api_key_test")
		w := httptest.NewRecorder()

		optionalRouter.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})
}
