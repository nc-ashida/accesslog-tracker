package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nc-ashida/accesslog-tracker/internal/domain/models"
	"github.com/nc-ashida/accesslog-tracker/internal/domain/validators"
	"github.com/nc-ashida/accesslog-tracker/internal/utils/iputil"
	"github.com/nc-ashida/accesslog-tracker/internal/utils/timeutil"
)

// TrackingRepository はトラッキングリポジトリのインターフェースです
type TrackingRepository interface {
	Create(ctx context.Context, data *models.TrackingData) error
	GetByID(ctx context.Context, id string) (*models.TrackingData, error)
	GetByAppID(ctx context.Context, appID string, limit, offset int) ([]*models.TrackingData, error)
	GetBySessionID(ctx context.Context, sessionID string) ([]*models.TrackingData, error)
	CountByAppID(ctx context.Context, appID string) (int64, error)
	Delete(ctx context.Context, id string) error
}

// TrackingService はトラッキングのビジネスロジックを提供します
type TrackingService struct {
	repo      TrackingRepository
	validator *validators.TrackingValidator
}

// NewTrackingService は新しいトラッキングサービスを作成します
func NewTrackingService(repo TrackingRepository) *TrackingService {
	return &TrackingService{
		repo:      repo,
		validator: validators.NewTrackingValidator(),
	}
}

// ProcessTrackingData はトラッキングデータを処理します
func (s *TrackingService) ProcessTrackingData(ctx context.Context, data *models.TrackingData) error {
	// バリデーション
	if err := s.validator.Validate(data); err != nil {
		return err
	}

	// カスタムパラメータのバリデーション
	if err := s.validator.ValidateCustomParams(data.CustomParams); err != nil {
		return err
	}

	// IDの生成
	if data.ID == "" {
		data.ID = uuid.New().String()
	}

	// タイムスタンプの設定
	if data.Timestamp.IsZero() {
		data.Timestamp = time.Now()
	}

	// IPアドレスの匿名化
	if data.IPAddress != "" {
		data.IPAddress = iputil.AnonymizeIP(data.IPAddress)
	}

	// セッションIDの生成（存在しない場合）
	if data.SessionID == "" {
		data.SessionID = s.generateSessionID(data)
	}

	// 作成時刻の設定
	data.CreatedAt = time.Now()

	// リポジトリに保存
	return s.repo.Create(ctx, data)
}

// GetByID はIDでトラッキングデータを取得します
func (s *TrackingService) GetByID(ctx context.Context, id string) (*models.TrackingData, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByAppID はアプリケーションIDでトラッキングデータ一覧を取得します
func (s *TrackingService) GetByAppID(ctx context.Context, appID string, limit, offset int) ([]*models.TrackingData, error) {
	return s.repo.GetByAppID(ctx, appID, limit, offset)
}

// GetBySessionID はセッションIDでトラッキングデータ一覧を取得します
func (s *TrackingService) GetBySessionID(ctx context.Context, sessionID string) ([]*models.TrackingData, error) {
	return s.repo.GetBySessionID(ctx, sessionID)
}

// CountByAppID はアプリケーションIDでトラッキングデータ数を取得します
func (s *TrackingService) CountByAppID(ctx context.Context, appID string) (int64, error) {
	return s.repo.CountByAppID(ctx, appID)
}

// Delete はトラッキングデータを削除します
func (s *TrackingService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// GetStatistics はトラッキング統計を取得します
func (s *TrackingService) GetStatistics(ctx context.Context, appID string, startDate, endDate time.Time) (*TrackingStatistics, error) {
	// 日付範囲の検証
	if startDate.After(endDate) {
		return nil, models.ErrStatisticsInvalidPeriod
	}

	// 統計データの取得
	stats := &TrackingStatistics{
		AppID:     appID,
		StartDate: startDate,
		EndDate:   endDate,
		Metrics:   make(map[string]interface{}),
	}

	// 基本的な統計情報を計算
	if err := s.calculateBasicStatistics(ctx, stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// TrackingStatistics はトラッキング統計を表します
type TrackingStatistics struct {
	AppID     string                 `json:"app_id"`
	StartDate time.Time              `json:"start_date"`
	EndDate   time.Time              `json:"end_date"`
	Metrics   map[string]interface{} `json:"metrics"`
}

// calculateBasicStatistics は基本的な統計情報を計算します
func (s *TrackingService) calculateBasicStatistics(ctx context.Context, stats *TrackingStatistics) error {
	// ここでは簡易的な実装
	// 実際の実装では、データベースから集計クエリを実行

	// 総トラッキング数
	totalCount, err := s.repo.CountByAppID(ctx, stats.AppID)
	if err != nil {
		return err
	}

	stats.Metrics["total_tracking_count"] = totalCount
	stats.Metrics["period_days"] = int(stats.EndDate.Sub(stats.StartDate).Hours() / 24)

	return nil
}

// generateSessionID はセッションIDを生成します
func (s *TrackingService) generateSessionID(data *models.TrackingData) string {
	// ユーザーエージェントとIPアドレスからハッシュを生成
	seed := data.UserAgent + data.IPAddress + data.AppID
	return uuid.NewSHA1(uuid.Nil, []byte(seed)).String()
}

// IsValidTrackingData はトラッキングデータが有効かどうかを判定します
func (s *TrackingService) IsValidTrackingData(data *models.TrackingData) bool {
	return s.validator.Validate(data) == nil
}

// GetTrackingDataByDateRange は日付範囲でトラッキングデータを取得します
func (s *TrackingService) GetTrackingDataByDateRange(ctx context.Context, appID string, startDate, endDate time.Time, limit, offset int) ([]*models.TrackingData, error) {
	// 日付範囲の検証
	if startDate.After(endDate) {
		return nil, models.ErrStatisticsInvalidPeriod
	}

	// 実際の実装では、データベースで日付範囲フィルタリングを実行
	// ここでは簡易的な実装として、全データを取得
	return s.repo.GetByAppID(ctx, appID, limit, offset)
}

// GetDailyStatistics は日別統計を取得します
func (s *TrackingService) GetDailyStatistics(ctx context.Context, appID string, date time.Time) (*DailyStatistics, error) {
	startOfDay := timeutil.GetStartOfDay(date)
	endOfDay := timeutil.GetEndOfDay(date)

	stats := &DailyStatistics{
		AppID: appID,
		Date:  date,
	}

	// 日別統計の計算
	if err := s.calculateDailyStatistics(ctx, stats, startOfDay, endOfDay); err != nil {
		return nil, err
	}

	return stats, nil
}

// DailyStatistics は日別統計を表します
type DailyStatistics struct {
	AppID           string    `json:"app_id"`
	Date            time.Time `json:"date"`
	TotalSessions   int64     `json:"total_sessions"`
	TotalPageViews  int64     `json:"total_page_views"`
	UniqueVisitors  int64     `json:"unique_visitors"`
	AverageSession  float64   `json:"average_session_duration"`
}

// calculateDailyStatistics は日別統計を計算します
func (s *TrackingService) calculateDailyStatistics(ctx context.Context, stats *DailyStatistics, startDate, endDate time.Time) error {
	// 実際の実装では、データベースから集計クエリを実行
	// ここでは簡易的な実装

	// 総トラッキング数
	totalCount, err := s.repo.CountByAppID(ctx, stats.AppID)
	if err != nil {
		return err
	}

	stats.TotalPageViews = totalCount
	stats.TotalSessions = totalCount / 10 // 簡易的な計算
	stats.UniqueVisitors = totalCount / 5  // 簡易的な計算
	stats.AverageSession = 300.0           // 5分（秒単位）

	return nil
}
