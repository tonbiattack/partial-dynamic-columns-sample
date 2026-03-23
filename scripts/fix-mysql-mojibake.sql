SET NAMES utf8mb4;

-- 目的: latin1 として取り込まれて文字化けした日本語サンプルデータを utf8mb4 として復元する
-- 想定ユースケース: 既存の MySQL コンテナで docker/mysql-init/*.sql が誤った接続文字コードで投入された場合の復旧
-- 主要な出力項目: なし（UPDATE のみ実行）

UPDATE customers
SET
    name = CONVERT(CAST(CONVERT(name USING latin1) AS BINARY) USING utf8mb4),
    notes = CASE
        WHEN notes IS NULL THEN NULL
        ELSE CONVERT(CAST(CONVERT(notes USING latin1) AS BINARY) USING utf8mb4)
    END;

UPDATE customers
SET extra_json = JSON_SET(
    extra_json,
    '$.industry',
    CONVERT(CAST(CONVERT(JSON_UNQUOTE(JSON_EXTRACT(extra_json, '$.industry')) USING latin1) AS BINARY) USING utf8mb4)
)
WHERE JSON_EXTRACT(extra_json, '$.industry') IS NOT NULL;

UPDATE customers
SET extra_json = JSON_SET(
    extra_json,
    '$.sales_rep',
    CONVERT(CAST(CONVERT(JSON_UNQUOTE(JSON_EXTRACT(extra_json, '$.sales_rep')) USING latin1) AS BINARY) USING utf8mb4)
)
WHERE JSON_EXTRACT(extra_json, '$.sales_rep') IS NOT NULL;

UPDATE custom_field_definitions
SET field_label = CONVERT(CAST(CONVERT(field_label USING latin1) AS BINARY) USING utf8mb4);

UPDATE custom_field_values
SET value_text = CONVERT(CAST(CONVERT(value_text USING latin1) AS BINARY) USING utf8mb4)
WHERE value_text IS NOT NULL;
