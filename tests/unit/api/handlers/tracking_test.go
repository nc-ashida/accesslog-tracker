package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"accesslog-tracker/internal/api/handlers"
	"accesslog-tracker/internal/api/models"
	"accesslog-tracker/internal/domain/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)



func setupTrackingTest() (*gin.Engine, *MockTrackingService, *MockLogger, *handlers.TrackingHandler) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	mockService := new(MockTrackingService)
	mockLogger := new(MockLogger)
	handler := handlers.NewTrackingHandler(mockService, mockLogger)
	
	return router, mockService, mockLogger, handler
}

func TestTrackingHandler_Track_Success(t *testing.T) {
	router, mockService, mockLogger, handler := setupTrackingTest()
	
	// テストデータ
	reqBody := models.TrackingRequest{
		AppID:       "test-app-id",
		UserAgent:   "Mozilla/5.0 (Test Browser)",
		URL:         "https://test.com/page",
		IPAddress:   "192.168.1.1",
		SessionID:   "test-session-id",
		Referrer:    "https://google.com",
		CustomParams: map[string]interface{}{"utm_source": "test"},
	}
	
	// モックの設定
	mockService.On("ProcessTrackingData", mock.Anything, mock.Anything).Return(nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	
	// リクエストの作成
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/track", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// ハンドラーの実行
	router.POST("/track", func(c *gin.Context) {
		c.Set("app_id", "test-app-id")
		handler.Track(c)
	})
	router.ServeHTTP(w, req)
	
	// アサーション
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestTrackingHandler_Track_InvalidRequest(t *testing.T) {
	router, _, mockLogger, handler := setupTrackingTest()
	
	// 無効なJSONリクエスト
	req := httptest.NewRequest("POST", "/track", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	
	router.POST("/track", handler.Track)
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	
	mockLogger.AssertExpectations(t)
}

func TestTrackingHandler_Track_ServiceError(t *testing.T) {
	router, mockService, mockLogger, handler := setupTrackingTest()
	
	reqBody := models.TrackingRequest{
		AppID:       "test-app-id",
		UserAgent:   "Mozilla/5.0 (Test Browser)",
		URL:         "https://test.com/page",
		IPAddress:   "192.168.1.1",
		SessionID:   "test-session-id",
		Referrer:    "https://google.com",
		CustomParams: map[string]interface{}{"utm_source": "test"},
	}
	
	mockService.On("ProcessTrackingData", mock.Anything, mock.Anything).Return(errors.New("database error"))
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/track", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router.POST("/track", func(c *gin.Context) {
		c.Set("app_id", "test-app-id")
		handler.Track(c)
	})
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", response.Error.Code)
	
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestTrackingHandler_GetStatistics_Success(t *testing.T) {
	router, mockService, mockLogger, handler := setupTrackingTest()
	
	stats := &services.TrackingStatistics{
		AppID:     "test-app-id",
		StartDate: time.Now().AddDate(0, 0, -7),
		EndDate:   time.Now(),
		Metrics: map[string]interface{}{
			"total_tracking_count": int64(100),
		},
	}
	
	mockService.On("GetStatistics", mock.Anything, "test-app-id", mock.Anything, mock.Anything).Return(stats, nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything)
	
	req := httptest.NewRequest("GET", "/statistics?app_id=test-app-id&start_date=2024-01-01&end_date=2024-01-31", nil)
	w := httptest.NewRecorder()
	
	router.GET("/statistics", func(c *gin.Context) {
		c.Set("app_id", "test-app-id")
		handler.GetStatistics(c)
	})
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestTrackingHandler_GetStatistics_MissingParameters(t *testing.T) {
	router, _, _, handler := setupTrackingTest()
	
	req := httptest.NewRequest("GET", "/statistics", nil)
	w := httptest.NewRecorder()
	
	router.GET("/statistics", handler.GetStatistics)
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
}

func TestTrackingHandler_GetStatistics_InvalidDateFormat(t *testing.T) {
	router, _, _, handler := setupTrackingTest()
	
	req := httptest.NewRequest("GET", "/statistics?app_id=test-app-id&start_date=invalid-date&end_date=2024-01-31", nil)
	w := httptest.NewRecorder()
	
	router.GET("/statistics", handler.GetStatistics)
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
}

func TestTrackingHandler_GetStatistics_ServiceError(t *testing.T) {
	router, mockService, mockLogger, handler := setupTrackingTest()
	
	mockService.On("GetStatistics", mock.Anything, "test-app-id", mock.Anything, mock.Anything).Return(nil, errors.New("database error"))
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	
	req := httptest.NewRequest("GET", "/statistics?app_id=test-app-id&start_date=2024-01-01&end_date=2024-01-31", nil)
	w := httptest.NewRecorder()
	
	router.GET("/statistics", func(c *gin.Context) {
		c.Set("app_id", "test-app-id")
		handler.GetStatistics(c)
	})
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", response.Error.Code)
	
	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestTrackingHandler_GetStatistics_AccessDenied(t *testing.T) {
	router, _, _, handler := setupTrackingTest()
	
	req := httptest.NewRequest("GET", "/statistics?app_id=different-app-id&start_date=2024-01-01&end_date=2024-01-31", nil)
	w := httptest.NewRecorder()
	
	router.GET("/statistics", func(c *gin.Context) {
		c.Set("app_id", "test-app-id")
		handler.GetStatistics(c)
	})
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusForbidden, w.Code)
	
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "FORBIDDEN", response.Error.Code)
}
