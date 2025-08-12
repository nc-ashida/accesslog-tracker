package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/your-username/accesslog-tracker/internal/domain/models"
	"github.com/your-username/accesslog-tracker/internal/domain/services"
)

// WebhookHandler Webhook管理ハンドラー
type WebhookHandler struct {
	webhookService *services.WebhookService
	applicationService *services.ApplicationService
}

// NewWebhookHandler Webhook管理ハンドラーを作成
func NewWebhookHandler(
	webhookService *services.WebhookService,
	applicationService *services.ApplicationService,
) *WebhookHandler {
	return &WebhookHandler{
		webhookService:    webhookService,
		applicationService: applicationService,
	}
}

// CreateWebhookRequest Webhook作成リクエスト
type CreateWebhookRequest struct {
	ApplicationID string   `json:"application_id" binding:"required"`
	Name          string   `json:"name" binding:"required"`
	URL           string   `json:"url" binding:"required"`
	Events        []string `json:"events" binding:"required"`
	Secret        string   `json:"secret"`
	IsActive      bool     `json:"is_active"`
}

// UpdateWebhookRequest Webhook更新リクエスト
type UpdateWebhookRequest struct {
	Name     string   `json:"name"`
	URL      string   `json:"url"`
	Events   []string `json:"events"`
	Secret   string   `json:"secret"`
	IsActive *bool    `json:"is_active"`
}

// CreateWebhook Webhookを作成
func (h *WebhookHandler) CreateWebhook(c *gin.Context) {
	var req CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Failed to bind create webhook request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// アプリケーションの存在確認
	app, err := h.applicationService.GetByID(req.ApplicationID)
	if err != nil {
		logrus.WithError(err).WithField("application_id", req.ApplicationID).Error("Application not found")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Application not found",
		})
		return
	}

	// 権限チェック
	userID, exists := c.Get("user_id")
	if !exists || userID.(string) != app.UserID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
		return
	}

	// Webhookを作成
	webhook := &models.Webhook{
		ID:            uuid.New().String(),
		ApplicationID: req.ApplicationID,
		Name:          req.Name,
		URL:           req.URL,
		Events:        req.Events,
		Secret:        req.Secret,
		IsActive:      req.IsActive,
	}

	err = h.webhookService.Create(webhook)
	if err != nil {
		logrus.WithError(err).Error("Failed to create webhook")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create webhook",
		})
		return
	}

	c.JSON(http.StatusCreated, webhook)
}

// GetWebhook Webhookを取得
func (h *WebhookHandler) GetWebhook(c *gin.Context) {
	webhookID := c.Param("webhook_id")
	if webhookID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Webhook ID is required",
		})
		return
	}

	// Webhookを取得
	webhook, err := h.webhookService.GetByID(webhookID)
	if err != nil {
		logrus.WithError(err).WithField("webhook_id", webhookID).Error("Webhook not found")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Webhook not found",
		})
		return
	}

	// アプリケーションを取得して権限チェック
	app, err := h.applicationService.GetByID(webhook.ApplicationID)
	if err != nil {
		logrus.WithError(err).Error("Application not found")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Application not found",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists || userID.(string) != app.UserID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
		return
	}

	c.JSON(http.StatusOK, webhook)
}

// UpdateWebhook Webhookを更新
func (h *WebhookHandler) UpdateWebhook(c *gin.Context) {
	webhookID := c.Param("webhook_id")
	if webhookID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Webhook ID is required",
		})
		return
	}

	var req UpdateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Failed to bind update webhook request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// 既存のWebhookを取得
	existingWebhook, err := h.webhookService.GetByID(webhookID)
	if err != nil {
		logrus.WithError(err).WithField("webhook_id", webhookID).Error("Webhook not found")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Webhook not found",
		})
		return
	}

	// アプリケーションを取得して権限チェック
	app, err := h.applicationService.GetByID(existingWebhook.ApplicationID)
	if err != nil {
		logrus.WithError(err).Error("Application not found")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Application not found",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists || userID.(string) != app.UserID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
		return
	}

	// 更新フィールドを設定
	if req.Name != "" {
		existingWebhook.Name = req.Name
	}
	if req.URL != "" {
		existingWebhook.URL = req.URL
	}
	if req.Events != nil {
		existingWebhook.Events = req.Events
	}
	if req.Secret != "" {
		existingWebhook.Secret = req.Secret
	}
	if req.IsActive != nil {
		existingWebhook.IsActive = *req.IsActive
	}

	// Webhookを更新
	err = h.webhookService.Update(existingWebhook)
	if err != nil {
		logrus.WithError(err).Error("Failed to update webhook")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update webhook",
		})
		return
	}

	c.JSON(http.StatusOK, existingWebhook)
}

// DeleteWebhook Webhookを削除
func (h *WebhookHandler) DeleteWebhook(c *gin.Context) {
	webhookID := c.Param("webhook_id")
	if webhookID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Webhook ID is required",
		})
		return
	}

	// 既存のWebhookを取得
	existingWebhook, err := h.webhookService.GetByID(webhookID)
	if err != nil {
		logrus.WithError(err).WithField("webhook_id", webhookID).Error("Webhook not found")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Webhook not found",
		})
		return
	}

	// アプリケーションを取得して権限チェック
	app, err := h.applicationService.GetByID(existingWebhook.ApplicationID)
	if err != nil {
		logrus.WithError(err).Error("Application not found")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Application not found",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists || userID.(string) != app.UserID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
		return
	}

	// Webhookを削除
	err = h.webhookService.Delete(webhookID)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete webhook")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete webhook",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Webhook deleted successfully",
	})
}

// ListWebhooks Webhook一覧を取得
func (h *WebhookHandler) ListWebhooks(c *gin.Context) {
	applicationID := c.Param("application_id")
	if applicationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Application ID is required",
		})
		return
	}

	// アプリケーションの存在確認と権限チェック
	app, err := h.applicationService.GetByID(applicationID)
	if err != nil {
		logrus.WithError(err).WithField("application_id", applicationID).Error("Application not found")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Application not found",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists || userID.(string) != app.UserID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
		return
	}

	// ページネーションパラメータを取得
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	// Webhook一覧を取得
	webhooks, total, err := h.webhookService.ListByApplicationID(applicationID, page, limit)
	if err != nil {
		logrus.WithError(err).Error("Failed to list webhooks")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list webhooks",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"webhooks": webhooks,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// TestWebhook Webhookをテスト
func (h *WebhookHandler) TestWebhook(c *gin.Context) {
	webhookID := c.Param("webhook_id")
	if webhookID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Webhook ID is required",
		})
		return
	}

	// Webhookを取得
	webhook, err := h.webhookService.GetByID(webhookID)
	if err != nil {
		logrus.WithError(err).WithField("webhook_id", webhookID).Error("Webhook not found")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Webhook not found",
		})
		return
	}

	// アプリケーションを取得して権限チェック
	app, err := h.applicationService.GetByID(webhook.ApplicationID)
	if err != nil {
		logrus.WithError(err).Error("Application not found")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Application not found",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists || userID.(string) != app.UserID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
		return
	}

	// テストイベントを送信
	err = h.webhookService.SendTestEvent(webhook)
	if err != nil {
		logrus.WithError(err).Error("Failed to send test webhook")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to send test webhook",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Test webhook sent successfully",
	})
}

// GetWebhookLogs Webhookログを取得
func (h *WebhookHandler) GetWebhookLogs(c *gin.Context) {
	webhookID := c.Param("webhook_id")
	if webhookID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Webhook ID is required",
		})
		return
	}

	// Webhookを取得
	webhook, err := h.webhookService.GetByID(webhookID)
	if err != nil {
		logrus.WithError(err).WithField("webhook_id", webhookID).Error("Webhook not found")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Webhook not found",
		})
		return
	}

	// アプリケーションを取得して権限チェック
	app, err := h.applicationService.GetByID(webhook.ApplicationID)
	if err != nil {
		logrus.WithError(err).Error("Application not found")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Application not found",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists || userID.(string) != app.UserID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
		return
	}

	// ページネーションパラメータを取得
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	// Webhookログを取得
	logs, total, err := h.webhookService.GetLogs(webhookID, page, limit)
	if err != nil {
		logrus.WithError(err).Error("Failed to get webhook logs")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get webhook logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs": logs,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}
