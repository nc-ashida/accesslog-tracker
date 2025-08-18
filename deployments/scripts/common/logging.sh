#!/bin/bash
# deployments/scripts/common/logging.sh

# ログ設定
LOG_LEVEL="${LOG_LEVEL:-INFO}"
LOG_FILE="${LOG_FILE:-/var/log/access-log-tracker/script.log}"
LOG_FORMAT="${LOG_FORMAT:-json}"

# ログレベル定義
declare -A LOG_LEVELS=(
    ["DEBUG"]=0
    ["INFO"]=1
    ["WARN"]=2
    ["ERROR"]=3
    ["FATAL"]=4
)

# 現在のログレベルを数値で取得
get_log_level_num() {
    echo "${LOG_LEVELS[${LOG_LEVEL^^}]:-1}"
}

# ログレベルチェック
should_log() {
    local level="$1"
    local current_level_num=$(get_log_level_num)
    local level_num="${LOG_LEVELS[${level^^}]:-1}"
    
    [ "$level_num" -ge "$current_level_num" ]
}

# JSON形式ログ出力
log_json() {
    local level="$1"
    local message="$2"
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%S.%3NZ")
    
    if should_log "$level"; then
        echo "{\"timestamp\":\"$timestamp\",\"level\":\"$level\",\"message\":\"$message\"}" >> "$LOG_FILE"
    fi
}

# プレーンテキスト形式ログ出力
log_plain() {
    local level="$1"
    local message="$2"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    if should_log "$level"; then
        echo "[$timestamp] [$level] $message" >> "$LOG_FILE"
    fi
}

# ログ出力関数
log_output() {
    local level="$1"
    local message="$2"
    
    case "$LOG_FORMAT" in
        "json")
            log_json "$level" "$message"
            ;;
        *)
            log_plain "$level" "$message"
            ;;
    esac
    
    # コンソールにも出力
    case "$level" in
        "ERROR"|"FATAL")
            echo "[$level] $message" >&2
            ;;
        *)
            echo "[$level] $message"
            ;;
    esac
}

# ログ関数
log_debug() {
    log_output "DEBUG" "$1"
}

log_info() {
    log_output "INFO" "$1"
}

log_warn() {
    log_output "WARN" "$1"
}

log_error() {
    log_output "ERROR" "$1"
}

log_fatal() {
    log_output "FATAL" "$1"
    exit 1
}

# 構造化ログ出力
log_structured() {
    local level="$1"
    local message="$2"
    shift 2
    local fields=("$@")
    
    local json_fields=""
    for field in "${fields[@]}"; do
        if [ -n "$json_fields" ]; then
            json_fields="$json_fields,"
        fi
        json_fields="$json_fields\"$field\""
    done
    
    if [ "$LOG_FORMAT" = "json" ]; then
        local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%S.%3NZ")
        echo "{\"timestamp\":\"$timestamp\",\"level\":\"$level\",\"message\":\"$message\",\"fields\":[$json_fields]}" >> "$LOG_FILE"
    else
        log_output "$level" "$message"
    fi
}

# ログローテーション
rotate_logs() {
    local log_file="$1"
    local max_size="${2:-100M}"
    local max_files="${3:-5}"
    
    if [ -f "$log_file" ]; then
        local size=$(stat -c%s "$log_file" 2>/dev/null || stat -f%z "$log_file" 2>/dev/null || echo 0)
        local max_size_bytes=$(numfmt --from=iec "$max_size" 2>/dev/null || echo 104857600)
        
        if [ "$size" -gt "$max_size_bytes" ]; then
            # ログローテーション
            for i in $(seq $((max_files-1)) -1 1); do
                if [ -f "${log_file}.$i" ]; then
                    mv "${log_file}.$i" "${log_file}.$((i+1))"
                fi
            done
            
            if [ -f "$log_file" ]; then
                mv "$log_file" "${log_file}.1"
            fi
            
            touch "$log_file"
            log_info "ログファイルをローテーションしました: $log_file"
        fi
    fi
}

# ログディレクトリ作成
setup_logging() {
    local log_dir=$(dirname "$LOG_FILE")
    
    if [ ! -d "$log_dir" ]; then
        mkdir -p "$log_dir"
        log_info "ログディレクトリを作成しました: $log_dir"
    fi
    
    # ログファイルが存在しない場合は作成
    if [ ! -f "$LOG_FILE" ]; then
        touch "$LOG_FILE"
        log_info "ログファイルを作成しました: $LOG_FILE"
    fi
}

# ログクリーンアップ
cleanup_old_logs() {
    local log_dir="$1"
    local days="${2:-30}"
    
    if [ -d "$log_dir" ]; then
        find "$log_dir" -name "*.log.*" -mtime +"$days" -delete
        log_info "古いログファイルを削除しました: $log_dir (${days}日以上)"
    fi
}

# ログ統計
get_log_stats() {
    local log_file="$1"
    local period="${2:-1h}"
    
    if [ -f "$log_file" ]; then
        local start_time=$(date -d "$period ago" +"%Y-%m-%d %H:%M:%S" 2>/dev/null || date -v-"$period" +"%Y-%m-%d %H:%M:%S" 2>/dev/null || echo "")
        
        if [ -n "$start_time" ]; then
            local error_count=$(awk -v start="$start_time" '$1" "$2 >= start && $0 ~ /ERROR/ {count++} END {print count+0}' "$log_file")
            local warn_count=$(awk -v start="$start_time" '$1" "$2 >= start && $0 ~ /WARN/ {count++} END {print count+0}' "$log_file")
            local total_count=$(awk -v start="$start_time" '$1" "$2 >= start {count++} END {print count+0}' "$log_file")
            
            echo "{\"period\":\"$period\",\"total\":$total_count,\"errors\":$error_count,\"warnings\":$warn_count}"
        fi
    fi
}
