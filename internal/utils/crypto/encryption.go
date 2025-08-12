package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// EncryptionAlgorithm は暗号化アルゴリズムを定義
type EncryptionAlgorithm string

const (
	AES256 EncryptionAlgorithm = "aes256"
)

// AESKey はAES暗号化キーを管理します
type AESKey struct {
	Key []byte
}

// NewAESKey は新しいAESキーを生成します
func NewAESKey(key []byte) (*AESKey, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("AES key must be 32 bytes, got %d", len(key))
	}
	return &AESKey{Key: key}, nil
}

// NewAESKeyFromString は文字列からAESキーを生成します
func NewAESKeyFromString(keyStr string) (*AESKey, error) {
	key, err := hex.DecodeString(keyStr)
	if err != nil {
		return nil, fmt.Errorf("invalid hex key: %w", err)
	}
	return NewAESKey(key)
}

// GenerateAESKey はランダムなAESキーを生成します
func GenerateAESKey() (*AESKey, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate AES key: %w", err)
	}
	return &AESKey{Key: key}, nil
}

// EncryptAES はAESでデータを暗号化します
func (a *AESKey) EncryptAES(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(a.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// GCMモードを使用
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// ノンスを生成
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// 暗号化
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// DecryptAES はAESでデータを復号化します
func (a *AESKey) DecryptAES(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(a.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// EncryptAESString は文字列をAESで暗号化します
func (a *AESKey) EncryptAESString(plaintext string) (string, error) {
	ciphertext, err := a.EncryptAES([]byte(plaintext))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAESString はAESで暗号化された文字列を復号化します
func (a *AESKey) DecryptAESString(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	plaintext, err := a.DecryptAES(data)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// JWTClaims はJWTクレームを定義します
type JWTClaims struct {
	UserID    string            `json:"user_id"`
	Username  string            `json:"username"`
	Email     string            `json:"email"`
	Roles     []string          `json:"roles"`
	Custom    map[string]string `json:"custom,omitempty"`
	jwt.RegisteredClaims
}

// JWTManager はJWTトークンの管理を行います
type JWTManager struct {
	SecretKey []byte
	Issuer    string
}

// NewJWTManager は新しいJWTマネージャーを作成します
func NewJWTManager(secretKey string, issuer string) *JWTManager {
	return &JWTManager{
		SecretKey: []byte(secretKey),
		Issuer:    issuer,
	}
}

// GenerateToken はJWTトークンを生成します
func (j *JWTManager) GenerateToken(claims *JWTClaims, expiration time.Duration) (string, error) {
	claims.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    j.Issuer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SecretKey)
}

// ValidateToken はJWTトークンを検証します
func (j *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.SecretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RefreshToken はJWTトークンを更新します
func (j *JWTManager) RefreshToken(tokenString string, expiration time.Duration) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// 新しい有効期限でトークンを生成
	return j.GenerateToken(claims, expiration)
}

// GenerateAccessToken はアクセストークンを生成します
func (j *JWTManager) GenerateAccessToken(userID, username, email string, roles []string, custom map[string]string) (string, error) {
	claims := &JWTClaims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Roles:    roles,
		Custom:   custom,
	}

	return j.GenerateToken(claims, 24*time.Hour) // 24時間
}

// GenerateRefreshToken はリフレッシュトークンを生成します
func (j *JWTManager) GenerateRefreshToken(userID string) (string, error) {
	claims := &JWTClaims{
		UserID: userID,
	}

	return j.GenerateToken(claims, 7*24*time.Hour) // 7日間
}

// ExtractUserID はトークンからユーザーIDを抽出します
func (j *JWTManager) ExtractUserID(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

// ExtractRoles はトークンからロールを抽出します
func (j *JWTManager) ExtractRoles(tokenString string) ([]string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	return claims.Roles, nil
}

// HasRole はトークンが指定されたロールを持っているかどうかを判定します
func (j *JWTManager) HasRole(tokenString string, role string) (bool, error) {
	roles, err := j.ExtractRoles(tokenString)
	if err != nil {
		return false, err
	}

	for _, r := range roles {
		if r == role {
			return true, nil
		}
	}
	return false, nil
}

// HasAnyRole はトークンが指定されたロールのいずれかを持っているかどうかを判定します
func (j *JWTManager) HasAnyRole(tokenString string, roles []string) (bool, error) {
	tokenRoles, err := j.ExtractRoles(tokenString)
	if err != nil {
		return false, err
	}

	for _, tokenRole := range tokenRoles {
		for _, role := range roles {
			if tokenRole == role {
				return true, nil
			}
		}
	}
	return false, nil
}

// GenerateAPIKey はAPIキーを生成します
func GenerateAPIKey(prefix string, length int) (string, error) {
	randomPart, err := GenerateRandomString(length - len(prefix))
	if err != nil {
		return "", err
	}
	return prefix + "_" + randomPart, nil
}

// HashAPIKey はAPIキーをハッシュ化します
func HashAPIKey(apiKey string) (string, error) {
	return HashString(apiKey, SHA256)
}

// VerifyAPIKey はAPIキーを検証します
func VerifyAPIKey(apiKey, hashValue string) (bool, error) {
	return VerifyHashString(apiKey, hashValue, SHA256)
}
