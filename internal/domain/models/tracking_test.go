package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewTrackingData(t *testing.T) {
	req := &TrackingRequest{
		ApplicationID: "test-app-123",
		SessionID:     "test-session-456",
		UserID:        "test-user-789",
		PageURL:       "https://example.com/page",
		Referrer:      "https://google.com",
		UserAgent:     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		IPAddress:     "192.168.1.1",
		ScreenWidth:   1920,
		ScreenHeight:  1080,
		ViewportWidth: 1200,
		ViewportHeight: 800,
		Language:      "ja-JP",
		Timezone:      "Asia/Tokyo",
		CustomParams: map[string]interface{}{
			"campaign": "summer2024",
			"source":   "google",
		},
	}

	trackingData := NewTrackingData(req)

	if trackingData.ApplicationID != req.ApplicationID {
		t.Errorf("Expected ApplicationID %s, got %s", req.ApplicationID, trackingData.ApplicationID)
	}

	if trackingData.SessionID != req.SessionID {
		t.Errorf("Expected SessionID %s, got %s", req.SessionID, trackingData.SessionID)
	}

	if trackingData.UserID != req.UserID {
		t.Errorf("Expected UserID %s, got %s", req.UserID, trackingData.UserID)
	}

	if trackingData.PageURL != req.PageURL {
		t.Errorf("Expected PageURL %s, got %s", req.PageURL, trackingData.PageURL)
	}

	if trackingData.Referrer != req.Referrer {
		t.Errorf("Expected Referrer %s, got %s", req.Referrer, trackingData.Referrer)
	}

	if trackingData.UserAgent != req.UserAgent {
		t.Errorf("Expected UserAgent %s, got %s", req.UserAgent, trackingData.UserAgent)
	}

	if trackingData.IPAddress != req.IPAddress {
		t.Errorf("Expected IPAddress %s, got %s", req.IPAddress, trackingData.IPAddress)
	}

	if trackingData.ScreenWidth != req.ScreenWidth {
		t.Errorf("Expected ScreenWidth %d, got %d", req.ScreenWidth, trackingData.ScreenWidth)
	}

	if trackingData.ScreenHeight != req.ScreenHeight {
		t.Errorf("Expected ScreenHeight %d, got %d", req.ScreenHeight, trackingData.ScreenHeight)
	}

	if trackingData.ViewportWidth != req.ViewportWidth {
		t.Errorf("Expected ViewportWidth %d, got %d", req.ViewportWidth, trackingData.ViewportWidth)
	}

	if trackingData.ViewportHeight != req.ViewportHeight {
		t.Errorf("Expected ViewportHeight %d, got %d", req.ViewportHeight, trackingData.ViewportHeight)
	}

	if trackingData.Language != req.Language {
		t.Errorf("Expected Language %s, got %s", req.Language, trackingData.Language)
	}

	if trackingData.Timezone != req.Timezone {
		t.Errorf("Expected Timezone %s, got %s", req.Timezone, trackingData.Timezone)
	}

	if len(trackingData.CustomParams) != len(req.CustomParams) {
		t.Errorf("Expected CustomParams length %d, got %d", len(req.CustomParams), len(trackingData.CustomParams))
	}

	if trackingData.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	if trackingData.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}

func TestTrackingData_ToJSON(t *testing.T) {
	trackingData := &TrackingData{
		ID:            "test-id",
		ApplicationID: "test-app",
		SessionID:     "test-session",
		PageURL:       "https://example.com",
		UserAgent:     "Mozilla/5.0",
		IPAddress:     "192.168.1.1",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	jsonData, err := trackingData.ToJSON()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("Expected JSON data to not be empty")
	}

	// デコードして検証
	var decoded TrackingData
	err = json.Unmarshal(jsonData, &decoded)
	if err != nil {
		t.Errorf("Expected no error when unmarshaling, got %v", err)
	}

	if decoded.ID != trackingData.ID {
		t.Errorf("Expected ID %s, got %s", trackingData.ID, decoded.ID)
	}
}

func TestTrackingData_FromJSON(t *testing.T) {
	original := &TrackingData{
		ID:            "test-id",
		ApplicationID: "test-app",
		SessionID:     "test-session",
		PageURL:       "https://example.com",
		UserAgent:     "Mozilla/5.0",
		IPAddress:     "192.168.1.1",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	jsonData, err := original.ToJSON()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	decoded := &TrackingData{}
	err = decoded.FromJSON(jsonData)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("Expected ID %s, got %s", original.ID, decoded.ID)
	}

	if decoded.ApplicationID != original.ApplicationID {
		t.Errorf("Expected ApplicationID %s, got %s", original.ApplicationID, decoded.ApplicationID)
	}
}

func TestTrackingData_GetCustomParam(t *testing.T) {
	trackingData := &TrackingData{
		CustomParams: map[string]interface{}{
			"campaign": "summer2024",
			"source":   "google",
			"count":    42,
		},
	}

	// 存在するパラメータの取得
	if campaign := trackingData.GetCustomParam("campaign"); campaign != "summer2024" {
		t.Errorf("Expected campaign 'summer2024', got %v", campaign)
	}

	if source := trackingData.GetCustomParam("source"); source != "google" {
		t.Errorf("Expected source 'google', got %v", source)
	}

	if count := trackingData.GetCustomParam("count"); count != 42 {
		t.Errorf("Expected count 42, got %v", count)
	}

	// 存在しないパラメータの取得
	if notFound := trackingData.GetCustomParam("notfound"); notFound != nil {
		t.Errorf("Expected nil for non-existent param, got %v", notFound)
	}

	// CustomParamsがnilの場合
	trackingData.CustomParams = nil
	if result := trackingData.GetCustomParam("any"); result != nil {
		t.Errorf("Expected nil when CustomParams is nil, got %v", result)
	}
}

func TestTrackingData_SetCustomParam(t *testing.T) {
	trackingData := &TrackingData{}

	// 新しいパラメータの設定
	trackingData.SetCustomParam("campaign", "summer2024")
	if campaign := trackingData.GetCustomParam("campaign"); campaign != "summer2024" {
		t.Errorf("Expected campaign 'summer2024', got %v", campaign)
	}

	// 既存パラメータの更新
	trackingData.SetCustomParam("campaign", "winter2024")
	if campaign := trackingData.GetCustomParam("campaign"); campaign != "winter2024" {
		t.Errorf("Expected campaign 'winter2024', got %v", campaign)
	}

	// 複数のパラメータの設定
	trackingData.SetCustomParam("source", "google")
	trackingData.SetCustomParam("count", 42)

	if len(trackingData.CustomParams) != 3 {
		t.Errorf("Expected 3 custom params, got %d", len(trackingData.CustomParams))
	}
}

func TestTrackingData_IsBot(t *testing.T) {
	tests := []struct {
		name      string
		userAgent string
		expected  bool
	}{
		{
			name:      "Google Bot",
			userAgent: "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
			expected:  true,
		},
		{
			name:      "Bing Bot",
			userAgent: "Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)",
			expected:  true,
		},
		{
			name:      "Regular Browser",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			expected:  false,
		},
		{
			name:      "Crawler",
			userAgent: "Mozilla/5.0 (compatible; crawler/1.0)",
			expected:  true,
		},
		{
			name:      "Spider",
			userAgent: "Mozilla/5.0 (compatible; spider/1.0)",
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trackingData := &TrackingData{UserAgent: tt.userAgent}
			if result := trackingData.IsBot(); result != tt.expected {
				t.Errorf("Expected IsBot() to be %v, got %v", tt.expected, result)
			}
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
			userAgent: "Mozilla/5.0 (Linux; Android 10; SM-G975F) AppleWebKit/537.36",
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
			name:      "BlackBerry",
			userAgent: "Mozilla/5.0 (BlackBerry; U; BlackBerry 9900; en) AppleWebKit/534.11",
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trackingData := &TrackingData{UserAgent: tt.userAgent}
			if result := trackingData.IsMobile(); result != tt.expected {
				t.Errorf("Expected IsMobile() to be %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestTrackingData_GetDeviceType(t *testing.T) {
	tests := []struct {
		name       string
		deviceType string
		userAgent  string
		expected   string
	}{
		{
			name:       "Predefined Mobile",
			deviceType: "mobile",
			userAgent:  "Mozilla/5.0",
			expected:   "mobile",
		},
		{
			name:       "Predefined Tablet",
			deviceType: "tablet",
			userAgent:  "Mozilla/5.0",
			expected:   "tablet",
		},
		{
			name:       "Detected Mobile",
			deviceType: "",
			userAgent:  "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)",
			expected:   "mobile",
		},
		{
			name:       "Detected Tablet",
			deviceType: "",
			userAgent:  "Mozilla/5.0 (iPad; CPU OS 14_0 like Mac OS X)",
			expected:   "tablet",
		},
		{
			name:       "Desktop Default",
			deviceType: "",
			userAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
			expected:   "desktop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trackingData := &TrackingData{
				DeviceType: tt.deviceType,
				UserAgent:  tt.userAgent,
			}
			if result := trackingData.GetDeviceType(); result != tt.expected {
				t.Errorf("Expected GetDeviceType() to be %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestTrackingData_GetBrowser(t *testing.T) {
	tests := []struct {
		name     string
		browser  string
		userAgent string
		expected string
	}{
		{
			name:     "Predefined Chrome",
			browser:  "chrome",
			userAgent: "Mozilla/5.0",
			expected: "chrome",
		},
		{
			name:     "Detected Chrome",
			browser:  "",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124",
			expected: "chrome",
		},
		{
			name:     "Detected Firefox",
			browser:  "",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0",
			expected: "firefox",
		},
		{
			name:     "Detected Safari",
			browser:  "",
			userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15",
			expected: "safari",
		},
		{
			name:     "Detected Edge",
			browser:  "",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 Edg/91.0.864.59",
			expected: "edge",
		},
		{
			name:     "Detected Opera",
			browser:  "",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 OPR/77.0.4054.254",
			expected: "opera",
		},
		{
			name:     "Unknown Browser",
			browser:  "",
			userAgent: "Some Unknown Browser",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trackingData := &TrackingData{
				Browser:   tt.browser,
				UserAgent: tt.userAgent,
			}
			if result := trackingData.GetBrowser(); result != tt.expected {
				t.Errorf("Expected GetBrowser() to be %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestTrackingData_GetOS(t *testing.T) {
	tests := []struct {
		name     string
		os       string
		userAgent string
		expected string
	}{
		{
			name:     "Predefined Windows",
			os:       "windows",
			userAgent: "Mozilla/5.0",
			expected: "windows",
		},
		{
			name:     "Detected Windows",
			os:       "",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			expected: "windows",
		},
		{
			name:     "Detected macOS",
			os:       "",
			userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
			expected: "macos",
		},
		{
			name:     "Detected Linux",
			os:       "",
			userAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
			expected: "linux",
		},
		{
			name:     "Detected Android",
			os:       "",
			userAgent: "Mozilla/5.0 (Linux; Android 10; SM-G975F) AppleWebKit/537.36",
			expected: "android",
		},
		{
			name:     "Detected iOS",
			os:       "",
			userAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15",
			expected: "ios",
		},
		{
			name:     "Unknown OS",
			os:       "",
			userAgent: "Some Unknown OS",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trackingData := &TrackingData{
				OS:        tt.os,
				UserAgent: tt.userAgent,
			}
			if result := trackingData.GetOS(); result != tt.expected {
				t.Errorf("Expected GetOS() to be %s, got %s", tt.expected, result)
			}
		})
	}
}
