package middleware

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"accesslog-tracker/internal/api/middleware"
	"accesslog-tracker/internal/api/models"
	domainmodels "accesslog-tracker/internal/domain/models"
	"accesslog-tracker/internal/utils/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockApplicationService はアプリケーションサービスのモックです
type MockApplicationService struct {
	mock.Mock
}

func (m *MockApplicationService) Create(ctx context.Context, app *domainmodels.Application) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

func (m *MockApplicationService) GetByID(ctx context.Context, appID string) (*domainmodels.Application, error) {
	args := m.Called(ctx, appID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainmodels.Application), args.Error(1)
}

func (m *MockApplicationService) GetByAPIKey(ctx context.Context, apiKey string) (*domainmodels.Application, error) {
	args := m.Called(ctx, apiKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainmodels.Application), args.Error(1)
}

func (m *MockApplicationService) Update(ctx context.Context, app *domainmodels.Application) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

func (m *MockApplicationService) List(ctx context.Context, limit, offset int) ([]*domainmodels.Application, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domainmodels.Application), args.Error(1)
}

func (m *MockApplicationService) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockApplicationService) Delete(ctx context.Context, appID string) error {
	args := m.Called(ctx, appID)
	return args.Error(0)
}

func (m *MockApplicationService) RegenerateAPIKey(ctx context.Context, appID string) (string, error) {
	args := m.Called(ctx, appID)
	return args.String(0), args.Error(1)
}

func (m *MockApplicationService) UpdateSettings(ctx context.Context, appID string, settings map[string]interface{}) error {
	args := m.Called(ctx, appID, settings)
	return args.Error(0)
}

// MockLogger はロガーのモックです
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(args ...interface{}) {
	m.Called(args)
}

func (m *MockLogger) Info(args ...interface{}) {
	m.Called(args)
}

func (m *MockLogger) Warn(args ...interface{}) {
	m.Called(args)
}

func (m *MockLogger) Error(args ...interface{}) {
	m.Called(args)
}

func (m *MockLogger) Fatal(args ...interface{}) {
	m.Called(args)
}

func (m *MockLogger) Panic(args ...interface{}) {
	m.Called(args)
}

func (m *MockLogger) Debugf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Infof(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Warnf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Errorf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Fatalf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Panicf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) WithField(key string, value interface{}) logger.Logger {
	m.Called(key, value)
	return m
}

func (m *MockLogger) WithFields(fields map[string]interface{}) logger.Logger {
	m.Called(fields)
	return m
}

func (m *MockLogger) WithError(err error) logger.Logger {
	m.Called(err)
	return m
}

func (m *MockLogger) SetLevel(level string) error {
	args := m.Called(level)
	return args.Error(0)
}

func (m *MockLogger) SetFormat(format string) error {
	args := m.Called(format)
	return args.Error(0)
}

func (m *MockLogger) SetOutput(output io.Writer) {
	m.Called(output)
}

func setupAuthTest() (*gin.Engine, *MockApplicationService, *MockLogger, *middleware.AuthMiddleware) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	mockService := new(MockApplicationService)
	mockLogger := new(MockLogger)
	authMiddleware := middleware.NewAuthMiddleware(mockService, mockLogger)
	
	return router, mockService, mockLogger, authMiddleware
}

func TestAuthMiddleware_Authenticate_Success(t *testing.T) {
	router, mockService, mockLogger, authMiddleware := setupAuthTest()
	
	expectedApp := &domainmodels.Application{
		AppID:  "test-app-id",
		Name:   "Test App",
		APIKey: "alt_test_api_key_123",
		Active: true,
	}
	
	mockService.On("GetByAPIKey", mock.Anything, "alt_test_api_key_123").Return(expectedApp, nil)
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "alt_test_api_key_123")
	w := httptest.NewRecorder()
	
	router.GET("/test", authMiddleware.Authenticate(), func(c *gin.Context) {
		appID, exists := c.Get("app_id")
		assert.True(t, exists)
		assert.Equal(t, "test-app-id", appID)
		
		app, exists := c.Get("application")
		assert.True(t, exists)
		assert.Equal(t, expectedApp, app)
		
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestAuthMiddleware_Authenticate_MissingAPIKey(t *testing.T) {
	router, _, mockLogger, authMiddleware := setupAuthTest()
	
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	router.GET("/test", authMiddleware.Authenticate(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "AUTHENTICATION_ERROR", response.Error.Code)
	
	mockLogger.AssertExpectations(t)
}

func TestAuthMiddleware_Authenticate_InvalidAPIKeyFormat(t *testing.T) {
	router, _, mockLogger, authMiddleware := setupAuthTest()
	
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "invalid_key_format")
	w := httptest.NewRecorder()
	
	router.GET("/test", authMiddleware.Authenticate(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "AUTHENTICATION_ERROR", response.Error.Code)
	
	mockLogger.AssertExpectations(t)
}

func TestAuthMiddleware_Authenticate_InvalidAPIKey(t *testing.T) {
	router, mockService, mockLogger, authMiddleware := setupAuthTest()
	
	mockService.On("GetByAPIKey", mock.Anything, "alt_invalid_key").Return(nil, assert.AnError)
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "alt_invalid_key")
	w := httptest.NewRecorder()
	
	router.GET("/test", authMiddleware.Authenticate(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "AUTHENTICATION_ERROR", response.Error.Code)
	
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestAuthMiddleware_Authenticate_InactiveApplication(t *testing.T) {
	router, mockService, mockLogger, authMiddleware := setupAuthTest()
	
	inactiveApp := &domainmodels.Application{
		AppID:  "inactive-app-id",
		Name:   "Inactive App",
		APIKey: "alt_inactive_key",
		Active: false,
	}
	
	mockService.On("GetByAPIKey", mock.Anything, "alt_inactive_key").Return(inactiveApp, nil)
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "alt_inactive_key")
	w := httptest.NewRecorder()
	
	router.GET("/test", authMiddleware.Authenticate(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusForbidden, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "APPLICATION_INACTIVE", response.Error.Code)
	
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestAuthMiddleware_OptionalAuth_WithValidKey(t *testing.T) {
	router, mockService, _, authMiddleware := setupAuthTest()
	
	expectedApp := &domainmodels.Application{
		AppID:  "test-app-id",
		Name:   "Test App",
		APIKey: "alt_test_api_key_123",
		Active: true,
	}
	
	mockService.On("GetByAPIKey", mock.Anything, "alt_test_api_key_123").Return(expectedApp, nil)
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "alt_test_api_key_123")
	w := httptest.NewRecorder()
	
	router.GET("/test", authMiddleware.OptionalAuth(), func(c *gin.Context) {
		appID, exists := c.Get("app_id")
		assert.True(t, exists)
		assert.Equal(t, "test-app-id", appID)
		
		app, exists := c.Get("application")
		assert.True(t, exists)
		assert.Equal(t, expectedApp, app)
		
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	mockService.AssertExpectations(t)
}

func TestAuthMiddleware_OptionalAuth_WithoutAPIKey(t *testing.T) {
	router, _, _, authMiddleware := setupAuthTest()
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	router.GET("/test", authMiddleware.OptionalAuth(), func(c *gin.Context) {
		// コンテキストにアプリケーション情報が設定されていないことを確認
		_, exists := c.Get("app_id")
		assert.False(t, exists)
		
		_, exists = c.Get("application")
		assert.False(t, exists)
		
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_OptionalAuth_WithInvalidKey(t *testing.T) {
	router, mockService, _, authMiddleware := setupAuthTest()
	
	mockService.On("GetByAPIKey", mock.Anything, "alt_invalid_key").Return(nil, assert.AnError)
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "alt_invalid_key")
	w := httptest.NewRecorder()
	
	router.GET("/test", authMiddleware.OptionalAuth(), func(c *gin.Context) {
		// コンテキストにアプリケーション情報が設定されていないことを確認
		_, exists := c.Get("app_id")
		assert.False(t, exists)
		
		_, exists = c.Get("application")
		assert.False(t, exists)
		
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	mockService.AssertExpectations(t)
}

func TestAuthMiddleware_OptionalAuth_WithInactiveApp(t *testing.T) {
	router, mockService, _, authMiddleware := setupAuthTest()
	
	inactiveApp := &domainmodels.Application{
		AppID:  "inactive-app-id",
		Name:   "Inactive App",
		APIKey: "alt_inactive_key",
		Active: false,
	}
	
	mockService.On("GetByAPIKey", mock.Anything, "alt_inactive_key").Return(inactiveApp, nil)
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "alt_inactive_key")
	w := httptest.NewRecorder()
	
	router.GET("/test", authMiddleware.OptionalAuth(), func(c *gin.Context) {
		// アプリケーションが非アクティブの場合、コンテキストに設定されない
		_, exists := c.Get("app_id")
		assert.False(t, exists)
		
		_, exists = c.Get("application")
		assert.False(t, exists)
		
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	mockService.AssertExpectations(t)
}
