package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"accesslog-tracker/internal/api/routes"
	"accesslog-tracker/internal/config"
	"accesslog-tracker/internal/domain/services"
	"accesslog-tracker/internal/infrastructure/database/postgresql"
	"accesslog-tracker/internal/infrastructure/cache/redis"
	"accesslog-tracker/internal/utils/logger"
)

// Server はAPIサーバーの構造体です
type Server struct {
	config             *config.Config
	logger             logger.Logger
	router             *gin.Engine
	httpServer         *http.Server
	trackingService    *services.TrackingService
	applicationService *services.ApplicationService
	dbConn             *postgresql.Connection
	redisConn          *redis.CacheService
}

// NewServer は新しいAPIサーバーを作成します
func NewServer(
	config *config.Config,
	logger logger.Logger,
	trackingService *services.TrackingService,
	applicationService *services.ApplicationService,
	dbConn *postgresql.Connection,
	redisConn *redis.CacheService,
) *Server {
	// Ginのモードを設定
	if config.App.Debug == false {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()

	server := &Server{
		config:             config,
		logger:             logger,
		router:             router,
		trackingService:    trackingService,
		applicationService: applicationService,
		dbConn:             dbConn,
		redisConn:          redisConn,
	}

	// ルートを設定
	routes.Setup(router, trackingService, applicationService, dbConn, redisConn, logger)

	// HTTPサーバーを作成
	server.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", config.App.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return server
}

// Start はサーバーを開始します
func (s *Server) Start() error {
	s.logger.Info("Starting API server", 
		"port", s.config.App.Port,
		"debug", s.config.App.Debug)

	// グレースフルシャットダウンの設定
	go s.gracefulShutdown()

	// サーバーを開始
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Error("Failed to start server", "error", err.Error())
		return err
	}

	return nil
}

// Stop はサーバーを停止します
func (s *Server) Stop() error {
	s.logger.Info("Stopping API server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("Failed to shutdown server gracefully", "error", err.Error())
		return err
	}

	s.logger.Info("Server stopped gracefully")
	return nil
}

// gracefulShutdown はグレースフルシャットダウンを処理します
func (s *Server) gracefulShutdown() {
	// シグナルを待機
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("Received shutdown signal")

	// サーバーを停止
	if err := s.Stop(); err != nil {
		s.logger.Error("Error during shutdown", "error", err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}

// GetRouter はルーターを取得します（テスト用）
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

// SetupTest はテスト用のルートを設定します
func (s *Server) SetupTest() {
	routes.SetupTest(s.router, s.trackingService, s.applicationService, s.dbConn, s.redisConn, s.logger)
}
