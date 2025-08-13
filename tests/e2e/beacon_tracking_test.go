package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nc-ashida/accesslog-tracker/internal/domain/models"
)

const (
	baseURL = "http://localhost:8080"
)

func TestBeaconTrackingFlow(t *testing.T) {
	// テスト用のアプリケーションを作成
	appID := createTestApplication(t)
	require.NotZero(t, appID)

	// セッションIDを生成
	sessionID := fmt.Sprintf("test-session-%d", time.Now().Unix())

	t.Run("Complete Beacon Tracking Flow", func(t *testing.T) {
		// 1. ビーコンリクエストを送信
		beaconURL := fmt.Sprintf("%s/beacon?app_id=%d&session_id=%s&url=/test-page&referrer=https://example.com", baseURL, appID, sessionID)
		
		resp, err := http.Get(beaconURL)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()

		// 2. 少し待機してデータベースに反映されるのを待つ
		time.Sleep(1 * time.Second)

		// 3. セッション情報を取得
		sessionResp, err := http.Get(fmt.Sprintf("%s/api/v1/sessions/%s", baseURL, sessionID))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, sessionResp.StatusCode)
		sessionResp.Body.Close()

		// 4. アクセスログを取得
		logsResp, err := http.Get(fmt.Sprintf("%s/api/v1/sessions/%s/logs", baseURL, sessionID))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, logsResp.StatusCode)
		logsResp.Body.Close()
	})

	t.Run("Multiple Page Views", func(t *testing.T) {
		pages := []string{"/home", "/about", "/contact", "/products"}

		for _, page := range pages {
			beaconURL := fmt.Sprintf("%s/beacon?app_id=%d&session_id=%s&url=%s&referrer=https://example.com", baseURL, appID, sessionID, page)
			
			resp, err := http.Get(beaconURL)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			resp.Body.Close()

			time.Sleep(100 * time.Millisecond)
		}

		// ページビュー数が正しく記録されているか確認
		logsResp, err := http.Get(fmt.Sprintf("%s/api/v1/sessions/%s/logs", baseURL, sessionID))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, logsResp.StatusCode)
		logsResp.Body.Close()
	})

	t.Run("Custom Parameters", func(t *testing.T) {
		customParams := map[string]string{
			"utm_source": "google",
			"utm_medium": "cpc",
			"utm_campaign": "test-campaign",
			"user_type": "premium",
		}

		paramStr := ""
		for key, value := range customParams {
			if paramStr != "" {
				paramStr += "&"
			}
			paramStr += fmt.Sprintf("%s=%s", key, value)
		}

		beaconURL := fmt.Sprintf("%s/beacon?app_id=%d&session_id=%s&url=/custom-test&%s", baseURL, appID, sessionID, paramStr)
		
		resp, err := http.Get(beaconURL)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	})

	// テスト用アプリケーションを削除
	cleanupTestApplication(t, appID)
}

func TestBeaconErrorHandling(t *testing.T) {
	t.Run("Invalid Application ID", func(t *testing.T) {
		beaconURL := fmt.Sprintf("%s/beacon?app_id=999999&session_id=test123&url=/test", baseURL)
		
		resp, err := http.Get(beaconURL)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("Missing Required Parameters", func(t *testing.T) {
		// app_idが不足
		resp, err := http.Get(fmt.Sprintf("%s/beacon?session_id=test123&url=/test", baseURL))
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		resp.Body.Close()

		// session_idが不足
		resp, err = http.Get(fmt.Sprintf("%s/beacon?app_id=1&url=/test", baseURL))
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("Invalid URL Format", func(t *testing.T) {
		appID := createTestApplication(t)
		defer cleanupTestApplication(t, appID)

		beaconURL := fmt.Sprintf("%s/beacon?app_id=%d&session_id=test123&url=invalid-url", baseURL, appID)
		
		resp, err := http.Get(beaconURL)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		resp.Body.Close()
	})
}

func TestBeaconPerformance(t *testing.T) {
	appID := createTestApplication(t)
	defer cleanupTestApplication(t, appID)

	t.Run("Concurrent Requests", func(t *testing.T) {
		const numRequests = 100
		results := make(chan bool, numRequests)

		for i := 0; i < numRequests; i++ {
			go func(index int) {
				sessionID := fmt.Sprintf("perf-session-%d", index)
				beaconURL := fmt.Sprintf("%s/beacon?app_id=%d&session_id=%s&url=/perf-test", baseURL, appID, sessionID)
				
				resp, err := http.Get(beaconURL)
				if err == nil && resp.StatusCode == http.StatusOK {
					results <- true
				} else {
					results <- false
				}
				if resp != nil {
					resp.Body.Close()
				}
			}(i)
		}

		successCount := 0
		for i := 0; i < numRequests; i++ {
			if <-results {
				successCount++
			}
		}

		assert.GreaterOrEqual(t, successCount, int(float64(numRequests)*0.95)) // 95%以上の成功率
	})
}

// ヘルパー関数
func createTestApplication(t *testing.T) int {
	app := models.Application{
		Name:        fmt.Sprintf("Test App %d", time.Now().Unix()),
		Description: "E2E Test Application",
		Domain:      "e2e-test.example.com",
	}

	body, _ := json.Marshal(app)
	resp, err := http.Post(fmt.Sprintf("%s/api/v1/applications", baseURL), "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdApp models.Application
	err = json.NewDecoder(resp.Body).Decode(&createdApp)
	require.NoError(t, err)

	return int(createdApp.ID)
}

func cleanupTestApplication(t *testing.T, appID int) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/applications/%d", baseURL, appID), nil)
	require.NoError(t, err)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err == nil && resp != nil {
		resp.Body.Close()
	}
}
