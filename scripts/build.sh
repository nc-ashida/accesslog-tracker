#!/bin/bash

# Access Log Tracker - ビルドスクリプト

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
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION=$(go version | awk '{print $3}')
LDFLAGS="-ldflags -X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GoVersion=$GO_VERSION"

# ヘルプ表示
show_help() {
    cat << EOF
Access Log Tracker - ビルドスクリプト

使用方法:
    $0 [オプション]

オプション:
    -h, --help          このヘルプを表示
    -v, --version       バージョンを表示
    -c, --clean         ビルド前にクリーンアップ
    -t, --test          ビルド後にテストを実行
    -d, --docker        Dockerイメージもビルド
    -a, --all           すべてのバイナリをビルド
    -o, --output DIR    出力ディレクトリを指定 (デフォルト: bin)
    -e, --env ENV       環境を指定 (dev, staging, prod)

例:
    $0                    # 基本的なビルド
    $0 -c -t             # クリーンアップしてテスト付きビルド
    $0 -d -e prod        # 本番用Dockerイメージをビルド
    $0 -a -o dist        # すべてのバイナリをdistディレクトリにビルド

EOF
}

# 初期化
init() {
    log_info "Access Log Tracker ビルドを開始します"
    log_info "バージョン: $VERSION"
    log_info "ビルド時刻: $BUILD_TIME"
    log_info "Go バージョン: $GO_VERSION"
}

# 依存関係チェック
check_dependencies() {
    log_info "依存関係をチェック中..."
    
    if ! command -v go &> /dev/null; then
        log_error "Go がインストールされていません"
        exit 1
    fi
    
    if ! command -v git &> /dev/null; then
        log_warning "Git がインストールされていません"
    fi
    
    log_success "依存関係チェック完了"
}

# クリーンアップ
clean() {
    log_info "ビルドディレクトリをクリーンアップ中..."
    rm -rf bin/
    rm -f coverage.out coverage.html
    log_success "クリーンアップ完了"
}

# 依存関係ダウンロード
download_deps() {
    log_info "依存関係をダウンロード中..."
    go mod download
    go mod tidy
    log_success "依存関係ダウンロード完了"
}

# テスト実行
run_tests() {
    log_info "テストを実行中..."
    go test -v ./...
    log_success "テスト完了"
}

# テストカバレッジ実行
run_test_coverage() {
    log_info "テストカバレッジを実行中..."
    go test -v -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    log_success "テストカバレッジ完了: coverage.html"
}

# リント実行
run_lint() {
    log_info "リントを実行中..."
    if command -v golangci-lint &> /dev/null; then
        golangci-lint run
        log_success "リント完了"
    else
        log_warning "golangci-lint がインストールされていません。スキップします"
    fi
}

# フォーマットチェック
check_format() {
    log_info "フォーマットをチェック中..."
    if [ -n "$(gofmt -l .)" ]; then
        log_error "フォーマットされていないファイルがあります:"
        gofmt -l .
        exit 1
    fi
    log_success "フォーマットチェック完了"
}

# 単一バイナリビルド
build_binary() {
    local binary_name=$1
    local source_path=$2
    local output_dir=${OUTPUT_DIR:-bin}
    
    log_info "$binary_name をビルド中..."
    
    mkdir -p "$output_dir"
    
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
        -o "$output_dir/$binary_name" \
        $LDFLAGS \
        "$source_path"
    
    log_success "$binary_name ビルド完了: $output_dir/$binary_name"
}

# すべてのバイナリビルド
build_all() {
    log_info "すべてのバイナリをビルド中..."
    
    build_binary "api" "./cmd/api"
    build_binary "worker" "./cmd/worker"
    build_binary "beacon-generator" "./cmd/beacon-generator"
    
    log_success "すべてのバイナリビルド完了"
}

# Dockerイメージビルド
build_docker() {
    log_info "Dockerイメージをビルド中..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker がインストールされていません"
        exit 1
    fi
    
    docker build -t "$APP_NAME:$VERSION" .
    docker tag "$APP_NAME:$VERSION" "$APP_NAME:latest"
    
    log_success "Dockerイメージビルド完了: $APP_NAME:$VERSION"
}

# セキュリティスキャン
security_scan() {
    log_info "セキュリティスキャンを実行中..."
    
    if command -v gosec &> /dev/null; then
        gosec ./...
        log_success "セキュリティスキャン完了"
    else
        log_warning "gosec がインストールされていません。スキップします"
    fi
}

# 依存関係監査
audit_deps() {
    log_info "依存関係の監査を実行中..."
    
    if command -v nancy &> /dev/null; then
        go list -json -deps ./... | nancy sleuth
        log_success "依存関係監査完了"
    else
        log_warning "nancy がインストールされていません。スキップします"
    fi
}

# メイン処理
main() {
    local clean_build=false
    local run_tests_flag=false
    local build_docker_flag=false
    local build_all_flag=false
    local run_coverage=false
    local run_lint_flag=false
    local security_scan_flag=false
    local audit_deps_flag=false
    local check_format_flag=false
    
    # オプション解析
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -v|--version)
                echo "$VERSION"
                exit 0
                ;;
            -c|--clean)
                clean_build=true
                shift
                ;;
            -t|--test)
                run_tests_flag=true
                shift
                ;;
            -d|--docker)
                build_docker_flag=true
                shift
                ;;
            -a|--all)
                build_all_flag=true
                shift
                ;;
            -o|--output)
                OUTPUT_DIR="$2"
                shift 2
                ;;
            -e|--env)
                BUILD_ENV="$2"
                shift 2
                ;;
            --coverage)
                run_coverage=true
                shift
                ;;
            --lint)
                run_lint_flag=true
                shift
                ;;
            --security)
                security_scan_flag=true
                shift
                ;;
            --audit)
                audit_deps_flag=true
                shift
                ;;
            --format-check)
                check_format_flag=true
                shift
                ;;
            *)
                log_error "不明なオプション: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 初期化
    init
    
    # 依存関係チェック
    check_dependencies
    
    # クリーンアップ
    if [ "$clean_build" = true ]; then
        clean
    fi
    
    # フォーマットチェック
    if [ "$check_format_flag" = true ]; then
        check_format
    fi
    
    # 依存関係ダウンロード
    download_deps
    
    # リント実行
    if [ "$run_lint_flag" = true ]; then
        run_lint
    fi
    
    # セキュリティスキャン
    if [ "$security_scan_flag" = true ]; then
        security_scan
    fi
    
    # 依存関係監査
    if [ "$audit_deps_flag" = true ]; then
        audit_deps
    fi
    
    # ビルド実行
    if [ "$build_all_flag" = true ]; then
        build_all
    else
        build_binary "api" "./cmd/api"
    fi
    
    # テスト実行
    if [ "$run_tests_flag" = true ]; then
        run_tests
    fi
    
    # テストカバレッジ実行
    if [ "$run_coverage" = true ]; then
        run_test_coverage
    fi
    
    # Dockerイメージビルド
    if [ "$build_docker_flag" = true ]; then
        build_docker
    fi
    
    log_success "ビルド完了！"
    
    # 出力情報
    if [ "$build_all_flag" = true ]; then
        echo ""
        echo "ビルドされたバイナリ:"
        ls -la "${OUTPUT_DIR:-bin}/"
    fi
    
    if [ "$build_docker_flag" = true ]; then
        echo ""
        echo "Dockerイメージ:"
        docker images | grep "$APP_NAME"
    fi
}

# スクリプト実行
main "$@"
