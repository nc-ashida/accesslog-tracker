package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"accesslog-tracker/internal/utils/iputil"
)

func TestIPUtil_IsValidIP(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid IPv4",
			input:    "192.168.1.1",
			expected: true,
		},
		{
			name:     "valid IPv4 with zeros",
			input:    "0.0.0.0",
			expected: true,
		},
		{
			name:     "valid IPv4 localhost",
			input:    "127.0.0.1",
			expected: true,
		},
		{
			name:     "valid IPv6",
			input:    "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			expected: true,
		},
		{
			name:     "valid IPv6 compressed",
			input:    "2001:db8::1",
			expected: true,
		},
		{
			name:     "invalid IP",
			input:    "invalid-ip",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "IPv4 with invalid octet",
			input:    "192.168.1.256",
			expected: false,
		},
		{
			name:     "IPv4 with too many octets",
			input:    "192.168.1.1.1",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := iputil.IsValidIP(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIPUtil_IsPrivateIP(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "private IPv4 class A",
			input:    "10.0.0.1",
			expected: true,
		},
		{
			name:     "private IPv4 class B",
			input:    "172.16.0.1",
			expected: true,
		},
		{
			name:     "private IPv4 class C",
			input:    "192.168.1.1",
			expected: true,
		},
		{
			name:     "public IPv4",
			input:    "203.0.113.1",
			expected: false,
		},
		{
			name:     "localhost",
			input:    "127.0.0.1",
			expected: true,
		},
		{
			name:     "invalid IP",
			input:    "invalid-ip",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := iputil.IsPrivateIP(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIPUtil_ExtractIPFromHeader(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		expected string
	}{
		{
			name: "X-Forwarded-For with single IP",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1",
			},
			expected: "192.168.1.1",
		},
		{
			name: "X-Forwarded-For with multiple IPs",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.1, 192.168.1.1",
			},
			expected: "203.0.113.1",
		},
		{
			name: "X-Real-IP",
			headers: map[string]string{
				"X-Real-IP": "192.168.1.1",
			},
			expected: "192.168.1.1",
		},
		{
			name: "X-Forwarded-For and X-Real-IP (X-Forwarded-For優先)",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.1",
				"X-Real-IP":       "192.168.1.1",
			},
			expected: "203.0.113.1",
		},
		{
			name: "no proxy headers",
			headers: map[string]string{
				"User-Agent": "Mozilla/5.0",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := iputil.ExtractIPFromHeader(tt.headers)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIPUtil_AnonymizeIP(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "IPv4 anonymization",
			input:    "192.168.1.100",
			expected: "192.168.1.0",
		},
		{
			name:     "IPv6 anonymization",
			input:    "2001:db8:85a3:0000:0000:8a2e:0370:7334",
			expected: "2001:db8:85a3::",
		},
		{
			name:     "invalid IP",
			input:    "invalid-ip",
			expected: "invalid-ip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := iputil.AnonymizeIP(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
