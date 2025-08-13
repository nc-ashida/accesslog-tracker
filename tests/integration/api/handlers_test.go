package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nc-ashida/accesslog-tracker/internal/api/handlers"
	"github.com/nc-ashida/accesslog-tracker/internal/domain/models"
)

func TestApplicationHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// テスト用のハンドラーを設定
	appHandler := &handlers.ApplicationHandler{}
	router.POST("/applications", appHandler.Create)
	router.GET("/applications/:id", appHandler.Get)
	router.PUT("/applications/:id", appHandler.Update)
	router.DELETE("/applications/:id", appHandler.Delete)

	t.Run("Create Application", func(t *testing.T) {
		app := models.Application{
			Name:        "Test App",
			Description: "Test Application",
			Domain:      "test.example.com",
		}
		
		body, _ := json.Marshal(app)
		req := httptest.NewRequest("POST", "/applications", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Get Application", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/applications/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestBeaconHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	beaconHandler := &handlers.BeaconHandler{}
	router.GET("/beacon", beaconHandler.Track)

	t.Run("Track Beacon", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/beacon?app_id=1&session_id=test123&url=/test&referrer=https://example.com", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestHealthHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	healthHandler := &handlers.HealthHandler{}
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)

	t.Run("Health Check", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "ok", response["status"])
	})

	t.Run("Readiness Check", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ready", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
