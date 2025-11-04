#!/bin/bash

set -e

# 色付きログ出力
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

# ブランチ名のチェック
if [ -z "$1" ]; then
    log_error "ブランチ名が指定されていません。使用方法: ./cleanup-preview.sh <branch-name>"
    exit 1
fi

BRANCH=$1
PREVIEW_DIR="/opt/chat-preview/${BRANCH}"

log_info "プレビュー環境をクリーンアップします: ${BRANCH}"

# ディレクトリの存在確認
if [ ! -d "$PREVIEW_DIR" ]; then
    log_warn "プレビュー環境が見つかりません: ${PREVIEW_DIR}"
    exit 0
fi

cd $PREVIEW_DIR

# Dockerコンテナを停止・削除
log_info "Dockerコンテナを停止・削除中..."
docker compose -f docker-compose.production.yml -p chat-preview-${BRANCH} down -v

# イメージも削除する場合
log_info "未使用のDockerイメージを削除中..."
docker image prune -f

# ディレクトリを削除
log_info "ディレクトリを削除中..."
cd /opt/chat-preview
rm -rf $PREVIEW_DIR

log_info "✅ プレビュー環境のクリーンアップが完了しました: ${BRANCH}"
