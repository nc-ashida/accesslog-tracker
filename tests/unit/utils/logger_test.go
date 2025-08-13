package utils_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/nc-ashida/accesslog-tracker/internal/utils/logger"
)

func TestLogger_NewLogger(t *testing.T) {
	log := logger.NewLogger()
	assert.NotNil(t, log)
}

func TestLogger_SetLevel(t *testing.T) {
	log := logger.NewLogger()
	
	tests := []struct {
		name     string
		level    string
		expected bool
	}{
		{
			name:     "debug level",
			level:    "debug",
			expected: true,
		},
		{
			name:     "info level",
			level:    "info",
			expected: true,
		},
		{
			name:     "warn level",
			level:    "warn",
			expected: true,
		},
		{
			name:     "error level",
			level:    "error",
			expected: true,
		},
		{
			name:     "invalid level",
			level:    "invalid",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := log.SetLevel(tt.level)
			if tt.expected {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestLogger_SetFormat(t *testing.T) {
	log := logger.NewLogger()
	
	tests := []struct {
		name     string
		format   string
		expected bool
	}{
		{
			name:     "json format",
			format:   "json",
			expected: true,
		},
		{
			name:     "text format",
			format:   "text",
			expected: true,
		},
		{
			name:     "invalid format",
			format:   "invalid",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := log.SetFormat(tt.format)
			if tt.expected {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestLogger_Logging(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLogger()
	log.SetOutput(&buf)
	log.SetFormat("json")
	
	// Infoログのテスト
	log.Info("test info message")
	
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)
	assert.Equal(t, "info", logEntry["level"])
	assert.Equal(t, "test info message", logEntry["msg"])
	
	// バッファをクリア
	buf.Reset()
	
	// Errorログのテスト
	log.Error("test error message")
	
	err = json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)
	assert.Equal(t, "error", logEntry["level"])
	assert.Equal(t, "test error message", logEntry["msg"])
}

func TestLogger_WithFields(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLogger()
	log.SetOutput(&buf)
	log.SetFormat("json")
	
	fields := map[string]interface{}{
		"user_id": 123,
		"action":  "login",
	}
	
	log.WithFields(fields).Info("user logged in")
	
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)
	assert.Equal(t, "info", logEntry["level"])
	assert.Equal(t, "user logged in", logEntry["msg"])
	assert.Equal(t, float64(123), logEntry["user_id"])
	assert.Equal(t, "login", logEntry["action"])
}

func TestLogger_WithError(t *testing.T) {
	var buf bytes.Buffer
	log := logger.NewLogger()
	log.SetOutput(&buf)
	log.SetFormat("json")
	
	testError := assert.AnError
	log.WithError(testError).Error("operation failed")
	
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)
	assert.Equal(t, "error", logEntry["level"])
	assert.Equal(t, "operation failed", logEntry["msg"])
	assert.NotEmpty(t, logEntry["error"])
}
