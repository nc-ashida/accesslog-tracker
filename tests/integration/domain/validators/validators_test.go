package validators_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"accesslog-tracker/internal/domain/models"
	"accesslog-tracker/internal/domain/validators"
)

func TestValidators_Integration(t *testing.T) {
	t.Run("ApplicationValidator", func(t *testing.T) {
		validator := validators.NewApplicationValidator()

		t.Run("Validate", func(t *testing.T) {
			// 有効なアプリケーション
			app := &models.Application{
				AppID:       "test-app",
				Name:        "Test Application",
				Description: "Test application for validation",
				Domain:      "test.example.com",
				APIKey:      "test-api-key-123456789012345678901234567890",
				Active:      true,
			}

			err := validator.Validate(app)
			assert.NoError(t, err)

			// 無効なアプリケーション（AppIDが空）
			invalidApp := &models.Application{
				AppID:       "",
				Name:        "Test Application",
				Description: "Test application for validation",
				Domain:      "test.example.com",
				APIKey:      "test-api-key-123456789012345678901234567890",
				Active:      true,
			}

			err = validator.Validate(invalidApp)
			assert.Error(t, err)
		})

		t.Run("ValidateCreate", func(t *testing.T) {
			// 有効な作成リクエスト
			createRequest := &models.Application{
				Name:        "New Application",
				Description: "New application for testing",
				Domain:      "new.example.com",
			}

			err := validator.ValidateCreate(createRequest)
			assert.NoError(t, err)

			// 無効な作成リクエスト（名前が空）
			invalidRequest := &models.Application{
				Name:        "",
				Description: "New application for testing",
				Domain:      "new.example.com",
			}

			err = validator.ValidateCreate(invalidRequest)
			assert.Error(t, err)
		})

		t.Run("ValidateUpdate", func(t *testing.T) {
			// 有効な更新リクエスト
			updateRequest := &models.Application{
				AppID:       "existing-app",
				Name:        "Updated Application",
				Description: "Updated application for testing",
				Domain:      "updated.example.com",
			}

			err := validator.ValidateUpdate(updateRequest)
			assert.NoError(t, err)

			// 無効な更新リクエスト（AppIDが空）
			invalidRequest := &models.Application{
				AppID:       "",
				Name:        "Updated Application",
				Description: "Updated application for testing",
				Domain:      "updated.example.com",
			}

			err = validator.ValidateUpdate(invalidRequest)
			assert.Error(t, err)
		})

		t.Run("ValidateAPIKey", func(t *testing.T) {
			// 有効なAPIキー
			validAPIKey := "test-api-key-123456789012345678901234567890"
			err := validator.ValidateAPIKey(validAPIKey)
			assert.NoError(t, err)

			// 無効なAPIキー（短すぎる）
			invalidAPIKey := "short"
			err = validator.ValidateAPIKey(invalidAPIKey)
			assert.Error(t, err)
		})

		t.Run("ValidateDomain", func(t *testing.T) {
			// 有効なドメイン
			validDomains := []string{
				"example.com",
				"sub.example.com",
				"test-domain.org",
				"localhost",
			}

			for _, domain := range validDomains {
				err := validator.ValidateDomain(domain)
				assert.NoError(t, err, "Domain %s should be valid", domain)
			}

			// 無効なドメイン
			invalidDomains := []string{
				"",
				"invalid-domain", // 特殊なケースで無効
			}

			for _, domain := range invalidDomains {
				err := validator.ValidateDomain(domain)
				assert.Error(t, err, "Domain %s should be invalid", domain)
			}
		})

		t.Run("ValidateName", func(t *testing.T) {
			// 有効な名前
			validNames := []string{
				"Test Application",
				"My App",
				"Application 123",
			}

			for _, name := range validNames {
				err := validator.ValidateName(name)
				assert.NoError(t, err, "Name %s should be valid", name)
			}

			// 無効な名前
			invalidNames := []string{
				"",
				"a", // 短すぎる
			}

			for _, name := range invalidNames {
				err := validator.ValidateName(name)
				assert.Error(t, err, "Name %s should be invalid", name)
			}
		})

		t.Run("ValidateAppID", func(t *testing.T) {
			// 有効なAppID
			validAppIDs := []string{
				"testapp123", // 8文字以上で文字で始まる
				"my_application",
				"app123456",
			}

			for _, appID := range validAppIDs {
				err := validator.ValidateAppID(appID)
				assert.NoError(t, err, "AppID %s should be valid", appID)
			}

			// 無効なAppID
			invalidAppIDs := []string{
				"",
				"test-app", // ハイフンは無効
				"app123", // 8文字未満
				"invalid app id", // スペースが含まれている
			}

			for _, appID := range invalidAppIDs {
				err := validator.ValidateAppID(appID)
				assert.Error(t, err, "AppID %s should be invalid", appID)
			}
		})
	})

	t.Run("TrackingValidator", func(t *testing.T) {
		validator := validators.NewTrackingValidator()

		t.Run("Validate", func(t *testing.T) {
			// 有効なトラッキングデータ
			tracking := &models.TrackingData{
				AppID:     "testapp123",
				UserAgent: "Mozilla/5.0 (Test Browser)",
				IPAddress: "192.168.1.1",
				URL:       "https://example.com/page",
				Timestamp: time.Now(),
			}

			err := validator.Validate(tracking)
			assert.NoError(t, err)

			// 無効なトラッキングデータ（AppIDが空）
			invalidTracking := &models.TrackingData{
				AppID:     "",
				UserAgent: "Mozilla/5.0 (Test Browser)",
				IPAddress: "192.168.1.1",
				URL:       "https://example.com/page",
				Timestamp: time.Now(),
			}

			err = validator.Validate(invalidTracking)
			assert.Error(t, err)
		})

		t.Run("ValidateAppID", func(t *testing.T) {
			// 有効なAppID
			validAppID := "testapp123"
			err := validator.ValidateAppID(validAppID)
			assert.NoError(t, err)

			// 無効なAppID
			invalidAppID := ""
			err = validator.ValidateAppID(invalidAppID)
			assert.Error(t, err)
		})

		t.Run("ValidateUserAgent", func(t *testing.T) {
			// 有効なUser-Agent
			validUserAgents := []string{
				"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
				"Mozilla/5.0 (iPhone; CPU iPhone OS 14_7_1 like Mac OS X) AppleWebKit/605.1.15",
				"Googlebot/2.1 (+http://www.google.com/bot.html)",
			}

			for _, ua := range validUserAgents {
				err := validator.ValidateUserAgent(ua)
				assert.NoError(t, err, "User-Agent %s should be valid", ua)
			}

			// 無効なUser-Agent
			invalidUserAgents := []string{
				"", // 空
			}

			for _, ua := range invalidUserAgents {
				err := validator.ValidateUserAgent(ua)
				assert.Error(t, err, "User-Agent %s should be invalid", ua)
			}
		})

		t.Run("ValidateTimestamp", func(t *testing.T) {
			// 有効なタイムスタンプ
			validTimestamp := time.Now()
			err := validator.ValidateTimestamp(validTimestamp)
			assert.NoError(t, err)

			// 無効なタイムスタンプ（未来の時間）
			invalidTimestamp := time.Now().Add(24 * time.Hour)
			err = validator.ValidateTimestamp(invalidTimestamp)
			assert.Error(t, err)
		})

		t.Run("ValidateIPAddress", func(t *testing.T) {
			// IPアドレスのバリデーションは内部メソッドのため、Validateメソッドを通じてテスト
			tracking := &models.TrackingData{
				AppID:     "test-app",
				UserAgent: "Mozilla/5.0 (Test Browser)",
				IPAddress: "192.168.1.1",
				URL:       "https://example.com/page",
				Timestamp: time.Now(),
			}

			err := validator.Validate(tracking)
			assert.NoError(t, err)

			// 無効なIPアドレス
			invalidTracking := &models.TrackingData{
				AppID:     "test-app",
				UserAgent: "Mozilla/5.0 (Test Browser)",
				IPAddress: "invalid-ip",
				URL:       "https://example.com/page",
				Timestamp: time.Now(),
			}

			err = validator.Validate(invalidTracking)
			assert.Error(t, err)
		})

		t.Run("ValidateURL", func(t *testing.T) {
			// 有効なURL
			validURLs := []string{
				"https://example.com",
				"http://localhost:8080",
				"https://sub.example.com/path?param=value",
			}

			for _, url := range validURLs {
				err := validator.ValidateURL(url)
				assert.NoError(t, err, "URL %s should be valid", url)
			}

			// 無効なURL
			invalidURLs := []string{
				"",
				"invalid-url",
				"ftp://example.com",
			}

			for _, url := range invalidURLs {
				err := validator.ValidateURL(url)
				assert.Error(t, err, "URL %s should be invalid", url)
			}
		})

		t.Run("ValidateReferrer", func(t *testing.T) {
			// リファラーのバリデーションは内部メソッドのため、Validateメソッドを通じてテスト
			tracking := &models.TrackingData{
				AppID:     "test-app",
				UserAgent: "Mozilla/5.0 (Test Browser)",
				IPAddress: "192.168.1.1",
				URL:       "https://example.com/page",
				Referrer:  "https://google.com",
				Timestamp: time.Now(),
			}

			err := validator.Validate(tracking)
			assert.NoError(t, err)

			// 無効なリファラー
			invalidTracking := &models.TrackingData{
				AppID:     "test-app",
				UserAgent: "Mozilla/5.0 (Test Browser)",
				IPAddress: "192.168.1.1",
				URL:       "https://example.com/page",
				Referrer:  "invalid-referrer",
				Timestamp: time.Now(),
			}

			err = validator.Validate(invalidTracking)
			assert.Error(t, err)
		})

		t.Run("ValidateCustomParams", func(t *testing.T) {
			// 有効なカスタムパラメータ
			validParams := map[string]interface{}{
				"utm_source":   "google",
				"utm_medium":   "cpc",
				"utm_campaign": "test_campaign",
				"page_type":    "product",
			}

			err := validator.ValidateCustomParams(validParams)
			assert.NoError(t, err)

			// 無効なカスタムパラメータ（空のキー）
			invalidParams := map[string]interface{}{
				"": "invalid_key",
			}

			err = validator.ValidateCustomParams(invalidParams)
			assert.Error(t, err)
		})

		t.Run("IsCrawler", func(t *testing.T) {
			// クローラーのUser-Agent
			crawlerUserAgents := []string{
				"Googlebot/2.1 (+http://www.google.com/bot.html)",
				"Bingbot/2.0 (+http://www.bing.com/bingbot.htm)",
				"Mozilla/5.0 (compatible; Baiduspider/2.0; +http://www.baidu.com/search/spider.html)",
			}

			for _, ua := range crawlerUserAgents {
				assert.True(t, validator.IsCrawler(ua), "User-Agent %s should be detected as crawler", ua)
			}

			// 通常のブラウザのUser-Agent
			normalUserAgents := []string{
				"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
				"Mozilla/5.0 (iPhone; CPU iPhone OS 14_7_1 like Mac OS X) AppleWebKit/605.1.15",
			}

			for _, ua := range normalUserAgents {
				assert.False(t, validator.IsCrawler(ua), "User-Agent %s should not be detected as crawler", ua)
			}
		})

		t.Run("Complex validation scenarios", func(t *testing.T) {
			// 複雑なバリデーションシナリオ
			tracking := &models.TrackingData{
				AppID:     "testapp123",
				UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 14_7_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Mobile/15E148 Safari/604.1",
				IPAddress: "192.168.1.100",
				URL:       "https://example.com/product/123?utm_source=google&utm_medium=cpc",
				Referrer:  "https://google.com/search?q=product",
				Timestamp: time.Now(),
				CustomParams: map[string]interface{}{
					"utm_source":   "google",
					"utm_medium":   "cpc",
					"utm_campaign": "summer_sale",
					"page_type":    "product",
					"product_id":   "123",
				},
			}

			// 全体のバリデーション
			err := validator.Validate(tracking)
			assert.NoError(t, err)

			// 個別のバリデーション
			err = validator.ValidateAppID(tracking.AppID)
			assert.NoError(t, err)

			err = validator.ValidateUserAgent(tracking.UserAgent)
			assert.NoError(t, err)

			err = validator.ValidateURL(tracking.URL)
			assert.NoError(t, err)

			err = validator.ValidateCustomParams(tracking.CustomParams)
			assert.NoError(t, err)

			// クローラーチェック
			assert.False(t, validator.IsCrawler(tracking.UserAgent))
		})

		t.Run("Edge cases", func(t *testing.T) {
			// 境界値のテスト
			validator := validators.NewTrackingValidator()

			// 最小限の有効なデータ
			minimalTracking := &models.TrackingData{
				AppID:     "testapp123",
				UserAgent: "Mozilla/5.0 (Test Browser)",
				IPAddress: "127.0.0.1",
				URL:       "http://a.com",
				Timestamp: time.Now(),
			}

			err := validator.Validate(minimalTracking)
			assert.NoError(t, err)

			// 空のカスタムパラメータ
			emptyParams := map[string]interface{}{}
			err = validator.ValidateCustomParams(emptyParams)
			assert.NoError(t, err)

			// nilのカスタムパラメータ
			err = validator.ValidateCustomParams(nil)
			assert.NoError(t, err)
		})
	})
}
