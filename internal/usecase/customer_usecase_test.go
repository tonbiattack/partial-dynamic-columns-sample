// usecase_test パッケージは CustomerUsecase のテスト。
//
// テスト方針:
//   - 実 DB（MySQL）を使った統合テスト。リポジトリのモックは使わない。
//   - usecase → repository（GORM 実装） → MySQL の実際の動作を通しで検証する。
//   - テスト間のデータ干渉を防ぐため、各テスト関数の末尾で defer クリーンアップを呼ぶ。
package usecase_test

import (
	"testing"

	"github.com/yourname/partial-dynamic-columns-sample/internal/infrastructure/persistence"
	"github.com/yourname/partial-dynamic-columns-sample/internal/usecase"
	testhelper "github.com/yourname/partial-dynamic-columns-sample/test"
)

// TestCreateCustomer は顧客作成ユースケースのテスト。
// 正常系・デフォルト値・バリデーションエラーの3ケースを検証する。
func TestCreateCustomer(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	defer testhelper.CleanupCustomers(t, db)

	repo := persistence.NewCustomerRepository(db)
	uc := usecase.NewCustomerUsecase(repo)

	t.Run("正常系_新規顧客を作成できること", func(t *testing.T) {
		input := usecase.CreateCustomerInput{
			Name:   "テスト太郎",
			Email:  "test@example.com",
			Status: "active",
		}
		customer, err := uc.CreateCustomer(input)
		if err != nil {
			t.Fatalf("顧客作成でエラーが発生しました: %v", err)
		}
		// DB が採番した ID が反映されていることを確認する
		if customer.ID == 0 {
			t.Error("ID が採番されていません")
		}
		if customer.Name != input.Name {
			t.Errorf("名前が一致しません: want=%s, got=%s", input.Name, customer.Name)
		}
		if customer.Status != "active" {
			t.Errorf("ステータスが一致しません: want=active, got=%s", customer.Status)
		}
	})

	t.Run("正常系_ステータス未指定の場合はactiveが設定されること", func(t *testing.T) {
		// Status を空にして渡す → usecase 内でデフォルト値 "active" が設定される
		input := usecase.CreateCustomerInput{
			Name:  "ステータス省略テスト",
			Email: "nostatus@example.com",
		}
		customer, err := uc.CreateCustomer(input)
		if err != nil {
			t.Fatalf("顧客作成でエラーが発生しました: %v", err)
		}
		if customer.Status != "active" {
			t.Errorf("デフォルトステータスが一致しません: want=active, got=%s", customer.Status)
		}
	})

	t.Run("異常系_名前が空の場合はエラーになること", func(t *testing.T) {
		// Name が空 → usecase のバリデーションでエラーになる（DB には到達しない）
		input := usecase.CreateCustomerInput{
			Name:  "",
			Email: "empty@example.com",
		}
		_, err := uc.CreateCustomer(input)
		if err == nil {
			t.Error("名前が空のときにエラーが返されませんでした")
		}
	})
}

// TestGetCustomerByID は顧客取得ユースケースのテスト。
// 存在する ID での取得と、存在しない ID でのエラーを検証する。
func TestGetCustomerByID(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	defer testhelper.CleanupCustomers(t, db)

	repo := persistence.NewCustomerRepository(db)
	uc := usecase.NewCustomerUsecase(repo)

	t.Run("正常系_存在するIDで顧客を取得できること", func(t *testing.T) {
		// 事前にテスト用の顧客を作成する
		created, err := uc.CreateCustomer(usecase.CreateCustomerInput{
			Name:  "取得テスト用顧客",
			Email: "find@example.com",
		})
		if err != nil {
			t.Fatalf("事前データ作成に失敗しました: %v", err)
		}

		found, err := uc.GetCustomerByID(created.ID)
		if err != nil {
			t.Fatalf("顧客取得でエラーが発生しました: %v", err)
		}
		if found.ID != created.ID {
			t.Errorf("ID が一致しません: want=%d, got=%d", created.ID, found.ID)
		}
		if found.Name != created.Name {
			t.Errorf("名前が一致しません: want=%s, got=%s", created.Name, found.Name)
		}
	})

	t.Run("異常系_存在しないIDではエラーになること", func(t *testing.T) {
		// 存在しない ID → リポジトリが nil を返し、usecase がエラーに変換する
		_, err := uc.GetCustomerByID(999999)
		if err == nil {
			t.Error("存在しない ID でエラーが返されませんでした")
		}
	})
}

// TestUpdateCustomerExtraJSON は extra_json 更新ユースケースのテスト。
// JSON パターンの典型的な操作（キーと値の追加）を検証する。
func TestUpdateCustomerExtraJSON(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	defer testhelper.CleanupCustomers(t, db)

	repo := persistence.NewCustomerRepository(db)
	uc := usecase.NewCustomerUsecase(repo)

	t.Run("正常系_extra_jsonにキーと値を追加できること", func(t *testing.T) {
		// 事前に顧客を作成する（extra_json は空）
		created, err := uc.CreateCustomer(usecase.CreateCustomerInput{
			Name:  "JSON更新テスト",
			Email: "json@example.com",
		})
		if err != nil {
			t.Fatalf("事前データ作成に失敗しました: %v", err)
		}

		// extra_json に "industry" キーで "IT" をセットして保存する
		updated, err := uc.UpdateCustomerExtraJSON(created.ID, "industry", "IT")
		if err != nil {
			t.Fatalf("extra_json 更新でエラーが発生しました: %v", err)
		}

		// 更新後のオブジェクトで値を検証する
		val, ok := updated.GetExtra("industry")
		if !ok {
			t.Error("extra_json にキー 'industry' が存在しません")
		}
		if val != "IT" {
			t.Errorf("extra_json の値が一致しません: want=IT, got=%v", val)
		}
	})
}
