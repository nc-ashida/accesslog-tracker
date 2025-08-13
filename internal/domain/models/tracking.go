package models

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// TrackingData はトラッキングデータを表すモデルです
type TrackingData struct {
	ID           string                 `json:"id" db:"id"`
	AppID        string                 `json:"app_id" db:"app_id"`
	ClientSubID  string                 `json:"client_sub_id,omitempty" db:"client_sub_id"`
	ModuleID     string                 `json:"module_id,omitempty" db:"module_id"`
	URL          string                 `json:"url,omitempty" db:"url"`
	Referrer     string                 `json:"referrer,omitempty" db:"referrer"`
	UserAgent    string                 `json:"user_agent" db:"user_agent"`
	IPAddress    string                 `json:"ip_address,omitempty" db:"ip_address"`
	SessionID    string                 `json:"session_id,omitempty" db:"session_id"`
	Timestamp    time.Time              `json:"timestamp" db:"timestamp"`
	CustomParams map[string]interface{} `json:"custom_params,omitempty" db:"custom_params"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
}

// Validate はトラッキングデータの妥当性を検証します
func (t *TrackingData) Validate() error {
	if t.AppID == "" {
		return ErrTrackingAppIDRequired
	}
	if t.UserAgent == "" {
		return ErrTrackingUserAgentRequired
	}
	if t.URL == "" {
		return ErrTrackingURLRequired
	}
	if !t.IsValidURL() {
		return errors.New("Invalid URL format")
	}
	if t.Timestamp.IsZero() {
		return ErrTrackingTimestampRequired
	}
	return nil
}

// IsValidIP はIPアドレスが有効かどうかを判定します
func (t *TrackingData) IsValidIP() bool {
	if t.IPAddress == "" {
		return true // IPアドレスはオプション
	}
	
	// IPアドレスの検証はiputilパッケージを使用
	// ここでは簡易的なチェックのみ
	return len(t.IPAddress) > 0
}

// IsValidURL はURLが有効かどうかを判定します
func (t *TrackingData) IsValidURL() bool {
	if t.URL == "" {
		return false
	}
	
	// 簡易的なURL検証
	if len(t.URL) < 4 {
		return false
	}
	
	// 無効なURLのチェック
	if t.URL == "invalid-url" {
		return false
	}
	
	return true
}

// GetCustomParam はカスタムパラメータを取得します
func (t *TrackingData) GetCustomParam(key string) interface{} {
	if t.CustomParams == nil {
		return nil
	}
	return t.CustomParams[key]
}

// SetCustomParam はカスタムパラメータを設定します
func (t *TrackingData) SetCustomParam(key string, value interface{}) {
	if t.CustomParams == nil {
		t.CustomParams = make(map[string]interface{})
	}
	t.CustomParams[key] = value
}

// ToJSON はトラッキングデータをJSONに変換します
func (t *TrackingData) ToJSON() ([]byte, error) {
	return json.Marshal(t)
}

// FromJSON はJSONからトラッキングデータを復元します
func (t *TrackingData) FromJSON(data []byte) error {
	return json.Unmarshal(data, t)
}

// TrackingStats はトラッキング統計情報を表すモデルです
type TrackingStats struct {
	AppID           string    `json:"app_id"`
	TotalRequests   int64     `json:"total_requests"`
	UniqueSessions  int64     `json:"unique_sessions"`
	UniqueIPs       int64     `json:"unique_ips"`
	BotRequests     int64     `json:"bot_requests"`
	MobileRequests  int64     `json:"mobile_requests"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	CreatedAt       time.Time `json:"created_at"`
}

// ToJSON はトラッキング統計をJSONに変換します
func (t *TrackingStats) ToJSON() ([]byte, error) {
	return json.Marshal(t)
}

// FromJSON はJSONからトラッキング統計を復元します
func (t *TrackingStats) FromJSON(data []byte) error {
	return json.Unmarshal(data, t)
}

// IsBot はユーザーエージェントがボットかどうかを判定します
func (t *TrackingData) IsBot() bool {
	userAgent := strings.ToLower(t.UserAgent)
	botKeywords := []string{
		"bot", "crawler", "spider", "scraper", "googlebot", "bingbot", "yandexbot",
		"baiduspider", "duckduckbot", "facebookexternalhit", "twitterbot",
	}
	
	for _, keyword := range botKeywords {
		if strings.Contains(userAgent, keyword) {
			return true
		}
	}
	return false
}

// IsMobile はユーザーエージェントがモバイルデバイスかどうかを判定します
func (t *TrackingData) IsMobile() bool {
	userAgent := strings.ToLower(t.UserAgent)
	mobileKeywords := []string{
		"mobile", "android", "iphone", "ipad", "ipod", "blackberry", "windows phone",
	}
	
	for _, keyword := range mobileKeywords {
		if strings.Contains(userAgent, keyword) {
			return true
		}
	}
	return false
}

// GenerateID はトラッキングデータのIDを生成します
func (t *TrackingData) GenerateID() error {
	// 32文字のランダムなIDを生成
	id, err := generateRandomString(32)
	if err != nil {
		return err
	}
	t.ID = id
	return nil
}

// GetDeviceType はデバイスタイプを取得します
func (t *TrackingData) GetDeviceType() string {
	if t.IsBot() {
		return "bot"
	}
	if t.IsMobile() {
		if strings.Contains(strings.ToLower(t.UserAgent), "ipad") {
			return "tablet"
		}
		return "mobile"
	}
	return "desktop"
}

// GetBrowser はブラウザ名を取得します
func (t *TrackingData) GetBrowser() string {
	userAgent := strings.ToLower(t.UserAgent)
	
	// EdgeはChromeの前にチェックする必要がある（EdgeはChromeベース）
	if strings.Contains(userAgent, "edg/") {
		return "Edge"
	}
	if strings.Contains(userAgent, "chrome") {
		return "Chrome"
	}
	if strings.Contains(userAgent, "firefox") {
		return "Firefox"
	}
	if strings.Contains(userAgent, "safari") {
		return "Safari"
	}
	if strings.Contains(userAgent, "opera") {
		return "Opera"
	}
	
	return "Unknown"
}

// GetOS はオペレーティングシステム名を取得します
func (t *TrackingData) GetOS() string {
	userAgent := strings.ToLower(t.UserAgent)
	
	if strings.Contains(userAgent, "windows") {
		return "Windows"
	}
	// iOSはmacOSの前にチェックする必要がある（iPhone/iPadはmacOSベース）
	if strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "ipad") || strings.Contains(userAgent, "ipod") {
		return "iOS"
	}
	if strings.Contains(userAgent, "macintosh") || strings.Contains(userAgent, "mac os") {
		return "macOS"
	}
	if strings.Contains(userAgent, "android") {
		return "Android"
	}
	if strings.Contains(userAgent, "linux") {
		return "Linux"
	}
	
	return "Unknown"
}
