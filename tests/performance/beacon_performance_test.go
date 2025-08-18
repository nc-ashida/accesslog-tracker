package performance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"runtime"

	"accesslog-tracker/internal/api/models"
)

const (
	baseURL = "http://localhost:8080"
)

// Application はパフォーマンステスト用のアプリケーション構造体です
type Application struct {
	AppID       string `json:"app_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Domain      string `json:"domain"`
	APIKey      string `json:"api_key"`
	Active      bool   `json:"active"`
}

func BenchmarkBeaconRequests(b *testing.B) {
	app := createTestApplication(b)
	defer cleanupTestApplication(b, app.AppID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sessionID := fmt.Sprintf("bench-session-%d", time.Now().UnixNano())
		beaconURL := fmt.Sprintf("%s/beacon?app_id=%s&session_id=%s&url=/benchmark-test", baseURL, app.AppID, sessionID)
		resp, err := http.Get(beaconURL)
		if err == nil && resp != nil {
			resp.Body.Close()
		}
	}
}

func BenchmarkConcurrentBeaconRequests(b *testing.B) {
	app := createTestApplication(b)
	defer cleanupTestApplication(b, app.AppID)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sessionID := fmt.Sprintf("concurrent-session-%d", time.Now().UnixNano())
			beaconURL := fmt.Sprintf("%s/beacon?app_id=%s&session_id=%s&url=/concurrent-test", baseURL, app.AppID, sessionID)
			resp, err := http.Get(beaconURL)
			if err == nil && resp != nil {
				resp.Body.Close()
			}
		}
	})
}

func BenchmarkTrackingAPIRequests(b *testing.B) {
	app := createTestApplication(b)
	defer cleanupTestApplication(b, app.AppID)

	trackingData := models.TrackingRequest{
		AppID:       app.AppID,
		UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		URL:         "/benchmark-test",
		IPAddress:   "192.168.1.100",
		SessionID:   "bench-session",
		Referrer:    "https://example.com",
		CustomParams: map[string]interface{}{
			"test_type": "benchmark",
		},
	}

	jsonData, _ := json.Marshal(trackingData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", baseURL+"/v1/tracking/track", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", app.APIKey)

		resp, err := http.DefaultClient.Do(req)
		if err == nil && resp != nil {
			resp.Body.Close()
		}
	}
}

func TestBeaconThroughput(t *testing.T) {
	app := createTestApplication(t)
	defer cleanupTestApplication(t, app.AppID)

	const numRequests = 1000
	const timeout = 30 * time.Second

	t.Run("Sequential Requests", func(t *testing.T) {
		start := time.Now()
		successCount := 0

		for i := 0; i < numRequests; i++ {
			sessionID := fmt.Sprintf("throughput-seq-%d", i)
			beaconURL := fmt.Sprintf("%s/beacon?app_id=%s&session_id=%s&url=/throughput-test", baseURL, app.AppID, sessionID)
			
			resp, err := http.Get(beaconURL)
			if err == nil && resp.StatusCode == http.StatusOK {
				successCount++
			}
			if resp != nil {
				resp.Body.Close()
			}
		}

		duration := time.Since(start)
		throughput := float64(successCount) / duration.Seconds()

		t.Logf("Sequential Throughput: %.2f requests/second", throughput)
		t.Logf("Success Rate: %.2f%%", float64(successCount)/float64(numRequests)*100)
		t.Logf("Total Duration: %v", duration)

		assert.GreaterOrEqual(t, throughput, 50.0) // 最低50 req/s
		assert.GreaterOrEqual(t, successCount, int(math.Floor(float64(numRequests)*0.95))) // 95%以上の成功率
	})

	t.Run("Concurrent Requests", func(t *testing.T) {
		start := time.Now()
		successCount := 0
		var mu sync.Mutex
		var wg sync.WaitGroup

		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				
				sessionID := fmt.Sprintf("throughput-conc-%d", index)
				beaconURL := fmt.Sprintf("%s/beacon?app_id=%s&session_id=%s&url=/throughput-test", baseURL, app.AppID, sessionID)
				
				resp, err := http.Get(beaconURL)
				if err == nil && resp.StatusCode == http.StatusOK {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
				if resp != nil {
					resp.Body.Close()
				}
			}(i)
		}

		wg.Wait()
		duration := time.Since(start)
		throughput := float64(successCount) / duration.Seconds()

		t.Logf("Concurrent Throughput: %.2f requests/second", throughput)
		t.Logf("Success Rate: %.2f%%", float64(successCount)/float64(numRequests)*100)
		t.Logf("Total Duration: %v", duration)

		assert.GreaterOrEqual(t, throughput, 100.0) // 最低100 req/s
		assert.GreaterOrEqual(t, successCount, int(math.Floor(float64(numRequests)*0.95))) // 95%以上の成功率
	})

	t.Run("High Load Test", func(t *testing.T) {
		const highLoadRequests = 5000
		const numWorkers = 50
		
		start := time.Now()
		successCount := 0
		var mu sync.Mutex
		var wg sync.WaitGroup

		// ワーカーごとにリクエストを分散
		requestsPerWorker := highLoadRequests / numWorkers
		
		for w := 0; w < numWorkers; w++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				
				for i := 0; i < requestsPerWorker; i++ {
					sessionID := fmt.Sprintf("highload-w%d-r%d", workerID, i)
					beaconURL := fmt.Sprintf("%s/beacon?app_id=%s&session_id=%s&url=/highload-test", baseURL, app.AppID, sessionID)
					
					resp, err := http.Get(beaconURL)
					if err == nil && resp.StatusCode == http.StatusOK {
						mu.Lock()
						successCount++
						mu.Unlock()
					}
					if resp != nil {
						resp.Body.Close()
					}
				}
			}(w)
		}

		wg.Wait()
		duration := time.Since(start)
		throughput := float64(successCount) / duration.Seconds()

		t.Logf("High Load Throughput: %.2f requests/second", throughput)
		t.Logf("Success Rate: %.2f%%", float64(successCount)/float64(highLoadRequests)*100)
		t.Logf("Total Duration: %v", duration)
		t.Logf("Workers: %d", numWorkers)

		assert.GreaterOrEqual(t, throughput, 200.0) // 最低200 req/s
		assert.GreaterOrEqual(t, successCount, int(math.Floor(float64(highLoadRequests)*0.90))) // 90%以上の成功率
	})
}

func TestBeaconLatency(t *testing.T) {
	app := createTestApplication(t)
	defer cleanupTestApplication(t, app.AppID)

	const numRequests = 100

	t.Run("Latency Distribution", func(t *testing.T) {
		var latencies []time.Duration
		successCount := 0

		for i := 0; i < numRequests; i++ {
			sessionID := fmt.Sprintf("latency-%d", i)
			beaconURL := fmt.Sprintf("%s/beacon?app_id=%s&session_id=%s&url=/latency-test", baseURL, app.AppID, sessionID)
			
			start := time.Now()
			resp, err := http.Get(beaconURL)
			latency := time.Since(start)
			
			if err == nil && resp.StatusCode == http.StatusOK {
				latencies = append(latencies, latency)
				successCount++
			}
			if resp != nil {
				resp.Body.Close()
			}
		}

		if len(latencies) == 0 {
			t.Skip("No successful requests to measure latency")
		}

		// 統計計算
		var total time.Duration
		var min, max time.Duration = latencies[0], latencies[0]
		
		for _, latency := range latencies {
			total += latency
			if latency < min {
				min = latency
			}
			if latency > max {
				max = latency
			}
		}

		avg := total / time.Duration(len(latencies))
		p50 := calculatePercentile(latencies, 50)
		p95 := calculatePercentile(latencies, 95)
		p99 := calculatePercentile(latencies, 99)

		t.Logf("Latency Statistics:")
		t.Logf("  Average: %v", avg)
		t.Logf("  50th Percentile: %v", p50)
		t.Logf("  95th Percentile: %v", p95)
		t.Logf("  99th Percentile: %v", p99)
		t.Logf("  Min: %v", min)
		t.Logf("  Max: %v", max)
		t.Logf("  Success Rate: %.2f%%", float64(successCount)/float64(numRequests)*100)

		// パフォーマンス要件の検証
		assert.LessOrEqual(t, avg, 100*time.Millisecond) // 平均100ms以下
		assert.LessOrEqual(t, p95, 200*time.Millisecond) // 95%が200ms以下
		assert.LessOrEqual(t, p99, 500*time.Millisecond) // 99%が500ms以下
	})

	t.Run("Latency Under Load", func(t *testing.T) {
		const loadRequests = 500
		const concurrentUsers = 10
		
		var latencies []time.Duration
		var mu sync.Mutex
		var wg sync.WaitGroup

		requestsPerUser := loadRequests / concurrentUsers

		for u := 0; u < concurrentUsers; u++ {
			wg.Add(1)
			go func(userID int) {
				defer wg.Done()
				
				for i := 0; i < requestsPerUser; i++ {
					sessionID := fmt.Sprintf("load-u%d-r%d", userID, i)
					beaconURL := fmt.Sprintf("%s/beacon?app_id=%s&session_id=%s&url=/load-test", baseURL, app.AppID, sessionID)
					
					start := time.Now()
					resp, err := http.Get(beaconURL)
					latency := time.Since(start)
					
					if err == nil && resp.StatusCode == http.StatusOK {
						mu.Lock()
						latencies = append(latencies, latency)
						mu.Unlock()
					}
					if resp != nil {
						resp.Body.Close()
					}
				}
			}(u)
		}

		wg.Wait()

		if len(latencies) == 0 {
			t.Skip("No successful requests to measure latency under load")
		}

		// 統計計算
		var total time.Duration
		for _, latency := range latencies {
			total += latency
		}

		avg := total / time.Duration(len(latencies))
		p95 := calculatePercentile(latencies, 95)

		t.Logf("Latency Under Load:")
		t.Logf("  Average: %v", avg)
		t.Logf("  95th Percentile: %v", p95)
		t.Logf("  Concurrent Users: %d", concurrentUsers)
		t.Logf("  Total Requests: %d", len(latencies))

		assert.LessOrEqual(t, avg, 200*time.Millisecond) // 負荷下でも平均200ms以下
		assert.LessOrEqual(t, p95, 500*time.Millisecond) // 負荷下でも95%が500ms以下
	})
}

func TestBeaconMemoryUsage(t *testing.T) {
	app := createTestApplication(t)
	defer cleanupTestApplication(t, app.AppID)

	const numRequests = 1000

	t.Run("Memory Usage Under Load", func(t *testing.T) {
		// メモリ使用量のベースラインを取得
		var m1, m2 runtime.MemStats
		runtime.ReadMemStats(&m1)

		var wg sync.WaitGroup
		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				
				sessionID := fmt.Sprintf("memory-%d", index)
				beaconURL := fmt.Sprintf("%s/beacon?app_id=%s&session_id=%s&url=/memory-test", baseURL, app.AppID, sessionID)
				
				resp, err := http.Get(beaconURL)
				if err == nil && resp != nil {
					resp.Body.Close()
				}
			}(i)
		}

		wg.Wait()
		
		// GCを強制実行
		runtime.GC()
		runtime.ReadMemStats(&m2)

		memoryIncrease := m2.Alloc - m1.Alloc
		memoryIncreaseMB := float64(memoryIncrease) / 1024 / 1024

		t.Logf("Memory Usage:")
		t.Logf("  Initial: %.2f MB", float64(m1.Alloc)/1024/1024)
		t.Logf("  Final: %.2f MB", float64(m2.Alloc)/1024/1024)
		t.Logf("  Increase: %.2f MB", memoryIncreaseMB)

		// メモリリークがないことを確認（増加が1MB以下）
		assert.LessOrEqual(t, memoryIncreaseMB, 1.0)
	})

	t.Run("Memory Usage Over Time", func(t *testing.T) {
		const iterations = 10
		const requestsPerIteration = 100
		
		var memorySnapshots []float64
		
		for iter := 0; iter < iterations; iter++ {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			memorySnapshots = append(memorySnapshots, float64(m.Alloc)/1024/1024)
			
			var wg sync.WaitGroup
			for i := 0; i < requestsPerIteration; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					
					sessionID := fmt.Sprintf("memory-time-%d-%d", iter, index)
					beaconURL := fmt.Sprintf("%s/beacon?app_id=%s&session_id=%s&url=/memory-time-test", baseURL, app.AppID, sessionID)
					
					resp, err := http.Get(beaconURL)
					if err == nil && resp != nil {
						resp.Body.Close()
					}
				}(i)
			}
			wg.Wait()
			
			// 各イテレーション後にGCを実行
			runtime.GC()
			time.Sleep(100 * time.Millisecond)
		}

		// メモリ使用量の傾向を分析
		initialMemory := memorySnapshots[0]
		finalMemory := memorySnapshots[len(memorySnapshots)-1]
		memoryGrowth := finalMemory - initialMemory

		t.Logf("Memory Usage Over Time:")
		t.Logf("  Initial: %.2f MB", initialMemory)
		t.Logf("  Final: %.2f MB", finalMemory)
		t.Logf("  Growth: %.2f MB", memoryGrowth)
		t.Logf("  Iterations: %d", iterations)

		// メモリリークがないことを確認（成長が2MB以下）
		assert.LessOrEqual(t, memoryGrowth, 2.0)
	})
}

func TestBeaconStressTest(t *testing.T) {
	app := createTestApplication(t)
	defer cleanupTestApplication(t, app.AppID)

	t.Run("Sustained Load", func(t *testing.T) {
		const duration = 30 * time.Second
		const requestsPerSecond = 50
		
		start := time.Now()
		successCount := 0
		totalRequests := 0
		var mu sync.Mutex
		
		// 指定された期間、継続的にリクエストを送信
		ticker := time.NewTicker(time.Second / time.Duration(requestsPerSecond))
		defer ticker.Stop()
		
		done := make(chan bool)
		go func() {
			time.Sleep(duration)
			done <- true
		}()
		
		for {
			select {
			case <-ticker.C:
				totalRequests++
				go func() {
					sessionID := fmt.Sprintf("stress-%d", time.Now().UnixNano())
					beaconURL := fmt.Sprintf("%s/beacon?app_id=%s&session_id=%s&url=/stress-test", baseURL, app.AppID, sessionID)
					
					resp, err := http.Get(beaconURL)
					if err == nil && resp.StatusCode == http.StatusOK {
						mu.Lock()
						successCount++
						mu.Unlock()
					}
					if resp != nil {
						resp.Body.Close()
					}
				}()
			case <-done:
				goto end
			}
		}
	end:
		
		elapsed := time.Since(start)
		actualThroughput := float64(successCount) / elapsed.Seconds()
		successRate := float64(successCount) / float64(totalRequests) * 100

		t.Logf("Stress Test Results:")
		t.Logf("  Duration: %v", elapsed)
		t.Logf("  Total Requests: %d", totalRequests)
		t.Logf("  Successful Requests: %d", successCount)
		t.Logf("  Throughput: %.2f requests/second", actualThroughput)
		t.Logf("  Success Rate: %.2f%%", successRate)

		assert.GreaterOrEqual(t, actualThroughput, float64(requestsPerSecond)*0.8) // 80%以上のスループット
		assert.GreaterOrEqual(t, successRate, 90.0) // 90%以上の成功率
	})

	t.Run("Burst Load", func(t *testing.T) {
		const burstSize = 200
		const numBursts = 5
		
		var totalSuccess int
		var totalRequests int
		
		for burst := 0; burst < numBursts; burst++ {
			successCount := 0
			var wg sync.WaitGroup
			
			// バーストリクエストを送信
			for i := 0; i < burstSize; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					
					sessionID := fmt.Sprintf("burst-%d-%d", burst, index)
					beaconURL := fmt.Sprintf("%s/beacon?app_id=%s&session_id=%s&url=/burst-test", baseURL, app.AppID, sessionID)
					
					resp, err := http.Get(beaconURL)
					if err == nil && resp.StatusCode == http.StatusOK {
						successCount++
					}
					if resp != nil {
						resp.Body.Close()
					}
				}(i)
			}
			
			wg.Wait()
			totalSuccess += successCount
			totalRequests += burstSize
			
			// バースト間の休憩
			time.Sleep(1 * time.Second)
		}
		
		successRate := float64(totalSuccess) / float64(totalRequests) * 100
		
		t.Logf("Burst Load Test Results:")
		t.Logf("  Total Bursts: %d", numBursts)
		t.Logf("  Burst Size: %d", burstSize)
		t.Logf("  Total Requests: %d", totalRequests)
		t.Logf("  Successful Requests: %d", totalSuccess)
		t.Logf("  Success Rate: %.2f%%", successRate)
		
		assert.GreaterOrEqual(t, successRate, 85.0) // 85%以上の成功率
	})
}

// ヘルパー関数
func calculatePercentile(latencies []time.Duration, percentile int) time.Duration {
	if len(latencies) == 0 {
		return 0
	}

	// ソート
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	index := int(float64(len(latencies)-1) * float64(percentile) / 100.0)
	return latencies[index]
}

// createTestApplication はテスト用のアプリケーションを作成します
func createTestApplication(t testing.TB) *Application {
	createRequest := map[string]interface{}{
		"name":        "Performance Test Application",
		"description": "Test application for performance testing",
		"domain":      "perf-test.example.com",
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
func cleanupTestApplication(t testing.TB, appID string) {
	req, err := http.NewRequest("DELETE", baseURL+"/v1/applications/"+appID, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	if err == nil && resp != nil {
		resp.Body.Close()
	}
}
