package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/your-username/accesslog-tracker/internal/domain/models"
	"github.com/your-username/accesslog-tracker/internal/infrastructure/cache/interfaces"
)

// WebhookService はWebhook送信のビジネスロジックを担当
type WebhookService struct {
	cacheService interfaces.CacheService
	httpClient   *http.Client
	logger       *logrus.Logger
}

// NewWebhookService は新しいWebhookサービスを作成
func NewWebhookService(
	cacheService interfaces.CacheService,
	logger *logrus.Logger,
) *WebhookService {
	return &WebhookService{
		cacheService: cacheService,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// SendWebhook はWebhookを送信
func (s *WebhookService) SendWebhook(ctx context.Context, webhook *models.Webhook, payload interface{}) error {
	// Webhookの有効性をチェック
	if !webhook.IsActive {
		return fmt.Errorf("webhook is not active")
	}

	// レート制限をチェック
	if err := s.checkRateLimit(ctx, webhook.ID); err != nil {
		return fmt.Errorf("rate limit exceeded: %w", err)
	}

	// ペイロードをJSONに変換
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	// HTTPリクエストを作成
	req, err := http.NewRequestWithContext(ctx, "POST", webhook.URL, bytes.NewBuffer(payloadJSON))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	// ヘッダーを設定
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "AccessLogTracker/1.0")
	req.Header.Set("X-Webhook-ID", webhook.ID)
	req.Header.Set("X-Webhook-Signature", s.generateSignature(payloadJSON, webhook.Secret))

	// カスタムヘッダーを追加
	for key, value := range webhook.Headers {
		req.Header.Set(key, value)
	}

	// Webhookを送信
	startTime := time.Now()
	resp, err := s.httpClient.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		s.logWebhookFailure(webhook, err, duration)
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	// レスポンスをログに記録
	s.logWebhookResponse(webhook, resp, duration)

	// 成功した場合、レート制限カウンターを更新
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		s.updateRateLimit(ctx, webhook.ID)
	}

	// エラーレスポンスの場合はエラーを返す
	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned error status: %d", resp.StatusCode)
	}

	return nil
}

// SendTrackingWebhook はトラッキングイベントのWebhookを送信
func (s *WebhookService) SendTrackingWebhook(ctx context.Context, webhook *models.Webhook, tracking *models.Tracking) error {
	payload := &TrackingWebhookPayload{
		Event:       "page_view",
		Timestamp:   time.Now(),
		Application: webhook.ApplicationID,
		Tracking:    tracking,
	}

	return s.SendWebhook(ctx, webhook, payload)
}

// SendSessionWebhook はセッションイベントのWebhookを送信
func (s *WebhookService) SendSessionWebhook(ctx context.Context, webhook *models.Webhook, session *models.Session, event string) error {
	payload := &SessionWebhookPayload{
		Event:       event,
		Timestamp:   time.Now(),
		Application: webhook.ApplicationID,
		Session:     session,
	}

	return s.SendWebhook(ctx, webhook, payload)
}

// SendStatisticsWebhook は統計イベントのWebhookを送信
func (s *WebhookService) SendStatisticsWebhook(ctx context.Context, webhook *models.Webhook, stats interface{}, period string) error {
	payload := &StatisticsWebhookPayload{
		Event:       "statistics_updated",
		Timestamp:   time.Now(),
		Application: webhook.ApplicationID,
		Period:      period,
		Statistics:  stats,
	}

	return s.SendWebhook(ctx, webhook, payload)
}

// SendApplicationWebhook はアプリケーションイベントのWebhookを送信
func (s *WebhookService) SendApplicationWebhook(ctx context.Context, webhook *models.Webhook, application *models.Application, event string) error {
	payload := &ApplicationWebhookPayload{
		Event:        event,
		Timestamp:    time.Now(),
		Application:  webhook.ApplicationID,
		AppData:      application,
	}

	return s.SendWebhook(ctx, webhook, payload)
}

// checkRateLimit はレート制限をチェック
func (s *WebhookService) checkRateLimit(ctx context.Context, webhookID string) error {
	rateLimitKey := interfaces.RateLimitKey(fmt.Sprintf("webhook:%s", webhookID))
	
	// 現在のカウントを取得
	currentCount, err := s.cacheService.Get(ctx, rateLimitKey)
	if err != nil {
		// キーが存在しない場合は初回なので許可
		return nil
	}

	// カウントを数値に変換
	var count int64
	if _, err := fmt.Sscanf(currentCount, "%d", &count); err != nil {
		s.logger.WithError(err).Warn("Failed to parse rate limit count")
		return nil
	}

	// 1分間に最大60回まで許可
	if count >= 60 {
		return fmt.Errorf("rate limit exceeded: %d requests per minute", count)
	}

	return nil
}

// updateRateLimit はレート制限カウンターを更新
func (s *WebhookService) updateRateLimit(ctx context.Context, webhookID string) {
	rateLimitKey := interfaces.RateLimitKey(fmt.Sprintf("webhook:%s", webhookID))
	
	// カウンターを増加
	if _, err := s.cacheService.Incr(ctx, rateLimitKey); err != nil {
		s.logger.WithError(err).Warn("Failed to increment rate limit counter")
		return
	}

	// 1分間の有効期限を設定（初回の場合）
	if err := s.cacheService.Expire(ctx, rateLimitKey, interfaces.DefaultRateLimitTTL); err != nil {
		s.logger.WithError(err).Warn("Failed to set rate limit expiration")
	}
}

// generateSignature はWebhook署名を生成
func (s *WebhookService) generateSignature(payload []byte, secret string) string {
	// 実際の実装では、HMAC-SHA256などを使用して署名を生成
	// ここでは簡略化のため、シンプルなハッシュを使用
	return fmt.Sprintf("sha256=%x", payload)
}

// logWebhookFailure はWebhook失敗をログに記録
func (s *WebhookService) logWebhookFailure(webhook *models.Webhook, err error, duration time.Duration) {
	s.logger.WithFields(logrus.Fields{
		"webhook_id": webhook.ID,
		"url":        webhook.URL,
		"error":      err.Error(),
		"duration":   duration,
	}).Error("Webhook delivery failed")
}

// logWebhookResponse はWebhookレスポンスをログに記録
func (s *WebhookService) logWebhookResponse(webhook *models.Webhook, resp *http.Response, duration time.Duration) {
	level := logrus.InfoLevel
	if resp.StatusCode >= 400 {
		level = logrus.ErrorLevel
	}

	s.logger.WithFields(logrus.Fields{
		"webhook_id":  webhook.ID,
		"url":         webhook.URL,
		"status_code": resp.StatusCode,
		"duration":    duration,
	}).Log(level, "Webhook delivered")
}

// ValidateWebhookURL はWebhook URLの有効性を検証
func (s *WebhookService) ValidateWebhookURL(ctx context.Context, url string) error {
	// 簡易的なURL検証
	if url == "" {
		return fmt.Errorf("webhook URL cannot be empty")
	}

	// HTTP/HTTPSプロトコルのチェック
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("webhook URL must use HTTP or HTTPS protocol")
	}

	// 実際の実装では、URLにHEADリクエストを送信して到達可能性をチェック
	// ここでは簡略化のため、基本的な形式チェックのみ

	return nil
}

// TestWebhook はWebhookのテスト送信を実行
func (s *WebhookService) TestWebhook(ctx context.Context, webhook *models.Webhook) error {
	testPayload := &TestWebhookPayload{
		Event:     "test",
		Timestamp: time.Now(),
		Message:   "This is a test webhook from Access Log Tracker",
		Data: map[string]interface{}{
			"webhook_id": webhook.ID,
			"url":        webhook.URL,
			"test_time":  time.Now().Format(time.RFC3339),
		},
	}

	return s.SendWebhook(ctx, webhook, testPayload)
}

// Webhookペイロード構造体
type TrackingWebhookPayload struct {
	Event       string           `json:"event"`
	Timestamp   time.Time        `json:"timestamp"`
	Application string           `json:"application"`
	Tracking    *models.Tracking `json:"tracking"`
}

type SessionWebhookPayload struct {
	Event       string         `json:"event"`
	Timestamp   time.Time      `json:"timestamp"`
	Application string         `json:"application"`
	Session     *models.Session `json:"session"`
}

type StatisticsWebhookPayload struct {
	Event       string      `json:"event"`
	Timestamp   time.Time   `json:"timestamp"`
	Application string      `json:"application"`
	Period      string      `json:"period"`
	Statistics  interface{} `json:"statistics"`
}

type ApplicationWebhookPayload struct {
	Event       string              `json:"event"`
	Timestamp   time.Time           `json:"timestamp"`
	Application string              `json:"application"`
	AppData     *models.Application `json:"app_data"`
}

type TestWebhookPayload struct {
	Event     string                 `json:"event"`
	Timestamp time.Time              `json:"timestamp"`
	Message   string                 `json:"message"`
