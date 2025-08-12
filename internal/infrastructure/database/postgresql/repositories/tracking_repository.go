package repositories

import (
	"context"
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

// TrackingRepository はトラッキングデータのPostgreSQL実装
type TrackingRepository struct {
	db     *postgresql.Connection
	logger *logrus.Logger
}

// NewTrackingRepository は新しいトラッキングリポジトリを作成
func NewTrackingRepository(db *postgresql.Connection, logger *logrus.Logger) interfaces.TrackingRepository {
	return &TrackingRepository{
		db:     db,
		logger: logger,
	}
}

// Create は新しいトラッキングレコードを作成
func (r *TrackingRepository) Create(ctx context.Context, tracking *models.Tracking) error {
	query := `
		INSERT INTO tracking (
			id, application_id, session_id, visitor_id, page_url, page_title,
			referrer, ip_address, user_agent, country, region, city,
			latitude, longitude, device_type, browser, os, screen_width,
			screen_height, language, timezone, load_time_ms,
			dom_content_loaded_ms, first_contentful_paint_ms,
			largest_contentful_paint_ms, first_input_delay_ms,
			cumulative_layout_shift, custom_parameters, timestamp
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
			$15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26,
			$27, $28, $29
		)
	`

	// カスタムパラメータをJSONに変換
	customParamsJSON, err := json.Marshal(tracking.CustomParameters)
	if err != nil {
		return fmt.Errorf("failed to marshal custom parameters: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		tracking.ID,
		tracking.ApplicationID,
		tracking.SessionID,
		tracking.VisitorID,
		tracking.PageURL,
		tracking.PageTitle,
		tracking.Referrer,
		tracking.IPAddress,
		tracking.UserAgent,
		tracking.Country,
		tracking.Region,
		tracking.City,
		tracking.Latitude,
		tracking.Longitude,
		tracking.DeviceType,
		tracking.Browser,
		tracking.OS,
		tracking.ScreenWidth,
		tracking.ScreenHeight,
		tracking.Language,
		tracking.Timezone,
		tracking.LoadTimeMs,
		tracking.DOMContentLoadedMs,
		tracking.FirstContentfulPaintMs,
		tracking.LargestContentfulPaintMs,
		tracking.FirstInputDelayMs,
		tracking.CumulativeLayoutShift,
		customParamsJSON,
		tracking.Timestamp,
	)

	if err != nil {
		return fmt.Errorf("failed to create tracking record: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"tracking_id": tracking.ID,
		"application_id": tracking.ApplicationID,
		"visitor_id": tracking.VisitorID,
	}).Debug("Created tracking record")

	return nil
}

// GetByID は指定されたIDのトラッキングレコードを取得
func (r *TrackingRepository) GetByID(ctx context.Context, id string) (*models.Tracking, error) {
	query := `
		SELECT id, application_id, session_id, visitor_id, page_url, page_title,
			referrer, ip_address, user_agent, country, region, city,
			latitude, longitude, device_type, browser, os, screen_width,
			screen_height, language, timezone, load_time_ms,
			dom_content_loaded_ms, first_contentful_paint_ms,
			largest_contentful_paint_ms, first_input_delay_ms,
			cumulative_layout_shift, custom_parameters, timestamp, created_at
		FROM tracking
		WHERE id = $1
	`

	var tracking models.Tracking
	var customParamsJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tracking.ID,
		&tracking.ApplicationID,
		&tracking.SessionID,
		&tracking.VisitorID,
		&tracking.PageURL,
		&tracking.PageTitle,
		&tracking.Referrer,
		&tracking.IPAddress,
		&tracking.UserAgent,
		&tracking.Country,
		&tracking.Region,
		&tracking.City,
		&tracking.Latitude,
		&tracking.Longitude,
		&tracking.DeviceType,
		&tracking.Browser,
		&tracking.OS,
		&tracking.ScreenWidth,
		&tracking.ScreenHeight,
		&tracking.Language,
		&tracking.Timezone,
		&tracking.LoadTimeMs,
		&tracking.DOMContentLoadedMs,
		&tracking.FirstContentfulPaintMs,
		&tracking.LargestContentfulPaintMs,
		&tracking.FirstInputDelayMs,
		&tracking.CumulativeLayoutShift,
		&customParamsJSON,
		&tracking.Timestamp,
		&tracking.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tracking record not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get tracking record: %w", err)
	}

	// カスタムパラメータをJSONから復元
	if err := json.Unmarshal(customParamsJSON, &tracking.CustomParameters); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom parameters: %w", err)
	}

	return &tracking, nil
}

// GetByApplicationID は指定されたアプリケーションIDのトラッキングレコードを取得
func (r *TrackingRepository) GetByApplicationID(ctx context.Context, applicationID string, limit, offset int) ([]*models.Tracking, error) {
	query := `
		SELECT id, application_id, session_id, visitor_id, page_url, page_title,
			referrer, ip_address, user_agent, country, region, city,
			latitude, longitude, device_type, browser, os, screen_width,
			screen_height, language, timezone, load_time_ms,
			dom_content_loaded_ms, first_contentful_paint_ms,
			largest_contentful_paint_ms, first_input_delay_ms,
			cumulative_layout_shift, custom_parameters, timestamp, created_at
		FROM tracking
		WHERE application_id = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, applicationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query tracking records: %w", err)
	}
	defer rows.Close()

	var trackings []*models.Tracking
	for rows.Next() {
		var tracking models.Tracking
		var customParamsJSON []byte

		err := rows.Scan(
			&tracking.ID,
			&tracking.ApplicationID,
			&tracking.SessionID,
			&tracking.VisitorID,
			&tracking.PageURL,
			&tracking.PageTitle,
			&tracking.Referrer,
			&tracking.IPAddress,
			&tracking.UserAgent,
			&tracking.Country,
			&tracking.Region,
			&tracking.City,
			&tracking.Latitude,
			&tracking.Longitude,
			&tracking.DeviceType,
			&tracking.Browser,
			&tracking.OS,
			&tracking.ScreenWidth,
			&tracking.ScreenHeight,
			&tracking.Language,
			&tracking.Timezone,
			&tracking.LoadTimeMs,
			&tracking.DOMContentLoadedMs,
			&tracking.FirstContentfulPaintMs,
			&tracking.LargestContentfulPaintMs,
			&tracking.FirstInputDelayMs,
			&tracking.CumulativeLayoutShift,
			&customParamsJSON,
			&tracking.Timestamp,
			&tracking.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan tracking record: %w", err)
		}

		// カスタムパラメータをJSONから復元
		if err := json.Unmarshal(customParamsJSON, &tracking.CustomParameters); err != nil {
			return nil, fmt.Errorf("failed to unmarshal custom parameters: %w", err)
		}

		trackings = append(trackings, &tracking)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tracking records: %w", err)
	}

	return trackings, nil
}

// GetBySessionID は指定されたセッションIDのトラッキングレコードを取得
func (r *TrackingRepository) GetBySessionID(ctx context.Context, sessionID string) ([]*models.Tracking, error) {
	query := `
		SELECT id, application_id, session_id, visitor_id, page_url, page_title,
			referrer, ip_address, user_agent, country, region, city,
			latitude, longitude, device_type, browser, os, screen_width,
			screen_height, language, timezone, load_time_ms,
			dom_content_loaded_ms, first_contentful_paint_ms,
			largest_contentful_paint_ms, first_input_delay_ms,
			cumulative_layout_shift, custom_parameters, timestamp, created_at
		FROM tracking
		WHERE session_id = $1
		ORDER BY timestamp ASC
	`

	rows, err := r.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tracking records by session: %w", err)
	}
	defer rows.Close()

	var trackings []*models.Tracking
	for rows.Next() {
		var tracking models.Tracking
		var customParamsJSON []byte

		err := rows.Scan(
			&tracking.ID,
			&tracking.ApplicationID,
			&tracking.SessionID,
			&tracking.VisitorID,
			&tracking.PageURL,
			&tracking.PageTitle,
			&tracking.Referrer,
			&tracking.IPAddress,
			&tracking.UserAgent,
			&tracking.Country,
			&tracking.Region,
			&tracking.City,
			&tracking.Latitude,
			&tracking.Longitude,
			&tracking.DeviceType,
			&tracking.Browser,
			&tracking.OS,
			&tracking.ScreenWidth,
			&tracking.ScreenHeight,
			&tracking.Language,
			&tracking.Timezone,
			&tracking.LoadTimeMs,
			&tracking.DOMContentLoadedMs,
			&tracking.FirstContentfulPaintMs,
			&tracking.LargestContentfulPaintMs,
			&tracking.FirstInputDelayMs,
			&tracking.CumulativeLayoutShift,
			&customParamsJSON,
			&tracking.Timestamp,
			&tracking.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan tracking record: %w", err)
		}

		// カスタムパラメータをJSONから復元
		if err := json.Unmarshal(customParamsJSON, &tracking.CustomParameters); err != nil {
			return nil, fmt.Errorf("failed to unmarshal custom parameters: %w", err)
		}

		trackings = append(trackings, &tracking)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tracking records: %w", err)
	}

	return trackings, nil
}

// GetByTimeRange は指定された時間範囲のトラッキングレコードを取得
func (r *TrackingRepository) GetByTimeRange(ctx context.Context, applicationID string, startTime, endTime time.Time, limit, offset int) ([]*models.Tracking, error) {
	query := `
		SELECT id, application_id, session_id, visitor_id, page_url, page_title,
			referrer, ip_address, user_agent, country, region, city,
			latitude, longitude, device_type, browser, os, screen_width,
			screen_height, language, timezone, load_time_ms,
			dom_content_loaded_ms, first_contentful_paint_ms,
			largest_contentful_paint_ms, first_input_delay_ms,
			cumulative_layout_shift, custom_parameters, timestamp, created_at
		FROM tracking
		WHERE application_id = $1 AND timestamp BETWEEN $2 AND $3
		ORDER BY timestamp DESC
		LIMIT $4 OFFSET $5
	`

	rows, err := r.db.QueryContext(ctx, query, applicationID, startTime, endTime, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query tracking records by time range: %w", err)
	}
	defer rows.Close()

	var trackings []*models.Tracking
	for rows.Next() {
		var tracking models.Tracking
		var customParamsJSON []byte

		err := rows.Scan(
			&tracking.ID,
			&tracking.ApplicationID,
			&tracking.SessionID,
			&tracking.VisitorID,
			&tracking.PageURL,
			&tracking.PageTitle,
			&tracking.Referrer,
			&tracking.IPAddress,
			&tracking.UserAgent,
			&tracking.Country,
			&tracking.Region,
			&tracking.City,
			&tracking.Latitude,
			&tracking.Longitude,
			&tracking.DeviceType,
			&tracking.Browser,
			&tracking.OS,
			&tracking.ScreenWidth,
			&tracking.ScreenHeight,
			&tracking.Language,
			&tracking.Timezone,
			&tracking.LoadTimeMs,
			&tracking.DOMContentLoadedMs,
			&tracking.FirstContentfulPaintMs,
			&tracking.LargestContentfulPaintMs,
			&tracking.FirstInputDelayMs,
			&tracking.CumulativeLayoutShift,
			&customParamsJSON,
			&tracking.Timestamp,
			&tracking.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan tracking record: %w", err)
		}

		// カスタムパラメータをJSONから復元
		if err := json.Unmarshal(customParamsJSON, &tracking.CustomParameters); err != nil {
			return nil, fmt.Errorf("failed to unmarshal custom parameters: %w", err)
		}

		trackings = append(trackings, &tracking)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tracking records: %w", err)
	}

	return trackings, nil
}

// GetStatistics は統計データを取得
func (r *TrackingRepository) GetStatistics(ctx context.Context, applicationID string, startTime, endTime time.Time) (*interfaces.TrackingStatistics, error) {
	// 基本統計クエリ
	basicQuery := `
		SELECT 
			COUNT(*) as total_requests,
			COUNT(DISTINCT visitor_id) as unique_visitors,
			COUNT(DISTINCT session_id) as unique_sessions
		FROM tracking
		WHERE application_id = $1 AND timestamp BETWEEN $2 AND $3
	`

	var stats interfaces.TrackingStatistics
	err := r.db.QueryRowContext(ctx, basicQuery, applicationID, startTime, endTime).Scan(
		&stats.TotalRequests,
		&stats.UniqueVisitors,
		&stats.UniqueSessions,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get basic statistics: %w", err)
	}

	// 平均セッション時間を取得
	durationQuery := `
		SELECT AVG(duration_seconds)
		FROM sessions
		WHERE application_id = $1 AND started_at BETWEEN $2 AND $3
	`
	var avgDuration sql.NullFloat64
	err = r.db.QueryRowContext(ctx, durationQuery, applicationID, startTime, endTime).Scan(&avgDuration)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get average session duration: %w", err)
	}
	if avgDuration.Valid {
		stats.AverageSessionDuration = time.Duration(avgDuration.Float64) * time.Second
	}

	// トップページを取得
	pagesQuery := `
		SELECT page_url, COUNT(*) as count
		FROM tracking
		WHERE application_id = $1 AND timestamp BETWEEN $2 AND $3
		GROUP BY page_url
		ORDER BY count DESC
		LIMIT 10
	`
	rows, err := r.db.QueryContext(ctx, pagesQuery, applicationID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get top pages: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var page interfaces.PageView
		if err := rows.Scan(&page.Page, &page.Count); err != nil {
			return nil, fmt.Errorf("failed to scan page view: %w", err)
		}
		stats.TopPages = append(stats.TopPages, page)
	}

	// トップリファラーを取得
	referrersQuery := `
		SELECT referrer, COUNT(*) as count
		FROM tracking
		WHERE application_id = $1 AND timestamp BETWEEN $2 AND $3 AND referrer IS NOT NULL
		GROUP BY referrer
		ORDER BY count DESC
		LIMIT 10
	`
	rows, err = r.db.QueryContext(ctx, referrersQuery, applicationID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get top referrers: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var referrer interfaces.Referrer
		if err := rows.Scan(&referrer.Referrer, &referrer.Count); err != nil {
			return nil, fmt.Errorf("failed to scan referrer: %w", err)
		}
		stats.TopReferrers = append(stats.TopReferrers, referrer)
	}

	// トップユーザーエージェントを取得
	userAgentsQuery := `
		SELECT user_agent, COUNT(*) as count
		FROM tracking
		WHERE application_id = $1 AND timestamp BETWEEN $2 AND $3 AND user_agent IS NOT NULL
		GROUP BY user_agent
		ORDER BY count DESC
		LIMIT 10
	`
	rows, err = r.db.QueryContext(ctx, userAgentsQuery, applicationID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get top user agents: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var userAgent interfaces.UserAgent
		if err := rows.Scan(&userAgent.UserAgent, &userAgent.Count); err != nil {
			return nil, fmt.Errorf("failed to scan user agent: %w", err)
		}
		stats.TopUserAgents = append(stats.TopUserAgents, userAgent)
	}

	// トップ国を取得
	countriesQuery := `
		SELECT country, COUNT(*) as count
		FROM tracking
		WHERE application_id = $1 AND timestamp BETWEEN $2 AND $3 AND country IS NOT NULL
		GROUP BY country
		ORDER BY count DESC
		LIMIT 10
	`
	rows, err = r.db.QueryContext(ctx, countriesQuery, applicationID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get top countries: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var country interfaces.Country
		if err := rows.Scan(&country.Country, &country.Count); err != nil {
			return nil, fmt.Errorf("failed to scan country: %w", err)
		}
		stats.TopCountries = append(stats.TopCountries, country)
	}

	return &stats, nil
}

// DeleteByApplicationID は指定されたアプリケーションIDのトラッキングレコードを削除
func (r *TrackingRepository) DeleteByApplicationID(ctx context.Context, applicationID string) error {
	query := `DELETE FROM tracking WHERE application_id = $1`
	
	result, err := r.db.ExecContext(ctx, query, applicationID)
	if err != nil {
		return fmt.Errorf("failed to delete tracking records: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"application_id": applicationID,
		"rows_affected": rowsAffected,
	}).Info("Deleted tracking records for application")

	return nil
}

// DeleteByTimeRange は指定された時間範囲のトラッキングレコードを削除
func (r *TrackingRepository) DeleteByTimeRange(ctx context.Context, applicationID string, startTime, endTime time.Time) error {
	query := `DELETE FROM tracking WHERE application_id = $1 AND timestamp BETWEEN $2 AND $3`
	
	result, err := r.db.ExecContext(ctx, query, applicationID, startTime, endTime)
	if err != nil {
		return fmt.Errorf("failed to delete tracking records by time range: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"application_id": applicationID,
		"start_time": startTime,
		"end_time": endTime,
		"rows_affected": rowsAffected,
	}).Info("Deleted tracking records by time range")

	return nil
}
