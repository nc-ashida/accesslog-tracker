package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sirupsen/logrus"

	"accesslog-tracker/internal/api/server"
	"accesslog-tracker/internal/config"
	"accesslog-tracker/internal/domain/services"
	"accesslog-tracker/internal/infrastructure/database/postgresql"
	"accesslog-tracker/internal/infrastructure/cache/redis"
	"accesslog-tracker/internal/utils/logger"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
	GoVersion = "unknown"
)

func main() {
	// 設定の読み込み
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// ロガーの初期化
	logger := logger.NewLogger()
	logger.WithFields(logrus.Fields{
		"version":    Version,
		"buildTime":  BuildTime,
		"goVersion":  GoVersion,
		"environment": cfg.Environment,
	}).Info("Starting Access Log Tracker API Server")

	// データベース接続の初期化
	dbConn, err := postgresql.NewConnection(cfg.Database.Host, cfg.Database.Port, cfg.Database.Name, cfg.Database.User, cfg.Database.Password)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer dbConn.Close()

	// Redis接続の初期化
	redisConn := redis.NewCacheService(fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port))
	if err := redisConn.Connect(); err != nil {
		logger.WithError(err).Fatal("Failed to connect to Redis")
	}
	defer redisConn.Close()

	// リポジトリの初期化
	trackingRepo := postgresql.NewTrackingRepository(dbConn.GetDB())
	applicationRepo := postgresql.NewApplicationRepository(dbConn.GetDB())

	// サービスの初期化
	trackingService := services.NewTrackingService(trackingRepo, redisConn)
	applicationService := services.NewApplicationService(applicationRepo)

	// APIサーバーの初期化
	apiServer := server.NewServer(
		cfg,
		logger,
		trackingService,
		applicationService,
		dbConn,
		redisConn,
	)

	// サーバーの開始
	if err := apiServer.Start(); err != nil {
		logger.WithError(err).Fatal("Failed to start server")
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
