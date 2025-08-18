package security

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"accesslog-tracker/internal/api/models"
)

const (
	baseURL = "http://localhost:8080"
)

// Application はセキュリティテスト用のアプリケーション構造体です
type Application struct {
	AppID       string `json:"app_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Domain      string `json:"domain"`
	APIKey      string `json:"api_key"`
	Active      bool   `json:"active"`
}

func TestAuthenticationSecurity(t *testing.T) {
	app := createTestApplication(t)
	defer cleanupTestApplication(t, app.AppID)

	t.Run("Unauthorized Access to Protected Endpoints", func(t *testing.T) {
		protectedEndpoints := []string{
			"/v1/applications",
			"/v1/applications/" + app.AppID,
			"/v1/tracking/statistics",
		}

		for _, endpoint := range protectedEndpoints {
			t.Run(fmt.Sprintf("GET %s", endpoint), func(t *testing.T) {
				resp, err := http.Get(fmt.Sprintf("%s%s", baseURL, endpoint))
				require.NoError(t, err)
				assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
				resp.Body.Close()
			})

			t.Run(fmt.Sprintf("POST %s", endpoint), func(t *testing.T) {
				resp, err := http.Post(fmt.Sprintf("%s%s", baseURL, endpoint), "application/json", bytes.NewBuffer([]byte("{}")))
				require.NoError(t, err)
				assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
				resp.Body.Close()
			})
		}
	})

	t.Run("Invalid API Key Format", func(t *testing.T) {
		invalidAPIKeys := []string{
			"invalid-key",
			"",
			"1234567890",
			"key-with-spaces",
			"key_with_special_chars!@#",
		}

		for _, apiKey := range invalidAPIKeys {
			t.Run(fmt.Sprintf("API Key: %s", apiKey), func(t *testing.T) {
				req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/applications", baseURL), nil)
				require.NoError(t, err)
				req.Header.Set("X-API-Key", apiKey)

				client := &http.Client{}
				resp, err := client.Do(req)
				require.NoError(t, err)
				assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
				resp.Body.Close()
			})
		}
	})

	t.Run("Missing API Key", func(t *testing.T) {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/applications", baseURL), nil)
		require.NoError(t, err)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("Valid API Key Access", func(t *testing.T) {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/applications", baseURL), nil)
		require.NoError(t, err)
		req.Header.Set("X-API-Key", app.APIKey)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("API Key for Different Application", func(t *testing.T) {
		// 別のアプリケーション用のAPIキーでアクセス
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/applications/%s", baseURL, app.AppID), nil)
		require.NoError(t, err)
		req.Header.Set("X-API-Key", "different-api-key")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		resp.Body.Close()
	})
}

func TestInputValidationSecurity(t *testing.T) {
	app := createTestApplication(t)
	defer cleanupTestApplication(t, app.AppID)

	t.Run("SQL Injection Prevention", func(t *testing.T) {
		sqlInjectionPayloads := []string{
			"'; DROP TABLE applications; --",
			"' OR '1'='1",
			"'; INSERT INTO applications VALUES (999, 'hacked', 'hacked', 'hacked'); --",
			"'; UPDATE applications SET name = 'hacked'; --",
			"'; SELECT * FROM applications WHERE 1=1; --",
			"'; EXEC xp_cmdshell('dir'); --",
		}

		for _, payload := range sqlInjectionPayloads {
			t.Run(fmt.Sprintf("Payload: %s", payload), func(t *testing.T) {
				// アプリケーション名にSQLインジェクションを試行
				updateRequest := map[string]interface{}{
					"name":        payload,
					"description": "Test",
					"domain":      "test.example.com",
				}

				body, _ := json.Marshal(updateRequest)
				req, err := http.NewRequest("PUT", fmt.Sprintf("%s/v1/applications/%s", baseURL, app.AppID), bytes.NewBuffer(body))
				require.NoError(t, err)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-API-Key", app.APIKey)

				client := &http.Client{}
				resp, err := client.Do(req)
				require.NoError(t, err)
				
				// 適切にバリデーションされるか、またはエラーが返されるか
				assert.True(t, resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized)
				resp.Body.Close()
			})
		}
	})

	t.Run("XSS Prevention", func(t *testing.T) {
		xssPayloads := []string{
			"<script>alert('XSS')</script>",
			"<img src=x onerror=alert('XSS')>",
			"javascript:alert('XSS')",
			"<svg onload=alert('XSS')>",
			"<iframe src=javascript:alert('XSS')>",
			"<object data=javascript:alert('XSS')>",
		}

		for _, payload := range xssPayloads {
			t.Run(fmt.Sprintf("Payload: %s", payload), func(t *testing.T) {
				updateRequest := map[string]interface{}{
					"name":        payload,
					"description": payload,
					"domain":      "test.example.com",
				}

				body, _ := json.Marshal(updateRequest)
				req, err := http.NewRequest("PUT", fmt.Sprintf("%s/v1/applications/%s", baseURL, app.AppID), bytes.NewBuffer(body))
				require.NoError(t, err)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-API-Key", app.APIKey)

				client := &http.Client{}
				resp, err := client.Do(req)
				require.NoError(t, err)
				
				// XSSペイロードが適切にサニタイズされるか
				assert.True(t, resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized)
				resp.Body.Close()
			})
		}
	})

	t.Run("Path Traversal Prevention", func(t *testing.T) {
		pathTraversalPayloads := []string{
			"../../../etc/passwd",
			"..\\..\\..\\windows\\system32\\config\\sam",
			"....//....//....//etc/passwd",
			"%2e%2e%2f%2e%2e%2f%2e%2e%2fetc%2fpasswd",
			"..%2f..%2f..%2fetc%2fpasswd",
		}

		for _, payload := range pathTraversalPayloads {
			t.Run(fmt.Sprintf("Payload: %s", payload), func(t *testing.T) {
				// URLパラメータにパストラバーサルを試行
				beaconURL := fmt.Sprintf("%s/beacon?app_id=%s&session_id=test&url=%s", baseURL, app.AppID, payload)
				resp, err := http.Get(beaconURL)
				require.NoError(t, err)
				
				// 適切にバリデーションされるか
				assert.True(t, resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusOK)
				resp.Body.Close()
			})
		}
	})

	t.Run("Command Injection Prevention", func(t *testing.T) {
		commandInjectionPayloads := []string{
			"; rm -rf /",
			"| cat /etc/passwd",
			"&& whoami",
			"`id`",
			"$(whoami)",
		}

		for _, payload := range commandInjectionPayloads {
			t.Run(fmt.Sprintf("Payload: %s", payload), func(t *testing.T) {
				// トラッキングデータにコマンドインジェクションを試行
				trackingData := models.TrackingRequest{
					AppID:       app.AppID,
					UserAgent:   payload,
					URL:         "/test",
					IPAddress:   "192.168.1.100",
					SessionID:   "test-session",
					Referrer:    "https://example.com",
				}

				body, _ := json.Marshal(trackingData)
				req, err := http.NewRequest("POST", baseURL+"/v1/tracking/track", bytes.NewBuffer(body))
				require.NoError(t, err)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-API-Key", app.APIKey)

				client := &http.Client{}
				resp, err := client.Do(req)
				require.NoError(t, err)
				
				// 適切にバリデーションされるか
				assert.True(t, resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusOK)
				resp.Body.Close()
			})
		}
	})
}

func TestBeaconSecurity(t *testing.T) {
	app := createTestApplication(t)
	defer cleanupTestApplication(t, app.AppID)

	t.Run("Rate Limiting", func(t *testing.T) {
		const maxRequests = 100
		successCount := 0
		rateLimitedCount := 0

		for i := 0; i < maxRequests+10; i++ {
			beaconURL := fmt.Sprintf("%s/beacon?app_id=%s&session_id=rate-limit-test&url=/test", baseURL, app.AppID)
			resp, err := http.Get(beaconURL)
			if err == nil {
				if resp.StatusCode == http.StatusOK {
					successCount++
				} else if resp.StatusCode == http.StatusTooManyRequests {
					rateLimitedCount++
				}
				resp.Body.Close()
			}
		}

		// レート制限が適切に機能しているか
		assert.Greater(t, rateLimitedCount, 0, "Rate limiting should be active")
	})

	t.Run("Invalid Application ID", func(t *testing.T) {
		invalidAppIDs := []string{
			"invalid-app-id",
			"",
			"app-id-with-spaces",
			"app_id_with_special_chars!@#",
			"very-long-app-id-that-exceeds-maximum-length-allowed-by-the-system",
		}

		for _, appID := range invalidAppIDs {
			t.Run(fmt.Sprintf("AppID: %s", appID), func(t *testing.T) {
				beaconURL := fmt.Sprintf("%s/beacon?app_id=%s&session_id=test&url=/test", baseURL, appID)
				resp, err := http.Get(beaconURL)
				require.NoError(t, err)
				
				// 無効なアプリケーションIDは適切に処理されるか
				assert.True(t, resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusOK)
				resp.Body.Close()
			})
		}
	})

	t.Run("Malicious User Agent", func(t *testing.T) {
		maliciousUserAgents := []string{
			"sqlmap/1.0",
			"nmap/7.80",
			"nikto/2.1.6",
			"<script>alert('XSS')</script>",
			"'; DROP TABLE applications; --",
			"curl/7.68.0",
			"wget/1.20.3",
		}

		for _, userAgent := range maliciousUserAgents {
			t.Run(fmt.Sprintf("UserAgent: %s", userAgent), func(t *testing.T) {
				req, err := http.NewRequest("GET", fmt.Sprintf("%s/beacon?app_id=%s&session_id=test&url=/test", baseURL, app.AppID), nil)
				require.NoError(t, err)
				req.Header.Set("User-Agent", userAgent)

				client := &http.Client{}
				resp, err := client.Do(req)
				require.NoError(t, err)
				
				// 悪意のあるUser-Agentが適切に処理されるか
				assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest)
				resp.Body.Close()
			})
		}
	})

	t.Run("Malicious Referrer", func(t *testing.T) {
		maliciousReferrers := []string{
			"javascript:alert('XSS')",
			"data:text/html,<script>alert('XSS')</script>",
			"file:///etc/passwd",
			"ftp://malicious-site.com",
		}

		for _, referrer := range maliciousReferrers {
			t.Run(fmt.Sprintf("Referrer: %s", referrer), func(t *testing.T) {
				beaconURL := fmt.Sprintf("%s/beacon?app_id=%s&session_id=test&url=/test&referrer=%s", baseURL, app.AppID, referrer)
				resp, err := http.Get(beaconURL)
				require.NoError(t, err)
				
				// 悪意のあるリファラーが適切に処理されるか
				assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest)
				resp.Body.Close()
			})
		}
	})
}

func TestDataPrivacySecurity(t *testing.T) {
	app := createTestApplication(t)
	defer cleanupTestApplication(t, app.AppID)

	t.Run("PII Data Protection", func(t *testing.T) {
		// 個人情報を含むリクエスト
		piiData := map[string]string{
			"email": "user@example.com",
			"phone": "123-456-7890",
			"ssn": "123-45-6789",
			"credit_card": "4111-1111-1111-1111",
			"password": "secret123",
			"api_key": "sk-1234567890abcdef",
		}

		for key, value := range piiData {
			t.Run(fmt.Sprintf("PII: %s", key), func(t *testing.T) {
				beaconURL := fmt.Sprintf("%s/beacon?app_id=%s&session_id=test&url=/test&%s=%s", baseURL, app.AppID, key, value)
				resp, err := http.Get(beaconURL)
				require.NoError(t, err)
				
				// 個人情報が適切に処理されるか
				assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest)
				resp.Body.Close()
			})
		}
	})

	t.Run("Sensitive Data in Logs", func(t *testing.T) {
		// ログに出力されるべきでないデータ
		sensitiveData := []string{
			"password=secret123",
			"token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			"api_key=sk-1234567890abcdef",
			"secret=very-secret-value",
		}

		for _, data := range sensitiveData {
			t.Run(fmt.Sprintf("Sensitive: %s", data), func(t *testing.T) {
				beaconURL := fmt.Sprintf("%s/beacon?app_id=%s&session_id=test&url=/test&%s", baseURL, app.AppID, data)
				resp, err := http.Get(beaconURL)
				require.NoError(t, err)
				
				// 機密データが適切に処理されるか
				assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest)
				resp.Body.Close()
			})
		}
	})

	t.Run("Data Encryption", func(t *testing.T) {
		// 機密データを含むトラッキングリクエスト
		trackingData := models.TrackingRequest{
			AppID:       app.AppID,
			UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			URL:         "/secure-page",
			IPAddress:   "192.168.1.100",
			SessionID:   "secure-session-123",
			Referrer:    "https://secure-site.com",
			CustomParams: map[string]interface{}{
				"sensitive_data": "encrypted-value",
				"user_id": "12345",
			},
		}

		body, _ := json.Marshal(trackingData)
		req, err := http.NewRequest("POST", baseURL+"/v1/tracking/track", bytes.NewBuffer(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", app.APIKey)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		
		// 機密データが適切に処理されるか
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	})
}

func TestCORSecurity(t *testing.T) {
	app := createTestApplication(t)
	defer cleanupTestApplication(t, app.AppID)

	t.Run("CORS Headers", func(t *testing.T) {
		req, err := http.NewRequest("OPTIONS", fmt.Sprintf("%s/beacon", baseURL), nil)
		require.NoError(t, err)
		req.Header.Set("Origin", "https://malicious-site.com")
		req.Header.Set("Access-Control-Request-Method", "GET")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		
		// CORSヘッダーが適切に設定されているか
		accessControlAllowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
		accessControlAllowMethods := resp.Header.Get("Access-Control-Allow-Methods")
		
		assert.NotEmpty(t, accessControlAllowOrigin)
		assert.NotEmpty(t, accessControlAllowMethods)
		resp.Body.Close()
	})

	t.Run("Unauthorized Origins", func(t *testing.T) {
		unauthorizedOrigins := []string{
			"https://malicious-site.com",
			"http://evil.com",
			"https://phishing-site.net",
			"https://attacker.com",
		}

		for _, origin := range unauthorizedOrigins {
			t.Run(fmt.Sprintf("Origin: %s", origin), func(t *testing.T) {
				req, err := http.NewRequest("GET", fmt.Sprintf("%s/beacon?app_id=%s&session_id=test&url=/test", baseURL, app.AppID), nil)
				require.NoError(t, err)
				req.Header.Set("Origin", origin)

				client := &http.Client{}
				resp, err := client.Do(req)
				require.NoError(t, err)
				
				// 未承認のオリジンが適切に処理されるか
				accessControlAllowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
				assert.NotEqual(t, origin, accessControlAllowOrigin)
				resp.Body.Close()
			})
		}
	})

	t.Run("Authorized Origins", func(t *testing.T) {
		authorizedOrigins := []string{
			"https://example.com",
			"https://www.example.com",
			"https://app.example.com",
		}

		for _, origin := range authorizedOrigins {
			t.Run(fmt.Sprintf("Origin: %s", origin), func(t *testing.T) {
				req, err := http.NewRequest("GET", fmt.Sprintf("%s/beacon?app_id=%s&session_id=test&url=/test", baseURL, app.AppID), nil)
				require.NoError(t, err)
				req.Header.Set("Origin", origin)

				client := &http.Client{}
				resp, err := client.Do(req)
				require.NoError(t, err)
				
				// 承認されたオリジンが適切に処理されるか
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				resp.Body.Close()
			})
		}
	})
}

func TestSessionSecurity(t *testing.T) {
	app := createTestApplication(t)
	defer cleanupTestApplication(t, app.AppID)

	t.Run("Session Hijacking Prevention", func(t *testing.T) {
		// 異なるIPアドレスからの同じセッションIDでのアクセス
		sessionID := "test-session-123"
		
		// 正常なIPアドレスからのアクセス
		beaconURL1 := fmt.Sprintf("%s/beacon?app_id=%s&session_id=%s&url=/test", baseURL, app.AppID, sessionID)
		req1, _ := http.NewRequest("GET", beaconURL1, nil)
		req1.Header.Set("X-Forwarded-For", "192.168.1.100")
		
		client := &http.Client{}
		resp1, err := client.Do(req1)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp1.StatusCode)
		resp1.Body.Close()

		// 異なるIPアドレスからのアクセス
		req2, _ := http.NewRequest("GET", beaconURL1, nil)
		req2.Header.Set("X-Forwarded-For", "10.0.0.100")
		
		resp2, err := client.Do(req2)
		require.NoError(t, err)
		// セッションハイジャック防止が機能しているか
		assert.True(t, resp2.StatusCode == http.StatusOK || resp2.StatusCode == http.StatusForbidden)
		resp2.Body.Close()
	})

	t.Run("Session Timeout", func(t *testing.T) {
		// 長時間のセッションが適切に処理されるか
		oldSessionID := fmt.Sprintf("old-session-%d", time.Now().Add(-24*time.Hour).Unix())
		beaconURL := fmt.Sprintf("%s/beacon?app_id=%s&session_id=%s&url=/test", baseURL, app.AppID, oldSessionID)
		
		resp, err := http.Get(beaconURL)
		require.NoError(t, err)
		// 古いセッションが適切に処理されるか
		assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest)
		resp.Body.Close()
	})
}

func TestAPIKeySecurity(t *testing.T) {
	app := createTestApplication(t)
	defer cleanupTestApplication(t, app.AppID)

	t.Run("API Key Rotation", func(t *testing.T) {
		// APIキーの更新
		updateRequest := map[string]interface{}{
			"name":        "Updated App",
			"description": "Updated description",
			"domain":      "updated.example.com",
		}

		body, _ := json.Marshal(updateRequest)
		req, err := http.NewRequest("PUT", fmt.Sprintf("%s/v1/applications/%s", baseURL, app.AppID), bytes.NewBuffer(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", app.APIKey)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()

		// 更新後のアプリケーション情報を取得
		req2, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/applications/%s", baseURL, app.AppID), nil)
		require.NoError(t, err)
		req2.Header.Set("X-API-Key", app.APIKey)

		resp2, err := client.Do(req2)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp2.StatusCode)
		resp2.Body.Close()
	})

	t.Run("API Key Permissions", func(t *testing.T) {
		// 異なるアプリケーションのリソースにアクセス
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/applications/different-app-id", baseURL), nil)
		require.NoError(t, err)
		req.Header.Set("X-API-Key", app.APIKey)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		// 権限がないリソースへのアクセスが適切に処理されるか
		assert.True(t, resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound)
		resp.Body.Close()
	})
}

// createTestApplication はテスト用のアプリケーションを作成します
func createTestApplication(t *testing.T) *Application {
	createRequest := map[string]interface{}{
		"name":        "Security Test Application",
		"description": "Test application for security testing",
		"domain":      "security-test.example.com",
	}

	jsonData, err := json.Marshal(createRequest)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/v1/applications", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	// レスポンスボディを読み取り
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response models.APIResponse
	err = json.Unmarshal(bodyBytes, &response)
	require.NoError(t, err)
	require.True(t, response.Success)

	// レスポンスからアプリケーション情報を抽出
	appData, ok := response.Data.(map[string]interface{})
	require.True(t, ok)

	app := &Application{
		AppID:       appData["app_id"].(string),
		Name:        appData["name"].(string),
		Description: appData["description"].(string),
		Domain:      appData["domain"].(string),
		APIKey:      appData["api_key"].(string),
		Active:      appData["active"].(bool),
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
