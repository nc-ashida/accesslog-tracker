package models

import (
	"encoding/json"
	"time"
)

// Application はアプリケーション情報を格納する構造体
type Application struct {
	ID          string            `json:"id" db:"id"`
	AppID       string            `json:"app_id" db:"app_id"`
	ClientSubID string            `json:"client_sub_id,omitempty" db:"client_sub_id"`
	ModuleID    string            `json:"module_id,omitempty" db:"module_id"`
	Name        string            `json:"name" db:"name"`
	Description string            `json:"description,omitempty" db:"description"`
	Domain      string            `json:"domain" db:"domain"`
	URL         string            `json:"url" db:"url"`
	Status      string            `json:"status" db:"status"`
	Settings    ApplicationSettings `json:"settings" db:"settings"`
	APIKey      string            `json:"api_key,omitempty" db:"api_key"`
	SecretKey   string            `json:"secret_key,omitempty" db:"secret_key"`
	OwnerID     string            `json:"owner_id" db:"owner_id"`
	TeamID      string            `json:"team_id,omitempty" db:"team_id"`
	Tags        []string          `json:"tags,omitempty" db:"tags"`
	Metadata    map[string]string `json:"metadata,omitempty" db:"metadata"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time        `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ApplicationSettings はアプリケーション設定
type ApplicationSettings struct {
	TrackingEnabled    bool     `json:"tracking_enabled" db:"tracking_enabled"`
	BotFiltering       bool     `json:"bot_filtering" db:"bot_filtering"`
	IPFiltering        bool     `json:"ip_filtering" db:"ip_filtering"`
	BlockedIPs         []string `json:"blocked_ips,omitempty" db:"blocked_ips"`
	AllowedDomains     []string `json:"allowed_domains,omitempty" db:"allowed_domains"`
	SessionTimeout     int      `json:"session_timeout" db:"session_timeout"`
	DataRetentionDays  int      `json:"data_retention_days" db:"data_retention_days"`
	MaxCustomParams    int      `json:"max_custom_params" db:"max_custom_params"`
	WebhookEnabled     bool     `json:"webhook_enabled" db:"webhook_enabled"`
	WebhookURL         string   `json:"webhook_url,omitempty" db:"webhook_url"`
	WebhookEvents      []string `json:"webhook_events,omitempty" db:"webhook_events"`
	PrivacyMode        bool     `json:"privacy_mode" db:"privacy_mode"`
	AnonymizeIP        bool     `json:"anonymize_ip" db:"anonymize_ip"`
	RespectDNT         bool     `json:"respect_dnt" db:"respect_dnt"`
	GDPRCompliant      bool     `json:"gdpr_compliant" db:"gdpr_compliant"`
	CCPACompliant      bool     `json:"ccpa_compliant" db:"ccpa_compliant"`
	CustomCSS          string   `json:"custom_css,omitempty" db:"custom_css"`
	CustomJS           string   `json:"custom_js,omitempty" db:"custom_js"`
}

// ApplicationRequest はアプリケーション作成・更新リクエスト
type ApplicationRequest struct {
	Name        string            `json:"name" validate:"required,min=1,max=100"`
	Description string            `json:"description,omitempty" validate:"max=500"`
	Domain      string            `json:"domain" validate:"required,url"`
	URL         string            `json:"url" validate:"required,url"`
	Settings    *ApplicationSettings `json:"settings,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// ApplicationResponse はアプリケーションAPIレスポンス
type ApplicationResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    *Application `json:"data,omitempty"`
	Error   *AppError   `json:"error,omitempty"`
}

// AppError はアプリケーションエラー情報
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

// ApplicationStats はアプリケーション統計情報
type ApplicationStats struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	TotalEvents     int64     `json:"total_events"`
	UniqueUsers     int64     `json:"unique_users"`
	UniqueSessions  int64     `json:"unique_sessions"`
	PageViews       int64     `json:"page_views"`
	BounceRate      float64   `json:"bounce_rate"`
	AvgSessionTime  float64   `json:"avg_session_time"`
	ConversionRate  float64   `json:"conversion_rate"`
	Revenue         float64   `json:"revenue"`
	LastActivity    time.Time `json:"last_activity"`
	Period          string    `json:"period"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
}

// ApplicationFilter はアプリケーションフィルタリング条件
type ApplicationFilter struct {
	OwnerID    string    `json:"owner_id,omitempty"`
	TeamID     string    `json:"team_id,omitempty"`
	Status     string    `json:"status,omitempty"`
	Domain     string    `json:"domain,omitempty"`
	Tags       []string  `json:"tags,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
	UpdatedAt  time.Time `json:"updated_at,omitempty"`
	Limit      int       `json:"limit,omitempty"`
	Offset     int       `json:"offset,omitempty"`
	SortBy     string    `json:"sort_by,omitempty"`
	SortOrder  string    `json:"sort_order,omitempty"`
}

// ApplicationStatus はアプリケーションのステータス定数
const (
	AppStatusActive   = "active"
	AppStatusInactive = "inactive"
	AppStatusSuspended = "suspended"
	AppStatusDeleted  = "deleted"
)

// NewApplication は新しいアプリケーションを作成
func NewApplication(req *ApplicationRequest, ownerID string) *Application {
	now := time.Now()
	
	app := &Application{
		Name:        req.Name,
		Description: req.Description,
		Domain:      req.Domain,
		URL:         req.URL,
		Status:      AppStatusActive,
		OwnerID:     ownerID,
		Tags:        req.Tags,
		Metadata:    req.Metadata,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	
	// デフォルト設定を適用
	if req.Settings != nil {
		app.Settings = *req.Settings
	} else {
		app.Settings = DefaultApplicationSettings()
	}
	
	return app
}

// DefaultApplicationSettings はデフォルトのアプリケーション設定を返す
func DefaultApplicationSettings() ApplicationSettings {
	return ApplicationSettings{
		TrackingEnabled:   true,
		BotFiltering:      true,
		IPFiltering:       false,
		SessionTimeout:    1800, // 30分
		DataRetentionDays: 365,  // 1年
		MaxCustomParams:   10,
		WebhookEnabled:    false,
		PrivacyMode:       false,
		AnonymizeIP:       false,
		RespectDNT:        true,
		GDPRCompliant:     false,
		CCPACompliant:     false,
	}
}

// ToJSON はアプリケーションをJSONに変換
func (a *Application) ToJSON() ([]byte, error) {
	return json.Marshal(a)
}

// FromJSON はJSONからアプリケーションを作成
func (a *Application) FromJSON(data []byte) error {
	return json.Unmarshal(data, a)
}

// IsActive はアプリケーションがアクティブかどうかを判定
func (a *Application) IsActive() bool {
	return a.Status == AppStatusActive
}

// IsDeleted はアプリケーションが削除されているかどうかを判定
func (a *Application) IsDeleted() bool {
	return a.Status == AppStatusDeleted || a.DeletedAt != nil
}

// Activate はアプリケーションをアクティブにする
func (a *Application) Activate() {
	a.Status = AppStatusActive
	a.UpdatedAt = time.Now()
}

// Deactivate はアプリケーションを非アクティブにする
func (a *Application) Deactivate() {
	a.Status = AppStatusInactive
	a.UpdatedAt = time.Now()
}

// Suspend はアプリケーションを一時停止する
func (a *Application) Suspend() {
	a.Status = AppStatusSuspended
	a.UpdatedAt = time.Now()
}

// Delete はアプリケーションを削除する
func (a *Application) Delete() {
	a.Status = AppStatusDeleted
	now := time.Now()
	a.DeletedAt = &now
	a.UpdatedAt = now
}

// Restore は削除されたアプリケーションを復元する
func (a *Application) Restore() {
	a.Status = AppStatusActive
	a.DeletedAt = nil
	a.UpdatedAt = time.Now()
}

// UpdateSettings はアプリケーション設定を更新
func (a *Application) UpdateSettings(settings ApplicationSettings) {
	a.Settings = settings
	a.UpdatedAt = time.Now()
}

// AddTag はタグを追加
func (a *Application) AddTag(tag string) {
	if a.Tags == nil {
		a.Tags = make([]string, 0)
	}
	
	// 重複チェック
	for _, existingTag := range a.Tags {
		if existingTag == tag {
			return
		}
	}
	
	a.Tags = append(a.Tags, tag)
	a.UpdatedAt = time.Now()
}

// RemoveTag はタグを削除
func (a *Application) RemoveTag(tag string) {
	if a.Tags == nil {
		return
	}
	
	for i, existingTag := range a.Tags {
		if existingTag == tag {
			a.Tags = append(a.Tags[:i], a.Tags[i+1:]...)
			a.UpdatedAt = time.Now()
			return
		}
	}
}

// HasTag は指定されたタグを持っているかどうかを判定
func (a *Application) HasTag(tag string) bool {
	if a.Tags == nil {
		return false
	}
	
	for _, existingTag := range a.Tags {
		if existingTag == tag {
			return true
		}
	}
	
	return false
}

// SetMetadata はメタデータを設定
func (a *Application) SetMetadata(key, value string) {
	if a.Metadata == nil {
		a.Metadata = make(map[string]string)
	}
	
	a.Metadata[key] = value
	a.UpdatedAt = time.Now()
}

// GetMetadata はメタデータを取得
func (a *Application) GetMetadata(key string) string {
	if a.Metadata == nil {
		return ""
	}
	
	return a.Metadata[key]
}

// RemoveMetadata はメタデータを削除
func (a *Application) RemoveMetadata(key string) {
	if a.Metadata == nil {
		return
	}
	
	delete(a.Metadata, key)
	a.UpdatedAt = time.Now()
}

// ValidateDomain はドメインが有効かどうかを検証
func (a *Application) ValidateDomain() bool {
	// 基本的なURL検証
	if a.Domain == "" {
		return false
	}
	
	// ドメイン形式の検証（簡易版）
	if len(a.Domain) < 4 || len(a.Domain) > 253 {
		return false
	}
	
	return true
}

// GetTrackingCode はトラッキングコードを生成
func (a *Application) GetTrackingCode() string {
	if !a.IsActive() || !a.Settings.TrackingEnabled {
		return ""
	}
	
	// 実際の実装では、テンプレートエンジンを使用して
	// カスタマイズされたトラッキングコードを生成
	return generateTrackingCode(a)
}

// generateTrackingCode はトラッキングコードを生成（簡易版）
func generateTrackingCode(app *Application) string {
	// 実際の実装では、テンプレートファイルから生成
	return `
<script>
(function() {
    var script = document.createElement('script');
    script.src = 'https://api.example.com/tracker.js';
    script.async = true;
    script.setAttribute('data-app-id', '` + app.ID + `');
    document.head.appendChild(script);
})();
</script>`
}
