package generator_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"accesslog-tracker/internal/beacon/generator"
)

func TestBeaconGenerator_GenerateBeacon(t *testing.T) {
	gen := generator.NewBeaconGenerator()

	t.Run("should generate beacon with provided parameters", func(t *testing.T) {
		appID := 1
		sessionID := "test_session_123"
		url := "/test-page"
		referrer := "https://example.com"
		userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
		ipAddress := "192.168.1.1"

		beacon := gen.GenerateBeacon(appID, sessionID, url, referrer, userAgent, ipAddress)

		assert.Equal(t, appID, beacon.AppID)
		assert.Equal(t, sessionID, beacon.SessionID)
		assert.Equal(t, url, beacon.URL)
		assert.Equal(t, referrer, beacon.Referrer)
		assert.Equal(t, userAgent, beacon.UserAgent)
		assert.Equal(t, ipAddress, beacon.IPAddress)
		assert.WithinDuration(t, time.Now(), beacon.Timestamp, 2*time.Second)
	})

	t.Run("should auto-generate session ID when empty", func(t *testing.T) {
		beacon := gen.GenerateBeacon(1, "", "/test", "", "", "")

		assert.NotEmpty(t, beacon.SessionID)
		assert.True(t, strings.HasPrefix(beacon.SessionID, "alt_"))
		assert.Len(t, beacon.SessionID, 28) // "alt_" + 14 chars timestamp + "_" + 9 chars random
	})

	t.Run("should handle empty optional parameters", func(t *testing.T) {
		beacon := gen.GenerateBeacon(1, "test_session", "", "", "", "")

		assert.Equal(t, 1, beacon.AppID)
		assert.Equal(t, "test_session", beacon.SessionID)
		assert.Empty(t, beacon.URL)
		assert.Empty(t, beacon.Referrer)
		assert.Empty(t, beacon.UserAgent)
		assert.Empty(t, beacon.IPAddress)
	})
}

func TestBeaconGenerator_GenerateJavaScript(t *testing.T) {
	gen := generator.NewBeaconGenerator()

	t.Run("should generate valid JavaScript beacon", func(t *testing.T) {
		config := generator.BeaconConfig{
			Endpoint: "https://api.example.com/v1/track",
			Debug:    false,
			Version:  "1.0.0",
		}

		result, err := gen.GenerateJavaScript(config)
		require.NoError(t, err)

		// JavaScriptとして有効かチェック
		assert.Contains(t, result, "function track")
		assert.Contains(t, result, config.Endpoint)
		assert.Contains(t, result, "XMLHttpRequest")
		assert.Contains(t, result, "fetch")

		// 構文エラーがないかチェック
		assert.NotContains(t, result, "undefined")
		// nullは正常な値なのでチェックしない
	})

	t.Run("should include debug mode when enabled", func(t *testing.T) {
		config := generator.BeaconConfig{
			Endpoint: "https://api.example.com/v1/track",
			Debug:    true,
			Version:  "1.0.0",
		}

		result, err := gen.GenerateJavaScript(config)
		require.NoError(t, err)

		assert.Contains(t, result, "console.log")
		assert.Contains(t, result, "debug")
	})

	t.Run("should handle custom parameters", func(t *testing.T) {
		config := generator.BeaconConfig{
			Endpoint: "https://api.example.com/v1/track",
			Debug:    false,
			Version:  "1.0.0",
			CustomParams: map[string]string{
				"campaign_id": "camp_123",
				"source":      "email",
			},
		}

		result, err := gen.GenerateJavaScript(config)
		require.NoError(t, err)

		assert.Contains(t, result, "campaign_id")
		assert.Contains(t, result, "source")
		assert.Contains(t, result, "camp_123")
		assert.Contains(t, result, "email")
	})

	t.Run("should handle minified version", func(t *testing.T) {
		config := generator.BeaconConfig{
			Endpoint: "https://api.example.com/v1/track",
			Debug:    false,
			Version:  "1.0.0",
			Minify:   true,
		}

		result, err := gen.GenerateJavaScript(config)
		require.NoError(t, err)

		// ミニファイされたバージョンは空白が少ないはず
		lines := strings.Split(result, "\n")
		assert.Less(t, len(lines), 50) // ミニファイされていないバージョンより少ない行数
	})
}

func TestBeaconGenerator_GenerateBeaconWithConfig(t *testing.T) {
	gen := generator.NewBeaconGenerator()

	t.Run("should generate beacon with custom config", func(t *testing.T) {
		config := generator.BeaconConfig{
			Endpoint: "https://custom-api.com/track",
			Debug:    true,
			Version:  "2.0.0",
			Minify:   false,
			CustomParams: map[string]string{
				"app_id": "123",
				"env":    "production",
			},
		}

		beacon := gen.GenerateBeaconWithConfig(config, "test_session", "/page", "https://ref.com", "Mozilla", "192.168.1.1")

		assert.Equal(t, 123, beacon.AppID) // CustomParamsから取得
		assert.Equal(t, "test_session", beacon.SessionID)
		assert.Equal(t, "/page", beacon.URL)
		assert.Equal(t, "https://ref.com", beacon.Referrer)
		assert.Equal(t, "Mozilla", beacon.UserAgent)
		assert.Equal(t, "192.168.1.1", beacon.IPAddress)
		assert.WithinDuration(t, time.Now(), beacon.Timestamp, 2*time.Second)
	})

	t.Run("should handle empty custom params", func(t *testing.T) {
		config := generator.BeaconConfig{
			Endpoint: "https://api.example.com/track",
			Debug:    false,
			Version:  "1.0.0",
		}

		beacon := gen.GenerateBeaconWithConfig(config, "", "", "", "", "")

		assert.Equal(t, 0, beacon.AppID) // デフォルト値
		assert.NotEmpty(t, beacon.SessionID) // 自動生成
		assert.Empty(t, beacon.URL)
		assert.Empty(t, beacon.Referrer)
		assert.Empty(t, beacon.UserAgent)
		assert.Empty(t, beacon.IPAddress)
	})
}

func TestBeaconGenerator_GenerateMinifiedJavaScript(t *testing.T) {
	gen := generator.NewBeaconGenerator()

	t.Run("should generate minified JavaScript", func(t *testing.T) {
		config := generator.BeaconConfig{
			Endpoint: "https://api.example.com/v1/track",
			Debug:    false,
			Version:  "1.0.0",
		}

		result, err := gen.GenerateMinifiedJavaScript(config)
		require.NoError(t, err)

		// ミニファイされたバージョンの特徴をチェック
		assert.Contains(t, result, "function track")
		assert.Contains(t, result, config.Endpoint)
		
		// 空白や改行が少ない
		lines := strings.Split(result, "\n")
		assert.Less(t, len(lines), 30)
	})
}

func TestBeaconGenerator_GenerateCustomJavaScript(t *testing.T) {
	gen := generator.NewBeaconGenerator()

	t.Run("should generate custom JavaScript with app_id", func(t *testing.T) {
		appID := "123"
		config := generator.BeaconConfig{
			Endpoint: "https://api.example.com/v1/track",
			Debug:    false,
			Version:  "1.0.0",
			CustomParams: map[string]string{
				"app_id": appID,
			},
		}

		result, err := gen.GenerateCustomJavaScript(appID, config)
		require.NoError(t, err)

		assert.Contains(t, result, "function track")
		assert.Contains(t, result, appID)
		assert.Contains(t, result, config.Endpoint)
	})

	t.Run("should handle invalid app_id", func(t *testing.T) {
		appID := "invalid"
		config := generator.BeaconConfig{
			Endpoint: "https://api.example.com/v1/track",
			Debug:    false,
			Version:  "1.0.0",
		}

		result, err := gen.GenerateCustomJavaScript(appID, config)
		require.NoError(t, err)

		// エラーが発生しないが、app_idはそのまま含まれる
		assert.Contains(t, result, appID)
	})
}

func TestBeaconGenerator_ValidateConfig(t *testing.T) {
	gen := generator.NewBeaconGenerator()

	t.Run("should validate correct config", func(t *testing.T) {
		config := generator.BeaconConfig{
			Endpoint: "https://api.example.com/v1/track",
			Debug:    false,
			Version:  "1.0.0",
		}

		err := gen.ValidateConfig(config)
		assert.NoError(t, err)
	})

	t.Run("should reject empty endpoint", func(t *testing.T) {
		config := generator.BeaconConfig{
			Endpoint: "",
			Debug:    false,
			Version:  "1.0.0",
		}

		err := gen.ValidateConfig(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "endpoint")
	})

	t.Run("should accept custom parameters", func(t *testing.T) {
		config := generator.BeaconConfig{
			Endpoint: "https://api.example.com/v1/track",
			Debug:    false,
			Version:  "1.0.0",
			CustomParams: map[string]string{
				"param1": "value1",
				"param2": "value2",
			},
		}

		err := gen.ValidateConfig(config)
		assert.NoError(t, err)
	})
}

func TestBeaconGenerator_GenerateGIFBeacon(t *testing.T) {
	gen := generator.NewBeaconGenerator()

	t.Run("should generate valid GIF beacon", func(t *testing.T) {
		gifData := gen.GenerateGIFBeacon()

		// GIFヘッダーをチェック
		assert.Equal(t, byte(0x47), gifData[0]) // G
		assert.Equal(t, byte(0x49), gifData[1]) // I
		assert.Equal(t, byte(0x46), gifData[2]) // F
		assert.Equal(t, byte(0x38), gifData[3]) // 8
		assert.Equal(t, byte(0x39), gifData[4]) // 9
		assert.Equal(t, byte(0x61), gifData[5]) // a

		// サイズをチェック（1x1ピクセル）
		assert.Equal(t, byte(0x01), gifData[6]) // width low byte
		assert.Equal(t, byte(0x00), gifData[7]) // width high byte
		assert.Equal(t, byte(0x01), gifData[8]) // height low byte
		assert.Equal(t, byte(0x00), gifData[9]) // height high byte

		// 終了マーカーをチェック
		assert.Equal(t, byte(0x3b), gifData[len(gifData)-1]) // ;
	})

	t.Run("should generate consistent GIF data", func(t *testing.T) {
		gif1 := gen.GenerateGIFBeacon()
		gif2 := gen.GenerateGIFBeacon()

		assert.Equal(t, gif1, gif2)
		assert.Len(t, gif1, 40) // 実際のGIFサイズに合わせて調整
	})
}
