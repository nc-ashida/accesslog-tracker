# Go/ビーコン実装ディレクトリ構成

## 1. プロジェクトルート構造

```
access-log-tracker/
├── README.md                           # プロジェクト概要
├── go.mod                              # Goモジュール定義
├── go.sum                              # 依存関係チェックサム
├── .gitignore                          # Git除外設定
├── .env.example                        # 環境変数テンプレート
├── docker-compose.yml                  # 開発環境用Docker Compose
├── Dockerfile                          # 本番用Dockerfile
├── Makefile                           # ビルド・デプロイ用Makefile
├── scripts/                           # 各種スクリプト
│   ├── build.sh                       # ビルドスクリプト
│   ├── deploy.sh                      # デプロイスクリプト
│   ├── migrate.sh                     # データベースマイグレーション
│   └── health-check.sh                # ヘルスチェックスクリプト
├── configs/                           # 設定ファイル
│   ├── nginx/                         # Nginx設定
│   │   ├── nginx.conf                 # メイン設定
│   │   ├── ssl/                       # SSL証明書
│   │   └── sites-enabled/             # サイト設定
│   ├── kubernetes/                    # Kubernetes設定
│   │   ├── namespace.yaml
│   │   ├── configmap.yaml
│   │   ├── secret.yaml
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   └── ingress.yaml
├── cmd/                               # 実行可能ファイル
│   ├── api/                           # APIサーバー
│   │   └── main.go                    # メインエントリーポイント
│   ├── worker/                        # バッチワーカー
│   │   └── main.go                    # ワーカーエントリーポイント
│   └── beacon-generator/              # ビーコン生成ツール
│       └── main.go                    # ビーコン生成エントリーポイント
├── internal/                          # 内部パッケージ
│   ├── api/                           # API層
│   │   ├── handlers/                  # HTTPハンドラー
│   │   │   ├── tracking.go            # トラッキングハンドラー
│   │   │   ├── health.go              # ヘルスチェックハンドラー
│   │   │   ├── statistics.go          # 統計ハンドラー
│   │   │   ├── applications.go        # アプリケーション管理ハンドラー
│   │   │   └── webhooks.go            # Webhookハンドラー
│   │   ├── middleware/                # ミドルウェア
│   │   │   ├── auth.go                # 認証ミドルウェア
│   │   │   ├── cors.go                # CORSミドルウェア
│   │   │   ├── rate_limit.go          # レート制限ミドルウェア
│   │   │   ├── logging.go             # ログミドルウェア
│   │   │   └── timeout.go             # タイムアウトミドルウェア
│   │   ├── routes/                    # ルーティング
│   │   │   ├── v1.go                  # v1 APIルート
│   │   │   └── routes.go              # ルート設定
│   │   └── server.go                  # サーバー設定
│   ├── domain/                        # ドメイン層
│   │   ├── models/                    # ドメインモデル
│   │   │   ├── tracking.go            # トラッキングデータモデル
│   │   │   ├── application.go         # アプリケーションモデル
│   │   │   ├── session.go             # セッションモデル
│   │   │   └── custom_params.go       # カスタムパラメータモデル
│   │   ├── services/                  # ドメインサービス
│   │   │   ├── tracking_service.go    # トラッキングサービス
│   │   │   ├── application_service.go # アプリケーションサービス
│   │   │   ├── statistics_service.go  # 統計サービス
│   │   │   └── webhook_service.go     # Webhookサービス
│   │   └── validators/                # バリデーター
│   │       ├── tracking_validator.go  # トラッキングバリデーター
│   │       └── application_validator.go # アプリケーションバリデーター
│   ├── infrastructure/                # インフラ層
│   │   ├── database/                  # データベース
│   │   │   ├── postgresql/            # PostgreSQL実装
│   │   │   │   ├── connection.go      # 接続管理
│   │   │   │   ├── repositories/      # リポジトリ実装
│   │   │   │   │   ├── tracking_repository.go
│   │   │   │   │   ├── application_repository.go
│   │   │   │   │   ├── session_repository.go
│   │   │   │   │   └── statistics_repository.go
│   │   │   │   ├── migrations/        # マイグレーション
│   │   │   │   │   ├── 001_initial_schema.sql
│   │   │   │   │   ├── 002_add_custom_params.sql
│   │   │   │   │   └── 003_add_partitioning.sql
│   │   │   │   └── partition_manager.go # パーティション管理
│   │   │   └── interfaces/            # データベースインターフェース
│   │   │       ├── tracking_repository.go
│   │   │       ├── application_repository.go
│   │   │       └── session_repository.go
│   │   ├── cache/                     # キャッシュ
│   │   │   ├── redis/                 # Redis実装
│   │   │   │   ├── connection.go      # Redis接続
│   │   │   │   └── cache_service.go   # キャッシュサービス
│   │   │   └── interfaces/            # キャッシュインターフェース
│   │   │       └── cache_service.go
│   │   ├── queue/                     # メッセージキュー
│   │   │   ├── sqs/                   # SQS実装
│   │   │   │   ├── connection.go      # SQS接続
│   │   │   │   └── queue_service.go   # キューサービス
│   │   │   └── interfaces/            # キューインターフェース
│   │   │       └── queue_service.go
│   │   └── storage/                   # ストレージ
│   │       ├── s3/                    # S3実装
│   │       │   ├── connection.go      # S3接続
│   │       │   └── storage_service.go # ストレージサービス
│   │       └── interfaces/            # ストレージインターフェース
│   │           └── storage_service.go
│   ├── beacon/                        # ビーコン関連
│   │   ├── generator/                 # ビーコン生成
│   │   │   ├── beacon_generator.go    # ビーコン生成器
│   │   │   ├── template.go            # テンプレート管理
│   │   │   └── minifier.go            # コード圧縮
│   │   ├── templates/                 # ビーコンテンプレート
│   │   │   ├── tracker.js             # 基本ビーコン
│   │   │   ├── tracker.min.js         # 圧縮版ビーコン
│   │   │   └── tracker.debug.js       # デバッグ版ビーコン
│   │   └── config/                    # ビーコン設定
│   │       ├── beacon_config.go       # ビーコン設定
│   │       └── cloudfront_config.go   # CloudFront設定
│   └── utils/                         # ユーティリティ
│       ├── logger/                    # ログ
│       │   ├── logger.go              # ロガー設定
│       │   └── formatter.go           # ログフォーマッター
│       ├── crypto/                    # 暗号化
│       │   ├── hash.go                # ハッシュ関数
│       │   └── encryption.go          # 暗号化関数
│       ├── timeutil/                  # 時間ユーティリティ
│       │   └── timeutil.go            # 時間処理
│       ├── iputil/                    # IPユーティリティ
│       │   └── iputil.go              # IP処理
│       └── jsonutil/                  # JSONユーティリティ
│           └── jsonutil.go            # JSON処理
├── pkg/                               # 公開パッケージ
│   ├── client/                        # クライアントライブラリ
│   │   ├── tracking_client.go         # トラッキングクライアント
│   │   └── statistics_client.go       # 統計クライアント
│   └── beacon/                        # ビーコンライブラリ
│       └── beacon_client.go           # ビーコンクライアント
├── web/                               # Webアセット
│   ├── static/                        # 静的ファイル
│   │   ├── js/                        # JavaScript
│   │   │   ├── tracker.js             # トラッキングビーコン
│   │   │   └── admin.js               # 管理画面用JS
│   │   ├── css/                       # CSS
│   │   │   └── admin.css              # 管理画面用CSS
│   │   └── images/                    # 画像
│   └── templates/                     # HTMLテンプレート
│       ├── admin/                     # 管理画面テンプレート
│       │   ├── dashboard.html
│       │   ├── applications.html
│       │   └── statistics.html
│       └── beacon/                    # ビーコンテンプレート
│           └── embed.html             # 埋め込み用HTML
├── tests/                             # テスト
│   ├── unit/                          # 単体テスト
│   │   ├── api/                       # APIテスト
│   │   ├── domain/                    # ドメインテスト
│   │   ├── infrastructure/            # インフラテスト
│   │   └── utils/                     # ユーティリティテスト
│   ├── integration/                   # 統合テスト
│   │   ├── api/                       # API統合テスト
│   │   ├── database/                  # データベース統合テスト
│   │   └── beacon/                    # ビーコン統合テスト
│   ├── e2e/                           # E2Eテスト
│   │   ├── tracking/                  # トラッキングE2Eテスト
│   │   └── beacon/                    # ビーコンE2Eテスト
│   ├── performance/                   # パフォーマンステスト
│   │   ├── load_test.yml              # 負荷テスト設定
│   │   └── stress_test.yml            # ストレステスト設定
│   ├── security/                      # セキュリティテスト
│   │   ├── authentication_test.go     # 認証テスト
│   │   └── authorization_test.go      # 認可テスト
│   ├── fixtures/                      # テストデータ
│   │   ├── tracking_data.json         # トラッキングテストデータ
│   │   ├── applications.json          # アプリケーションテストデータ
│   │   └── custom_params.json         # カスタムパラメータテストデータ
│   ├── mocks/                         # モック
│   │   ├── database_mock.go           # データベースモック
│   │   ├── cache_mock.go              # キャッシュモック
│   │   └── queue_mock.go              # キューモック
│   └── helpers/                       # テストヘルパー
│       ├── database_helper.go         # データベースヘルパー
│       ├── test_server.go             # テストサーバー
│       └── test_data_generator.go     # テストデータ生成器
├── deployments/                       # デプロイメント設定
│   ├── aws/                           # AWS設定
│   │   ├── cloudformation/            # CloudFormation
│   │   │   ├── infrastructure.yml     # インフラ設定
│   │   │   ├── alb.yml                # ALB設定
│   │   │   ├── ec2.yml                # EC2設定
│   │   │   ├── rds.yml                # RDS設定
│   │   │   └── cloudfront.yml         # CloudFront設定
│   │   ├── terraform/                 # Terraform設定
│   │   │   ├── main.tf                # メイン設定
│   │   │   ├── variables.tf           # 変数定義
│   │   │   ├── outputs.tf             # 出力定義
│   │   │   └── modules/               # Terraformモジュール
│   │   │       ├── vpc/               # VPCモジュール
│   │   │       ├── alb/               # ALBモジュール
│   │   │       ├── ec2/               # EC2モジュール
│   │   │       └── rds/               # RDSモジュール
│   │   └── lambda/                    # Lambda関数
│   │       ├── edge-functions/        # Lambda@Edge
│   │       │   ├── viewer-request.js  # ビューアリクエスト
│   │       │   └── origin-response.js # オリジンレスポンス
│   │       └── workers/               # ワーカー関数
│   │           ├── batch-processor.js  # バッチ処理
│   │           └── statistics-generator.js # 統計生成
│   ├── kubernetes/                    # Kubernetes設定
│   │   ├── namespace.yaml
│   │   ├── configmap.yaml
│   │   ├── secret.yaml
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   ├── ingress.yaml
│   │   └── hpa.yaml                   # Horizontal Pod Autoscaler
│   └── docker/                        # Docker設定
│       ├── Dockerfile.dev             # 開発用Dockerfile
│       ├── Dockerfile.prod            # 本番用Dockerfile
│       └── docker-compose.yml         # Docker Compose
├── monitoring/                        # 監視設定
│   ├── prometheus/                    # Prometheus設定
│   │   ├── prometheus.yml             # Prometheus設定
│   │   └── rules/                     # アラートルール
│   │       ├── api_alerts.yml         # APIアラート
│   │       └── system_alerts.yml      # システムアラート
│   ├── grafana/                       # Grafana設定
│   │   ├── dashboards/                # ダッシュボード
│   │   │   ├── api_dashboard.json     # APIダッシュボード
│   │   │   ├── beacon_dashboard.json  # ビーコンダッシュボード
│   │   │   └── system_dashboard.json  # システムダッシュボード
│   │   └── datasources/               # データソース
│   │       └── prometheus.yml         # Prometheusデータソース
│   └── cloudwatch/                    # CloudWatch設定
│       ├── alarms/                    # アラーム
│       │   ├── api_alarms.yml         # APIアラーム
│       │   └── system_alarms.yml      # システムアラーム
│       └── dashboards/                # CloudWatchダッシュボード
│           └── main_dashboard.json    # メインダッシュボード
└── tools/                             # 開発ツール
    ├── codegen/                       # コード生成
    │   ├── generate_models.go         # モデル生成
    │   └── generate_handlers.go       # ハンドラー生成
    ├── migration/                     # マイグレーションツール
    │   ├── create_migration.go        # マイグレーション作成
    │   └── run_migrations.go          # マイグレーション実行
    └── beacon-builder/                # ビーコンビルダー
        ├── build_beacon.go            # ビーコンビルド
        └── deploy_beacon.go           # ビーコンデプロイ
```

## 2. 主要ファイルの詳細

### 2.1 メインエントリーポイント

#### cmd/api/main.go
```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "access-log-tracker/internal/api"
    "access-log-tracker/internal/infrastructure/database/postgresql"
    "access-log-tracker/internal/infrastructure/cache/redis"
    "access-log-tracker/internal/utils/logger"
)

func main() {
    // ロガー初期化
    logger := logger.New()
    
    // データベース接続
    db, err := postgresql.NewConnection()
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()
    
    // Redis接続
    redisClient, err := redis.NewConnection()
    if err != nil {
        log.Fatal("Failed to connect to Redis:", err)
    }
    defer redisClient.Close()
    
    // APIサーバー初期化
    server := api.NewServer(db, redisClient, logger)
    
    // グレースフルシャットダウン
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    go func() {
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
        <-sigChan
        cancel()
    }()
    
    // サーバー起動
    if err := server.Start(ctx); err != nil {
        log.Fatal("Server failed:", err)
    }
}
```

### 2.2 ドメインモデル

#### internal/domain/models/tracking.go
```go
package models

import (
    "time"
    "encoding/json"
)

// TrackingData トラッキングデータモデル
type TrackingData struct {
    ID            int64                  `json:"id" db:"id"`
    AppID         string                 `json:"app_id" db:"app_id" binding:"required"`
    ClientSubID   string                 `json:"client_sub_id,omitempty" db:"client_sub_id"`
    ModuleID      string                 `json:"module_id,omitempty" db:"module_id"`
    URL           string                 `json:"url,omitempty" db:"url"`
    Referrer      string                 `json:"referrer,omitempty" db:"referrer"`
    UserAgent     string                 `json:"user_agent" db:"user_agent" binding:"required"`
    IPAddress     string                 `json:"ip_address,omitempty" db:"ip_address"`
    SessionID     string                 `json:"session_id,omitempty" db:"session_id"`
    ScreenRes     string                 `json:"screen_resolution,omitempty" db:"screen_resolution"`
    Language      string                 `json:"language,omitempty" db:"language"`
    Timezone      string                 `json:"timezone,omitempty" db:"timezone"`
    CustomParams  json.RawMessage        `json:"custom_params,omitempty" db:"custom_params"`
    CreatedAt     time.Time              `json:"created_at" db:"created_at"`
}

// TrackingRequest トラッキングリクエスト
type TrackingRequest struct {
    AppID        string                 `json:"app_id" binding:"required"`
    ClientSubID  string                 `json:"client_sub_id,omitempty"`
    ModuleID     string                 `json:"module_id,omitempty"`
    URL          string                 `json:"url,omitempty"`
    Referrer     string                 `json:"referrer,omitempty"`
    UserAgent    string                 `json:"user_agent" binding:"required"`
    IPAddress    string                 `json:"ip_address,omitempty"`
    SessionID    string                 `json:"session_id,omitempty"`
    ScreenRes    string                 `json:"screen_resolution,omitempty"`
    Language     string                 `json:"language,omitempty"`
    Timezone     string                 `json:"timezone,omitempty"`
    CustomParams map[string]interface{} `json:"custom_params,omitempty"`
}

// TrackingResponse トラッキングレスポンス
type TrackingResponse struct {
    Success   bool        `json:"success"`
    Data      interface{} `json:"data,omitempty"`
    Message   string      `json:"message"`
    Timestamp string      `json:"timestamp"`
}
```

### 2.3 APIハンドラー

#### internal/api/handlers/tracking.go
```go
package handlers

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "access-log-tracker/internal/domain/models"
    "access-log-tracker/internal/domain/services"
    "access-log-tracker/internal/domain/validators"
)

type TrackingHandler struct {
    trackingService *services.TrackingService
    validator       *validators.TrackingValidator
}

func NewTrackingHandler(trackingService *services.TrackingService, validator *validators.TrackingValidator) *TrackingHandler {
    return &TrackingHandler{
        trackingService: trackingService,
        validator:       validator,
    }
}

// Track トラッキングデータ受信ハンドラー
func (h *TrackingHandler) Track(c *gin.Context) {
    var req models.TrackingRequest
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.TrackingResponse{
            Success:   false,
            Message:   "Invalid request data",
            Timestamp: time.Now().UTC().Format(time.RFC3339),
        })
        return
    }
    
    // バリデーション
    if err := h.validator.Validate(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.TrackingResponse{
            Success:   false,
            Message:   err.Error(),
            Timestamp: time.Now().UTC().Format(time.RFC3339),
        })
        return
    }
    
    // トラッキングデータ保存
    trackingData := &models.TrackingData{
        AppID:        req.AppID,
        ClientSubID:  req.ClientSubID,
        ModuleID:     req.ModuleID,
        URL:          req.URL,
        Referrer:     req.Referrer,
        UserAgent:    req.UserAgent,
        IPAddress:    req.IPAddress,
        SessionID:    req.SessionID,
        ScreenRes:    req.ScreenRes,
        Language:     req.Language,
        Timezone:     req.Timezone,
        CustomParams: req.CustomParams,
        CreatedAt:    time.Now(),
    }
    
    if err := h.trackingService.SaveTrackingData(trackingData); err != nil {
        c.JSON(http.StatusInternalServerError, models.TrackingResponse{
            Success:   false,
            Message:   "Failed to save tracking data",
            Timestamp: time.Now().UTC().Format(time.RFC3339),
        })
        return
    }
    
    c.JSON(http.StatusOK, models.TrackingResponse{
        Success:   true,
        Message:   "Tracking data recorded successfully",
        Timestamp: time.Now().UTC().Format(time.RFC3339),
    })
}
```

### 2.4 ビーコン生成

#### internal/beacon/generator/beacon_generator.go
```go
package generator

import (
    "bytes"
    "embed"
    "text/template"
    "access-log-tracker/internal/beacon/config"
)

//go:embed templates/*
var templates embed.FS

// BeaconGenerator ビーコン生成器
type BeaconGenerator struct {
    config *config.BeaconConfig
}

// NewBeaconGenerator 新しいビーコン生成器を作成
func NewBeaconGenerator(config *config.BeaconConfig) *BeaconGenerator {
    return &BeaconGenerator{
        config: config,
    }
}

// GenerateTracker トラッキングビーコンを生成
func (g *BeaconGenerator) GenerateTracker() ([]byte, error) {
    tmpl, err := template.ParseFS(templates, "templates/tracker.js")
    if err != nil {
        return nil, err
    }
    
    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, g.config); err != nil {
        return nil, err
    }
    
    return buf.Bytes(), nil
}

// GenerateMinifiedTracker 圧縮版トラッキングビーコンを生成
func (g *BeaconGenerator) GenerateMinifiedTracker() ([]byte, error) {
    data, err := g.GenerateTracker()
    if err != nil {
        return nil, err
    }
    
    return g.minify(data), nil
}

// minify JavaScriptコードを圧縮
func (g *BeaconGenerator) minify(data []byte) []byte {
    // 簡易的な圧縮処理
    return data
}
```

### 2.5 データベースリポジトリ

#### internal/infrastructure/database/postgresql/repositories/tracking_repository.go
```go
package repositories

import (
    "database/sql"
    "encoding/json"
    "time"
    
    "access-log-tracker/internal/domain/models"
)

type TrackingRepository struct {
    db *sql.DB
}

func NewTrackingRepository(db *sql.DB) *TrackingRepository {
    return &TrackingRepository{db: db}
}

// SaveTrackingData トラッキングデータを保存
func (r *TrackingRepository) SaveTrackingData(data *models.TrackingData) error {
    query := `
        INSERT INTO access_logs (
            app_id, client_sub_id, module_id, url, referrer,
            user_agent, ip_address, session_id, screen_resolution,
            language, timezone, custom_params, created_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
    `
    
    customParamsJSON, _ := json.Marshal(data.CustomParams)
    
    _, err := r.db.Exec(query,
        data.AppID, data.ClientSubID, data.ModuleID, data.URL, data.Referrer,
        data.UserAgent, data.IPAddress, data.SessionID, data.ScreenRes,
        data.Language, data.Timezone, customParamsJSON, data.CreatedAt,
    )
    
    return err
}

// GetStatistics 統計情報を取得
func (r *TrackingRepository) GetStatistics(appID string, startDate, endDate time.Time) (*models.Statistics, error) {
    query := `
        SELECT 
            COUNT(*) as total_requests,
            COUNT(DISTINCT session_id) as unique_sessions,
            COUNT(DISTINCT ip_address) as unique_visitors
        FROM access_logs 
        WHERE app_id = $1 
        AND created_at BETWEEN $2 AND $3
    `
    
    var stats models.Statistics
    err := r.db.QueryRow(query, appID, startDate, endDate).Scan(
        &stats.TotalRequests,
        &stats.UniqueSessions,
        &stats.UniqueVisitors,
    )
    
    if err != nil {
        return nil, err
    }
    
    return &stats, nil
}
```

## 3. 設定ファイル

### 3.1 go.mod
```go
module access-log-tracker

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/lib/pq v1.10.9
    github.com/go-redis/redis/v8 v8.11.5
    github.com/aws/aws-sdk-go v1.48.0
    github.com/sirupsen/logrus v1.9.3
    github.com/gin-contrib/cors v1.4.0
    github.com/gin-contrib/timeout v0.0.3
    github.com/stretchr/testify v1.8.4
    github.com/golang-migrate/migrate/v4 v4.16.2
    github.com/prometheus/client_golang v1.17.0
    github.com/gin-contrib/prometheus v0.0.0-20230501144526-8c036d44e6b7
)
```

### 3.2 Dockerfile
```dockerfile
# ビルドステージ
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 依存関係コピー
COPY go.mod go.sum ./
RUN go mod download

# ソースコードコピー
COPY . .

# ビルド
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# 実行ステージ
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# バイナリコピー
COPY --from=builder /app/main .

# 設定ファイルコピー
COPY --from=builder /app/configs/ ./configs/

EXPOSE 8080

CMD ["./main"]
```

### 3.3 docker-compose.yml
```yaml
version: '3.8'

services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=access_log_tracker
      - DB_USER=postgres
      - DB_PASSWORD=password
      - REDIS_HOST=redis
      - REDIS_PORT=6379
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:14
    environment:
      - POSTGRES_DB=access_log_tracker
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./internal/infrastructure/database/postgresql/migrations:/docker-entrypoint-initdb.d

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  postgres_data:
```

## 4. テスト構造

### 4.1 単体テスト例

#### tests/unit/api/handlers/tracking_test.go
```go
package handlers_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"

    "access-log-tracker/internal/api/handlers"
    "access-log-tracker/internal/domain/models"
    "access-log-tracker/internal/domain/services"
    "access-log-tracker/internal/domain/validators"
)

func TestTrackingHandler_Track(t *testing.T) {
    // モックサービス
    mockTrackingService := &services.MockTrackingService{}
    mockValidator := &validators.MockTrackingValidator{}
    
    handler := handlers.NewTrackingHandler(mockTrackingService, mockValidator)
    
    // テストケース
    tests := []struct {
        name           string
        requestBody    models.TrackingRequest
        expectedStatus int
        setupMocks     func()
    }{
        {
            name: "valid tracking data",
            requestBody: models.TrackingRequest{
                AppID:     "test_app",
                UserAgent: "Mozilla/5.0",
                URL:       "https://example.com",
            },
            expectedStatus: http.StatusOK,
            setupMocks: func() {
                mockValidator.On("Validate", mock.Anything).Return(nil)
                mockTrackingService.On("SaveTrackingData", mock.Anything).Return(nil)
            },
        },
        {
            name: "invalid app_id",
            requestBody: models.TrackingRequest{
                UserAgent: "Mozilla/5.0",
                URL:       "https://example.com",
            },
            expectedStatus: http.StatusBadRequest,
            setupMocks: func() {
                mockValidator.On("Validate", mock.Anything).Return(errors.New("app_id is required"))
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // モック設定
            tt.setupMocks()
            
            // リクエスト作成
            body, _ := json.Marshal(tt.requestBody)
            req := httptest.NewRequest("POST", "/v1/track", bytes.NewBuffer(body))
            req.Header.Set("Content-Type", "application/json")
            
            // レスポンス作成
            w := httptest.NewRecorder()
            
            // Ginコンテキスト作成
            gin.SetMode(gin.TestMode)
            c, _ := gin.CreateTestContext(w)
            c.Request = req
            
            // ハンドラー実行
            handler.Track(c)
            
            // アサーション
            assert.Equal(t, tt.expectedStatus, w.Code)
        })
    }
}
```

## 5. デプロイメント設定

### 5.1 Kubernetes Deployment

#### deployments/kubernetes/deployment.yaml
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: alt-api
  namespace: access-log-tracker
spec:
  replicas: 3
  selector:
    matchLabels:
      app: alt-api
  template:
    metadata:
      labels:
        app: alt-api
    spec:
      containers:
      - name: api
        image: access-log-tracker:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          valueFrom:
            configMapKeyRef:
              name: alt-config
              key: DB_HOST
        - name: DB_PORT
          valueFrom:
            configMapKeyRef:
              name: alt-config
              key: DB_PORT
        - name: DB_NAME
          valueFrom:
            configMapKeyRef:
              name: alt-config
              key: DB_NAME
        - name: DB_USER
          valueFrom:
            configMapKeyRef:
              name: alt-config
              key: DB_USER
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: alt-secrets
              key: DB_PASSWORD
        - name: REDIS_HOST
          valueFrom:
            configMapKeyRef:
              name: alt-config
              key: REDIS_HOST
        - name: REDIS_PORT
          valueFrom:
            configMapKeyRef:
              name: alt-config
              key: REDIS_PORT
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /v1/health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /v1/health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

## 6. 監視設定

### 6.1 Prometheus設定

#### monitoring/prometheus/prometheus.yml
```yaml
global:
  scrape_interval: 15s

scrape_configs:
- job_name: 'alt-api'
  static_configs:
  - targets: ['alt-api-service:8080']
  metrics_path: /metrics
  scrape_interval: 10s

- job_name: 'postgres'
  static_configs:
  - targets: ['postgres:5432']

- job_name: 'redis'
  static_configs:
  - targets: ['redis:6379']
```

### 6.2 Grafanaダッシュボード

#### monitoring/grafana/dashboards/api_dashboard.json
```json
{
  "dashboard": {
    "title": "Access Log Tracker API Dashboard",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "{{app_id}}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total{status=~\"5..\"}[5m])",
            "legendFormat": "5xx errors"
          }
        ]
      }
    ]
  }
}
```

このディレクトリ構成により、仕様書に基づいた包括的で実用的なGo/ビーコン実装が可能になります。各層が適切に分離され、テスト可能で、スケーラブルなアーキテクチャを実現できます。
