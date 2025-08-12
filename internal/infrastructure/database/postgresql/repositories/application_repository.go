package repositories

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/your-username/accesslog-tracker/internal/domain/models"
	"github.com/your-username/accesslog-tracker/internal/infrastructure/database/interfaces"
	"github.com/your-username/accesslog-tracker/internal/infrastructure/database/postgresql"
)

// ApplicationRepository はアプリケーション管理のPostgreSQL実装
type ApplicationRepository struct {
	db     *postgresql.Connection
	logger *logrus.Logger
}

// NewApplicationRepository は新しいアプリケーションリポジトリを作成
func NewApplicationRepository(db *postgresql.Connection, logger *logrus.Logger) interfaces.ApplicationRepository {
	return &ApplicationRepository{
		db:     db,
		logger: logger,
	}
}

// generateAPIKey は新しいAPIキーを生成
func (r *ApplicationRepository) generateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return fmt.Sprintf("%x", bytes), nil
}

// Create は新しいアプリケーションを作成
func (r *ApplicationRepository) Create(ctx context.Context, application *models.Application) error {
	// APIキーが設定されていない場合は生成
	if application.APIKey == "" {
		apiKey, err := r.generateAPIKey()
		if err != nil {
			return fmt.Errorf("failed to generate API key: %w", err)
		}
		application.APIKey = apiKey
	}

	query := `
		INSERT INTO applications (
			id, name, description, api_key, user_id, settings, is_active
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
	`

	// 設定をJSONに変換
	settingsJSON, err := json.Marshal(application.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		application.ID,
		application.Name,
		application.Description,
		application.APIKey,
		application.UserID,
		settingsJSON,
		application.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to create application: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"application_id": application.ID,
		"name": application.Name,
		"user_id": application.UserID,
	}).Info("Created application")

	return nil
}

// GetByID は指定されたIDのアプリケーションを取得
func (r *ApplicationRepository) GetByID(ctx context.Context, id string) (*models.Application, error) {
	query := `
		SELECT id, name, description, api_key, user_id, settings, is_active, created_at, updated_at
		FROM applications
		WHERE id = $1
	`

	var application models.Application
	var settingsJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&application.ID,
		&application.Name,
		&application.Description,
		&application.APIKey,
		&application.UserID,
		&settingsJSON,
		&application.IsActive,
		&application.CreatedAt,
		&application.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("application not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get application: %w", err)
	}

	// 設定をJSONから復元
	if err := json.Unmarshal(settingsJSON, &application.Settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	return &application, nil
}

// GetByAPIKey は指定されたAPIキーのアプリケーションを取得
func (r *ApplicationRepository) GetByAPIKey(ctx context.Context, apiKey string) (*models.Application, error) {
	query := `
		SELECT id, name, description, api_key, user_id, settings, is_active, created_at, updated_at
		FROM applications
		WHERE api_key = $1 AND is_active = true
	`

	var application models.Application
	var settingsJSON []byte

	err := r.db.QueryRowContext(ctx, query, apiKey).Scan(
		&application.ID,
		&application.Name,
		&application.Description,
		&application.APIKey,
		&application.UserID,
		&settingsJSON,
		&application.IsActive,
		&application.CreatedAt,
		&application.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("application not found for API key: %s", apiKey)
		}
		return nil, fmt.Errorf("failed to get application by API key: %w", err)
	}

	// 設定をJSONから復元
	if err := json.Unmarshal(settingsJSON, &application.Settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	return &application, nil
}

// GetByUserID は指定されたユーザーIDのアプリケーション一覧を取得
func (r *ApplicationRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Application, error) {
	query := `
		SELECT id, name, description, api_key, user_id, settings, is_active, created_at, updated_at
		FROM applications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query applications: %w", err)
	}
	defer rows.Close()

	var applications []*models.Application
	for rows.Next() {
		var application models.Application
		var settingsJSON []byte

		err := rows.Scan(
			&application.ID,
			&application.Name,
			&application.Description,
			&application.APIKey,
			&application.UserID,
			&settingsJSON,
			&application.IsActive,
			&application.CreatedAt,
			&application.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan application: %w", err)
		}

		// 設定をJSONから復元
		if err := json.Unmarshal(settingsJSON, &application.Settings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
		}

		applications = append(applications, &application)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating applications: %w", err)
	}

	return applications, nil
}

// Update はアプリケーション情報を更新
func (r *ApplicationRepository) Update(ctx context.Context, application *models.Application) error {
	query := `
		UPDATE applications
		SET name = $2, description = $3, settings = $4, is_active = $5, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	// 設定をJSONに変換
	settingsJSON, err := json.Marshal(application.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query,
		application.ID,
		application.Name,
		application.Description,
		settingsJSON,
		application.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to update application: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("application not found: %s", application.ID)
	}

	r.logger.WithFields(logrus.Fields{
		"application_id": application.ID,
		"name": application.Name,
	}).Info("Updated application")

	return nil
}

// Delete はアプリケーションを削除
func (r *ApplicationRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM applications WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete application: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("application not found: %s", id)
	}

	r.logger.WithFields(logrus.Fields{
		"application_id": id,
	}).Info("Deleted application")

	return nil
}

// List はアプリケーション一覧を取得
func (r *ApplicationRepository) List(ctx context.Context, limit, offset int) ([]*models.Application, error) {
	query := `
		SELECT id, name, description, api_key, user_id, settings, is_active, created_at, updated_at
		FROM applications
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query applications: %w", err)
	}
	defer rows.Close()

	var applications []*models.Application
	for rows.Next() {
		var application models.Application
		var settingsJSON []byte

		err := rows.Scan(
			&application.ID,
			&application.Name,
			&application.Description,
			&application.APIKey,
			&application.UserID,
			&settingsJSON,
			&application.IsActive,
			&application.CreatedAt,
			&application.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan application: %w", err)
		}

		// 設定をJSONから復元
		if err := json.Unmarshal(settingsJSON, &application.Settings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
		}

		applications = append(applications, &application)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating applications: %w", err)
	}

	return applications, nil
}

// Count はアプリケーション数を取得
func (r *ApplicationRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM applications`
	
	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count applications: %w", err)
	}

	return count, nil
}

// RegenerateAPIKey はAPIキーを再生成
func (r *ApplicationRepository) RegenerateAPIKey(ctx context.Context, id string) (string, error) {
	apiKey, err := r.generateAPIKey()
	if err != nil {
		return "", fmt.Errorf("failed to generate API key: %w", err)
	}

	query := `UPDATE applications SET api_key = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id, apiKey)
	if err != nil {
		return "", fmt.Errorf("failed to regenerate API key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return "", fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return "", fmt.Errorf("application not found: %s", id)
	}

	r.logger.WithFields(logrus.Fields{
		"application_id": id,
	}).Info("Regenerated API key")

	return apiKey, nil
}

// UpdateSettings はアプリケーション設定を更新
func (r *ApplicationRepository) UpdateSettings(ctx context.Context, id string, settings map[string]interface{}) error {
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	query := `UPDATE applications SET settings = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id, settingsJSON)
	if err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("application not found: %s", id)
	}

	r.logger.WithFields(logrus.Fields{
		"application_id": id,
	}).Info("Updated application settings")

	return nil
}
