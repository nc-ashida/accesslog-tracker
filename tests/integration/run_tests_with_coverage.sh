#!/bin/bash

# フェーズ4 APIテスト カバレッジ測定スクリプト
# このスクリプトは統合テストを実行し、カバレッジを測定します

set -e

echo "=== フェーズ4 APIテスト カバレッジ測定開始 ==="

# テストコンテナ環境の起動
echo "1. テストコンテナ環境を起動中..."
# コンテナは既に起動しているため、スキップ

# コンテナの起動を待つ
echo "2. コンテナの起動を待機中..."
sleep 5

# テストデータベースの初期化
echo "3. テストデータベースを初期化中..."
# データベースは既に初期化されているため、スキップ

# 統合テストの実行（カバレッジ測定付き）
echo "4. 統合テストを実行中（カバレッジ測定付き）..."
go test -v -coverprofile=coverage.out -covermode=atomic \
  -coverpkg=./internal/api/...,./internal/domain/...,./internal/infrastructure/... \
  ./tests/integration/...

# カバレッジレポートの生成
echo "5. カバレッジレポートを生成中..."
go tool cover -html=coverage.out -o coverage.html

# カバレッジの詳細表示
echo "6. カバレッジ詳細を表示中..."
go tool cover -func=coverage.out

# カバレッジレポートをホストにコピー
echo "7. カバレッジレポートをホストにコピー中..."
cp coverage.html /workspace/
cp coverage.out /workspace/

echo "=== カバレッジ測定完了 ==="
echo "カバレッジレポート: coverage.html"
echo "カバレッジデータ: coverage.out"

# テストコンテナ環境の停止
echo "8. テストコンテナ環境を停止中..."
# コンテナは停止しない（他のテストで使用中）

echo "=== フェーズ4 APIテスト カバレッジ測定終了 ==="
