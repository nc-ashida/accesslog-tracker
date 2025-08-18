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
