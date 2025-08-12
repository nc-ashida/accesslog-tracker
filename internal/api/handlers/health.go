package handlers

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// HealthHandler ヘルスチェックハンドラー
type HealthHandler struct {
	db          *gorm.DB
	redisClient *redis.Connection
}

// NewHealthHandler ヘルスチェックハンドラーを作成
func NewHealthHandler(db *gorm.DB, redisClient *redis.Connection) *HealthHandler {
	return &HealthHandler{
		db:          db,
		redisClient: redisClient,
	}
}

// HealthResponse ヘルスチェックレスポンス構造体
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version"`
	Services  map[string]ServiceInfo `json:"services"`
	System    SystemInfo             `json:"system"`
}

// ServiceInfo サービス情報構造体
type ServiceInfo struct {
	Status    string        `json:"status"`
	Latency   time.Duration `json:"latency,omitempty"`
	Error     string        `json:"error,omitempty"`
	LastCheck time.Time     `json:"last_check"`
}

// SystemInfo システム情報構造体
type SystemInfo struct {
	GoVersion    string `json:"go_version"`
	Architecture string `json:"architecture"`
	OS           string `json:"os"`
	NumCPU       int    `json:"num_cpu"`
	NumGoroutine int    `json:"num_goroutine"`
	Memory       MemoryInfo `json:"memory"`
}

// MemoryInfo メモリ情報構造体
type MemoryInfo struct {
	Alloc      uint64 `json:"alloc"`
	TotalAlloc uint64 `json:"total_alloc"`
	Sys        uint64 `json:"sys"`
	NumGC      uint32 `json:"num_gc"`
}

// Health ヘルスチェック
func (h *HealthHandler) Health(c *gin.Context) {
	start := time.Now()
	
	// サービスチェック
	services := make(map[string]ServiceInfo)
	
	// データベースチェック
	dbStatus := h.checkDatabase()
	services["database"] = dbStatus
	
	// Redisチェック
	redisStatus := h.checkRedis()
	services["redis"] = redisStatus
	
	// システム情報
	systemInfo := h.getSystemInfo()
	
	// 全体のステータスを決定
	overallStatus := "healthy"
	for _, service := range services {
		if service.Status == "unhealthy" {
			overallStatus = "unhealthy"
			break
		}
	}
	
	response := HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Version:   "1.0.0", // TODO: バージョン情報を設定ファイルから取得
		Services:  services,
		System:    systemInfo,
	}
	
	// レスポンス時間を計算
	latency := time.Since(start)
	
	// ログ出力
	logrus.WithFields(logrus.Fields{
		"status":  overallStatus,
		"latency": latency,
		"services": services,
	}).Info("Health check completed")
	
	// ステータスコードを決定
	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}
	
	c.JSON(statusCode, response)
}

// Liveness ライブネスチェック（軽量）
func (h *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"timestamp": time.Now(),
	})
}

// Readiness レディネスチェック（依存関係チェック）
func (h *HealthHandler) Readiness(c *gin.Context) {
	services := make(map[string]ServiceInfo)
	
	// データベースチェック
	dbStatus := h.checkDatabase()
	services["database"] = dbStatus
	
	// Redisチェック
	redisStatus := h.checkRedis()
	services["redis"] = redisStatus
	
	// 全体のステータスを決定
	overallStatus := "ready"
	for _, service := range services {
		if service.Status == "unhealthy" {
			overallStatus = "not_ready"
			break
		}
	}
	
	statusCode := http.StatusOK
	if overallStatus == "not_ready" {
		statusCode = http.StatusServiceUnavailable
	}
	
	c.JSON(statusCode, gin.H{
		"status":    overallStatus,
		"timestamp": time.Now(),
		"services":  services,
	})
}

// checkDatabase データベース接続チェック
func (h *HealthHandler) checkDatabase() ServiceInfo {
	start := time.Now()
	
	// シンプルなクエリを実行
	sqlDB, err := h.db.DB()
	if err != nil {
		return ServiceInfo{
			Status:    "unhealthy",
			Error:     err.Error(),
			LastCheck: time.Now(),
		}
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	err = sqlDB.PingContext(ctx)
	latency := time.Since(start)
	
	if err != nil {
		return ServiceInfo{
			Status:    "unhealthy",
			Error:     err.Error(),
			Latency:   latency,
			LastCheck: time.Now(),
		}
	}
	
	return ServiceInfo{
		Status:    "healthy",
		Latency:   latency,
		LastCheck: time.Now(),
	}
}

// checkRedis Redis接続チェック
func (h *HealthHandler) checkRedis() ServiceInfo {
	start := time.Now()
	
	err := h.redisClient.Ping()
	latency := time.Since(start)
	
	if err != nil {
		return ServiceInfo{
			Status:    "unhealthy",
			Error:     err.Error(),
			Latency:   latency,
			LastCheck: time.Now(),
		}
	}
	
	return ServiceInfo{
		Status:    "healthy",
		Latency:   latency,
		LastCheck: time.Now(),
	}
}

// getSystemInfo システム情報を取得
func (h *HealthHandler) getSystemInfo() SystemInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return SystemInfo{
		GoVersion:    runtime.Version(),
		Architecture: runtime.GOARCH,
		OS:           runtime.GOOS,
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
		Memory: MemoryInfo{
			Alloc:      m.Alloc,
			TotalAlloc: m.TotalAlloc,
			Sys:        m.Sys,
			NumGC:      m.NumGC,
		},
	}
}

// Metrics Prometheusメトリクスエンドポイント（将来の実装用）
func (h *HealthHandler) Metrics(c *gin.Context) {
	// TODO: Prometheusメトリクスを実装
	c.JSON(http.StatusOK, gin.H{
		"message": "Metrics endpoint - to be implemented",
	})
}
