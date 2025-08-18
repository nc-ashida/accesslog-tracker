#!/bin/bash
# deployments/scripts/common/utils.sh

# 共通設定
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$(dirname "$SCRIPT_DIR")")")"

# ログ関数
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# エラーログ関数
error_log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] ERROR: $1" >&2
}

# 成功ログ関数
success_log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] SUCCESS: $1"
}

# 警告ログ関数
warn_log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] WARNING: $1"
}

# エラーハンドリング関数
error_exit() {
    error_log "$1"
    exit 1
}

# 確認関数
confirm() {
    local message="$1"
    echo -n "$message (y/N): "
    read -r response
    case "$response" in
        [yY][eE][sS]|[yY])
            return 0
            ;;
        *)
            return 1
            ;;
    esac
}

# バックアップ関数
backup_file() {
    local file="$1"
    local backup_dir="$2"
    
    if [ -f "$file" ]; then
        local backup_file="$backup_dir/$(basename "$file").backup.$(date +%Y%m%d_%H%M%S)"
        log "バックアップを作成中: $file -> $backup_file"
        cp "$file" "$backup_file"
        echo "$backup_file"
    fi
}

# 復元関数
restore_file() {
    local backup_file="$1"
    local target_file="$2"
    
    if [ -f "$backup_file" ]; then
        log "ファイルを復元中: $backup_file -> $target_file"
        cp "$backup_file" "$target_file"
        return 0
    else
        error_log "バックアップファイルが見つかりません: $backup_file"
        return 1
    fi
}

# サービス状態チェック関数
check_service_status() {
    local service_name="$1"
    
    if systemctl is-active --quiet "$service_name"; then
        return 0
    else
        return 1
    fi
}

# ポート使用状況チェック関数
check_port() {
    local port="$1"
    
    if netstat -tuln | grep -q ":$port "; then
        return 0
    else
        return 1
    fi
}

# ディスク使用量チェック関数
check_disk_usage() {
    local threshold="${1:-80}"
    local usage=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
    
    if [ "$usage" -gt "$threshold" ]; then
        warn_log "ディスク使用量が高いです: ${usage}%"
        return 1
    else
        return 0
    fi
}

# メモリ使用量チェック関数
check_memory_usage() {
    local threshold="${1:-85}"
    local usage=$(free | awk 'NR==2{printf "%.0f", $3*100/$2}')
    
    if [ "$usage" -gt "$threshold" ]; then
        warn_log "メモリ使用量が高いです: ${usage}%"
        return 1
    else
        return 0
    fi
}

# バージョン比較関数
compare_versions() {
    local version1="$1"
    local version2="$2"
    
    if [ "$version1" = "$version2" ]; then
        return 0
    fi
    
    local IFS=.
    local i ver1=($version1) ver2=($version2)
    
    for ((i=${#ver1[@]}; i<${#ver2[@]}; i++)); do
        ver1[i]=0
    done
    
    for ((i=0; i<${#ver1[@]}; i++)); do
        if [[ -z ${ver2[i]} ]]; then
            ver2[i]=0
        fi
        
        if ((10#${ver1[i]} > 10#${ver2[i]})); then
            return 1
        fi
        
        if ((10#${ver1[i]} < 10#${ver2[i]})); then
            return 2
        fi
    done
    
    return 0
}

# 環境変数チェック関数
check_required_env() {
    local missing_vars=()
    
    for var in "$@"; do
        if [ -z "${!var}" ]; then
            missing_vars+=("$var")
        fi
    done
    
    if [ ${#missing_vars[@]} -gt 0 ]; then
        error_log "必要な環境変数が設定されていません: ${missing_vars[*]}"
        return 1
    fi
    
    return 0
}

# ファイル存在チェック関数
check_required_files() {
    local missing_files=()
    
    for file in "$@"; do
        if [ ! -f "$file" ]; then
            missing_files+=("$file")
        fi
    done
    
    if [ ${#missing_files[@]} -gt 0 ]; then
        error_log "必要なファイルが見つかりません: ${missing_files[*]}"
        return 1
    fi
    
    return 0
}

# ディレクトリ存在チェック関数
check_required_dirs() {
    local missing_dirs=()
    
    for dir in "$@"; do
        if [ ! -d "$dir" ]; then
            missing_dirs+=("$dir")
        fi
    done
    
    if [ ${#missing_dirs[@]} -gt 0 ]; then
        error_log "必要なディレクトリが見つかりません: ${missing_dirs[*]}"
        return 1
    fi
    
    return 0
}
