-- 初期スキーママイグレーション
-- 作成日: 2024-01-01
-- 説明: Access Log Trackerの初期データベーススキーマ

-- 拡張機能の有効化
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- アプリケーションテーブル
CREATE TABLE applications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    api_key VARCHAR(64) UNIQUE NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    settings JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- アプリケーション名のインデックス
CREATE INDEX idx_applications_name ON applications(name);
CREATE INDEX idx_applications_user_id ON applications(user_id);
CREATE INDEX idx_applications_api_key ON applications(api_key);
CREATE INDEX idx_applications_created_at ON applications(created_at);

-- セッションテーブル
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    visitor_id VARCHAR(255) NOT NULL,
    session_id VARCHAR(255) NOT NULL,
    ip_address INET,
    user_agent TEXT,
    referrer TEXT,
    country VARCHAR(2),
    region VARCHAR(255),
    city VARCHAR(255),
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    device_type VARCHAR(50),
    browser VARCHAR(100),
    os VARCHAR(100),
    screen_width INTEGER,
    screen_height INTEGER,
    language VARCHAR(10),
    timezone VARCHAR(50),
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_activity TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP WITH TIME ZONE,
    duration_seconds INTEGER DEFAULT 0,
    page_views INTEGER DEFAULT 1,
    is_bounce BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- セッションのインデックス
CREATE INDEX idx_sessions_application_id ON sessions(application_id);
CREATE INDEX idx_sessions_visitor_id ON sessions(visitor_id);
CREATE INDEX idx_sessions_session_id ON sessions(session_id);
CREATE INDEX idx_sessions_started_at ON sessions(started_at);
CREATE INDEX idx_sessions_last_activity ON sessions(last_activity);
CREATE INDEX idx_sessions_ip_address ON sessions(ip_address);
CREATE INDEX idx_sessions_country ON sessions(country);
CREATE INDEX idx_sessions_device_type ON sessions(device_type);

-- トラッキングテーブル
CREATE TABLE tracking (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    session_id UUID REFERENCES sessions(id) ON DELETE CASCADE,
    visitor_id VARCHAR(255) NOT NULL,
    page_url TEXT NOT NULL,
    page_title VARCHAR(500),
    referrer TEXT,
    ip_address INET,
    user_agent TEXT,
    country VARCHAR(2),
    region VARCHAR(255),
    city VARCHAR(255),
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    device_type VARCHAR(50),
    browser VARCHAR(100),
    os VARCHAR(100),
    screen_width INTEGER,
    screen_height INTEGER,
    language VARCHAR(10),
    timezone VARCHAR(50),
    load_time_ms INTEGER,
    dom_content_loaded_ms INTEGER,
    first_contentful_paint_ms INTEGER,
    largest_contentful_paint_ms INTEGER,
    first_input_delay_ms INTEGER,
    cumulative_layout_shift DECIMAL(5, 4),
    custom_parameters JSONB DEFAULT '{}',
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- トラッキングのインデックス
CREATE INDEX idx_tracking_application_id ON tracking(application_id);
CREATE INDEX idx_tracking_session_id ON tracking(session_id);
CREATE INDEX idx_tracking_visitor_id ON tracking(visitor_id);
CREATE INDEX idx_tracking_timestamp ON tracking(timestamp);
CREATE INDEX idx_tracking_page_url ON tracking(page_url);
CREATE INDEX idx_tracking_ip_address ON tracking(ip_address);
CREATE INDEX idx_tracking_country ON tracking(country);
CREATE INDEX idx_tracking_device_type ON tracking(device_type);
CREATE INDEX idx_tracking_browser ON tracking(browser);
CREATE INDEX idx_tracking_os ON tracking(os);

-- パーティショニング用のインデックス（日付範囲クエリ用）
CREATE INDEX idx_tracking_application_timestamp ON tracking(application_id, timestamp);

-- 更新日時の自動更新トリガー関数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- アプリケーションテーブルの更新日時トリガー
CREATE TRIGGER update_applications_updated_at 
    BEFORE UPDATE ON applications 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- セッションテーブルの更新日時トリガー
CREATE TRIGGER update_sessions_updated_at 
    BEFORE UPDATE ON sessions 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- 統計ビュー（リアルタイム統計用）
CREATE VIEW session_statistics AS
SELECT 
    application_id,
    COUNT(*) as total_sessions,
    COUNT(CASE WHEN ended_at IS NULL THEN 1 END) as active_sessions,
    AVG(duration_seconds) as avg_duration_seconds,
    AVG(page_views) as avg_page_views,
    COUNT(CASE WHEN is_bounce = true THEN 1 END) * 100.0 / COUNT(*) as bounce_rate
FROM sessions
GROUP BY application_id;

-- ページビュー統計ビュー
CREATE VIEW page_view_statistics AS
SELECT 
    application_id,
    page_url,
    COUNT(*) as view_count,
    COUNT(DISTINCT visitor_id) as unique_visitors,
    COUNT(DISTINCT session_id) as unique_sessions
FROM tracking
GROUP BY application_id, page_url;

-- デバイス統計ビュー
CREATE VIEW device_statistics AS
SELECT 
    application_id,
    device_type,
    COUNT(*) as count
FROM tracking
GROUP BY application_id, device_type;

-- ブラウザ統計ビュー
CREATE VIEW browser_statistics AS
SELECT 
    application_id,
    browser,
    COUNT(*) as count
FROM tracking
GROUP BY application_id, browser;

-- 国別統計ビュー
CREATE VIEW country_statistics AS
SELECT 
    application_id,
    country,
    COUNT(*) as count
FROM tracking
WHERE country IS NOT NULL
GROUP BY application_id, country;
