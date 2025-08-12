package redis

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConnection(t *testing.T) {
	// テスト用の設定
	config := &Config{
		Host:         "localhost",
		Port:         6379,
		Password:     "",
		Database:     0,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		IdleTimeout:  5 * time.Minute,
	}

	logger := logrus.New()

	// 実際のRedisがない場合のテスト
	// このテストは実際のRedisインスタンスが必要
	t.Skip("Skipping test that requires actual Redis instance")

	conn, err := NewConnection(config, logger)
	require.NoError(t, err)
	defer conn.Close()

	assert.NotNil(t, conn)
	assert.NotNil(t, conn.GetClient())
}

func TestConnection_Ping(t *testing.T) {
	// 実際のRedisがない場合のテスト
	t.Skip("Skipping test that requires actual Redis instance")

	config := &Config{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		Database: 0,
	}

	logger := logrus.New()
	conn, err := NewConnection(config, logger)
	require.NoError(t, err)
	defer conn.Close()

	err = conn.Ping()
	assert.NoError(t, err)
}

func TestConnection_GetStats(t *testing.T) {
	// 実際のRedisがない場合のテスト
	t.Skip("Skipping test that requires actual Redis instance")

	config := &Config{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		Database: 0,
	}

	logger := logrus.New()
	conn, err := NewConnection(config, logger)
	require.NoError(t, err)
	defer conn.Close()

	stats := conn.GetStats()
	assert.NotNil(t, stats)
}
