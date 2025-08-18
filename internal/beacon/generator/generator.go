package generator

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"text/template"
	"time"
	"accesslog-tracker/internal/utils/crypto"
)

// Beacon はトラッキングビーコンの構造体です
type Beacon struct {
	AppID     int       `json:"app_id"`
	SessionID string    `json:"session_id"`
	URL       string    `json:"url"`
	Referrer  string    `json:"referrer"`
	UserAgent string    `json:"user_agent"`
	IPAddress string    `json:"ip_address"`
	Timestamp time.Time `json:"timestamp"`
}

// BeaconConfig はビーコン生成の設定構造体です
type BeaconConfig struct {
	Endpoint     string            `json:"endpoint"`
	Debug        bool              `json:"debug"`
	Version      string            `json:"version"`
	Minify       bool              `json:"minify"`
	CustomParams map[string]string `json:"custom_params,omitempty"`
}

// BeaconGenerator はビーコン生成器の構造体です
type BeaconGenerator struct{}

// NewBeaconGenerator は新しいビーコン生成器を作成します
func NewBeaconGenerator() *BeaconGenerator {
	return &BeaconGenerator{}
}

// GenerateBeacon は新しいビーコンを生成します
func (bg *BeaconGenerator) GenerateBeacon(appID int, sessionID, url, referrer, userAgent, ipAddress string) *Beacon {
	if sessionID == "" {
		sessionID = "alt_" + time.Now().Format("20060102150405") + "_" + crypto.GenerateRandomString(9)
	}

	return &Beacon{
		AppID:     appID,
		SessionID: sessionID,
		URL:       url,
		Referrer:  referrer,
		UserAgent: userAgent,
		IPAddress: ipAddress,
		Timestamp: time.Now(),
	}
}

// ValidateConfig はビーコン設定を検証します
func (bg *BeaconGenerator) ValidateConfig(config BeaconConfig) error {
	if config.Endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}

	if _, err := url.Parse(config.Endpoint); err != nil {
		return fmt.Errorf("invalid endpoint URL: %v", err)
	}

	if config.Version == "" {
		return fmt.Errorf("version is required")
	}

	return nil
}

// GenerateJavaScript はJavaScriptビーコンを生成します
func (bg *BeaconGenerator) GenerateJavaScript(config BeaconConfig) (string, error) {
	if err := bg.ValidateConfig(config); err != nil {
		return "", err
	}

	// JavaScriptテンプレート
	jsTemplate := `
(function() {
    'use strict';
    
    // 設定
    var config = {
        endpoint: '{{.Endpoint}}',
        version: '{{.Version}}',
        debug: {{.Debug}},
        customParams: {{.CustomParamsJSON}}
    };
    
    // デバッグログ
    function log(message) {
        if (config.debug) {
            console.log('[ALT Tracker]', message);
        }
    }
    
    // データ収集
    function collectData() {
        var data = {
            app_id: window.ALT_CONFIG ? window.ALT_CONFIG.app_id : null,
            client_sub_id: window.ALT_CONFIG ? window.ALT_CONFIG.client_sub_id : null,
            module_id: window.ALT_CONFIG ? window.ALT_CONFIG.module_id : null,
            url: window.location.href,
            referrer: document.referrer,
            user_agent: navigator.userAgent,
            screen_res: screen.width + 'x' + screen.height,
            language: navigator.language,
            timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
            timestamp: new Date().toISOString()
        };
        
        // カスタムパラメータを追加
        if (config.customParams) {
            for (var key in config.customParams) {
                data[key] = config.customParams[key];
            }
        }
        
        return data;
    }
    
    // データ送信
    function sendData(data) {
        log('Sending tracking data: ' + JSON.stringify(data));
        
        // fetch APIを使用
        if (window.fetch) {
            fetch(config.endpoint, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-API-Key': window.ALT_CONFIG ? window.ALT_CONFIG.api_key : null
                },
                body: JSON.stringify(data)
            })
            .then(function(response) {
                if (response.ok) {
                    log('Data sent successfully');
                } else {
                    log('Failed to send data: ' + response.status);
                }
            })
            .catch(function(error) {
                log('Error sending data: ' + error.message);
            });
        } else {
            // フォールバック: XMLHttpRequest
            var xhr = new XMLHttpRequest();
            xhr.open('POST', config.endpoint, true);
            xhr.setRequestHeader('Content-Type', 'application/json');
            if (window.ALT_CONFIG && window.ALT_CONFIG.api_key) {
                xhr.setRequestHeader('X-API-Key', window.ALT_CONFIG.api_key);
            }
            xhr.onreadystatechange = function() {
                if (xhr.readyState === 4) {
                    if (xhr.status === 200) {
                        log('Data sent successfully');
                    } else {
                        log('Failed to send data: ' + xhr.status);
                    }
                }
            };
            xhr.send(JSON.stringify(data));
        }
    }
    
    // メイン関数
    function track() {
        try {
            var data = collectData();
            sendData(data);
        } catch (error) {
            log('Error in track function: ' + error.message);
        }
    }
    
    // ページ読み込み時に実行
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', track);
    } else {
        track();
    }
    
    // グローバル関数として公開
    window.ALT_Track = track;
    
    log('ALT Tracker v' + config.version + ' loaded');
})();`

	// カスタムパラメータをJSONに変換
	customParamsJSON := "{}"
	if len(config.CustomParams) > 0 {
		params := make([]string, 0, len(config.CustomParams))
		for key, value := range config.CustomParams {
			params = append(params, fmt.Sprintf(`"%s": "%s"`, key, value))
		}
		customParamsJSON = "{" + strings.Join(params, ", ") + "}"
	}

	// テンプレートデータ
	templateData := struct {
		Endpoint         string
		Version          string
		Debug            bool
		CustomParamsJSON string
	}{
		Endpoint:         config.Endpoint,
		Version:          config.Version,
		Debug:            config.Debug,
		CustomParamsJSON: customParamsJSON,
	}

	// テンプレートを実行
	tmpl, err := template.New("beacon").Parse(jsTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %v", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", fmt.Errorf("failed to execute template: %v", err)
	}

	result := buf.String()

	// ミニファイ処理
	if config.Minify {
		result = bg.minify(result)
	}

	return result, nil
}

// minify はJavaScriptコードを圧縮します
func (bg *BeaconGenerator) minify(code string) string {
	// コメントを削除
	code = regexp.MustCompile(`//.*$`).ReplaceAllString(code, "")
	code = regexp.MustCompile(`/\*.*?\*/`).ReplaceAllString(code, "")
	
	// 不要な空白と改行を削除
	code = regexp.MustCompile(`\s+`).ReplaceAllString(code, " ")
	code = strings.TrimSpace(code)
	
	return code
}

// GenerateBeaconWithConfig はカスタム設定でビーコンを生成します
func (bg *BeaconGenerator) GenerateBeaconWithConfig(config BeaconConfig, sessionID, url, referrer, userAgent, ipAddress string) *Beacon {
	appID := 0
	if appIDStr, exists := config.CustomParams["app_id"]; exists {
		if id, err := fmt.Sscanf(appIDStr, "%d", &appID); err != nil || id != 1 {
			appID = 0
		}
	}

	if sessionID == "" {
		sessionID = "alt_" + time.Now().Format("20060102150405") + "_" + crypto.GenerateRandomString(9)
	}

	return &Beacon{
		AppID:     appID,
		SessionID: sessionID,
		URL:       url,
		Referrer:  referrer,
		UserAgent: userAgent,
		IPAddress: ipAddress,
		Timestamp: time.Now(),
	}
}

// GenerateMinifiedJavaScript はミニファイされたJavaScriptを生成します
func (bg *BeaconGenerator) GenerateMinifiedJavaScript(config BeaconConfig) (string, error) {
	javascript, err := bg.GenerateJavaScript(config)
	if err != nil {
		return "", err
	}

	return bg.minify(javascript), nil
}

// GenerateCustomJavaScript はカスタムapp_idでJavaScriptを生成します
func (bg *BeaconGenerator) GenerateCustomJavaScript(appID string, config BeaconConfig) (string, error) {
	// app_idをカスタムパラメータに追加
	if config.CustomParams == nil {
		config.CustomParams = make(map[string]string)
	}
	config.CustomParams["app_id"] = appID

	return bg.GenerateJavaScript(config)
}

// GenerateGIFBeacon は1x1ピクセルの透明GIFビーコンを生成します
func (bg *BeaconGenerator) GenerateGIFBeacon() []byte {
	// 1x1ピクセルの透明GIF画像（Base64エンコード）
	// これは最小限のGIF画像データです
	return []byte{
		0x47, 0x49, 0x46, 0x38, 0x39, 0x61, // GIF89a
		0x01, 0x00, 0x01, 0x00, // 1x1ピクセル
		0x80, 0x00, 0x00, // 背景色（透明）
		0x00, 0x00, 0x00, // パレット
		0x21, 0xf9, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00, // グラフィック制御拡張
		0x2c, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, // 画像記述子
		0x02, 0x02, 0x44, 0x01, 0x00, // 画像データ
		0x3b, // 終了
	}
}
