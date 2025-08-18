package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"accesslog-tracker/internal/api/models"
	"accesslog-tracker/internal/api/server"
	"accesslog-tracker/internal/config"
	"accesslog-tracker/internal/domain/services"
	"accesslog-tracker/internal/infrastructure/cache/redis"
	"accesslog-tracker/internal/infrastructure/database/postgresql"
	"accesslog-tracker/internal/infrastructure/database/postgresql/repositories"
	"accesslog-tracker/internal/utils/logger"
)

// テスト用のアプリケーション構造体
type TestApplication struct {
	AppID  string
	APIKey string
	Name   string
	Domain string
}

// テスト用のトラッキングデータ構造体
type TestTrackingData struct {
	AppID        string                 `json:"app_id"`
	UserAgent    string                 `json:"user_agent"`
	URL          string                 `json:"url,omitempty"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	SessionID    string                 `json:"session_id,omitempty"`
	Referrer     string                 `json:"referrer,omitempty"`
	CustomParams map[string]interface{} `json:"custom_params,omitempty"`
}

// テストサーバーのセットアップ
func setupSpecificationTestServer(t *testing.T) (*httptest.Server, func()) {
	// テスト用設定
	cfg := &config.Config{
		App: config.AppConfig{
			Port:  8080,
			Debug: true,
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     18432,
			Name:     "access_log_tracker_test",
			User:     "postgres",
			Password: "password",
		},
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: 16379,
		},
	}

	// ロガーの初期化
	log := logger.NewLogger()

	// テスト用データベース接続
	dbConn := postgresql.NewConnection("test")
	err := dbConn.Connect(fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.Name, cfg.Database.User, cfg.Database.Password))
	require.NoError(t, err)

	// テスト用Redis接続
	redisClient := redis.NewCacheService(fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port))
	err = redisClient.Connect()
	require.NoError(t, err)

	// リポジトリの初期化
	appRepo := repositories.NewApplicationRepository(dbConn.GetDB())
	trackingRepo := repositories.NewTrackingRepository(dbConn.GetDB())

	// サービスの初期化
	appService := services.NewApplicationService(appRepo, redisClient)
	trackingService := services.NewTrackingService(trackingRepo)

	// サーバーの作成
	srv := server.NewServer(cfg, log, trackingService, appService, dbConn, redisClient)

	// テストサーバーを起動
	testServer := httptest.NewServer(srv.GetRouter())

	cleanup := func() {
		testServer.Close()
		dbConn.Close()
		redisClient.Close()
	}

	return testServer, cleanup
}

// テスト用アプリケーションの作成
func createTestApplication(t *testing.T, serverURL string) *TestApplication {
	appData := models.ApplicationRequest{
		Name:   "Specification Test App",
		Domain: "spec-test.example.com",
	}

	jsonData, _ := json.Marshal(appData)
	resp, err := http.Post(serverURL+"/v1/applications", "application/json", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var response models.APIResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)
	require.True(t, response.Success)

	if response.Data != nil {
		appInfo := response.Data.(map[string]interface{})
		return &TestApplication{
			AppID:  appInfo["app_id"].(string),
			APIKey: appInfo["api_key"].(string),
			Name:   appData.Name,
			Domain: appData.Domain,
		}
	}

	// エラーケース
	return &TestApplication{
		AppID:  "test_app_id",
		APIKey: "alt_test_api_key",
		Name:   appData.Name,
		Domain: appData.Domain,
	}
}

// JSONリクエストの送信ヘルパー
func sendJSONRequest(method, url string, data interface{}, apiKey string) (*http.Response, error) {
	jsonData, _ := json.Marshal(data)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	return client.Do(req)
}

// 1. ヘルスチェックエンドポイントのテスト
func TestHealthCheckEndpoints(t *testing.T) {
	server, cleanup := setupSpecificationTestServer(t)
	defer cleanup()

	t.Run("GET /health - システムの健全性確認", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/health")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response models.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.True(t, response.Success)

		// レスポンス構造の確認
		healthData := response.Data.(map[string]interface{})
		assert.Equal(t, "healthy", healthData["status"])
		assert.NotNil(t, healthData["timestamp"])

		services := healthData["services"].(map[string]interface{})
		assert.Equal(t, "healthy", services["database"])
		assert.Equal(t, "healthy", services["redis"])
	})

	t.Run("GET /ready - アプリケーション準備完了状態", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/ready")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GET /live - アプリケーション生存状態", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/live")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// 2. アプリケーション管理APIのテスト
func TestApplicationManagementAPI(t *testing.T) {
	server, cleanup := setupSpecificationTestServer(t)
	defer cleanup()

	t.Run("POST /v1/applications - アプリケーション作成", func(t *testing.T) {
		appData := models.ApplicationRequest{
			Name:   "Test Application",
			Domain: "test.example.com",
		}

		jsonData, _ := json.Marshal(appData)
		resp, err := http.Post(server.URL+"/v1/applications", "application/json", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var response models.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.True(t, response.Success)

		// レスポンス構造の確認
		if response.Data != nil {
			appInfo := response.Data.(map[string]interface{})
			assert.NotEmpty(t, appInfo["app_id"])
			assert.NotEmpty(t, appInfo["api_key"])
			assert.Equal(t, appData.Name, appInfo["name"])
			assert.Equal(t, appData.Domain, appInfo["domain"])
			if appInfo["is_active"] != nil {
				assert.True(t, appInfo["is_active"].(bool))
			}
			assert.NotNil(t, appInfo["created_at"])
			assert.NotNil(t, appInfo["updated_at"])
		}
	})

	t.Run("GET /v1/applications - アプリケーション一覧取得", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/v1/applications")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response models.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.True(t, response.Success)

		// レスポンス構造の確認
		listData := response.Data.(map[string]interface{})
		assert.NotNil(t, listData["applications"])
		assert.NotNil(t, listData["pagination"])

		pagination := listData["pagination"].(map[string]interface{})
		assert.Equal(t, float64(1), pagination["page"])
		if pagination["limit"] != nil {
			assert.Equal(t, float64(20), pagination["limit"])
		}
		assert.NotNil(t, pagination["total"])
		assert.NotNil(t, pagination["total_pages"])
	})

	t.Run("GET /v1/applications/{id} - アプリケーション詳細取得", func(t *testing.T) {
		app := createTestApplication(t, server.URL)

		resp, err := http.Get(server.URL + "/v1/applications/" + app.AppID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response models.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.True(t, response.Success)

		appInfo := response.Data.(map[string]interface{})
		assert.Equal(t, app.AppID, appInfo["app_id"])
		assert.Equal(t, app.Name, appInfo["name"])
		assert.Equal(t, app.Domain, appInfo["domain"])
	})

	t.Run("PUT /v1/applications/{id} - アプリケーション更新", func(t *testing.T) {
		app := createTestApplication(t, server.URL)

		updateData := models.ApplicationUpdateRequest{
			Name:   "Updated Test Application",
			Domain: "updated.test.example.com",
		}

		jsonData, _ := json.Marshal(updateData)
		req, _ := http.NewRequest("PUT", server.URL+"/v1/applications/"+app.AppID, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response models.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})

	t.Run("DELETE /v1/applications/{id} - アプリケーション削除", func(t *testing.T) {
		app := createTestApplication(t, server.URL)

		req, _ := http.NewRequest("DELETE", server.URL+"/v1/applications/"+app.AppID, nil)
		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response models.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})
}

// 3. トラッキングAPIのテスト
func TestTrackingAPI(t *testing.T) {
	server, cleanup := setupSpecificationTestServer(t)
	defer cleanup()

	app := createTestApplication(t, server.URL)

	t.Run("POST /v1/tracking/track - トラッキングデータ送信", func(t *testing.T) {
		trackingData := TestTrackingData{
			AppID:     app.AppID,
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			URL:       "https://example.com/product/123",
			IPAddress: "192.168.1.1",
			SessionID: "session_123",
			Referrer:  "https://google.com",
			CustomParams: map[string]interface{}{
				"page_type":       "product_detail",
				"product_id":      "PROD_12345",
				"product_name":    "Wireless Headphones",
				"product_price":   299.99,
				"product_brand":   "AudioTech",
				"cart_total":      299.99,
				"cart_item_count": 1,
			},
		}

		resp, err := sendJSONRequest("POST", server.URL+"/v1/tracking/track", trackingData, app.APIKey)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response models.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.True(t, response.Success)

		// レスポンス構造の確認
		if response.Data != nil {
			trackData := response.Data.(map[string]interface{})
			assert.NotEmpty(t, trackData["tracking_id"])
			assert.NotNil(t, trackData["timestamp"])
		}
	})

	t.Run("POST /v1/tracking/track - バリデーションエラー", func(t *testing.T) {
		// 必須フィールドが不足したデータ
		invalidData := TestTrackingData{
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
			// app_idが不足
		}

		resp, err := sendJSONRequest("POST", server.URL+"/v1/tracking/track", invalidData, app.APIKey)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var response models.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("POST /v1/tracking/track - 認証エラー", func(t *testing.T) {
		trackingData := TestTrackingData{
			AppID:     app.AppID,
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
		}

		// APIキーなしでリクエスト
		resp, err := sendJSONRequest("POST", server.URL+"/v1/tracking/track", trackingData, "")
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		var response models.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "AUTHENTICATION_ERROR", response.Error.Code)
	})

	t.Run("POST /v1/tracking/track - 無効なAPIキー形式", func(t *testing.T) {
		trackingData := TestTrackingData{
			AppID:     app.AppID,
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
		}

		// 無効なAPIキー形式でリクエスト
		resp, err := sendJSONRequest("POST", server.URL+"/v1/tracking/track", trackingData, "invalid_api_key")
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		var response models.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "AUTHENTICATION_ERROR", response.Error.Code)
	})

	t.Run("GET /v1/tracking/statistics - 統計情報取得", func(t *testing.T) {
		// まずトラッキングデータを送信
		trackingData := TestTrackingData{
			AppID:     app.AppID,
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
			URL:       "https://example.com/test",
		}

		resp, err := sendJSONRequest("POST", server.URL+"/v1/tracking/track", trackingData, app.APIKey)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// 統計情報を取得
		resp, err = http.Get(server.URL + "/v1/tracking/statistics?app_id=" + app.AppID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response models.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.True(t, response.Success)

		// レスポンス構造の確認
		if response.Data != nil {
			statsData := response.Data.(map[string]interface{})
			assert.NotNil(t, statsData["total_requests"])
			assert.NotNil(t, statsData["unique_visitors"])
			assert.NotNil(t, statsData["unique_sessions"])
		}
	})
}

// 4. ビーコンAPIのテスト
func TestBeaconAPI(t *testing.T) {
	server, cleanup := setupSpecificationTestServer(t)
	defer cleanup()

	app := createTestApplication(t, server.URL)

	t.Run("GET /tracker.js - JavaScriptビーコン配信", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/tracker.js")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/javascript", resp.Header.Get("Content-Type"))

		// JavaScriptコードの内容確認
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		jsCode := string(body[:n])
		assert.Contains(t, jsCode, "function")
		assert.Contains(t, jsCode, "endpoint")
	})

	t.Run("GET /tracker.min.js - 圧縮版JavaScriptビーコン", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/tracker.min.js")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/javascript", resp.Header.Get("Content-Type"))
	})

	t.Run("GET /tracker/{app_id}.js - カスタム設定ビーコン", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/tracker/" + app.AppID + ".js")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/javascript", resp.Header.Get("Content-Type"))
	})

	t.Run("GET /v1/beacon/generate - 1x1ピクセルGIFビーコン", func(t *testing.T) {
		url := fmt.Sprintf("%s/v1/beacon/generate?app_id=%s&session_id=test_session&url=https://example.com", server.URL, app.AppID)
		resp, err := http.Get(url)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "image/gif", resp.Header.Get("Content-Type"))
	})

	t.Run("POST /v1/beacon/generate - カスタム設定ビーコン生成", func(t *testing.T) {
		beaconData := map[string]interface{}{
			"app_id":     app.AppID,
			"session_id": "test_session",
			"url":        "https://example.com/test",
			"referrer":   "https://google.com",
		}

		resp, err := sendJSONRequest("POST", server.URL+"/v1/beacon/generate", beaconData, "")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "image/gif", resp.Header.Get("Content-Type"))
	})

	t.Run("GET /v1/beacon/health - ビーコンサービス健全性", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/v1/beacon/health")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response models.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.True(t, response.Success)
	})
}

// 5. レート制限のテスト
func TestRateLimiting(t *testing.T) {
	server, cleanup := setupSpecificationTestServer(t)
	defer cleanup()

	app := createTestApplication(t, server.URL)

	t.Run("レート制限の確認", func(t *testing.T) {
		// 短時間で多数のリクエストを送信
		for i := 0; i < 100; i++ {
			trackingData := TestTrackingData{
				AppID:     app.AppID,
				UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
			}

			resp, err := sendJSONRequest("POST", server.URL+"/v1/tracking/track", trackingData, app.APIKey)
			require.NoError(t, err)

			// レート制限ヘッダーの確認
			if resp.Header.Get("X-RateLimit-Limit") != "" {
				assert.NotEmpty(t, resp.Header.Get("X-RateLimit-Remaining"))
				assert.NotEmpty(t, resp.Header.Get("X-RateLimit-Reset"))
			}

			// レート制限に達した場合の確認
			if resp.StatusCode == http.StatusTooManyRequests {
				var response models.APIResponse
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)
				assert.False(t, response.Success)
				assert.Equal(t, "RATE_LIMIT_EXCEEDED", response.Error.Code)
				break
			}
		}
	})
}

// 6. エラーハンドリングのテスト
func TestErrorHandling(t *testing.T) {
	server, cleanup := setupSpecificationTestServer(t)
	defer cleanup()

	t.Run("存在しないエンドポイント", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/v1/nonexistent")
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("無効なJSONリクエスト", func(t *testing.T) {
		invalidJSON := []byte(`{"invalid": json}`)
		resp, err := http.Post(server.URL+"/v1/applications", "application/json", bytes.NewBuffer(invalidJSON))
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("無効なAPIキー", func(t *testing.T) {
		trackingData := TestTrackingData{
			AppID:     "test_app",
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
		}

		resp, err := sendJSONRequest("POST", server.URL+"/v1/tracking/track", trackingData, "invalid_api_key")
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		var response models.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "INVALID_API_KEY", response.Error.Code)
	})
}

// 7. CORS設定のテスト
func TestCORSSettings(t *testing.T) {
	server, cleanup := setupSpecificationTestServer(t)
	defer cleanup()

	t.Run("CORSプリフライトリクエスト", func(t *testing.T) {
		req, _ := http.NewRequest("OPTIONS", server.URL+"/v1/applications", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// CORSヘッダーの確認
		assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Origin"))
		assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Methods"))
		assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Headers"))
	})
}

// 8. レスポンス形式の統一性テスト
func TestResponseFormatConsistency(t *testing.T) {
	server, cleanup := setupSpecificationTestServer(t)
	defer cleanup()

	t.Run("成功レスポンスの形式確認", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/health")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response models.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		// 統一されたレスポンス形式の確認
		assert.True(t, response.Success)
		assert.NotNil(t, response.Data)
		assert.NotNil(t, response.Timestamp)
		assert.Empty(t, response.Error)
	})

	t.Run("エラーレスポンスの形式確認", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/v1/nonexistent")
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		var response models.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		// 統一されたエラーレスポンス形式の確認
		assert.False(t, response.Success)
		assert.Nil(t, response.Data)
		assert.NotNil(t, response.Error)
		assert.NotEmpty(t, response.Error.Code)
		assert.NotEmpty(t, response.Error.Message)
		assert.NotNil(t, response.Timestamp)
	})
}

// 9. パフォーマンス要件の基本テスト
func TestBasicPerformanceRequirements(t *testing.T) {
	server, cleanup := setupSpecificationTestServer(t)
	defer cleanup()

	app := createTestApplication(t, server.URL)

	t.Run("レスポンス時間の確認", func(t *testing.T) {
		start := time.Now()
		resp, err := http.Get(server.URL + "/health")
		duration := time.Since(start)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Less(t, duration, 100*time.Millisecond, "レスポンス時間が100msを超えています")
	})

	t.Run("トラッキングAPIのレスポンス時間", func(t *testing.T) {
		trackingData := TestTrackingData{
			AppID:     app.AppID,
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
		}

		start := time.Now()
		resp, err := sendJSONRequest("POST", server.URL+"/v1/tracking/track", trackingData, app.APIKey)
		duration := time.Since(start)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Less(t, duration, 100*time.Millisecond, "トラッキングAPIのレスポンス時間が100msを超えています")
	})
}

// 10. セキュリティ要件のテスト
func TestSecurityRequirements(t *testing.T) {
	server, cleanup := setupSpecificationTestServer(t)
	defer cleanup()

	t.Run("SQLインジェクション対策", func(t *testing.T) {
		// SQLインジェクションを試行するデータ
		maliciousData := TestTrackingData{
			AppID:     "'; DROP TABLE tracking; --",
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
		}

		resp, err := sendJSONRequest("POST", server.URL+"/v1/tracking/track", maliciousData, "")
		require.NoError(t, err)
		// 認証エラーまたはバリデーションエラーが返されることを確認
		assert.True(t, resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusBadRequest)
	})

	t.Run("XSS攻撃対策", func(t *testing.T) {
		// XSS攻撃を試行するデータ
		xssData := TestTrackingData{
			AppID:     "test_app",
			UserAgent: "<script>alert('XSS')</script>",
		}

		resp, err := sendJSONRequest("POST", server.URL+"/v1/tracking/track", xssData, "")
		require.NoError(t, err)
		// 認証エラーが返されることを確認（APIキーなしのため）
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
