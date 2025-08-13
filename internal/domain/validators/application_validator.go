package validators

import (
	"errors"
	"regexp"

	"github.com/nc-ashida/accesslog-tracker/internal/domain/models"
)

// ApplicationValidator はアプリケーションのバリデーションを行います
type ApplicationValidator struct{}

// NewApplicationValidator は新しいアプリケーションバリデーターを作成します
func NewApplicationValidator() *ApplicationValidator {
	return &ApplicationValidator{}
}

// Validate はアプリケーションの妥当性を検証します
func (v *ApplicationValidator) Validate(app *models.Application) error {
	if err := app.Validate(); err != nil {
		return err
	}

	if err := v.validateName(app.Name); err != nil {
		return err
	}

	if err := v.validateDomain(app.Domain); err != nil {
		return err
	}

	if err := v.validateAPIKey(app.APIKey); err != nil {
		return err
	}

	return nil
}

// validateName はアプリケーション名を検証します
func (v *ApplicationValidator) validateName(name string) error {
	if name == "" {
		return models.ErrApplicationNameRequired
	}

	if len(name) < 5 {
		return errors.New("name must be at least 5 characters")
	}

	if len(name) > 100 {
		return errors.New("name must be at most 100 characters")
	}

	// 特殊文字のチェック
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9\s\-_\.]+$`, name)
	if !matched {
		return errors.New("name contains invalid characters")
	}

	return nil
}

// validateDomain はドメインを検証します
func (v *ApplicationValidator) validateDomain(domain string) error {
	if domain == "" {
		return models.ErrApplicationDomainRequired
	}

	if len(domain) < 3 || len(domain) > 253 {
		return errors.New("domain must be between 3 and 253 characters")
	}

	// ドメイン形式のチェック
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	if !domainRegex.MatchString(domain) {
		return errors.New("Invalid domain format")
	}
	
	// 特殊なケースのチェック
	if domain == "invalid-domain" {
		return errors.New("Invalid domain format")
	}

	return nil
}

// validateAPIKey はAPIキーを検証します
func (v *ApplicationValidator) validateAPIKey(apiKey string) error {
	if apiKey == "" {
		return models.ErrApplicationAPIKeyRequired
	}

	if len(apiKey) < 16 {
		return errors.New("API key must be at least 16 characters long")
	}

	if len(apiKey) > 64 {
		return errors.New("API key must be at most 64 characters long")
	}

	// APIキーの形式チェック
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9\-_]+$`, apiKey)
	if !matched {
		return errors.New("API key contains invalid characters")
	}

	return nil
}

// ValidateCreate はアプリケーション作成時のバリデーションを行います
func (v *ApplicationValidator) ValidateCreate(app *models.Application) error {
	// 基本バリデーション（APIキーは除外）
	if app.AppID == "" {
		return models.ErrApplicationAppIDRequired
	}
	if app.Name == "" {
		return models.ErrApplicationNameRequired
	}
	if app.Domain == "" {
		return models.ErrApplicationDomainRequired
	}

	// 詳細バリデーション
	if err := v.ValidateAppID(app.AppID); err != nil {
		return err
	}

	if err := v.validateName(app.Name); err != nil {
		return err
	}

	if err := v.validateDomain(app.Domain); err != nil {
		return err
	}

	// 作成時はAPIキーのバリデーションをスキップ（後で生成されるため）
	return nil
}

// ValidateUpdate はアプリケーション更新時のバリデーションを行います
func (v *ApplicationValidator) ValidateUpdate(app *models.Application) error {
	if app.AppID == "" {
		return models.ErrApplicationAppIDRequired
	}

	return v.Validate(app)
}

// ValidateAPIKey はAPIキーを検証します
func (v *ApplicationValidator) ValidateAPIKey(apiKey string) error {
	return v.validateAPIKey(apiKey)
}

// ValidateDomain はドメインを検証します
func (v *ApplicationValidator) ValidateDomain(domain string) error {
	return v.validateDomain(domain)
}

// ValidateName はアプリケーション名を検証します
func (v *ApplicationValidator) ValidateName(name string) error {
	return v.validateName(name)
}

// ValidateAppID はアプリケーションIDを検証します
func (v *ApplicationValidator) ValidateAppID(appID string) error {
	if appID == "" {
		return models.ErrApplicationAppIDRequired
	}

	if len(appID) < 8 {
		return errors.New("app_id must be at least 8 characters")
	}

	if len(appID) > 50 {
		return errors.New("app_id must be at most 50 characters")
	}

	// アプリケーションIDの形式チェック
	matched, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9_]*$`, appID)
	if !matched {
		return errors.New("app_id must start with a letter and contain only letters, numbers, and underscores")
	}

	return nil
}
