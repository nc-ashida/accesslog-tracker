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
```json
// package.json
{
  "devDependencies": {
    "jest": "^29.0.0",
    "supertest": "^6.0.0",
    "puppeteer": "^21.0.0",
    "artillery": "^2.0.0"
  },
  "scripts": {
    "test": "jest",
    "test:unit": "jest --testPathPattern=unit",
    "test:integration": "jest --testPathPattern=integration",
    "test:e2e": "jest --testPathPattern=e2e",
    "test:coverage": "jest --coverage",
    "test:watch": "jest --watch"
  }
}
```

### 3.2 ユーティリティ関数のテスト
```javascript
// tests/unit/utils/tracking-validator.test.js
const { validateTrackingData, isCrawler } = require('../../../src/utils/tracking-validator');

describe('Tracking Validator', () => {
  describe('validateTrackingData', () => {
    test('should validate correct tracking data', () => {
      const validData = {
        app_id: 'test_app_123',
        user_agent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
        url: 'https://example.com'
      };
      
      const result = validateTrackingData(validData);
      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });
    
    test('should reject missing app_id', () => {
      const invalidData = {
        user_agent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
        url: 'https://example.com'
      };
      
      const result = validateTrackingData(invalidData);
      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('app_id is required');
    });
    
    test('should reject invalid URL format', () => {
      const invalidData = {
        app_id: 'test_app_123',
        user_agent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
        url: 'invalid-url'
      };
      
      const result = validateTrackingData(invalidData);
      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('Invalid URL format');
    });
  });
  
  describe('isCrawler', () => {
    test('should detect Googlebot', () => {
      const userAgent = 'Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)';
      expect(isCrawler(userAgent)).toBe(true);
    });
    
    test('should detect Bingbot', () => {
      const userAgent = 'Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)';
      expect(isCrawler(userAgent)).toBe(true);
    });
    
    test('should not detect regular browser', () => {
      const userAgent = 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36';
      expect(isCrawler(userAgent)).toBe(false);
    });
  });
});
```

### 3.3 データベース関数のテスト
```javascript
// tests/unit/database/tracking-repository.test.js
const { TrackingRepository } = require('../../../src/database/tracking-repository');
const { mockDatabase } = require('../../mocks/database');

describe('TrackingRepository', () => {
  let repository;
  let mockDb;
  
  beforeEach(() => {
    mockDb = mockDatabase();
    repository = new TrackingRepository(mockDb);
  });
  
  describe('saveTrackingData', () => {
    test('should save tracking data successfully', async () => {
      const trackingData = {
        app_id: 'test_app_123',
        user_agent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
        url: 'https://example.com',
        created_at: new Date()
      };
      
      const result = await repository.saveTrackingData(trackingData);
      
      expect(result.id).toBeDefined();
      expect(mockDb.query).toHaveBeenCalledWith(
        expect.stringContaining('INSERT INTO access_logs'),
        expect.arrayContaining([trackingData.app_id])
      );
    });
    
    test('should handle database errors', async () => {
      mockDb.query.mockRejectedValue(new Error('Database connection failed'));
      
      const trackingData = {
        app_id: 'test_app_123',
        user_agent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
        url: 'https://example.com'
      };
      
      await expect(repository.saveTrackingData(trackingData))
        .rejects.toThrow('Database connection failed');
    });
  });
  
  describe('getStatistics', () => {
    test('should return correct statistics', async () => {
      const mockStats = {
        total_requests: 1000,
        unique_visitors: 500,
        unique_sessions: 750
      };
      
      mockDb.query.mockResolvedValue({ rows: [mockStats] });
      
      const result = await repository.getStatistics('test_app_123', {
        start_date: '2024-01-01',
        end_date: '2024-01-31'
      });
      
      expect(result).toEqual(mockStats);
      expect(mockDb.query).toHaveBeenCalledWith(
        expect.stringContaining('SELECT'),
        expect.arrayContaining(['test_app_123'])
      );
    });
  });
});
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