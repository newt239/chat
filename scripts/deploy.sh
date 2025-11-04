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

# 環境変数のチェック
if [ -z "$1" ]; then
    log_error "環境が指定されていません。使用方法: ./deploy.sh <production|preview>"
    exit 1
fi

ENVIRONMENT=$1
BRANCH=${2:-main}

log_info "デプロイ環境: ${ENVIRONMENT}"
log_info "ブランチ: ${BRANCH}"

# 本番環境へのデプロイ
if [ "$ENVIRONMENT" = "production" ]; then
    log_info "本番環境へデプロイします..."

    # .envファイルの存在確認
    if [ ! -f .env.production ]; then
        log_error ".env.productionファイルが見つかりません"
        exit 1
    fi

    # 最新コードを取得
    log_info "最新のコードを取得中..."
    git fetch origin
    git reset --hard origin/${BRANCH}

    # 環境変数をロード
    export $(cat .env.production | grep -v '^#' | xargs)

    # Dockerコンテナを停止
    log_info "既存のコンテナを停止中..."
    docker compose -f docker-compose.production.yml down

    # Dockerイメージをビルドして起動
    log_info "Dockerイメージをビルド中..."
    docker compose -f docker-compose.production.yml build --no-cache

    log_info "コンテナを起動中..."
    docker compose -f docker-compose.production.yml up -d

    # ヘルスチェック
    log_info "ヘルスチェックを実行中..."
    sleep 10

    if docker compose -f docker-compose.production.yml ps | grep -q "Up"; then
        log_info "✅ デプロイが完了しました"
        docker compose -f docker-compose.production.yml ps
    else
        log_error "❌ デプロイに失敗しました"
        docker compose -f docker-compose.production.yml logs --tail=50
        exit 1
    fi

# プレビュー環境へのデプロイ
elif [ "$ENVIRONMENT" = "preview" ]; then
    log_info "プレビュー環境へデプロイします..."

    if [ -z "$BRANCH" ]; then
        log_error "ブランチ名が指定されていません"
        exit 1
    fi

    # プレビュー環境のディレクトリを作成
    PREVIEW_DIR="/opt/chat-preview/${BRANCH}"
    mkdir -p $PREVIEW_DIR

    cd $PREVIEW_DIR

    # リポジトリのクローンまたは更新
    if [ ! -d .git ]; then
        log_info "リポジトリをクローン中..."
        git clone $(git config --get remote.origin.url) .
    fi

    log_info "最新のコードを取得中..."
    git fetch origin
    git checkout ${BRANCH}
    git reset --hard origin/${BRANCH}

    # .envファイルの存在確認（プレビュー環境用）
    if [ ! -f .env.preview ]; then
        log_warn ".env.previewファイルが見つかりません。.env.production.exampleをコピーしてください"
    else
        export $(cat .env.preview | grep -v '^#' | xargs)
    fi

    # Dockerコンテナを停止
    log_info "既存のコンテナを停止中..."
    docker compose -f docker-compose.production.yml -p chat-preview-${BRANCH} down

    # Dockerイメージをビルドして起動
    log_info "Dockerイメージをビルド中..."
    docker compose -f docker-compose.production.yml -p chat-preview-${BRANCH} build --no-cache

    log_info "コンテナを起動中..."
    docker compose -f docker-compose.production.yml -p chat-preview-${BRANCH} up -d

    # ヘルスチェック
    log_info "ヘルスチェックを実行中..."
    sleep 10

    if docker compose -f docker-compose.production.yml -p chat-preview-${BRANCH} ps | grep -q "Up"; then
        log_info "✅ プレビュー環境のデプロイが完了しました"
        docker compose -f docker-compose.production.yml -p chat-preview-${BRANCH} ps
    else
        log_error "❌ デプロイに失敗しました"
        docker compose -f docker-compose.production.yml -p chat-preview-${BRANCH} logs --tail=50
        exit 1
    fi
else
    log_error "不正な環境が指定されました: ${ENVIRONMENT}"
    log_error "使用可能な環境: production, preview"
    exit 1
fi
