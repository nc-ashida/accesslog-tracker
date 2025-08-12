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

// TrackingService はトラッキングデータのビジネスロジックを担当
type TrackingService struct {
	trackingRepo    interfaces.TrackingRepository
	applicationRepo interfaces.ApplicationRepository
	sessionRepo     interfaces.SessionRepository
	cacheService    interfaces.CacheService
	logger          *logrus.Logger
}

// NewTrackingService は新しいトラッキングサービスを作成
func NewTrackingService(
	trackingRepo interfaces.TrackingRepository,
	applicationRepo interfaces.ApplicationRepository,
	sessionRepo interfaces.SessionRepository,
	cacheService interfaces.CacheService,
	logger *logrus.Logger,
) *TrackingService {
	return &TrackingService{
		trackingRepo:    trackingRepo,
		applicationRepo: applicationRepo,
		sessionRepo:     sessionRepo,
		cacheService:    cacheService,
		logger:          logger,
	}
}

// TrackPageView はページビューをトラッキング
func (s *TrackingService) TrackPageView(ctx context.Context, request *models.TrackingRequest) error {
	// アプリケーションの存在確認
	application, err := s.applicationRepo.GetByAPIKey(ctx, request.APIKey)
	if err != nil {
		return fmt.Errorf("invalid API key: %w", err)
	}

	if !application.IsActive {
		return fmt.Errorf("application is not active")
	}

	// トラッキングデータのバリデーション
	if err := validators.ValidateTrackingRequest(request); err != nil {
		return fmt.Errorf("invalid tracking request: %w", err)
	}

	// セッション管理
	session, err := s.getOrCreateSession(ctx, application.ID, request)
	if err != nil {
		return fmt.Errorf("failed to manage session: %w", err)
	}

	// トラッキングレコードの作成
	tracking := &models.Tracking{
		ID:                        uuid.New().String(),
		ApplicationID:             application.ID,
		SessionID:                 session.ID,
		VisitorID:                 request.VisitorID,
		PageURL:                   request.PageURL,
		PageTitle:                 request.PageTitle,
		Referrer:                  request.Referrer,
		IPAddress:                 request.IPAddress,
		UserAgent:                 request.UserAgent,
		Country:                   request.Country,
		Region:                    request.Region,
		City:                      request.City,
		Latitude:                  request.Latitude,
		Longitude:                 request.Longitude,
		DeviceType:                request.DeviceType,
		Browser:                   request.Browser,
		OS:                        request.OS,
		ScreenWidth:               request.ScreenWidth,
		ScreenHeight:              request.ScreenHeight,
		Language:                  request.Language,
		Timezone:                  request.Timezone,
		LoadTimeMs:                request.LoadTimeMs,
		DOMContentLoadedMs:        request.DOMContentLoadedMs,
		FirstContentfulPaintMs:    request.FirstContentfulPaintMs,
		LargestContentfulPaintMs:  request.LargestContentfulPaintMs,
		FirstInputDelayMs:         request.FirstInputDelayMs,
		CumulativeLayoutShift:     request.CumulativeLayoutShift,
		CustomParameters:          request.CustomParameters,
		Timestamp:                 time.Now(),
	}

	// データベースに保存
	if err := s.trackingRepo.Create(ctx, tracking); err != nil {
		return fmt.Errorf("failed to save tracking data: %w", err)
	}

	// セッション情報を更新
	session.PageViews++
	session.LastActivity = time.Now()
	session.IsBounce = false // 複数ページビューがあるのでバウンスではない

	if err := s.sessionRepo.Update(ctx, session); err != nil {
		s.logger.WithError(err).Warn("Failed to update session")
	}

	// キャッシュに統計データを更新
	s.updateStatisticsCache(ctx, application.ID, tracking)

	s.logger.WithFields(logrus.Fields{
		"tracking_id":    tracking.ID,
		"application_id": application.ID,
		"visitor_id":     request.VisitorID,
		"page_url":       request.PageURL,
	}).Info("Tracked page view")

	return nil
}

// getOrCreateSession はセッションを取得または作成
func (s *TrackingService) getOrCreateSession(ctx context.Context, applicationID string, request *models.TrackingRequest) (*models.Session, error) {
	// 既存のセッションを確認
	sessionKey := interfaces.SessionKey(request.SessionID)
	sessionData, err := s.cacheService.Get(ctx, sessionKey)
	if err == nil {
		// キャッシュからセッションを取得
		session, err := s.sessionRepo.GetByID(ctx, sessionData)
		if err == nil {
			return session, nil
		}
	}

	// 新しいセッションを作成
	session := &models.Session{
		ID:            uuid.New().String(),
		ApplicationID: applicationID,
		VisitorID:     request.VisitorID,
		SessionID:     request.SessionID,
		IPAddress:     request.IPAddress,
		UserAgent:     request.UserAgent,
		Referrer:      request.Referrer,
		Country:       request.Country,
		Region:        request.Region,
		City:          request.City,
		Latitude:      request.Latitude,
		Longitude:     request.Longitude,
		DeviceType:    request.DeviceType,
		Browser:       request.Browser,
		OS:            request.OS,
		ScreenWidth:   request.ScreenWidth,
		ScreenHeight:  request.ScreenHeight,
		Language:      request.Language,
		Timezone:      request.Timezone,
		StartedAt:     time.Now(),
		LastActivity:  time.Now(),
		PageViews:     1,
		IsBounce:      true, // 最初はバウンスとして設定
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// セッションをキャッシュに保存
	if err := s.cacheService.Set(ctx, sessionKey, session.ID, interfaces.DefaultSessionTTL); err != nil {
		s.logger.WithError(err).Warn("Failed to cache session")
	}

	return session, nil
}

// updateStatisticsCache は統計データのキャッシュを更新
func (s *TrackingService) updateStatisticsCache(ctx context.Context, applicationID string, tracking *models.Tracking) {
	// リアルタイム統計を更新
	statsKey := interfaces.StatsKey(applicationID, "realtime")
	
	// ページビューカウンターを増加
	pageViewKey := interfaces.PageViewKey(applicationID)
	if err := s.cacheService.ZIncrBy(ctx, pageViewKey, 1, tracking.PageURL); err != nil {
		s.logger.WithError(err).Warn("Failed to update page view cache")
	}

	// デバイスタイプ統計を更新
	if tracking.DeviceType != "" {
		deviceKey := interfaces.DeviceKey(applicationID)
		if err := s.cacheService.ZIncrBy(ctx, deviceKey, 1, tracking.DeviceType); err != nil {
			s.logger.WithError(err).Warn("Failed to update device cache")
		}
	}

	// 国別統計を更新
	if tracking.Country != "" {
		countryKey := interfaces.CountryKey(applicationID)
		if err := s.cacheService.ZIncrBy(ctx, countryKey, 1, tracking.Country); err != nil {
			s.logger.WithError(err).Warn("Failed to update country cache")
		}
	}

	// リファラー統計を更新
	if tracking.Referrer != "" {
		referrerKey := interfaces.ReferrerKey(applicationID)
		if err := s.cacheService.ZIncrBy(ctx, referrerKey, 1, tracking.Referrer); err != nil {
			s.logger.WithError(err).Warn("Failed to update referrer cache")
		}
	}
}

// GetTrackingData はトラッキングデータを取得
func (s *TrackingService) GetTrackingData(ctx context.Context, applicationID string, limit, offset int) ([]*models.Tracking, error) {
	return s.trackingRepo.GetByApplicationID(ctx, applicationID, limit, offset)
}

// GetTrackingDataByTimeRange は指定された時間範囲のトラッキングデータを取得
func (s *TrackingService) GetTrackingDataByTimeRange(ctx context.Context, applicationID string, startTime, endTime time.Time, limit, offset int) ([]*models.Tracking, error) {
	return s.trackingRepo.GetByTimeRange(ctx, applicationID, startTime, endTime, limit, offset)
}

// GetTrackingDataBySession は指定されたセッションのトラッキングデータを取得
func (s *TrackingService) GetTrackingDataBySession(ctx context.Context, sessionID string) ([]*models.Tracking, error) {
	return s.trackingRepo.GetBySessionID(ctx, sessionID)
}

// GetTrackingStatistics はトラッキング統計を取得
func (s *TrackingService) GetTrackingStatistics(ctx context.Context, applicationID string, startTime, endTime time.Time) (*interfaces.TrackingStatistics, error) {
	// まずキャッシュから取得を試行
	cacheKey := fmt.Sprintf("stats:%s:%d:%d", applicationID, startTime.Unix(), endTime.Unix())
	
	// キャッシュから取得を試行（実装は簡略化）
	// 実際の実装では、キャッシュからJSONを取得してデシリアライズする
	
	// データベースから統計を取得
	stats, err := s.trackingRepo.GetStatistics(ctx, applicationID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get tracking statistics: %w", err)
	}

	// 統計をキャッシュに保存（1時間）
	if err := s.cacheService.Set(ctx, cacheKey, "stats_data", interfaces.DefaultStatsTTL); err != nil {
		s.logger.WithError(err).Warn("Failed to cache statistics")
	}

	return stats, nil
}

// DeleteTrackingData はトラッキングデータを削除
func (s *TrackingService) DeleteTrackingData(ctx context.Context, applicationID string) error {
	return s.trackingRepo.DeleteByApplicationID(ctx, applicationID)
}

// DeleteTrackingDataByTimeRange は指定された時間範囲のトラッキングデータを削除
func (s *TrackingService) DeleteTrackingDataByTimeRange(ctx context.Context, applicationID string, startTime, endTime time.Time) error {
	return s.trackingRepo.DeleteByTimeRange(ctx, applicationID, startTime, endTime)
}
