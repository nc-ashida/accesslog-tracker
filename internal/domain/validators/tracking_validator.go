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

// TrackingValidator はトラッキングデータのバリデーター
type TrackingValidator struct {
	// バリデーション設定
	MaxCustomParams int
	MaxURLLength    int
	MaxUserAgentLength int
	BlockedIPs      []string
	AllowedDomains  []string
}

// NewTrackingValidator は新しいトラッキングバリデーターを作成
func NewTrackingValidator() *TrackingValidator {
	return &TrackingValidator{
		MaxCustomParams:    10,
		MaxURLLength:       2048,
		MaxUserAgentLength: 500,
		BlockedIPs:         []string{},
		AllowedDomains:     []string{},
	}
}

// ValidateTrackingRequest はトラッキングリクエストを検証
func (tv *TrackingValidator) ValidateTrackingRequest(req *models.TrackingRequest) error {
	var errors []string

	// 必須フィールドの検証
	if err := tv.validateRequiredFields(req); err != nil {
		errors = append(errors, err.Error())
	}

	// アプリケーションIDの検証
	if err := tv.validateAppID(req.AppID); err != nil {
		errors = append(errors, err.Error())
	}

	// URLの検証
	if err := tv.validateURL(req.URL); err != nil {
		errors = append(errors, err.Error())
	}

	// UserAgentの検証
	if err := tv.validateUserAgent(req.UserAgent); err != nil {
		errors = append(errors, err.Error())
	}

	// IPアドレスの検証
	if req.IPAddress != "" {
		if err := tv.validateIPAddress(req.IPAddress); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// 画面解像度の検証
	if req.ScreenResolution != "" {
		if err := tv.validateScreenResolution(req.ScreenResolution); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// 言語の検証
	if req.Language != "" {
		if err := tv.validateLanguage(req.Language); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// タイムゾーンの検証
	if req.Timezone != "" {
		if err := tv.validateTimezone(req.Timezone); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// カスタムパラメータの検証
	if len(req.CustomParams) > 0 {
		if err := tv.validateCustomParams(req.CustomParams); err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// validateRequiredFields は必須フィールドを検証
func (tv *TrackingValidator) validateRequiredFields(req *models.TrackingRequest) error {
	if req.AppID == "" {
		return fmt.Errorf("application_id is required")
	}
	if req.URL == "" {
		return fmt.Errorf("page_url is required")
	}
	if req.UserAgent == "" {
		return fmt.Errorf("user_agent is required")
	}
	return nil
}

// validateAppID はアプリケーションIDを検証
func (tv *TrackingValidator) validateAppID(appID string) error {
	if appID == "" {
		return fmt.Errorf("application_id cannot be empty")
	}

	// アプリケーションIDの形式検証（UUIDまたはカスタム形式）
	if !tv.isValidApplicationID(appID) {
		return fmt.Errorf("invalid application_id format")
	}

	return nil
}

// validateURL はページURLを検証
func (tv *TrackingValidator) validateURL(pageURL string) error {
	if pageURL == "" {
		return fmt.Errorf("page_url cannot be empty")
	}

	if len(pageURL) > tv.MaxURLLength {
		return fmt.Errorf("page_url length exceeds maximum limit of %d characters", tv.MaxURLLength)
	}

	// URL形式の検証
	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		return fmt.Errorf("invalid page_url format: %v", err)
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("page_url must have valid scheme and host")
	}

	// 許可されたドメインのチェック
	if len(tv.AllowedDomains) > 0 {
		if !tv.isAllowedDomain(parsedURL.Host) {
			return fmt.Errorf("page_url domain is not allowed")
		}
	}

	return nil
}

// validateReferrer はリファラーを検証
func (tv *TrackingValidator) validateReferrer(referrer string) error {
	if len(referrer) > tv.MaxURLLength {
		return fmt.Errorf("referrer length exceeds maximum limit of %d characters", tv.MaxURLLength)
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

// validateUserAgent はUserAgentを検証
func (tv *TrackingValidator) validateUserAgent(userAgent string) error {
	if userAgent == "" {
		return fmt.Errorf("user_agent cannot be empty")
	}

	if len(userAgent) > tv.MaxUserAgentLength {
		return fmt.Errorf("user_agent length exceeds maximum limit of %d characters", tv.MaxUserAgentLength)
	}

	// 基本的なUserAgent形式の検証
	if !tv.isValidUserAgent(userAgent) {
		return fmt.Errorf("invalid user_agent format")
	}

	return nil
}

// validateIPAddress はIPアドレスを検証
func (tv *TrackingValidator) validateIPAddress(ipAddress string) error {
	if !iputil.IsValidIP(ipAddress) {
		return fmt.Errorf("invalid ip_address format")
	}

	// ブロックされたIPのチェック
	if tv.isBlockedIP(ipAddress) {
		return fmt.Errorf("ip_address is blocked")
	}

	return nil
}

// validateScreenResolution は画面解像度を検証
func (tv *TrackingValidator) validateScreenResolution(resolution string) error {
	if resolution == "" {
		return nil
	}

	// 画面解像度の形式: "1920x1080"
	pattern := `^\d+x\d+$`
	matched, err := regexp.MatchString(pattern, resolution)
	if err != nil {
		return fmt.Errorf("failed to validate screen resolution: %w", err)
	}

	if !matched {
		return fmt.Errorf("invalid screen resolution format, expected: WIDTHxHEIGHT")
	}

	return nil
}

// validateScreenSize は画面サイズを検証
func (tv *TrackingValidator) validateScreenSize(width, height int) error {
	if width < 0 || height < 0 {
		return fmt.Errorf("screen dimensions cannot be negative")
	}

	if width > tv.MaxScreenWidth || height > tv.MaxScreenHeight {
		return fmt.Errorf("screen dimensions exceed maximum allowed size")
	}

	return nil
}

// validateViewportSize はビューポートサイズを検証
func (tv *TrackingValidator) validateViewportSize(width, height int) error {
	if width < 0 {
		return fmt.Errorf("viewport_width cannot be negative")
	}
	if height < 0 {
		return fmt.Errorf("viewport_height cannot be negative")
	}
	if width > 10000 {
		return fmt.Errorf("viewport_width exceeds maximum limit of 10000")
	}
	if height > 10000 {
		return fmt.Errorf("viewport_height exceeds maximum limit of 10000")
	}
	return nil
}

// validateCustomParams はカスタムパラメータを検証
func (tv *TrackingValidator) validateCustomParams(params map[string]interface{}) error {
	if params == nil {
		return nil
	}

	if len(params) > tv.MaxCustomParams {
		return fmt.Errorf("custom_params count exceeds maximum limit of %d", tv.MaxCustomParams)
	}

	for key, value := range params {
		// キーの検証
		if err := tv.validateCustomParamKey(key); err != nil {
			return err
		}

		// 値の検証
		if err := tv.validateCustomParamValue(key, value); err != nil {
			return err
		}
	}

	return nil
}

// validateCustomParamKey はカスタムパラメータのキーを検証
func (tv *TrackingValidator) validateCustomParamKey(key string) error {
	if key == "" {
		return fmt.Errorf("custom_param key cannot be empty")
	}

	if len(key) > 50 {
		return fmt.Errorf("custom_param key length exceeds maximum limit of 50 characters")
	}

	// キーの形式検証（英数字、アンダースコア、ハイフンのみ許可）
	keyRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !keyRegex.MatchString(key) {
		return fmt.Errorf("custom_param key contains invalid characters")
	}

	return nil
}

// validateCustomParamValue はカスタムパラメータの値を検証
func (tv *TrackingValidator) validateCustomParamValue(key string, value interface{}) error {
	if value == nil {
		return nil // null値は許可
	}

	switch v := value.(type) {
	case string:
		if len(v) > 1000 {
			return fmt.Errorf("custom_param value for key '%s' exceeds maximum length of 1000 characters", key)
		}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		// 数値は制限なし
	case float32, float64:
		// 浮動小数点は制限なし
	case bool:
		// 真偽値は制限なし
	case []interface{}:
		if len(v) > 100 {
			return fmt.Errorf("custom_param array for key '%s' exceeds maximum length of 100 items", key)
		}
		// 配列の各要素も検証
		for i, item := range v {
			if err := tv.validateCustomParamValue(fmt.Sprintf("%s[%d]", key, i), item); err != nil {
				return err
			}
		}
	case map[string]interface{}:
		if len(v) > 50 {
			return fmt.Errorf("custom_param object for key '%s' exceeds maximum length of 50 items", key)
		}
		// オブジェクトの各プロパティも検証
		for subKey, subValue := range v {
			if err := tv.validateCustomParamValue(fmt.Sprintf("%s.%s", key, subKey), subValue); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("custom_param value for key '%s' has unsupported type", key)
	}

	return nil
}

// validateLanguage は言語コードを検証
func (tv *TrackingValidator) validateLanguage(language string) error {
	if language == "" {
		return nil
	}

	// 言語コードの形式検証（ISO 639-1形式）
	langRegex := regexp.MustCompile(`^[a-z]{2}(-[A-Z]{2})?$`)
	if !langRegex.MatchString(language) {
		return fmt.Errorf("invalid language format")
	}

	return nil
}

// validateTimezone はタイムゾーンを検証
func (tv *TrackingValidator) validateTimezone(timezone string) error {
	if timezone == "" {
		return nil
	}

	// タイムゾーンの形式検証
	_, err := time.LoadLocation(timezone)
	if err != nil {
		return fmt.Errorf("invalid timezone format: %v", err)
	}

	return nil
}

// isValidApplicationID はアプリケーションIDの形式を検証
func (tv *TrackingValidator) isValidApplicationID(appID string) bool {
	// UUID形式またはカスタム形式を許可
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	customRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]{1,50}$`)
	
	return uuidRegex.MatchString(appID) || customRegex.MatchString(appID)
}

// isValidSessionID はセッションIDの形式を検証
func (tv *TrackingValidator) isValidSessionID(sessionID string) bool {
	// UUID形式またはカスタム形式を許可
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	customRegex := regexp.MatchString(`^[a-zA-Z0-9_-]{1,100}$`, sessionID)
	
	return uuidRegex.MatchString(sessionID) || customRegex
}

// isValidUserAgent はUserAgentの形式を検証
func (tv *TrackingValidator) isValidUserAgent(userAgent string) bool {
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

// isAllowedDomain はドメインが許可されているかどうかをチェック
func (tv *TrackingValidator) isAllowedDomain(domain string) bool {
	if len(tv.AllowedDomains) == 0 {
		return true // 制限がない場合はすべて許可
	}
	
	for _, allowedDomain := range tv.AllowedDomains {
		if domain == allowedDomain || strings.HasSuffix(domain, "."+allowedDomain) {
			return true
		}
	}
	
	return false
}

// isBlockedIP はIPアドレスがブロックされているかどうかをチェック
func (tv *TrackingValidator) isBlockedIP(ipAddress string) bool {
	for _, blockedIP := range tv.BlockedIPs {
		if ipAddress == blockedIP {
			return true
		}
	}
	
	return false
}

// SetMaxCustomParams はカスタムパラメータの最大数を設定
func (tv *TrackingValidator) SetMaxCustomParams(max int) {
	tv.MaxCustomParams = max
}

// SetMaxURLLength はURLの最大長を設定
func (tv *TrackingValidator) SetMaxURLLength(max int) {
	tv.MaxURLLength = max
}

// SetMaxUserAgentLength はUserAgentの最大長を設定
func (tv *TrackingValidator) SetMaxUserAgentLength(max int) {
	tv.MaxUserAgentLength = max
}

// SetBlockedIPs はブロックされたIPリストを設定
func (tv *TrackingValidator) SetBlockedIPs(ips []string) {
	tv.BlockedIPs = ips
}

// SetAllowedDomains は許可されたドメインリストを設定
func (tv *TrackingValidator) SetAllowedDomains(domains []string) {
	tv.AllowedDomains = domains
}
