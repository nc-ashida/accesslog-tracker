-- 初期データベーススキーマ
-- 作成日: 2024年12月
-- 説明: accesslog-trackerの初期テーブル作成

-- アプリケーションテーブル
CREATE TABLE IF NOT EXISTS applications (
    app_id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(255) NOT NULL,
    api_key VARCHAR(255) UNIQUE NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- トラッキングデータテーブル
CREATE TABLE IF NOT EXISTS tracking_data (
    id VARCHAR(255) PRIMARY KEY,
    app_id VARCHAR(255) NOT NULL,
    client_sub_id VARCHAR(255),
    module_id VARCHAR(255),
    url TEXT,
    referrer TEXT,
    user_agent TEXT NOT NULL,
    ip_address INET,
    session_id VARCHAR(255),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    custom_params JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (app_id) REFERENCES applications(app_id) ON DELETE CASCADE
);

-- インデックスの作成
CREATE INDEX IF NOT EXISTS idx_tracking_data_app_id ON tracking_data(app_id);
CREATE INDEX IF NOT EXISTS idx_tracking_data_timestamp ON tracking_data(timestamp);
CREATE INDEX IF NOT EXISTS idx_tracking_data_session_id ON tracking_data(session_id);
CREATE INDEX IF NOT EXISTS idx_tracking_data_ip_address ON tracking_data(ip_address);
CREATE INDEX IF NOT EXISTS idx_applications_api_key ON applications(api_key);
CREATE INDEX IF NOT EXISTS idx_applications_domain ON applications(domain);

-- パーティショニング用のインデックス（将来の拡張用）
CREATE INDEX IF NOT EXISTS idx_tracking_data_app_timestamp ON tracking_data(app_id, timestamp);

-- 統計情報用のビュー
CREATE OR REPLACE VIEW tracking_stats AS
SELECT 
    app_id,
    COUNT(*) as total_requests,
    COUNT(DISTINCT session_id) as unique_sessions,
    COUNT(DISTINCT ip_address) as unique_ips,
    COUNT(CASE WHEN user_agent ILIKE '%bot%' OR user_agent ILIKE '%crawler%' THEN 1 END) as bot_requests,
    COUNT(CASE WHEN user_agent ILIKE '%mobile%' OR user_agent ILIKE '%android%' OR user_agent ILIKE '%iphone%' THEN 1 END) as mobile_requests,
    MIN(timestamp) as first_request,
    MAX(timestamp) as last_request
FROM tracking_data
GROUP BY app_id;

-- セッション統計用のビュー
CREATE OR REPLACE VIEW session_stats AS
SELECT 
    app_id,
    session_id,
    COUNT(*) as page_views,
    MIN(timestamp) as session_start,
    MAX(timestamp) as session_end,
    EXTRACT(EPOCH FROM (MAX(timestamp) - MIN(timestamp))) as session_duration_seconds
FROM tracking_data
WHERE session_id IS NOT NULL
GROUP BY app_id, session_id;

-- 更新日時の自動更新トリガー
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_applications_updated_at 
    BEFORE UPDATE ON applications 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- コメントの追加
COMMENT ON TABLE applications IS 'アプリケーション情報を管理するテーブル';
COMMENT ON TABLE tracking_data IS 'トラッキングデータを保存するテーブル';
COMMENT ON VIEW tracking_stats IS 'トラッキング統計情報を提供するビュー';
COMMENT ON VIEW session_stats IS 'セッション統計情報を提供するビュー';
