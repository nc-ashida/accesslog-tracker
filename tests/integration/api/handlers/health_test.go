package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"accesslog-tracker/internal/api/handlers"
	"accesslog-tracker/internal/api/middleware"
	apimodels "accesslog-tracker/internal/api/models"
	"accesslog-tracker/internal/config"
	redisCache "accesslog-tracker/internal/infrastructure/cache/redis"
	"accesslog-tracker/internal/infrastructure/database/postgresql"
	"accesslog-tracker/internal/utils/logger"
	apihelpers "accesslog-tracker/tests/integration/api"
)

func TestHealthHandlerIntegration(t *testing.T) {
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

	// ハンドラーを初期化
	healthHandler := handlers.NewHealthHandler(dbConn, cacheService, log)

	// ルーターをセットアップ
	router := gin.New()
	router.Use(middleware.Logging(log))
	router.Use(middleware.ErrorHandler(log))

	// ヘルスチェックルートを追加
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Readiness)
	router.GET("/live", healthHandler.Liveness)

	t.Run("should_return_health_status", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		var response apimodels.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("should_return_readiness_status", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/ready", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		var response apimodels.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("should_return_liveness_status", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/live", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		var response apimodels.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("should_handle_invalid_http_method", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 404, w.Code)
	})

	t.Run("should_handle_nonexistent_endpoint", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/nonexistent", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 404, w.Code)
	})
}
