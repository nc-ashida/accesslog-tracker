package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"accesslog-tracker/internal/domain/models"
	"accesslog-tracker/internal/domain/services"
)

// MockTrackingRepository はトラッキングリポジトリのモック
type MockTrackingRepository struct {
	mock.Mock
}

func (m *MockTrackingRepository) Create(ctx context.Context, data *models.TrackingData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockTrackingRepository) GetByID(ctx context.Context, id string) (*models.TrackingData, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TrackingData), args.Error(1)
}

func (m *MockTrackingRepository) GetByAppID(ctx context.Context, appID string, limit, offset int) ([]*models.TrackingData, error) {
	args := m.Called(ctx, appID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TrackingData), args.Error(1)
}

func (m *MockTrackingRepository) GetBySessionID(ctx context.Context, sessionID string) ([]*models.TrackingData, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TrackingData), args.Error(1)
}

func (m *MockTrackingRepository) CountByAppID(ctx context.Context, appID string) (int64, error) {
	args := m.Called(ctx, appID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTrackingRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestTrackingService_ProcessTrackingData(t *testing.T) {
	repo := new(MockTrackingRepository)
	service := services.NewTrackingService(repo)

	ctx := context.Background()
	data := &models.TrackingData{
		AppID:     "test_app_123",
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		IPAddress: "192.168.1.100",
		URL:       "https://example.com/page",
		Timestamp: time.Now(),
	}

	repo.On("Create", ctx, mock.AnythingOfType("*models.TrackingData")).Return(nil)

	err := service.ProcessTrackingData(ctx, data)

	assert.NoError(t, err)
	assert.NotEmpty(t, data.ID)
	assert.NotEmpty(t, data.SessionID)
	assert.False(t, data.CreatedAt.IsZero())
	// IPアドレスが匿名化されていることを確認
	assert.Equal(t, "192.168.1.0", data.IPAddress)
	repo.AssertExpectations(t)
}

func TestTrackingService_GetByID(t *testing.T) {
	repo := new(MockTrackingRepository)
	service := services.NewTrackingService(repo)

	ctx := context.Background()
	trackingID := "tracking123"
	expectedData := &models.TrackingData{
		ID:        trackingID,
		AppID:     "app123",
		UserAgent: "Mozilla/5.0",
		Timestamp: time.Now(),
	}

	repo.On("GetByID", ctx, trackingID).Return(expectedData, nil)

	data, err := service.GetByID(ctx, trackingID)

	assert.NoError(t, err)
	assert.Equal(t, expectedData, data)
	repo.AssertExpectations(t)
}

func TestTrackingService_GetByAppID(t *testing.T) {
	repo := new(MockTrackingRepository)
	service := services.NewTrackingService(repo)

	ctx := context.Background()
	appID := "app123"
	expectedData := []*models.TrackingData{
		{
			ID:        "tracking1",
			AppID:     appID,
			UserAgent: "Mozilla/5.0",
			Timestamp: time.Now(),
		},
		{
			ID:        "tracking2",
			AppID:     appID,
			UserAgent: "Chrome/91.0",
			Timestamp: time.Now(),
		},
	}

	repo.On("GetByAppID", ctx, appID, 10, 0).Return(expectedData, nil)

	data, err := service.GetByAppID(ctx, appID, 10, 0)

	assert.NoError(t, err)
	assert.Equal(t, expectedData, data)
	repo.AssertExpectations(t)
}

func TestTrackingService_GetBySessionID(t *testing.T) {
	repo := new(MockTrackingRepository)
	service := services.NewTrackingService(repo)

	ctx := context.Background()
	sessionID := "session123"
	expectedData := []*models.TrackingData{
		{
			ID:        "tracking1",
			SessionID: sessionID,
			AppID:     "app123",
			UserAgent: "Mozilla/5.0",
			Timestamp: time.Now(),
		},
	}

	repo.On("GetBySessionID", ctx, sessionID).Return(expectedData, nil)

	data, err := service.GetBySessionID(ctx, sessionID)

	assert.NoError(t, err)
	assert.Equal(t, expectedData, data)
	repo.AssertExpectations(t)
}

func TestTrackingService_CountByAppID(t *testing.T) {
	repo := new(MockTrackingRepository)
	service := services.NewTrackingService(repo)

	ctx := context.Background()
	appID := "app123"
	expectedCount := int64(100)

	repo.On("CountByAppID", ctx, appID).Return(expectedCount, nil)

	count, err := service.CountByAppID(ctx, appID)

	assert.NoError(t, err)
	assert.Equal(t, expectedCount, count)
	repo.AssertExpectations(t)
}

func TestTrackingService_Delete(t *testing.T) {
	repo := new(MockTrackingRepository)
	service := services.NewTrackingService(repo)

	ctx := context.Background()
	trackingID := "tracking123"

	repo.On("Delete", ctx, trackingID).Return(nil)

	err := service.Delete(ctx, trackingID)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestTrackingService_GetStatistics(t *testing.T) {
	repo := new(MockTrackingRepository)
	service := services.NewTrackingService(repo)

	ctx := context.Background()
	appID := "app123"
	startDate := time.Now().AddDate(0, 0, -7)
	endDate := time.Now()
	expectedCount := int64(100)

	repo.On("CountByAppID", ctx, appID).Return(expectedCount, nil)

	stats, err := service.GetStatistics(ctx, appID, startDate, endDate)

	assert.NoError(t, err)
	assert.Equal(t, appID, stats.AppID)
	assert.Equal(t, startDate, stats.StartDate)
	assert.Equal(t, endDate, stats.EndDate)
	assert.Equal(t, expectedCount, stats.Metrics["total_tracking_count"])
	repo.AssertExpectations(t)
}

func TestTrackingService_GetDailyStatistics(t *testing.T) {
	repo := new(MockTrackingRepository)
	service := services.NewTrackingService(repo)

	ctx := context.Background()
	appID := "app123"
	date := time.Now()
	expectedCount := int64(50)

	repo.On("CountByAppID", ctx, appID).Return(expectedCount, nil)

	stats, err := service.GetDailyStatistics(ctx, appID, date)

	assert.NoError(t, err)
	assert.Equal(t, appID, stats.AppID)
	assert.Equal(t, date, stats.Date)
	assert.Equal(t, expectedCount, stats.TotalPageViews)
	repo.AssertExpectations(t)
}

func TestTrackingService_IsValidTrackingData(t *testing.T) {
	repo := new(MockTrackingRepository)
	service := services.NewTrackingService(repo)

	// 有効なデータ
	validData := &models.TrackingData{
		AppID:     "test_app_123",
		UserAgent: "Mozilla/5.0",
		URL:       "https://example.com",
		Timestamp: time.Now(),
	}

	isValid := service.IsValidTrackingData(validData)
	assert.True(t, isValid)

	// 無効なデータ（UserAgentが空）
	invalidData := &models.TrackingData{
		AppID:     "test_app_123",
		UserAgent: "",
		URL:       "https://example.com",
		Timestamp: time.Now(),
	}

	isValid = service.IsValidTrackingData(invalidData)
	assert.False(t, isValid)
}

func TestTrackingService_GetTrackingDataByDateRange(t *testing.T) {
	repo := new(MockTrackingRepository)
	service := services.NewTrackingService(repo)

	ctx := context.Background()
	appID := "app123"
	startDate := time.Now().AddDate(0, 0, -7)
	endDate := time.Now()
	expectedData := []*models.TrackingData{
		{
			ID:        "tracking1",
			AppID:     appID,
			UserAgent: "Mozilla/5.0",
			Timestamp: time.Now(),
		},
	}

	repo.On("GetByAppID", ctx, appID, 10, 0).Return(expectedData, nil)

	data, err := service.GetTrackingDataByDateRange(ctx, appID, startDate, endDate, 10, 0)

	assert.NoError(t, err)
	assert.Equal(t, expectedData, data)
	repo.AssertExpectations(t)
}
