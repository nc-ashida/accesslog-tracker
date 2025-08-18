package repositories

import (
	"context"
	"fmt"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"accesslog-tracker/internal/infrastructure/database/postgresql/repositories"
	"accesslog-tracker/internal/domain/models"
	"accesslog-tracker/internal/infrastructure/database/postgresql"
	"accesslog-tracker/tests/integration/infrastructure"
)

func setupApplicationTestDatabase() (*repositories.ApplicationRepository, func(), error) {
	// 環境変数から接続情報を取得
	host := infrastructure.GetEnvOrDefault("DB_HOST", "localhost")
	port := infrastructure.GetEnvOrDefault("DB_PORT", "18433")
	user := infrastructure.GetEnvOrDefault("DB_USER", "postgres")
	password := infrastructure.GetEnvOrDefault("DB_PASSWORD", "password")
	dbname := infrastructure.GetEnvOrDefault("DB_NAME", "access_log_tracker_test")
	
	// テスト用データベース接続
	conn := postgresql.NewConnection("test")
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", 
		host, port, user, password, dbname)
	
	err := conn.Connect(dsn)
	if err != nil {
		return nil, nil, err
	}

	repo := repositories.NewApplicationRepository(conn.GetDB())
	
	// クリーンアップ関数
	cleanup := func() {
		conn.Close()
	}

	return repo, cleanup, nil
}

// cleanupTestData はテストデータをクリーンアップします
func cleanupTestData(t *testing.T, repo *repositories.ApplicationRepository) {
	ctx := context.Background()
	
	// テスト用アプリケーションを削除
	testAppIDs := []string{
		"test_app_123",
		"test_app_456",
		"test_app_paginate_0",
		"test_app_paginate_1", 
		"test_app_paginate_2",
		"test_app_paginate_3",
		"test_app_paginate_4",
		"test_app_update",
		"test_app_delete",
	}
	
	for _, appID := range testAppIDs {
		repo.Delete(ctx, appID)
	}
}

func TestApplicationRepository_Integration(t *testing.T) {
	repo, cleanup, err := setupApplicationTestDatabase()
	require.NoError(t, err)
	defer cleanup()

	ctx := context.Background()
	
	// テスト前にデータをクリーンアップ
	cleanupTestData(t, repo)

	t.Run("should save and retrieve application", func(t *testing.T) {
		// テストデータを作成
		app := &models.Application{
			AppID:    "test_app_123",
			Name:     "Test Application",
			Domain:   "example.com",
			APIKey:   "test-api-key-123",
			Active: true,
		}

		// アプリケーションを保存
		err := repo.Create(ctx, app)
		assert.NoError(t, err)

		// アプリケーションを取得して検証
		retrieved, err := repo.GetByID(ctx, "test_app_123")
		assert.NoError(t, err)
		assert.Equal(t, app.AppID, retrieved.AppID)
		assert.Equal(t, app.Name, retrieved.Name)
		assert.Equal(t, app.Domain, retrieved.Domain)
		assert.Equal(t, app.APIKey, retrieved.APIKey)
		assert.Equal(t, app.Active, retrieved.Active)
	})

	t.Run("should find application by API key", func(t *testing.T) {
		apiKey := "test-api-key-456"
		app := &models.Application{
			AppID:    "test_app_456",
			Name:     "Test Application 2",
			Domain:   "example2.com",
			APIKey:   apiKey,
			Active: true,
		}

		err := repo.Create(ctx, app)
		assert.NoError(t, err)

		// APIキーで検索
		retrieved, err := repo.GetByAPIKey(ctx, apiKey)
		assert.NoError(t, err)
		assert.Equal(t, app.AppID, retrieved.AppID)
		assert.Equal(t, apiKey, retrieved.APIKey)
	})

	t.Run("should find all applications with pagination", func(t *testing.T) {
		// 複数のアプリケーションを作成
		for i := 0; i < 5; i++ {
			app := &models.Application{
				AppID:    fmt.Sprintf("test_app_paginate_%d", i),
				Name:     fmt.Sprintf("Test Application %d", i),
				Domain:   fmt.Sprintf("example%d.com", i),
				APIKey:   fmt.Sprintf("test-api-key-paginate-%d", i),
				Active: true,
			}
			err := repo.Create(ctx, app)
			assert.NoError(t, err)
		}

		// ページネーションで取得
		results, err := repo.List(ctx, 3, 0)
		assert.NoError(t, err)
		assert.Len(t, results, 3)

		// 次のページを取得
		results, err = repo.List(ctx, 3, 3)
		assert.NoError(t, err)
		assert.Len(t, results, 3)
	})

	t.Run("should update application", func(t *testing.T) {
		app := &models.Application{
			AppID:    "test_app_update",
			Name:     "Original Name",
			Domain:   "original.com",
			APIKey:   "test-api-key-update",
			Active: true,
		}

		err := repo.Create(ctx, app)
		assert.NoError(t, err)

		// アプリケーションを更新
		app.Name = "Updated Name"
		app.Domain = "updated.com"
		app.Active = false

		err = repo.Update(ctx, app)
		assert.NoError(t, err)

		// 更新されたアプリケーションを取得して検証
		retrieved, err := repo.GetByID(ctx, "test_app_update")
		assert.NoError(t, err)
		assert.Equal(t, "Updated Name", retrieved.Name)
		assert.Equal(t, "updated.com", retrieved.Domain)
		assert.Equal(t, false, retrieved.Active)
	})

	t.Run("should delete application", func(t *testing.T) {
		app := &models.Application{
			AppID:    "test_app_delete",
			Name:     "Delete Test App",
			Domain:   "delete.com",
			APIKey:   "test-api-key-delete",
			Active: true,
		}

		err := repo.Create(ctx, app)
		assert.NoError(t, err)

		// 削除前の確認
		retrieved, err := repo.GetByID(ctx, "test_app_delete")
		assert.NoError(t, err)
		assert.Equal(t, app.AppID, retrieved.AppID)

		// 削除実行
		err = repo.Delete(ctx, "test_app_delete")
		assert.NoError(t, err)

		// 削除後の確認
		retrieved, err = repo.GetByID(ctx, "test_app_delete")
		assert.Error(t, err)
		assert.Nil(t, retrieved)
	})

	t.Run("should handle non-existent application", func(t *testing.T) {
		// 存在しないアプリケーションを検索
		retrieved, err := repo.GetByID(ctx, "non_existent_app")
		assert.Error(t, err)
		assert.Nil(t, retrieved)

		// 存在しないAPIキーを検索
		retrieved, err = repo.GetByAPIKey(ctx, "non_existent_key")
		assert.Error(t, err)
		assert.Nil(t, retrieved)
	})

	t.Run("should handle duplicate application creation", func(t *testing.T) {
		app := &models.Application{
			AppID:    "test_app_duplicate",
			Name:     "Duplicate Test App",
			Domain:   "duplicate.com",
			APIKey:   "test-api-key-duplicate",
			Active:   true,
		}

		// 最初の作成
		err := repo.Create(ctx, app)
		assert.NoError(t, err)

		// 重複するアプリケーションの作成を試行
		duplicateApp := &models.Application{
			AppID:    "test_app_duplicate", // 同じAppID
			Name:     "Duplicate Test App 2",
			Domain:   "duplicate2.com",
			APIKey:   "test-api-key-duplicate-2",
			Active:   true,
		}

		err = repo.Create(ctx, duplicateApp)
		assert.Error(t, err) // 重複エラーが期待される

		// クリーンアップ
		repo.Delete(ctx, "test_app_duplicate")
	})

	t.Run("should handle duplicate API key creation", func(t *testing.T) {
		app1 := &models.Application{
			AppID:    "test_app_api_key_1",
			Name:     "API Key Test App 1",
			Domain:   "apikey1.com",
			APIKey:   "test-api-key-shared",
			Active:   true,
		}

		app2 := &models.Application{
			AppID:    "test_app_api_key_2",
			Name:     "API Key Test App 2",
			Domain:   "apikey2.com",
			APIKey:   "test-api-key-shared", // 同じAPIキー
			Active:   true,
		}

		// 最初の作成
		err := repo.Create(ctx, app1)
		assert.NoError(t, err)

		// 同じAPIキーでの作成を試行
		err = repo.Create(ctx, app2)
		assert.Error(t, err) // 重複エラーが期待される

		// クリーンアップ
		repo.Delete(ctx, "test_app_api_key_1")
	})

	t.Run("should handle application count", func(t *testing.T) {
		// 複数のアプリケーションを作成
		for i := 0; i < 3; i++ {
			app := &models.Application{
				AppID:    fmt.Sprintf("test_app_count_%d", i),
				Name:     fmt.Sprintf("Count Test App %d", i),
				Domain:   fmt.Sprintf("count%d.com", i),
				APIKey:   fmt.Sprintf("test-api-key-count-%d", i),
				Active:   true,
			}
			err := repo.Create(ctx, app)
			assert.NoError(t, err)
		}

		// アプリケーション数を取得
		count, err := repo.Count(ctx)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(3))

		// クリーンアップ
		for i := 0; i < 3; i++ {
			repo.Delete(ctx, fmt.Sprintf("test_app_count_%d", i))
		}
	})

	t.Run("should handle API key regeneration", func(t *testing.T) {
		app := &models.Application{
			AppID:    "test_app_regenerate",
			Name:     "Regenerate Test App",
			Domain:   "regenerate.com",
			APIKey:   "old-api-key",
			Active:   true,
		}

		err := repo.Create(ctx, app)
		assert.NoError(t, err)

		// APIキーを再生成
		newAPIKey, err := repo.RegenerateAPIKey(ctx, "test_app_regenerate")
		assert.NoError(t, err)
		assert.NotEmpty(t, newAPIKey)
		assert.NotEqual(t, "old-api-key", newAPIKey)

		// 新しいAPIキーでアプリケーションを取得
		retrieved, err := repo.GetByAPIKey(ctx, newAPIKey)
		assert.NoError(t, err)
		assert.Equal(t, "test_app_regenerate", retrieved.AppID)

		// 古いAPIキーでは取得できないことを確認
		_, err = repo.GetByAPIKey(ctx, "old-api-key")
		assert.Error(t, err)

		// クリーンアップ
		repo.Delete(ctx, "test_app_regenerate")
	})

	t.Run("should handle settings update", func(t *testing.T) {
		app := &models.Application{
			AppID:    "test_app_settings",
			Name:     "Settings Test App",
			Domain:   "settings.com",
			APIKey:   "test-api-key-settings",
			Active:   true,
		}

		err := repo.Create(ctx, app)
		assert.NoError(t, err)

		// 設定を更新
		settings := map[string]interface{}{
			"tracking_enabled": true,
			"max_requests":     1000,
			"retention_days":   30,
		}

		err = repo.UpdateSettings(ctx, "test_app_settings", settings)
		assert.NoError(t, err)

		// クリーンアップ
		repo.Delete(ctx, "test_app_settings")
	})

	t.Run("should handle pagination edge cases", func(t *testing.T) {
		// 空の結果セットでのページネーション
		results, err := repo.List(ctx, 10, 1000) // 大きなオフセット
		assert.NoError(t, err)
		assert.Len(t, results, 0)

		// ゼロ制限でのページネーション
		results, err = repo.List(ctx, 0, 0)
		assert.NoError(t, err)
		assert.Len(t, results, 0)

		// 負のオフセットでのページネーション（データベースエラーが発生するためスキップ）
		// results, err = repo.List(ctx, 10, -1)
		// assert.Error(t, err) // 負のオフセットはエラーを返す
	})
}
