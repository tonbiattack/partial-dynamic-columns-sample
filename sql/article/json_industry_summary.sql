-- 目的: JSON に保持した業種ごとの顧客件数を集計する
-- 想定ユースケース: 補足項目として持っていた業種が、ダッシュボードや簡易集計でも使われ始めた場面
-- 主要な出力項目: 業種、顧客件数
-- 実装方針:
--   - customers.extra_json の industry を集計軸として使う
--   - GROUP BY / ORDER BY でも JSON パスの指定が必要になる例を示す

SELECT
    -- extra_json から業種を取り出し、集計結果のキーとして使う
    JSON_UNQUOTE(JSON_EXTRACT(extra_json, '$.industry')) AS industry,
    -- 同じ業種を持つ顧客件数を数える
    COUNT(*) AS customer_count
FROM customers
-- industry キーが存在する顧客だけを集計対象にする
WHERE JSON_EXTRACT(extra_json, '$.industry') IS NOT NULL
-- JSON から取り出した業種単位でグルーピングする
GROUP BY JSON_UNQUOTE(JSON_EXTRACT(extra_json, '$.industry'))
-- 件数の多い順に並べ、同件数なら業種名で安定化する
ORDER BY customer_count DESC, industry ASC;
