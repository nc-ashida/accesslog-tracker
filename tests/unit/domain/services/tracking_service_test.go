package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"accesslog-tracker/internal/domain/models"
	"accesslog-tracker/internal/domain/services"
)

// MockTrackingRepository はトラッキングリポジトリのモックです
type MockTrackingRepository struct {
	mock.Mock
}

func (m *MockTrackingRepository) Create(ctx context.Context, tracking *models.TrackingData) error {
	args := m.Called(ctx, tracking)
	return args.Error(0)
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

func (m *MockTrackingRepository) GetByID(ctx context.Context, id string) (*models.TrackingData, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TrackingData), args.Error(1)
}

func (m *MockTrackingRepository) CountByAppID(ctx context.Context, appID string) (int64, error) {
	args := m.Called(ctx, appID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTrackingRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestNewTrackingService(t *testing.T) {
	mockRepo := &MockTrackingRepository{}

	service := services.NewTrackingService(mockRepo)

	assert.NotNil(t, service)
}

func TestTrackingService_ProcessTrackingData(t *testing.T) {
	mockRepo := &MockTrackingRepository{}
	service := services.NewTrackingService(mockRepo)

	ctx := context.Background()
	trackingData := &models.TrackingData{
		AppID:     "test_app_123",
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		URL:       "https://example.com/page1",
		IPAddress: "192.168.1.1",
		Timestamp: time.Now(),
	}

	t.Run("should process tracking data successfully", func(t *testing.T) {
		mockRepo.On("Create", ctx, mock.AnythingOfType("*models.TrackingData")).Return(nil).Once()

		err := service.ProcessTrackingData(ctx, trackingData)

		assert.NoError(t, err)
		assert.NotEmpty(t, trackingData.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should handle repository error", func(t *testing.T) {
		mockRepo.On("Create", ctx, mock.AnythingOfType("*models.TrackingData")).Return(assert.AnError).Once()

		err := service.ProcessTrackingData(ctx, trackingData)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestTrackingService_GetByID(t *testing.T) {
	mockRepo := &MockTrackingRepository{}
	service := services.NewTrackingService(mockRepo)

	ctx := context.Background()
	expectedData := &models.TrackingData{
		ID:        "alt_1234567890_abc123",
		AppID:     "test_app_123",
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		URL:       "https://example.com/page1",
		IPAddress: "192.168.1.1",
		Timestamp: time.Now(),
	}

	t.Run("should get tracking data by ID successfully", func(t *testing.T) {
		mockRepo.On("GetByID", ctx, "alt_1234567890_abc123").Return(expectedData, nil).Once()

		result, err := service.GetByID(ctx, "alt_1234567890_abc123")

		assert.NoError(t, err)
		assert.Equal(t, expectedData, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should handle repository error", func(t *testing.T) {
		mockRepo.On("GetByID", ctx, "alt_1234567890_abc123").Return(nil, assert.AnError).Once()

		result, err := service.GetByID(ctx, "alt_1234567890_abc123")

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestTrackingService_GetByAppID(t *testing.T) {
	mockRepo := &MockTrackingRepository{}
	service := services.NewTrackingService(mockRepo)

	ctx := context.Background()
	expectedData := []*models.TrackingData{
		{
			ID:        "alt_1234567890_abc123",
			AppID:     "test_app_123",
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			URL:       "https://example.com/page1",
			IPAddress: "192.168.1.1",
			Timestamp: time.Now(),
		},
	}

	t.Run("should get tracking data by app ID successfully", func(t *testing.T) {
		mockRepo.On("GetByAppID", ctx, "test_app_123", 10, 0).Return(expectedData, nil).Once()

		result, err := service.GetByAppID(ctx, "test_app_123", 10, 0)

		assert.NoError(t, err)
		assert.Equal(t, expectedData, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should handle repository error", func(t *testing.T) {
		mockRepo.On("GetByAppID", ctx, "test_app_123", 10, 0).Return(nil, assert.AnError).Once()

		result, err := service.GetByAppID(ctx, "test_app_123", 10, 0)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestTrackingService_GetBySessionID(t *testing.T) {
	mockRepo := &MockTrackingRepository{}
	service := services.NewTrackingService(mockRepo)

	ctx := context.Background()
	expectedData := []*models.TrackingData{
		{
			ID:        "alt_1234567890_abc123",
			AppID:     "test_app_123",
			SessionID: "session_123",
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			URL:       "https://example.com/page1",
			IPAddress: "192.168.1.1",
			Timestamp: time.Now(),
		},
	}

	t.Run("should get tracking data by session ID successfully", func(t *testing.T) {
		mockRepo.On("GetBySessionID", ctx, "session_123").Return(expectedData, nil).Once()

		result, err := service.GetBySessionID(ctx, "session_123")

		assert.NoError(t, err)
		assert.Equal(t, expectedData, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should handle repository error", func(t *testing.T) {
		mockRepo.On("GetBySessionID", ctx, "session_123").Return(nil, assert.AnError).Once()

		result, err := service.GetBySessionID(ctx, "session_123")

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestTrackingService_CountByAppID(t *testing.T) {
	mockRepo := &MockTrackingRepository{}
	service := services.NewTrackingService(mockRepo)

	ctx := context.Background()

	t.Run("should count tracking data by app ID successfully", func(t *testing.T) {
		mockRepo.On("CountByAppID", ctx, "test_app_123").Return(int64(100), nil).Once()

		count, err := service.CountByAppID(ctx, "test_app_123")

		assert.NoError(t, err)
		assert.Equal(t, int64(100), count)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should handle repository error", func(t *testing.T) {
		mockRepo.On("CountByAppID", ctx, "test_app_123").Return(int64(0), assert.AnError).Once()

		count, err := service.CountByAppID(ctx, "test_app_123")

		assert.Error(t, err)
		assert.Equal(t, int64(0), count)
		mockRepo.AssertExpectations(t)
	})
}

func TestTrackingService_Delete(t *testing.T) {
	mockRepo := &MockTrackingRepository{}
	service := services.NewTrackingService(mockRepo)

	ctx := context.Background()

	t.Run("should delete tracking data successfully", func(t *testing.T) {
		mockRepo.On("Delete", ctx, "alt_1234567890_abc123").Return(nil).Once()

		err := service.Delete(ctx, "alt_1234567890_abc123")

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should handle repository error", func(t *testing.T) {
		mockRepo.On("Delete", ctx, "alt_1234567890_abc123").Return(assert.AnError).Once()

		err := service.Delete(ctx, "alt_1234567890_abc123")

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}
