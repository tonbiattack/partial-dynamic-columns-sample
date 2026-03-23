-- 目的: EAV に移した業種と契約ランクで顧客を絞り込む
-- 想定ユースケース: 可変項目が検索条件の主役になり、JSON ではなく EAV で扱うようになった場面
-- 主要な出力項目: 顧客ID、顧客名
-- 実装方針:
--   - custom_field_definitions で対象フィールドを特定する
--   - custom_field_values をフィールドごとに結合し、業種=IT かつ 契約ランク=A を満たす顧客を抽出する

SELECT
    -- 顧客の主キー
    c.id,
    -- 顧客の固定カラム名
    c.name
FROM customers c
-- まず tenant_id=1 / customer 向けの industry 定義を特定する
INNER JOIN custom_field_definitions d1
    ON d1.tenant_id = 1
   AND d1.entity_type = 'customer'
   AND d1.field_key = 'industry'
-- 上で特定した industry 定義に対応する値を顧客ごとに結合する
INNER JOIN custom_field_values v1
    ON v1.entity_id = c.id
   AND v1.field_definition_id = d1.id
-- 同様に contract_rank 定義を特定する
INNER JOIN custom_field_definitions d2
    ON d2.tenant_id = 1
   AND d2.entity_type = 'customer'
   AND d2.field_key = 'contract_rank'
-- 上で特定した contract_rank の値を顧客ごとに結合する
INNER JOIN custom_field_values v2
    ON v2.entity_id = c.id
   AND v2.field_definition_id = d2.id
-- industry='IT' かつ contract_rank='A' を両方満たす顧客だけを返す
WHERE v1.value_text = 'IT'
  AND v2.value_text = 'A'
-- テストや比較で安定するよう id 順で返す
ORDER BY c.id ASC;
