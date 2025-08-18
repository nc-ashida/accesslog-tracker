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

// MockApplicationRepository はアプリケーションリポジトリのモックです
type MockApplicationRepository struct {
	mock.Mock
}

func (m *MockApplicationRepository) Create(ctx context.Context, app *models.Application) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

func (m *MockApplicationRepository) GetByID(ctx context.Context, appID string) (*models.Application, error) {
	args := m.Called(ctx, appID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Application), args.Error(1)
}

func (m *MockApplicationRepository) GetByAPIKey(ctx context.Context, apiKey string) (*models.Application, error) {
	args := m.Called(ctx, apiKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Application), args.Error(1)
}

func (m *MockApplicationRepository) List(ctx context.Context, limit, offset int) ([]*models.Application, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Application), args.Error(1)
}

func (m *MockApplicationRepository) Update(ctx context.Context, app *models.Application) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

func (m *MockApplicationRepository) Delete(ctx context.Context, appID string) error {
	args := m.Called(ctx, appID)
	return args.Error(0)
}

func (m *MockApplicationRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockApplicationRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Application, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Application), args.Error(1)
}

func (m *MockApplicationRepository) RegenerateAPIKey(ctx context.Context, appID string) (string, error) {
	args := m.Called(ctx, appID)
	return args.String(0), args.Error(1)
}

func (m *MockApplicationRepository) UpdateSettings(ctx context.Context, appID string, settings map[string]interface{}) error {
	args := m.Called(ctx, appID, settings)
	return args.Error(0)
}

// MockCacheService はキャッシュサービスのモックです
type MockCacheService struct {
	mock.Mock
}

func (m *MockCacheService) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCacheService) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func TestApplicationService_Create(t *testing.T) {
	mockRepo := &MockApplicationRepository{}
	mockCache := &MockCacheService{}
	service := services.NewApplicationService(mockRepo, mockCache)

	ctx := context.Background()
	app := &models.Application{
		Name:        "Test App",
		Description: "Test application",
		Domain:      "test.example.com",
	}

	t.Run("should create application successfully", func(t *testing.T) {
		mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Application")).Return(nil).Once()
		mockCache.On("Set", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil).Twice()

		err := service.Create(ctx, app)

		assert.NoError(t, err)
		assert.NotEmpty(t, app.AppID)
		assert.NotEmpty(t, app.APIKey)
		assert.True(t, app.Active)
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("should handle repository error", func(t *testing.T) {
		mockRepo.On("Create", ctx, mock.AnythingOfType("*models.Application")).Return(assert.AnError).Once()

		err := service.Create(ctx, app)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestApplicationService_GetByID(t *testing.T) {
	mockRepo := &MockApplicationRepository{}
	mockCache := &MockCacheService{}
	service := services.NewApplicationService(mockRepo, mockCache)

	ctx := context.Background()
	expectedApp := &models.Application{
		AppID:       "test_app_123",
		Name:        "Test App",
		Description: "Test application",
		Domain:      "test.example.com",
		APIKey:      "test_api_key",
		Active:      true,
	}

	t.Run("should get application from cache", func(t *testing.T) {
		cacheKey := "app:id:test_app_123"
		cachedAppID := "test_app_123"

		mockCache.On("Get", ctx, cacheKey).Return(cachedAppID, nil).Once()
		mockRepo.On("GetByID", ctx, cachedAppID).Return(expectedApp, nil).Once()

		result, err := service.GetByID(ctx, "test_app_123")

		assert.NoError(t, err)
		assert.Equal(t, expectedApp.AppID, result.AppID)
		assert.Equal(t, expectedApp.Name, result.Name)
		mockCache.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should get application from repository when not in cache", func(t *testing.T) {
		cacheKey := "app:id:test_app_123"
		apiKeyCacheKey := "app:apikey:test_api_key"

		mockCache.On("Get", ctx, cacheKey).Return("", assert.AnError).Once()
		mockRepo.On("GetByID", ctx, "test_app_123").Return(expectedApp, nil).Once()
		mockCache.On("Set", ctx, cacheKey, mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil).Once()
		mockCache.On("Set", ctx, apiKeyCacheKey, mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil).Once()

		result, err := service.GetByID(ctx, "test_app_123")

		assert.NoError(t, err)
		assert.Equal(t, expectedApp.AppID, result.AppID)
		assert.Equal(t, expectedApp.Name, result.Name)
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("should handle repository error", func(t *testing.T) {
		cacheKey := "app:id:test_app_123"

		mockCache.On("Get", ctx, cacheKey).Return("", assert.AnError).Once()
		mockRepo.On("GetByID", ctx, "test_app_123").Return(nil, assert.AnError).Once()

		result, err := service.GetByID(ctx, "test_app_123")

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})
}

func TestApplicationService_GetByAPIKey(t *testing.T) {
	mockRepo := &MockApplicationRepository{}
	mockCache := &MockCacheService{}
	service := services.NewApplicationService(mockRepo, mockCache)

	ctx := context.Background()
	expectedApp := &models.Application{
		AppID:       "test_app_123",
		Name:        "Test App",
		Description: "Test application",
		Domain:      "test.example.com",
		APIKey:      "test_api_key",
		Active:      true,
	}

	t.Run("should get application by API key", func(t *testing.T) {
		cacheKey := "app:apikey:test_api_key"
		idCacheKey := "app:id:test_app_123"

		mockCache.On("Get", ctx, cacheKey).Return("", assert.AnError).Once()
		mockRepo.On("GetByAPIKey", ctx, "test_api_key").Return(expectedApp, nil).Once()
		mockCache.On("Set", ctx, idCacheKey, mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil).Once()
		mockCache.On("Set", ctx, cacheKey, mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil).Once()

		result, err := service.GetByAPIKey(ctx, "test_api_key")

		assert.NoError(t, err)
		assert.Equal(t, expectedApp.AppID, result.AppID)
		assert.Equal(t, expectedApp.APIKey, result.APIKey)
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("should handle repository error", func(t *testing.T) {
		cacheKey := "app:apikey:invalid_key"

		mockCache.On("Get", ctx, cacheKey).Return("", assert.AnError).Once()
		mockRepo.On("GetByAPIKey", ctx, "invalid_key").Return(nil, assert.AnError).Once()

		result, err := service.GetByAPIKey(ctx, "invalid_key")

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})
}

func TestApplicationService_Update(t *testing.T) {
	mockRepo := &MockApplicationRepository{}
	mockCache := &MockCacheService{}
	service := services.NewApplicationService(mockRepo, mockCache)

	ctx := context.Background()
	app := &models.Application{
		AppID:       "test_app_123",
		Name:        "Updated App",
		Description: "Updated application",
		Domain:      "updated.example.com",
		APIKey:      "test_api_key",
		Active:      true,
	}

	t.Run("should update application successfully", func(t *testing.T) {
		mockRepo.On("Update", ctx, app).Return(nil).Once()
		mockCache.On("Set", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil).Twice()

		err := service.Update(ctx, app)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("should handle repository error", func(t *testing.T) {
		mockRepo.On("Update", ctx, app).Return(assert.AnError).Once()

		err := service.Update(ctx, app)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestApplicationService_Delete(t *testing.T) {
	mockRepo := &MockApplicationRepository{}
	mockCache := &MockCacheService{}
	service := services.NewApplicationService(mockRepo, mockCache)

	ctx := context.Background()

	t.Run("should delete application successfully", func(t *testing.T) {
		mockRepo.On("Delete", ctx, "test_app_123").Return(nil).Once()
		mockCache.On("Set", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil).Once()

		err := service.Delete(ctx, "test_app_123")

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("should handle repository error", func(t *testing.T) {
		mockRepo.On("Delete", ctx, "test_app_123").Return(assert.AnError).Once()

		err := service.Delete(ctx, "test_app_123")

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestApplicationService_List(t *testing.T) {
	mockRepo := &MockApplicationRepository{}
	mockCache := &MockCacheService{}
	service := services.NewApplicationService(mockRepo, mockCache)

	ctx := context.Background()
	expectedApps := []*models.Application{
		{
			AppID:       "test_app_1",
			Name:        "Test App 1",
			Description: "Test application 1",
			Domain:      "test1.example.com",
			APIKey:      "test_api_key_1",
			Active:      true,
		},
		{
			AppID:       "test_app_2",
			Name:        "Test App 2",
			Description: "Test application 2",
			Domain:      "test2.example.com",
			APIKey:      "test_api_key_2",
			Active:      true,
		},
	}

	t.Run("should list applications successfully", func(t *testing.T) {
		mockRepo.On("List", ctx, 10, 0).Return(expectedApps, nil).Once()

		result, err := service.List(ctx, 10, 0)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, expectedApps[0].AppID, result[0].AppID)
		assert.Equal(t, expectedApps[1].AppID, result[1].AppID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should handle repository error", func(t *testing.T) {
		mockRepo.On("List", ctx, 10, 0).Return(nil, assert.AnError).Once()

		result, err := service.List(ctx, 10, 0)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestApplicationService_Count(t *testing.T) {
	mockRepo := &MockApplicationRepository{}
	mockCache := &MockCacheService{}
	service := services.NewApplicationService(mockRepo, mockCache)

	ctx := context.Background()

	t.Run("should count applications successfully", func(t *testing.T) {
		mockRepo.On("Count", ctx).Return(int64(5), nil).Once()

		count, err := service.Count(ctx)

		assert.NoError(t, err)
		assert.Equal(t, int64(5), count)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should handle repository error", func(t *testing.T) {
		mockRepo.On("Count", ctx).Return(int64(0), assert.AnError).Once()

		count, err := service.Count(ctx)

		assert.Error(t, err)
		assert.Equal(t, int64(0), count)
		mockRepo.AssertExpectations(t)
	})
}

func TestApplicationService_RegenerateAPIKey(t *testing.T) {
	mockRepo := &MockApplicationRepository{}
	mockCache := &MockCacheService{}
	service := services.NewApplicationService(mockRepo, mockCache)

	ctx := context.Background()

	t.Run("should regenerate API key successfully", func(t *testing.T) {
		newAPIKey := "new_api_key_123"
		mockRepo.On("RegenerateAPIKey", ctx, "test_app_123").Return(newAPIKey, nil).Once()
		mockCache.On("Set", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil).Once()

		result, err := service.RegenerateAPIKey(ctx, "test_app_123")

		assert.NoError(t, err)
		assert.Equal(t, newAPIKey, result)
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("should handle repository error", func(t *testing.T) {
		mockRepo.On("RegenerateAPIKey", ctx, "test_app_123").Return("", assert.AnError).Once()

		result, err := service.RegenerateAPIKey(ctx, "test_app_123")

		assert.Error(t, err)
		assert.Empty(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestApplicationService_UpdateSettings(t *testing.T) {
	mockRepo := &MockApplicationRepository{}
	mockCache := &MockCacheService{}
	service := services.NewApplicationService(mockRepo, mockCache)

	ctx := context.Background()
	settings := map[string]interface{}{
		"setting1": "value1",
		"setting2": "value2",
	}

	t.Run("should update settings successfully", func(t *testing.T) {
		mockRepo.On("UpdateSettings", ctx, "test_app_123", settings).Return(nil).Once()
		mockCache.On("Set", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil).Once()

		err := service.UpdateSettings(ctx, "test_app_123", settings)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("should handle repository error", func(t *testing.T) {
		mockRepo.On("UpdateSettings", ctx, "test_app_123", settings).Return(assert.AnError).Once()

		err := service.UpdateSettings(ctx, "test_app_123", settings)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}
