# Go/ビーコン実装順序ガイド

## 概要

このドキュメントは、Access Log TrackerのGo/ビーコン実装における推奨実装順序を定義します。段階的なアプローチにより、早期にMVP（Minimum Viable Product）を提供し、その後機能を拡張していく構成となっています。

## 実装フェーズ概要

### フェーズ1: 基盤構築（1-2週間）
- プロジェクト初期化
- ユーティリティ層実装
- 開発環境構築

### フェーズ2: ドメイン層実装（1週間）
- ドメインモデル定義
- バリデーター実装

### フェーズ3: インフラ層実装（1-2週間）
- データベース接続・リポジトリ
- キャッシュ機能

### フェーズ4: ドメインサービス実装（1週間）
- ビジネスロジック実装

### フェーズ5: API層実装（1-2週間）
- ミドルウェア
- ハンドラー
- ルーティング

### フェーズ6: エントリーポイント実装（1週間）
- メインアプリケーション

### フェーズ7: ビーコン機能実装（1-2週間）
- ビーコン生成
- Webアセット

### フェーズ8: テスト実装（1-2週間）
- 単体・統合・E2Eテスト

### フェーズ9: デプロイメント設定（1週間）
- AWS・Kubernetes・Docker設定

### フェーズ10: 監視設定（1週間）
- Prometheus・Grafana・CloudWatch

## 詳細実装順序

### フェーズ1: 基盤構築（1-2週間）

#### 1.1 プロジェクト初期化
**優先度: 最高**
**推定時間: 1日**

```
access-log-tracker/
├── go.mod                              # Goモジュール定義
├── go.sum                              # 依存関係チェックサム
├── .gitignore                          # Git除外設定
├── .env.example                        # 環境変数テンプレート
└── README.md                           # プロジェクト概要
```

**実装内容:**
- `go.mod`でモジュール定義と依存関係を設定
- 主要なGoライブラリ（Gin、PostgreSQL、Redis等）を追加
- `.gitignore`で適切な除外設定
- `.env.example`で環境変数テンプレートを作成

#### 1.2 ユーティリティ層実装
**優先度: 最高**
**推定時間: 3日**

```
internal/utils/
├── logger/
│   ├── logger.go                        # ロガー設定
│   └── formatter.go                     # ログフォーマッター
├── timeutil/
│   └── timeutil.go                      # 時間処理ユーティリティ
├── iputil/
│   └── iputil.go                        # IP処理ユーティリティ
├── crypto/
│   ├── hash.go                          # ハッシュ関数
│   └── encryption.go                    # 暗号化関数
└── jsonutil/
    └── jsonutil.go                      # JSON処理ユーティリティ
```

**実装内容:**
- 構造化ログ機能（logrus使用）
- 時間処理ヘルパー（タイムゾーン、フォーマット等）
- IPアドレス処理（検証、正規化等）
- 暗号化・ハッシュ機能
- JSON処理ヘルパー

#### 1.3 開発環境構築
**優先度: 最高**
**推定時間: 2日**

```
├── docker-compose.yml                   # 開発環境用Docker Compose
├── Dockerfile                           # 本番用Dockerfile
├── Makefile                            # ビルド・デプロイ用Makefile
└── scripts/                            # 各種スクリプト
    ├── build.sh                        # ビルドスクリプト
    ├── deploy.sh                       # デプロイスクリプト
    ├── migrate.sh                      # データベースマイグレーション
    └── health-check.sh                 # ヘルスチェックスクリプト
```

**実装内容:**
- Docker Composeで開発環境（PostgreSQL、Redis）
- マルチステージDockerfile
- Makefileでビルド・デプロイ自動化
- 各種運用スクリプト

### フェーズ2: ドメイン層実装（1週間）

#### 2.1 ドメインモデル定義
**優先度: 最高**
**推定時間: 3日**

```
internal/domain/models/
├── tracking.go                          # トラッキングデータモデル
├── application.go                       # アプリケーションモデル
├── session.go                           # セッションモデル
└── custom_params.go                     # カスタムパラメータモデル
```

**実装内容:**
- `TrackingData`: トラッキングデータの構造体
- `TrackingRequest`: APIリクエスト用構造体
- `TrackingResponse`: APIレスポンス用構造体
- `Application`: アプリケーション管理用構造体
- `Session`: セッション管理用構造体
- `CustomParams`: カスタムパラメータ用構造体

#### 2.2 バリデーター実装
**優先度: 高**
**推定時間: 2日**

```
internal/domain/validators/
├── tracking_validator.go                # トラッキングバリデーター
└── application_validator.go             # アプリケーションバリデーター
```

**実装内容:**
- トラッキングデータのバリデーション
- アプリケーションIDの検証
- 必須フィールドのチェック
- データ型・形式の検証

### フェーズ3: インフラ層実装（1-2週間）

#### 3.1 データベース接続・リポジトリ
**優先度: 最高**
**推定時間: 5日**

```
internal/infrastructure/database/
├── postgresql/
│   ├── connection.go                    # 接続管理
│   ├── repositories/                    # リポジトリ実装
│   │   ├── tracking_repository.go      # トラッキングリポジトリ
│   │   ├── application_repository.go   # アプリケーションリポジトリ
│   │   └── session_repository.go       # セッションリポジトリ
│   └── migrations/                     # マイグレーション
│       ├── 001_initial_schema.sql      # 初期スキーマ
│       └── 002_add_custom_params.sql   # カスタムパラメータ追加
└── interfaces/                         # データベースインターフェース
    ├── tracking_repository.go          # トラッキングリポジトリインターフェース
    ├── application_repository.go       # アプリケーションリポジトリインターフェース
    └── session_repository.go           # セッションリポジトリインターフェース
```

**実装内容:**
- PostgreSQL接続プール管理
- リポジトリパターン実装
- マイグレーション管理
- インターフェース定義（テスト用）

#### 3.2 キャッシュ機能
**優先度: 高**
**推定時間: 3日**

```
internal/infrastructure/cache/
├── redis/
│   ├── connection.go                    # Redis接続
│   └── cache_service.go                # キャッシュサービス
└── interfaces/
    └── cache_service.go                # キャッシュインターフェース
```

**実装内容:**
- Redis接続管理
- セッションキャッシュ
- 統計データキャッシュ
- キャッシュインターフェース

### フェーズ4: ドメインサービス実装（1週間）

#### 4.1 ビジネスロジック実装
**優先度: 最高**
**推定時間: 5日**

```
internal/domain/services/
├── tracking_service.go                  # トラッキングサービス
├── application_service.go               # アプリケーションサービス
├── statistics_service.go                # 統計サービス
└── webhook_service.go                  # Webhookサービス
```

**実装内容:**
- トラッキングデータの保存ロジック
- アプリケーション管理ロジック
- 統計データ生成ロジック
- Webhook送信ロジック

### フェーズ5: API層実装（1-2週間）

#### 5.1 ミドルウェア実装
**優先度: 高**
**推定時間: 3日**

```
internal/api/middleware/
├── auth.go                             # 認証ミドルウェア
├── cors.go                             # CORSミドルウェア
├── rate_limit.go                       # レート制限ミドルウェア
├── logging.go                          # ログミドルウェア
└── timeout.go                          # タイムアウトミドルウェア
```

**実装内容:**
- JWT認証
- CORS設定
- レート制限（Redis使用）
- 構造化ログ
- リクエストタイムアウト

#### 5.2 ハンドラー実装
**優先度: 最高**
**推定時間: 5日**

```
internal/api/handlers/
├── tracking.go                          # トラッキングハンドラー
├── health.go                            # ヘルスチェックハンドラー
├── statistics.go                        # 統計ハンドラー
├── applications.go                      # アプリケーション管理ハンドラー
└── webhooks.go                         # Webhookハンドラー
```

**実装内容:**
- トラッキングデータ受信エンドポイント
- ヘルスチェックエンドポイント
- 統計データ取得エンドポイント
- アプリケーション管理エンドポイント
- Webhook設定エンドポイント

#### 5.3 ルーティング・サーバー設定
**優先度: 高**
**推定時間: 2日**

```
internal/api/
├── routes/
│   ├── v1.go                           # v1 APIルート
│   └── routes.go                        # ルート設定
└── server.go                           # サーバー設定
```

**実装内容:**
- RESTful APIルート定義
- バージョニング対応
- サーバー設定・起動処理

### フェーズ6: エントリーポイント実装（1週間）

#### 6.1 メインアプリケーション
**優先度: 最高**
**推定時間: 3日**

```
cmd/
├── api/
│   └── main.go                         # APIサーバーメイン
├── worker/
│   └── main.go                         # バッチワーカーメイン
└── beacon-generator/
    └── main.go                         # ビーコン生成メイン
```

**実装内容:**
- APIサーバー起動処理
- グレースフルシャットダウン
- 環境変数読み込み
- 依存関係注入

### フェーズ7: ビーコン機能実装（1-2週間）

#### 7.1 ビーコン生成
**優先度: 中**
**推定時間: 5日**

```
internal/beacon/
├── generator/
│   ├── beacon_generator.go              # ビーコン生成器
│   ├── template.go                      # テンプレート管理
│   └── minifier.go                     # コード圧縮
├── templates/
│   ├── tracker.js                       # 基本ビーコン
│   ├── tracker.min.js                   # 圧縮版ビーコン
│   └── tracker.debug.js                 # デバッグ版ビーコン
└── config/
    ├── beacon_config.go                 # ビーコン設定
    └── cloudfront_config.go             # CloudFront設定
```

**実装内容:**
- JavaScriptビーコン生成
- テンプレートエンジン
- コード圧縮機能
- 設定管理

#### 7.2 Webアセット
**優先度: 中**
**推定時間: 3日**

```
web/
├── static/
│   ├── js/
│   │   ├── tracker.js                   # トラッキングビーコン
│   │   └── admin.js                     # 管理画面用JS
│   ├── css/
│   │   └── admin.css                    # 管理画面用CSS
│   └── images/                          # 画像
└── templates/
    ├── admin/                           # 管理画面テンプレート
    │   ├── dashboard.html
    │   ├── applications.html
    │   └── statistics.html
    └── beacon/
        └── embed.html                   # 埋め込み用HTML
```

**実装内容:**
- フロントエンドJavaScript
- 管理画面CSS
- HTMLテンプレート
- 埋め込み用HTML

### フェーズ8: テスト実装（1-2週間）

#### 8.1 単体テスト
**優先度: 高**
**推定時間: 5日**

```
tests/unit/
├── api/                                # APIテスト
├── domain/                             # ドメインテスト
├── infrastructure/                      # インフラテスト
└── utils/                              # ユーティリティテスト
```

**実装内容:**
- ハンドラーの単体テスト
- サービスの単体テスト
- リポジトリの単体テスト
- ユーティリティの単体テスト

#### 8.2 統合テスト・E2Eテスト
**優先度: 中**
**推定時間: 5日**

```
tests/
├── integration/                         # 統合テスト
│   ├── api/                            # API統合テスト
│   ├── database/                        # データベース統合テスト
│   └── beacon/                         # ビーコン統合テスト
├── e2e/                                # E2Eテスト
│   ├── tracking/                        # トラッキングE2Eテスト
│   └── beacon/                         # ビーコンE2Eテスト
├── fixtures/                            # テストデータ
├── mocks/                              # モック
└── helpers/                            # テストヘルパー
```

**実装内容:**
- API統合テスト
- データベース統合テスト
- ビーコンE2Eテスト
- テストデータ・モック

### フェーズ9: デプロイメント設定（1週間）

#### 9.1 AWS設定
**優先度: 中**
**推定時間: 3日**

```
deployments/aws/
├── cloudformation/                      # CloudFormation
│   ├── infrastructure.yml               # インフラ設定
│   ├── alb.yml                         # ALB設定
│   ├── ec2.yml                         # EC2設定
│   ├── rds.yml                         # RDS設定
│   └── cloudfront.yml                  # CloudFront設定
├── terraform/                          # Terraform設定
│   ├── main.tf                         # メイン設定
│   ├── variables.tf                    # 変数定義
│   ├── outputs.tf                      # 出力定義
│   └── modules/                        # Terraformモジュール
└── lambda/                             # Lambda関数
    ├── edge-functions/                  # Lambda@Edge
    └── workers/                         # ワーカー関数
```

**実装内容:**
- CloudFormation/Terraform設定
- ALB・EC2・RDS設定
- CloudFront設定
- Lambda関数（バッチ処理）

#### 9.2 Kubernetes・Docker設定
**優先度: 中**
**推定時間: 2日**

```
deployments/
├── kubernetes/                          # Kubernetes設定
│   ├── namespace.yaml
│   ├── configmap.yaml
│   ├── secret.yaml
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── ingress.yaml
│   └── hpa.yaml                        # Horizontal Pod Autoscaler
└── docker/                             # Docker設定
    ├── Dockerfile.dev                   # 開発用Dockerfile
    ├── Dockerfile.prod                  # 本番用Dockerfile
    └── docker-compose.yml               # Docker Compose
```

**実装内容:**
- Kubernetesマニフェスト
- Docker設定
- オートスケーリング設定

### フェーズ10: 監視設定（1週間）

#### 10.1 監視・ログ設定
**優先度: 低**
**推定時間: 5日**

```
monitoring/
├── prometheus/                          # Prometheus設定
│   ├── prometheus.yml                   # Prometheus設定
│   └── rules/                          # アラートルール
│       ├── api_alerts.yml              # APIアラート
│       └── system_alerts.yml           # システムアラート
├── grafana/                             # Grafana設定
│   ├── dashboards/                      # ダッシュボード
│   │   ├── api_dashboard.json          # APIダッシュボード
│   │   ├── beacon_dashboard.json       # ビーコンダッシュボード
│   │   └── system_dashboard.json       # システムダッシュボード
│   └── datasources/                     # データソース
│       └── prometheus.yml              # Prometheusデータソース
└── cloudwatch/                          # CloudWatch設定
    ├── alarms/                          # アラーム
    │   ├── api_alarms.yml              # APIアラーム
    │   └── system_alarms.yml           # システムアラーム
    └── dashboards/                      # CloudWatchダッシュボード
        └── main_dashboard.json         # メインダッシュボード
```

**実装内容:**
- Prometheus設定
- Grafanaダッシュボード
- CloudWatchアラーム
- メトリクス収集

## 実装の優先順位

### 高優先度（MVP - 4-6週間）
1. **フェーズ1-3**: 基盤構築とデータベース
2. **フェーズ4**: ドメインサービス
3. **フェーズ5**: API層（トラッキングハンドラーのみ）
4. **フェーズ6**: APIサーバーメイン

### 中優先度（機能拡張 - 3-4週間）
1. **フェーズ7**: ビーコン機能
2. **フェーズ8**: テスト実装
3. **フェーズ9**: デプロイメント設定

### 低優先度（運用・監視 - 2週間）
1. **フェーズ10**: 監視設定
2. **フェーズ5**: その他のハンドラー
3. **フェーズ6**: ワーカーとビーコン生成ツール

## 推奨実装アプローチ

### 1. 段階的実装
- 各フェーズを完了してから次に進む
- 各フェーズで動作確認を行う
- 早期にMVPを提供

### 2. テスト駆動開発
- 各機能の実装前にテストを作成
- 継続的インテグレーション
- カバレッジ目標: 80%以上

### 3. AWS Lambda優先
- ユーザーの好みに合わせてLambda関数を優先実装
- サーバーレスアーキテクチャの活用
- コスト最適化

### 4. コスト最適化
- 運用コストを最小化する設計
- リソース使用量の監視
- 自動スケーリングの活用

## リスク管理

### 技術的リスク
- **データベース設計**: 早期にスキーマを確定
- **パフォーマンス**: 負荷テストを早期に実施
- **セキュリティ**: 認証・認可を最初に実装

### スケジュールリスク
- **依存関係**: 外部ライブラリの選定を早期に
- **学習コスト**: 新しい技術の習得時間を考慮
- **統合テスト**: 各フェーズでの動作確認

### 運用リスク
- **監視**: 本番環境での監視体制
- **バックアップ**: データベースバックアップ戦略
- **障害対応**: 障害時の復旧手順

## 成功指標

### 技術指標
- **レスポンス時間**: 95%ile < 200ms
- **可用性**: 99.9%以上
- **エラー率**: 1%以下
- **テストカバレッジ**: 80%以上

### ビジネス指標
- **MVP提供**: 6週間以内
- **機能完成**: 12週間以内
- **本番稼働**: 16週間以内

この実装順序により、段階的に機能を構築しながら、早期にMVPを提供し、その後機能を拡張していくことが可能になります。
