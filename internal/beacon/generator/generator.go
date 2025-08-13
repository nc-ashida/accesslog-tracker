package generator

import (
	"time"
	"accesslog-tracker/internal/utils/crypto"
)

// Beacon はトラッキングビーコンの構造体です
type Beacon struct {
	AppID     int       `json:"app_id"`
	SessionID string    `json:"session_id"`
	URL       string    `json:"url"`
	Referrer  string    `json:"referrer"`
	UserAgent string    `json:"user_agent"`
	IPAddress string    `json:"ip_address"`
	Timestamp time.Time `json:"timestamp"`
}

// BeaconGenerator はビーコン生成器の構造体です
type BeaconGenerator struct{}

// NewBeaconGenerator は新しいビーコン生成器を作成します
func NewBeaconGenerator() *BeaconGenerator {
	return &BeaconGenerator{}
}

// GenerateBeacon は新しいビーコンを生成します
func (bg *BeaconGenerator) GenerateBeacon(appID int, sessionID, url, referrer, userAgent, ipAddress string) *Beacon {
	if sessionID == "" {
		sessionID = "alt_" + time.Now().Format("20060102150405") + "_" + crypto.GenerateRandomString(9)
	}

	return &Beacon{
		AppID:     appID,
		SessionID: sessionID,
		URL:       url,
		Referrer:  referrer,
		UserAgent: userAgent,
		IPAddress: ipAddress,
		Timestamp: time.Now(),
	}
}
