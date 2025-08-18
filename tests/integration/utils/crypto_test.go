package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"accesslog-tracker/internal/utils/crypto"
)

func TestCryptoUtils_Integration(t *testing.T) {
	t.Run("Hash Generation", func(t *testing.T) {
		// ハッシュ生成のテスト
		data := "test data for hashing"
		hash := crypto.HashSHA256(data)

		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 64) // SHA-256は64文字

		// 同じデータから同じハッシュが生成されることを確認
		hash2 := crypto.HashSHA256(data)
		assert.Equal(t, hash, hash2)

		// 異なるデータから異なるハッシュが生成されることを確認
		hash3 := crypto.HashSHA256("different data")
		assert.NotEqual(t, hash, hash3)
	})

	t.Run("Random String Generation", func(t *testing.T) {
		// ランダム文字列生成のテスト
		length := 32
		randomStr := crypto.GenerateRandomString(length)

		assert.NotEmpty(t, randomStr)
		assert.Len(t, randomStr, length)

		// 複数回生成して異なる文字列が生成されることを確認
		randomStr2 := crypto.GenerateRandomString(length)
		assert.NotEqual(t, randomStr, randomStr2)
	})

	t.Run("API Key Generation", func(t *testing.T) {
		// APIキー生成のテスト
		apiKey := crypto.GenerateAPIKey()

		assert.NotEmpty(t, apiKey)
		assert.Len(t, apiKey, 32)

		// APIキーの検証
		assert.True(t, crypto.ValidateAPIKey(apiKey))
	})

	t.Run("API Key Validation", func(t *testing.T) {
		// 有効なAPIキー
		assert.True(t, crypto.ValidateAPIKey("valid-api-key-123"))
		assert.True(t, crypto.ValidateAPIKey("another-valid-key"))

		// 無効なAPIキー
		assert.False(t, crypto.ValidateAPIKey("short"))
		assert.False(t, crypto.ValidateAPIKey("invalid@key"))
		assert.False(t, crypto.ValidateAPIKey("invalid key"))
	})

	t.Run("Secure Token Generation", func(t *testing.T) {
		// セキュアトークン生成のテスト
		token, err := crypto.GenerateSecureToken()
		require.NoError(t, err)

		assert.NotEmpty(t, token)
		assert.Len(t, token, 64) // hexエンコードされた32バイト

		// 複数回生成して異なるトークンが生成されることを確認
		token2, err := crypto.GenerateSecureToken()
		require.NoError(t, err)
		assert.NotEqual(t, token, token2)
	})

	t.Run("Password Hashing", func(t *testing.T) {
		// パスワードハッシュのテスト
		password := "testpassword123"

		// パスワードのハッシュ化
		hashedPassword := crypto.HashPassword(password)
		assert.NotEmpty(t, hashedPassword)
		assert.NotEqual(t, password, hashedPassword)

		// パスワードの検証
		assert.True(t, crypto.VerifyPassword(password, hashedPassword))
		assert.False(t, crypto.VerifyPassword("wrongpassword", hashedPassword))
	})
}
