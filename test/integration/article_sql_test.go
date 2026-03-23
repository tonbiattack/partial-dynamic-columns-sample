package integration_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	testhelper "github.com/yourname/partial-dynamic-columns-sample/test"
)

func TestArticleSQLSamples(t *testing.T) {
	db := testhelper.SetupTestDB(t)

	t.Run("記事SQL_JSONの検索クエリで業種と契約ランクを取得できること", func(t *testing.T) {
		query := loadSQLFile(t, "sql/article/json_industry_customers.sql")

		var rows []struct {
			ID           uint
			Name         string
			Industry     string
			ContractRank string
		}
		if err := db.Raw(query).Scan(&rows).Error; err != nil {
			t.Fatalf("JSON検索SQLの実行に失敗しました: %v", err)
		}

		if len(rows) != 1 {
			t.Fatalf("取得件数が一致しません: want=1, got=%d", len(rows))
		}
		if rows[0].Name != "田中 太郎" {
			t.Errorf("顧客名が一致しません: want=田中 太郎, got=%s", rows[0].Name)
		}
		if rows[0].Industry != "IT" {
			t.Errorf("業種が一致しません: want=IT, got=%s", rows[0].Industry)
		}
		if rows[0].ContractRank != "A" {
			t.Errorf("契約ランクが一致しません: want=A, got=%s", rows[0].ContractRank)
		}
	})

	t.Run("記事SQL_JSONの集計クエリで業種別件数を取得できること", func(t *testing.T) {
		query := loadSQLFile(t, "sql/article/json_industry_summary.sql")

		var rows []struct {
			Industry      string
			CustomerCount int
		}
		if err := db.Raw(query).Scan(&rows).Error; err != nil {
			t.Fatalf("JSON集計SQLの実行に失敗しました: %v", err)
		}

		if len(rows) != 2 {
			t.Fatalf("取得件数が一致しません: want=2, got=%d", len(rows))
		}
		if rows[0].Industry != "IT" || rows[0].CustomerCount != 1 {
			t.Errorf("1件目の集計結果が一致しません: got=%+v", rows[0])
		}
		if rows[1].Industry != "製造" || rows[1].CustomerCount != 1 {
			t.Errorf("2件目の集計結果が一致しません: got=%+v", rows[1])
		}
	})

	t.Run("記事SQL_EAVの検索クエリで業種と契約ランクの両方で絞り込めること", func(t *testing.T) {
		query := loadSQLFile(t, "sql/article/eav_industry_contract_rank_customers.sql")

		var rows []struct {
			ID   uint
			Name string
		}
		if err := db.Raw(query).Scan(&rows).Error; err != nil {
			t.Fatalf("EAV検索SQLの実行に失敗しました: %v", err)
		}

		if len(rows) != 1 {
			t.Fatalf("取得件数が一致しません: want=1, got=%d", len(rows))
		}
		if rows[0].Name != "田中 太郎" {
			t.Errorf("顧客名が一致しません: want=田中 太郎, got=%s", rows[0].Name)
		}
	})

	t.Run("記事SQL_JSONからEAVへのバックフィルクエリで値を移送できること", func(t *testing.T) {
		query := loadSQLFile(t, "sql/article/backfill_extra_json_to_eav_industry.sql")

		if err := db.Exec("DELETE FROM custom_field_values WHERE entity_id IN (1, 2) AND field_definition_id = 1").Error; err != nil {
			t.Fatalf("事前クリーンアップに失敗しました: %v", err)
		}
		if err := db.Exec(query).Error; err != nil {
			t.Fatalf("バックフィルSQLの実行に失敗しました: %v", err)
		}

		var rows []struct {
			EntityID  uint
			ValueText string
		}
		if err := db.Raw(`
			SELECT entity_id, value_text
			FROM custom_field_values
			WHERE field_definition_id = 1
			ORDER BY entity_id ASC
		`).Scan(&rows).Error; err != nil {
			t.Fatalf("バックフィル結果の取得に失敗しました: %v", err)
		}

		if len(rows) != 2 {
			t.Fatalf("バックフィル件数が一致しません: want=2, got=%d", len(rows))
		}
		if rows[0].EntityID != 1 || rows[0].ValueText != "IT" {
			t.Errorf("1件目のバックフィル結果が一致しません: got=%+v", rows[0])
		}
		if rows[1].EntityID != 2 || rows[1].ValueText != "製造" {
			t.Errorf("2件目のバックフィル結果が一致しません: got=%+v", rows[1])
		}
	})
}

func loadSQLFile(t *testing.T, relPath string) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("テストファイルの位置を取得できません")
	}
	rootDir := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
	path := filepath.Join(rootDir, filepath.Clean(relPath))
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("SQLファイルの読み込みに失敗しました: path=%s err=%v", path, err)
	}
	return string(b)
}
