package interfaces

import (
	"context"
	"time"

	"github.com/your-username/accesslog-tracker/internal/domain/models"
)

// TrackingRepository はトラッキングデータのリポジトリインターフェース
type TrackingRepository interface {
	// Create は新しいトラッキングデータを作成
	Create(ctx context.Context, tracking *models.TrackingData) error

	// GetByID はIDでトラッキングデータを取得
	GetByID(ctx context.Context, id string) (*models.TrackingData, error)

	// GetByApplicationID はアプリケーションIDでトラッキングデータを取得
	GetByApplicationID(ctx context.Context, applicationID string, limit, offset int) ([]*models.TrackingData, error)

	// GetBySessionID はセッションIDでトラッキングデータを取得
	GetBySessionID(ctx context.Context, sessionID string) ([]*models.TrackingData, error)

	// GetStatistics は統計情報を取得
	GetStatistics(ctx context.Context, applicationID string, startDate, endDate time.Time) (*models.TrackingStats, error)

	// DeleteByID はIDでトラッキングデータを削除
	DeleteByID(ctx context.Context, id string) error
}
