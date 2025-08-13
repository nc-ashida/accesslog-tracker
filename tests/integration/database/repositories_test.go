package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nc-ashida/accesslog-tracker/internal/domain/models"
	"github.com/nc-ashida/accesslog-tracker/internal/infrastructure/database/postgresql/repositories"
)

func TestApplicationRepository(t *testing.T) {
	// テスト用データベース接続を設定
	// 実際のテストでは、テスト用のデータベースを使用する
	
	t.Run("Create Application", func(t *testing.T) {
		repo := &repositories.ApplicationRepository{}
		
		app := &models.Application{
			Name:        "Test App",
			Description: "Test Application",
			Domain:      "test.example.com",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		
		err := repo.Create(context.Background(), app)
		require.NoError(t, err)
		assert.NotZero(t, app.ID)
	})

	t.Run("Get Application", func(t *testing.T) {
		repo := &repositories.ApplicationRepository{}
		
		app, err := repo.GetByID(context.Background(), 1)
		require.NoError(t, err)
		assert.NotNil(t, app)
		assert.Equal(t, "Test App", app.Name)
	})

	t.Run("Update Application", func(t *testing.T) {
		repo := &repositories.ApplicationRepository{}
		
		app := &models.Application{
			ID:          1,
			Name:        "Updated Test App",
			Description: "Updated Test Application",
			Domain:      "updated.test.example.com",
			UpdatedAt:   time.Now(),
		}
		
		err := repo.Update(context.Background(), app)
		require.NoError(t, err)
	})

	t.Run("Delete Application", func(t *testing.T) {
		repo := &repositories.ApplicationRepository{}
		
		err := repo.Delete(context.Background(), 1)
		require.NoError(t, err)
	})
}

func TestSessionRepository(t *testing.T) {
	t.Run("Create Session", func(t *testing.T) {
		repo := &repositories.SessionRepository{}
		
		session := &models.Session{
			ApplicationID: 1,
			SessionID:     "test-session-123",
			UserAgent:     "Mozilla/5.0 (Test Browser)",
			IPAddress:     "192.168.1.1",
			CreatedAt:     time.Now(),
		}
		
		err := repo.Create(context.Background(), session)
		require.NoError(t, err)
		assert.NotZero(t, session.ID)
	})

	t.Run("Get Session", func(t *testing.T) {
		repo := &repositories.SessionRepository{}
		
		session, err := repo.GetBySessionID(context.Background(), "test-session-123")
		require.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, "test-session-123", session.SessionID)
	})
}

func TestTrackingRepository(t *testing.T) {
	t.Run("Create Access Log", func(t *testing.T) {
		repo := &repositories.TrackingRepository{}
		
		accessLog := &models.AccessLog{
			SessionID:    1,
			URL:          "/test-page",
			Referrer:     "https://example.com",
			UserAgent:    "Mozilla/5.0 (Test Browser)",
			IPAddress:    "192.168.1.1",
			Timestamp:    time.Now(),
		}
		
		err := repo.CreateAccessLog(context.Background(), accessLog)
		require.NoError(t, err)
		assert.NotZero(t, accessLog.ID)
	})

	t.Run("Get Access Logs by Session", func(t *testing.T) {
		repo := &repositories.TrackingRepository{}
		
		logs, err := repo.GetAccessLogsBySession(context.Background(), 1)
		require.NoError(t, err)
		assert.NotEmpty(t, logs)
	})
}
