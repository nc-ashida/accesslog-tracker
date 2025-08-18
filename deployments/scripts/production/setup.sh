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
