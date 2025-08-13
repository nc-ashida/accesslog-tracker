package models

import (
	"time"
)

// TrackingRequest はトラッキングAPIのリクエスト構造体です
type TrackingRequest struct {
	AppID       string                 `json:"app_id" binding:"required"`
	UserAgent   string                 `json:"user_agent"`
	URL         string                 `json:"url"`
	IPAddress   string                 `json:"ip_address"`
	SessionID   string                 `json:"session_id"`
	Referrer    string                 `json:"referrer"`
	CustomParams map[string]interface{} `json:"custom_params"`
}

// StatisticsRequest は統計APIのリクエスト構造体です
type StatisticsRequest struct {
	AppID     string `json:"app_id" binding:"required"`
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
	GroupBy   string `json:"group_by"` // "day", "hour", "page", "referrer"
	Limit     int    `json:"limit"`    // 結果の制限数
}

// ApplicationRequest はアプリケーションAPIのリクエスト構造体です
type ApplicationRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Domain      string `json:"domain" binding:"required"`
}

// ApplicationUpdateRequest はアプリケーション更新APIのリクエスト構造体です
type ApplicationUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Domain      string `json:"domain"`
	Active      *bool  `json:"active"`
}

// PaginationRequest はページネーション用のリクエスト構造体です
type PaginationRequest struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"page_size" form:"page_size"`
}

// FilterRequest はフィルタリング用のリクエスト構造体です
type FilterRequest struct {
	StartDate *time.Time `json:"start_date" form:"start_date"`
	EndDate   *time.Time `json:"end_date" form:"end_date"`
	AppID     string     `json:"app_id" form:"app_id"`
	SessionID string     `json:"session_id" form:"session_id"`
	IPAddress string     `json:"ip_address" form:"ip_address"`
}
