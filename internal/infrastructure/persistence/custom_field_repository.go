package persistence

import (
	"github.com/yourname/partial-dynamic-columns-sample/internal/domain"
	"github.com/yourname/partial-dynamic-columns-sample/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// customFieldRepository は repository.CustomFieldRepository インターフェースの GORM 実装。
// EAV パターン（フィールド定義・値）の永続化を担当する。
type customFieldRepository struct {
	db *gorm.DB
}

// NewCustomFieldRepository は CustomFieldRepository の実装を生成して返す。
func NewCustomFieldRepository(db *gorm.DB) repository.CustomFieldRepository {
	return &customFieldRepository{db: db}
}

// FindDefinitionsByTenantAndEntityType は指定テナント・エンティティ種別のフィールド定義一覧を返す。
//
// 並び順は display_order ASC, id ASC とすることで、
// 同一 display_order 内では登録順（id 順）に並ぶ。
// これによりテストで ORDER BY を明示でき、結果の比較が安定する。
func (r *customFieldRepository) FindDefinitionsByTenantAndEntityType(tenantID uint, entityType string) ([]*domain.CustomFieldDefinition, error) {
	var models []CustomFieldDefinitionModel
	result := r.db.
		Where("tenant_id = ? AND entity_type = ?", tenantID, entityType).
		Order("display_order ASC, id ASC").
		Find(&models)
	if result.Error != nil {
		return nil, result.Error
	}
	defs := make([]*domain.CustomFieldDefinition, 0, len(models))
	for i := range models {
		defs = append(defs, toDomainDefinition(&models[i]))
	}
	return defs, nil
}

// CreateDefinition はカスタムフィールド定義を新規作成する。
// INSERT 後に採番された ID と作成日時をドメインモデルに反映する。
func (r *customFieldRepository) CreateDefinition(def *domain.CustomFieldDefinition) error {
	model := toDefinitionModel(def)
	result := r.db.Create(model)
	if result.Error != nil {
		return result.Error
	}
	// DB が採番した ID と作成日時をドメインモデルに反映する
	def.ID = model.ID
	def.CreatedAt = model.CreatedAt
	return nil
}

// UpsertValue はカスタムフィールドの値を登録または更新する。
//
// custom_field_values テーブルには (entity_id, field_definition_id) のユニーク制約がある。
// clause.OnConflict を使い、重複時は値カラムを上書きする（ON DUPLICATE KEY UPDATE 相当）。
//
// この設計により、「値がなければ INSERT、あれば UPDATE」の処理を
// アプリ側で事前チェックせずに DB 側で一括処理できる。
func (r *customFieldRepository) UpsertValue(val *domain.CustomFieldValue) error {
	model := toValueModel(val)
	result := r.db.Clauses(clause.OnConflict{
		// 競合を検出するカラム（ユニーク制約に対応）
		Columns: []clause.Column{
			{Name: "entity_id"},
			{Name: "field_definition_id"},
		},
		// 競合時に上書きするカラム（値カラムのみ。id や created_at は変更しない）
		DoUpdates: clause.AssignmentColumns([]string{
			"value_text",
			"value_number",
			"value_date",
			"value_boolean",
		}),
	}).Create(model)
	if result.Error != nil {
		return result.Error
	}
	val.ID = model.ID
	val.CreatedAt = model.CreatedAt
	return nil
}

// FindValuesByEntityID は指定エンティティに紐づくカスタムフィールド値の一覧を返す。
// field_definition_id 昇順で返すことで、テストでの比較が安定する。
func (r *customFieldRepository) FindValuesByEntityID(entityID uint) ([]*domain.CustomFieldValue, error) {
	var models []CustomFieldValueModel
	result := r.db.
		Where("entity_id = ?", entityID).
		Order("field_definition_id ASC").
		Find(&models)
	if result.Error != nil {
		return nil, result.Error
	}
	values := make([]*domain.CustomFieldValue, 0, len(models))
	for i := range models {
		values = append(values, toDomainValue(&models[i]))
	}
	return values, nil
}

// toDomainDefinition は GORM モデルをドメインモデルに変換する。
func toDomainDefinition(m *CustomFieldDefinitionModel) *domain.CustomFieldDefinition {
	return &domain.CustomFieldDefinition{
		ID:           m.ID,
		TenantID:     m.TenantID,
		EntityType:   m.EntityType,
		FieldKey:     m.FieldKey,
		FieldLabel:   m.FieldLabel,
		FieldType:    m.FieldType,
		IsRequired:   m.IsRequired,
		IsSearchable: m.IsSearchable,
		DisplayOrder: m.DisplayOrder,
		CreatedAt:    m.CreatedAt,
	}
}

// toDefinitionModel はドメインモデルを GORM モデルに変換する。
func toDefinitionModel(d *domain.CustomFieldDefinition) *CustomFieldDefinitionModel {
	return &CustomFieldDefinitionModel{
		ID:           d.ID,
		TenantID:     d.TenantID,
		EntityType:   d.EntityType,
		FieldKey:     d.FieldKey,
		FieldLabel:   d.FieldLabel,
		FieldType:    d.FieldType,
		IsRequired:   d.IsRequired,
		IsSearchable: d.IsSearchable,
		DisplayOrder: d.DisplayOrder,
		CreatedAt:    d.CreatedAt,
	}
}

// toDomainValue は GORM モデルをドメインモデルに変換する。
// 値カラムはすべてポインタ型で持つため、nil の場合はそのまま nil を渡す。
func toDomainValue(m *CustomFieldValueModel) *domain.CustomFieldValue {
	return &domain.CustomFieldValue{
		ID:                m.ID,
		EntityID:          m.EntityID,
		FieldDefinitionID: m.FieldDefinitionID,
		ValueText:         m.ValueText,
		ValueNumber:       m.ValueNumber,
		ValueDate:         m.ValueDate,
		ValueBoolean:      m.ValueBoolean,
		CreatedAt:         m.CreatedAt,
	}
}

// toValueModel はドメインモデルを GORM モデルに変換する。
func toValueModel(v *domain.CustomFieldValue) *CustomFieldValueModel {
	return &CustomFieldValueModel{
		ID:                v.ID,
		EntityID:          v.EntityID,
		FieldDefinitionID: v.FieldDefinitionID,
		ValueText:         v.ValueText,
		ValueNumber:       v.ValueNumber,
		ValueDate:         v.ValueDate,
		ValueBoolean:      v.ValueBoolean,
		CreatedAt:         v.CreatedAt,
	}
}
