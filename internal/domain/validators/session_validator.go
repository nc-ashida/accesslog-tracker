package validators

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/your-username/accesslog-tracker/internal/domain/models"
	"github.com/your-username/accesslog-tracker/internal/utils/iputil"
)

// SessionValidator はセッションのバリデーター
type SessionValidator struct {
	// バリデーション設定
	MaxCustomDataCount     int
	MaxCustomDataKeyLength int
	MaxCustomDataValueLength int
	MaxEntryPageLength     int
	MaxExitPageLength      int
	MaxLanguageLength      int
	MaxTimezoneLength      int
	MinSessionDuration     int
	MaxSessionDuration     int
}

// NewSessionValidator は新しいセッションバリデーターを作成
func NewSessionValidator() *SessionValidator {
	return &SessionValidator{
		MaxCustomDataCount:     20,
		MaxCustomDataKeyLength: 50,
		MaxCustomDataValueLength: 500,
		MaxEntryPageLength:     2048,
		MaxExitPageLength:      2048,
		MaxLanguageLength:      10,
		MaxTimezoneLength:      50,
		MinSessionDuration:     0,
		MaxSessionDuration:     86400, // 24時間
	}
}

// ValidateSessionRequest はセッションリクエストを検証
func (sv *SessionValidator) ValidateSessionRequest(req *models.SessionRequest) error {
	var errors []string

	// 必須フィールドの検証
	if err := sv.validateRequiredFields(req); err != nil {
		errors = append(errors, err.Error())
	}

	// アプリケーションIDの検証
	if err := sv.validateApplicationID(req.ApplicationID); err != nil {
		errors = append(errors, err.Error())
	}

	// IPアドレスの検証
	if err := sv.validateIPAddress(req.IPAddress); err != nil {
		errors = append(errors, err.Error())
	}

	// UserAgentの検証
	if err := sv.validateUserAgent(req.UserAgent); err != nil {
		errors = append(errors, err.Error())
	}

	// リファラーの検証
	if req.Referrer != "" {
		if err := sv.validateReferrer(req.Referrer); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// エントリーページの検証
	if err := sv.validateEntryPage(req.EntryPage); err != nil {
		errors = append(errors, err.Error())
	}

	// デバイスタイプの検証
	if req.DeviceType != "" {
		if err := sv.validateDeviceType(req.DeviceType); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// ブラウザの検証
	if req.Browser != "" {
		if err := sv.validateBrowser(req.Browser); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// OSの検証
	if req.OS != "" {
		if err := sv.validateOS(req.OS); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// 国の検証
	if req.Country != "" {
		if err := sv.validateCountry(req.Country); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// 地域の検証
	if req.Region != "" {
		if err := sv.validateRegion(req.Region); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// 都市の検証
	if req.City != "" {
		if err := sv.validateCity(req.City); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// 言語の検証
	if req.Language != "" {
		if err := sv.validateLanguage(req.Language); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// タイムゾーンの検証
	if req.Timezone != "" {
		if err := sv.validateTimezone(req.Timezone); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// カスタムデータの検証
	if len(req.CustomData) > 0 {
		if err := sv.validateCustomData(req.CustomData); err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// ValidateSession はセッションを検証
func (sv *SessionValidator) ValidateSession(session *models.Session) error {
	var errors []string

	// 必須フィールドの検証
	if session.ID == "" {
		errors = append(errors, "id is required")
	}
	if session.ApplicationID == "" {
		errors = append(errors, "application_id is required")
	}
	if session.IPAddress == "" {
		errors = append(errors, "ip_address is required")
	}
	if session.UserAgent == "" {
		errors = append(errors, "user_agent is required")
	}
	if session.EntryPage == "" {
		errors = append(errors, "entry_page is required")
	}

	// アプリケーションIDの検証
	if err := sv.validateApplicationID(session.ApplicationID); err != nil {
		errors = append(errors, err.Error())
	}

	// IPアドレスの検証
	if err := sv.validateIPAddress(session.IPAddress); err != nil {
		errors = append(errors, err.Error())
	}

	// UserAgentの検証
	if err := sv.validateUserAgent(session.UserAgent); err != nil {
		errors = append(errors, err.Error())
	}

	// エントリーページの検証
	if err := sv.validateEntryPage(session.EntryPage); err != nil {
		errors = append(errors, err.Error())
	}

	// エグジットページの検証
	if session.ExitPage != "" {
		if err := sv.validateExitPage(session.ExitPage); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// ページビュー数の検証
	if session.PageViews < 1 {
		errors = append(errors, "page_views must be at least 1")
	}

	// セッション時間の検証
	if session.Duration < sv.MinSessionDuration || session.Duration > sv.MaxSessionDuration {
		errors = append(errors, fmt.Sprintf("duration must be between %d and %d seconds", sv.MinSessionDuration, sv.MaxSessionDuration))
	}

	// デバイスタイプの検証
	if session.DeviceType != "" {
		if err := sv.validateDeviceType(session.DeviceType); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// ブラウザの検証
	if session.Browser != "" {
		if err := sv.validateBrowser(session.Browser); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// OSの検証
	if session.OS != "" {
		if err := sv.validateOS(session.OS); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// 国の検証
	if session.Country != "" {
		if err := sv.validateCountry(session.Country); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// 地域の検証
	if session.Region != "" {
		if err := sv.validateRegion(session.Region); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// 都市の検証
	if session.City != "" {
		if err := sv.validateCity(session.City); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// 言語の検証
	if session.Language != "" {
		if err := sv.validateLanguage(session.Language); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// タイムゾーンの検証
	if session.Timezone != "" {
		if err := sv.validateTimezone(session.Timezone); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// カスタムデータの検証
	if len(session.CustomData) > 0 {
		if err := sv.validateCustomData(session.CustomData); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// 日時の検証
	if err := sv.validateTimestamps(session); err != nil {
		errors = append(errors, err.Error())
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// validateRequiredFields は必須フィールドを検証
func (sv *SessionValidator) validateRequiredFields(req *models.SessionRequest) error {
	if req.ApplicationID == "" {
		return fmt.Errorf("application_id is required")
	}
	if req.IPAddress == "" {
		return fmt.Errorf("ip_address is required")
	}
	if req.UserAgent == "" {
		return fmt.Errorf("user_agent is required")
	}
	if req.EntryPage == "" {
		return fmt.Errorf("entry_page is required")
	}
	return nil
}

// validateApplicationID はアプリケーションIDを検証
func (sv *SessionValidator) validateApplicationID(appID string) error {
	if appID == "" {
		return fmt.Errorf("application_id cannot be empty")
	}

	// アプリケーションIDの形式検証（UUIDまたはカスタム形式）
	if !sv.isValidApplicationID(appID) {
		return fmt.Errorf("invalid application_id format")
	}

	return nil
}

// validateIPAddress はIPアドレスを検証
func (sv *SessionValidator) validateIPAddress(ipAddress string) error {
	if !iputil.IsValidIP(ipAddress) {
		return fmt.Errorf("invalid ip_address format")
	}

	return nil
}

// validateUserAgent はUserAgentを検証
func (sv *SessionValidator) validateUserAgent(userAgent string) error {
	if userAgent == "" {
		return fmt.Errorf("user_agent cannot be empty")
	}

	if len(userAgent) > 500 {
		return fmt.Errorf("user_agent length exceeds maximum limit of 500 characters")
	}

	// 基本的なUserAgent形式の検証
	if !sv.isValidUserAgent(userAgent) {
		return fmt.Errorf("invalid user_agent format")
	}

	return nil
}

// validateReferrer はリファラーを検証
func (sv *SessionValidator) validateReferrer(referrer string) error {
	if len(referrer) > 2048 {
		return fmt.Errorf("referrer length exceeds maximum limit of 2048 characters")
	}

	// URL形式の検証
	parsedURL, err := url.Parse(referrer)
	if err != nil {
		return fmt.Errorf("invalid referrer format: %v", err)
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("referrer must have valid scheme and host")
	}

	return nil
}

// validateEntryPage はエントリーページを検証
func (sv *SessionValidator) validateEntryPage(entryPage string) error {
	if entryPage == "" {
		return fmt.Errorf("entry_page cannot be empty")
	}

	if len(entryPage) > sv.MaxEntryPageLength {
		return fmt.Errorf("entry_page length exceeds maximum limit of %d characters", sv.MaxEntryPageLength)
	}

	// URL形式の検証
	parsedURL, err := url.Parse(entryPage)
	if err != nil {
		return fmt.Errorf("invalid entry_page format: %v", err)
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("entry_page must have valid scheme and host")
	}

	return nil
}

// validateExitPage はエグジットページを検証
func (sv *SessionValidator) validateExitPage(exitPage string) error {
	if len(exitPage) > sv.MaxExitPageLength {
		return fmt.Errorf("exit_page length exceeds maximum limit of %d characters", sv.MaxExitPageLength)
	}

	// URL形式の検証
	parsedURL, err := url.Parse(exitPage)
	if err != nil {
		return fmt.Errorf("invalid exit_page format: %v", err)
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("exit_page must have valid scheme and host")
	}

	return nil
}

// validateDeviceType はデバイスタイプを検証
func (sv *SessionValidator) validateDeviceType(deviceType string) error {
	validDeviceTypes := []string{"desktop", "mobile", "tablet", "unknown"}
	
	for _, validType := range validDeviceTypes {
		if deviceType == validType {
			return nil
		}
	}
	
	return fmt.Errorf("invalid device_type: %s", deviceType)
}

// validateBrowser はブラウザを検証
func (sv *SessionValidator) validateBrowser(browser string) error {
	validBrowsers := []string{"chrome", "firefox", "safari", "edge", "opera", "unknown"}
	
	for _, validBrowser := range validBrowsers {
		if browser == validBrowser {
			return nil
		}
	}
	
	return fmt.Errorf("invalid browser: %s", browser)
}

// validateOS はOSを検証
func (sv *SessionValidator) validateOS(os string) error {
	validOS := []string{"windows", "macos", "linux", "android", "ios", "unknown"}
	
	for _, validOSName := range validOS {
		if os == validOSName {
			return nil
		}
	}
	
	return fmt.Errorf("invalid os: %s", os)
}

// validateCountry は国を検証
func (sv *SessionValidator) validateCountry(country string) error {
	if len(country) > 100 {
		return fmt.Errorf("country length exceeds maximum limit of 100 characters")
	}
	
	// 国名の形式検証（英数字、スペース、ハイフンのみ許可）
	countryRegex := regexp.MustCompile(`^[a-zA-Z0-9\s\-]+$`)
	if !countryRegex.MatchString(country) {
		return fmt.Errorf("country contains invalid characters")
	}
	
	return nil
}

// validateRegion は地域を検証
func (sv *SessionValidator) validateRegion(region string) error {
	if len(region) > 100 {
		return fmt.Errorf("region length exceeds maximum limit of 100 characters")
	}
	
	// 地域名の形式検証（英数字、スペース、ハイフンのみ許可）
	regionRegex := regexp.MustCompile(`^[a-zA-Z0-9\s\-]+$`)
	if !regionRegex.MatchString(region) {
		return fmt.Errorf("region contains invalid characters")
	}
	
	return nil
}

// validateCity は都市を検証
func (sv *SessionValidator) validateCity(city string) error {
	if len(city) > 100 {
		return fmt.Errorf("city length exceeds maximum limit of 100 characters")
	}
	
	// 都市名の形式検証（英数字、スペース、ハイフンのみ許可）
	cityRegex := regexp.MustCompile(`^[a-zA-Z0-9\s\-]+$`)
	if !cityRegex.MatchString(city) {
		return fmt.Errorf("city contains invalid characters")
	}
	
	return nil
}

// validateLanguage は言語を検証
func (sv *SessionValidator) validateLanguage(language string) error {
	if len(language) > sv.MaxLanguageLength {
		return fmt.Errorf("language length exceeds maximum limit of %d characters", sv.MaxLanguageLength)
	}

	// 言語コードの形式検証（ISO 639-1形式）
	langRegex := regexp.MustCompile(`^[a-z]{2}(-[A-Z]{2})?$`)
	if !langRegex.MatchString(language) {
		return fmt.Errorf("invalid language format")
	}

	return nil
}

// validateTimezone はタイムゾーンを検証
func (sv *SessionValidator) validateTimezone(timezone string) error {
	if len(timezone) > sv.MaxTimezoneLength {
		return fmt.Errorf("timezone length exceeds maximum limit of %d characters", sv.MaxTimezoneLength)
	}

	// タイムゾーンの形式検証
	_, err := time.LoadLocation(timezone)
	if err != nil {
		return fmt.Errorf("invalid timezone format: %v", err)
	}

	return nil
}

// validateCustomData はカスタムデータを検証
func (sv *SessionValidator) validateCustomData(customData map[string]string) error {
	if len(customData) > sv.MaxCustomDataCount {
		return fmt.Errorf("custom_data count exceeds maximum limit of %d", sv.MaxCustomDataCount)
	}

	for key, value := range customData {
		// キーの検証
		if key == "" {
			return fmt.Errorf("custom_data key cannot be empty")
		}

		if len(key) > sv.MaxCustomDataKeyLength {
			return fmt.Errorf("custom_data key length exceeds maximum limit of %d characters", sv.MaxCustomDataKeyLength)
		}

		// キーの形式検証（英数字、ハイフン、アンダースコアのみ許可）
		keyRegex := regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)
		if !keyRegex.MatchString(key) {
			return fmt.Errorf("custom_data key contains invalid characters: %s", key)
		}

		// 値の検証
		if len(value) > sv.MaxCustomDataValueLength {
			return fmt.Errorf("custom_data value for key '%s' length exceeds maximum limit of %d characters", key, sv.MaxCustomDataValueLength)
		}
	}

	return nil
}

// validateTimestamps はタイムスタンプを検証
func (sv *SessionValidator) validateTimestamps(session *models.Session) error {
	now := time.Now()

	// StartedAtの検証
	if session.StartedAt.IsZero() {
		return fmt.Errorf("started_at cannot be zero")
	}

	if session.StartedAt.After(now) {
		return fmt.Errorf("started_at cannot be in the future")
	}

	// LastActivityの検証
	if session.LastActivity.IsZero() {
		return fmt.Errorf("last_activity cannot be zero")
	}

	if session.LastActivity.Before(session.StartedAt) {
		return fmt.Errorf("last_activity cannot be before started_at")
	}

	if session.LastActivity.After(now) {
		return fmt.Errorf("last_activity cannot be in the future")
	}

	// EndedAtの検証
	if session.EndedAt != nil {
		if session.EndedAt.Before(session.StartedAt) {
			return fmt.Errorf("ended_at cannot be before started_at")
		}

		if session.EndedAt.Before(session.LastActivity) {
			return fmt.Errorf("ended_at cannot be before last_activity")
		}

		if session.EndedAt.After(now) {
			return fmt.Errorf("ended_at cannot be in the future")
		}
	}

	return nil
}

// isValidApplicationID はアプリケーションIDの形式を検証
func (sv *SessionValidator) isValidApplicationID(appID string) bool {
	// UUID形式またはカスタム形式を許可
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	customRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]{1,50}$`)
	
	return uuidRegex.MatchString(appID) || customRegex.MatchString(appID)
}

// isValidUserAgent はUserAgentの形式を検証
func (sv *SessionValidator) isValidUserAgent(userAgent string) bool {
	// 基本的なUserAgent形式の検証
	if len(userAgent) < 10 {
		return false
	}
	
	// 一般的なUserAgentパターンのチェック
	patterns := []string{
		`Mozilla/`,
		`Chrome/`,
		`Safari/`,
		`Firefox/`,
		`Edge/`,
		`Opera/`,
		`bot`,
		`crawler`,
		`spider`,
	}
	
	for _, pattern := range patterns {
		if strings.Contains(userAgent, pattern) {
			return true
		}
	}
	
	return false
}

// SetMaxCustomDataCount はカスタムデータの最大数を設定
func (sv *SessionValidator) SetMaxCustomDataCount(max int) {
	sv.MaxCustomDataCount = max
}

// SetMaxCustomDataKeyLength はカスタムデータキーの最大長を設定
func (sv *SessionValidator) SetMaxCustomDataKeyLength(max int) {
	sv.MaxCustomDataKeyLength = max
}

// SetMaxCustomDataValueLength はカスタムデータ値の最大長を設定
func (sv *SessionValidator) SetMaxCustomDataValueLength(max int) {
	sv.MaxCustomDataValueLength = max
}

// SetMaxEntryPageLength はエントリーページの最大長を設定
func (sv *SessionValidator) SetMaxEntryPageLength(max int) {
	sv.MaxEntryPageLength = max
}

// SetMaxExitPageLength はエグジットページの最大長を設定
func (sv *SessionValidator) SetMaxExitPageLength(max int) {
	sv.MaxExitPageLength = max
}

// SetMaxLanguageLength は言語の最大長を設定
func (sv *SessionValidator) SetMaxLanguageLength(max int) {
	sv.MaxLanguageLength = max
}

// SetMaxTimezoneLength はタイムゾーンの最大長を設定
func (sv *SessionValidator) SetMaxTimezoneLength(max int) {
	sv.MaxTimezoneLength = max
}

// SetMinSessionDuration はセッション時間の最小値を設定
func (sv *SessionValidator) SetMinSessionDuration(min int) {
	sv.MinSessionDuration = min
}

// SetMaxSessionDuration はセッション時間の最大値を設定
func (sv *SessionValidator) SetMaxSessionDuration(max int) {
	sv.MaxSessionDuration = max
}
