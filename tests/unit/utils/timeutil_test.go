package utils_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/nc-ashida/accesslog-tracker/internal/utils/timeutil"
)

func TestTimeUtil_FormatTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "format UTC timestamp",
			input:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			expected: "2024-01-01T12:00:00Z",
		},
		{
			name:     "format JST timestamp",
			input:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.FixedZone("JST", 9*60*60)),
			expected: "2024-01-01T03:00:00Z", // UTCに変換される
		},
		{
			name:     "format with milliseconds",
			input:    time.Date(2024, 1, 1, 12, 0, 0, 500*1000000, time.UTC),
			expected: "2024-01-01T12:00:00Z", // ミリ秒は切り捨て
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := timeutil.FormatTimestamp(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTimeUtil_ParseTimestamp(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    time.Time
		expectError bool
	}{
		{
			name:        "parse valid UTC timestamp",
			input:       "2024-01-01T12:00:00Z",
			expected:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			expectError: false,
		},
		{
			name:        "parse invalid timestamp",
			input:       "invalid-timestamp",
			expected:    time.Time{},
			expectError: true,
		},
		{
			name:        "parse empty string",
			input:       "",
			expected:    time.Time{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := timeutil.ParseTimestamp(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestTimeUtil_IsToday(t *testing.T) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())
	yesterday := today.AddDate(0, 0, -1)
	tomorrow := today.AddDate(0, 0, 1)

	tests := []struct {
		name     string
		input    time.Time
		expected bool
	}{
		{
			name:     "today",
			input:    today,
			expected: true,
		},
		{
			name:     "yesterday",
			input:    yesterday,
			expected: false,
		},
		{
			name:     "tomorrow",
			input:    tomorrow,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := timeutil.IsToday(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTimeUtil_GetStartOfDay(t *testing.T) {
	input := time.Date(2024, 1, 1, 15, 30, 45, 123456789, time.UTC)
	expected := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	result := timeutil.GetStartOfDay(input)
	assert.Equal(t, expected, result)
}

func TestTimeUtil_GetEndOfDay(t *testing.T) {
	input := time.Date(2024, 1, 1, 15, 30, 45, 123456789, time.UTC)
	expected := time.Date(2024, 1, 1, 23, 59, 59, 999999999, time.UTC)

	result := timeutil.GetEndOfDay(input)
	assert.Equal(t, expected, result)
}
