package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"accesslog-tracker/internal/utils/jsonutil"
)

func TestJSONUtil_Marshal(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expected    string
		expectError bool
	}{
		{
			name: "simple struct",
			input: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{
				Name: "John",
				Age:  30,
			},
			expected:    `{"name":"John","age":30}`,
			expectError: false,
		},
		{
			name: "map",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": 123,
			},
			expected:    `{"key1":"value1","key2":123}`,
			expectError: false,
		},
		{
			name:        "nil",
			input:       nil,
			expected:    "null",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := jsonutil.Marshal(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, string(result))
			}
		})
	}
}

func TestJSONUtil_Unmarshal(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		target      interface{}
		expectError bool
	}{
		{
			name:  "valid JSON to struct",
			input: `{"name":"John","age":30}`,
			target: &struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{},
			expectError: false,
		},
		{
			name:        "invalid JSON",
			input:       `{"name":"John","age":30`,
			target:      &struct{}{},
			expectError: true,
		},
		{
			name:        "empty string",
			input:       "",
			target:      &struct{}{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := jsonutil.Unmarshal([]byte(tt.input), tt.target)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestJSONUtil_MarshalIndent(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		prefix      string
		indent      string
		expectError bool
	}{
		{
			name: "simple struct with indent",
			input: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{
				Name: "John",
				Age:  30,
			},
			prefix:      "",
			indent:      "  ",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := jsonutil.MarshalIndent(tt.input, tt.prefix, tt.indent)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, string(result), "name")
				assert.Contains(t, string(result), "age")
			}
		})
	}
}

func TestJSONUtil_IsValidJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid JSON object",
			input:    `{"name":"John","age":30}`,
			expected: true,
		},
		{
			name:     "valid JSON array",
			input:    `[1,2,3]`,
			expected: true,
		},
		{
			name:     "valid JSON string",
			input:    `"hello"`,
			expected: true,
		},
		{
			name:     "invalid JSON",
			input:    `{"name":"John","age":30`,
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := jsonutil.IsValidJSON(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJSONUtil_Merge(t *testing.T) {
	tests := []struct {
		name     string
		base     map[string]interface{}
		override map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "merge simple maps",
			base: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			override: map[string]interface{}{
				"key2": "new_value2",
				"key3": "value3",
			},
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": "new_value2",
				"key3": "value3",
			},
		},
		{
			name: "merge with nil override",
			base: map[string]interface{}{
				"key1": "value1",
			},
			override: nil,
			expected: map[string]interface{}{
				"key1": "value1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := jsonutil.Merge(tt.base, tt.override)
			assert.Equal(t, tt.expected, result)
		})
	}
}
