package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"accesslog-tracker/internal/api/models"
	domainmodels "accesslog-tracker/internal/domain/models"
	"accesslog-tracker/internal/domain/services"
	"accesslog-tracker/internal/utils/logger"
)

// ApplicationHandler はアプリケーションAPIのハンドラーです
type ApplicationHandler struct {
	applicationService *services.ApplicationService
	logger            logger.Logger
}

// NewApplicationHandler は新しいアプリケーションハンドラーを作成します
func NewApplicationHandler(applicationService *services.ApplicationService, logger logger.Logger) *ApplicationHandler {
	return &ApplicationHandler{
		applicationService: applicationService,
		logger:            logger,
	}
}

// Create は新しいアプリケーションを作成します
func (h *ApplicationHandler) Create(c *gin.Context) {
	var req models.ApplicationRequest
	
	// リクエストボディをバインディング
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid application request", "error", err.Error())
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

	// アプリケーションを作成
	app := &domainmodels.Application{
		Name:        req.Name,
		Description: req.Description,
		Domain:      req.Domain,
	}
	err := h.applicationService.Create(c.Request.Context(), app)
	if err != nil {
		h.logger.Error("Failed to create application", "error", err.Error())
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_SERVER_ERROR",
				Message: "Failed to create application",
			},
		})
		return
	}

	// レスポンスを作成
	response := models.ApplicationResponse{
		AppID:      app.AppID,
		Name:       app.Name,
		Description: app.Description,
		Domain:     app.Domain,
		APIKey:     app.APIKey,
		CreatedAt:  app.CreatedAt,
		UpdatedAt:  app.UpdatedAt,
	}

	h.logger.Info("Application created successfully", "app_id", app.AppID, "name", app.Name)

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    response,
	})
}

// Get はアプリケーションの詳細を取得します
func (h *ApplicationHandler) Get(c *gin.Context) {
	appID := c.Param("id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Application ID is required",
			},
		})
		return
	}

	// アプリケーションを取得
	app, err := h.applicationService.GetByID(c.Request.Context(), appID)
	if err != nil {
		h.logger.Error("Failed to get application", "error", err.Error(), "app_id", appID)
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "NOT_FOUND",
				Message: "Application not found",
			},
		})
		return
	}

	// レスポンスを作成
	response := models.ApplicationResponse{
		AppID:      app.AppID,
		Name:       app.Name,
		Description: app.Description,
		Domain:     app.Domain,
		APIKey:     app.APIKey,
		CreatedAt:  app.CreatedAt,
		UpdatedAt:  app.UpdatedAt,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    response,
	})
}

// Update はアプリケーションを更新します
func (h *ApplicationHandler) Update(c *gin.Context) {
	appID := c.Param("id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Application ID is required",
			},
		})
		return
	}

	var req models.ApplicationUpdateRequest
	
	// リクエストボディをバインディング
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid application update request", "error", err.Error())
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

	// アプリケーションを更新
	app := &domainmodels.Application{
		AppID:       appID,
		Name:        req.Name,
		Description: req.Description,
		Domain:      req.Domain,
		Active:      *req.Active,
	}
	err := h.applicationService.Update(c.Request.Context(), app)
	if err != nil {
		h.logger.Error("Failed to update application", "error", err.Error(), "app_id", appID)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_SERVER_ERROR",
				Message: "Failed to update application",
			},
		})
		return
	}

	// レスポンスを作成
	response := models.ApplicationResponse{
		AppID:      app.AppID,
		Name:       app.Name,
		Description: app.Description,
		Domain:     app.Domain,
		APIKey:     app.APIKey,
		CreatedAt:  app.CreatedAt,
		UpdatedAt:  app.UpdatedAt,
	}

	h.logger.Info("Application updated successfully", "app_id", appID)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    response,
	})
}

// List はアプリケーション一覧を取得します
func (h *ApplicationHandler) List(c *gin.Context) {
	// ページネーションパラメータを取得
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// アプリケーション一覧を取得
	limit := pageSize
	offset := (page - 1) * pageSize
	apps, err := h.applicationService.List(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list applications", "error", err.Error())
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_SERVER_ERROR",
				Message: "Failed to list applications",
			},
		})
		return
	}
	
	total, err := h.applicationService.Count(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list applications", "error", err.Error())
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_SERVER_ERROR",
				Message: "Failed to list applications",
			},
		})
		return
	}

	// レスポンスを作成
	responses := make([]models.ApplicationResponse, len(apps))
	for i, app := range apps {
		responses[i] = models.ApplicationResponse{
			AppID:      app.AppID,
			Name:       app.Name,
			Description: app.Description,
			Domain:     app.Domain,
			APIKey:     app.APIKey,
			CreatedAt:  app.CreatedAt,
			UpdatedAt:  app.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: gin.H{
			"applications": responses,
			"pagination": gin.H{
				"page":       page,
				"page_size":  pageSize,
				"total":      total,
				"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
			},
		},
	})
}

// Delete はアプリケーションを削除します
func (h *ApplicationHandler) Delete(c *gin.Context) {
	appID := c.Param("id")
	if appID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Application ID is required",
			},
		})
		return
	}

	// アプリケーションを削除
	err := h.applicationService.Delete(c.Request.Context(), appID)
	if err != nil {
		h.logger.Error("Failed to delete application", "error", err.Error(), "app_id", appID)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INTERNAL_SERVER_ERROR",
				Message: "Failed to delete application",
			},
		})
		return
	}

	h.logger.Info("Application deleted successfully", "app_id", appID)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: gin.H{
			"message": "Application deleted successfully",
		},
	})
}
