package performance

import (
	"fmt"
	"net/http"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"runtime"
)

const (
	baseURL = "http://localhost:8080"
)

func BenchmarkBeaconRequests(b *testing.B) {
	appID := 1 // テスト用アプリケーションID
	sessionID := fmt.Sprintf("bench-session-%d", time.Now().Unix())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		beaconURL := fmt.Sprintf("%s/beacon?app_id=%d&session_id=%s&url=/benchmark-test", baseURL, appID, sessionID)
		resp, err := http.Get(beaconURL)
		if err == nil && resp != nil {
			resp.Body.Close()
		}
	}
}

func BenchmarkConcurrentBeaconRequests(b *testing.B) {
	appID := 1
	numGoroutines := 10

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		sessionID := fmt.Sprintf("concurrent-session-%d", time.Now().UnixNano())
		for pb.Next() {
			beaconURL := fmt.Sprintf("%s/beacon?app_id=%d&session_id=%s&url=/concurrent-test", baseURL, appID, sessionID)
			resp, err := http.Get(beaconURL)
			if err == nil && resp != nil {
				resp.Body.Close()
			}
		}
	})
}

func TestBeaconThroughput(t *testing.T) {
	appID := 1
	const numRequests = 1000
	const timeout = 30 * time.Second

	t.Run("Sequential Requests", func(t *testing.T) {
		start := time.Now()
		successCount := 0

		for i := 0; i < numRequests; i++ {
			sessionID := fmt.Sprintf("throughput-seq-%d", i)
			beaconURL := fmt.Sprintf("%s/beacon?app_id=%d&session_id=%s&url=/throughput-test", baseURL, appID, sessionID)
			
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
		assert.GreaterOrEqual(t, successCount, int(float64(numRequests)*0.95)) // 95%以上の成功率
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
				beaconURL := fmt.Sprintf("%s/beacon?app_id=%d&session_id=%s&url=/throughput-test", baseURL, appID, sessionID)
				
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
		assert.GreaterOrEqual(t, successCount, int(float64(numRequests)*0.95)) // 95%以上の成功率
	})
}

func TestBeaconLatency(t *testing.T) {
	appID := 1
	const numRequests = 100

	t.Run("Latency Distribution", func(t *testing.T) {
		var latencies []time.Duration
		successCount := 0

		for i := 0; i < numRequests; i++ {
			sessionID := fmt.Sprintf("latency-%d", i)
			beaconURL := fmt.Sprintf("%s/beacon?app_id=%d&session_id=%s&url=/latency-test", baseURL, appID, sessionID)
			
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
		p95 := calculatePercentile(latencies, 95)
		p99 := calculatePercentile(latencies, 99)

		t.Logf("Latency Statistics:")
		t.Logf("  Average: %v", avg)
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
}

func TestBeaconMemoryUsage(t *testing.T) {
	appID := 1
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
				beaconURL := fmt.Sprintf("%s/beacon?app_id=%d&session_id=%s&url=/memory-test", baseURL, appID, sessionID)
				
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
