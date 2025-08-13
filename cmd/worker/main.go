package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

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
	}).Info("Starting Access Log Tracker Worker")

	// 環境変数の取得
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

	// コンテキストの作成
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ワーカープロセスの開始
	logger.Info("Starting worker processes...")

	// データ処理ワーカー
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.Info("Data processing worker stopped")
				return
			case <-ticker.C:
				// データ処理ロジックをここに実装
				logger.Debug("Processing data...")
			}
		}
	}()

	// クリーンアップワーカー
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.Info("Cleanup worker stopped")
				return
			case <-ticker.C:
				// クリーンアップロジックをここに実装
				logger.Debug("Running cleanup...")
			}
		}
	}()

	// グレースフルシャットダウンの設定
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down worker...")
	cancel()

	// ワーカーの停止を待機
	time.Sleep(5 * time.Second)
	logger.Info("Worker stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
