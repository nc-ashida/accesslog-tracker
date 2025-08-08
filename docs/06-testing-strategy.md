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
```javascript
// tests/integration/api/tracking.test.js
const request = require('supertest');
const app = require('../../../src/app');
const { setupTestDatabase, cleanupTestDatabase } = require('../../helpers/database');

describe('Tracking API', () => {
  beforeAll(async () => {
    await setupTestDatabase();
  });
  
  afterAll(async () => {
    await cleanupTestDatabase();
  });
  
  describe('POST /v1/track', () => {
    test('should accept valid tracking data', async () => {
      const trackingData = {
        app_id: 'test_app_123',
        user_agent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
        url: 'https://example.com',
        session_id: 'alt_1234567890_abc123'
      };
      
      const response = await request(app)
        .post('/v1/track')
        .set('X-API-Key', 'test_api_key')
        .send(trackingData)
        .expect(200);
      
      expect(response.body.success).toBe(true);
      expect(response.body.data.tracking_id).toBeDefined();
    });
    
    test('should reject invalid API key', async () => {
      const trackingData = {
        app_id: 'test_app_123',
        user_agent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
        url: 'https://example.com'
      };
      
      const response = await request(app)
        .post('/v1/track')
        .set('X-API-Key', 'invalid_key')
        .send(trackingData)
        .expect(401);
      
      expect(response.body.success).toBe(false);
      expect(response.body.error.code).toBe('AUTHENTICATION_ERROR');
    });
    
    test('should handle rate limiting', async () => {
      const trackingData = {
        app_id: 'test_app_123',
        user_agent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
        url: 'https://example.com'
      };
      
      // 1001回リクエストを送信（制限: 1000 req/min）
      const requests = Array.from({ length: 1001 }, () =>
        request(app)
          .post('/v1/track')
          .set('X-API-Key', 'test_api_key')
          .send(trackingData)
      );
      
      const responses = await Promise.all(requests);
      const rateLimitedResponses = responses.filter(r => r.status === 429);
      
      expect(rateLimitedResponses.length).toBeGreaterThan(0);
    });
  });
  
  describe('GET /v1/statistics', () => {
    test('should return statistics for valid app_id', async () => {
      const response = await request(app)
        .get('/v1/statistics')
        .set('X-API-Key', 'test_api_key')
        .query({
          app_id: 'test_app_123',
          start_date: '2024-01-01',
          end_date: '2024-01-31'
        })
        .expect(200);
      
      expect(response.body.success).toBe(true);
      expect(response.body.data).toHaveProperty('total_requests');
      expect(response.body.data).toHaveProperty('unique_visitors');
    });
  });
});
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
```javascript
// tests/e2e/tracking-beacon.test.js
const puppeteer = require('puppeteer');

describe('Tracking Beacon E2E', () => {
  let browser;
  let page;
  
  beforeAll(async () => {
    browser = await puppeteer.launch({
      headless: true,
      args: ['--no-sandbox', '--disable-setuid-sandbox']
    });
  });
  
  afterAll(async () => {
    await browser.close();
  });
  
  beforeEach(async () => {
    page = await browser.newPage();
    
    // ネットワークリクエストを監視
    await page.setRequestInterception(true);
    const requests = [];
    page.on('request', request => {
      requests.push(request.url());
      request.continue();
    });
    
    page.requests = requests;
  });
  
  test('should send tracking data when page loads', async () => {
    // テスト用HTMLページを作成
    const testHtml = `
      <!DOCTYPE html>
      <html>
        <head>
          <script>
            window.ALT_CONFIG = {
              app_id: 'test_app_123',
              endpoint: 'http://localhost:3000/v1/track'
            };
          </script>
          <script async src="http://localhost:3000/tracker.js"></script>
        </head>
        <body>
          <h1>Test Page</h1>
        </body>
      </html>
    `;
    
    await page.setContent(testHtml);
    
    // ページ読み込み完了を待機
    await page.waitForTimeout(2000);
    
    // トラッキングリクエストが送信されたことを確認
    const trackingRequests = page.requests.filter(url => 
      url.includes('/v1/track')
    );
    
    expect(trackingRequests.length).toBeGreaterThan(0);
  });
  
  test('should respect DNT setting', async () => {
    // DNTヘッダーを設定
    await page.setExtraHTTPHeaders({
      'DNT': '1'
    });
    
    const testHtml = `
      <!DOCTYPE html>
      <html>
        <head>
          <script>
            window.ALT_CONFIG = {
              app_id: 'test_app_123',
              endpoint: 'http://localhost:3000/v1/track',
              respect_dnt: true
            };
          </script>
          <script async src="http://localhost:3000/tracker.js"></script>
        </head>
        <body>
          <h1>Test Page with DNT</h1>
        </body>
      </html>
    `;
    
    await page.setContent(testHtml);
    await page.waitForTimeout(2000);
    
    // DNTが有効な場合、トラッキングリクエストが送信されないことを確認
    const trackingRequests = page.requests.filter(url => 
      url.includes('/v1/track')
    );
    
    expect(trackingRequests.length).toBe(0);
  });
  
  test('should detect crawlers and skip tracking', async () => {
    // ユーザーエージェントをGooglebotに設定
    await page.setUserAgent('Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)');
    
    const testHtml = `
      <!DOCTYPE html>
      <html>
        <head>
          <script>
            window.ALT_CONFIG = {
              app_id: 'test_app_123',
              endpoint: 'http://localhost:3000/v1/track'
            };
          </script>
          <script async src="http://localhost:3000/tracker.js"></script>
        </head>
        <body>
          <h1>Test Page for Crawler</h1>
        </body>
      </html>
    `;
    
    await page.setContent(testHtml);
    await page.waitForTimeout(2000);
    
    // クローラーの場合、トラッキングリクエストが送信されないことを確認
    const trackingRequests = page.requests.filter(url => 
      url.includes('/v1/track')
    );
    
    expect(trackingRequests.length).toBe(0);
  });
});
```

## 6. パフォーマンステスト

### 6.1 負荷テスト
```javascript
// tests/performance/load-test.yml
config:
  target: 'http://localhost:3000'
  phases:
    - duration: 60
      arrivalRate: 10
      name: "Warm up"
    - duration: 300
      arrivalRate: 100
      name: "Sustained load"
    - duration: 60
      arrivalRate: 500
      name: "Peak load"
  defaults:
    headers:
      X-API-Key: 'test_api_key'

scenarios:
  - name: "Tracking data submission"
    weight: 80
    requests:
      - post:
          url: "/v1/track"
          json:
            app_id: "perf_test_app"
            user_agent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)"
            url: "https://example.com"
            session_id: "alt_{{ $randomNumber }}_{{ $randomString }}"
  
  - name: "Statistics retrieval"
    weight: 20
    requests:
      - get:
          url: "/v1/statistics"
          qs:
            app_id: "perf_test_app"
            start_date: "2024-01-01"
            end_date: "2024-01-31"
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
```javascript
// tests/security/authentication.test.js
const request = require('supertest');
const app = require('../../src/app');

describe('Security Tests', () => {
  describe('API Key Authentication', () => {
    test('should reject requests without API key', async () => {
      const response = await request(app)
        .post('/v1/track')
        .send({
          app_id: 'test_app_123',
          user_agent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
          url: 'https://example.com'
        })
        .expect(401);
      
      expect(response.body.error.code).toBe('AUTHENTICATION_ERROR');
    });
    
    test('should reject expired API key', async () => {
      const response = await request(app)
        .post('/v1/track')
        .set('X-API-Key', 'expired_api_key')
        .send({
          app_id: 'test_app_123',
          user_agent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
          url: 'https://example.com'
        })
        .expect(401);
      
      expect(response.body.error.code).toBe('AUTHENTICATION_ERROR');
    });
  });
  
  describe('Input Validation', () => {
    test('should reject SQL injection attempts', async () => {
      const maliciousData = {
        app_id: "'; DROP TABLE access_logs; --",
        user_agent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
        url: 'https://example.com'
      };
      
      const response = await request(app)
        .post('/v1/track')
        .set('X-API-Key', 'test_api_key')
        .send(maliciousData)
        .expect(400);
      
      expect(response.body.error.code).toBe('VALIDATION_ERROR');
    });
    
    test('should reject XSS attempts', async () => {
      const maliciousData = {
        app_id: 'test_app_123',
        user_agent: '<script>alert("XSS")</script>',
        url: 'https://example.com'
      };
      
      const response = await request(app)
        .post('/v1/track')
        .set('X-API-Key', 'test_api_key')
        .send(maliciousData)
        .expect(400);
      
      expect(response.body.error.code).toBe('VALIDATION_ERROR');
    });
  });
  
  describe('Rate Limiting', () => {
    test('should enforce rate limits per API key', async () => {
      const trackingData = {
        app_id: 'test_app_123',
        user_agent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
        url: 'https://example.com'
      };
      
      // 制限を超えるリクエストを送信
      const requests = Array.from({ length: 1001 }, () =>
        request(app)
          .post('/v1/track')
          .set('X-API-Key', 'test_api_key')
          .send(trackingData)
      );
      
      const responses = await Promise.all(requests);
      const rateLimitedResponses = responses.filter(r => r.status === 429);
      
      expect(rateLimitedResponses.length).toBeGreaterThan(0);
    });
  });
});
```

## 8. テスト実行とレポート

### 8.1 テスト実行スクリプト
```json
// package.json
{
  "scripts": {
    "test:all": "npm run test:unit && npm run test:integration && npm run test:e2e",
    "test:ci": "npm run test:coverage && npm run test:performance",
    "test:security": "npm run test:security",
    "test:report": "jest --coverage --coverageReporters=html --coverageReporters=text"
  }
}
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
```javascript
// tests/report-generator.js
class TestReportGenerator {
  static generateReport(testResults) {
    return {
      summary: {
        total: testResults.numTotalTests,
        passed: testResults.numPassedTests,
        failed: testResults.numFailedTests,
        coverage: testResults.coverage
      },
      performance: {
        averageResponseTime: testResults.avgResponseTime,
        throughput: testResults.throughput,
        errorRate: testResults.errorRate
      },
      security: {
        vulnerabilities: testResults.vulnerabilities,
        authenticationTests: testResults.authTests,
        authorizationTests: testResults.authzTests
      }
    };
  }
}
``` 