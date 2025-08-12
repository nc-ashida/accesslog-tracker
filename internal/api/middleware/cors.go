package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

// CORSConfig CORS設定
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           time.Duration
}

// DefaultCORSConfig デフォルトCORS設定
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{
			"http://localhost:3000",
			"http://localhost:8080",
			"https://yourdomain.com",
		},
		AllowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
		},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"X-API-Key",
		},
		ExposedHeaders: []string{
			"Content-Length",
			"X-Total-Count",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
}

// CORS CORSミドルウェア
func CORS(config CORSConfig) gin.HandlerFunc {
	corsConfig := cors.Config{
		AllowOrigins:     config.AllowedOrigins,
		AllowMethods:     config.AllowedMethods,
		AllowHeaders:     config.AllowedHeaders,
		ExposeHeaders:    config.ExposedHeaders,
		AllowCredentials: config.AllowCredentials,
		MaxAge:           config.MaxAge,
	}

	return cors.New(corsConfig)
}

// TrackingCORS トラッキング用CORS設定（より緩い設定）
func TrackingCORS() gin.HandlerFunc {
	config := CORSConfig{
		AllowedOrigins: []string{"*"}, // トラッキングは全てのオリジンを許可
		AllowedMethods: []string{
			"GET",
			"POST",
			"OPTIONS",
		},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"User-Agent",
			"Referer",
		},
		ExposedHeaders: []string{
			"Content-Length",
		},
		AllowCredentials: false, // トラッキングではクレデンシャルは不要
		MaxAge:           1 * time.Hour,
	}

	return CORS(config)
}
