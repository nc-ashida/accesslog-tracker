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
