package interfaces

import (
	"context"
	"time"

	"github.com/your-username/accesslog-tracker/internal/domain/models"
)

// SessionRepository はセッション管理を担当するインターフェース
type SessionRepository interface {
	// Create は新しいセッションを作成
	Create(ctx context.Context, session *models.Session) error
	
	// GetByID は指定されたIDのセッションを取得
	GetByID(ctx context.Context, id string) (*models.Session, error)
	
	// GetByApplicationID は指定されたアプリケーションIDのセッション一覧を取得
	GetByApplicationID(ctx context.Context, applicationID string, limit, offset int) ([]*models.Session, error)
	
	// GetByVisitorID は指定されたビジターIDのセッション一覧を取得
	GetByVisitorID(ctx context.Context, visitorID string) ([]*models.Session, error)
	
	// GetActiveSessions はアクティブなセッション一覧を取得
	GetActiveSessions(ctx context.Context, applicationID string, limit, offset int) ([]*models.Session, error)
	
	// Update はセッション情報を更新
	Update(ctx context.Context, session *models.Session) error
	
	// UpdateLastActivity はセッションの最終アクティビティを更新
	UpdateLastActivity(ctx context.Context, sessionID string, lastActivity time.Time) error
	
	// Delete はセッションを削除
	Delete(ctx context.Context, id string) error
	
	// DeleteExpired は期限切れのセッションを削除
	DeleteExpired(ctx context.Context, before time.Time) error
	
	// GetSessionStatistics はセッション統計を取得
	GetSessionStatistics(ctx context.Context, applicationID string, startTime, endTime time.Time) (*SessionStatistics, error)
}

// SessionStatistics はセッション統計データを表す
type SessionStatistics struct {
	TotalSessions        int64     `json:"total_sessions"`
	ActiveSessions       int64     `json:"active_sessions"`
	AverageSessionDuration time.Duration `json:"average_session_duration"`
	AveragePagesPerSession float64  `json:"average_pages_per_session"`
	BounceRate           float64   `json:"bounce_rate"`
	ReturningVisitors    int64     `json:"returning_visitors"`
	NewVisitors          int64     `json:"new_visitors"`
}
