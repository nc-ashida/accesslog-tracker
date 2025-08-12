# Access Log Tracker

高性能なWebアクセスログトラッキングシステムです。リアルタイムでのアクセス分析、カスタムパラメータの追跡、Webhook通知機能を提供します。

## 機能

- **リアルタイムトラッキング**: JavaScriptビーコンによる軽量なアクセスログ収集
- **カスタムパラメータ**: 柔軟なカスタムデータの追跡
- **統計分析**: リアルタイムでのアクセス統計とレポート
- **Webhook通知**: イベント発生時の自動通知
- **高可用性**: マイクロサービスアーキテクチャによるスケーラブルな設計
- **監視・メトリクス**: Prometheus、Grafana、CloudWatchによる包括的な監視

## アーキテクチャ

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Client    │    │   Mobile App    │    │   API Client    │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────▼─────────────┐
                    │    Load Balancer (ALB)    │
                    └─────────────┬─────────────┘
                                  │
                    ┌─────────────▼─────────────┐
                    │   API Gateway (CloudFront)│
                    └─────────────┬─────────────┘
                                  │
                    ┌─────────────▼─────────────┐
                    │   API Server (Go/Gin)     │
                    └─────────────┬─────────────┘
                                  │
          ┌───────────────────────┼───────────────────────┐
          │                       │                       │
┌─────────▼─────────┐  ┌─────────▼─────────┐  ┌─────────▼─────────┐
│   PostgreSQL      │  │   Redis Cache     │  │   S3 Storage      │
│   (Primary DB)    │  │   (Session/Stats) │  │   (Assets)        │
└───────────────────┘  └───────────────────┘  └───────────────────┘
```

## 技術スタック

### バックエンド
- **言語**: Go 1.21+
- **Webフレームワーク**: Gin
- **データベース**: PostgreSQL
- **キャッシュ**: Redis
- **認証**: JWT
- **メトリクス**: Prometheus

### インフラストラクチャ
- **コンテナ**: Docker
- **オーケストレーション**: Kubernetes
- **クラウド**: AWS (EC2, RDS, ElastiCache, S3, CloudFront)
- **監視**: CloudWatch, Grafana

## クイックスタート

### 前提条件

- Go 1.21以上
- Docker & Docker Compose
- PostgreSQL 14以上
- Redis 6以上

### 1. リポジトリのクローン

```bash
git clone https://github.com/your-org/access-log-tracker.git
cd access-log-tracker
```

### 2. 環境設定

```bash
cp env.example .env
# .envファイルを編集して環境変数を設定
```

### 3. 依存関係のインストール

```bash
go mod download
```

### 4. 開発環境の起動

```bash
# Docker Composeで開発環境を起動
docker-compose up -d

# データベースマイグレーション
make migrate

# アプリケーションの起動
make run
```

### 5. 動作確認

```bash
# ヘルスチェック
curl http://localhost:8080/health

# APIドキュメント
open http://localhost:8080/docs
```

## 開発

### プロジェクト構造

```
access-log-tracker/
├── cmd/                    # エントリーポイント
│   ├── api/               # APIサーバー
│   ├── worker/            # バッチワーカー
│   └── beacon-generator/  # ビーコン生成ツール
├── internal/              # 内部パッケージ
│   ├── api/              # API層
│   ├── domain/           # ドメイン層
│   ├── infrastructure/   # インフラ層
│   └── utils/            # ユーティリティ
├── web/                  # フロントエンド
├── tests/                # テスト
├── deployments/          # デプロイメント設定
├── monitoring/           # 監視設定
└── docs/                # ドキュメント
```

### 開発コマンド

```bash
# アプリケーションの起動
make run

# テストの実行
make test

# ビルド
make build

# リント
make lint

# フォーマット
make fmt

# カバレッジレポート
make coverage
```

### テスト

```bash
# 単体テスト
go test ./...

# 統合テスト
go test ./tests/integration/...

# E2Eテスト
go test ./tests/e2e/...

# カバレッジ
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## デプロイメント

### Docker

```bash
# イメージのビルド
docker build -t access-log-tracker .

# コンテナの起動
docker run -p 8080:8080 access-log-tracker
```

### Kubernetes

```bash
# デプロイメント
kubectl apply -f deployments/kubernetes/

# サービスの確認
kubectl get services

# ログの確認
kubectl logs -f deployment/access-log-tracker
```

### AWS

```bash
# CloudFormationでデプロイ
aws cloudformation deploy \
  --template-file deployments/aws/cloudformation/infrastructure.yml \
  --stack-name access-log-tracker \
  --capabilities CAPABILITY_IAM

# Terraformでデプロイ
cd deployments/aws/terraform
terraform init
terraform plan
terraform apply
```

## API ドキュメント

### トラッキングエンドポイント

```http
POST /api/v1/track
Content-Type: application/json

{
  "application_id": "app_123",
  "session_id": "sess_456",
  "page_url": "https://example.com/page",
  "referrer": "https://google.com",
  "user_agent": "Mozilla/5.0...",
  "ip_address": "192.168.1.1",
  "custom_params": {
    "user_id": "user_789",
    "category": "electronics"
  }
}
```

### 統計エンドポイント

```http
GET /api/v1/statistics?application_id=app_123&period=24h
Authorization: Bearer <token>
```

## 監視・メトリクス

### Prometheus メトリクス

- `http_requests_total`: リクエスト数
- `http_request_duration_seconds`: レスポンス時間
- `tracking_events_total`: トラッキングイベント数
- `database_connections`: データベース接続数

### Grafana ダッシュボード

- API パフォーマンス
- トラッキング統計
- システムリソース

## ライセンス

MIT License

## コントリビューション

1. Fork する
2. フィーチャーブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'Add amazing feature'`)
4. ブランチにプッシュ (`git push origin feature/amazing-feature`)
5. Pull Request を作成

## サポート

- ドキュメント: [docs/](docs/)
- イシュー: [GitHub Issues](https://github.com/your-org/access-log-tracker/issues)
- ディスカッション: [GitHub Discussions](https://github.com/your-org/access-log-tracker/discussions)
