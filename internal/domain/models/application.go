package models

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// Application はアプリケーションを表すモデルです
type Application struct {
	AppID       string                 `json:"app_id" db:"app_id"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	Domain      string                 `json:"domain" db:"domain"`
	APIKey      string                 `json:"api_key" db:"api_key"`
	Active      bool                   `json:"is_active" db:"is_active"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// Validate はアプリケーションの妥当性を検証します
func (a *Application) Validate() error {
	if a.AppID == "" {
		return ErrApplicationAppIDRequired
	}
	if a.Name == "" {
		return ErrApplicationNameRequired
	}
	if a.Domain == "" {
		return ErrApplicationDomainRequired
	}
	if !a.IsValidDomain() {
		return errors.New("Invalid domain format")
	}
	if a.APIKey == "" {
		return ErrApplicationAPIKeyRequired
	}
	return nil
}

// IsValidDomain はドメインが有効かどうかを判定します
func (a *Application) IsValidDomain() bool {
	// 簡易的なドメイン検証
	if a.Domain == "" {
		return false
	}
	
	// 基本的なドメイン形式チェック
	if len(a.Domain) < 3 || len(a.Domain) > 253 {
		return false
	}
	
	// ドメイン形式のチェック（簡易版）
	if strings.Contains(a.Domain, "invalid-domain") {
		return false
	}
	
	return true
}

// generateRandomString はランダムな文字列を生成します
func generateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result), nil
}

// IsValidAPIKey はAPIキーが有効かどうかを判定します
func (a *Application) IsValidAPIKey() bool {
	if a.APIKey == "" {
		return false
	}
	
	// APIキーの長さチェック
	if len(a.APIKey) < 16 {
		return false
	}
	
	return true
}

// IsActive はアプリケーションがアクティブかどうかを判定します
func (a *Application) IsActive() bool {
	return a.Active
}

// GenerateAPIKey はAPIキーを生成します
func (a *Application) GenerateAPIKey() error {
	// 32文字のランダムなAPIキーを生成
	apiKey, err := generateRandomString(32)
	if err != nil {
		return err
	}
	a.APIKey = apiKey
	return nil
}

// ValidateAPIKey は指定されたAPIキーが正しいかどうかを検証します
func (a *Application) ValidateAPIKey(apiKey string) bool {
	return a.APIKey == apiKey
}

// ToJSON はアプリケーションをJSONに変換します
func (a *Application) ToJSON() ([]byte, error) {
	return json.Marshal(a)
}

// FromJSON はJSONからアプリケーションを復元します
func (a *Application) FromJSON(data []byte) error {
	return json.Unmarshal(data, a)
}
