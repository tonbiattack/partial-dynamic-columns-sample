// このファイルでは「固定カラム + EAV」パターンを検証する。
//
// EAV（Entity-Attribute-Value）パターンの概要:
//   - custom_field_definitions テーブルでフィールド定義（属性）を管理する
//   - custom_field_values テーブルで各エンティティの値を管理する
//   - 値は型ごとに専用カラム（value_text / value_number / value_date / value_boolean）に格納する
//
// 検証内容:
//   - フィールド定義を作成して ID が採番されること
//   - テナントとエンティティ種別でフィールド定義を絞り込めること
//   - テキスト値・数値を正しく保存・取得できること
//   - 同一エンティティ・フィールドの値を Upsert で更新できること
package integration_test

import (
	"testing"

	"github.com/yourname/partial-dynamic-columns-sample/internal/domain"
	"github.com/yourname/partial-dynamic-columns-sample/internal/infrastructure/persistence"
	testhelper "github.com/yourname/partial-dynamic-columns-sample/test"
)

// TestCustomFieldDefinitionAndValue は EAV パターンの統合テスト。
// フィールド定義の作成・値の保存（Upsert）・取得を通しで検証する。
func TestCustomFieldDefinitionAndValue(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	// 削除順序に注意: 外部キー制約のため値 → 定義 → 顧客の順でクリーンアップする
	defer testhelper.CleanupCustomFieldValues(t, db)
	defer testhelper.CleanupCustomFieldDefinitions(t, db)
	defer testhelper.CleanupCustomers(t, db)

	customerRepo := persistence.NewCustomerRepository(db)
	fieldRepo := persistence.NewCustomFieldRepository(db)

	t.Run("EAVパターン_フィールド定義を作成できること", func(t *testing.T) {
		def := &domain.CustomFieldDefinition{
			TenantID:     99,
			EntityType:   "customer",
			FieldKey:     "test_field",
			FieldLabel:   "テストフィールド",
			FieldType:    "text",
			IsRequired:   false,
			IsSearchable: true,
			DisplayOrder: 10,
		}

		if err := fieldRepo.CreateDefinition(def); err != nil {
			t.Fatalf("フィールド定義の作成に失敗しました: %v", err)
		}
		// DB が採番した ID がドメインモデルに反映されていることを確認する
		if def.ID == 0 {
			t.Error("フィールド定義の ID が採番されていません")
		}
	})

	t.Run("EAVパターン_テナントとエンティティ種別でフィールド定義を取得できること", func(t *testing.T) {
		// テナント 98 に2つのフィールド定義を作成する
		def1 := &domain.CustomFieldDefinition{
			TenantID: 98, EntityType: "customer",
			FieldKey: "field_a", FieldLabel: "フィールドA", FieldType: "text",
			DisplayOrder: 1,
		}
		def2 := &domain.CustomFieldDefinition{
			TenantID: 98, EntityType: "customer",
			FieldKey: "field_b", FieldLabel: "フィールドB", FieldType: "number",
			DisplayOrder: 2,
		}
		// テナント 97 のフィールド定義（テナント 98 の検索結果に含まれないことを確認するため）
		defOther := &domain.CustomFieldDefinition{
			TenantID: 97, EntityType: "customer",
			FieldKey: "other_field", FieldLabel: "別テナント", FieldType: "text",
		}

		for _, def := range []*domain.CustomFieldDefinition{def1, def2, defOther} {
			if err := fieldRepo.CreateDefinition(def); err != nil {
				t.Fatalf("フィールド定義の作成に失敗しました: %v", err)
			}
		}

		// テナント 98 のフィールド定義のみ取得できることを検証する
		// display_order ASC で並ぶため field_a → field_b の順になる
		defs, err := fieldRepo.FindDefinitionsByTenantAndEntityType(98, "customer")
		if err != nil {
			t.Fatalf("フィールド定義の取得に失敗しました: %v", err)
		}
		if len(defs) != 2 {
			t.Errorf("取得件数が一致しません: want=2, got=%d", len(defs))
		}
		if defs[0].FieldKey != "field_a" {
			t.Errorf("1件目のフィールドキーが一致しません: want=field_a, got=%s", defs[0].FieldKey)
		}
		if defs[1].FieldKey != "field_b" {
			t.Errorf("2件目のフィールドキーが一致しません: want=field_b, got=%s", defs[1].FieldKey)
		}
	})

	t.Run("EAVパターン_テキスト値を保存して取得できること", func(t *testing.T) {
		// 事前に顧客とフィールド定義を作成する
		customer := &domain.Customer{
			Name: "EAVテスト顧客", Email: "eav@example.com", Status: "active",
		}
		if err := customerRepo.Create(customer); err != nil {
			t.Fatalf("顧客の作成に失敗しました: %v", err)
		}

		def := &domain.CustomFieldDefinition{
			TenantID: 96, EntityType: "customer",
			FieldKey: "region", FieldLabel: "地域", FieldType: "text",
		}
		if err := fieldRepo.CreateDefinition(def); err != nil {
			t.Fatalf("フィールド定義の作成に失敗しました: %v", err)
		}

		// FieldType = "text" なので ValueText に格納する
		// ValueNumber / ValueDate / ValueBoolean は nil のまま
		textVal := "関東"
		val := &domain.CustomFieldValue{
			EntityID:          customer.ID,
			FieldDefinitionID: def.ID,
			ValueText:         &textVal,
		}
		if err := fieldRepo.UpsertValue(val); err != nil {
			t.Fatalf("カスタムフィールド値の保存に失敗しました: %v", err)
		}

		// DB から取得して ValueText が正しく保存されていることを確認する
		values, err := fieldRepo.FindValuesByEntityID(customer.ID)
		if err != nil {
			t.Fatalf("カスタムフィールド値の取得に失敗しました: %v", err)
		}
		if len(values) != 1 {
			t.Fatalf("取得件数が一致しません: want=1, got=%d", len(values))
		}
		if values[0].ValueText == nil {
			t.Fatal("ValueText が nil です")
		}
		if *values[0].ValueText != "関東" {
			t.Errorf("ValueText が一致しません: want=関東, got=%s", *values[0].ValueText)
		}
	})

	t.Run("EAVパターン_数値値を保存して取得できること", func(t *testing.T) {
		customer := &domain.Customer{
			Name: "数値EAVテスト", Email: "number-eav@example.com", Status: "active",
		}
		if err := customerRepo.Create(customer); err != nil {
			t.Fatalf("顧客の作成に失敗しました: %v", err)
		}

		// FieldType = "number" のフィールド定義を作成する
		def := &domain.CustomFieldDefinition{
			TenantID: 95, EntityType: "customer",
			FieldKey: "annual_revenue", FieldLabel: "年間売上", FieldType: "number",
		}
		if err := fieldRepo.CreateDefinition(def); err != nil {
			t.Fatalf("フィールド定義の作成に失敗しました: %v", err)
		}

		// ValueNumber に数値を格納する
		// DB カラムは DECIMAL(20,4) のため小数点4桁まで保持される
		revenue := 12345.67
		val := &domain.CustomFieldValue{
			EntityID:          customer.ID,
			FieldDefinitionID: def.ID,
			ValueNumber:       &revenue,
		}
		if err := fieldRepo.UpsertValue(val); err != nil {
			t.Fatalf("カスタムフィールド値の保存に失敗しました: %v", err)
		}

		values, err := fieldRepo.FindValuesByEntityID(customer.ID)
		if err != nil {
			t.Fatalf("カスタムフィールド値の取得に失敗しました: %v", err)
		}
		if len(values) != 1 {
			t.Fatalf("取得件数が一致しません: want=1, got=%d", len(values))
		}
		if values[0].ValueNumber == nil {
			t.Fatal("ValueNumber が nil です")
		}
		if *values[0].ValueNumber != 12345.67 {
			t.Errorf("ValueNumber が一致しません: want=12345.67, got=%v", *values[0].ValueNumber)
		}
	})

	t.Run("EAVパターン_同一エンティティ・フィールドの値をUpsertで更新できること", func(t *testing.T) {
		customer := &domain.Customer{
			Name: "Upsertテスト顧客", Email: "upsert@example.com", Status: "active",
		}
		if err := customerRepo.Create(customer); err != nil {
			t.Fatalf("顧客の作成に失敗しました: %v", err)
		}

		def := &domain.CustomFieldDefinition{
			TenantID: 94, EntityType: "customer",
			FieldKey: "priority", FieldLabel: "優先度", FieldType: "text",
		}
		if err := fieldRepo.CreateDefinition(def); err != nil {
			t.Fatalf("フィールド定義の作成に失敗しました: %v", err)
		}

		// 初回: "優先度" = "低" で登録する
		val1 := "低"
		if err := fieldRepo.UpsertValue(&domain.CustomFieldValue{
			EntityID:          customer.ID,
			FieldDefinitionID: def.ID,
			ValueText:         &val1,
		}); err != nil {
			t.Fatalf("初回 Upsert に失敗しました: %v", err)
		}

		// 2回目: 同じ (entity_id, field_definition_id) で "高" に更新する
		// ON DUPLICATE KEY UPDATE が働き、INSERT ではなく UPDATE になる
		val2 := "高"
		if err := fieldRepo.UpsertValue(&domain.CustomFieldValue{
			EntityID:          customer.ID,
			FieldDefinitionID: def.ID,
			ValueText:         &val2,
		}); err != nil {
			t.Fatalf("更新 Upsert に失敗しました: %v", err)
		}

		// Upsert 後はレコードが1件のみで、値が "高" になっていることを確認する
		values, err := fieldRepo.FindValuesByEntityID(customer.ID)
		if err != nil {
			t.Fatalf("カスタムフィールド値の取得に失敗しました: %v", err)
		}
		if len(values) != 1 {
			t.Fatalf("Upsert 後のレコード件数が一致しません: want=1, got=%d", len(values))
		}
		if *values[0].ValueText != "高" {
			t.Errorf("Upsert 後の値が一致しません: want=高, got=%s", *values[0].ValueText)
		}
	})
}
