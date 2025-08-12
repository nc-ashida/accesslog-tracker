package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/your-username/accesslog-tracker/internal/beacon/generator"
)

// BeaconHandler ビーコンハンドラー
type BeaconHandler struct {
	beaconGenerator *generator.BeaconGenerator
}

// NewBeaconHandler ビーコンハンドラーを作成
func NewBeaconHandler(beaconGenerator *generator.BeaconGenerator) *BeaconHandler {
	return &BeaconHandler{
		beaconGenerator: beaconGenerator,
	}
}

// BeaconRequest ビーコン生成リクエスト
type BeaconRequest struct {
	AppID         string `json:"app_id" binding:"required"`
	ClientSubID   string `json:"client_sub_id,omitempty"`
	ModuleID      string `json:"module_id,omitempty"`
	Endpoint      string `json:"endpoint,omitempty"`
	Debug         bool   `json:"debug,omitempty"`
	RespectDNT    bool   `json:"respect_dnt,omitempty"`
	SessionTimeout int   `json:"session_timeout,omitempty"`
}

// BeaconResponse ビーコン生成レスポンス
type BeaconResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// GenerateBeacon ビーコンを生成
func (h *BeaconHandler) GenerateBeacon(c *gin.Context) {
	var req BeaconRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Failed to bind beacon request")
		c.JSON(http.StatusBadRequest, gin.H{
			"success":   false,
			"message":   "Invalid request format",
			"timestamp": "2024-01-01T00:00:00Z",
		})
		return
	}

	// デフォルト値の設定
	if req.Endpoint == "" {
		req.Endpoint = "https://api.access-log-tracker.com/v1/track"
	}
	if req.SessionTimeout == 0 {
		req.SessionTimeout = 1800000 // 30分
	}

	config := generator.BeaconConfig{
		AppID:         req.AppID,
		ClientSubID:   req.ClientSubID,
		ModuleID:      req.ModuleID,
		Endpoint:      req.Endpoint,
		Debug:         req.Debug,
		RespectDNT:    req.RespectDNT,
		SessionTimeout: req.SessionTimeout,
	}

	// ビーコンコードを生成
	beaconCode, err := h.beaconGenerator.GenerateBeacon(config)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate beacon")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":   false,
			"message":   "Failed to generate beacon",
			"timestamp": "2024-01-01T00:00:00Z",
		})
		return
	}

	// 埋め込みコードを生成
	embedCode, err := h.beaconGenerator.GenerateEmbedCode(config)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate embed code")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":   false,
			"message":   "Failed to generate embed code",
			"timestamp": "2024-01-01T00:00:00Z",
		})
		return
	}

	// 圧縮版も生成
	minifiedCode := h.beaconGenerator.MinifyBeacon(beaconCode)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"beacon_code":     beaconCode,
			"minified_code":   minifiedCode,
			"embed_code":      embedCode,
			"app_id":          req.AppID,
			"client_sub_id":   req.ClientSubID,
			"module_id":       req.ModuleID,
			"endpoint":        req.Endpoint,
		},
		"message":   "Beacon generated successfully",
		"timestamp": "2024-01-01T00:00:00Z",
	})
}

// GetBeaconFile ビーコンファイルを配信
func (h *BeaconHandler) GetBeaconFile(c *gin.Context) {
	appID := c.Query("app_id")
	clientSubID := c.Query("client_sub_id")
	moduleID := c.Query("module_id")
	debug := c.Query("debug") == "true"

	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":   false,
			"message":   "app_id is required",
			"timestamp": "2024-01-01T00:00:00Z",
		})
		return
	}

	config := generator.BeaconConfig{
		AppID:         appID,
		ClientSubID:   clientSubID,
		ModuleID:      moduleID,
		Endpoint:      "https://api.access-log-tracker.com/v1/track",
		Debug:         debug,
		RespectDNT:    true,
		SessionTimeout: 1800000,
	}

	beaconCode, err := h.beaconGenerator.GenerateBeacon(config)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate beacon file")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":   false,
			"message":   "Failed to generate beacon file",
			"timestamp": "2024-01-01T00:00:00Z",
		})
		return
	}

	// JavaScriptファイルとして配信
	c.Header("Content-Type", "application/javascript")
	c.Header("Cache-Control", "public, max-age=3600")
	c.String(http.StatusOK, beaconCode)
}
