package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"accesslog-tracker/internal/api/handlers"
	"accesslog-tracker/internal/api/middleware"
	"accesslog-tracker/internal/api/routes"
	"accesslog-tracker/internal/config"
	"accesslog-tracker/internal/domain/models"
	"accesslog-tracker/internal/domain/services"
	"accesslog-tracker/internal/domain/validators"
	"accesslog-tracker/internal/infrastructure/cache/redis"
	"accesslog-tracker/internal/infrastructure/database/postgresql"
	"accesslog-tracker/internal/infrastructure/database/postgresql/repositories"
	"accesslog-tracker/internal/utils/logger"
)

func setupTestServer(t *testing.T) (*gin.Engine, *sql.DB, *redis.Client) {
	// テスト用データベース接続
	db := SetupTestDatabase(t)
	
	// テスト用Redis接続
	redisClient := SetupTestRedis(t)
	
	// Redisに接続
	err := redisClient.Ping(context.Background()).Err()
	require.NoError(t, err)
	
	// リポジトリの初期化
	trackingRepo := repositories.NewTrackingRepository(db)
	applicationRepo := repositories.NewApplicationRepository(db)
	
	// キャッシュサービスの初期化
	cacheService := redis.NewCacheService("redis:6379")
	err = cacheService.Connect()
	require.NoError(t, err)
	
	// サービスの初期化
	trackingService := services.NewTrackingService(trackingRepo)
	applicationService := services.NewApplicationService(applicationRepo, cacheService)
	
	// ロガーの初期化
	log := logger.NewLogger()
	
	// ハンドラーの初期化
	trackingHandler := handlers.NewTrackingHandler(trackingService, log)
	applicationHandler := handlers.NewApplicationHandler(applicationService, log)
	
	// ルーターの設定
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.Logging(log))
	
	// 認証ミドルウェアの初期化
	authMiddleware := middleware.NewAuthMiddleware(applicationService, log)
	
	// ルートの設定
	v1 := router.Group("/v1")
	{
		v1.POST("/track", authMiddleware.Authenticate(), trackingHandler.Track)
		v1.GET("/statistics", authMiddleware.Authenticate(), trackingHandler.GetStatistics)
		v1.POST("/applications", applicationHandler.Create)
	}
	
	return router, db, redisClient
}

func TestTrackingAPI_Integration(t *testing.T) {
	router, db, redisClient := setupTestServer(t)
	defer db.Close()
	defer redisClient.Close()
	
	// テスト用アプリケーションを作成
	app := CreateTestApplication(t, db)
	defer CleanupTestData(t, db)
	
	t.Run("POST /v1/track - should accept valid tracking data", func(t *testing.T) {
		trackingData := models.TrackingRequest{
			AppID:     app.AppID,
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
			URL:       "https://example.com",
			IPAddress: "192.168.1.100",
			SessionID: "alt_1234567890_abc123",
		}
		
		jsonData, _ := json.Marshal(trackingData)
		req := httptest.NewRequest("POST", "/v1/track", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", app.APIKey)
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, true, response.Success)
		
		if response.Data != nil {
			data := response.Data.(map[string]interface{})
			assert.NotNil(t, data["tracking_id"])
			assert.Equal(t, app.AppID, data["app_id"])
			assert.Equal(t, trackingData.SessionID, data["session_id"])
		}
	})
	
	t.Run("POST /v1/track - should reject invalid API key", func(t *testing.T) {
		trackingData := models.TrackingRequest{
			AppID:     app.AppID,
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
			URL:       "https://example.com",
			IPAddress: "192.168.1.100",
		}
		
		jsonData, _ := json.Marshal(trackingData)
		req := httptest.NewRequest("POST", "/v1/track", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "invalid_key")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		
		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, false, response.Success)
		assert.Equal(t, "AUTHENTICATION_ERROR", response.Error.Code)
	})
	
	t.Run("POST /v1/track - should reject missing API key", func(t *testing.T) {
		trackingData := models.TrackingRequest{
			AppID:     app.AppID,
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
			URL:       "https://example.com",
			IPAddress: "192.168.1.100",
		}
		
		jsonData, _ := json.Marshal(trackingData)
		req := httptest.NewRequest("POST", "/v1/track", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		// X-API-Keyヘッダーを設定しない
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		
		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, false, response.Success)
		assert.Equal(t, "AUTHENTICATION_ERROR", response.Error.Code)
	})
	
	t.Run("POST /v1/track - should reject invalid request format", func(t *testing.T) {
		invalidJSON := `{"app_id": "test", "invalid_field": "value"`
		
		req := httptest.NewRequest("POST", "/v1/track", bytes.NewBufferString(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", app.APIKey)
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, false, response.Success)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})
	
	t.Run("GET /v1/statistics - should return statistics for valid app_id", func(t *testing.T) {
		// テストデータを作成
		CreateTestTrackingData(t, db, app.AppID)
		CreateTestTrackingData(t, db, app.AppID)
		
		startDate := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
		endDate := time.Now().Format("2006-01-02")
		
		req := httptest.NewRequest("GET", fmt.Sprintf("/v1/statistics?app_id=%s&start_date=%s&end_date=%s", app.AppID, startDate, endDate), nil)
		req.Header.Set("X-API-Key", app.APIKey)
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, true, response.Success)
		
		if response.Data != nil {
			data := response.Data.(map[string]interface{})
			assert.NotNil(t, data["total_requests"])
			assert.NotNil(t, data["unique_visitors"])
		}
	})
	
	t.Run("GET /v1/statistics - should reject missing parameters", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/statistics", nil)
		req.Header.Set("X-API-Key", app.APIKey)
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, false, response.Success)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})
	
	t.Run("GET /v1/statistics - should reject invalid date format", func(t *testing.T) {
		req := httptest.NewRequest("GET", fmt.Sprintf("/v1/statistics?app_id=%s&start_date=invalid-date&end_date=2024-01-31", app.AppID), nil)
		req.Header.Set("X-API-Key", app.APIKey)
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, false, response.Success)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})
}

func TestTrackingHandlerIntegration(t *testing.T) {
	// テスト用データベースをセットアップ
	db := SetupTestDatabase(t)
	defer db.Close()

	// テスト用Redisをセットアップ
	redisClient := SetupTestRedis(t)
	defer redisClient.Close()

	// 設定を読み込み
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "postgres",
			Port:     5432,
			User:     "postgres",
			Password: "password",
			DBName:   "access_log_tracker_test",
			SSLMode:  "disable",
		},
		Redis: config.RedisConfig{
			Host:     "redis",
			Port:     6379,
			Password: "",
			DB:       0,
		},
	}

	// ロガーを初期化
	log := logger.NewLogger()
	log.SetLevel("debug")

	// データベース接続を初期化
	dbConn := postgresql.NewConnection(cfg)
	err := dbConn.Connect()
	require.NoError(t, err)
	defer dbConn.Close()

	// Redisキャッシュサービスを初期化
	cacheService := redis.NewCacheService(cfg)
	err = cacheService.Connect()
	require.NoError(t, err)
	defer cacheService.Close()

	// リポジトリを初期化
	appRepo := repositories.NewApplicationRepository(dbConn)
	trackingRepo := repositories.NewTrackingRepository(dbConn)

	// バリデーターを初期化
	appValidator := validators.NewApplicationValidator()
	trackingValidator := validators.NewTrackingValidator()

	// サービスを初期化
	appService := services.NewApplicationService(appRepo, cacheService, appValidator, log)
	trackingService := services.NewTrackingService(trackingRepo, appService, trackingValidator, log)

	// ハンドラーを初期化
	trackingHandler := handlers.NewTrackingHandler(trackingService, log)

	// ルーターをセットアップ
	router := gin.New()
	router.Use(middleware.CORS())
	router.Use(middleware.Logging(log))
	router.Use(middleware.ErrorHandler(log))

	// テスト用ルートを設定
	routes.SetupTest(router, trackingHandler)

	// テストケース1: 正常なトラッキングデータの送信
	t.Run("Track valid data", func(t *testing.T) {
		trackingData := models.TrackingData{
			AppID:     1,
			URL:       "https://example.com/page1",
			Referrer:  "https://google.com",
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			IPAddress: "192.168.1.1",
			Timestamp: time.Now(),
			CustomParams: map[string]string{
				"utm_source": "google",
				"utm_medium": "cpc",
			},
		}

		jsonData, err := json.Marshal(trackingData)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/v1/track", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", "192.168.1.1")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "success", response["status"])
	})

	// テストケース2: 無効なトラッキングデータの送信
	t.Run("Track invalid data", func(t *testing.T) {
		trackingData := models.TrackingData{
			AppID:     0, // 無効なAppID
			URL:       "",
			UserAgent: "",
			IPAddress: "invalid-ip",
			Timestamp: time.Now(),
		}

		jsonData, err := json.Marshal(trackingData)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/v1/track", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// テストケース3: 統計データの取得
	t.Run("Get statistics", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/statistics?app_id=1&start_date=2024-01-01&end_date=2024-12-31", nil)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "success", response["status"])
	})
}
