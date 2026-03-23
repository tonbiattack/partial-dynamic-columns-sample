// integration_test パッケージは実 DB を使った統合テスト。
//
// このファイルでは「固定カラム + JSON」パターンを検証する。
//
// 検証内容:
//   - extra_json に複数キーを保存し、DB から再取得して値が正しいことを確認する
//   - extra_json が空の顧客も正常に登録・取得できることを確認する
//   - extra_json の値を更新（上書き・追加）できることを確認する
//
// テスト方針:
//   - 実 DB（MySQL）に接続し、実際の JSON 型カラムの動作を通しで検証する
//   - 各テストケースは独立して実行できるよう、テスト終了時にデータをクリーンアップする
package integration_test

import (
	"testing"

	"github.com/yourname/partial-dynamic-columns-sample/internal/domain"
	"github.com/yourname/partial-dynamic-columns-sample/internal/infrastructure/persistence"
	testhelper "github.com/yourname/partial-dynamic-columns-sample/test"
)

// TestCustomerWithExtraJSON は固定カラム + JSON パターンの統合テスト。
// extra_json に値を保存し、正しく取得・更新できることを検証する。
func TestCustomerWithExtraJSON(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	defer testhelper.CleanupCustomers(t, db)

	repo := persistence.NewCustomerRepository(db)

	t.Run("JSONパターン_extra_jsonに複数キーを保存して取得できること", func(t *testing.T) {
		// extra_json に業種・ランク・担当者の3つのキーを持つ顧客を作成する
		c := &domain.Customer{
			Name:   "JSONテスト株式会社",
			Email:  "json-test@example.com",
			Status: "active",
			ExtraJSON: map[string]interface{}{
				"industry":      "IT",
				"contract_rank": "A",
				"sales_rep":     "山田",
			},
			Notes: "extra_json パターンの統合テスト",
		}

		if err := repo.Create(c); err != nil {
			t.Fatalf("顧客の作成に失敗しました: %v", err)
		}
		if c.ID == 0 {
			t.Fatal("ID が採番されていません")
		}

		// DB から再取得することで、JSONMap の Value()/Scan() を通じた
		// シリアライズ・デシリアライズが正しく動作するかを確認する
		found, err := repo.FindByID(c.ID)
		if err != nil {
			t.Fatalf("顧客の取得に失敗しました: %v", err)
		}
		if found == nil {
			t.Fatal("顧客が見つかりません")
		}

		// 固定カラムの値が正しく保存されていることを確認する
		if found.Name != c.Name {
			t.Errorf("名前が一致しません: want=%s, got=%s", c.Name, found.Name)
		}
		if found.Status != c.Status {
			t.Errorf("ステータスが一致しません: want=%s, got=%s", c.Status, found.Status)
		}

		// extra_json の各キーが正しく保存・復元されていることを確認する
		checkExtraJSON(t, found, "industry", "IT")
		checkExtraJSON(t, found, "contract_rank", "A")
		checkExtraJSON(t, found, "sales_rep", "山田")
	})

	t.Run("JSONパターン_extra_jsonが空の顧客を登録・取得できること", func(t *testing.T) {
		// extra_json が nil の顧客 → DB には "{}" として保存される
		c := &domain.Customer{
			Name:      "extra_json空テスト",
			Email:     "empty-json@example.com",
			Status:    "inactive",
			ExtraJSON: nil,
		}

		if err := repo.Create(c); err != nil {
			t.Fatalf("顧客の作成に失敗しました: %v", err)
		}

		found, err := repo.FindByID(c.ID)
		if err != nil {
			t.Fatalf("顧客の取得に失敗しました: %v", err)
		}
		// "{}" から復元された場合は空 map になる（len=0）ことを確認する
		if len(found.ExtraJSON) != 0 {
			t.Errorf("extra_json が空であるべきですが値が存在します: %v", found.ExtraJSON)
		}
	})

	t.Run("JSONパターン_extra_jsonの値を更新できること", func(t *testing.T) {
		// 初期状態: "industry" = "製造" の顧客を作成する
		c := &domain.Customer{
			Name:   "JSON更新テスト",
			Email:  "json-update@example.com",
			Status: "active",
			ExtraJSON: map[string]interface{}{
				"industry": "製造",
			},
		}

		if err := repo.Create(c); err != nil {
			t.Fatalf("顧客の作成に失敗しました: %v", err)
		}

		// "industry" を "IT" に上書きし、新しいキー "contract_rank" を追加する
		c.SetExtra("industry", "IT")
		c.SetExtra("contract_rank", "S")
		if err := repo.Update(c); err != nil {
			t.Fatalf("顧客の更新に失敗しました: %v", err)
		}

		// DB から再取得して更新後の値を確認する
		found, err := repo.FindByID(c.ID)
		if err != nil {
			t.Fatalf("更新後の顧客取得に失敗しました: %v", err)
		}

		checkExtraJSON(t, found, "industry", "IT")
		checkExtraJSON(t, found, "contract_rank", "S")
	})
}

// checkExtraJSON は extra_json の指定キーに期待する文字列値が格納されているかを検証するヘルパー。
// t.Helper() を使うことで、このヘルパー内でのエラーが呼び出し元のテスト行に表示される。
func checkExtraJSON(t *testing.T, c *domain.Customer, key, want string) {
	t.Helper()
	val, ok := c.GetExtra(key)
	if !ok {
		t.Errorf("extra_json にキー '%s' が存在しません", key)
		return
	}
	got, ok := val.(string)
	if !ok {
		t.Errorf("extra_json[%s] の型が string ではありません: %T", key, val)
		return
	}
	if got != want {
		t.Errorf("extra_json[%s] の値が一致しません: want=%s, got=%s", key, want, got)
	}
}
