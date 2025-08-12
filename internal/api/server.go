package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/your-username/accesslog-tracker/internal/api/middleware"
	"github.com/your-username/accesslog-tracker/internal/api/routes"
	"github.com/your-username/accesslog-tracker/internal/infrastructure/cache/redis"
	"gorm.io/gorm"
)

// Server APIサーバー
type Server struct {
	router         *gin.Engine
	httpServer     *http.Server
	db             *gorm.DB
	redisClient    *redis.Connection
	authConfig     middleware.AuthConfig
	rateLimitConfig middleware.RateLimitConfig
	serverConfig   routes.ServerConfig
}

// NewServer サーバーを作成
func NewServer(
	db *gorm.DB,
	redisClient *redis.Connection,
	authConfig middleware.AuthConfig,
	rateLimitConfig middleware.RateLimitConfig,
	serverConfig routes.ServerConfig,
) *Server {
	// ルートを設定
	router := routes.SetupRoutes(db, redisClient, authConfig, rateLimitConfig, serverConfig)

	// 開発用ルートを設定
	routes.SetupDevelopmentRoutes(router, db, redisClient, authConfig, rateLimitConfig)

	// HTTPサーバーを作成
	httpServer := &http.Server{
		Addr:           serverConfig.Port,
		Handler:        router,
		ReadTimeout:    serverConfig.ReadTimeout,
		WriteTimeout:   serverConfig.WriteTimeout,
		MaxHeaderBytes: serverConfig.MaxHeaderBytes,
	}

	return &Server{
		router:          router,
		httpServer:      httpServer,
		db:              db,
		redisClient:     redisClient,
		authConfig:      authConfig,
		rateLimitConfig: rateLimitConfig,
		serverConfig:    serverConfig,
	}
}

// Start サーバーを開始
func (s *Server) Start() error {
	// サーバーを非同期で開始
	go func() {
		logrus.WithFields(logrus.Fields{
			"port":    s.serverConfig.Port,
			"mode":    gin.Mode(),
		}).Info("Starting API server")

		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Fatal("Failed to start server")
		}
	}()

	// グレースフルシャットダウンのためのシグナルを待機
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down server...")

	// グレースフルシャットダウン
	return s.Shutdown()
}

// Shutdown サーバーをシャットダウン
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// HTTPサーバーをシャットダウン
	if err := s.httpServer.Shutdown(ctx); err != nil {
		logrus.WithError(err).Error("Server forced to shutdown")
		return err
	}

	// データベース接続を閉じる
	sqlDB, err := s.db.DB()
	if err == nil {
		if err := sqlDB.Close(); err != nil {
			logrus.WithError(err).Error("Failed to close database connection")
		}
	}

	// Redis接続を閉じる
	if err := s.redisClient.Close(); err != nil {
		logrus.WithError(err).Error("Failed to close Redis connection")
	}

	logrus.Info("Server exited")
	return nil
}

// GetRouter ルーターを取得
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

// GetHTTPServer HTTPサーバーを取得
func (s *Server) GetHTTPServer() *http.Server {
	return s.httpServer
}

// HealthCheck ヘルスチェック
func (s *Server) HealthCheck() error {
	// データベース接続チェック
	sqlDB, err := s.db.DB()
	if err != nil {
		return fmt.Errorf("database connection error: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping error: %w", err)
	}

	// Redis接続チェック
	if err := s.redisClient.Ping(); err != nil {
		return fmt.Errorf("redis ping error: %w", err)
	}

	return nil
}

// DefaultConfig デフォルト設定を作成
func DefaultConfig() (
	middleware.AuthConfig,
	middleware.RateLimitConfig,
	routes.ServerConfig,
) {
	authConfig := middleware.AuthConfig{
		SecretKey:     "your-secret-key", // 本番環境では環境変数から取得
		TokenDuration: 24 * time.Hour,
	}

	rateLimitConfig := middleware.RateLimitConfig{
		RequestsPerMinute: 100,
		RequestsPerHour:   1000,
		BurstSize:         10,
		RedisClient:       nil, // 後で設定
	}

	serverConfig := routes.DefaultServerConfig()

	return authConfig, rateLimitConfig, serverConfig
}
