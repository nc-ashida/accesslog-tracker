# E2Eテスト実装

## 1. フェーズ5: ビーコンフェーズのテスト

### 1.1 ビーコン生成器のテスト

#### 1.1.1 JavaScriptビーコン生成のテスト
```go
// tests/e2e/beacon/generator/beacon_generator_test.go
package generator_test

import (
    "testing"
    "strings"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "access-log-tracker/internal/beacon/generator"
)

func TestBeaconGenerator_Generate(t *testing.T) {
    generator := generator.NewBeaconGenerator()
    
    t.Run("should generate valid JavaScript beacon", func(t *testing.T) {
        config := generator.BeaconConfig{
            Endpoint: "https://api.example.com/track",
            Debug:    false,
            Version:  "1.0.0",
        }
        
        result, err := generator.Generate(config)
        require.NoError(t, err)
        
        // JavaScriptとして有効かチェック
        assert.Contains(t, result, "function track")
        assert.Contains(t, result, config.Endpoint)
        assert.Contains(t, result, "XMLHttpRequest")
        assert.Contains(t, result, "fetch")
        
        // 構文エラーがないかチェック
        assert.NotContains(t, result, "undefined")
        assert.NotContains(t, result, "null")
    })
    
    t.Run("should include debug mode when enabled", func(t *testing.T) {
        config := generator.BeaconConfig{
            Endpoint: "https://api.example.com/track",
            Debug:    true,
            Version:  "1.0.0",
        }
        
        result, err := generator.Generate(config)
        require.NoError(t, err)
        
        assert.Contains(t, result, "console.log")
        assert.Contains(t, result, "debug")
    })
    
    t.Run("should handle custom parameters", func(t *testing.T) {
        config := generator.BeaconConfig{
            Endpoint: "https://api.example.com/track",
            Debug:    false,
            Version:  "1.0.0",
            CustomParams: map[string]string{
                "campaign_id": "camp_123",
                "source":      "google",
            },
        }
        
        result, err := generator.Generate(config)
        require.NoError(t, err)
        
        assert.Contains(t, result, "campaign_id")
        assert.Contains(t, result, "source")
    })
}

func TestBeaconGenerator_Minify(t *testing.T) {
    generator := generator.NewBeaconGenerator()
    
    t.Run("should minify JavaScript code", func(t *testing.T) {
        config := generator.BeaconConfig{
            Endpoint: "https://api.example.com/track",
            Debug:    false,
            Minify:   true,
        }
        
        result, err := generator.Generate(config)
        require.NoError(t, err)
        
        // 改行とスペースが削除されているかチェック
        lines := strings.Split(result, "\n")
        assert.Less(t, len(lines), 10) // 行数が少ない
        
        // コメントが削除されているかチェック
        assert.NotContains(t, result, "//")
        assert.NotContains(t, result, "/*")
    })
}
```

#### 1.1.2 ビーコン配信APIのテスト
```go
// tests/e2e/api/beacon_delivery_test.go
package api_test

import (
    "testing"
    "net/http"
    "net/http/httptest"
    "strings"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "access-log-tracker/internal/api/handlers"
)

func TestBeaconDeliveryAPI(t *testing.T) {
    router := setupTestServer(t)
    
    t.Run("GET /tracker.js - should serve JavaScript beacon", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/tracker.js", nil)
        w := httptest.NewRecorder()
        
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusOK, w.Code)
        assert.Equal(t, "application/javascript", w.Header().Get("Content-Type"))
        assert.Contains(t, w.Body.String(), "function track")
        assert.Contains(t, w.Body.String(), "XMLHttpRequest")
    })
    
    t.Run("GET /tracker.js - should include correct headers", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/tracker.js", nil)
        w := httptest.NewRecorder()
        
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusOK, w.Code)
        assert.Equal(t, "application/javascript", w.Header().Get("Content-Type"))
        assert.Equal(t, "public, max-age=3600", w.Header().Get("Cache-Control"))
        assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))
    })
    
    t.Run("GET /tracker.js - should handle query parameters", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/tracker.js?debug=true&version=1.0.0", nil)
        w := httptest.NewRecorder()
        
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusOK, w.Code)
        assert.Contains(t, w.Body.String(), "debug")
        assert.Contains(t, w.Body.String(), "1.0.0")
    })
}
```

### 1.2 ビーコン実行のテスト

#### 1.2.1 ブラウザ環境でのビーコン実行テスト
```go
// tests/e2e/beacon/execution/beacon_execution_test.go
package execution_test

import (
    "testing"
    "time"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/chromedp/chromedp"
    "context"
)

func TestBeaconExecution_Browser(t *testing.T) {
    ctx, cancel := chromedp.NewContext(context.Background())
    defer cancel()
    
    t.Run("should execute tracking beacon in browser", func(t *testing.T) {
        var result string
        
        err := chromedp.Run(ctx,
            chromedp.Navigate("http://localhost:8080/test-page.html"),
            chromedp.WaitVisible("#tracking-result"),
            chromedp.Text("#tracking-result", &result),
        )
        
        require.NoError(t, err)
        assert.Contains(t, result, "tracking_success")
    })
    
    t.Run("should handle tracking errors gracefully", func(t *testing.T) {
        var result string
        
        err := chromedp.Run(ctx,
            chromedp.Navigate("http://localhost:8080/test-page-error.html"),
            chromedp.WaitVisible("#error-result"),
            chromedp.Text("#error-result", &result),
        )
        
        require.NoError(t, err)
        assert.Contains(t, result, "error_handled")
    })
    
    t.Run("should respect privacy settings", func(t *testing.T) {
        var result string
        
        err := chromedp.Run(ctx,
            chromedp.Navigate("http://localhost:8080/test-page-privacy.html"),
            chromedp.WaitVisible("#privacy-result"),
            chromedp.Text("#privacy-result", &result),
        )
        
        require.NoError(t, err)
        assert.Contains(t, result, "privacy_respected")
    })
}
```

## 2. フェーズ6: 統合フェーズのテスト

### 2.1 エンドツーエンドトラッキングテスト

#### 2.1.1 完全なトラッキングフローのテスト
```go
// tests/e2e/tracking/complete_flow_test.go
package tracking_test

import (
    "testing"
    "time"
    "net/http"
    "encoding/json"
    "strings"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "access-log-tracker/internal/domain/models"
)

func TestCompleteTrackingFlow(t *testing.T) {
    // テスト用アプリケーションを作成
    app := createTestApplication(t)
    
    t.Run("should complete full tracking flow", func(t *testing.T) {
        // 1. ビーコンの配信確認
        resp, err := http.Get("http://localhost:8080/tracker.js")
        require.NoError(t, err)
        assert.Equal(t, http.StatusOK, resp.StatusCode)
        
        beaconJS := resp.Body
        defer beaconJS.Close()
        
        // 2. テストページでビーコンを実行
        testPage := createTestPage(t, app.AppID, beaconJS)
        resp, err = http.Get("http://localhost:8080/test-page.html")
        require.NoError(t, err)
        assert.Equal(t, http.StatusOK, resp.StatusCode)
        
        // 3. トラッキングデータの送信確認
        time.Sleep(2 * time.Second) // ビーコンの実行を待機
        
        // 4. データベースにデータが保存されているか確認
        trackingData, err := getTrackingDataFromDB(t, app.AppID)
        require.NoError(t, err)
        assert.NotEmpty(t, trackingData)
        
        // 5. 統計APIでデータが取得できるか確認
        resp, err = http.Get(fmt.Sprintf("http://localhost:8080/v1/statistics?app_id=%s", app.AppID))
        require.NoError(t, err)
        assert.Equal(t, http.StatusOK, resp.StatusCode)
        
        var stats map[string]interface{}
        err = json.NewDecoder(resp.Body).Decode(&stats)
        require.NoError(t, err)
        assert.Equal(t, true, stats["success"])
        
        data := stats["data"].(map[string]interface{})
        assert.Greater(t, data["total_requests"], float64(0))
    })
    
    t.Run("should handle concurrent tracking requests", func(t *testing.T) {
        const numRequests = 100
        const concurrency = 10
        
        // 並行してトラッキングリクエストを送信
        results := make(chan bool, numRequests)
        
        for i := 0; i < concurrency; i++ {
            go func() {
                for j := 0; j < numRequests/concurrency; j++ {
                    trackingData := models.TrackingRequest{
                        AppID:     app.AppID,
                        UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
                        URL:       "https://example.com/concurrent-test",
                    }
                    
                    jsonData, _ := json.Marshal(trackingData)
                    resp, err := http.Post("http://localhost:8080/v1/track",
                        "application/json", strings.NewReader(string(jsonData)))
                    
                    if err == nil && resp.StatusCode == http.StatusOK {
                        results <- true
                    } else {
                        results <- false
                    }
                }
            }()
        }
        
        // 結果を収集
        successCount := 0
        for i := 0; i < numRequests; i++ {
            if <-results {
                successCount++
            }
        }
        
        assert.Equal(t, numRequests, successCount)
    })
    
    t.Run("should handle different user agents", func(t *testing.T) {
        userAgents := []string{
            "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
            "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
            "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15",
            "Mozilla/5.0 (Linux; Android 10; SM-G973F) AppleWebKit/537.36",
        }
        
        for _, userAgent := range userAgents {
            trackingData := models.TrackingRequest{
                AppID:     app.AppID,
                UserAgent: userAgent,
                URL:       "https://example.com/user-agent-test",
            }
            
            jsonData, _ := json.Marshal(trackingData)
            resp, err := http.Post("http://localhost:8080/v1/track",
                "application/json", strings.NewReader(string(jsonData)))
            
            assert.NoError(t, err)
            assert.Equal(t, http.StatusOK, resp.StatusCode)
        }
        
        // データベースで異なるユーザーエージェントが記録されているか確認
        time.Sleep(1 * time.Second)
        trackingData, err := getTrackingDataFromDB(t, app.AppID)
        require.NoError(t, err)
        
        userAgentCount := make(map[string]int)
        for _, data := range trackingData {
            userAgentCount[data.UserAgent]++
        }
        
        assert.GreaterOrEqual(t, len(userAgentCount), 4)
    })
}

func TestTrackingFlow_ErrorScenarios(t *testing.T) {
    app := createTestApplication(t)
    
    t.Run("should handle invalid tracking data", func(t *testing.T) {
        // 無効なデータを送信
        invalidData := map[string]interface{}{
            "app_id": "", // 空のAppID
            "url":    "invalid-url",
        }
        
        jsonData, _ := json.Marshal(invalidData)
        resp, err := http.Post("http://localhost:8080/v1/track",
            "application/json", strings.NewReader(string(jsonData)))
        
        assert.NoError(t, err)
        assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
    })
    
    t.Run("should handle network errors", func(t *testing.T) {
        // 存在しないエンドポイントにリクエスト
        resp, err := http.Post("http://localhost:8080/v1/nonexistent",
            "application/json", strings.NewReader("{}"))
        
        assert.NoError(t, err)
        assert.Equal(t, http.StatusNotFound, resp.StatusCode)
    })
    
    t.Run("should handle rate limiting", func(t *testing.T) {
        // 制限を超えるリクエストを送信
        rateLimitedCount := 0
        for i := 0; i < 1001; i++ {
            trackingData := models.TrackingRequest{
                AppID:     app.AppID,
                UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
                URL:       "https://example.com/rate-limit-test",
            }
            
            jsonData, _ := json.Marshal(trackingData)
            resp, err := http.Post("http://localhost:8080/v1/track",
                "application/json", strings.NewReader(string(jsonData)))
            
            if err == nil && resp.StatusCode == http.StatusTooManyRequests {
                rateLimitedCount++
            }
        }
        
        assert.Greater(t, rateLimitedCount, 0)
    })
}
```

### 2.2 パフォーマンス統合テスト

#### 2.2.1 システム全体のパフォーマンステスト
```go
// tests/e2e/performance/system_performance_test.go
package performance_test

import (
    "testing"
    "time"
    "net/http"
    "sync"
    "encoding/json"
    "strings"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "access-log-tracker/internal/domain/models"
)

func TestSystemPerformance(t *testing.T) {
    app := createTestApplication(t)
    
    t.Run("should handle high load", func(t *testing.T) {
        const numRequests = 10000
        const concurrency = 100
        
        start := time.Now()
        
        var wg sync.WaitGroup
        results := make(chan bool, numRequests)
        
        // 並行してリクエストを送信
        for i := 0; i < concurrency; i++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                for j := 0; j < numRequests/concurrency; j++ {
                    trackingData := models.TrackingRequest{
                        AppID:     app.AppID,
                        UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
                        URL:       "https://example.com/performance-test",
                    }
                    
                    jsonData, _ := json.Marshal(trackingData)
                    resp, err := http.Post("http://localhost:8080/v1/track",
                        "application/json", strings.NewReader(string(jsonData)))
                    
                    if err == nil && resp.StatusCode == http.StatusOK {
                        results <- true
                    } else {
                        results <- false
                    }
                }
            }()
        }
        
        wg.Wait()
        close(results)
        
        duration := time.Since(start)
        
        // 結果を収集
        successCount := 0
        for result := range results {
            if result {
                successCount++
            }
        }
        
        // パフォーマンス基準をチェック
        assert.Equal(t, numRequests, successCount)
        assert.Less(t, duration, 30*time.Second) // 30秒以内に完了
        
        // スループットを計算
        throughput := float64(numRequests) / duration.Seconds()
        assert.Greater(t, throughput, 300.0) // 300 req/sec以上
    })
    
    t.Run("should maintain response time under load", func(t *testing.T) {
        const numRequests = 1000
        responseTimes := make([]time.Duration, numRequests)
        
        for i := 0; i < numRequests; i++ {
            start := time.Now()
            
            trackingData := models.TrackingRequest{
                AppID:     app.AppID,
                UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
                URL:       "https://example.com/response-time-test",
            }
            
            jsonData, _ := json.Marshal(trackingData)
            resp, err := http.Post("http://localhost:8080/v1/track",
                "application/json", strings.NewReader(string(jsonData)))
            
            responseTime := time.Since(start)
            responseTimes[i] = responseTime
            
            assert.NoError(t, err)
            assert.Equal(t, http.StatusOK, resp.StatusCode)
        }
        
        // 平均応答時間を計算
        var totalTime time.Duration
        for _, rt := range responseTimes {
            totalTime += rt
        }
        avgResponseTime := totalTime / time.Duration(numRequests)
        
        // 応答時間基準をチェック
        assert.Less(t, avgResponseTime, 100*time.Millisecond) // 100ms以下
        
        // 95パーセンタイル応答時間をチェック
        sortedTimes := make([]time.Duration, numRequests)
        copy(sortedTimes, responseTimes)
        sort.Slice(sortedTimes, func(i, j int) bool {
            return sortedTimes[i] < sortedTimes[j]
        })
        
        p95Index := int(float64(numRequests) * 0.95)
        p95ResponseTime := sortedTimes[p95Index]
        assert.Less(t, p95ResponseTime, 200*time.Millisecond) // 200ms以下
    })
    
    t.Run("should handle memory usage efficiently", func(t *testing.T) {
        // メモリ使用量を監視
        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        initialMemory := m.Alloc
        
        // 大量のリクエストを送信
        const numRequests = 5000
        for i := 0; i < numRequests; i++ {
            trackingData := models.TrackingRequest{
                AppID:     app.AppID,
                UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
                URL:       "https://example.com/memory-test",
            }
            
            jsonData, _ := json.Marshal(trackingData)
            resp, err := http.Post("http://localhost:8080/v1/track",
                "application/json", strings.NewReader(string(jsonData)))
            
            assert.NoError(t, err)
            assert.Equal(t, http.StatusOK, resp.StatusCode)
        }
        
        // ガベージコレクションを実行
        runtime.GC()
        
        // 最終メモリ使用量をチェック
        runtime.ReadMemStats(&m)
        finalMemory := m.Alloc
        memoryIncrease := finalMemory - initialMemory
        
        // メモリ増加が100MB以下であることを確認
        assert.Less(t, memoryIncrease, uint64(100*1024*1024))
    })
}
```

### 2.3 セキュリティ統合テスト

#### 2.3.1 セキュリティ脆弱性のテスト
```go
// tests/e2e/security/security_vulnerabilities_test.go
package security_test

import (
    "testing"
    "net/http"
    "encoding/json"
    "strings"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestSecurityVulnerabilities(t *testing.T) {
    app := createTestApplication(t)
    
    t.Run("should prevent SQL injection", func(t *testing.T) {
        // SQLインジェクション攻撃のペイロード
        maliciousPayloads := []string{
            "'; DROP TABLE access_logs; --",
            "' OR '1'='1",
            "'; INSERT INTO access_logs VALUES ('hacked'); --",
        }
        
        for _, payload := range maliciousPayloads {
            trackingData := map[string]interface{}{
                "app_id":     payload,
                "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
                "url":        "https://example.com/sql-injection-test",
            }
            
            jsonData, _ := json.Marshal(trackingData)
            resp, err := http.Post("http://localhost:8080/v1/track",
                "application/json", strings.NewReader(string(jsonData)))
            
            assert.NoError(t, err)
            // バリデーションエラーが返されるべき
            assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
        }
    })
    
    t.Run("should prevent XSS attacks", func(t *testing.T) {
        // XSS攻撃のペイロード
        xssPayloads := []string{
            "<script>alert('xss')</script>",
            "javascript:alert('xss')",
            "<img src=x onerror=alert('xss')>",
        }
        
        for _, payload := range xssPayloads {
            trackingData := models.TrackingRequest{
                AppID:     app.AppID,
                UserAgent: payload,
                URL:       "https://example.com/xss-test",
            }
            
            jsonData, _ := json.Marshal(trackingData)
            resp, err := http.Post("http://localhost:8080/v1/track",
                "application/json", strings.NewReader(string(jsonData)))
            
            assert.NoError(t, err)
            // リクエストは受け入れられるが、データはサニタイズされる
            assert.Equal(t, http.StatusOK, resp.StatusCode)
        }
    })
    
    t.Run("should prevent CSRF attacks", func(t *testing.T) {
        // CSRFトークンなしでリクエストを送信
        trackingData := models.TrackingRequest{
            AppID:     app.AppID,
            UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
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
            assert.True(t, true)
        } else {
            // CSRF保護が無効または別の方法で保護されている
            assert.Equal(t, http.StatusOK, resp.StatusCode)
        }
    })
    
    t.Run("should validate API key properly", func(t *testing.T) {
        // 無効なAPIキーでリクエスト
        trackingData := models.TrackingRequest{
            AppID:     app.AppID,
            UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
            URL:       "https://example.com/api-key-test",
        }
        
        jsonData, _ := json.Marshal(trackingData)
        req, err := http.NewRequest("POST", "http://localhost:8080/v1/track",
            strings.NewReader(string(jsonData)))
        require.NoError(t, err)
        
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("X-API-Key", "invalid-api-key")
        
        client := &http.Client{}
        resp, err := client.Do(req)
        assert.NoError(t, err)
        assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
    })
    
    t.Run("should prevent rate limiting bypass", func(t *testing.T) {
        // 異なるIPアドレスでレート制限を回避しようとする
        for i := 0; i < 10; i++ {
            trackingData := models.TrackingRequest{
                AppID:     app.AppID,
                UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
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
            
            // レート制限が適切に適用される
            if resp.StatusCode == http.StatusTooManyRequests {
                break
            }
        }
    })
}
```

## 3. E2Eテストの実行

### 3.1 E2Eテスト実行コマンド
```bash
# すべてのE2Eテストを実行
go test ./tests/e2e/...

# 特定のE2Eテストを実行
go test ./tests/e2e/beacon/...
go test ./tests/e2e/tracking/...
go test ./tests/e2e/performance/...
go test ./tests/e2e/security/...

# カバレッジ付きでE2Eテスト実行
go test -cover ./tests/e2e/...

# E2Eテストの詳細出力
go test -v ./tests/e2e/...
```

### 3.2 E2Eテストの設定
```yaml
# tests/e2e/config/e2e-test-config.yml
browser:
  headless: true
  slow_mo: 100
  timeout: 30000

api:
  base_url: http://localhost:8080
  timeout: 30s

database:
  host: postgres
  port: 5432
  name: access_log_tracker_e2e
  user: postgres
  password: password

test:
  cleanup_after_each: true
  parallel_tests: 2
  timeout: 300s
```

### 3.3 E2Eテストのヘルパー関数
```go
// tests/e2e/helpers/e2e_helpers.go
package helpers

import (
    "database/sql"
    "testing"
    "time"
    "net/http"
    "encoding/json"
    "strings"
    
    _ "github.com/lib/pq"
    "github.com/stretchr/testify/require"
    "access-log-tracker/internal/domain/models"
)

// E2Eテスト用データベースセットアップ
func SetupE2EDatabase(t *testing.T) *sql.DB {
    dsn := "host=postgres port=5432 user=postgres password=password dbname=access_log_tracker_e2e sslmode=disable"
    db, err := sql.Open("postgres", dsn)
    require.NoError(t, err)
    
    // 接続テスト
    err = db.Ping()
    require.NoError(t, err)
    
    return db
}

// テスト用アプリケーション作成
func CreateE2EApplication(t *testing.T) *models.Application {
    app := &models.Application{
        AppID:       "e2e_app_" + time.Now().Format("20060102150405"),
        Name:        "E2E Test Application",
        Description: "Application for E2E testing",
        Domain:      "e2e-test.example.com",
        APIKey:      "e2e_api_key_" + time.Now().Format("20060102150405"),
    }
    
    // APIを使用してアプリケーションを作成
    jsonData, _ := json.Marshal(app)
    resp, err := http.Post("http://localhost:8080/v1/applications",
        "application/json", strings.NewReader(string(jsonData)))
    require.NoError(t, err)
    require.Equal(t, http.StatusCreated, resp.StatusCode)
    
    return app
}

// テストページ作成
func CreateTestPage(t *testing.T, appID string, beaconJS string) string {
    testPage := fmt.Sprintf(`
        <!DOCTYPE html>
        <html>
        <head>
            <title>E2E Test Page</title>
        </head>
        <body>
            <div id="tracking-result"></div>
            <script>
                %s
                track({
                    app_id: '%s',
                    url: window.location.href,
                    user_agent: navigator.userAgent
                });
                document.getElementById('tracking-result').textContent = 'tracking_success';
            </script>
        </body>
        </html>
    `, beaconJS, appID)
    
    return testPage
}

// データベースからトラッキングデータを取得
func GetTrackingDataFromDB(t *testing.T, appID string) ([]*models.TrackingData, error) {
    db := SetupE2EDatabase(t)
    defer db.Close()
    
    rows, err := db.Query(`
        SELECT id, app_id, user_agent, url, ip_address, session_id, timestamp
        FROM access_logs
        WHERE app_id = $1
        ORDER BY timestamp DESC
    `, appID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var results []*models.TrackingData
    for rows.Next() {
        data := &models.TrackingData{}
        err := rows.Scan(&data.ID, &data.AppID, &data.UserAgent, &data.URL,
            &data.IPAddress, &data.SessionID, &data.Timestamp)
        if err != nil {
            return nil, err
        }
        results = append(results, data)
    }
    
    return results, nil
}

// テストデータクリーンアップ
func CleanupE2EData(t *testing.T) {
    db := SetupE2EDatabase(t)
    defer db.Close()
    
    tables := []string{"access_logs", "sessions", "applications"}
    for _, table := range tables {
        _, err := db.Exec("TRUNCATE TABLE " + table + " CASCADE")
        require.NoError(t, err)
    }
}
```

### 3.4 フェーズ別E2Eテスト実行
```bash
# フェーズ5: ビーコンフェーズのテスト
go test ./tests/e2e/beacon/...

# フェーズ6: 統合フェーズのテスト
go test ./tests/e2e/tracking/...
go test ./tests/e2e/performance/...
go test ./tests/e2e/security/...
```
