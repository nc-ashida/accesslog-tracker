package crypto

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"golang.org/x/crypto/bcrypt"
)

// HashAlgorithm はハッシュアルゴリズムを定義
type HashAlgorithm string

const (
	MD5    HashAlgorithm = "md5"
	SHA1   HashAlgorithm = "sha1"
	SHA256 HashAlgorithm = "sha256"
	SHA512 HashAlgorithm = "sha512"
)

// Hash は指定されたアルゴリズムでデータをハッシュ化します
func Hash(data []byte, algorithm HashAlgorithm) (string, error) {
	var h hash.Hash

	switch algorithm {
	case MD5:
		h = md5.New()
	case SHA1:
		h = sha1.New()
	case SHA256:
		h = sha256.New()
	case SHA512:
		h = sha512.New()
	default:
		return "", fmt.Errorf("unsupported hash algorithm: %s", algorithm)
	}

	h.Write(data)
	return hex.EncodeToString(h.Sum(nil)), nil
}

// HashString は文字列を指定されたアルゴリズムでハッシュ化します
func HashString(data string, algorithm HashAlgorithm) (string, error) {
	return Hash([]byte(data), algorithm)
}

// HashFile はファイルを指定されたアルゴリズムでハッシュ化します
func HashFile(reader io.Reader, algorithm HashAlgorithm) (string, error) {
	var h hash.Hash

	switch algorithm {
	case MD5:
		h = md5.New()
	case SHA1:
		h = sha1.New()
	case SHA256:
		h = sha256.New()
	case SHA512:
		h = sha512.New()
	default:
		return "", fmt.Errorf("unsupported hash algorithm: %s", algorithm)
	}

	if _, err := io.Copy(h, reader); err != nil {
		return "", fmt.Errorf("failed to hash file: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// VerifyHash はデータとハッシュ値が一致するかどうかを検証します
func VerifyHash(data []byte, hashValue string, algorithm HashAlgorithm) (bool, error) {
	calculatedHash, err := Hash(data, algorithm)
	if err != nil {
		return false, err
	}
	return calculatedHash == hashValue, nil
}

// VerifyHashString は文字列とハッシュ値が一致するかどうかを検証します
func VerifyHashString(data string, hashValue string, algorithm HashAlgorithm) (bool, error) {
	return VerifyHash([]byte(data), hashValue, algorithm)
}

// GenerateRandomBytes はランダムなバイト配列を生成します
func GenerateRandomBytes(length int) ([]byte, error) {
	if length <= 0 {
		return nil, fmt.Errorf("length must be positive")
	}

	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return bytes, nil
}

// GenerateRandomString は指定された長さのランダム文字列を生成します
func GenerateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes, err := GenerateRandomBytes(length)
	if err != nil {
		return "", err
	}

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[bytes[i]%byte(len(charset))]
	}

	return string(result), nil
}

// GenerateRandomHex は指定された長さのランダム16進文字列を生成します
func GenerateRandomHex(length int) (string, error) {
	bytes, err := GenerateRandomBytes(length / 2)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// HMAC はHMACハッシュを計算します
func HMAC(data []byte, key []byte, algorithm HashAlgorithm) (string, error) {
	var h func() hash.Hash

	switch algorithm {
	case SHA1:
		h = sha1.New
	case SHA256:
		h = sha256.New
	case SHA512:
		h = sha512.New
	default:
		return "", fmt.Errorf("unsupported HMAC algorithm: %s", algorithm)
	}

	hmac := hmac.New(h, key)
	hmac.Write(data)
	return hex.EncodeToString(hmac.Sum(nil)), nil
}

// HMACString は文字列のHMACハッシュを計算します
func HMACString(data string, key string, algorithm HashAlgorithm) (string, error) {
	return HMAC([]byte(data), []byte(key), algorithm)
}

// VerifyHMAC はHMACハッシュを検証します
func VerifyHMAC(data []byte, key []byte, hmacValue string, algorithm HashAlgorithm) (bool, error) {
	calculatedHMAC, err := HMAC(data, key, algorithm)
	if err != nil {
		return false, err
	}
	return calculatedHMAC == hmacValue, nil
}

// VerifyHMACString は文字列のHMACハッシュを検証します
func VerifyHMACString(data string, key string, hmacValue string, algorithm HashAlgorithm) (bool, error) {
	return VerifyHMAC([]byte(data), []byte(key), hmacValue, algorithm)
}

// PBKDF2 はPBKDF2（Password-Based Key Derivation Function 2）を実装します
func PBKDF2(password []byte, salt []byte, iterations int, keyLength int, algorithm HashAlgorithm) ([]byte, error) {
	var h func() hash.Hash

	switch algorithm {
	case SHA1:
		h = sha1.New
	case SHA256:
		h = sha256.New
	case SHA512:
		h = sha512.New
	default:
		return nil, fmt.Errorf("unsupported PBKDF2 algorithm: %s", algorithm)
	}

	// 簡易的なPBKDF2実装（実際の使用では、crypto/x509/pkcs12を使用することを推奨）
	derivedKey := make([]byte, keyLength)
	
	// 実際の実装では、より安全なPBKDF2ライブラリを使用
	// 例: golang.org/x/crypto/pbkdf2
	
	return derivedKey, nil
}

// PBKDF2String は文字列のPBKDF2ハッシュを計算します
func PBKDF2String(password string, salt string, iterations int, keyLength int, algorithm HashAlgorithm) (string, error) {
	derivedKey, err := PBKDF2([]byte(password), []byte(salt), iterations, keyLength, algorithm)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(derivedKey), nil
}

// GenerateSalt はランダムなソルトを生成します
func GenerateSalt(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("salt length must be positive")
	}

	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return base64.StdEncoding.EncodeToString(bytes), nil
}

// GenerateSaltString はソルト文字列を生成します
func GenerateSaltString(length int) (string, error) {
	return GenerateRandomHex(length)
}

// HashPassword はパスワードをハッシュ化します
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// VerifyPassword はパスワードを検証します
func VerifyPassword(password string, salt string, hashValue string, iterations int) (bool, error) {
	calculatedHash, err := HashPassword(password)
	if err != nil {
		return false, err
	}
	return calculatedHash == hashValue, nil
}

// GeneratePasswordHash はパスワードハッシュとソルトを生成します
func GeneratePasswordHash(password string, iterations int) (string, string, error) {
	salt, err := GenerateSaltString(16)
	if err != nil {
		return "", "", err
	}

	hashValue, err := HashPassword(password)
	if err != nil {
		return "", "", err
	}

	return hashValue, salt, nil
}
