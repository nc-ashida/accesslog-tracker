package models_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"accesslog-tracker/internal/domain/models"
)

func TestApplicationModel_Integration(t *testing.T) {
	t.Run("Application Validation", func(t *testing.T) {
		app := &models.Application{
			AppID:       "test-app-123",
			Name:        "Test Application",
			Description: "Test application for integration testing",
			Domain:      "test.example.com",
			APIKey:      "test-api-key-123",
			Active:      true,
		}

		err := app.Validate()
		assert.NoError(t, err)
	})

	t.Run("Application Domain Validation", func(t *testing.T) {
		app := &models.Application{
			AppID:       "test-app-123",
			Name:        "Test Application",
			Description: "Test application for integration testing",
			Domain:      "test.example.com",
			APIKey:      "test-api-key-123",
			Active:      true,
		}

		// 有効なドメイン
		assert.True(t, app.IsValidDomain())

		// 無効なドメイン
		app.Domain = ""
		assert.False(t, app.IsValidDomain())
	})

	t.Run("Application API Key Validation", func(t *testing.T) {
		app := &models.Application{
			AppID:       "test-app-123",
			Name:        "Test Application",
			Description: "Test application for integration testing",
			Domain:      "test.example.com",
			APIKey:      "test-api-key-123",
			Active:      true,
		}

		// 有効なAPIキー
		assert.True(t, app.IsValidAPIKey())

		// 無効なAPIキー
		app.APIKey = ""
		assert.False(t, app.IsValidAPIKey())
	})

	t.Run("Application Active Status", func(t *testing.T) {
		app := &models.Application{
			AppID:       "test-app-123",
			Name:        "Test Application",
			Description: "Test application for integration testing",
			Domain:      "test.example.com",
			APIKey:      "test-api-key-123",
			Active:      true,
		}

		assert.True(t, app.IsActive())

		app.Active = false
		assert.False(t, app.IsActive())
	})

	t.Run("Application API Key Generation", func(t *testing.T) {
		app := &models.Application{
			AppID:       "test-app-123",
			Name:        "Test Application",
			Description: "Test application for integration testing",
			Domain:      "test.example.com",
			APIKey:      "",
			Active:      true,
		}

		err := app.GenerateAPIKey()
		require.NoError(t, err)
		assert.NotEmpty(t, app.APIKey)
		assert.True(t, app.IsValidAPIKey())
	})

	t.Run("Application API Key Validation Method", func(t *testing.T) {
		app := &models.Application{
			AppID:       "test-app-123",
			Name:        "Test Application",
			Description: "Test application for integration testing",
			Domain:      "test.example.com",
			APIKey:      "test-api-key-123",
			Active:      true,
		}

		// 有効なAPIキー
		err := app.ValidateAPIKey("test-api-key-123")
		assert.NoError(t, err)

		// 無効なAPIキー
		err = app.ValidateAPIKey("invalid-api-key")
		assert.Error(t, err)
	})

	t.Run("Application JSON Serialization", func(t *testing.T) {
		app := &models.Application{
			AppID:       "test-app-123",
			Name:        "Test Application",
			Description: "Test application for integration testing",
			Domain:      "test.example.com",
			APIKey:      "test-api-key-123",
			Active:      true,
		}

		// JSONにシリアライズ
		jsonData, err := app.ToJSON()
		require.NoError(t, err)
		assert.NotEmpty(t, jsonData)

		// JSONからデシリアライズ
		newApp := &models.Application{}
		err = newApp.FromJSON(jsonData)
		require.NoError(t, err)

		assert.Equal(t, app.AppID, newApp.AppID)
		assert.Equal(t, app.Name, newApp.Name)
		assert.Equal(t, app.Description, newApp.Description)
		assert.Equal(t, app.Domain, newApp.Domain)
		assert.Equal(t, app.APIKey, newApp.APIKey)
		assert.Equal(t, app.Active, newApp.Active)
	})

	t.Run("Application Static Domain Validation", func(t *testing.T) {
		// 有効なドメイン
		assert.True(t, models.IsValidDomain("example.com"))
		assert.True(t, models.IsValidDomain("test.example.com"))
		assert.True(t, models.IsValidDomain("sub.test.example.com"))

		// 無効なドメイン
		assert.False(t, models.IsValidDomain(""))
		assert.False(t, models.IsValidDomain("ab")) // 3文字未満
		assert.False(t, models.IsValidDomain("invalid-domain"))
		assert.False(t, models.IsValidDomain("test example.com"))
	})

	t.Run("Application Static API Key Validation", func(t *testing.T) {
		// 有効なAPIキー
		assert.True(t, models.IsValidAPIKey("alt_test_api_key_123"))
		assert.True(t, models.IsValidAPIKey("alt_valid_api_key"))

		// 無効なAPIキー
		assert.False(t, models.IsValidAPIKey(""))
		assert.False(t, models.IsValidAPIKey("invalid"))
		assert.False(t, models.IsValidAPIKey("test-api-key-123"))
		assert.False(t, models.IsValidAPIKey("alt_test@key"))
	})
}
