#!/bin/bash

set -e

# 色付きログ出力
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# 使用方法
usage() {
    cat <<EOF
使用方法: $0 <action> [environment]

Actions:
  apply       - マイグレーションを適用（entの自動マイグレーション）
  seed        - シードデータを投入

Environment:
  production  - 本番環境（デフォルト）
  preview     - プレビュー環境（全ブランチ共通のデータベース）

注意:
  - プレビュー環境は専用のデータベースを使用します
  - プレビュー環境のデータは全てのブランチで共有されます

例:
  $0 apply production    # 本番環境にマイグレーションを適用
  $0 apply preview       # プレビュー環境にマイグレーションを適用
  $0 seed production     # 本番環境にシードデータを投入
  $0 seed preview        # プレビュー環境にシードデータを投入
EOF
    exit 1
}

# 引数チェック
if [ -z "$1" ]; then
    log_error "アクションが指定されていません"
    usage
fi

ACTION=$1
ENVIRONMENT=${2:-production}

log_info "アクション: ${ACTION}"
log_info "環境: ${ENVIRONMENT}"

# 環境に応じた設定
if [ "$ENVIRONMENT" = "production" ]; then
    CONTAINER="chat-backend-prod"
    ENV_NAME="本番環境"
elif [ "$ENVIRONMENT" = "preview" ]; then
    # プレビュー環境は任意のブランチのbackendコンテナを使用
    # 複数ある場合は最初のものを使用
    CONTAINER=$(docker ps --filter "name=chat-backend-preview-" --format "{{.Names}}" | head -n 1)
    if [ -z "$CONTAINER" ]; then
        log_error "プレビュー環境のバックエンドコンテナが見つかりません"
        log_error "先にプレビュー環境をデプロイしてください"
        exit 1
    fi
    ENV_NAME="プレビュー環境（全ブランチ共通）"
else
    log_error "不正な環境が指定されました: ${ENVIRONMENT}"
    usage
fi

log_info "対象コンテナ: ${CONTAINER}"

# アクションの実行
case "$ACTION" in
    apply)
        log_step "マイグレーションを適用します（entの自動マイグレーション）"
        log_info "環境: ${ENV_NAME}"

        # entの自動マイグレーションを実行
        docker exec ${CONTAINER} ./migrate

        log_info "✅ マイグレーションの適用が完了しました"
        ;;

    seed)
        log_step "シードデータを投入します"
        log_info "環境: ${ENV_NAME}"

        if [ "$ENVIRONMENT" = "production" ]; then
            log_warn "⚠️  本番環境にシードデータを投入します"
        else
            log_warn "⚠️  プレビュー環境にシードデータを投入します"
            log_warn "⚠️  全てのプレビューブランチで同じデータが使用されます"
        fi
        log_warn "⚠️  既存のデータに影響を与える可能性があります"
        read -p "本当に実行しますか? (yes/no): " confirm
        if [ "$confirm" != "yes" ]; then
            log_info "キャンセルされました"
            exit 0
        fi

        docker exec ${CONTAINER} ./seed

        log_info "✅ シードデータの投入が完了しました"
        ;;

    *)
        log_error "不正なアクションが指定されました: ${ACTION}"
        usage
        ;;
esac
