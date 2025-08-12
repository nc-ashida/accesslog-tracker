package models

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

// CustomParam はカスタムパラメータの基本構造体
type CustomParam struct {
	Key         string      `json:"key" db:"key"`
	Value       interface{} `json:"value" db:"value"`
	Type        string      `json:"type" db:"type"`
	Description string      `json:"description,omitempty" db:"description"`
	Required    bool        `json:"required" db:"required"`
	Validation  string      `json:"validation,omitempty" db:"validation"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

// CustomParams はカスタムパラメータのコレクション
type CustomParams map[string]interface{}

// ParamType はパラメータタイプの定数
const (
	ParamTypeString  = "string"
	ParamTypeNumber  = "number"
	ParamTypeBoolean = "boolean"
	ParamTypeArray   = "array"
	ParamTypeObject  = "object"
	ParamTypeDate    = "date"
	ParamTypeEmail   = "email"
	ParamTypeURL     = "url"
)

// NewCustomParam は新しいカスタムパラメータを作成
func NewCustomParam(key string, value interface{}) *CustomParam {
	now := time.Now()
	
	return &CustomParam{
		Key:       key,
		Value:     value,
		Type:      inferType(value),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// inferType は値から型を推測
func inferType(value interface{}) string {
	if value == nil {
		return ParamTypeString
	}
	
	switch v := value.(type) {
	case string:
		return ParamTypeString
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return ParamTypeNumber
	case float32, float64:
		return ParamTypeNumber
	case bool:
		return ParamTypeBoolean
	case []interface{}:
		return ParamTypeArray
	case map[string]interface{}:
		return ParamTypeObject
	case time.Time:
		return ParamTypeDate
	default:
		rt := reflect.TypeOf(v)
		switch rt.Kind() {
		case reflect.String:
			return ParamTypeString
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			 reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return ParamTypeNumber
		case reflect.Float32, reflect.Float64:
			return ParamTypeNumber
		case reflect.Bool:
			return ParamTypeBoolean
		case reflect.Slice, reflect.Array:
			return ParamTypeArray
		case reflect.Map:
			return ParamTypeObject
		default:
			return ParamTypeString
		}
	}
}

// ToJSON はカスタムパラメータをJSONに変換
func (cp *CustomParam) ToJSON() ([]byte, error) {
	return json.Marshal(cp)
}

// FromJSON はJSONからカスタムパラメータを作成
func (cp *CustomParam) FromJSON(data []byte) error {
	return json.Unmarshal(data, cp)
}

// Validate はカスタムパラメータを検証
func (cp *CustomParam) Validate() error {
	if cp.Key == "" {
		return fmt.Errorf("key is required")
	}
	
	if len(cp.Key) > 50 {
		return fmt.Errorf("key length must be less than 50 characters")
	}
	
	if cp.Value == nil && cp.Required {
		return fmt.Errorf("value is required for key: %s", cp.Key)
	}
	
	return nil
}

// GetString は文字列値として取得
func (cp *CustomParam) GetString() string {
	if str, ok := cp.Value.(string); ok {
		return str
	}
	return fmt.Sprintf("%v", cp.Value)
}

// GetNumber は数値として取得
func (cp *CustomParam) GetNumber() float64 {
	switch v := cp.Value.(type) {
	case int:
		return float64(v)
	case int8:
		return float64(v)
	case int16:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case uint:
		return float64(v)
	case uint8:
		return float64(v)
	case uint16:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
		return float64(v)
	case float32:
		return float64(v)
	case float64:
		return v
	default:
		return 0
	}
}

// GetBoolean は真偽値として取得
func (cp *CustomParam) GetBoolean() bool {
	if b, ok := cp.Value.(bool); ok {
		return b
	}
	return false
}

// SetValue は値を設定
func (cp *CustomParam) SetValue(value interface{}) {
	cp.Value = value
	cp.Type = inferType(value)
	cp.UpdatedAt = time.Now()
}

// CustomParams のメソッド

// Get はキーに対応する値を取得
func (cp CustomParams) Get(key string) interface{} {
	return cp[key]
}

// GetString はキーに対応する文字列値を取得
func (cp CustomParams) GetString(key string) string {
	if value, exists := cp[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
		return fmt.Sprintf("%v", value)
	}
	return ""
}

// GetNumber はキーに対応する数値を取得
func (cp CustomParams) GetNumber(key string) float64 {
	if value, exists := cp[key]; exists {
		switch v := value.(type) {
		case int:
			return float64(v)
		case int8:
			return float64(v)
		case int16:
			return float64(v)
		case int32:
			return float64(v)
		case int64:
			return float64(v)
		case uint:
			return float64(v)
		case uint8:
			return float64(v)
		case uint16:
			return float64(v)
		case uint32:
			return float64(v)
		case uint64:
			return float64(v)
		case float32:
			return float64(v)
		case float64:
			return v
		}
	}
	return 0
}

// GetBoolean はキーに対応する真偽値を取得
func (cp CustomParams) GetBoolean(key string) bool {
	if value, exists := cp[key]; exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return false
}

// Set はキーと値を設定
func (cp CustomParams) Set(key string, value interface{}) {
	cp[key] = value
}

// Delete はキーを削除
func (cp CustomParams) Delete(key string) {
	delete(cp, key)
}

// Has はキーが存在するかどうかを判定
func (cp CustomParams) Has(key string) bool {
	_, exists := cp[key]
	return exists
}

// Keys はすべてのキーを取得
func (cp CustomParams) Keys() []string {
	keys := make([]string, 0, len(cp))
	for key := range cp {
		keys = append(keys, key)
	}
	return keys
}

// ToJSON はカスタムパラメータをJSONに変換
func (cp CustomParams) ToJSON() ([]byte, error) {
	return json.Marshal(cp)
}

// FromJSON はJSONからカスタムパラメータを作成
func (cp CustomParams) FromJSON(data []byte) error {
	return json.Unmarshal(data, &cp)
}
