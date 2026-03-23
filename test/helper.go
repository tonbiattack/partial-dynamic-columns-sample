// test パッケージはテスト全体で共通して使うヘルパー関数を提供する。
//
// このプロジェクトのテスト方針（古典派）:
//   - DB アクセスを伴うテストは実 DB（MySQL）を使う。モックは使わない。
//   - 各テストは独立して実行できるよう、テスト後にデータをクリーンアップする。
//   - t.Helper() を使い、ヘルパー関数内のエラーがテストの呼び出し元行に表示されるようにする。
package test

import (
	"os"
	"testing"

	"github.com/yourname/partial-dynamic-columns-sample/internal/infrastructure/persistence"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// defaultDSN はテスト用 DB のデフォルト接続文字列。
// Docker Compose で起動した MySQL（root/password）に接続する。
// parseTime=true を付けることで DATETIME カラムを time.Time として読み込める。
const defaultDSN = "root:password@tcp(localhost:3306)/partial_dynamic_columns?charset=utf8mb4&collation=utf8mb4_0900_ai_ci&parseTime=true"

// SetupTestDB はテスト用 DB への接続を確立して返す。
//
// 接続先 DSN は環境変数 TEST_DSN から取得する。
// 未設定の場合は defaultDSN（Docker Compose の MySQL）を使う。
// これにより CI 環境でも環境変数で接続先を切り替えられる。
//
// テスト中は GORM の SQL ログを Silent にして出力を抑える。
func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DSN")
	if dsn == "" {
		dsn = defaultDSN
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		// テスト実行時の出力を抑えるため Silent モードにする
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("テスト用 DB への接続に失敗しました: %v", err)
	}
	return db
}

// CleanupCustomers はテストで追加した customers テーブルのレコードを削除する。
//
// サンプルデータ（id=1〜3）は残し、テストで追加したレコード（id > 3）のみ削除する。
// defer と組み合わせてテスト終了時に呼ぶことで、テスト間のデータ干渉を防ぐ。
func CleanupCustomers(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.Exec("DELETE FROM customers WHERE id > 3").Error; err != nil {
		t.Errorf("customers テーブルのクリーンアップに失敗しました: %v", err)
	}
}

// CleanupCustomFieldValues はテストで追加した custom_field_values のレコードを削除する。
//
// entity_id > 3 のレコードを削除する（サンプルデータの顧客 id=1〜3 に紐づく値は残す）。
// CleanupCustomers の前に呼ぶこと（外部キー制約により顧客削除前に値を削除する必要がある）。
func CleanupCustomFieldValues(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.Exec("DELETE FROM custom_field_values WHERE entity_id > 3").Error; err != nil {
		t.Errorf("custom_field_values テーブルのクリーンアップに失敗しました: %v", err)
	}
}

// CleanupCustomFieldDefinitions はテストで追加した custom_field_definitions のレコードを削除する。
//
// id > 3 のレコードを削除する（サンプルデータのフィールド定義は残す）。
// CleanupCustomFieldValues の後に呼ぶこと（フィールド定義の削除は値の CASCADE DELETE を引き起こす）。
func CleanupCustomFieldDefinitions(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.Exec("DELETE FROM custom_field_definitions WHERE id > 3").Error; err != nil {
		t.Errorf("custom_field_definitions テーブルのクリーンアップに失敗しました: %v", err)
	}
}

// CreateTestCustomer はテスト用の顧客レコードを直接 DB に作成して返すヘルパー。
//
// ユースケース層を経由せず DB に直接書き込むため、
// リポジトリ・インフラ層のテストで事前データを素早く準備するのに使う。
func CreateTestCustomer(t *testing.T, db *gorm.DB, name, email, status string) *persistence.CustomerModel {
	t.Helper()
	model := &persistence.CustomerModel{
		Name:   name,
		Email:  email,
		Status: status,
	}
	if err := db.Create(model).Error; err != nil {
		t.Fatalf("テスト用顧客の作成に失敗しました: %v", err)
	}
	return model
}
