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
