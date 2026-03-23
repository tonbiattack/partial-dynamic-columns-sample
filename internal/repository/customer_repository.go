// repository パッケージは永続化層のインターフェースを定義する。
//
// クリーンアーキテクチャにおける依存関係の逆転を実現するための層。
// usecase 層はこのインターフェースにのみ依存し、
// GORM や MySQL などの具体的な実装には依存しない。
//
// インターフェースの実装は infrastructure/persistence パッケージに置く。
package repository

import "github.com/yourname/partial-dynamic-columns-sample/internal/domain"

// CustomerRepository は顧客の永続化操作を抽象化するインターフェース。
//
// usecase 層はこのインターフェースを通じてデータアクセスを行う。
// テストでは実 DB（MySQL）に接続した実装を使用し、モックは使わない。
type CustomerRepository interface {
	// FindByID は指定 ID の顧客を返す。見つからない場合は (nil, nil) を返す。
	FindByID(id uint) (*domain.Customer, error)

	// FindAll はすべての顧客を ID 昇順で返す。
	FindAll() ([]*domain.Customer, error)

	// Create は新規顧客を作成し、採番された ID を c.ID にセットする。
	Create(c *domain.Customer) error

	// Update は顧客情報（extra_json を含む全フィールド）を上書き保存する。
	Update(c *domain.Customer) error
}

// CustomFieldRepository はカスタムフィールド（EAV パターン）の永続化操作を抽象化するインターフェース。
//
// EAV（Entity-Attribute-Value）パターンでは、
// フィールド定義（Attribute）と値（Value）を別テーブルで管理する。
// このインターフェースはその両方の操作をまとめて提供する。
type CustomFieldRepository interface {
	// FindDefinitionsByTenantAndEntityType は指定テナント・エンティティ種別の
	// フィールド定義一覧を display_order 昇順で返す。
	FindDefinitionsByTenantAndEntityType(tenantID uint, entityType string) ([]*domain.CustomFieldDefinition, error)

	// CreateDefinition は新規フィールド定義を作成し、採番された ID を def.ID にセットする。
	CreateDefinition(def *domain.CustomFieldDefinition) error

	// UpsertValue はカスタムフィールドの値を登録または更新する。
	// (entity_id, field_definition_id) の組み合わせが既に存在する場合は値を上書きする。
	UpsertValue(val *domain.CustomFieldValue) error

	// FindValuesByEntityID は指定エンティティに紐づくカスタムフィールド値を返す。
	FindValuesByEntityID(entityID uint) ([]*domain.CustomFieldValue, error)
}
