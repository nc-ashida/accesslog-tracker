package models

import "errors"

// アプリケーション関連のエラー
var (
	ErrApplicationAppIDRequired    = errors.New("app_id is required")
	ErrApplicationNameRequired     = errors.New("name is required")
	ErrApplicationDomainRequired   = errors.New("domain is required")
	ErrApplicationAPIKeyRequired   = errors.New("api_key is required")
	ErrApplicationNotFound         = errors.New("application not found")
	ErrApplicationAlreadyExists    = errors.New("application already exists")
	ErrApplicationInvalidAPIKey    = errors.New("invalid API key")
)

// トラッキングデータ関連のエラー
var (
	ErrTrackingAppIDRequired       = errors.New("app_id is required")
	ErrTrackingUserAgentRequired   = errors.New("user_agent is required")
	ErrTrackingURLRequired         = errors.New("url is required")
	ErrTrackingTimestampRequired   = errors.New("timestamp is required")
	ErrTrackingDataNotFound        = errors.New("tracking data not found")
	ErrTrackingInvalidData         = errors.New("invalid tracking data")
)

// セッション関連のエラー
var (
	ErrSessionNotFound             = errors.New("session not found")
	ErrSessionExpired              = errors.New("session expired")
	ErrSessionInvalid              = errors.New("invalid session")
)

// 統計関連のエラー
var (
	ErrStatisticsNotFound          = errors.New("statistics not found")
	ErrStatisticsInvalidPeriod     = errors.New("invalid statistics period")
	ErrStatisticsInvalidMetric     = errors.New("invalid statistics metric")
)

// バリデーション関連のエラー
var (
	ErrValidationError             = errors.New("validation error")
)
