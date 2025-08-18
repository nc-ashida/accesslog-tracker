package validators

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"accesslog-tracker/internal/domain/models"
	"accesslog-tracker/internal/domain/validators"
)

func TestApplicationValidator_Validate(t *testing.T) {
	validator := validators.NewApplicationValidator()

	tests := []struct {
		name    string
		app     *models.Application
		wantErr bool
	}{
		{
			name: "valid application",
			app: &models.Application{
				AppID:       "test_app_123",
				Name:        "Test Application",
				Description: "Test application description",
				Domain:      "example.com",
				APIKey:      "alt_test_api_key_123",
				Active:      true,
			},
			wantErr: false,
		},
		{
			name: "missing app_id",
			app: &models.Application{
				AppID:       "",
				Name:        "Test Application",
				Description: "Test application description",
				Domain:      "example.com",
				APIKey:      "alt_test_api_key_123",
				Active:      true,
			},
			wantErr: true,
		},
		{
			name: "missing name",
			app: &models.Application{
				AppID:       "test_app_123",
				Name:        "",
				Description: "Test application description",
				Domain:      "example.com",
				APIKey:      "alt_test_api_key_123",
				Active:      true,
			},
			wantErr: true,
		},
		{
			name: "missing domain",
			app: &models.Application{
				AppID:       "test_app_123",
				Name:        "Test Application",
				Description: "Test application description",
				Domain:      "",
				APIKey:      "alt_test_api_key_123",
				Active:      true,
			},
			wantErr: true,
		},
		{
			name: "missing api_key",
			app: &models.Application{
				AppID:       "test_app_123",
				Name:        "Test Application",
				Description: "Test application description",
				Domain:      "example.com",
				APIKey:      "",
				Active:      true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.app)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestApplicationValidator_ValidateCreate(t *testing.T) {
	validator := validators.NewApplicationValidator()

	tests := []struct {
		name    string
		app     *models.Application
		wantErr bool
	}{
		{
			name: "valid create request",
			app: &models.Application{
				Name:        "Test Application",
				Description: "Test application description",
				Domain:      "example.com",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			app: &models.Application{
				Name:        "",
				Description: "Test application description",
				Domain:      "example.com",
			},
			wantErr: true,
		},
		{
			name: "missing domain",
			app: &models.Application{
				Name:        "Test Application",
				Description: "Test application description",
				Domain:      "",
			},
			wantErr: true,
		},
		{
			name: "invalid domain",
			app: &models.Application{
				Name:        "Test Application",
				Description: "Test application description",
				Domain:      "invalid-domain",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCreate(tt.app)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestApplicationValidator_ValidateUpdate(t *testing.T) {
	validator := validators.NewApplicationValidator()

	tests := []struct {
		name    string
		app     *models.Application
		wantErr bool
	}{
		{
			name: "valid update request",
			app: &models.Application{
				AppID:       "test_app_123",
				Name:        "Updated Test Application",
				Description: "Updated test application description",
				Domain:      "updated.example.com",
			},
			wantErr: false,
		},
		{
			name: "missing app_id",
			app: &models.Application{
				AppID:       "",
				Name:        "Updated Test Application",
				Description: "Updated test application description",
				Domain:      "updated.example.com",
			},
			wantErr: true,
		},
		{
			name: "invalid domain",
			app: &models.Application{
				AppID:       "test_app_123",
				Name:        "Updated Test Application",
				Description: "Updated test application description",
				Domain:      "invalid-domain",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateUpdate(tt.app)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestApplicationValidator_ValidateAPIKey(t *testing.T) {
	validator := validators.NewApplicationValidator()

	tests := []struct {
		name   string
		apiKey string
		want   error
	}{
		{"valid api key", "alt_test_api_key_123", nil},
		{"invalid api key", "invalid_key", errors.New("API key must be at least 16 characters long")},
		{"empty api key", "", models.ErrApplicationAPIKeyRequired},
		{"api key without prefix", "test_api_key_123", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.ValidateAPIKey(tt.apiKey)
			if tt.want == nil {
				assert.NoError(t, got)
			} else {
				assert.Error(t, got)
				if got != nil {
					assert.Equal(t, tt.want.Error(), got.Error())
				}
			}
		})
	}
}

func TestApplicationValidator_ValidateDomain(t *testing.T) {
	validator := validators.NewApplicationValidator()

	tests := []struct {
		name   string
		domain string
		want   error
	}{
		{"valid domain", "example.com", nil},
		{"valid subdomain", "sub.example.com", nil},
		{"valid domain with www", "www.example.com", nil},
		{"invalid domain", "in", errors.New("domain must be between 3 and 253 characters")},
		{"empty domain", "", models.ErrApplicationDomainRequired},
		{"domain with invalid chars", "example@.com", errors.New("Invalid domain format")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.ValidateDomain(tt.domain)
			if tt.want == nil {
				assert.NoError(t, got)
			} else {
				assert.Error(t, got)
				if got != nil {
					assert.Equal(t, tt.want.Error(), got.Error())
				}
			}
		})
	}
}

func TestApplicationValidator_ValidateName(t *testing.T) {
	validator := validators.NewApplicationValidator()

	tests := []struct {
		testName string
		name     string
		want     error
	}{
		{"valid name", "Test Application", nil},
		{"valid name with numbers", "Test App 123", nil},
		{"empty name", "", models.ErrApplicationNameRequired},
		{"name too short", "A", errors.New("name must be at least 5 characters")},
		{"name too long", "This is a very long application name that exceeds the maximum allowed length", nil},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := validator.ValidateName(tt.name)
			if tt.want == nil {
				assert.NoError(t, got)
			} else {
				assert.Error(t, got)
				if got != nil {
					assert.Equal(t, tt.want.Error(), got.Error())
				}
			}
		})
	}
}

func TestApplicationValidator_ValidateAppID(t *testing.T) {
	validator := validators.NewApplicationValidator()

	tests := []struct {
		name  string
		appID string
		want  error
	}{
		{"valid app id", "test_app_123", nil},
		{"valid app id with numbers", "app_123456", nil},
		{"empty app id", "", models.ErrApplicationAppIDRequired},
		{"app id too short", "a", errors.New("app_id must be at least 8 characters")},
		{"app id with invalid chars", "test@app", errors.New("app_id must start with a letter and contain only letters, numbers, and underscores")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.ValidateAppID(tt.appID)
			if tt.want == nil {
				assert.NoError(t, got)
			} else {
				assert.Error(t, got)
				if got != nil {
					assert.Equal(t, tt.want.Error(), got.Error())
				}
			}
		})
	}
}
