#!/bin/bash

# 仕様書準拠テスト実行スクリプト
# 実際のアプリケーションコードのカバレッジを測定

set -e

echo "=== 仕様書準拠テスト実行 ==="
echo "日時: $(date)"
echo ""

# テスト環境の確認
echo "1. テスト環境の確認"
docker-compose ps | grep -E "(postgres|redis)" || {
    echo "エラー: テスト用データベースが起動していません"
    echo "docker-compose up -d postgres redis を実行してください"
    exit 1
}

# データベースの状態確認
echo "2. データベースの状態確認"
docker-compose exec postgres psql -U postgres -d access_log_tracker_test -c "\dt" || {
    echo "エラー: テスト用データベースに接続できません"
    exit 1
}

# テストの実行
echo "3. 仕様書準拠テストの実行"
echo "テストファイル: tests/integration/specification_compliance_test.go"
echo ""

# 実際のアプリケーションコードのカバレッジを測定
go test ./tests/integration/specification_compliance_test.go \
    -coverprofile=specification_coverage.out \
    -coverpkg=./internal/... \
    -v

# カバレッジレポートの生成
echo ""
echo "4. カバレッジレポートの生成"
go tool cover -func=specification_coverage.out

# 全体カバレッジの表示
echo ""
echo "5. 全体カバレッジ"
TOTAL_COVERAGE=$(go tool cover -func=specification_coverage.out | tail -1 | awk '{print $3}' | sed 's/%//')
echo "全体カバレッジ: ${TOTAL_COVERAGE}%"

# コンポーネント別カバレッジの表示
echo ""
echo "6. コンポーネント別カバレッジ"
echo "=== API層 ==="
go tool cover -func=specification_coverage.out | grep "internal/api" || echo "API層: カバレッジなし"

echo ""
echo "=== Domain層 ==="
go tool cover -func=specification_coverage.out | grep "internal/domain" || echo "Domain層: カバレッジなし"

echo ""
echo "=== Infrastructure層 ==="
go tool cover -func=specification_coverage.out | grep "internal/infrastructure" || echo "Infrastructure層: カバレッジなし"

echo ""
echo "=== Utils層 ==="
go tool cover -func=specification_coverage.out | grep "internal/utils" || echo "Utils層: カバレッジなし"

# HTMLレポートの生成
echo ""
echo "7. HTMLカバレッジレポートの生成"
go tool cover -html=specification_coverage.out -o specification_coverage.html
echo "HTMLレポート: specification_coverage.html"

# テスト結果の要約
echo ""
echo "=== テスト結果要約 ==="
echo "実行日時: $(date)"
echo "全体カバレッジ: ${TOTAL_COVERAGE}%"
echo "カバレッジファイル: specification_coverage.out"
echo "HTMLレポート: specification_coverage.html"

# 仕様書準拠の確認
echo ""
echo "=== 仕様書準拠確認 ==="
if [ "$TOTAL_COVERAGE" -ge 80 ]; then
    echo "✅ カバレッジ80%以上を達成"
else
    echo "⚠️  カバレッジ80%未満 (${TOTAL_COVERAGE}%)"
fi

echo ""
echo "=== テスト完了 ==="
