package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/nc-ashida/accesslog-tracker/internal/api/handlers"
	"github.com/nc-ashida/accesslog-tracker/internal/api/middleware"
	"github.com/nc-ashida/accesslog-tracker/internal/api/routes"
	"github.com/nc-ashida/accesslog-tracker/internal/infrastructure/database/postgresql"
	"github.com/nc-ashida/accesslog-tracker/internal/infrastructure/cache/redis"
	"github.com/nc-ashida/accesslog-tracker/internal/utils/logger"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
	GoVersion = "unknown"
)

func main() {
	// ロガーの初期化
	logger := logger.NewLogger()
	logger.WithFields(logrus.Fields{
		"version":    Version,
		"buildTime":  BuildTime,
		"goVersion":  GoVersion,
	}).Info("Starting Access Log Tracker API Server")

	// 環境変数の取得
	port := getEnv("PORT", "8080")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbName := getEnv("DB_NAME", "access_log_tracker")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "password")
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")

	// データベース接続の初期化
	db, err := postgresql.NewConnection(dbHost, dbPort, dbName, dbUser, dbPassword)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	// Redis接続の初期化
	redisClient, err := redis.NewConnection(redisHost, redisPort, redisPassword)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to Redis")
	}
	defer redisClient.Close()

	// Ginルーターの初期化
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// ミドルウェアの設定
	router.Use(middleware.Logger(logger))
	router.Use(middleware.CORS())
	router.Use(middleware.Recovery(logger))

	// ハンドラーの初期化
	appHandler := handlers.NewApplicationHandler(nil, logger) // TODO: ApplicationServiceの実装後に修正
	beaconHandler := handlers.NewBeaconHandler(db, redisClient, logger)
	healthHandler := handlers.NewHealthHandler(db, redisClient, logger)
	sessionHandler := handlers.NewSessionHandler(db, logger)
	statisticsHandler := handlers.NewStatisticsHandler(db, logger)

	// ルートの設定
	routes.SetupRoutes(router, appHandler, beaconHandler, healthHandler, sessionHandler, statisticsHandler)

	// HTTPサーバーの設定
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// グレースフルシャットダウンの設定
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		logger.Info("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.WithError(err).Error("Server forced to shutdown")
		}
	}()

	// サーバーの起動
	logger.WithField("port", port).Info("Server starting")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.WithError(err).Fatal("Server failed to start")
	}

	logger.Info("Server stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
