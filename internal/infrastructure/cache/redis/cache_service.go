package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/your-username/accesslog-tracker/internal/infrastructure/cache/interfaces"
)

// CacheService はRedisキャッシュサービスの実装
type CacheService struct {
	conn   *Connection
	logger *logrus.Logger
}

// NewCacheService は新しいキャッシュサービスを作成
func NewCacheService(conn *Connection, logger *logrus.Logger) interfaces.CacheService {
	return &CacheService{
		conn:   conn,
		logger: logger,
	}
}

// Get はキーに対応する値を取得
func (c *CacheService) Get(ctx context.Context, key string) (string, error) {
	value, err := c.conn.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key not found: %s", key)
		}
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}
	return value, nil
}

// Set はキー・バリューペアを設定
func (c *CacheService) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	err := c.conn.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}
	return nil
}

// Delete はキーを削除
func (c *CacheService) Delete(ctx context.Context, key string) error {
	err := c.conn.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}
	return nil
}

// Exists はキーが存在するかチェック
func (c *CacheService) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.conn.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence of key %s: %w", key, err)
	}
	return result > 0, nil
}

// HGet はハッシュフィールドの値を取得
func (c *CacheService) HGet(ctx context.Context, key, field string) (string, error) {
	value, err := c.conn.client.HGet(ctx, key, field).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("field %s not found in hash %s", field, key)
		}
		return "", fmt.Errorf("failed to get field %s from hash %s: %w", field, key, err)
	}
	return value, nil
}

// HSet はハッシュフィールドの値を設定
func (c *CacheService) HSet(ctx context.Context, key, field, value string) error {
	err := c.conn.client.HSet(ctx, key, field, value).Err()
	if err != nil {
		return fmt.Errorf("failed to set field %s in hash %s: %w", field, key, err)
	}
	return nil
}

// HGetAll はハッシュの全フィールドを取得
func (c *CacheService) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	result, err := c.conn.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get all fields from hash %s: %w", key, err)
	}
	return result, nil
}

// HDel はハッシュフィールドを削除
func (c *CacheService) HDel(ctx context.Context, key string, fields ...string) error {
	err := c.conn.client.HDel(ctx, key, fields...).Err()
	if err != nil {
		return fmt.Errorf("failed to delete fields from hash %s: %w", key, err)
	}
	return nil
}

// LPush はリストの左端に値を追加
func (c *CacheService) LPush(ctx context.Context, key string, values ...string) error {
	err := c.conn.client.LPush(ctx, key, values).Err()
	if err != nil {
		return fmt.Errorf("failed to push values to list %s: %w", key, err)
	}
	return nil
}

// RPush はリストの右端に値を追加
func (c *CacheService) RPush(ctx context.Context, key string, values ...string) error {
	err := c.conn.client.RPush(ctx, key, values).Err()
	if err != nil {
		return fmt.Errorf("failed to push values to list %s: %w", key, err)
	}
	return nil
}

// LPop はリストの左端から値を取得
func (c *CacheService) LPop(ctx context.Context, key string) (string, error) {
	value, err := c.conn.client.LPop(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("list %s is empty", key)
		}
		return "", fmt.Errorf("failed to pop from list %s: %w", key, err)
	}
	return value, nil
}

// RPop はリストの右端から値を取得
func (c *CacheService) RPop(ctx context.Context, key string) (string, error) {
	value, err := c.conn.client.RPop(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("list %s is empty", key)
		}
		return "", fmt.Errorf("failed to pop from list %s: %w", key, err)
	}
	return value, nil
}

// LRange はリストの範囲を取得
func (c *CacheService) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	result, err := c.conn.client.LRange(ctx, key, start, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get range from list %s: %w", key, err)
	}
	return result, nil
}

// SAdd はセットにメンバーを追加
func (c *CacheService) SAdd(ctx context.Context, key string, members ...string) error {
	err := c.conn.client.SAdd(ctx, key, members).Err()
	if err != nil {
		return fmt.Errorf("failed to add members to set %s: %w", key, err)
	}
	return nil
}

// SRem はセットからメンバーを削除
func (c *CacheService) SRem(ctx context.Context, key string, members ...string) error {
	err := c.conn.client.SRem(ctx, key, members).Err()
	if err != nil {
		return fmt.Errorf("failed to remove members from set %s: %w", key, err)
	}
	return nil
}

// SMembers はセットの全メンバーを取得
func (c *CacheService) SMembers(ctx context.Context, key string) ([]string, error) {
	result, err := c.conn.client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get members from set %s: %w", key, err)
	}
	return result, nil
}

// SIsMember はメンバーがセットに含まれているかチェック
func (c *CacheService) SIsMember(ctx context.Context, key, member string) (bool, error) {
	result, err := c.conn.client.SIsMember(ctx, key, member).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check membership in set %s: %w", key, err)
	}
	return result, nil
}

// ZAdd はソート済みセットにメンバーを追加
func (c *CacheService) ZAdd(ctx context.Context, key string, score float64, member string) error {
	err := c.conn.client.ZAdd(ctx, key, redis.Z{Score: score, Member: member}).Err()
	if err != nil {
		return fmt.Errorf("failed to add member to sorted set %s: %w", key, err)
	}
	return nil
}

// ZRange はソート済みセットの範囲を取得
func (c *CacheService) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	result, err := c.conn.client.ZRange(ctx, key, start, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get range from sorted set %s: %w", key, err)
	}
	return result, nil
}

// ZRangeWithScores はソート済みセットの範囲をスコア付きで取得
func (c *CacheService) ZRangeWithScores(ctx context.Context, key string, start, stop int64) (map[string]float64, error) {
	result, err := c.conn.client.ZRangeWithScores(ctx, key, start, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get range with scores from sorted set %s: %w", key, err)
	}
	
	scores := make(map[string]float64)
	for _, z := range result {
		if member, ok := z.Member.(string); ok {
			scores[member] = z.Score
		}
	}
	return scores, nil
}

// ZRem はソート済みセットからメンバーを削除
func (c *CacheService) ZRem(ctx context.Context, key string, members ...string) error {
	err := c.conn.client.ZRem(ctx, key, members).Err()
	if err != nil {
		return fmt.Errorf("failed to remove members from sorted set %s: %w", key, err)
	}
	return nil
}

// Incr はカウンターを1増加
func (c *CacheService) Incr(ctx context.Context, key string) (int64, error) {
	result, err := c.conn.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s: %w", key, err)
	}
	return result, nil
}

// IncrBy はカウンターを指定値増加
func (c *CacheService) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	result, err := c.conn.client.IncrBy(ctx, key, value).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s by %d: %w", key, value, err)
	}
	return result, nil
}

// Decr はカウンターを1減少
func (c *CacheService) Decr(ctx context.Context, key string) (int64, error) {
	result, err := c.conn.client.Decr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to decrement key %s: %w", key, err)
	}
	return result, nil
}

// DecrBy はカウンターを指定値減少
func (c *CacheService) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	result, err := c.conn.client.DecrBy(ctx, key, value).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to decrement key %s by %d: %w", key, value, err)
	}
	return result, nil
}

// Expire はキーの有効期限を設定
func (c *CacheService) Expire(ctx context.Context, key string, expiration time.Duration) error {
	err := c.conn.client.Expire(ctx, key, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set expiration for key %s: %w", key, err)
	}
	return nil
}

// TTL はキーの残り有効期限を取得
func (c *CacheService) TTL(ctx context.Context, key string) (time.Duration, error) {
	result, err := c.conn.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL for key %s: %w", key, err)
	}
	return result, nil
}

// Keys はパターンにマッチするキーを取得
func (c *CacheService) Keys(ctx context.Context, pattern string) ([]string, error) {
	result, err := c.conn.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys matching pattern %s: %w", pattern, err)
	}
	return result, nil
}

// Pipeline はパイプライン操作を開始
func (c *CacheService) Pipeline() interfaces.Pipeline {
	return &Pipeline{
		pipe: c.conn.client.Pipeline(),
	}
}

// Ping はRedis接続をテスト
func (c *CacheService) Ping(ctx context.Context) error {
	return c.conn.Ping()
}

// Close はRedis接続を閉じる
func (c *CacheService) Close() error {
	return c.conn.Close()
}

// Pipeline はRedisパイプライン操作の実装
type Pipeline struct {
	pipe redis.Pipeliner
}

// Get はパイプラインでGet操作を追加
func (p *Pipeline) Get(key string) *interfaces.Result {
	cmd := p.pipe.Get(context.Background(), key)
	return &interfaces.Result{Value: cmd}
}

// Set はパイプラインでSet操作を追加
func (p *Pipeline) Set(key, value string, expiration time.Duration) *interfaces.Result {
	cmd := p.pipe.Set(context.Background(), key, value, expiration)
	return &interfaces.Result{Value: cmd}
}

// Delete はパイプラインでDelete操作を追加
func (p *Pipeline) Delete(key string) *interfaces.Result {
	cmd := p.pipe.Del(context.Background(), key)
	return &interfaces.Result{Value: cmd}
}

// Incr はパイプラインでIncr操作を追加
func (p *Pipeline) Incr(key string) *interfaces.Result {
	cmd := p.pipe.Incr(context.Background(), key)
	return &interfaces.Result{Value: cmd}
}

// IncrBy はパイプラインでIncrBy操作を追加
func (p *Pipeline) IncrBy(key string, value int64) *interfaces.Result {
	cmd := p.pipe.IncrBy(context.Background(), key, value)
	return &interfaces.Result{Value: cmd}
}

// Expire はパイプラインでExpire操作を追加
func (p *Pipeline) Expire(key string, expiration time.Duration) *interfaces.Result {
	cmd := p.pipe.Expire(context.Background(), key, expiration)
	return &interfaces.Result{Value: cmd}
}

// Exec はパイプライン操作を実行
func (p *Pipeline) Exec(ctx context.Context) error {
	_, err := p.pipe.Exec(ctx)
	return err
}
