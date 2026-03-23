SET NAMES utf8mb4;

-- 目的: customers.extra_json に保存した値を custom_field_values にバックフィルする
-- 想定ユースケース: JSON で保持していた補足項目のうち、検索・集計対象になったキーを EAV に移行したい場合
-- 主要な出力項目: なし（INSERT ... SELECT / UPSERT のみ実行）
-- 実装方針:
--   - 先に custom_field_definitions に対象キーを定義する
--   - customers.extra_json から JSON_EXTRACT で値を取り出す
--   - (entity_id, field_definition_id) のユニーク制約を使って ON DUPLICATE KEY UPDATE で冪等に流し込む

-- 例: tenant_id=1 / entity_type='customer' の industry, contract_rank を EAV に移す

INSERT INTO custom_field_definitions (
    -- 対象テナントID
    tenant_id,
    -- 対象エンティティ種別
    entity_type,
    -- アプリ内で使うフィールドキー
    field_key,
    -- 画面表示用ラベル
    field_label,
    -- 値の型
    field_type,
    -- 必須入力かどうか
    is_required,
    -- 検索対象に含めるかどうか
    is_searchable,
    -- 画面表示順
    display_order
)
-- industry 定義が未登録なら初回だけ作成する
SELECT 1, 'customer', 'industry', '業種', 'text', FALSE, TRUE, 1
WHERE NOT EXISTS (
    SELECT 1
    FROM custom_field_definitions
    WHERE tenant_id = 1
      AND entity_type = 'customer'
      AND field_key = 'industry'
);

INSERT INTO custom_field_definitions (
    -- 対象テナントID
    tenant_id,
    -- 対象エンティティ種別
    entity_type,
    -- アプリ内で使うフィールドキー
    field_key,
    -- 画面表示用ラベル
    field_label,
    -- 値の型
    field_type,
    -- 必須入力かどうか
    is_required,
    -- 検索対象に含めるかどうか
    is_searchable,
    -- 画面表示順
    display_order
)
-- contract_rank 定義が未登録なら初回だけ作成する
SELECT 1, 'customer', 'contract_rank', '契約ランク', 'text', FALSE, TRUE, 2
WHERE NOT EXISTS (
    SELECT 1
    FROM custom_field_definitions
    WHERE tenant_id = 1
      AND entity_type = 'customer'
      AND field_key = 'contract_rank'
);

INSERT INTO custom_field_values (
    -- 値が属する顧客ID
    entity_id,
    -- industry 定義ID
    field_definition_id,
    -- 業種の文字列値
    value_text
)
SELECT
    -- 顧客IDをそのまま使う
    c.id,
    -- industry の定義ID
    d.id,
    -- extra_json から業種を取り出して value_text にセットする
    JSON_UNQUOTE(JSON_EXTRACT(c.extra_json, '$.industry'))
FROM customers c
-- 事前に作成済みの industry 定義を結合する
INNER JOIN custom_field_definitions d
    ON d.tenant_id = 1
   AND d.entity_type = 'customer'
   AND d.field_key = 'industry'
-- industry を持つ顧客だけを対象にする
WHERE JSON_EXTRACT(c.extra_json, '$.industry') IS NOT NULL
-- 既存値があれば上書きし、何度流しても同じ結果になるようにする
ON DUPLICATE KEY UPDATE
    value_text = VALUES(value_text);

INSERT INTO custom_field_values (
    -- 値が属する顧客ID
    entity_id,
    -- contract_rank 定義ID
    field_definition_id,
    -- 契約ランクの文字列値
    value_text
)
SELECT
    -- 顧客IDをそのまま使う
    c.id,
    -- contract_rank の定義ID
    d.id,
    -- extra_json から契約ランクを取り出して value_text にセットする
    JSON_UNQUOTE(JSON_EXTRACT(c.extra_json, '$.contract_rank'))
FROM customers c
-- 事前に作成済みの contract_rank 定義を結合する
INNER JOIN custom_field_definitions d
    ON d.tenant_id = 1
   AND d.entity_type = 'customer'
   AND d.field_key = 'contract_rank'
-- contract_rank を持つ顧客だけを対象にする
WHERE JSON_EXTRACT(c.extra_json, '$.contract_rank') IS NOT NULL
-- 既存値があれば上書きし、何度流しても同じ結果になるようにする
ON DUPLICATE KEY UPDATE
    value_text = VALUES(value_text);
