# partial-dynamic-columns-sample

ビジネスアプリで「一部だけ可変カラム」を安全に設計する方法のサンプルプロジェクトです。

## このサンプルで扱う設計パターン

- 固定カラムのみ（顧客の主要項目）
- 固定カラム + JSON（補足情報を `extra_json` に持つ）
- 固定カラム + EAV（カスタムフィールドを定義・管理できる）
- JSON から EAV への移行例（`scripts/backfill-extra-json-to-eav.sql`）

## 技術スタック

- Go 1.22
- GORM + MySQL 8.0
- Docker Compose

## ディレクトリ構成

```
.
├── docker/
│   └── mysql-init/
│       ├── 01-schema.sql        # テーブル定義
│       └── 02-data.sql          # サンプルデータ
├── internal/
│   ├── domain/
│   │   └── customer.go          # ドメインモデル
│   ├── repository/
│   │   └── customer_repository.go  # リポジトリインターフェース
│   ├── usecase/
│   │   ├── customer_usecase.go
│   │   └── customer_usecase_test.go
│   └── infrastructure/
│       └── persistence/
│           ├── customer_model.go           # GORMモデル
│           ├── customer_repository.go      # CustomerRepository実装
│           └── custom_field_repository.go  # CustomFieldRepository実装
├── test/
│   ├── helper.go                # テストヘルパー
│   └── integration/
│       ├── customer_json_test.go   # JSONパターンの統合テスト
│       └── customer_eav_test.go    # EAVパターンの統合テスト
├── docker-compose.yml
└── go.mod
```

## セットアップ

### 1. MySQL を起動する

```bash
docker compose up -d
```

MySQL は `utf8mb4` で起動するよう固定しています。日本語のサンプルデータが文字化けした場合は、初期化済みコンテナを削除して再作成してください。

```bash
docker compose down
docker compose up -d --force-recreate
```

既存DBのデータだけを復旧したい場合は、次の復旧SQLを流してください。

```bash
cat scripts/fix-mysql-mojibake.sql | docker exec -i <mysql-container> mysql -uroot -ppassword --default-character-set=utf8mb4 partial_dynamic_columns
```

### 2. 依存モジュールをダウンロードする

```bash
go mod tidy
```

### 3. テストを実行する

```bash
go test ./...
```

テスト用の DSN を変更する場合は環境変数で指定します。

```bash
TEST_DSN="user:password@tcp(host:port)/dbname?parseTime=true" go test ./...
```

日本語データを扱う DSN では `charset=utf8mb4` を付ける前提にしてください。

## テスト方針

- 実DB（MySQL）を使った統合テストで動作を検証する
- プロセス内処理（リポジトリ、サービス層）はモックしない
- テスト関数名は英語、`t.Run` の表示名は日本語

## 補助スクリプト

- `scripts/fix-mysql-mojibake.sql`: 文字化けした日本語サンプルデータの復旧
- `scripts/backfill-extra-json-to-eav.sql`: `customers.extra_json` の値を `custom_field_values` へバックフィルする移行例

## 記事対応SQL

- `sql/article/json_industry_customers.sql`: JSON を使った検索SQL
- `sql/article/json_industry_summary.sql`: JSON を使った集計SQL
- `sql/article/eav_industry_contract_rank_customers.sql`: EAV を使った検索SQL
- `sql/article/backfill_extra_json_to_eav_industry.sql`: JSON から EAV へのバックフィルSQL
