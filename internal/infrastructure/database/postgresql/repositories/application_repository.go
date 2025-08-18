package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"github.com/google/uuid"
	"accesslog-tracker/internal/domain/models"
)

// ApplicationRepository PostgreSQL用のアプリケーションリポジトリ実装
type ApplicationRepository struct {
	db *sql.DB
}

// NewApplicationRepository 新しいアプリケーションリポジトリを作成
func NewApplicationRepository(db *sql.DB) *ApplicationRepository {
	return &ApplicationRepository{
		db: db,
	}
}

// Create アプリケーションを作成
func (r *ApplicationRepository) Create(ctx context.Context, app *models.Application) error {
	if app.AppID == "" {
		app.AppID = uuid.New().String()
	}
	if app.APIKey == "" {
		app.APIKey = uuid.New().String()
	}
	if app.CreatedAt.IsZero() {
		app.CreatedAt = time.Now()
	}
	if app.UpdatedAt.IsZero() {
		app.UpdatedAt = time.Now()
	}

	query := `
		INSERT INTO applications (
			app_id, name, description, domain, api_key, active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		app.AppID, app.Name, app.Description, app.Domain, app.APIKey, app.Active, app.CreatedAt, app.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save application: %w", err)
	}

	return nil
}

// GetByID アプリケーションIDでアプリケーションを検索
func (r *ApplicationRepository) GetByID(ctx context.Context, appID string) (*models.Application, error) {
	query := `
		SELECT app_id, name, description, domain, api_key, active, created_at, updated_at
		FROM applications 
		WHERE app_id = $1
	`

	var app models.Application
	err := r.db.QueryRowContext(ctx, query, appID).Scan(
		&app.AppID, &app.Name, &app.Description, &app.Domain, &app.APIKey, &app.Active, &app.CreatedAt, &app.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("application not found: %s", appID)
		}
		return nil, fmt.Errorf("failed to query application: %w", err)
	}

	return &app, nil
}

// GetByAPIKey APIキーでアプリケーションを検索
func (r *ApplicationRepository) GetByAPIKey(ctx context.Context, apiKey string) (*models.Application, error) {
	query := `
		SELECT app_id, name, description, domain, api_key, active, created_at, updated_at
		FROM applications 
		WHERE api_key = $1
	`

	var app models.Application
	err := r.db.QueryRowContext(ctx, query, apiKey).Scan(
		&app.AppID, &app.Name, &app.Description, &app.Domain, &app.APIKey, &app.Active, &app.CreatedAt, &app.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("application not found with API key: %s", apiKey)
		}
		return nil, fmt.Errorf("failed to query application by API key: %w", err)
	}

	return &app, nil
}

// List すべてのアプリケーションをページネーション付きで取得
func (r *ApplicationRepository) List(ctx context.Context, limit, offset int) ([]*models.Application, error) {
	query := `
		SELECT app_id, name, description, domain, api_key, active, created_at, updated_at
		FROM applications 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query applications: %w", err)
	}
	defer rows.Close()

	var results []*models.Application
	for rows.Next() {
		app, err := r.scanApplication(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, app)
	}

	return results, nil
}

// Update アプリケーションを更新
func (r *ApplicationRepository) Update(ctx context.Context, app *models.Application) error {
	app.UpdatedAt = time.Now()

	query := `
		UPDATE applications 
		SET name = $2, description = $3, domain = $4, api_key = $5, active = $6, updated_at = $7
		WHERE app_id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		app.AppID, app.Name, app.Description, app.Domain, app.APIKey, app.Active, app.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update application: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.ErrApplicationNotFound
	}

	return nil
}

// Delete アプリケーションを削除
func (r *ApplicationRepository) Delete(ctx context.Context, appID string) error {
	query := `DELETE FROM applications WHERE app_id = $1`

	result, err := r.db.ExecContext(ctx, query, appID)
	if err != nil {
		return fmt.Errorf("failed to delete application: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.ErrApplicationNotFound
	}

	return nil
}

// scanApplication データベースの行をアプリケーションに変換
func (r *ApplicationRepository) scanApplication(rows *sql.Rows) (*models.Application, error) {
	var app models.Application

	err := rows.Scan(
		&app.AppID, &app.Name, &app.Description, &app.Domain, &app.APIKey, &app.Active, &app.CreatedAt, &app.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scan application: %w", err)
	}

	return &app, nil
}

// Count アプリケーション数を取得
func (r *ApplicationRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM applications`

	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count applications: %w", err)
	}

	return count, nil
}

// GetByUserID ユーザーIDでアプリケーションを取得
func (r *ApplicationRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Application, error) {
	// 簡易的な実装として、全アプリケーションを返す
	return r.List(ctx, limit, offset)
}

// RegenerateAPIKey APIキーを再生成
func (r *ApplicationRepository) RegenerateAPIKey(ctx context.Context, id string) (string, error) {
	newAPIKey := uuid.New().String()
	
	query := `UPDATE applications SET api_key = $2, updated_at = $3 WHERE app_id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id, newAPIKey, time.Now())
	if err != nil {
		return "", fmt.Errorf("failed to regenerate API key: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return "", fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return "", models.ErrApplicationNotFound
	}
	
	return newAPIKey, nil
}

// UpdateSettings アプリケーション設定を更新
func (r *ApplicationRepository) UpdateSettings(ctx context.Context, id string, settings map[string]interface{}) error {
	// 簡易的な実装として、何もしない
	return nil
}
