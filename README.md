# Access Log Tracker (ALT)

複数のWebアプリケーションから利用される汎用アクセスログトラッキングシステム

## 概要

Access Log Tracker (ALT) は、Webサイトのアクセスログを収集・分析するためのトラッキングシステムです。トラッキング用ビーコンの生成・管理とデータ書き込み機能に特化し、月間2000万PVの大規模トラフィックに対応します。

### 主要機能
- ✅ **軽量JavaScriptビーコン**（5KB以下、非同期読み込み）
- ✅ **複数アプリケーション対応**（アプリケーションID管理）
- ✅ **クッキーレストラッキング**（DNT対応、セッション管理）
- ✅ **クローラー検出・除外機能**
- ✅ **高パフォーマンス書き込み**（直接PostgreSQL書き込み）
- ✅ **セキュリティ対策**（API Key認証、レート制限）
- ✅ **月別パーティショニング**（データ保持2年間）

### 技術スタック
- **バックエンド**: Go + Gin Framework
- **データベース**: PostgreSQL 15（月別パーティショニング）
- **キャッシュ**: Redis 7
- **フロントエンド**: Vanilla JavaScript（軽量トラッキングビーコン）
- **開発環境**: Docker + Docker Compose
- **本番環境**: AWS（ALB + EC2 + RDS PostgreSQL）

## 前提条件

- Docker 20.10以上
- Docker Compose 2.0以上
- Go 1.21以上
- Make

## クイックスタート

### 1. 開発環境のセットアップ

```bash
# 環境変数ファイルを作成
make dev-setup

# 開発環境を起動
make dev-up
```

### 2. アプリケーションの起動（ホットリロード）

```bash
# アプリケーションのみを起動（コード変更時に自動リロード）
make dev-up-app
```

### 3. テストの実行

```bash
# すべてのテストを実行（カバレッジ81.8%達成）
make test-all-container

# 単体テストのみ実行
make test-unit

# E2Eテストを実行
make test-e2e-container
```

## トラッキングビーコンの実装

### 基本実装

```html
<!-- 基本実装 -->
<script>
(function() {
    var script = document.createElement('script');
    script.async = true;
    script.src = 'http://localhost:8080/tracker.js'; // 開発環境
    script.setAttribute('data-app-id', 'YOUR_APP_ID');
    var firstScript = document.getElementsByTagName('script')[0];
    firstScript.parentNode.insertBefore(script, firstScript);
})();
</script>
```

### 高度な実装

```html
<!-- カスタム設定付き実装 -->
<script>
window.ALT_CONFIG = {
    app_id: 'YOUR_APP_ID',
    client_sub_id: 'OPTIONAL_SUB_ID',
    module_id: 'OPTIONAL_MODULE_ID',
    endpoint: 'http://localhost:8080/v1/tracking/track',
    debug: false,
    respect_dnt: true,
    session_timeout: 1800000 // 30分
};
</script>
<script async src="http://localhost:8080/tracker.js"></script>
```

## API使用例

### トラッキングデータ送信

```javascript
// トラッキングデータ送信
const response = await fetch('http://localhost:8080/v1/tracking/track', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': 'your_api_key'
  },
  body: JSON.stringify({
    app_id: 'your_app_id',
    user_agent: navigator.userAgent,
    url: window.location.href,
    custom_params: {
      page_type: 'product',
      product_id: '12345',
      product_name: 'Sample Product'
    }
  })
});
```

## 利用可能なコマンド

### 開発環境

```bash
make dev-setup          # 開発環境をセットアップ
make dev-up             # 開発環境を起動
make dev-up-app         # アプリケーションのみを起動（ホットリロード）
make dev-shell          # 開発コンテナにシェルで接続
make dev-logs           # 開発環境のログを表示
make dev-logs-app       # アプリケーションのログを表示
make dev-down           # 開発環境を停止
make dev-clean          # 開発環境をクリーンアップ
```

### ビルド

```bash
make build              # ローカルでビルド
make build-container    # コンテナ内でビルド
make build-all          # すべてのバイナリをローカルでビルド
make build-all-container # すべてのバイナリをコンテナ内でビルド
make build-docker       # Dockerイメージをビルド
```

### テスト

```bash
make test               # ローカルでテスト実行
make test-container     # コンテナ内でテスト実行
make test-all           # すべてのテストをローカルで実行
make test-all-container # すべてのテストをコンテナ内で実行（カバレッジ81.8%）
make test-unit          # 単体テストを実行
make test-integration   # 統合テストを実行
make test-e2e           # E2Eテストを実行
make test-e2e-container # E2Eテストをコンテナ環境で実行
make test-performance   # パフォーマンステストを実行
make test-security      # セキュリティテストを実行
make test-coverage      # テストカバレッジを実行
```

### コード品質

```bash
make lint               # ローカルでリント実行
make lint-container     # コンテナ内でリント実行
make fmt                # ローカルでフォーマット
make fmt-container      # コンテナ内でフォーマット
make fmt-check          # ローカルでフォーマットチェック
make fmt-check-container # コンテナ内でフォーマットチェック
```

### データベース

```bash
make migrate            # マイグレーションを実行
make migrate-create     # 新しいマイグレーションファイルを作成
```

## 環境構成

### 開発環境（docker-compose.yml）

- **アプリケーション**: http://localhost:8080
- **PostgreSQL**: localhost:18432
- **Redis**: localhost:16379
- **pgAdmin**: http://localhost:18081
- **Redis Commander**: http://localhost:18082
- **Prometheus**: http://localhost:19090
- **Grafana**: http://localhost:13000
- **Jaeger**: http://localhost:16686
- **Mailhog**: http://localhost:18025

### テスト環境（docker-compose.test.yml）

- **テスト用PostgreSQL**: localhost:18433
- **テスト用Redis**: localhost:16380
- **テスト用アプリケーション**: http://localhost:8081

## システム構成

### ディレクトリ構造

```
accesslog-tracker/
├── cmd/                    # アプリケーションエントリーポイント
│   ├── api/               # APIサーバー
│   ├── beacon-generator/  # ビーコン生成ツール
│   └── worker/            # ワーカープロセス
├── internal/              # 内部パッケージ
│   ├── api/              # API層（ハンドラー、ミドルウェア、ルート）
│   ├── beacon/           # ビーコン生成
│   ├── config/           # 設定管理
│   ├── domain/           # ドメイン層（モデル、サービス、バリデーター）
│   ├── infrastructure/   # インフラ層（データベース、キャッシュ）
│   └── utils/            # ユーティリティ
├── tests/                # テストファイル
│   ├── unit/            # 単体テスト
│   ├── integration/     # 統合テスト
│   ├── e2e/             # E2Eテスト
│   ├── performance/     # パフォーマンステスト
│   └── security/        # セキュリティテスト
├── deployments/          # デプロイメント設定
├── docs/                # ドキュメント
├── docker-compose.yml   # 開発環境設定
├── docker-compose.test.yml # テスト環境設定
├── Dockerfile           # 本番用Dockerfile
├── Dockerfile.dev       # 開発用Dockerfile
├── Makefile            # ビルド・テスト・デプロイスクリプト
└── README.md           # このファイル
```

## パフォーマンス仕様

- **スループット**: 2000 req/sec以上（簡素化による最適化）
- **レスポンス時間**: 100ms以下（簡素化による安定性）
- **同時接続**: 5000以上（簡素化による安定性）
- **データ保持**: 2年間
- **可用性**: 99.9%以上

## 対応ブラウザ

- Chrome 60+
- Firefox 55+
- Safari 11+
- Edge 79+
- IE 11（制限あり）

## 開発ワークフロー

### 1. 新機能開発

```bash
# 開発環境を起動
make dev-up

# アプリケーションをホットリロードで起動
make dev-up-app

# コードを編集（自動リロード）
# ...

# テストを実行
make test-all-container

# リントを実行
make lint-container
```

### 2. テスト駆動開発（TDD）

```bash
# テストを先に書く
# tests/unit/...

# テストを実行（失敗することを確認）
make test-unit

# 実装を書く
# internal/...

# テストを再実行（成功することを確認）
make test-unit

# 統合テストを実行
make test-integration
```

### 3. 品質チェック

```bash
# フォーマットチェック
make fmt-check-container

# リントチェック
make lint-container

# セキュリティチェック
make test-security

# カバレッジチェック（81.8%達成）
make test-coverage
```

## 実装進捗状況

### 完了済み機能 ✅

#### フェーズ1-3: 基盤実装完了
- ✅ **ドメイン層**: モデル、サービス、バリデーター
- ✅ **インフラ層**: PostgreSQL、Redis接続・リポジトリ
- ✅ **API層**: ハンドラー、ミドルウェア、ルーティング
- ✅ **ビーコン生成**: JavaScriptビーコン生成器
- ✅ **ユーティリティ**: 暗号化、IP処理、JSON処理、ログ、時間処理

#### テスト実装状況
- ✅ **単体テスト**: 全コンポーネント実装完了
- ✅ **統合テスト**: データベース・キャッシュ統合テスト
- ✅ **E2Eテスト**: 基本的な機能テスト
- ✅ **テストカバレッジ**: 81.8%達成（目標80%を超過）

### 次のフェーズ 🚀

#### フェーズ4: ドメインサービス層
- トラッキングサービス（ビジネスロジック）
- アプリケーションサービス（アプリケーション管理）
- セッションサービス（セッション管理）
- 統計サービス（統計情報集計）

#### フェーズ5: API層
- HTTPハンドラー
- ルーティング設定
- ミドルウェア（認証、CORS、レート制限）
- APIエンドポイント

## トラブルシューティング

### よくある問題

1. **ポートが既に使用されている**
   ```bash
   # 使用中のポートを確認
   lsof -i :8080
   
   # コンテナを停止
   make dev-down
   ```

2. **データベース接続エラー**
   ```bash
   # データベースの状態を確認
   docker-compose logs postgres
   
   # データベースを再起動
   docker-compose restart postgres
   ```

3. **テストが失敗する**
   ```bash
   # テスト環境をクリーンアップ
   make test-e2e-cleanup
   
   # テスト環境を再セットアップ
   make test-e2e-setup
   ```

### ログの確認

```bash
# 全サービスのログ
make dev-logs

# 特定サービスのログ
docker-compose logs -f app
docker-compose logs -f postgres
docker-compose logs -f redis
```

## ドキュメント

詳細なドキュメントは `docs/` フォルダにあります：

- [システム概要](./docs/01-overview.md) - プロジェクト概要とシステム要件
- [API仕様書](./docs/02-api-specification.md) - RESTful API設計とエンドポイント
- [トラッキングビーコン仕様書](./docs/03-tracking-beacon.md) - JavaScriptビーコン設計
- [データベース設計仕様書](./docs/04-database-design.md) - PostgreSQL設計とパーティショニング
- [デプロイメントガイド](./docs/05-deployment-guide.md) - 環境セットアップとデプロイ
- [テスト戦略仕様書](./docs/06-testing-strategy.md) - テスト方針と実装
- [カバレッジレポート](./docs/07-coverage-report.md) - テストカバレッジ詳細

## 貢献

1. このリポジトリをフォーク
2. 機能ブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'Add some amazing feature'`)
4. ブランチにプッシュ (`git push origin feature/amazing-feature`)
5. プルリクエストを作成

## ライセンス

このプロジェクトはMITライセンスの下で公開されています。

## 更新履歴

### v4.0.0 (2025-08-17)
- 簡素化構成への移行
- 直接PostgreSQL書き込み
- シンプルな構成による安定性向上
- コスト最適化（月間$50）
- 複雑性の大幅削減

### v3.0.0 (2025-08-15)
- コスト最適化構成への移行
- SQS + Lambdaによるサーバーレス処理
- ElastiCacheによる高速バッファリング
- 大幅なコスト削減（78.6%削減）
- 性能向上（同時接続数20,000以上、スループット10000 req/sec以上）

### v2.0.0 (2025-08-11)
- Go + Nginx構成への移行
- 性能向上（同時接続数10,000以上、スループット5000 req/sec以上）
- レスポンス時間短縮（50ms以下）
- メモリ使用量削減

### v1.0.0 (2025-08-10)
- 初回リリース
- 基本トラッキング機能
- API実装
- データベース設計
- ドキュメント整備
