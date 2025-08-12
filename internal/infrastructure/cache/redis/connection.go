package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ConnectionConfig はRedis接続の設定
type ConnectionConfig struct {
	Address  string
	Password string
	DB       int
	PoolSize int
}

// Connection はRedis接続を管理
type Connection struct {
	client *redis.Client
	config ConnectionConfig
}

// NewConnection は新しいRedis接続を作成
func NewConnection(config ConnectionConfig) (*Connection, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Address,
		Password: config.Password,
		DB:       config.DB,
		PoolSize: config.PoolSize,
	})

	// 接続テスト
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Connection{
		client: rdb,
		config: config,
	}, nil
}

// Ping はRedis接続をテスト
func (c *Connection) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	_, err := c.client.Ping(ctx).Result()
	return err
}

// Close はRedis接続を閉じる
func (c *Connection) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// GetClient はRedisクライアントインスタンスを取得
func (c *Connection) GetClient() *redis.Client {
	return c.client
}

// GetStats は接続プールの統計情報を取得
func (c *Connection) GetStats() *redis.PoolStats {
	return c.client.PoolStats()
}
