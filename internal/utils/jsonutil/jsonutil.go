package jsonutil

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Marshal はオブジェクトをJSONにマーシャリングします
func Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// MarshalIndent はオブジェクトをインデント付きJSONにマーシャリングします
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}

// Unmarshal はJSONをオブジェクトにアンマーシャリングします
func Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// MarshalToString はオブジェクトをJSON文字列にマーシャリングします
func MarshalToString(v interface{}) (string, error) {
	data, err := Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// UnmarshalFromString はJSON文字列をオブジェクトにアンマーシャリングします
func UnmarshalFromString(data string, v interface{}) error {
	return Unmarshal([]byte(data), v)
}

// PrettyPrint はJSONを整形して出力します
func PrettyPrint(v interface{}) (string, error) {
	return MarshalToStringIndent(v, "", "  ")
}

// MarshalToStringIndent はオブジェクトをインデント付きJSON文字列にマーシャリングします
func MarshalToStringIndent(v interface{}, prefix, indent string) (string, error) {
	data, err := MarshalIndent(v, prefix, indent)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// IsValidJSON は文字列が有効なJSONかどうかを判定します
func IsValidJSON(data string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(data), &js) == nil
}

// IsValidJSONBytes はバイト配列が有効なJSONかどうかを判定します
func IsValidJSONBytes(data []byte) bool {
	var js json.RawMessage
	return json.Unmarshal(data, &js) == nil
}

// GetJSONValue はJSONから指定されたパスの値を取得します
func GetJSONValue(data []byte, path string) (interface{}, error) {
	var jsonData interface{}
	if err := Unmarshal(data, &jsonData); err != nil {
		return nil, err
	}

	keys := strings.Split(path, ".")
	current := jsonData

	for _, key := range keys {
		if key == "" {
			continue
		}

		switch v := current.(type) {
		case map[string]interface{}:
			if value, exists := v[key]; exists {
				current = value
			} else {
				return nil, fmt.Errorf("key not found: %s", key)
			}
		case []interface{}:
			index, err := strconv.Atoi(key)
			if err != nil {
				return nil, fmt.Errorf("invalid array index: %s", key)
			}
			if index < 0 || index >= len(v) {
				return nil, fmt.Errorf("array index out of bounds: %d", index)
			}
			current = v[index]
		default:
			return nil, fmt.Errorf("cannot access key %s in type %T", key, current)
		}
	}

	return current, nil
}

// GetJSONString はJSONから指定されたパスの文字列値を取得します
func GetJSONString(data []byte, path string) (string, error) {
	value, err := GetJSONValue(data, path)
	if err != nil {
		return "", err
	}

	switch v := value.(type) {
	case string:
		return v, nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case int:
		return strconv.Itoa(v), nil
	case bool:
		return strconv.FormatBool(v), nil
	default:
		return "", fmt.Errorf("value is not a string: %T", value)
	}
}

// GetJSONInt はJSONから指定されたパスの整数値を取得します
func GetJSONInt(data []byte, path string) (int, error) {
	value, err := GetJSONValue(data, path)
	if err != nil {
		return 0, err
	}

	switch v := value.(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("value is not a number: %T", value)
	}
}

// GetJSONFloat はJSONから指定されたパスの浮動小数点値を取得します
func GetJSONFloat(data []byte, path string) (float64, error) {
	value, err := GetJSONValue(data, path)
	if err != nil {
		return 0, err
	}

	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("value is not a number: %T", value)
	}
}

// GetJSONBool はJSONから指定されたパスの真偽値を取得します
func GetJSONBool(data []byte, path string) (bool, error) {
	value, err := GetJSONValue(data, path)
	if err != nil {
		return false, err
	}

	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	default:
		return false, fmt.Errorf("value is not a boolean: %T", value)
	}
}

// SetJSONValue はJSONの指定されたパスに値を設定します
func SetJSONValue(data []byte, path string, value interface{}) ([]byte, error) {
	var jsonData interface{}
	if err := Unmarshal(data, &jsonData); err != nil {
		return nil, err
	}

	keys := strings.Split(path, ".")
	current := &jsonData

	for i, key := range keys {
		if key == "" {
			continue
		}

		if i == len(keys)-1 {
			// 最後のキーなので値を設定
			switch v := (*current).(type) {
			case map[string]interface{}:
				v[key] = value
			default:
				return nil, fmt.Errorf("cannot set value at path: %s", path)
			}
		} else {
			// 中間のキーなので次のレベルに移動
			switch v := (*current).(type) {
			case map[string]interface{}:
				if _, exists := v[key]; !exists {
					v[key] = make(map[string]interface{})
				}
				nextValue := v[key]
				current = &nextValue
			default:
				return nil, fmt.Errorf("cannot access path: %s", path)
			}
		}
	}

	return Marshal(jsonData)
}

// MergeJSON は2つのJSONオブジェクトをマージします
func MergeJSON(data1, data2 []byte) ([]byte, error) {
	var obj1, obj2 map[string]interface{}

	if err := Unmarshal(data1, &obj1); err != nil {
		return nil, err
	}

	if err := Unmarshal(data2, &obj2); err != nil {
		return nil, err
	}

	// obj2の値をobj1にマージ
	for key, value := range obj2 {
		obj1[key] = value
	}

	return Marshal(obj1)
}

// DeepMergeJSON は2つのJSONオブジェクトを深くマージします
func DeepMergeJSON(data1, data2 []byte) ([]byte, error) {
	var obj1, obj2 map[string]interface{}

	if err := Unmarshal(data1, &obj1); err != nil {
		return nil, err
	}

	if err := Unmarshal(data2, &obj2); err != nil {
		return nil, err
	}

	merged := deepMerge(obj1, obj2)
	return Marshal(merged)
}

// deepMerge は再帰的にオブジェクトをマージします
func deepMerge(obj1, obj2 map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// obj1の値をコピー
	for key, value := range obj1 {
		result[key] = value
	}

	// obj2の値をマージ
	for key, value := range obj2 {
		if existingValue, exists := result[key]; exists {
			// 既存の値がある場合、深くマージ
			if existingMap, ok := existingValue.(map[string]interface{}); ok {
				if newMap, ok := value.(map[string]interface{}); ok {
					result[key] = deepMerge(existingMap, newMap)
					continue
				}
			}
		}
		// 新しい値または上書き
		result[key] = value
	}

	return result
}

// FlattenJSON はJSONオブジェクトをフラット化します
func FlattenJSON(data []byte) (map[string]interface{}, error) {
	var obj map[string]interface{}
	if err := Unmarshal(data, &obj); err != nil {
		return nil, err
	}

	flattened := make(map[string]interface{})
	flattenObject(obj, "", flattened)
	return flattened, nil
}

// flattenObject はオブジェクトを再帰的にフラット化します
func flattenObject(obj interface{}, prefix string, result map[string]interface{}) {
	switch v := obj.(type) {
	case map[string]interface{}:
		for key, value := range v {
			newKey := key
			if prefix != "" {
				newKey = prefix + "." + key
			}
			flattenObject(value, newKey, result)
		}
	case []interface{}:
		for i, value := range v {
			newKey := strconv.Itoa(i)
			if prefix != "" {
				newKey = prefix + "." + newKey
			}
			flattenObject(value, newKey, result)
		}
	default:
		if prefix != "" {
			result[prefix] = v
		}
	}
}

// UnflattenJSON はフラット化されたJSONを元の構造に戻します
func UnflattenJSON(flattened map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for key, value := range flattened {
		keys := strings.Split(key, ".")
		current := result

		for i, k := range keys {
			if i == len(keys)-1 {
				// 最後のキーなので値を設定
				current[k] = value
			} else {
				// 中間のキーなのでマップを作成
				if _, exists := current[k]; !exists {
					current[k] = make(map[string]interface{})
				}
				current = current[k].(map[string]interface{})
			}
		}
	}

	return result, nil
}

// ConvertToMap は任意の構造体をmap[string]interface{}に変換します
func ConvertToMap(v interface{}) (map[string]interface{}, error) {
	data, err := Marshal(v)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// ConvertFromMap はmap[string]interface{}を指定された構造体に変換します
func ConvertFromMap(data map[string]interface{}, v interface{}) error {
	jsonData, err := Marshal(data)
	if err != nil {
		return err
	}

	return Unmarshal(jsonData, v)
}

// ValidateJSONSchema はJSONスキーマに基づいてJSONを検証します（簡易実装）
func ValidateJSONSchema(data []byte, schema map[string]interface{}) error {
	// 実際の実装では、github.com/xeipuuv/gojsonschema などのライブラリを使用
	// ここでは簡易的な実装例を示す

	var jsonData interface{}
	if err := Unmarshal(data, &jsonData); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// スキーマの型をチェック
	if schemaType, exists := schema["type"]; exists {
		if err := validateType(jsonData, schemaType.(string)); err != nil {
			return err
		}
	}

	// 必須フィールドをチェック
	if required, exists := schema["required"]; exists {
		if requiredList, ok := required.([]interface{}); ok {
			if obj, ok := jsonData.(map[string]interface{}); ok {
				for _, field := range requiredList {
					if fieldName, ok := field.(string); ok {
						if _, exists := obj[fieldName]; !exists {
							return fmt.Errorf("required field missing: %s", fieldName)
						}
					}
				}
			}
		}
	}

	return nil
}

// validateType は値の型を検証します
func validateType(value interface{}, expectedType string) error {
	switch expectedType {
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("expected object, got %T", value)
		}
	case "array":
		if _, ok := value.([]interface{}); !ok {
			return fmt.Errorf("expected array, got %T", value)
		}
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "number":
		switch value.(type) {
		case float64, int:
			// OK
		default:
			return fmt.Errorf("expected number, got %T", value)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case "null":
		if value != nil {
			return fmt.Errorf("expected null, got %T", value)
		}
	}
	return nil
}
