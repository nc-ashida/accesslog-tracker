-- スキーマ更新マイグレーション（仕様書準拠）
-- 作成日: 2024-01-15
-- 説明: API仕様書に準拠したスキーマ更新

-- アプリケーションテーブルの更新
ALTER TABLE applications 
ADD COLUMN app_id VARCHAR(64) UNIQUE NOT NULL DEFAULT uuid_generate_v4()::text,
ADD COLUMN client_sub_id VARCHAR(64),
ADD COLUMN module_id VARCHAR(64),
ADD COLUMN domain VARCHAR(255),
ADD COLUMN url VARCHAR(500);

-- 既存のレコードに対してapp_idを設定
UPDATE applications SET app_id = id::text WHERE app_id IS NULL;

-- トラッキングテーブルの更新
ALTER TABLE tracking 
ADD COLUMN client_sub_id VARCHAR(64),
ADD COLUMN module_id VARCHAR(64),
ADD COLUMN screen_resolution VARCHAR(20),
ADD COLUMN custom_params JSONB DEFAULT '{}';

-- 既存のcustom_parametersをcustom_paramsに移行
UPDATE tracking SET custom_params = custom_parameters WHERE custom_params IS NULL;

-- 不要なカラムを削除（後で削除）
-- ALTER TABLE tracking DROP COLUMN custom_parameters;

-- 新しいインデックスを作成
CREATE INDEX idx_applications_app_id ON applications(app_id);
CREATE INDEX idx_tracking_client_sub_id ON tracking(client_sub_id);
CREATE INDEX idx_tracking_module_id ON tracking(module_id);
CREATE INDEX idx_tracking_screen_resolution ON tracking(screen_resolution);

-- コメントを追加
COMMENT ON COLUMN applications.app_id IS 'アプリケーションID（仕様書準拠）';
COMMENT ON COLUMN applications.client_sub_id IS 'クライアントサブID';
COMMENT ON COLUMN applications.module_id IS 'モジュールID';
COMMENT ON COLUMN tracking.client_sub_id IS 'クライアントサブID';
COMMENT ON COLUMN tracking.module_id IS 'モジュールID';
COMMENT ON COLUMN tracking.screen_resolution IS '画面解像度';
COMMENT ON COLUMN tracking.custom_params IS 'カスタムパラメータ（仕様書準拠）';
