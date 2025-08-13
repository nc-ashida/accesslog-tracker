package redis_test

import (
	"context"
	"fmt"
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"accesslog-tracker/internal/infrastructure/cache/redis"
	"accesslog-tracker/tests/integration/infrastructure"
)

func setupTestRedis() (*redis.CacheService, func(), error) {
	// 環境変数から接続情報を取得
	host := infrastructure.GetEnvOrDefault("REDIS_HOST", "localhost")
	port := infrastructure.GetEnvOrDefault("REDIS_PORT", "16380")
	
	// テスト用Redis接続
	addr := fmt.Sprintf("%s:%s", host, port)
	cache := redis.NewCacheService(addr)
	
	err := cache.Connect()
	if err != nil {
		return nil, nil, err
	}

	// クリーンアップ関数
	cleanup := func() {
		cache.Close()
	}

	return cache, cleanup, nil
}

func TestCacheService_Integration(t *testing.T) {
	cache, cleanup, err := setupTestRedis()
	require.NoError(t, err)
	defer cleanup()

	ctx := context.Background()

	t.Run("should set and get string value", func(t *testing.T) {
		key := "test_string_key"
		value := "test_string_value"
		ttl := time.Minute

		// 値を設定
		err := cache.Set(ctx, key, value, ttl)
		assert.NoError(t, err)

		// 値を取得
		result, err := cache.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value, result)
	})

	t.Run("should set and get JSON value", func(t *testing.T) {
		key := "test_json_key"
		value := map[string]interface{}{
			"name":  "test",
			"count": 123,
			"active": true,
		}
		ttl := time.Minute

		// JSON値を設定
		err := cache.SetJSON(ctx, key, value, ttl)
		assert.NoError(t, err)

		// JSON値を取得
		var result map[string]interface{}
		err = cache.GetJSON(ctx, key, &result)
		assert.NoError(t, err)
		assert.Equal(t, value["name"], result["name"])
		assert.Equal(t, float64(value["count"].(int)), result["count"])
		assert.Equal(t, value["active"], result["active"])
	})

	t.Run("should handle expiration", func(t *testing.T) {
		key := "test_expire_key"
		value := "test_expire_value"
		ttl := time.Millisecond * 100

		// 値を設定（短いTTL）
		err := cache.Set(ctx, key, value, ttl)
		assert.NoError(t, err)

		// 即座に取得
		result, err := cache.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value, result)

		// TTL後に取得（期限切れ）
		time.Sleep(time.Millisecond * 150)
		result, err = cache.Get(ctx, key)
		assert.Error(t, err)
		assert.Empty(t, result)
	})

	t.Run("should delete key", func(t *testing.T) {
		key := "test_delete_key"
		value := "test_delete_value"
		ttl := time.Minute

		// 値を設定
		err := cache.Set(ctx, key, value, ttl)
		assert.NoError(t, err)

		// 削除前の確認
		result, err := cache.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value, result)

		// キーを削除
		err = cache.Delete(ctx, key)
		assert.NoError(t, err)

		// 削除後の確認
		result, err = cache.Get(ctx, key)
		assert.Error(t, err)
		assert.Empty(t, result)
	})

	t.Run("should check if key exists", func(t *testing.T) {
		key := "test_exists_key"
		value := "test_exists_value"
		ttl := time.Minute

		// 存在しないキーの確認
		exists, err := cache.Exists(ctx, key)
		assert.NoError(t, err)
		assert.False(t, exists)

		// 値を設定
		err = cache.Set(ctx, key, value, ttl)
		assert.NoError(t, err)

		// 存在するキーの確認
		exists, err = cache.Exists(ctx, key)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("should increment and decrement counter", func(t *testing.T) {
		key := "test_counter_key"
		ttl := time.Minute

		// カウンターを初期化
		err := cache.Set(ctx, key, "0", ttl)
		assert.NoError(t, err)

		// インクリメント
		value, err := cache.Incr(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), value)

		value, err = cache.Incr(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), value)

		// デクリメント
		value, err = cache.Decr(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), value)
	})

	t.Run("should handle multiple keys", func(t *testing.T) {
		keys := []string{"test_multi_1", "test_multi_2", "test_multi_3"}
		values := []string{"value1", "value2", "value3"}
		ttl := time.Minute

		// 複数のキーを設定
		for i, key := range keys {
			err := cache.Set(ctx, key, values[i], ttl)
			assert.NoError(t, err)
		}

		// 複数のキーを取得
		results, err := cache.MGet(ctx, keys...)
		assert.NoError(t, err)
		assert.Len(t, results, 3)
		assert.Equal(t, values[0], results[0])
		assert.Equal(t, values[1], results[1])
		assert.Equal(t, values[2], results[2])

		// 複数のキーを削除
		err = cache.MDelete(ctx, keys...)
		assert.NoError(t, err)

		// 削除後の確認
		for _, key := range keys {
			result, err := cache.Get(ctx, key)
			assert.Error(t, err)
			assert.Empty(t, result)
		}
	})

	t.Run("should handle hash operations", func(t *testing.T) {
		key := "test_hash_key"
		field1 := "field1"
		field2 := "field2"
		value1 := "value1"
		value2 := "value2"

		// ハッシュフィールドを設定
		err := cache.HSet(ctx, key, field1, value1)
		assert.NoError(t, err)
		err = cache.HSet(ctx, key, field2, value2)
		assert.NoError(t, err)

		// ハッシュフィールドを取得
		result, err := cache.HGet(ctx, key, field1)
		assert.NoError(t, err)
		assert.Equal(t, value1, result)

		// ハッシュ全体を取得
		hash, err := cache.HGetAll(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value1, hash[field1])
		assert.Equal(t, value2, hash[field2])

		// ハッシュフィールドを削除
		err = cache.HDel(ctx, key, field1)
		assert.NoError(t, err)

		// 削除後の確認
		result, err = cache.HGet(ctx, key, field1)
		assert.Error(t, err)
		assert.Empty(t, result)
	})
}
