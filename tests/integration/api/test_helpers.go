package api

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"accesslog-tracker/internal/domain/models"
)

// getTestDBHost はテスト用データベースのホストを取得します
func getTestDBHost() string {
	if host := os.Getenv("TEST_DB_HOST"); host != "" {
		return host
	}
	return "localhost"
}

// SetupTestDatabase はテスト用データベースをセットアップします
func SetupTestDatabase(t *testing.T) *sql.DB {
	host := getTestDBHost()
	dsn := "host=" + host + " port=5432 user=postgres password=password dbname=access_log_tracker_test sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	
	// 接続テスト
	err = db.Ping()
	require.NoError(t, err)
	
	return db
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
	app := &models.Application{
		AppID:       "test_app_" + time.Now().Format("20060102150405") + "_" + randomString(5),
		Name:        "Test Application",
		Description: "Test application for integration testing",
		Domain:      "test.example.com",
		APIKey:      "alt_test_api_key_" + time.Now().Format("20060102150405") + "_" + randomString(5),
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
		ID:        "alt_" + time.Now().Format("20060102150405") + "_" + randomString(9),
		AppID:     appID,
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		URL:       "https://example.com/test",
		IPAddress: "192.168.1.100",
		SessionID: "alt_" + time.Now().Format("20060102150405") + "_" + randomString(9),
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

// randomString はランダム文字列を生成します
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
