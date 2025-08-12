package interfaces

import (
	"context"

	"github.com/your-username/accesslog-tracker/internal/domain/models"
)

// ApplicationRepository はアプリケーション管理を担当するインターフェース
type ApplicationRepository interface {
	// Create は新しいアプリケーションを作成
	Create(ctx context.Context, application *models.Application) error
	
	// GetByID は指定されたIDのアプリケーションを取得
	GetByID(ctx context.Context, id string) (*models.Application, error)
	
	// GetByAPIKey は指定されたAPIキーのアプリケーションを取得
	GetByAPIKey(ctx context.Context, apiKey string) (*models.Application, error)
	
	// GetByUserID は指定されたユーザーIDのアプリケーション一覧を取得
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Application, error)
	
	// Update はアプリケーション情報を更新
	Update(ctx context.Context, application *models.Application) error
	
	// Delete はアプリケーションを削除
	Delete(ctx context.Context, id string) error
	
	// List はアプリケーション一覧を取得
	List(ctx context.Context, limit, offset int) ([]*models.Application, error)
	
	// Count はアプリケーション数を取得
	Count(ctx context.Context) (int64, error)
	
	// RegenerateAPIKey はAPIキーを再生成
	RegenerateAPIKey(ctx context.Context, id string) (string, error)
	
	// UpdateSettings はアプリケーション設定を更新
	UpdateSettings(ctx context.Context, id string, settings map[string]interface{}) error
}
