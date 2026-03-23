-- 目的: JSON に保持した業種と契約ランクで顧客一覧を取得する
-- 想定ユースケース: 詳細画面でしか使っていなかった補足項目が、一覧検索や並び替えに使われ始めた場面
-- 主要な出力項目: 顧客ID、顧客名、業種、契約ランク
-- 実装方針:
--   - customers.extra_json から業種と契約ランクを取り出す
--   - JSON の値を直接 WHERE / ORDER BY に使うことで、JSON検索の書き味を示す

SELECT
    -- 顧客の主キー。画面遷移や詳細取得に使うことを想定
    id,
    -- 固定カラムとして保持している顧客名
    name,
    -- extra_json から業種を取り出して一覧表示用の列にする
    JSON_UNQUOTE(JSON_EXTRACT(extra_json, '$.industry')) AS industry,
    -- extra_json から契約ランクを取り出して一覧表示用の列にする
    JSON_UNQUOTE(JSON_EXTRACT(extra_json, '$.contract_rank')) AS contract_rank
FROM customers
-- JSON 内の industry が 'IT' の顧客だけに絞り込む
WHERE JSON_UNQUOTE(JSON_EXTRACT(extra_json, '$.industry')) = 'IT'
-- 契約ランク順で並べ、同ランク内は id 順にして結果を安定させる
ORDER BY JSON_UNQUOTE(JSON_EXTRACT(extra_json, '$.contract_rank')) ASC, id ASC;
