-- データベーススキーマの不整合修正
-- 作成日: 2025年8月18日
-- 説明: applicationsテーブルのactiveカラムをis_activeに変更

-- activeカラムをis_activeに変更
ALTER TABLE applications RENAME COLUMN active TO is_active;

-- コメントの更新
COMMENT ON COLUMN applications.is_active IS 'アプリケーションのアクティブ状態';
