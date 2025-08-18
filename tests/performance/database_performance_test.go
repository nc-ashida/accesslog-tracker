package performance

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"sort"
	"sync"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"accesslog-tracker/internal/domain/models"
	"accesslog-tracker/internal/infrastructure/database/postgresql"
	"accesslog-tracker/internal/infrastructure/database/postgresql/repositories"
)

const (
	testDBHost     = "postgres"
	testDBPort     = 5432
	testDBName     = "access_log_tracker_test"
	testDBUser     = "postgres"
	testDBPassword = "password"
)

// DatabasePerformanceTest はデータベースパフォーマンステスト用の構造体です
type DatabasePerformanceTest struct {
	conn         *postgresql.Connection
	appRepo      *repositories.ApplicationRepository
	trackingRepo *repositories.TrackingRepository
}

func BenchmarkDatabaseConnection(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn := postgresql.NewConnection("test")
		dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
			testDBHost, testDBPort, testDBName, testDBUser, testDBPassword)
		err := conn.Connect(dsn)
		if err == nil {
			conn.Close()
		}
	}
}

func BenchmarkApplicationRepositoryOperations(b *testing.B) {
	test := setupDatabaseTest(b)
	defer test.cleanup()

	// テストデータの準備
	app := &models.Application{
		Name:        "Benchmark Test App",
		Description: "Test application for database benchmarking",
		Domain:      "benchmark-test.example.com",
		Active:      true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// アプリケーション作成
		err := test.appRepo.Create(context.Background(), app)
		if err == nil {
			// アプリケーション取得
			_, err = test.appRepo.GetByID(context.Background(), app.AppID)
			if err == nil {
				// アプリケーション削除
				test.appRepo.Delete(context.Background(), app.AppID)
			}
		}
	}
}

func BenchmarkConcurrentApplicationRepositoryOperations(b *testing.B) {
	test := setupDatabaseTest(b)
	defer test.cleanup()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		app := &models.Application{
			Name:        "Concurrent Benchmark Test App",
			Description: "Test application for concurrent database benchmarking",
			Domain:      "concurrent-benchmark-test.example.com",
			Active:      true,
		}

		for pb.Next() {
			// アプリケーション作成
			err := test.appRepo.Create(context.Background(), app)
			if err == nil {
				// アプリケーション取得
				_, err = test.appRepo.GetByID(context.Background(), app.AppID)
				if err == nil {
					// アプリケーション削除
					test.appRepo.Delete(context.Background(), app.AppID)
				}
			}
		}
	})
}

func BenchmarkTrackingRepositoryOperations(b *testing.B) {
	test := setupDatabaseTest(b)
	defer test.cleanup()

	// テストアプリケーションの作成
	app := &models.Application{
		Name:        "Tracking Benchmark Test App",
		Description: "Test application for tracking database benchmarking",
		Domain:      "tracking-benchmark-test.example.com",
		Active:      true,
	}
	err := test.appRepo.Create(context.Background(), app)
	require.NoError(b, err)

	// テストデータの準備
	tracking := &models.TrackingData{
		AppID:     app.AppID,
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		URL:       "/benchmark-test",
		IPAddress: "192.168.1.100",
		SessionID: "bench-session",
		Referrer:  "https://example.com",
		CustomParams: map[string]interface{}{
			"test_type": "benchmark",
		},
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// トラッキングデータ作成
		err := test.trackingRepo.Create(context.Background(), tracking)
		if err == nil {
			// トラッキングデータ取得
			_, err = test.trackingRepo.GetByID(context.Background(), tracking.ID)
			if err == nil {
				// トラッキングデータ削除
				test.trackingRepo.Delete(context.Background(), tracking.ID)
			}
		}
	}

	// テストアプリケーションの削除
	test.appRepo.Delete(context.Background(), app.AppID)
}

func BenchmarkConcurrentTrackingRepositoryOperations(b *testing.B) {
	test := setupDatabaseTest(b)
	defer test.cleanup()

	// テストアプリケーションの作成
	app := &models.Application{
		Name:        "Concurrent Tracking Benchmark Test App",
		Description: "Test application for concurrent tracking database benchmarking",
		Domain:      "concurrent-tracking-benchmark-test.example.com",
		Active:      true,
	}
	err := test.appRepo.Create(context.Background(), app)
	require.NoError(b, err)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		tracking := &models.TrackingData{
			AppID:     app.AppID,
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			URL:       "/concurrent-benchmark-test",
			IPAddress: "192.168.1.100",
			SessionID: "concurrent-bench-session",
			Referrer:  "https://example.com",
			CustomParams: map[string]interface{}{
				"test_type": "concurrent_benchmark",
			},
			Timestamp: time.Now(),
		}

		for pb.Next() {
			// トラッキングデータ作成
			err := test.trackingRepo.Create(context.Background(), tracking)
			if err == nil {
				// トラッキングデータ取得
				_, err = test.trackingRepo.GetByID(context.Background(), tracking.ID)
				if err == nil {
					// トラッキングデータ削除
					test.trackingRepo.Delete(context.Background(), tracking.ID)
				}
			}
		}
	})

	// テストアプリケーションの削除
	test.appRepo.Delete(context.Background(), app.AppID)
}

func TestDatabaseThroughput(t *testing.T) {
	test := setupDatabaseTest(t)
	defer test.cleanup()

	const numOperations = 1000

	t.Run("Application Repository Sequential Operations", func(t *testing.T) {
		start := time.Now()
		successCount := 0

		for i := 0; i < numOperations; i++ {
			app := &models.Application{
				Name:        fmt.Sprintf("Throughput Test App %d", i),
				Description: "Test application for database throughput testing",
				Domain:      fmt.Sprintf("throughput-test-%d.example.com", i),
				Active:      true,
			}

			err := test.appRepo.Create(context.Background(), app)
			if err == nil {
				successCount++
				// 作成したアプリケーションを削除
				test.appRepo.Delete(context.Background(), app.AppID)
			}
		}

		duration := time.Since(start)
		throughput := float64(successCount) / duration.Seconds()

		t.Logf("Application Repository Sequential Throughput: %.2f operations/second", throughput)
		t.Logf("Success Rate: %.2f%%", float64(successCount)/float64(numOperations)*100)
		t.Logf("Total Duration: %v", duration)

		assert.GreaterOrEqual(t, throughput, 50.0)                                           // 最低50 ops/s
		assert.GreaterOrEqual(t, successCount, int(math.Floor(float64(numOperations)*0.95))) // 95%以上の成功率
	})

	t.Run("Application Repository Concurrent Operations", func(t *testing.T) {
		start := time.Now()
		successCount := 0
		var mu sync.Mutex
		var wg sync.WaitGroup

		for i := 0; i < numOperations; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				app := &models.Application{
					Name:        fmt.Sprintf("Concurrent Throughput Test App %d", index),
					Description: "Test application for concurrent database throughput testing",
					Domain:      fmt.Sprintf("concurrent-throughput-test-%d.example.com", index),
					Active:      true,
				}

				err := test.appRepo.Create(context.Background(), app)
				if err == nil {
					mu.Lock()
					successCount++
					mu.Unlock()
					// 作成したアプリケーションを削除
					test.appRepo.Delete(context.Background(), app.AppID)
				}
			}(i)
		}

		wg.Wait()
		duration := time.Since(start)
		throughput := float64(successCount) / duration.Seconds()

		t.Logf("Application Repository Concurrent Throughput: %.2f operations/second", throughput)
		t.Logf("Success Rate: %.2f%%", float64(successCount)/float64(numOperations)*100)
		t.Logf("Total Duration: %v", duration)

		assert.GreaterOrEqual(t, throughput, 100.0)                                          // 最低100 ops/s
		assert.GreaterOrEqual(t, successCount, int(math.Floor(float64(numOperations)*0.95))) // 95%以上の成功率
	})

	t.Run("Tracking Repository High Load Test", func(t *testing.T) {
		const highLoadOperations = 5000
		const numWorkers = 50

		// テストアプリケーションの作成
		app := &models.Application{
			Name:        "High Load Tracking Test App",
			Description: "Test application for high load tracking database testing",
			Domain:      "highload-tracking-test.example.com",
			Active:      true,
		}
		err := test.appRepo.Create(context.Background(), app)
		require.NoError(t, err)
		defer test.appRepo.Delete(context.Background(), app.AppID)

		start := time.Now()
		successCount := 0
		var mu sync.Mutex
		var wg sync.WaitGroup

		// ワーカーごとにリクエストを分散
		operationsPerWorker := highLoadOperations / numWorkers

		for w := 0; w < numWorkers; w++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				for i := 0; i < operationsPerWorker; i++ {
					tracking := &models.TrackingData{
						AppID:     app.AppID,
						UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
						URL:       fmt.Sprintf("/highload-test-%d-%d", workerID, i),
						IPAddress: "192.168.1.100",
						SessionID: fmt.Sprintf("highload-session-%d-%d", workerID, i),
						Referrer:  "https://example.com",
						CustomParams: map[string]interface{}{
							"test_type":    "highload",
							"worker_id":    workerID,
							"operation_id": i,
						},
						Timestamp: time.Now(),
					}

					err := test.trackingRepo.Create(context.Background(), tracking)
					if err == nil {
						mu.Lock()
						successCount++
						mu.Unlock()
						// 作成したトラッキングデータを削除
						test.trackingRepo.Delete(context.Background(), tracking.ID)
					}
				}
			}(w)
		}

		wg.Wait()
		duration := time.Since(start)
		throughput := float64(successCount) / duration.Seconds()

		t.Logf("Tracking Repository High Load Throughput: %.2f operations/second", throughput)
		t.Logf("Success Rate: %.2f%%", float64(successCount)/float64(highLoadOperations)*100)
		t.Logf("Total Duration: %v", duration)
		t.Logf("Workers: %d", numWorkers)

		assert.GreaterOrEqual(t, throughput, 200.0)                                               // 最低200 ops/s
		assert.GreaterOrEqual(t, successCount, int(math.Floor(float64(highLoadOperations)*0.90))) // 90%以上の成功率
	})
}

func TestDatabaseLatency(t *testing.T) {
	test := setupDatabaseTest(t)
	defer test.cleanup()

	const numOperations = 100

	t.Run("Application Repository Latency Distribution", func(t *testing.T) {
		var latencies []time.Duration
		successCount := 0

		for i := 0; i < numOperations; i++ {
			app := &models.Application{
				Name:        fmt.Sprintf("Latency Test App %d", i),
				Description: "Test application for database latency testing",
				Domain:      fmt.Sprintf("latency-test-%d.example.com", i),
				Active:      true,
			}

			start := time.Now()
			err := test.appRepo.Create(context.Background(), app)
			latency := time.Since(start)

			if err == nil {
				latencies = append(latencies, latency)
				successCount++
				// 作成したアプリケーションを削除
				test.appRepo.Delete(context.Background(), app.AppID)
			}
		}

		if len(latencies) == 0 {
			t.Skip("No successful operations to measure latency")
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

		// パーセンタイル計算
		sort.Slice(latencies, func(i, j int) bool {
			return latencies[i] < latencies[j]
		})
		p50 := latencies[int(float64(len(latencies)-1)*0.5)]
		p95 := latencies[int(float64(len(latencies)-1)*0.95)]
		p99 := latencies[int(float64(len(latencies)-1)*0.99)]

		t.Logf("Application Repository Latency Statistics:")
		t.Logf("  Average: %v", avg)
		t.Logf("  50th Percentile: %v", p50)
		t.Logf("  95th Percentile: %v", p95)
		t.Logf("  99th Percentile: %v", p99)
		t.Logf("  Min: %v", min)
		t.Logf("  Max: %v", max)
		t.Logf("  Success Rate: %.2f%%", float64(successCount)/float64(numOperations)*100)

		// パフォーマンス要件の検証
		assert.LessOrEqual(t, avg, 50*time.Millisecond)  // 平均50ms以下
		assert.LessOrEqual(t, p95, 100*time.Millisecond) // 95%が100ms以下
		assert.LessOrEqual(t, p99, 200*time.Millisecond) // 99%が200ms以下
	})

	t.Run("Tracking Repository Latency Under Load", func(t *testing.T) {
		const loadOperations = 500
		const concurrentUsers = 10

		// テストアプリケーションの作成
		app := &models.Application{
			Name:        "Load Tracking Test App",
			Description: "Test application for load tracking database testing",
			Domain:      "load-tracking-test.example.com",
			Active:      true,
		}
		err := test.appRepo.Create(context.Background(), app)
		require.NoError(t, err)
		defer test.appRepo.Delete(context.Background(), app.AppID)

		var latencies []time.Duration
		var mu sync.Mutex
		var wg sync.WaitGroup

		operationsPerUser := loadOperations / concurrentUsers

		for u := 0; u < concurrentUsers; u++ {
			wg.Add(1)
			go func(userID int) {
				defer wg.Done()

				for i := 0; i < operationsPerUser; i++ {
					tracking := &models.TrackingData{
						AppID:     app.AppID,
						UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
						URL:       fmt.Sprintf("/load-test-%d-%d", userID, i),
						IPAddress: "192.168.1.100",
						SessionID: fmt.Sprintf("load-session-%d-%d", userID, i),
						Referrer:  "https://example.com",
						CustomParams: map[string]interface{}{
							"test_type":    "load",
							"user_id":      userID,
							"operation_id": i,
						},
						Timestamp: time.Now(),
					}

					start := time.Now()
					err := test.trackingRepo.Create(context.Background(), tracking)
					latency := time.Since(start)

					if err == nil {
						mu.Lock()
						latencies = append(latencies, latency)
						mu.Unlock()
						// 作成したトラッキングデータを削除
						test.trackingRepo.Delete(context.Background(), tracking.ID)
					}
				}
			}(u)
		}

		wg.Wait()

		if len(latencies) == 0 {
			t.Skip("No successful operations to measure latency under load")
		}

		// 統計計算
		var total time.Duration
		for _, latency := range latencies {
			total += latency
		}

		avg := total / time.Duration(len(latencies))

		// パーセンタイル計算
		sort.Slice(latencies, func(i, j int) bool {
			return latencies[i] < latencies[j]
		})
		p95 := latencies[int(float64(len(latencies)-1)*0.95)]

		t.Logf("Tracking Repository Latency Under Load:")
		t.Logf("  Average: %v", avg)
		t.Logf("  95th Percentile: %v", p95)
		t.Logf("  Concurrent Users: %d", concurrentUsers)
		t.Logf("  Total Operations: %d", len(latencies))

		assert.LessOrEqual(t, avg, 100*time.Millisecond) // 負荷下でも平均100ms以下
		assert.LessOrEqual(t, p95, 200*time.Millisecond) // 負荷下でも95%が200ms以下
	})
}

func TestDatabaseMemoryUsage(t *testing.T) {
	test := setupDatabaseTest(t)
	defer test.cleanup()

	const numOperations = 1000

	t.Run("Database Memory Usage Under Load", func(t *testing.T) {
		// メモリ使用量のベースラインを取得
		var m1, m2 runtime.MemStats
		runtime.ReadMemStats(&m1)

		var wg sync.WaitGroup
		for i := 0; i < numOperations; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				app := &models.Application{
					Name:        fmt.Sprintf("Memory Test App %d", index),
					Description: "Test application for database memory usage testing",
					Domain:      fmt.Sprintf("memory-test-%d.example.com", index),
					Active:      true,
				}

				err := test.appRepo.Create(context.Background(), app)
				if err == nil {
					// 作成したアプリケーションを削除
					test.appRepo.Delete(context.Background(), app.AppID)
				}
			}(i)
		}

		wg.Wait()

		// GCを強制実行
		runtime.GC()
		runtime.ReadMemStats(&m2)

		memoryIncrease := int64(m2.Alloc) - int64(m1.Alloc)
		memoryIncreaseMB := float64(memoryIncrease) / 1024 / 1024

		t.Logf("Database Memory Usage:")
		t.Logf("  Initial: %.2f MB", float64(m1.Alloc)/1024/1024)
		t.Logf("  Final: %.2f MB", float64(m2.Alloc)/1024/1024)
		t.Logf("  Increase: %.2f MB", memoryIncreaseMB)

		// メモリリークがないことを確認（増加が1MB以下）
		assert.LessOrEqual(t, memoryIncreaseMB, 1.0)
	})

	t.Run("Database Memory Usage Over Time", func(t *testing.T) {
		const iterations = 10
		const operationsPerIteration = 100

		var memorySnapshots []float64

		for iter := 0; iter < iterations; iter++ {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			memorySnapshots = append(memorySnapshots, float64(m.Alloc)/1024/1024)

			var wg sync.WaitGroup
			for i := 0; i < operationsPerIteration; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()

					app := &models.Application{
						Name:        fmt.Sprintf("Memory Time Test App %d-%d", iter, index),
						Description: "Test application for database memory usage over time testing",
						Domain:      fmt.Sprintf("memory-time-test-%d-%d.example.com", iter, index),
						Active:      true,
					}

					err := test.appRepo.Create(context.Background(), app)
					if err == nil {
						// 作成したアプリケーションを削除
						test.appRepo.Delete(context.Background(), app.AppID)
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

		t.Logf("Database Memory Usage Over Time:")
		t.Logf("  Initial: %.2f MB", initialMemory)
		t.Logf("  Final: %.2f MB", finalMemory)
		t.Logf("  Growth: %.2f MB", memoryGrowth)
		t.Logf("  Iterations: %d", iterations)

		// メモリリークがないことを確認（成長が2MB以下）
		assert.LessOrEqual(t, memoryGrowth, 2.0)
	})
}

func TestDatabaseStressTest(t *testing.T) {
	test := setupDatabaseTest(t)
	defer test.cleanup()

	t.Run("Database Sustained Load", func(t *testing.T) {
		const duration = 30 * time.Second
		const operationsPerSecond = 50

		start := time.Now()
		successCount := 0
		totalOperations := 0
		var mu sync.Mutex

		// 指定された期間、継続的に操作を実行
		ticker := time.NewTicker(time.Second / time.Duration(operationsPerSecond))
		defer ticker.Stop()

		done := make(chan bool)
		go func() {
			time.Sleep(duration)
			done <- true
		}()

		for {
			select {
			case <-ticker.C:
				totalOperations++
				go func() {
					app := &models.Application{
						Name:        fmt.Sprintf("Stress Test App %d", time.Now().UnixNano()),
						Description: "Test application for database stress testing",
						Domain:      fmt.Sprintf("stress-test-%d.example.com", time.Now().UnixNano()),
						Active:      true,
					}

					err := test.appRepo.Create(context.Background(), app)
					if err == nil {
						mu.Lock()
						successCount++
						mu.Unlock()
						// 作成したアプリケーションを削除
						test.appRepo.Delete(context.Background(), app.AppID)
					}
				}()
			case <-done:
				goto end
			}
		}
	end:

		elapsed := time.Since(start)
		actualThroughput := float64(successCount) / elapsed.Seconds()
		successRate := float64(successCount) / float64(totalOperations) * 100

		t.Logf("Database Stress Test Results:")
		t.Logf("  Duration: %v", elapsed)
		t.Logf("  Total Operations: %d", totalOperations)
		t.Logf("  Successful Operations: %d", successCount)
		t.Logf("  Throughput: %.2f operations/second", actualThroughput)
		t.Logf("  Success Rate: %.2f%%", successRate)

		assert.GreaterOrEqual(t, actualThroughput, float64(operationsPerSecond)*0.8) // 80%以上のスループット
		assert.GreaterOrEqual(t, successRate, 90.0)                                  // 90%以上の成功率
	})

	t.Run("Database Burst Load", func(t *testing.T) {
		const burstSize = 200
		const numBursts = 5

		var totalSuccess int
		var totalOperations int

		for burst := 0; burst < numBursts; burst++ {
			successCount := 0
			var wg sync.WaitGroup

			// バースト操作を実行
			for i := 0; i < burstSize; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()

					app := &models.Application{
						Name:        fmt.Sprintf("Burst Test App %d-%d", burst, index),
						Description: "Test application for database burst load testing",
						Domain:      fmt.Sprintf("burst-test-%d-%d.example.com", burst, index),
						Active:      true,
					}

					err := test.appRepo.Create(context.Background(), app)
					if err == nil {
						successCount++
						// 作成したアプリケーションを削除
						test.appRepo.Delete(context.Background(), app.AppID)
					}
				}(i)
			}

			wg.Wait()
			totalSuccess += successCount
			totalOperations += burstSize

			// バースト間の休憩
			time.Sleep(1 * time.Second)
		}

		successRate := float64(totalSuccess) / float64(totalOperations) * 100

		t.Logf("Database Burst Load Test Results:")
		t.Logf("  Total Bursts: %d", numBursts)
		t.Logf("  Burst Size: %d", burstSize)
		t.Logf("  Total Operations: %d", totalOperations)
		t.Logf("  Successful Operations: %d", totalSuccess)
		t.Logf("  Success Rate: %.2f%%", successRate)

		assert.GreaterOrEqual(t, successRate, 85.0) // 85%以上の成功率
	})
}

// setupDatabaseTest はデータベーステストのセットアップを行います
func setupDatabaseTest(t testing.TB) *DatabasePerformanceTest {
	// データベース接続
	conn := postgresql.NewConnection("test")
	dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		testDBHost, testDBPort, testDBName, testDBUser, testDBPassword)
	err := conn.Connect(dsn)
	require.NoError(t, err)

	// リポジトリの初期化
	appRepo := repositories.NewApplicationRepository(conn.GetDB())
	trackingRepo := repositories.NewTrackingRepository(conn.GetDB())

	return &DatabasePerformanceTest{
		conn:         conn,
		appRepo:      appRepo,
		trackingRepo: trackingRepo,
	}
}

// cleanup はデータベーステストのクリーンアップを行います
func (dpt *DatabasePerformanceTest) cleanup() {
	if dpt.conn != nil {
		dpt.conn.Close()
	}
}
