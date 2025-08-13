# テスト環境設定

## 1. 環境構成

### 1.1 環境別設定
```yaml
# test-environments.yml
environments:
  unit:
    database: sqlite
    redis: mock
    dependencies: mocked
    
  integration:
    database: postgresql_test
    redis: redis_test
    dependencies: real
    
  e2e:
    database: postgresql_e2e
    redis: redis_e2e
    browser: puppeteer
    
  performance:
    database: postgresql_perf
    redis: redis_perf
    load_generator: artillery
```

### 1.2 環境変数設定
```bash
# .env.test
# テスト環境共通設定
APP_ENV=test
LOG_LEVEL=error

# データベース設定
DB_HOST=localhost
DB_PORT=5432
DB_NAME=access_log_tracker_test
DB_USER=postgres
DB_PASSWORD=password
DB_SSL_MODE=disable
DB_MAX_OPEN_CONNS=10
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=300s

# Redis設定
REDIS_HOST=localhost
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
```

### 1.3 テスト用データベース設定
```sql
-- test-database-setup.sql
-- テスト用データベース作成
CREATE DATABASE access_log_tracker_test;
CREATE DATABASE access_log_tracker_e2e;
CREATE DATABASE access_log_tracker_perf;
CREATE DATABASE access_log_tracker_security;

-- テスト用ユーザー作成
CREATE USER test_user WITH PASSWORD 'test_password';
GRANT ALL PRIVILEGES ON DATABASE access_log_tracker_test TO test_user;
GRANT ALL PRIVILEGES ON DATABASE access_log_tracker_e2e TO test_user;
GRANT ALL PRIVILEGES ON DATABASE access_log_tracker_perf TO test_user;
GRANT ALL PRIVILEGES ON DATABASE access_log_tracker_security TO test_user;
```

## 2. テスト用設定ファイル

### 2.1 Goテスト設定
```go
// tests/config/test_config.go
package config

import (
    "os"
    "time"
)

type TestConfig struct {
    Database DatabaseConfig
    Redis    RedisConfig
    API      APIConfig
    Test     TestSettings
}

type DatabaseConfig struct {
    Host            string
    Port            int
    Name            string
    User            string
    Password        string
    SSLMode         string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
}

type RedisConfig struct {
    Host     string
    Port     int
    Password string
    DB       int
    PoolSize int
}

type APIConfig struct {
    Port         int
    CORSOrigin   string
    RateLimit    RateLimitConfig
}

type RateLimitConfig struct {
    Window time.Duration
    Max    int
}

type TestSettings struct {
    Timeout  time.Duration
    Parallel int
}

func LoadTestConfig() *TestConfig {
    return &TestConfig{
        Database: DatabaseConfig{
            Host:            getEnv("DB_HOST", "localhost"),
            Port:            getEnvAsInt("DB_PORT", 5432),
            Name:            getEnv("DB_NAME", "access_log_tracker_test"),
            User:            getEnv("DB_USER", "postgres"),
            Password:        getEnv("DB_PASSWORD", "password"),
            SSLMode:         getEnv("DB_SSL_MODE", "disable"),
            MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 10),
            MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
            ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 300*time.Second),
        },
        Redis: RedisConfig{
            Host:     getEnv("REDIS_HOST", "localhost"),
            Port:     getEnvAsInt("REDIS_PORT", 6379),
            Password: getEnv("REDIS_PASSWORD", ""),
            DB:       getEnvAsInt("REDIS_DB", 0),
            PoolSize: getEnvAsInt("REDIS_POOL_SIZE", 5),
        },
        API: APIConfig{
            Port:       getEnvAsInt("API_PORT", 3001),
            CORSOrigin: getEnv("CORS_ORIGIN", "http://localhost:3000"),
            RateLimit: RateLimitConfig{
                Window: getEnvAsDuration("RATE_LIMIT_WINDOW", 60*time.Second),
                Max:    getEnvAsInt("RATE_LIMIT_MAX", 1000),
            },
        },
        Test: TestSettings{
            Timeout:  getEnvAsDuration("TEST_TIMEOUT", 30*time.Second),
            Parallel: getEnvAsInt("TEST_PARALLEL", 4),
        },
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
    if value := os.Getenv(key); value != "" {
        if duration, err := time.ParseDuration(value); err == nil {
            return duration
        }
    }
    return defaultValue
}
```

### 2.2 テストセットアップファイル
```go
// tests/setup/setup.go
package setup

import (
    "context"
    "database/sql"
    "log"
    "time"
    
    _ "github.com/lib/pq"
    "github.com/redis/go-redis/v9"
    
    "access-log-tracker/tests/config"
)

type TestSetup struct {
    DB    *sql.DB
    Redis *redis.Client
    Config *config.TestConfig
}

func NewTestSetup() *TestSetup {
    cfg := config.LoadTestConfig()
    
    return &TestSetup{
        Config: cfg,
    }
}

func (ts *TestSetup) SetupDatabase() error {
    dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        ts.Config.Database.Host,
        ts.Config.Database.Port,
        ts.Config.Database.User,
        ts.Config.Database.Password,
        ts.Config.Database.Name,
        ts.Config.Database.SSLMode,
    )
    
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return err
    }
    
    // 接続設定
    db.SetMaxOpenConns(ts.Config.Database.MaxOpenConns)
    db.SetMaxIdleConns(ts.Config.Database.MaxIdleConns)
    db.SetConnMaxLifetime(ts.Config.Database.ConnMaxLifetime)
    
    // 接続テスト
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := db.PingContext(ctx); err != nil {
        return err
    }
    
    ts.DB = db
    return nil
}

func (ts *TestSetup) SetupRedis() error {
    rdb := redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%d", ts.Config.Redis.Host, ts.Config.Redis.Port),
        Password: ts.Config.Redis.Password,
        DB:       ts.Config.Redis.DB,
        PoolSize: ts.Config.Redis.PoolSize,
    })
    
    // 接続テスト
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := rdb.Ping(ctx).Err(); err != nil {
        return err
    }
    
    ts.Redis = rdb
    return nil
}

func (ts *TestSetup) Teardown() error {
    if ts.DB != nil {
        if err := ts.DB.Close(); err != nil {
            log.Printf("Error closing database: %v", err)
        }
    }
    
    if ts.Redis != nil {
        if err := ts.Redis.Close(); err != nil {
            log.Printf("Error closing Redis: %v", err)
        }
    }
    
    return nil
}

func (ts *TestSetup) ClearTestData() error {
    if ts.DB != nil {
        tables := []string{"access_logs", "sessions", "applications", "custom_parameters"}
        for _, table := range tables {
            if _, err := ts.DB.Exec("TRUNCATE TABLE " + table + " CASCADE"); err != nil {
                return err
            }
        }
    }
    
    if ts.Redis != nil {
        ctx := context.Background()
        if err := ts.Redis.FlushDB(ctx).Err(); err != nil {
            return err
        }
    }
    
    return nil
}
```

## 3. テスト用ユーティリティ

### 3.1 データベースユーティリティ
```go
// tests/utils/database.go
package utils

import (
    "context"
    "database/sql"
    "fmt"
    "time"
    
    "access-log-tracker/internal/domain/models"
)

type TestDatabase struct {
    DB *sql.DB
}

func NewTestDatabase(db *sql.DB) *TestDatabase {
    return &TestDatabase{DB: db}
}

func (td *TestDatabase) Setup() error {
    // マイグレーション実行
    if err := td.runMigrations(); err != nil {
        return err
    }
    
    // テスト用データ投入
    if err := td.seedTestData(); err != nil {
        return err
    }
    
    return nil
}

func (td *TestDatabase) Teardown() error {
    // テストデータクリア
    if err := td.clearAllData(); err != nil {
        return err
    }
    
    return nil
}

func (td *TestDatabase) ClearAllData() error {
    tables := []string{"access_logs", "sessions", "applications", "custom_parameters"}
    
    for _, table := range tables {
        if _, err := td.DB.Exec("TRUNCATE TABLE " + table + " CASCADE"); err != nil {
            return err
        }
    }
    
    return nil
}

func (td *TestDatabase) runMigrations() error {
    // マイグレーションファイルを実行
    migrationFiles := []string{
        "001_create_applications_table.sql",
        "002_create_access_logs_table.sql",
        "003_create_sessions_table.sql",
        "004_create_custom_parameters_table.sql",
    }
    
    for _, file := range migrationFiles {
        sql, err := readMigrationFile(file)
        if err != nil {
            return err
        }
        
        if _, err := td.DB.Exec(sql); err != nil {
            return err
        }
    }
    
    return nil
}

func (td *TestDatabase) seedTestData() error {
    // テスト用アプリケーションデータ
    _, err := td.DB.Exec(`
        INSERT INTO applications (app_id, name, description, domain, api_key, created_at, updated_at)
        VALUES 
            ('test_app_123', 'Test Application', 'Test application for unit testing', 'test.example.com', 'test_api_key', NOW(), NOW()),
            ('test_app_456', 'Another Test App', 'Another test application', 'another-test.example.com', 'another_test_api_key', NOW(), NOW())
        ON CONFLICT (app_id) DO NOTHING
    `)
    
    return err
}

func readMigrationFile(filename string) (string, error) {
    // マイグレーションファイルを読み込む実装
    // 実際の実装では、ファイルシステムから読み込むか、
    // 埋め込まれたSQLを使用する
    return "", nil
}
```

### 3.2 Redisユーティリティ
```go
// tests/utils/redis.go
package utils

import (
    "context"
    "encoding/json"
    "fmt"
    
    "github.com/redis/go-redis/v9"
)

type TestRedis struct {
    Client *redis.Client
}

func NewTestRedis(client *redis.Client) *TestRedis {
    return &TestRedis{Client: client}
}

func (tr *TestRedis) Setup() error {
    // Redis接続テスト
    ctx := context.Background()
    if err := tr.Client.Ping(ctx).Err(); err != nil {
        return err
    }
    
    // テスト用データ投入
    if err := tr.seedTestData(); err != nil {
        return err
    }
    
    return nil
}

func (tr *TestRedis) Teardown() error {
    // テストデータクリア
    if err := tr.clearAllData(); err != nil {
        return err
    }
    
    return nil
}

func (tr *TestRedis) ClearAllData() error {
    ctx := context.Background()
    return tr.Client.FlushDB(ctx).Err()
}

func (tr *TestRedis) seedTestData() error {
    ctx := context.Background()
    
    // テスト用キャッシュデータ
    appData := map[string]interface{}{
        "app_id":  "test_app_123",
        "name":    "Test Application",
        "api_key": "test_api_key",
    }
    
    appDataJSON, err := json.Marshal(appData)
    if err != nil {
        return err
    }
    
    return tr.Client.Set(ctx, "test:app:test_app_123", appDataJSON, 0).Err()
}
```

## 4. 環境別テスト設定

### 4.1 単体テスト環境
```go
// tests/unit/config.go
package config

type UnitTestConfig struct {
    Database DatabaseConfig
    Redis    RedisConfig
    Logging  LoggingConfig
}

func LoadUnitTestConfig() *UnitTestConfig {
    return &UnitTestConfig{
        Database: DatabaseConfig{
            Type: "sqlite",
            File: ":memory:",
        },
        Redis: RedisConfig{
            Type:    "mock",
            Enabled: false,
        },
        Logging: LoggingConfig{
            Level:   "error",
            Enabled: false,
        },
    }
}
```

### 4.2 統合テスト環境
```go
// tests/integration/config.go
package config

func LoadIntegrationTestConfig() *TestConfig {
    return &TestConfig{
        Database: DatabaseConfig{
            Type: "postgresql",
            Host: "localhost",
            Port: 5432,
            Name: "access_log_tracker_test",
            User: "postgres",
            Password: "password",
        },
        Redis: RedisConfig{
            Type: "redis",
            Host: "localhost",
            Port: 6379,
            Password: "",
            DB: 0,
        },
        Logging: LoggingConfig{
            Level:   "info",
            Enabled: true,
        },
    }
}
```

### 4.3 E2Eテスト環境
```go
// tests/e2e/config.go
package config

func LoadE2ETestConfig() *TestConfig {
    return &TestConfig{
        Database: DatabaseConfig{
            Type: "postgresql",
            Host: "localhost",
            Port: 5432,
            Name: "access_log_tracker_e2e",
            User: "postgres",
            Password: "password",
        },
        Redis: RedisConfig{
            Type: "redis",
            Host: "localhost",
            Port: 6379,
            Password: "",
            DB: 1,
        },
        Browser: BrowserConfig{
            Type:    "puppeteer",
            Headless: true,
            SlowMo:  100,
        },
        Logging: LoggingConfig{
            Level:   "debug",
            Enabled: true,
        },
    }
}
```

## 5. 環境切り替えスクリプト

### 5.1 環境切り替えユーティリティ
```go
// scripts/test-environment.go
package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "path/filepath"
    
    "access-log-tracker/tests/config"
)

func main() {
    environment := flag.String("env", "unit", "Test environment (unit, integration, e2e, performance)")
    flag.Parse()
    
    environments := []string{"unit", "integration", "e2e", "performance"}
    
    valid := false
    for _, env := range environments {
        if *environment == env {
            valid = true
            break
        }
    }
    
    if !valid {
        log.Fatalf("Invalid environment: %s", *environment)
    }
    
    if err := switchEnvironment(*environment); err != nil {
        log.Fatalf("Failed to switch environment: %v", err)
    }
    
    fmt.Printf("Switched to %s test environment\n", *environment)
}

func switchEnvironment(environment string) error {
    // 環境設定ファイルをコピー
    sourceConfig := filepath.Join("tests", "config", environment+".go")
    targetConfig := filepath.Join("config", "test.go")
    
    // 実際の実装では、設定ファイルのコピーや
    // 環境変数の設定を行う
    fmt.Printf("Copying %s to %s\n", sourceConfig, targetConfig)
    
    return nil
}
```

### 5.2 環境切り替えコマンド
```bash
#!/bin/bash
# scripts/switch-test-env.sh

ENVIRONMENT=$1

if [ -z "$ENVIRONMENT" ]; then
  echo "Usage: $0 <environment>"
  echo "Available environments: unit, integration, e2e, performance"
  exit 1
fi

# Goスクリプトを実行して環境を切り替え
go run scripts/test-environment.go -env $ENVIRONMENT

echo "Test environment switched to: $ENVIRONMENT"
```

## 6. 環境検証

### 6.1 環境検証スクリプト
```go
// scripts/verify-test-environment.go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "access-log-tracker/tests/setup"
)

func main() {
    fmt.Println("Verifying test environment...")
    
    testSetup := setup.NewTestSetup()
    
    // データベース接続テスト
    if err := testSetup.SetupDatabase(); err != nil {
        log.Fatalf("❌ Database connection failed: %v", err)
    }
    fmt.Println("✅ Database connection: OK")
    
    // Redis接続テスト
    if err := testSetup.SetupRedis(); err != nil {
        log.Fatalf("❌ Redis connection failed: %v", err)
    }
    fmt.Println("✅ Redis connection: OK")
    
    // クリーンアップ
    if err := testSetup.Teardown(); err != nil {
        log.Printf("Warning: Teardown failed: %v", err)
    }
    
    fmt.Println("✅ Test environment verification completed successfully")
}
```

### 6.2 環境検証コマンド
```bash
#!/bin/bash
# scripts/verify-env.sh

echo "Verifying test environment..."

# データベース接続確認
echo "Checking database connection..."
go run scripts/verify-test-environment.go

if [ $? -eq 0 ]; then
  echo "✅ Test environment is ready"
else
  echo "❌ Test environment verification failed"
  exit 1
fi
```
