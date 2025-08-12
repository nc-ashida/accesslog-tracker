package models

import (
	"encoding/json"
	"strings"
	"time"
)

// TrackingData はトラッキングデータの基本構造体
type TrackingData struct {
	ID            string                 `json:"id" db:"id"`
	ApplicationID string                 `json:"application_id" db:"application_id"`
	AppID         string                 `json:"app_id" db:"app_id"`
	ClientSubID   string                 `json:"client_sub_id,omitempty" db:"client_sub_id"`
	ModuleID      string                 `json:"module_id,omitempty" db:"module_id"`
	SessionID     string                 `json:"session_id" db:"session_id"`
	UserID        string                 `json:"user_id,omitempty" db:"user_id"`
	PageURL       string                 `json:"page_url" db:"page_url"`
	URL           string                 `json:"url,omitempty" db:"url"`
	Referrer      string                 `json:"referrer,omitempty" db:"referrer"`
	UserAgent     string                 `json:"user_agent" db:"user_agent"`
	IPAddress     string                 `json:"ip_address" db:"ip_address"`
	Country       string                 `json:"country,omitempty" db:"country"`
	Region        string                 `json:"region,omitempty" db:"region"`
	City          string                 `json:"city,omitempty" db:"city"`
	ISP           string                 `json:"isp,omitempty" db:"isp"`
	DeviceType    string                 `json:"device_type,omitempty" db:"device_type"`
	Browser       string                 `json:"browser,omitempty" db:"browser"`
	OS            string                 `json:"os,omitempty" db:"os"`
	ScreenWidth   int                    `json:"screen_width,omitempty" db:"screen_width"`
	ScreenHeight  int                    `json:"screen_height,omitempty" db:"screen_height"`
	ScreenResolution string              `json:"screen_resolution,omitempty" db:"screen_resolution"`
	ViewportWidth int                    `json:"viewport_width,omitempty" db:"viewport_width"`
	ViewportHeight int                   `json:"viewport_height,omitempty" db:"viewport_height"`
	Language      string                 `json:"language,omitempty" db:"language"`
	Timezone      string                 `json:"timezone,omitempty" db:"timezone"`
	CustomParams  map[string]interface{} `json:"custom_params,omitempty" db:"custom_params"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
}

// TrackingRequest はAPIリクエスト用の構造体
type TrackingRequest struct {
	AppID         string                 `json:"app_id" validate:"required"`
	ClientSubID   string                 `json:"client_sub_id,omitempty"`
	ModuleID      string                 `json:"module_id,omitempty"`
	URL           string                 `json:"url,omitempty"`
	Referrer      string                 `json:"referrer,omitempty"`
	UserAgent     string                 `json:"user_agent" validate:"required"`
	IPAddress     string                 `json:"ip_address,omitempty"`
	SessionID     string                 `json:"session_id,omitempty"`
	ScreenResolution string              `json:"screen_resolution,omitempty"`
	Language      string                 `json:"language,omitempty"`
	Timezone      string                 `json:"timezone,omitempty"`
	CustomParams  map[string]interface{} `json:"custom_params,omitempty"`
}

// TrackingResponse はAPIレスポンス用の構造体
type TrackingResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    struct {
		ID        string    `json:"id"`
		Timestamp time.Time `json:"timestamp"`
	} `json:"data,omitempty"`
	Error *TrackingError `json:"error,omitempty"`
}

// TrackingError はエラー情報を格納する構造体
type TrackingError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

// TrackingStats は統計情報を格納する構造体
type TrackingStats struct {
	TotalEvents     int64     `json:"total_events"`
	UniqueUsers     int64     `json:"unique_users"`
	UniqueSessions  int64     `json:"unique_sessions"`
	PageViews       int64     `json:"page_views"`
	BounceRate      float64   `json:"bounce_rate"`
	AvgSessionTime  float64   `json:"avg_session_time"`
	TopPages        []PageStat `json:"top_pages"`
	TopReferrers    []ReferrerStat `json:"top_referrers"`
	TopCountries    []CountryStat `json:"top_countries"`
	TopBrowsers     []BrowserStat `json:"top_browsers"`
	TopDevices      []DeviceStat `json:"top_devices"`
	Period          string    `json:"period"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
}

// PageStat はページ統計情報
type PageStat struct {
	URL         string `json:"url"`
	PageViews   int64  `json:"page_views"`
	UniqueViews int64  `json:"unique_views"`
	BounceRate  float64 `json:"bounce_rate"`
}

// ReferrerStat はリファラー統計情報
type ReferrerStat struct {
	Referrer    string `json:"referrer"`
	Visits      int64  `json:"visits"`
	UniqueVisits int64 `json:"unique_visits"`
}

// CountryStat は国別統計情報
type CountryStat struct {
	Country     string `json:"country"`
	Visits      int64  `json:"visits"`
	UniqueVisits int64 `json:"unique_visits"`
}

// BrowserStat はブラウザ統計情報
type BrowserStat struct {
	Browser     string `json:"browser"`
	Visits      int64  `json:"visits"`
	UniqueVisits int64 `json:"unique_visits"`
	Version     string `json:"version,omitempty"`
}

// DeviceStat はデバイス統計情報
type DeviceStat struct {
	DeviceType  string `json:"device_type"`
	Visits      int64  `json:"visits"`
	UniqueVisits int64 `json:"unique_visits"`
}

// TrackingFilter はトラッキングデータのフィルタリング条件
type TrackingFilter struct {
	ApplicationID string    `json:"application_id,omitempty"`
	SessionID     string    `json:"session_id,omitempty"`
	UserID        string    `json:"user_id,omitempty"`
	PageURL       string    `json:"page_url,omitempty"`
	Referrer      string    `json:"referrer,omitempty"`
	IPAddress     string    `json:"ip_address,omitempty"`
	Country       string    `json:"country,omitempty"`
	DeviceType    string    `json:"device_type,omitempty"`
	Browser       string    `json:"browser,omitempty"`
	OS            string    `json:"os,omitempty"`
	StartDate     time.Time `json:"start_date,omitempty"`
	EndDate       time.Time `json:"end_date,omitempty"`
	Limit         int       `json:"limit,omitempty"`
	Offset        int       `json:"offset,omitempty"`
	SortBy        string    `json:"sort_by,omitempty"`
	SortOrder     string    `json:"sort_order,omitempty"`
}

// TrackingQuery はトラッキングデータのクエリ条件
type TrackingQuery struct {
	Filter    TrackingFilter `json:"filter"`
	GroupBy   []string       `json:"group_by,omitempty"`
	Aggregate []string       `json:"aggregate,omitempty"`
	TimeRange string         `json:"time_range,omitempty"`
}

// NewTrackingData は新しいトラッキングデータを作成
func NewTrackingData(req *TrackingRequest) *TrackingData {
	now := time.Now()
	
	return &TrackingData{
		AppID:         req.AppID,
		ClientSubID:   req.ClientSubID,
		ModuleID:      req.ModuleID,
		SessionID:     req.SessionID,
		URL:           req.URL,
		Referrer:      req.Referrer,
		UserAgent:     req.UserAgent,
		IPAddress:     req.IPAddress,
		ScreenResolution: req.ScreenResolution,
		Language:      req.Language,
		Timezone:      req.Timezone,
		CustomParams:  req.CustomParams,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// ToJSON はトラッキングデータをJSONに変換
func (t *TrackingData) ToJSON() ([]byte, error) {
	return json.Marshal(t)
}

// FromJSON はJSONからトラッキングデータを作成
func (t *TrackingData) FromJSON(data []byte) error {
	return json.Unmarshal(data, t)
}

// GetCustomParam はカスタムパラメータを取得
func (t *TrackingData) GetCustomParam(key string) interface{} {
	if t.CustomParams == nil {
		return nil
	}
	return t.CustomParams[key]
}

// SetCustomParam はカスタムパラメータを設定
func (t *TrackingData) SetCustomParam(key string, value interface{}) {
	if t.CustomParams == nil {
		t.CustomParams = make(map[string]interface{})
	}
	t.CustomParams[key] = value
}

// IsBot はボットかどうかを判定
func (t *TrackingData) IsBot() bool {
	botKeywords := []string{
		"bot", "crawler", "spider", "scraper", "robot",
		"googlebot", "bingbot", "slurp", "duckduckbot",
		"facebookexternalhit", "twitterbot", "linkedinbot",
	}
	
	userAgent := t.UserAgent
	for _, keyword := range botKeywords {
		if strings.Contains(strings.ToLower(userAgent), strings.ToLower(keyword)) {
			return true
		}
	}
	
	return false
}

// IsMobile はモバイルデバイスかどうかを判定
func (t *TrackingData) IsMobile() bool {
	mobileKeywords := []string{
		"mobile", "android", "iphone", "ipad", "ipod",
		"blackberry", "windows phone", "opera mini",
	}
	
	userAgent := t.UserAgent
	for _, keyword := range mobileKeywords {
		if strings.Contains(strings.ToLower(userAgent), strings.ToLower(keyword)) {
			return true
		}
	}
	
	return false
}

// GetDeviceType はデバイスタイプを取得
func (t *TrackingData) GetDeviceType() string {
	if t.DeviceType != "" {
		return t.DeviceType
	}
	
	if t.IsMobile() {
		return "mobile"
	}
	
	if strings.Contains(strings.ToLower(t.UserAgent), "tablet") {
		return "tablet"
	}
	
	return "desktop"
}

// GetBrowser はブラウザ情報を取得
func (t *TrackingData) GetBrowser() string {
	if t.Browser != "" {
		return t.Browser
	}
	
	userAgent := t.UserAgent
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
func (t *TrackingData) GetOS() string {
	if t.OS != "" {
		return t.OS
	}
	
	userAgent := t.UserAgent
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
