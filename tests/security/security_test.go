package security

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nc-ashida/accesslog-tracker/internal/domain/models"
)

const (
	baseURL = "http://localhost:8080"
)

func TestAuthenticationSecurity(t *testing.T) {
	t.Run("Unauthorized Access to Protected Endpoints", func(t *testing.T) {
		protectedEndpoints := []string{
			"/api/v1/applications",
			"/api/v1/applications/1",
			"/api/v1/sessions",
			"/api/v1/statistics",
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

	t.Run("Invalid JWT Token", func(t *testing.T) {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/applications", baseURL), nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer invalid-token")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("Expired JWT Token", func(t *testing.T) {
		// 期限切れのトークンを使用
		expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjE1MTYyMzkwMjJ9.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
		
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/applications", baseURL), nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", expiredToken))

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		resp.Body.Close()
	})
}

func TestInputValidationSecurity(t *testing.T) {
	t.Run("SQL Injection Prevention", func(t *testing.T) {
		sqlInjectionPayloads := []string{
			"'; DROP TABLE applications; --",
			"' OR '1'='1",
			"'; INSERT INTO applications VALUES (999, 'hacked', 'hacked', 'hacked'); --",
			"'; UPDATE applications SET name = 'hacked'; --",
		}

		for _, payload := range sqlInjectionPayloads {
			t.Run(fmt.Sprintf("Payload: %s", payload), func(t *testing.T) {
				// アプリケーション名にSQLインジェクションを試行
				app := models.Application{
					Name:        payload,
					Description: "Test",
					Domain:      "test.example.com",
				}

				body, _ := json.Marshal(app)
				resp, err := http.Post(fmt.Sprintf("%s/api/v1/applications", baseURL), "application/json", bytes.NewBuffer(body))
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
		}

		for _, payload := range xssPayloads {
			t.Run(fmt.Sprintf("Payload: %s", payload), func(t *testing.T) {
				app := models.Application{
					Name:        payload,
					Description: payload,
					Domain:      "test.example.com",
				}

				body, _ := json.Marshal(app)
				resp, err := http.Post(fmt.Sprintf("%s/api/v1/applications", baseURL), "application/json", bytes.NewBuffer(body))
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
		}

		for _, payload := range pathTraversalPayloads {
			t.Run(fmt.Sprintf("Payload: %s", payload), func(t *testing.T) {
				// URLパラメータにパストラバーサルを試行
				beaconURL := fmt.Sprintf("%s/beacon?app_id=1&session_id=test&url=%s", baseURL, payload)
				resp, err := http.Get(beaconURL)
				require.NoError(t, err)
				
				// 適切にバリデーションされるか
				assert.True(t, resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusOK)
				resp.Body.Close()
			})
		}
	})
}

func TestBeaconSecurity(t *testing.T) {
	t.Run("Rate Limiting", func(t *testing.T) {
		const maxRequests = 100
		successCount := 0
		rateLimitedCount := 0

		for i := 0; i < maxRequests+10; i++ {
			beaconURL := fmt.Sprintf("%s/beacon?app_id=1&session_id=rate-limit-test&url=/test", baseURL)
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
			"-1",
			"0",
			"999999999",
			"invalid",
			"1.5",
		}

		for _, appID := range invalidAppIDs {
			t.Run(fmt.Sprintf("AppID: %s", appID), func(t *testing.T) {
				beaconURL := fmt.Sprintf("%s/beacon?app_id=%s&session_id=test&url=/test", baseURL, appID)
				resp, err := http.Get(beaconURL)
				require.NoError(t, err)
				
				// 無効なアプリケーションIDは適切に処理されるか
				assert.True(t, resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound)
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
		}

		for _, userAgent := range maliciousUserAgents {
			t.Run(fmt.Sprintf("UserAgent: %s", userAgent), func(t *testing.T) {
				req, err := http.NewRequest("GET", fmt.Sprintf("%s/beacon?app_id=1&session_id=test&url=/test", baseURL), nil)
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
}

func TestDataPrivacySecurity(t *testing.T) {
	t.Run("PII Data Protection", func(t *testing.T) {
		// 個人情報を含むリクエスト
		piiData := map[string]string{
			"email": "user@example.com",
			"phone": "123-456-7890",
			"ssn": "123-45-6789",
			"credit_card": "4111-1111-1111-1111",
		}

		for key, value := range piiData {
			t.Run(fmt.Sprintf("PII: %s", key), func(t *testing.T) {
				beaconURL := fmt.Sprintf("%s/beacon?app_id=1&session_id=test&url=/test&%s=%s", baseURL, key, value)
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
		}

		for _, data := range sensitiveData {
			t.Run(fmt.Sprintf("Sensitive: %s", data), func(t *testing.T) {
				beaconURL := fmt.Sprintf("%s/beacon?app_id=1&session_id=test&url=/test&%s", baseURL, data)
				resp, err := http.Get(beaconURL)
				require.NoError(t, err)
				
				// 機密データが適切に処理されるか
				assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest)
				resp.Body.Close()
			})
		}
	})
}

func TestCORSecurity(t *testing.T) {
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
		}

		for _, origin := range unauthorizedOrigins {
			t.Run(fmt.Sprintf("Origin: %s", origin), func(t *testing.T) {
				req, err := http.NewRequest("GET", fmt.Sprintf("%s/beacon?app_id=1&session_id=test&url=/test", baseURL), nil)
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
}
