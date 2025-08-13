package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS はCORSミドルウェアを設定します
func CORS() gin.HandlerFunc {
	config := cors.DefaultConfig()
	
	// 許可するオリジンを設定
	config.AllowOrigins = []string{
		"http://localhost:3000",
		"http://localhost:8080",
		"https://example.com",
		"https://*.example.com",
	}
	
	// 許可するHTTPメソッドを設定
	config.AllowMethods = []string{
		"GET",
		"POST",
		"PUT",
		"DELETE",
		"OPTIONS",
	}
	
	// 許可するヘッダーを設定
	config.AllowHeaders = []string{
		"Origin",
		"Content-Type",
		"Accept",
		"Authorization",
		"X-API-Key",
		"X-Requested-With",
	}
	
	// 認証情報の送信を許可
	config.AllowCredentials = true
	
	// プリフライトリクエストのキャッシュ時間を設定
	config.MaxAge = 86400 // 24時間
	
	return cors.New(config)
}
