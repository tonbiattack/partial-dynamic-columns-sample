SET NAMES utf8mb4;

-- 目的: 3パターン（固定カラムのみ / 固定+JSON / 固定+EAV）のサンプルスキーマ定義
-- 出力: 顧客テーブル・カスタムフィールド定義テーブル・カスタムフィールド値テーブル
-- 実装方針: customers に extra_json を持たせることでJSONパターンをカバーし、
--           custom_field_definitions / custom_field_values でEAVパターンをカバーする

-- 顧客テーブル（固定カラム + JSON カラムを含む）
CREATE TABLE IF NOT EXISTS customers (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '顧客ID',
    name       VARCHAR(255)    NOT NULL COMMENT '顧客名',
    email      VARCHAR(255)    NOT NULL DEFAULT '' COMMENT 'メールアドレス',
    status     VARCHAR(50)     NOT NULL DEFAULT 'active' COMMENT '顧客ステータス',
    extra_json JSON            NOT NULL DEFAULT ('{}') COMMENT '可変の補足項目を保持するJSON',
    notes      TEXT            COMMENT '自由記述の備考欄',
    created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '作成日時',
    PRIMARY KEY (id),
    INDEX idx_customers_status (status),
    INDEX idx_customers_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='顧客の固定項目とJSON補足項目を保持するテーブル';

-- カスタムフィールド定義テーブル（EAVパターン）
-- テナントごと・エンティティ種別ごとに追加フィールドを定義できる
CREATE TABLE IF NOT EXISTS custom_field_definitions (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'カスタムフィールド定義ID',
    tenant_id     BIGINT UNSIGNED NOT NULL COMMENT 'テナントID',
    entity_type   VARCHAR(100)    NOT NULL COMMENT '対象エンティティ種別',
    field_key     VARCHAR(100)    NOT NULL COMMENT 'アプリケーション内部で使うキー名',
    field_label   VARCHAR(255)    NOT NULL COMMENT '画面表示用の項目名',
    field_type    ENUM('text','number','date','boolean') NOT NULL DEFAULT 'text' COMMENT 'フィールド値の型',
    is_required   BOOLEAN         NOT NULL DEFAULT FALSE COMMENT '必須入力かどうか',
    is_searchable BOOLEAN         NOT NULL DEFAULT FALSE COMMENT '検索条件として利用可能かどうか',
    display_order INT             NOT NULL DEFAULT 0 COMMENT '画面表示順',
    created_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '作成日時',
    PRIMARY KEY (id),
    UNIQUE KEY uq_tenant_entity_field (tenant_id, entity_type, field_key),
    INDEX idx_cfd_tenant_entity (tenant_id, entity_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='EAVパターンのカスタムフィールド定義を保持するテーブル';

-- カスタムフィールド値テーブル（EAVパターン）
-- フィールド型ごとに専用カラムを用意し、型安全に値を格納する
CREATE TABLE IF NOT EXISTS custom_field_values (
    id                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'カスタムフィールド値ID',
    entity_id           BIGINT UNSIGNED NOT NULL COMMENT '値が紐づくエンティティID',
    field_definition_id BIGINT UNSIGNED NOT NULL COMMENT '対応するカスタムフィールド定義ID',
    value_text          TEXT            COMMENT '文字列型の値',
    value_number        DECIMAL(20,4)   COMMENT '数値型の値',
    value_date          DATE            COMMENT '日付型の値',
    value_boolean       BOOLEAN         COMMENT '真偽値型の値',
    created_at          DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '作成日時',
    PRIMARY KEY (id),
    UNIQUE KEY uq_entity_field (entity_id, field_definition_id),
    INDEX idx_cfv_entity (entity_id),
    INDEX idx_cfv_definition (field_definition_id),
    CONSTRAINT fk_cfv_definition
        FOREIGN KEY (field_definition_id)
        REFERENCES custom_field_definitions (id)
        ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='EAVパターンのカスタムフィールド値を保持するテーブル';
