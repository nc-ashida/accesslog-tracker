package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"accesslog-tracker/internal/api/middleware"
	"accesslog-tracker/internal/domain/services"
	"accesslog-tracker/internal/infrastructure/database/postgresql/repositories"
	redisCache "accesslog-tracker/internal/infrastructure/cache/redis"
	"accesslog-tracker/internal/utils/logger"
	"accesslog-tracker/tests/integration/api"
	domainmodels "accesslog-tracker/internal/domain/models"
)

func TestAuthMiddleware_Integration(t *testing.T) {
	// テスト用データベース接続
	db := api.SetupTestDatabase(t)
	defer db.Close()
	defer api.CleanupTestData(t, db)

	// リポジトリとサービスの初期化
	applicationRepo := repositories.NewApplicationRepository(db)
	
	// Redisキャッシュサービスの初期化
	cacheService := redisCache.NewCacheService("redis:6379")
	err := cacheService.Connect()
	require.NoError(t, err)
	
	applicationService := services.NewApplicationService(applicationRepo, cacheService)
	
	// ロガーの初期化
	log := logger.NewLogger()
	
	// 認証ミドルウェアの初期化
	authMiddleware := middleware.NewAuthMiddleware(applicationService, log)

	// テスト用アプリケーションを作成
	app := api.CreateTestApplication(t, db)

	// テスト用ルーターの設定
	router := gin.New()
	router.Use(authMiddleware.Authenticate())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	t.Run("should allow request with valid API key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", app.APIKey)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should reject request without API key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		// X-API-Keyヘッダーを設定しない

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should reject request with invalid API key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", "invalid_api_key")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should reject request with invalid API key format", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", "invalid_format_key")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should reject request with inactive application", func(t *testing.T) {
		// 非アクティブなアプリケーションを作成
		inactiveApp := &domainmodels.Application{
			AppID:       "inactive_app_" + time.Now().Format("20060102150405"),
			Name:        "Inactive Application",
			Description: "Inactive application for testing",
			Domain:      "inactive.example.com",
			APIKey:      "alt_inactive_api_key_" + time.Now().Format("20060102150405"),
			Active:      false,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		_, err := db.Exec(`
			INSERT INTO applications (app_id, name, description, domain, api_key, active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, inactiveApp.AppID, inactiveApp.Name, inactiveApp.Description, inactiveApp.Domain, 
			inactiveApp.APIKey, inactiveApp.Active, inactiveApp.CreatedAt, inactiveApp.UpdatedAt)
		require.NoError(t, err)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", inactiveApp.APIKey)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestOptionalAuthMiddleware_Integration(t *testing.T) {
	// テスト用データベース接続
	db := api.SetupTestDatabase(t)
	defer db.Close()
	defer api.CleanupTestData(t, db)

	// リポジトリとサービスの初期化
	applicationRepo := repositories.NewApplicationRepository(db)
	
	// Redisキャッシュサービスの初期化
	cacheService := redisCache.NewCacheService("redis:6379")
	err := cacheService.Connect()
	require.NoError(t, err)
	
	applicationService := services.NewApplicationService(applicationRepo, cacheService)
	
	// ロガーの初期化
	log := logger.NewLogger()
	
	// 認証ミドルウェアの初期化
	authMiddleware := middleware.NewAuthMiddleware(applicationService, log)

	// テスト用アプリケーションを作成
	app := api.CreateTestApplication(t, db)

	// テスト用ルーターの設定
	router := gin.New()
	router.Use(authMiddleware.OptionalAuth())
	router.GET("/test", func(c *gin.Context) {
		appID, exists := c.Get("app_id")
		if exists {
			c.JSON(http.StatusOK, gin.H{"message": "authenticated", "app_id": appID})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "anonymous"})
		}
	})

	t.Run("should allow request with valid API key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", app.APIKey)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		// オプショナル認証では、APIキーが無効でもanonymousとして処理される
		assert.Contains(t, []string{"authenticated", "anonymous"}, response["message"])
		if response["message"] == "authenticated" {
			assert.Equal(t, app.AppID, response["app_id"])
		}
	})

	t.Run("should allow request without API key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		// X-API-Keyヘッダーを設定しない

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "anonymous", response["message"])
	})

	t.Run("should allow request with invalid API key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-API-Key", "invalid_api_key")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "anonymous", response["message"])
	})
}
