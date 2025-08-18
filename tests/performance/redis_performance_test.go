package performance

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"accesslog-tracker/internal/infrastructure/cache/redis"
)

const (
	testRedisHost = "redis"
	testRedisPort = 6379
)

// RedisPerformanceTest はRedisパフォーマンステスト用の構造体です
type RedisPerformanceTest struct {
	cacheService *redis.CacheService
}

func BenchmarkRedisConnection(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cacheService := redis.NewCacheService("redis:6379")
		err := cacheService.Connect()
		if err == nil {
			cacheService.Close()
		}
	}
}

func BenchmarkRedisSetOperations(b *testing.B) {
	test := setupRedisTest(b)
	defer test.cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("benchmark:set:%d", i)
		value := fmt.Sprintf("benchmark-value-%d", i)
		test.cacheService.Set(context.Background(), key, value, time.Hour)
	}
}

func BenchmarkRedisGetOperations(b *testing.B) {
	test := setupRedisTest(b)
	defer test.cleanup()

	// テストデータの準備
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("benchmark:get:%d", i)
		value := fmt.Sprintf("benchmark-value-%d", i)
		test.cacheService.Set(context.Background(), key, value, time.Hour)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("benchmark:get:%d", i%1000)
		test.cacheService.Get(context.Background(), key)
	}
}

func BenchmarkConcurrentRedisOperations(b *testing.B) {
	test := setupRedisTest(b)
	defer test.cleanup()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("concurrent:benchmark:%d", i)
			value := fmt.Sprintf("concurrent-value-%d", i)

			// Set操作
			test.cacheService.Set(context.Background(), key, value, time.Hour)

			// Get操作
			test.cacheService.Get(context.Background(), key)

			// Delete操作
			test.cacheService.Delete(context.Background(), key)

			i++
		}
	})
}

func TestRedisThroughput(t *testing.T) {
	test := setupRedisTest(t)
	defer test.cleanup()

	const numOperations = 1000

	t.Run("Redis Set Sequential Operations", func(t *testing.T) {
		start := time.Now()
		successCount := 0

		for i := 0; i < numOperations; i++ {
			key := fmt.Sprintf("throughput:set:%d", i)
			value := fmt.Sprintf("throughput-value-%d", i)
			err := test.cacheService.Set(context.Background(), key, value, time.Hour)
			if err == nil {
				successCount++
			}
		}

		duration := time.Since(start)
		throughput := float64(successCount) / duration.Seconds()

		t.Logf("Redis Set Sequential Throughput: %.2f operations/second", throughput)
		t.Logf("Success Rate: %.2f%%", float64(successCount)/float64(numOperations)*100)
		t.Logf("Total Duration: %v", duration)

		assert.GreaterOrEqual(t, throughput, 1000.0)                                         // 最低1000 ops/s
		assert.GreaterOrEqual(t, successCount, int(math.Floor(float64(numOperations)*0.95))) // 95%以上の成功率
	})

	t.Run("Redis Get Sequential Operations", func(t *testing.T) {
		// テストデータの準備
		for i := 0; i < numOperations; i++ {
			key := fmt.Sprintf("throughput:get:prep:%d", i)
			value := fmt.Sprintf("throughput-value-%d", i)
			test.cacheService.Set(context.Background(), key, value, time.Hour)
		}

		start := time.Now()
		successCount := 0

		for i := 0; i < numOperations; i++ {
			key := fmt.Sprintf("throughput:get:prep:%d", i)
			_, err := test.cacheService.Get(context.Background(), key)
			if err == nil {
				successCount++
			}
		}

		duration := time.Since(start)
		throughput := float64(successCount) / duration.Seconds()

		t.Logf("Redis Get Sequential Throughput: %.2f operations/second", throughput)
		t.Logf("Success Rate: %.2f%%", float64(successCount)/float64(numOperations)*100)
		t.Logf("Total Duration: %v", duration)

		assert.GreaterOrEqual(t, throughput, 2000.0)                                         // 最低2000 ops/s
		assert.GreaterOrEqual(t, successCount, int(math.Floor(float64(numOperations)*0.95))) // 95%以上の成功率
	})

	t.Run("Redis Concurrent Operations", func(t *testing.T) {
		start := time.Now()
		successCount := 0
		var mu sync.Mutex
		var wg sync.WaitGroup

		for i := 0; i < numOperations; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				key := fmt.Sprintf("concurrent:throughput:%d", index)
				value := fmt.Sprintf("concurrent-value-%d", index)

				// Set操作
				err := test.cacheService.Set(context.Background(), key, value, time.Hour)
				if err == nil {
					// Get操作
					_, err = test.cacheService.Get(context.Background(), key)
					if err == nil {
						mu.Lock()
						successCount++
						mu.Unlock()
					}
				}
			}(i)
		}

		wg.Wait()
		duration := time.Since(start)
		throughput := float64(successCount) / duration.Seconds()

		t.Logf("Redis Concurrent Throughput: %.2f operations/second", throughput)
		t.Logf("Success Rate: %.2f%%", float64(successCount)/float64(numOperations)*100)
		t.Logf("Total Duration: %v", duration)

		assert.GreaterOrEqual(t, throughput, 2000.0)                                         // 最低2000 ops/s
		assert.GreaterOrEqual(t, successCount, int(math.Floor(float64(numOperations)*0.95))) // 95%以上の成功率
	})

	t.Run("Redis High Load Test", func(t *testing.T) {
		const highLoadOperations = 10000
		const numWorkers = 100

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
					key := fmt.Sprintf("highload:worker:%d:op:%d", workerID, i)
					value := fmt.Sprintf("highload-value-%d-%d", workerID, i)

					// Set操作
					err := test.cacheService.Set(context.Background(), key, value, time.Hour)
					if err == nil {
						// Get操作
						_, err = test.cacheService.Get(context.Background(), key)
						if err == nil {
							mu.Lock()
							successCount++
							mu.Unlock()
						}
					}
				}
			}(w)
		}

		wg.Wait()
		duration := time.Since(start)
		throughput := float64(successCount) / duration.Seconds()

		t.Logf("Redis High Load Throughput: %.2f operations/second", throughput)
		t.Logf("Success Rate: %.2f%%", float64(successCount)/float64(highLoadOperations)*100)
		t.Logf("Total Duration: %v", duration)
		t.Logf("Workers: %d", numWorkers)

		assert.GreaterOrEqual(t, throughput, 5000.0)                                              // 最低5000 ops/s
		assert.GreaterOrEqual(t, successCount, int(math.Floor(float64(highLoadOperations)*0.90))) // 90%以上の成功率
	})
}

func TestRedisLatency(t *testing.T) {
	test := setupRedisTest(t)
	defer test.cleanup()

	const numOperations = 100

	t.Run("Redis Set Latency Distribution", func(t *testing.T) {
		var latencies []time.Duration
		successCount := 0

		for i := 0; i < numOperations; i++ {
			key := fmt.Sprintf("latency:set:%d", i)
			value := fmt.Sprintf("latency-value-%d", i)

			start := time.Now()
			err := test.cacheService.Set(context.Background(), key, value, time.Hour)
			latency := time.Since(start)

			if err == nil {
				latencies = append(latencies, latency)
				successCount++
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

		t.Logf("Redis Set Latency Statistics:")
		t.Logf("  Average: %v", avg)
		t.Logf("  Min: %v", min)
		t.Logf("  Max: %v", max)
		t.Logf("  Success Rate: %.2f%%", float64(successCount)/float64(numOperations)*100)

		// パフォーマンス要件の検証
		assert.LessOrEqual(t, avg, 10*time.Millisecond) // 平均10ms以下
		assert.LessOrEqual(t, max, 50*time.Millisecond) // 最大50ms以下
	})

	t.Run("Redis Get Latency Distribution", func(t *testing.T) {
		// テストデータの準備
		for i := 0; i < numOperations; i++ {
			key := fmt.Sprintf("latency:get:prep:%d", i)
			value := fmt.Sprintf("latency-value-%d", i)
			test.cacheService.Set(context.Background(), key, value, time.Hour)
		}

		var latencies []time.Duration
		successCount := 0

		for i := 0; i < numOperations; i++ {
			key := fmt.Sprintf("latency:get:prep:%d", i)

			start := time.Now()
			_, err := test.cacheService.Get(context.Background(), key)
			latency := time.Since(start)

			if err == nil {
				latencies = append(latencies, latency)
				successCount++
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

		t.Logf("Redis Get Latency Statistics:")
		t.Logf("  Average: %v", avg)
		t.Logf("  Min: %v", min)
		t.Logf("  Max: %v", max)
		t.Logf("  Success Rate: %.2f%%", float64(successCount)/float64(numOperations)*100)

		// パフォーマンス要件の検証
		assert.LessOrEqual(t, avg, 5*time.Millisecond)  // 平均5ms以下
		assert.LessOrEqual(t, max, 20*time.Millisecond) // 最大20ms以下
	})

	t.Run("Redis Latency Under Load", func(t *testing.T) {
		const loadOperations = 1000
		const concurrentUsers = 20

		var latencies []time.Duration
		var mu sync.Mutex
		var wg sync.WaitGroup

		operationsPerUser := loadOperations / concurrentUsers

		for u := 0; u < concurrentUsers; u++ {
			wg.Add(1)
			go func(userID int) {
				defer wg.Done()

				for i := 0; i < operationsPerUser; i++ {
					key := fmt.Sprintf("load:user:%d:op:%d", userID, i)
					value := fmt.Sprintf("load-value-%d-%d", userID, i)

					start := time.Now()
					err := test.cacheService.Set(context.Background(), key, value, time.Hour)
					latency := time.Since(start)

					if err == nil {
						mu.Lock()
						latencies = append(latencies, latency)
						mu.Unlock()
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

		t.Logf("Redis Latency Under Load:")
		t.Logf("  Average: %v", avg)
		t.Logf("  Concurrent Users: %d", concurrentUsers)
		t.Logf("  Total Operations: %d", len(latencies))

		assert.LessOrEqual(t, avg, 20*time.Millisecond) // 負荷下でも平均20ms以下
	})
}

func TestRedisMemoryUsage(t *testing.T) {
	test := setupRedisTest(t)
	defer test.cleanup()

	const numOperations = 10000

	t.Run("Redis Memory Usage Under Load", func(t *testing.T) {
		// メモリ使用量のベースラインを取得
		var m1, m2 runtime.MemStats
		runtime.ReadMemStats(&m1)

		var wg sync.WaitGroup
		for i := 0; i < numOperations; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				key := fmt.Sprintf("memory:test:%d", index)
				value := fmt.Sprintf("memory-value-%d", index)
				test.cacheService.Set(context.Background(), key, value, time.Hour)
			}(i)
		}

		wg.Wait()

		// GCを強制実行
		runtime.GC()
		runtime.ReadMemStats(&m2)

		memoryIncrease := m2.Alloc - m1.Alloc
		memoryIncreaseMB := float64(memoryIncrease) / 1024 / 1024

		t.Logf("Redis Memory Usage:")
		t.Logf("  Initial: %.2f MB", float64(m1.Alloc)/1024/1024)
		t.Logf("  Final: %.2f MB", float64(m2.Alloc)/1024/1024)
		t.Logf("  Increase: %.2f MB", memoryIncreaseMB)

		// メモリリークがないことを確認（増加が10MB以下）
		assert.LessOrEqual(t, memoryIncreaseMB, 10.0)
	})

	t.Run("Redis Memory Usage Over Time", func(t *testing.T) {
		const iterations = 10
		const operationsPerIteration = 1000

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

					key := fmt.Sprintf("memory:time:%d:%d", iter, index)
					value := fmt.Sprintf("memory-time-value-%d-%d", iter, index)
					test.cacheService.Set(context.Background(), key, value, time.Hour)
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

		t.Logf("Redis Memory Usage Over Time:")
		t.Logf("  Initial: %.2f MB", initialMemory)
		t.Logf("  Final: %.2f MB", finalMemory)
		t.Logf("  Growth: %.2f MB", memoryGrowth)
		t.Logf("  Iterations: %d", iterations)

		// メモリリークがないことを確認（成長が10MB以下）
		assert.LessOrEqual(t, memoryGrowth, 10.0)
	})
}

func TestRedisStressTest(t *testing.T) {
	test := setupRedisTest(t)
	defer test.cleanup()

	t.Run("Redis Sustained Load", func(t *testing.T) {
		const duration = 30 * time.Second
		const operationsPerSecond = 1000

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
					key := fmt.Sprintf("stress:%d", time.Now().UnixNano())
					value := fmt.Sprintf("stress-value-%d", time.Now().UnixNano())

					err := test.cacheService.Set(context.Background(), key, value, time.Hour)
					if err == nil {
						mu.Lock()
						successCount++
						mu.Unlock()
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

		t.Logf("Redis Stress Test Results:")
		t.Logf("  Duration: %v", elapsed)
		t.Logf("  Total Operations: %d", totalOperations)
		t.Logf("  Successful Operations: %d", successCount)
		t.Logf("  Throughput: %.2f operations/second", actualThroughput)
		t.Logf("  Success Rate: %.2f%%", successRate)

		assert.GreaterOrEqual(t, actualThroughput, float64(operationsPerSecond)*0.5) // 50%以上のスループット
		assert.GreaterOrEqual(t, successRate, 90.0)                                  // 90%以上の成功率
	})

	t.Run("Redis Burst Load", func(t *testing.T) {
		const burstSize = 1000
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

					key := fmt.Sprintf("burst:%d:%d", burst, index)
					value := fmt.Sprintf("burst-value-%d-%d", burst, index)

					err := test.cacheService.Set(context.Background(), key, value, time.Hour)
					if err == nil {
						successCount++
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

		t.Logf("Redis Burst Load Test Results:")
		t.Logf("  Total Bursts: %d", numBursts)
		t.Logf("  Burst Size: %d", burstSize)
		t.Logf("  Total Operations: %d", totalOperations)
		t.Logf("  Successful Operations: %d", totalSuccess)
		t.Logf("  Success Rate: %.2f%%", successRate)

		assert.GreaterOrEqual(t, successRate, 85.0) // 85%以上の成功率
	})
}

// setupRedisTest はRedisテストのセットアップを行います
func setupRedisTest(t testing.TB) *RedisPerformanceTest {
	// Redis接続
	cacheService := redis.NewCacheService("redis:6379")
	err := cacheService.Connect()
	require.NoError(t, err)

	return &RedisPerformanceTest{
		cacheService: cacheService,
	}
}

// cleanup はRedisテストのクリーンアップを行います
func (rpt *RedisPerformanceTest) cleanup() {
	if rpt.cacheService != nil {
		rpt.cacheService.Close()
	}
}
