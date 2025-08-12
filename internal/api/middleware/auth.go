package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// JWTClaims JWTクレーム構造体
type JWTClaims struct {
	ApplicationID string `json:"application_id"`
	UserID        string `json:"user_id"`
	jwt.RegisteredClaims
}

// AuthConfig 認証設定
type AuthConfig struct {
	SecretKey     string
	TokenDuration time.Duration
}

// APIKeyAuthMiddleware API Key認証ミドルウェア（仕様書準拠）
func APIKeyAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// X-API-KeyヘッダーからAPI Keyを取得
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success":   false,
				"message":   "X-API-Key header is required",
				"timestamp": time.Now().Format(time.RFC3339),
			})
			c.Abort()
			return
		}

		// API Keyの形式をチェック（64文字の英数字）
		if len(apiKey) != 64 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success":   false,
				"message":   "Invalid API key format",
				"timestamp": time.Now().Format(time.RFC3339),
			})
			c.Abort()
			return
		}

		// コンテキストにAPI Keyを設定
		c.Set("api_key", apiKey)
		c.Next()
	}
}

// AuthMiddleware 認証ミドルウェア
func AuthMiddleware(config AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Authorizationヘッダーからトークンを取得
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Bearerトークンの形式をチェック
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// JWTトークンを検証
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.SecretKey), nil
		})

		if err != nil {
			logrus.WithError(err).Error("Failed to parse JWT token")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		// クレームを取得
		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token claims",
			})
			c.Abort()
			return
		}

		// トークンの有効期限をチェック
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token has expired",
			})
			c.Abort()
			return
		}

		// コンテキストにクレーム情報を設定
		c.Set("application_id", claims.ApplicationID)
		c.Set("user_id", claims.UserID)
		c.Set("claims", claims)

		c.Next()
	}
}

// GenerateToken JWTトークンを生成
func GenerateToken(applicationID, userID string, config AuthConfig) (string, error) {
	claims := &JWTClaims{
		ApplicationID: applicationID,
		UserID:        userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.TokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.SecretKey))
}

// OptionalAuthMiddleware オプショナル認証ミドルウェア（認証が失敗しても続行）
func OptionalAuthMiddleware(config AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := tokenParts[1]

		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.SecretKey), nil
		})

		if err != nil || !token.Valid {
			c.Next()
			return
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			c.Next()
			return
		}

		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			c.Next()
			return
		}

		c.Set("application_id", claims.ApplicationID)
		c.Set("user_id", claims.UserID)
		c.Set("claims", claims)

		c.Next()
	}
}
