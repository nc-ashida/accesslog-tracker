package postgresql

import (
	"database/sql"
	"fmt"
	"time"
	_ "github.com/lib/pq"
)

// Connection PostgreSQL接続を管理する構造体
type Connection struct {
	name string
	db   *sql.DB
}

// NewConnection 新しいPostgreSQL接続を作成
func NewConnection(name string) *Connection {
	return &Connection{
		name: name,
	}
}

// Connect データベースに接続
func (c *Connection) Connect(dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// 接続設定
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	c.db = db
	return nil
}

// GetDB データベースインスタンスを取得
func (c *Connection) GetDB() *sql.DB {
	return c.db
}

// Ping データベース接続をテスト
func (c *Connection) Ping() error {
	if c.db == nil {
		return fmt.Errorf("database connection not established")
	}
	return c.db.Ping()
}

// Close データベース接続を閉じる
func (c *Connection) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// Begin トランザクションを開始
func (c *Connection) Begin() (*sql.Tx, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database connection not established")
	}
	return c.db.Begin()
}

// Exec クエリを実行
func (c *Connection) Exec(query string, args ...interface{}) (sql.Result, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database connection not established")
	}
	return c.db.Exec(query, args...)
}

// Query クエリを実行して結果を取得
func (c *Connection) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database connection not established")
	}
	return c.db.Query(query, args...)
}

// QueryRow 単一行のクエリを実行
func (c *Connection) QueryRow(query string, args ...interface{}) *sql.Row {
	if c.db == nil {
		return nil
	}
	return c.db.QueryRow(query, args...)
}
