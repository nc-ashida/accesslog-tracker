package models_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"accesslog-tracker/internal/domain/models"
)

func TestApplication_Validate(t *testing.T) {
	tests := []struct {
		name    string
		app     *models.Application
		wantErr bool
	}{
		{
			name: "valid application",
			app: &models.Application{
				AppID:       "test_app_123",
				Name:        "Test App",
				Description: "Test Description",
				Domain:      "example.com",
				APIKey:      "alt_test_key_123",
			},
			wantErr: false,
		},
		{
			name: "empty app_id",
			app: &models.Application{
				AppID:       "",
				Name:        "Test App",
				Description: "Test Description",
				Domain:      "example.com",
				APIKey:      "alt_test_key_123",
			},
			wantErr: true,
		},
		{
			name: "empty name",
			app: &models.Application{
				AppID:       "test_app_123",
				Name:        "",
				Description: "Test Description",
				Domain:      "example.com",
				APIKey:      "alt_test_key_123",
			},
			wantErr: true,
		},
		{
			name: "empty domain",
			app: &models.Application{
				AppID:       "test_app_123",
				Name:        "Test App",
				Description: "Test Description",
				Domain:      "",
				APIKey:      "alt_test_key_123",
			},
			wantErr: true,
		},
		{
			name: "empty api_key",
			app: &models.Application{
				AppID:       "test_app_123",
				Name:        "Test App",
				Description: "Test Description",
				Domain:      "example.com",
				APIKey:      "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.app.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestApplication_IsValidDomain(t *testing.T) {
	tests := []struct {
		name     string
		domain   string
		expected bool
	}{
		{"valid domain", "example.com", true},
		{"valid subdomain", "sub.example.com", true},
		{"empty domain", "", false},
		{"invalid domain", "invalid-domain", false},
		{"too short", "ab", false},
		{"too long", strings.Repeat("a", 254), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &models.Application{Domain: tt.domain}
			result := app.IsValidDomain()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestApplication_IsValidAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		expected bool
	}{
		{"valid api key", "alt_test_key_1234567890123456", true},
		{"empty api key", "", false},
		{"too short", "alt_short", false},
		{"no prefix", "test_key_1234567890123456", true},              // プレフィックスは必須ではない
		{"with special chars", "alt_test@key_1234567890123456", true}, // 特殊文字も許可
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &models.Application{APIKey: tt.apiKey}
			result := app.IsValidAPIKey()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestApplication_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		active   bool
		expected bool
	}{
		{"active application", true, true},
		{"inactive application", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &models.Application{Active: tt.active}
			result := app.IsActive()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestApplication_GenerateAPIKey(t *testing.T) {
	app := &models.Application{}
	err := app.GenerateAPIKey()

	assert.NoError(t, err)
	assert.NotEmpty(t, app.APIKey)
	assert.Len(t, app.APIKey, 32)
}

func TestApplication_ValidateAPIKey(t *testing.T) {
	app := &models.Application{APIKey: "alt_test_key_1234567890123456"}

	t.Run("valid api key", func(t *testing.T) {
		err := app.ValidateAPIKey("alt_test_key_1234567890123456")
		assert.NoError(t, err)
	})

	t.Run("invalid api key", func(t *testing.T) {
		err := app.ValidateAPIKey("wrong_key")
		assert.Error(t, err)
	})

	t.Run("empty api key", func(t *testing.T) {
		err := app.ValidateAPIKey("")
		assert.Error(t, err)
	})
}

func TestApplication_ToJSON(t *testing.T) {
	app := &models.Application{
		AppID:       "test_app_123",
		Name:        "Test App",
		Description: "Test Description",
		Domain:      "example.com",
		APIKey:      "alt_test_key_123",
		Active:      true,
		CreatedAt:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	jsonData, err := app.ToJSON()
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// JSONとしてパースできることを確認
	var parsedApp models.Application
	err = json.Unmarshal(jsonData, &parsedApp)
	assert.NoError(t, err)
	assert.Equal(t, app.AppID, parsedApp.AppID)
	assert.Equal(t, app.Name, parsedApp.Name)
}

func TestApplication_FromJSON(t *testing.T) {
	jsonData := `{
		"app_id": "test_app_123",
		"name": "Test App",
		"description": "Test Description",
		"domain": "example.com",
		"api_key": "alt_test_key_123",
		"is_active": true
	}`

	app := &models.Application{}
	err := app.FromJSON([]byte(jsonData))

	assert.NoError(t, err)
	assert.Equal(t, "test_app_123", app.AppID)
	assert.Equal(t, "Test App", app.Name)
	assert.Equal(t, "Test Description", app.Description)
	assert.Equal(t, "example.com", app.Domain)
	assert.Equal(t, "alt_test_key_123", app.APIKey)
	assert.True(t, app.Active)
}

func TestIsValidDomain(t *testing.T) {
	tests := []struct {
		name     string
		domain   string
		expected bool
	}{
		{"valid domain", "example.com", true},
		{"valid subdomain", "sub.example.com", true},
		{"empty domain", "", false},
		{"invalid domain", "invalid-domain", false},
		{"domain with space", "example com", false},
		{"too short", "ab", false},
		{"too long", strings.Repeat("a", 254), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := models.IsValidDomain(tt.domain)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		expected bool
	}{
		{"valid api key", "alt_test_key_1234567890123456", true},
		{"empty api key", "", false},
		{"no prefix", "test_key_1234567890123456", false},
		{"too short", "alt_short", false},
		{"with special chars", "alt_test@key_1234567890123456", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := models.IsValidAPIKey(tt.apiKey)
			assert.Equal(t, tt.expected, result)
		})
	}
}
