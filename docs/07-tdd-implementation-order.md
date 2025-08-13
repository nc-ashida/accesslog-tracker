# TDD実装順序ガイド

## 1. 概要

### 1.1 TDD実装方針
- **テストファースト**: 各機能の実装前にテストを書く
- **段階的実装**: 依存関係の少ない部分から順次実装
- **継続的テスト**: 実装とテストを並行して進める
- **品質保証**: 各段階で品質を確認しながら進める

### 1.2 実装フェーズ
1. **基盤フェーズ**: ユーティリティ・設定・インフラ層
2. **ドメインフェーズ**: ビジネスロジック・モデル・バリデーション
3. **インフラフェーズ**: データベース・キャッシュ・ストレージ
4. **APIフェーズ**: HTTPハンドラー・ミドルウェア・ルーティング
5. **ビーコンフェーズ**: JavaScriptビーコン・配信システム
6. **統合フェーズ**: E2Eテスト・パフォーマンステスト・セキュリティテスト

## 2. フェーズ1: 基盤フェーズ

### 2.1 プロジェクト初期化
```bash
# 1. プロジェクト構造作成
mkdir -p access-log-tracker/{cmd,internal,pkg,web,tests,configs,scripts}
mkdir -p access-log-tracker/internal/{api,domain,infrastructure,beacon,utils}
mkdir -p access-log-tracker/tests/{unit,integration,e2e,performance,security}

# 2. Goモジュール初期化
cd access-log-tracker
go mod init access-log-tracker

# 3. 基本設定ファイル作成
touch go.mod go.sum .gitignore README.md Makefile
```

### 2.2 ユーティリティ関数の実装（TDD）

#### 2.2.1 時間ユーティリティ
```go
// tests/unit/utils/timeutil_test.go
package utils_test

import (
    "testing"
    "time"
    "github.com/stretchr/testify/assert"
    "access-log-tracker/internal/utils/timeutil"
)

func TestTimeUtil_FormatTimestamp(t *testing.T) {
    tests := []struct {
        name     string
        input    time.Time
        expected string
    }{
        {
            name:     "format UTC timestamp",
            input:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
            expected: "2024-01-01T12:00:00Z",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := timeutil.FormatTimestamp(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

```go
// internal/utils/timeutil/timeutil.go
package timeutil

import "time"

func FormatTimestamp(t time.Time) string {
    return t.UTC().Format(time.RFC3339)
}
```

#### 2.2.2 IPユーティリティ
```go
// tests/unit/utils/iputil_test.go
package utils_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "access-log-tracker/internal/utils/iputil"
)

func TestIPUtil_IsValidIP(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected bool
    }{
        {
            name:     "valid IPv4",
            input:    "192.168.1.1",
            expected: true,
        },
        {
            name:     "invalid IP",
            input:    "invalid-ip",
            expected: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := iputil.IsValidIP(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### 2.3 設定管理の実装
```go
// tests/unit/config/config_test.go
package config_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "access-log-tracker/internal/config"
)

func TestConfig_Load(t *testing.T) {
    cfg := config.New()
    err := cfg.Load("testdata/config.yaml")
    
    assert.NoError(t, err)
    assert.Equal(t, "localhost", cfg.Database.Host)
    assert.Equal(t, 5432, cfg.Database.Port)
}
```

## 3. フェーズ2: ドメインフェーズ

### 3.1 ドメインモデルの実装

#### 3.1.1 トラッキングデータモデル
```go
// tests/unit/domain/models/tracking_test.go
package models_test

import (
    "testing"
    "time"
    "github.com/stretchr/testify/assert"
    "access-log-tracker/internal/domain/models"
)

func TestTrackingData_Validate(t *testing.T) {
    tests := []struct {
        name    string
        data    models.TrackingData
        isValid bool
    }{
        {
            name: "valid tracking data",
            data: models.TrackingData{
                AppID:     "test_app_123",
                UserAgent: "Mozilla/5.0",
                Timestamp: time.Now(),
            },
            isValid: true,
        },
        {
            name: "missing app_id",
            data: models.TrackingData{
                UserAgent: "Mozilla/5.0",
                Timestamp: time.Now(),
            },
            isValid: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.data.Validate()
            if tt.isValid {
                assert.NoError(t, err)
            } else {
                assert.Error(t, err)
            }
        })
    }
}
```

```go
// internal/domain/models/tracking.go
package models

import (
    "errors"
    "time"
)

type TrackingData struct {
    ID            string                 `json:"id"`
    AppID         string                 `json:"app_id"`
    ClientSubID   string                 `json:"client_sub_id,omitempty"`
    ModuleID      string                 `json:"module_id,omitempty"`
    URL           string                 `json:"url,omitempty"`
    Referrer      string                 `json:"referrer,omitempty"`
    UserAgent     string                 `json:"user_agent"`
    IPAddress     string                 `json:"ip_address,omitempty"`
    SessionID     string                 `json:"session_id,omitempty"`
    Timestamp     time.Time              `json:"timestamp"`
    CustomParams  map[string]interface{} `json:"custom_params,omitempty"`
}

func (t *TrackingData) Validate() error {
    if t.AppID == "" {
        return errors.New("app_id is required")
    }
    if t.UserAgent == "" {
        return errors.New("user_agent is required")
    }
    return nil
}
```

#### 3.1.2 アプリケーションモデル
```go
// tests/unit/domain/models/application_test.go
package models_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "access-log-tracker/internal/domain/models"
)

func TestApplication_Validate(t *testing.T) {
    tests := []struct {
        name    string
        app     models.Application
        isValid bool
    }{
        {
            name: "valid application",
            app: models.Application{
                AppID:    "test_app_123",
                Name:     "Test Application",
                Domain:   "example.com",
                APIKey:   "test-api-key",
            },
            isValid: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.app.Validate()
            if tt.isValid {
                assert.NoError(t, err)
            } else {
                assert.Error(t, err)
            }
        })
    }
}
```

### 3.2 ドメインサービスの実装

#### 3.2.1 トラッキングサービス
```go
// tests/unit/domain/services/tracking_service_test.go
package services_test

import (
    "testing"
    "time"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "access-log-tracker/internal/domain/services"
    "access-log-tracker/internal/domain/models"
)

func TestTrackingService_ProcessTrackingData(t *testing.T) {
    mockRepo := &MockTrackingRepository{}
    service := services.NewTrackingService(mockRepo)
    
    data := models.TrackingData{
        AppID:     "test_app_123",
        UserAgent: "Mozilla/5.0",
        Timestamp: time.Now(),
    }
    
    mockRepo.On("Save", mock.AnythingOfType("*models.TrackingData")).Return(nil)
    
    err := service.ProcessTrackingData(&data)
    
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}
```

## 4. フェーズ3: インフラフェーズ

### 4.1 データベース接続の実装

#### 4.1.1 PostgreSQL接続
```go
// tests/unit/infrastructure/database/postgresql/connection_test.go
package postgresql_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "access-log-tracker/internal/infrastructure/database/postgresql"
)

func TestPostgreSQLConnection_Connect(t *testing.T) {
    conn := postgresql.NewConnection("test")
    err := conn.Connect("host=localhost port=5432 user=test password=test dbname=test")
    
    assert.NoError(t, err)
    defer conn.Close()
}
```

#### 4.1.2 リポジトリの実装
```go
// tests/unit/infrastructure/database/postgresql/repositories/tracking_repository_test.go
package repositories_test

import (
    "testing"
    "time"
    "github.com/stretchr/testify/assert"
    "access-log-tracker/internal/infrastructure/database/postgresql/repositories"
    "access-log-tracker/internal/domain/models"
)

func TestTrackingRepository_Save(t *testing.T) {
    repo := repositories.NewTrackingRepository(testDB)
    
    data := &models.TrackingData{
        AppID:     "test_app_123",
        UserAgent: "Mozilla/5.0",
        Timestamp: time.Now(),
    }
    
    err := repo.Save(data)
    
    assert.NoError(t, err)
    assert.NotEmpty(t, data.ID)
}
```

### 4.2 キャッシュの実装
```go
// tests/unit/infrastructure/cache/redis/cache_service_test.go
package redis_test

import (
    "testing"
    "time"
    "github.com/stretchr/testify/assert"
    "access-log-tracker/internal/infrastructure/cache/redis"
)

func TestCacheService_SetAndGet(t *testing.T) {
    cache := redis.NewCacheService("localhost:6379")
    
    key := "test_key"
    value := "test_value"
    ttl := time.Minute
    
    err := cache.Set(key, value, ttl)
    assert.NoError(t, err)
    
    result, err := cache.Get(key)
    assert.NoError(t, err)
    assert.Equal(t, value, result)
}
```

## 5. フェーズ4: APIフェーズ

### 5.1 HTTPハンドラーの実装

#### 5.1.1 トラッキングハンドラー
```go
// tests/unit/api/handlers/tracking_test.go
package handlers_test

import (
    "testing"
    "net/http"
    "net/http/httptest"
    "strings"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "access-log-tracker/internal/api/handlers"
)

func TestTrackingHandler_Track(t *testing.T) {
    gin.SetMode(gin.TestMode)
    router := gin.New()
    
    mockService := &MockTrackingService{}
    handler := handlers.NewTrackingHandler(mockService)
    
    router.POST("/track", handler.Track)
    
    payload := `{
        "app_id": "test_app_123",
        "user_agent": "Mozilla/5.0"
    }`
    
    req := httptest.NewRequest("POST", "/track", strings.NewReader(payload))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-API-Key", "test-api-key")
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
}
```

#### 5.1.2 ミドルウェアの実装
```go
// tests/unit/api/middleware/auth_test.go
package middleware_test

import (
    "testing"
    "net/http"
    "net/http/httptest"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "access-log-tracker/internal/api/middleware"
)

func TestAuthMiddleware_ValidAPIKey(t *testing.T) {
    gin.SetMode(gin.TestMode)
    router := gin.New()
    
    authMiddleware := middleware.NewAuthMiddleware()
    router.Use(authMiddleware.Authenticate())
    
    router.GET("/test", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    req := httptest.NewRequest("GET", "/test", nil)
    req.Header.Set("X-API-Key", "valid-api-key")
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
}
```

### 5.2 ルーティングの実装
```go
// tests/unit/api/routes/routes_test.go
package routes_test

import (
    "testing"
    "net/http"
    "net/http/httptest"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "access-log-tracker/internal/api/routes"
)

func TestRoutes_Setup(t *testing.T) {
    gin.SetMode(gin.TestMode)
    router := gin.New()
    
    routes.Setup(router)
    
    req := httptest.NewRequest("GET", "/health", nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
}
```

## 6. フェーズ5: ビーコンフェーズ

### 6.1 JavaScriptビーコンの実装

#### 6.1.1 ビーコン生成器
```go
// tests/unit/beacon/generator/beacon_generator_test.go
package generator_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "access-log-tracker/internal/beacon/generator"
)

func TestBeaconGenerator_Generate(t *testing.T) {
    generator := generator.NewBeaconGenerator()
    
    config := generator.BeaconConfig{
        Endpoint: "https://api.example.com/track",
        Debug:    false,
    }
    
    result, err := generator.Generate(config)
    
    assert.NoError(t, err)
    assert.Contains(t, result, "function track")
    assert.Contains(t, result, config.Endpoint)
}
```

#### 6.1.2 ビーコン配信API
```go
// tests/unit/api/handlers/beacon_test.go
package handlers_test

import (
    "testing"
    "net/http"
    "net/http/httptest"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "access-log-tracker/internal/api/handlers"
)

func TestBeaconHandler_Serve(t *testing.T) {
    gin.SetMode(gin.TestMode)
    router := gin.New()
    
    handler := handlers.NewBeaconHandler()
    router.GET("/tracker.js", handler.Serve)
    
    req := httptest.NewRequest("GET", "/tracker.js", nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
    assert.Equal(t, "application/javascript", w.Header().Get("Content-Type"))
    assert.Contains(t, w.Body.String(), "function track")
}
```

## 7. フェーズ6: 統合フェーズ

### 7.1 E2Eテストの実装
```go
// tests/e2e/beacon_tracking_test.go
package e2e_test

import (
    "testing"
    "net/http"
    "time"
    "github.com/stretchr/testify/assert"
)

func TestBeaconTracking_EndToEnd(t *testing.T) {
    // 1. ビーコンの配信確認
    resp, err := http.Get("http://localhost:8080/tracker.js")
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    // 2. トラッキングデータの送信確認
    payload := `{
        "app_id": "test_app_123",
        "user_agent": "Mozilla/5.0"
    }`
    
    resp, err = http.Post("http://localhost:8080/v1/track", 
        "application/json", strings.NewReader(payload))
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
    
    // 3. データベースへの保存確認
    time.Sleep(1 * time.Second) // 非同期処理の完了待ち
    
    // データベースからデータを取得して確認
    // ...
}
```

### 7.2 パフォーマンステストの実装
```go
// tests/performance/beacon_performance_test.go
package performance_test

import (
    "testing"
    "net/http"
    "sync"
    "time"
    "github.com/stretchr/testify/assert"
)

func TestBeaconPerformance_ConcurrentRequests(t *testing.T) {
    const numRequests = 1000
    const concurrency = 100
    
    var wg sync.WaitGroup
    results := make(chan int, numRequests)
    
    start := time.Now()
    
    for i := 0; i < concurrency; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < numRequests/concurrency; j++ {
                resp, err := http.Post("http://localhost:8080/v1/track",
                    "application/json", strings.NewReader(`{"app_id":"test"}`))
                if err == nil {
                    results <- resp.StatusCode
                }
            }
        }()
    }
    
    wg.Wait()
    close(results)
    
    duration := time.Since(start)
    
    successCount := 0
    for statusCode := range results {
        if statusCode == http.StatusOK {
            successCount++
        }
    }
    
    assert.Equal(t, numRequests, successCount)
    assert.Less(t, duration, 10*time.Second) // 10秒以内に完了
}
```

## 8. 実装順序の詳細

### 8.1 フェーズ別実装順序

#### フェーズ1: 基盤フェーズ（1-2週間）
1. **プロジェクト初期化**
   - ディレクトリ構造作成
   - Goモジュール設定
   - 基本設定ファイル

2. **ユーティリティ関数**
   - 時間ユーティリティ（TDD）
   - IPユーティリティ（TDD）
   - JSONユーティリティ（TDD）
   - 暗号化ユーティリティ（TDD）

3. **設定管理**
   - 設定ファイル読み込み（TDD）
   - 環境変数管理（TDD）
   - バリデーション（TDD）

4. **ログ機能**
   - ロガー設定（TDD）
   - ログフォーマッター（TDD）

#### フェーズ2: ドメインフェーズ（2-3週間）
1. **ドメインモデル**
   - トラッキングデータモデル（TDD）
   - アプリケーションモデル（TDD）
   - セッションモデル（TDD）
   - カスタムパラメータモデル（TDD）

2. **バリデーター**
   - トラッキングバリデーター（TDD）
   - アプリケーションバリデーター（TDD）

3. **ドメインサービス**
   - トラッキングサービス（TDD）
   - アプリケーションサービス（TDD）
   - 統計サービス（TDD）

#### フェーズ3: インフラフェーズ（2-3週間）
1. **データベース接続**
   - PostgreSQL接続管理（TDD）
   - コネクションプール（TDD）

2. **リポジトリ実装**
   - トラッキングリポジトリ（TDD）
   - アプリケーションリポジトリ（TDD）
   - セッションリポジトリ（TDD）

3. **キャッシュ実装**
   - Redis接続（TDD）
   - キャッシュサービス（TDD）

4. **マイグレーション**
   - データベースマイグレーション（TDD）
   - パーティション管理（TDD）

#### フェーズ4: APIフェーズ（2-3週間）
1. **HTTPハンドラー**
   - トラッキングハンドラー（TDD）
   - ヘルスチェックハンドラー（TDD）
   - 統計ハンドラー（TDD）

2. **ミドルウェア**
   - 認証ミドルウェア（TDD）
   - CORSミドルウェア（TDD）
   - レート制限ミドルウェア（TDD）
   - ログミドルウェア（TDD）

3. **ルーティング**
   - APIルート設定（TDD）
   - エラーハンドリング（TDD）

4. **サーバー設定**
   - HTTPサーバー設定（TDD）
   - グレースフルシャットダウン（TDD）

#### フェーズ5: ビーコンフェーズ（1-2週間）
1. **ビーコン生成器**
   - JavaScriptテンプレート（TDD）
   - ビーコン生成器（TDD）
   - コード圧縮（TDD）

2. **ビーコン配信**
   - ビーコン配信API（TDD）
   - CloudFront設定（TDD）

#### フェーズ6: 統合フェーズ（2-3週間）
1. **E2Eテスト**
   - ビーコントラッキングテスト（TDD）
   - API統合テスト（TDD）

2. **パフォーマンステスト**
   - 負荷テスト（TDD）
   - スループットテスト（TDD）

3. **セキュリティテスト**
   - 認証テスト（TDD）
   - 入力値検証テスト（TDD）

### 8.2 依存関係マップ

```
基盤フェーズ
├── ユーティリティ関数
├── 設定管理
└── ログ機能

ドメインフェーズ
├── ドメインモデル (基盤フェーズに依存)
├── バリデーター (ドメインモデルに依存)
└── ドメインサービス (ドメインモデル・バリデーターに依存)

インフラフェーズ
├── データベース接続 (基盤フェーズに依存)
├── リポジトリ実装 (ドメインフェーズ・データベース接続に依存)
├── キャッシュ実装 (基盤フェーズに依存)
└── マイグレーション (データベース接続に依存)

APIフェーズ
├── HTTPハンドラー (ドメインフェーズ・インフラフェーズに依存)
├── ミドルウェア (基盤フェーズに依存)
├── ルーティング (HTTPハンドラー・ミドルウェアに依存)
└── サーバー設定 (ルーティングに依存)

ビーコンフェーズ
├── ビーコン生成器 (基盤フェーズに依存)
└── ビーコン配信 (APIフェーズに依存)

統合フェーズ
├── E2Eテスト (全フェーズに依存)
├── パフォーマンステスト (全フェーズに依存)
└── セキュリティテスト (全フェーズに依存)
```

## 9. テスト戦略

### 9.1 テストピラミッド
```
    E2Eテスト (10%)
   /           \
統合テスト (20%)  セキュリティテスト (5%)
   \           /
   単体テスト (65%)
```

### 9.2 テスト実行順序
1. **単体テスト**: 各機能の個別テスト
2. **統合テスト**: コンポーネント間の連携テスト
3. **E2Eテスト**: システム全体の動作テスト
4. **パフォーマンステスト**: 負荷・スループットテスト
5. **セキュリティテスト**: 脆弱性・認証テスト

### 9.3 継続的テスト
- **開発時**: 単体テストを頻繁に実行
- **コミット時**: 単体テスト + 統合テスト
- **プルリクエスト時**: 全テスト実行
- **デプロイ時**: E2Eテスト + パフォーマンステスト

## 10. 品質保証

### 10.1 コードカバレッジ
- **目標**: 80%以上
- **必須**: ビジネスロジック部分は90%以上
- **測定**: go test -coverprofile

### 10.2 静的解析
- **golangci-lint**: コード品質チェック
- **gosec**: セキュリティ脆弱性チェック
- **govet**: コードの問題点チェック

### 10.3 パフォーマンス基準
- **API応答時間**: 50ms以下
- **スループット**: 5000 req/sec以上
- **メモリ使用量**: 100MB以下
- **CPU使用率**: 70%以下

## 11. 実装チェックリスト

### フェーズ1: 基盤フェーズ
- [ ] プロジェクト構造作成
- [ ] Goモジュール設定
- [ ] ユーティリティ関数実装（TDD）
- [ ] 設定管理実装（TDD）
- [ ] ログ機能実装（TDD）
- [ ] 単体テスト作成
- [ ] 統合テスト作成

### フェーズ2: ドメインフェーズ
- [ ] ドメインモデル実装（TDD）
- [ ] バリデーター実装（TDD）
- [ ] ドメインサービス実装（TDD）
- [ ] 単体テスト作成
- [ ] 統合テスト作成

### フェーズ3: インフラフェーズ
- [ ] データベース接続実装（TDD）
- [ ] リポジトリ実装（TDD）
- [ ] キャッシュ実装（TDD）
- [ ] マイグレーション実装（TDD）
- [ ] 単体テスト作成
- [ ] 統合テスト作成

### フェーズ4: APIフェーズ
- [ ] HTTPハンドラー実装（TDD）
- [ ] ミドルウェア実装（TDD）
- [ ] ルーティング実装（TDD）
- [ ] サーバー設定実装（TDD）
- [ ] 単体テスト作成
- [ ] 統合テスト作成

### フェーズ5: ビーコンフェーズ
- [ ] ビーコン生成器実装（TDD）
- [ ] ビーコン配信実装（TDD）
- [ ] 単体テスト作成
- [ ] 統合テスト作成

### フェーズ6: 統合フェーズ
- [ ] E2Eテスト実装（TDD）
- [ ] パフォーマンステスト実装（TDD）
- [ ] セキュリティテスト実装（TDD）
- [ ] 全テスト実行
- [ ] 品質評価
- [ ] ドキュメント更新

## 12. リスク管理

### 12.1 技術的リスク
- **データベース性能**: パーティショニング・インデックス最適化
- **API性能**: キャッシュ・非同期処理の実装
- **セキュリティ**: 認証・認可・入力値検証の徹底

### 12.2 スケジュールリスク
- **依存関係**: フェーズ間の依存関係を最小化
- **並行開発**: 独立したコンポーネントの並行実装
- **早期テスト**: 各フェーズでの早期テスト実行

### 12.3 品質リスク
- **TDD徹底**: テストファーストの徹底
- **コードレビュー**: 各フェーズでのコードレビュー
- **継続的改善**: テスト結果に基づく改善

## 13. 成功指標

### 13.1 技術指標
- **テストカバレッジ**: 80%以上
- **API応答時間**: 50ms以下
- **スループット**: 5000 req/sec以上
- **エラー率**: 0.1%以下

### 13.2 プロセス指標
- **TDD遵守率**: 100%
- **テスト実行時間**: 5分以内
- **デプロイ成功率**: 95%以上
- **バグ発見率**: 早期発見90%以上

### 13.3 ビジネス指標
- **開発速度**: 計画通り
- **品質**: 要件満足
- **保守性**: 高保守性
- **拡張性**: 高拡張性

この実装順序に従うことで、TDDを徹底し、高品質なシステムを段階的に構築できます。各フェーズでテストを先行して実装し、継続的に品質を確認しながら開発を進めることが重要です。
