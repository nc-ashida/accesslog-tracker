package handlers

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"accesslog-tracker/internal/api/models"
	"accesslog-tracker/internal/beacon/generator"
)

// BeaconHandler はビーコン配信ハンドラーです
type BeaconHandler struct {
	generator *generator.BeaconGenerator
}

// NewBeaconHandler は新しいビーコンハンドラーを作成します
func NewBeaconHandler() *BeaconHandler {
	return &BeaconHandler{
		generator: generator.NewBeaconGenerator(),
	}
}

// Serve はJavaScriptビーコンを配信します
func (h *BeaconHandler) Serve(c *gin.Context) {
	config := generator.BeaconConfig{
		Endpoint: "https://api.access-log-tracker.com/v1/track",
		Debug:    false,
		Version:  "1.0.0",
		Minify:   false,
	}

	javascript, err := h.generator.GenerateJavaScript(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "BEACON_GENERATION_ERROR",
				Message: "Failed to generate beacon",
				Details: err.Error(),
			},
			Timestamp: time.Now(),
		})
		return
	}

	// ETagを生成
	etag := fmt.Sprintf("\"%x\"", md5.Sum([]byte(javascript)))
	c.Header("ETag", etag)
	c.Header("Cache-Control", "public, max-age=3600")
	c.Header("Content-Type", "application/javascript")

	// 条件付きリクエストをチェック
	if match := c.GetHeader("If-None-Match"); match == etag {
		c.Status(http.StatusNotModified)
		return
	}

	c.Data(http.StatusOK, "application/javascript", []byte(javascript))
}

// ServeMinified は圧縮版JavaScriptビーコンを配信します
func (h *BeaconHandler) ServeMinified(c *gin.Context) {
	config := generator.BeaconConfig{
		Endpoint: "https://api.access-log-tracker.com/v1/track",
		Debug:    false,
		Version:  "1.0.0",
		Minify:   true,
	}

	javascript, err := h.generator.GenerateJavaScript(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "BEACON_GENERATION_ERROR",
				Message: "Failed to generate minified beacon",
				Details: err.Error(),
			},
			Timestamp: time.Now(),
		})
		return
	}

	// ETagを生成
	etag := fmt.Sprintf("\"%x\"", md5.Sum([]byte(javascript)))
	c.Header("ETag", etag)
	c.Header("Cache-Control", "public, max-age=86400") // 24時間キャッシュ
	c.Header("Content-Type", "application/javascript")

	// 条件付きリクエストをチェック
	if match := c.GetHeader("If-None-Match"); match == etag {
		c.Status(http.StatusNotModified)
		return
	}

	c.Data(http.StatusOK, "application/javascript", []byte(javascript))
}

// ServeCustom はカスタム設定のビーコンを配信します
func (h *BeaconHandler) ServeCustom(c *gin.Context) {
	appIDStr := c.Param("app_id")
	// デバッグ用ログ
	fmt.Printf("ServeCustom: app_id param = '%s'\n", appIDStr)

	if appIDStr == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_APP_ID",
				Message: "Invalid application ID",
				Details: "app_id parameter is required",
			},
			Timestamp: time.Now(),
		})
		return
	}

	// app_idから.js拡張子を除去（UUID等も許容）
	appIDStr = strings.TrimSuffix(appIDStr, ".js")

	config := generator.BeaconConfig{
		Endpoint: "https://api.access-log-tracker.com/v1/track",
		Debug:    false,
		Version:  "1.0.0",
		Minify:   false,
		CustomParams: map[string]string{
			"app_id": appIDStr,
		},
	}

	javascript, err := h.generator.GenerateJavaScript(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "BEACON_GENERATION_ERROR",
				Message: "Failed to generate custom beacon",
				Details: err.Error(),
			},
			Timestamp: time.Now(),
		})
		return
	}

	// ETagを生成
	etag := fmt.Sprintf("\"%x\"", md5.Sum([]byte(javascript)))
	c.Header("ETag", etag)
	c.Header("Cache-Control", "public, max-age=3600")
	c.Header("Content-Type", "application/javascript")

	// 条件付きリクエストをチェック
	if match := c.GetHeader("If-None-Match"); match == etag {
		c.Status(http.StatusNotModified)
		return
	}

	c.Data(http.StatusOK, "application/javascript", []byte(javascript))
}

// GenerateBeacon はデフォルト設定でビーコンを生成します
func (h *BeaconHandler) GenerateBeacon(c *gin.Context) {
	// クエリパラメータを取得
	appID := c.Query("app_id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "app_id parameter is required",
			},
			Timestamp: time.Now(),
		})
		return
	}

	// トラッキングデータを収集
	sessionID := c.Query("session_id")
	url := c.Query("url")
	if url == "" {
		url = "/"
	}
	referrer := c.Query("referrer")
	userAgent := c.GetHeader("User-Agent")

	// IPアドレスを取得
	ipAddress := c.ClientIP()
	if ipAddress == "" {
		ipAddress = c.Request.RemoteAddr
	}

	// カスタムパラメータを収集
	customParams := make(map[string]interface{})
	for key, values := range c.Request.URL.Query() {
		if key != "app_id" && key != "session_id" && key != "url" && key != "referrer" {
			if len(values) > 0 {
				customParams[key] = values[0]
			}
		}
	}

	// 1x1ピクセルの透明GIF画像（Base64エンコード）
	// これは最小限のGIF画像データです
	gifData := []byte{
		0x47, 0x49, 0x46, 0x38, 0x39, 0x61, // GIF89a
		0x01, 0x00, 0x01, 0x00, // 1x1ピクセル
		0x80, 0x00, 0x00, // 背景色（透明）
		0x00, 0x00, 0x00, // パレット
		0x21, 0xf9, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00, // グラフィック制御拡張
		0x2c, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, // 画像記述子
		0x02, 0x02, 0x44, 0x01, 0x00, // 画像データ
		0x3b, // 終了
	}

	// レスポンスヘッダーを設定
	c.Header("Content-Type", "image/gif")
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	// GIF画像を返す
	c.Data(http.StatusOK, "image/gif", gifData)

	// 注意: 実際のトラッキングデータの保存は非同期で行うべきです
	// ここではログ出力のみ行います
	fmt.Printf("Beacon tracking: app_id=%s, session_id=%s, url=%s, referrer=%s, user_agent=%s, ip=%s, custom_params=%v\n",
		appID, sessionID, url, referrer, userAgent, ipAddress, customParams)
}

// ServeGIF は1x1ピクセルGIFビーコンを配信します
func (h *BeaconHandler) ServeGIF(c *gin.Context) {
	// クエリパラメータを取得
	appID := c.Query("app_id")
	sessionID := c.Query("session_id")
	url := c.Query("url")
	if url == "" {
		url = "/"
	}
	referrer := c.Query("referrer")
	userAgent := c.GetHeader("User-Agent")

	// IPアドレスを取得
	ipAddress := c.ClientIP()
	if ipAddress == "" {
		ipAddress = c.Request.RemoteAddr
	}

	// カスタムパラメータを収集
	customParams := make(map[string]interface{})
	for key, values := range c.Request.URL.Query() {
		if key != "app_id" && key != "session_id" && key != "url" && key != "referrer" {
			if len(values) > 0 {
				customParams[key] = values[0]
			}
		}
	}

	// これは最小限のGIF画像データです
	gifData := []byte{
		0x47, 0x49, 0x46, 0x38, 0x39, 0x61, // GIF89a
		0x01, 0x00, 0x01, 0x00, // 1x1ピクセル
		0x80, 0x00, 0x00, // 背景色（透明）
		0x00, 0x00, 0x00, // パレット
		0x21, 0xf9, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00, // グラフィック制御拡張
		0x2c, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, // 画像記述子
		0x02, 0x02, 0x44, 0x01, 0x00, // 画像データ
		0x3b, // 終了
	}

	// レスポンスヘッダーを設定
	c.Header("Content-Type", "image/gif")
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	// GIF画像を返す
	c.Data(http.StatusOK, "image/gif", gifData)

	// 注意: 実際のトラッキングデータの保存は非同期で行うべきです
	// ここではログ出力のみ行います
	fmt.Printf("Beacon tracking: app_id=%s, session_id=%s, url=%s, referrer=%s, user_agent=%s, ip=%s, custom_params=%v\n",
		appID, sessionID, url, referrer, userAgent, ipAddress, customParams)
}

// GenerateBeaconWithConfig はカスタム設定でビーコンを生成します
func (h *BeaconHandler) GenerateBeaconWithConfig(c *gin.Context) {
	var config generator.BeaconConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_CONFIG",
				Message: "Invalid configuration",
				Details: err.Error(),
			},
			Timestamp: time.Now(),
		})
		return
	}

	// 仕様: POSTもGIFを返す
	gifData := []byte{
		0x47, 0x49, 0x46, 0x38, 0x39, 0x61,
		0x01, 0x00, 0x01, 0x00,
		0x80, 0x00, 0x00,
		0x00, 0x00, 0x00,
		0x21, 0xf9, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00,
		0x2c, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00,
		0x02, 0x02, 0x44, 0x01, 0x00,
		0x3b,
	}

	c.Header("Content-Type", "image/gif")
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.Data(http.StatusOK, "image/gif", gifData)
}

// ProcessBeacon はビーコンリクエストを処理します
func (h *BeaconHandler) ProcessBeacon(c *gin.Context) {
	// クエリパラメータを取得
	appID := c.Query("app_id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "app_id parameter is required",
			},
			Timestamp: time.Now(),
		})
		return
	}

	// トラッキングデータを収集
	sessionID := c.Query("session_id")
	url := c.Query("url")
	if url == "" {
		url = "/"
	}
	referrer := c.Query("referrer")
	userAgent := c.GetHeader("User-Agent")

	// IPアドレスを取得
	ipAddress := c.ClientIP()
	if ipAddress == "" {
		ipAddress = c.Request.RemoteAddr
	}

	// カスタムパラメータを収集
	customParams := make(map[string]interface{})
	for key, values := range c.Request.URL.Query() {
		if key != "app_id" && key != "session_id" && key != "url" && key != "referrer" {
			if len(values) > 0 {
				customParams[key] = values[0]
			}
		}
	}

	// 1x1ピクセルの透明GIF画像（Base64エンコード）
	gifData := []byte{
		0x47, 0x49, 0x46, 0x38, 0x39, 0x61, // GIF89a
		0x01, 0x00, 0x01, 0x00, // 1x1ピクセル
		0x80, 0x00, 0x00, // 背景色（透明）
		0x00, 0x00, 0x00, // パレット
		0x21, 0xf9, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00, // グラフィック制御拡張
		0x2c, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, // 画像記述子
		0x02, 0x02, 0x44, 0x01, 0x00, // 画像データ
		0x3b, // 終了
	}

	// レスポンスヘッダーを設定
	c.Header("Content-Type", "image/gif")
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	// GIF画像を返す
	c.Data(http.StatusOK, "image/gif", gifData)

	// 注意: 実際のトラッキングデータの保存は非同期で行うべきです
	// ここではログ出力のみ行います
	fmt.Printf("Beacon tracking: app_id=%s, session_id=%s, url=%s, referrer=%s, user_agent=%s, ip=%s, custom_params=%v\n",
		appID, sessionID, url, referrer, userAgent, ipAddress, customParams)
}

// Health はビーコンサービスの健全性を返します
func (h *BeaconHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"status": "healthy",
			"beacon": map[string]interface{}{
				"version": "1.0.0",
				"uptime":  time.Since(time.Now()).String(),
			},
		},
		Timestamp: time.Now(),
	})
}
