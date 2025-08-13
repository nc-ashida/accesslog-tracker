package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"accesslog-tracker/internal/api/models"
	"accesslog-tracker/internal/infrastructure/database/postgresql"
	"accesslog-tracker/internal/infrastructure/cache/redis"
	"accesslog-tracker/internal/utils/logger"
)

// HealthHandler はヘルスチェックAPIのハンドラーです
type HealthHandler struct {
	dbConn    *postgresql.Connection
	redisConn *redis.CacheService
	logger    logger.Logger
}

// NewHealthHandler は新しいヘルスチェックハンドラーを作成します
func NewHealthHandler(dbConn *postgresql.Connection, redisConn *redis.CacheService, logger logger.Logger) *HealthHandler {
	return &HealthHandler{
		dbConn:    dbConn,
		redisConn: redisConn,
		logger:    logger,
	}
}

// Health はヘルスチェックを実行します
func (h *HealthHandler) Health(c *gin.Context) {
	status := "healthy"
	services := make(map[string]string)

	// データベースのヘルスチェック
	if err := h.dbConn.Ping(); err != nil {
		status = "unhealthy"
		services["database"] = "unhealthy"
		h.logger.Error("Database health check failed", "error", err.Error())
	} else {
		services["database"] = "healthy"
	}

	// Redisのヘルスチェック
	if err := h.redisConn.Ping(c.Request.Context()); err != nil {
		status = "unhealthy"
		services["redis"] = "unhealthy"
		h.logger.Error("Redis health check failed", "error", err.Error())
	} else {
		services["redis"] = "healthy"
	}

	// レスポンスを作成
	response := models.HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Services:  services,
	}

	// ステータスコードを決定
	statusCode := http.StatusOK
	if status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, models.APIResponse{
		Success: status == "healthy",
		Data:    response,
	})
}

// Readiness はアプリケーションの準備完了状態をチェックします
func (h *HealthHandler) Readiness(c *gin.Context) {
	status := "ready"
	services := make(map[string]string)

	// データベースの準備完了チェック
	if err := h.dbConn.Ping(); err != nil {
		status = "not_ready"
		services["database"] = "not_ready"
		h.logger.Error("Database readiness check failed", "error", err.Error())
	} else {
		services["database"] = "ready"
	}

	// Redisの準備完了チェック
	if err := h.redisConn.Ping(c.Request.Context()); err != nil {
		status = "not_ready"
		services["redis"] = "not_ready"
		h.logger.Error("Redis readiness check failed", "error", err.Error())
	} else {
		services["redis"] = "ready"
	}

	// レスポンスを作成
	response := models.HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Services:  services,
	}

	// ステータスコードを決定
	statusCode := http.StatusOK
	if status == "not_ready" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, models.APIResponse{
		Success: status == "ready",
		Data:    response,
	})
}

// Liveness はアプリケーションの生存状態をチェックします
func (h *HealthHandler) Liveness(c *gin.Context) {
	response := models.HealthResponse{
		Status:    "alive",
		Timestamp: time.Now(),
		Services: map[string]string{
			"application": "alive",
		},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    response,
	})
}
