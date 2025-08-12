#!/bin/bash

# Access Log Tracker - デプロイスクリプト

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
DEFAULT_ENV="development"

# ヘルプ表示
show_help() {
    cat << EOF
Access Log Tracker - デプロイスクリプト

使用方法:
    $0 [オプション] [環境]

オプション:
    -h, --help          このヘルプを表示
    -v, --version       バージョンを表示
    -e, --env ENV       環境を指定 (dev, staging, prod)
    -t, --target TARGET デプロイターゲット (local, docker, k8s, aws)
    -c, --config FILE   設定ファイルを指定
    -d, --dry-run       ドライラン（実際のデプロイは行わない）
    -f, --force         確認なしでデプロイ
    -r, --rollback      ロールバックを実行
    -s, --status        デプロイ状況を確認

環境:
    dev                 開発環境
    staging             ステージング環境
    prod                本番環境

ターゲット:
    local              ローカル環境
    docker             Docker Compose
    k8s                Kubernetes
    aws                AWS (CloudFormation/Terraform)

例:
    $0 dev                    # 開発環境にデプロイ
    $0 -e prod -t aws        # 本番環境にAWSデプロイ
    $0 -e staging -t k8s     # ステージング環境にK8sデプロイ
    $0 -d -e prod            # 本番環境のドライラン
    $0 -s -e prod            # 本番環境の状況確認

EOF
}

# 初期化
init() {
    log_info "Access Log Tracker デプロイを開始します"
    log_info "バージョン: $VERSION"
    log_info "環境: $DEPLOY_ENV"
    log_info "ターゲット: $DEPLOY_TARGET"
}

# 依存関係チェック
check_dependencies() {
    log_info "依存関係をチェック中..."
    
    case $DEPLOY_TARGET in
        docker)
            if ! command -v docker &> /dev/null; then
                log_error "Docker がインストールされていません"
                exit 1
            fi
            if ! command -v docker-compose &> /dev/null; then
                log_error "Docker Compose がインストールされていません"
                exit 1
            fi
            ;;
        k8s)
            if ! command -v kubectl &> /dev/null; then
                log_error "kubectl がインストールされていません"
                exit 1
            fi
            ;;
        aws)
            if ! command -v aws &> /dev/null; then
                log_error "AWS CLI がインストールされていません"
                exit 1
            fi
            ;;
    esac
    
    log_success "依存関係チェック完了"
}

# 環境変数読み込み
load_environment() {
    local env_file=".env.$DEPLOY_ENV"
    
    if [ -f "$env_file" ]; then
        log_info "環境変数を読み込み中: $env_file"
        export $(cat "$env_file" | grep -v '^#' | xargs)
    else
        log_warning "環境変数ファイルが見つかりません: $env_file"
    fi
}

# 設定ファイル検証
validate_config() {
    log_info "設定ファイルを検証中..."
    
    # 必須設定のチェック
    local required_vars=("APP_NAME" "APP_PORT")
    
    for var in "${required_vars[@]}"; do
        if [ -z "${!var}" ]; then
            log_error "必須環境変数が設定されていません: $var"
            exit 1
        fi
    done
    
    log_success "設定ファイル検証完了"
}

# ビルド実行
build_application() {
    log_info "アプリケーションをビルド中..."
    
    if [ "$DRY_RUN" = true ]; then
        log_info "ドライラン: ビルドをスキップ"
        return
    fi
    
    ./scripts/build.sh -c -t -d
    log_success "アプリケーションビルド完了"
}

# ローカルデプロイ
deploy_local() {
    log_info "ローカル環境にデプロイ中..."
    
    if [ "$DRY_RUN" = true ]; then
        log_info "ドライラン: ローカルデプロイをスキップ"
        return
    fi
    
    # アプリケーションを停止
    pkill -f "$APP_NAME" || true
    
    # アプリケーションを起動
    nohup ./bin/api > logs/app.log 2>&1 &
    
    log_success "ローカルデプロイ完了"
}

# Dockerデプロイ
deploy_docker() {
    log_info "Docker環境にデプロイ中..."
    
    if [ "$DRY_RUN" = true ]; then
        log_info "ドライラン: Dockerデプロイをスキップ"
        return
    fi
    
    # Docker Composeでデプロイ
    docker-compose -f docker-compose.yml down
    docker-compose -f docker-compose.yml up -d
    
    log_success "Dockerデプロイ完了"
}

# Kubernetesデプロイ
deploy_kubernetes() {
    log_info "Kubernetes環境にデプロイ中..."
    
    if [ "$DRY_RUN" = true ]; then
        log_info "ドライラン: Kubernetesデプロイをスキップ"
        return
    fi
    
    # 名前空間を作成
    kubectl create namespace "$APP_NAME" --dry-run=client -o yaml | kubectl apply -f -
    
    # ConfigMapとSecretを適用
    kubectl apply -f deployments/kubernetes/configmap.yaml
    kubectl apply -f deployments/kubernetes/secret.yaml
    
    # デプロイメントを適用
    kubectl apply -f deployments/kubernetes/deployment.yaml
    kubectl apply -f deployments/kubernetes/service.yaml
    kubectl apply -f deployments/kubernetes/ingress.yaml
    
    # デプロイメントの確認
    kubectl rollout status deployment/"$APP_NAME" -n "$APP_NAME"
    
    log_success "Kubernetesデプロイ完了"
}

# AWSデプロイ
deploy_aws() {
    log_info "AWS環境にデプロイ中..."
    
    if [ "$DRY_RUN" = true ]; then
        log_info "ドライラン: AWSデプロイをスキップ"
        return
    fi
    
    # CloudFormationでデプロイ
    aws cloudformation deploy \
        --template-file deployments/aws/cloudformation/infrastructure.yml \
        --stack-name "$APP_NAME-$DEPLOY_ENV" \
        --parameter-overrides \
            Environment="$DEPLOY_ENV" \
            Version="$VERSION" \
        --capabilities CAPABILITY_IAM \
        --tags \
            Environment="$DEPLOY_ENV" \
            Application="$APP_NAME" \
            Version="$VERSION"
    
    log_success "AWSデプロイ完了"
}

# ヘルスチェック
health_check() {
    log_info "ヘルスチェックを実行中..."
    
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -f "http://localhost:$APP_PORT/health" > /dev/null 2>&1; then
            log_success "ヘルスチェック成功"
            return 0
        fi
        
        log_info "ヘルスチェック試行 $attempt/$max_attempts"
        sleep 2
        ((attempt++))
    done
    
    log_error "ヘルスチェック失敗"
    return 1
}

# ロールバック
rollback() {
    log_info "ロールバックを実行中..."
    
    case $DEPLOY_TARGET in
        docker)
            docker-compose -f docker-compose.yml down
            docker-compose -f docker-compose.yml up -d
            ;;
        k8s)
            kubectl rollout undo deployment/"$APP_NAME" -n "$APP_NAME"
            kubectl rollout status deployment/"$APP_NAME" -n "$APP_NAME"
            ;;
        aws)
            aws cloudformation rollback-stack \
                --stack-name "$APP_NAME-$DEPLOY_ENV"
            ;;
        *)
            log_error "サポートされていないターゲット: $DEPLOY_TARGET"
            exit 1
            ;;
    esac
    
    log_success "ロールバック完了"
}

# デプロイ状況確認
check_status() {
    log_info "デプロイ状況を確認中..."
    
    case $DEPLOY_TARGET in
        local)
            if pgrep -f "$APP_NAME" > /dev/null; then
                log_success "アプリケーションは実行中です"
            else
                log_error "アプリケーションは停止中です"
            fi
            ;;
        docker)
            docker-compose -f docker-compose.yml ps
            ;;
        k8s)
            kubectl get pods -n "$APP_NAME"
            kubectl get services -n "$APP_NAME"
            kubectl get ingress -n "$APP_NAME"
            ;;
        aws)
            aws cloudformation describe-stacks \
                --stack-name "$APP_NAME-$DEPLOY_ENV" \
                --query 'Stacks[0].StackStatus' \
                --output text
            ;;
        *)
            log_error "サポートされていないターゲット: $DEPLOY_TARGET"
            exit 1
            ;;
    esac
}

# 確認プロンプト
confirm_deploy() {
    if [ "$FORCE" = true ]; then
        return 0
    fi
    
    echo ""
    echo "デプロイ情報:"
    echo "  アプリケーション: $APP_NAME"
    echo "  バージョン: $VERSION"
    echo "  環境: $DEPLOY_ENV"
    echo "  ターゲット: $DEPLOY_TARGET"
    echo ""
    
    read -p "デプロイを続行しますか？ (y/N): " -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "デプロイをキャンセルしました"
        exit 0
    fi
}

# メイン処理
main() {
    local show_help_flag=false
    local show_version_flag=false
    local dry_run_flag=false
    local force_flag=false
    local rollback_flag=false
    local status_flag=false
    local deploy_env="$DEFAULT_ENV"
    local deploy_target="local"
    local config_file=""
    
    # オプション解析
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help_flag=true
                shift
                ;;
            -v|--version)
                show_version_flag=true
                shift
                ;;
            -e|--env)
                deploy_env="$2"
                shift 2
                ;;
            -t|--target)
                deploy_target="$2"
                shift 2
                ;;
            -c|--config)
                config_file="$2"
                shift 2
                ;;
            -d|--dry-run)
                dry_run_flag=true
                shift
                ;;
            -f|--force)
                force_flag=true
                shift
                ;;
            -r|--rollback)
                rollback_flag=true
                shift
                ;;
            -s|--status)
                status_flag=true
                shift
                ;;
            dev|staging|prod)
                deploy_env="$1"
                shift
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
    
    # バージョン表示
    if [ "$show_version_flag" = true ]; then
        echo "$VERSION"
        exit 0
    fi
    
    # グローバル変数設定
    DEPLOY_ENV="$deploy_env"
    DEPLOY_TARGET="$deploy_target"
    DRY_RUN="$dry_run_flag"
    FORCE="$force_flag"
    
    # 初期化
    init
    
    # 依存関係チェック
    check_dependencies
    
    # 環境変数読み込み
    load_environment
    
    # 設定ファイル検証
    validate_config
    
    # ロールバック実行
    if [ "$rollback_flag" = true ]; then
        rollback
        exit 0
    fi
    
    # 状況確認
    if [ "$status_flag" = true ]; then
        check_status
        exit 0
    fi
    
    # 確認プロンプト
    confirm_deploy
    
    # ビルド実行
    build_application
    
    # デプロイ実行
    case $DEPLOY_TARGET in
        local)
            deploy_local
            ;;
        docker)
            deploy_docker
            ;;
        k8s)
            deploy_kubernetes
            ;;
        aws)
            deploy_aws
            ;;
        *)
            log_error "サポートされていないターゲット: $DEPLOY_TARGET"
            exit 1
            ;;
    esac
    
    # ヘルスチェック
    if [ "$DRY_RUN" = false ]; then
        health_check
    fi
    
    log_success "デプロイ完了！"
    
    # デプロイ情報表示
    echo ""
    echo "デプロイ情報:"
    echo "  アプリケーション: $APP_NAME"
    echo "  バージョン: $VERSION"
    echo "  環境: $DEPLOY_ENV"
    echo "  ターゲット: $DEPLOY_TARGET"
    echo "  ヘルスチェック: http://localhost:$APP_PORT/health"
    echo ""
}

# スクリプト実行
main "$@"
