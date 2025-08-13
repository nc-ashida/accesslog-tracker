package models_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"accesslog-tracker/internal/domain/models"
)

func TestTrackingData_Validate(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.data.Validate()
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				for _, expectedError := range tt.errors {
					assert.Contains(t, err.Error(), expectedError)
				}
			}
		})
	}
}

func TestTrackingData_ToJSON(t *testing.T) {
	trackingData := &models.TrackingData{
		ID:        "alt_1234567890_abc123",
		AppID:     "test_app_123",
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
		URL:       "https://example.com",
		Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		CustomParams: map[string]interface{}{
			"campaign_id": "camp_123",
			"source":      "google",
		},
	}

	jsonData, err := trackingData.ToJSON()
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// JSONの構造を検証
	var parsed map[string]interface{}
	err = json.Unmarshal(jsonData, &parsed)
	require.NoError(t, err)
	assert.Equal(t, trackingData.ID, parsed["id"])
	assert.Equal(t, trackingData.AppID, parsed["app_id"])
	assert.Equal(t, trackingData.UserAgent, parsed["user_agent"])
	assert.Equal(t, trackingData.URL, parsed["url"])
	assert.Equal(t, "camp_123", parsed["custom_params"].(map[string]interface{})["campaign_id"])
}

func TestTrackingData_FromJSON(t *testing.T) {
	jsonData := `{
		"id": "alt_1234567890_abc123",
		"app_id": "test_app_123",
		"user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
		"url": "https://example.com",
		"timestamp": "2024-01-15T10:30:00Z",
		"custom_params": {
			"campaign_id": "camp_123",
			"source": "google"
		}
	}`

	trackingData := &models.TrackingData{}
	err := trackingData.FromJSON([]byte(jsonData))

	require.NoError(t, err)
	assert.Equal(t, "alt_1234567890_abc123", trackingData.ID)
	assert.Equal(t, "test_app_123", trackingData.AppID)
	assert.Equal(t, "Mozilla/5.0 (Windows NT 10.0; Win64; x64)", trackingData.UserAgent)
	assert.Equal(t, "https://example.com", trackingData.URL)
	assert.Equal(t, "camp_123", trackingData.CustomParams["campaign_id"])
	assert.Equal(t, "google", trackingData.CustomParams["source"])
}

func TestTrackingData_IsBot(t *testing.T) {
	tests := []struct {
		name      string
		userAgent string
		expected  bool
	}{
		{
			name:      "Googlebot",
			userAgent: "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
			expected:  true,
		},
		{
			name:      "Bingbot",
			userAgent: "Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)",
			expected:  true,
		},
		{
			name:      "YandexBot",
			userAgent: "Mozilla/5.0 (compatible; YandexBot/3.0; +http://yandex.com/bots)",
			expected:  true,
		},
		{
			name:      "Regular browser",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			expected:  false,
		},
		{
			name:      "Mobile browser",
			userAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trackingData := &models.TrackingData{UserAgent: tt.userAgent}
			result := trackingData.IsBot()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrackingData_IsMobile(t *testing.T) {
	tests := []struct {
		name      string
		userAgent string
		expected  bool
	}{
		{
			name:      "iPhone",
			userAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15",
			expected:  true,
		},
		{
			name:      "Android",
			userAgent: "Mozilla/5.0 (Linux; Android 10; SM-G973F) AppleWebKit/537.36",
			expected:  true,
		},
		{
			name:      "iPad",
			userAgent: "Mozilla/5.0 (iPad; CPU OS 14_0 like Mac OS X) AppleWebKit/605.1.15",
			expected:  true,
		},
		{
			name:      "Desktop",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			expected:  false,
		},
		{
			name:      "Bot",
			userAgent: "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trackingData := &models.TrackingData{UserAgent: tt.userAgent}
			result := trackingData.IsMobile()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrackingData_GenerateID(t *testing.T) {
	trackingData := &models.TrackingData{
		AppID:     "test_app_123",
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
		URL:       "https://example.com",
		Timestamp: time.Now(),
	}

	err := trackingData.GenerateID()
	require.NoError(t, err)
	assert.NotEmpty(t, trackingData.ID)
	assert.Len(t, trackingData.ID, 32) // 32文字のID
}

func TestTrackingData_GetDeviceType(t *testing.T) {
	tests := []struct {
		name      string
		userAgent string
		expected  string
	}{
		{
			name:      "Desktop",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			expected:  "desktop",
		},
		{
			name:      "Mobile",
			userAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15",
			expected:  "mobile",
		},
		{
			name:      "Tablet",
			userAgent: "Mozilla/5.0 (iPad; CPU OS 14_0 like Mac OS X) AppleWebKit/605.1.15",
			expected:  "tablet",
		},
		{
			name:      "Bot",
			userAgent: "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
			expected:  "bot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trackingData := &models.TrackingData{UserAgent: tt.userAgent}
			result := trackingData.GetDeviceType()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrackingData_GetBrowser(t *testing.T) {
	tests := []struct {
		name      string
		userAgent string
		expected  string
	}{
		{
			name:      "Chrome",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			expected:  "Chrome",
		},
		{
			name:      "Firefox",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0",
			expected:  "Firefox",
		},
		{
			name:      "Safari",
			userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15",
			expected:  "Safari",
		},
		{
			name:      "Edge",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 Edg/91.0.864.59",
			expected:  "Edge",
		},
		{
			name:      "Unknown",
			userAgent: "Unknown Browser",
			expected:  "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trackingData := &models.TrackingData{UserAgent: tt.userAgent}
			result := trackingData.GetBrowser()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrackingData_GetOS(t *testing.T) {
	tests := []struct {
		name      string
		userAgent string
		expected  string
	}{
		{
			name:      "Windows",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			expected:  "Windows",
		},
		{
			name:      "macOS",
			userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15",
			expected:  "macOS",
		},
		{
			name:      "iOS",
			userAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15",
			expected:  "iOS",
		},
		{
			name:      "Android",
			userAgent: "Mozilla/5.0 (Linux; Android 10; SM-G973F) AppleWebKit/537.36",
			expected:  "Android",
		},
		{
			name:      "Linux",
			userAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
			expected:  "Linux",
		},
		{
			name:      "Unknown",
			userAgent: "Unknown OS",
			expected:  "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trackingData := &models.TrackingData{UserAgent: tt.userAgent}
			result := trackingData.GetOS()
			assert.Equal(t, tt.expected, result)
		})
	}
}
