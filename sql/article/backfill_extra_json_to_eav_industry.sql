-- 目的: customers.extra_json の industry を custom_field_values へバックフィルする
-- 想定ユースケース: JSON で保持していた可変項目のうち、検索対象になったキーを EAV に移す場面
-- 主要な出力項目: なし（INSERT ... SELECT / UPSERT のみ実行）
-- 実装方針:
--   - custom_field_definitions の industry 定義を参照する
--   - customers.extra_json から industry を取り出して custom_field_values に流し込む
--   - (entity_id, field_definition_id) のユニーク制約を使って冪等に更新する

INSERT INTO custom_field_values (
    -- 値が属する顧客ID
    entity_id,
    -- industry フィールド定義のID
    field_definition_id,
    -- extra_json から取り出した業種文字列
    value_text
)
SELECT
    -- 顧客ごとのIDをそのまま値テーブルへ移す
    c.id,
    -- tenant_id=1 / customer / industry に対応する定義IDを使う
    d.id,
    -- extra_json から industry を取り出して value_text に入れる
    JSON_UNQUOTE(JSON_EXTRACT(c.extra_json, '$.industry'))
FROM customers c
-- industry フィールド定義を先に特定してから値を組み立てる
INNER JOIN custom_field_definitions d
    ON d.tenant_id = 1
   AND d.entity_type = 'customer'
   AND d.field_key = 'industry'
-- industry を持つ顧客だけをバックフィル対象にする
WHERE JSON_EXTRACT(c.extra_json, '$.industry') IS NOT NULL
-- 既に値が存在する場合は上書きし、何度流しても同じ状態になるようにする
ON DUPLICATE KEY UPDATE
    value_text = VALUES(value_text);
