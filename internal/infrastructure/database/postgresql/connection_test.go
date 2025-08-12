package postgresql

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
		Host:            "localhost",
		Port:            5432,
		User:            "test_user",
		Password:        "test_password",
		Database:        "test_db",
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 1 * time.Hour,
	}

	logger := logrus.New()

	// 実際のデータベースがない場合のテスト
	// このテストは実際のPostgreSQLインスタンスが必要
	t.Skip("Skipping test that requires actual PostgreSQL instance")

	conn, err := NewConnection(config, logger)
	require.NoError(t, err)
	defer conn.Close()

	assert.NotNil(t, conn)
	assert.NotNil(t, conn.GetDB())
}

func TestConnection_Ping(t *testing.T) {
	// 実際のデータベースがない場合のテスト
	t.Skip("Skipping test that requires actual PostgreSQL instance")

	config := &Config{
		Host:     "localhost",
		Port:     5432,
		User:     "test_user",
		Password: "test_password",
		Database: "test_db",
		SSLMode:  "disable",
	}

	logger := logrus.New()
	conn, err := NewConnection(config, logger)
	require.NoError(t, err)
	defer conn.Close()

	err = conn.Ping()
	assert.NoError(t, err)
}

func TestConnection_Stats(t *testing.T) {
	// 実際のデータベースがない場合のテスト
	t.Skip("Skipping test that requires actual PostgreSQL instance")

	config := &Config{
		Host:     "localhost",
		Port:     5432,
		User:     "test_user",
		Password: "test_password",
		Database: "test_db",
		SSLMode:  "disable",
	}

	logger := logrus.New()
	conn, err := NewConnection(config, logger)
	require.NoError(t, err)
	defer conn.Close()

	stats := conn.Stats()
	assert.NotNil(t, stats)
}
