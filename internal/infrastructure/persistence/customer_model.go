// persistence パッケージは GORM を使った MySQL への永続化実装を提供する。
//
// クリーンアーキテクチャにおける最も外側の層（infrastructure 層）にあたる。
// domain 層のモデルと DB のテーブル構造のマッピングはこのパッケージ内で完結させ、
// domain 層に GORM のタグや型が漏れないようにする。
package persistence

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// JSONMap は MySQL の JSON 型カラムを Go の map として読み書きするための GORM カスタム型。
//
// GORM はカラムへの書き込み時に driver.Valuer インターフェースを、
// カラムからの読み出し時に sql.Scanner インターフェースを呼び出す。
// この型を定義することで、JSON 型カラムを map[string]interface{} として透過的に扱える。
type JSONMap map[string]interface{}

// Value は DB への書き込み時に呼ばれる。
// map を JSON 文字列にシリアライズして返す。
// map が nil の場合は空オブジェクト "{}" を返す。
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	b, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// Scan は DB からの読み出し時に呼ばれる。
// JSON 文字列を map にデシリアライズして j にセットする。
// MySQL ドライバは JSON 型を []byte か string で渡してくるため、両方に対応している。
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONMap)
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("JSONMap: 未対応の型 %T", value)
	}
	result := make(JSONMap)
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}
	*j = result
	return nil
}

// CustomerModel は customers テーブルに対応する GORM モデル。
//
// domain.Customer との違い:
//   - ExtraJSON は JSONMap 型（driver.Valuer / sql.Scanner 実装済み）で持つ。
//     domain 層では map[string]interface{} として扱うため、相互変換が必要。
//   - gorm タグでカラム制約（not null・autoIncrement など）を指定している。
type CustomerModel struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"not null"`
	Email     string
	Status    string    `gorm:"not null"`
	ExtraJSON JSONMap   `gorm:"type:json;not null;default:'{}'"`
	Notes     string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"not null;autoCreateTime"`
}

// TableName は GORM がテーブル名を解決するために参照するメソッド。
// デフォルトでは構造体名をスネークケース複数形にするが、明示的に指定することで変更を防ぐ。
func (CustomerModel) TableName() string {
	return "customers"
}

// CustomFieldDefinitionModel は custom_field_definitions テーブルに対応する GORM モデル。
//
// EAV パターンの「フィールド定義」部分。
// テナントごとにどんな項目を持つかを管理する。
// (tenant_id, entity_type, field_key) にユニーク制約があり、同一テナント内での重複を防ぐ。
type CustomFieldDefinitionModel struct {
	ID           uint      `gorm:"primaryKey;autoIncrement"`
	TenantID     uint      `gorm:"not null"`
	EntityType   string    `gorm:"not null"`
	FieldKey     string    `gorm:"not null"`
	FieldLabel   string    `gorm:"not null"`
	FieldType    string    `gorm:"not null"` // "text" / "number" / "date" / "boolean"
	IsRequired   bool      `gorm:"not null;default:false"`
	IsSearchable bool      `gorm:"not null;default:false"`
	DisplayOrder int       `gorm:"not null;default:0"`
	CreatedAt    time.Time `gorm:"not null;autoCreateTime"`
}

// TableName は GORM がテーブル名を解決するために参照するメソッド。
func (CustomFieldDefinitionModel) TableName() string {
	return "custom_field_definitions"
}

// CustomFieldValueModel は custom_field_values テーブルに対応する GORM モデル。
//
// EAV パターンの「値」部分。
// 値を単一の文字列カラムに詰め込まず、型ごとに専用カラムを持つ設計にしている。
// 使用するカラムは FieldType によって決まり、それ以外は nil になる。
//
//   - FieldType = "text"    → ValueText に格納
//   - FieldType = "number"  → ValueNumber に格納（DECIMAL(20,4)）
//   - FieldType = "date"    → ValueDate に格納（"YYYY-MM-DD" 形式の文字列）
//   - FieldType = "boolean" → ValueBoolean に格納
//
// (entity_id, field_definition_id) にユニーク制約があり、
// 同一エンティティ・同一フィールドの値が複数登録されないようにしている。
// Upsert（INSERT ... ON DUPLICATE KEY UPDATE）でこの制約を活用する。
type CustomFieldValueModel struct {
	ID                uint      `gorm:"primaryKey;autoIncrement"`
	EntityID          uint      `gorm:"not null"`
	FieldDefinitionID uint      `gorm:"not null"`
	ValueText         *string   `gorm:"type:text"`
	ValueNumber       *float64  `gorm:"type:decimal(20,4)"`
	ValueDate         *string   `gorm:"type:date"`
	ValueBoolean      *bool
	CreatedAt         time.Time `gorm:"not null;autoCreateTime"`
}

// TableName は GORM がテーブル名を解決するために参照するメソッド。
func (CustomFieldValueModel) TableName() string {
	return "custom_field_values"
}
