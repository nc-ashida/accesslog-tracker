# パフォーマンステスト実装

## 1. フェーズ6: 統合フェーズのパフォーマンステスト

### 1.1 負荷テスト

#### 1.1.1 トラッキングAPIの負荷テスト
```go
// tests/performance/load/tracking_api_load_test.go
package load_test

import (
    "testing"
    "time"
    "net/http"
    "sync"
    "encoding/json"
    "strings"
    "runtime"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "access-log-tracker/internal/domain/models"
)

func TestTrackingAPILoad(t *testing.T) {
    app := createTestApplication(t)
    
    t.Run("should handle 1000 concurrent requests", func(t *testing.T) {
        const numRequests = 1000
        const concurrency = 100
        
        start := time.Now()
        
        var wg sync.WaitGroup
        results := make(chan bool, numRequests)
        errors := make(chan error, numRequests)
        
        // 並行してリクエストを送信
        for i := 0; i < concurrency; i++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                for j := 0; j < numRequests/concurrency; j++ {
                    trackingData := models.TrackingRequest{
                        AppID:     app.AppID,
                        UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
                        URL:       "https://example.com/load-test",
                    }
                    
                    jsonData, _ := json.Marshal(trackingData)
                    resp, err := http.Post("http://localhost:8080/v1/track",
                        "application/json", strings.NewReader(string(jsonData)))
                    
                    if err != nil {
                        errors <- err
                        results <- false
                    } else if resp.StatusCode == http.StatusOK {
                        results <- true
                    } else {
                        results <- false
                    }
                }
            }()
        }
        
        wg.Wait()
        close(results)
        close(errors)
        
        duration := time.Since(start)
        
        // 結果を収集
        successCount := 0
        errorCount := 0
        for result := range results {
            if result {
                successCount++
            }
        }
        for range errors {
            errorCount++
        }
        
        // パフォーマンス基準をチェック
        assert.Equal(t, numRequests, successCount)
        assert.Equal(t, 0, errorCount)
        assert.Less(t, duration, 30*time.Second) // 30秒以内に完了
        
        // スループットを計算
        throughput := float64(numRequests) / duration.Seconds()
        assert.Greater(t, throughput, 30.0) // 30 req/sec以上
        
        t.Logf("Load test completed: %d requests in %v (%.2f req/sec)", 
            numRequests, duration, throughput)
    })
    
    t.Run("should handle 10000 requests with memory monitoring", func(t *testing.T) {
        const numRequests = 10000
        
        // 初期メモリ使用量を記録
        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        initialMemory := m.Alloc
        initialHeap := m.HeapAlloc
        
        start := time.Now()
        
        // 大量のリクエストを送信
        for i := 0; i < numRequests; i++ {
            trackingData := models.TrackingRequest{
                AppID:     app.AppID,
                UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
                URL:       fmt.Sprintf("https://example.com/load-test-%d", i),
            }
            
            jsonData, _ := json.Marshal(trackingData)
            resp, err := http.Post("http://localhost:8080/v1/track",
                "application/json", strings.NewReader(string(jsonData)))
            
            assert.NoError(t, err)
            assert.Equal(t, http.StatusOK, resp.StatusCode)
        }
        
        duration := time.Since(start)
        
        // ガベージコレクションを実行
        runtime.GC()
        
        // 最終メモリ使用量をチェック
        runtime.ReadMemStats(&m)
        finalMemory := m.Alloc
        finalHeap := m.HeapAlloc
        
        memoryIncrease := finalMemory - initialMemory
        heapIncrease := finalHeap - initialHeap
        
        // メモリ使用量基準をチェック
        assert.Less(t, memoryIncrease, uint64(100*1024*1024)) // 100MB以下
        assert.Less(t, heapIncrease, uint64(50*1024*1024))    // 50MB以下
        
        // スループットを計算
        throughput := float64(numRequests) / duration.Seconds()
        assert.Greater(t, throughput, 300.0) // 300 req/sec以上
        
        t.Logf("Memory test completed: %d requests in %v (%.2f req/sec, memory: +%d bytes, heap: +%d bytes)", 
            numRequests, duration, throughput, memoryIncrease, heapIncrease)
    })
    
    t.Run("should maintain response time under sustained load", func(t *testing.T) {
        const numRequests = 5000
        const duration = 60 * time.Second // 60秒間の持続負荷
        
        responseTimes := make([]time.Duration, 0, numRequests)
        var responseTimesMutex sync.Mutex
        
        start := time.Now()
        endTime := start.Add(duration)
        
        var wg sync.WaitGroup
        for i := 0; i < 10; i++ { // 10個のゴルーチンで並行実行
            wg.Add(1)
            go func() {
                defer wg.Done()
                for time.Now().Before(endTime) {
                    requestStart := time.Now()
                    
                    trackingData := models.TrackingRequest{
                        AppID:     app.AppID,
                        UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
                        URL:       "https://example.com/sustained-load-test",
                    }
                    
                    jsonData, _ := json.Marshal(trackingData)
                    resp, err := http.Post("http://localhost:8080/v1/track",
                        "application/json", strings.NewReader(string(jsonData)))
                    
                    responseTime := time.Since(requestStart)
                    
                    if err == nil && resp.StatusCode == http.StatusOK {
                        responseTimesMutex.Lock()
                        responseTimes = append(responseTimes, responseTime)
                        responseTimesMutex.Unlock()
                    }
                    
                    time.Sleep(10 * time.Millisecond) // 10ms間隔
                }
            }()
        }
        
        wg.Wait()
        actualDuration := time.Since(start)
        
        // 応答時間の統計を計算
        if len(responseTimes) > 0 {
            var totalTime time.Duration
            for _, rt := range responseTimes {
                totalTime += rt
            }
            avgResponseTime := totalTime / time.Duration(len(responseTimes))
            
            // 95パーセンタイル応答時間を計算
            sortedTimes := make([]time.Duration, len(responseTimes))
            copy(sortedTimes, responseTimes)
            sort.Slice(sortedTimes, func(i, j int) bool {
                return sortedTimes[i] < sortedTimes[j]
            })
            
            p95Index := int(float64(len(sortedTimes)) * 0.95)
            p95ResponseTime := sortedTimes[p95Index]
            
            // 応答時間基準をチェック
            assert.Less(t, avgResponseTime, 100*time.Millisecond) // 平均100ms以下
            assert.Less(t, p95ResponseTime, 200*time.Millisecond) // 95パーセンタイル200ms以下
            
            t.Logf("Sustained load test completed: %d requests in %v (avg: %v, p95: %v)", 
                len(responseTimes), actualDuration, avgResponseTime, p95ResponseTime)
        }
    })
}
```

#### 1.1.2 統計APIの負荷テスト
```go
// tests/performance/load/statistics_api_load_test.go
package load_test

import (
    "testing"
    "time"
    "net/http"
    "sync"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestStatisticsAPILoad(t *testing.T) {
    app := createTestApplication(t)
    
    // 事前に大量のトラッキングデータを作成
    createBulkTrackingData(t, app.AppID, 10000)
    
    t.Run("should handle concurrent statistics requests", func(t *testing.T) {
        const numRequests = 100
        const concurrency = 20
        
        start := time.Now()
        
        var wg sync.WaitGroup
        results := make(chan bool, numRequests)
        
        // 並行して統計リクエストを送信
        for i := 0; i < concurrency; i++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                for j := 0; j < numRequests/concurrency; j++ {
                    url := fmt.Sprintf("http://localhost:8080/v1/statistics?app_id=%s&start_date=2024-01-01&end_date=2024-12-31", app.AppID)
                    req, err := http.NewRequest("GET", url, nil)
                    require.NoError(t, err)
                    req.Header.Set("X-API-Key", app.APIKey)
                    
                    client := &http.Client{Timeout: 30 * time.Second}
                    resp, err := client.Do(req)
                    
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
        assert.Less(t, duration, 60*time.Second) // 60秒以内に完了
        
        // スループットを計算
        throughput := float64(numRequests) / duration.Seconds()
        assert.Greater(t, throughput, 1.0) // 1 req/sec以上
        
        t.Logf("Statistics load test completed: %d requests in %v (%.2f req/sec)", 
            numRequests, duration, throughput)
    })
    
    t.Run("should handle large date range queries", func(t *testing.T) {
        // 1年間のデータをクエリ
        start := time.Now()
        
        url := fmt.Sprintf("http://localhost:8080/v1/statistics?app_id=%s&start_date=2023-01-01&end_date=2024-12-31", app.AppID)
        req, err := http.NewRequest("GET", url, nil)
        require.NoError(t, err)
        req.Header.Set("X-API-Key", app.APIKey)
        
        client := &http.Client{Timeout: 60 * time.Second}
        resp, err := client.Do(req)
        
        duration := time.Since(start)
        
        assert.NoError(t, err)
        assert.Equal(t, http.StatusOK, resp.StatusCode)
        assert.Less(t, duration, 10*time.Second) // 10秒以内に完了
        
        t.Logf("Large date range query completed in %v", duration)
    })
}

// 大量のトラッキングデータを作成するヘルパー関数
func createBulkTrackingData(t *testing.T, appID string, count int) {
    var wg sync.WaitGroup
    const batchSize = 100
    
    for i := 0; i < count; i += batchSize {
        wg.Add(1)
        go func(startIndex int) {
            defer wg.Done()
            
            for j := 0; j < batchSize && startIndex+j < count; j++ {
                trackingData := models.TrackingRequest{
                    AppID:     appID,
                    UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
                    URL:       fmt.Sprintf("https://example.com/bulk-test-%d", startIndex+j),
                }
                
                jsonData, _ := json.Marshal(trackingData)
                resp, err := http.Post("http://localhost:8080/v1/track",
                    "application/json", strings.NewReader(string(jsonData)))
                
                if err != nil || resp.StatusCode != http.StatusOK {
                    t.Logf("Failed to create tracking data: %v", err)
                }
            }
        }(i)
    }
    
    wg.Wait()
}
```

### 1.2 ストレステスト

#### 1.2.1 極限負荷テスト
```go
// tests/performance/stress/extreme_load_test.go
package stress_test

import (
    "testing"
    "time"
    "net/http"
    "sync"
    "encoding/json"
    "strings"
    "runtime"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "access-log-tracker/internal/domain/models"
)

func TestExtremeLoad(t *testing.T) {
    app := createTestApplication(t)
    
    t.Run("should handle burst traffic", func(t *testing.T) {
        const burstSize = 1000
        const burstDuration = 5 * time.Second
        
        start := time.Now()
        
        var wg sync.WaitGroup
        results := make(chan bool, burstSize)
        
        // バーストトラフィックをシミュレート
        for i := 0; i < burstSize; i++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                
                trackingData := models.TrackingRequest{
                    AppID:     app.AppID,
                    UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
                    URL:       "https://example.com/burst-test",
                }
                
                jsonData, _ := json.Marshal(trackingData)
                resp, err := http.Post("http://localhost:8080/v1/track",
                    "application/json", strings.NewReader(string(jsonData)))
                
                if err == nil && resp.StatusCode == http.StatusOK {
                    results <- true
                } else {
                    results <- false
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
        
        // バースト処理基準をチェック
        successRate := float64(successCount) / float64(burstSize)
        assert.Greater(t, successRate, 0.95) // 95%以上の成功率
        assert.Less(t, duration, burstDuration) // 指定時間内に完了
        
        t.Logf("Burst test completed: %d/%d successful (%.2f%%) in %v", 
            successCount, burstSize, successRate*100, duration)
    })
    
    t.Run("should handle memory pressure", func(t *testing.T) {
        const numRequests = 50000
        
        // メモリ使用量を監視
        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        initialMemory := m.Alloc
        initialHeap := m.HeapAlloc
        
        start := time.Now()
        
        // 大量のリクエストを送信してメモリ圧迫をシミュレート
        var wg sync.WaitGroup
        for i := 0; i < 10; i++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                for j := 0; j < numRequests/10; j++ {
                    trackingData := models.TrackingRequest{
                        AppID:     app.AppID,
                        UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
                        URL:       fmt.Sprintf("https://example.com/memory-pressure-test-%d", j),
                    }
                    
                    jsonData, _ := json.Marshal(trackingData)
                    resp, err := http.Post("http://localhost:8080/v1/track",
                        "application/json", strings.NewReader(string(jsonData)))
                    
                    if err != nil || resp.StatusCode != http.StatusOK {
                        t.Logf("Request failed: %v", err)
                    }
                }
            }()
        }
        
        wg.Wait()
        duration := time.Since(start)
        
        // ガベージコレクションを実行
        runtime.GC()
        
        // 最終メモリ使用量をチェック
        runtime.ReadMemStats(&m)
        finalMemory := m.Alloc
        finalHeap := m.HeapAlloc
        
        memoryIncrease := finalMemory - initialMemory
        heapIncrease := finalHeap - initialHeap
        
        // メモリ使用量基準をチェック
        assert.Less(t, memoryIncrease, uint64(200*1024*1024)) // 200MB以下
        assert.Less(t, heapIncrease, uint64(100*1024*1024))   // 100MB以下
        
        t.Logf("Memory pressure test completed: %d requests in %v (memory: +%d bytes, heap: +%d bytes)", 
            numRequests, duration, memoryIncrease, heapIncrease)
    })
    
    t.Run("should handle connection exhaustion", func(t *testing.T) {
        const numConnections = 1000
        
        // 大量の同時接続をシミュレート
        var wg sync.WaitGroup
        results := make(chan bool, numConnections)
        
        for i := 0; i < numConnections; i++ {
            wg.Add(1)
            go func() {
                defer wg.Done()
                
                // 各接続で複数のリクエストを送信
                client := &http.Client{Timeout: 30 * time.Second}
                
                for j := 0; j < 10; j++ {
                    trackingData := models.TrackingRequest{
                        AppID:     app.AppID,
                        UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
                        URL:       "https://example.com/connection-test",
                    }
                    
                    jsonData, _ := json.Marshal(trackingData)
                    req, err := http.NewRequest("POST", "http://localhost:8080/v1/track",
                        strings.NewReader(string(jsonData)))
                    require.NoError(t, err)
                    req.Header.Set("Content-Type", "application/json")
                    req.Header.Set("X-API-Key", app.APIKey)
                    
                    resp, err := client.Do(req)
                    
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
        
        // 結果を収集
        successCount := 0
        for result := range results {
            if result {
                successCount++
            }
        }
        
        // 接続処理基準をチェック
        successRate := float64(successCount) / float64(numConnections*10)
        assert.Greater(t, successRate, 0.90) // 90%以上の成功率
        
        t.Logf("Connection exhaustion test completed: %d/%d successful (%.2f%%)", 
            successCount, numConnections*10, successRate*100)
    })
}
```

### 1.3 スループットテスト

#### 1.3.1 最大スループットテスト
```go
// tests/performance/throughput/max_throughput_test.go
package throughput_test

import (
    "testing"
    "time"
    "net/http"
    "sync"
    "encoding/json"
    "strings"
    "runtime"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "access-log-tracker/internal/domain/models"
)

func TestMaxThroughput(t *testing.T) {
    app := createTestApplication(t)
    
    t.Run("should achieve maximum throughput", func(t *testing.T) {
        const testDuration = 30 * time.Second
        const targetThroughput = 1000 // 1000 req/sec
        
        start := time.Now()
        endTime := start.Add(testDuration)
        
        var wg sync.WaitGroup
        requestCount := int64(0)
        var requestCountMutex sync.Mutex
        
        // 目標スループットを維持するようにリクエストを送信
        for i := 0; i < runtime.NumCPU()*2; i++ { // CPUコア数の2倍のゴルーチン
            wg.Add(1)
            go func() {
                defer wg.Done()
                
                for time.Now().Before(endTime) {
                    trackingData := models.TrackingRequest{
                        AppID:     app.AppID,
                        UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
                        URL:       "https://example.com/throughput-test",
                    }
                    
                    jsonData, _ := json.Marshal(trackingData)
                    resp, err := http.Post("http://localhost:8080/v1/track",
                        "application/json", strings.NewReader(string(jsonData)))
                    
                    if err == nil && resp.StatusCode == http.StatusOK {
                        requestCountMutex.Lock()
                        requestCount++
                        requestCountMutex.Unlock()
                    }
                    
                    // 目標スループットに合わせて間隔を調整
                    time.Sleep(time.Duration(1000000/targetThroughput) * time.Microsecond)
                }
            }()
        }
        
        wg.Wait()
        actualDuration := time.Since(start)
        
        // スループットを計算
        actualThroughput := float64(requestCount) / actualDuration.Seconds()
        
        // スループット基準をチェック
        assert.Greater(t, actualThroughput, float64(targetThroughput)*0.8) // 目標の80%以上
        
        t.Logf("Throughput test completed: %.2f req/sec over %v", actualThroughput, actualDuration)
    })
    
    t.Run("should maintain consistent throughput", func(t *testing.T) {
        const testDuration = 60 * time.Second
        const interval = 10 * time.Second
        
        start := time.Now()
        endTime := start.Add(testDuration)
        
        throughputs := make([]float64, 0)
        var throughputsMutex sync.Mutex
        
        // 10秒間隔でスループットを測定
        for intervalStart := start; intervalStart.Before(endTime); intervalStart = intervalStart.Add(interval) {
            intervalEnd := intervalStart.Add(interval)
            if intervalEnd.After(endTime) {
                intervalEnd = endTime
            }
            
            var wg sync.WaitGroup
            requestCount := int64(0)
            var requestCountMutex sync.Mutex
            
            // この間隔でリクエストを送信
            for i := 0; i < 10; i++ {
                wg.Add(1)
                go func() {
                    defer wg.Done()
                    
                    for time.Now().Before(intervalEnd) {
                        trackingData := models.TrackingRequest{
                            AppID:     app.AppID,
                            UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
                            URL:       "https://example.com/consistent-throughput-test",
                        }
                        
                        jsonData, _ := json.Marshal(trackingData)
                        resp, err := http.Post("http://localhost:8080/v1/track",
                            "application/json", strings.NewReader(string(jsonData)))
                        
                        if err == nil && resp.StatusCode == http.StatusOK {
                            requestCountMutex.Lock()
                            requestCount++
                            requestCountMutex.Unlock()
                        }
                    }
                }()
            }
            
            wg.Wait()
            
            // この間隔のスループットを計算
            intervalDuration := intervalEnd.Sub(intervalStart)
            intervalThroughput := float64(requestCount) / intervalDuration.Seconds()
            
            throughputsMutex.Lock()
            throughputs = append(throughputs, intervalThroughput)
            throughputsMutex.Unlock()
        }
        
        // スループットの一貫性をチェック
        if len(throughputs) > 1 {
            var total float64
            for _, t := range throughputs {
                total += t
            }
            avgThroughput := total / float64(len(throughputs))
            
            // 各間隔のスループットが平均の±20%以内であることを確認
            for i, throughput := range throughputs {
                deviation := (throughput - avgThroughput) / avgThroughput
                assert.Less(t, deviation, 0.2) // 20%以内
                assert.Greater(t, deviation, -0.2) // -20%以内
                
                t.Logf("Interval %d: %.2f req/sec (deviation: %.2f%%)", 
                    i+1, throughput, deviation*100)
            }
            
            t.Logf("Average throughput: %.2f req/sec", avgThroughput)
        }
    })
}
```

## 2. パフォーマンステストの実行

### 2.1 パフォーマンステスト実行コマンド
```bash
# すべてのパフォーマンステストを実行
go test ./tests/performance/...

# 特定のパフォーマンステストを実行
go test ./tests/performance/load/...
go test ./tests/performance/stress/...
go test ./tests/performance/throughput/...

# ベンチマークテストを実行
go test -bench=. ./tests/performance/...

# パフォーマンステストの詳細出力
go test -v ./tests/performance/...
```

### 2.2 パフォーマンステストの設定
```yaml
# tests/performance/config/performance-test-config.yml
load:
  concurrent_users: 100
  ramp_up_time: 30s
  test_duration: 300s
  target_throughput: 1000

stress:
  max_concurrent_users: 1000
  burst_size: 1000
  memory_limit: 200MB
  cpu_limit: 80%

throughput:
  target_throughput: 1000
  test_duration: 60s
  consistency_threshold: 0.2

monitoring:
  enable_metrics: true
  metrics_endpoint: http://localhost:9090
  log_level: info
```

### 2.3 パフォーマンステストのヘルパー関数
```go
// tests/performance/helpers/performance_helpers.go
package helpers

import (
    "testing"
    "time"
    "net/http"
    "encoding/json"
    "strings"
    "sync"
    "runtime"
    "github.com/stretchr/testify/require"
    "access-log-tracker/internal/domain/models"
)

// パフォーマンステスト用アプリケーション作成
func CreatePerformanceTestApplication(t *testing.T) *models.Application {
    app := &models.Application{
        AppID:       "perf_app_" + time.Now().Format("20060102150405"),
        Name:        "Performance Test Application",
        Description: "Application for performance testing",
        Domain:      "perf-test.example.com",
        APIKey:      "perf_api_key_" + time.Now().Format("20060102150405"),
    }
    
    // APIを使用してアプリケーションを作成
    jsonData, _ := json.Marshal(app)
    resp, err := http.Post("http://localhost:8080/v1/applications",
        "application/json", strings.NewReader(string(jsonData)))
    require.NoError(t, err)
    require.Equal(t, http.StatusCreated, resp.StatusCode)
    
    return app
}

// メモリ使用量を監視
func MonitorMemoryUsage(t *testing.T) func() (uint64, uint64) {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    initialMemory := m.Alloc
    initialHeap := m.HeapAlloc
    
    return func() (uint64, uint64) {
        runtime.ReadMemStats(&m)
        memoryIncrease := m.Alloc - initialMemory
        heapIncrease := m.HeapAlloc - initialHeap
        return memoryIncrease, heapIncrease
    }
}

// 並行リクエストを送信
func SendConcurrentRequests(t *testing.T, app *models.Application, count int, concurrency int) int {
    var wg sync.WaitGroup
    results := make(chan bool, count)
    
    for i := 0; i < concurrency; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < count/concurrency; j++ {
                trackingData := models.TrackingRequest{
                    AppID:     app.AppID,
                    UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
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
    
    wg.Wait()
    close(results)
    
    successCount := 0
    for result := range results {
        if result {
            successCount++
        }
    }
    
    return successCount
}

// スループットを測定
func MeasureThroughput(t *testing.T, app *models.Application, duration time.Duration, targetThroughput int) float64 {
    start := time.Now()
    endTime := start.Add(duration)
    
    var wg sync.WaitGroup
    requestCount := int64(0)
    var requestCountMutex sync.Mutex
    
    // 目標スループットを維持するようにリクエストを送信
    for i := 0; i < runtime.NumCPU()*2; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            
            for time.Now().Before(endTime) {
                trackingData := models.TrackingRequest{
                    AppID:     app.AppID,
                    UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
                    URL:       "https://example.com/throughput-test",
                }
                
                jsonData, _ := json.Marshal(trackingData)
                resp, err := http.Post("http://localhost:8080/v1/track",
                    "application/json", strings.NewReader(string(jsonData)))
                
                if err == nil && resp.StatusCode == http.StatusOK {
                    requestCountMutex.Lock()
                    requestCount++
                    requestCountMutex.Unlock()
                }
                
                // 目標スループットに合わせて間隔を調整
                time.Sleep(time.Duration(1000000/targetThroughput) * time.Microsecond)
            }
        }()
    }
    
    wg.Wait()
    actualDuration := time.Since(start)
    
    return float64(requestCount) / actualDuration.Seconds()
}
```

### 2.4 フェーズ別パフォーマンステスト実行
```bash
# フェーズ6: 統合フェーズのパフォーマンステスト
go test ./tests/performance/load/...
go test ./tests/performance/stress/...
go test ./tests/performance/throughput/...
```
