package persistence

import (
	"errors"

	"github.com/yourname/partial-dynamic-columns-sample/internal/domain"
	"github.com/yourname/partial-dynamic-columns-sample/internal/repository"
	"gorm.io/gorm"
)

// customerRepository は repository.CustomerRepository インターフェースの GORM 実装。
// 外部からは NewCustomerRepository() を通じてのみ生成し、
// 型を repository.CustomerRepository として返すことで実装の詳細を隠蔽する。
type customerRepository struct {
	db *gorm.DB
}

// NewCustomerRepository は CustomerRepository の実装を生成して返す。
// 戻り値の型をインターフェースにすることで、呼び出し元が具体的な実装に依存しないようにする。
func NewCustomerRepository(db *gorm.DB) repository.CustomerRepository {
	return &customerRepository{db: db}
}

// FindByID は指定 ID の顧客を取得する。
// レコードが存在しない場合は (nil, nil) を返す（エラーにはしない）。
// 「見つからない」と「DB エラー」を呼び出し元で区別できるようにするための設計。
func (r *customerRepository) FindByID(id uint) (*domain.Customer, error) {
	var model CustomerModel
	result := r.db.First(&model, id)
	if result.Error != nil {
		// gorm.ErrRecordNotFound はレコード未存在を示す。エラーではなく nil を返す。
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return toDomainCustomer(&model), nil
}

// FindAll はすべての顧客を ID 昇順で取得する。
// テスト時の比較を安定させるため ORDER BY id ASC を必ず指定する。
func (r *customerRepository) FindAll() ([]*domain.Customer, error) {
	var models []CustomerModel
	result := r.db.Order("id ASC").Find(&models)
	if result.Error != nil {
		return nil, result.Error
	}
	customers := make([]*domain.Customer, 0, len(models))
	for i := range models {
		customers = append(customers, toDomainCustomer(&models[i]))
	}
	return customers, nil
}

// Create は新規顧客を DB に INSERT する。
// GORM の Create は INSERT 後に AUTO_INCREMENT で採番された ID を model.ID にセットするため、
// それをドメインモデルにも反映して呼び出し元が ID を参照できるようにする。
func (r *customerRepository) Create(c *domain.Customer) error {
	model := toCustomerModel(c)
	result := r.db.Create(model)
	if result.Error != nil {
		return result.Error
	}
	// DB が採番した ID と作成日時をドメインモデルに反映する
	c.ID = model.ID
	c.CreatedAt = model.CreatedAt
	return nil
}

// Update は顧客情報を全フィールド上書きで保存する。
// GORM の Save は主キーが存在する場合に UPDATE、存在しない場合に INSERT を行う。
// extra_json を含むすべてのフィールドが更新対象になる。
func (r *customerRepository) Update(c *domain.Customer) error {
	model := toCustomerModel(c)
	result := r.db.Save(model)
	return result.Error
}

// toDomainCustomer は GORM モデルをドメインモデルに変換する。
// 永続化の詳細（JSONMap 型など）をドメイン層に漏らさないために、
// インフラ層内部でのみ使用するプライベート関数として定義する。
func toDomainCustomer(m *CustomerModel) *domain.Customer {
	return &domain.Customer{
		ID:        m.ID,
		Name:      m.Name,
		Email:     m.Email,
		Status:    m.Status,
		ExtraJSON: map[string]interface{}(m.ExtraJSON), // JSONMap → map に変換
		Notes:     m.Notes,
		CreatedAt: m.CreatedAt,
	}
}

// toCustomerModel はドメインモデルを GORM モデルに変換する。
// domain.Customer の ExtraJSON（map）を JSONMap 型にキャストして渡す。
// JSONMap は driver.Valuer を実装しているため、GORM が DB への書き込み時に
// 自動的に JSON 文字列へシリアライズする。
func toCustomerModel(c *domain.Customer) *CustomerModel {
	return &CustomerModel{
		ID:        c.ID,
		Name:      c.Name,
		Email:     c.Email,
		Status:    c.Status,
		ExtraJSON: JSONMap(c.ExtraJSON), // map → JSONMap にキャスト
		Notes:     c.Notes,
		CreatedAt: c.CreatedAt,
	}
}
