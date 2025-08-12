package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/your-username/accesslog-tracker/internal/domain/models"
	"github.com/your-username/accesslog-tracker/internal/domain/services"
)

// MockTrackingService モックトラッキングサービス
type MockTrackingService struct {
	mock.Mock
}

func (m *MockTrackingService) Track(data *models.TrackingData) error {
	args := m.Called(data)
	return args.Error(0)
}

// MockApplicationService モックアプリケーションサービス
type MockApplicationService struct {
	mock.Mock
}

func (m *MockApplicationService) GetByID(id string) (*models.Application, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Application), args.Error(1)
}

func TestTrackHandler_Success(t *testing.T) {
	// テストケース: 正常なトラッキングリクエスト
	gin.SetMode(gin.TestMode)
	
	mockTrackingService := new(MockTrackingService)
	mockApplicationService := new(MockApplicationService)
	
	handler := &TrackingHandler{
		trackingService:    mockTrackingService,
		applicationService: mockApplicationService,
	}
	
	// テストデータ
	requestData := TrackRequest{
		AppID:         "test-app-id",
		ClientSubID:   "test-client-sub-id",
		ModuleID:      "test-module-id",
		URL:           "https://example.com",
		Referrer:      "https://google.com",
		UserAgent:     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		IPAddress:     "192.168.1.1",
		SessionID:     "test-session-id",
		ScreenResolution: "1920x1080",
		Language:      "ja-JP",
		Timezone:      "Asia/Tokyo",
		CustomParams: map[string]interface{}{
			"page_type": "product",
			"product_id": "12345",
		},
	}
	
	// モックの設定
	mockApplication := &models.Application{
		ID:     "test-app-id",
		AppID:  "test-app-id",
		Status: "active",
	}
	
	mockApplicationService.On("GetByID", "test-app-id").Return(mockApplication, nil)
	mockTrackingService.On("Track", mock.AnythingOfType("*models.TrackingData")).Return(nil)
	
	// リクエストの作成
	jsonData, _ := json.Marshal(requestData)
	req, _ := http.NewRequest("POST", "/api/v1/track", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-api-key-64-chars-long-12345678901234567890123456789012")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	// ハンドラーを実行
	handler.Track(c)
	
	// アサーション
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response TrackResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Tracking data saved successfully", response.Message)
	
	// モックの検証
	mockApplicationService.AssertExpectations(t)
	mockTrackingService.AssertExpectations(t)
}

func TestTrackHandler_InvalidRequest(t *testing.T) {
	// テストケース: 無効なリクエスト
	gin.SetMode(gin.TestMode)
	
	mockTrackingService := new(MockTrackingService)
	mockApplicationService := new(MockApplicationService)
	
	handler := &TrackingHandler{
		trackingService:    mockTrackingService,
		applicationService: mockApplicationService,
	}
	
	// 無効なリクエストデータ（app_idが不足）
	requestData := map[string]interface{}{
		"user_agent": "Mozilla/5.0",
	}
	
	jsonData, _ := json.Marshal(requestData)
	req, _ := http.NewRequest("POST", "/api/v1/track", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-api-key-64-chars-long-12345678901234567890123456789012")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	// ハンドラーを実行
	handler.Track(c)
	
	// アサーション
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid request format", response["error"])
}

func TestTrackHandler_ApplicationNotFound(t *testing.T) {
	// テストケース: アプリケーションが見つからない
	gin.SetMode(gin.TestMode)
	
	mockTrackingService := new(MockTrackingService)
	mockApplicationService := new(MockApplicationService)
	
	handler := &TrackingHandler{
		trackingService:    mockTrackingService,
		applicationService: mockApplicationService,
	}
	
	requestData := TrackRequest{
		AppID:     "non-existent-app",
		UserAgent: "Mozilla/5.0",
	}
	
	// モックの設定
	mockApplicationService.On("GetByID", "non-existent-app").Return((*models.Application)(nil), assert.AnError)
	
	jsonData, _ := json.Marshal(requestData)
	req, _ := http.NewRequest("POST", "/api/v1/track", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-api-key-64-chars-long-12345678901234567890123456789012")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	
	// ハンドラーを実行
	handler.Track(c)
	
	// アサーション
	assert.Equal(t, http.StatusNotFound, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Application not found", response["error"])
	
	mockApplicationService.AssertExpectations(t)
}
