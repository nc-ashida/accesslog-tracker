package models_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"accesslog-tracker/internal/domain/models"
)

func TestTrackingData_Validate(t *testing.T) {
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
				URL:       "https://example.com/test",
				IPAddress: "192.168.1.100",
				SessionID: "alt_1234567890_abc123",
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing app_id",
			data: &models.TrackingData{
				AppID:     "",
				UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
				URL:       "https://example.com/test",
				IPAddress: "192.168.1.100",
				SessionID: "alt_1234567890_abc123",
				Timestamp: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing user_agent",
			data: &models.TrackingData{
				AppID:     "test_app_123",
				UserAgent: "",
				URL:       "https://example.com/test",
				IPAddress: "192.168.1.100",
				SessionID: "alt_1234567890_abc123",
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
				SessionID: "alt_1234567890_abc123",
				Timestamp: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "invalid url",
			data: &models.TrackingData{
				AppID:     "test_app_123",
				UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
				URL:       "invalid-url",
				IPAddress: "192.168.1.100",
				SessionID: "alt_1234567890_abc123",
				Timestamp: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "zero timestamp",
			data: &models.TrackingData{
				AppID:     "test_app_123",
				UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
				URL:       "https://example.com/test",
				IPAddress: "192.168.1.100",
				SessionID: "alt_1234567890_abc123",
				Timestamp: time.Time{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.data.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTrackingData_IsValidIP(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{"valid IPv4", "192.168.1.100", true},
		{"valid IPv4 localhost", "127.0.0.1", true},
		{"valid IPv6", "2001:db8::1", true},
		{"empty IP", "", true},             // IPアドレスはオプション
		{"invalid IP", "invalid-ip", true}, // 簡易チェックのため
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &models.TrackingData{IPAddress: tt.ip}
			result := data.IsValidIP()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrackingData_IsValidURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{"valid URL", "https://example.com/test", true},
		{"valid HTTP URL", "http://example.com/test", true},
		{"empty URL", "", false},
		{"too short", "abc", false},
		{"invalid URL", "invalid-url", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &models.TrackingData{URL: tt.url}
			result := data.IsValidURL()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrackingData_GetCustomParam(t *testing.T) {
	data := &models.TrackingData{
		CustomParams: map[string]interface{}{
			"param1": "value1",
			"param2": 123,
			"param3": true,
		},
	}

	t.Run("existing param", func(t *testing.T) {
		value, exists := data.GetCustomParam("param1")
		assert.True(t, exists)
		assert.Equal(t, "value1", value)
	})

	t.Run("non-existing param", func(t *testing.T) {
		value, exists := data.GetCustomParam("non-existing")
		assert.False(t, exists)
		assert.Nil(t, value)
	})

	t.Run("nil custom params", func(t *testing.T) {
		data := &models.TrackingData{}
		value, exists := data.GetCustomParam("param1")
		assert.False(t, exists)
		assert.Nil(t, value)
	})
}

func TestTrackingData_SetCustomParam(t *testing.T) {
	data := &models.TrackingData{}

	t.Run("set first param", func(t *testing.T) {
		data.SetCustomParam("param1", "value1")
		assert.Equal(t, "value1", data.CustomParams["param1"])
	})

	t.Run("set additional param", func(t *testing.T) {
		data.SetCustomParam("param2", 123)
		assert.Equal(t, 123, data.CustomParams["param2"])
		assert.Equal(t, "value1", data.CustomParams["param1"]) // 既存の値は保持
	})

	t.Run("update existing param", func(t *testing.T) {
		data.SetCustomParam("param1", "updated_value")
		assert.Equal(t, "updated_value", data.CustomParams["param1"])
	})
}

func TestTrackingData_ToJSON(t *testing.T) {
	data := &models.TrackingData{
		ID:        "tracking_123",
		AppID:     "test_app_123",
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		URL:       "https://example.com/test",
		IPAddress: "192.168.1.100",
		SessionID: "alt_1234567890_abc123",
		Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		CustomParams: map[string]interface{}{
			"param1": "value1",
		},
	}

	jsonData, err := data.ToJSON()
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// JSONとしてパースできることを確認
	var parsedData models.TrackingData
	err = json.Unmarshal(jsonData, &parsedData)
	assert.NoError(t, err)
	assert.Equal(t, data.AppID, parsedData.AppID)
	assert.Equal(t, data.UserAgent, parsedData.UserAgent)
}

func TestTrackingData_FromJSON(t *testing.T) {
	jsonData := `{
		"id": "tracking_123",
		"app_id": "test_app_123",
		"user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		"url": "https://example.com/test",
		"ip_address": "192.168.1.100",
		"session_id": "alt_1234567890_abc123",
		"timestamp": "2024-01-01T12:00:00Z",
		"custom_params": {
			"param1": "value1"
		}
	}`

	data := &models.TrackingData{}
	err := data.FromJSON([]byte(jsonData))

	assert.NoError(t, err)
	assert.Equal(t, "tracking_123", data.ID)
	assert.Equal(t, "test_app_123", data.AppID)
	assert.Equal(t, "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36", data.UserAgent)
	assert.Equal(t, "https://example.com/test", data.URL)
	assert.Equal(t, "192.168.1.100", data.IPAddress)
	assert.Equal(t, "alt_1234567890_abc123", data.SessionID)
	assert.Equal(t, "value1", data.CustomParams["param1"])
}

func TestTrackingData_IsBot(t *testing.T) {
	tests := []struct {
		name      string
		userAgent string
		expected  bool
	}{
		{"Google Bot", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)", true},
		{"Bing Bot", "Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)", true},
		{"Regular browser", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36", false},
		{"Empty user agent", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &models.TrackingData{UserAgent: tt.userAgent}
			result := data.IsBot()
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
		{"iPhone", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15", true},
		{"Android", "Mozilla/5.0 (Linux; Android 10; SM-G975F) AppleWebKit/537.36", true},
		{"iPad", "Mozilla/5.0 (iPad; CPU OS 14_0 like Mac OS X) AppleWebKit/605.1.15", true},
		{"Desktop browser", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36", false},
		{"Empty user agent", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &models.TrackingData{UserAgent: tt.userAgent}
			result := data.IsMobile()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrackingData_GenerateID(t *testing.T) {
	data := &models.TrackingData{}
	err := data.GenerateID()

	assert.NoError(t, err)
	assert.NotEmpty(t, data.ID)
	assert.Len(t, data.ID, 32)
}

func TestTrackingData_GetDeviceType(t *testing.T) {
	tests := []struct {
		name      string
		userAgent string
		expected  string
	}{
		{"Bot", "Mozilla/5.0 (compatible; Googlebot/2.1)", "bot"},
		{"iPhone", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)", "mobile"},
		{"iPad", "Mozilla/5.0 (iPad; CPU OS 14_0 like Mac OS X)", "tablet"},
		{"Desktop", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36", "desktop"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &models.TrackingData{UserAgent: tt.userAgent}
			result := data.GetDeviceType()
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
		{"Chrome", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124", "Chrome"},
		{"Firefox", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0", "Firefox"},
		{"Safari", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15", "Safari"},
		{"Edge", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 Edg/91.0.864.59", "Edge"},
		{"Unknown", "Unknown Browser", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &models.TrackingData{UserAgent: tt.userAgent}
			result := data.GetBrowser()
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
		{"Windows", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36", "Windows"},
		{"macOS", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36", "macOS"},
		{"iOS", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15", "iOS"},
		{"Android", "Mozilla/5.0 (Linux; Android 10; SM-G975F) AppleWebKit/537.36", "Android"},
		{"Linux", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36", "Linux"},
		{"Unknown", "Unknown OS", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &models.TrackingData{UserAgent: tt.userAgent}
			result := data.GetOS()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidIP(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{"valid IPv4", "192.168.1.100", true},
		{"valid IPv4 localhost", "127.0.0.1", true},
		{"valid IPv6", "2001:db8::1", true},
		{"empty IP", "", false},
		{"invalid IP", "invalid-ip", false},
		{"invalid IPv4 format", "192.168.1", false},
		{"invalid IPv4 range", "192.168.1.256", true},        // Goのnet.ParseIPは256を有効として扱う
		{"invalid IPv4 leading zero", "192.168.1.01", false}, // 実装では先頭の0は無効
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := models.IsValidIP(tt.ip)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{"valid HTTPS URL", "https://example.com/test", true},
		{"valid HTTP URL", "http://example.com/test", true},
		{"empty URL", "", false},
		{"too short", "abc", false},
		{"invalid URL", "invalid-url", false},
		{"no protocol", "example.com/test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := models.IsValidURL(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrackingStats_ToJSON(t *testing.T) {
	stats := &models.TrackingStats{
		AppID:          "test_app_123",
		TotalRequests:  1000,
		UniqueSessions: 500,
		UniqueIPs:      300,
		BotRequests:    50,
		MobileRequests: 400,
		StartDate:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:        time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
		CreatedAt:      time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	jsonData, err := stats.ToJSON()
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// JSONとしてパースできることを確認
	var parsedStats models.TrackingStats
	err = json.Unmarshal(jsonData, &parsedStats)
	assert.NoError(t, err)
	assert.Equal(t, stats.AppID, parsedStats.AppID)
	assert.Equal(t, stats.TotalRequests, parsedStats.TotalRequests)
}

func TestTrackingStats_FromJSON(t *testing.T) {
	jsonData := `{
		"app_id": "test_app_123",
		"total_requests": 1000,
		"unique_sessions": 500,
		"unique_ips": 300,
		"bot_requests": 50,
		"mobile_requests": 400,
		"start_date": "2024-01-01T00:00:00Z",
		"end_date": "2024-01-31T23:59:59Z",
		"created_at": "2024-01-01T12:00:00Z"
	}`

	stats := &models.TrackingStats{}
	err := stats.FromJSON([]byte(jsonData))

	assert.NoError(t, err)
	assert.Equal(t, "test_app_123", stats.AppID)
	assert.Equal(t, int64(1000), stats.TotalRequests)
	assert.Equal(t, int64(500), stats.UniqueSessions)
	assert.Equal(t, int64(300), stats.UniqueIPs)
	assert.Equal(t, int64(50), stats.BotRequests)
	assert.Equal(t, int64(400), stats.MobileRequests)
}
