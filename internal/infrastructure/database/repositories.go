package database

import (
	"context"
	"time"
	"accesslog-tracker/internal/domain/models"
)

// TrackingRepository トラッキングデータのリポジトリインターフェース
type TrackingRepository interface {
	Save(ctx context.Context, data *models.TrackingData) error
	FindByAppID(ctx context.Context, appID string, limit, offset int) ([]*models.TrackingData, error)
	FindBySessionID(ctx context.Context, sessionID string) ([]*models.TrackingData, error)
	FindByDateRange(ctx context.Context, appID string, start, end time.Time) ([]*models.TrackingData, error)
	GetStatsByAppID(ctx context.Context, appID string, start, end time.Time) (*models.TrackingStats, error)
	DeleteByAppID(ctx context.Context, appID string) error
}

// ApplicationRepository アプリケーションのリポジトリインターフェース
type ApplicationRepository interface {
	Save(ctx context.Context, app *models.Application) error
	FindByAppID(ctx context.Context, appID string) (*models.Application, error)
	FindByAPIKey(ctx context.Context, apiKey string) (*models.Application, error)
	FindAll(ctx context.Context, limit, offset int) ([]*models.Application, error)
	Update(ctx context.Context, app *models.Application) error
	Delete(ctx context.Context, appID string) error
}
