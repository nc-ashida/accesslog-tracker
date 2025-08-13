package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nc-ashida/accesslog-tracker/internal/domain/models"
	"github.com/nc-ashida/accesslog-tracker/internal/domain/validators"
	"github.com/nc-ashida/accesslog-tracker/internal/utils/crypto"
)

// ApplicationRepository はアプリケーションリポジトリのインターフェースです
type ApplicationRepository interface {
	Create(ctx context.Context, application *models.Application) error
	GetByID(ctx context.Context, id string) (*models.Application, error)
	GetByAPIKey(ctx context.Context, apiKey string) (*models.Application, error)
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Application, error)
	Update(ctx context.Context, application *models.Application) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*models.Application, error)
	Count(ctx context.Context) (int64, error)
	RegenerateAPIKey(ctx context.Context, id string) (string, error)
	UpdateSettings(ctx context.Context, id string, settings map[string]interface{}) error
}

// CacheService はキャッシュサービスのインターフェースです
type CacheService interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
}

// ApplicationService はアプリケーションのビジネスロジックを提供します
type ApplicationService struct {
	repo     ApplicationRepository
	cache    CacheService
	validator *validators.ApplicationValidator
}

// NewApplicationService は新しいアプリケーションサービスを作成します
func NewApplicationService(repo ApplicationRepository, cache CacheService) *ApplicationService {
	return &ApplicationService{
		repo:      repo,
		cache:     cache,
		validator: validators.NewApplicationValidator(),
	}
}

// Create は新しいアプリケーションを作成します
func (s *ApplicationService) Create(ctx context.Context, app *models.Application) error {
	// バリデーション
	if err := s.validator.ValidateCreate(app); err != nil {
		return err
	}

	// AppIDの生成
	if app.AppID == "" {
		app.AppID = uuid.New().String()
	}

	// APIキーの生成
	if app.APIKey == "" {
		app.APIKey = crypto.GenerateAPIKey()
	}

	// タイムスタンプの設定
	now := time.Now()
	app.CreatedAt = now
	app.UpdatedAt = now

	// デフォルト設定
	app.Active = true

	// リポジトリに保存
	if err := s.repo.Create(ctx, app); err != nil {
		return err
	}

	// キャッシュに保存
	s.cacheApplication(ctx, app)

	return nil
}

// GetByID はIDでアプリケーションを取得します
func (s *ApplicationService) GetByID(ctx context.Context, id string) (*models.Application, error) {
	// キャッシュから取得を試行
	if cached := s.getCachedApplication(ctx, id); cached != nil {
		return cached, nil
	}

	// リポジトリから取得
	app, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// キャッシュに保存
	s.cacheApplication(ctx, app)

	return app, nil
}

// GetByAPIKey はAPIキーでアプリケーションを取得します
func (s *ApplicationService) GetByAPIKey(ctx context.Context, apiKey string) (*models.Application, error) {
	// キャッシュから取得を試行
	cacheKey := "app:apikey:" + apiKey
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != "" {
		return s.GetByID(ctx, cached)
	}

	// リポジトリから取得
	app, err := s.repo.GetByAPIKey(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	// キャッシュに保存
	s.cacheApplication(ctx, app)

	return app, nil
}

// GetByUserID はユーザーIDでアプリケーション一覧を取得します
func (s *ApplicationService) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Application, error) {
	return s.repo.GetByUserID(ctx, userID, limit, offset)
}

// Update はアプリケーションを更新します
func (s *ApplicationService) Update(ctx context.Context, app *models.Application) error {
	// バリデーション
	if err := s.validator.ValidateUpdate(app); err != nil {
		return err
	}

	// タイムスタンプの更新
	app.UpdatedAt = time.Now()

	// リポジトリに保存
	if err := s.repo.Update(ctx, app); err != nil {
		return err
	}

	// キャッシュを更新
	s.cacheApplication(ctx, app)

	return nil
}

// Delete はアプリケーションを削除します
func (s *ApplicationService) Delete(ctx context.Context, id string) error {
	// リポジトリから削除
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// キャッシュから削除
	s.deleteCachedApplication(ctx, id)

	return nil
}

// List はアプリケーション一覧を取得します
func (s *ApplicationService) List(ctx context.Context, limit, offset int) ([]*models.Application, error) {
	return s.repo.List(ctx, limit, offset)
}

// Count はアプリケーション数を取得します
func (s *ApplicationService) Count(ctx context.Context) (int64, error) {
	return s.repo.Count(ctx)
}

// RegenerateAPIKey はAPIキーを再生成します
func (s *ApplicationService) RegenerateAPIKey(ctx context.Context, id string) (string, error) {
	// リポジトリで更新
	newAPIKey, err := s.repo.RegenerateAPIKey(ctx, id)
	if err != nil {
		return "", err
	}

	// キャッシュを削除（APIキーが変更されたため）
	s.deleteCachedApplication(ctx, id)

	return newAPIKey, nil
}

// UpdateSettings はアプリケーション設定を更新します
func (s *ApplicationService) UpdateSettings(ctx context.Context, id string, settings map[string]interface{}) error {
	// リポジトリで更新
	if err := s.repo.UpdateSettings(ctx, id, settings); err != nil {
		return err
	}

	// キャッシュを削除
	s.deleteCachedApplication(ctx, id)

	return nil
}

// ValidateAPIKey はAPIキーの妥当性を検証します
func (s *ApplicationService) ValidateAPIKey(ctx context.Context, apiKey string) bool {
	if !crypto.ValidateAPIKey(apiKey) {
		return false
	}

	// データベースで存在確認
	_, err := s.GetByAPIKey(ctx, apiKey)
	return err == nil
}

// cacheApplication はアプリケーションをキャッシュに保存します
func (s *ApplicationService) cacheApplication(ctx context.Context, app *models.Application) {
	if app == nil || app.AppID == "" {
		return
	}

	// アプリケーションIDでキャッシュ
	cacheKey := "app:id:" + app.AppID
	s.cache.Set(ctx, cacheKey, app.AppID, 30*time.Minute)

	// APIキーでキャッシュ
	if app.APIKey != "" {
		apiKeyCacheKey := "app:apikey:" + app.APIKey
		s.cache.Set(ctx, apiKeyCacheKey, app.AppID, 30*time.Minute)
	}
}

// getCachedApplication はキャッシュからアプリケーションを取得します
func (s *ApplicationService) getCachedApplication(ctx context.Context, id string) *models.Application {
	cacheKey := "app:id:" + id
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != "" {
		// キャッシュからIDを取得して、実際のデータを取得
		if app, err := s.repo.GetByID(ctx, cached); err == nil {
			return app
		}
	}
	return nil
}

// deleteCachedApplication はキャッシュからアプリケーションを削除します
func (s *ApplicationService) deleteCachedApplication(ctx context.Context, id string) {
	cacheKey := "app:id:" + id
	s.cache.Set(ctx, cacheKey, "", 1*time.Second) // 即座に期限切れにする
}
