package validators

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/nc-ashida/accesslog-tracker/internal/domain/models"
	"github.com/nc-ashida/accesslog-tracker/internal/utils/iputil"
)

// TrackingValidator はトラッキングデータのバリデーションを行います
type TrackingValidator struct{}

// NewTrackingValidator は新しいトラッキングバリデーターを作成します
func NewTrackingValidator() *TrackingValidator {
	return &TrackingValidator{}
}

// Validate はトラッキングデータの妥当性を検証します
func (v *TrackingValidator) Validate(data *models.TrackingData) error {
	if err := data.Validate(); err != nil {
		return err
	}

	if err := v.validateAppID(data.AppID); err != nil {
		return err
	}

	if err := v.validateUserAgent(data.UserAgent); err != nil {
		return err
	}

	if err := v.validateTimestamp(data.Timestamp); err != nil {
		return err
	}

	if err := v.validateIPAddress(data.IPAddress); err != nil {
		return err
	}

	if err := v.validateURL(data.URL); err != nil {
		return err
	}

	if err := v.validateReferrer(data.Referrer); err != nil {
		return err
	}

	return nil
}

// validateAppID はアプリケーションIDを検証します
func (v *TrackingValidator) validateAppID(appID string) error {
	if appID == "" {
		return models.ErrTrackingAppIDRequired
	}

	if len(appID) < 8 {
		return errors.New("app_id must be at least 8 characters")
	}

	if len(appID) > 50 {
		return errors.New("app_id must be at most 50 characters")
	}

	// アプリケーションIDの形式チェック
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9\-_]+$`, appID)
	if !matched {
		return errors.New("app_id contains invalid characters")
	}

	return nil
}

// validateUserAgent はユーザーエージェントを検証します
func (v *TrackingValidator) validateUserAgent(userAgent string) error {
	if userAgent == "" {
		return models.ErrTrackingUserAgentRequired
	}

	if len(userAgent) < 10 {
		return errors.New("user agent must be at least 10 characters")
	}

	if len(userAgent) > 140 {
		return errors.New("user agent must be at most 140 characters")
	}

	// nullバイトのチェック
	if strings.Contains(userAgent, "\x00") {
		return errors.New("user agent contains null bytes")
	}

	return nil
}

// validateTimestamp はタイムスタンプを検証します
func (v *TrackingValidator) validateTimestamp(timestamp time.Time) error {
	if timestamp.IsZero() {
		return models.ErrTrackingTimestampRequired
	}

	// 未来の日時は許可しない
	if timestamp.After(time.Now().Add(1 * time.Minute)) {
		return errors.New("timestamp cannot be in the future")
	}

	// 過去すぎる日時は許可しない（10年前まで）
	if timestamp.Before(time.Now().AddDate(-10, 0, 0)) {
		return errors.New("timestamp is too old")
	}

	return nil
}

// validateIPAddress はIPアドレスを検証します
func (v *TrackingValidator) validateIPAddress(ipAddress string) error {
	if ipAddress == "" {
		return nil // IPアドレスはオプション
	}

	if !iputil.IsValidIP(ipAddress) {
		return errors.New("invalid IP address format")
	}

	return nil
}

// validateURL はURLを検証します
func (v *TrackingValidator) validateURL(urlStr string) error {
	if urlStr == "" {
		return models.ErrTrackingURLRequired
	}

	if len(urlStr) > 2048 {
		return errors.New("URL must be at most 2048 characters")
	}

	// URL形式のチェック
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return errors.New("Invalid URL format")
	}

	if parsedURL.Scheme == "" {
		return errors.New("URL must have a scheme (http:// or https://)")
	}

	if parsedURL.Host == "" {
		return errors.New("URL must have a host")
	}

	// 許可されていないプロトコルのチェック
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.New("URL must use http or https protocol")
	}

	// XSS攻撃の可能性がある文字列のチェック
	if strings.Contains(urlStr, "<script>") || strings.Contains(urlStr, "javascript:") {
		return errors.New("URL contains potentially dangerous content")
	}

	return nil
}

// validateReferrer はリファラーを検証します
func (v *TrackingValidator) validateReferrer(referrer string) error {
	if referrer == "" {
		return nil // リファラーはオプション
	}

	if len(referrer) > 2048 {
		return errors.New("referrer must be at most 2048 characters")
	}

	// リファラーの形式チェック
	parsedURL, err := url.Parse(referrer)
	if err != nil {
		return errors.New("invalid referrer format")
	}

	if parsedURL.Scheme == "" {
		return errors.New("referrer must have a scheme (http:// or https://)")
	}

	if parsedURL.Host == "" {
		return errors.New("referrer must have a host")
	}

	return nil
}

// ValidateCustomParams はカスタムパラメータを検証します
func (v *TrackingValidator) ValidateCustomParams(params map[string]interface{}) error {
	if params == nil {
		return nil
	}

	if len(params) > 50 {
		return errors.New("custom params cannot exceed 50 items")
	}

	for key, value := range params {
		if err := v.validateCustomParamKey(key); err != nil {
			return err
		}

		if err := v.validateCustomParamValue(value); err != nil {
			return err
		}
	}

	return nil
}

// validateCustomParamKey はカスタムパラメータのキーを検証します
func (v *TrackingValidator) validateCustomParamKey(key string) error {
	if key == "" {
		return errors.New("custom param key cannot be empty")
	}

	if len(key) > 50 {
		return errors.New("custom param key must be at most 50 characters")
	}

	// キーの形式チェック
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, key)
	if !matched {
		return errors.New("custom param key contains invalid characters")
	}

	return nil
}

// validateCustomParamValue はカスタムパラメータの値を検証します
func (v *TrackingValidator) validateCustomParamValue(value interface{}) error {
	switch val := value.(type) {
	case string:
		if len(val) > 140 {
			return errors.New("custom param string value must be at most 140 characters")
		}
	case int, int64, float64, bool:
		// 数値とブール値は制限なし
	default:
		return errors.New("custom param value must be string, number, or boolean")
	}

	return nil
}

// IsCrawler はユーザーエージェントがクローラーかどうかを判定します
func (v *TrackingValidator) IsCrawler(userAgent string) bool {
	userAgentLower := strings.ToLower(userAgent)
	crawlerKeywords := []string{
		"bot", "crawler", "spider", "scraper", "googlebot", "bingbot", "yandexbot",
		"baiduspider", "duckduckbot", "facebookexternalhit", "twitterbot",
	}
	
	for _, keyword := range crawlerKeywords {
		if strings.Contains(userAgentLower, keyword) {
			return true
		}
	}
	return false
}

// ValidateAppID はアプリケーションIDを検証します
func (v *TrackingValidator) ValidateAppID(appID string) error {
	return v.validateAppID(appID)
}

// ValidateUserAgent はユーザーエージェントを検証します
func (v *TrackingValidator) ValidateUserAgent(userAgent string) error {
	return v.validateUserAgent(userAgent)
}

// ValidateURL はURLを検証します
func (v *TrackingValidator) ValidateURL(url string) error {
	return v.validateURL(url)
}

// ValidateTimestamp はタイムスタンプを検証します
func (v *TrackingValidator) ValidateTimestamp(timestamp time.Time) error {
	return v.validateTimestamp(timestamp)
}
