package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"accesslog-tracker/internal/utils/crypto"
)

func TestCryptoUtil_HashSHA256(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    "hello",
			expected: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := crypto.HashSHA256(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCryptoUtil_GenerateRandomString(t *testing.T) {
	tests := []struct {
		name  string
		length int
	}{
		{
			name:   "16 characters",
			length: 16,
		},
		{
			name:   "32 characters",
			length: 32,
		},
		{
			name:   "64 characters",
			length: 64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := crypto.GenerateRandomString(tt.length)
			assert.Equal(t, tt.length, len(result))
			assert.NotEmpty(t, result)
		})
	}
}

func TestCryptoUtil_GenerateAPIKey(t *testing.T) {
	result := crypto.GenerateAPIKey()
	assert.Equal(t, 32, len(result))
	assert.NotEmpty(t, result)
}

func TestCryptoUtil_ValidateAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		expected bool
	}{
		{
			name:     "valid API key",
			apiKey:   "valid-api-key-32-chars-long",
			expected: true,
		},
		{
			name:     "too short",
			apiKey:   "short",
			expected: false,
		},
		{
			name:     "empty string",
			apiKey:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := crypto.ValidateAPIKey(tt.apiKey)
			assert.Equal(t, tt.expected, result)
		})
	}
}
