package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/your-username/accesslog-tracker/internal/infrastructure/cache/interfaces"
	"github.com/your-username/accesslog-tracker/internal/infrastructure/database/interfaces"
)

// StatisticsService は統計データ生成のビジネスロジックを担当
type StatisticsService struct {
	trackingRepo    interfaces.TrackingRepository
	sessionRepo     interfaces.SessionRepository
	cacheService    interfaces.CacheService
	logger          *logrus.Logger
}

// NewStatisticsService は新しい統計サービスを作成
func NewStatisticsService(
	trackingRepo interfaces.TrackingRepository,
	sessionRepo interfaces.SessionRepository,
	cacheService interfaces.CacheService,
	logger *logrus.Logger,
) *StatisticsService {
	return &StatisticsService{
		trackingRepo: trackingRepo,
		sessionRepo:  sessionRepo,
		cacheService: cacheService,
		logger:       logger,
	}
}

// GetRealTimeStatistics はリアルタイム統計を取得
func (s *StatisticsService) GetRealTimeStatistics(ctx context.Context, applicationID string) (*RealTimeStatistics, error) {
	// キャッシュからリアルタイム統計を取得
	cacheKey := interfaces.StatsKey(applicationID, "realtime")
	
	// 現在時刻から過去1時間の統計を取得
	endTime := time.Now()
	startTime := endTime.Add(-1 * time.Hour)

	// トラッキング統計を取得
	trackingStats, err := s.trackingRepo.GetStatistics(ctx, applicationID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get tracking statistics: %w", err)
	}

	// セッション統計を取得
	sessionStats, err := s.sessionRepo.GetSessionStatistics(ctx, applicationID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get session statistics: %w", err)
	}

	// リアルタイム統計を構築
	realTimeStats := &RealTimeStatistics{
		ApplicationID:           applicationID,
		Period:                  "1h",
		TotalRequests:           trackingStats.TotalRequests,
		UniqueVisitors:          trackingStats.UniqueVisitors,
		UniqueSessions:          trackingStats.UniqueSessions,
		ActiveSessions:          sessionStats.ActiveSessions,
		AverageSessionDuration:  sessionStats.AverageSessionDuration,
		AveragePagesPerSession:  sessionStats.AveragePagesPerSession,
		BounceRate:              sessionStats.BounceRate,
		TopPages:                trackingStats.TopPages,
		TopReferrers:            trackingStats.TopReferrers,
		TopUserAgents:           trackingStats.TopUserAgents,
		TopCountries:            trackingStats.TopCountries,
		GeneratedAt:             time.Now(),
	}

	// 統計をキャッシュに保存（5分間）
	statsJSON, err := json.Marshal(realTimeStats)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to marshal real-time statistics")
	} else {
		if err := s.cacheService.Set(ctx, cacheKey, string(statsJSON), 5*time.Minute); err != nil {
			s.logger.WithError(err).Warn("Failed to cache real-time statistics")
		}
	}

	return realTimeStats, nil
}

// GetDailyStatistics は日次統計を取得
func (s *StatisticsService) GetDailyStatistics(ctx context.Context, applicationID string, date time.Time) (*DailyStatistics, error) {
	// 指定日の開始と終了時刻を計算
	startTime := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endTime := startTime.Add(24 * time.Hour)

	// キャッシュキーを生成
	cacheKey := fmt.Sprintf("stats:%s:daily:%s", applicationID, date.Format("2006-01-02"))

	// キャッシュから取得を試行
	if cachedData, err := s.cacheService.Get(ctx, cacheKey); err == nil {
		var stats DailyStatistics
		if err := json.Unmarshal([]byte(cachedData), &stats); err == nil {
			return &stats, nil
		}
	}

	// トラッキング統計を取得
	trackingStats, err := s.trackingRepo.GetStatistics(ctx, applicationID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get tracking statistics: %w", err)
	}

	// セッション統計を取得
	sessionStats, err := s.sessionRepo.GetSessionStatistics(ctx, applicationID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get session statistics: %w", err)
	}

	// 時間別統計を取得
	hourlyStats, err := s.getHourlyStatistics(ctx, applicationID, startTime, endTime)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get hourly statistics")
	}

	// 日次統計を構築
	dailyStats := &DailyStatistics{
		ApplicationID:           applicationID,
		Date:                    date,
		TotalRequests:           trackingStats.TotalRequests,
		UniqueVisitors:          trackingStats.UniqueVisitors,
		UniqueSessions:          trackingStats.UniqueSessions,
		ActiveSessions:          sessionStats.ActiveSessions,
		AverageSessionDuration:  sessionStats.AverageSessionDuration,
		AveragePagesPerSession:  sessionStats.AveragePagesPerSession,
		BounceRate:              sessionStats.BounceRate,
		ReturningVisitors:       sessionStats.ReturningVisitors,
		NewVisitors:             sessionStats.NewVisitors,
		TopPages:                trackingStats.TopPages,
		TopReferrers:            trackingStats.TopReferrers,
		TopUserAgents:           trackingStats.TopUserAgents,
		TopCountries:            trackingStats.TopCountries,
		HourlyStats:             hourlyStats,
		GeneratedAt:             time.Now(),
	}

	// 統計をキャッシュに保存（1時間）
	statsJSON, err := json.Marshal(dailyStats)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to marshal daily statistics")
	} else {
		if err := s.cacheService.Set(ctx, cacheKey, string(statsJSON), interfaces.DefaultStatsTTL); err != nil {
			s.logger.WithError(err).Warn("Failed to cache daily statistics")
		}
	}

	return dailyStats, nil
}

// GetWeeklyStatistics は週次統計を取得
func (s *StatisticsService) GetWeeklyStatistics(ctx context.Context, applicationID string, startDate time.Time) (*WeeklyStatistics, error) {
	// 週の開始と終了時刻を計算
	weekStart := startDate
	weekEnd := weekStart.AddDate(0, 0, 7)

	// キャッシュキーを生成
	cacheKey := fmt.Sprintf("stats:%s:weekly:%s", applicationID, startDate.Format("2006-01-02"))

	// キャッシュから取得を試行
	if cachedData, err := s.cacheService.Get(ctx, cacheKey); err == nil {
		var stats WeeklyStatistics
		if err := json.Unmarshal([]byte(cachedData), &stats); err == nil {
			return &stats, nil
		}
	}

	// トラッキング統計を取得
	trackingStats, err := s.trackingRepo.GetStatistics(ctx, applicationID, weekStart, weekEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get tracking statistics: %w", err)
	}

	// セッション統計を取得
	sessionStats, err := s.sessionRepo.GetSessionStatistics(ctx, applicationID, weekStart, weekEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get session statistics: %w", err)
	}

	// 日別統計を取得
	dailyStats, err := s.getDailyStatisticsForWeek(ctx, applicationID, weekStart, weekEnd)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get daily statistics for week")
	}

	// 週次統計を構築
	weeklyStats := &WeeklyStatistics{
		ApplicationID:           applicationID,
		WeekStart:               weekStart,
		WeekEnd:                 weekEnd,
		TotalRequests:           trackingStats.TotalRequests,
		UniqueVisitors:          trackingStats.UniqueVisitors,
		UniqueSessions:          trackingStats.UniqueSessions,
		ActiveSessions:          sessionStats.ActiveSessions,
		AverageSessionDuration:  sessionStats.AverageSessionDuration,
		AveragePagesPerSession:  sessionStats.AveragePagesPerSession,
		BounceRate:              sessionStats.BounceRate,
		ReturningVisitors:       sessionStats.ReturningVisitors,
		NewVisitors:             sessionStats.NewVisitors,
		TopPages:                trackingStats.TopPages,
		TopReferrers:            trackingStats.TopReferrers,
		TopUserAgents:           trackingStats.TopUserAgents,
		TopCountries:            trackingStats.TopCountries,
		DailyStats:              dailyStats,
		GeneratedAt:             time.Now(),
	}

	// 統計をキャッシュに保存（6時間）
	statsJSON, err := json.Marshal(weeklyStats)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to marshal weekly statistics")
	} else {
		if err := s.cacheService.Set(ctx, cacheKey, string(statsJSON), 6*time.Hour); err != nil {
			s.logger.WithError(err).Warn("Failed to cache weekly statistics")
		}
	}

	return weeklyStats, nil
}

// GetMonthlyStatistics は月次統計を取得
func (s *StatisticsService) GetMonthlyStatistics(ctx context.Context, applicationID string, year int, month time.Month) (*MonthlyStatistics, error) {
	// 月の開始と終了時刻を計算
	monthStart := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)

	// キャッシュキーを生成
	cacheKey := fmt.Sprintf("stats:%s:monthly:%d-%02d", applicationID, year, month)

	// キャッシュから取得を試行
	if cachedData, err := s.cacheService.Get(ctx, cacheKey); err == nil {
		var stats MonthlyStatistics
		if err := json.Unmarshal([]byte(cachedData), &stats); err == nil {
			return &stats, nil
		}
	}

	// トラッキング統計を取得
	trackingStats, err := s.trackingRepo.GetStatistics(ctx, applicationID, monthStart, monthEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get tracking statistics: %w", err)
	}

	// セッション統計を取得
	sessionStats, err := s.sessionRepo.GetSessionStatistics(ctx, applicationID, monthStart, monthEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get session statistics: %w", err)
	}

	// 週別統計を取得
	weeklyStats, err := s.getWeeklyStatisticsForMonth(ctx, applicationID, monthStart, monthEnd)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get weekly statistics for month")
	}

	// 月次統計を構築
	monthlyStats := &MonthlyStatistics{
		ApplicationID:           applicationID,
		Year:                    year,
		Month:                   month,
		TotalRequests:           trackingStats.TotalRequests,
		UniqueVisitors:          trackingStats.UniqueVisitors,
		UniqueSessions:          trackingStats.UniqueSessions,
		ActiveSessions:          sessionStats.ActiveSessions,
		AverageSessionDuration:  sessionStats.AverageSessionDuration,
		AveragePagesPerSession:  sessionStats.AveragePagesPerSession,
		BounceRate:              sessionStats.BounceRate,
		ReturningVisitors:       sessionStats.ReturningVisitors,
		NewVisitors:             sessionStats.NewVisitors,
		TopPages:                trackingStats.TopPages,
		TopReferrers:            trackingStats.TopReferrers,
		TopUserAgents:           trackingStats.TopUserAgents,
		TopCountries:            trackingStats.TopCountries,
		WeeklyStats:             weeklyStats,
		GeneratedAt:             time.Now(),
	}

	// 統計をキャッシュに保存（24時間）
	statsJSON, err := json.Marshal(monthlyStats)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to marshal monthly statistics")
	} else {
		if err := s.cacheService.Set(ctx, cacheKey, string(statsJSON), 24*time.Hour); err != nil {
			s.logger.WithError(err).Warn("Failed to cache monthly statistics")
		}
	}

	return monthlyStats, nil
}

// getHourlyStatistics は時間別統計を取得
func (s *StatisticsService) getHourlyStatistics(ctx context.Context, applicationID string, startTime, endTime time.Time) ([]HourlyStat, error) {
	var hourlyStats []HourlyStat

	for hour := 0; hour < 24; hour++ {
		hourStart := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), hour, 0, 0, 0, startTime.Location())
		hourEnd := hourStart.Add(1 * time.Hour)

		if hourStart.Before(startTime) || hourStart.After(endTime) {
			continue
		}

		// 1時間の統計を取得
		stats, err := s.trackingRepo.GetStatistics(ctx, applicationID, hourStart, hourEnd)
		if err != nil {
			s.logger.WithError(err).Warnf("Failed to get statistics for hour %d", hour)
			continue
		}

		hourlyStats = append(hourlyStats, HourlyStat{
			Hour:           hour,
			TotalRequests:  stats.TotalRequests,
			UniqueVisitors: stats.UniqueVisitors,
			UniqueSessions: stats.UniqueSessions,
		})
	}

	return hourlyStats, nil
}

// getDailyStatisticsForWeek は週の日別統計を取得
func (s *StatisticsService) getDailyStatisticsForWeek(ctx context.Context, applicationID string, weekStart, weekEnd time.Time) ([]DailyStat, error) {
	var dailyStats []DailyStat

	for day := 0; day < 7; day++ {
		dayStart := weekStart.AddDate(0, 0, day)
		dayEnd := dayStart.Add(24 * time.Hour)

		if dayStart.After(weekEnd) {
			break
		}

		// 1日の統計を取得
		stats, err := s.trackingRepo.GetStatistics(ctx, applicationID, dayStart, dayEnd)
		if err != nil {
			s.logger.WithError(err).Warnf("Failed to get statistics for day %d", day)
			continue
		}

		dailyStats = append(dailyStats, DailyStat{
			Date:           dayStart,
			TotalRequests:  stats.TotalRequests,
			UniqueVisitors: stats.UniqueVisitors,
			UniqueSessions: stats.UniqueSessions,
		})
	}

	return dailyStats, nil
}

// getWeeklyStatisticsForMonth は月の週別統計を取得
func (s *StatisticsService) getWeeklyStatisticsForMonth(ctx context.Context, applicationID string, monthStart, monthEnd time.Time) ([]WeeklyStat, error) {
	var weeklyStats []WeeklyStat

	// 月の最初の週の開始日を計算
	weekStart := monthStart
	for weekStart.Weekday() != time.Monday {
		weekStart = weekStart.AddDate(0, 0, -1)
	}

	for weekStart.Before(monthEnd) {
		weekEnd := weekStart.AddDate(0, 0, 7)

		// 1週間の統計を取得
		stats, err := s.trackingRepo.GetStatistics(ctx, applicationID, weekStart, weekEnd)
		if err != nil {
			s.logger.WithError(err).Warnf("Failed to get statistics for week starting %s", weekStart.Format("2006-01-02"))
			weekStart = weekEnd
			continue
		}

		weeklyStats = append(weeklyStats, WeeklyStat{
			WeekStart:      weekStart,
			WeekEnd:        weekEnd,
			TotalRequests:  stats.TotalRequests,
			UniqueVisitors: stats.UniqueVisitors,
			UniqueSessions: stats.UniqueSessions,
		})

		weekStart = weekEnd
	}

	return weeklyStats, nil
}

// ClearStatisticsCache は統計キャッシュをクリア
func (s *StatisticsService) ClearStatisticsCache(ctx context.Context, applicationID string) error {
	pattern := fmt.Sprintf("stats:%s:*", applicationID)
	keys, err := s.cacheService.Keys(ctx, pattern)
	if err != nil {
		return fmt.Errorf("failed to get cache keys: %w", err)
	}

	for _, key := range keys {
		if err := s.cacheService.Delete(ctx, key); err != nil {
			s.logger.WithError(err).Warnf("Failed to delete cache key: %s", key)
		}
	}

	s.logger.WithField("application_id", applicationID).Info("Cleared statistics cache")
	return nil
}

// 統計データ構造体
type RealTimeStatistics struct {
	ApplicationID           string                    `json:"application_id"`
	Period                  string                    `json:"period"`
	TotalRequests           int64                     `json:"total_requests"`
	UniqueVisitors          int64                     `json:"unique_visitors"`
	UniqueSessions          int64                     `json:"unique_sessions"`
	ActiveSessions          int64                     `json:"active_sessions"`
	AverageSessionDuration  time.Duration             `json:"average_session_duration"`
	AveragePagesPerSession  float64                   `json:"average_pages_per_session"`
	BounceRate              float64                   `json:"bounce_rate"`
	TopPages                []interfaces.PageView     `json:"top_pages"`
	TopReferrers            []interfaces.Referrer     `json:"top_referrers"`
	TopUserAgents           []interfaces.UserAgent    `json:"top_user_agents"`
	TopCountries            []interfaces.Country      `json:"top_countries"`
	GeneratedAt             time.Time                 `json:"generated_at"`
}

type DailyStatistics struct {
	ApplicationID           string                    `json:"application_id"`
	Date                    time.Time                 `json:"date"`
	TotalRequests           int64                     `json:"total_requests"`
	UniqueVisitors          int64                     `json:"unique_visitors"`
	UniqueSessions          int64                     `json:"unique_sessions"`
	ActiveSessions          int64                     `json:"active_sessions"`
	AverageSessionDuration  time.Duration             `json:"average_session_duration"`
	AveragePagesPerSession  float64                   `json:"average_pages_per_session"`
	BounceRate              float64                   `json:"bounce_rate"`
	ReturningVisitors       int64                     `json:"returning_visitors"`
	NewVisitors             int64                     `json:"new_visitors"`
	TopPages                []interfaces.PageView     `json:"top_pages"`
	TopReferrers            []interfaces.Referrer     `json:"top_referrers"`
	TopUserAgents           []interfaces.UserAgent    `json:"top_user_agents"`
	TopCountries            []interfaces.Country      `json:"top_countries"`
	HourlyStats             []HourlyStat              `json:"hourly_stats"`
	GeneratedAt             time.Time                 `json:"generated_at"`
}

type WeeklyStatistics struct {
	ApplicationID           string                    `json:"application_id"`
	WeekStart               time.Time                 `json:"week_start"`
	WeekEnd                 time.Time                 `json:"week_end"`
	TotalRequests           int64                     `json:"total_requests"`
	UniqueVisitors          int64                     `json:"unique_visitors"`
	UniqueSessions          int64                     `json:"unique_sessions"`
	ActiveSessions          int64                     `json:"active_sessions"`
	AverageSessionDuration  time.Duration             `json:"average_session_duration"`
	AveragePagesPerSession  float64                   `json:"average_pages_per_session"`
	BounceRate              float64                   `json:"bounce_rate"`
	ReturningVisitors       int64                     `json:"returning_visitors"`
	NewVisitors             int64                     `json:"new_visitors"`
	TopPages                []interfaces.PageView     `json:"top_pages"`
	TopReferrers            []interfaces.Referrer     `json:"top_referrers"`
	TopUserAgents           []interfaces.UserAgent    `json:"top_user_agents"`
	TopCountries            []interfaces.Country      `json:"top_countries"`
	DailyStats              []DailyStat               `json:"daily_stats"`
	GeneratedAt             time.Time                 `json:"generated_at"`
}

type MonthlyStatistics struct {
	ApplicationID           string                    `json:"application_id"`
	Year                    int                       `json:"year"`
	Month                   time.Month                `json:"month"`
	TotalRequests           int64                     `json:"total_requests"`
	UniqueVisitors          int64                     `json:"unique_visitors"`
	UniqueSessions          int64                     `json:"unique_sessions"`
	ActiveSessions          int64                     `json:"active_sessions"`
	AverageSessionDuration  time.Duration             `json:"average_session_duration"`
	AveragePagesPerSession  float64                   `json:"average_pages_per_session"`
	BounceRate              float64                   `json:"bounce_rate"`
	ReturningVisitors       int64                     `json:"returning_visitors"`
	NewVisitors             int64                     `json:"new_visitors"`
	TopPages                []interfaces.PageView     `json:"top_pages"`
	TopReferrers            []interfaces.Referrer     `json:"top_referrers"`
	TopUserAgents           []interfaces.UserAgent    `json:"top_user_agents"`
	TopCountries            []interfaces.Country      `json:"top_countries"`
	WeeklyStats             []WeeklyStat              `json:"weekly_stats"`
	GeneratedAt             time.Time                 `json:"generated_at"`
}

type HourlyStat struct {
	Hour           int   `json:"hour"`
	TotalRequests  int64 `json:"total_requests"`
	UniqueVisitors int64 `json:"unique_visitors"`
	UniqueSessions int64 `json:"unique_sessions"`
}

type DailyStat struct {
	Date           time.Time `json:"date"`
	TotalRequests  int64     `json:"total_requests"`
	UniqueVisitors int64     `json:"unique_visitors"`
	UniqueSessions int64     `json:"unique_sessions"`
}

type WeeklyStat struct {
	WeekStart      time.Time `json:"week_start"`
	WeekEnd        time.Time `json:"week_end"`
	TotalRequests  int64     `json:"total_requests"`
	UniqueVisitors int64     `json:"unique_visitors"`
	UniqueSessions int64     `json:"unique_sessions"`
}
