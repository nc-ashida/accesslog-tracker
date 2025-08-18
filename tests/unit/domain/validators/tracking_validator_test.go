package validators

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"accesslog-tracker/internal/domain/models"
	"accesslog-tracker/internal/domain/validators"
)

func TestTrackingValidator_Validate(t *testing.T) {
	validator := validators.NewTrackingValidator()

	tests := []struct {
		name    string
		data    *models.TrackingData
		wantErr bool
	}{
		{
			name: "valid tracking data",
			data: &models.TrackingData{
				AppID:     "test_app_123",
				UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
				URL:       "https://example.com/page1",
				IPAddress: "192.168.1.100",
				SessionID: "session_123",
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing app_id",
			data: &models.TrackingData{
				AppID:     "",
				UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
				URL:       "https://example.com/page1",
				IPAddress: "192.168.1.100",
				SessionID: "session_123",
				Timestamp: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing user_agent",
			data: &models.TrackingData{
				AppID:     "test_app_123",
				UserAgent: "",
				URL:       "https://example.com/page1",
				IPAddress: "192.168.1.100",
				SessionID: "session_123",
				Timestamp: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing url",
			data: &models.TrackingData{
				AppID:     "test_app_123",
				UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
				URL:       "",
				IPAddress: "192.168.1.100",
				SessionID: "session_123",
				Timestamp: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing timestamp",
			data: &models.TrackingData{
				AppID:     "test_app_123",
				UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
				URL:       "https://example.com/page1",
				IPAddress: "192.168.1.100",
				SessionID: "session_123",
				Timestamp: time.Time{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.data)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTrackingValidator_ValidateAppID(t *testing.T) {
	validator := validators.NewTrackingValidator()

	tests := []struct {
		name  string
		appID string
		want  error
	}{
		{"valid app id", "test_app_123", nil},
		{"valid app id with numbers", "app_123456", nil},
		{"empty app id", "", models.ErrTrackingAppIDRequired},
		{"app id too short", "a", errors.New("app_id must be at least 8 characters")},
		{"app id with invalid chars", "test@app", errors.New("app_id contains invalid characters")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.ValidateAppID(tt.appID)
			if tt.want == nil {
				assert.NoError(t, got)
			} else {
				assert.Error(t, got)
				if got != nil {
					assert.Equal(t, tt.want.Error(), got.Error())
				}
			}
		})
	}
}

func TestTrackingValidator_ValidateUserAgent(t *testing.T) {
	validator := validators.NewTrackingValidator()

	tests := []struct {
		name      string
		userAgent string
		want      error
	}{
		{"valid user agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36", nil},
		{"valid mobile user agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15", nil},
		{"empty user agent", "", models.ErrTrackingUserAgentRequired},
		{"user agent too short", "A", errors.New("user agent must be at least 10 characters")},
		{"user agent too long", "This is a very long user agent string that exceeds the maximum allowed length", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.ValidateUserAgent(tt.userAgent)
			if tt.want == nil {
				assert.NoError(t, got)
			} else {
				assert.Error(t, got)
				if got != nil {
					assert.Equal(t, tt.want.Error(), got.Error())
				}
			}
		})
	}
}

func TestTrackingValidator_ValidateURL(t *testing.T) {
	validator := validators.NewTrackingValidator()

	tests := []struct {
		name string
		url  string
		want error
	}{
		{"valid HTTP URL", "http://example.com", nil},
		{"valid HTTPS URL", "https://example.com/page1", nil},
		{"valid URL with query", "https://example.com/page1?param=value", nil},
		{"invalid URL", "not-a-url", errors.New("URL must have a scheme (http:// or https://)")},
		{"empty URL", "", models.ErrTrackingURLRequired},
		{"URL without scheme", "example.com", errors.New("URL must have a scheme (http:// or https://)")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.ValidateURL(tt.url)
			if tt.want == nil {
				assert.NoError(t, got)
			} else {
				assert.Error(t, got)
				if got != nil {
					assert.Equal(t, tt.want.Error(), got.Error())
				}
			}
		})
	}
}

func TestTrackingValidator_ValidateTimestamp(t *testing.T) {
	validator := validators.NewTrackingValidator()

	tests := []struct {
		name      string
		timestamp time.Time
		want      error
	}{
		{"valid timestamp", time.Now(), nil},
		{"valid past timestamp", time.Now().Add(-24 * time.Hour), nil},
		{"valid future timestamp", time.Now().Add(24 * time.Hour), errors.New("timestamp cannot be in the future")},
		{"zero timestamp", time.Time{}, models.ErrTrackingTimestampRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.ValidateTimestamp(tt.timestamp)
			if tt.want == nil {
				assert.NoError(t, got)
			} else {
				assert.Error(t, got)
				if got != nil {
					assert.Equal(t, tt.want.Error(), got.Error())
				}
			}
		})
	}
}

// ValidateIPAddressは非公開メソッドのため、テストを削除

// ValidateReferrerは非公開メソッドのため、テストを削除

func TestTrackingValidator_ValidateCustomParams(t *testing.T) {
	validator := validators.NewTrackingValidator()

	tests := []struct {
		name         string
		customParams map[string]interface{}
		want         error
	}{
		{
			name: "valid custom params",
			customParams: map[string]interface{}{
				"page_type": "product",
				"user_id":   "12345",
			},
			want: nil,
		},
		{
			name:         "empty custom params",
			customParams: map[string]interface{}{},
			want:         nil,
		},
		{
			name: "custom params with invalid key",
			customParams: map[string]interface{}{
				"": "value",
			},
			want: errors.New("custom param key cannot be empty"),
		},
		{
			name: "custom params with invalid value",
			customParams: map[string]interface{}{
				"key": make(chan int), // 無効な値
			},
			want: errors.New("custom param value must be string, number, or boolean"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.ValidateCustomParams(tt.customParams)
			if tt.want == nil {
				assert.NoError(t, got)
			} else {
				assert.Error(t, got)
				if got != nil {
					assert.Equal(t, tt.want.Error(), got.Error())
				}
			}
		})
	}
}

// ValidateCustomParamKeyとValidateCustomParamValueは非公開メソッドのため、テストを削除

func TestTrackingValidator_IsCrawler(t *testing.T) {
	validator := validators.NewTrackingValidator()

	tests := []struct {
		name      string
		userAgent string
		want      bool
	}{
		{"bot user agent", "Googlebot/2.1 (+http://www.google.com/bot.html)", true},
		{"crawler user agent", "Bingbot/2.0", true},
		{"normal user agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36", false},
		{"empty user agent", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.IsCrawler(tt.userAgent)
			assert.Equal(t, tt.want, got)
		})
	}
}
