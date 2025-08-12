# テスト戦略仕様書

## 1. 概要

### 1.1 テスト方針
- **品質保証**: 高可用性と高パフォーマンスの確保
- **自動化**: CI/CDパイプラインでの自動テスト実行
- **包括性**: 単体テスト、統合テスト、E2Eテストの網羅
- **継続的改善**: テスト結果に基づく品質向上

### 1.2 テストレベル
1. **単体テスト**: 個別の関数・クラスのテスト
2. **統合テスト**: APIエンドポイント・データベース連携のテスト
3. **E2Eテスト**: トラッキングビーコンの動作テスト
4. **パフォーマンステスト**: 負荷・スループットテスト
5. **セキュリティテスト**: 脆弱性・認証テスト

## 2. テスト環境

### 2.1 環境構成
```yaml
# test-environments.yml
environments:
  unit:
    database: sqlite
    redis: mock
    dependencies: mocked
    
  integration:
    database: postgresql_test
    redis: redis_test
    dependencies: real
    
  e2e:
    database: postgresql_e2e
    redis: redis_e2e
    browser: puppeteer
    
  performance:
    database: postgresql_perf
    redis: redis_perf
    load_generator: artillery
```

### 2.2 Dockerコンテナ環境でのテスト実行
```bash
# 開発環境の起動
make dev-up

# テスト実行（Dockerコンテナ内で実行）
make test-in-container

# 統合テスト実行（Dockerコンテナ環境を使用）
make test-integration-container

# E2Eテスト実行（Dockerコンテナ環境を使用）
make test-e2e-container

# パフォーマンステスト実行（Dockerコンテナ環境を使用）
make test-performance-container
```

### 2.3 コンテナ内テスト実行の設定
```yaml
# docker-compose.test.yml
version: '3.8'

services:
  test-runner:
    build:
      context: .
      dockerfile: Dockerfile.test
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=access_log_tracker_test
      - DB_USER=postgres
      - DB_PASSWORD=password
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_PASSWORD=
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    volumes:
      - ./:/app
      - /app/vendor
    command: ["make", "test-all"]

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: access_log_tracker_test
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    ports:
      - "18433:5432"
    volumes:
      - postgres_test_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres -d access_log_tracker_test"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "16380:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres_test_data:
```

### 2.2 テストデータ管理
```javascript
// test-data-generator.js
class TestDataGenerator {
  static generateTrackingData(count = 100) {
    return Array.from({ length: count }, () => ({
      app_id: `test_app_${Math.random().toString(36).substr(2, 9)}`,
      client_sub_id: `sub_${Math.random().toString(36).substr(2, 9)}`,
      module_id: `module_${Math.random().toString(36).substr(2, 9)}`,
      url: `https://example.com/page/${Math.floor(Math.random() * 1000)}`,
      user_agent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
      ip_address: `192.168.1.${Math.floor(Math.random() * 255)}`,
      session_id: `alt_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
    }));
  }
  
  static generateApplicationData() {
    return {
      name: `Test App ${Date.now()}`,
      description: 'Test application for unit testing',
      domain: 'test.example.com',
      api_key: `test_key_${Math.random().toString(36).substr(2, 20)}`
    };
  }
}
```

## 3. 単体テスト

### 3.1 テストフレームワーク
```go
// go.mod
module access-log-tracker

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/lib/pq v1.10.9
    github.com/go-redis/redis/v8 v8.11.5
    github.com/stretchr/testify v1.8.4
    github.com/golang-migrate/migrate/v4 v4.16.2
    github.com/prometheus/client_golang v1.17.0
    github.com/gin-contrib/prometheus v0.0.0-20230501144526-8c036d44e6b7
)

// Makefile
test:
	go test ./...
test-unit:
	go test ./tests/unit/...
test-integration:
	go test ./tests/integration/...
test-e2e:
	go test ./tests/e2e/...
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
```

### 3.2 ユーティリティ関数のテスト
```go
// tests/unit/utils/tracking_validator_test.go
package utils_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "access-log-tracker/internal/domain/validators"
    "access-log-tracker/internal/domain/models"
)

func TestTrackingValidator_ValidateTrackingData(t *testing.T) {
    validator := validators.NewTrackingValidator()
    
    tests := []struct {
        name    string
        data    models.TrackingRequest
        isValid bool
        errors  []string
    }{
        {
            name: "valid tracking data",
            data: models.TrackingRequest{
                AppID:     "test_app_123",
                UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
                URL:       "https://example.com",
            },
            isValid: true,
            errors:  []string{},
        },
        {
            name: "missing app_id",
            data: models.TrackingRequest{
                UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
                URL:       "https://example.com",
            },
            isValid: false,
            errors:  []string{"app_id is required"},
        },
        {
            name: "invalid URL format",
            data: models.TrackingRequest{
                AppID:     "test_app_123",
                UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
                URL:       "invalid-url",
            },
            isValid: false,
            errors:  []string{"Invalid URL format"},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := validator.Validate(&tt.data)
            
            if tt.isValid {
                assert.NoError(t, result)
            } else {
                assert.Error(t, result)
                assert.Contains(t, result.Error(), tt.errors[0])
            }
        })
    }
}

func TestTrackingValidator_IsCrawler(t *testing.T) {
    validator := validators.NewTrackingValidator()
    
    tests := []struct {
        name      string
        userAgent string
        expected  bool
    }{
        {
            name:      "detect Googlebot",
            userAgent: "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
            expected:  true,
        },
        {
            name:      "detect Bingbot",
            userAgent: "Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)",
            expected:  true,
        },
        {
            name:      "regular browser",
            userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
            expected:  false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := validator.IsCrawler(tt.userAgent)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### 3.3 データベース関数のテスト
```go
// tests/unit/infrastructure/database/tracking_repository_test.go
package database_test

import (
    "database/sql"
    "testing"
    "time"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "access-log-tracker/internal/infrastructure/database/postgresql/repositories"
    "access-log-tracker/internal/domain/models"
)

type MockDB struct {
    mock.Mock
}

func (m *MockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
    mockArgs := m.Called(query, args)
    return mockArgs.Get(0).(sql.Result), mockArgs.Error(1)
}

func (m *MockDB) QueryRow(query string, args ...interface{}) *sql.Row {
    mockArgs := m.Called(query, args)
    return mockArgs.Get(0).(*sql.Row)
}

func TestTrackingRepository_SaveTrackingData(t *testing.T) {
    mockDB := &MockDB{}
    repository := repositories.NewTrackingRepository(mockDB)
    
    trackingData := &models.TrackingData{
        AppID:     "test_app_123",
        UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
        URL:       "https://example.com",
        CreatedAt: time.Now(),
    }
    
    // 成功ケース
    t.Run("should save tracking data successfully", func(t *testing.T) {
        mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(
            &MockResult{lastInsertId: 1}, nil,
        ).Once()
        
        err := repository.SaveTrackingData(trackingData)
        
        assert.NoError(t, err)
        mockDB.AssertExpectations(t)
    })
    
    // エラーケース
    t.Run("should handle database errors", func(t *testing.T) {
        mockDB.On("Exec", mock.AnythingOfType("string"), mock.Anything).Return(
            nil, sql.ErrConnDone,
        ).Once()
        
        err := repository.SaveTrackingData(trackingData)
        
        assert.Error(t, err)
        assert.Equal(t, sql.ErrConnDone, err)
    })
}

func TestTrackingRepository_GetStatistics(t *testing.T) {
    mockDB := &MockDB{}
    repository := repositories.NewTrackingRepository(mockDB)
    
    startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
    endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)
    
    t.Run("should return correct statistics", func(t *testing.T) {
        expectedStats := &models.Statistics{
            TotalRequests:  1000,
            UniqueVisitors: 500,
            UniqueSessions: 750,
        }
        
        mockRow := &MockRow{stats: expectedStats}
        mockDB.On("QueryRow", mock.AnythingOfType("string"), mock.Anything).Return(mockRow).Once()
        
        result, err := repository.GetStatistics("test_app_123", startDate, endDate)
        
        assert.NoError(t, err)
        assert.Equal(t, expectedStats, result)
        mockDB.AssertExpectations(t)
    })
}

// モックヘルパー
type MockResult struct {
    lastInsertId int64
    rowsAffected int64
}

func (m *MockResult) LastInsertId() (int64, error) {
    return m.lastInsertId, nil
}

func (m *MockResult) RowsAffected() (int64, error) {
    return m.rowsAffected, nil
}

type MockRow struct {
    stats *models.Statistics
}

func (m *MockRow) Scan(dest ...interface{}) error {
    if len(dest) >= 3 {
        if totalRequests, ok := dest[0].(*int64); ok {
            *totalRequests = m.stats.TotalRequests
        }
        if uniqueSessions, ok := dest[1].(*int64); ok {
            *uniqueSessions = m.stats.UniqueSessions
        }
        if uniqueVisitors, ok := dest[2].(*int64); ok {
            *uniqueVisitors = m.stats.UniqueVisitors
        }
    }
    return nil
}
```

## 4. 統合テスト

### 4.1 APIエンドポイントのテスト
```go
// tests/integration/api/tracking_test.go
package api_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "access-log-tracker/internal/api/handlers"
    "access-log-tracker/internal/api/middleware"
    "access-log-tracker/internal/domain/models"
    "access-log-tracker/internal/infrastructure/database/postgresql/repositories"
    "access-log-tracker/internal/infrastructure/cache/redis"
)

func setupTestServer(t *testing.T) *gin.Engine {
    // テスト用データベース接続
    db, err := setupTestDatabase()
    require.NoError(t, err)
    
    // テスト用Redis接続
    redisClient, err := setupTestRedis()
    require.NoError(t, err)
    
    // リポジトリの初期化
    trackingRepo := repositories.NewTrackingRepository(db)
    applicationRepo := repositories.NewApplicationRepository(db)
    
    // ハンドラーの初期化
    trackingHandler := handlers.NewTrackingHandler(trackingRepo, redisClient)
    applicationHandler := handlers.NewApplicationHandler(applicationRepo)
    
    // ルーターの設定
    router := gin.New()
    router.Use(gin.Recovery())
    router.Use(middleware.CORS())
    router.Use(middleware.Logging())
    router.Use(middleware.RateLimit(redisClient))
    
    // ルートの設定
    v1 := router.Group("/v1")
    {
        v1.POST("/track", middleware.Auth(applicationRepo), trackingHandler.Track)
        v1.GET("/statistics", middleware.Auth(applicationRepo), trackingHandler.GetStatistics)
        v1.POST("/applications", applicationHandler.Create)
    }
    
    return router
}

func TestTrackingAPI(t *testing.T) {
    router := setupTestServer(t)
    
    t.Run("POST /v1/track - should accept valid tracking data", func(t *testing.T) {
        trackingData := models.TrackingRequest{
            AppID:     "test_app_123",
            UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
            URL:       "https://example.com",
            SessionID: "alt_1234567890_abc123",
        }
        
        jsonData, _ := json.Marshal(trackingData)
        req := httptest.NewRequest("POST", "/v1/track", bytes.NewBuffer(jsonData))
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("X-API-Key", "test_api_key")
        
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusOK, w.Code)
        
        var response map[string]interface{}
        err := json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err)
        assert.Equal(t, true, response["success"])
        assert.NotNil(t, response["data"].(map[string]interface{})["tracking_id"])
    })
    
    t.Run("POST /v1/track - should reject invalid API key", func(t *testing.T) {
        trackingData := models.TrackingRequest{
            AppID:     "test_app_123",
            UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
            URL:       "https://example.com",
        }
        
        jsonData, _ := json.Marshal(trackingData)
        req := httptest.NewRequest("POST", "/v1/track", bytes.NewBuffer(jsonData))
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("X-API-Key", "invalid_key")
        
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusUnauthorized, w.Code)
        
        var response map[string]interface{}
        err := json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err)
        assert.Equal(t, false, response["success"])
        assert.Equal(t, "AUTHENTICATION_ERROR", response["error"].(map[string]interface{})["code"])
    })
    
    t.Run("POST /v1/track - should handle rate limiting", func(t *testing.T) {
        trackingData := models.TrackingRequest{
            AppID:     "test_app_123",
            UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
            URL:       "https://example.com",
        }
        
        jsonData, _ := json.Marshal(trackingData)
        
        // 1001回リクエストを送信（制限: 1000 req/min）
        rateLimitedCount := 0
        for i := 0; i < 1001; i++ {
            req := httptest.NewRequest("POST", "/v1/track", bytes.NewBuffer(jsonData))
            req.Header.Set("Content-Type", "application/json")
            req.Header.Set("X-API-Key", "test_api_key")
            
            w := httptest.NewRecorder()
            router.ServeHTTP(w, req)
            
            if w.Code == http.StatusTooManyRequests {
                rateLimitedCount++
            }
        }
        
        assert.Greater(t, rateLimitedCount, 0)
    })
    
    t.Run("GET /v1/statistics - should return statistics for valid app_id", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/v1/statistics?app_id=test_app_123&start_date=2024-01-01&end_date=2024-01-31", nil)
        req.Header.Set("X-API-Key", "test_api_key")
        
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusOK, w.Code)
        
        var response map[string]interface{}
        err := json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err)
        assert.Equal(t, true, response["success"])
        
        data := response["data"].(map[string]interface{})
        assert.NotNil(t, data["total_requests"])
        assert.NotNil(t, data["unique_visitors"])
    })
}

// テスト用データベースセットアップ
func setupTestDatabase() (*sql.DB, error) {
    // Dockerコンテナ内のPostgreSQLに接続
    dsn := "host=postgres port=5432 user=postgres password=password dbname=access_log_tracker_test sslmode=disable"
    return sql.Open("postgres", dsn)
}

// テスト用Redisセットアップ
func setupTestRedis() (*redis.Client, error) {
    // Dockerコンテナ内のRedisに接続
    rdb := redis.NewClient(&redis.Options{
        Addr:     "redis:6379",
        Password: "",
        DB:       0,
    })
    
    // 接続テスト
    _, err := rdb.Ping(context.Background()).Result()
    return rdb, err
}
```

### 4.2 データベース統合テスト
```javascript
// tests/integration/database/partitioning.test.js
const { createPartition, dropPartition } = require('../../../src/database/partition-manager');
const { TrackingRepository } = require('../../../src/database/tracking-repository');

describe('Database Partitioning', () => {
  let repository;
  
  beforeAll(async () => {
    repository = new TrackingRepository();
  });
  
  test('should create monthly partition', async () => {
    const partitionName = 'access_logs_2024_03';
    const startDate = '2024-03-01';
    const endDate = '2024-04-01';
    
    await createPartition(partitionName, startDate, endDate);
    
    // パーティションが作成されたことを確認
    const partitions = await repository.getPartitions();
    expect(partitions).toContain(partitionName);
  });
  
  test('should insert data into correct partition', async () => {
    const trackingData = {
      app_id: 'test_app_123',
      user_agent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
      url: 'https://example.com',
      created_at: new Date('2024-03-15')
    };
    
    const result = await repository.saveTrackingData(trackingData);
    
    // データが正しいパーティションに挿入されたことを確認
    const partitionData = await repository.getDataFromPartition('access_logs_2024_03');
    expect(partitionData).toContainEqual(
      expect.objectContaining({ id: result.id })
    );
  });
});
```

## 5. E2Eテスト

### 5.1 トラッキングビーコンのテスト
```go
// tests/e2e/tracking_beacon_test.go
package e2e_test

import (
    "context"
    "database/sql"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "access-log-tracker/internal/api/handlers"
    "access-log-tracker/internal/api/middleware"
    "access-log-tracker/internal/domain/models"
    "access-log-tracker/internal/infrastructure/database/postgresql/repositories"
    "access-log-tracker/internal/infrastructure/cache/redis"
)

func TestTrackingBeaconE2E(t *testing.T) {
    // テスト用サーバーのセットアップ
    server := setupE2EServer(t)
    defer server.Close()
    
    t.Run("should send tracking data when beacon loads", func(t *testing.T) {
        // テスト用アプリケーションを作成
        appID := createTestApplication(t, server)
        
        // トラッキングビーコンのリクエストをシミュレート
        trackingData := models.TrackingRequest{
            AppID:     appID,
            UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
            URL:       "https://example.com/test-page",
            SessionID: "alt_1234567890_abc123",
            IPAddress: "192.168.1.100",
        }
        
        // トラッキングリクエストを送信
        response := sendTrackingRequest(t, server, trackingData, "test_api_key")
        
        assert.Equal(t, http.StatusOK, response.Code)
        
        var responseBody map[string]interface{}
        err := json.Unmarshal(response.Body.Bytes(), &responseBody)
        assert.NoError(t, err)
        assert.Equal(t, true, responseBody["success"])
        
        // データベースにデータが保存されたことを確認
        trackingID := responseBody["data"].(map[string]interface{})["tracking_id"].(string)
        savedData := getTrackingDataFromDB(t, trackingID)
        assert.Equal(t, appID, savedData.AppID)
        assert.Equal(t, trackingData.URL, savedData.URL)
    })
    
    t.Run("should respect DNT setting", func(t *testing.T) {
        appID := createTestApplication(t, server)
        
        trackingData := models.TrackingRequest{
            AppID:     appID,
            UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
            URL:       "https://example.com/test-page",
            SessionID: "alt_1234567890_abc123",
            IPAddress: "192.168.1.100",
        }
        
        // DNTヘッダー付きでリクエストを送信
        req := createTrackingRequest(t, server, trackingData, "test_api_key")
        req.Header.Set("DNT", "1")
        
        w := httptest.NewRecorder()
        server.Config.Handler.ServeHTTP(w, req)
        
        // DNTが有効な場合、データが保存されないことを確認
        assert.Equal(t, http.StatusOK, w.Code)
        
        var responseBody map[string]interface{}
        err := json.Unmarshal(w.Body.Bytes(), &responseBody)
        assert.NoError(t, err)
        assert.Equal(t, true, responseBody["success"])
        
        // データベースにデータが保存されていないことを確認
        trackingID := responseBody["data"].(map[string]interface{})["tracking_id"].(string)
        savedData := getTrackingDataFromDB(t, trackingID)
        assert.Nil(t, savedData) // データが保存されていない
    })
    
    t.Run("should detect crawlers and skip tracking", func(t *testing.T) {
        appID := createTestApplication(t, server)
        
        trackingData := models.TrackingRequest{
            AppID:     appID,
            UserAgent: "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
            URL:       "https://example.com/test-page",
            SessionID: "alt_1234567890_abc123",
            IPAddress: "192.168.1.100",
        }
        
        // クローラーのユーザーエージェントでリクエストを送信
        response := sendTrackingRequest(t, server, trackingData, "test_api_key")
        
        assert.Equal(t, http.StatusOK, response.Code)
        
        var responseBody map[string]interface{}
        err := json.Unmarshal(response.Body.Bytes(), &responseBody)
        assert.NoError(t, err)
        assert.Equal(t, true, responseBody["success"])
        
        // クローラーの場合、データが保存されないことを確認
        trackingID := responseBody["data"].(map[string]interface{})["tracking_id"].(string)
        savedData := getTrackingDataFromDB(t, trackingID)
        assert.Nil(t, savedData) // データが保存されていない
    })
}

// E2Eテスト用サーバーセットアップ
func setupE2EServer(t *testing.T) *httptest.Server {
    // テスト用データベース接続
    db, err := setupE2EDatabase()
    require.NoError(t, err)
    
    // テスト用Redis接続
    redisClient, err := setupE2ERedis()
    require.NoError(t, err)
    
    // リポジトリの初期化
    trackingRepo := repositories.NewTrackingRepository(db)
    applicationRepo := repositories.NewApplicationRepository(db)
    
    // ハンドラーの初期化
    trackingHandler := handlers.NewTrackingHandler(trackingRepo, redisClient)
    applicationHandler := handlers.NewApplicationHandler(applicationRepo)
    
    // ルーターの設定
    router := setupTestRouter(trackingHandler, applicationHandler)
    
    return httptest.NewServer(router)
}

// テスト用データベースセットアップ（E2E用）
func setupE2EDatabase() (*sql.DB, error) {
    // Dockerコンテナ内のPostgreSQLに接続
    dsn := "host=postgres port=5432 user=postgres password=password dbname=access_log_tracker_e2e sslmode=disable"
    return sql.Open("postgres", dsn)
}

// テスト用Redisセットアップ（E2E用）
func setupE2ERedis() (*redis.Client, error) {
    // Dockerコンテナ内のRedisに接続
    rdb := redis.NewClient(&redis.Options{
        Addr:     "redis:6379",
        Password: "",
        DB:       1, // E2Eテスト用のDB
    })
    
    // 接続テスト
    _, err := rdb.Ping(context.Background()).Result()
    return rdb, err
}

// テスト用アプリケーション作成
func createTestApplication(t *testing.T, server *httptest.Server) string {
    appData := models.Application{
        Name:        "E2E Test App",
        Description: "Test application for E2E testing",
        Domain:      "e2e-test.example.com",
        APIKey:      "test_api_key",
    }
    
    jsonData, _ := json.Marshal(appData)
    req := httptest.NewRequest("POST", "/v1/applications", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    server.Config.Handler.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusCreated, w.Code)
    
    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    
    return response["data"].(map[string]interface{})["app_id"].(string)
}

// トラッキングリクエスト送信ヘルパー
func sendTrackingRequest(t *testing.T, server *httptest.Server, data models.TrackingRequest, apiKey string) *httptest.ResponseRecorder {
    req := createTrackingRequest(t, server, data, apiKey)
    w := httptest.NewRecorder()
    server.Config.Handler.ServeHTTP(w, req)
    return w
}

// トラッキングリクエスト作成ヘルパー
func createTrackingRequest(t *testing.T, server *httptest.Server, data models.TrackingRequest, apiKey string) *http.Request {
    jsonData, _ := json.Marshal(data)
    req := httptest.NewRequest("POST", "/v1/track", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-API-Key", apiKey)
    return req
}

// データベースからトラッキングデータ取得
func getTrackingDataFromDB(t *testing.T, trackingID string) *models.TrackingData {
    db, err := setupE2EDatabase()
    require.NoError(t, err)
    defer db.Close()
    
    var data models.TrackingData
    err = db.QueryRow("SELECT * FROM access_logs WHERE tracking_id = $1", trackingID).Scan(
        &data.ID, &data.TrackingID, &data.AppID, &data.UserAgent, &data.URL,
        &data.IPAddress, &data.SessionID, &data.CreatedAt,
    )
    
    if err == sql.ErrNoRows {
        return nil
    }
    require.NoError(t, err)
    
    return &data
}
```

## 6. パフォーマンステスト

### 6.1 負荷テスト
```go
// tests/performance/load_test.go
package performance_test

import (
    "context"
    "database/sql"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "sync"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "access-log-tracker/internal/api/handlers"
    "access-log-tracker/internal/api/middleware"
    "access-log-tracker/internal/domain/models"
    "access-log-tracker/internal/infrastructure/database/postgresql/repositories"
    "access-log-tracker/internal/infrastructure/cache/redis"
)

func TestLoadPerformance(t *testing.T) {
    server := setupPerformanceServer(t)
    defer server.Close()
    
    t.Run("should handle sustained load", func(t *testing.T) {
        const (
            concurrentUsers = 100
            requestsPerUser = 10
            totalRequests   = concurrentUsers * requestsPerUser
        )
        
        startTime := time.Now()
        
        // 並行リクエストを実行
        var wg sync.WaitGroup
        successCount := 0
        errorCount := 0
        var mu sync.Mutex
        
        for i := 0; i < concurrentUsers; i++ {
            wg.Add(1)
            go func(userID int) {
                defer wg.Done()
                
                for j := 0; j < requestsPerUser; j++ {
                    trackingData := models.TrackingRequest{
                        AppID:     "perf_test_app",
                        UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
                        URL:       "https://example.com",
                        SessionID: fmt.Sprintf("alt_%d_%d_%d", time.Now().Unix(), userID, j),
                    }
                    
                    response := sendTrackingRequest(t, server, trackingData, "test_api_key")
                    
                    mu.Lock()
                    if response.Code == http.StatusOK {
                        successCount++
                    } else {
                        errorCount++
                    }
                    mu.Unlock()
                }
            }(i)
        }
        
        wg.Wait()
        duration := time.Since(startTime)
        
        // パフォーマンス指標を計算
        requestsPerSecond := float64(totalRequests) / duration.Seconds()
        successRate := float64(successCount) / float64(totalRequests) * 100
        
        t.Logf("Total requests: %d", totalRequests)
        t.Logf("Duration: %v", duration)
        t.Logf("Requests per second: %.2f", requestsPerSecond)
        t.Logf("Success rate: %.2f%%", successRate)
        t.Logf("Errors: %d", errorCount)
        
        // パフォーマンス要件をチェック
        assert.GreaterOrEqual(t, requestsPerSecond, 100.0, "Should handle at least 100 requests per second")
        assert.GreaterOrEqual(t, successRate, 95.0, "Success rate should be at least 95%")
        assert.LessOrEqual(t, errorCount, totalRequests/20, "Error rate should be less than 5%")
    })
    
    t.Run("should handle peak load", func(t *testing.T) {
        const (
            peakUsers = 500
            duration  = 30 * time.Second
        )
        
        startTime := time.Now()
        var wg sync.WaitGroup
        successCount := 0
        errorCount := 0
        var mu sync.Mutex
        
        // 指定時間内で継続的にリクエストを送信
        for time.Since(startTime) < duration {
            for i := 0; i < peakUsers; i++ {
                wg.Add(1)
                go func() {
                    defer wg.Done()
                    
                    trackingData := models.TrackingRequest{
                        AppID:     "perf_test_app",
                        UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
                        URL:       "https://example.com",
                        SessionID: fmt.Sprintf("alt_%d_%d", time.Now().UnixNano(), i),
                    }
                    
                    response := sendTrackingRequest(t, server, trackingData, "test_api_key")
                    
                    mu.Lock()
                    if response.Code == http.StatusOK {
                        successCount++
                    } else {
                        errorCount++
                    }
                    mu.Unlock()
                }()
            }
            time.Sleep(100 * time.Millisecond) // 少し待機
        }
        
        wg.Wait()
        totalDuration := time.Since(startTime)
        
        totalRequests := successCount + errorCount
        requestsPerSecond := float64(totalRequests) / totalDuration.Seconds()
        successRate := float64(successCount) / float64(totalRequests) * 100
        
        t.Logf("Peak load test completed")
        t.Logf("Total requests: %d", totalRequests)
        t.Logf("Duration: %v", totalDuration)
        t.Logf("Requests per second: %.2f", requestsPerSecond)
        t.Logf("Success rate: %.2f%%", successRate)
        
        assert.GreaterOrEqual(t, requestsPerSecond, 200.0, "Should handle at least 200 requests per second under peak load")
        assert.GreaterOrEqual(t, successRate, 90.0, "Success rate should be at least 90% under peak load")
    })
}

// パフォーマンステスト用サーバーセットアップ
func setupPerformanceServer(t *testing.T) *httptest.Server {
    // テスト用データベース接続
    db, err := setupPerformanceDatabase()
    require.NoError(t, err)
    
    // テスト用Redis接続
    redisClient, err := setupPerformanceRedis()
    require.NoError(t, err)
    
    // リポジトリの初期化
    trackingRepo := repositories.NewTrackingRepository(db)
    applicationRepo := repositories.NewApplicationRepository(db)
    
    // ハンドラーの初期化
    trackingHandler := handlers.NewTrackingHandler(trackingRepo, redisClient)
    applicationHandler := handlers.NewApplicationHandler(applicationRepo)
    
    // ルーターの設定
    router := setupTestRouter(trackingHandler, applicationHandler)
    
    return httptest.NewServer(router)
}

// パフォーマンステスト用データベースセットアップ
func setupPerformanceDatabase() (*sql.DB, error) {
    // Dockerコンテナ内のPostgreSQLに接続
    dsn := "host=postgres port=5432 user=postgres password=password dbname=access_log_tracker_perf sslmode=disable"
    return sql.Open("postgres", dsn)
}

// パフォーマンステスト用Redisセットアップ
func setupPerformanceRedis() (*redis.Client, error) {
    // Dockerコンテナ内のRedisに接続
    rdb := redis.NewClient(&redis.Options{
        Addr:     "redis:6379",
        Password: "",
        DB:       2, // パフォーマンステスト用のDB
    })
    
    // 接続テスト
    _, err := rdb.Ping(context.Background()).Result()
    return rdb, err
}

// トラッキングリクエスト送信ヘルパー
func sendTrackingRequest(t *testing.T, server *httptest.Server, data models.TrackingRequest, apiKey string) *httptest.ResponseRecorder {
    jsonData, _ := json.Marshal(data)
    req := httptest.NewRequest("POST", "/v1/track", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-API-Key", apiKey)
    
    w := httptest.NewRecorder()
    server.Config.Handler.ServeHTTP(w, req)
    return w
}
```

### 6.2 スループットテスト
```javascript
// tests/performance/throughput.test.js
const { performance } = require('perf_hooks');
const { TrackingRepository } = require('../../src/database/tracking-repository');

describe('Performance Tests', () => {
  let repository;
  
  beforeAll(async () => {
    repository = new TrackingRepository();
  });
  
  test('should handle 1000 requests per second', async () => {
    const startTime = performance.now();
    const requestCount = 1000;
    
    const trackingDataArray = Array.from({ length: requestCount }, (_, i) => ({
      app_id: `perf_app_${i}`,
      user_agent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
      url: `https://example.com/page/${i}`,
      session_id: `alt_${Date.now()}_${i}`
    }));
    
    const promises = trackingDataArray.map(data => 
      repository.saveTrackingData(data)
    );
    
    await Promise.all(promises);
    
    const endTime = performance.now();
    const duration = (endTime - startTime) / 1000; // 秒単位
    const requestsPerSecond = requestCount / duration;
    
    expect(requestsPerSecond).toBeGreaterThan(1000);
  });
  
  test('should maintain performance under concurrent load', async () => {
    const concurrentUsers = 100;
    const requestsPerUser = 10;
    
    const userPromises = Array.from({ length: concurrentUsers }, async (_, userIndex) => {
      const userRequests = Array.from({ length: requestsPerUser }, (_, requestIndex) => ({
        app_id: `concurrent_app_${userIndex}`,
        user_agent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
        url: `https://example.com/user/${userIndex}/page/${requestIndex}`,
        session_id: `alt_${Date.now()}_${userIndex}_${requestIndex}`
      }));
      
      return Promise.all(userRequests.map(data => repository.saveTrackingData(data)));
    });
    
    const startTime = performance.now();
    await Promise.all(userPromises);
    const endTime = performance.now();
    
    const totalRequests = concurrentUsers * requestsPerUser;
    const duration = (endTime - startTime) / 1000;
    const requestsPerSecond = totalRequests / duration;
    
    expect(requestsPerSecond).toBeGreaterThan(500);
  });
});
```

## 7. セキュリティテスト

### 7.1 認証・認可テスト
```go
// tests/security/authentication_test.go
package security_test

import (
    "bytes"
    "context"
    "database/sql"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "access-log-tracker/internal/api/handlers"
    "access-log-tracker/internal/api/middleware"
    "access-log-tracker/internal/domain/models"
    "access-log-tracker/internal/infrastructure/database/postgresql/repositories"
    "access-log-tracker/internal/infrastructure/cache/redis"
)

func TestSecurityAuthentication(t *testing.T) {
    server := setupSecurityServer(t)
    defer server.Close()
    
    t.Run("should reject requests without API key", func(t *testing.T) {
        trackingData := models.TrackingRequest{
            AppID:     "test_app_123",
            UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
            URL:       "https://example.com",
        }
        
        jsonData, _ := json.Marshal(trackingData)
        req := httptest.NewRequest("POST", "/v1/track", bytes.NewBuffer(jsonData))
        req.Header.Set("Content-Type", "application/json")
        // X-API-Keyヘッダーを設定しない
        
        w := httptest.NewRecorder()
        server.Config.Handler.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusUnauthorized, w.Code)
        
        var response map[string]interface{}
        err := json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err)
        assert.Equal(t, "AUTHENTICATION_ERROR", response["error"].(map[string]interface{})["code"])
    })
    
    t.Run("should reject expired API key", func(t *testing.T) {
        trackingData := models.TrackingRequest{
            AppID:     "test_app_123",
            UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
            URL:       "https://example.com",
        }
        
        jsonData, _ := json.Marshal(trackingData)
        req := httptest.NewRequest("POST", "/v1/track", bytes.NewBuffer(jsonData))
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("X-API-Key", "expired_api_key")
        
        w := httptest.NewRecorder()
        server.Config.Handler.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusUnauthorized, w.Code)
        
        var response map[string]interface{}
        err := json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err)
        assert.Equal(t, "AUTHENTICATION_ERROR", response["error"].(map[string]interface{})["code"])
    })
}

func TestSecurityInputValidation(t *testing.T) {
    server := setupSecurityServer(t)
    defer server.Close()
    
    t.Run("should reject SQL injection attempts", func(t *testing.T) {
        maliciousData := models.TrackingRequest{
            AppID:     "'; DROP TABLE access_logs; --",
            UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
            URL:       "https://example.com",
        }
        
        jsonData, _ := json.Marshal(maliciousData)
        req := httptest.NewRequest("POST", "/v1/track", bytes.NewBuffer(jsonData))
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("X-API-Key", "test_api_key")
        
        w := httptest.NewRecorder()
        server.Config.Handler.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusBadRequest, w.Code)
        
        var response map[string]interface{}
        err := json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err)
        assert.Equal(t, "VALIDATION_ERROR", response["error"].(map[string]interface{})["code"])
    })
    
    t.Run("should reject XSS attempts", func(t *testing.T) {
        maliciousData := models.TrackingRequest{
            AppID:     "test_app_123",
            UserAgent: "<script>alert(\"XSS\")</script>",
            URL:       "https://example.com",
        }
        
        jsonData, _ := json.Marshal(maliciousData)
        req := httptest.NewRequest("POST", "/v1/track", bytes.NewBuffer(jsonData))
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("X-API-Key", "test_api_key")
        
        w := httptest.NewRecorder()
        server.Config.Handler.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusBadRequest, w.Code)
        
        var response map[string]interface{}
        err := json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err)
        assert.Equal(t, "VALIDATION_ERROR", response["error"].(map[string]interface{})["code"])
    })
}

func TestSecurityRateLimiting(t *testing.T) {
    server := setupSecurityServer(t)
    defer server.Close()
    
    t.Run("should enforce rate limits per API key", func(t *testing.T) {
        trackingData := models.TrackingRequest{
            AppID:     "test_app_123",
            UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
            URL:       "https://example.com",
        }
        
        jsonData, _ := json.Marshal(trackingData)
        
        // 制限を超えるリクエストを送信（1001回）
        rateLimitedCount := 0
        for i := 0; i < 1001; i++ {
            req := httptest.NewRequest("POST", "/v1/track", bytes.NewBuffer(jsonData))
            req.Header.Set("Content-Type", "application/json")
            req.Header.Set("X-API-Key", "test_api_key")
            
            w := httptest.NewRecorder()
            server.Config.Handler.ServeHTTP(w, req)
            
            if w.Code == http.StatusTooManyRequests {
                rateLimitedCount++
            }
        }
        
        assert.Greater(t, rateLimitedCount, 0, "Should enforce rate limiting")
    })
}

// セキュリティテスト用サーバーセットアップ
func setupSecurityServer(t *testing.T) *httptest.Server {
    // テスト用データベース接続
    db, err := setupSecurityDatabase()
    require.NoError(t, err)
    
    // テスト用Redis接続
    redisClient, err := setupSecurityRedis()
    require.NoError(t, err)
    
    // リポジトリの初期化
    trackingRepo := repositories.NewTrackingRepository(db)
    applicationRepo := repositories.NewApplicationRepository(db)
    
    // ハンドラーの初期化
    trackingHandler := handlers.NewTrackingHandler(trackingRepo, redisClient)
    applicationHandler := handlers.NewApplicationHandler(applicationRepo)
    
    // ルーターの設定
    router := setupTestRouter(trackingHandler, applicationHandler)
    
    return httptest.NewServer(router)
}

// セキュリティテスト用データベースセットアップ
func setupSecurityDatabase() (*sql.DB, error) {
    // Dockerコンテナ内のPostgreSQLに接続
    dsn := "host=postgres port=5432 user=postgres password=password dbname=access_log_tracker_security sslmode=disable"
    return sql.Open("postgres", dsn)
}

// セキュリティテスト用Redisセットアップ
func setupSecurityRedis() (*redis.Client, error) {
    // Dockerコンテナ内のRedisに接続
    rdb := redis.NewClient(&redis.Options{
        Addr:     "redis:6379",
        Password: "",
        DB:       3, // セキュリティテスト用のDB
    })
    
    // 接続テスト
    _, err := rdb.Ping(context.Background()).Result()
    return rdb, err
}
```

## 8. テスト実行とレポート

### 8.1 テスト実行スクリプト
```makefile
# Makefile テスト関連コマンド
.PHONY: test-all
test-all: ## すべてのテストを実行
	@echo "すべてのテストを実行中..."
	go test -v ./...

.PHONY: test-unit
test-unit: ## 単体テストを実行
	@echo "単体テストを実行中..."
	go test -v ./internal/domain/...
	go test -v ./internal/utils/...

.PHONY: test-integration
test-integration: ## 統合テストを実行
	@echo "統合テストを実行中..."
	go test -v ./internal/infrastructure/...
	go test -v ./internal/api/...

.PHONY: test-e2e
test-e2e: ## E2Eテストを実行
	@echo "E2Eテストを実行中..."
	go test -v ./tests/e2e/...

.PHONY: test-performance
test-performance: ## パフォーマンステストを実行
	@echo "パフォーマンステストを実行中..."
	go test -v -bench=. -benchmem ./tests/performance/...

.PHONY: test-security
test-security: ## セキュリティテストを実行
	@echo "セキュリティテストを実行中..."
	go test -v ./tests/security/...

# Dockerコンテナ環境でのテスト実行
.PHONY: test-in-container
test-in-container: ## Dockerコンテナ内でテストを実行
	@echo "Dockerコンテナ内でテストを実行中..."
	docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit

.PHONY: test-integration-container
test-integration-container: ## Dockerコンテナ環境で統合テストを実行
	@echo "Dockerコンテナ環境で統合テストを実行中..."
	docker-compose -f docker-compose.test.yml run --rm test-runner make test-integration

.PHONY: test-e2e-container
test-e2e-container: ## Dockerコンテナ環境でE2Eテストを実行
	@echo "Dockerコンテナ環境でE2Eテストを実行中..."
	docker-compose -f docker-compose.test.yml run --rm test-runner make test-e2e

.PHONY: test-performance-container
test-performance-container: ## Dockerコンテナ環境でパフォーマンステストを実行
	@echo "Dockerコンテナ環境でパフォーマンステストを実行中..."
	docker-compose -f docker-compose.test.yml run --rm test-runner make test-performance

.PHONY: test-coverage-container
test-coverage-container: ## Dockerコンテナ環境でカバレッジテストを実行
	@echo "Dockerコンテナ環境でカバレッジテストを実行中..."
	docker-compose -f docker-compose.test.yml run --rm test-runner make test-coverage
```

### 8.2 カバレッジ設定
```javascript
// jest.config.js
module.exports = {
  collectCoverageFrom: [
    'src/**/*.js',
    '!src/**/*.test.js',
    '!src/config/**',
    '!src/database/migrations/**'
  ],
  coverageThreshold: {
    global: {
      branches: 80,
      functions: 80,
      lines: 80,
      statements: 80
    }
  },
  coverageReporters: ['text', 'html', 'lcov'],
  testEnvironment: 'node',
  setupFilesAfterEnv: ['<rootDir>/tests/setup.js']
};
```

### 8.3 テストレポート
```go
// tests/report/report_generator.go
package report

import (
    "encoding/json"
    "fmt"
    "os"
    "time"
)

type TestReport struct {
    Summary struct {
        Total  int     `json:"total"`
        Passed int     `json:"passed"`
        Failed int     `json:"failed"`
        Coverage float64 `json:"coverage"`
    } `json:"summary"`
    Performance struct {
        AverageResponseTime float64 `json:"average_response_time"`
        Throughput         float64 `json:"throughput"`
        ErrorRate          float64 `json:"error_rate"`
    } `json:"performance"`
    Security struct {
        Vulnerabilities     []string `json:"vulnerabilities"`
        AuthenticationTests int      `json:"authentication_tests"`
        AuthorizationTests  int      `json:"authorization_tests"`
    } `json:"security"`
    Timestamp time.Time `json:"timestamp"`
}

type TestReportGenerator struct{}

func (g *TestReportGenerator) GenerateReport(testResults map[string]interface{}) *TestReport {
    report := &TestReport{
        Timestamp: time.Now(),
    }
    
    // サマリー情報を設定
    if summary, ok := testResults["summary"].(map[string]interface{}); ok {
        report.Summary.Total = int(summary["total"].(float64))
        report.Summary.Passed = int(summary["passed"].(float64))
        report.Summary.Failed = int(summary["failed"].(float64))
        report.Summary.Coverage = summary["coverage"].(float64)
    }
    
    // パフォーマンス情報を設定
    if performance, ok := testResults["performance"].(map[string]interface{}); ok {
        report.Performance.AverageResponseTime = performance["average_response_time"].(float64)
        report.Performance.Throughput = performance["throughput"].(float64)
        report.Performance.ErrorRate = performance["error_rate"].(float64)
    }
    
    // セキュリティ情報を設定
    if security, ok := testResults["security"].(map[string]interface{}); ok {
        if vulnerabilities, ok := security["vulnerabilities"].([]interface{}); ok {
            for _, v := range vulnerabilities {
                report.Security.Vulnerabilities = append(report.Security.Vulnerabilities, v.(string))
            }
        }
        report.Security.AuthenticationTests = int(security["authentication_tests"].(float64))
        report.Security.AuthorizationTests = int(security["authorization_tests"].(float64))
    }
    
    return report
}

func (g *TestReportGenerator) SaveReport(report *TestReport, filename string) error {
    data, err := json.MarshalIndent(report, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal report: %w", err)
    }
    
    return os.WriteFile(filename, data, 0644)
}
```

### 8.4 Dockerコンテナ環境でのテスト実行ガイド

#### 8.4.1 テスト環境の準備
```bash
# 開発環境を起動
make dev-up

# テスト用データベースの準備
docker-compose exec postgres psql -U postgres -c "CREATE DATABASE access_log_tracker_test;"
docker-compose exec postgres psql -U postgres -c "CREATE DATABASE access_log_tracker_e2e;"
docker-compose exec postgres psql -U postgres -c "CREATE DATABASE access_log_tracker_perf;"
docker-compose exec postgres psql -U postgres -c "CREATE DATABASE access_log_tracker_security;"
```

#### 8.4.2 テスト実行コマンド
```bash
# すべてのテストを実行
make test-all

# 単体テストのみ実行
make test-unit

# 統合テストのみ実行
make test-integration

# E2Eテストのみ実行
make test-e2e

# パフォーマンステストのみ実行
make test-performance

# セキュリティテストのみ実行
make test-security

# カバレッジテスト実行
make test-coverage
```

#### 8.4.3 Dockerコンテナ内でのテスト実行
```bash
# Dockerコンテナ内でテストを実行
make test-in-container

# 特定のテストをコンテナ内で実行
make test-integration-container
make test-e2e-container
make test-performance-container
make test-coverage-container
```

#### 8.4.4 テスト結果の確認
```bash
# テストログの確認
docker-compose logs test-runner

# カバレッジレポートの確認
open coverage.html

# テストレポートの確認
cat test-report.json
```

#### 8.4.5 CI/CDパイプラインでのテスト実行
```yaml
# .github/workflows/test.yml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_DB: access_log_tracker_test
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: password
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      
      redis:
        image: redis:7-alpine
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Run tests
      run: |
        make test-all
        make test-coverage
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
``` 