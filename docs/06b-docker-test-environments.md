# Dockerテスト環境

## 1. Dockerコンテナ環境でのテスト実行

### 1.1 テスト用Docker Compose設定
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
      - APP_ENV=test
      - CGO_ENABLED=0
      - GOOS=linux
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    volumes:
      - ./:/app
      - /app/vendor
      - ~/.ssh:/root/.ssh:ro
      - go-cache:/go
    networks:
      - test-network
    profiles:
      - test

  test-app:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: access-log-tracker-test-app
    environment:
      - APP_ENV=test
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=access_log_tracker_test
      - DB_USER=postgres
      - DB_PASSWORD=password
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_PASSWORD=
    ports:
      - "8081:8080"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - test-network
    profiles:
      - e2e

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: access_log_tracker_test
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
      POSTGRES_INITDB_ARGS: "--encoding=UTF-8 --lc-collate=C --lc-ctype=C"
    ports:
      - "18433:5432"
    volumes:
      - postgres_test_data:/var/lib/postgresql/data
      - ./deployments/database/init:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres -d access_log_tracker_test"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - test-network

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes --requirepass ""
    ports:
      - "16380:6379"
    volumes:
      - redis_test_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - test-network

volumes:
  postgres_test_data:
    driver: local
  redis_test_data:
    driver: local
  go-cache:
    driver: local

networks:
  test-network:
    driver: bridge
```

### 1.2 テスト用Dockerfile
```dockerfile
# Dockerfile.test
FROM golang:1.21-alpine

# 必要なパッケージをインストール
RUN apk add --no-cache git make openssh-client wget curl

# 作業ディレクトリを設定
WORKDIR /app

# SSH設定ディレクトリを作成
RUN mkdir -p /root/.ssh && chmod 700 /root/.ssh

# 依存関係をコピーしてダウンロード
COPY go.mod go.sum ./
RUN go mod download
RUN go mod tidy

# テストツールをインストール
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
RUN go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
RUN go install github.com/sonatype-nexus-community/nancy@latest

# ソースコードをコピー
COPY . .

# テスト実行用のエントリーポイント
ENTRYPOINT ["make"]
CMD ["test-all"]
```

### 1.3 テスト実行コマンド
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

# カバレッジテスト実行（Dockerコンテナ環境を使用）
make test-coverage-container
```

## 2. コンテナ内テスト実行の設定

### 2.1 テスト用Makefile
```makefile
# Makefile テスト関連コマンド
.PHONY: test-all
test-all: ## すべてのテストを実行
	@echo "すべてのテストを実行中..."
	go test -v ./...

.PHONY: test-unit
test-unit: ## 単体テストを実行
	@echo "単体テストを実行中..."
	go test -v ./tests/unit/...
	go test -v ./internal/domain/...
	go test -v ./internal/utils/...

.PHONY: test-integration
test-integration: ## 統合テストを実行
	@echo "統合テストを実行中..."
	go test -v ./tests/integration/...
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

.PHONY: test-coverage
test-coverage: ## テストカバレッジを実行
	@echo "テストカバレッジを実行中..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "カバレッジレポート: coverage.html"

# Dockerコンテナ環境でのテスト実行
.PHONY: test-in-container
test-in-container: ## Dockerコンテナ内でテストを実行
	@echo "Dockerコンテナ内でテストを実行中..."
	docker-compose -f docker-compose.test.yml --profile test up --build --abort-on-container-exit

.PHONY: test-integration-container
test-integration-container: ## Dockerコンテナ環境で統合テストを実行
	@echo "Dockerコンテナ環境で統合テストを実行中..."
	docker-compose -f docker-compose.test.yml --profile test run --rm test-runner make test-integration

.PHONY: test-e2e-container
test-e2e-container: ## Dockerコンテナ環境でE2Eテストを実行
	@echo "Dockerコンテナ環境でE2Eテストを実行中..."
	docker-compose -f docker-compose.test.yml --profile e2e up --build --abort-on-container-exit

.PHONY: test-e2e-setup
test-e2e-setup: ## E2Eテスト環境をセットアップ
	@echo "E2Eテスト環境をセットアップ中..."
	docker-compose -f docker-compose.test.yml --profile e2e up -d postgres redis
	@echo "E2Eテスト環境のセットアップが完了しました"

.PHONY: test-e2e-run
test-e2e-run: ## E2Eテストを実行（環境は起動済み）
	@echo "E2Eテストを実行中..."
	docker-compose -f docker-compose.test.yml --profile e2e run --rm test-runner make test-e2e

.PHONY: test-e2e-cleanup
test-e2e-cleanup: ## E2Eテスト環境をクリーンアップ
	@echo "E2Eテスト環境をクリーンアップ中..."
	docker-compose -f docker-compose.test.yml --profile e2e down -v

.PHONY: test-performance-container
test-performance-container: ## Dockerコンテナ環境でパフォーマンステストを実行
	@echo "Dockerコンテナ環境でパフォーマンステストを実行中..."
	docker-compose -f docker-compose.test.yml --profile test run --rm test-runner make test-performance

.PHONY: test-coverage-container
test-coverage-container: ## Dockerコンテナ環境でカバレッジテストを実行
	@echo "Dockerコンテナ環境でカバレッジテストを実行中..."
	docker-compose -f docker-compose.test.yml --profile test run --rm test-runner make test-coverage
```

### 2.2 テスト用環境変数
```bash
# .env.test.docker
# Dockerコンテナ内でのテスト環境変数
APP_ENV=test
LOG_LEVEL=debug

# データベース設定（コンテナ内）
DB_HOST=postgres
DB_PORT=5432
DB_NAME=access_log_tracker_test
DB_USER=postgres
DB_PASSWORD=password
DB_SSL_MODE=disable
DB_MAX_OPEN_CONNS=10
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=300s

# Redis設定（コンテナ内）
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=5

# アプリケーション設定
API_PORT=3001
CORS_ORIGIN=http://localhost:3000
RATE_LIMIT_WINDOW=60000
RATE_LIMIT_MAX=1000

# テスト設定
TEST_TIMEOUT=30000
TEST_PARALLEL=4
CGO_ENABLED=0
GOOS=linux
```

## 3. テスト環境の準備

### 3.1 テスト用データベースの準備
```bash
# 開発環境を起動
make dev-up

# テスト用データベースの準備
docker-compose exec postgres psql -U postgres -c "CREATE DATABASE access_log_tracker_test;"
docker-compose exec postgres psql -U postgres -c "CREATE DATABASE access_log_tracker_e2e;"
docker-compose exec postgres psql -U postgres -c "CREATE DATABASE access_log_tracker_perf;"
docker-compose exec postgres psql -U postgres -c "CREATE DATABASE access_log_tracker_security;"
```

### 3.2 テスト用マイグレーション実行
```bash
# テスト用データベースにマイグレーションを実行
docker-compose exec postgres psql -U postgres -d access_log_tracker_test -f /app/deployments/database/migrations/001_create_applications_table.sql
docker-compose exec postgres psql -U postgres -d access_log_tracker_test -f /app/deployments/database/migrations/002_create_access_logs_table.sql
docker-compose exec postgres psql -U postgres -d access_log_tracker_test -f /app/deployments/database/migrations/003_create_sessions_table.sql
docker-compose exec postgres psql -U postgres -d access_log_tracker_test -f /app/deployments/database/migrations/004_create_custom_parameters_table.sql
```

## 4. テスト実行とレポート

### 4.1 テスト実行コマンド
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

### 4.2 Dockerコンテナ内でのテスト実行
```bash
# Dockerコンテナ内でテストを実行
make test-in-container

# 特定のテストをコンテナ内で実行
make test-integration-container
make test-e2e-container
make test-performance-container
make test-coverage-container

# E2Eテストの段階的実行
make test-e2e-setup
make test-e2e-run
make test-e2e-cleanup
```

### 4.3 テスト結果の確認
```bash
# テストログの確認
docker-compose logs test-runner

# カバレッジレポートの確認
open coverage.html

# テストレポートの確認
cat test-report.json
```

## 5. CI/CDパイプラインでのテスト実行

### 5.1 GitHub Actions設定
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

### 5.2 GitLab CI設定
```yaml
# .gitlab-ci.yml
stages:
  - test

test:
  stage: test
  image: golang:1.21-alpine
  services:
    - postgres:15-alpine
    - redis:7-alpine
  variables:
    POSTGRES_DB: access_log_tracker_test
    POSTGRES_USER: postgres
    POSTGRES_PASSWORD: password
    REDIS_HOST: redis
    REDIS_PORT: 6379
  before_script:
    - apk add --no-cache git make postgresql-client redis
    - go mod download
  script:
    - make test-all
    - make test-coverage
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.xml
    paths:
      - coverage.html
      - test-report.json
```

## 6. テスト環境のトラブルシューティング

### 6.1 よくある問題と解決方法

#### 6.1.1 データベース接続エラー
```bash
# 問題: データベースに接続できない
# 解決方法: コンテナの状態を確認
docker-compose ps

# データベースコンテナを再起動
docker-compose restart postgres

# データベース接続をテスト
docker-compose exec postgres pg_isready -U postgres
```

#### 6.1.2 Redis接続エラー
```bash
# 問題: Redisに接続できない
# 解決方法: Redisコンテナの状態を確認
docker-compose ps redis

# Redisコンテナを再起動
docker-compose restart redis

# Redis接続をテスト
docker-compose exec redis redis-cli ping
```

#### 6.1.3 テストタイムアウト
```bash
# 問題: テストがタイムアウトする
# 解決方法: テストタイムアウトを延長
export TEST_TIMEOUT=60000

# または、Makefileで設定
test-all:
	go test -v -timeout 60s ./...
```

### 6.2 ログ確認方法
```bash
# テストランナーのログを確認
docker-compose logs test-runner

# データベースのログを確認
docker-compose logs postgres

# Redisのログを確認
docker-compose logs redis

# すべてのログを確認
docker-compose logs
```

### 6.3 テスト環境のクリーンアップ
```bash
# テスト環境を停止
docker-compose -f docker-compose.test.yml down

# テスト用ボリュームを削除
docker-compose -f docker-compose.test.yml down -v

# テスト用イメージを削除
docker rmi accesslog-tracker_test-runner

# すべてのテスト関連コンテナを削除
docker container prune -f
```

## 7. パフォーマンス最適化

### 7.1 テスト実行時間の短縮
```yaml
# docker-compose.test.yml の最適化
services:
  test-runner:
    # 並列テスト実行
    environment:
      - TEST_PARALLEL=4
      - GO_TEST_TIMEOUT=30s
    # キャッシュボリュームの追加
    volumes:
      - ./:/app
      - /app/vendor
      - go-cache:/go
      - test-cache:/app/.cache

volumes:
  go-cache:
  test-cache:
```

### 7.2 リソース制限の設定
```yaml
# docker-compose.test.yml のリソース制限
services:
  test-runner:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '1'
          memory: 1G
  
  postgres:
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 1G
  
  redis:
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 512M
```
