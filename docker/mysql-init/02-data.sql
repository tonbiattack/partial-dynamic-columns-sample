SET NAMES utf8mb4;

-- 目的: サンプルデータの投入（テナントID=1 で顧客3件・フィールド定義3件・値を挿入）
-- 注意: 冪等性確保のため INSERT IGNORE を使用する

-- 顧客サンプルデータ（固定カラム+JSON）
INSERT IGNORE INTO customers (id, name, email, status, extra_json, notes) VALUES
(1, '田中 太郎', 'tanaka@example.com', 'active',
    '{"industry": "IT", "contract_rank": "A", "sales_rep": "山田"}',
    '大口顧客。四半期ごとに定期訪問。'),
(2, '鈴木 花子', 'suzuki@example.com', 'active',
    '{"industry": "製造", "contract_rank": "B"}',
    NULL),
(3, '佐藤 次郎', 'sato@example.com', 'inactive',
    '{}',
    '解約済み。再獲得候補。');

-- カスタムフィールド定義（テナントID=1、エンティティ種別=customer）
INSERT IGNORE INTO custom_field_definitions
    (id, tenant_id, entity_type, field_key, field_label, field_type, is_required, is_searchable, display_order)
VALUES
(1, 1, 'customer', 'industry',       '業種',           'text',    FALSE, TRUE,  1),
(2, 1, 'customer', 'contract_rank',  '契約ランク',     'text',    FALSE, TRUE,  2),
(3, 1, 'customer', 'annual_revenue', '年間売上(万円)', 'number',  FALSE, FALSE, 3);

-- カスタムフィールド値（EAVパターン）
-- 田中太郎（entity_id=1）の値
INSERT IGNORE INTO custom_field_values (entity_id, field_definition_id, value_text, value_number, value_date, value_boolean)
VALUES
(1, 1, 'IT',   NULL,     NULL, NULL),
(1, 2, 'A',    NULL,     NULL, NULL),
(1, 3, NULL,   50000.00, NULL, NULL);

-- 鈴木花子（entity_id=2）の値
INSERT IGNORE INTO custom_field_values (entity_id, field_definition_id, value_text, value_number, value_date, value_boolean)
VALUES
(2, 1, '製造', NULL,     NULL, NULL),
(2, 2, 'B',   NULL,     NULL, NULL);
