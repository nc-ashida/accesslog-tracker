package models

import (
	"time"
)

// APIResponse は標準的なAPIレスポンスの構造体です
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// APIError はAPIエラーの構造体です
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// TrackingResponse はトラッキングAPIのレスポンス構造体です
type TrackingResponse struct {
	TrackingID string    `json:"tracking_id"`
	AppID      string    `json:"app_id"`
	SessionID  string    `json:"session_id"`
	Timestamp  time.Time `json:"timestamp"`
}

// StatisticsResponse は統計APIのレスポンス構造体です
type StatisticsResponse struct {
	AppID          string    `json:"app_id"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	TotalRequests  int64     `json:"total_requests"`
	UniqueVisitors int64     `json:"unique_visitors"`
	TopPages       []PageStats `json:"top_pages"`
	TopReferrers   []ReferrerStats `json:"top_referrers"`
}

// PageStats はページ統計の構造体です
type PageStats struct {
	URL   string `json:"url"`
	Count int64  `json:"count"`
}

// ReferrerStats はリファラー統計の構造体です
type ReferrerStats struct {
	Referrer string `json:"referrer"`
	Count    int64  `json:"count"`
}

// ApplicationResponse はアプリケーションAPIのレスポンス構造体です
type ApplicationResponse struct {
	AppID      string    `json:"app_id"`
	Name       string    `json:"name"`
	Description string   `json:"description"`
	Domain     string    `json:"domain"`
	APIKey     string    `json:"api_key"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// HealthResponse はヘルスチェックAPIのレスポンス構造体です
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}
