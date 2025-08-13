package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/nc-ashida/accesslog-tracker/internal/config"
)

func TestConfig_Load(t *testing.T) {
	// テスト用の設定ファイルを作成
	testConfig := `
app:
  name: "test-app"
  port: 8080
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
`
	
	// 一時ファイルを作成
	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	
	_, err = tmpFile.WriteString(testConfig)
	assert.NoError(t, err)
	tmpFile.Close()
	
	// 設定を読み込み
	cfg := config.New()
	err = cfg.Load(tmpFile.Name())
	
	assert.NoError(t, err)
	assert.Equal(t, "test-app", cfg.App.Name)
	assert.Equal(t, 8080, cfg.App.Port)
	assert.True(t, cfg.App.Debug)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "test_db", cfg.Database.Name)
}

func TestConfig_LoadFromEnv(t *testing.T) {
	// 環境変数を設定
	os.Setenv("APP_NAME", "env-test-app")
	os.Setenv("APP_PORT", "9090")
	os.Setenv("DB_HOST", "env-db-host")
	os.Setenv("DB_PORT", "5433")
	
	defer func() {
		os.Unsetenv("APP_NAME")
		os.Unsetenv("APP_PORT")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
	}()
	
	cfg := config.New()
	err := cfg.LoadFromEnv()
	
	assert.NoError(t, err)
	assert.Equal(t, "env-test-app", cfg.App.Name)
	assert.Equal(t, 9090, cfg.App.Port)
	assert.Equal(t, "env-db-host", cfg.Database.Host)
	assert.Equal(t, 5433, cfg.Database.Port)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.Config
		isValid bool
	}{
		{
			name: "valid config",
			config: &config.Config{
				App: config.AppConfig{
					Name: "test-app",
					Port: 8080,
				},
				Database: config.DatabaseConfig{
					Host: "localhost",
					Port: 5432,
					Name: "test_db",
					User: "test_user",
				},
				Redis: config.RedisConfig{
					Host: "localhost",
					Port: 6379,
				},
			},
			isValid: true,
		},
		{
			name: "missing app name",
			config: &config.Config{
				App: config.AppConfig{
					Port: 8080,
				},
				Database: config.DatabaseConfig{
					Host: "localhost",
					Port: 5432,
					Name: "test_db",
					User: "test_user",
				},
				Redis: config.RedisConfig{
					Host: "localhost",
					Port: 6379,
				},
			},
			isValid: false,
		},
		{
			name: "invalid port",
			config: &config.Config{
				App: config.AppConfig{
					Name: "test-app",
					Port: -1,
				},
				Database: config.DatabaseConfig{
					Host: "localhost",
					Port: 5432,
					Name: "test_db",
					User: "test_user",
				},
				Redis: config.RedisConfig{
					Host: "localhost",
					Port: 6379,
				},
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
