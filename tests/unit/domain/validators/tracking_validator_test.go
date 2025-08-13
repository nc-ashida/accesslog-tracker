package validators_test

import (
	"fmt"
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
		data    models.TrackingData
		isValid bool
		errors  []string
	}{
		{
			name: "valid tracking data",
			data: models.TrackingData{
				AppID:     "test_app_123",
				UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
				URL:       "https://example.com",
				Timestamp: time.Now(),
			},
			isValid: true,
			errors:  []string{},
		},
		{
			name: "missing app_id",
			data: models.TrackingData{
				UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
				URL:       "https://example.com",
				Timestamp: time.Now(),
			},
			isValid: false,
			errors:  []string{"app_id is required"},
		},
		{
			name: "missing user_agent",
			data: models.TrackingData{
				AppID:     "test_app_123",
				URL:       "https://example.com",
				Timestamp: time.Now(),
			},
			isValid: false,
			errors:  []string{"user_agent is required"},
		},
		{
			name: "invalid URL format",
			data: models.TrackingData{
				AppID:     "test_app_123",
				UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
				URL:       "invalid-url",
				Timestamp: time.Now(),
			},
			isValid: false,
			errors:  []string{"Invalid URL format"},
		},
		{
			name: "empty URL",
			data: models.TrackingData{
				AppID:     "test_app_123",
				UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
				URL:       "",
				Timestamp: time.Now(),
			},
			isValid: false,
			errors:  []string{"url is required"},
		},
		{
			name: "zero timestamp",
			data: models.TrackingData{
				AppID:     "test_app_123",
				UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
				URL:       "https://example.com",
				Timestamp: time.Time{},
			},
			isValid: false,
			errors:  []string{"timestamp is required"},
		},
		{
			name: "future timestamp",
			data: models.TrackingData{
				AppID:     "test_app_123",
				UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
				URL:       "https://example.com",
				Timestamp: time.Now().Add(24 * time.Hour),
			},
			isValid: false,
			errors:  []string{"timestamp cannot be in the future"},
		},
		{
			name: "very old timestamp",
			data: models.TrackingData{
				AppID:     "test_app_123",
				UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
				URL:       "https://example.com",
				Timestamp: time.Now().AddDate(-10, 0, 0),
			},
			isValid: false,
			errors:  []string{"timestamp is too old"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Validate(&tt.data)
			if tt.isValid {
				assert.NoError(t, result)
			} else {
				assert.Error(t, result)
				for _, expectedError := range tt.errors {
					assert.Contains(t, result.Error(), expectedError)
				}
			}
		})
	}
}

func TestTrackingValidator_IsCrawler(t *testing.T) {
	validator := validators.NewTrackingValidator()

	tests := []struct {
		name      string
		userAgent string
		expected  bool
	}{
		{
			name:      "detect Googlebot",
			userAgent: "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
			expected:  true,
		},
		{
			name:      "detect Bingbot",
			userAgent: "Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)",
			expected:  true,
		},
		{
			name:      "detect YandexBot",
			userAgent: "Mozilla/5.0 (compatible; YandexBot/3.0; +http://yandex.com/bots)",
			expected:  true,
		},
		{
			name:      "detect Baiduspider",
			userAgent: "Mozilla/5.0 (compatible; Baiduspider/2.0; +http://www.baidu.com/search/spider.html)",
			expected:  true,
		},
		{
			name:      "detect DuckDuckBot",
			userAgent: "DuckDuckBot/1.0; (+http://duckduckgo.com/duckduckbot.html)",
			expected:  true,
		},
		{
			name:      "regular browser",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			expected:  false,
		},
		{
			name:      "mobile browser",
			userAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15",
			expected:  false,
		},
		{
			name:      "empty user agent",
			userAgent: "",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.IsCrawler(tt.userAgent)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrackingValidator_ValidateURL(t *testing.T) {
	validator := validators.NewTrackingValidator()

	tests := []struct {
		name        string
		url         string
		expectValid bool
	}{
		{
			name:        "valid HTTPS URL",
			url:         "https://example.com",
			expectValid: true,
		},
		{
			name:        "valid HTTP URL",
			url:         "http://example.com",
			expectValid: true,
		},
		{
			name:        "valid URL with path",
			url:         "https://example.com/path/to/page",
			expectValid: true,
		},
		{
			name:        "valid URL with query parameters",
			url:         "https://example.com/page?param1=value1&param2=value2",
			expectValid: true,
		},
		{
			name:        "valid URL with fragment",
			url:         "https://example.com/page#section",
			expectValid: true,
		},
		{
			name:        "invalid URL format",
			url:         "invalid-url",
			expectValid: false,
		},
		{
			name:        "empty URL",
			url:         "",
			expectValid: false,
		},
		{
			name:        "URL with invalid protocol",
			url:         "ftp://example.com",
			expectValid: false,
		},
		{
			name:        "URL with invalid characters",
			url:         "https://example.com/page<script>alert('xss')</script>",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateURL(tt.url)
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestTrackingValidator_ValidateUserAgent(t *testing.T) {
	validator := validators.NewTrackingValidator()

	tests := []struct {
		name        string
		userAgent   string
		expectValid bool
	}{
		{
			name:        "valid user agent",
			userAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			expectValid: true,
		},
		{
			name:        "empty user agent",
			userAgent:   "",
			expectValid: false,
		},
		{
			name:        "user agent too short",
			userAgent:   "Short",
			expectValid: false,
		},
		{
			name:        "user agent too long",
			userAgent:   "This is a very long user agent string that exceeds the maximum allowed length and should be rejected because it is too long for the database field",
			expectValid: false,
		},
		{
			name:        "user agent with null bytes",
			userAgent:   "Mozilla/5.0\x00(Windows NT 10.0; Win64; x64)",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateUserAgent(tt.userAgent)
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestTrackingValidator_ValidateTimestamp(t *testing.T) {
	validator := validators.NewTrackingValidator()

	now := time.Now()
	tests := []struct {
		name        string
		timestamp   time.Time
		expectValid bool
	}{
		{
			name:        "valid timestamp",
			timestamp:   now,
			expectValid: true,
		},
		{
			name:        "zero timestamp",
			timestamp:   time.Time{},
			expectValid: false,
		},
		{
			name:        "future timestamp",
			timestamp:   now.Add(24 * time.Hour),
			expectValid: false,
		},
		{
			name:        "very old timestamp",
			timestamp:   now.AddDate(-10, 0, 0),
			expectValid: false,
		},
		{
			name:        "recent past timestamp",
			timestamp:   now.Add(-1 * time.Hour),
			expectValid: true,
		},
		{
			name:        "timestamp within allowed range",
			timestamp:   now.AddDate(-1, 0, 0),
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTimestamp(tt.timestamp)
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestTrackingValidator_ValidateAppID(t *testing.T) {
	validator := validators.NewTrackingValidator()

	tests := []struct {
		name        string
		appID       string
		expectValid bool
	}{
		{
			name:        "valid app_id",
			appID:       "test_app_123",
			expectValid: true,
		},
		{
			name:        "empty app_id",
			appID:       "",
			expectValid: false,
		},
		{
			name:        "app_id too short",
			appID:       "short",
			expectValid: false,
		},
		{
			name:        "app_id too long",
			appID:       "this_is_a_very_long_application_id_that_exceeds_the_maximum_allowed_length_and_should_be_rejected",
			expectValid: false,
		},
		{
			name:        "app_id with invalid characters",
			appID:       "test-app-123!@#",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateAppID(tt.appID)
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestTrackingValidator_ValidateCustomParams(t *testing.T) {
	validator := validators.NewTrackingValidator()

	tests := []struct {
		name        string
		customParams map[string]interface{}
		expectValid bool
	}{
		{
			name: "valid custom params",
			customParams: map[string]interface{}{
				"campaign_id": "camp_123",
				"source":      "google",
				"medium":      "cpc",
			},
			expectValid: true,
		},
		{
			name:        "empty custom params",
			customParams: map[string]interface{}{},
			expectValid: true,
		},
		{
			name:        "nil custom params",
			customParams: nil,
			expectValid: true,
		},
		{
			name: "too many custom params",
			customParams: func() map[string]interface{} {
				params := make(map[string]interface{})
				for i := 0; i < 51; i++ {
					params[fmt.Sprintf("param_%d", i)] = fmt.Sprintf("value_%d", i)
				}
				return params
			}(),
			expectValid: false,
		},
		{
			name: "custom param key too long",
			customParams: map[string]interface{}{
				"this_is_a_very_long_parameter_key_that_exceeds_the_maximum_allowed_length_and_should_be_rejected": "value",
			},
			expectValid: false,
		},
		{
			name: "custom param value too long",
			customParams: map[string]interface{}{
				"key": "this_is_a_very_long_parameter_value_that_exceeds_the_maximum_allowed_length_and_should_be_rejected_because_it_is_too_long_for_the_database_field",
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCustomParams(tt.customParams)
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
