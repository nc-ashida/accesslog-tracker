package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := AuthConfig{
		SecretKey:     "test-secret-key",
		TokenDuration: 1 * time.Hour,
	}

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "No Authorization Header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Authorization header is required"}`,
		},
		{
			name:           "Invalid Authorization Format",
			authHeader:     "InvalidFormat",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid authorization header format"}`,
		},
		{
			name:           "Invalid Token",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid token"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(AuthMiddleware(config))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestAuthMiddlewareValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := AuthConfig{
		SecretKey:     "test-secret-key",
		TokenDuration: 1 * time.Hour,
	}

	// 有効なトークンを生成
	token, err := GenerateToken("test-app", "test-user", config)
	assert.NoError(t, err)

	router := gin.New()
	router.Use(AuthMiddleware(config))
	router.GET("/test", func(c *gin.Context) {
		appID, exists := c.Get("application_id")
		assert.True(t, exists)
		assert.Equal(t, "test-app", appID)

		userID, exists := c.Get("user_id")
		assert.True(t, exists)
		assert.Equal(t, "test-user", userID)

		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, `{"message":"success"}`, w.Body.String())
}

func TestGenerateToken(t *testing.T) {
	config := AuthConfig{
		SecretKey:     "test-secret-key",
		TokenDuration: 1 * time.Hour,
	}

	token, err := GenerateToken("test-app", "test-user", config)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// トークンを検証
	parsedToken, err := jwt.ParseWithClaims(token, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.SecretKey), nil
	})

	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(*JWTClaims)
	assert.True(t, ok)
	assert.Equal(t, "test-app", claims.ApplicationID)
	assert.Equal(t, "test-user", claims.UserID)
}

func TestOptionalAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := AuthConfig{
		SecretKey:     "test-secret-key",
		TokenDuration: 1 * time.Hour,
	}

	router := gin.New()
	router.Use(OptionalAuthMiddleware(config))
	router.GET("/test", func(c *gin.Context) {
		appID, exists := c.Get("application_id")
		if exists {
			assert.Equal(t, "test-app", appID)
		}
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 認証ヘッダーなしでもアクセス可能
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, `{"message":"success"}`, w.Body.String())
}
