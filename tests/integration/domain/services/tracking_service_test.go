package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"accesslog-tracker/internal/domain/models"
	"accesslog-tracker/internal/domain/services"
	"accesslog-tracker/internal/infrastructure/database/postgresql"
	"accesslog-tracker/internal/infrastructure/database/postgresql/repositories"
	apihelpers "accesslog-tracker/tests/integration/api"
)

func TestTrackingService_GetByAppID(t *testing.T) {
	// テスト用データベースをセットアップ
	db := apihelpers.SetupTestDatabase(t)
	defer db.Close()

	trackingRepo := repositories.NewTrackingRepository(db)
	trackingService := services.NewTrackingService(trackingRepo)

	ctx := context.Background()

	// テスト用アプリケーションを先に作成
	appRepo := repositories.NewApplicationRepository(db)
	testApp := &models.Application{
		AppID:  "test-app-123",
		Name:   "Test App for GetByAppID",
		Domain: "example.com",
		APIKey: "test-api-key-123",
		Active: true,
	}
	err := appRepo.Create(ctx, testApp)
	require.NoError(t, err)

	// テストデータの準備
	trackingData := &models.TrackingData{
		ID:        "test-tracking-123",
		AppID:     "test-app-123",
		SessionID: "test-session-123",
		URL:       "https://example.com/page",
		Referrer:  "https://google.com",
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		IPAddress: "192.168.1.1",
		Timestamp: time.Now(),
		CustomParams: map[string]interface{}{
			"utm_source": "google",
			"utm_medium": "cpc",
		},
	}

	// トラッキングデータを保存
	err = trackingService.ProcessTrackingData(ctx, trackingData)
	require.NoError(t, err)

	// GetByAppIDをテスト
	trackings, err := trackingService.GetByAppID(ctx, "test-app-123", 10, 0)
	require.NoError(t, err)
	assert.NotEmpty(t, trackings)
	assert.Equal(t, 1, len(trackings))
	assert.Equal(t, "test-tracking-123", trackings[0].ID)
	assert.Equal(t, "test-app-123", trackings[0].AppID)

	// 存在しないアプリケーションIDでテスト
	emptyTrackings, err := trackingService.GetByAppID(ctx, "non-existent-app", 10, 0)
	require.NoError(t, err)
	assert.Empty(t, emptyTrackings)

	// 制限とオフセットのテスト
	limitedTrackings, err := trackingService.GetByAppID(ctx, "test-app-123", 5, 0)
	require.NoError(t, err)
	assert.Len(t, limitedTrackings, 1)

	// オフセットのテスト
	offsetTrackings, err := trackingService.GetByAppID(ctx, "test-app-123", 10, 10)
	require.NoError(t, err)
	assert.Empty(t, offsetTrackings)
}

func TestTrackingService_GetTrackingDataByDateRange(t *testing.T) {
	// テスト用データベースをセットアップ
	db := apihelpers.SetupTestDatabase(t)
	defer db.Close()

	trackingRepo := repositories.NewTrackingRepository(db)
	trackingService := services.NewTrackingService(trackingRepo)

	ctx := context.Background()

	// テスト用アプリケーションを先に作成
	appRepo := repositories.NewApplicationRepository(db)
	testApp := &models.Application{
		AppID:  "test-app-date-123",
		Name:   "Test App for DateRange",
		Domain: "example.com",
		APIKey: "test-api-key-date-123",
		Active: true,
	}
	err := appRepo.Create(ctx, testApp)
	require.NoError(t, err)

	// テストデータの準備
	trackingData := &models.TrackingData{
		ID:        "test-tracking-date-123",
		AppID:     "test-app-date-123",
		SessionID: "test-session-date-123",
		URL:       "https://example.com/page",
		Referrer:  "https://google.com",
		UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
		IPAddress: "10.0.0.1",
		Timestamp: time.Now(),
		CustomParams: map[string]interface{}{
			"utm_source": "facebook",
			"utm_medium": "social",
		},
	}

	// トラッキングデータを保存
	err = trackingService.ProcessTrackingData(ctx, trackingData)
	require.NoError(t, err)

	// 日付範囲でトラッキングデータを取得
	startDate := time.Now().AddDate(0, 0, -1) // 昨日
	endDate := time.Now().AddDate(0, 0, 1)    // 明日

	trackings, err := trackingService.GetTrackingDataByDateRange(ctx, "test-app-date-123", startDate, endDate, 10, 0)
	require.NoError(t, err)
	assert.NotEmpty(t, trackings)
	assert.Equal(t, 1, len(trackings))
	assert.Equal(t, "test-tracking-date-123", trackings[0].ID)
	assert.Equal(t, "test-app-date-123", trackings[0].AppID)
}



func TestTrackingService_GetTrackingDataByDateRange_InvalidDateRange(t *testing.T) {
	// テスト環境のセットアップ
	db := postgresql.NewConnection("test")
	err := db.Connect("host=localhost port=5432 user=postgres password=password dbname=access_log_tracker_test sslmode=disable")
	require.NoError(t, err)
	defer db.Close()

	trackingRepo := repositories.NewTrackingRepository(db.GetDB())
	trackingService := services.NewTrackingService(trackingRepo)

	ctx := context.Background()

	// 無効な日付範囲（開始日が終了日より後）
	startDate := time.Now().AddDate(0, 0, 1) // 明日
	endDate := time.Now().AddDate(0, 0, -1)  // 昨日

	_, err = trackingService.GetTrackingDataByDateRange(ctx, "test-app-invalid", startDate, endDate, 10, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid statistics period")
}

func TestTrackingService_GetStatistics_InvalidDateRange(t *testing.T) {
	// テスト環境のセットアップ
	db := postgresql.NewConnection("test")
	err := db.Connect("host=localhost port=5432 user=postgres password=password dbname=access_log_tracker_test sslmode=disable")
	require.NoError(t, err)
	defer db.Close()

	trackingRepo := repositories.NewTrackingRepository(db.GetDB())
	trackingService := services.NewTrackingService(trackingRepo)

	ctx := context.Background()

	// 無効な日付範囲（開始日が終了日より後）
	startDate := time.Now().AddDate(0, 0, 1) // 明日
	endDate := time.Now().AddDate(0, 0, -1)  // 昨日

	_, err = trackingService.GetStatistics(ctx, "test-app-invalid", startDate, endDate)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid statistics period")
}

func TestTrackingService_IsValidTrackingData(t *testing.T) {
	// テスト環境のセットアップ
	db := postgresql.NewConnection("test")
	err := db.Connect("host=localhost port=5432 user=postgres password=password dbname=access_log_tracker_test sslmode=disable")
	require.NoError(t, err)
	defer db.Close()

	trackingRepo := repositories.NewTrackingRepository(db.GetDB())
	trackingService := services.NewTrackingService(trackingRepo)

	// 有効なトラッキングデータ
	validTrackingData := &models.TrackingData{
		ID:        "test-tracking-valid-123",
		AppID:     "test-app-valid-123",
		SessionID: "test-session-valid-123",
		URL:       "https://example.com/page",
		Referrer:  "https://google.com",
		UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15",
		IPAddress: "203.0.113.1",
		Timestamp: time.Now(),
		CustomParams: map[string]interface{}{
			"utm_source": "direct",
			"utm_medium": "none",
		},
	}

	isValid := trackingService.IsValidTrackingData(validTrackingData)
	assert.True(t, isValid)

	// 無効なトラッキングデータ（AppIDが空）
	invalidTrackingData := &models.TrackingData{
		ID:           "test-tracking-invalid-123",
		AppID:        "", // 空のAppID
		SessionID:    "test-session-invalid-123",
		URL:          "https://example.com/page",
		Referrer:     "https://google.com",
		UserAgent:    "Mozilla/5.0",
		IPAddress:    "192.168.1.1",
		Timestamp:    time.Now(),
		CustomParams: map[string]interface{}{},
	}

	isValid = trackingService.IsValidTrackingData(invalidTrackingData)
	assert.False(t, isValid)
}

func TestTrackingService_GetByID(t *testing.T) {
	// テスト環境のセットアップ
	db := postgresql.NewConnection("test")
	err := db.Connect("postgres://postgres:password@postgres:5432/access_log_tracker_test?sslmode=disable")
	require.NoError(t, err)
	defer db.Close()

	trackingRepo := repositories.NewTrackingRepository(db.GetDB())
	trackingService := services.NewTrackingService(trackingRepo)

	ctx := context.Background()

	// テスト用アプリケーションを先に作成
	appRepo := repositories.NewApplicationRepository(db.GetDB())
	testApp := &models.Application{
		AppID:  "test-app-getbyid-123",
		Name:   "Test App for GetByID",
		Domain: "getbyid.example.com",
		APIKey: "test-api-key-getbyid-123",
		Active: true,
	}
	err = appRepo.Create(ctx, testApp)
	require.NoError(t, err)

	// テストデータの準備
	trackingData := &models.TrackingData{
		ID:        "test-tracking-getbyid-123",
		AppID:     "test-app-getbyid-123", // 作成したアプリケーションのIDを使用
		SessionID: "test-session-getbyid-123",
		URL:       "https://example.com/page",
		Referrer:  "https://google.com",
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		IPAddress: "192.168.1.1",
		Timestamp: time.Now(),
		CustomParams: map[string]interface{}{
			"utm_source": "google",
			"utm_medium": "cpc",
		},
	}

	// トラッキングデータを保存
	err = trackingService.ProcessTrackingData(ctx, trackingData)
	require.NoError(t, err)

	// GetByIDをテスト
	foundTracking, err := trackingService.GetByID(ctx, trackingData.ID)
	require.NoError(t, err)
	assert.NotNil(t, foundTracking)
	assert.Equal(t, trackingData.ID, foundTracking.ID)
	assert.Equal(t, trackingData.AppID, foundTracking.AppID)

	// 存在しないIDでテスト
	_, err = trackingService.GetByID(ctx, "non-existent-id")
	assert.Error(t, err)

	// クリーンアップ
	appRepo.Delete(ctx, "test-app-getbyid-123")
}

func TestTrackingService_GetBySessionID(t *testing.T) {
	// テスト環境のセットアップ
	db := postgresql.NewConnection("test")
	err := db.Connect("postgres://postgres:password@postgres:5432/access_log_tracker_test?sslmode=disable")
	require.NoError(t, err)
	defer db.Close()

	trackingRepo := repositories.NewTrackingRepository(db.GetDB())
	trackingService := services.NewTrackingService(trackingRepo)

	ctx := context.Background()

	// テスト用アプリケーションを先に作成
	appRepo := repositories.NewApplicationRepository(db.GetDB())
	testApp := &models.Application{
		AppID:  "test-app-session-123",
		Name:   "Test App for Session",
		Domain: "session.example.com",
		APIKey: "test-api-key-session-123",
		Active: true,
	}
	err = appRepo.Create(ctx, testApp)
	require.NoError(t, err)

	// テストデータの準備
	trackingData := &models.TrackingData{
		ID:        "test-tracking-session-123",
		AppID:     "test-app-session-123", // 作成したアプリケーションのIDを使用
		SessionID: "test-session-session-123",
		URL:       "https://example.com/page",
		Referrer:  "https://google.com",
		UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
		IPAddress: "10.0.0.1",
		Timestamp: time.Now(),
		CustomParams: map[string]interface{}{
			"utm_source": "facebook",
			"utm_medium": "social",
		},
	}

	// トラッキングデータを保存
	err = trackingService.ProcessTrackingData(ctx, trackingData)
	require.NoError(t, err)

	// GetBySessionIDをテスト
	sessionTrackings, err := trackingService.GetBySessionID(ctx, trackingData.SessionID)
	require.NoError(t, err)
	assert.NotEmpty(t, sessionTrackings)
	assert.Equal(t, 1, len(sessionTrackings))
	assert.Equal(t, trackingData.SessionID, sessionTrackings[0].SessionID)

	// 存在しないSessionIDでテスト
	emptyTrackings, err := trackingService.GetBySessionID(ctx, "non-existent-session")
	require.NoError(t, err)
	assert.Empty(t, emptyTrackings)

	// クリーンアップ
	appRepo.Delete(ctx, "test-app-session-123")
}

func TestTrackingService_CountByAppID(t *testing.T) {
	// テスト環境のセットアップ
	db := postgresql.NewConnection("test")
	err := db.Connect("postgres://postgres:password@postgres:5432/access_log_tracker_test?sslmode=disable")
	require.NoError(t, err)
	defer db.Close()

	trackingRepo := repositories.NewTrackingRepository(db.GetDB())
	trackingService := services.NewTrackingService(trackingRepo)

	ctx := context.Background()

	// テスト用アプリケーションを先に作成
	appRepo := repositories.NewApplicationRepository(db.GetDB())
	testApp := &models.Application{
		AppID:  "test-app-count-123",
		Name:   "Test App for Count",
		Domain: "count.example.com",
		APIKey: "test-api-key-count-123",
		Active: true,
	}
	err = appRepo.Create(ctx, testApp)
	require.NoError(t, err)

	// テストデータの準備
	trackingData := &models.TrackingData{
		ID:        "test-tracking-count-123",
		AppID:     "test-app-count-123", // 作成したアプリケーションのIDを使用
		SessionID: "test-session-count-123",
		URL:       "https://example.com/page",
		Referrer:  "https://google.com",
		UserAgent: "Mozilla/5.0 (Linux; Android 10) AppleWebKit/537.36",
		IPAddress: "172.16.0.1",
		Timestamp: time.Now(),
		CustomParams: map[string]interface{}{
			"utm_source": "twitter",
			"utm_medium": "social",
		},
	}

	// トラッキングデータを保存
	err = trackingService.ProcessTrackingData(ctx, trackingData)
	require.NoError(t, err)

	// CountByAppIDをテスト
	count, err := trackingService.CountByAppID(ctx, trackingData.AppID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(1))

	// 存在しないAppIDでテスト
	zeroCount, err := trackingService.CountByAppID(ctx, "non-existent-app")
	require.NoError(t, err)
	assert.Equal(t, int64(0), zeroCount)

	// クリーンアップ
	appRepo.Delete(ctx, "test-app-count-123")
}

func TestTrackingService_Delete(t *testing.T) {
	// テスト環境のセットアップ
	db := postgresql.NewConnection("test")
	err := db.Connect("postgres://postgres:password@postgres:5432/access_log_tracker_test?sslmode=disable")
	require.NoError(t, err)
	defer db.Close()

	trackingRepo := repositories.NewTrackingRepository(db.GetDB())
	trackingService := services.NewTrackingService(trackingRepo)

	ctx := context.Background()

	// テスト用アプリケーションを先に作成
	appRepo := repositories.NewApplicationRepository(db.GetDB())
	testApp := &models.Application{
		AppID:  "test-app-delete-123",
		Name:   "Test App for Delete",
		Domain: "delete.example.com",
		APIKey: "test-api-key-delete-123",
		Active: true,
	}
	err = appRepo.Create(ctx, testApp)
	require.NoError(t, err)

	// テストデータの準備
	trackingData := &models.TrackingData{
		ID:        "test-tracking-delete-123",
		AppID:     "test-app-delete-123", // 作成したアプリケーションのIDを使用
		SessionID: "test-session-delete-123",
		URL:       "https://example.com/page",
		Referrer:  "https://google.com",
		UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15",
		IPAddress: "203.0.113.1",
		Timestamp: time.Now(),
		CustomParams: map[string]interface{}{
			"utm_source": "direct",
			"utm_medium": "none",
		},
	}

	// トラッキングデータを保存
	err = trackingService.ProcessTrackingData(ctx, trackingData)
	require.NoError(t, err)

	// 削除前にデータが存在することを確認
	foundTracking, err := trackingService.GetByID(ctx, trackingData.ID)
	require.NoError(t, err)
	assert.NotNil(t, foundTracking)

	// Deleteをテスト
	err = trackingService.Delete(ctx, trackingData.ID)
	require.NoError(t, err)

	// 削除後にデータが存在しないことを確認
	_, err = trackingService.GetByID(ctx, trackingData.ID)
	assert.Error(t, err)

	// 存在しないIDで削除を試行
	err = trackingService.Delete(ctx, "non-existent-id")
	assert.Error(t, err)

	// クリーンアップ
	appRepo.Delete(ctx, "test-app-delete-123")
}

func TestTrackingService_GetDailyStatistics(t *testing.T) {
	// テスト環境のセットアップ
	db := postgresql.NewConnection("test")
	err := db.Connect("postgres://postgres:password@postgres:5432/access_log_tracker_test?sslmode=disable")
	require.NoError(t, err)
	defer db.Close()

	trackingRepo := repositories.NewTrackingRepository(db.GetDB())
	trackingService := services.NewTrackingService(trackingRepo)

	ctx := context.Background()

	// テスト用アプリケーションを先に作成
	appRepo := repositories.NewApplicationRepository(db.GetDB())
	testApp := &models.Application{
		AppID:  "test-app-daily-123",
		Name:   "Test App for Daily Stats",
		Domain: "daily.example.com",
		APIKey: "test-api-key-daily-123",
		Active: true,
	}
	err = appRepo.Create(ctx, testApp)
	require.NoError(t, err)

	// テストデータの準備
	trackingData := &models.TrackingData{
		ID:        "test-tracking-daily-123",
		AppID:     "test-app-daily-123",
		SessionID: "test-session-daily-123",
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
	err = trackingService.ProcessTrackingData(ctx, trackingData)
	require.NoError(t, err)

	// GetDailyStatisticsをテスト
	dailyStats, err := trackingService.GetDailyStatistics(ctx, "test-app-daily-123", time.Now())
	require.NoError(t, err)
	assert.NotNil(t, dailyStats)
	assert.Equal(t, "test-app-daily-123", dailyStats.AppID)
	assert.NotZero(t, dailyStats.TotalPageViews)
	assert.NotZero(t, dailyStats.TotalSessions)
	assert.NotZero(t, dailyStats.UniqueVisitors)
	assert.NotZero(t, dailyStats.AverageSession)

	// 存在しないアプリケーションIDでテスト
	emptyStats, err := trackingService.GetDailyStatistics(ctx, "non-existent-app", time.Now())
	require.NoError(t, err)
	assert.NotNil(t, emptyStats)
	assert.Equal(t, "non-existent-app", emptyStats.AppID)
	assert.Zero(t, emptyStats.TotalPageViews)
	assert.Zero(t, emptyStats.TotalSessions)
	assert.Zero(t, emptyStats.UniqueVisitors)

	// クリーンアップ
	appRepo.Delete(ctx, "test-app-daily-123")
}
