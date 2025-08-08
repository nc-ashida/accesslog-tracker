# データベース設計仕様書

## 1. 概要

### 1.1 データベース概要
- **DBMS**: Aurora PostgreSQL 14以上
- **文字エンコーディング**: UTF-8
- **タイムゾーン**: UTC
- **パーティショニング**: 月別パーティショニング
- **コネクションプール**: RDS Proxy対応
- **バックアップ**: 自動バックアップ（7日間保持）

### 1.2 設計方針
- 高パフォーマンスな書き込み処理
- 大量データの効率的な管理
- スケーラビリティの確保
- データ整合性の維持
- カスタムパラメータの柔軟な保存
- バッチ処理による書き込み最適化

### 1.3 Aurora PostgreSQL設定

#### 1.3.1 インスタンス設定
```yaml
# CloudFormation テンプレート
Resources:
  AuroraCluster:
    Type: AWS::RDS::DBCluster
    Properties:
      Engine: aurora-postgresql
      EngineVersion: '14.7'
      EngineMode: provisioned
      DBClusterInstanceClass: db.r6g.large
      MasterUsername: alt_admin
      MasterUserPassword: !Ref DBPassword
      BackupRetentionPeriod: 7
      PreferredBackupWindow: '03:00-04:00'
      PreferredMaintenanceWindow: 'sun:04:00-sun:05:00'
      StorageEncrypted: true
      DeletionProtection: true
      EnableCloudwatchLogsExports:
        - postgresql
      ScalingConfiguration:
        MinCapacity: 2
        MaxCapacity: 16
        AutoPause: true
        SecondsUntilAutoPause: 300

  RDSProxy:
    Type: AWS::RDS::DBProxy
    Properties:
      DBProxyName: access-log-proxy
      EngineFamily: POSTGRESQL
      RequireTLS: true
      IdleClientTimeout: 1800
      MaxConnectionsPercent: 100
      MaxIdleConnectionsPercent: 50
      Auth:
        - AuthScheme: SECRETS
          SecretArn: !Ref DBSecretArn
      RoleArn: !GetAtt RDSProxyRole.Arn
```

#### 1.3.2 パフォーマンス最適化設定
```sql
-- 書き込み性能最適化設定
ALTER SYSTEM SET max_connections = 1000;
ALTER SYSTEM SET shared_buffers = '2GB';
ALTER SYSTEM SET effective_cache_size = '6GB';
ALTER SYSTEM SET maintenance_work_mem = '256MB';
ALTER SYSTEM SET checkpoint_completion_target = 0.9;
ALTER SYSTEM SET wal_buffers = '16MB';
ALTER SYSTEM SET default_statistics_target = 100;
ALTER SYSTEM SET random_page_cost = 1.1;
ALTER SYSTEM SET effective_io_concurrency = 200;
ALTER SYSTEM SET work_mem = '4MB';
ALTER SYSTEM SET min_wal_size = '1GB';
ALTER SYSTEM SET max_wal_size = '4GB';

-- 設定の再読み込み
SELECT pg_reload_conf();
```

## 2. テーブル設計

### 2.1 アプリケーション管理テーブル

#### applications
```sql
CREATE TABLE applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    domain VARCHAR(255),
    api_key_hash VARCHAR(255) NOT NULL,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended')),
    rate_limit_per_minute INTEGER DEFAULT 1000,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- インデックス
CREATE INDEX idx_applications_app_id ON applications(app_id);
CREATE INDEX idx_applications_status ON applications(status);
CREATE INDEX idx_applications_created_at ON applications(created_at);
```

#### application_settings
```sql
CREATE TABLE application_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    setting_key VARCHAR(100) NOT NULL,
    setting_value TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(application_id, setting_key)
);

-- インデックス
CREATE INDEX idx_application_settings_app_id ON application_settings(application_id);
```

### 2.2 アクセスログテーブル（パーティショニング）

#### access_logs（親テーブル）
```sql
CREATE TABLE access_logs (
    id BIGSERIAL,
    application_id UUID NOT NULL,
    app_id VARCHAR(50) NOT NULL,
    client_sub_id VARCHAR(100),
    module_id VARCHAR(100),
    url TEXT,
    referrer TEXT,
    user_agent TEXT,
    ip_address INET,
    session_id VARCHAR(100),
    screen_resolution VARCHAR(20),
    language VARCHAR(10),
    timezone VARCHAR(50),
    event_name VARCHAR(100),
    event_data JSONB,
    -- カスタムパラメータ対応
    custom_params JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- インデックス（親テーブル）
CREATE INDEX idx_access_logs_app_id ON access_logs(app_id);
CREATE INDEX idx_access_logs_session_id ON access_logs(session_id);
CREATE INDEX idx_access_logs_created_at ON access_logs(created_at);
CREATE INDEX idx_access_logs_client_sub_id ON access_logs(client_sub_id);
CREATE INDEX idx_access_logs_module_id ON access_logs(module_id);
-- カスタムパラメータ用インデックス
CREATE INDEX idx_access_logs_custom_params ON access_logs USING GIN (custom_params);
CREATE INDEX idx_access_logs_custom_params_page_type ON access_logs USING GIN ((custom_params->>'page_type'));
```

#### パーティションテーブル例
```sql
-- 2024年1月のパーティション
CREATE TABLE access_logs_2024_01 PARTITION OF access_logs
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

-- 2024年2月のパーティション
CREATE TABLE access_logs_2024_02 PARTITION OF access_logs
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');

-- パーティション固有のインデックス
CREATE INDEX idx_access_logs_2024_01_app_id ON access_logs_2024_01(app_id);
CREATE INDEX idx_access_logs_2024_01_created_at ON access_logs_2024_01(created_at);
```

### 2.3 バッチ処理テーブル

#### batch_jobs
```sql
CREATE TABLE batch_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    batch_size INTEGER DEFAULT 100,
    processed_count INTEGER DEFAULT 0,
    error_count INTEGER DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- インデックス
CREATE INDEX idx_batch_jobs_status ON batch_jobs(status);
CREATE INDEX idx_batch_jobs_created_at ON batch_jobs(created_at);
```

#### batch_queue
```sql
CREATE TABLE batch_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tracking_data JSONB NOT NULL,
    priority INTEGER DEFAULT 0,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    status VARCHAR(20) DEFAULT 'queued' CHECK (status IN ('queued', 'processing', 'completed', 'failed')),
    processed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- インデックス
CREATE INDEX idx_batch_queue_status ON batch_queue(status);
CREATE INDEX idx_batch_queue_priority ON batch_queue(priority);
CREATE INDEX idx_batch_queue_created_at ON batch_queue(created_at);
```

### 2.3 カスタムパラメータ管理テーブル

#### custom_parameter_definitions
```sql
CREATE TABLE custom_parameter_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    parameter_key VARCHAR(100) NOT NULL,
    parameter_name VARCHAR(255) NOT NULL,
    parameter_type VARCHAR(50) NOT NULL CHECK (parameter_type IN ('string', 'number', 'boolean', 'array', 'object')),
    description TEXT,
    is_required BOOLEAN DEFAULT false,
    default_value TEXT,
    validation_rules JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(application_id, parameter_key)
);

-- インデックス
CREATE INDEX idx_custom_param_defs_app_id ON custom_parameter_definitions(application_id);
CREATE INDEX idx_custom_param_defs_key ON custom_parameter_definitions(parameter_key);
```

#### custom_parameter_values
```sql
CREATE TABLE custom_parameter_values (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    access_log_id BIGINT NOT NULL,
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    parameter_key VARCHAR(100) NOT NULL,
    parameter_value TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (access_log_id, created_at) REFERENCES access_logs(id, created_at) ON DELETE CASCADE
);

-- インデックス
CREATE INDEX idx_custom_param_values_log_id ON custom_parameter_values(access_log_id);
CREATE INDEX idx_custom_param_values_app_id ON custom_parameter_values(application_id);
CREATE INDEX idx_custom_param_values_key ON custom_parameter_values(parameter_key);
CREATE INDEX idx_custom_param_values_key_value ON custom_parameter_values(parameter_key, parameter_value);
```

### 2.4 セッション管理テーブル

#### sessions
```sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id VARCHAR(100) UNIQUE NOT NULL,
    application_id UUID NOT NULL REFERENCES applications(id),
    app_id VARCHAR(50) NOT NULL,
    client_sub_id VARCHAR(100),
    module_id VARCHAR(100),
    user_agent TEXT,
    ip_address INET,
    first_accessed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_accessed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    page_views INTEGER DEFAULT 1,
    is_active BOOLEAN DEFAULT true,
    -- セッション全体のカスタムパラメータ
    session_custom_params JSONB
);

-- インデックス
CREATE INDEX idx_sessions_session_id ON sessions(session_id);
CREATE INDEX idx_sessions_app_id ON sessions(app_id);
CREATE INDEX idx_sessions_last_accessed_at ON sessions(last_accessed_at);
CREATE INDEX idx_sessions_is_active ON sessions(is_active);
CREATE INDEX idx_sessions_custom_params ON sessions USING GIN (session_custom_params);
```

### 2.5 統計情報テーブル

#### daily_statistics
```sql
CREATE TABLE daily_statistics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id),
    app_id VARCHAR(50) NOT NULL,
    date DATE NOT NULL,
    total_requests BIGINT DEFAULT 0,
    unique_visitors BIGINT DEFAULT 0,
    unique_sessions BIGINT DEFAULT 0,
    -- カスタムパラメータ別統計
    custom_param_stats JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(application_id, date)
);

-- インデックス
CREATE INDEX idx_daily_statistics_app_id ON daily_statistics(app_id);
CREATE INDEX idx_daily_statistics_date ON daily_statistics(date);
CREATE INDEX idx_daily_statistics_custom_params ON daily_statistics USING GIN (custom_param_stats);
```

#### hourly_statistics
```sql
CREATE TABLE hourly_statistics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id),
    app_id VARCHAR(50) NOT NULL,
    date DATE NOT NULL,
    hour INTEGER NOT NULL CHECK (hour >= 0 AND hour <= 23),
    total_requests BIGINT DEFAULT 0,
    unique_visitors BIGINT DEFAULT 0,
    -- カスタムパラメータ別統計
    custom_param_stats JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(application_id, date, hour)
);

-- インデックス
CREATE INDEX idx_hourly_statistics_app_id ON hourly_statistics(app_id);
CREATE INDEX idx_hourly_statistics_date_hour ON hourly_statistics(date, hour);
CREATE INDEX idx_hourly_statistics_custom_params ON hourly_statistics USING GIN (custom_param_stats);
```

### 2.6 システム管理テーブル

#### api_keys
```sql
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    key_hash VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    permissions JSONB DEFAULT '{}',
    last_used_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- インデックス
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_application_id ON api_keys(application_id);
```

#### webhooks
```sql
CREATE TABLE webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    events JSONB NOT NULL,
    secret VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    last_triggered_at TIMESTAMP WITH TIME ZONE,
    failure_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- インデックス
CREATE INDEX idx_webhooks_application_id ON webhooks(application_id);
CREATE INDEX idx_webhooks_is_active ON webhooks(is_active);
```

## 3. カスタムパラメータの活用例

### 3.1 Eコマースサイトのカスタムパラメータ例
```sql
-- 商品詳細ページのカスタムパラメータ
INSERT INTO custom_parameter_definitions (
    application_id, parameter_key, parameter_name, parameter_type, description
) VALUES 
    ('app-uuid-1', 'page_type', 'ページタイプ', 'string', 'ページの種類（product_detail, cart, checkout等）'),
    ('app-uuid-1', 'product_id', '商品ID', 'string', '商品の一意識別子'),
    ('app-uuid-1', 'product_name', '商品名', 'string', '商品の名称'),
    ('app-uuid-1', 'product_category', '商品カテゴリ', 'string', '商品のカテゴリ'),
    ('app-uuid-1', 'product_price', '商品価格', 'number', '商品の価格'),
    ('app-uuid-1', 'product_brand', 'ブランド', 'string', '商品のブランド名'),
    ('app-uuid-1', 'product_availability', '在庫状況', 'string', '商品の在庫状況'),
    ('app-uuid-1', 'product_rating', '評価', 'number', '商品の評価（1-5）'),
    ('app-uuid-1', 'product_review_count', 'レビュー数', 'number', 'レビューの件数'),
    ('app-uuid-1', 'cart_total', 'カート合計', 'number', 'カート内の商品合計金額'),
    ('app-uuid-1', 'cart_item_count', 'カート内商品数', 'number', 'カート内の商品数'),
    ('app-uuid-1', 'user_segment', 'ユーザーセグメント', 'string', 'ユーザーのセグメント（premium, regular等）');
```

### 3.2 ニュースサイトのカスタムパラメータ例
```sql
-- ニュース記事ページのカスタムパラメータ
INSERT INTO custom_parameter_definitions (
    application_id, parameter_key, parameter_name, parameter_type, description
) VALUES 
    ('app-uuid-2', 'page_type', 'ページタイプ', 'string', 'ページの種類（article, category, search等）'),
    ('app-uuid-2', 'article_id', '記事ID', 'string', '記事の一意識別子'),
    ('app-uuid-2', 'article_title', '記事タイトル', 'string', '記事のタイトル'),
    ('app-uuid-2', 'article_category', '記事カテゴリ', 'string', '記事のカテゴリ'),
    ('app-uuid-2', 'article_author', '著者', 'string', '記事の著者名'),
    ('app-uuid-2', 'article_publish_date', '公開日', 'string', '記事の公開日'),
    ('app-uuid-2', 'article_read_time', '読了時間', 'number', '推定読了時間（分）'),
    ('app-uuid-2', 'article_tags', 'タグ', 'array', '記事のタグ一覧'),
    ('app-uuid-2', 'article_word_count', '文字数', 'number', '記事の文字数'),
    ('app-uuid-2', 'article_comment_count', 'コメント数', 'number', 'コメントの件数');
```

## 4. パーティショニング戦略

### 4.1 月別パーティショニング
```sql
-- パーティション作成関数
CREATE OR REPLACE FUNCTION create_monthly_partition(year_month TEXT)
RETURNS VOID AS $$
BEGIN
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS access_logs_%s PARTITION OF access_logs
        FOR VALUES FROM (%L) TO (%L)
    ', 
    year_month,
    year_month || '-01',
    (year_month || '-01')::date + INTERVAL '1 month'
    );
END;
$$ LANGUAGE plpgsql;

-- 自動パーティション作成（cron等で実行）
SELECT create_monthly_partition('2024-03');
```

## 5. テーブル定義詳細

### 5.1 アクセスログテーブル（access_logs）

| カラム名            | データ型                 | NULL     | デフォルト値 | 説明                           |
| ------------------- | ------------------------ | -------- | ------------ | ------------------------------ |
| `id`                | BIGSERIAL                | NOT NULL | -            | 主キー（自動採番）             |
| `application_id`    | UUID                     | NOT NULL | -            | アプリケーションID（外部キー） |
| `app_id`            | VARCHAR(50)              | NOT NULL | -            | アプリケーション識別子         |
| `client_sub_id`     | VARCHAR(100)             | NULL     | NULL         | クライアントサブID             |
| `module_id`         | VARCHAR(100)             | NULL     | NULL         | モジュールID                   |
| `url`               | TEXT                     | NULL     | NULL         | アクセスURL                    |
| `referrer`          | TEXT                     | NULL     | NULL         | リファラーページURL            |
| `user_agent`        | TEXT                     | NULL     | NULL         | ブラウザUser-Agent             |
| `ip_address`        | INET                     | NULL     | NULL         | クライアントIPアドレス         |
| `session_id`        | VARCHAR(100)             | NULL     | NULL         | セッションID                   |
| `screen_resolution` | VARCHAR(20)              | NULL     | NULL         | 画面解像度                     |
| `language`          | VARCHAR(10)              | NULL     | NULL         | ブラウザ言語設定               |
| `timezone`          | VARCHAR(50)              | NULL     | NULL         | タイムゾーン                   |
| `event_name`        | VARCHAR(100)             | NULL     | NULL         | イベント名                     |
| `event_data`        | JSONB                    | NULL     | NULL         | イベント詳細データ             |
| `custom_params`     | JSONB                    | NULL     | NULL         | カスタムパラメータ             |
| `created_at`        | TIMESTAMP WITH TIME ZONE | NOT NULL | NOW()        | 作成日時                       |

**インデックス:**
- `idx_access_logs_app_id` (app_id)
- `idx_access_logs_session_id` (session_id)
- `idx_access_logs_created_at` (created_at)
- `idx_access_logs_client_sub_id` (client_sub_id)
- `idx_access_logs_module_id` (module_id)
- `idx_access_logs_custom_params` (custom_params) - GIN
- `idx_access_logs_custom_params_page_type` ((custom_params->>'page_type')) - GIN

### 5.2 アプリケーション管理テーブル（applications）

| カラム名                | データ型                 | NULL     | デフォルト値      | 説明                                    |
| ----------------------- | ------------------------ | -------- | ----------------- | --------------------------------------- |
| `id`                    | UUID                     | NOT NULL | gen_random_uuid() | 主キー                                  |
| `app_id`                | VARCHAR(50)              | NOT NULL | -                 | アプリケーション識別子（ユニーク）      |
| `name`                  | VARCHAR(255)             | NOT NULL | -                 | アプリケーション名                      |
| `description`           | TEXT                     | NULL     | NULL              | 説明                                    |
| `domain`                | VARCHAR(255)             | NULL     | NULL              | ドメイン                                |
| `api_key_hash`          | VARCHAR(255)             | NOT NULL | -                 | APIキーハッシュ                         |
| `status`                | VARCHAR(20)              | NOT NULL | 'active'          | ステータス（active/inactive/suspended） |
| `rate_limit_per_minute` | INTEGER                  | NOT NULL | 1000              | 分間レート制限                          |
| `created_at`            | TIMESTAMP WITH TIME ZONE | NOT NULL | NOW()             | 作成日時                                |
| `updated_at`            | TIMESTAMP WITH TIME ZONE | NOT NULL | NOW()             | 更新日時                                |
| `deleted_at`            | TIMESTAMP WITH TIME ZONE | NULL     | NULL              | 削除日時                                |

**インデックス:**
- `idx_applications_app_id` (app_id)
- `idx_applications_status` (status)
- `idx_applications_created_at` (created_at)

### 5.3 セッション管理テーブル（sessions）

| カラム名                | データ型                 | NULL     | デフォルト値      | 説明                               |
| ----------------------- | ------------------------ | -------- | ----------------- | ---------------------------------- |
| `id`                    | UUID                     | NOT NULL | gen_random_uuid() | 主キー                             |
| `session_id`            | VARCHAR(100)             | NOT NULL | -                 | セッションID（ユニーク）           |
| `application_id`        | UUID                     | NOT NULL | -                 | アプリケーションID（外部キー）     |
| `app_id`                | VARCHAR(50)              | NOT NULL | -                 | アプリケーション識別子             |
| `client_sub_id`         | VARCHAR(100)             | NULL     | NULL              | クライアントサブID                 |
| `module_id`             | VARCHAR(100)             | NULL     | NULL              | モジュールID                       |
| `user_agent`            | TEXT                     | NULL     | NULL              | ブラウザUser-Agent                 |
| `ip_address`            | INET                     | NULL     | NULL              | クライアントIPアドレス             |
| `first_accessed_at`     | TIMESTAMP WITH TIME ZONE | NOT NULL | NOW()             | 初回アクセス日時                   |
| `last_accessed_at`      | TIMESTAMP WITH TIME ZONE | NOT NULL | NOW()             | 最終アクセス日時                   |
| `page_views`            | INTEGER                  | NOT NULL | 1                 | ページビュー数                     |
| `is_active`             | BOOLEAN                  | NOT NULL | true              | アクティブフラグ                   |
| `session_custom_params` | JSONB                    | NULL     | NULL              | セッション全体のカスタムパラメータ |

**インデックス:**
- `idx_sessions_session_id` (session_id)
- `idx_sessions_app_id` (app_id)
- `idx_sessions_last_accessed_at` (last_accessed_at)
- `idx_sessions_is_active` (is_active)
- `idx_sessions_custom_params` (session_custom_params) - GIN

### 5.4 カスタムパラメータ定義テーブル（custom_parameter_definitions）

| カラム名           | データ型                 | NULL     | デフォルト値      | 説明                                               |
| ------------------ | ------------------------ | -------- | ----------------- | -------------------------------------------------- |
| `id`               | UUID                     | NOT NULL | gen_random_uuid() | 主キー                                             |
| `application_id`   | UUID                     | NOT NULL | -                 | アプリケーションID（外部キー）                     |
| `parameter_key`    | VARCHAR(100)             | NOT NULL | -                 | パラメータキー                                     |
| `parameter_name`   | VARCHAR(255)             | NOT NULL | -                 | パラメータ名                                       |
| `parameter_type`   | VARCHAR(50)              | NOT NULL | -                 | パラメータ型（string/number/boolean/array/object） |
| `description`      | TEXT                     | NULL     | NULL              | 説明                                               |
| `is_required`      | BOOLEAN                  | NOT NULL | false             | 必須フラグ                                         |
| `default_value`    | TEXT                     | NULL     | NULL              | デフォルト値                                       |
| `validation_rules` | JSONB                    | NULL     | NULL              | バリデーションルール                               |
| `created_at`       | TIMESTAMP WITH TIME ZONE | NOT NULL | NOW()             | 作成日時                                           |
| `updated_at`       | TIMESTAMP WITH TIME ZONE | NOT NULL | NOW()             | 更新日時                                           |

**インデックス:**
- `idx_custom_param_defs_app_id` (application_id)
- `idx_custom_param_defs_key` (parameter_key)

### 5.5 カスタムパラメータ値テーブル（custom_parameter_values）

| カラム名          | データ型                 | NULL     | デフォルト値      | 説明                           |
| ----------------- | ------------------------ | -------- | ----------------- | ------------------------------ |
| `id`              | UUID                     | NOT NULL | gen_random_uuid() | 主キー                         |
| `access_log_id`   | BIGINT                   | NOT NULL | -                 | アクセスログID（外部キー）     |
| `application_id`  | UUID                     | NOT NULL | -                 | アプリケーションID（外部キー） |
| `parameter_key`   | VARCHAR(100)             | NOT NULL | -                 | パラメータキー                 |
| `parameter_value` | TEXT                     | NULL     | NULL              | パラメータ値                   |
| `created_at`      | TIMESTAMP WITH TIME ZONE | NOT NULL | NOW()             | 作成日時                       |

**インデックス:**
- `idx_custom_param_values_log_id` (access_log_id)
- `idx_custom_param_values_app_id` (application_id)
- `idx_custom_param_values_key` (parameter_key)
- `idx_custom_param_values_key_value` (parameter_key, parameter_value)

### 5.6 日次統計テーブル（daily_statistics）

| カラム名             | データ型                 | NULL     | デフォルト値      | 説明                           |
| -------------------- | ------------------------ | -------- | ----------------- | ------------------------------ |
| `id`                 | UUID                     | NOT NULL | gen_random_uuid() | 主キー                         |
| `application_id`     | UUID                     | NOT NULL | -                 | アプリケーションID（外部キー） |
| `app_id`             | VARCHAR(50)              | NOT NULL | -                 | アプリケーション識別子         |
| `date`               | DATE                     | NOT NULL | -                 | 統計日                         |
| `total_requests`     | BIGINT                   | NOT NULL | 0                 | 総リクエスト数                 |
| `unique_visitors`    | BIGINT                   | NOT NULL | 0                 | ユニークビジター数             |
| `unique_sessions`    | BIGINT                   | NOT NULL | 0                 | ユニークセッション数           |
| `custom_param_stats` | JSONB                    | NULL     | NULL              | カスタムパラメータ別統計       |
| `created_at`         | TIMESTAMP WITH TIME ZONE | NOT NULL | NOW()             | 作成日時                       |
| `updated_at`         | TIMESTAMP WITH TIME ZONE | NOT NULL | NOW()             | 更新日時                       |

**インデックス:**
- `idx_daily_statistics_app_id` (app_id)
- `idx_daily_statistics_date` (date)
- `idx_daily_statistics_custom_params` (custom_param_stats) - GIN

### 5.7 時間別統計テーブル（hourly_statistics）

| カラム名             | データ型                 | NULL     | デフォルト値      | 説明                           |
| -------------------- | ------------------------ | -------- | ----------------- | ------------------------------ |
| `id`                 | UUID                     | NOT NULL | gen_random_uuid() | 主キー                         |
| `application_id`     | UUID                     | NOT NULL | -                 | アプリケーションID（外部キー） |
| `app_id`             | VARCHAR(50)              | NOT NULL | -                 | アプリケーション識別子         |
| `date`               | DATE                     | NOT NULL | -                 | 統計日                         |
| `hour`               | INTEGER                  | NOT NULL | -                 | 時間（0-23）                   |
| `total_requests`     | BIGINT                   | NOT NULL | 0                 | 総リクエスト数                 |
| `unique_visitors`    | BIGINT                   | NOT NULL | 0                 | ユニークビジター数             |
| `custom_param_stats` | JSONB                    | NULL     | NULL              | カスタムパラメータ別統計       |
| `created_at`         | TIMESTAMP WITH TIME ZONE | NOT NULL | NOW()             | 作成日時                       |
| `updated_at`         | TIMESTAMP WITH TIME ZONE | NOT NULL | NOW()             | 更新日時                       |

**インデックス:**
- `idx_hourly_statistics_app_id` (app_id)
- `idx_hourly_statistics_date_hour` (date, hour)
- `idx_hourly_statistics_custom_params` (custom_param_stats) - GIN

### 5.8 バッチ処理テーブル（batch_jobs）

| カラム名          | データ型                 | NULL     | デフォルト値      | 説明                                              |
| ----------------- | ------------------------ | -------- | ----------------- | ------------------------------------------------- |
| `id`              | UUID                     | NOT NULL | gen_random_uuid() | 主キー                                            |
| `job_type`        | VARCHAR(50)              | NOT NULL | -                 | ジョブタイプ                                      |
| `status`          | VARCHAR(20)              | NOT NULL | 'pending'         | ステータス（pending/processing/completed/failed） |
| `batch_size`      | INTEGER                  | NOT NULL | 100               | バッチサイズ                                      |
| `processed_count` | INTEGER                  | NOT NULL | 0                 | 処理済み件数                                      |
| `error_count`     | INTEGER                  | NOT NULL | 0                 | エラー件数                                        |
| `started_at`      | TIMESTAMP WITH TIME ZONE | NULL     | NULL              | 開始日時                                          |
| `completed_at`    | TIMESTAMP WITH TIME ZONE | NULL     | NULL              | 完了日時                                          |
| `error_message`   | TEXT                     | NULL     | NULL              | エラーメッセージ                                  |
| `created_at`      | TIMESTAMP WITH TIME ZONE | NOT NULL | NOW()             | 作成日時                                          |
| `updated_at`      | TIMESTAMP WITH TIME ZONE | NOT NULL | NOW()             | 更新日時                                          |

**インデックス:**
- `idx_batch_jobs_status` (status)
- `idx_batch_jobs_created_at` (created_at)

### 5.9 バッチキュー（batch_queue）

| カラム名        | データ型                 | NULL     | デフォルト値      | 説明                                             |
| --------------- | ------------------------ | -------- | ----------------- | ------------------------------------------------ |
| `id`            | UUID                     | NOT NULL | gen_random_uuid() | 主キー                                           |
| `tracking_data` | JSONB                    | NOT NULL | -                 | トラッキングデータ                               |
| `priority`      | INTEGER                  | NOT NULL | 0                 | 優先度                                           |
| `retry_count`   | INTEGER                  | NOT NULL | 0                 | リトライ回数                                     |
| `max_retries`   | INTEGER                  | NOT NULL | 3                 | 最大リトライ回数                                 |
| `status`        | VARCHAR(20)              | NOT NULL | 'queued'          | ステータス（queued/processing/completed/failed） |
| `processed_at`  | TIMESTAMP WITH TIME ZONE | NULL     | NULL              | 処理日時                                         |
| `error_message` | TEXT                     | NULL     | NULL              | エラーメッセージ                                 |
| `created_at`    | TIMESTAMP WITH TIME ZONE | NOT NULL | NOW()             | 作成日時                                         |
| `updated_at`    | TIMESTAMP WITH TIME ZONE | NOT NULL | NOW()             | 更新日時                                         |

**インデックス:**
- `idx_batch_queue_status` (status)
- `idx_batch_queue_priority` (priority)
- `idx_batch_queue_created_at` (created_at)

### 5.10 APIキー管理テーブル（api_keys）

| カラム名         | データ型                 | NULL     | デフォルト値      | 説明                           |
| ---------------- | ------------------------ | -------- | ----------------- | ------------------------------ |
| `id`             | UUID                     | NOT NULL | gen_random_uuid() | 主キー                         |
| `application_id` | UUID                     | NOT NULL | -                 | アプリケーションID（外部キー） |
| `key_hash`       | VARCHAR(255)             | NOT NULL | -                 | キーハッシュ（ユニーク）       |
| `name`           | VARCHAR(255)             | NULL     | NULL              | キー名                         |
| `permissions`    | JSONB                    | NOT NULL | '{}'              | 権限設定                       |
| `last_used_at`   | TIMESTAMP WITH TIME ZONE | NULL     | NULL              | 最終使用日時                   |
| `expires_at`     | TIMESTAMP WITH TIME ZONE | NULL     | NULL              | 有効期限                       |
| `is_active`      | BOOLEAN                  | NOT NULL | true              | アクティブフラグ               |
| `created_at`     | TIMESTAMP WITH TIME ZONE | NOT NULL | NOW()             | 作成日時                       |

**インデックス:**
- `idx_api_keys_key_hash` (key_hash)
- `idx_api_keys_application_id` (application_id)

### 5.11 Webhook管理テーブル（webhooks）

| カラム名            | データ型                 | NULL     | デフォルト値      | 説明                           |
| ------------------- | ------------------------ | -------- | ----------------- | ------------------------------ |
| `id`                | UUID                     | NOT NULL | gen_random_uuid() | 主キー                         |
| `application_id`    | UUID                     | NOT NULL | -                 | アプリケーションID（外部キー） |
| `url`               | TEXT                     | NOT NULL | -                 | Webhook URL                    |
| `events`            | JSONB                    | NOT NULL | -                 | 対象イベント                   |
| `secret`            | VARCHAR(255)             | NULL     | NULL              | Webhook シークレット           |
| `is_active`         | BOOLEAN                  | NOT NULL | true              | アクティブフラグ               |
| `last_triggered_at` | TIMESTAMP WITH TIME ZONE | NULL     | NULL              | 最終実行日時                   |
| `failure_count`     | INTEGER                  | NOT NULL | 0                 | 失敗回数                       |
| `created_at`        | TIMESTAMP WITH TIME ZONE | NOT NULL | NOW()             | 作成日時                       |
| `updated_at`        | TIMESTAMP WITH TIME ZONE | NOT NULL | NOW()             | 更新日時                       |

**インデックス:**
- `idx_webhooks_application_id` (application_id)
- `idx_webhooks_is_active` (is_active)