package iputil

import (
	"net"
	"strings"
)

// IsValidIP はIPアドレスが有効かどうかを判定します
func IsValidIP(ip string) bool {
	if ip == "" {
		return false
	}
	
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil
}

// IsPrivateIP はIPアドレスがプライベートIPかどうかを判定します
func IsPrivateIP(ip string) bool {
	if !IsValidIP(ip) {
		return false
	}
	
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	
	// プライベートIP範囲のチェック
	privateRanges := []struct {
		start net.IP
		end   net.IP
	}{
		{net.ParseIP("10.0.0.0"), net.ParseIP("10.255.255.255")},
		{net.ParseIP("172.16.0.0"), net.ParseIP("172.31.255.255")},
		{net.ParseIP("192.168.0.0"), net.ParseIP("192.168.255.255")},
		{net.ParseIP("127.0.0.0"), net.ParseIP("127.255.255.255")},
		{net.ParseIP("169.254.0.0"), net.ParseIP("169.254.255.255")},
		{net.ParseIP("::1"), net.ParseIP("::1")}, // IPv6 localhost
		{net.ParseIP("fe80::"), net.ParseIP("febf::")}, // IPv6 link-local
		{net.ParseIP("fc00::"), net.ParseIP("fdff::")}, // IPv6 unique local
	}
	
	for _, r := range privateRanges {
		if inRange(parsedIP, r.start, r.end) {
			return true
		}
	}
	
	return false
}

// ExtractIPFromHeader はHTTPヘッダーからIPアドレスを抽出します
func ExtractIPFromHeader(headers map[string]string) string {
	// X-Forwarded-Forヘッダーを優先的にチェック
	if xff := headers["X-Forwarded-For"]; xff != "" {
		// カンマ区切りの最初のIPを取得
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if IsValidIP(ip) {
				return ip
			}
		}
	}
	
	// X-Real-IPヘッダーをチェック
	if xri := headers["X-Real-IP"]; xri != "" {
		ip := strings.TrimSpace(xri)
		if IsValidIP(ip) {
			return ip
		}
	}
	
	// X-Client-IPヘッダーをチェック
	if xci := headers["X-Client-IP"]; xci != "" {
		ip := strings.TrimSpace(xci)
		if IsValidIP(ip) {
			return ip
		}
	}
	
	return ""
}

// GetClientIP はHTTPリクエストからクライアントIPを取得します
func GetClientIP(headers map[string]string, remoteAddr string) string {
	// ヘッダーからIPを抽出
	if ip := ExtractIPFromHeader(headers); ip != "" {
		return ip
	}
	
	// リモートアドレスからIPを抽出
	if remoteAddr != "" {
		// "ip:port"形式からIPを抽出
		if colonIndex := strings.LastIndex(remoteAddr, ":"); colonIndex != -1 {
			ip := remoteAddr[:colonIndex]
			if IsValidIP(ip) {
				return ip
			}
		}
	}
	
	return ""
}

// AnonymizeIP はIPアドレスを匿名化します
func AnonymizeIP(ip string) string {
	if !IsValidIP(ip) {
		return ip
	}
	
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return ip
	}
	
	// IPv4の場合
	if ipv4 := parsedIP.To4(); ipv4 != nil {
		// 最後のオクテットを0にする
		ipv4[3] = 0
		return ipv4.String()
	}
	
	// IPv6の場合
	if ipv6 := parsedIP.To16(); ipv6 != nil {
		// 最後の64ビットを0にする
		for i := 8; i < 16; i++ {
			ipv6[i] = 0
		}
		return ipv6.String()
	}
	
	return ip
}

// GetIPVersion はIPアドレスのバージョンを返します
func GetIPVersion(ip string) int {
	if !IsValidIP(ip) {
		return 0
	}
	
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return 0
	}
	
	if parsedIP.To4() != nil {
		return 4
	}
	
	return 6
}

// IsIPv4 はIPアドレスがIPv4かどうかを判定します
func IsIPv4(ip string) bool {
	return GetIPVersion(ip) == 4
}

// IsIPv6 はIPアドレスがIPv6かどうかを判定します
func IsIPv6(ip string) bool {
	return GetIPVersion(ip) == 6
}

// inRange はIPアドレスが指定された範囲内かどうかを判定します
func inRange(ip, start, end net.IP) bool {
	// IPv4の場合
	if ip.To4() != nil && start.To4() != nil && end.To4() != nil {
		return bytes2Int(ip.To4()) >= bytes2Int(start.To4()) && bytes2Int(ip.To4()) <= bytes2Int(end.To4())
	}
	
	// IPv6の場合
	if ip.To16() != nil && start.To16() != nil && end.To16() != nil {
		// IPv6の場合は文字列比較を使用（より安全）
		ipStr := ip.String()
		startStr := start.String()
		endStr := end.String()
		
		// 特殊なケース: ::1 (localhost)
		if startStr == "::1" && endStr == "::1" {
			return ipStr == "::1"
		}
		
		// プレフィックスベースの比較
		if strings.HasPrefix(startStr, "fe80:") && strings.HasPrefix(endStr, "febf:") {
			return strings.HasPrefix(ipStr, "fe8") || strings.HasPrefix(ipStr, "fe9") || 
				   strings.HasPrefix(ipStr, "fea") || strings.HasPrefix(ipStr, "feb")
		}
		
		if strings.HasPrefix(startStr, "fc00:") && strings.HasPrefix(endStr, "fdff:") {
			return strings.HasPrefix(ipStr, "fc") || strings.HasPrefix(ipStr, "fd")
		}
	}
	
	return false
}

// bytes2Int はIPv4アドレスを整数に変換します
func bytes2Int(ip net.IP) uint32 {
	return uint32(ip[0])<<24 + uint32(ip[1])<<16 + uint32(ip[2])<<8 + uint32(ip[3])
}

// bytes2IntIPv6 はIPv6アドレスの前半64ビットを整数に変換します
func bytes2IntIPv6(ip net.IP) uint64 {
	return uint64(ip[0])<<56 + uint64(ip[1])<<48 + uint64(ip[2])<<40 + uint64(ip[3])<<32 +
		   uint64(ip[4])<<24 + uint64(ip[5])<<16 + uint64(ip[6])<<8 + uint64(ip[7])
}
