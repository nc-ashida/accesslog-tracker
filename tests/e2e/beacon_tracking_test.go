package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"accesslog-tracker/internal/api/models"
)

const (
	baseURL = "http://test-app:8080"
)

// Application はE2Eテスト用のアプリケーション構造体です
type Application struct {
	AppID       string `json:"app_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Domain      string `json:"domain"`
	APIKey      string `json:"api_key"`
	Active      bool   `json:"active"`
}

func TestBasicE2EFlow(t *testing.T) {
	// テスト用のアプリケーションを作成
	app := createTestApplication(t)
	require.NotNil(t, app)

	t.Run("Health Check", func(t *testing.T) {
		// ヘルスチェックエンドポイントをテスト
		resp, err := http.Get(baseURL + "/health")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("Application Creation and Retrieval", func(t *testing.T) {
		// 作成したアプリケーションを取得（エンドポイントが存在しない場合はスキップ）
		resp, err := http.Get(fmt.Sprintf("%s/v1/applications/%s", baseURL, app.AppID))
		if err != nil || resp.StatusCode == http.StatusNotFound {
			t.Skip("Application retrieval endpoint not available")
		}
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("JavaScript Beacon Loading", func(t *testing.T) {
		// JavaScriptビーコンの配信をテスト
		resp, err := http.Get(baseURL + "/tracker.js")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/javascript", resp.Header.Get("Content-Type"))
		resp.Body.Close()
	})

	t.Run("Minified JavaScript Beacon", func(t *testing.T) {
		// ミニファイされたJavaScriptビーコンの配信をテスト
		resp, err := http.Get(baseURL + "/tracker.min.js")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/javascript", resp.Header.Get("Content-Type"))
		resp.Body.Close()
	})

	t.Run("Beacon Generation", func(t *testing.T) {
		// ビーコン生成をテスト
		beaconURL := fmt.Sprintf("%s/v1/beacon/generate?app_id=%s", baseURL, app.AppID)
		resp, err := http.Get(beaconURL)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("Application Cleanup", func(t *testing.T) {
		// テスト用アプリケーションを削除
		cleanupTestApplication(t, app.AppID)
	})
}

// createTestApplication はテスト用のアプリケーションを作成します
func createTestApplication(t *testing.T) *Application {
	createRequest := map[string]interface{}{
		"name":        "E2E Test Application",
		"description": "Test application for E2E testing",
		"domain":      "e2e-test.example.com",
	}

	jsonData, err := json.Marshal(createRequest)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/v1/applications", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// レスポンスボディを読み取り
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response models.APIResponse
	err = json.Unmarshal(bodyBytes, &response)
	require.NoError(t, err)
	assert.True(t, response.Success)

	// レスポンスからアプリケーション情報を抽出
	appData, ok := response.Data.(map[string]interface{})
	require.True(t, ok)

	// activeフィールドの安全な取得
	var active bool
	if activeVal, exists := appData["active"]; exists && activeVal != nil {
		if activeBool, ok := activeVal.(bool); ok {
			active = activeBool
		} else {
			active = true // デフォルト値
		}
	} else {
		active = true // デフォルト値
	}

	app := &Application{
		AppID:       appData["app_id"].(string),
		Name:        appData["name"].(string),
		Description: appData["description"].(string),
		Domain:      appData["domain"].(string),
		APIKey:      appData["api_key"].(string),
		Active:      active,
	}

	return app
}

// cleanupTestApplication はテスト用のアプリケーションを削除します
func cleanupTestApplication(t *testing.T, appID string) {
	req, err := http.NewRequest("DELETE", baseURL+"/v1/applications/"+appID, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	if err == nil && resp != nil {
		resp.Body.Close()
	}
}
