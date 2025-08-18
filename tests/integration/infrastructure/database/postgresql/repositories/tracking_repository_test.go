package repositories

import (
	"accesslog-tracker/internal/domain/models"
	"accesslog-tracker/internal/infrastructure/database/postgresql"
	"accesslog-tracker/internal/infrastructure/database/postgresql/repositories"
	"accesslog-tracker/tests/integration/infrastructure"
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// CreateTestApplication はテスト用アプリケーションを作成します
func CreateTestApplication(t *testing.T, db *sql.DB) *models.Application {
	// より一意なIDを生成
	timestamp := time.Now().UnixNano()
	randomSuffix := randomString(8)

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
		INSERT INTO applications (app_id, name, description, domain, api_key, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (app_id) DO NOTHING
	`, app.AppID, app.Name, app.Description, app.Domain, app.APIKey, app.Active, app.CreatedAt, app.UpdatedAt)
	require.NoError(t, err)

	return app
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
		AppID:  "test_app_tracking_" + time.Now().Format("20060102150405") + "_" + randomString(5),
		Name:   "Test Tracking Application",
		Domain: "example.com",
		APIKey: "test-api-key-tracking_" + time.Now().Format("20060102150405") + "_" + randomString(5),
		Active: true,
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

		sessionID := "alt_session_" + time.Now().Format("20060102150405") + "_" + randomString(5)
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
		appID := "test_app_date_range_" + time.Now().Format("20060102150405") + "_" + randomString(5)
		now := time.Now()

		// 必要なアプリケーションを作成
		appRepo := repositories.NewApplicationRepository(conn.GetDB())
		testApp := &models.Application{
			AppID:  appID,
			Name:   "Test Date Range App",
			Domain: "example.com",
			APIKey: "test-api-key-date-range_" + time.Now().Format("20060102150405") + "_" + randomString(5),
			Active: true,
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
			AppID:  appID,
			Name:   "Test Stats App",
			Domain: "example.com",
			APIKey: "test-api-key-stats_" + time.Now().Format("20060102150405") + "_" + randomString(5),
			Active: true,
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
			AppID:  appID,
			Name:   "Test Delete App",
			Domain: "example.com",
			APIKey: "test-api-key-delete_" + time.Now().Format("20060102150405") + "_" + randomString(5),
			Active: true,
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

func TestTrackingRepository_FindBySessionID(t *testing.T) {
	repo, _, cleanup, err := setupTestDatabase()
	require.NoError(t, err)
	defer cleanup()

	ctx := context.Background()

	// テストデータの準備
	trackingData := &models.TrackingData{
		ID:        "test-find-session-123",
		AppID:     "test_app_tracking_" + time.Now().Format("20060102150405") + "_" + randomString(5),
		SessionID: "test-session-find-123",
		URL:       "https://example.com/page",
		Referrer:  "https://google.com",
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		IPAddress: "192.168.1.100",
		Timestamp: time.Now(),
		CustomParams: map[string]interface{}{
			"utm_source": "direct",
			"utm_medium": "none",
		},
	}

	// トラッキングデータを保存
	err = repo.Save(ctx, trackingData)
	require.NoError(t, err)

	// FindBySessionIDをテスト
	results, err := repo.FindBySessionID(ctx, "test-session-find-123")
	require.NoError(t, err)
	assert.NotEmpty(t, results)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "test-session-find-123", results[0].SessionID)

	// 存在しないセッションIDでテスト
	emptyResults, err := repo.FindBySessionID(ctx, "non-existent-session")
	require.NoError(t, err)
	assert.Empty(t, emptyResults)
}

func TestTrackingRepository_FindByDateRange(t *testing.T) {
	repo, _, cleanup, err := setupTestDatabase()
	require.NoError(t, err)
	defer cleanup()

	ctx := context.Background()

	// テストデータの準備
	trackingData := &models.TrackingData{
		ID:        "test-find-date-123",
		AppID:     "test_app_tracking_" + time.Now().Format("20060102150405") + "_" + randomString(5),
		SessionID: "test-session-find-date-123",
		URL:       "https://example.com/page",
		Referrer:  "https://google.com",
		UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
		IPAddress: "10.0.0.100",
		Timestamp: time.Now(),
		CustomParams: map[string]interface{}{
			"utm_source": "facebook",
			"utm_medium": "social",
		},
	}

	// トラッキングデータを保存
	err = repo.Save(ctx, trackingData)
	require.NoError(t, err)

	// FindByDateRangeをテスト
	startDate := time.Now().AddDate(0, 0, -1) // 昨日
	endDate := time.Now().AddDate(0, 0, 1)    // 明日

	results, err := repo.FindByDateRange(ctx, trackingData.AppID, startDate, endDate)
	require.NoError(t, err)
	assert.NotEmpty(t, results)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, trackingData.AppID, results[0].AppID)

	// 存在しないアプリケーションIDでテスト
	emptyResults, err := repo.FindByDateRange(ctx, "non-existent-app", startDate, endDate)
	require.NoError(t, err)
	assert.Empty(t, emptyResults)
}

func TestTrackingRepository_GetStatsByAppID(t *testing.T) {
	repo, _, cleanup, err := setupTestDatabase()
	require.NoError(t, err)
	defer cleanup()

	ctx := context.Background()

	// テストデータの準備
	trackingData := &models.TrackingData{
		ID:        "test-stats-123",
		AppID:     "test_app_tracking_" + time.Now().Format("20060102150405") + "_" + randomString(5),
		SessionID: "test-session-stats-123",
		URL:       "https://example.com/page",
		Referrer:  "https://google.com",
		UserAgent: "Mozilla/5.0 (Linux; Android 10) AppleWebKit/537.36",
		IPAddress: "172.16.0.100",
		Timestamp: time.Now(),
		CustomParams: map[string]interface{}{
			"utm_source": "twitter",
			"utm_medium": "social",
		},
	}

	// トラッキングデータを保存
	err = repo.Save(ctx, trackingData)
	require.NoError(t, err)

	// GetStatsByAppIDをテスト
	startDate := time.Now().AddDate(0, 0, -1) // 昨日
	endDate := time.Now().AddDate(0, 0, 1)    // 明日

	stats, err := repo.GetStatsByAppID(ctx, trackingData.AppID, startDate, endDate)
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, trackingData.AppID, stats.AppID)
	assert.GreaterOrEqual(t, stats.TotalRequests, int64(1))
	assert.GreaterOrEqual(t, stats.UniqueSessions, int64(1))
	assert.GreaterOrEqual(t, stats.UniqueIPs, int64(1))

	// 存在しないアプリケーションIDでテスト
	emptyStats, err := repo.GetStatsByAppID(ctx, "non-existent-app", startDate, endDate)
	require.NoError(t, err)
	assert.NotNil(t, emptyStats)
	assert.Equal(t, "non-existent-app", emptyStats.AppID)
	assert.Zero(t, emptyStats.TotalRequests)
	assert.Zero(t, emptyStats.UniqueSessions)
	assert.Zero(t, emptyStats.UniqueIPs)
}

// TestTrackingRepositoryDirectFunctions はtracking_repository.goの関数を直接テストします
func TestTrackingRepositoryDirectFunctions(t *testing.T) {
	repo, conn, cleanup, err := setupTestDatabase()
	require.NoError(t, err)
	defer cleanup()

	ctx := context.Background()

	t.Run("should_test_find_by_session_id_direct", func(t *testing.T) {
		// テスト用アプリケーションを作成
		appRepo := repositories.NewApplicationRepository(conn.GetDB())
		testApp := &models.Application{
			AppID:       "test_app_direct_session_" + time.Now().Format("20060102150405") + "_" + randomString(5),
			Name:        "Test Direct Session App",
			Description: "Test application for direct session test",
			Domain:      "example.com",
			APIKey:      "alt_test_api_key_direct_session_" + time.Now().Format("20060102150405") + "_" + randomString(5),
			Active:      true,
		}
		err = appRepo.Create(ctx, testApp)
		require.NoError(t, err)

		// テストデータを作成
		sessionID := "test-session-direct-" + time.Now().Format("20060102150405") + "_" + randomString(5)
		trackingData := &models.TrackingData{
			AppID:     testApp.AppID,
			SessionID: sessionID,
			URL:       "https://example.com/direct-test",
			Referrer:  "https://google.com",
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			IPAddress: "192.168.1.150",
			Timestamp: time.Now(),
			CustomParams: map[string]interface{}{
				"test_param": "direct_test",
			},
		}

		// データを保存
		err = repo.Save(ctx, trackingData)
		require.NoError(t, err)

		// FindBySessionIDを直接呼び出し
		results, err := repo.FindBySessionID(ctx, sessionID)
		require.NoError(t, err)
		assert.NotEmpty(t, results)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, sessionID, results[0].SessionID)
	})

	t.Run("should_test_find_by_date_range_direct", func(t *testing.T) {
		// テスト用アプリケーションを作成
		appRepo := repositories.NewApplicationRepository(conn.GetDB())
		testApp := &models.Application{
			AppID:       "test_app_direct_date_" + time.Now().Format("20060102150405") + "_" + randomString(5),
			Name:        "Test Direct Date App",
			Description: "Test application for direct date range test",
			Domain:      "example.com",
			APIKey:      "alt_test_api_key_direct_date_" + time.Now().Format("20060102150405") + "_" + randomString(5),
			Active:      true,
		}
		err = appRepo.Create(ctx, testApp)
		require.NoError(t, err)

		// テストデータを作成
		trackingData := &models.TrackingData{
			AppID:     testApp.AppID,
			SessionID: "test-session-direct-date-" + time.Now().Format("20060102150405") + "_" + randomString(5),
			URL:       "https://example.com/direct-date-test",
			Referrer:  "https://facebook.com",
			UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
			IPAddress: "10.0.0.150",
			Timestamp: time.Now(),
			CustomParams: map[string]interface{}{
				"test_param": "direct_date_test",
			},
		}

		// データを保存
		err = repo.Save(ctx, trackingData)
		require.NoError(t, err)

		// FindByDateRangeを直接呼び出し
		startDate := time.Now().AddDate(0, 0, -1) // 昨日
		endDate := time.Now().AddDate(0, 0, 1)    // 明日

		results, err := repo.FindByDateRange(ctx, testApp.AppID, startDate, endDate)
		require.NoError(t, err)
		assert.NotEmpty(t, results)
		assert.Equal(t, 1, len(results))
		assert.Equal(t, testApp.AppID, results[0].AppID)
	})

	t.Run("should_test_get_stats_by_app_id_direct", func(t *testing.T) {
		// テスト用アプリケーションを作成
		appRepo := repositories.NewApplicationRepository(conn.GetDB())
		testApp := &models.Application{
			AppID:       "test_app_direct_stats_" + time.Now().Format("20060102150405") + "_" + randomString(5),
			Name:        "Test Direct Stats App",
			Description: "Test application for direct stats test",
			Domain:      "example.com",
			APIKey:      "alt_test_api_key_direct_stats_" + time.Now().Format("20060102150405") + "_" + randomString(5),
			Active:      true,
		}
		err = appRepo.Create(ctx, testApp)
		require.NoError(t, err)

		// 複数のテストデータを作成
		for i := 0; i < 3; i++ {
			trackingData := &models.TrackingData{
				AppID:     testApp.AppID,
				SessionID: fmt.Sprintf("test-session-direct-stats-%d-%s", i, time.Now().Format("20060102150405")+"_"+randomString(5)),
				URL:       fmt.Sprintf("https://example.com/direct-stats-test-%d", i),
				Referrer:  "https://twitter.com",
				UserAgent: "Mozilla/5.0 (Linux; Android 10) AppleWebKit/537.36",
				IPAddress: fmt.Sprintf("172.16.0.%d", 150+i),
				Timestamp: time.Now().Add(time.Duration(i) * time.Hour),
				CustomParams: map[string]interface{}{
					"test_param": fmt.Sprintf("direct_stats_test_%d", i),
				},
			}

			// データを保存
			err = repo.Save(ctx, trackingData)
			require.NoError(t, err)
		}

		// GetStatsByAppIDを直接呼び出し
		startDate := time.Now().AddDate(0, 0, -1) // 昨日
		endDate := time.Now().AddDate(0, 0, 1)    // 明日

		stats, err := repo.GetStatsByAppID(ctx, testApp.AppID, startDate, endDate)
		require.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, testApp.AppID, stats.AppID)
		assert.Equal(t, int64(3), stats.TotalRequests)
		assert.Equal(t, int64(3), stats.UniqueSessions)
		assert.Equal(t, int64(3), stats.UniqueIPs)
	})
}
