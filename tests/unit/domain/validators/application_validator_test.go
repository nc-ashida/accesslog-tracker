package validators_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nc-ashida/accesslog-tracker/internal/domain/models"
	"github.com/nc-ashida/accesslog-tracker/internal/domain/validators"
)

func TestApplicationValidator_ValidateCreate(t *testing.T) {
	validator := validators.NewApplicationValidator()

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
			},
			isValid: false,
			errors:  []string{"name is required"},
		},
		{
			name: "invalid domain format",
			app: models.Application{
				AppID:       "test_app_123",
				Name:        "Test Application",
				Description: "Test application for unit testing",
				Domain:      "invalid-domain",
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
			},
			isValid: false,
			errors:  []string{"domain is required"},
		},
		{
			name: "app_id too short",
			app: models.Application{
				AppID:       "short",
				Name:        "Test Application",
				Description: "Test application for unit testing",
				Domain:      "example.com",
			},
			isValid: false,
			errors:  []string{"app_id must be at least 8 characters"},
		},
		{
			name: "name too short",
			app: models.Application{
				AppID:       "test_app_123",
				Name:        "Test",
				Description: "Test application for unit testing",
				Domain:      "example.com",
			},
			isValid: false,
			errors:  []string{"name must be at least 5 characters"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCreate(&tt.app)
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

func TestApplicationValidator_ValidateUpdate(t *testing.T) {
	validator := validators.NewApplicationValidator()

	tests := []struct {
		name    string
		app     models.Application
		isValid bool
		errors  []string
	}{
		{
			name: "valid application update",
			app: models.Application{
				AppID:       "test_app_123",
				Name:        "Updated Test Application",
				Description: "Updated test application for unit testing",
				Domain:      "updated.example.com",
				APIKey:      "test-api-key-123",
			},
			isValid: true,
			errors:  []string{},
		},
		{
			name: "missing app_id",
			app: models.Application{
				Name:        "Updated Test Application",
				Description: "Updated test application for unit testing",
				Domain:      "updated.example.com",
				APIKey:      "test-api-key-123",
			},
			isValid: false,
			errors:  []string{"app_id is required"},
		},
		{
			name: "missing api_key",
			app: models.Application{
				AppID:       "test_app_123",
				Name:        "Updated Test Application",
				Description: "Updated test application for unit testing",
				Domain:      "updated.example.com",
			},
			isValid: false,
			errors:  []string{"api_key is required"},
		},
		{
			name: "invalid domain format",
			app: models.Application{
				AppID:       "test_app_123",
				Name:        "Updated Test Application",
				Description: "Updated test application for unit testing",
				Domain:      "invalid-domain",
				APIKey:      "test-api-key-123",
			},
			isValid: false,
			errors:  []string{"Invalid domain format"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateUpdate(&tt.app)
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

func TestApplicationValidator_ValidateAPIKey(t *testing.T) {
	validator := validators.NewApplicationValidator()

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
			name:        "API key too short",
			apiKey:      "short",
			expectValid: false,
		},
		{
			name:        "empty API key",
			apiKey:      "",
			expectValid: false,
		},
		{
			name:        "API key with invalid characters",
			apiKey:      "test-api-key-123!@#",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateAPIKey(tt.apiKey)
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestApplicationValidator_ValidateDomain(t *testing.T) {
	validator := validators.NewApplicationValidator()

	tests := []struct {
		name        string
		domain      string
		expectValid bool
	}{
		{
			name:        "valid domain",
			domain:      "example.com",
			expectValid: true,
		},
		{
			name:        "valid subdomain",
			domain:      "sub.example.com",
			expectValid: true,
		},
		{
			name:        "valid domain with www",
			domain:      "www.example.com",
			expectValid: true,
		},
		{
			name:        "invalid domain format",
			domain:      "invalid-domain",
			expectValid: false,
		},
		{
			name:        "domain with invalid characters",
			domain:      "example.com!@#",
			expectValid: false,
		},
		{
			name:        "empty domain",
			domain:      "",
			expectValid: false,
		},
		{
			name:        "domain starting with dash",
			domain:      "-example.com",
			expectValid: false,
		},
		{
			name:        "domain ending with dash",
			domain:      "example-.com",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateDomain(tt.domain)
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestApplicationValidator_ValidateName(t *testing.T) {
	validator := validators.NewApplicationValidator()

	tests := []struct {
		name        string
		appName     string
		expectValid bool
	}{
		{
			name:        "valid name",
			appName:     "Test Application",
			expectValid: true,
		},
		{
			name:        "name too short",
			appName:     "Test",
			expectValid: false,
		},
		{
			name:        "name too long",
			appName:     "This is a very long application name that exceeds the maximum allowed length of 100 characters and should be rejected",
			expectValid: false,
		},
		{
			name:        "empty name",
			appName:     "",
			expectValid: false,
		},
		{
			name:        "name with special characters",
			appName:     "Test App!@#",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateName(tt.appName)
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestApplicationValidator_ValidateAppID(t *testing.T) {
	validator := validators.NewApplicationValidator()

	tests := []struct {
		name        string
		appID       string
		expectValid bool
	}{
		{
			name:        "valid app_id",
			appID:       "test_app_123",
			expectValid: true,
		},
		{
			name:        "app_id too short",
			appID:       "short",
			expectValid: false,
		},
		{
			name:        "app_id too long",
			appID:       "this_is_a_very_long_application_id_that_exceeds_the_maximum_allowed_length_and_should_be_rejected",
			expectValid: false,
		},
		{
			name:        "empty app_id",
			appID:       "",
			expectValid: false,
		},
		{
			name:        "app_id with invalid characters",
			appID:       "test-app-123!@#",
			expectValid: false,
		},
		{
			name:        "app_id starting with number",
			appID:       "123test_app",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateAppID(tt.appID)
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
