package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	"github.com/google/uuid"
	"accesslog-tracker/internal/domain/models"
)

// TrackingRepository PostgreSQL用のトラッキングリポジトリ実装
type TrackingRepository struct {
	db *sql.DB
}

// NewTrackingRepository 新しいトラッキングリポジトリを作成
func NewTrackingRepository(db *sql.DB) *TrackingRepository {
	return &TrackingRepository{
		db: db,
	}
}

// Save トラッキングデータを保存
func (r *TrackingRepository) Save(ctx context.Context, data *models.TrackingData) error {
	if data.ID == "" {
		data.ID = uuid.New().String()
	}
	if data.CreatedAt.IsZero() {
		data.CreatedAt = time.Now()
	}

	// カスタムパラメータをJSONに変換
	customParamsJSON, err := json.Marshal(data.CustomParams)
	if err != nil {
		return fmt.Errorf("failed to marshal custom params: %w", err)
	}

	query := `
		INSERT INTO access_logs (
			id, app_id, user_agent, url, ip_address, session_id, referrer, 
			timestamp, custom_params, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err = r.db.ExecContext(ctx, query,
		data.ID, data.AppID, data.UserAgent, data.URL, data.IPAddress, data.SessionID, 
		data.Referrer, data.Timestamp, customParamsJSON, data.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save tracking data: %w", err)
	}

	return nil
}

// FindByAppID アプリケーションIDでトラッキングデータを検索
func (r *TrackingRepository) FindByAppID(ctx context.Context, appID string, limit, offset int) ([]*models.TrackingData, error) {
	query := `
		SELECT id, app_id, user_agent, url, ip_address, session_id, referrer, 
		       timestamp, custom_params, created_at
		FROM access_logs 
		WHERE app_id = $1 
		ORDER BY timestamp DESC 
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, appID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query tracking data: %w", err)
	}
	defer rows.Close()

	var results []*models.TrackingData
	for rows.Next() {
		data, err := r.scanTrackingData(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, data)
	}

	return results, nil
}

// GetByAppID アプリケーションIDでトラッキングデータを検索（インターフェース互換性）
func (r *TrackingRepository) GetByAppID(ctx context.Context, appID string, limit, offset int) ([]*models.TrackingData, error) {
	return r.FindByAppID(ctx, appID, limit, offset)
}

// GetBySessionID セッションIDでトラッキングデータを検索
func (r *TrackingRepository) GetBySessionID(ctx context.Context, sessionID string) ([]*models.TrackingData, error) {
	query := `
		SELECT id, app_id, user_agent, url, ip_address, session_id, referrer, 
		       timestamp, custom_params, created_at
		FROM access_logs 
		WHERE session_id = $1 
		ORDER BY timestamp ASC
	`

	rows, err := r.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tracking data by session: %w", err)
	}
	defer rows.Close()

	var results []*models.TrackingData
	for rows.Next() {
		data, err := r.scanTrackingData(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, data)
	}

	return results, nil
}

// FindBySessionID セッションIDでトラッキングデータを検索（インターフェース互換性）
func (r *TrackingRepository) FindBySessionID(ctx context.Context, sessionID string) ([]*models.TrackingData, error) {
	return r.GetBySessionID(ctx, sessionID)
}

// FindByDateRange 日付範囲でトラッキングデータを検索
func (r *TrackingRepository) FindByDateRange(ctx context.Context, appID string, start, end time.Time) ([]*models.TrackingData, error) {
	query := `
		SELECT id, app_id, user_agent, url, ip_address, session_id, referrer, 
		       timestamp, custom_params, created_at
		FROM access_logs 
		WHERE app_id = $1 AND timestamp BETWEEN $2 AND $3
		ORDER BY timestamp DESC
	`

	rows, err := r.db.QueryContext(ctx, query, appID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query tracking data by date range: %w", err)
	}
	defer rows.Close()

	var results []*models.TrackingData
	for rows.Next() {
		data, err := r.scanTrackingData(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, data)
	}

	return results, nil
}

// GetStatsByAppID アプリケーションIDの統計情報を取得
func (r *TrackingRepository) GetStatsByAppID(ctx context.Context, appID string, start, end time.Time) (*models.TrackingStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_requests,
			COUNT(DISTINCT session_id) as unique_sessions,
			COUNT(DISTINCT ip_address) as unique_ips,
			COUNT(CASE WHEN user_agent ILIKE '%bot%' OR user_agent ILIKE '%crawler%' THEN 1 END) as bot_requests,
			COUNT(CASE WHEN user_agent ILIKE '%mobile%' OR user_agent ILIKE '%android%' OR user_agent ILIKE '%iphone%' THEN 1 END) as mobile_requests
		FROM access_logs 
		WHERE app_id = $1 AND timestamp BETWEEN $2 AND $3
	`

	var stats models.TrackingStats
	err := r.db.QueryRowContext(ctx, query, appID, start, end).Scan(
		&stats.TotalRequests,
		&stats.UniqueSessions,
		&stats.UniqueIPs,
		&stats.BotRequests,
		&stats.MobileRequests,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get tracking stats: %w", err)
	}

	stats.AppID = appID
	stats.StartDate = start
	stats.EndDate = end
	stats.CreatedAt = time.Now()

	return &stats, nil
}

// DeleteByAppID アプリケーションIDのトラッキングデータを削除
func (r *TrackingRepository) DeleteByAppID(ctx context.Context, appID string) error {
	query := `DELETE FROM access_logs WHERE app_id = $1`

	_, err := r.db.ExecContext(ctx, query, appID)
	if err != nil {
		return fmt.Errorf("failed to delete tracking data: %w", err)
	}

	return nil
}

// scanTrackingData データベースの行をトラッキングデータに変換
func (r *TrackingRepository) scanTrackingData(rows *sql.Rows) (*models.TrackingData, error) {
	var data models.TrackingData
	var customParamsJSON []byte

	err := rows.Scan(
		&data.ID, &data.AppID, &data.UserAgent, &data.URL, &data.IPAddress, &data.SessionID, &data.Referrer,
		&data.Timestamp, &customParamsJSON, &data.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scan tracking data: %w", err)
	}

	// カスタムパラメータをJSONから復元
	if len(customParamsJSON) > 0 {
		err = json.Unmarshal(customParamsJSON, &data.CustomParams)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal custom params: %w", err)
		}
	}

	return &data, nil
}

// Create トラッキングデータを作成
func (r *TrackingRepository) Create(ctx context.Context, data *models.TrackingData) error {
	return r.Save(ctx, data)
}

// GetByID IDでトラッキングデータを取得
func (r *TrackingRepository) GetByID(ctx context.Context, id string) (*models.TrackingData, error) {
	query := `
		SELECT id, app_id, user_agent, url, ip_address, session_id, referrer, 
		       timestamp, custom_params, created_at
		FROM access_logs 
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	data, err := r.scanTrackingDataFromRow(row)
	if err != nil {
		return nil, fmt.Errorf("failed to get tracking data by ID: %w", err)
	}

	return data, nil
}

// CountByAppID アプリケーションIDでトラッキングデータ数を取得
func (r *TrackingRepository) CountByAppID(ctx context.Context, appID string) (int64, error) {
	query := `SELECT COUNT(*) FROM access_logs WHERE app_id = $1`

	var count int64
	err := r.db.QueryRowContext(ctx, query, appID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count tracking data: %w", err)
	}

	return count, nil
}

// Delete IDでトラッキングデータを削除
func (r *TrackingRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM access_logs WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete tracking data: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tracking data not found: %s", id)
	}

	return nil
}

// scanTrackingDataFromRow データベースの行をトラッキングデータに変換
func (r *TrackingRepository) scanTrackingDataFromRow(row *sql.Row) (*models.TrackingData, error) {
	var data models.TrackingData
	var customParamsJSON []byte

	err := row.Scan(
		&data.ID, &data.AppID, &data.UserAgent, &data.URL, &data.IPAddress, &data.SessionID, &data.Referrer,
		&data.Timestamp, &customParamsJSON, &data.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scan tracking data: %w", err)
	}

	// カスタムパラメータをJSONから復元
	if len(customParamsJSON) > 0 {
		err = json.Unmarshal(customParamsJSON, &data.CustomParams)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal custom params: %w", err)
		}
	}

	return &data, nil
}
