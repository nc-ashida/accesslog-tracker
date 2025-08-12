#!/bin/bash

# Access Log Tracker - ヘルスチェックスクリプト

set -e

# 色付き出力
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ログ関数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 変数定義
APP_NAME="access-log-tracker"
DEFAULT_PORT="8080"
DEFAULT_TIMEOUT="5"

# ヘルプ表示
show_help() {
    cat << EOF
Access Log Tracker - ヘルスチェックスクリプト

使用方法:
    $0 [オプション]

オプション:
    -h, --help          このヘルプを表示
    -p, --port PORT     ポート番号を指定 (デフォルト: 8080)
    -t, --timeout SEC   タイムアウト秒数を指定 (デフォルト: 5)
    -a, --all           すべてのサービスをチェック
    -d, --detailed      詳細な情報を表示
    -c, --continuous    継続的にチェック
    -i, --interval SEC  継続チェックの間隔 (デフォルト: 30)
    -f, --format FORMAT 出力フォーマット (text, json, csv)
    -o, --output FILE   出力ファイルを指定

チェック項目:
    app                 アプリケーション
    db                  データベース (PostgreSQL)
    cache               キャッシュ (Redis)
    disk                ディスク使用量
    memory              メモリ使用量
    network             ネットワーク接続

例:
    $0                    # 基本的なヘルスチェック
    $0 -a                 # すべてのサービスをチェック
    $0 -c -i 60          # 1分間隔で継続チェック
    $0 -f json -o health.json  # JSON形式でファイル出力

EOF
}

# 初期化
init() {
    log_info "ヘルスチェックを開始します"
    log_info "アプリケーション: $APP_NAME"
    log_info "ポート: $HEALTH_PORT"
    log_info "タイムアウト: $HEALTH_TIMEOUT秒"
}

# アプリケーションヘルスチェック
check_app() {
    log_info "アプリケーションのヘルスチェック中..."
    
    local url="http://localhost:$HEALTH_PORT/health"
    local response=$(curl -s -w "%{http_code}" -o /tmp/health_response --max-time "$HEALTH_TIMEOUT" "$url" || echo "000")
    
    if [ "$response" = "200" ]; then
        log_success "アプリケーションは正常です"
        if [ "$DETAILED" = true ]; then
            cat /tmp/health_response
        fi
        return 0
    else
        log_error "アプリケーションが異常です (HTTP: $response)"
        return 1
    fi
}

# データベースヘルスチェック
check_database() {
    log_info "データベースのヘルスチェック中..."
    
    if command -v docker &> /dev/null; then
        if docker exec access-log-tracker-postgres pg_isready -U postgres -d access_log_tracker > /dev/null 2>&1; then
            log_success "データベースは正常です"
            return 0
        else
            log_error "データベースが異常です"
            return 1
        fi
    else
        # Dockerがない場合は直接接続を試行
        if command -v psql &> /dev/null; then
            if PGPASSWORD=password psql -h localhost -U postgres -d access_log_tracker -c "SELECT 1;" > /dev/null 2>&1; then
                log_success "データベースは正常です"
                return 0
            else
                log_error "データベースが異常です"
                return 1
            fi
        else
            log_warning "PostgreSQLクライアントがインストールされていません"
            return 2
        fi
    fi
}

# キャッシュヘルスチェック
check_cache() {
    log_info "キャッシュのヘルスチェック中..."
    
    if command -v docker &> /dev/null; then
        if docker exec access-log-tracker-redis redis-cli ping > /dev/null 2>&1; then
            log_success "キャッシュは正常です"
            return 0
        else
            log_error "キャッシュが異常です"
            return 1
        fi
    else
        # Dockerがない場合は直接接続を試行
        if command -v redis-cli &> /dev/null; then
            if redis-cli -h localhost ping > /dev/null 2>&1; then
                log_success "キャッシュは正常です"
                return 0
            else
                log_error "キャッシュが異常です"
                return 1
            fi
        else
            log_warning "Redisクライアントがインストールされていません"
            return 2
        fi
    fi
}

# ディスク使用量チェック
check_disk() {
    log_info "ディスク使用量をチェック中..."
    
    local usage=$(df / | tail -1 | awk '{print $5}' | sed 's/%//')
    local threshold=90
    
    if [ "$usage" -lt "$threshold" ]; then
        log_success "ディスク使用量は正常です ($usage%)"
        return 0
    else
        log_error "ディスク使用量が高いです ($usage%)"
        return 1
    fi
}

# メモリ使用量チェック
check_memory() {
    log_info "メモリ使用量をチェック中..."
    
    local total=$(free | grep Mem | awk '{print $2}')
    local used=$(free | grep Mem | awk '{print $3}')
    local usage=$((used * 100 / total))
    local threshold=90
    
    if [ "$usage" -lt "$threshold" ]; then
        log_success "メモリ使用量は正常です ($usage%)"
        return 0
    else
        log_error "メモリ使用量が高いです ($usage%)"
        return 1
    fi
}

# ネットワーク接続チェック
check_network() {
    log_info "ネットワーク接続をチェック中..."
    
    local endpoints=("google.com:80" "github.com:443" "cloudflare.com:443")
    local failed=0
    
    for endpoint in "${endpoints[@]}"; do
        local host=$(echo "$endpoint" | cut -d: -f1)
        local port=$(echo "$endpoint" | cut -d: -f2)
        
        if timeout "$HEALTH_TIMEOUT" bash -c "</dev/tcp/$host/$port" 2>/dev/null; then
            log_success "$host:$port に接続可能です"
        else
            log_error "$host:$port に接続できません"
            ((failed++))
        fi
    done
    
    if [ "$failed" -eq 0 ]; then
        return 0
    else
        return 1
    fi
}

# プロセスチェック
check_process() {
    log_info "プロセスをチェック中..."
    
    if pgrep -f "$APP_NAME" > /dev/null; then
        log_success "アプリケーションプロセスは実行中です"
        return 0
    else
        log_error "アプリケーションプロセスが停止しています"
        return 1
    fi
}

# Dockerコンテナチェック
check_docker() {
    log_info "Dockerコンテナをチェック中..."
    
    if ! command -v docker &> /dev/null; then
        log_warning "Dockerがインストールされていません"
        return 2
    fi
    
    local containers=("access-log-tracker-postgres" "access-log-tracker-redis")
    local failed=0
    
    for container in "${containers[@]}"; do
        if docker ps --format "table {{.Names}}" | grep -q "$container"; then
            log_success "コンテナ $container は実行中です"
        else
            log_error "コンテナ $container が停止しています"
            ((failed++))
        fi
    done
    
    if [ "$failed" -eq 0 ]; then
        return 0
    else
        return 1
    fi
}

# すべてのチェック実行
check_all() {
    local results=()
    local total=0
    local passed=0
    
    # アプリケーション
    if check_app; then
        results+=("app:OK")
        ((passed++))
    else
        results+=("app:FAIL")
    fi
    ((total++))
    
    # プロセス
    if check_process; then
        results+=("process:OK")
        ((passed++))
    else
        results+=("process:FAIL")
    fi
    ((total++))
    
    # データベース
    if check_database; then
        results+=("database:OK")
        ((passed++))
    else
        results+=("database:FAIL")
    fi
    ((total++))
    
    # キャッシュ
    if check_cache; then
        results+=("cache:OK")
        ((passed++))
    else
        results+=("cache:FAIL")
    fi
    ((total++))
    
    # Docker
    if check_docker; then
        results+=("docker:OK")
        ((passed++))
    else
        results+=("docker:FAIL")
    fi
    ((total++))
    
    # ディスク
    if check_disk; then
        results+=("disk:OK")
        ((passed++))
    else
        results+=("disk:FAIL")
    fi
    ((total++))
    
    # メモリ
    if check_memory; then
        results+=("memory:OK")
        ((passed++))
    else
        results+=("memory:FAIL")
    fi
    ((total++))
    
    # ネットワーク
    if check_network; then
        results+=("network:OK")
        ((passed++))
    else
        results+=("network:FAIL")
    fi
    ((total++))
    
    # 結果表示
    echo ""
    log_info "ヘルスチェック結果:"
    for result in "${results[@]}"; do
        local service=$(echo "$result" | cut -d: -f1)
        local status=$(echo "$result" | cut -d: -f2)
        
        if [ "$status" = "OK" ]; then
            echo -e "  ${GREEN}✓${NC} $service"
        else
            echo -e "  ${RED}✗${NC} $service"
        fi
    done
    
    echo ""
    log_info "総合結果: $passed/$total チェックが成功"
    
    if [ "$passed" -eq "$total" ]; then
        log_success "すべてのヘルスチェックが成功しました"
        return 0
    else
        log_error "一部のヘルスチェックが失敗しました"
        return 1
    fi
}

# JSON形式で出力
output_json() {
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    local results=()
    
    # 各チェックを実行して結果を収集
    results+=("{\"service\":\"app\",\"status\":\"$(check_app && echo 'OK' || echo 'FAIL')\"}")
    results+=("{\"service\":\"process\",\"status\":\"$(check_process && echo 'OK' || echo 'FAIL')\"}")
    results+=("{\"service\":\"database\",\"status\":\"$(check_database && echo 'OK' || echo 'FAIL')\"}")
    results+=("{\"service\":\"cache\",\"status\":\"$(check_cache && echo 'OK' || echo 'FAIL')\"}")
    results+=("{\"service\":\"docker\",\"status\":\"$(check_docker && echo 'OK' || echo 'FAIL')\"}")
    results+=("{\"service\":\"disk\",\"status\":\"$(check_disk && echo 'OK' || echo 'FAIL')\"}")
    results+=("{\"service\":\"memory\",\"status\":\"$(check_memory && echo 'OK' || echo 'FAIL')\"}")
    results+=("{\"service\":\"network\",\"status\":\"$(check_network && echo 'OK' || echo 'FAIL')\"}")
    
    # JSON形式で出力
    echo "{"
    echo "  \"timestamp\": \"$timestamp\","
    echo "  \"application\": \"$APP_NAME\","
    echo "  \"checks\": ["
    printf "    %s" "${results[0]}"
    for ((i=1; i<${#results[@]}; i++)); do
        printf ",\n    %s" "${results[i]}"
    done
    echo ""
    echo "  ]"
    echo "}"
}

# CSV形式で出力
output_csv() {
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    echo "timestamp,service,status"
    echo "$timestamp,app,$(check_app && echo 'OK' || echo 'FAIL')"
    echo "$timestamp,process,$(check_process && echo 'OK' || echo 'FAIL')"
    echo "$timestamp,database,$(check_database && echo 'OK' || echo 'FAIL')"
    echo "$timestamp,cache,$(check_cache && echo 'OK' || echo 'FAIL')"
    echo "$timestamp,docker,$(check_docker && echo 'OK' || echo 'FAIL')"
    echo "$timestamp,disk,$(check_disk && echo 'OK' || echo 'FAIL')"
    echo "$timestamp,memory,$(check_memory && echo 'OK' || echo 'FAIL')"
    echo "$timestamp,network,$(check_network && echo 'OK' || echo 'FAIL')"
}

# 継続チェック
continuous_check() {
    log_info "継続チェックを開始します (間隔: ${HEALTH_INTERVAL}秒)"
    
    while true; do
        echo ""
        echo "=== $(date) ==="
        
        if [ "$OUTPUT_FORMAT" = "json" ]; then
            output_json
        elif [ "$OUTPUT_FORMAT" = "csv" ]; then
            output_csv
        else
            check_all
        fi
        
        sleep "$HEALTH_INTERVAL"
    done
}

# メイン処理
main() {
    local show_help_flag=false
    local check_all_flag=false
    local continuous_flag=false
    local detailed_flag=false
    local health_port="$DEFAULT_PORT"
    local health_timeout="$DEFAULT_TIMEOUT"
    local health_interval="30"
    local output_format="text"
    local output_file=""
    
    # オプション解析
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help_flag=true
                shift
                ;;
            -p|--port)
                health_port="$2"
                shift 2
                ;;
            -t|--timeout)
                health_timeout="$2"
                shift 2
                ;;
            -a|--all)
                check_all_flag=true
                shift
                ;;
            -c|--continuous)
                continuous_flag=true
                shift
                ;;
            -i|--interval)
                health_interval="$2"
                shift 2
                ;;
            -d|--detailed)
                detailed_flag=true
                shift
                ;;
            -f|--format)
                output_format="$2"
                shift 2
                ;;
            -o|--output)
                output_file="$2"
                shift 2
                ;;
            *)
                log_error "不明なオプション: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # ヘルプ表示
    if [ "$show_help_flag" = true ]; then
        show_help
        exit 0
    fi
    
    # グローバル変数設定
    HEALTH_PORT="$health_port"
    HEALTH_TIMEOUT="$health_timeout"
    HEALTH_INTERVAL="$health_interval"
    DETAILED="$detailed_flag"
    OUTPUT_FORMAT="$output_format"
    
    # 初期化
    init
    
    # 出力先設定
    if [ -n "$output_file" ]; then
        exec > "$output_file"
    fi
    
    # 継続チェック
    if [ "$continuous_flag" = true ]; then
        continuous_check
        exit 0
    fi
    
    # 全チェック実行
    if [ "$check_all_flag" = true ]; then
        if [ "$output_format" = "json" ]; then
            output_json
        elif [ "$output_format" = "csv" ]; then
            output_csv
        else
            check_all
        fi
    else
        # 基本的なアプリケーションチェックのみ
        check_app
    fi
}

# スクリプト実行
main "$@"
