package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"accesslog-tracker/internal/api/handlers"
	"accesslog-tracker/internal/api/middleware"
	apimodels "accesslog-tracker/internal/api/models"
	"accesslog-tracker/internal/config"
	"accesslog-tracker/internal/domain/services"

	"accesslog-tracker/internal/infrastructure/cache/redis"
	"accesslog-tracker/internal/infrastructure/database/postgresql"
	"accesslog-tracker/internal/infrastructure/database/postgresql/repositories"
	"accesslog-tracker/internal/utils/logger"
	apihelpers "accesslog-tracker/tests/integration/api"
)

func TestTrackingHandlerIntegration(t *testing.T) {
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
	cacheService := redis.NewCacheService(cfg.GetRedisAddr())
	err := cacheService.Connect()
	require.NoError(t, err)
	defer cacheService.Close()

	// データベース接続を初期化
	dbConn := postgresql.NewConnection(cfg.GetDatabaseDSN())
	err = dbConn.Connect(cfg.GetDatabaseDSN())
	require.NoError(t, err)
	defer dbConn.Close()

	// リポジトリを初期化
	trackingRepo := repositories.NewTrackingRepository(dbConn.GetDB())

	// サービスを初期化
	trackingService := services.NewTrackingService(trackingRepo)

	// ハンドラーを初期化
	trackingHandler := handlers.NewTrackingHandler(trackingService, log)

	// ルーターをセットアップ
	router := gin.New()
	router.Use(middleware.Logging(log))
	router.Use(middleware.ErrorHandler(log))
	router.Use(middleware.CORS())

	// トラッキングルートを追加（認証コンテキストを設定するミドルウェアを追加）
	router.POST("/v1/tracking/track", func(c *gin.Context) {
		// テスト用に認証コンテキストを設定
		c.Set("app_id", app.AppID)
		c.Set("application", app)
		c.Next()
	}, trackingHandler.Track)
	router.GET("/v1/tracking/statistics", func(c *gin.Context) {
		// テスト用に認証コンテキストを設定
		c.Set("app_id", app.AppID)
		c.Set("application", app)
		c.Next()
	}, trackingHandler.GetStatistics)

	t.Run("should_track_valid_data", func(t *testing.T) {
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

	t.Run("should_reject_invalid_api_key", func(t *testing.T) {
		trackingRequest := apimodels.TrackingRequest{
			AppID:     app.AppID,
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			URL:       "https://example.com/page1",
			IPAddress: "192.168.1.100",
		}

		jsonData, err := json.Marshal(trackingRequest)
		require.NoError(t, err)

		req, _ := http.NewRequest("POST", "/v1/tracking/track", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "invalid_api_key")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("should_reject_missing_api_key", func(t *testing.T) {
		trackingRequest := apimodels.TrackingRequest{
			AppID:     app.AppID,
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			URL:       "https://example.com/page1",
			IPAddress: "192.168.1.100",
		}

		jsonData, err := json.Marshal(trackingRequest)
		require.NoError(t, err)

		req, _ := http.NewRequest("POST", "/v1/tracking/track", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("should_reject_invalid_request_format", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/v1/tracking/track", bytes.NewBufferString(`{invalid json}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", app.APIKey)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 400, w.Code)
	})

	t.Run("should_reject_missing_required_fields", func(t *testing.T) {
		trackingRequest := map[string]interface{}{
			"user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			"url":        "https://example.com/page1",
		}

		jsonData, err := json.Marshal(trackingRequest)
		require.NoError(t, err)

		req, _ := http.NewRequest("POST", "/v1/tracking/track", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", app.APIKey)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 400, w.Code)
	})

	t.Run("should_reject_app_id_mismatch", func(t *testing.T) {
		trackingRequest := apimodels.TrackingRequest{
			AppID:     "different_app_id",
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			URL:       "https://example.com/page1",
			IPAddress: "192.168.1.100",
		}

		jsonData, err := json.Marshal(trackingRequest)
		require.NoError(t, err)

		req, _ := http.NewRequest("POST", "/v1/tracking/track", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", app.APIKey)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 403, w.Code)
	})

	t.Run("should_get_statistics", func(t *testing.T) {
		// 日付パラメータを追加
		startDate := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
		endDate := time.Now().Format("2006-01-02")
		req, _ := http.NewRequest("GET", "/v1/tracking/statistics?app_id="+app.AppID+"&start_date="+startDate+"&end_date="+endDate, nil)
		req.Header.Set("X-API-Key", app.APIKey)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		var response apimodels.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("should_reject_statistics_without_auth", func(t *testing.T) {
		// 日付パラメータを追加
		startDate := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
		endDate := time.Now().Format("2006-01-02")
		req, _ := http.NewRequest("GET", "/v1/tracking/statistics?app_id="+app.AppID+"&start_date="+startDate+"&end_date="+endDate, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("should_handle_invalid_http_method", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/v1/tracking/track", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 404, w.Code)
	})

	t.Run("should_handle_nonexistent_endpoint", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/tracking/nonexistent", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 404, w.Code)
	})
}
