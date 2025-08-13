package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
)

// HashSHA256 は文字列をSHA256ハッシュに変換します
func HashSHA256(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// GenerateRandomString は指定された長さのランダム文字列を生成します
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		// crypto/randを使用してセキュアな乱数を生成
		randomBytes := make([]byte, 1)
		rand.Read(randomBytes)
		b[i] = charset[int(randomBytes[0])%len(charset)]
	}
	return string(b)
}

// GenerateAPIKey は32文字のAPIキーを生成します
func GenerateAPIKey() string {
	return GenerateRandomString(32)
}

// ValidateAPIKey はAPIキーが有効かどうかを判定します
func ValidateAPIKey(apiKey string) bool {
	if len(apiKey) < 16 {
		return false
	}
	
	// 英数字とハイフンのみ許可
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9-]+$`, apiKey)
	return matched
}

// GenerateSecureToken はセキュアなトークンを生成します
func GenerateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// HashPassword はパスワードをハッシュ化します
func HashPassword(password string) string {
	// 実際の実装ではbcryptなどを使用することを推奨
	return HashSHA256(password)
}

// VerifyPassword はパスワードがハッシュと一致するかどうかを確認します
func VerifyPassword(password, hash string) bool {
	return HashPassword(password) == hash
}
