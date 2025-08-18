-- テスト用データベース初期化スクリプト
-- 作成日: 2024年12月
-- 説明: テスト環境用のデータベーススキーマ初期化

-- アプリケーションテーブル
CREATE TABLE IF NOT EXISTS applications (
    app_id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    domain VARCHAR(255) NOT NULL,
    api_key VARCHAR(255) UNIQUE NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- アクセスログテーブル
CREATE TABLE IF NOT EXISTS access_logs (
    id VARCHAR(255) PRIMARY KEY,
    app_id VARCHAR(255) NOT NULL,
    user_agent TEXT NOT NULL,
    url TEXT,
    ip_address INET,
    session_id VARCHAR(255),
    referrer TEXT,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    custom_params JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (app_id) REFERENCES applications(app_id) ON DELETE CASCADE
);

-- セッションテーブル
CREATE TABLE IF NOT EXISTS sessions (
    id VARCHAR(255) PRIMARY KEY,
    app_id VARCHAR(255) NOT NULL,
    session_id VARCHAR(255) UNIQUE NOT NULL,
    ip_address INET,
    user_agent TEXT,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_activity TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    page_views INTEGER DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (app_id) REFERENCES applications(app_id) ON DELETE CASCADE
);

-- カスタムパラメータテーブル
CREATE TABLE IF NOT EXISTS custom_parameters (
    id SERIAL PRIMARY KEY,
    access_log_id VARCHAR(255) NOT NULL,
    param_key VARCHAR(255) NOT NULL,
    param_value TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (access_log_id) REFERENCES access_logs(id) ON DELETE CASCADE
);

-- インデックスの作成
CREATE INDEX IF NOT EXISTS idx_access_logs_app_id ON access_logs(app_id);
CREATE INDEX IF NOT EXISTS idx_access_logs_timestamp ON access_logs(timestamp);
CREATE INDEX IF NOT EXISTS idx_access_logs_session_id ON access_logs(session_id);
CREATE INDEX IF NOT EXISTS idx_access_logs_ip_address ON access_logs(ip_address);
CREATE INDEX IF NOT EXISTS idx_applications_api_key ON applications(api_key);
CREATE INDEX IF NOT EXISTS idx_applications_domain ON applications(domain);
CREATE INDEX IF NOT EXISTS idx_sessions_app_id ON sessions(app_id);
CREATE INDEX IF NOT EXISTS idx_sessions_session_id ON sessions(session_id);
CREATE INDEX IF NOT EXISTS idx_custom_parameters_access_log_id ON custom_parameters(access_log_id);

-- 統計情報用のビュー
CREATE OR REPLACE VIEW access_log_stats AS
SELECT 
    app_id,
    COUNT(*) as total_requests,
    COUNT(DISTINCT session_id) as unique_sessions,
    COUNT(DISTINCT ip_address) as unique_visitors,
    COUNT(CASE WHEN user_agent ILIKE '%bot%' OR user_agent ILIKE '%crawler%' THEN 1 END) as bot_requests,
    COUNT(CASE WHEN user_agent ILIKE '%mobile%' OR user_agent ILIKE '%android%' OR user_agent ILIKE '%iphone%' THEN 1 END) as mobile_requests,
    MIN(timestamp) as first_request,
    MAX(timestamp) as last_request
FROM access_logs
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
FROM access_logs
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

-- テスト用のサンプルデータを挿入
INSERT INTO applications (app_id, name, description, domain, api_key, is_active) VALUES
('test_app_001', 'Test Application 1', 'Test application for integration testing', 'test1.example.com', 'alt_test_api_key_001', true),
('test_app_002', 'Test Application 2', 'Test application for integration testing', 'test2.example.com', 'alt_test_api_key_002', true),
('test_app_inactive', 'Inactive Test Application', 'Inactive test application', 'inactive.example.com', 'alt_test_api_key_inactive', false)
ON CONFLICT (app_id) DO NOTHING;

-- コメントの追加
COMMENT ON TABLE applications IS 'アプリケーション情報を管理するテーブル';
COMMENT ON TABLE access_logs IS 'アクセスログデータを保存するテーブル';
COMMENT ON TABLE sessions IS 'セッション情報を管理するテーブル';
COMMENT ON TABLE custom_parameters IS 'カスタムパラメータを保存するテーブル';
COMMENT ON VIEW access_log_stats IS 'アクセスログ統計情報を提供するビュー';
COMMENT ON VIEW session_stats IS 'セッション統計情報を提供するビュー';
