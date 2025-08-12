package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/your-username/accesslog-tracker/internal/domain/models"
	"github.com/your-username/accesslog-tracker/internal/infrastructure/cache/interfaces"
	"github.com/your-username/accesslog-tracker/internal/infrastructure/database/interfaces"
)

// MockTrackingRepository はトラッキングリポジトリのモック
type MockTrackingRepository struct {
	mock.Mock
}

func (m *MockTrackingRepository) Create(ctx context.Context, tracking *models.Tracking) error {
	args := m.Called(ctx, tracking)
	return args.Error(0)
}

func (m *MockTrackingRepository) GetByID(ctx context.Context, id string) (*models.Tracking, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Tracking), args.Error(1)
}

func (m *MockTrackingRepository) GetByApplicationID(ctx context.Context, applicationID string, limit, offset int) ([]*models.Tracking, error) {
	args := m.Called(ctx, applicationID, limit, offset)
	return args.Get(0).([]*models.Tracking), args.Error(1)
}

func (m *MockTrackingRepository) GetBySessionID(ctx context.Context, sessionID string) ([]*models.Tracking, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).([]*models.Tracking), args.Error(1)
}

func (m *MockTrackingRepository) GetByTimeRange(ctx context.Context, applicationID string, startTime, endTime time.Time, limit, offset int) ([]*models.Tracking, error) {
	args := m.Called(ctx, applicationID, startTime, endTime, limit, offset)
	return args.Get(0).([]*models.Tracking), args.Error(1)
}

func (m *MockTrackingRepository) GetStatistics(ctx context.Context, applicationID string, startTime, endTime time.Time) (*interfaces.TrackingStatistics, error) {
	args := m.Called(ctx, applicationID, startTime, endTime)
	return args.Get(0).(*interfaces.TrackingStatistics), args.Error(1)
}

func (m *MockTrackingRepository) DeleteByApplicationID(ctx context.Context, applicationID string) error {
	args := m.Called(ctx, applicationID)
	return args.Error(0)
}

func (m *MockTrackingRepository) DeleteByTimeRange(ctx context.Context, applicationID string, startTime, endTime time.Time) error {
	args := m.Called(ctx, applicationID, startTime, endTime)
	return args.Error(0)
}

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
	return args.Get(0).(*models.Application), args.Error(1)
}

func (m *MockApplicationRepository) GetByAPIKey(ctx context.Context, apiKey string) (*models.Application, error) {
	args := m.Called(ctx, apiKey)
	return args.Get(0).(*models.Application), args.Error(1)
}

func (m *MockApplicationRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Application, error) {
	args := m.Called(ctx, userID, limit, offset)
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

// MockSessionRepository はセッションリポジトリのモック
type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Create(ctx context.Context, session *models.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) GetByID(ctx context.Context, id string) (*models.Session, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockSessionRepository) GetByApplicationID(ctx context.Context, applicationID string, limit, offset int) ([]*models.Session, error) {
	args := m.Called(ctx, applicationID, limit, offset)
	return args.Get(0).([]*models.Session), args.Error(1)
}

func (m *MockSessionRepository) GetByVisitorID(ctx context.Context, visitorID string) ([]*models.Session, error) {
	args := m.Called(ctx, visitorID)
	return args.Get(0).([]*models.Session), args.Error(1)
}

func (m *MockSessionRepository) GetActiveSessions(ctx context.Context, applicationID string, limit, offset int) ([]*models.Session, error) {
	args := m.Called(ctx, applicationID, limit, offset)
	return args.Get(0).([]*models.Session), args.Error(1)
}

func (m *MockSessionRepository) Update(ctx context.Context, session *models.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) UpdateLastActivity(ctx context.Context, sessionID string, lastActivity time.Time) error {
	args := m.Called(ctx, sessionID, lastActivity)
	return args.Error(0)
}

func (m *MockSessionRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSessionRepository) DeleteExpired(ctx context.Context, before time.Time) error {
	args := m.Called(ctx, before)
	return args.Error(0)
}

func (m *MockSessionRepository) GetSessionStatistics(ctx context.Context, applicationID string, startTime, endTime time.Time) (*interfaces.SessionStatistics, error) {
	args := m.Called(ctx, applicationID, startTime, endTime)
	return args.Get(0).(*interfaces.SessionStatistics), args.Error(1)
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

func (m *MockCacheService) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacheService) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *MockCacheService) HGet(ctx context.Context, key, field string) (string, error) {
	args := m.Called(ctx, key, field)
	return args.String(0), args.Error(1)
}

func (m *MockCacheService) HSet(ctx context.Context, key, field, value string) error {
	args := m.Called(ctx, key, field, value)
	return args.Error(0)
}

func (m *MockCacheService) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockCacheService) HDel(ctx context.Context, key string, fields ...string) error {
	args := m.Called(ctx, key, fields)
	return args.Error(0)
}

func (m *MockCacheService) LPush(ctx context.Context, key string, values ...string) error {
	args := m.Called(ctx, key, values)
	return args.Error(0)
}

func (m *MockCacheService) RPush(ctx context.Context, key string, values ...string) error {
	args := m.Called(ctx, key, values)
	return args.Error(0)
}

func (m *MockCacheService) LPop(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCacheService) RPop(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCacheService) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	args := m.Called(ctx, key, start, stop)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockCacheService) SAdd(ctx context.Context, key string, members ...string) error {
	args := m.Called(ctx, key, members)
	return args.Error(0)
}

func (m *MockCacheService) SRem(ctx context.Context, key string, members ...string) error {
	args := m.Called(ctx, key, members)
	return args.Error(0)
}

func (m *MockCacheService) SMembers(ctx context.Context, key string) ([]string, error) {
	args := m.Called(ctx, key)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockCacheService) SIsMember(ctx context.Context, key, member string) (bool, error) {
	args := m.Called(ctx, key, member)
	return args.Bool(0), args.Error(1)
}

func (m *MockCacheService) ZAdd(ctx context.Context, key string, score float64, member string) error {
	args := m.Called(ctx, key, score, member)
	return args.Error(0)
}

func (m *MockCacheService) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	args := m.Called(ctx, key, start, stop)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockCacheService) ZRangeWithScores(ctx context.Context, key string, start, stop int64) (map[string]float64, error) {
	args := m.Called(ctx, key, start, stop)
	return args.Get(0).(map[string]float64), args.Error(1)
}

func (m *MockCacheService) ZRem(ctx context.Context, key string, members ...string) error {
	args := m.Called(ctx, key, members)
	return args.Error(0)
}

func (m *MockCacheService) Incr(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCacheService) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	args := m.Called(ctx, key, value)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCacheService) Decr(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCacheService) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	args := m.Called(ctx, key, value)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCacheService) Expire(ctx context.Context, key string, expiration time.Duration) error {
	args := m.Called(ctx, key, expiration)
	return args.Error(0)
}

func (m *MockCacheService) TTL(ctx context.Context, key string) (time.Duration, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(time.Duration), args.Error(1)
}

func (m *MockCacheService) Keys(ctx context.Context, pattern string) ([]string, error) {
	args := m.Called(ctx, pattern)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockCacheService) Pipeline() interfaces.Pipeline {
	args := m.Called()
	return args.Get(0).(interfaces.Pipeline)
}

func (m *MockCacheService) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCacheService) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestTrackingService_TrackPageView(t *testing.T) {
	// モックの設定
	mockTrackingRepo := new(MockTrackingRepository)
	mockApplicationRepo := new(MockApplicationRepository)
	mockSessionRepo := new(MockSessionRepository)
	mockCacheService := new(MockCacheService)
	logger := logrus.New()

	service := NewTrackingService(mockTrackingRepo, mockApplicationRepo, mockSessionRepo, mockCacheService, logger)

	ctx := context.Background()
	request := &models.TrackingRequest{
		APIKey:     "test-api-key",
		VisitorID:  "visitor-123",
		SessionID:  "session-456",
		PageURL:    "https://example.com/page",
		PageTitle:  "Test Page",
		IPAddress:  "192.168.1.1",
		UserAgent:  "Mozilla/5.0",
		Timestamp:  time.Now(),
	}

	application := &models.Application{
		ID:       uuid.New().String(),
		Name:     "Test App",
		APIKey:   "test-api-key",
		IsActive: true,
	}

	// アプリケーション取得のモック
	mockApplicationRepo.On("GetByAPIKey", ctx, request.APIKey).Return(application, nil)

	// セッション取得のモック（キャッシュにない場合）
	mockCacheService.On("Get", ctx, interfaces.SessionKey(request.SessionID)).Return("", fmt.Errorf("not found"))

	// セッション作成のモック
	mockSessionRepo.On("Create", ctx, mock.AnythingOfType("*models.Session")).Return(nil)

	// セッションキャッシュ保存のモック
	mockCacheService.On("Set", ctx, interfaces.SessionKey(request.SessionID), mock.AnythingOfType("string"), interfaces.DefaultSessionTTL).Return(nil)

	// トラッキング作成のモック
	mockTrackingRepo.On("Create", ctx, mock.AnythingOfType("*models.Tracking")).Return(nil)

	// セッション更新のモック
	mockSessionRepo.On("Update", ctx, mock.AnythingOfType("*models.Session")).Return(nil)

	// 統計キャッシュ更新のモック
	mockCacheService.On("ZIncrBy", ctx, interfaces.PageViewKey(application.ID), int64(1), request.PageURL).Return(nil)

	// テスト実行
	err := service.TrackPageView(ctx, request)

	// アサーション
	assert.NoError(t, err)
	mockApplicationRepo.AssertExpectations(t)
	mockSessionRepo.AssertExpectations(t)
	mockTrackingRepo.AssertExpectations(t)
	mockCacheService.AssertExpectations(t)
}
