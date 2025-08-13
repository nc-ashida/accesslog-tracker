package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config はアプリケーション全体の設定を表します
type Config struct {
	App      AppConfig      `yaml:"app"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	CORS     CORSConfig     `yaml:"cors"`
	Logging  LoggingConfig  `yaml:"logging"`
}

// AppConfig はアプリケーション固有の設定を表します
type AppConfig struct {
	Name string `yaml:"name" env:"APP_NAME"`
	Port int    `yaml:"port" env:"APP_PORT"`
	Host string `yaml:"host" env:"APP_HOST"`
	Debug bool  `yaml:"debug" env:"APP_DEBUG"`
}

// DatabaseConfig はデータベース接続設定を表します
type DatabaseConfig struct {
	Host            string `yaml:"host" env:"DB_HOST"`
	Port            int    `yaml:"port" env:"DB_PORT"`
	Name            string `yaml:"name" env:"DB_NAME"`
	User            string `yaml:"user" env:"DB_USER"`
	Password        string `yaml:"password" env:"DB_PASSWORD"`
	SSLMode         string `yaml:"ssl_mode" env:"DB_SSL_MODE"`
	MaxOpenConns    int    `yaml:"max_open_conns" env:"DB_MAX_OPEN_CONNS"`
	MaxIdleConns    int    `yaml:"max_idle_conns" env:"DB_MAX_IDLE_CONNS"`
	ConnMaxLifetime string `yaml:"conn_max_lifetime" env:"DB_CONN_MAX_LIFETIME"`
}

// RedisConfig はRedis接続設定を表します
type RedisConfig struct {
	Host     string `yaml:"host" env:"REDIS_HOST"`
	Port     int    `yaml:"port" env:"REDIS_PORT"`
	Password string `yaml:"password" env:"REDIS_PASSWORD"`
	DB       int    `yaml:"db" env:"REDIS_DB"`
	PoolSize int    `yaml:"pool_size" env:"REDIS_POOL_SIZE"`
}

// JWTConfig はJWT設定を表します
type JWTConfig struct {
	Secret           string `yaml:"secret" env:"JWT_SECRET"`
	Expiration       string `yaml:"expiration" env:"JWT_EXPIRATION"`
	RefreshExpiration string `yaml:"refresh_expiration" env:"JWT_REFRESH_EXPIRATION"`
}

// CORSConfig はCORS設定を表します
type CORSConfig struct {
	AllowedOrigins   string `yaml:"allowed_origins" env:"CORS_ALLOWED_ORIGINS"`
	AllowedMethods   string `yaml:"allowed_methods" env:"CORS_ALLOWED_METHODS"`
	AllowedHeaders   string `yaml:"allowed_headers" env:"CORS_ALLOWED_HEADERS"`
	ExposedHeaders   string `yaml:"exposed_headers" env:"CORS_EXPOSED_HEADERS"`
	AllowCredentials bool   `yaml:"allow_credentials" env:"CORS_ALLOW_CREDENTIALS"`
	MaxAge           int    `yaml:"max_age" env:"CORS_MAX_AGE"`
}

// LoggingConfig はログ設定を表します
type LoggingConfig struct {
	Level  string `yaml:"level" env:"LOG_LEVEL"`
	Format string `yaml:"format" env:"LOG_FORMAT"`
	Output string `yaml:"output" env:"LOG_OUTPUT"`
}

// New は新しい設定インスタンスを作成します
func New() *Config {
	return &Config{
		App: AppConfig{
			Name:  "access-log-tracker",
			Port:  8080,
			Host:  "0.0.0.0",
			Debug: false,
		},
		Database: DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			Name:            "access_log_tracker",
			User:            "postgres",
			Password:        "password",
			SSLMode:         "disable",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: "300s",
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
			PoolSize: 10,
		},
		JWT: JWTConfig{
			Secret:           "your-super-secret-jwt-key-change-in-production",
			Expiration:       "24h",
			RefreshExpiration: "168h",
		},
		CORS: CORSConfig{
			AllowedOrigins:   "http://localhost:3000,http://localhost:8080",
			AllowedMethods:   "GET,POST,PUT,DELETE,OPTIONS",
			AllowedHeaders:   "Content-Type,Authorization,X-Requested-With",
			ExposedHeaders:   "Content-Length",
			AllowCredentials: true,
			MaxAge:           86400,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
	}
}

// Load はYAMLファイルから設定を読み込みます
func (c *Config) Load(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	
	err = yaml.Unmarshal(data, c)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}
	
	return c.Validate()
}

// LoadFromEnv は環境変数から設定を読み込みます
func (c *Config) LoadFromEnv() error {
	// App設定
	if val := os.Getenv("APP_NAME"); val != "" {
		c.App.Name = val
	}
	if val := os.Getenv("APP_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			c.App.Port = port
		}
	}
	if val := os.Getenv("APP_HOST"); val != "" {
		c.App.Host = val
	}
	if val := os.Getenv("APP_DEBUG"); val != "" {
		c.App.Debug = val == "true"
	}
	
	// Database設定
	if val := os.Getenv("DB_HOST"); val != "" {
		c.Database.Host = val
	}
	if val := os.Getenv("DB_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			c.Database.Port = port
		}
	}
	if val := os.Getenv("DB_NAME"); val != "" {
		c.Database.Name = val
	}
	if val := os.Getenv("DB_USER"); val != "" {
		c.Database.User = val
	}
	if val := os.Getenv("DB_PASSWORD"); val != "" {
		c.Database.Password = val
	}
	if val := os.Getenv("DB_SSL_MODE"); val != "" {
		c.Database.SSLMode = val
	}
	
	// Redis設定
	if val := os.Getenv("REDIS_HOST"); val != "" {
		c.Redis.Host = val
	}
	if val := os.Getenv("REDIS_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			c.Redis.Port = port
		}
	}
	if val := os.Getenv("REDIS_PASSWORD"); val != "" {
		c.Redis.Password = val
	}
	if val := os.Getenv("REDIS_DB"); val != "" {
		if db, err := strconv.Atoi(val); err == nil {
			c.Redis.DB = db
		}
	}
	
	return c.Validate()
}

// Validate は設定の妥当性を検証します
func (c *Config) Validate() error {
	// App設定の検証
	if c.App.Name == "" {
		return errors.New("app name is required")
	}
	if c.App.Port <= 0 || c.App.Port > 65535 {
		return errors.New("app port must be between 1 and 65535")
	}
	
	// Database設定の検証
	if c.Database.Host == "" {
		return errors.New("database host is required")
	}
	if c.Database.Port <= 0 || c.Database.Port > 65535 {
		return errors.New("database port must be between 1 and 65535")
	}
	if c.Database.Name == "" {
		return errors.New("database name is required")
	}
	if c.Database.User == "" {
		return errors.New("database user is required")
	}
	
	// Redis設定の検証
	if c.Redis.Host == "" {
		return errors.New("redis host is required")
	}
	if c.Redis.Port <= 0 || c.Redis.Port > 65535 {
		return errors.New("redis port must be between 1 and 65535")
	}
	
	return nil
}

// GetDatabaseDSN はデータベース接続文字列を返します
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host, c.Database.Port, c.Database.User, c.Database.Password,
		c.Database.Name, c.Database.SSLMode)
}

// GetRedisAddr はRedis接続アドレスを返します
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

// GetCORSAllowedOrigins はCORS許可オリジンのリストを返します
func (c *Config) GetCORSAllowedOrigins() []string {
	if c.CORS.AllowedOrigins == "" {
		return []string{}
	}
	return strings.Split(c.CORS.AllowedOrigins, ",")
}

// GetCORSAllowedMethods はCORS許可メソッドのリストを返します
func (c *Config) GetCORSAllowedMethods() []string {
	if c.CORS.AllowedMethods == "" {
		return []string{}
	}
	return strings.Split(c.CORS.AllowedMethods, ",")
}

// GetCORSAllowedHeaders はCORS許可ヘッダーのリストを返します
func (c *Config) GetCORSAllowedHeaders() []string {
	if c.CORS.AllowedHeaders == "" {
		return []string{}
	}
	return strings.Split(c.CORS.AllowedHeaders, ",")
}
