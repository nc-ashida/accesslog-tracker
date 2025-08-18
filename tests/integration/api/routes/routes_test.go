package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"accesslog-tracker/internal/api/routes"
	apimodels "accesslog-tracker/internal/api/models"
	"accesslog-tracker/internal/config"
	"accesslog-tracker/internal/domain/services"
	redisCache "accesslog-tracker/internal/infrastructure/cache/redis"
	"accesslog-tracker/internal/infrastructure/database/postgresql"
	"accesslog-tracker/internal/infrastructure/database/postgresql/repositories"
	"accesslog-tracker/internal/utils/logger"
	apihelpers "accesslog-tracker/tests/integration/api"
)

func TestRoutesIntegration(t *testing.T) {
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

	// データベース接続を初期化
	dbConn := postgresql.NewConnection(cfg.GetDatabaseDSN())
	err = dbConn.Connect(cfg.GetDatabaseDSN())
	require.NoError(t, err)
	defer dbConn.Close()

	// リポジトリを初期化
	appRepo := repositories.NewApplicationRepository(db)
	trackingRepo := repositories.NewTrackingRepository(db)

	// サービスを初期化
	appService := services.NewApplicationService(appRepo, cacheService)
	trackingService := services.NewTrackingService(trackingRepo)

	// ルーターをセットアップ
	router := gin.New()
	routes.Setup(router, trackingService, appService, dbConn, cacheService, log)

	t.Run("should_handle_health_check", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		var response apimodels.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("should_handle_readiness_check", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/ready", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		var response apimodels.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("should_handle_liveness_check", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/live", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		var response apimodels.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("should_handle_application_creation", func(t *testing.T) {
		createRequest := map[string]interface{}{
			"name":        "Test Application",
			"description": "Test application for integration testing",
			"domain":      "test.example.com",
		}

		jsonData, err := json.Marshal(createRequest)
		require.NoError(t, err)

		req, _ := http.NewRequest("POST", "/v1/applications", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 201, w.Code)

		var response apimodels.APIResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("should_handle_application_listing", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/applications", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		var response apimodels.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("should_handle_tracking_with_authentication", func(t *testing.T) {
		trackingRequest := apimodels.TrackingRequest{
			AppID:       app.AppID,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			URL:         "https://example.com/page1",
			IPAddress:   "192.168.1.100",
			SessionID:   "test_session_123",
			Referrer:    "https://google.com",
			CustomParams: map[string]interface{}{
				"page_type": "product",
				"user_id":   "12345",
			},
		}

		jsonData, err := json.Marshal(trackingRequest)
		require.NoError(t, err)

		req, _ := http.NewRequest("POST", "/v1/tracking/track", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", app.APIKey)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		var response apimodels.APIResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("should_handle_tracking_without_authentication", func(t *testing.T) {
		trackingRequest := apimodels.TrackingRequest{
			AppID:       app.AppID,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			URL:         "https://example.com/page1",
			IPAddress:   "192.168.1.100",
			SessionID:   "test_session_123",
			Referrer:    "https://google.com",
			CustomParams: map[string]interface{}{
				"page_type": "product",
				"user_id":   "12345",
			},
		}

		jsonData, err := json.Marshal(trackingRequest)
		require.NoError(t, err)

		req, _ := http.NewRequest("POST", "/v1/tracking/track", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 401, w.Code)
	})

	t.Run("should_handle_beacon_generation", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/beacon/generate?app_id="+app.AppID, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "image/gif", w.Header().Get("Content-Type"))
	})

	t.Run("should_handle_beacon_serving", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/tracker.js", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "application/javascript", w.Header().Get("Content-Type"))
	})

	t.Run("should_handle_custom_beacon_serving", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/tracker/123.js", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 400, w.Code)
		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
	})

	t.Run("should_handle_404_error", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/nonexistent", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 404, w.Code)

		var response apimodels.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "NOT_FOUND", response.Error.Code)
	})

	t.Run("should_handle_method_not_allowed", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 404, w.Code)

		var response apimodels.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "NOT_FOUND", response.Error.Code)
	})
}

func TestRoutesSetupTest(t *testing.T) {
	// テスト用データベースをセットアップ
	db := apihelpers.SetupTestDatabase(t)
	defer db.Close()

	// テスト用Redisをセットアップ
	redisClient := apihelpers.SetupTestRedis(t)
	defer redisClient.Close()

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

	// データベース接続を初期化
	dbConn := postgresql.NewConnection(cfg.GetDatabaseDSN())
	err = dbConn.Connect(cfg.GetDatabaseDSN())
	require.NoError(t, err)
	defer dbConn.Close()

	// リポジトリを初期化
	appRepo := repositories.NewApplicationRepository(db)
	trackingRepo := repositories.NewTrackingRepository(db)

	// サービスを初期化
	appService := services.NewApplicationService(appRepo, cacheService)
	trackingService := services.NewTrackingService(trackingRepo)

	// テスト用ルーターをセットアップ
	router := gin.New()
	routes.SetupTest(router, trackingService, appService, dbConn, cacheService, log)

	t.Run("should_handle_test_health_check", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		var response apimodels.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})
}
