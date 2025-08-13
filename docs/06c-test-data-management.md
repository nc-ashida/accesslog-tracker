# テストデータ管理

## 1. テストデータ生成

### 1.1 テストデータジェネレーター
```go
// tests/utils/test_data_generator.go
package utils

import (
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "math/big"
    "net"
    "time"
    
    "access-log-tracker/internal/domain/models"
)

type TestDataGenerator struct{}

func NewTestDataGenerator() *TestDataGenerator {
    return &TestDataGenerator{}
}

func (g *TestDataGenerator) GenerateTrackingData(count int) []*models.TrackingData {
    data := make([]*models.TrackingData, count)
    
    for i := 0; i < count; i++ {
        data[i] = &models.TrackingData{
            AppID:       fmt.Sprintf("test_app_%s", g.randomString(9)),
            ClientSubID: fmt.Sprintf("sub_%s", g.randomString(9)),
            ModuleID:    fmt.Sprintf("module_%s", g.randomString(9)),
            URL:         fmt.Sprintf("https://example.com/page/%d", i),
            UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
            IPAddress:   g.randomIPAddress(),
            SessionID:   fmt.Sprintf("alt_%d_%s", time.Now().Unix(), g.randomString(9)),
            Timestamp:   time.Now(),
        }
    }
    
    return data
}

func (g *TestDataGenerator) GenerateApplicationData() *models.Application {
    return &models.Application{
        Name:        fmt.Sprintf("Test App %d", time.Now().Unix()),
        Description: "Test application for unit testing",
        Domain:      "test.example.com",
        APIKey:      fmt.Sprintf("test_key_%s", g.randomString(20)),
    }
}

func (g *TestDataGenerator) GenerateSessionData() *models.Session {
    return &models.Session{
        SessionID:    fmt.Sprintf("alt_%d_%s", time.Now().Unix(), g.randomString(9)),
        AppID:        fmt.Sprintf("test_app_%s", g.randomString(9)),
        UserAgent:    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
        IPAddress:    g.randomIPAddress(),
        StartedAt:    time.Now(),
        LastActivity: time.Now(),
        IsActive:     true,
    }
}

func (g *TestDataGenerator) GenerateCustomParameters() map[string]interface{} {
    sources := []string{"google", "facebook", "twitter", "direct"}
    mediums := []string{"cpc", "social", "email", "organic"}
    
    sourceIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(sources))))
    mediumIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(mediums))))
    
    return map[string]interface{}{
        "campaign_id": fmt.Sprintf("camp_%s", g.randomString(9)),
        "source":      sources[sourceIndex.Int64()],
        "medium":      mediums[mediumIndex.Int64()],
        "term":        fmt.Sprintf("keyword_%s", g.randomString(9)),
        "content":     fmt.Sprintf("content_%s", g.randomString(9)),
    }
}

func (g *TestDataGenerator) randomString(length int) string {
    bytes := make([]byte, length/2)
    rand.Read(bytes)
    return hex.EncodeToString(bytes)[:length]
}

func (g *TestDataGenerator) randomIPAddress() string {
    // 192.168.1.x の範囲でランダムIPを生成
    thirdOctet := 1
    fourthOctet, _ := rand.Int(rand.Reader, big.NewInt(255))
    return fmt.Sprintf("192.168.%d.%d", thirdOctet, fourthOctet.Int64())
}
```

### 1.2 テストデータファクトリー
```go
// tests/factories/tracking_data_factory.go
package factories

import (
    "time"
    
    "access-log-tracker/internal/domain/models"
)

type TrackingDataFactory struct{}

func NewTrackingDataFactory() *TrackingDataFactory {
    return &TrackingDataFactory{}
}

func (f *TrackingDataFactory) CreateValidTrackingData(overrides map[string]interface{}) *models.TrackingData {
    data := &models.TrackingData{
        AppID:       fmt.Sprintf("test_app_%d", time.Now().Unix()),
        UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
        URL:         "https://example.com/test-page",
        SessionID:   fmt.Sprintf("alt_%d_%s", time.Now().Unix(), randomString(9)),
        IPAddress:   "192.168.1.100",
        Timestamp:   time.Now(),
    }
    
    // オーバーライドの適用
    for key, value := range overrides {
        switch key {
        case "app_id":
            if str, ok := value.(string); ok {
                data.AppID = str
            }
        case "user_agent":
            if str, ok := value.(string); ok {
                data.UserAgent = str
            }
        case "url":
            if str, ok := value.(string); ok {
                data.URL = str
            }
        }
    }
    
    return data
}

func (f *TrackingDataFactory) CreateInvalidTrackingData() *models.TrackingData {
    return &models.TrackingData{
        // AppIDが欠けている
        UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
        URL:       "invalid-url", // 無効なURL
        SessionID: "invalid_session_id",
        Timestamp: time.Now(),
    }
}

func (f *TrackingDataFactory) CreateCrawlerTrackingData() *models.TrackingData {
    return &models.TrackingData{
        AppID:       fmt.Sprintf("test_app_%d", time.Now().Unix()),
        UserAgent:   "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
        URL:         "https://example.com/test-page",
        SessionID:   fmt.Sprintf("alt_%d_%s", time.Now().Unix(), randomString(9)),
        IPAddress:   "192.168.1.100",
        Timestamp:   time.Now(),
    }
}

func (f *TrackingDataFactory) CreateMobileTrackingData() *models.TrackingData {
    return &models.TrackingData{
        AppID:       fmt.Sprintf("test_app_%d", time.Now().Unix()),
        UserAgent:   "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15",
        URL:         "https://example.com/test-page",
        SessionID:   fmt.Sprintf("alt_%d_%s", time.Now().Unix(), randomString(9)),
        IPAddress:   "192.168.1.100",
        Timestamp:   time.Now(),
    }
}

func randomString(length int) string {
    const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
    b := make([]byte, length)
    for i := range b {
        b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
    }
    return string(b)
}
```

### 1.3 テストデータセット
```go
// tests/datasets/tracking_datasets.go
package datasets

import "access-log-tracker/internal/domain/models"

var TrackingDatasets = struct {
    ValidTrackingData     []*models.TrackingData
    InvalidTrackingData   []*models.TrackingData
    CrawlerTrackingData   []*models.TrackingData
    MobileTrackingData    []*models.TrackingData
}{
    ValidTrackingData: []*models.TrackingData{
        {
            AppID:       "test_app_123",
            UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
            URL:         "https://example.com/page1",
            SessionID:   "alt_1234567890_abc123",
            IPAddress:   "192.168.1.100",
            Timestamp:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
        },
        {
            AppID:       "test_app_456",
            UserAgent:   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
            URL:         "https://example.com/page2",
            SessionID:   "alt_1234567890_def456",
            IPAddress:   "192.168.1.101",
            Timestamp:   time.Date(2024, 1, 15, 11, 30, 0, 0, time.UTC),
        },
    },
    
    InvalidTrackingData: []*models.TrackingData{
        {
            // AppIDが欠けている
            UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
            URL:       "https://example.com/page1",
            SessionID: "alt_1234567890_abc123",
            Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
        },
        {
            AppID:       "test_app_123",
            UserAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
            URL:         "invalid-url", // 無効なURL
            SessionID:   "alt_1234567890_abc123",
            Timestamp:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
        },
    },
    
    CrawlerTrackingData: []*models.TrackingData{
        {
            AppID:       "test_app_123",
            UserAgent:   "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
            URL:         "https://example.com/page1",
            SessionID:   "alt_1234567890_abc123",
            IPAddress:   "192.168.1.100",
            Timestamp:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
        },
        {
            AppID:       "test_app_456",
            UserAgent:   "Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)",
            URL:         "https://example.com/page2",
            SessionID:   "alt_1234567890_def456",
            IPAddress:   "192.168.1.101",
            Timestamp:   time.Date(2024, 1, 15, 11, 30, 0, 0, time.UTC),
        },
    },
    
    MobileTrackingData: []*models.TrackingData{
        {
            AppID:       "test_app_123",
            UserAgent:   "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15",
            URL:         "https://example.com/page1",
            SessionID:   "alt_1234567890_abc123",
            IPAddress:   "192.168.1.100",
            Timestamp:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
        },
        {
            AppID:       "test_app_456",
            UserAgent:   "Mozilla/5.0 (Linux; Android 10; SM-G973F) AppleWebKit/537.36",
            URL:         "https://example.com/page2",
            SessionID:   "alt_1234567890_def456",
            IPAddress:   "192.168.1.101",
            Timestamp:   time.Date(2024, 1, 15, 11, 30, 0, 0, time.UTC),
        },
    },
}
```

## 2. テストデータベース管理

### 2.1 テストデータベースセットアップ
```go
// tests/utils/database_setup.go
package utils

import (
    "database/sql"
    "fmt"
    "time"
    
    _ "github.com/lib/pq"
    "access-log-tracker/internal/domain/models"
)

type TestDatabaseSetup struct {
    DB *sql.DB
}

func NewTestDatabaseSetup(db *sql.DB) *TestDatabaseSetup {
    return &TestDatabaseSetup{DB: db}
}

func (tds *TestDatabaseSetup) Setup() error {
    if err := tds.createTables(); err != nil {
        return err
    }
    
    if err := tds.insertSeedData(); err != nil {
        return err
    }
    
    return nil
}

func (tds *TestDatabaseSetup) Teardown() error {
    if err := tds.clearAllData(); err != nil {
        return err
    }
    
    return tds.DB.Close()
}

func (tds *TestDatabaseSetup) createTables() error {
    createApplicationsTable := `
        CREATE TABLE IF NOT EXISTS applications (
            id SERIAL PRIMARY KEY,
            app_id VARCHAR(255) UNIQUE NOT NULL,
            name VARCHAR(255) NOT NULL,
            description TEXT,
            domain VARCHAR(255),
            api_key VARCHAR(255) UNIQUE NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    `

    createAccessLogsTable := `
        CREATE TABLE IF NOT EXISTS access_logs (
            id SERIAL PRIMARY KEY,
            tracking_id VARCHAR(255) UNIQUE NOT NULL,
            app_id VARCHAR(255) NOT NULL,
            user_agent TEXT,
            url TEXT,
            ip_address INET,
            session_id VARCHAR(255),
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (app_id) REFERENCES applications(app_id)
        );
    `

    createSessionsTable := `
        CREATE TABLE IF NOT EXISTS sessions (
            id SERIAL PRIMARY KEY,
            session_id VARCHAR(255) UNIQUE NOT NULL,
            app_id VARCHAR(255) NOT NULL,
            user_agent TEXT,
            ip_address INET,
            started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            is_active BOOLEAN DEFAULT TRUE,
            FOREIGN KEY (app_id) REFERENCES applications(app_id)
        );
    `

    if _, err := tds.DB.Exec(createApplicationsTable); err != nil {
        return err
    }
    if _, err := tds.DB.Exec(createAccessLogsTable); err != nil {
        return err
    }
    if _, err := tds.DB.Exec(createSessionsTable); err != nil {
        return err
    }
    
    return nil
}

func (tds *TestDatabaseSetup) insertSeedData() error {
    // テスト用アプリケーションデータ
    _, err := tds.DB.Exec(`
        INSERT INTO applications (app_id, name, description, domain, api_key)
        VALUES 
            ('test_app_123', 'Test Application', 'Test application for unit testing', 'test.example.com', 'test_api_key'),
            ('test_app_456', 'Another Test App', 'Another test application', 'another-test.example.com', 'another_test_api_key')
        ON CONFLICT (app_id) DO NOTHING;
    `)
    if err != nil {
        return err
    }

    // テスト用セッションデータ
    _, err = tds.DB.Exec(`
        INSERT INTO sessions (session_id, app_id, user_agent, ip_address)
        VALUES 
            ('alt_1234567890_abc123', 'test_app_123', 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36', '192.168.1.100'),
            ('alt_1234567890_def456', 'test_app_456', 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36', '192.168.1.101')
        ON CONFLICT (session_id) DO NOTHING;
    `)
    
    return err
}

func (tds *TestDatabaseSetup) clearAllData() error {
    tables := []string{"access_logs", "sessions", "applications"}
    
    for _, table := range tables {
        if _, err := tds.DB.Exec("TRUNCATE TABLE " + table + " CASCADE"); err != nil {
            return err
        }
    }
    
    return nil
}
```

### 2.2 テストデータクリーンアップ
```go
// tests/utils/data_cleanup.go
package utils

import (
    "context"
    "database/sql"
    
    "github.com/redis/go-redis/v9"
)

type DataCleanup struct{}

func NewDataCleanup() *DataCleanup {
    return &DataCleanup{}
}

func (dc *DataCleanup) CleanupTrackingData(db *sql.DB) error {
    _, err := db.Exec("DELETE FROM access_logs WHERE app_id LIKE 'test_app_%'")
    return err
}

func (dc *DataCleanup) CleanupSessionData(db *sql.DB) error {
    _, err := db.Exec("DELETE FROM sessions WHERE app_id LIKE 'test_app_%'")
    return err
}

func (dc *DataCleanup) CleanupApplicationData(db *sql.DB) error {
    _, err := db.Exec("DELETE FROM applications WHERE app_id LIKE 'test_app_%'")
    return err
}

func (dc *DataCleanup) CleanupAllTestData(db *sql.DB) error {
    if err := dc.CleanupTrackingData(db); err != nil {
        return err
    }
    if err := dc.CleanupSessionData(db); err != nil {
        return err
    }
    if err := dc.CleanupApplicationData(db); err != nil {
        return err
    }
    return nil
}

func (dc *DataCleanup) CleanupRedisData(redisClient *redis.Client) error {
    ctx := context.Background()
    keys, err := redisClient.Keys(ctx, "test:*").Result()
    if err != nil {
        return err
    }
    
    if len(keys) > 0 {
        _, err = redisClient.Del(ctx, keys...).Result()
        return err
    }
    
    return nil
}
```

## 3. テストデータ検証

### 3.1 データ検証ユーティリティ
```go
// tests/utils/data_validator.go
package utils

import (
    "net"
    "net/url"
    "regexp"
    "strings"
    
    "access-log-tracker/internal/domain/models"
)

type DataValidator struct{}

func NewDataValidator() *DataValidator {
    return &DataValidator{}
}

type ValidationResult struct {
    IsValid bool
    Errors  []string
}

func (dv *DataValidator) ValidateTrackingData(data *models.TrackingData) ValidationResult {
    errors := []string{}

    if data.AppID == "" {
        errors = append(errors, "app_id is required")
    }

    if data.UserAgent == "" {
        errors = append(errors, "user_agent is required")
    }

    if data.URL == "" {
        errors = append(errors, "url is required")
    } else if !dv.isValidURL(data.URL) {
        errors = append(errors, "Invalid URL format")
    }

    if data.IPAddress != "" && !dv.isValidIPAddress(data.IPAddress) {
        errors = append(errors, "Invalid IP address format")
    }

    return ValidationResult{
        IsValid: len(errors) == 0,
        Errors:  errors,
    }
}

func (dv *DataValidator) ValidateApplicationData(data *models.Application) ValidationResult {
    errors := []string{}

    if data.Name == "" {
        errors = append(errors, "name is required")
    }

    if data.APIKey == "" {
        errors = append(errors, "api_key is required")
    }

    if data.Domain != "" && !dv.isValidDomain(data.Domain) {
        errors = append(errors, "Invalid domain format")
    }

    return ValidationResult{
        IsValid: len(errors) == 0,
        Errors:  errors,
    }
}

func (dv *DataValidator) isValidURL(urlStr string) bool {
    _, err := url.ParseRequestURI(urlStr)
    return err == nil
}

func (dv *DataValidator) isValidIPAddress(ip string) bool {
    return net.ParseIP(ip) != nil
}

func (dv *DataValidator) isValidDomain(domain string) bool {
    domainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
    return domainRegex.MatchString(domain)
}
```

### 3.2 データ比較ユーティリティ
```go
// tests/utils/data_comparator.go
package utils

import (
    "reflect"
    
    "access-log-tracker/internal/domain/models"
)

type DataComparator struct{}

func NewDataComparator() *DataComparator {
    return &DataComparator{}
}

type ComparisonResult struct {
    IsEqual     bool
    Differences []string
}

func (dc *DataComparator) CompareTrackingData(expected, actual *models.TrackingData) ComparisonResult {
    differences := []string{}

    if expected.AppID != actual.AppID {
        differences = append(differences, fmt.Sprintf("app_id: expected %s, got %s", expected.AppID, actual.AppID))
    }

    if expected.UserAgent != actual.UserAgent {
        differences = append(differences, fmt.Sprintf("user_agent: expected %s, got %s", expected.UserAgent, actual.UserAgent))
    }

    if expected.URL != actual.URL {
        differences = append(differences, fmt.Sprintf("url: expected %s, got %s", expected.URL, actual.URL))
    }

    if expected.IPAddress != actual.IPAddress {
        differences = append(differences, fmt.Sprintf("ip_address: expected %s, got %s", expected.IPAddress, actual.IPAddress))
    }

    if expected.SessionID != actual.SessionID {
        differences = append(differences, fmt.Sprintf("session_id: expected %s, got %s", expected.SessionID, actual.SessionID))
    }

    return ComparisonResult{
        IsEqual:     len(differences) == 0,
        Differences: differences,
    }
}

func (dc *DataComparator) CompareApplicationData(expected, actual *models.Application) ComparisonResult {
    differences := []string{}

    if expected.AppID != actual.AppID {
        differences = append(differences, fmt.Sprintf("app_id: expected %s, got %s", expected.AppID, actual.AppID))
    }

    if expected.Name != actual.Name {
        differences = append(differences, fmt.Sprintf("name: expected %s, got %s", expected.Name, actual.Name))
    }

    if expected.Description != actual.Description {
        differences = append(differences, fmt.Sprintf("description: expected %s, got %s", expected.Description, actual.Description))
    }

    if expected.Domain != actual.Domain {
        differences = append(differences, fmt.Sprintf("domain: expected %s, got %s", expected.Domain, actual.Domain))
    }

    if expected.APIKey != actual.APIKey {
        differences = append(differences, fmt.Sprintf("api_key: expected %s, got %s", expected.APIKey, actual.APIKey))
    }

    return ComparisonResult{
        IsEqual:     len(differences) == 0,
        Differences: differences,
    }
}

func (dc *DataComparator) CompareArrays(expected, actual interface{}, comparator func(interface{}, interface{}) ComparisonResult) ComparisonResult {
    expectedVal := reflect.ValueOf(expected)
    actualVal := reflect.ValueOf(actual)

    if expectedVal.Len() != actualVal.Len() {
        return ComparisonResult{
            IsEqual: false,
            Differences: []string{fmt.Sprintf("Array length mismatch: expected %d, got %d", expectedVal.Len(), actualVal.Len())},
        }
    }

    differences := []string{}
    for i := 0; i < expectedVal.Len(); i++ {
        comparison := comparator(expectedVal.Index(i).Interface(), actualVal.Index(i).Interface())
        if !comparison.IsEqual {
            differences = append(differences, fmt.Sprintf("Index %d: %v", i, comparison.Differences))
        }
    }

    return ComparisonResult{
        IsEqual:     len(differences) == 0,
        Differences: differences,
    }
}
```

## 4. テストデータ管理スクリプト

### 4.1 データ生成スクリプト
```go
// scripts/generate_test_data.go
package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "os"
    "path/filepath"
    
    "access-log-tracker/tests/utils"
)

func main() {
    count := flag.Int("count", 1000, "Number of tracking data records to generate")
    outputDir := flag.String("output", "tests/data", "Output directory for test data")
    flag.Parse()

    fmt.Println("Generating test data...")

    // 出力ディレクトリを作成
    if err := os.MkdirAll(*outputDir, 0755); err != nil {
        log.Fatalf("Failed to create output directory: %v", err)
    }

    generator := utils.NewTestDataGenerator()

    // トラッキングデータを生成
    trackingData := generator.GenerateTrackingData(*count)
    if err := saveToFile(filepath.Join(*outputDir, "tracking-data.json"), trackingData); err != nil {
        log.Fatalf("Failed to save tracking data: %v", err)
    }
    fmt.Printf("Generated tracking-data.json with %d records\n", len(trackingData))

    // アプリケーションデータを生成
    applicationData := make([]*models.Application, 10)
    for i := 0; i < 10; i++ {
        applicationData[i] = generator.GenerateApplicationData()
    }
    if err := saveToFile(filepath.Join(*outputDir, "application-data.json"), applicationData); err != nil {
        log.Fatalf("Failed to save application data: %v", err)
    }
    fmt.Printf("Generated application-data.json with %d records\n", len(applicationData))

    // セッションデータを生成
    sessionData := make([]*models.Session, 100)
    for i := 0; i < 100; i++ {
        sessionData[i] = generator.GenerateSessionData()
    }
    if err := saveToFile(filepath.Join(*outputDir, "session-data.json"), sessionData); err != nil {
        log.Fatalf("Failed to save session data: %v", err)
    }
    fmt.Printf("Generated session-data.json with %d records\n", len(sessionData))

    fmt.Println("Test data generation completed")
}

func saveToFile(filename string, data interface{}) error {
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    return encoder.Encode(data)
}
```

### 4.2 データクリーンアップスクリプト
```go
// scripts/cleanup_test_data.go
package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    
    "database/sql"
    _ "github.com/lib/pq"
    "github.com/redis/go-redis/v9"
    
    "access-log-tracker/tests/utils"
)

func main() {
    dbHost := flag.String("db-host", "localhost", "Database host")
    dbPort := flag.Int("db-port", 5432, "Database port")
    dbName := flag.String("db-name", "access_log_tracker_test", "Database name")
    dbUser := flag.String("db-user", "postgres", "Database user")
    dbPassword := flag.String("db-password", "password", "Database password")
    
    redisHost := flag.String("redis-host", "localhost", "Redis host")
    redisPort := flag.Int("redis-port", 6379, "Redis port")
    redisPassword := flag.String("redis-password", "", "Redis password")
    redisDB := flag.Int("redis-db", 0, "Redis database")
    
    flag.Parse()

    fmt.Println("Cleaning up test data...")

    // データベース接続
    dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
        *dbHost, *dbPort, *dbUser, *dbPassword, *dbName)
    
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()

    // Redis接続
    rdb := redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%d", *redisHost, *redisPort),
        Password: *redisPassword,
        DB:       *redisDB,
    })
    defer rdb.Close()

    cleanup := utils.NewDataCleanup()

    try {
        // データベースのテストデータをクリア
        if err := cleanup.CleanupAllTestData(db); err != nil {
            log.Fatalf("Failed to cleanup database test data: %v", err)
        }
        fmt.Println("✅ Database test data cleaned up")

        // Redisのテストデータをクリア
        if err := cleanup.CleanupRedisData(rdb); err != nil {
            log.Fatalf("Failed to cleanup Redis test data: %v", err)
        }
        fmt.Println("✅ Redis test data cleaned up")

        fmt.Println("Test data cleanup completed successfully")
    } catch (err error) {
        log.Fatalf("❌ Test data cleanup failed: %v", err)
    }
}
```

## 5. テストデータの使用例

### 5.1 単体テストでの使用
```go
// tests/unit/tracking_service_test.go
package unit_test

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    
    "access-log-tracker/tests/factories"
    "access-log-tracker/tests/utils"
)

func TestTrackingService(t *testing.T) {
    factory := factories.NewTrackingDataFactory()
    validator := utils.NewDataValidator()

    t.Run("should process valid tracking data", func(t *testing.T) {
        trackingData := factory.CreateValidTrackingData(nil)
        validation := validator.ValidateTrackingData(trackingData)
        
        assert.True(t, validation.IsValid)
        assert.Empty(t, validation.Errors)
    })

    t.Run("should reject invalid tracking data", func(t *testing.T) {
        trackingData := factory.CreateInvalidTrackingData()
        validation := validator.ValidateTrackingData(trackingData)
        
        assert.False(t, validation.IsValid)
        assert.NotEmpty(t, validation.Errors)
    })
}
```

### 5.2 統合テストでの使用
```go
// tests/integration/tracking_api_test.go
package integration_test

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    
    "access-log-tracker/tests/datasets"
    "access-log-tracker/tests/utils"
)

func TestTrackingAPI(t *testing.T) {
    comparator := utils.NewDataComparator()

    t.Run("should accept valid tracking data", func(t *testing.T) {
        testData := datasets.TrackingDatasets.ValidTrackingData[0]
        
        // APIリクエストを送信
        response := sendTrackingRequest(t, testData)
        
        // レスポンスを検証
        assert.Equal(t, 200, response.StatusCode)
        
        // データを比較
        comparison := comparator.CompareTrackingData(testData, response.Data)
        assert.True(t, comparison.IsEqual)
    })
}

func sendTrackingRequest(t *testing.T, data *models.TrackingData) *TrackingResponse {
    // APIリクエスト送信の実装
    return &TrackingResponse{}
}
```
