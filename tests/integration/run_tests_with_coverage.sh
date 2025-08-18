#!/bin/bash

# フェーズ4 APIテスト カバレッジ測定スクリプト
# このスクリプトは統合テストを実行し、カバレッジを測定します

set -e

echo "=== フェーズ4 APIテスト カバレッジ測定開始 ==="

# テストコンテナ環境の起動
echo "1. テストコンテナ環境を起動中..."
docker-compose -f docker-compose.test.yml up -d postgres redis

# コンテナの起動を待つ
echo "2. コンテナの起動を待機中..."
sleep 10

# テストデータベースの初期化
echo "3. テストデータベースを初期化中..."
docker-compose -f docker-compose.test.yml exec -T postgres psql -U postgres -d access_log_tracker_test -f /docker-entrypoint-initdb.d/01_init_test_db.sql

# 統合テストの実行（カバレッジ測定付き）
echo "4. 統合テストを実行中（カバレッジ測定付き）..."
docker-compose -f docker-compose.test.yml run --rm \
  -e TEST_DB_HOST=postgres \
  -e TEST_REDIS_HOST=redis \
  -e CGO_ENABLED=0 \
  test-runner \
  go test -v -coverprofile=coverage.out -covermode=atomic \
  -coverpkg=./internal/api/...,./internal/domain/...,./internal/infrastructure/... \
  ./tests/integration/...

# カバレッジレポートの生成
echo "5. カバレッジレポートを生成中..."
docker-compose -f docker-compose.test.yml run --rm test-runner \
  go tool cover -html=coverage.out -o coverage.html

# カバレッジの詳細表示
echo "6. カバレッジ詳細を表示中..."
docker-compose -f docker-compose.test.yml run --rm test-runner \
  go tool cover -func=coverage.out

# カバレッジレポートをホストにコピー
echo "7. カバレッジレポートをホストにコピー中..."
docker-compose -f docker-compose.test.yml run --rm test-runner \
  cp coverage.html /workspace/
docker-compose -f docker-compose.test.yml run --rm test-runner \
  cp coverage.out /workspace/

echo "=== カバレッジ測定完了 ==="
echo "カバレッジレポート: coverage.html"
echo "カバレッジデータ: coverage.out"

# テストコンテナ環境の停止
echo "8. テストコンテナ環境を停止中..."
docker-compose -f docker-compose.test.yml down

echo "=== フェーズ4 APIテスト カバレッジ測定終了 ==="
