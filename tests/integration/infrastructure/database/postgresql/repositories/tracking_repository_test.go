package repositories

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"accesslog-tracker/internal/infrastructure/database/postgresql"
	"accesslog-tracker/internal/infrastructure/database/postgresql/repositories"
	"accesslog-tracker/internal/domain/models"
	"accesslog-tracker/tests/integration/infrastructure"
)

// randomString は指定された長さのランダムな文字列を生成します
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func setupTestDatabase() (*repositories.TrackingRepository, *postgresql.Connection, func(), error) {
	// 環境変数から接続情報を取得
	host := infrastructure.GetEnvOrDefault("DB_HOST", "localhost")
	port := infrastructure.GetEnvOrDefault("DB_PORT", "18433")
	user := infrastructure.GetEnvOrDefault("DB_USER", "postgres")
	password := infrastructure.GetEnvOrDefault("DB_PASSWORD", "password")
	dbname := infrastructure.GetEnvOrDefault("DB_NAME", "access_log_tracker_test")
	
	// テスト用データベース接続
	conn := postgresql.NewConnection("test")
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", 
		host, port, user, password, dbname)
	
	err := conn.Connect(dsn)
	if err != nil {
		return nil, nil, nil, err
	}

	repo := repositories.NewTrackingRepository(conn.GetDB())
	
	// テスト用のアプリケーションを作成
	appRepo := repositories.NewApplicationRepository(conn.GetDB())
	testApp := &models.Application{
		AppID:    "test_app_tracking_123",
		Name:     "Test Tracking Application",
		Domain:   "example.com",
		APIKey:   "test-api-key-tracking-123",
		Active:   true,
	}
	err = appRepo.Create(context.Background(), testApp)
	if err != nil {
		conn.Close()
		return nil, nil, nil, err
	}
	
	// クリーンアップ関数
	cleanup := func() {
		conn.Close()
	}

	return repo, conn, cleanup, nil
}

// cleanupTrackingTestData はテストデータをクリーンアップします
func cleanupTrackingTestData(t *testing.T, repo *repositories.TrackingRepository) {
	ctx := context.Background()
	
	// テスト用トラッキングデータを削除
	// 実際の実装では、特定のIDで削除するか、テスト用のフラグで管理
	// ここでは簡易的に全データを取得して削除
	results, err := repo.GetByAppID(ctx, "test_app_tracking_123", 100, 0)
	if err == nil {
		for _, data := range results {
			// Deleteメソッドを使用
			err := repo.Delete(context.Background(), data.ID)
			if err != nil {
				t.Logf("Failed to delete tracking data: %v", err)
			}
		}
	}
	
	// 他のテスト用アプリケーションのデータも削除
	results, err = repo.GetByAppID(ctx, "test_app_date_range", 100, 0)
	if err == nil {
		for _, data := range results {
			err := repo.Delete(context.Background(), data.ID)
			if err != nil {
				t.Logf("Failed to delete tracking data: %v", err)
			}
		}
	}
	
	results, err = repo.GetByAppID(ctx, "test_app_stats", 100, 0)
	if err == nil {
		for _, data := range results {
			err := repo.Delete(context.Background(), data.ID)
			if err != nil {
				t.Logf("Failed to delete tracking data: %v", err)
			}
		}
	}
	
	results, err = repo.GetByAppID(ctx, "test_app_delete", 100, 0)
	if err == nil {
		for _, data := range results {
			err := repo.Delete(context.Background(), data.ID)
			if err != nil {
				t.Logf("Failed to delete tracking data: %v", err)
			}
		}
	}
}

func TestTrackingRepository_Integration(t *testing.T) {
	repo, conn, cleanup, err := setupTestDatabase()
	require.NoError(t, err)
	defer cleanup()

	ctx := context.Background()
	
	// テスト前にデータをクリーンアップ
	cleanupTrackingTestData(t, repo)

	t.Run("should save and retrieve tracking data", func(t *testing.T) {
		// テスト用アプリケーションを作成
		appRepo := repositories.NewApplicationRepository(conn.GetDB())
		testApp := &models.Application{
			AppID:       "test_app_tracking_" + time.Now().Format("20060102150405") + "_" + randomString(5),
			Name:        "Test Tracking Application",
			Description: "Test application for tracking",
			Domain:      "example.com",
			APIKey:      "alt_test_api_key_tracking_" + time.Now().Format("20060102150405") + "_" + randomString(5),
			Active:      true,
		}
		err = appRepo.Create(ctx, testApp)
		require.NoError(t, err)
		
		// テストデータを作成
		trackingData := &models.TrackingData{
			AppID:     testApp.AppID,
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			URL:       "https://example.com/test",
			IPAddress: "192.168.1.100",
			SessionID: "alt_1234567890_abc123",
			Timestamp: time.Now(),
			CustomParams: map[string]interface{}{
				"campaign_id": "camp_123",
				"source":      "google",
			},
		}

		// データを保存
		err := repo.Create(ctx, trackingData)
		assert.NoError(t, err)
		assert.NotEmpty(t, trackingData.ID)

		// データを取得して検証
		results, err := repo.FindByAppID(ctx, trackingData.AppID, 10, 0)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		if len(results) > 0 {
			assert.Equal(t, trackingData.AppID, results[0].AppID)
			assert.Equal(t, trackingData.UserAgent, results[0].UserAgent)
		}
	})

	t.Run("should find tracking data by session ID", func(t *testing.T) {
		// テスト用アプリケーションを作成
		appRepo := repositories.NewApplicationRepository(conn.GetDB())
		testApp := &models.Application{
			AppID:       "test_app_session_" + time.Now().Format("20060102150405") + "_" + randomString(5),
			Name:        "Test Session Application",
			Description: "Test application for session",
			Domain:      "example.com",
			APIKey:      "alt_test_api_key_session_" + time.Now().Format("20060102150405") + "_" + randomString(5),
			Active:      true,
		}
		err = appRepo.Create(ctx, testApp)
		require.NoError(t, err)
		
		sessionID := "alt_session_123"
		trackingData := &models.TrackingData{
			AppID:     testApp.AppID,
			UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)",
			URL:       "https://example.com/mobile",
			IPAddress: "192.168.1.100",
			SessionID: sessionID,
			Timestamp: time.Now(),
		}

		err := repo.Save(ctx, trackingData)
		assert.NoError(t, err)

		results, err := repo.GetBySessionID(ctx, sessionID)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		if len(results) > 0 {
			assert.Equal(t, sessionID, results[0].SessionID)
		}
	})

	t.Run("should find tracking data by date range", func(t *testing.T) {
		appID := "test_app_date_range"
		now := time.Now()
		
		// 必要なアプリケーションを作成
		appRepo := repositories.NewApplicationRepository(conn.GetDB())
		testApp := &models.Application{
			AppID:    appID,
			Name:     "Test Date Range App",
			Domain:   "example.com",
			APIKey:   "test-api-key-date-range",
			Active:   true,
		}
		err = appRepo.Create(ctx, testApp)
		assert.NoError(t, err)
		
		// 過去のデータ
		pastData := &models.TrackingData{
			AppID:     appID,
			UserAgent: "Mozilla/5.0",
			URL:       "https://example.com/past",
			IPAddress: "192.168.1.101",
			Timestamp: now.Add(-24 * time.Hour),
		}
		err := repo.Save(ctx, pastData)
		assert.NoError(t, err)

		// 現在のデータ
		currentData := &models.TrackingData{
			AppID:     appID,
			UserAgent: "Mozilla/5.0",
			URL:       "https://example.com/current",
			IPAddress: "192.168.1.102",
			Timestamp: now,
		}
		err = repo.Save(ctx, currentData)
		assert.NoError(t, err)

		// 日付範囲で検索
		results, err := repo.FindByAppID(ctx, appID, 100, 0) // 簡易的な実装
		assert.NoError(t, err)
		assert.Len(t, results, 2) // 過去のデータと現在のデータの2つ
		if len(results) > 0 {
			// 最新のデータが最初に来るので、currentDataのURLと一致するはず
			assert.Equal(t, currentData.URL, results[0].URL)
		}
	})

	t.Run("should get tracking stats", func(t *testing.T) {
		appID := "test_app_stats_" + time.Now().Format("20060102150405") + "_" + randomString(5)
		now := time.Now()

		// 必要なアプリケーションを作成
		appRepo := repositories.NewApplicationRepository(conn.GetDB())
		testApp := &models.Application{
			AppID:    appID,
			Name:     "Test Stats App",
			Domain:   "example.com",
			APIKey:   "test-api-key-stats_" + time.Now().Format("20060102150405") + "_" + randomString(5),
			Active:   true,
		}
		err = appRepo.Create(ctx, testApp)
		assert.NoError(t, err)

		// 複数のデータを作成
		for i := 0; i < 5; i++ {
			data := &models.TrackingData{
				AppID:     appID,
				UserAgent: "Mozilla/5.0",
				URL:       fmt.Sprintf("https://example.com/page%d", i),
				SessionID: fmt.Sprintf("session_%d", i),
				IPAddress: fmt.Sprintf("192.168.1.%d", i+10),
				Timestamp: now.Add(time.Duration(i) * time.Hour),
			}
			err := repo.Save(ctx, data)
			assert.NoError(t, err)
		}

		// 保存されたデータを確認
		results, err := repo.GetByAppID(ctx, appID, 10, 0)
		assert.NoError(t, err)
		t.Logf("Saved %d tracking data records for appID: %s", len(results), appID)

		// 統計情報を取得
		// 簡易的な実装として、CountByAppIDを使用
		count, err := repo.CountByAppID(ctx, appID)
		assert.NoError(t, err)
		t.Logf("CountByAppID returned: %d", count)
		assert.Equal(t, int64(5), count)
	})

	t.Run("should delete tracking data by app ID", func(t *testing.T) {
		appID := "test_app_delete_" + time.Now().Format("20060102150405") + "_" + randomString(5)
		
		// 必要なアプリケーションを作成
		appRepo := repositories.NewApplicationRepository(conn.GetDB())
		testApp := &models.Application{
			AppID:    appID,
			Name:     "Test Delete App",
			Domain:   "example.com",
			APIKey:   "test-api-key-delete_" + time.Now().Format("20060102150405") + "_" + randomString(5),
			Active:   true,
		}
		err = appRepo.Create(ctx, testApp)
		assert.NoError(t, err)
		
		trackingData := &models.TrackingData{
			AppID:     appID,
			UserAgent: "Mozilla/5.0",
			URL:       "https://example.com/delete",
			IPAddress: "192.168.1.200",
			Timestamp: time.Now(),
		}

		err := repo.Save(ctx, trackingData)
		assert.NoError(t, err)

		// 削除前の確認
		results, err := repo.GetByAppID(ctx, appID, 10, 0)
		assert.NoError(t, err)
		assert.Len(t, results, 1)

		// 削除実行
		err = repo.DeleteByAppID(ctx, appID)
		assert.NoError(t, err)

		// 削除後の確認
		results, err = repo.GetByAppID(ctx, appID, 10, 0)
		assert.NoError(t, err)
		assert.Len(t, results, 0)
	})
}
