package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/your-username/accesslog-tracker/internal/domain/models"
	"github.com/your-username/accesslog-tracker/internal/infrastructure/database/interfaces"
	"github.com/your-username/accesslog-tracker/internal/infrastructure/database/postgresql"
)

// SessionRepository はセッション管理のPostgreSQL実装
type SessionRepository struct {
	db     *postgresql.Connection
	logger *logrus.Logger
}

// NewSessionRepository は新しいセッションリポジトリを作成
func NewSessionRepository(db *postgresql.Connection, logger *logrus.Logger) interfaces.SessionRepository {
	return &SessionRepository{
		db:     db,
		logger: logger,
	}
}

// Create は新しいセッションを作成
func (r *SessionRepository) Create(ctx context.Context, session *models.Session) error {
	query := `
		INSERT INTO sessions (
			id, application_id, visitor_id, session_id, ip_address, user_agent,
			referrer, country, region, city, latitude, longitude, device_type,
			browser, os, screen_width, screen_height, language, timezone,
			started_at, last_activity, ended_at, duration_seconds, page_views, is_bounce
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
			$15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		session.ID,
		session.ApplicationID,
		session.VisitorID,
		session.SessionID,
		session.IPAddress,
		session.UserAgent,
		session.Referrer,
		session.Country,
		session.Region,
		session.City,
		session.Latitude,
		session.Longitude,
		session.DeviceType,
		session.Browser,
		session.OS,
		session.ScreenWidth,
		session.ScreenHeight,
		session.Language,
		session.Timezone,
		session.StartedAt,
		session.LastActivity,
		session.EndedAt,
		session.DurationSeconds,
		session.PageViews,
		session.IsBounce,
	)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"session_id": session.ID,
		"application_id": session.ApplicationID,
		"visitor_id": session.VisitorID,
	}).Debug("Created session")

	return nil
}

// GetByID は指定されたIDのセッションを取得
func (r *SessionRepository) GetByID(ctx context.Context, id string) (*models.Session, error) {
	query := `
		SELECT id, application_id, visitor_id, session_id, ip_address, user_agent,
			referrer, country, region, city, latitude, longitude, device_type,
			browser, os, screen_width, screen_height, language, timezone,
			started_at, last_activity, ended_at, duration_seconds, page_views,
			is_bounce, created_at, updated_at
		FROM sessions
		WHERE id = $1
	`

	var session models.Session

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID,
		&session.ApplicationID,
		&session.VisitorID,
		&session.SessionID,
		&session.IPAddress,
		&session.UserAgent,
		&session.Referrer,
		&session.Country,
		&session.Region,
		&session.City,
		&session.Latitude,
		&session.Longitude,
		&session.DeviceType,
		&session.Browser,
		&session.OS,
		&session.ScreenWidth,
		&session.ScreenHeight,
		&session.Language,
		&session.Timezone,
		&session.StartedAt,
		&session.LastActivity,
		&session.EndedAt,
		&session.DurationSeconds,
		&session.PageViews,
		&session.IsBounce,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

// GetByApplicationID は指定されたアプリケーションIDのセッション一覧を取得
func (r *SessionRepository) GetByApplicationID(ctx context.Context, applicationID string, limit, offset int) ([]*models.Session, error) {
	query := `
		SELECT id, application_id, visitor_id, session_id, ip_address, user_agent,
			referrer, country, region, city, latitude, longitude, device_type,
			browser, os, screen_width, screen_height, language, timezone,
			started_at, last_activity, ended_at, duration_seconds, page_views,
			is_bounce, created_at, updated_at
		FROM sessions
		WHERE application_id = $1
		ORDER BY started_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, applicationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*models.Session
	for rows.Next() {
		var session models.Session

		err := rows.Scan(
			&session.ID,
			&session.ApplicationID,
			&session.VisitorID,
			&session.SessionID,
			&session.IPAddress,
			&session.UserAgent,
			&session.Referrer,
			&session.Country,
			&session.Region,
			&session.City,
			&session.Latitude,
			&session.Longitude,
			&session.DeviceType,
			&session.Browser,
			&session.OS,
			&session.ScreenWidth,
			&session.ScreenHeight,
			&session.Language,
			&session.Timezone,
			&session.StartedAt,
			&session.LastActivity,
			&session.EndedAt,
			&session.DurationSeconds,
			&session.PageViews,
			&session.IsBounce,
			&session.CreatedAt,
			&session.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		sessions = append(sessions, &session)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return sessions, nil
}

// GetByVisitorID は指定されたビジターIDのセッション一覧を取得
func (r *SessionRepository) GetByVisitorID(ctx context.Context, visitorID string) ([]*models.Session, error) {
	query := `
		SELECT id, application_id, visitor_id, session_id, ip_address, user_agent,
			referrer, country, region, city, latitude, longitude, device_type,
			browser, os, screen_width, screen_height, language, timezone,
			started_at, last_activity, ended_at, duration_seconds, page_views,
			is_bounce, created_at, updated_at
		FROM sessions
		WHERE visitor_id = $1
		ORDER BY started_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, visitorID)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions by visitor: %w", err)
	}
	defer rows.Close()

	var sessions []*models.Session
	for rows.Next() {
		var session models.Session

		err := rows.Scan(
			&session.ID,
			&session.ApplicationID,
			&session.VisitorID,
			&session.SessionID,
			&session.IPAddress,
			&session.UserAgent,
			&session.Referrer,
			&session.Country,
			&session.Region,
			&session.City,
			&session.Latitude,
			&session.Longitude,
			&session.DeviceType,
			&session.Browser,
			&session.OS,
			&session.ScreenWidth,
			&session.ScreenHeight,
			&session.Language,
			&session.Timezone,
			&session.StartedAt,
			&session.LastActivity,
			&session.EndedAt,
			&session.DurationSeconds,
			&session.PageViews,
			&session.IsBounce,
			&session.CreatedAt,
			&session.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		sessions = append(sessions, &session)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return sessions, nil
}

// GetActiveSessions はアクティブなセッション一覧を取得
func (r *SessionRepository) GetActiveSessions(ctx context.Context, applicationID string, limit, offset int) ([]*models.Session, error) {
	query := `
		SELECT id, application_id, visitor_id, session_id, ip_address, user_agent,
			referrer, country, region, city, latitude, longitude, device_type,
			browser, os, screen_width, screen_height, language, timezone,
			started_at, last_activity, ended_at, duration_seconds, page_views,
			is_bounce, created_at, updated_at
		FROM sessions
		WHERE application_id = $1 AND ended_at IS NULL
		ORDER BY last_activity DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, applicationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query active sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*models.Session
	for rows.Next() {
		var session models.Session

		err := rows.Scan(
			&session.ID,
			&session.ApplicationID,
			&session.VisitorID,
			&session.SessionID,
			&session.IPAddress,
			&session.UserAgent,
			&session.Referrer,
			&session.Country,
			&session.Region,
			&session.City,
			&session.Latitude,
			&session.Longitude,
			&session.DeviceType,
			&session.Browser,
			&session.OS,
			&session.ScreenWidth,
			&session.ScreenHeight,
			&session.Language,
			&session.Timezone,
			&session.StartedAt,
			&session.LastActivity,
			&session.EndedAt,
			&session.DurationSeconds,
			&session.PageViews,
			&session.IsBounce,
			&session.CreatedAt,
			&session.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		sessions = append(sessions, &session)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return sessions, nil
}

// Update はセッション情報を更新
func (r *SessionRepository) Update(ctx context.Context, session *models.Session) error {
	query := `
		UPDATE sessions
		SET visitor_id = $2, ip_address = $3, user_agent = $4, referrer = $5,
			country = $6, region = $7, city = $8, latitude = $9, longitude = $10,
			device_type = $11, browser = $12, os = $13, screen_width = $14,
			screen_height = $15, language = $16, timezone = $17, last_activity = $18,
			ended_at = $19, duration_seconds = $20, page_views = $21, is_bounce = $22,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		session.ID,
		session.VisitorID,
		session.IPAddress,
		session.UserAgent,
		session.Referrer,
		session.Country,
		session.Region,
		session.City,
		session.Latitude,
		session.Longitude,
		session.DeviceType,
		session.Browser,
		session.OS,
		session.ScreenWidth,
		session.ScreenHeight,
		session.Language,
		session.Timezone,
		session.LastActivity,
		session.EndedAt,
		session.DurationSeconds,
		session.PageViews,
		session.IsBounce,
	)

	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", session.ID)
	}

	r.logger.WithFields(logrus.Fields{
		"session_id": session.ID,
		"visitor_id": session.VisitorID,
	}).Debug("Updated session")

	return nil
}

// UpdateLastActivity はセッションの最終アクティビティを更新
func (r *SessionRepository) UpdateLastActivity(ctx context.Context, sessionID string, lastActivity time.Time) error {
	query := `UPDATE sessions SET last_activity = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, sessionID, lastActivity)
	if err != nil {
		return fmt.Errorf("failed to update last activity: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	return nil
}

// Delete はセッションを削除
func (r *SessionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM sessions WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", id)
	}

	r.logger.WithFields(logrus.Fields{
		"session_id": id,
	}).Info("Deleted session")

	return nil
}

// DeleteExpired は期限切れのセッションを削除
func (r *SessionRepository) DeleteExpired(ctx context.Context, before time.Time) error {
	query := `DELETE FROM sessions WHERE last_activity < $1 AND ended_at IS NOT NULL`
	
	result, err := r.db.ExecContext(ctx, query, before)
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"before": before,
		"rows_affected": rowsAffected,
	}).Info("Deleted expired sessions")

	return nil
}

// GetSessionStatistics はセッション統計を取得
func (r *SessionRepository) GetSessionStatistics(ctx context.Context, applicationID string, startTime, endTime time.Time) (*interfaces.SessionStatistics, error) {
	// 基本統計クエリ
	basicQuery := `
		SELECT 
			COUNT(*) as total_sessions,
			COUNT(CASE WHEN ended_at IS NULL THEN 1 END) as active_sessions,
			AVG(duration_seconds) as avg_duration_seconds,
			AVG(page_views) as avg_page_views,
			COUNT(CASE WHEN is_bounce = true THEN 1 END) * 100.0 / COUNT(*) as bounce_rate
		FROM sessions
		WHERE application_id = $1 AND started_at BETWEEN $2 AND $3
	`

	var stats interfaces.SessionStatistics
	var avgDuration sql.NullFloat64
	var avgPageViews sql.NullFloat64
	var bounceRate sql.NullFloat64

	err := r.db.QueryRowContext(ctx, basicQuery, applicationID, startTime, endTime).Scan(
		&stats.TotalSessions,
		&stats.ActiveSessions,
		&avgDuration,
		&avgPageViews,
		&bounceRate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get basic session statistics: %w", err)
	}

	if avgDuration.Valid {
		stats.AverageSessionDuration = time.Duration(avgDuration.Float64) * time.Second
	}
	if avgPageViews.Valid {
		stats.AveragePagesPerSession = avgPageViews.Float64
	}
	if bounceRate.Valid {
		stats.BounceRate = bounceRate.Float64
	}

	// リピートビジターと新規ビジターを取得
	visitorQuery := `
		WITH visitor_counts AS (
			SELECT visitor_id, COUNT(*) as session_count
			FROM sessions
			WHERE application_id = $1 AND started_at BETWEEN $2 AND $3
			GROUP BY visitor_id
		)
		SELECT 
			COUNT(CASE WHEN session_count > 1 THEN 1 END) as returning_visitors,
			COUNT(CASE WHEN session_count = 1 THEN 1 END) as new_visitors
		FROM visitor_counts
	`

	err = r.db.QueryRowContext(ctx, visitorQuery, applicationID, startTime, endTime).Scan(
		&stats.ReturningVisitors,
		&stats.NewVisitors,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get visitor statistics: %w", err)
	}

	return &stats, nil
}
