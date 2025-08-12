package generator

import (
	"bytes"
	"html/template"
	"strings"
)

// BeaconConfig はビーコンの設定
type BeaconConfig struct {
	AppID         string
	ClientSubID   string
	ModuleID      string
	Endpoint      string
	Debug         bool
	RespectDNT    bool
	SessionTimeout int
}

// BeaconGenerator はビーコン生成器
type BeaconGenerator struct {
	templates map[string]*template.Template
}

// NewBeaconGenerator は新しいビーコン生成器を作成
func NewBeaconGenerator() *BeaconGenerator {
	return &BeaconGenerator{
		templates: make(map[string]*template.Template),
	}
}

// GenerateBeacon はビーコンを生成
func (g *BeaconGenerator) GenerateBeacon(config BeaconConfig) (string, error) {
	tmpl, err := template.New("tracker").Parse(beaconTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, config)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// GenerateEmbedCode は埋め込みコードを生成
func (g *BeaconGenerator) GenerateEmbedCode(config BeaconConfig) (string, error) {
	tmpl, err := template.New("embed").Parse(embedTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, config)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// MinifyBeacon はビーコンを圧縮
func (g *BeaconGenerator) MinifyBeacon(code string) string {
	// 基本的な圧縮処理
	code = strings.ReplaceAll(code, "\n", "")
	code = strings.ReplaceAll(code, "\t", "")
	code = strings.ReplaceAll(code, "  ", " ")
	return strings.TrimSpace(code)
}

// beaconTemplate はビーコンのテンプレート
const beaconTemplate = `
(function() {
    'use strict';
    
    // 設定
    var config = {
        appId: '{{.AppID}}',
        clientSubId: '{{.ClientSubID}}',
        moduleId: '{{.ModuleID}}',
        endpoint: '{{.Endpoint}}',
        debug: {{.Debug}},
        respectDNT: {{.RespectDNT}},
        sessionTimeout: {{.SessionTimeout}}
    };
    
    // DNTチェック
    if (config.respectDNT && navigator.doNotTrack === '1') {
        return;
    }
    
    // セッション管理
    var sessionId = localStorage.getItem('alt_session_id');
    if (!sessionId) {
        sessionId = generateUUID();
        localStorage.setItem('alt_session_id', sessionId);
    }
    
    // 画面情報取得
    var screenRes = screen.width + 'x' + screen.height;
    
    // トラッキングデータ送信
    function sendTrackingData() {
        var data = {
            app_id: config.appId,
            client_sub_id: config.clientSubId,
            module_id: config.moduleId,
            url: window.location.href,
            referrer: document.referrer,
            user_agent: navigator.userAgent,
            session_id: sessionId,
            screen_resolution: screenRes,
            language: navigator.language,
            timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
            custom_params: {}
        };
        
        // カスタムパラメータの収集
        if (window.ALT_CUSTOM_PARAMS) {
            data.custom_params = window.ALT_CUSTOM_PARAMS;
        }
        
        // 送信
        fetch(config.endpoint, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        }).catch(function(error) {
            if (config.debug) {
                console.error('ALT Tracking Error:', error);
            }
        });
    }
    
    // UUID生成
    function generateUUID() {
        return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
            var r = Math.random() * 16 | 0;
            var v = c == 'x' ? r : (r & 0x3 | 0x8);
            return v.toString(16);
        });
    }
    
    // ページ読み込み完了時に送信
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', sendTrackingData);
    } else {
        sendTrackingData();
    }
    
    // ページ離脱時の送信
    window.addEventListener('beforeunload', function() {
        navigator.sendBeacon(config.endpoint, JSON.stringify({
            app_id: config.appId,
            client_sub_id: config.clientSubId,
            module_id: config.moduleId,
            url: window.location.href,
            user_agent: navigator.userAgent,
            session_id: sessionId,
            screen_resolution: screenRes,
            custom_params: { event_type: 'page_exit' }
        }));
    });
    
    // グローバル関数として公開
    window.ALT = {
        track: sendTrackingData,
        setCustomParams: function(params) {
            window.ALT_CUSTOM_PARAMS = params;
        }
    };
})();
`

// embedTemplate は埋め込みコードのテンプレート
const embedTemplate = `
<script>
(function() {
    var script = document.createElement('script');
    script.async = true;
    script.src = 'https://d1234567890.cloudfront.net/tracker.js';
    script.setAttribute('data-app-id', '{{.AppID}}');
    {{if .ClientSubID}}script.setAttribute('data-client-sub-id', '{{.ClientSubID}}');{{end}}
    {{if .ModuleID}}script.setAttribute('data-module-id', '{{.ModuleID}}');{{end}}
    var firstScript = document.getElementsByTagName('script')[0];
    firstScript.parentNode.insertBefore(script, firstScript);
})();
</script>
`
