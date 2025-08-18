package handlers

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

	"accesslog-tracker/internal/api/handlers"
	"accesslog-tracker/internal/api/middleware"
	apimodels "accesslog-tracker/internal/api/models"
	"accesslog-tracker/internal/config"
	redisCache "accesslog-tracker/internal/infrastructure/cache/redis"
	"accesslog-tracker/internal/infrastructure/database/postgresql"
	"accesslog-tracker/internal/utils/logger"
	apihelpers "accesslog-tracker/tests/integration/api"
)

func TestBeaconHandlerIntegration(t *testing.T) {
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

	// ハンドラーを初期化
	beaconHandler := handlers.NewBeaconHandler()

	// ルーターをセットアップ
	router := gin.New()
	router.Use(middleware.Logging(log))
	router.Use(middleware.ErrorHandler(log))
	router.Use(middleware.CORS())

	// ビーコンルートを追加（順序が重要）
	router.GET("/tracker/:app_id", beaconHandler.ServeCustom) // パラメータのみ
	router.GET("/tracker.js", beaconHandler.Serve)
	router.GET("/tracker.min.js", beaconHandler.ServeMinified)
	router.GET("/beacon", beaconHandler.ProcessBeacon)
	router.GET("/beacon.gif", beaconHandler.ServeGIF)
	router.POST("/beacon/config", beaconHandler.GenerateBeaconWithConfig)
	router.GET("/beacon/health", beaconHandler.Health)

	t.Run("should_generate_beacon_successfully", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/beacon?app_id="+app.AppID, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "image/gif", w.Header().Get("Content-Type"))
	})

	t.Run("should_generate_beacon_with_session_id", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/beacon?app_id="+app.AppID+"&session_id=test_session_123", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		req.RemoteAddr = "192.168.1.2:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "image/gif", w.Header().Get("Content-Type"))
	})

	t.Run("should_generate_beacon_with_referrer", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/beacon?app_id="+app.AppID+"&referrer=https://example.com", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		req.RemoteAddr = "192.168.1.3:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "image/gif", w.Header().Get("Content-Type"))
	})

	t.Run("should_generate_beacon_with_custom_parameters", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/beacon?app_id="+app.AppID+"&custom_param1=value1&custom_param2=value2", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		req.RemoteAddr = "192.168.1.4:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "image/gif", w.Header().Get("Content-Type"))
	})

	t.Run("should_generate_beacon_with_x_forwarded_for", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/beacon?app_id="+app.AppID, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		req.Header.Set("X-Forwarded-For", "203.0.113.1")
		req.RemoteAddr = "192.168.1.5:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "image/gif", w.Header().Get("Content-Type"))
	})

	t.Run("should_generate_beacon_without_user_agent", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/beacon?app_id="+app.AppID, nil)
		req.RemoteAddr = "192.168.1.6:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "image/gif", w.Header().Get("Content-Type"))
	})

	t.Run("should_generate_beacon_without_remote_addr", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/beacon?app_id="+app.AppID, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "image/gif", w.Header().Get("Content-Type"))
	})

	t.Run("should_handle_missing_app_id", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/beacon", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		req.RemoteAddr = "192.168.1.7:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 400, w.Code)
	})

	t.Run("should_handle_invalid_app_id", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/beacon?app_id=invalid_app_id", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		req.RemoteAddr = "192.168.1.8:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "image/gif", w.Header().Get("Content-Type"))
	})

	t.Run("should_handle_post_request", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/beacon?app_id="+app.AppID, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		req.RemoteAddr = "192.168.1.9:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 404, w.Code)
	})

	t.Run("should_handle_beacon_with_url_parameter", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/beacon?app_id="+app.AppID+"&url=https://example.com/page1", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		req.RemoteAddr = "192.168.1.10:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "image/gif", w.Header().Get("Content-Type"))
	})

	t.Run("should_handle_beacon_with_ip_address_parameter", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/beacon?app_id="+app.AppID+"&ip_address=192.168.1.100", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		req.RemoteAddr = "192.168.1.11:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "image/gif", w.Header().Get("Content-Type"))
	})

	t.Run("should_handle_beacon_with_encoded_parameters", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/beacon?app_id="+app.AppID+"&url=https%3A//example.com/page%3Fparam%3Dvalue", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		req.RemoteAddr = "192.168.1.12:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "image/gif", w.Header().Get("Content-Type"))
	})

	t.Run("should_handle_beacon_with_empty_parameters", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/beacon?app_id="+app.AppID+"&session_id=&referrer=", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		req.RemoteAddr = "192.168.1.13:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "image/gif", w.Header().Get("Content-Type"))
	})

	// ServeMinified のテスト
	t.Run("should_serve_minified_javascript_beacon", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/tracker.min.js", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "application/javascript", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Body.String(), "function track")
		assert.Contains(t, w.Header().Get("Cache-Control"), "max-age=86400")
	})

	// ServeCustom のテスト
	t.Run("should_serve_custom_javascript_beacon", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/tracker/123", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "application/javascript", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Body.String(), "function track")
		assert.Contains(t, w.Body.String(), "123")
	})

	t.Run("should_handle_invalid_app_id_in_custom_beacon", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/tracker/invalid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 400, w.Code)
		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
	})

	// ServeGIF のテスト
	t.Run("should_serve_gif_beacon", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/beacon.gif?app_id="+app.AppID, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		req.RemoteAddr = "192.168.1.14:12345"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "image/gif", w.Header().Get("Content-Type"))
		assert.Equal(t, "no-cache, no-store, must-revalidate", w.Header().Get("Cache-Control"))
	})

	// GenerateBeaconWithConfig のテスト
	t.Run("should_generate_beacon_with_config", func(t *testing.T) {
		config := map[string]interface{}{
			"endpoint": "https://api.example.com/track",
			"debug":    true,
			"version":  "1.0.0",
			"minify":   false,
		}

		jsonData, _ := json.Marshal(config)
		req, _ := http.NewRequest("POST", "/beacon/config", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

		var response apimodels.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("should_handle_invalid_config", func(t *testing.T) {
		config := map[string]interface{}{
			"endpoint": "", // 無効なエンドポイント
			"version":  "1.0.0",
		}

		jsonData, _ := json.Marshal(config)
		req, _ := http.NewRequest("POST", "/beacon/config", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 400, w.Code)
	})

	// Health のテスト
	t.Run("should_return_health_status", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/beacon/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

		var response apimodels.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "healthy", response.Data.(map[string]interface{})["status"])
	})

	// ETag とキャッシュ制御のテスト
	t.Run("should_handle_etag_caching", func(t *testing.T) {
		// 最初のリクエスト
		req, _ := http.NewRequest("GET", "/tracker.js", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		etag := w.Header().Get("ETag")
		assert.NotEmpty(t, etag)

		// 同じETagでリクエスト
		req2, _ := http.NewRequest("GET", "/tracker.js", nil)
		req2.Header.Set("If-None-Match", etag)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		assert.Equal(t, 304, w2.Code) // Not Modified
	})

	// ProcessBeacon のテスト
	t.Run("should_process_beacon_request", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/beacon?app_id="+app.AppID+"&session_id=test-session&url=/test-page&referrer=https://example.com&custom_param=value", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		req.RemoteAddr = "192.168.1.15:12346"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "image/gif", w.Header().Get("Content-Type"))
		assert.Equal(t, "no-cache, no-store, must-revalidate", w.Header().Get("Cache-Control"))
		assert.Equal(t, "no-cache", w.Header().Get("Pragma"))
		assert.Equal(t, "0", w.Header().Get("Expires"))
	})

	t.Run("should_handle_beacon_request_without_app_id", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/beacon?session_id=test-session&url=/test-page", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 400, w.Code)
		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

		var response apimodels.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
		assert.Equal(t, "app_id parameter is required", response.Error.Message)
	})

	t.Run("should_handle_beacon_request_with_default_url", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/beacon?app_id="+app.AppID+"&session_id=test-session", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		req.RemoteAddr = "192.168.1.16:12347"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "image/gif", w.Header().Get("Content-Type"))
	})

	t.Run("should_handle_beacon_request_with_custom_params", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/beacon?app_id="+app.AppID+"&session_id=test-session&url=/test-page&param1=value1&param2=value2&param3=value3", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		req.RemoteAddr = "192.168.1.17:12348"
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "image/gif", w.Header().Get("Content-Type"))
	})
}
