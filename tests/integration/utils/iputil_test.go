package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"accesslog-tracker/internal/utils/iputil"
)

// bytes2IntIPv6はプライベートメソッドのため、直接テストできません
// 代わりに、IPv6アドレスの検証を通じて間接的にテストします
func TestIPUtil_IPv6Validation(t *testing.T) {
	// IPv6アドレスの検証テスト
	ipv6Addresses := []string{
		"2001:db8::1",
		"2001:0db8:0000:0000:0000:0000:0000:0001",
		"::1",
		"fe80::1",
		"fd00::1",
	}

	for _, ip := range ipv6Addresses {
		t.Run(ip, func(t *testing.T) {
			assert.True(t, iputil.IsValidIP(ip))
			assert.True(t, iputil.IsIPv6(ip))
			assert.False(t, iputil.IsIPv4(ip))
		})
	}
}

func TestIPUtil_ComplexIPValidation(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{
			name:     "valid IPv4 with leading zeros",
			ip:       "192.168.001.001",
			expected: false,
		},
		{
			name:     "valid IPv4 with maximum values",
			ip:       "255.255.255.255",
			expected: true,
		},
		{
			name:     "valid IPv4 with minimum values",
			ip:       "0.0.0.0",
			expected: true,
		},
		{
			name:     "invalid IPv4 with out of range values",
			ip:       "256.256.256.256",
			expected: false,
		},
		{
			name:     "invalid IPv4 with negative values",
			ip:       "-1.-1.-1.-1",
			expected: false,
		},
		{
			name:     "valid IPv6 with mixed notation",
			ip:       "2001:db8::1",
			expected: true,
		},
		{
			name:     "valid IPv6 with full notation",
			ip:       "2001:0db8:0000:0000:0000:0000:0000:0001",
			expected: true,
		},
		{
			name:     "valid IPv6 with leading zeros",
			ip:       "2001:0db8:0000:0000:0000:0000:0000:0001",
			expected: true,
		},
		{
			name:     "invalid IPv6 with too many segments",
			ip:       "2001:db8::1::2",
			expected: false,
		},
		{
			name:     "invalid IPv6 with invalid characters",
			ip:       "2001:db8::g",
			expected: false,
		},
		{
			name:     "empty string",
			ip:       "",
			expected: false,
		},
		{
			name:     "whitespace only",
			ip:       "   ",
			expected: false,
		},
		{
			name:     "invalid format with dots and colons",
			ip:       "192.168.1.1:8080",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := iputil.IsValidIP(tt.ip)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIPUtil_ComplexPrivateIPDetection(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{
			name:     "private IPv4 - Class A",
			ip:       "10.0.0.1",
			expected: true,
		},
		{
			name:     "private IPv4 - Class A boundary",
			ip:       "10.255.255.255",
			expected: true,
		},
		{
			name:     "private IPv4 - Class B",
			ip:       "172.16.0.1",
			expected: true,
		},
		{
			name:     "private IPv4 - Class B boundary",
			ip:       "172.31.255.255",
			expected: true,
		},
		{
			name:     "private IPv4 - Class C",
			ip:       "192.168.0.1",
			expected: true,
		},
		{
			name:     "private IPv4 - Class C boundary",
			ip:       "192.168.255.255",
			expected: true,
		},
		{
			name:     "public IPv4 - not in private ranges",
			ip:       "8.8.8.8",
			expected: false,
		},
		{
			name:     "public IPv4 - Google DNS",
			ip:       "8.8.4.4",
			expected: false,
		},
		{
			name:     "public IPv4 - Cloudflare DNS",
			ip:       "1.1.1.1",
			expected: false,
		},
		{
			name:     "loopback IPv4",
			ip:       "127.0.0.1",
			expected: true,
		},
		{
			name:     "link-local IPv4",
			ip:       "169.254.0.1",
			expected: true,
		},
		{
			name:     "private IPv6 - unique local",
			ip:       "fd00::1",
			expected: true,
		},
		{
			name:     "private IPv6 - link-local",
			ip:       "fe80::1",
			expected: true,
		},
		{
			name:     "public IPv6",
			ip:       "2001:db8::1",
			expected: false,
		},
		{
			name:     "invalid IP",
			ip:       "invalid-ip",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := iputil.IsPrivateIP(tt.ip)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIPUtil_ComplexIPVersionDetection(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected int
	}{
		{
			name:     "IPv4 address",
			ip:       "192.168.1.1",
			expected: 4,
		},
		{
			name:     "IPv6 address",
			ip:       "2001:db8::1",
			expected: 6,
		},
		{
			name:     "IPv6 address with full notation",
			ip:       "2001:0db8:0000:0000:0000:0000:0000:0001",
			expected: 6,
		},
		{
			name:     "invalid IP",
			ip:       "invalid-ip",
			expected: 0,
		},
		{
			name:     "empty string",
			ip:       "",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := iputil.GetIPVersion(tt.ip)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIPUtil_ComplexIPExtraction(t *testing.T) {
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
				"X-Forwarded-For": "203.0.113.1, 192.168.1.1, 10.0.0.1",
			},
			expected: "203.0.113.1",
		},
		{
			name: "X-Real-IP header",
			headers: map[string]string{
				"X-Real-IP": "192.168.1.2",
			},
			expected: "192.168.1.2",
		},
		{
			name: "CF-Connecting-IP header",
			headers: map[string]string{
				"CF-Connecting-IP": "203.0.113.2",
			},
			expected: "",
		},
		{
			name: "X-Forwarded-For with IPv6",
			headers: map[string]string{
				"X-Forwarded-For": "2001:db8::1",
			},
			expected: "2001:db8::1",
		},
		{
			name: "X-Forwarded-For with mixed IPv4 and IPv6",
			headers: map[string]string{
				"X-Forwarded-For": "2001:db8::1, 192.168.1.1",
			},
			expected: "2001:db8::1",
		},
		{
			name: "no IP headers",
			headers: map[string]string{
				"User-Agent": "Mozilla/5.0",
			},
			expected: "",
		},
		{
			name:     "empty headers",
			headers:  map[string]string{},
			expected: "",
		},
		{
			name: "X-Forwarded-For with invalid IPs",
			headers: map[string]string{
				"X-Forwarded-For": "invalid-ip, another-invalid",
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

func TestIPUtil_ComplexIPAnonymization(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected string
	}{
		{
			name:     "IPv4 address",
			ip:       "192.168.1.100",
			expected: "192.168.1.0",
		},
		{
			name:     "IPv4 address with different last octet",
			ip:       "192.168.1.200",
			expected: "192.168.1.0",
		},
		{
			name:     "IPv4 address with zero last octet",
			ip:       "192.168.1.0",
			expected: "192.168.1.0",
		},
		{
			name:     "IPv4 address with 255 last octet",
			ip:       "192.168.1.255",
			expected: "192.168.1.0",
		},
		{
			name:     "IPv6 address",
			ip:       "2001:db8:1234:5678:9abc:def0:1234:5678",
			expected: "2001:db8:1234:5678::",
		},
		{
			name:     "IPv6 address with zero last segment",
			ip:       "2001:db8:1234:5678:9abc:def0:1234:0000",
			expected: "2001:db8:1234:5678::",
		},
		{
			name:     "IPv6 address with ffff last segment",
			ip:       "2001:db8:1234:5678:9abc:def0:1234:ffff",
			expected: "2001:db8:1234:5678::",
		},
		{
			name:     "invalid IP",
			ip:       "invalid-ip",
			expected: "invalid-ip",
		},
		{
			name:     "empty string",
			ip:       "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := iputil.AnonymizeIP(tt.ip)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIPUtil_ComplexClientIPExtraction(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expected   string
	}{
		{
			name: "X-Forwarded-For header present",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.1",
			},
			remoteAddr: "192.168.1.1:8080",
			expected:   "203.0.113.1",
		},
		{
			name: "X-Real-IP header present",
			headers: map[string]string{
				"X-Real-IP": "203.0.113.2",
			},
			remoteAddr: "192.168.1.1:8080",
			expected:   "203.0.113.2",
		},
		{
			name: "CF-Connecting-IP header present",
			headers: map[string]string{
				"CF-Connecting-IP": "203.0.113.3",
			},
			remoteAddr: "192.168.1.1:8080",
			expected:   "192.168.1.1",
		},
		{
			name: "no IP headers, use remote address",
			headers: map[string]string{
				"User-Agent": "Mozilla/5.0",
			},
			remoteAddr: "192.168.1.100:8080",
			expected:   "192.168.1.100",
		},
		{
			name:       "remote address without port",
			headers:    map[string]string{},
			remoteAddr: "192.168.1.200",
			expected:   "",
		},
		{
			name:       "empty remote address",
			headers:    map[string]string{},
			remoteAddr: "",
			expected:   "",
		},
		{
			name:       "invalid remote address format",
			headers:    map[string]string{},
			remoteAddr: "invalid-address",
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := iputil.GetClientIP(tt.headers, tt.remoteAddr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIPUtil_Bytes2IntIPv6(t *testing.T) {
	// bytes2IntIPv6はプライベートメソッドのため、IPv6アドレスの検証を通じて間接的にテストします
	t.Run("IPv6 address validation", func(t *testing.T) {
		// 有効なIPv6アドレス
		validIPv6 := []string{
			"2001:db8::1",
			"2001:0db8:0000:0000:0000:0000:0000:0001",
			"::1",
			"fe80::1",
			"fd00::1",
			"2001:db8:1234:5678:9abc:def0:1234:5678",
		}

		for _, ip := range validIPv6 {
			t.Run(ip, func(t *testing.T) {
				assert.True(t, iputil.IsValidIP(ip))
				assert.True(t, iputil.IsIPv6(ip))
				assert.Equal(t, 6, iputil.GetIPVersion(ip))
			})
		}
	})

	t.Run("IPv6 private address detection", func(t *testing.T) {
		// プライベートIPv6アドレス
		privateIPv6 := []string{
			"fe80::1",
			"fd00::1",
			"::1",
		}

		for _, ip := range privateIPv6 {
			t.Run(ip, func(t *testing.T) {
				assert.True(t, iputil.IsValidIP(ip))
				assert.True(t, iputil.IsIPv6(ip))
				assert.True(t, iputil.IsPrivateIP(ip))
			})
		}
	})
}
