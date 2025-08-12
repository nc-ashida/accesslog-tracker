package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// Config はPostgreSQL接続設定を表す
type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// Connection はPostgreSQL接続を管理
type Connection struct {
	db     *sql.DB
	config *Config
	logger *logrus.Logger
}

// NewConnection は新しいPostgreSQL接続を作成
func NewConnection(config *Config, logger *logrus.Logger) (*Connection, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Database, config.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// 接続プール設定
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	conn := &Connection{
		db:     db,
		config: config,
		logger: logger,
	}

	// 接続テスト
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("PostgreSQL connection established successfully")
	return conn, nil
}

// Ping はデータベース接続をテスト
func (c *Connection) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.db.PingContext(ctx)
}

// Close はデータベース接続を閉じる
func (c *Connection) Close() error {
	if c.db != nil {
		c.logger.Info("Closing PostgreSQL connection")
		return c.db.Close()
	}
	return nil
}

// GetDB はデータベースインスタンスを取得
func (c *Connection) GetDB() *sql.DB {
	return c.db
}

// BeginTx はトランザクションを開始
func (c *Connection) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return c.db.BeginTx(ctx, opts)
}

// ExecContext はSQLを実行（結果を返さない）
func (c *Connection) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return c.db.ExecContext(ctx, query, args...)
}

// QueryContext はSQLクエリを実行
func (c *Connection) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return c.db.QueryContext(ctx, query, args...)
}

// QueryRowContext は単一行のSQLクエリを実行
func (c *Connection) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return c.db.QueryRowContext(ctx, query, args...)
}

// Stats は接続プールの統計情報を取得
func (c *Connection) Stats() sql.DBStats {
	return c.db.Stats()
}
