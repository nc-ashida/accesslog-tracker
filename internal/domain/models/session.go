package models

import (
	"encoding/json"
	"strings"
	"time"
)

// Session はセッション情報を格納する構造体
type Session struct {
	ID            string            `json:"id" db:"id"`
	ApplicationID string            `json:"application_id" db:"application_id"`
	UserID        string            `json:"user_id,omitempty" db:"user_id"`
	IPAddress     string            `json:"ip_address" db:"ip_address"`
	UserAgent     string            `json:"user_agent" db:"user_agent"`
	Referrer      string            `json:"referrer,omitempty" db:"referrer"`
	EntryPage     string            `json:"entry_page" db:"entry_page"`
	ExitPage      string            `json:"exit_page,omitempty" db:"exit_page"`
	PageViews     int               `json:"page_views" db:"page_views"`
	Duration      int               `json:"duration" db:"duration"` // 秒単位
	Bounce        bool              `json:"bounce" db:"bounce"`
	DeviceType    string            `json:"device_type,omitempty" db:"device_type"`
	Browser       string            `json:"browser,omitempty" db:"browser"`
	OS            string            `json:"os,omitempty" db:"os"`
	Country       string            `json:"country,omitempty" db:"country"`
	Region        string            `json:"region,omitempty" db:"region"`
	City          string            `json:"city,omitempty" db:"city"`
	Language      string            `json:"language,omitempty" db:"language"`
	Timezone      string            `json:"timezone,omitempty" db:"timezone"`
	CustomData    map[string]string `json:"custom_data,omitempty" db:"custom_data"`
	StartedAt     time.Time         `json:"started_at" db:"started_at"`
	LastActivity  time.Time         `json:"last_activity" db:"last_activity"`
	EndedAt       *time.Time        `json:"ended_at,omitempty" db:"ended_at"`
	CreatedAt     time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at" db:"updated_at"`
}

// SessionRequest はセッション作成リクエスト
type SessionRequest struct {
	ApplicationID string            `json:"application_id" validate:"required"`
	UserID        string            `json:"user_id,omitempty"`
	IPAddress     string            `json:"ip_address" validate:"required"`
	UserAgent     string            `json:"user_agent" validate:"required"`
	Referrer      string            `json:"referrer,omitempty"`
	EntryPage     string            `json:"entry_page" validate:"required"`
	DeviceType    string            `json:"device_type,omitempty"`
	Browser       string            `json:"browser,omitempty"`
	OS            string            `json:"os,omitempty"`
	Country       string            `json:"country,omitempty"`
	Region        string            `json:"region,omitempty"`
	City          string            `json:"city,omitempty"`
	Language      string            `json:"language,omitempty"`
	Timezone      string            `json:"timezone,omitempty"`
	CustomData    map[string]string `json:"custom_data,omitempty"`
}

// SessionResponse はセッションAPIレスポンス
type SessionResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message,omitempty"`
	Data    *Session `json:"data,omitempty"`
	Error   *SessionError `json:"error,omitempty"`
}

// SessionError はセッションエラー情報
type SessionError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

// SessionStats はセッション統計情報
type SessionStats struct {
	TotalSessions    int64     `json:"total_sessions"`
	ActiveSessions   int64     `json:"active_sessions"`
	AvgSessionTime   float64   `json:"avg_session_time"`
	BounceRate       float64   `json:"bounce_rate"`
	AvgPageViews     float64   `json:"avg_page_views"`
	TopEntryPages    []PageStat `json:"top_entry_pages"`
	TopExitPages     []PageStat `json:"top_exit_pages"`
	DeviceBreakdown  []DeviceStat `json:"device_breakdown"`
	BrowserBreakdown []BrowserStat `json:"browser_breakdown"`
	CountryBreakdown []CountryStat `json:"country_breakdown"`
	Period           string    `json:"period"`
	StartDate        time.Time `json:"start_date"`
	EndDate          time.Time `json:"end_date"`
}

// SessionFilter はセッションフィルタリング条件
type SessionFilter struct {
	ApplicationID string    `json:"application_id,omitempty"`
	UserID        string    `json:"user_id,omitempty"`
	IPAddress     string    `json:"ip_address,omitempty"`
	DeviceType    string    `json:"device_type,omitempty"`
	Browser       string    `json:"browser,omitempty"`
	OS            string    `json:"os,omitempty"`
	Country       string    `json:"country,omitempty"`
	Bounce        *bool     `json:"bounce,omitempty"`
	StartedAt     time.Time `json:"started_at,omitempty"`
	EndedAt       time.Time `json:"ended_at,omitempty"`
	Limit         int       `json:"limit,omitempty"`
	Offset        int       `json:"offset,omitempty"`
	SortBy        string    `json:"sort_by,omitempty"`
	SortOrder     string    `json:"sort_order,omitempty"`
}

// SessionEvent はセッションイベント
type SessionEvent struct {
	ID          string                 `json:"id" db:"id"`
	SessionID   string                 `json:"session_id" db:"session_id"`
	EventType   string                 `json:"event_type" db:"event_type"`
	EventName   string                 `json:"event_name" db:"event_name"`
	PageURL     string                 `json:"page_url" db:"page_url"`
	EventData   map[string]interface{} `json:"event_data,omitempty" db:"event_data"`
	Timestamp   time.Time              `json:"timestamp" db:"timestamp"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
}

// NewSession は新しいセッションを作成
func NewSession(req *SessionRequest) *Session {
	now := time.Now()
	
	return &Session{
		ApplicationID: req.ApplicationID,
		UserID:        req.UserID,
		IPAddress:     req.IPAddress,
		UserAgent:     req.UserAgent,
		Referrer:      req.Referrer,
		EntryPage:     req.EntryPage,
		PageViews:     1, // 初期ページビュー
		Duration:      0,
		Bounce:        true, // 初期はバウンス
		DeviceType:    req.DeviceType,
		Browser:       req.Browser,
		OS:            req.OS,
		Country:       req.Country,
		Region:        req.Region,
		City:          req.City,
		Language:      req.Language,
		Timezone:      req.Timezone,
		CustomData:    req.CustomData,
		StartedAt:     now,
		LastActivity:  now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// ToJSON はセッションをJSONに変換
func (s *Session) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

// FromJSON はJSONからセッションを作成
func (s *Session) FromJSON(data []byte) error {
	return json.Unmarshal(data, s)
}

// IsActive はセッションがアクティブかどうかを判定
func (s *Session) IsActive() bool {
	return s.EndedAt == nil
}

// IsExpired はセッションが期限切れかどうかを判定
func (s *Session) IsExpired(timeout int) bool {
	if s.IsActive() {
		return time.Since(s.LastActivity) > time.Duration(timeout)*time.Second
	}
	return false
}

// UpdateActivity はセッションの最終アクティビティを更新
func (s *Session) UpdateActivity(pageURL string) {
	s.LastActivity = time.Now()
	s.PageViews++
	s.Bounce = false // 複数ページビューがあるのでバウンスではない
	s.UpdatedAt = time.Now()
}

// EndSession はセッションを終了
func (s *Session) EndSession(exitPage string) {
	now := time.Now()
	s.ExitPage = exitPage
	s.EndedAt = &now
	s.Duration = int(now.Sub(s.StartedAt).Seconds())
	s.UpdatedAt = now
}

// GetDuration はセッションの継続時間を取得（秒単位）
func (s *Session) GetDuration() int {
	if s.EndedAt != nil {
		return int(s.EndedAt.Sub(s.StartedAt).Seconds())
	}
	return int(time.Since(s.StartedAt).Seconds())
}

// IsBounce はバウンスセッションかどうかを判定
func (s *Session) IsBounce() bool {
	return s.PageViews <= 1
}

// GetCustomData はカスタムデータを取得
func (s *Session) GetCustomData(key string) string {
	if s.CustomData == nil {
		return ""
	}
	return s.CustomData[key]
}

// SetCustomData はカスタムデータを設定
func (s *Session) SetCustomData(key, value string) {
	if s.CustomData == nil {
		s.CustomData = make(map[string]string)
	}
	s.CustomData[key] = value
	s.UpdatedAt = time.Now()
}

// RemoveCustomData はカスタムデータを削除
func (s *Session) RemoveCustomData(key string) {
	if s.CustomData == nil {
		return
	}
	delete(s.CustomData, key)
	s.UpdatedAt = time.Now()
}

// GetDeviceType はデバイスタイプを取得
func (s *Session) GetDeviceType() string {
	if s.DeviceType != "" {
		return s.DeviceType
	}
	
	// UserAgentから推測
	if strings.Contains(strings.ToLower(s.UserAgent), "mobile") || strings.Contains(strings.ToLower(s.UserAgent), "android") || 
	   strings.Contains(strings.ToLower(s.UserAgent), "iphone") || strings.Contains(strings.ToLower(s.UserAgent), "ipad") {
		return "mobile"
	}
	
	if strings.Contains(strings.ToLower(s.UserAgent), "tablet") {
		return "tablet"
	}
	
	return "desktop"
}

// GetBrowser はブラウザ情報を取得
func (s *Session) GetBrowser() string {
	if s.Browser != "" {
		return s.Browser
	}
	
	userAgent := s.UserAgent
	if strings.Contains(strings.ToLower(userAgent), "chrome") {
		return "chrome"
	} else if strings.Contains(strings.ToLower(userAgent), "firefox") {
		return "firefox"
	} else if strings.Contains(strings.ToLower(userAgent), "safari") {
		return "safari"
	} else if strings.Contains(strings.ToLower(userAgent), "edge") {
		return "edge"
	} else if strings.Contains(strings.ToLower(userAgent), "opera") {
		return "opera"
	}
	
	return "unknown"
}

// GetOS はOS情報を取得
func (s *Session) GetOS() string {
	if s.OS != "" {
		return s.OS
	}
	
	userAgent := s.UserAgent
	if strings.Contains(strings.ToLower(userAgent), "windows") {
		return "windows"
	} else if strings.Contains(strings.ToLower(userAgent), "mac") {
		return "macos"
	} else if strings.Contains(strings.ToLower(userAgent), "linux") {
		return "linux"
	} else if strings.Contains(strings.ToLower(userAgent), "android") {
		return "android"
	} else if strings.Contains(strings.ToLower(userAgent), "ios") {
		return "ios"
	}
	
	return "unknown"
}

// IsBot はボットかどうかを判定
func (s *Session) IsBot() bool {
	botKeywords := []string{
		"bot", "crawler", "spider", "scraper", "robot",
		"googlebot", "bingbot", "slurp", "duckduckbot",
		"facebookexternalhit", "twitterbot", "linkedinbot",
	}
	
	userAgent := s.UserAgent
	for _, keyword := range botKeywords {
		if strings.Contains(strings.ToLower(userAgent), strings.ToLower(keyword)) {
			return true
		}
	}
	
	return false
}
