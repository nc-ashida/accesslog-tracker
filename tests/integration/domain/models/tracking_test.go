package models_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"accesslog-tracker/internal/domain/models"
)

func TestTrackingModel_Integration(t *testing.T) {
	t.Run("IsValidIP", func(t *testing.T) {
		// 有効なIPアドレス
		validIPs := []string{
			"192.168.1.1",
			"10.0.0.1",
			"172.16.0.1",
			"8.8.8.8",
			"2001:db8::1",
			"::1",
		}

		for _, ip := range validIPs {
			assert.True(t, models.IsValidIP(ip), "IP %s should be valid", ip)
		}

		// 無効なIPアドレス
		invalidIPs := []string{
			"",
			"invalid",
			"192.168.1",
			"192.168.1.1.1",
		}

		for _, ip := range invalidIPs {
			assert.False(t, models.IsValidIP(ip), "IP %s should be invalid", ip)
		}
	})

	t.Run("GetCustomParam", func(t *testing.T) {
		tracking := &models.TrackingData{
			CustomParams: map[string]interface{}{
				"utm_source":   "google",
				"utm_medium":   "cpc",
				"utm_campaign": "test_campaign",
			},
		}

		// 存在するパラメータ
		value, exists := tracking.GetCustomParam("utm_source")
		assert.True(t, exists)
		assert.Equal(t, "google", value)

		// 存在しないパラメータ
		value, exists = tracking.GetCustomParam("nonexistent")
		assert.False(t, exists)
		assert.Nil(t, value)
	})

	t.Run("SetCustomParam", func(t *testing.T) {
		tracking := &models.TrackingData{
			CustomParams: make(map[string]interface{}),
		}

		// 新しいパラメータを設定
		tracking.SetCustomParam("utm_source", "google")
		assert.Equal(t, "google", tracking.CustomParams["utm_source"])

		// 既存のパラメータを更新
		tracking.SetCustomParam("utm_source", "facebook")
		assert.Equal(t, "facebook", tracking.CustomParams["utm_source"])
	})

	t.Run("ToJSON", func(t *testing.T) {
		tracking := &models.TrackingData{
			ID:        "test-id",
			AppID:     "test-app",
			UserAgent: "Mozilla/5.0 (Test Browser)",
			IPAddress: "192.168.1.1",
			URL:       "https://example.com/page",
			Timestamp: time.Date(2023, 12, 25, 10, 30, 0, 0, time.UTC),
		}

		jsonData, err := tracking.ToJSON()
		require.NoError(t, err)
		assert.NotEmpty(t, jsonData)
		assert.Contains(t, string(jsonData), "test-id")
		assert.Contains(t, string(jsonData), "test-app")
	})

	t.Run("FromJSON", func(t *testing.T) {
		jsonData := `{
			"id": "test-id",
			"app_id": "test-app",
			"user_agent": "Mozilla/5.0 (Test Browser)",
			"ip_address": "192.168.1.1",
			"url": "https://example.com/page",
			"timestamp": "2023-12-25T10:30:00Z"
		}`

		tracking := &models.TrackingData{}
		err := tracking.FromJSON([]byte(jsonData))
		require.NoError(t, err)
		assert.Equal(t, "test-id", tracking.ID)
		assert.Equal(t, "test-app", tracking.AppID)
		assert.Equal(t, "Mozilla/5.0 (Test Browser)", tracking.UserAgent)
		assert.Equal(t, "192.168.1.1", tracking.IPAddress)
		assert.Equal(t, "https://example.com/page", tracking.URL)
	})

	t.Run("IsBot", func(t *testing.T) {
		// ボットのUser-Agent
		botUserAgents := []string{
			"Googlebot/2.1 (+http://www.google.com/bot.html)",
			"Bingbot/2.0 (+http://www.bing.com/bingbot.htm)",
			"Mozilla/5.0 (compatible; Baiduspider/2.0; +http://www.baidu.com/search/spider.html)",
			"Mozilla/5.0 (compatible; YandexBot/3.0; +http://yandex.com/bots)",
		}

		for _, ua := range botUserAgents {
			tracking := &models.TrackingData{UserAgent: ua}
			assert.True(t, tracking.IsBot(), "User-Agent %s should be detected as bot", ua)
		}

		// 通常のブラウザのUser-Agent
		normalUserAgents := []string{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
		}

		for _, ua := range normalUserAgents {
			tracking := &models.TrackingData{UserAgent: ua}
			assert.False(t, tracking.IsBot(), "User-Agent %s should not be detected as bot", ua)
		}
	})

	t.Run("IsMobile", func(t *testing.T) {
		// モバイルのUser-Agent
		mobileUserAgents := []string{
			"Mozilla/5.0 (iPhone; CPU iPhone OS 14_7_1 like Mac OS X) AppleWebKit/605.1.15",
			"Mozilla/5.0 (Linux; Android 11; SM-G991B) AppleWebKit/537.36",
			"Mozilla/5.0 (iPad; CPU OS 14_7_1 like Mac OS X) AppleWebKit/605.1.15",
		}

		for _, ua := range mobileUserAgents {
			tracking := &models.TrackingData{UserAgent: ua}
			assert.True(t, tracking.IsMobile(), "User-Agent %s should be detected as mobile", ua)
		}

		// デスクトップのUser-Agent
		desktopUserAgents := []string{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
		}

		for _, ua := range desktopUserAgents {
			tracking := &models.TrackingData{UserAgent: ua}
			assert.False(t, tracking.IsMobile(), "User-Agent %s should not be detected as mobile", ua)
		}
	})

	t.Run("GenerateID", func(t *testing.T) {
		tracking := &models.TrackingData{}
		err := tracking.GenerateID()
		require.NoError(t, err)
		assert.NotEmpty(t, tracking.ID)
		assert.Len(t, tracking.ID, 32) // 32文字のランダムID
	})

	t.Run("GetDeviceType", func(t *testing.T) {
		// デスクトップ
		desktopUA := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
		tracking := &models.TrackingData{UserAgent: desktopUA}
		assert.Equal(t, "desktop", tracking.GetDeviceType())

		// モバイル
		mobileUA := "Mozilla/5.0 (iPhone; CPU iPhone OS 14_7_1 like Mac OS X) AppleWebKit/605.1.15"
		tracking = &models.TrackingData{UserAgent: mobileUA}
		assert.Equal(t, "mobile", tracking.GetDeviceType())

		// タブレット
		tabletUA := "Mozilla/5.0 (iPad; CPU OS 14_7_1 like Mac OS X) AppleWebKit/605.1.15"
		tracking = &models.TrackingData{UserAgent: tabletUA}
		assert.Equal(t, "tablet", tracking.GetDeviceType())
	})

	t.Run("GetBrowser", func(t *testing.T) {
		// Chrome
		chromeUA := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
		tracking := &models.TrackingData{UserAgent: chromeUA}
		assert.Equal(t, "Chrome", tracking.GetBrowser())

		// Firefox
		firefoxUA := "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0"
		tracking = &models.TrackingData{UserAgent: firefoxUA}
		assert.Equal(t, "Firefox", tracking.GetBrowser())

		// Safari
		safariUA := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15"
		tracking = &models.TrackingData{UserAgent: safariUA}
		assert.Equal(t, "Safari", tracking.GetBrowser())

		// Edge
		edgeUA := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 Edg/91.0.864.59"
		tracking = &models.TrackingData{UserAgent: edgeUA}
		assert.Equal(t, "Edge", tracking.GetBrowser())
	})

	t.Run("GetOS", func(t *testing.T) {
		// Windows
		windowsUA := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
		tracking := &models.TrackingData{UserAgent: windowsUA}
		assert.Equal(t, "Windows", tracking.GetOS())

		// macOS
		macosUA := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"
		tracking = &models.TrackingData{UserAgent: macosUA}
		assert.Equal(t, "macOS", tracking.GetOS())

		// Linux
		linuxUA := "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36"
		tracking = &models.TrackingData{UserAgent: linuxUA}
		assert.Equal(t, "Linux", tracking.GetOS())

		// iOS
		iosUA := "Mozilla/5.0 (iPhone; CPU iPhone OS 14_7_1 like Mac OS X) AppleWebKit/605.1.15"
		tracking = &models.TrackingData{UserAgent: iosUA}
		assert.Equal(t, "iOS", tracking.GetOS())

		// Android
		androidUA := "Mozilla/5.0 (Linux; Android 11; SM-G991B) AppleWebKit/537.36"
		tracking = &models.TrackingData{UserAgent: androidUA}
		assert.Equal(t, "Android", tracking.GetOS())
	})

	t.Run("Complex tracking data operations", func(t *testing.T) {
		// 複雑なトラッキングデータのテスト
		tracking := &models.TrackingData{
			AppID:     "test-app",
			UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_7_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Mobile/15E148 Safari/604.1",
			IPAddress: "192.168.1.100",
			URL:       "https://example.com/product/123",
			Referrer:  "https://google.com/search?q=product",
			Timestamp: time.Now(),
			CustomParams: map[string]interface{}{
				"utm_source":   "google",
				"utm_medium":   "cpc",
				"utm_campaign": "summer_sale",
				"page_type":    "product",
				"product_id":   "123",
			},
		}

		// バリデーション
		err := tracking.Validate()
		require.NoError(t, err)

		// デバイスタイプの確認
		assert.Equal(t, "mobile", tracking.GetDeviceType())
		assert.Equal(t, "Safari", tracking.GetBrowser())
		assert.Equal(t, "iOS", tracking.GetOS())
		assert.False(t, tracking.IsBot())
		assert.True(t, tracking.IsMobile())

		// カスタムパラメータの確認
		value, exists := tracking.GetCustomParam("utm_source")
		assert.True(t, exists)
		assert.Equal(t, "google", value)

		// JSON変換
		jsonData, err := tracking.ToJSON()
		require.NoError(t, err)
		assert.NotEmpty(t, jsonData)

		// JSONから復元
		newTracking := &models.TrackingData{}
		err = newTracking.FromJSON(jsonData)
		require.NoError(t, err)
		assert.Equal(t, tracking.ID, newTracking.ID)
		assert.Equal(t, tracking.AppID, newTracking.AppID)
		assert.Equal(t, tracking.UserAgent, newTracking.UserAgent)
	})

	t.Run("Edge cases", func(t *testing.T) {
		// 空のUser-Agent
		tracking := &models.TrackingData{UserAgent: ""}
		assert.Equal(t, "desktop", tracking.GetDeviceType()) // 空のUser-Agentはデスクトップとして扱う
		assert.Equal(t, "Unknown", tracking.GetBrowser())
		assert.Equal(t, "Unknown", tracking.GetOS())
		assert.False(t, tracking.IsBot())
		assert.False(t, tracking.IsMobile())

		// 空のカスタムパラメータ
		tracking = &models.TrackingData{CustomParams: nil}
		value, exists := tracking.GetCustomParam("test")
		assert.False(t, exists)
		assert.Nil(t, value)

		// 無効なJSON
		invalidJSON := []byte(`{"invalid": json}`)
		tracking = &models.TrackingData{}
		err := tracking.FromJSON(invalidJSON)
		assert.Error(t, err)
	})

	t.Run("URL validation", func(t *testing.T) {
		// 有効なURL
		validURLs := []string{
			"https://example.com",
			"http://localhost:8080",
			"https://sub.example.com/path?param=value",
		}

		for _, url := range validURLs {
			assert.True(t, models.IsValidURL(url), "URL %s should be valid", url)
		}

		// 無効なURL
		invalidURLs := []string{
			"",
			"invalid-url",
			"ftp://example.com",
			"not-a-url",
		}

		for _, url := range invalidURLs {
			assert.False(t, models.IsValidURL(url), "URL %s should be invalid", url)
		}
	})
}
