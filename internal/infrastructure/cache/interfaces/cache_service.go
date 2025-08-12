package interfaces

import (
	"context"
	"time"
)

// CacheService はキャッシュ機能を提供するインターフェース
type CacheService interface {
	// 基本的なキー・バリュー操作
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	
	// ハッシュ操作
	HGet(ctx context.Context, key, field string) (string, error)
	HSet(ctx context.Context, key, field, value string) error
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HDel(ctx context.Context, key string, fields ...string) error
	
	// リスト操作
	LPush(ctx context.Context, key string, values ...string) error
	RPush(ctx context.Context, key string, values ...string) error
	LPop(ctx context.Context, key string) (string, error)
	RPop(ctx context.Context, key string) (string, error)
	LRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	
	// セット操作
	SAdd(ctx context.Context, key string, members ...string) error
	SRem(ctx context.Context, key string, members ...string) error
	SMembers(ctx context.Context, key string) ([]string, error)
	SIsMember(ctx context.Context, key, member string) (bool, error)
	
	// ソート済みセット操作
	ZAdd(ctx context.Context, key string, score float64, member string) error
	ZRange(ctx context.Context, key string, start, stop int64) ([]string, error)
	ZRangeWithScores(ctx context.Context, key string, start, stop int64) (map[string]float64, error)
	ZRem(ctx context.Context, key string, members ...string) error
	
	// カウンター操作
	Incr(ctx context.Context, key string) (int64, error)
	IncrBy(ctx context.Context, key string, value int64) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)
	DecrBy(ctx context.Context, key string, value int64) (int64, error)
	
	// 有効期限操作
	Expire(ctx context.Context, key string, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	
	// パターンマッチング
	Keys(ctx context.Context, pattern string) ([]string, error)
	
	// バッチ操作
	Pipeline() Pipeline
	
	// 接続管理
	Ping(ctx context.Context) error
	Close() error
}

// Pipeline はRedisパイプライン操作を提供するインターフェース
type Pipeline interface {
	Get(key string) *Result
	Set(key, value string, expiration time.Duration) *Result
	Delete(key string) *Result
	Incr(key string) *Result
	IncrBy(key string, value int64) *Result
	Expire(key string, expiration time.Duration) *Result
	Exec(ctx context.Context) error
}

// Result はパイプライン操作の結果を表す
type Result struct {
	Value interface{}
	Err   error
}

// セッションキャッシュ用の定数
const (
	// セッション関連のキー
	SessionKeyPrefix     = "session:"
	VisitorKeyPrefix     = "visitor:"
	ApplicationKeyPrefix = "app:"
	
	// 統計関連のキー
	StatsKeyPrefix       = "stats:"
	PageViewKeyPrefix    = "pageview:"
	ReferrerKeyPrefix    = "referrer:"
	DeviceKeyPrefix      = "device:"
	CountryKeyPrefix     = "country:"
	
	// レート制限関連のキー
	RateLimitKeyPrefix   = "ratelimit:"
	
	// デフォルト有効期限
	DefaultSessionTTL    = 30 * time.Minute
	DefaultStatsTTL      = 1 * time.Hour
	DefaultRateLimitTTL  = 1 * time.Minute
)

// セッションキャッシュ用のヘルパー関数
func SessionKey(sessionID string) string {
	return SessionKeyPrefix + sessionID
}

func VisitorKey(visitorID string) string {
	return VisitorKeyPrefix + visitorID
}

func ApplicationKey(applicationID string) string {
	return ApplicationKeyPrefix + applicationID
}

func StatsKey(applicationID, period string) string {
	return StatsKeyPrefix + applicationID + ":" + period
}

func PageViewKey(applicationID string) string {
	return PageViewKeyPrefix + applicationID
}

func ReferrerKey(applicationID string) string {
	return ReferrerKeyPrefix + applicationID
}

func DeviceKey(applicationID string) string {
	return DeviceKeyPrefix + applicationID
}

func CountryKey(applicationID string) string {
	return CountryKeyPrefix + applicationID
}

func RateLimitKey(identifier string) string {
	return RateLimitKeyPrefix + identifier
}
