# パフォーマンステスト

このディレクトリには、Access Log Trackerアプリケーションのパフォーマンステストが含まれています。

## 概要

パフォーマンステストは以下のコンポーネントを対象としています：

1. **ビーコントラッキング** (`beacon_performance_test.go`)
   - ビーコンエンドポイントのパフォーマンス測定
   - スループット、レイテンシー、メモリ使用量のテスト

2. **データベース操作** (`database_performance_test.go`)
   - PostgreSQLデータベースのパフォーマンス測定
   - アプリケーションリポジトリとトラッキングリポジトリのテスト

3. **Redisキャッシュ** (`redis_performance_test.go`)
   - Redisキャッシュのパフォーマンス測定
   - Set/Get操作のスループットとレイテンシーテスト

## テスト環境の準備

### 1. Dockerコンテナ環境の起動

```bash
# テスト環境の起動
docker-compose -f docker-compose.test.yml up -d

# 環境の確認
docker-compose -f docker-compose.test.yml ps
```

### 2. テスト用データベースのセットアップ

```bash
# データベースのセットアップ
make test-setup-db
```

## テストの実行

### 基本的なパフォーマンステスト実行

```bash
# すべてのパフォーマンステストを実行
make test-performance

# Dockerコンテナ環境でパフォーマンステストを実行
make test-performance-container
```

### 個別のテスト実行

```bash
# ビーコントラッキングテストのみ実行
go test -v -bench=. -benchmem ./tests/performance/beacon_performance_test.go

# データベーステストのみ実行
go test -v -bench=. -benchmem ./tests/performance/database_performance_test.go

# Redisテストのみ実行
go test -v -bench=. -benchmem ./tests/performance/redis_performance_test.go
```

### ベンチマークテストの実行

```bash
# ベンチマークテストのみ実行
go test -bench=. -benchmem ./tests/performance/

# 特定のベンチマークテスト実行
go test -bench=BenchmarkBeaconRequests -benchmem ./tests/performance/
```

## テスト内容

### 1. ビーコントラッキングテスト

#### ベンチマークテスト
- `BenchmarkBeaconRequests`: シーケンシャルビーコンリクエスト
- `BenchmarkConcurrentBeaconRequests`: 並行ビーコンリクエスト
- `BenchmarkTrackingAPIRequests`: トラッキングAPIリクエスト

#### スループットテスト
- **シーケンシャルリクエスト**: 最低50 req/s
- **並行リクエスト**: 最低100 req/s
- **高負荷テスト**: 最低200 req/s（5000リクエスト、50ワーカー）

#### レイテンシーテスト
- **平均レイテンシー**: 100ms以下
- **95パーセンタイル**: 200ms以下
- **99パーセンタイル**: 500ms以下

#### メモリ使用量テスト
- **メモリリーク検出**: 1MB以下の増加
- **長時間実行**: 2MB以下の成長

#### ストレステスト
- **継続負荷**: 30秒間、50 req/s
- **バースト負荷**: 5回のバースト、各200リクエスト

### 2. データベーステスト

#### ベンチマークテスト
- `BenchmarkDatabaseConnection`: データベース接続
- `BenchmarkApplicationRepositoryOperations`: アプリケーションリポジトリ操作
- `BenchmarkTrackingRepositoryOperations`: トラッキングリポジトリ操作

#### スループットテスト
- **シーケンシャル操作**: 最低50 ops/s
- **並行操作**: 最低100 ops/s
- **高負荷テスト**: 最低200 ops/s（5000操作、50ワーカー）

#### レイテンシーテスト
- **平均レイテンシー**: 50ms以下
- **95パーセンタイル**: 100ms以下
- **99パーセンタイル**: 200ms以下

#### メモリ使用量テスト
- **メモリリーク検出**: 1MB以下の増加
- **長時間実行**: 2MB以下の成長

### 3. Redisテスト

#### ベンチマークテスト
- `BenchmarkRedisConnection`: Redis接続
- `BenchmarkRedisSetOperations`: Set操作
- `BenchmarkRedisGetOperations`: Get操作

#### スループットテスト
- **Set操作**: 最低1000 ops/s
- **Get操作**: 最低2000 ops/s
- **並行操作**: 最低2000 ops/s
- **高負荷テスト**: 最低5000 ops/s（10000操作、100ワーカー）

#### レイテンシーテスト
- **Set操作**: 平均10ms以下、最大50ms以下
- **Get操作**: 平均5ms以下、最大20ms以下
- **負荷下**: 平均20ms以下

#### メモリ使用量テスト
- **メモリリーク検出**: 5MB以下の増加
- **長時間実行**: 10MB以下の成長

## パフォーマンス要件

### ビーコントラッキング
- **スループット**: 200 req/s以上
- **レイテンシー**: 平均100ms以下、95%が200ms以下
- **成功率**: 95%以上
- **メモリ**: リークなし

### データベース
- **スループット**: 200 ops/s以上
- **レイテンシー**: 平均50ms以下、95%が100ms以下
- **成功率**: 95%以上
- **メモリ**: リークなし

### Redisキャッシュ
- **スループット**: 5000 ops/s以上
- **レイテンシー**: 平均10ms以下、95%が20ms以下
- **成功率**: 95%以上
- **メモリ**: リークなし

## テスト結果の解釈

### スループット
- **目標値以上**: パフォーマンス要件を満たしている
- **目標値未満**: パフォーマンスの最適化が必要

### レイテンシー
- **目標値以下**: レスポンス時間要件を満たしている
- **目標値超過**: レスポンス時間の改善が必要

### 成功率
- **95%以上**: 安定性要件を満たしている
- **95%未満**: エラーハンドリングの改善が必要

### メモリ使用量
- **リークなし**: メモリ管理が適切
- **リークあり**: メモリリークの修正が必要

## トラブルシューティング

### よくある問題

1. **テストがタイムアウトする**
   ```bash
   # タイムアウトを延長
   go test -timeout 60s ./tests/performance/
   ```

2. **データベース接続エラー**
   ```bash
   # データベースの再起動
   docker-compose -f docker-compose.test.yml restart postgres
   ```

3. **Redis接続エラー**
   ```bash
   # Redisの再起動
   docker-compose -f docker-compose.test.yml restart redis
   ```

4. **メモリ不足**
   ```bash
   # メモリ制限を調整
   docker-compose -f docker-compose.test.yml run --rm --memory=2g test-runner make test-performance
   ```

### ログの確認

```bash
# テストランナーのログ
docker-compose -f docker-compose.test.yml logs test-runner

# データベースのログ
docker-compose -f docker-compose.test.yml logs postgres

# Redisのログ
docker-compose -f docker-compose.test.yml logs redis
```

## CI/CDでの利用

### GitHub Actionsでの実行例

```yaml
- name: Run Performance Tests
  run: |
    docker-compose -f docker-compose.test.yml up -d
    sleep 10
    make test-setup-db
    make test-performance-container
```

### パフォーマンステストの自動化

```bash
# 定期実行用スクリプト
#!/bin/bash
set -e

echo "Starting performance tests..."
docker-compose -f docker-compose.test.yml up -d
sleep 10
make test-setup-db
make test-performance-container

echo "Performance tests completed successfully"
```

## 注意事項

1. **テスト環境の分離**: 本番環境とは完全に分離されたテスト環境を使用
2. **データのクリーンアップ**: テスト後にテストデータを適切に削除
3. **リソース管理**: テスト実行時のリソース使用量を監視
4. **結果の記録**: パフォーマンステスト結果を記録・分析

## 関連ドキュメント

- [テスト戦略](../README.md)
- [統合テスト](../integration/README.md)
- [E2Eテスト](../e2e/README.md)
- [セキュリティテスト](../security/README.md)
