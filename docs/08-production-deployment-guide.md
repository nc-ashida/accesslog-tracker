# 本番環境構築手順書

## 1. 概要

### 1.1 本番環境構成
- **ロードバランサー**: AWS Application Load Balancer (ALB)
- **アプリケーションサーバー**: EC2インスタンス 3台（他のアプリケーションと共存）
- **データベース**: Aurora for PostgreSQL
- **キャッシュ**: ElastiCache for Redis
- **監視**: CloudWatch
- **ログ**: CloudWatch Logs

### 1.2 アーキテクチャ図
```
┌─────────────────────────────────────────────────────────────┐
│                    Internet                                 │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│              Application Load Balancer                      │
│                    (ALB)                                    │
└─────────────┬───────────────┬───────────────┬───────────────┘
              │               │               │
┌─────────────▼─────┐ ┌───────▼──────┐ ┌─────▼─────────────┐
│   EC2 Instance 1  │ │ EC2 Instance 2│ │  EC2 Instance 3  │
│  (Port 8080)      │ │ (Port 8080)  │ │  (Port 8080)     │
│  [ALT + Others]   │ │[ALT + Others]│ │ [ALT + Others]   │
└─────────┬─────────┘ └──────┬───────┘ └─────┬─────────────┘
          │                  │               │
          └──────────────────┼───────────────┘
                             │
┌────────────────────────────▼────────────────────────────────┐
│                    VPC                                      │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐ │
│  │ Aurora Cluster  │  │ ElastiCache     │  │ Security    │ │
│  │ PostgreSQL      │  │ Redis           │  │ Groups      │ │
│  └─────────────────┘  └─────────────────┘  └─────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### 1.3 技術スタック
- **アプリケーション**: Go + Gin Framework
- **Webサーバー**: Nginx（リバースプロキシ）
- **プロセス管理**: systemd
- **データベース**: Aurora for PostgreSQL
- **キャッシュ**: ElastiCache for Redis
- **ロードバランサー**: AWS ALB
- **監視**: CloudWatch + CloudWatch Logs

## 2. 前提条件

### 2.1 AWSアカウント設定
- AWS CLIの設定完了
- 適切なIAM権限の設定
- VPC、サブネット、セキュリティグループの準備

### 2.2 必要なAWSサービス
- EC2（t3.medium以上推奨）
- Application Load Balancer
- Aurora for PostgreSQL
- ElastiCache for Redis
- CloudWatch
- CloudWatch Logs
- IAM

## 3. インフラ構築手順

### 3.1 VPC・ネットワーク設定

#### 3.1.1 VPC作成
```bash
# VPC作成
aws ec2 create-vpc \
  --cidr-block 10.0.0.0/16 \
  --tag-specifications 'ResourceType=vpc,Tags=[{Key=Name,Value=alt-production-vpc}]'

# サブネット作成（パブリック）
aws ec2 create-subnet \
  --vpc-id vpc-xxxxxxxxx \
  --cidr-block 10.0.1.0/24 \
  --availability-zone ap-northeast-1a \
  --tag-specifications 'ResourceType=subnet,Tags=[{Key=Name,Value=alt-public-subnet-1a}]'

# サブネット作成（プライベート）
aws ec2 create-subnet \
  --vpc-id vpc-xxxxxxxxx \
  --cidr-block 10.0.2.0/24 \
  --availability-zone ap-northeast-1a \
  --tag-specifications 'ResourceType=subnet,Tags=[{Key=Name,Value=alt-private-subnet-1a}]'
```

#### 3.1.2 セキュリティグループ作成
```bash
# ALB用セキュリティグループ
aws ec2 create-security-group \
  --group-name alt-alb-sg \
  --description "Security group for ALB" \
  --vpc-id vpc-xxxxxxxxx

# EC2用セキュリティグループ
aws ec2 create-security-group \
  --group-name alt-ec2-sg \
  --description "Security group for EC2 instances" \
  --vpc-id vpc-xxxxxxxxx

# Aurora用セキュリティグループ
aws ec2 create-security-group \
  --group-name alt-aurora-sg \
  --description "Security group for Aurora PostgreSQL" \
  --vpc-id vpc-xxxxxxxxx

# ElastiCache用セキュリティグループ
aws ec2 create-security-group \
  --group-name alt-elasticache-sg \
  --description "Security group for ElastiCache Redis" \
  --vpc-id vpc-xxxxxxxxx
```

#### 3.1.3 セキュリティグループルール設定
```bash
# ALBセキュリティグループルール
aws ec2 authorize-security-group-ingress \
  --group-id sg-xxxxxxxxx \
  --protocol tcp \
  --port 80 \
  --cidr 0.0.0.0/0

aws ec2 authorize-security-group-ingress \
  --group-id sg-xxxxxxxxx \
  --protocol tcp \
  --port 443 \
  --cidr 0.0.0.0/0

# EC2セキュリティグループルール
aws ec2 authorize-security-group-ingress \
  --group-id sg-xxxxxxxxx \
  --protocol tcp \
  --port 8080 \
  --source-group sg-xxxxxxxxx  # ALBのセキュリティグループ

# Auroraセキュリティグループルール
aws ec2 authorize-security-group-ingress \
  --group-id sg-xxxxxxxxx \
  --protocol tcp \
  --port 5432 \
  --source-group sg-xxxxxxxxx  # EC2のセキュリティグループ

# ElastiCacheセキュリティグループルール
aws ec2 authorize-security-group-ingress \
  --group-id sg-xxxxxxxxx \
  --protocol tcp \
  --port 6379 \
  --source-group sg-xxxxxxxxx  # EC2のセキュリティグループ
```

### 3.2 Aurora for PostgreSQL設定

#### 3.2.1 サブネットグループ作成
```bash
aws rds create-db-subnet-group \
  --db-subnet-group-name alt-aurora-subnet-group \
  --db-subnet-group-description "Subnet group for Aurora PostgreSQL" \
  --subnet-ids subnet-xxxxxxxxx subnet-yyyyyyyyy
```

#### 3.2.2 Auroraクラスター作成
```bash
aws rds create-db-cluster \
  --db-cluster-identifier alt-aurora-cluster \
  --engine aurora-postgresql \
  --engine-version 15.4 \
  --master-username alt_admin \
  --master-user-password "secure_password_here" \
  --db-subnet-group-name alt-aurora-subnet-group \
  --vpc-security-group-ids sg-xxxxxxxxx \
  --backup-retention-period 7 \
  --preferred-backup-window "03:00-04:00" \
  --preferred-maintenance-window "sun:04:00-sun:05:00" \
  --storage-encrypted \
  --deletion-protection
```

#### 3.2.3 Auroraインスタンス作成
```bash
aws rds create-db-instance \
  --db-instance-identifier alt-aurora-writer \
  --db-cluster-identifier alt-aurora-cluster \
  --engine aurora-postgresql \
  --db-instance-class db.r6g.large \
  --allocated-storage 100 \
  --storage-type gp3 \
  --storage-encrypted \
  --monitoring-interval 60 \
  --monitoring-role-arn arn:aws:iam::xxxxxxxxx:role/rds-monitoring-role

# リーダーインスタンス（オプション）
aws rds create-db-instance \
  --db-instance-identifier alt-aurora-reader \
  --db-cluster-identifier alt-aurora-cluster \
  --engine aurora-postgresql \
  --db-instance-class db.r6g.large \
  --allocated-storage 100 \
  --storage-type gp3 \
  --storage-encrypted \
  --monitoring-interval 60 \
  --monitoring-role-arn arn:aws:iam::xxxxxxxxx:role/rds-monitoring-role
```

### 3.3 ElastiCache for Redis設定

#### 3.3.1 サブネットグループ作成
```bash
aws elasticache create-cache-subnet-group \
  --cache-subnet-group-name alt-redis-subnet-group \
  --cache-subnet-group-description "Subnet group for ElastiCache Redis" \
  --subnet-ids subnet-xxxxxxxxx subnet-yyyyyyyyy
```

#### 3.3.2 Redisクラスター作成
```bash
aws elasticache create-cache-cluster \
  --cache-cluster-id alt-redis-cluster \
  --engine redis \
  --cache-node-type cache.t3.micro \
  --num-cache-nodes 1 \
  --cache-subnet-group-name alt-redis-subnet-group \
  --security-group-ids sg-xxxxxxxxx \
  --port 6379 \
  --preferred-availability-zone ap-northeast-1a
```

### 3.4 Application Load Balancer設定

#### 3.4.1 ALB作成
```bash
aws elbv2 create-load-balancer \
  --name alt-production-alb \
  --subnets subnet-xxxxxxxxx subnet-yyyyyyyyy \
  --security-groups sg-xxxxxxxxx \
  --scheme internet-facing \
  --type application
```

#### 3.4.2 ターゲットグループ作成
```bash
aws elbv2 create-target-group \
  --name alt-production-tg \
  --protocol HTTP \
  --port 8080 \
  --vpc-id vpc-xxxxxxxxx \
  --target-type instance \
  --health-check-protocol HTTP \
  --health-check-path /health \
  --health-check-interval-seconds 30 \
  --health-check-timeout-seconds 5 \
  --healthy-threshold-count 2 \
  --unhealthy-threshold-count 2
```

#### 3.4.3 リスナー作成
```bash
# HTTPリスナー
aws elbv2 create-listener \
  --load-balancer-arn arn:aws:elasticloadbalancing:ap-northeast-1:xxxxxxxxx:loadbalancer/app/alt-production-alb/xxxxxxxxx \
  --protocol HTTP \
  --port 80 \
  --default-actions Type=forward,TargetGroupArn=arn:aws:elasticloadbalancing:ap-northeast-1:xxxxxxxxx:targetgroup/alt-production-tg/xxxxxxxxx

# HTTPSリスナー（SSL証明書がある場合）
aws elbv2 create-listener \
  --load-balancer-arn arn:aws:elasticloadbalancing:ap-northeast-1:xxxxxxxxx:loadbalancer/app/alt-production-alb/xxxxxxxxx \
  --protocol HTTPS \
  --port 443 \
  --certificates CertificateArn=arn:aws:acm:ap-northeast-1:xxxxxxxxx:certificate/xxxxxxxxx \
  --default-actions Type=forward,TargetGroupArn=arn:aws:elasticloadbalancing:ap-northeast-1:xxxxxxxxx:targetgroup/alt-production-tg/xxxxxxxxx
```

## 4. EC2インスタンス設定

### 4.1 AMI・インスタンスタイプ選択
```bash
# Amazon Linux 2023 AMIを使用
AMI_ID="ami-xxxxxxxxx"  # ap-northeast-1のAmazon Linux 2023 AMI
INSTANCE_TYPE="t3.medium"
```

### 4.2 EC2インスタンス作成
```bash
# 3台のEC2インスタンスを作成
for i in {1..3}; do
  aws ec2 run-instances \
    --image-id $AMI_ID \
    --count 1 \
    --instance-type $INSTANCE_TYPE \
    --key-name your-key-pair \
    --security-group-ids sg-xxxxxxxxx \
    --subnet-id subnet-xxxxxxxxx \
    --iam-instance-profile Name=alt-ec2-role \
    --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=alt-production-ec2-'$i'},{Key=Environment,Value=production},{Key=Application,Value=access-log-tracker}]' \
    --user-data file://user-data.sh
done
```

### 4.3 本番用サービス登録スクリプト

#### 4.3.1 スクリプト配置場所
```bash
# プロジェクト構造
deployments/
├── scripts/
│   ├── production/
│   │   ├── register-service.sh    # サービス登録スクリプト
│   │   ├── deploy.sh             # デプロイスクリプト
│   │   ├── setup.sh              # 初期セットアップスクリプト
│   │   └── health-check.sh       # ヘルスチェックスクリプト
│   └── common/
│       ├── utils.sh              # 共通ユーティリティ
│       └── logging.sh            # ログ機能
```

#### 4.3.2 サービス登録スクリプト（register-service.sh）
```bash
#!/bin/bash
# deployments/scripts/production/register-service.sh

set -e

# 設定
SERVICE_NAME="access-log-tracker"
SERVICE_USER="ec2-user"
APP_DIR="/opt/access-log-tracker"
LOG_DIR="/var/log/access-log-tracker"
CONFIG_DIR="/etc/access-log-tracker"

# ログ関数
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# エラーハンドリング
error_exit() {
    log "ERROR: $1"
    exit 1
}

# ディレクトリ作成
log "ディレクトリを作成中..."
sudo mkdir -p $APP_DIR $LOG_DIR $CONFIG_DIR
sudo chown $SERVICE_USER:$SERVICE_USER $APP_DIR $LOG_DIR $CONFIG_DIR

# systemdサービスファイル作成
log "systemdサービスファイルを作成中..."
sudo tee /etc/systemd/system/$SERVICE_NAME.service > /dev/null << EOF
[Unit]
Description=Access Log Tracker
After=network.target postgresql.service redis.service
Wants=postgresql.service redis.service

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
WorkingDirectory=$APP_DIR
ExecStart=$APP_DIR/access-log-tracker
ExecReload=/bin/kill -HUP \$MAINPID
Restart=always
RestartSec=5
StartLimitInterval=0

# 環境変数
Environment=ENVIRONMENT=production
Environment=DB_HOST=alt-aurora-cluster.cluster-xxxxxxxxx.ap-northeast-1.rds.amazonaws.com
Environment=DB_PORT=5432
Environment=DB_NAME=access_log_tracker_prod
Environment=DB_USER=alt_admin
Environment=DB_PASSWORD=secure_password_here
Environment=DB_SSL_MODE=require
Environment=REDIS_HOST=alt-redis-cluster.xxxxxxxxx.cache.amazonaws.com
Environment=REDIS_PORT=6379
Environment=REDIS_PASSWORD=
Environment=REDIS_DB=0
Environment=API_PORT=8080
Environment=API_HOST=0.0.0.0
Environment=LOG_LEVEL=info

# セキュリティ設定
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$APP_DIR $LOG_DIR $CONFIG_DIR

# ログ設定
StandardOutput=journal
StandardError=journal
SyslogIdentifier=$SERVICE_NAME

[Install]
WantedBy=multi-user.target
EOF

# サービス有効化
log "サービスを有効化中..."
sudo systemctl daemon-reload
sudo systemctl enable $SERVICE_NAME

# 権限設定
log "権限を設定中..."
sudo chmod 755 $APP_DIR
sudo chmod 644 /etc/systemd/system/$SERVICE_NAME.service

log "サービス登録が完了しました"
log "サービスを起動するには: sudo systemctl start $SERVICE_NAME"
log "サービス状態を確認するには: sudo systemctl status $SERVICE_NAME"
```

#### 4.3.3 デプロイスクリプト（deploy.sh）
```bash
#!/bin/bash
# deployments/scripts/production/deploy.sh

set -e

# 設定
SERVICE_NAME="access-log-tracker"
APP_DIR="/opt/access-log-tracker"
BACKUP_DIR="/opt/backups/access-log-tracker"
S3_BUCKET="your-deployment-bucket"
APP_VERSION="$1"

# ログ関数
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# エラーハンドリング
error_exit() {
    log "ERROR: $1"
    # ロールバック
    if [ -f "$BACKUP_DIR/access-log-tracker.backup" ]; then
        log "ロールバックを実行中..."
        sudo systemctl stop $SERVICE_NAME || true
        sudo cp $BACKUP_DIR/access-log-tracker.backup $APP_DIR/access-log-tracker
        sudo chmod +x $APP_DIR/access-log-tracker
        sudo systemctl start $SERVICE_NAME
        log "ロールバックが完了しました"
    fi
    exit 1
}

# 引数チェック
if [ -z "$APP_VERSION" ]; then
    error_exit "アプリケーションバージョンを指定してください"
fi

log "デプロイを開始します: バージョン $APP_VERSION"

# バックアップディレクトリ作成
sudo mkdir -p $BACKUP_DIR

# 現在のアプリケーションをバックアップ
if [ -f "$APP_DIR/access-log-tracker" ]; then
    log "現在のアプリケーションをバックアップ中..."
    sudo cp $APP_DIR/access-log-tracker $BACKUP_DIR/access-log-tracker.backup
fi

# アプリケーション停止
log "アプリケーションを停止中..."
sudo systemctl stop $SERVICE_NAME || true

# 新しいバージョンをダウンロード
log "新しいバージョンをダウンロード中..."
aws s3 cp s3://$S3_BUCKET/access-log-tracker-$APP_VERSION $APP_DIR/access-log-tracker
sudo chmod +x $APP_DIR/access-log-tracker

# アプリケーション起動
log "アプリケーションを起動中..."
sudo systemctl start $SERVICE_NAME

# ヘルスチェック
log "ヘルスチェックを実行中..."
sleep 10

for i in {1..30}; do
    if curl -f http://localhost:8080/health > /dev/null 2>&1; then
        log "ヘルスチェックが成功しました"
        break
    fi
    
    if [ $i -eq 30 ]; then
        error_exit "ヘルスチェックが失敗しました"
    fi
    
    log "ヘルスチェックを再試行中... ($i/30)"
    sleep 2
done

# 古いバックアップを削除（7日以上）
log "古いバックアップを削除中..."
find $BACKUP_DIR -name "*.backup" -mtime +7 -delete

log "デプロイが完了しました"
```

#### 4.3.4 初期セットアップスクリプト（setup.sh）
```bash
#!/bin/bash
# deployments/scripts/production/setup.sh

set -e

# 設定
SERVICE_NAME="access-log-tracker"
APP_DIR="/opt/access-log-tracker"
LOG_DIR="/var/log/access-log-tracker"
CONFIG_DIR="/etc/access-log-tracker"

# ログ関数
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# エラーハンドリング
error_exit() {
    log "ERROR: $1"
    exit 1
}

log "初期セットアップを開始します"

# システムアップデート
log "システムをアップデート中..."
sudo yum update -y

# 必要なパッケージをインストール
log "必要なパッケージをインストール中..."
sudo yum install -y \
    nginx \
    golang \
    amazon-cloudwatch-agent \
    jq \
    curl \
    wget \
    unzip

# ディレクトリ作成
log "ディレクトリを作成中..."
sudo mkdir -p $APP_DIR $LOG_DIR $CONFIG_DIR
sudo chown ec2-user:ec2-user $APP_DIR $LOG_DIR $CONFIG_DIR

# Nginx設定
log "Nginxを設定中..."
sudo tee /etc/nginx/conf.d/access-log-tracker.conf > /dev/null << 'EOF'
upstream access_log_tracker {
    server 127.0.0.1:8080;
    keepalive 32;
}

server {
    listen 80;
    server_name _;

    # セキュリティヘッダー
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    # ログ設定
    access_log /var/log/nginx/access-log-tracker.access.log;
    error_log /var/log/nginx/access-log-tracker.error.log;

    # メインアプリケーション
    location / {
        proxy_pass http://access_log_tracker;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Connection "";
        proxy_http_version 1.1;
        
        # タイムアウト設定
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }

    # ヘルスチェック
    location /health {
        proxy_pass http://access_log_tracker/health;
        access_log off;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # ビーコンファイル
    location ~ ^/tracker(\.min)?\.js$ {
        proxy_pass http://access_log_tracker;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        expires 1h;
        add_header Cache-Control "public, immutable";
    }

    # 静的ファイル
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
        access_log off;
    }
}
EOF

# Nginx設定テスト
sudo nginx -t || error_exit "Nginx設定が無効です"

# サービス有効化・起動
log "Nginxを有効化・起動中..."
sudo systemctl enable nginx
sudo systemctl start nginx

# logrotate設定
log "logrotateを設定中..."
sudo tee /etc/logrotate.d/access-log-tracker > /dev/null << 'EOF'
/var/log/access-log-tracker.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 644 ec2-user ec2-user
    postrotate
        systemctl reload access-log-tracker
    endscript
}
EOF

# CloudWatch設定
log "CloudWatchを設定中..."
sudo tee /opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json > /dev/null << 'EOF'
{
  "logs": {
    "logs_collected": {
      "files": {
        "collect_list": [
          {
            "file_path": "/var/log/access-log-tracker.log",
            "log_group_name": "/aws/ec2/access-log-tracker",
            "log_stream_name": "{instance_id}",
            "timezone": "UTC"
          },
          {
            "file_path": "/var/log/nginx/access-log-tracker.access.log",
            "log_group_name": "/aws/ec2/nginx-access",
            "log_stream_name": "{instance_id}",
            "timezone": "UTC"
          },
          {
            "file_path": "/var/log/nginx/access-log-tracker.error.log",
            "log_group_name": "/aws/ec2/nginx-error",
            "log_stream_name": "{instance_id}",
            "timezone": "UTC"
          }
        ]
      }
    }
  },
  "metrics": {
    "metrics_collected": {
      "disk": {
        "measurement": ["used_percent"],
        "metrics_collection_interval": 60,
        "resources": ["*"]
      },
      "mem": {
        "measurement": ["mem_used_percent"],
        "metrics_collection_interval": 60
      },
      "netstat": {
        "measurement": ["tcp_established", "tcp_time_wait"],
        "metrics_collection_interval": 60
      }
    }
  }
}
EOF

# CloudWatchエージェント起動
sudo systemctl enable amazon-cloudwatch-agent
sudo systemctl start amazon-cloudwatch-agent

log "初期セットアップが完了しました"
log "次のステップ:"
log "1. サービス登録スクリプトを実行: ./deployments/scripts/production/register-service.sh"
log "2. アプリケーションをデプロイ: ./deployments/scripts/production/deploy.sh <version>"
```

#### 4.3.5 ヘルスチェックスクリプト（health-check.sh）
```bash
#!/bin/bash
# deployments/scripts/production/health-check.sh

set -e

# 設定
SERVICE_NAME="access-log-tracker"
HEALTH_URL="http://localhost:8080/health"
LOG_FILE="/var/log/access-log-tracker/health-check.log"

# ログ関数
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a $LOG_FILE
}

# ヘルスチェック実行
log "ヘルスチェックを開始します"

# サービス状態チェック
if ! systemctl is-active --quiet $SERVICE_NAME; then
    log "ERROR: サービスが停止しています"
    systemctl status $SERVICE_NAME
    exit 1
fi

# HTTPヘルスチェック
if ! curl -f -s $HEALTH_URL > /dev/null; then
    log "ERROR: HTTPヘルスチェックが失敗しました"
    exit 1
fi

# データベース接続チェック
if ! curl -f -s "$HEALTH_URL/db" > /dev/null; then
    log "ERROR: データベース接続チェックが失敗しました"
    exit 1
fi

# Redis接続チェック
if ! curl -f -s "$HEALTH_URL/redis" > /dev/null; then
    log "ERROR: Redis接続チェックが失敗しました"
    exit 1
fi

log "ヘルスチェックが成功しました"
exit 0
```

### 4.4 ユーザーデータスクリプト（user-data.sh）
```bash
#!/bin/bash
yum update -y
yum install -y nginx

# Go 1.21のインストール
yum install -y golang

# アプリケーションディレクトリ作成
mkdir -p /opt/access-log-tracker
cd /opt/access-log-tracker

# アプリケーションのダウンロード・設定
# （実際のデプロイでは、S3からダウンロードまたはGitからクローン）

# Nginx設定
cat > /etc/nginx/conf.d/access-log-tracker.conf << 'EOF'
upstream access_log_tracker {
    server 127.0.0.1:8080;
}

server {
    listen 80;
    server_name _;

    location / {
        proxy_pass http://access_log_tracker;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /health {
        proxy_pass http://access_log_tracker/health;
        access_log off;
    }
}
EOF

# systemdサービスファイル作成
cat > /etc/systemd/system/access-log-tracker.service << 'EOF'
[Unit]
Description=Access Log Tracker
After=network.target

[Service]
Type=simple
User=ec2-user
WorkingDirectory=/opt/access-log-tracker
ExecStart=/opt/access-log-tracker/access-log-tracker
Restart=always
RestartSec=5
Environment=ENVIRONMENT=production
Environment=DB_HOST=alt-aurora-cluster.cluster-xxxxxxxxx.ap-northeast-1.rds.amazonaws.com
Environment=DB_PORT=5432
Environment=DB_NAME=access_log_tracker_prod
Environment=DB_USER=alt_admin
Environment=DB_PASSWORD=secure_password_here
Environment=DB_SSL_MODE=require
Environment=REDIS_HOST=alt-redis-cluster.xxxxxxxxx.cache.amazonaws.com
Environment=REDIS_PORT=6379
Environment=REDIS_PASSWORD=
Environment=REDIS_DB=0
Environment=API_PORT=8080
Environment=API_HOST=0.0.0.0
Environment=LOG_LEVEL=info

[Install]
WantedBy=multi-user.target
EOF

# サービス有効化・起動
systemctl daemon-reload
systemctl enable nginx
systemctl enable access-log-tracker
systemctl start nginx
systemctl start access-log-tracker
```

### 4.4 ターゲットグループにインスタンス登録
```bash
# インスタンスIDを取得してターゲットグループに登録
INSTANCE_IDS=$(aws ec2 describe-instances \
  --filters "Name=tag:Name,Values=alt-production-ec2-*" "Name=instance-state-name,Values=running" \
  --query 'Reservations[].Instances[].InstanceId' \
  --output text)

for instance_id in $INSTANCE_IDS; do
  aws elbv2 register-targets \
    --target-group-arn arn:aws:elasticloadbalancing:ap-northeast-1:xxxxxxxxx:targetgroup/alt-production-tg/xxxxxxxxx \
    --targets Id=$instance_id
done
```

## 5. アプリケーションデプロイ

### 5.1 アプリケーションビルド
```bash
# ローカルでビルド
make build-docker

# DockerイメージをECRにプッシュ
aws ecr get-login-password --region ap-northeast-1 | docker login --username AWS --password-stdin xxxxxxxxx.dkr.ecr.ap-northeast-1.amazonaws.com

docker tag access-log-tracker:latest xxxxxxxxx.dkr.ecr.ap-northeast-1.amazonaws.com/access-log-tracker:latest
docker push xxxxxxxxx.dkr.ecr.ap-northeast-1.amazonaws.com/access-log-tracker:latest
```

### 5.2 デプロイスクリプト（deploy.sh）
```bash
#!/bin/bash
# EC2インスタンス上で実行

# アプリケーション停止
systemctl stop access-log-tracker

# 新しいバージョンをダウンロード
cd /opt/access-log-tracker
aws s3 cp s3://your-deployment-bucket/access-log-tracker /opt/access-log-tracker/access-log-tracker
chmod +x /opt/access-log-tracker/access-log-tracker

# アプリケーション起動
systemctl start access-log-tracker

# ヘルスチェック
sleep 10
curl -f http://localhost:8080/health || exit 1
```

### 5.3 データベースマイグレーション
```bash
# マイグレーション実行
cd /opt/access-log-tracker
./access-log-tracker migrate
```

## 6. 監視・ログ設定

### 6.1 CloudWatch設定
```bash
# CloudWatchエージェントインストール
yum install -y amazon-cloudwatch-agent

# CloudWatchエージェント設定
cat > /opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json << 'EOF'
{
  "logs": {
    "logs_collected": {
      "files": {
        "collect_list": [
          {
            "file_path": "/var/log/access-log-tracker.log",
            "log_group_name": "/aws/ec2/access-log-tracker",
            "log_stream_name": "{instance_id}",
            "timezone": "UTC"
          },
          {
            "file_path": "/var/log/nginx/access.log",
            "log_group_name": "/aws/ec2/nginx-access",
            "log_stream_name": "{instance_id}",
            "timezone": "UTC"
          }
        ]
      }
    }
  },
  "metrics": {
    "metrics_collected": {
      "disk": {
        "measurement": ["used_percent"],
        "metrics_collection_interval": 60,
        "resources": ["*"]
      },
      "mem": {
        "measurement": ["mem_used_percent"],
        "metrics_collection_interval": 60
      }
    }
  }
}
EOF

# CloudWatchエージェント起動
systemctl enable amazon-cloudwatch-agent
systemctl start amazon-cloudwatch-agent
```

### 6.2 アラート設定
```bash
# CPU使用率アラート
aws cloudwatch put-metric-alarm \
  --alarm-name "ALT-HighCPU" \
  --alarm-description "High CPU usage for Access Log Tracker" \
  --metric-name CPUUtilization \
  --namespace AWS/EC2 \
  --statistic Average \
  --period 300 \
  --threshold 80 \
  --comparison-operator GreaterThanThreshold \
  --evaluation-periods 2 \
  --alarm-actions arn:aws:sns:ap-northeast-1:xxxxxxxxx:your-sns-topic

# メモリ使用率アラート
aws cloudwatch put-metric-alarm \
  --alarm-name "ALT-HighMemory" \
  --alarm-description "High memory usage for Access Log Tracker" \
  --metric-name MemoryUtilization \
  --namespace System/Linux \
  --statistic Average \
  --period 300 \
  --threshold 85 \
  --comparison-operator GreaterThanThreshold \
  --evaluation-periods 2 \
  --alarm-actions arn:aws:sns:ap-northeast-1:xxxxxxxxx:your-sns-topic
```

## 7. セキュリティ設定

### 7.1 IAMロール設定
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "rds-db:connect",
        "elasticache:DescribeCacheClusters"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject"
      ],
      "Resource": "arn:aws:s3:::your-deployment-bucket/*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "*"
    }
  ]
}
```

### 7.2 セキュリティグループ詳細設定
```bash
# EC2セキュリティグループ（最小権限）
aws ec2 authorize-security-group-ingress \
  --group-id sg-xxxxxxxxx \
  --protocol tcp \
  --port 22 \
  --cidr your-office-ip/32  # SSHアクセス（必要最小限）

# Auroraセキュリティグループ
aws ec2 authorize-security-group-ingress \
  --group-id sg-xxxxxxxxx \
  --protocol tcp \
  --port 5432 \
  --source-group sg-xxxxxxxxx  # EC2からのみアクセス許可
```

## 8. バックアップ・復旧

### 8.1 Auroraバックアップ設定
```bash
# 自動バックアップ設定（既にクラスター作成時に設定済み）
# 手動スナップショット作成
aws rds create-db-cluster-snapshot \
  --db-cluster-snapshot-identifier alt-aurora-snapshot-$(date +%Y%m%d-%H%M%S) \
  --db-cluster-identifier alt-aurora-cluster
```

### 8.2 アプリケーションデータバックアップ
```bash
# 設定ファイルのバックアップ
aws s3 cp /opt/access-log-tracker/config s3://your-backup-bucket/config/$(date +%Y%m%d)/ --recursive

# ログファイルのバックアップ
aws s3 cp /var/log/access-log-tracker.log s3://your-backup-bucket/logs/$(date +%Y%m%d)/
```

## 9. 運用・メンテナンス

### 9.1 ログローテーション設定
```bash
# logrotate設定
cat > /etc/logrotate.d/access-log-tracker << 'EOF'
/var/log/access-log-tracker.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 644 ec2-user ec2-user
    postrotate
        systemctl reload access-log-tracker
    endscript
}
EOF
```

### 9.2 定期メンテナンススクリプト
```bash
#!/bin/bash
# /opt/scripts/maintenance.sh

# ログファイルのクリーンアップ
find /var/log -name "*.log.*" -mtime +30 -delete

# ディスク使用量チェック
DISK_USAGE=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
if [ $DISK_USAGE -gt 80 ]; then
    echo "High disk usage: ${DISK_USAGE}%" | mail -s "Disk Usage Alert" admin@example.com
fi

# アプリケーションのヘルスチェック
if ! curl -f http://localhost:8080/health > /dev/null 2>&1; then
    echo "Application health check failed" | mail -s "Health Check Alert" admin@example.com
    systemctl restart access-log-tracker
fi
```

### 9.3 スケーリング手順
```bash
# 新しいEC2インスタンス追加
aws ec2 run-instances \
  --image-id $AMI_ID \
  --count 1 \
  --instance-type $INSTANCE_TYPE \
  --key-name your-key-pair \
  --security-group-ids sg-xxxxxxxxx \
  --subnet-id subnet-xxxxxxxxx \
  --iam-instance-profile Name=alt-ec2-role \
  --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=alt-production-ec2-4},{Key=Environment,Value=production},{Key=Application,Value=access-log-tracker}]' \
  --user-data file://user-data.sh

# ターゲットグループに追加
aws elbv2 register-targets \
  --target-group-arn arn:aws:elasticloadbalancing:ap-northeast-1:xxxxxxxxx:targetgroup/alt-production-tg/xxxxxxxxx \
  --targets Id=i-xxxxxxxxx
```

## 10. トラブルシューティング

### 10.1 よくある問題と対処法

#### アプリケーションが起動しない
```bash
# ログ確認
journalctl -u access-log-tracker -f

# 設定確認
systemctl status access-log-tracker

# 手動起動テスト
cd /opt/access-log-tracker
./access-log-tracker
```

#### データベース接続エラー
```bash
# 接続テスト
psql -h alt-aurora-cluster.cluster-xxxxxxxxx.ap-northeast-1.rds.amazonaws.com -U alt_admin -d access_log_tracker_prod

# セキュリティグループ確認
aws ec2 describe-security-groups --group-ids sg-xxxxxxxxx
```

#### ALBヘルスチェック失敗
```bash
# ターゲットグループのヘルス状態確認
aws elbv2 describe-target-health \
  --target-group-arn arn:aws:elasticloadbalancing:ap-northeast-1:xxxxxxxxx:targetgroup/alt-production-tg/xxxxxxxxx

# インスタンスレベルでのヘルスチェック
curl -f http://localhost:8080/health
```

### 10.2 ログ分析
```bash
# CloudWatchログの確認
aws logs describe-log-groups --log-group-name-prefix "/aws/ec2/access-log-tracker"

# 特定のログストリームからログ取得
aws logs get-log-events \
  --log-group-name "/aws/ec2/access-log-tracker" \
  --log-stream-name "i-xxxxxxxxx" \
  --start-time $(date -d '1 hour ago' +%s)000
```

## 11. コスト最適化

### 11.1 リソース最適化
- **EC2**: t3.medium（必要に応じてt3.smallにダウングレード）
- **Aurora**: db.r6g.large（読み取り専用インスタンスは必要時のみ起動）
- **ElastiCache**: cache.t3.micro（必要に応じてスケールアップ）

### 11.2 コスト監視
```bash
# コストアラート設定
aws ce create-cost-allocation-tag \
  --tag-key "Environment" \
  --tag-values "production"

# 月次コストレポート
aws ce get-cost-and-usage \
  --time-period Start=2024-01-01,End=2024-02-01 \
  --granularity MONTHLY \
  --metrics BlendedCost \
  --group-by Type=DIMENSION,Key=SERVICE
```

## 12. セキュリティチェックリスト

### 12.1 定期的なセキュリティ確認
- [ ] セキュリティグループの見直し
- [ ] IAMロールの権限確認
- [ ] SSL証明書の有効期限確認
- [ ] セキュリティパッチの適用状況確認
- [ ] ログの監査

### 12.2 セキュリティ強化
```bash
# 不要なポートの閉鎖
aws ec2 revoke-security-group-ingress \
  --group-id sg-xxxxxxxxx \
  --protocol tcp \
  --port 22 \
  --cidr 0.0.0.0/0

# 特定のIPからのみSSHアクセス許可
aws ec2 authorize-security-group-ingress \
  --group-id sg-xxxxxxxxx \
  --protocol tcp \
  --port 22 \
  --cidr your-office-ip/32
```

## 13. パフォーマンス監視

### 13.1 パフォーマンスメトリクス
```bash
# レスポンス時間監視
aws cloudwatch put-metric-alarm \
  --alarm-name "ALT-HighResponseTime" \
  --alarm-description "High response time for Access Log Tracker" \
  --metric-name TargetResponseTime \
  --namespace AWS/ApplicationELB \
  --statistic Average \
  --period 300 \
  --threshold 1.0 \
  --comparison-operator GreaterThanThreshold \
  --evaluation-periods 2 \
  --alarm-actions arn:aws:sns:ap-northeast-1:xxxxxxxxx:your-sns-topic

# エラー率監視
aws cloudwatch put-metric-alarm \
  --alarm-name "ALT-HighErrorRate" \
  --alarm-description "High error rate for Access Log Tracker" \
  --metric-name HTTPCode_Target_5XX_Count \
  --namespace AWS/ApplicationELB \
  --statistic Sum \
  --period 300 \
  --threshold 10 \
  --comparison-operator GreaterThanThreshold \
  --evaluation-periods 2 \
  --alarm-actions arn:aws:sns:ap-northeast-1:xxxxxxxxx:your-sns-topic
```

## 14. 更新・メンテナンス手順

### 14.1 アプリケーション更新手順
```bash
#!/bin/bash
# /opt/scripts/update.sh

# 1. 新しいバージョンをダウンロード
aws s3 cp s3://your-deployment-bucket/access-log-tracker-new /tmp/access-log-tracker-new

# 2. バックアップ作成
cp /opt/access-log-tracker/access-log-tracker /opt/access-log-tracker/access-log-tracker.backup

# 3. アプリケーション停止
systemctl stop access-log-tracker

# 4. 新しいバージョンを配置
mv /tmp/access-log-tracker-new /opt/access-log-tracker/access-log-tracker
chmod +x /opt/access-log-tracker/access-log-tracker

# 5. アプリケーション起動
systemctl start access-log-tracker

# 6. ヘルスチェック
sleep 10
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    echo "Update successful"
    rm /opt/access-log-tracker/access-log-tracker.backup
else
    echo "Update failed, rolling back"
    mv /opt/access-log-tracker/access-log-tracker.backup /opt/access-log-tracker/access-log-tracker
    systemctl restart access-log-tracker
fi
```

### 14.2 ロールアウト戦略
1. **ブルー・グリーンデプロイ**: 新しいインスタンスを追加してから古いインスタンスを削除
2. **ローリングアップデート**: インスタンスを1台ずつ更新
3. **カナリアデプロイ**: 一部のインスタンスのみ新しいバージョンをデプロイ

## 15. 緊急時対応

### 15.1 障害時の対応手順
```bash
#!/bin/bash
# /opt/scripts/emergency.sh

# 1. 状況確認
systemctl status access-log-tracker
journalctl -u access-log-tracker --since "10 minutes ago"

# 2. データベース接続確認
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT 1;"

# 3. 必要に応じてアプリケーション再起動
systemctl restart access-log-tracker

# 4. 管理者に通知
echo "Emergency: Access Log Tracker service restarted" | mail -s "Emergency Alert" admin@example.com
```

### 15.2 復旧手順
1. **インスタンス障害**: 新しいインスタンスを起動してターゲットグループに追加
2. **データベース障害**: Auroraの自動フェイルオーバーを確認
3. **ALB障害**: 新しいALBを作成してDNSを更新

## 16. ドキュメント管理

### 16.1 更新履歴
- **2024-01-15**: 初版作成
- **2024-01-20**: セキュリティ設定追加
- **2024-01-25**: 監視設定詳細化

### 16.2 関連ドキュメント
- [システム概要](./01-overview.md)
- [API仕様書](./02-api-specification.md)
- [データベース設計](./04-database-design.md)
- [テスト戦略](./06-testing-strategy.md)

---

**注意**: この手順書は本番環境向けの構築手順です。実際の運用前に、テスト環境での検証を必ず行ってください。
