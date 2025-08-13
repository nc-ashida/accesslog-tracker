package jsonutil

import (
	"encoding/json"
	"reflect"
)

// Marshal はオブジェクトをJSONバイト配列に変換します
func Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal はJSONバイト配列をオブジェクトに変換します
func Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// MarshalIndent はオブジェクトをインデント付きJSONバイト配列に変換します
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}

// IsValidJSON は文字列が有効なJSONかどうかを判定します
func IsValidJSON(s string) bool {
	if s == "" {
		return false
	}
	
	var js json.RawMessage
	return json.Unmarshal([]byte(s), &js) == nil
}

// Merge は2つのマップをマージします（overrideが優先）
func Merge(base, override map[string]interface{}) map[string]interface{} {
	if base == nil {
		base = make(map[string]interface{})
	}
	
	if override == nil {
		return base
	}
	
	result := make(map[string]interface{})
	
	// ベースマップをコピー
	for k, v := range base {
		result[k] = v
	}
	
	// オーバーライドマップで上書き
	for k, v := range override {
		result[k] = v
	}
	
	return result
}

// DeepMerge は2つのマップを深くマージします
func DeepMerge(base, override map[string]interface{}) map[string]interface{} {
	if base == nil {
		base = make(map[string]interface{})
	}
	
	if override == nil {
		return base
	}
	
	result := make(map[string]interface{})
	
	// ベースマップをコピー
	for k, v := range base {
		result[k] = v
	}
	
	// オーバーライドマップで深くマージ
	for k, v := range override {
		if baseVal, exists := base[k]; exists {
			// 両方ともマップの場合、再帰的にマージ
			if baseMap, baseOk := baseVal.(map[string]interface{}); baseOk {
				if overrideMap, overrideOk := v.(map[string]interface{}); overrideOk {
					result[k] = DeepMerge(baseMap, overrideMap)
					continue
				}
			}
		}
		result[k] = v
	}
	
	return result
}

// GetNestedValue はネストしたマップから値を取得します
func GetNestedValue(data map[string]interface{}, keys ...string) (interface{}, bool) {
	current := data
	
	for i, key := range keys {
		if current == nil {
			return nil, false
		}
		
		val, exists := current[key]
		if !exists {
			return nil, false
		}
		
		// 最後のキーの場合、値を返す
		if i == len(keys)-1 {
			return val, true
		}
		
		// 中間のキーの場合、マップであることを確認
		if nextMap, ok := val.(map[string]interface{}); ok {
			current = nextMap
		} else {
			return nil, false
		}
	}
	
	return nil, false
}

// SetNestedValue はネストしたマップに値を設定します
func SetNestedValue(data map[string]interface{}, value interface{}, keys ...string) {
	if data == nil {
		data = make(map[string]interface{})
	}
	
	current := data
	
	for i, key := range keys {
		// 最後のキーの場合、値を設定
		if i == len(keys)-1 {
			current[key] = value
			return
		}
		
		// 中間のキーの場合、マップを作成または取得
		if nextMap, exists := current[key]; exists {
			if mapVal, ok := nextMap.(map[string]interface{}); ok {
				current = mapVal
			} else {
				// 既存の値がマップでない場合、新しいマップで上書き
				newMap := make(map[string]interface{})
				current[key] = newMap
				current = newMap
			}
		} else {
			// キーが存在しない場合、新しいマップを作成
			newMap := make(map[string]interface{})
			current[key] = newMap
			current = newMap
		}
	}
}

// ToMap は任意の構造体をマップに変換します
func ToMap(v interface{}) (map[string]interface{}, error) {
	data, err := Marshal(v)
	if err != nil {
		return nil, err
	}
	
	var result map[string]interface{}
	err = Unmarshal(data, &result)
	return result, err
}

// FromMap はマップを構造体に変換します
func FromMap(data map[string]interface{}, v interface{}) error {
	jsonData, err := Marshal(data)
	if err != nil {
		return err
	}
	
	return Unmarshal(jsonData, v)
}

// IsEmpty は値が空かどうかを判定します
func IsEmpty(v interface{}) bool {
	if v == nil {
		return true
	}
	
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.String:
		return val.Len() == 0
	case reflect.Array, reflect.Slice, reflect.Map:
		return val.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return val.IsNil()
	default:
		return false
	}
}
