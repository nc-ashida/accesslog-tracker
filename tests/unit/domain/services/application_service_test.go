package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/nc-ashida/accesslog-tracker/internal/domain/models"
	"github.com/nc-ashida/accesslog-tracker/internal/domain/services"
)

// MockApplicationRepository はアプリケーションリポジトリのモック
type MockApplicationRepository struct {
	mock.Mock
}

func (m *MockApplicationRepository) Create(ctx context.Context, application *models.Application) error {
	args := m.Called(ctx, application)
	return args.Error(0)
}

func (m *MockApplicationRepository) GetByID(ctx context.Context, id string) (*models.Application, error) {
	args := m.Called(ctx, id)
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

func (m *MockApplicationRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Application, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Application), args.Error(1)
}

func (m *MockApplicationRepository) Update(ctx context.Context, application *models.Application) error {
	args := m.Called(ctx, application)
	return args.Error(0)
}

func (m *MockApplicationRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockApplicationRepository) List(ctx context.Context, limit, offset int) ([]*models.Application, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Application), args.Error(1)
}

func (m *MockApplicationRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockApplicationRepository) RegenerateAPIKey(ctx context.Context, id string) (string, error) {
	args := m.Called(ctx, id)
	return args.String(0), args.Error(1)
}

func (m *MockApplicationRepository) UpdateSettings(ctx context.Context, id string, settings map[string]interface{}) error {
	args := m.Called(ctx, id, settings)
	return args.Error(0)
}

// MockCacheService はキャッシュサービスのモック
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
	repo := new(MockApplicationRepository)
	cache := new(MockCacheService)
	service := services.NewApplicationService(repo, cache)

	ctx := context.Background()
	app := &models.Application{
		AppID:       "test_app_123",
		Name:        "Test App",
		Description: "Test Description",
		Domain:      "example.com",
	}

	repo.On("Create", ctx, mock.AnythingOfType("*models.Application")).Return(nil)
	// cacheApplicationでSetが2回呼ばれる（ID用とAPIキー用）
	cache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Twice()

	err := service.Create(ctx, app)

	assert.NoError(t, err)
	assert.NotEmpty(t, app.AppID)
	assert.NotEmpty(t, app.APIKey)
	assert.False(t, app.CreatedAt.IsZero())
	assert.False(t, app.UpdatedAt.IsZero())
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestApplicationService_GetByID(t *testing.T) {
	repo := new(MockApplicationRepository)
	cache := new(MockCacheService)
	service := services.NewApplicationService(repo, cache)

	ctx := context.Background()
	appID := "app123"
	expectedApp := &models.Application{
		AppID:  appID,
		Name:   "Test App",
		Domain: "example.com",
	}

	// キャッシュから取得を試行するため、Getメソッドのモックを設定
	cache.On("Get", ctx, "app:id:"+appID).Return("", assert.AnError)
	repo.On("GetByID", ctx, appID).Return(expectedApp, nil)
	cache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	app, err := service.GetByID(ctx, appID)

	assert.NoError(t, err)
	assert.Equal(t, expectedApp, app)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestApplicationService_GetByAPIKey(t *testing.T) {
	repo := new(MockApplicationRepository)
	cache := new(MockCacheService)
	service := services.NewApplicationService(repo, cache)

	ctx := context.Background()
	apiKey := "test-api-key-1234567890123456"
	expectedApp := &models.Application{
		AppID:  "app123",
		APIKey: apiKey,
		Name:   "Test App",
		Domain: "example.com",
	}

	// キャッシュから取得を試行するため、Getメソッドのモックを設定
	cache.On("Get", ctx, "app:apikey:"+apiKey).Return("", assert.AnError)
	repo.On("GetByAPIKey", ctx, apiKey).Return(expectedApp, nil)
	// cacheApplicationでSetが2回呼ばれる（ID用とAPIキー用）
	cache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Twice()

	app, err := service.GetByAPIKey(ctx, apiKey)

	assert.NoError(t, err)
	assert.Equal(t, expectedApp, app)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestApplicationService_Update(t *testing.T) {
	repo := new(MockApplicationRepository)
	cache := new(MockCacheService)
	service := services.NewApplicationService(repo, cache)

	ctx := context.Background()
	app := &models.Application{
		AppID:  "app123",
		Name:   "Updated App",
		Domain: "example.com",
		APIKey: "valid-api-key-1234567890123456",
	}

	repo.On("Update", ctx, app).Return(nil)
	// cacheApplicationでSetが2回呼ばれる（ID用とAPIキー用）
	cache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Twice()

	err := service.Update(ctx, app)

	assert.NoError(t, err)
	assert.False(t, app.UpdatedAt.IsZero())
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestApplicationService_Delete(t *testing.T) {
	repo := new(MockApplicationRepository)
	cache := new(MockCacheService)
	service := services.NewApplicationService(repo, cache)

	ctx := context.Background()
	appID := "app123"

	repo.On("Delete", ctx, appID).Return(nil)
	cache.On("Set", mock.Anything, mock.Anything, "", mock.Anything).Return(nil)

	err := service.Delete(ctx, appID)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestApplicationService_ValidateAPIKey(t *testing.T) {
	repo := new(MockApplicationRepository)
	cache := new(MockCacheService)
	service := services.NewApplicationService(repo, cache)

	ctx := context.Background()
	validAPIKey := "valid-api-key-1234567890123456"
	invalidAPIKey := "short"

	// 有効なAPIキーのテスト
	expectedApp := &models.Application{
		AppID:  "app123",
		APIKey: validAPIKey,
		Name:   "Test App",
		Domain: "example.com",
	}

	// ValidateAPIKeyがGetByAPIKeyを呼び出すため、キャッシュのモック設定が必要
	cache.On("Get", ctx, "app:apikey:"+validAPIKey).Return("", assert.AnError)
	repo.On("GetByAPIKey", ctx, validAPIKey).Return(expectedApp, nil)
	// GetByAPIKeyがcacheApplicationを呼び出すため、Setメソッドのモック設定が必要
	cache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Twice()

	isValid := service.ValidateAPIKey(ctx, validAPIKey)
	assert.True(t, isValid)

	// 無効なAPIキーのテスト
	isValid = service.ValidateAPIKey(ctx, invalidAPIKey)
	assert.False(t, isValid)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}
