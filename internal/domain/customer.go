// domain パッケージはビジネスルールとエンティティを定義する。
// 外部ライブラリ（GORM・Gin など）への依存を持たず、最も内側の層として位置づけられる。
// 永続化の詳細（テーブル構造や型変換）はここには書かない。
package domain

import (
	"encoding/json"
	"time"
)

// Customer は顧客を表すドメインモデル。
//
// 設計方針:
//   - 主要業務データ（名前・メール・ステータス）は固定フィールドとして持つ。
//     これにより SQL での検索・集計・一覧表示が素直に書ける。
//   - テナントごとに異なる補足項目は ExtraJSON（map）に格納する。
//     検索や集計に使わない補足情報はここに入れることで、スキーマ変更なく拡張できる。
//   - EAV パターン（CustomFieldValue）と組み合わせることで、
//     項目定義の管理・型制約・検索可能な可変項目も扱える。
type Customer struct {
	ID        uint
	Name      string
	Email     string
	Status    string
	ExtraJSON map[string]interface{} // テナントごとの補足項目（固定カラム + JSON パターン）
	Notes     string                 // 自由記述欄。検索・集計・入力制御が不要な備考に使う
	CreatedAt time.Time
}

// CustomFieldDefinition はカスタムフィールドの定義を表すドメインモデル。
//
// EAV（Entity-Attribute-Value）パターンの「Attribute」にあたる。
// テナントごと・エンティティ種別ごとに項目定義を追加できる。
//
// 設計ポイント:
//   - フィールドの追加・変更をスキーマ変更なく行える。
//   - FieldType で型を管理し、値テーブルの対応カラムに格納する。
//   - IsSearchable を true にした項目は一覧検索の対象にする想定。
type CustomFieldDefinition struct {
	ID           uint
	TenantID     uint   // どのテナントの定義か（マルチテナント対応）
	EntityType   string // 対象エンティティの種別（例: "customer", "contract"）
	FieldKey     string // アプリケーション内で参照するキー名（例: "industry"）
	FieldLabel   string // 画面表示用のラベル（例: "業種"）
	FieldType    string // 値の型: "text" / "number" / "date" / "boolean"
	IsRequired   bool   // 必須入力かどうか
	IsSearchable bool   // 一覧・検索画面で絞り込み条件に使えるか
	DisplayOrder int    // 画面上の表示順
	CreatedAt    time.Time
}

// CustomFieldValue はカスタムフィールドの値を表すドメインモデル。
//
// EAV パターンの「Value」にあたる。
// 値を単一の文字列カラムに詰め込まず、型ごとに専用カラムを持つ設計にしている。
// これにより型チェックや数値での範囲検索が現実的になる。
//
// 例: FieldType が "number" なら ValueNumber に格納し、ValueText は nil にする。
type CustomFieldValue struct {
	ID                uint
	EntityID          uint    // 値が紐づくエンティティの ID（例: customers.id）
	FieldDefinitionID uint    // どのフィールド定義に対応するか
	ValueText         *string  // FieldType = "text" の場合に使用
	ValueNumber       *float64 // FieldType = "number" の場合に使用（DECIMAL(20,4) 相当）
	ValueDate         *string  // FieldType = "date" の場合に使用（"YYYY-MM-DD" 形式）
	ValueBoolean      *bool    // FieldType = "boolean" の場合に使用
	CreatedAt         time.Time
}

// SetExtra は ExtraJSON に指定キーで値をセットする。
// ExtraJSON が nil の場合は初期化してからセットする。
func (c *Customer) SetExtra(key string, value interface{}) {
	if c.ExtraJSON == nil {
		c.ExtraJSON = make(map[string]interface{})
	}
	c.ExtraJSON[key] = value
}

// GetExtra は ExtraJSON から指定キーの値を取得する。
// キーが存在しない場合は (nil, false) を返す。
func (c *Customer) GetExtra(key string) (interface{}, bool) {
	if c.ExtraJSON == nil {
		return nil, false
	}
	v, ok := c.ExtraJSON[key]
	return v, ok
}

// MarshalExtraJSON は ExtraJSON を JSON 文字列にシリアライズして返す。
// ExtraJSON が nil の場合は "{}" を返す。
func (c *Customer) MarshalExtraJSON() (string, error) {
	if c.ExtraJSON == nil {
		return "{}", nil
	}
	b, err := json.Marshal(c.ExtraJSON)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
