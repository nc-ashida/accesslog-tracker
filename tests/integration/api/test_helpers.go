package api

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"accesslog-tracker/internal/domain/models"
)

// GetTestDBHost はテスト用データベースのホストを取得します
func GetTestDBHost() string {
	if host := os.Getenv("DB_HOST"); host != "" {
		return host
	}
	if host := os.Getenv("TEST_DB_HOST"); host != "" {
		return host
	}
	return "localhost"
}

// GetTestRedisHost はテスト用Redisのホストを取得します
func GetTestRedisHost() string {
	if host := os.Getenv("REDIS_HOST"); host != "" {
		return host
	}
	if host := os.Getenv("TEST_REDIS_HOST"); host != "" {
		return host
	}
	return "localhost"
}

// SetupTestDatabase はテスト用データベースをセットアップします
func SetupTestDatabase(t *testing.T) *sql.DB {
	host := GetTestDBHost()
	dsn := "host=" + host + " port=5432 user=postgres password=password dbname=access_log_tracker_test sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	
	// 接続テスト
	err = db.Ping()
	require.NoError(t, err)
	
	return db
}

// SetupTestRedis はテスト用Redisをセットアップします
func SetupTestRedis(t *testing.T) *redis.Client {
	host := GetTestRedisHost()
	client := redis.NewClient(&redis.Options{
		Addr:     host + ":6379",
		Password: "",
		DB:       0,
	})
	
	// 接続テスト
	ctx := context.Background()
	err := client.Ping(ctx).Err()
	require.NoError(t, err)
	
	return client
}

// CleanupTestData はテストデータをクリーンアップします
func CleanupTestData(t *testing.T, db *sql.DB) {
	tables := []string{"custom_parameters", "access_logs", "sessions", "applications"}
	
	for _, table := range tables {
		_, err := db.Exec("TRUNCATE TABLE " + table + " CASCADE")
		require.NoError(t, err)
	}
}

// CreateTestApplication はテスト用アプリケーションを作成します
func CreateTestApplication(t *testing.T, db *sql.DB) *models.Application {
	// より一意なIDを生成
	timestamp := time.Now().UnixNano()
	randomSuffix := RandomString(8)
	
	app := &models.Application{
		AppID:       fmt.Sprintf("test_app_%d_%s", timestamp, randomSuffix),
		Name:        "Test Application",
		Description: "Test application for integration testing",
		Domain:      "test.example.com",
		APIKey:      fmt.Sprintf("alt_test_api_key_%d_%s", timestamp, randomSuffix),
		Active:      true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	_, err := db.Exec(`
		INSERT INTO applications (app_id, name, description, domain, api_key, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (app_id) DO NOTHING
	`, app.AppID, app.Name, app.Description, app.Domain, app.APIKey, app.Active, app.CreatedAt, app.UpdatedAt)
	require.NoError(t, err)
	
	return app
}

// CreateTestTrackingData はテスト用トラッキングデータを作成します
func CreateTestTrackingData(t *testing.T, db *sql.DB, appID string) *models.TrackingData {
	trackingData := &models.TrackingData{
		ID:        "alt_" + time.Now().Format("20060102150405") + "_" + RandomString(9),
		AppID:     appID,
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		URL:       "https://example.com/test",
		IPAddress: "192.168.1.100",
		SessionID: "alt_" + time.Now().Format("20060102150405") + "_" + RandomString(9),
		Timestamp: time.Now(),
	}
	
	_, err := db.Exec(`
		INSERT INTO access_logs (id, app_id, user_agent, url, ip_address, session_id, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO NOTHING
	`, trackingData.ID, trackingData.AppID, trackingData.UserAgent, 
		trackingData.URL, trackingData.IPAddress, trackingData.SessionID, trackingData.Timestamp)
	require.NoError(t, err)
	
	return trackingData
}

// CreateTestCustomParameter はテスト用カスタムパラメータを作成します
func CreateTestCustomParameter(t *testing.T, db *sql.DB, trackingID string, key, value string) {
	_, err := db.Exec(`
		INSERT INTO custom_parameters (tracking_id, param_key, param_value)
		VALUES ($1, $2, $3)
		ON CONFLICT (tracking_id, param_key) DO NOTHING
	`, trackingID, key, value)
	require.NoError(t, err)
}

// GetTestApplicationByID は指定されたIDのテストアプリケーションを取得します
func GetTestApplicationByID(t *testing.T, db *sql.DB, appID string) *models.Application {
	var app models.Application
	err := db.QueryRow(`
		SELECT app_id, name, description, domain, api_key, active, created_at, updated_at
		FROM applications WHERE app_id = $1
	`, appID).Scan(&app.AppID, &app.Name, &app.Description, &app.Domain, &app.APIKey, &app.Active, &app.CreatedAt, &app.UpdatedAt)
	
	if err != nil {
		return nil
	}
	
	return &app
}

// GetTestTrackingDataByID は指定されたIDのテストトラッキングデータを取得します
func GetTestTrackingDataByID(t *testing.T, db *sql.DB, trackingID string) *models.TrackingData {
	var tracking models.TrackingData
	err := db.QueryRow(`
		SELECT id, app_id, user_agent, url, ip_address, session_id, timestamp
		FROM access_logs WHERE id = $1
	`, trackingID).Scan(&tracking.ID, &tracking.AppID, &tracking.UserAgent, &tracking.URL, &tracking.IPAddress, &tracking.SessionID, &tracking.Timestamp)
	
	if err != nil {
		return nil
	}
	
	return &tracking
}

// CountTestApplications はテストアプリケーションの数を取得します
func CountTestApplications(t *testing.T, db *sql.DB) int {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM applications").Scan(&count)
	require.NoError(t, err)
	return count
}

// CountTestTrackingData はテストトラッキングデータの数を取得します
func CountTestTrackingData(t *testing.T, db *sql.DB) int {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM access_logs").Scan(&count)
	require.NoError(t, err)
	return count
}

// CountTestSessions はテストセッションの数を取得します
func CountTestSessions(t *testing.T, db *sql.DB) int {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&count)
	require.NoError(t, err)
	return count
}

// RandomString はランダム文字列を生成します
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// SetupTestEnvironment はテスト環境全体をセットアップします
func SetupTestEnvironment(t *testing.T) (*sql.DB, *redis.Client) {
	// データベースをセットアップ
	db := SetupTestDatabase(t)
	
	// Redisをセットアップ
	redisClient := SetupTestRedis(t)
	
	// テストデータをクリーンアップ
	CleanupTestData(t, db)
	
	return db, redisClient
}

// TeardownTestEnvironment はテスト環境をクリーンアップします
func TeardownTestEnvironment(t *testing.T, db *sql.DB, redisClient *redis.Client) {
	// テストデータをクリーンアップ
	CleanupTestData(t, db)
	
	// 接続を閉じる
	if db != nil {
		db.Close()
	}
	if redisClient != nil {
		redisClient.Close()
	}
}
