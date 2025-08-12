package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/your-username/accesslog-tracker/internal/domain/models"
	"github.com/your-username/accesslog-tracker/internal/domain/validators"
	"github.com/your-username/accesslog-tracker/internal/infrastructure/cache/interfaces"
	"github.com/your-username/accesslog-tracker/internal/infrastructure/database/interfaces"
)

// ApplicationService はアプリケーション管理のビジネスロジックを担当
type ApplicationService struct {
	applicationRepo interfaces.ApplicationRepository
	cacheService    interfaces.CacheService
	logger          *logrus.Logger
}

// NewApplicationService は新しいアプリケーションサービスを作成
func NewApplicationService(
	applicationRepo interfaces.ApplicationRepository,
	cacheService interfaces.CacheService,
	logger *logrus.Logger,
) *ApplicationService {
	return &ApplicationService{
		applicationRepo: applicationRepo,
		cacheService:    cacheService,
		logger:          logger,
	}
}

// CreateApplication は新しいアプリケーションを作成
func (s *ApplicationService) CreateApplication(ctx context.Context, request *models.CreateApplicationRequest) (*models.Application, error) {
	// リクエストのバリデーション
	if err := validators.ValidateCreateApplicationRequest(request); err != nil {
		return nil, fmt.Errorf("invalid create application request: %w", err)
	}

	// アプリケーション名の重複チェック
	applications, err := s.applicationRepo.GetByUserID(ctx, request.UserID, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing applications: %w", err)
	}

	for _, app := range applications {
		if app.Name == request.Name {
			return nil, fmt.Errorf("application name already exists: %s", request.Name)
		}
	}

	// アプリケーションを作成
	application := &models.Application{
		ID:          uuid.New().String(),
		Name:        request.Name,
		Description: request.Description,
		UserID:      request.UserID,
		Settings:    request.Settings,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.applicationRepo.Create(ctx, application); err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	// アプリケーションをキャッシュに保存
	appKey := interfaces.ApplicationKey(application.ID)
	if err := s.cacheService.Set(ctx, appKey, application.ID, 24*time.Hour); err != nil {
		s.logger.WithError(err).Warn("Failed to cache application")
	}

	s.logger.WithFields(logrus.Fields{
		"application_id": application.ID,
		"name":           application.Name,
		"user_id":        application.UserID,
	}).Info("Created application")

	return application, nil
}

// GetApplication はアプリケーションを取得
func (s *ApplicationService) GetApplication(ctx context.Context, id string) (*models.Application, error) {
	// まずキャッシュから取得を試行
	appKey := interfaces.ApplicationKey(id)
	if _, err := s.cacheService.Get(ctx, appKey); err == nil {
		// キャッシュに存在する場合は有効
	}

	application, err := s.applicationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get application: %w", err)
	}

	return application, nil
}

// GetApplicationByAPIKey はAPIキーでアプリケーションを取得
func (s *ApplicationService) GetApplicationByAPIKey(ctx context.Context, apiKey string) (*models.Application, error) {
	application, err := s.applicationRepo.GetByAPIKey(ctx, apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get application by API key: %w", err)
	}

	return application, nil
}

// GetApplicationsByUser はユーザーのアプリケーション一覧を取得
func (s *ApplicationService) GetApplicationsByUser(ctx context.Context, userID string, limit, offset int) ([]*models.Application, error) {
	applications, err := s.applicationRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get applications by user: %w", err)
	}

	return applications, nil
}

// UpdateApplication はアプリケーションを更新
func (s *ApplicationService) UpdateApplication(ctx context.Context, id string, request *models.UpdateApplicationRequest) (*models.Application, error) {
	// リクエストのバリデーション
	if err := validators.ValidateUpdateApplicationRequest(request); err != nil {
		return nil, fmt.Errorf("invalid update application request: %w", err)
	}

	// 既存のアプリケーションを取得
	application, err := s.applicationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get application: %w", err)
	}

	// 名前の重複チェック（変更がある場合）
	if request.Name != "" && request.Name != application.Name {
		applications, err := s.applicationRepo.GetByUserID(ctx, application.UserID, 100, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing applications: %w", err)
		}

		for _, app := range applications {
			if app.ID != id && app.Name == request.Name {
				return nil, fmt.Errorf("application name already exists: %s", request.Name)
			}
		}
	}

	// アプリケーション情報を更新
	if request.Name != "" {
		application.Name = request.Name
	}
	if request.Description != "" {
		application.Description = request.Description
	}
	if request.Settings != nil {
		application.Settings = request.Settings
	}
	if request.IsActive != nil {
		application.IsActive = *request.IsActive
	}

	application.UpdatedAt = time.Now()

	if err := s.applicationRepo.Update(ctx, application); err != nil {
		return nil, fmt.Errorf("failed to update application: %w", err)
	}

	// キャッシュを更新
	appKey := interfaces.ApplicationKey(application.ID)
	if err := s.cacheService.Set(ctx, appKey, application.ID, 24*time.Hour); err != nil {
		s.logger.WithError(err).Warn("Failed to update application cache")
	}

	s.logger.WithFields(logrus.Fields{
		"application_id": application.ID,
		"name":           application.Name,
	}).Info("Updated application")

	return application, nil
}

// DeleteApplication はアプリケーションを削除
func (s *ApplicationService) DeleteApplication(ctx context.Context, id string) error {
	// アプリケーションの存在確認
	application, err := s.applicationRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get application: %w", err)
	}

	// アプリケーションを削除
	if err := s.applicationRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete application: %w", err)
	}

	// キャッシュから削除
	appKey := interfaces.ApplicationKey(id)
	if err := s.cacheService.Delete(ctx, appKey); err != nil {
		s.logger.WithError(err).Warn("Failed to delete application from cache")
	}

	s.logger.WithFields(logrus.Fields{
		"application_id": id,
		"name":           application.Name,
	}).Info("Deleted application")

	return nil
}

// RegenerateAPIKey はAPIキーを再生成
func (s *ApplicationService) RegenerateAPIKey(ctx context.Context, id string) (string, error) {
	// アプリケーションの存在確認
	application, err := s.applicationRepo.GetByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to get application: %w", err)
	}

	// APIキーを再生成
	newAPIKey, err := s.applicationRepo.RegenerateAPIKey(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to regenerate API key: %w", err)
	}

	// キャッシュを更新
	appKey := interfaces.ApplicationKey(id)
	if err := s.cacheService.Set(ctx, appKey, id, 24*time.Hour); err != nil {
		s.logger.WithError(err).Warn("Failed to update application cache")
	}

	s.logger.WithFields(logrus.Fields{
		"application_id": id,
		"name":           application.Name,
	}).Info("Regenerated API key")

	return newAPIKey, nil
}

// UpdateApplicationSettings はアプリケーション設定を更新
func (s *ApplicationService) UpdateApplicationSettings(ctx context.Context, id string, settings map[string]interface{}) error {
	// アプリケーションの存在確認
	application, err := s.applicationRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get application: %w", err)
	}

	// 設定を更新
	if err := s.applicationRepo.UpdateSettings(ctx, id, settings); err != nil {
		return fmt.Errorf("failed to update application settings: %w", err)
	}

	// キャッシュを更新
	appKey := interfaces.ApplicationKey(id)
	if err := s.cacheService.Set(ctx, appKey, id, 24*time.Hour); err != nil {
		s.logger.WithError(err).Warn("Failed to update application cache")
	}

	s.logger.WithFields(logrus.Fields{
		"application_id": id,
		"name":           application.Name,
	}).Info("Updated application settings")

	return nil
}

// ListApplications はアプリケーション一覧を取得
func (s *ApplicationService) ListApplications(ctx context.Context, limit, offset int) ([]*models.Application, error) {
	applications, err := s.applicationRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list applications: %w", err)
	}

	return applications, nil
}

// CountApplications はアプリケーション数を取得
func (s *ApplicationService) CountApplications(ctx context.Context) (int64, error) {
	count, err := s.applicationRepo.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count applications: %w", err)
	}

	return count, nil
}

// ValidateAPIKey はAPIキーの有効性を検証
func (s *ApplicationService) ValidateAPIKey(ctx context.Context, apiKey string) (*models.Application, error) {
	application, err := s.applicationRepo.GetByAPIKey(ctx, apiKey)
	if err != nil {
		return nil, fmt.Errorf("invalid API key: %w", err)
	}

	if !application.IsActive {
		return nil, fmt.Errorf("application is not active")
	}

	return application, nil
}
