package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/your-username/accesslog-tracker/internal/domain/services"
)

// StatisticsHandler 統計ハンドラー
type StatisticsHandler struct {
	statisticsService *services.StatisticsService
	applicationService *services.ApplicationService
}

// NewStatisticsHandler 統計ハンドラーを作成
func NewStatisticsHandler(
	statisticsService *services.StatisticsService,
	applicationService *services.ApplicationService,
) *StatisticsHandler {
	return &StatisticsHandler{
		statisticsService:  statisticsService,
		applicationService: applicationService,
	}
}

// GetApplicationStats アプリケーション統計を取得
func (h *StatisticsHandler) GetApplicationStats(c *gin.Context) {
	applicationID := c.Param("application_id")
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

	// 期間パラメータを取得
	period := c.DefaultQuery("period", "24h")
	startTime, endTime := h.parseTimeRange(period)

	// 統計データを取得
	stats, err := h.statisticsService.GetApplicationStats(applicationID, startTime, endTime)
	if err != nil {
		logrus.WithError(err).Error("Failed to get application stats")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get application stats",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetPageViews ページビュー統計を取得
func (h *StatisticsHandler) GetPageViews(c *gin.Context) {
	applicationID := c.Param("application_id")
	if applicationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Application ID is required",
		})
		return
	}

	// 期間パラメータを取得
	period := c.DefaultQuery("period", "24h")
	startTime, endTime := h.parseTimeRange(period)

	// ページビュー統計を取得
	pageViews, err := h.statisticsService.GetPageViews(applicationID, startTime, endTime)
	if err != nil {
		logrus.WithError(err).Error("Failed to get page views")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get page views",
		})
		return
	}

	c.JSON(http.StatusOK, pageViews)
}

// GetReferrers リファラー統計を取得
func (h *StatisticsHandler) GetReferrers(c *gin.Context) {
	applicationID := c.Param("application_id")
	if applicationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Application ID is required",
		})
		return
	}

	// 期間パラメータを取得
	period := c.DefaultQuery("period", "24h")
	startTime, endTime := h.parseTimeRange(period)

	// リファラー統計を取得
	referrers, err := h.statisticsService.GetReferrers(applicationID, startTime, endTime)
	if err != nil {
		logrus.WithError(err).Error("Failed to get referrers")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get referrers",
		})
		return
	}

	c.JSON(http.StatusOK, referrers)
}

// GetUserAgents ユーザーエージェント統計を取得
func (h *StatisticsHandler) GetUserAgents(c *gin.Context) {
	applicationID := c.Param("application_id")
	if applicationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Application ID is required",
		})
		return
	}

	// 期間パラメータを取得
	period := c.DefaultQuery("period", "24h")
	startTime, endTime := h.parseTimeRange(period)

	// ユーザーエージェント統計を取得
	userAgents, err := h.statisticsService.GetUserAgents(applicationID, startTime, endTime)
	if err != nil {
		logrus.WithError(err).Error("Failed to get user agents")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user agents",
		})
		return
	}

	c.JSON(http.StatusOK, userAgents)
}

// GetGeographicStats 地理的統計を取得
func (h *StatisticsHandler) GetGeographicStats(c *gin.Context) {
	applicationID := c.Param("application_id")
	if applicationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Application ID is required",
		})
		return
	}

	// 期間パラメータを取得
	period := c.DefaultQuery("period", "24h")
	startTime, endTime := h.parseTimeRange(period)

	// 地理的統計を取得
	geoStats, err := h.statisticsService.GetGeographicStats(applicationID, startTime, endTime)
	if err != nil {
		logrus.WithError(err).Error("Failed to get geographic stats")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get geographic stats",
		})
		return
	}

	c.JSON(http.StatusOK, geoStats)
}

// GetTimeSeriesData 時系列データを取得
func (h *StatisticsHandler) GetTimeSeriesData(c *gin.Context) {
	applicationID := c.Param("application_id")
	if applicationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Application ID is required",
		})
		return
	}

	// 期間パラメータを取得
	period := c.DefaultQuery("period", "24h")
	startTime, endTime := h.parseTimeRange(period)

	// 間隔パラメータを取得
	interval := c.DefaultQuery("interval", "1h")

	// 時系列データを取得
	timeSeries, err := h.statisticsService.GetTimeSeriesData(applicationID, startTime, endTime, interval)
	if err != nil {
		logrus.WithError(err).Error("Failed to get time series data")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get time series data",
		})
		return
	}

	c.JSON(http.StatusOK, timeSeries)
}

// GetCustomParamStats カスタムパラメータ統計を取得
func (h *StatisticsHandler) GetCustomParamStats(c *gin.Context) {
	applicationID := c.Param("application_id")
	if applicationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Application ID is required",
		})
		return
	}

	paramName := c.Param("param_name")
	if paramName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Parameter name is required",
		})
		return
	}

	// 期間パラメータを取得
	period := c.DefaultQuery("period", "24h")
	startTime, endTime := h.parseTimeRange(period)

	// カスタムパラメータ統計を取得
	paramStats, err := h.statisticsService.GetCustomParamStats(applicationID, paramName, startTime, endTime)
	if err != nil {
		logrus.WithError(err).Error("Failed to get custom parameter stats")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get custom parameter stats",
		})
		return
	}

	c.JSON(http.StatusOK, paramStats)
}

// GetRealTimeStats リアルタイム統計を取得
func (h *StatisticsHandler) GetRealTimeStats(c *gin.Context) {
	applicationID := c.Param("application_id")
	if applicationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Application ID is required",
		})
		return
	}

	// リアルタイム統計を取得
	realTimeStats, err := h.statisticsService.GetRealTimeStats(applicationID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get real-time stats")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get real-time stats",
		})
		return
	}

	c.JSON(http.StatusOK, realTimeStats)
}

// parseTimeRange 時間範囲を解析
func (h *StatisticsHandler) parseTimeRange(period string) (time.Time, time.Time) {
	endTime := time.Now()
	var startTime time.Time

	switch period {
	case "1h":
		startTime = endTime.Add(-1 * time.Hour)
	case "6h":
		startTime = endTime.Add(-6 * time.Hour)
	case "24h":
		startTime = endTime.Add(-24 * time.Hour)
	case "7d":
		startTime = endTime.Add(-7 * 24 * time.Hour)
	case "30d":
		startTime = endTime.Add(-30 * 24 * time.Hour)
	case "90d":
		startTime = endTime.Add(-90 * 24 * time.Hour)
	default:
		startTime = endTime.Add(-24 * time.Hour)
	}

	return startTime, endTime
}
