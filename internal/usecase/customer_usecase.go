// usecase パッケージはアプリケーション固有のビジネスロジックを実装する。
//
// クリーンアーキテクチャにおいて usecase 層は：
//   - domain 層のモデルとインターフェースのみに依存する
//   - GORM・MySQL・Gin などのフレームワークには依存しない
//   - ハンドラー（interface 層）から呼び出され、リポジトリ（infrastructure 層）を操作する
//
// この設計により、DB の実装を変更しても usecase 層は変更不要になる。
package usecase

import (
	"fmt"

	"github.com/yourname/partial-dynamic-columns-sample/internal/domain"
	"github.com/yourname/partial-dynamic-columns-sample/internal/repository"
)

// CustomerUsecase は顧客に関するアプリケーションユースケースをまとめた構造体。
// repository.CustomerRepository インターフェースに依存するため、
// 実際の DB 実装に依存しない。
type CustomerUsecase struct {
	customerRepo repository.CustomerRepository
}

// NewCustomerUsecase は CustomerUsecase を生成する。
// repository.CustomerRepository インターフェースを受け取るため、
// テストでは実 DB に接続した実装を渡す。
func NewCustomerUsecase(r repository.CustomerRepository) *CustomerUsecase {
	return &CustomerUsecase{customerRepo: r}
}

// CreateCustomerInput は顧客作成ユースケースの入力パラメータ。
// ハンドラー層（HTTP リクエスト・CLI コマンドなど）から渡される。
type CreateCustomerInput struct {
	Name      string
	Email     string
	Status    string
	ExtraJSON map[string]interface{} // 補足項目（固定カラム + JSON パターン）
	Notes     string
}

// CreateCustomer は新規顧客を作成して返す。
//
// バリデーション:
//   - Name は必須。空の場合はエラーを返す。
//   - Status が未指定の場合はデフォルト値 "active" を使用する。
func (uc *CustomerUsecase) CreateCustomer(input CreateCustomerInput) (*domain.Customer, error) {
	// 必須項目のバリデーション
	if input.Name == "" {
		return nil, fmt.Errorf("顧客名は必須です")
	}
	// Status のデフォルト値設定
	status := input.Status
	if status == "" {
		status = "active"
	}

	c := &domain.Customer{
		Name:      input.Name,
		Email:     input.Email,
		Status:    status,
		ExtraJSON: input.ExtraJSON,
		Notes:     input.Notes,
	}
	if err := uc.customerRepo.Create(c); err != nil {
		return nil, fmt.Errorf("顧客の作成に失敗しました: %w", err)
	}
	return c, nil
}

// GetCustomerByID は指定 ID の顧客を取得する。
// 顧客が存在しない場合はエラーを返す。
// （リポジトリの FindByID が nil を返した場合を「見つからない」として扱う）
func (uc *CustomerUsecase) GetCustomerByID(id uint) (*domain.Customer, error) {
	c, err := uc.customerRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("顧客の取得に失敗しました: %w", err)
	}
	if c == nil {
		return nil, fmt.Errorf("顧客が見つかりません: id=%d", id)
	}
	return c, nil
}

// GetAllCustomers はすべての顧客を取得する。
func (uc *CustomerUsecase) GetAllCustomers() ([]*domain.Customer, error) {
	customers, err := uc.customerRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("顧客一覧の取得に失敗しました: %w", err)
	}
	return customers, nil
}

// UpdateCustomerExtraJSON は顧客の extra_json に指定キーで値をセットして保存する。
//
// JSON パターンの典型的なユースケース：
// テナントごとに異なる補足項目（業種・担当者など）を、
// スキーマ変更なく追加・更新できる。
//
// 既存のキーがある場合は上書きし、新しいキーの場合は追加する。
// 他のキーの値は変わらない（SetExtra は map への部分的な書き込み）。
func (uc *CustomerUsecase) UpdateCustomerExtraJSON(id uint, key string, value interface{}) (*domain.Customer, error) {
	// 最新の状態を DB から取得する（extra_json の現在値を保持するため）
	c, err := uc.customerRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("顧客の取得に失敗しました: %w", err)
	}
	if c == nil {
		return nil, fmt.Errorf("顧客が見つかりません: id=%d", id)
	}
	c.SetExtra(key, value)
	if err := uc.customerRepo.Update(c); err != nil {
		return nil, fmt.Errorf("顧客の更新に失敗しました: %w", err)
	}
	return c, nil
}
