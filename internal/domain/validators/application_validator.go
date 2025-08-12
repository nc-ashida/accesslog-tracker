package validators

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/your-username/accesslog-tracker/internal/domain/models"
)

// ApplicationValidator はアプリケーションのバリデーター
type ApplicationValidator struct {
	// バリデーション設定
	MaxNameLength        int
	MaxDescriptionLength int
	MaxDomainLength      int
	MaxURLLength         int
	MaxTagsCount         int
	MaxTagLength         int
	MaxMetadataCount     int
	MaxMetadataKeyLength int
	MaxMetadataValueLength int
}

// NewApplicationValidator は新しいアプリケーションバリデーターを作成
func NewApplicationValidator() *ApplicationValidator {
	return &ApplicationValidator{
		MaxNameLength:        100,
		MaxDescriptionLength: 500,
		MaxDomainLength:      253,
		MaxURLLength:         2048,
		MaxTagsCount:         20,
		MaxTagLength:         50,
		MaxMetadataCount:     50,
		MaxMetadataKeyLength: 100,
		MaxMetadataValueLength: 1000,
	}
}

// ValidateApplicationRequest はアプリケーションリクエストを検証
func (av *ApplicationValidator) ValidateApplicationRequest(req *models.ApplicationRequest) error {
	var errors []string

	// 必須フィールドの検証
	if err := av.validateRequiredFields(req); err != nil {
		errors = append(errors, err.Error())
	}

	// 名前の検証
	if err := av.validateName(req.Name); err != nil {
		errors = append(errors, err.Error())
	}

	// 説明の検証
	if req.Description != "" {
		if err := av.validateDescription(req.Description); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// ドメインの検証
	if err := av.validateDomain(req.Domain); err != nil {
		errors = append(errors, err.Error())
	}

	// URLの検証
	if err := av.validateURL(req.URL); err != nil {
		errors = append(errors, err.Error())
	}

	// 設定の検証
	if req.Settings != nil {
		if err := av.validateSettings(req.Settings); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// タグの検証
	if len(req.Tags) > 0 {
		if err := av.validateTags(req.Tags); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// メタデータの検証
	if len(req.Metadata) > 0 {
		if err := av.validateMetadata(req.Metadata); err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// ValidateApplication はアプリケーションを検証
func (av *ApplicationValidator) ValidateApplication(app *models.Application) error {
	var errors []string

	// 必須フィールドの検証
	if app.ID == "" {
		errors = append(errors, "id is required")
	}
	if app.Name == "" {
		errors = append(errors, "name is required")
	}
	if app.Domain == "" {
		errors = append(errors, "domain is required")
	}
	if app.URL == "" {
		errors = append(errors, "url is required")
	}
	if app.OwnerID == "" {
		errors = append(errors, "owner_id is required")
	}

	// 名前の検証
	if err := av.validateName(app.Name); err != nil {
		errors = append(errors, err.Error())
	}

	// 説明の検証
	if app.Description != "" {
		if err := av.validateDescription(app.Description); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// ドメインの検証
	if err := av.validateDomain(app.Domain); err != nil {
		errors = append(errors, err.Error())
	}

	// URLの検証
	if err := av.validateURL(app.URL); err != nil {
		errors = append(errors, err.Error())
	}

	// ステータスの検証
	if err := av.validateStatus(app.Status); err != nil {
		errors = append(errors, err.Error())
	}

	// 設定の検証
	if err := av.validateSettings(&app.Settings); err != nil {
		errors = append(errors, err.Error())
	}

	// タグの検証
	if len(app.Tags) > 0 {
		if err := av.validateTags(app.Tags); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// メタデータの検証
	if len(app.Metadata) > 0 {
		if err := av.validateMetadata(app.Metadata); err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// validateRequiredFields は必須フィールドを検証
func (av *ApplicationValidator) validateRequiredFields(req *models.ApplicationRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Domain == "" {
		return fmt.Errorf("domain is required")
	}
	if req.URL == "" {
		return fmt.Errorf("url is required")
	}
	return nil
}

// validateName は名前を検証
func (av *ApplicationValidator) validateName(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if len(name) > av.MaxNameLength {
		return fmt.Errorf("name length exceeds maximum limit of %d characters", av.MaxNameLength)
	}

	// 名前の形式検証（英数字、スペース、ハイフン、アンダースコアのみ許可）
	nameRegex := regexp.MustCompile(`^[a-zA-Z0-9\s\-_]+$`)
	if !nameRegex.MatchString(name) {
		return fmt.Errorf("name contains invalid characters")
	}

	return nil
}

// validateDescription は説明を検証
func (av *ApplicationValidator) validateDescription(description string) error {
	if len(description) > av.MaxDescriptionLength {
		return fmt.Errorf("description length exceeds maximum limit of %d characters", av.MaxDescriptionLength)
	}

	return nil
}

// validateDomain はドメインを検証
func (av *ApplicationValidator) validateDomain(domain string) error {
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	if len(domain) > av.MaxDomainLength {
		return fmt.Errorf("domain length exceeds maximum limit of %d characters", av.MaxDomainLength)
	}

	// ドメイン形式の検証
	if !av.isValidDomain(domain) {
		return fmt.Errorf("invalid domain format")
	}

	return nil
}

// validateURL はURLを検証
func (av *ApplicationValidator) validateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("url cannot be empty")
	}

	if len(urlStr) > av.MaxURLLength {
		return fmt.Errorf("url length exceeds maximum limit of %d characters", av.MaxURLLength)
	}

	// URL形式の検証
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid url format: %v", err)
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("url must have valid scheme and host")
	}

	// サポートされているスキームのチェック
	supportedSchemes := []string{"http", "https"}
	schemeValid := false
	for _, scheme := range supportedSchemes {
		if parsedURL.Scheme == scheme {
			schemeValid = true
			break
		}
	}
	if !schemeValid {
		return fmt.Errorf("url scheme must be http or https")
	}

	return nil
}

// validateStatus はステータスを検証
func (av *ApplicationValidator) validateStatus(status string) error {
	validStatuses := []string{
		models.AppStatusActive,
		models.AppStatusInactive,
		models.AppStatusSuspended,
		models.AppStatusDeleted,
	}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			return nil
		}
	}

	return fmt.Errorf("invalid status: %s", status)
}

// validateSettings は設定を検証
func (av *ApplicationValidator) validateSettings(settings *models.ApplicationSettings) error {
	// セッションタイムアウトの検証
	if settings.SessionTimeout < 60 || settings.SessionTimeout > 86400 {
		return fmt.Errorf("session_timeout must be between 60 and 86400 seconds")
	}

	// データ保持期間の検証
	if settings.DataRetentionDays < 1 || settings.DataRetentionDays > 3650 {
		return fmt.Errorf("data_retention_days must be between 1 and 3650 days")
	}

	// カスタムパラメータの最大数の検証
	if settings.MaxCustomParams < 1 || settings.MaxCustomParams > 100 {
		return fmt.Errorf("max_custom_params must be between 1 and 100")
	}

	// Webhook URLの検証
	if settings.WebhookEnabled && settings.WebhookURL != "" {
		if err := av.validateURL(settings.WebhookURL); err != nil {
			return fmt.Errorf("invalid webhook_url: %v", err)
		}
	}

	// ブロックされたIPの検証
	for _, ip := range settings.BlockedIPs {
		if !av.isValidIP(ip) {
			return fmt.Errorf("invalid blocked_ip: %s", ip)
		}
	}

	// 許可されたドメインの検証
	for _, domain := range settings.AllowedDomains {
		if !av.isValidDomain(domain) {
			return fmt.Errorf("invalid allowed_domain: %s", domain)
		}
	}

	return nil
}

// validateTags はタグを検証
func (av *ApplicationValidator) validateTags(tags []string) error {
	if len(tags) > av.MaxTagsCount {
		return fmt.Errorf("tags count exceeds maximum limit of %d", av.MaxTagsCount)
	}

	for i, tag := range tags {
		if tag == "" {
			return fmt.Errorf("tag at index %d cannot be empty", i)
		}

		if len(tag) > av.MaxTagLength {
			return fmt.Errorf("tag at index %d length exceeds maximum limit of %d characters", i, av.MaxTagLength)
		}

		// タグの形式検証（英数字、ハイフン、アンダースコアのみ許可）
		tagRegex := regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)
		if !tagRegex.MatchString(tag) {
			return fmt.Errorf("tag at index %d contains invalid characters", i)
		}
	}

	return nil
}

// validateMetadata はメタデータを検証
func (av *ApplicationValidator) validateMetadata(metadata map[string]string) error {
	if len(metadata) > av.MaxMetadataCount {
		return fmt.Errorf("metadata count exceeds maximum limit of %d", av.MaxMetadataCount)
	}

	for key, value := range metadata {
		// キーの検証
		if key == "" {
			return fmt.Errorf("metadata key cannot be empty")
		}

		if len(key) > av.MaxMetadataKeyLength {
			return fmt.Errorf("metadata key length exceeds maximum limit of %d characters", av.MaxMetadataKeyLength)
		}

		// キーの形式検証（英数字、ハイフン、アンダースコアのみ許可）
		keyRegex := regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)
		if !keyRegex.MatchString(key) {
			return fmt.Errorf("metadata key contains invalid characters: %s", key)
		}

		// 値の検証
		if len(value) > av.MaxMetadataValueLength {
			return fmt.Errorf("metadata value for key '%s' length exceeds maximum limit of %d characters", key, av.MaxMetadataValueLength)
		}
	}

	return nil
}

// isValidDomain はドメインの形式を検証
func (av *ApplicationValidator) isValidDomain(domain string) bool {
	// 基本的なドメイン形式の検証
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	return domainRegex.MatchString(domain)
}

// isValidIP はIPアドレスの形式を検証
func (av *ApplicationValidator) isValidIP(ip string) bool {
	// IPv4形式の検証
	ipv4Regex := regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
	if ipv4Regex.MatchString(ip) {
		parts := strings.Split(ip, ".")
		for _, part := range parts {
			if len(part) > 1 && part[0] == '0' {
				return false
			}
			if part == "" || len(part) > 3 {
				return false
			}
			if num := 0; num < 0 || num > 255 {
				return false
			}
		}
		return true
	}

	// IPv6形式の検証（簡易版）
	ipv6Regex := regexp.MustCompile(`^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$`)
	return ipv6Regex.MatchString(ip)
}

// SetMaxNameLength は名前の最大長を設定
func (av *ApplicationValidator) SetMaxNameLength(max int) {
	av.MaxNameLength = max
}

// SetMaxDescriptionLength は説明の最大長を設定
func (av *ApplicationValidator) SetMaxDescriptionLength(max int) {
	av.MaxDescriptionLength = max
}

// SetMaxDomainLength はドメインの最大長を設定
func (av *ApplicationValidator) SetMaxDomainLength(max int) {
	av.MaxDomainLength = max
}

// SetMaxURLLength はURLの最大長を設定
func (av *ApplicationValidator) SetMaxURLLength(max int) {
	av.MaxURLLength = max
}

// SetMaxTagsCount はタグの最大数を設定
func (av *ApplicationValidator) SetMaxTagsCount(max int) {
	av.MaxTagsCount = max
}

// SetMaxTagLength はタグの最大長を設定
func (av *ApplicationValidator) SetMaxTagLength(max int) {
	av.MaxTagLength = max
}

// SetMaxMetadataCount はメタデータの最大数を設定
func (av *ApplicationValidator) SetMaxMetadataCount(max int) {
	av.MaxMetadataCount = max
}

// SetMaxMetadataKeyLength はメタデータキーの最大長を設定
func (av *ApplicationValidator) SetMaxMetadataKeyLength(max int) {
	av.MaxMetadataKeyLength = max
}

// SetMaxMetadataValueLength はメタデータ値の最大長を設定
func (av *ApplicationValidator) SetMaxMetadataValueLength(max int) {
	av.MaxMetadataValueLength = max
}
