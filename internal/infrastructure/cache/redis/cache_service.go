package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"github.com/redis/go-redis/v9"
)

// CacheService Redisキャッシュサービス
type CacheService struct {
	client *redis.Client
	addr   string
}

// NewCacheService 新しいRedisキャッシュサービスを作成
func NewCacheService(addr string) *CacheService {
	return &CacheService{
		addr: addr,
	}
}

// Connect Redisに接続
func (c *CacheService) Connect() error {
	c.client = redis.NewClient(&redis.Options{
		Addr:     c.addr,
		Password: "", // 必要に応じて設定
		DB:       0,  // デフォルトDB
	})

	// 接続テスト
	ctx := context.Background()
	_, err := c.client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return nil
}

// Close Redis接続を閉じる
func (c *CacheService) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// GetClient Redisクライアントを取得
func (c *CacheService) GetClient() *redis.Client {
	return c.client
}

// Ping Redis接続をテスト
func (c *CacheService) Ping(ctx context.Context) error {
	if c.client == nil {
		return fmt.Errorf("Redis client not connected")
	}

	_, err := c.client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to ping Redis: %w", err)
	}

	return nil
}

// Set キーと値を設定
func (c *CacheService) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	if c.client == nil {
		return fmt.Errorf("Redis client not connected")
	}

	err := c.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	return nil
}

// Get キーの値を取得
func (c *CacheService) Get(ctx context.Context, key string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("Redis client not connected")
	}

	value, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key not found: %s", key)
		}
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}

	return value, nil
}

// SetJSON JSON値を設定
func (c *CacheService) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if c.client == nil {
		return fmt.Errorf("Redis client not connected")
	}

	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	err = c.client.Set(ctx, key, jsonData, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set JSON key %s: %w", key, err)
	}

	return nil
}

// GetJSON JSON値を取得
func (c *CacheService) GetJSON(ctx context.Context, key string, dest interface{}) error {
	if c.client == nil {
		return fmt.Errorf("Redis client not connected")
	}

	jsonData, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found: %s", key)
		}
		return fmt.Errorf("failed to get key %s: %w", key, err)
	}

	err = json.Unmarshal([]byte(jsonData), dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

// Delete キーを削除
func (c *CacheService) Delete(ctx context.Context, key string) error {
	if c.client == nil {
		return fmt.Errorf("Redis client not connected")
	}

	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}

	return nil
}

// Exists キーが存在するかチェック
func (c *CacheService) Exists(ctx context.Context, key string) (bool, error) {
	if c.client == nil {
		return false, fmt.Errorf("Redis client not connected")
	}

	result, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence %s: %w", key, err)
	}

	return result > 0, nil
}

// Incr カウンターをインクリメント
func (c *CacheService) Incr(ctx context.Context, key string) (int64, error) {
	if c.client == nil {
		return 0, fmt.Errorf("Redis client not connected")
	}

	result, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s: %w", key, err)
	}

	return result, nil
}

// Decr カウンターをデクリメント
func (c *CacheService) Decr(ctx context.Context, key string) (int64, error) {
	if c.client == nil {
		return 0, fmt.Errorf("Redis client not connected")
	}

	result, err := c.client.Decr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to decrement key %s: %w", key, err)
	}

	return result, nil
}

// MGet 複数のキーの値を取得
func (c *CacheService) MGet(ctx context.Context, keys ...string) ([]string, error) {
	if c.client == nil {
		return nil, fmt.Errorf("Redis client not connected")
	}

	results, err := c.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get multiple keys: %w", err)
	}

	values := make([]string, len(results))
	for i, result := range results {
		if result == nil {
			values[i] = ""
		} else {
			values[i] = result.(string)
		}
	}

	return values, nil
}

// MDelete 複数のキーを削除
func (c *CacheService) MDelete(ctx context.Context, keys ...string) error {
	if c.client == nil {
		return fmt.Errorf("Redis client not connected")
	}

	err := c.client.Del(ctx, keys...).Err()
	if err != nil {
		return fmt.Errorf("failed to delete multiple keys: %w", err)
	}

	return nil
}

// HSet ハッシュフィールドを設定
func (c *CacheService) HSet(ctx context.Context, key, field, value string) error {
	if c.client == nil {
		return fmt.Errorf("Redis client not connected")
	}

	err := c.client.HSet(ctx, key, field, value).Err()
	if err != nil {
		return fmt.Errorf("failed to set hash field %s:%s: %w", key, field, err)
	}

	return nil
}

// HGet ハッシュフィールドを取得
func (c *CacheService) HGet(ctx context.Context, key, field string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("Redis client not connected")
	}

	value, err := c.client.HGet(ctx, key, field).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("hash field not found: %s:%s", key, field)
		}
		return "", fmt.Errorf("failed to get hash field %s:%s: %w", key, field, err)
	}

	return value, nil
}

// HGetAll ハッシュ全体を取得
func (c *CacheService) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	if c.client == nil {
		return nil, fmt.Errorf("Redis client not connected")
	}

	hash, err := c.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get hash %s: %w", key, err)
	}

	return hash, nil
}

// HDel ハッシュフィールドを削除
func (c *CacheService) HDel(ctx context.Context, key, field string) error {
	if c.client == nil {
		return fmt.Errorf("Redis client not connected")
	}

	err := c.client.HDel(ctx, key, field).Err()
	if err != nil {
		return fmt.Errorf("failed to delete hash field %s:%s: %w", key, field, err)
	}

	return nil
}

// TTL キーのTTLを取得
func (c *CacheService) TTL(ctx context.Context, key string) (time.Duration, error) {
	if c.client == nil {
		return 0, fmt.Errorf("Redis client not connected")
	}

	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL for key %s: %w", key, err)
	}

	return ttl, nil
}

// Expire キーのTTLを設定
func (c *CacheService) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if c.client == nil {
		return fmt.Errorf("Redis client not connected")
	}

	err := c.client.Expire(ctx, key, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set TTL for key %s: %w", key, err)
	}

	return nil
}
