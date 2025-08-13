# Makefile for Access Log Tracker

# 変数定義
APP_NAME := access-log-tracker
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION := $(shell go version | awk '{print $$3}')
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GoVersion=$(GO_VERSION)"

# デフォルトターゲット
.DEFAULT_GOAL := help

# ヘルプ
.PHONY: help
help: ## 利用可能なコマンドを表示
	@echo "Access Log Tracker - 利用可能なコマンド:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# 開発環境
.PHONY: dev-setup
dev-setup: ## 開発環境をセットアップ
	@echo "開発環境をセットアップ中..."
	cp env.example .env
	@echo "環境変数ファイルを作成しました: .env"
	@echo "必要に応じて .env ファイルを編集してください"

.PHONY: dev-up
dev-up: ## 開発環境を起動
	@echo "開発環境を起動中..."
	docker-compose up -d
	@echo "開発環境が起動しました"
	@echo "アプリケーション: http://localhost:8080"
	@echo "PostgreSQL: localhost:18432"
	@echo "Redis: localhost:16379"
	@echo "pgAdmin: http://localhost:18081"
	@echo "Redis Commander: http://localhost:18082"
	@echo "Prometheus: http://localhost:19090"
	@echo "Grafana: http://localhost:13000"
	@echo "Jaeger: http://localhost:16686"
	@echo "Mailhog: http://localhost:18025"

.PHONY: dev-up-app
dev-up-app: ## アプリケーションのみを起動（ホットリロード）
	@echo "アプリケーションを起動中（ホットリロード）..."
	docker-compose up app

.PHONY: dev-shell
dev-shell: ## 開発コンテナにシェルで接続
	@echo "開発コンテナに接続中..."
	docker-compose run --rm builder /bin/sh

.PHONY: dev-logs
dev-logs: ## 開発環境のログを表示
	docker-compose logs -f

.PHONY: dev-logs-app
dev-logs-app: ## アプリケーションのログを表示
	docker-compose logs -f app

.PHONY: dev-down
dev-down: ## 開発環境を停止
	@echo "開発環境を停止中..."
	docker-compose down

.PHONY: dev-clean
dev-clean: ## 開発環境をクリーンアップ
	@echo "開発環境をクリーンアップ中..."
	docker-compose down -v
	docker system prune -f

# 依存関係
.PHONY: deps
deps: ## 依存関係をダウンロード
	@echo "依存関係をダウンロード中..."
	go mod download
	go mod tidy

.PHONY: deps-update
deps-update: ## 依存関係を更新
	@echo "依存関係を更新中..."
	go get -u ./...
	go mod tidy

# ビルド
.PHONY: build
build: ## アプリケーションをビルド
	@echo "アプリケーションをビルド中..."
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/api $(LDFLAGS) ./cmd/api

.PHONY: build-container
build-container: ## コンテナ内でアプリケーションをビルド
	@echo "コンテナ内でアプリケーションをビルド中..."
	docker-compose run --rm builder make build

.PHONY: build-all
build-all: ## すべてのバイナリをビルド
	@echo "すべてのバイナリをビルド中..."
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/api $(LDFLAGS) ./cmd/api
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/worker $(LDFLAGS) ./cmd/worker
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/beacon-generator $(LDFLAGS) ./cmd/beacon-generator

.PHONY: build-all-container
build-all-container: ## コンテナ内ですべてのバイナリをビルド
	@echo "コンテナ内ですべてのバイナリをビルド中..."
	docker-compose run --rm builder make build-all

.PHONY: build-docker
build-docker: ## Dockerイメージをビルド
	@echo "Dockerイメージをビルド中..."
	docker build -t $(APP_NAME):$(VERSION) .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest

# テスト
.PHONY: test
test: ## テストを実行
	@echo "テストを実行中..."
	go test -v ./...

.PHONY: test-container
test-container: ## コンテナ内でテストを実行
	@echo "コンテナ内でテストを実行中..."
	docker-compose -f docker-compose.test.yml --profile test run --rm test-runner make test

.PHONY: test-all
test-all: ## すべてのテストを実行
	@echo "すべてのテストを実行中..."
	make test-unit
	make test-integration
	make test-e2e
	make test-performance
	make test-security
	make test-coverage

.PHONY: test-all-container
test-all-container: ## コンテナ内ですべてのテストを実行
	@echo "コンテナ内ですべてのテストを実行中..."
	docker-compose -f docker-compose.test.yml --profile test run --rm test-runner make test-all

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

.PHONY: test-benchmark
test-benchmark: ## ベンチマークテストを実行
	@echo "ベンチマークテストを実行中..."
	go test -bench=. -benchmem ./...

.PHONY: test-race
test-race: ## レースコンディションテストを実行
	@echo "レースコンディションテストを実行中..."
	go test -race ./...

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
	docker-compose -f docker-compose.test.yml run --rm test-runner make test-performance

.PHONY: test-coverage-container
test-coverage-container: ## Dockerコンテナ環境でカバレッジテストを実行
	@echo "Dockerコンテナ環境でカバレッジテストを実行中..."
	docker-compose -f docker-compose.test.yml run --rm test-runner make test-coverage

.PHONY: test-setup-db
test-setup-db: ## テスト用データベースをセットアップ
	@echo "テスト用データベースをセットアップ中..."
	docker-compose exec postgres psql -U postgres -c "CREATE DATABASE IF NOT EXISTS access_log_tracker_test;"
	docker-compose exec postgres psql -U postgres -c "CREATE DATABASE IF NOT EXISTS access_log_tracker_e2e;"
	docker-compose exec postgres psql -U postgres -c "CREATE DATABASE IF NOT EXISTS access_log_tracker_perf;"
	docker-compose exec postgres psql -U postgres -c "CREATE DATABASE IF NOT EXISTS access_log_tracker_security;"
	@echo "テスト用データベースのセットアップが完了しました"

# リント・フォーマット
.PHONY: lint
lint: ## コードをリント
	@echo "コードをリント中..."
	golangci-lint run

.PHONY: lint-container
lint-container: ## コンテナ内でコードをリント
	@echo "コンテナ内でコードをリント中..."
	docker-compose run --rm builder golangci-lint run

.PHONY: fmt
fmt: ## コードをフォーマット
	@echo "コードをフォーマット中..."
	go fmt ./...

.PHONY: fmt-container
fmt-container: ## コンテナ内でコードをフォーマット
	@echo "コンテナ内でコードをフォーマット中..."
	docker-compose run --rm builder go fmt ./...

.PHONY: fmt-check
fmt-check: ## フォーマットをチェック
	@echo "フォーマットをチェック中..."
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "フォーマットされていないファイルがあります:"; \
		gofmt -l .; \
		exit 1; \
	fi

.PHONY: fmt-check-container
fmt-check-container: ## コンテナ内でフォーマットをチェック
	@echo "コンテナ内でフォーマットをチェック中..."
	docker-compose run --rm builder make fmt-check

# データベース
.PHONY: migrate
migrate: ## データベースマイグレーションを実行
	@echo "データベースマイグレーションを実行中..."
	@if [ -f "./bin/migrate" ]; then \
		./bin/migrate up; \
	else \
		echo "マイグレーションツールが見つかりません"; \
	fi

.PHONY: migrate-create
migrate-create: ## 新しいマイグレーションファイルを作成
	@echo "新しいマイグレーションファイルを作成中..."
	@read -p "マイグレーション名を入力してください: " name; \
	./bin/migrate create -ext sql -dir deployments/database/migrations $$name

# 実行
.PHONY: run
run: ## アプリケーションを実行
	@echo "アプリケーションを実行中..."
	go run ./cmd/api

.PHONY: run-dev
run-dev: ## 開発モードでアプリケーションを実行
	@echo "開発モードでアプリケーションを実行中..."
	APP_ENV=development go run ./cmd/api

# デプロイ
.PHONY: deploy-local
deploy-local: ## ローカルにデプロイ
	@echo "ローカルにデプロイ中..."
	docker-compose -f docker-compose.prod.yml up -d

.PHONY: deploy-k8s
deploy-k8s: ## Kubernetesにデプロイ
	@echo "Kubernetesにデプロイ中..."
	kubectl apply -f deployments/kubernetes/

.PHONY: deploy-aws
deploy-aws: ## AWSにデプロイ
	@echo "AWSにデプロイ中..."
	aws cloudformation deploy \
		--template-file deployments/aws/cloudformation/infrastructure.yml \
		--stack-name $(APP_NAME) \
		--capabilities CAPABILITY_IAM

# 監視・ログ
.PHONY: logs
logs: ## アプリケーションログを表示
	@echo "アプリケーションログを表示中..."
	docker logs -f $(APP_NAME)

.PHONY: monitor
monitor: ## 監視ダッシュボードを開く
	@echo "監視ダッシュボードを開く中..."
	open http://localhost:3000  # Grafana
	open http://localhost:9090  # Prometheus
	open http://localhost:16686 # Jaeger

# クリーンアップ
.PHONY: clean
clean: ## ビルドファイルをクリーンアップ
	@echo "ビルドファイルをクリーンアップ中..."
	rm -rf bin/
	rm -f coverage.out coverage.html

.PHONY: clean-docker
clean-docker: ## Dockerイメージをクリーンアップ
	@echo "Dockerイメージをクリーンアップ中..."
	docker rmi $(APP_NAME):$(VERSION) $(APP_NAME):latest 2>/dev/null || true

# セキュリティ
.PHONY: security-scan
security-scan: ## セキュリティスキャンを実行
	@echo "セキュリティスキャンを実行中..."
	gosec ./...

.PHONY: audit
audit: ## 依存関係の監査を実行
	@echo "依存関係の監査を実行中..."
	go list -json -deps ./... | nancy sleuth

# ドキュメント
.PHONY: docs
docs: ## ドキュメントを生成
	@echo "ドキュメントを生成中..."
	godoc -http=:6060 &
	@echo "GoDoc: http://localhost:6060"

.PHONY: swagger
swagger: ## Swaggerドキュメントを生成
	@echo "Swaggerドキュメントを生成中..."
	swag init -g cmd/api/main.go

# リリース
.PHONY: release
release: ## リリースを作成
	@echo "リリースを作成中..."
	@if [ -z "$(VERSION)" ]; then \
		echo "バージョンを指定してください: make release VERSION=v1.0.0"; \
		exit 1; \
	fi
	git tag $(VERSION)
	git push origin $(VERSION)

# 開発者向け
.PHONY: dev-install-tools
dev-install-tools: ## 開発ツールをインストール
	@echo "開発ツールをインストール中..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install github.com/sonatype-nexus-community/nancy@latest
	go install github.com/swaggo/swag/cmd/swag@latest

.PHONY: dev-check
dev-check: ## 開発環境のチェック
	@echo "開発環境をチェック中..."
	@echo "Go version: $(GO_VERSION)"
	@echo "Git version: $(shell git --version)"
	@echo "Docker version: $(shell docker --version)"
	@echo "Docker Compose version: $(shell docker-compose --version)"
	@echo "golangci-lint: $(shell golangci-lint --version 2>/dev/null || echo 'not installed')"
	@echo "gosec: $(shell gosec --version 2>/dev/null || echo 'not installed')"

# ヘルスチェック
.PHONY: health-check
health-check: ## ヘルスチェックを実行
	@echo "ヘルスチェックを実行中..."
	@curl -f http://localhost:8080/health || echo "アプリケーションが起動していません"
	@curl -f http://localhost:5432 || echo "PostgreSQLが起動していません"
	@curl -f http://localhost:6379 || echo "Redisが起動していません"

# バックアップ・リストア
.PHONY: backup
backup: ## データベースをバックアップ
	@echo "データベースをバックアップ中..."
	@mkdir -p backups
	docker exec access-log-tracker-postgres pg_dump -U postgres access_log_tracker > backups/backup_$(shell date +%Y%m%d_%H%M%S).sql

.PHONY: restore
restore: ## データベースをリストア
	@echo "データベースをリストア中..."
	@if [ -z "$(FILE)" ]; then \
		echo "バックアップファイルを指定してください: make restore FILE=backups/backup_20231201_120000.sql"; \
		exit 1; \
	fi
	docker exec -i access-log-tracker-postgres psql -U postgres access_log_tracker < $(FILE)
