# Access Log Tracker - テストガイド

このディレクトリには、Access Log Trackerプロジェクトの包括的なテストスイートが含まれています。

## テスト構造

```
tests/
├── unit/                    # 単体テスト
│   └── domain/
│       └── services/
│           ├── application_service_test.go
│           └── tracking_service_test.go
├── integration/             # 統合テスト
│   ├── api/
│   │   └── handlers_test.go
│   └── database/
│       └── repositories_test.go
├── e2e/                     # E2Eテスト
│   └── beacon_tracking_test.go
├── performance/             # パフォーマンステスト
│   └── beacon_performance_test.go
├── security/                # セキュリティテスト
│   └── security_test.go
├── test-config.yml          # テスト設定ファイル
└── README.md               # このファイル
```

## テストの種類

### 1. 単体テスト (`unit/`)
- **目的**: 個別の関数、メソッド、クラスの動作を検証
- **対象**: ドメインサービス、ユーティリティ関数
- **実行時間**: 短い（数秒）
- **依存関係**: 外部依存なし

### 2. 統合テスト (`integration/`)
- **目的**: 複数のコンポーネント間の連携を検証
- **対象**: APIハンドラー、データベースリポジトリ
- **実行時間**: 中程度（数十秒）
- **依存関係**: データベース、Redis

### 3. E2Eテスト (`e2e/`)
- **目的**: エンドツーエンドのユーザーシナリオを検証
- **対象**: ビーコントラッキングフロー全体
- **実行時間**: 長い（数分）
- **依存関係**: 完全なアプリケーション環境

### 4. パフォーマンステスト (`performance/`)
- **目的**: システムの性能とスケーラビリティを検証
- **対象**: スループット、レイテンシー、メモリ使用量
- **実行時間**: 長い（数分〜数十分）
- **依存関係**: 本番環境に近い設定

### 5. セキュリティテスト (`security/`)
- **目的**: セキュリティ脆弱性を検出
- **対象**: 認証、認可、入力検証、データ保護
- **実行時間**: 中程度（数分）
- **依存関係**: セキュリティ設定

## テストの実行

### 基本的な実行方法

```bash
# すべてのテストを実行
make test-all

# 特定のテストタイプを実行
make test-unit
make test-integration
make test-e2e
make test-performance
make test-security
make test-coverage
```

### テストスクリプトを使用

```bash
# すべてのテストを実行
./scripts/run-tests.sh

# 特定のテストタイプを実行
./scripts/run-tests.sh unit
./scripts/run-tests.sh integration
./scripts/run-tests.sh e2e
./scripts/run-tests.sh performance
./scripts/run-tests.sh security
./scripts/run-tests.sh coverage

# Dockerコンテナ内でテスト実行
./scripts/run-tests.sh docker
```

### Dockerコンテナ内での実行

```bash
# テスト用コンテナを起動してテスト実行
make test-in-container

# 特定のテストをコンテナ内で実行
make test-integration-container
make test-e2e-container
make test-performance-container
make test-coverage-container
```

## テスト環境の準備

### 前提条件

1. **Go 1.21以上**がインストールされている
2. **Docker**と**Docker Compose**がインストールされている
3. **PostgreSQL**と**Redis**が利用可能

### 環境セットアップ

```bash
# 開発環境を起動
make dev-up

# テスト用データベースをセットアップ
make test-setup-db
```

### テスト用データベース

以下のデータベースが自動的に作成されます：

- `access_log_tracker_test` - 単体・統合テスト用
- `access_log_tracker_e2e` - E2Eテスト用
- `access_log_tracker_perf` - パフォーマンステスト用
- `access_log_tracker_security` - セキュリティテスト用

## テスト設定

テスト設定は `test-config.yml` ファイルで管理されています：

```yaml
# データベース設定
database:
  test:
    host: "localhost"
    port: 5432
    name: "access_log_tracker_test"
    # ...

# テスト設定
test:
  unit:
    timeout: 30s
    parallel: 4
  # ...
```

## テストカバレッジ

カバレッジレポートを生成：

```bash
make test-coverage
```

生成されるファイル：
- `coverage.out` - 生のカバレッジデータ
- `coverage.html` - HTML形式のカバレッジレポート

## テストレポート

テスト結果は以下の形式で出力されます：

- **コンソール出力**: リアルタイムのテスト結果
- **JUnit XML**: CI/CDパイプライン用
- **HTMLレポート**: ブラウザでの確認用
- **JSONレポート**: 自動化ツール用

## ベストプラクティス

### テストの書き方

1. **テスト名は明確に**: 何をテストしているかが分かる名前にする
2. **AAAパターン**: Arrange（準備）、Act（実行）、Assert（検証）
3. **独立性**: テストは互いに独立していること
4. **再現性**: 同じ結果が得られること

### テストデータ

1. **フィクスチャ**: 再利用可能なテストデータを作成
2. **クリーンアップ**: テスト後にデータをクリーンアップ
3. **ランダム化**: 必要に応じてランダムデータを使用

### パフォーマンス

1. **並列実行**: 可能な限り並列でテストを実行
2. **モック**: 外部依存はモックを使用
3. **タイムアウト**: 適切なタイムアウトを設定

## トラブルシューティング

### よくある問題

1. **データベース接続エラー**
   ```bash
   # データベースが起動しているか確認
   docker-compose ps
   
   # テスト用データベースを再作成
   make test-setup-db
   ```

2. **テストがタイムアウトする**
   - テスト設定のタイムアウト値を増やす
   - 並列度を下げる

3. **カバレッジが低い**
   - テストケースを追加
   - 除外パターンを確認

### デバッグ

```bash
# 詳細なログでテスト実行
go test -v -count=1 ./tests/...

# 特定のテストのみ実行
go test -v -run TestSpecificFunction ./tests/...

# レースコンディション検出
go test -race ./tests/...
```

## CI/CD統合

### GitHub Actions

```yaml
- name: Run tests
  run: |
    make test-all
    make test-coverage
```

### GitLab CI

```yaml
test:
  script:
    - make test-all
    - make test-coverage
```

## 貢献

新しいテストを追加する際は：

1. 適切なディレクトリに配置
2. テスト名は `Test` で始める
3. テスト設定を `test-config.yml` に追加
4. ドキュメントを更新

## 参考資料

- [Go Testing Package](https://golang.org/pkg/testing/)
- [Testify](https://github.com/stretchr/testify)
- [Gin Testing](https://gin-gonic.com/docs/testing/)
- [Docker Testing Best Practices](https://docs.docker.com/develop/dev-best-practices/)
