package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/your-username/accesslog-tracker/internal/domain/models"
	"github.com/your-username/accesslog-tracker/internal/domain/services"
	"github.com/your-username/accesslog-tracker/internal/utils/iputil"
)

// TrackingHandler トラッキングハンドラー
type TrackingHandler struct {
	trackingService *services.TrackingService
	applicationService *services.ApplicationService
}

// NewTrackingHandler トラッキングハンドラーを作成
func NewTrackingHandler(
	trackingService *services.TrackingService,
	applicationService *services.ApplicationService,
) *TrackingHandler {
	return &TrackingHandler{
		trackingService:    trackingService,
		applicationService: applicationService,
	}
}

// TrackRequest トラッキングリクエスト構造体
type TrackRequest struct {
	AppID         string                 `json:"app_id" binding:"required"`
	ClientSubID   string                 `json:"client_sub_id,omitempty"`
	ModuleID      string                 `json:"module_id,omitempty"`
	URL           string                 `json:"url,omitempty"`
	Referrer      string                 `json:"referrer,omitempty"`
	UserAgent     string                 `json:"user_agent" binding:"required"`
	IPAddress     string                 `json:"ip_address,omitempty"`
	SessionID     string                 `json:"session_id,omitempty"`
	ScreenResolution string              `json:"screen_resolution,omitempty"`
	Language      string                 `json:"language,omitempty"`
	Timezone      string                 `json:"timezone,omitempty"`
	CustomParams  map[string]interface{} `json:"custom_params,omitempty"`
}

// TrackResponse トラッキングレスポンス構造体
type TrackResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// Track トラッキングデータを受信
func (h *TrackingHandler) Track(c *gin.Context) {
	var req TrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Failed to bind tracking request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// アプリケーションの存在確認
	app, err := h.applicationService.GetByID(req.AppID)
	if err != nil {
		logrus.WithError(err).WithField("application_id", req.AppID).Error("Application not found")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Application not found",
		})
		return
	}

	// セッションIDがなければ生成
	if req.SessionID == "" {
		req.SessionID = uuid.New().String()
	}

	// クライアントIPを取得
	clientIP := iputil.GetClientIP(c.Request)

	// トラッキングデータを作成
	trackingData := &models.Tracking{
		ID:            uuid.New().String(),
		ApplicationID: req.AppID,
		SessionID:     req.SessionID,
		PageURL:       req.URL,
		Referrer:      req.Referrer,
		UserAgent:     req.UserAgent,
		ClientIP:      clientIP,
		ScreenWidth:   req.ScreenResolution,
		ScreenHeight:  req.ScreenResolution,
		Language:      req.Language,
		Timezone:      req.Timezone,
		Timestamp:     time.Now(),
		CustomParams:  req.CustomParams,
	}

	// トラッキングデータを保存
	err = h.trackingService.Track(trackingData)
	if err != nil {
		logrus.WithError(err).Error("Failed to save tracking data")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save tracking data",
		})
		return
	}

	// レスポンスを返す
	c.JSON(http.StatusOK, TrackResponse{
		Success: true,
		Message: "Tracking data saved successfully",
		Timestamp: time.Now(),
	})
}

// TrackBeacon ビーコン用トラッキング（GET/POST両対応）
func (h *TrackingHandler) TrackBeacon(c *gin.Context) {
	// クエリパラメータからデータを取得
	applicationID := c.Query("aid")
	if applicationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Application ID is required",
		})
		return
	}

	// アプリケーションの存在確認
	app, err := h.applicationService.GetByID(applicationID)
	if err != nil {
		logrus.WithError(err).WithField("application_id", applicationID).Error("Application not found")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Application not found",
		})
		return
	}

	// セッションIDを取得または生成
	sessionID := c.Query("sid")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	// クライアントIPを取得
	clientIP := iputil.GetClientIP(c.Request)

	// トラッキングデータを作成
	trackingData := &models.Tracking{
		ID:            uuid.New().String(),
		ApplicationID: applicationID,
		SessionID:     sessionID,
		PageURL:       c.Query("url"),
		Referrer:      c.Query("ref"),
		UserAgent:     c.Request.UserAgent(),
		ClientIP:      clientIP,
		Timestamp:     time.Now(),
	}

	// カスタムパラメータを処理
	customParams := make(map[string]interface{})
	for key, values := range c.Request.URL.Query() {
		if key != "aid" && key != "sid" && key != "url" && key != "ref" {
			if len(values) == 1 {
				customParams[key] = values[0]
			} else {
				customParams[key] = values
			}
		}
	}
	trackingData.CustomParams = customParams

	// トラッキングデータを保存
	err = h.trackingService.Track(trackingData)
	if err != nil {
		logrus.WithError(err).Error("Failed to save tracking data")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save tracking data",
		})
		return
	}

	// 1x1ピクセルの透明GIFを返す
	c.Header("Content-Type", "image/gif")
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	
	// 透明GIFのバイトデータ
	transparentGIF := []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00, 0x80, 0x00, 0x00, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x21, 0xf9, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00, 0x2c, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44, 0x01, 0x00, 0x3b}
	c.Data(http.StatusOK, "image/gif", transparentGIF)
}

// GetTrackingStats トラッキング統計を取得
func (h *TrackingHandler) GetTrackingStats(c *gin.Context) {
	applicationID := c.Param("application_id")
	if applicationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Application ID is required",
		})
		return
	}

	// 期間パラメータを取得
	period := c.DefaultQuery("period", "24h")
	
	var startTime time.Time
	switch period {
	case "1h":
		startTime = time.Now().Add(-1 * time.Hour)
	case "24h":
		startTime = time.Now().Add(-24 * time.Hour)
	case "7d":
		startTime = time.Now().Add(-7 * 24 * time.Hour)
	case "30d":
		startTime = time.Now().Add(-30 * 24 * time.Hour)
	default:
		startTime = time.Now().Add(-24 * time.Hour)
	}

	// 統計データを取得
	stats, err := h.trackingService.GetStats(applicationID, startTime, time.Now())
	if err != nil {
		logrus.WithError(err).Error("Failed to get tracking stats")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get tracking stats",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetSessionData セッションデータを取得
func (h *TrackingHandler) GetSessionData(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Session ID is required",
		})
		return
	}

	// セッションデータを取得
	sessionData, err := h.trackingService.GetSessionData(sessionID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get session data")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get session data",
		})
		return
	}

	c.JSON(http.StatusOK, sessionData)
}
