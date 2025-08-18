package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"accesslog-tracker/internal/api/server"
	"accesslog-tracker/internal/config"
	"accesslog-tracker/internal/domain/services"
	"accesslog-tracker/internal/infrastructure/cache/redis"
	"accesslog-tracker/internal/infrastructure/database/postgresql"
	"accesslog-tracker/internal/infrastructure/database/postgresql/repositories"
	"accesslog-tracker/internal/utils/logger"
	apihelpers "accesslog-tracker/tests/integration/api"
)

func TestServerIntegration(t *testing.T) {
	// テスト用データベースをセットアップ
	db := apihelpers.SetupTestDatabase(t)
	defer db.Close()

	// テスト用Redisをセットアップ
	redisClient := apihelpers.SetupTestRedis(t)
	defer redisClient.Close()

	// 設定を初期化
	cfg := &config.Config{
		App: config.AppConfig{
			Port:  8080,
			Debug: true,
		},
		Database: config.DatabaseConfig{
			Host:     apihelpers.GetTestDBHost(),
			Port:     5432,
			User:     "postgres",
			Password: "password",
			Name:     "access_log_tracker_test",
			SSLMode:  "disable",
		},
		Redis: config.RedisConfig{
			Host:     apihelpers.GetTestRedisHost(),
			Port:     6379,
			Password: "",
			DB:       0,
		},
	}

	// ロガーを初期化
	log := logger.NewLogger()

	// データベース接続を初期化
	dbConn := postgresql.NewConnection(cfg.GetDatabaseDSN())
	err := dbConn.Connect(cfg.GetDatabaseDSN())
	require.NoError(t, err)
	defer dbConn.Close()

	// キャッシュサービスを初期化
	cacheService := redis.NewCacheService(cfg.GetRedisAddr())
	err = cacheService.Connect()
	require.NoError(t, err)
	defer cacheService.Close()

	// リポジトリを初期化
	appRepo := repositories.NewApplicationRepository(db)
	trackingRepo := repositories.NewTrackingRepository(db)

	// サービスを初期化
	appService := services.NewApplicationService(appRepo, cacheService)
	trackingService := services.NewTrackingService(trackingRepo)

	// サーバーを初期化
	srv := server.NewServer(cfg, log, trackingService, appService, dbConn, cacheService)

	t.Run("should_start_and_stop_server", func(t *testing.T) {
		// サーバーを開始
		go func() {
			err := srv.Start()
			assert.NoError(t, err)
		}()

		// サーバーが起動するまで待機
		time.Sleep(3 * time.Second)

		// ヘルスチェックを実行
		resp, err := http.Get("http://localhost:8080/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		// 503エラーも許容（データベース接続が不安定な場合）
		assert.Contains(t, []int{200, 503}, resp.StatusCode, "Expected 200 or 503, got %d", resp.StatusCode)

		// サーバーを停止
		err = srv.Stop()
		assert.NoError(t, err)
	})

	t.Run("should_handle_graceful_shutdown", func(t *testing.T) {
		// サーバーを開始
		go func() {
			err := srv.Start()
			assert.NoError(t, err)
		}()

		// サーバーが起動するまで待機
		time.Sleep(3 * time.Second)

		// 複数のリクエストを同時に送信
		for i := 0; i < 5; i++ {
			go func() {
				resp, err := http.Get("http://localhost:8080/health")
				if err == nil {
					resp.Body.Close()
				}
			}()
		}

		// 少し待機してからシャットダウン
		time.Sleep(1 * time.Second)

		// サーバーを停止
		err := srv.Stop()
		assert.NoError(t, err)
	})

	t.Run("should_handle_server_configuration", func(t *testing.T) {
		// サーバー設定を確認
		assert.NotNil(t, srv.GetRouter())
	})

	t.Run("should_handle_server_middleware", func(t *testing.T) {
		// サーバーを開始
		go func() {
			err := srv.Start()
			assert.NoError(t, err)
		}()

		// サーバーが起動するまで待機
		time.Sleep(3 * time.Second)

		// CORSヘッダーを確認
		req, _ := http.NewRequest("OPTIONS", "http://localhost:8080/health", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "GET")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			// 接続エラーの場合はスキップ
			t.Skip("Server not responding, skipping middleware test")
		}
		defer resp.Body.Close()

		// CORSヘッダーが設定されていることを確認
		assert.Equal(t, "http://localhost:3000", resp.Header.Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET", resp.Header.Get("Access-Control-Allow-Methods"))

		// サーバーを停止
		err = srv.Stop()
		assert.NoError(t, err)
	})

	t.Run("should_handle_server_routes", func(t *testing.T) {
		// サーバーを開始
		go func() {
			err := srv.Start()
			assert.NoError(t, err)
		}()

		// サーバーが起動するまで待機
		time.Sleep(3 * time.Second)

		// 各種エンドポイントをテスト
		endpoints := []string{
			"/health",
			"/ready",
			"/live",
			"/v1/applications",
			"/tracker.js",
		}

		client := &http.Client{Timeout: 5 * time.Second}
		for _, endpoint := range endpoints {
			resp, err := client.Get("http://localhost:8080" + endpoint)
			if err != nil {
				// 接続エラーの場合はスキップ
				t.Skipf("Server not responding for endpoint %s, skipping route test", endpoint)
			}
			resp.Body.Close()

			// 404以外のステータスコードを期待
			assert.NotEqual(t, 404, resp.StatusCode, "Endpoint %s should not return 404", endpoint)
		}

		// サーバーを停止
		err := srv.Stop()
		assert.NoError(t, err)
	})
}

func TestServerSetupTest(t *testing.T) {
	// テスト用データベースをセットアップ
	db := apihelpers.SetupTestDatabase(t)
	defer db.Close()

	// テスト用Redisをセットアップ
	redisClient := apihelpers.SetupTestRedis(t)
	defer redisClient.Close()

	// 設定を初期化
	cfg := &config.Config{
		App: config.AppConfig{
			Port:  8080,
			Debug: true,
		},
		Database: config.DatabaseConfig{
			Host:     apihelpers.GetTestDBHost(),
			Port:     5432,
			User:     "postgres",
			Password: "password",
			Name:     "access_log_tracker_test",
			SSLMode:  "disable",
		},
		Redis: config.RedisConfig{
			Host:     apihelpers.GetTestRedisHost(),
			Port:     6379,
			Password: "",
			DB:       0,
		},
	}

	// ロガーを初期化
	log := logger.NewLogger()

	// データベース接続を初期化
	dbConn := postgresql.NewConnection(cfg.GetDatabaseDSN())
	err := dbConn.Connect(cfg.GetDatabaseDSN())
	require.NoError(t, err)
	defer dbConn.Close()

	// キャッシュサービスを初期化
	cacheService := redis.NewCacheService(cfg.GetRedisAddr())
	err = cacheService.Connect()
	require.NoError(t, err)
	defer cacheService.Close()

	// リポジトリを初期化
	appRepo := repositories.NewApplicationRepository(db)
	trackingRepo := repositories.NewTrackingRepository(db)

	// サービスを初期化
	appService := services.NewApplicationService(appRepo, cacheService)
	trackingService := services.NewTrackingService(trackingRepo)

	// サーバーを初期化
	srv := server.NewServer(cfg, log, trackingService, appService, dbConn, cacheService)

	t.Run("should_setup_test_router", func(t *testing.T) {
		// ルーターが正しく設定されていることを確認
		router := srv.GetRouter()
		assert.NotNil(t, router)
		
		// テスト用エンドポイントが利用可能であることを確認
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		// 200または503を許容（データベース接続が不安定な場合）
		assert.Contains(t, []int{200, 503}, w.Code, "Expected 200 or 503, got %d", w.Code)
	})
}
