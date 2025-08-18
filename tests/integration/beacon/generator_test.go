package generator_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"accesslog-tracker/internal/beacon/generator"
)

func TestBeaconGenerator_GenerateBeaconWithConfig(t *testing.T) {
	bg := generator.NewBeaconGenerator()

	config := generator.BeaconConfig{
		Version: "1.0.0",
		Debug:   true,
		CustomParams: map[string]string{
			"app_id": "123",
			"custom": "value",
		},
	}

	sessionID := "test_session_123"
	url := "https://example.com/page"
	referrer := "https://google.com"
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
	ipAddress := "192.168.1.1"

	beacon := bg.GenerateBeaconWithConfig(config, sessionID, url, referrer, userAgent, ipAddress)

	// アサーション
	assert.Equal(t, 123, beacon.AppID)
	assert.Equal(t, sessionID, beacon.SessionID)
	assert.Equal(t, url, beacon.URL)
	assert.Equal(t, referrer, beacon.Referrer)
	assert.Equal(t, userAgent, beacon.UserAgent)
	assert.Equal(t, ipAddress, beacon.IPAddress)
	assert.WithinDuration(t, time.Now(), beacon.Timestamp, 2*time.Second)
}

func TestBeaconGenerator_GenerateBeaconWithConfig_EmptySessionID(t *testing.T) {
	bg := generator.NewBeaconGenerator()

	config := generator.BeaconConfig{
		Version: "1.0.0",
		Debug:   false,
		CustomParams: map[string]string{
			"app_id": "456",
		},
	}

	url := "https://example.com/page"
	referrer := "https://google.com"
	userAgent := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"
	ipAddress := "10.0.0.1"

	beacon := bg.GenerateBeaconWithConfig(config, "", url, referrer, userAgent, ipAddress)

	// アサーション
	assert.Equal(t, 456, beacon.AppID)
	assert.NotEmpty(t, beacon.SessionID)
	assert.True(t, len(beacon.SessionID) > 20) // alt_ + timestamp + random string
	assert.Equal(t, url, beacon.URL)
	assert.Equal(t, referrer, beacon.Referrer)
	assert.Equal(t, userAgent, beacon.UserAgent)
	assert.Equal(t, ipAddress, beacon.IPAddress)
	assert.WithinDuration(t, time.Now(), beacon.Timestamp, 2*time.Second)
}

func TestBeaconGenerator_GenerateBeaconWithConfig_InvalidAppID(t *testing.T) {
	bg := generator.NewBeaconGenerator()

	config := generator.BeaconConfig{
		Version: "1.0.0",
		Debug:   false,
		CustomParams: map[string]string{
			"app_id": "invalid",
		},
	}

	sessionID := "test_session_456"
	url := "https://example.com/page"
	referrer := "https://google.com"
	userAgent := "Mozilla/5.0 (Linux; Android 10) AppleWebKit/537.36"
	ipAddress := "172.16.0.1"

	beacon := bg.GenerateBeaconWithConfig(config, sessionID, url, referrer, userAgent, ipAddress)

	// アサーション
	assert.Equal(t, 0, beacon.AppID) // 無効なapp_idの場合は0になる
	assert.Equal(t, sessionID, beacon.SessionID)
	assert.Equal(t, url, beacon.URL)
	assert.Equal(t, referrer, beacon.Referrer)
	assert.Equal(t, userAgent, beacon.UserAgent)
	assert.Equal(t, ipAddress, beacon.IPAddress)
	assert.WithinDuration(t, time.Now(), beacon.Timestamp, 2*time.Second)
}

func TestBeaconGenerator_GenerateMinifiedJavaScript(t *testing.T) {
	bg := generator.NewBeaconGenerator()

	config := generator.BeaconConfig{
		Endpoint: "https://example.com/track",
		Version:  "1.0.0",
		Debug:    true,
		Minify:   true,
		CustomParams: map[string]string{
			"app_id": "789",
			"custom": "value",
		},
	}

	javascript, err := bg.GenerateMinifiedJavaScript(config)
	require.NoError(t, err)

		// アサーション
	assert.NotEmpty(t, javascript)
	// ミニファイされたJavaScriptでは文字列が圧縮されているため、基本的な構造のみ確認
	assert.Contains(t, javascript, "function")
	
	// ミニファイされていることを確認（コメントや不要な空白が削除されている）
	assert.NotContains(t, javascript, "//")
	assert.NotContains(t, javascript, "/*")
	assert.NotContains(t, javascript, "*/")
}

func TestBeaconGenerator_GenerateCustomJavaScript(t *testing.T) {
	bg := generator.NewBeaconGenerator()

	config := generator.BeaconConfig{
		Endpoint: "https://example.com/track",
		Version:  "1.0.0",
		Debug:    false,
		CustomParams: map[string]string{
			"custom": "value",
		},
	}

	appID := "999"
	javascript, err := bg.GenerateCustomJavaScript(appID, config)
	require.NoError(t, err)

	// アサーション
	assert.NotEmpty(t, javascript)
	// 基本的な構造のみ確認
	assert.Contains(t, javascript, "function")
}

func TestBeaconGenerator_GenerateCustomJavaScript_EmptyCustomParams(t *testing.T) {
	bg := generator.NewBeaconGenerator()

	config := generator.BeaconConfig{
		Endpoint: "https://example.com/track",
		Version:  "1.0.0",
		Debug:    false,
		// CustomParams が nil
	}

	appID := "888"
	javascript, err := bg.GenerateCustomJavaScript(appID, config)
	require.NoError(t, err)

	// アサーション
	assert.NotEmpty(t, javascript)
	// 基本的な構造のみ確認
	assert.Contains(t, javascript, "function")
}

func TestBeaconGenerator_GenerateGIFBeacon(t *testing.T) {
	bg := generator.NewBeaconGenerator()

	gifData := bg.GenerateGIFBeacon()

	// アサーション
	assert.NotEmpty(t, gifData)
	assert.Equal(t, 40, len(gifData)) // 1x1ピクセルの透明GIFの標準サイズ

	// GIFヘッダーの確認
	assert.Equal(t, byte(0x47), gifData[0]) // G
	assert.Equal(t, byte(0x49), gifData[1]) // I
	assert.Equal(t, byte(0x46), gifData[2]) // F
	assert.Equal(t, byte(0x38), gifData[3]) // 8
	assert.Equal(t, byte(0x39), gifData[4]) // 9
	assert.Equal(t, byte(0x61), gifData[5]) // a

	// GIF終了マーカーの確認
	assert.Equal(t, byte(0x3b), gifData[len(gifData)-1]) // ;
}

func TestBeaconGenerator_GenerateBeaconWithConfig_NoAppID(t *testing.T) {
	bg := generator.NewBeaconGenerator()

	config := generator.BeaconConfig{
		Version: "1.0.0",
		Debug:   false,
		CustomParams: map[string]string{
			"other": "value",
		},
	}

	sessionID := "test_session_no_app_id"
	url := "https://example.com/page"
	referrer := "https://google.com"
	userAgent := "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15"
	ipAddress := "203.0.113.1"

	beacon := bg.GenerateBeaconWithConfig(config, sessionID, url, referrer, userAgent, ipAddress)

	// アサーション
	assert.Equal(t, 0, beacon.AppID) // app_idが設定されていない場合は0
	assert.Equal(t, sessionID, beacon.SessionID)
	assert.Equal(t, url, beacon.URL)
	assert.Equal(t, referrer, beacon.Referrer)
	assert.Equal(t, userAgent, beacon.UserAgent)
	assert.Equal(t, ipAddress, beacon.IPAddress)
	assert.WithinDuration(t, time.Now(), beacon.Timestamp, 2*time.Second)
}
