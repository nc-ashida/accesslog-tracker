package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"accesslog-tracker/internal/api/models"
	domainmodels "accesslog-tracker/internal/domain/models"
	"accesslog-tracker/internal/domain/services"
	"accesslog-tracker/internal/utils/logger"
	"accesslog-tracker/internal/utils/timeutil"
)

// TrackingHandler はトラッキングAPIのハンドラーです
type TrackingHandler struct {
	trackingService services.TrackingServiceInterface
	logger          logger.Logger
}

// NewTrackingHandler は新しいトラッキングハンドラーを作成します
func NewTrackingHandler(trackingService services.TrackingServiceInterface, logger logger.Logger) *TrackingHandler {
	return &TrackingHandler{
		trackingService: trackingService,
		logger:          logger,
	}
}

// Track はトラッキングデータを受け取って保存します
func (h *TrackingHandler) Track(c *gin.Context) {
	var req models.TrackingRequest
	
	// リクエストボディをバインディング
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid tracking request", "error", err.Error(), "ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid request format",
				Details: err.Error(),
			},
		})
		return
	}

	// アプリケーションIDをコンテキストから取得
	appID, exists := c.Get("app_id")
	if !exists {
		h.logger.Error("App ID not found in context")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_SERVER_ERROR",
				Message: "Application ID not found",
			},
		})
		return
	}

	// リクエストのAppIDと認証されたAppIDが一致するかチェック
	if req.AppID != appID {
		h.logger.Warn("App ID mismatch", "request_app_id", req.AppID, "auth_app_id", appID)
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FORBIDDEN",
				Message: "App ID mismatch",
			},
		})
		return
	}

	// トラッキングデータを作成
	trackingData := &domainmodels.TrackingData{
		AppID:       req.AppID,
		UserAgent:   req.UserAgent,
		URL:         req.URL,
		IPAddress:   req.IPAddress,
		SessionID:   req.SessionID,
		Referrer:    req.Referrer,
		CustomParams: req.CustomParams,
		Timestamp:   time.Now(),
	}

	// トラッキングデータを保存
	err := h.trackingService.ProcessTrackingData(c.Request.Context(), trackingData)
	if err != nil {
		h.logger.Error("Failed to save tracking data", "error", err.Error(), "app_id", req.AppID)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_SERVER_ERROR",
				Message: "Failed to save tracking data",
			},
		})
		return
	}

	// レスポンスを作成
	response := models.TrackingResponse{
		TrackingID: trackingData.ID,
		AppID:      trackingData.AppID,
		SessionID:  trackingData.SessionID,
		Timestamp:  trackingData.Timestamp,
	}

	h.logger.Info("Tracking data saved successfully", 
		"tracking_id", trackingData.ID, 
		"app_id", req.AppID, 
		"session_id", req.SessionID)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    response,
	})
}

// GetStatistics は統計データを取得します
func (h *TrackingHandler) GetStatistics(c *gin.Context) {
	// クエリパラメータを取得
	appID := c.Query("app_id")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// 必須パラメータのチェック
	if appID == "" || startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "app_id, start_date, and end_date are required",
			},
		})
		return
	}

	// 日付のパース
	startDate, err := timeutil.ParseDate(startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid start_date format",
				Details: err.Error(),
			},
		})
		return
	}

	endDate, err := timeutil.ParseDate(endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid end_date format",
				Details: err.Error(),
			},
		})
		return
	}

	// 認証されたアプリケーションIDと一致するかチェック
	authAppID, exists := c.Get("app_id")
	if !exists || appID != authAppID {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "FORBIDDEN",
				Message: "Access denied to this application's data",
			},
		})
		return
	}

	// 統計データを取得
	stats, err := h.trackingService.GetStatistics(c.Request.Context(), appID, startDate, endDate)
	if err != nil {
		h.logger.Error("Failed to get statistics", "error", err.Error(), "app_id", appID)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_SERVER_ERROR",
				Message: "Failed to get statistics",
			},
		})
		return
	}

	// レスポンスを作成
	response := models.StatisticsResponse{
		AppID:         appID,
		StartDate:     startDate,
		EndDate:       endDate,
		TotalRequests: int64(stats.Metrics["total_tracking_count"].(int64)),
		UniqueVisitors: int64(stats.Metrics["total_tracking_count"].(int64)) / 5, // 簡易的な計算
		TopPages:      []models.PageStats{},
		TopReferrers:  []models.ReferrerStats{},
	}

	h.logger.Info("Statistics retrieved successfully", "app_id", appID)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    response,
	})
}
