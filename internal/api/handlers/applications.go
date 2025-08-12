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

// ApplicationHandler アプリケーション管理ハンドラー
type ApplicationHandler struct {
	applicationService *services.ApplicationService
}

// NewApplicationHandler アプリケーション管理ハンドラーを作成
func NewApplicationHandler(applicationService *services.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{
		applicationService: applicationService,
	}
}

// CreateApplicationRequest アプリケーション作成リクエスト
type CreateApplicationRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Domain      string `json:"domain"`
	Settings    map[string]interface{} `json:"settings"`
}

// UpdateApplicationRequest アプリケーション更新リクエスト
type UpdateApplicationRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Domain      string `json:"domain"`
	Settings    map[string]interface{} `json:"settings"`
	IsActive    *bool  `json:"is_active"`
}

// CreateApplication アプリケーションを作成
func (h *ApplicationHandler) CreateApplication(c *gin.Context) {
	var req CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Failed to bind create application request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// ユーザーIDを取得（認証ミドルウェアから）
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found",
		})
		return
	}

	// アプリケーションを作成
	application := &models.Application{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Domain:      req.Domain,
		UserID:      userID.(string),
		Settings:    req.Settings,
		IsActive:    true,
	}

	err := h.applicationService.Create(application)
	if err != nil {
		logrus.WithError(err).Error("Failed to create application")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create application",
		})
		return
	}

	c.JSON(http.StatusCreated, application)
}

// GetApplication アプリケーションを取得
func (h *ApplicationHandler) GetApplication(c *gin.Context) {
	applicationID := c.Param("application_id")
	if applicationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Application ID is required",
		})
		return
	}

	// アプリケーションを取得
	application, err := h.applicationService.GetByID(applicationID)
	if err != nil {
		logrus.WithError(err).WithField("application_id", applicationID).Error("Application not found")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Application not found",
		})
		return
	}

	// 権限チェック（ユーザーIDが一致するか）
	userID, exists := c.Get("user_id")
	if !exists || userID.(string) != application.UserID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
		return
	}

	c.JSON(http.StatusOK, application)
}

// UpdateApplication アプリケーションを更新
func (h *ApplicationHandler) UpdateApplication(c *gin.Context) {
	applicationID := c.Param("application_id")
	if applicationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Application ID is required",
		})
		return
	}

	var req UpdateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Failed to bind update application request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// 既存のアプリケーションを取得
	existingApp, err := h.applicationService.GetByID(applicationID)
	if err != nil {
		logrus.WithError(err).WithField("application_id", applicationID).Error("Application not found")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Application not found",
		})
		return
	}

	// 権限チェック
	userID, exists := c.Get("user_id")
	if !exists || userID.(string) != existingApp.UserID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
		return
	}

	// 更新フィールドを設定
	if req.Name != "" {
		existingApp.Name = req.Name
	}
	if req.Description != "" {
		existingApp.Description = req.Description
	}
	if req.Domain != "" {
		existingApp.Domain = req.Domain
	}
	if req.Settings != nil {
		existingApp.Settings = req.Settings
	}
	if req.IsActive != nil {
		existingApp.IsActive = *req.IsActive
	}

	// アプリケーションを更新
	err = h.applicationService.Update(existingApp)
	if err != nil {
		logrus.WithError(err).Error("Failed to update application")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update application",
		})
		return
	}

	c.JSON(http.StatusOK, existingApp)
}

// DeleteApplication アプリケーションを削除
func (h *ApplicationHandler) DeleteApplication(c *gin.Context) {
	applicationID := c.Param("application_id")
	if applicationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Application ID is required",
		})
		return
	}

	// 既存のアプリケーションを取得
	existingApp, err := h.applicationService.GetByID(applicationID)
	if err != nil {
		logrus.WithError(err).WithField("application_id", applicationID).Error("Application not found")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Application not found",
		})
		return
	}

	// 権限チェック
	userID, exists := c.Get("user_id")
	if !exists || userID.(string) != existingApp.UserID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
		return
	}

	// アプリケーションを削除
	err = h.applicationService.Delete(applicationID)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete application")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete application",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Application deleted successfully",
	})
}

// ListApplications アプリケーション一覧を取得
func (h *ApplicationHandler) ListApplications(c *gin.Context) {
	// ユーザーIDを取得
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found",
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

	// アプリケーション一覧を取得
	applications, total, err := h.applicationService.ListByUserID(userID.(string), page, limit)
	if err != nil {
		logrus.WithError(err).Error("Failed to list applications")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list applications",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"applications": applications,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetApplicationAPIKey アプリケーションのAPIキーを取得
func (h *ApplicationHandler) GetApplicationAPIKey(c *gin.Context) {
	applicationID := c.Param("application_id")
	if applicationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Application ID is required",
		})
		return
	}

	// アプリケーションを取得
	application, err := h.applicationService.GetByID(applicationID)
	if err != nil {
		logrus.WithError(err).WithField("application_id", applicationID).Error("Application not found")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Application not found",
		})
		return
	}

	// 権限チェック
	userID, exists := c.Get("user_id")
	if !exists || userID.(string) != application.UserID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"api_key": application.APIKey,
	})
}

// RegenerateAPIKey APIキーを再生成
func (h *ApplicationHandler) RegenerateAPIKey(c *gin.Context) {
	applicationID := c.Param("application_id")
	if applicationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Application ID is required",
		})
		return
	}

	// アプリケーションを取得
	application, err := h.applicationService.GetByID(applicationID)
	if err != nil {
		logrus.WithError(err).WithField("application_id", applicationID).Error("Application not found")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Application not found",
		})
		return
	}

	// 権限チェック
	userID, exists := c.Get("user_id")
	if !exists || userID.(string) != application.UserID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
		return
	}

	// APIキーを再生成
	newAPIKey, err := h.applicationService.RegenerateAPIKey(applicationID)
	if err != nil {
		logrus.WithError(err).Error("Failed to regenerate API key")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to regenerate API key",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"api_key": newAPIKey,
		"message": "API key regenerated successfully",
	})
}
