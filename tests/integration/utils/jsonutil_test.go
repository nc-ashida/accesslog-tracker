package utils_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"accesslog-tracker/internal/utils/jsonutil"
)

func TestJSONUtil_Integration(t *testing.T) {
	t.Run("Marshal", func(t *testing.T) {
		data := map[string]interface{}{
			"name": "test",
			"age":  25,
		}

		result, err := jsonutil.Marshal(data)
		require.NoError(t, err)
		assert.NotEmpty(t, result)

		var decoded map[string]interface{}
		err = json.Unmarshal(result, &decoded)
		require.NoError(t, err)
		assert.Equal(t, "test", decoded["name"])
		assert.Equal(t, float64(25), decoded["age"])
	})

	t.Run("Unmarshal", func(t *testing.T) {
		jsonData := `{"name":"test","age":25}`
		var result map[string]interface{}

		err := jsonutil.Unmarshal([]byte(jsonData), &result)
		require.NoError(t, err)
		assert.Equal(t, "test", result["name"])
		assert.Equal(t, float64(25), result["age"])
	})

	t.Run("MarshalIndent", func(t *testing.T) {
		data := map[string]interface{}{
			"name": "test",
			"age":  25,
		}

		result, err := jsonutil.MarshalIndent(data, "", "  ")
		require.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Contains(t, string(result), "\n")
	})

	t.Run("IsValidJSON", func(t *testing.T) {
		// 有効なJSON
		assert.True(t, jsonutil.IsValidJSON(`{"name":"test"}`))
		assert.True(t, jsonutil.IsValidJSON(`[1,2,3]`))
		assert.True(t, jsonutil.IsValidJSON(`"string"`))

		// 無効なJSON
		assert.False(t, jsonutil.IsValidJSON(""))
		assert.False(t, jsonutil.IsValidJSON(`{"name":}`))
		assert.False(t, jsonutil.IsValidJSON(`invalid`))
	})

	t.Run("Merge", func(t *testing.T) {
		base := map[string]interface{}{
			"name": "base",
			"age":  25,
		}
		override := map[string]interface{}{
			"name": "override",
			"city": "tokyo",
		}

		result := jsonutil.Merge(base, override)
		assert.Equal(t, "override", result["name"])
		assert.Equal(t, 25, result["age"])
		assert.Equal(t, "tokyo", result["city"])
	})

	t.Run("Merge with nil base", func(t *testing.T) {
		override := map[string]interface{}{
			"name": "override",
		}

		result := jsonutil.Merge(nil, override)
		assert.Equal(t, "override", result["name"])
	})

	t.Run("Merge with nil override", func(t *testing.T) {
		base := map[string]interface{}{
			"name": "base",
		}

		result := jsonutil.Merge(base, nil)
		assert.Equal(t, "base", result["name"])
	})

	t.Run("DeepMerge", func(t *testing.T) {
		base := map[string]interface{}{
			"user": map[string]interface{}{
				"name": "base",
				"age":  25,
			},
			"settings": map[string]interface{}{
				"theme": "dark",
			},
		}
		override := map[string]interface{}{
			"user": map[string]interface{}{
				"name": "override",
				"city": "tokyo",
			},
			"settings": map[string]interface{}{
				"language": "ja",
			},
		}

		result := jsonutil.DeepMerge(base, override)
		user := result["user"].(map[string]interface{})
		settings := result["settings"].(map[string]interface{})

		assert.Equal(t, "override", user["name"])
		assert.Equal(t, 25, user["age"])
		assert.Equal(t, "tokyo", user["city"])
		assert.Equal(t, "dark", settings["theme"])
		assert.Equal(t, "ja", settings["language"])
	})

	t.Run("GetNestedValue", func(t *testing.T) {
		data := map[string]interface{}{
			"user": map[string]interface{}{
				"profile": map[string]interface{}{
					"name": "test",
					"age":  25,
				},
			},
		}

		// 正常なケース
		value, exists := jsonutil.GetNestedValue(data, "user", "profile", "name")
		assert.True(t, exists)
		assert.Equal(t, "test", value)

		// 存在しないキー
		value, exists = jsonutil.GetNestedValue(data, "user", "profile", "nonexistent")
		assert.False(t, exists)
		assert.Nil(t, value)

		// 中間のキーがマップでない
		value, exists = jsonutil.GetNestedValue(data, "user", "profile", "name", "invalid")
		assert.False(t, exists)
		assert.Nil(t, value)
	})

	t.Run("SetNestedValue", func(t *testing.T) {
		data := make(map[string]interface{})

		// 新しいネストした値を設定
		jsonutil.SetNestedValue(data, "test", "user", "profile", "name")
		jsonutil.SetNestedValue(data, 25, "user", "profile", "age")

		user := data["user"].(map[string]interface{})
		profile := user["profile"].(map[string]interface{})
		assert.Equal(t, "test", profile["name"])
		assert.Equal(t, 25, profile["age"])

		// 既存の値を更新
		jsonutil.SetNestedValue(data, "updated", "user", "profile", "name")
		assert.Equal(t, "updated", profile["name"])
	})

	t.Run("ToMap", func(t *testing.T) {
		type TestStruct struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		original := TestStruct{
			Name: "test",
			Age:  25,
		}

		result, err := jsonutil.ToMap(original)
		require.NoError(t, err)
		assert.Equal(t, "test", result["name"])
		assert.Equal(t, float64(25), result["age"])
	})

	t.Run("FromMap", func(t *testing.T) {
		type TestStruct struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		data := map[string]interface{}{
			"name": "test",
			"age":  25,
		}

		var result TestStruct
		err := jsonutil.FromMap(data, &result)
		require.NoError(t, err)
		assert.Equal(t, "test", result.Name)
		assert.Equal(t, 25, result.Age)
	})

	t.Run("IsEmpty", func(t *testing.T) {
		// nil
		assert.True(t, jsonutil.IsEmpty(nil))

		// 空の文字列
		assert.True(t, jsonutil.IsEmpty(""))
		assert.False(t, jsonutil.IsEmpty("test"))

		// 空のスライス
		assert.True(t, jsonutil.IsEmpty([]string{}))
		assert.False(t, jsonutil.IsEmpty([]string{"test"}))

		// 空のマップ
		assert.True(t, jsonutil.IsEmpty(map[string]string{}))
		assert.False(t, jsonutil.IsEmpty(map[string]string{"key": "value"}))

		// ポインタ
		var ptr *string
		assert.True(t, jsonutil.IsEmpty(ptr))
		str := "test"
		ptr = &str
		assert.False(t, jsonutil.IsEmpty(ptr))

		// 数値
		assert.False(t, jsonutil.IsEmpty(0))
		assert.False(t, jsonutil.IsEmpty(42))

		// 構造体
		type TestStruct struct {
			Name string
		}
		assert.False(t, jsonutil.IsEmpty(TestStruct{}))
		assert.False(t, jsonutil.IsEmpty(TestStruct{Name: "test"}))
	})

	t.Run("Complex JSON operations", func(t *testing.T) {
		// 複雑なJSONデータの処理
		complexData := map[string]interface{}{
			"users": []map[string]interface{}{
				{
					"id":   1,
					"name": "user1",
					"settings": map[string]interface{}{
						"theme": "dark",
						"notifications": map[string]interface{}{
							"email": true,
							"sms":   false,
						},
					},
				},
				{
					"id":   2,
					"name": "user2",
					"settings": map[string]interface{}{
						"theme": "light",
						"notifications": map[string]interface{}{
							"email": false,
							"sms":   true,
						},
					},
				},
			},
			"metadata": map[string]interface{}{
				"total": 2,
				"page":  1,
			},
		}

		// Marshal
		jsonData, err := jsonutil.Marshal(complexData)
		require.NoError(t, err)
		assert.NotEmpty(t, jsonData)

		// Unmarshal
		var decoded map[string]interface{}
		err = jsonutil.Unmarshal(jsonData, &decoded)
		require.NoError(t, err)

		// ネストした値の取得（配列の要素にアクセス）
		users := decoded["users"].([]interface{})
		user0 := users[0].(map[string]interface{})
		settings := user0["settings"].(map[string]interface{})
		theme := settings["theme"]
		assert.Equal(t, "dark", theme)

		// ネストした値の設定
		settings["theme"] = "blue"
		theme = settings["theme"]
		assert.Equal(t, "blue", theme)

		// マージ操作
		update := map[string]interface{}{
			"metadata": map[string]interface{}{
				"updated": true,
			},
		}
		merged := jsonutil.DeepMerge(decoded, update)
		metadata := merged["metadata"].(map[string]interface{})
		assert.Equal(t, float64(2), metadata["total"])
		assert.Equal(t, float64(1), metadata["page"])
		assert.Equal(t, true, metadata["updated"])
	})

	t.Run("Error handling", func(t *testing.T) {
		// 無効なJSONのUnmarshal
		var result map[string]interface{}
		err := jsonutil.Unmarshal([]byte(`{"invalid": json}`), &result)
		assert.Error(t, err)

		// 無効な構造体のToMap
		invalidStruct := make(chan int) // マーシャル不可能な型
		_, err = jsonutil.ToMap(invalidStruct)
		assert.Error(t, err)
	})
}
