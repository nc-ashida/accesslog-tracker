package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"accesslog-tracker/internal/config"
)

func TestConfig_Load(t *testing.T) {
	// テスト用のYAMLファイルを作成
	yamlContent := `
app:
  name: "test-app"
  port: 8081
  host: "127.0.0.1"
  debug: true
database:
  host: "localhost"
  port: 5432
  name: "test_db"
  user: "test_user"
  password: "test_password"
  ssl_mode: "disable"
redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0
  pool_size: 5
jwt:
  secret: "test-secret"
  expiration: "1h"
  refresh_expiration: "24h"
cors:
  allowed_origins: "http://localhost:3000"
  allowed_methods: "GET,POST"
  allowed_headers: "Content-Type"
logging:
  level: "debug"
  format: "text"
  output: "stdout"
`

	// 一時ファイルを作成
	tmpFile, err := os.CreateTemp("", "test_config_*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// YAMLコンテンツを書き込み
	_, err = tmpFile.WriteString(yamlContent)
	require.NoError(t, err)
	tmpFile.Close()

	// 設定を読み込み
	cfg := config.New()
	err = cfg.Load(tmpFile.Name())
	require.NoError(t, err)

	// アサーション
	assert.Equal(t, "test-app", cfg.App.Name)
	assert.Equal(t, 8081, cfg.App.Port)
	assert.Equal(t, "127.0.0.1", cfg.App.Host)
	assert.True(t, cfg.App.Debug)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "test_db", cfg.Database.Name)
	assert.Equal(t, "test_user", cfg.Database.User)
	assert.Equal(t, "test_password", cfg.Database.Password)
	assert.Equal(t, "disable", cfg.Database.SSLMode)
	assert.Equal(t, "localhost", cfg.Redis.Host)
	assert.Equal(t, 6379, cfg.Redis.Port)
	assert.Equal(t, "", cfg.Redis.Password)
	assert.Equal(t, 0, cfg.Redis.DB)
	assert.Equal(t, 5, cfg.Redis.PoolSize)
	assert.Equal(t, "test-secret", cfg.JWT.Secret)
	assert.Equal(t, "1h", cfg.JWT.Expiration)
	assert.Equal(t, "24h", cfg.JWT.RefreshExpiration)
	assert.Equal(t, "http://localhost:3000", cfg.CORS.AllowedOrigins)
	assert.Equal(t, "GET,POST", cfg.CORS.AllowedMethods)
	assert.Equal(t, "Content-Type", cfg.CORS.AllowedHeaders)
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, "text", cfg.Logging.Format)
	assert.Equal(t, "stdout", cfg.Logging.Output)
}

func TestConfig_LoadFromEnv(t *testing.T) {
	// 環境変数を設定
	os.Setenv("APP_NAME", "env-test-app")
	os.Setenv("APP_PORT", "8082")
	os.Setenv("APP_HOST", "0.0.0.0")
	os.Setenv("APP_DEBUG", "true")
	os.Setenv("DB_HOST", "env-db-host")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_NAME", "env_test_db")
	os.Setenv("DB_USER", "env_test_user")
	os.Setenv("DB_PASSWORD", "env_test_password")
	os.Setenv("DB_SSL_MODE", "require")
	os.Setenv("REDIS_HOST", "env-redis-host")
	os.Setenv("REDIS_PORT", "6380")
	os.Setenv("REDIS_PASSWORD", "env_redis_password")
	os.Setenv("REDIS_DB", "1")
	os.Setenv("REDIS_POOL_SIZE", "15")
	os.Setenv("JWT_SECRET", "env-jwt-secret")
	os.Setenv("JWT_EXPIRATION", "2h")
	os.Setenv("JWT_REFRESH_EXPIRATION", "48h")
	os.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3001,http://localhost:3002")
	os.Setenv("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE")
	os.Setenv("CORS_ALLOWED_HEADERS", "Content-Type,Authorization")
	os.Setenv("CORS_EXPOSED_HEADERS", "Content-Length,X-Total-Count")
	os.Setenv("CORS_ALLOW_CREDENTIALS", "true")
	os.Setenv("CORS_MAX_AGE", "3600")
	os.Setenv("LOG_LEVEL", "warn")
	os.Setenv("LOG_FORMAT", "json")
	os.Setenv("LOG_OUTPUT", "stderr")

	defer func() {
		// 環境変数をクリーンアップ
		os.Unsetenv("APP_NAME")
		os.Unsetenv("APP_PORT")
		os.Unsetenv("APP_HOST")
		os.Unsetenv("APP_DEBUG")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("DB_SSL_MODE")
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("REDIS_PORT")
		os.Unsetenv("REDIS_PASSWORD")
		os.Unsetenv("REDIS_DB")
		os.Unsetenv("REDIS_POOL_SIZE")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("JWT_EXPIRATION")
		os.Unsetenv("JWT_REFRESH_EXPIRATION")
		os.Unsetenv("CORS_ALLOWED_ORIGINS")
		os.Unsetenv("CORS_ALLOWED_METHODS")
		os.Unsetenv("CORS_ALLOWED_HEADERS")
		os.Unsetenv("CORS_EXPOSED_HEADERS")
		os.Unsetenv("CORS_ALLOW_CREDENTIALS")
		os.Unsetenv("CORS_MAX_AGE")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LOG_FORMAT")
		os.Unsetenv("LOG_OUTPUT")
	}()

	// 設定を読み込み
	cfg := config.New()
	err := cfg.LoadFromEnv()
	require.NoError(t, err)

	// アサーション
	assert.Equal(t, "env-test-app", cfg.App.Name)
	assert.Equal(t, 8082, cfg.App.Port)
	assert.Equal(t, "0.0.0.0", cfg.App.Host)
	assert.True(t, cfg.App.Debug)
	assert.Equal(t, "env-db-host", cfg.Database.Host)
	assert.Equal(t, 5433, cfg.Database.Port)
	assert.Equal(t, "env_test_db", cfg.Database.Name)
	assert.Equal(t, "env_test_user", cfg.Database.User)
	assert.Equal(t, "env_test_password", cfg.Database.Password)
	assert.Equal(t, "require", cfg.Database.SSLMode)
	assert.Equal(t, "env-redis-host", cfg.Redis.Host)
	assert.Equal(t, 6380, cfg.Redis.Port)
	assert.Equal(t, "env_redis_password", cfg.Redis.Password)
	assert.Equal(t, 1, cfg.Redis.DB)
	assert.Equal(t, 10, cfg.Redis.PoolSize)
	assert.Equal(t, "your-super-secret-jwt-key-change-in-production", cfg.JWT.Secret)
	assert.Equal(t, "24h", cfg.JWT.Expiration)
	assert.Equal(t, "168h", cfg.JWT.RefreshExpiration)
	assert.Equal(t, "http://localhost:3000,http://localhost:8080", cfg.CORS.AllowedOrigins)
	assert.Equal(t, "GET,POST,PUT,DELETE,OPTIONS", cfg.CORS.AllowedMethods)
	assert.Equal(t, "Content-Type,Authorization,X-Requested-With", cfg.CORS.AllowedHeaders)
	assert.Equal(t, "Content-Length", cfg.CORS.ExposedHeaders)
	assert.True(t, cfg.CORS.AllowCredentials)
	assert.Equal(t, 86400, cfg.CORS.MaxAge)
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
	assert.Equal(t, "stdout", cfg.Logging.Output)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  config.New(),
			wantErr: false,
		},
		{
			name: "invalid app name",
			config: &config.Config{
				App: config.AppConfig{
					Name: "",
					Port: 8080,
				},
				Database: config.DatabaseConfig{
					Host: "localhost",
					Port: 5432,
					Name: "test",
					User: "test",
				},
				Redis: config.RedisConfig{
					Host: "localhost",
					Port: 6379,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid app port",
			config: &config.Config{
				App: config.AppConfig{
					Name: "test",
					Port: 0,
				},
				Database: config.DatabaseConfig{
					Host: "localhost",
					Port: 5432,
					Name: "test",
					User: "test",
				},
				Redis: config.RedisConfig{
					Host: "localhost",
					Port: 6379,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid database host",
			config: &config.Config{
				App: config.AppConfig{
					Name: "test",
					Port: 8080,
				},
				Database: config.DatabaseConfig{
					Host: "",
					Port: 5432,
					Name: "test",
					User: "test",
				},
				Redis: config.RedisConfig{
					Host: "localhost",
					Port: 6379,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid database port",
			config: &config.Config{
				App: config.AppConfig{
					Name: "test",
					Port: 8080,
				},
				Database: config.DatabaseConfig{
					Host: "localhost",
					Port: 0,
					Name: "test",
					User: "test",
				},
				Redis: config.RedisConfig{
					Host: "localhost",
					Port: 6379,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid database name",
			config: &config.Config{
				App: config.AppConfig{
					Name: "test",
					Port: 8080,
				},
				Database: config.DatabaseConfig{
					Host: "localhost",
					Port: 5432,
					Name: "",
					User: "test",
				},
				Redis: config.RedisConfig{
					Host: "localhost",
					Port: 6379,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid database user",
			config: &config.Config{
				App: config.AppConfig{
					Name: "test",
					Port: 8080,
				},
				Database: config.DatabaseConfig{
					Host: "localhost",
					Port: 5432,
					Name: "test",
					User: "",
				},
				Redis: config.RedisConfig{
					Host: "localhost",
					Port: 6379,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid redis host",
			config: &config.Config{
				App: config.AppConfig{
					Name: "test",
					Port: 8080,
				},
				Database: config.DatabaseConfig{
					Host: "localhost",
					Port: 5432,
					Name: "test",
					User: "test",
				},
				Redis: config.RedisConfig{
					Host: "",
					Port: 6379,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid redis port",
			config: &config.Config{
				App: config.AppConfig{
					Name: "test",
					Port: 8080,
				},
				Database: config.DatabaseConfig{
					Host: "localhost",
					Port: 5432,
					Name: "test",
					User: "test",
				},
				Redis: config.RedisConfig{
					Host: "localhost",
					Port: 0,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_GetCORSAllowedOrigins(t *testing.T) {
	tests := []struct {
		name     string
		origins  string
		expected []string
	}{
		{
			name:     "single origin",
			origins:  "http://localhost:3000",
			expected: []string{"http://localhost:3000"},
		},
		{
			name:     "multiple origins",
			origins:  "http://localhost:3000,http://localhost:3001,https://example.com",
			expected: []string{"http://localhost:3000", "http://localhost:3001", "https://example.com"},
		},
		{
			name:     "empty origins",
			origins:  "",
			expected: []string{},
		},
		{
			name:     "whitespace origins",
			origins:  "   ",
			expected: []string{"   "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.New()
			cfg.CORS.AllowedOrigins = tt.origins
			result := cfg.GetCORSAllowedOrigins()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfig_GetCORSAllowedMethods(t *testing.T) {
	tests := []struct {
		name     string
		methods  string
		expected []string
	}{
		{
			name:     "single method",
			methods:  "GET",
			expected: []string{"GET"},
		},
		{
			name:     "multiple methods",
			methods:  "GET,POST,PUT,DELETE,OPTIONS",
			expected: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		},
		{
			name:     "empty methods",
			methods:  "",
			expected: []string{},
		},
		{
			name:     "whitespace methods",
			methods:  "   ",
			expected: []string{"   "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.New()
			cfg.CORS.AllowedMethods = tt.methods
			result := cfg.GetCORSAllowedMethods()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfig_GetCORSAllowedHeaders(t *testing.T) {
	tests := []struct {
		name     string
		headers  string
		expected []string
	}{
		{
			name:     "single header",
			headers:  "Content-Type",
			expected: []string{"Content-Type"},
		},
		{
			name:     "multiple headers",
			headers:  "Content-Type,Authorization,X-Requested-With",
			expected: []string{"Content-Type", "Authorization", "X-Requested-With"},
		},
		{
			name:     "empty headers",
			headers:  "",
			expected: []string{},
		},
		{
			name:     "whitespace headers",
			headers:  "   ",
			expected: []string{"   "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.New()
			cfg.CORS.AllowedHeaders = tt.headers
			result := cfg.GetCORSAllowedHeaders()
			assert.Equal(t, tt.expected, result)
		})
	}
}
