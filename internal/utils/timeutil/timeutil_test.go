package timeutil

import (
	"testing"
	"time"
)

func TestNow(t *testing.T) {
	now := Now()
	if now.IsZero() {
		t.Error("Now() should return a non-zero time")
	}
}

func TestFormatTime(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 12, 30, 45, 0, time.UTC)
	
	tests := []struct {
		name     string
		format   string
		expected string
	}{
		{
			name:     "RFC3339 format",
			format:   FormatRFC3339,
			expected: "2024-01-15T12:30:45Z",
		},
		{
			name:     "Date format",
			format:   FormatDate,
			expected: "2024-01-15",
		},
		{
			name:     "Unix timestamp",
			format:   FormatUnix,
			expected: "1705321845",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTime(testTime, tt.format)
			if result != tt.expected {
				t.Errorf("FormatTime() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		name     string
		timeStr  string
		format   string
		expected time.Time
		hasError bool
	}{
		{
			name:     "RFC3339 format",
			timeStr:  "2024-01-15T12:30:45Z",
			format:   FormatRFC3339,
			expected: time.Date(2024, 1, 15, 12, 30, 45, 0, time.UTC),
			hasError: false,
		},
		{
			name:     "Date format",
			timeStr:  "2024-01-15",
			format:   FormatDate,
			expected: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			hasError: false,
		},
		{
			name:     "Unix timestamp",
			timeStr:  "1705321845",
			format:   FormatUnix,
			expected: time.Date(2024, 1, 15, 12, 30, 45, 0, time.UTC),
			hasError: false,
		},
		{
			name:     "Invalid format",
			timeStr:  "invalid",
			format:   FormatDate,
			hasError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTime(tt.timeStr, tt.format)
			
			if tt.hasError {
				if err == nil {
					t.Error("ParseTime() should return an error")
				}
			} else {
				if err != nil {
					t.Errorf("ParseTime() returned unexpected error: %v", err)
				}
				if !result.Equal(tt.expected) {
					t.Errorf("ParseTime() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestStartOfDay(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 12, 30, 45, 123456789, time.UTC)
	expected := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	
	result := StartOfDay(testTime)
	if !result.Equal(expected) {
		t.Errorf("StartOfDay() = %v, want %v", result, expected)
	}
}

func TestEndOfDay(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 12, 30, 45, 123456789, time.UTC)
	expected := time.Date(2024, 1, 15, 23, 59, 59, 999999999, time.UTC)
	
	result := EndOfDay(testTime)
	if !result.Equal(expected) {
		t.Errorf("EndOfDay() = %v, want %v", result, expected)
	}
}
