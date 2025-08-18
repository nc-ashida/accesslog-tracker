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
	apimodels "accesslog-tracker/internal/api/models"
	domainmodels "accesslog-tracker/internal/domain/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTest() (*gin.Engine, *MockApplicationService, *MockLogger, *handlers.ApplicationHandler) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockService := new(MockApplicationService)
	mockLogger := new(MockLogger)
	handler := handlers.NewApplicationHandler(mockService, mockLogger)

	return router, mockService, mockLogger, handler
}

func TestApplicationHandler_Create_Success(t *testing.T) {
	router, mockService, mockLogger, handler := setupTest()

	// テストデータ
	reqBody := apimodels.ApplicationRequest{
		Name:        "Test App",
		Description: "Test Description",
		Domain:      "test.com",
	}

	// モックの設定
	mockService.On("Create", mock.Anything, mock.AnythingOfType("*domainmodels.Application")).Return(nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	// リクエストの作成
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/applications", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// ハンドラーの実行
	router.POST("/applications", handler.Create)
	router.ServeHTTP(w, req)

	// アサーション
	assert.Equal(t, http.StatusCreated, w.Code)

	var response apimodels.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestApplicationHandler_Create_InvalidRequest(t *testing.T) {
	router, _, mockLogger, handler := setupTest()

	// 無効なJSONリクエスト
	req := httptest.NewRequest("POST", "/applications", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything)

	router.POST("/applications", handler.Create)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response apimodels.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)

	mockLogger.AssertExpectations(t)
}

func TestApplicationHandler_Create_ServiceError(t *testing.T) {
	router, mockService, mockLogger, handler := setupTest()

	reqBody := apimodels.ApplicationRequest{
		Name:        "Test App",
		Description: "Test Description",
		Domain:      "test.com",
	}

	mockService.On("Create", mock.Anything, mock.AnythingOfType("*domainmodels.Application")).Return(errors.New("database error"))
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything)

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/applications", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.POST("/applications", handler.Create)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response apimodels.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", response.Error.Code)

	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestApplicationHandler_Get_Success(t *testing.T) {
	router, mockService, _, handler := setupTest()

	mockService.On("GetByID", mock.Anything, "test-app-id").Return(&domainmodels.Application{
		AppID:       "test-app-id",
		Name:        "Test App",
		Description: "Test Description",
		Domain:      "test.com",
		APIKey:      "test-api-key",
		Active:      true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil)

	req := httptest.NewRequest("GET", "/applications/test-app-id", nil)
	w := httptest.NewRecorder()

	router.GET("/applications/:id", handler.Get)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response apimodels.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	mockService.AssertExpectations(t)
}

func TestApplicationHandler_Get_MissingID(t *testing.T) {
	router, _, _, handler := setupTest()

	req := httptest.NewRequest("GET", "/applications/", nil)
	w := httptest.NewRecorder()

	router.GET("/applications/:id", handler.Get)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response apimodels.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
}

func TestApplicationHandler_Get_NotFound(t *testing.T) {
	router, mockService, mockLogger, handler := setupTest()

	mockService.On("GetByID", mock.Anything, "non-existent").Return(nil, errors.New("application not found"))
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	req := httptest.NewRequest("GET", "/applications/non-existent", nil)
	w := httptest.NewRecorder()

	router.GET("/applications/:id", handler.Get)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response apimodels.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "NOT_FOUND", response.Error.Code)

	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestApplicationHandler_Update_Success(t *testing.T) {
	router, mockService, mockLogger, handler := setupTest()

	reqBody := apimodels.ApplicationUpdateRequest{
		Name:        "Updated App",
		Description: "Updated Description",
		Domain:      "updated.com",
		Active:      &[]bool{true}[0],
	}

	existingApp := &domainmodels.Application{
		AppID:       "test-app-id",
		Name:        "Old Name",
		Description: "Old Description",
		Domain:      "old.com",
		APIKey:      "test-api-key",
		Active:      true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockService.On("GetByID", mock.Anything, "test-app-id").Return(existingApp, nil)
	mockService.On("Update", mock.Anything, mock.AnythingOfType("*domainmodels.Application")).Return(nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything)

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/applications/test-app-id", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.PUT("/applications/:id", handler.Update)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response apimodels.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestApplicationHandler_Update_InvalidRequest(t *testing.T) {
	router, _, mockLogger, handler := setupTest()

	req := httptest.NewRequest("PUT", "/applications/test-app-id", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mockLogger.On("Warn", mock.Anything, mock.Anything, mock.Anything)

	router.PUT("/applications/:id", handler.Update)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response apimodels.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)

	mockLogger.AssertExpectations(t)
}

func TestApplicationHandler_Update_NotFound(t *testing.T) {
	router, mockService, mockLogger, handler := setupTest()

	reqBody := apimodels.ApplicationUpdateRequest{
		Name:        "Updated App",
		Description: "Updated Description",
		Domain:      "updated.com",
	}

	mockService.On("GetByID", mock.Anything, "non-existent").Return(nil, errors.New("application not found"))
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/applications/non-existent", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.PUT("/applications/:id", handler.Update)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response apimodels.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "NOT_FOUND", response.Error.Code)

	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestApplicationHandler_List_Success(t *testing.T) {
	router, mockService, _, handler := setupTest()

	apps := []*domainmodels.Application{
		{
			AppID:       "app1",
			Name:        "App 1",
			Description: "Description 1",
			Domain:      "app1.com",
			APIKey:      "key1",
			Active:      true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			AppID:       "app2",
			Name:        "App 2",
			Description: "Description 2",
			Domain:      "app2.com",
			APIKey:      "key2",
			Active:      false,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	mockService.On("List", mock.Anything, 10, 0).Return(apps, nil)
	mockService.On("Count", mock.Anything).Return(int64(2), nil)

	req := httptest.NewRequest("GET", "/applications", nil)
	w := httptest.NewRecorder()

	router.GET("/applications", handler.List)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response apimodels.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	mockService.AssertExpectations(t)
}

func TestApplicationHandler_List_WithPagination(t *testing.T) {
	router, mockService, _, handler := setupTest()

	apps := []*domainmodels.Application{}

	mockService.On("List", mock.Anything, 5, 10).Return(apps, nil)
	mockService.On("Count", mock.Anything).Return(int64(0), nil)

	req := httptest.NewRequest("GET", "/applications?page=3&page_size=5", nil)
	w := httptest.NewRecorder()

	router.GET("/applications", handler.List)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response apimodels.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	mockService.AssertExpectations(t)
}

func TestApplicationHandler_List_ServiceError(t *testing.T) {
	router, mockService, mockLogger, handler := setupTest()

	mockService.On("List", mock.Anything, 10, 0).Return(nil, errors.New("database error"))
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything)

	req := httptest.NewRequest("GET", "/applications", nil)
	w := httptest.NewRecorder()

	router.GET("/applications", handler.List)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response apimodels.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", response.Error.Code)

	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestApplicationHandler_Delete_Success(t *testing.T) {
	router, mockService, mockLogger, handler := setupTest()

	mockService.On("Delete", mock.Anything, "test-app-id").Return(nil)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything)

	req := httptest.NewRequest("DELETE", "/applications/test-app-id", nil)
	w := httptest.NewRecorder()

	router.DELETE("/applications/:id", handler.Delete)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response apimodels.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestApplicationHandler_Delete_MissingID(t *testing.T) {
	router, _, _, handler := setupTest()

	req := httptest.NewRequest("DELETE", "/applications/", nil)
	w := httptest.NewRecorder()

	router.DELETE("/applications/:id", handler.Delete)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response apimodels.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
}

func TestApplicationHandler_Delete_NotFound(t *testing.T) {
	router, mockService, mockLogger, handler := setupTest()

	mockService.On("Delete", mock.Anything, "non-existent").Return(domainmodels.ErrApplicationNotFound)
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	req := httptest.NewRequest("DELETE", "/applications/non-existent", nil)
	w := httptest.NewRecorder()

	router.DELETE("/applications/:id", handler.Delete)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response apimodels.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "NOT_FOUND", response.Error.Code)

	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestApplicationHandler_Delete_ServiceError(t *testing.T) {
	router, mockService, mockLogger, handler := setupTest()

	mockService.On("Delete", mock.Anything, "test-app-id").Return(errors.New("database error"))
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	req := httptest.NewRequest("DELETE", "/applications/test-app-id", nil)
	w := httptest.NewRecorder()

	router.DELETE("/applications/:id", handler.Delete)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response apimodels.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", response.Error.Code)

	mockService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
