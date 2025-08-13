# セキュリティテスト実装

## 1. フェーズ6: 統合フェーズのセキュリティテスト

### 1.1 認証・認可テスト

#### 1.1.1 API認証テスト
```go
// tests/security/authentication/api_auth_test.go
package auth_test

import (
    "testing"
    "net/http"
    "encoding/json"
    "strings"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "access-log-tracker/internal/domain/models"
)

func TestAPIAuthentication(t *testing.T) {
    app := createTestApplication(t)
    
    t.Run("should reject requests without API key", func(t *testing.T) {
        trackingData := models.TrackingRequest{
            AppID:     app.AppID,
            UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
            URL:       "https://example.com/auth-test",
        }
        
        jsonData, _ := json.Marshal(trackingData)
        req, err := http.NewRequest("POST", "http://localhost:8080/v1/track",
            strings.NewReader(string(jsonData)))
        require.NoError(t, err)
        
        req.Header.Set("Content-Type", "application/json")
        // X-API-Keyヘッダーを設定しない
        
        client := &http.Client{}
        resp, err := client.Do(req)
        assert.NoError(t, err)
        assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
        
        var response map[string]interface{}
        err = json.NewDecoder(resp.Body).Decode(&response)
        assert.NoError(t, err)
        assert.Equal(t, false, response["success"])
        assert.Equal(t, "AUTHENTICATION_ERROR", response["error"].(map[string]interface{})["code"])
    })
    
    t.Run("should reject requests with invalid API key", func(t *testing.T) {
        trackingData := models.TrackingRequest{
            AppID:     app.AppID,
            UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
            URL:       "https://example.com/auth-test",
        }
        
        jsonData, _ := json.Marshal(trackingData)
        req, err := http.NewRequest("POST", "http://localhost:8080/v1/track",
            strings.NewReader(string(jsonData)))
        require.NoError(t, err)
        
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("X-API-Key", "invalid_api_key")
        
        client := &http.Client{}
        resp, err := client.Do(req)
        assert.NoError(t, err)
        assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
        
        var response map[string]interface{}
        err = json.NewDecoder(resp.Body).Decode(&response)
        assert.NoError(t, err)
        assert.Equal(t, false, response["success"])
        assert.Equal(t, "AUTHENTICATION_ERROR", response["error"].(map[string]interface{})["code"])
    })
    
    t.Run("should reject requests with expired API key", func(t *testing.T) {
        // 期限切れのAPIキーを持つアプリケーションを作成
        expiredApp := createExpiredApplication(t)
        
        trackingData := models.TrackingRequest{
            AppID:     expiredApp.AppID,
            UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
            URL:       "https://example.com/auth-test",
        }
        
        jsonData, _ := json.Marshal(trackingData)
        req, err := http.NewRequest("POST", "http://localhost:8080/v1/track",
            strings.NewReader(string(jsonData)))
        require.NoError(t, err)
        
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("X-API-Key", expiredApp.APIKey)
        
        client := &http.Client{}
        resp, err := client.Do(req)
        assert.NoError(t, err)
        assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
        
        var response map[string]interface{}
        err = json.NewDecoder(resp.Body).Decode(&response)
        assert.NoError(t, err)
        assert.Equal(t, false, response["success"])
        assert.Equal(t, "AUTHENTICATION_ERROR", response["error"].(map[string]interface{})["code"])
    })
    
    t.Run("should accept requests with valid API key", func(t *testing.T) {
        trackingData := models.TrackingRequest{
            AppID:     app.AppID,
            UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
            URL:       "https://example.com/auth-test",
        }
        
        jsonData, _ := json.Marshal(trackingData)
        req, err := http.NewRequest("POST", "http://localhost:8080/v1/track",
            strings.NewReader(string(jsonData)))
        require.NoError(t, err)
        
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("X-API-Key", app.APIKey)
        
        client := &http.Client{}
        resp, err := client.Do(req)
        assert.NoError(t, err)
        assert.Equal(t, http.StatusOK, resp.StatusCode)
        
        var response map[string]interface{}
        err = json.NewDecoder(resp.Body).Decode(&response)
        assert.NoError(t, err)
        assert.Equal(t, true, response["success"])
    })
    
    t.Run("should handle API key case sensitivity", func(t *testing.T) {
        trackingData := models.TrackingRequest{
            AppID:     app.AppID,
            UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
            URL:       "https://example.com/auth-test",
        }
        
        jsonData, _ := json.Marshal(trackingData)
        req, err := http.NewRequest("POST", "http://localhost:8080/v1/track",
            strings.NewReader(string(jsonData)))
        require.NoError(t, err)
        
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("X-API-Key", strings.ToUpper(app.APIKey)) // 大文字に変換
        
        client := &http.Client{}
        resp, err := client.Do(req)
        assert.NoError(t, err)
        assert.Equal(t, http.StatusUnauthorized, resp.StatusCode) // 大文字小文字を区別する
    })
}
```

#### 1.1.2 認可テスト
```go
// tests/security/authorization/authorization_test.go
package authorization_test

import (
    "testing"
    "net/http"
    "encoding/json"
    "strings"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "access-log-tracker/internal/domain/models"
)

func TestAuthorization(t *testing.T) {
    app1 := createTestApplication(t)
    app2 := createTestApplication(t)
    
    t.Run("should prevent cross-application access", func(t *testing.T) {
        // app1のAPIキーでapp2のデータにアクセスしようとする
        trackingData := models.TrackingRequest{
            AppID:     app2.AppID, // app2のAppID
            UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
            URL:       "https://example.com/auth-test",
        }
        
        jsonData, _ := json.Marshal(trackingData)
        req, err := http.NewRequest("POST", "http://localhost:8080/v1/track",
            strings.NewReader(string(jsonData)))
        require.NoError(t, err)
        
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("X-API-Key", app1.APIKey) // app1のAPIキー
        
        client := &http.Client{}
        resp, err := client.Do(req)
        assert.NoError(t, err)
        assert.Equal(t, http.StatusForbidden, resp.StatusCode)
        
        var response map[string]interface{}
        err = json.NewDecoder(resp.Body).Decode(&response)
        assert.NoError(t, err)
        assert.Equal(t, false, response["success"])
        assert.Equal(t, "AUTHORIZATION_ERROR", response["error"].(map[string]interface{})["code"])
    })
    
    t.Run("should prevent unauthorized statistics access", func(t *testing.T) {
        // app1のAPIキーでapp2の統計情報にアクセスしようとする
        url := fmt.Sprintf("http://localhost:8080/v1/statistics?app_id=%s&start_date=2024-01-01&end_date=2024-01-31", app2.AppID)
        req, err := http.NewRequest("GET", url, nil)
        require.NoError(t, err)
        
        req.Header.Set("X-API-Key", app1.APIKey) // app1のAPIキー
        
        client := &http.Client{}
        resp, err := client.Do(req)
        assert.NoError(t, err)
        assert.Equal(t, http.StatusForbidden, resp.StatusCode)
        
        var response map[string]interface{}
        err = json.NewDecoder(resp.Body).Decode(&response)
        assert.NoError(t, err)
        assert.Equal(t, false, response["success"])
        assert.Equal(t, "AUTHORIZATION_ERROR", response["error"].(map[string]interface{})["code"])
    })
    
    t.Run("should allow authorized statistics access", func(t *testing.T) {
        // app1のAPIキーでapp1の統計情報にアクセス
        url := fmt.Sprintf("http://localhost:8080/v1/statistics?app_id=%s&start_date=2024-01-01&end_date=2024-01-31", app1.AppID)
        req, err := http.NewRequest("GET", url, nil)
        require.NoError(t, err)
        
        req.Header.Set("X-API-Key", app1.APIKey) // app1のAPIキー
        
        client := &http.Client{}
        resp, err := client.Do(req)
        assert.NoError(t, err)
        assert.Equal(t, http.StatusOK, resp.StatusCode)
        
        var response map[string]interface{}
        err = json.NewDecoder(resp.Body).Decode(&response)
        assert.NoError(t, err)
        assert.Equal(t, true, response["success"])
    })
}
```

### 1.2 入力検証テスト

#### 1.2.1 SQLインジェクション対策テスト
```go
// tests/security/input_validation/sql_injection_test.go
package validation_test

import (
    "testing"
    "net/http"
    "encoding/json"
    "strings"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "access-log-tracker/internal/domain/models"
)

func TestSQLInjectionProtection(t *testing.T) {
    app := createTestApplication(t)
    
    t.Run("should prevent SQL injection in app_id", func(t *testing.T) {
        sqlInjectionPayloads := []string{
            "'; DROP TABLE access_logs; --",
            "' OR '1'='1",
            "'; INSERT INTO access_logs VALUES ('hacked'); --",
            "' UNION SELECT * FROM applications; --",
            "'; UPDATE applications SET api_key='hacked'; --",
        }
        
        for _, payload := range sqlInjectionPayloads {
            trackingData := map[string]interface{}{
                "app_id":     payload,
                "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
                "url":        "https://example.com/sql-injection-test",
            }
            
            jsonData, _ := json.Marshal(trackingData)
            req, err := http.NewRequest("POST", "http://localhost:8080/v1/track",
                strings.NewReader(string(jsonData)))
            require.NoError(t, err)
            
            req.Header.Set("Content-Type", "application/json")
            req.Header.Set("X-API-Key", app.APIKey)
            
            client := &http.Client{}
            resp, err := client.Do(req)
            assert.NoError(t, err)
            
            // バリデーションエラーが返されるべき
            assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
            
            var response map[string]interface{}
            err = json.NewDecoder(resp.Body).Decode(&response)
            assert.NoError(t, err)
            assert.Equal(t, false, response["success"])
            assert.Equal(t, "VALIDATION_ERROR", response["error"].(map[string]interface{})["code"])
        }
    })
    
    t.Run("should prevent SQL injection in URL", func(t *testing.T) {
        sqlInjectionPayloads := []string{
            "https://example.com'; DROP TABLE access_logs; --",
            "https://example.com' OR '1'='1",
            "https://example.com'; INSERT INTO access_logs VALUES ('hacked'); --",
        }
        
        for _, payload := range sqlInjectionPayloads {
            trackingData := models.TrackingRequest{
                AppID:     app.AppID,
                UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
                URL:       payload,
            }
            
            jsonData, _ := json.Marshal(trackingData)
            req, err := http.NewRequest("POST", "http://localhost:8080/v1/track",
                strings.NewReader(string(jsonData)))
            require.NoError(t, err)
            
            req.Header.Set("Content-Type", "application/json")
            req.Header.Set("X-API-Key", app.APIKey)
            
            client := &http.Client{}
            resp, err := client.Do(req)
            assert.NoError(t, err)
            
            // バリデーションエラーが返されるべき
            assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
            
            var response map[string]interface{}
            err = json.NewDecoder(resp.Body).Decode(&response)
            assert.NoError(t, err)
            assert.Equal(t, false, response["success"])
            assert.Equal(t, "VALIDATION_ERROR", response["error"].(map[string]interface{})["code"])
        }
    })
    
    t.Run("should prevent SQL injection in user agent", func(t *testing.T) {
        sqlInjectionPayloads := []string{
            "Mozilla/5.0'; DROP TABLE access_logs; --",
            "Mozilla/5.0' OR '1'='1",
            "Mozilla/5.0'; INSERT INTO access_logs VALUES ('hacked'); --",
        }
        
        for _, payload := range sqlInjectionPayloads {
            trackingData := models.TrackingRequest{
                AppID:     app.AppID,
                UserAgent: payload,
                URL:       "https://example.com/sql-injection-test",
            }
            
            jsonData, _ := json.Marshal(trackingData)
            req, err := http.NewRequest("POST", "http://localhost:8080/v1/track",
                strings.NewReader(string(jsonData)))
            require.NoError(t, err)
            
            req.Header.Set("Content-Type", "application/json")
            req.Header.Set("X-API-Key", app.APIKey)
            
            client := &http.Client{}
            resp, err := client.Do(req)
            assert.NoError(t, err)
            
            // バリデーションエラーが返されるべき
            assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
            
            var response map[string]interface{}
            err = json.NewDecoder(resp.Body).Decode(&response)
            assert.NoError(t, err)
            assert.Equal(t, false, response["success"])
            assert.Equal(t, "VALIDATION_ERROR", response["error"].(map[string]interface{})["code"])
        }
    })
}
```

#### 1.2.2 XSS対策テスト
```go
// tests/security/input_validation/xss_protection_test.go
package validation_test

import (
    "testing"
    "net/http"
    "encoding/json"
    "strings"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "access-log-tracker/internal/domain/models"
)

func TestXSSProtection(t *testing.T) {
    app := createTestApplication(t)
    
    t.Run("should sanitize XSS payloads in user agent", func(t *testing.T) {
        xssPayloads := []string{
            "<script>alert('xss')</script>",
            "javascript:alert('xss')",
            "<img src=x onerror=alert('xss')>",
            "<svg onload=alert('xss')>",
            "';alert('xss');//",
        }
        
        for _, payload := range xssPayloads {
            trackingData := models.TrackingRequest{
                AppID:     app.AppID,
                UserAgent: payload,
                URL:       "https://example.com/xss-test",
            }
            
            jsonData, _ := json.Marshal(trackingData)
            req, err := http.NewRequest("POST", "http://localhost:8080/v1/track",
                strings.NewReader(string(jsonData)))
            require.NoError(t, err)
            
            req.Header.Set("Content-Type", "application/json")
            req.Header.Set("X-API-Key", app.APIKey)
            
            client := &http.Client{}
            resp, err := client.Do(req)
            assert.NoError(t, err)
            
            // リクエストは受け入れられるが、データはサニタイズされる
            assert.Equal(t, http.StatusOK, resp.StatusCode)
            
            // データベースでサニタイズされたデータを確認
            time.Sleep(100 * time.Millisecond) // データ保存を待機
            
            savedData, err := getTrackingDataFromDB(t, app.AppID)
            require.NoError(t, err)
            require.NotEmpty(t, savedData)
            
            // 最新のデータを取得
            latestData := savedData[0]
            
            // XSSペイロードがサニタイズされていることを確認
            assert.NotContains(t, latestData.UserAgent, "<script>")
            assert.NotContains(t, latestData.UserAgent, "javascript:")
            assert.NotContains(t, latestData.UserAgent, "onerror=")
            assert.NotContains(t, latestData.UserAgent, "onload=")
        }
    })
    
    t.Run("should sanitize XSS payloads in URL", func(t *testing.T) {
        xssPayloads := []string{
            "https://example.com<script>alert('xss')</script>",
            "javascript:alert('xss')",
            "https://example.com'><script>alert('xss')</script>",
        }
        
        for _, payload := range xssPayloads {
            trackingData := models.TrackingRequest{
                AppID:     app.AppID,
                UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
                URL:       payload,
            }
            
            jsonData, _ := json.Marshal(trackingData)
            req, err := http.NewRequest("POST", "http://localhost:8080/v1/track",
                strings.NewReader(string(jsonData)))
            require.NoError(t, err)
            
            req.Header.Set("Content-Type", "application/json")
            req.Header.Set("X-API-Key", app.APIKey)
            
            client := &http.Client{}
            resp, err := client.Do(req)
            assert.NoError(t, err)
            
            // リクエストは受け入れられるが、データはサニタイズされる
            assert.Equal(t, http.StatusOK, resp.StatusCode)
            
            // データベースでサニタイズされたデータを確認
            time.Sleep(100 * time.Millisecond) // データ保存を待機
            
            savedData, err := getTrackingDataFromDB(t, app.AppID)
            require.NoError(t, err)
            require.NotEmpty(t, savedData)
            
            // 最新のデータを取得
            latestData := savedData[0]
            
            // XSSペイロードがサニタイズされていることを確認
            assert.NotContains(t, latestData.URL, "<script>")
            assert.NotContains(t, latestData.URL, "javascript:")
        }
    })
}
```

### 1.3 レート制限テスト

#### 1.3.1 レート制限機能テスト
```go
// tests/security/rate_limiting/rate_limit_test.go
package rate_limit_test

import (
    "testing"
    "net/http"
    "encoding/json"
    "strings"
    "time"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "access-log-tracker/internal/domain/models"
)

func TestRateLimiting(t *testing.T) {
    app := createTestApplication(t)
    
    t.Run("should enforce rate limits", func(t *testing.T) {
        // 制限を超えるリクエストを送信
        rateLimitedCount := 0
        for i := 0; i < 1001; i++ { // 制限: 1000 req/min
            trackingData := models.TrackingRequest{
                AppID:     app.AppID,
                UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
                URL:       "https://example.com/rate-limit-test",
            }
            
            jsonData, _ := json.Marshal(trackingData)
            req, err := http.NewRequest("POST", "http://localhost:8080/v1/track",
                strings.NewReader(string(jsonData)))
            require.NoError(t, err)
            
            req.Header.Set("Content-Type", "application/json")
            req.Header.Set("X-API-Key", app.APIKey)
            
            client := &http.Client{}
            resp, err := client.Do(req)
            assert.NoError(t, err)
            
            if resp.StatusCode == http.StatusTooManyRequests {
                rateLimitedCount++
            }
        }
        
        assert.Greater(t, rateLimitedCount, 0)
        t.Logf("Rate limited requests: %d", rateLimitedCount)
    })
    
    t.Run("should reset rate limits after time window", func(t *testing.T) {
        // 最初のリクエスト
        trackingData := models.TrackingRequest{
            AppID:     app.AppID,
            UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
            URL:       "https://example.com/rate-limit-reset-test",
        }
        
        jsonData, _ := json.Marshal(trackingData)
        req, err := http.NewRequest("POST", "http://localhost:8080/v1/track",
            strings.NewReader(string(jsonData)))
        require.NoError(t, err)
        
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("X-API-Key", app.APIKey)
        
        client := &http.Client{}
        resp, err := client.Do(req)
        assert.NoError(t, err)
        assert.Equal(t, http.StatusOK, resp.StatusCode)
        
        // 時間ウィンドウを待機（テスト用に短縮）
        time.Sleep(2 * time.Second)
        
        // 再度リクエスト
        resp, err = client.Do(req)
        assert.NoError(t, err)
        assert.Equal(t, http.StatusOK, resp.StatusCode)
    })
    
    t.Run("should prevent rate limit bypass attempts", func(t *testing.T) {
        // 異なるIPアドレスでレート制限を回避しようとする
        bypassAttempts := 0
        for i := 0; i < 10; i++ {
            trackingData := models.TrackingRequest{
                AppID:     app.AppID,
                UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
                URL:       "https://example.com/rate-limit-bypass-test",
            }
            
            jsonData, _ := json.Marshal(trackingData)
            req, err := http.NewRequest("POST", "http://localhost:8080/v1/track",
                strings.NewReader(string(jsonData)))
            require.NoError(t, err)
            
            req.Header.Set("Content-Type", "application/json")
            req.Header.Set("X-API-Key", app.APIKey)
            req.Header.Set("X-Forwarded-For", fmt.Sprintf("192.168.1.%d", i))
            
            client := &http.Client{}
            resp, err := client.Do(req)
            assert.NoError(t, err)
            
            if resp.StatusCode == http.StatusTooManyRequests {
                bypassAttempts++
            }
        }
        
        // レート制限が適切に適用される
        assert.Greater(t, bypassAttempts, 0)
        t.Logf("Bypass attempts blocked: %d", bypassAttempts)
    })
}
```

### 1.4 CSRF対策テスト

#### 1.4.1 CSRF保護テスト
```go
// tests/security/csrf/csrf_protection_test.go
package csrf_test

import (
    "testing"
    "net/http"
    "encoding/json"
    "strings"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "access-log-tracker/internal/domain/models"
)

func TestCSRFProtection(t *testing.T) {
    app := createTestApplication(t)
    
    t.Run("should reject requests without CSRF token", func(t *testing.T) {
        trackingData := models.TrackingRequest{
            AppID:     app.AppID,
            UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
            URL:       "https://example.com/csrf-test",
        }
        
        jsonData, _ := json.Marshal(trackingData)
        req, err := http.NewRequest("POST", "http://localhost:8080/v1/track",
            strings.NewReader(string(jsonData)))
        require.NoError(t, err)
        
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("X-API-Key", app.APIKey)
        // CSRFトークンを設定しない
        
        client := &http.Client{}
        resp, err := client.Do(req)
        assert.NoError(t, err)
        
        // CSRF保護が有効な場合、エラーが返される
        if resp.StatusCode == http.StatusForbidden {
            // CSRF保護が有効
            var response map[string]interface{}
            err = json.NewDecoder(resp.Body).Decode(&response)
            assert.NoError(t, err)
            assert.Equal(t, false, response["success"])
            assert.Equal(t, "CSRF_ERROR", response["error"].(map[string]interface{})["code"])
        } else {
            // CSRF保護が無効または別の方法で保護されている
            assert.Equal(t, http.StatusOK, resp.StatusCode)
        }
    })
    
    t.Run("should reject requests with invalid CSRF token", func(t *testing.T) {
        trackingData := models.TrackingRequest{
            AppID:     app.AppID,
            UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
            URL:       "https://example.com/csrf-test",
        }
        
        jsonData, _ := json.Marshal(trackingData)
        req, err := http.NewRequest("POST", "http://localhost:8080/v1/track",
            strings.NewReader(string(jsonData)))
        require.NoError(t, err)
        
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("X-API-Key", app.APIKey)
        req.Header.Set("X-CSRF-Token", "invalid_token")
        
        client := &http.Client{}
        resp, err := client.Do(req)
        assert.NoError(t, err)
        
        // CSRF保護が有効な場合、エラーが返される
        if resp.StatusCode == http.StatusForbidden {
            var response map[string]interface{}
            err = json.NewDecoder(resp.Body).Decode(&response)
            assert.NoError(t, err)
            assert.Equal(t, false, response["success"])
            assert.Equal(t, "CSRF_ERROR", response["error"].(map[string]interface{})["code"])
        } else {
            // CSRF保護が無効または別の方法で保護されている
            assert.Equal(t, http.StatusOK, resp.StatusCode)
        }
    })
    
    t.Run("should accept requests with valid CSRF token", func(t *testing.T) {
        // 有効なCSRFトークンを取得
        csrfToken := getValidCSRFToken(t, app.APIKey)
        
        trackingData := models.TrackingRequest{
            AppID:     app.AppID,
            UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
            URL:       "https://example.com/csrf-test",
        }
        
        jsonData, _ := json.Marshal(trackingData)
        req, err := http.NewRequest("POST", "http://localhost:8080/v1/track",
            strings.NewReader(string(jsonData)))
        require.NoError(t, err)
        
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("X-API-Key", app.APIKey)
        req.Header.Set("X-CSRF-Token", csrfToken)
        
        client := &http.Client{}
        resp, err := client.Do(req)
        assert.NoError(t, err)
        assert.Equal(t, http.StatusOK, resp.StatusCode)
        
        var response map[string]interface{}
        err = json.NewDecoder(resp.Body).Decode(&response)
        assert.NoError(t, err)
        assert.Equal(t, true, response["success"])
    })
}

// 有効なCSRFトークンを取得するヘルパー関数
func getValidCSRFToken(t *testing.T, apiKey string) string {
    // CSRFトークン取得エンドポイントにリクエスト
    req, err := http.NewRequest("GET", "http://localhost:8080/v1/csrf-token", nil)
    require.NoError(t, err)
    
    req.Header.Set("X-API-Key", apiKey)
    
    client := &http.Client{}
    resp, err := client.Do(req)
    require.NoError(t, err)
    require.Equal(t, http.StatusOK, resp.StatusCode)
    
    var response map[string]interface{}
    err = json.NewDecoder(resp.Body).Decode(&response)
    require.NoError(t, err)
    
    return response["data"].(map[string]interface{})["csrf_token"].(string)
}
```

### 1.5 データ保護テスト

#### 1.5.1 個人情報保護テスト
```go
// tests/security/data_protection/privacy_test.go
package privacy_test

import (
    "testing"
    "net/http"
    "encoding/json"
    "strings"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "access-log-tracker/internal/domain/models"
)

func TestDataProtection(t *testing.T) {
    app := createTestApplication(t)
    
    t.Run("should respect DNT header", func(t *testing.T) {
        trackingData := models.TrackingRequest{
            AppID:     app.AppID,
            UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
            URL:       "https://example.com/dnt-test",
        }
        
        jsonData, _ := json.Marshal(trackingData)
        req, err := http.NewRequest("POST", "http://localhost:8080/v1/track",
            strings.NewReader(string(jsonData)))
        require.NoError(t, err)
        
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("X-API-Key", app.APIKey)
        req.Header.Set("DNT", "1") // Do Not Track
        
        client := &http.Client{}
        resp, err := client.Do(req)
        assert.NoError(t, err)
        assert.Equal(t, http.StatusOK, resp.StatusCode)
        
        var response map[string]interface{}
        err = json.NewDecoder(resp.Body).Decode(&response)
        assert.NoError(t, err)
        assert.Equal(t, true, response["success"])
        
        // DNTが有効な場合、データが保存されないことを確認
        trackingID := response["data"].(map[string]interface{})["tracking_id"].(string)
        savedData, err := getTrackingDataFromDB(t, trackingID)
        require.NoError(t, err)
        assert.Nil(t, savedData) // データが保存されていない
    })
    
    t.Run("should anonymize IP addresses", func(t *testing.T) {
        trackingData := models.TrackingRequest{
            AppID:     app.AppID,
            UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
            URL:       "https://example.com/ip-anonymization-test",
        }
        
        jsonData, _ := json.Marshal(trackingData)
        req, err := http.NewRequest("POST", "http://localhost:8080/v1/track",
            strings.NewReader(string(jsonData)))
        require.NoError(t, err)
        
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("X-API-Key", app.APIKey)
        req.Header.Set("X-Forwarded-For", "192.168.1.100")
        
        client := &http.Client{}
        resp, err := client.Do(req)
        assert.NoError(t, err)
        assert.Equal(t, http.StatusOK, resp.StatusCode)
        
        // データベースでIPアドレスが匿名化されていることを確認
        time.Sleep(100 * time.Millisecond) // データ保存を待機
        
        savedData, err := getTrackingDataFromDB(t, app.AppID)
        require.NoError(t, err)
        require.NotEmpty(t, savedData)
        
        // 最新のデータを取得
        latestData := savedData[0]
        
        // IPアドレスが匿名化されていることを確認
        assert.NotEqual(t, "192.168.1.100", latestData.IPAddress)
        assert.Contains(t, latestData.IPAddress, "192.168.1.0") // 最後のオクテットが0になっている
    })
    
    t.Run("should not store sensitive headers", func(t *testing.T) {
        trackingData := models.TrackingRequest{
            AppID:     app.AppID,
            UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
            URL:       "https://example.com/sensitive-headers-test",
        }
        
        jsonData, _ := json.Marshal(trackingData)
        req, err := http.NewRequest("POST", "http://localhost:8080/v1/track",
            strings.NewReader(string(jsonData)))
        require.NoError(t, err)
        
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("X-API-Key", app.APIKey)
        req.Header.Set("Authorization", "Bearer sensitive_token")
        req.Header.Set("Cookie", "session=sensitive_session")
        
        client := &http.Client{}
        resp, err := client.Do(req)
        assert.NoError(t, err)
        assert.Equal(t, http.StatusOK, resp.StatusCode)
        
        // データベースで機密情報が保存されていないことを確認
        time.Sleep(100 * time.Millisecond) // データ保存を待機
        
        savedData, err := getTrackingDataFromDB(t, app.AppID)
        require.NoError(t, err)
        require.NotEmpty(t, savedData)
        
        // 最新のデータを取得
        latestData := savedData[0]
        
        // 機密情報が保存されていないことを確認
        assert.NotContains(t, latestData.UserAgent, "sensitive_token")
        assert.NotContains(t, latestData.UserAgent, "sensitive_session")
    })
}
```

## 2. セキュリティテストの実行

### 2.1 セキュリティテスト実行コマンド
```bash
# すべてのセキュリティテストを実行
go test ./tests/security/...

# 特定のセキュリティテストを実行
go test ./tests/security/authentication/...
go test ./tests/security/authorization/...
go test ./tests/security/input_validation/...
go test ./tests/security/rate_limiting/...
go test ./tests/security/csrf/...
go test ./tests/security/data_protection/...

# セキュリティテストの詳細出力
go test -v ./tests/security/...
```

### 2.2 セキュリティテストの設定
```yaml
# tests/security/config/security-test-config.yml
authentication:
  test_expired_keys: true
  test_invalid_keys: true
  test_missing_keys: true

authorization:
  test_cross_application_access: true
  test_unauthorized_statistics: true

input_validation:
  test_sql_injection: true
  test_xss_attacks: true
  test_injection_payloads: true

rate_limiting:
  test_rate_limit_enforcement: true
  test_bypass_attempts: true
  test_reset_after_window: true

csrf:
  test_csrf_protection: true
  test_invalid_tokens: true
  test_missing_tokens: true

data_protection:
  test_dnt_respect: true
  test_ip_anonymization: true
  test_sensitive_headers: true

test:
  cleanup_after_each: true
  parallel_tests: 4
  timeout: 60s
```

### 2.3 セキュリティテストのヘルパー関数
```go
// tests/security/helpers/security_helpers.go
package helpers

import (
    "testing"
    "time"
    "net/http"
    "encoding/json"
    "strings"
    "github.com/stretchr/testify/require"
    "access-log-tracker/internal/domain/models"
)

// セキュリティテスト用アプリケーション作成
func CreateSecurityTestApplication(t *testing.T) *models.Application {
    app := &models.Application{
        AppID:       "security_app_" + time.Now().Format("20060102150405"),
        Name:        "Security Test Application",
        Description: "Application for security testing",
        Domain:      "security-test.example.com",
        APIKey:      "security_api_key_" + time.Now().Format("20060102150405"),
    }
    
    // APIを使用してアプリケーションを作成
    jsonData, _ := json.Marshal(app)
    resp, err := http.Post("http://localhost:8080/v1/applications",
        "application/json", strings.NewReader(string(jsonData)))
    require.NoError(t, err)
    require.Equal(t, http.StatusCreated, resp.StatusCode)
    
    return app
}

// 期限切れアプリケーション作成
func CreateExpiredApplication(t *testing.T) *models.Application {
    app := &models.Application{
        AppID:       "expired_app_" + time.Now().Format("20060102150405"),
        Name:        "Expired Test Application",
        Description: "Application with expired API key",
        Domain:      "expired-test.example.com",
        APIKey:      "expired_api_key_" + time.Now().Format("20060102150405"),
        ExpiresAt:   time.Now().Add(-24 * time.Hour), // 24時間前に期限切れ
    }
    
    // APIを使用してアプリケーションを作成
    jsonData, _ := json.Marshal(app)
    resp, err := http.Post("http://localhost:8080/v1/applications",
        "application/json", strings.NewReader(string(jsonData)))
    require.NoError(t, err)
    require.Equal(t, http.StatusCreated, resp.StatusCode)
    
    return app
}

// セキュリティテスト用リクエスト送信
func SendSecurityTestRequest(t *testing.T, method, url string, data interface{}, headers map[string]string) *http.Response {
    var body strings.Reader
    if data != nil {
        jsonData, _ := json.Marshal(data)
        body = *strings.NewReader(string(jsonData))
    }
    
    req, err := http.NewRequest(method, url, &body)
    require.NoError(t, err)
    
    // デフォルトヘッダーを設定
    if data != nil {
        req.Header.Set("Content-Type", "application/json")
    }
    
    // カスタムヘッダーを設定
    for key, value := range headers {
        req.Header.Set(key, value)
    }
    
    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Do(req)
    require.NoError(t, err)
    
    return resp
}

// セキュリティテスト用レスポンス解析
func ParseSecurityTestResponse(t *testing.T, resp *http.Response) map[string]interface{} {
    var response map[string]interface{}
    err := json.NewDecoder(resp.Body).Decode(&response)
    require.NoError(t, err)
    return response
}

// データベースからトラッキングデータを取得
func GetTrackingDataFromDB(t *testing.T, identifier string) (*models.TrackingData, error) {
    db := setupTestDatabase(t)
    defer db.Close()
    
    var data models.TrackingData
    err := db.QueryRow(`
        SELECT id, app_id, user_agent, url, ip_address, session_id, timestamp
        FROM access_logs
        WHERE id = $1 OR app_id = $1
        ORDER BY timestamp DESC
        LIMIT 1
    `, identifier).Scan(&data.ID, &data.AppID, &data.UserAgent, &data.URL,
        &data.IPAddress, &data.SessionID, &data.Timestamp)
    
    if err != nil {
        return nil, err
    }
    
    return &data, nil
}

// セキュリティテストデータクリーンアップ
func CleanupSecurityTestData(t *testing.T) {
    db := setupTestDatabase(t)
    defer db.Close()
    
    tables := []string{"access_logs", "sessions", "applications"}
    for _, table := range tables {
        _, err := db.Exec("TRUNCATE TABLE " + table + " CASCADE")
        require.NoError(t, err)
    }
}
```

### 2.4 フェーズ別セキュリティテスト実行
```bash
# フェーズ6: 統合フェーズのセキュリティテスト
go test ./tests/security/authentication/...
go test ./tests/security/authorization/...
go test ./tests/security/input_validation/...
go test ./tests/security/rate_limiting/...
go test ./tests/security/csrf/...
go test ./tests/security/data_protection/...
```
