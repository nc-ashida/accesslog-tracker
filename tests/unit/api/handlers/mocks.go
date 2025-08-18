package handlers

import (
	"context"
	"io"
	"time"

	"accesslog-tracker/internal/domain/models"
	"accesslog-tracker/internal/domain/services"
	"accesslog-tracker/internal/utils/logger"

	"github.com/stretchr/testify/mock"
)

// MockLogger はロガーのモックです
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) Info(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) Warn(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) Error(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) Fatal(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) Panic(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) Debugf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...)...)
}

func (m *MockLogger) Infof(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...)...)
}

func (m *MockLogger) Warnf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...)...)
}

func (m *MockLogger) Errorf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...)...)
}

func (m *MockLogger) Fatalf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...)...)
}

func (m *MockLogger) Panicf(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...)...)
}

func (m *MockLogger) WithField(key string, value interface{}) logger.Logger {
	args := m.Called(key, value)
	return args.Get(0).(logger.Logger)
}

func (m *MockLogger) WithFields(fields map[string]interface{}) logger.Logger {
	args := m.Called(fields)
	return args.Get(0).(logger.Logger)
}

func (m *MockLogger) WithError(err error) logger.Logger {
	args := m.Called(err)
	return args.Get(0).(logger.Logger)
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

// MockApplicationService はアプリケーションサービスのモックです
type MockApplicationService struct {
	mock.Mock
}

func (m *MockApplicationService) Create(ctx context.Context, app *models.Application) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

func (m *MockApplicationService) GetByID(ctx context.Context, appID string) (*models.Application, error) {
	args := m.Called(ctx, appID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Application), args.Error(1)
}

func (m *MockApplicationService) GetByAPIKey(ctx context.Context, apiKey string) (*models.Application, error) {
	args := m.Called(ctx, apiKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Application), args.Error(1)
}

func (m *MockApplicationService) Update(ctx context.Context, app *models.Application) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

func (m *MockApplicationService) List(ctx context.Context, limit, offset int) ([]*models.Application, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Application), args.Error(1)
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

// MockTrackingService はトラッキングサービスのモックです
type MockTrackingService struct {
	mock.Mock
}

func (m *MockTrackingService) ProcessTrackingData(ctx context.Context, data *models.TrackingData) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func (m *MockTrackingService) GetStatistics(ctx context.Context, appID string, startDate, endDate time.Time) (*services.TrackingStatistics, error) {
	args := m.Called(ctx, appID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.TrackingStatistics), args.Error(1)
}

func (m *MockTrackingService) GetByID(ctx context.Context, id string) (*models.TrackingData, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TrackingData), args.Error(1)
}

func (m *MockTrackingService) GetByAppID(ctx context.Context, appID string, limit, offset int) ([]*models.TrackingData, error) {
	args := m.Called(ctx, appID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TrackingData), args.Error(1)
}

func (m *MockTrackingService) GetBySessionID(ctx context.Context, sessionID string) ([]*models.TrackingData, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TrackingData), args.Error(1)
}

func (m *MockTrackingService) CountByAppID(ctx context.Context, appID string) (int64, error) {
	args := m.Called(ctx, appID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTrackingService) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
