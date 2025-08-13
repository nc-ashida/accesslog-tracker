package models_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"accesslog-tracker/internal/domain/models"
)

func TestApplication_Validate(t *testing.T) {
	tests := []struct {
		name    string
		app     models.Application
		isValid bool
		errors  []string
	}{
		{
			name: "valid application",
			app: models.Application{
				AppID:       "test_app_123",
				Name:        "Test Application",
				Description: "Test application for unit testing",
				Domain:      "example.com",
				APIKey:      "test-api-key-123",
			},
			isValid: true,
			errors:  []string{},
		},
		{
			name: "missing app_id",
			app: models.Application{
				Name:        "Test Application",
				Description: "Test application for unit testing",
				Domain:      "example.com",
				APIKey:      "test-api-key-123",
			},
			isValid: false,
			errors:  []string{"app_id is required"},
		},
		{
			name: "missing name",
			app: models.Application{
				AppID:       "test_app_123",
				Description: "Test application for unit testing",
				Domain:      "example.com",
				APIKey:      "test-api-key-123",
			},
			isValid: false,
			errors:  []string{"name is required"},
		},
		{
			name: "missing api_key",
			app: models.Application{
				AppID:       "test_app_123",
				Name:        "Test Application",
				Description: "Test application for unit testing",
				Domain:      "example.com",
			},
			isValid: false,
			errors:  []string{"api_key is required"},
		},
		{
			name: "invalid domain format",
			app: models.Application{
				AppID:       "test_app_123",
				Name:        "Test Application",
				Description: "Test application for unit testing",
				Domain:      "invalid-domain",
				APIKey:      "test-api-key-123",
			},
			isValid: false,
			errors:  []string{"Invalid domain format"},
		},
		{
			name: "empty domain",
			app: models.Application{
				AppID:       "test_app_123",
				Name:        "Test Application",
				Description: "Test application for unit testing",
				Domain:      "",
				APIKey:      "test-api-key-123",
			},
			isValid: false,
			errors:  []string{"domain is required"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.app.Validate()
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				for _, expectedError := range tt.errors {
					assert.Contains(t, err.Error(), expectedError)
				}
			}
		})
	}
}

func TestApplication_GenerateAPIKey(t *testing.T) {
	app := &models.Application{
		AppID:       "test_app_123",
		Name:        "Test Application",
		Description: "Test application for unit testing",
		Domain:      "example.com",
	}

	err := app.GenerateAPIKey()
	require.NoError(t, err)
	assert.NotEmpty(t, app.APIKey)
	assert.Len(t, app.APIKey, 32) // 32文字のAPIキー
}

func TestApplication_IsActive(t *testing.T) {
	now := time.Now()
	activeApp := &models.Application{
		AppID:       "test_app_123",
		Name:        "Active App",
		Description: "Active application",
		Domain:      "example.com",
		APIKey:      "test-api-key-123",
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	inactiveApp := &models.Application{
		AppID:       "test_app_456",
		Name:        "Inactive App",
		Description: "Inactive application",
		Domain:      "example.com",
		APIKey:      "test-api-key-456",
		Active:      false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	assert.True(t, activeApp.IsActive())
	assert.False(t, inactiveApp.IsActive())
}

func TestApplication_ToJSON(t *testing.T) {
	app := &models.Application{
		AppID:       "test_app_123",
		Name:        "Test Application",
		Description: "Test application for unit testing",
		Domain:      "example.com",
		APIKey:      "test-api-key-123",
		Active:      true,
		CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	jsonData, err := app.ToJSON()
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// JSONの構造を検証
	var parsed map[string]interface{}
	err = json.Unmarshal(jsonData, &parsed)
	require.NoError(t, err)
	assert.Equal(t, app.AppID, parsed["app_id"])
	assert.Equal(t, app.Name, parsed["name"])
	assert.Equal(t, app.Description, parsed["description"])
	assert.Equal(t, app.Domain, parsed["domain"])
	assert.Equal(t, app.APIKey, parsed["api_key"])
	assert.Equal(t, app.Active, parsed["is_active"])
}

func TestApplication_FromJSON(t *testing.T) {
	jsonData := `{
		"app_id": "test_app_123",
		"name": "Test Application",
		"description": "Test application for unit testing",
		"domain": "example.com",
		"api_key": "test-api-key-123",
		"is_active": true,
		"created_at": "2024-01-15T10:30:00Z",
		"updated_at": "2024-01-15T10:30:00Z"
	}`

	app := &models.Application{}
	err := app.FromJSON([]byte(jsonData))

	require.NoError(t, err)
	assert.Equal(t, "test_app_123", app.AppID)
	assert.Equal(t, "Test Application", app.Name)
	assert.Equal(t, "Test application for unit testing", app.Description)
	assert.Equal(t, "example.com", app.Domain)
	assert.Equal(t, "test-api-key-123", app.APIKey)
	assert.True(t, app.IsActive())
}

func TestApplication_ValidateAPIKey(t *testing.T) {
	app := &models.Application{
		AppID:       "test_app_123",
		Name:        "Test Application",
		Description: "Test application for unit testing",
		Domain:      "example.com",
		APIKey:      "test-api-key-123",
		Active:      true,
	}

	tests := []struct {
		name        string
		apiKey      string
		expectValid bool
	}{
		{
			name:        "valid API key",
			apiKey:      "test-api-key-123",
			expectValid: true,
		},
		{
			name:        "invalid API key",
			apiKey:      "invalid-api-key",
			expectValid: false,
		},
		{
			name:        "empty API key",
			apiKey:      "",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := app.ValidateAPIKey(tt.apiKey)
			assert.Equal(t, tt.expectValid, result)
		})
	}
}
