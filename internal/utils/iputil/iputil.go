package iputil

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

// IPVersion はIPバージョンを定義
type IPVersion int

const (
	IPv4 IPVersion = 4
	IPv6 IPVersion = 6
)

// IsValidIP はIPアドレスが有効かどうかを判定します
func IsValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// IsValidIPv4 はIPv4アドレスが有効かどうかを判定します
func IsValidIPv4(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	return parsedIP.To4() != nil
}

// IsValidIPv6 はIPv6アドレスが有効かどうかを判定します
func IsValidIPv6(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	return parsedIP.To4() == nil
}

// GetIPVersion はIPアドレスのバージョンを取得します
func GetIPVersion(ip string) (IPVersion, error) {
	if IsValidIPv4(ip) {
		return IPv4, nil
	}
	if IsValidIPv6(ip) {
		return IPv6, nil
	}
	return 0, fmt.Errorf("invalid IP address: %s", ip)
}

// NormalizeIP はIPアドレスを正規化します
func NormalizeIP(ip string) (string, error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return "", fmt.Errorf("invalid IP address: %s", ip)
	}
	return parsedIP.String(), nil
}

// IsPrivateIP はプライベートIPアドレスかどうかを判定します
func IsPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// IPv4 プライベートアドレス範囲
	privateIPv4Ranges := []struct {
		start net.IP
		end   net.IP
	}{
		{net.ParseIP("10.0.0.0"), net.ParseIP("10.255.255.255")},
		{net.ParseIP("172.16.0.0"), net.ParseIP("172.31.255.255")},
		{net.ParseIP("192.168.0.0"), net.ParseIP("192.168.255.255")},
	}

	// IPv6 プライベートアドレス範囲
	privateIPv6Ranges := []struct {
		start net.IP
		end   net.IP
	}{
		{net.ParseIP("fc00::"), net.ParseIP("fdff:ffff:ffff:ffff:ffff:ffff:ffff:ffff")},
		{net.ParseIP("fe80::"), net.ParseIP("febf:ffff:ffff:ffff:ffff:ffff:ffff:ffff")},
	}

	// IPv4 チェック
	if parsedIP.To4() != nil {
		for _, r := range privateIPv4Ranges {
			if inRange(parsedIP, r.start, r.end) {
				return true
			}
		}
		return false
	}

	// IPv6 チェック
	for _, r := range privateIPv6Ranges {
		if inRange(parsedIP, r.start, r.end) {
			return true
		}
	}

	return false
}

// IsPublicIP はパブリックIPアドレスかどうかを判定します
func IsPublicIP(ip string) bool {
	if !IsValidIP(ip) {
		return false
	}
	return !IsPrivateIP(ip) && !IsLoopbackIP(ip) && !IsLinkLocalIP(ip)
}

// IsLoopbackIP はループバックIPアドレスかどうかを判定します
func IsLoopbackIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// IPv4 ループバック
	if parsedIP.To4() != nil {
		return parsedIP.IsLoopback()
	}

	// IPv6 ループバック
	return parsedIP.IsLoopback()
}

// IsLinkLocalIP はリンクローカルIPアドレスかどうかを判定します
func IsLinkLocalIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// IPv4 リンクローカル
	if parsedIP.To4() != nil {
		return parsedIP.IsLinkLocalUnicast()
	}

	// IPv6 リンクローカル
	return parsedIP.IsLinkLocalUnicast()
}

// ExtractIPFromString は文字列からIPアドレスを抽出します
func ExtractIPFromString(text string) []string {
	ipv4Pattern := `\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`
	ipv6Pattern := `\b(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\b`

	var ips []string

	// IPv4 アドレスを抽出
	re := regexp.MustCompile(ipv4Pattern)
	matches := re.FindAllString(text, -1)
	for _, match := range matches {
		if IsValidIPv4(match) {
			ips = append(ips, match)
		}
	}

	// IPv6 アドレスを抽出
	re = regexp.MustCompile(ipv6Pattern)
	matches = re.FindAllString(text, -1)
	for _, match := range matches {
		if IsValidIPv6(match) {
			ips = append(ips, match)
		}
	}

	return ips
}

// GetClientIP はHTTPリクエストからクライアントIPを取得します
func GetClientIP(headers map[string]string, remoteAddr string) string {
	// X-Forwarded-For ヘッダーをチェック
	if xff := headers["X-Forwarded-For"]; xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if IsValidIP(ip) {
				return ip
			}
		}
	}

	// X-Real-IP ヘッダーをチェック
	if xri := headers["X-Real-IP"]; xri != "" {
		ip := strings.TrimSpace(xri)
		if IsValidIP(ip) {
			return ip
		}
	}

	// X-Client-IP ヘッダーをチェック
	if xci := headers["X-Client-IP"]; xci != "" {
		ip := strings.TrimSpace(xci)
		if IsValidIP(ip) {
			return ip
		}
	}

	// CF-Connecting-IP ヘッダーをチェック（Cloudflare）
	if cfip := headers["CF-Connecting-IP"]; cfip != "" {
		ip := strings.TrimSpace(cfip)
		if IsValidIP(ip) {
			return ip
		}
	}

	// True-Client-IP ヘッダーをチェック（Akamai）
	if tcip := headers["True-Client-IP"]; tcip != "" {
		ip := strings.TrimSpace(tcip)
		if IsValidIP(ip) {
			return ip
		}
	}

	// RemoteAddr からIPを抽出
	if remoteAddr != "" {
		host, _, err := net.SplitHostPort(remoteAddr)
		if err == nil && IsValidIP(host) {
			return host
		}
		// ポートがない場合
		if IsValidIP(remoteAddr) {
			return remoteAddr
		}
	}

	return ""
}

// IsInSubnet はIPアドレスが指定されたサブネットに含まれるかどうかを判定します
func IsInSubnet(ip, subnet string) bool {
	_, network, err := net.ParseCIDR(subnet)
	if err != nil {
		return false
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	return network.Contains(parsedIP)
}

// GetSubnetInfo はサブネットの情報を取得します
func GetSubnetInfo(subnet string) (net.IP, net.IP, net.IP, error) {
	ip, network, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, nil, nil, err
	}

	// ネットワークアドレス
	networkIP := network.IP

	// ブロードキャストアドレス（IPv4の場合）
	var broadcastIP net.IP
	if ip.To4() != nil {
		broadcastIP = make(net.IP, len(networkIP))
		copy(broadcastIP, networkIP)
		for i := range broadcastIP {
			broadcastIP[i] |= ^network.Mask[i]
		}
	}

	return networkIP, ip, broadcastIP, nil
}

// inRange はIPアドレスが指定された範囲内にあるかどうかを判定します
func inRange(ip, start, end net.IP) bool {
	return bytes2Int(ip) >= bytes2Int(start) && bytes2Int(ip) <= bytes2Int(end)
}

// bytes2Int はIPアドレスを整数に変換します
func bytes2Int(ip net.IP) uint32 {
	ip = ip.To4()
	return uint32(ip[0])<<24 + uint32(ip[1])<<16 + uint32(ip[2])<<8 + uint32(ip[3])
}

// Int2Bytes は整数をIPアドレスに変換します
func Int2Bytes(n uint32) net.IP {
	return net.IPv4(byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}

// GetCountryFromIP はIPアドレスから国コードを取得します（実装例）
func GetCountryFromIP(ip string) (string, error) {
	// 実際の実装では、GeoIPデータベース（MaxMind等）を使用
	// ここでは簡易的な実装例を示す
	
	if !IsValidIP(ip) {
		return "", fmt.Errorf("invalid IP address: %s", ip)
	}

	// プライベートIPの場合は "LOCAL" を返す
	if IsPrivateIP(ip) {
		return "LOCAL", nil
	}

	// 実際の実装では、GeoIPデータベースを参照
	// 例: geoip2.Reader.Lookup(ip)
	
	return "UNKNOWN", nil
}

// GetISPFromIP はIPアドレスからISP情報を取得します（実装例）
func GetISPFromIP(ip string) (string, error) {
	// 実際の実装では、GeoIPデータベース（MaxMind等）を使用
	// ここでは簡易的な実装例を示す
	
	if !IsValidIP(ip) {
		return "", fmt.Errorf("invalid IP address: %s", ip)
	}

	// プライベートIPの場合は "LOCAL" を返す
	if IsPrivateIP(ip) {
		return "LOCAL", nil
	}

	// 実際の実装では、GeoIPデータベースを参照
	// 例: geoip2.Reader.Lookup(ip)
	
	return "UNKNOWN", nil
}
