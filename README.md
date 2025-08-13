# Access Log Tracker

アクセスログトラッキングシステムのDocker環境

## 概要

このプロジェクトは、Webサイトのアクセスログを収集・分析するためのトラッキングシステムです。Dockerを使用して開発・テスト・本番環境を統一し、TDD（テスト駆動開発）を徹底した高品質なシステムを構築します。

## 前提条件

- Docker
- Docker Compose
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
# すべてのテストを実行
make test-all-container

# 単体テストのみ実行
make test-unit

# E2Eテストを実行
make test-e2e-container
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
make test-all-container # すべてのテストをコンテナ内で実行
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

## プロファイル

### 開発プロファイル

```bash
# 全サービスを起動
docker-compose up -d

# アプリケーションのみ起動
docker-compose up app

# ビルドサービスを使用
docker-compose --profile build run --rm builder make build
```

### テストプロファイル

```bash
# テスト環境を起動
docker-compose -f docker-compose.test.yml --profile test up -d

# E2Eテスト環境を起動
docker-compose -f docker-compose.test.yml --profile e2e up -d
```

## ディレクトリ構造

```
accesslog-tracker/
├── cmd/                    # アプリケーションエントリーポイント
├── internal/               # 内部パッケージ
├── pkg/                    # 公開パッケージ
├── tests/                  # テストファイル
├── deployments/            # デプロイメント設定
├── docs/                   # ドキュメント
├── docker-compose.yml      # 開発環境設定
├── docker-compose.test.yml # テスト環境設定
├── Dockerfile              # 本番用Dockerfile
├── Dockerfile.dev          # 開発用Dockerfile
├── Dockerfile.test         # テスト用Dockerfile
├── .air.toml              # Air設定（ホットリロード）
├── Makefile               # ビルド・テスト・デプロイスクリプト
└── README.md              # このファイル
```

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

# カバレッジチェック
make test-coverage
```

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

## 貢献

1. このリポジトリをフォーク
2. 機能ブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'Add some amazing feature'`)
4. ブランチにプッシュ (`git push origin feature/amazing-feature`)
5. プルリクエストを作成

## ライセンス

このプロジェクトはMITライセンスの下で公開されています。
