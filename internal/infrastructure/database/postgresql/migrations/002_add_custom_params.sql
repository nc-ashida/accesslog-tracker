-- カスタムパラメータ追加マイグレーション
-- 作成日: 2024-01-01
-- 説明: カスタムパラメータ機能の追加

-- カスタムパラメータテーブル（アプリケーション固有のパラメータ定義）
CREATE TABLE custom_parameters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    parameter_name VARCHAR(100) NOT NULL,
    parameter_type VARCHAR(50) NOT NULL DEFAULT 'string', -- string, number, boolean, array
    description TEXT,
    is_required BOOLEAN DEFAULT false,
    default_value TEXT,
    validation_rules JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(application_id, parameter_name)
);

-- カスタムパラメータのインデックス
CREATE INDEX idx_custom_parameters_application_id ON custom_parameters(application_id);
CREATE INDEX idx_custom_parameters_name ON custom_parameters(parameter_name);

-- カスタムパラメータテーブルの更新日時トリガー
CREATE TRIGGER update_custom_parameters_updated_at 
    BEFORE UPDATE ON custom_parameters 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- トラッキングテーブルにカスタムパラメータのインデックスを追加
CREATE INDEX idx_tracking_custom_parameters ON tracking USING GIN (custom_parameters);

-- カスタムパラメータ統計ビュー
CREATE VIEW custom_parameter_statistics AS
SELECT 
    t.application_id,
    cp.parameter_name,
    cp.parameter_type,
    COUNT(*) as usage_count,
    COUNT(DISTINCT t.visitor_id) as unique_visitors,
    COUNT(DISTINCT t.session_id) as unique_sessions
FROM tracking t
JOIN custom_parameters cp ON t.application_id = cp.application_id
WHERE t.custom_parameters ? cp.parameter_name
GROUP BY t.application_id, cp.parameter_name, cp.parameter_type;

-- パフォーマンス最適化のためのパーティショニング準備
-- トラッキングテーブルを日付でパーティショニングする準備
-- 注: 実際のパーティショニングは運用時に実装

-- アーカイブテーブル（古いデータ用）
CREATE TABLE tracking_archive (
    LIKE tracking INCLUDING ALL
) PARTITION BY RANGE (timestamp);

-- アーカイブテーブルのインデックス
CREATE INDEX idx_tracking_archive_application_id ON tracking_archive(application_id);
CREATE INDEX idx_tracking_archive_timestamp ON tracking_archive(timestamp);

-- データ保持ポリシー用のビュー
CREATE VIEW data_retention_stats AS
SELECT 
    application_id,
    COUNT(*) as total_records,
    MIN(timestamp) as oldest_record,
    MAX(timestamp) as newest_record,
    COUNT(*) FILTER (WHERE timestamp < CURRENT_TIMESTAMP - INTERVAL '90 days') as records_older_than_90_days,
    COUNT(*) FILTER (WHERE timestamp < CURRENT_TIMESTAMP - INTERVAL '365 days') as records_older_than_1_year
FROM tracking
GROUP BY application_id;

-- データクリーンアップ用の関数
CREATE OR REPLACE FUNCTION cleanup_old_tracking_data(
    p_application_id UUID,
    p_days_to_keep INTEGER DEFAULT 90
)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM tracking 
    WHERE application_id = p_application_id 
    AND timestamp < CURRENT_TIMESTAMP - (p_days_to_keep || ' days')::INTERVAL;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- セッションクリーンアップ用の関数
CREATE OR REPLACE FUNCTION cleanup_expired_sessions(
    p_application_id UUID,
    p_hours_to_keep INTEGER DEFAULT 24
)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM sessions 
    WHERE application_id = p_application_id 
    AND last_activity < CURRENT_TIMESTAMP - (p_hours_to_keep || ' hours')::INTERVAL
    AND ended_at IS NOT NULL;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;
