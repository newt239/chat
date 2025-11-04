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
    log_error "環境が指定されていません。使用方法: ./deploy.sh <production|preview> [branch]"
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

    # デプロイ前バックアップ（Wasabiへアップロード）
    log_info "デプロイ前バックアップを実行します..."
    if ./scripts/backup.sh; then
        log_info "バックアップ完了"
    else
        log_error "バックアップに失敗しました。デプロイを中止します。"
        exit 1
    fi

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
    log_info "プレビュー環境へデプロイします（共通DB使用）..."

    if [ -z "$BRANCH" ]; then
        log_error "ブランチ名が指定されていません"
        exit 1
    fi

    # ブランチ名をサニタイズ（スラッシュをハイフンに変換）
    SANITIZED_BRANCH=$(echo $BRANCH | sed 's/\//-/g')

    # プレビュー環境のディレクトリを作成
    PREVIEW_DIR="/opt/chat-preview/${SANITIZED_BRANCH}"
    mkdir -p $PREVIEW_DIR

    cd $PREVIEW_DIR

    # リポジトリのクローンまたは更新
    if [ ! -d .git ]; then
        log_info "リポジトリをクローン中..."
        git clone /opt/chat/.git .
    fi

    log_info "最新のコードを取得中..."
    git fetch origin
    git checkout ${BRANCH}
    git reset --hard origin/${BRANCH}

    # .envファイルの存在確認（プレビュー環境用）
    if [ ! -f .env.preview ]; then
        log_warn ".env.previewファイルが見つかりません。本番環境の設定をコピーします"
        cp /opt/chat/.env.production .env.preview
        log_warn "⚠️  .env.previewを確認して、以下の設定を変更してください:"
        log_warn "  - PREVIEW_PORT (例: 8081, 8082など、本番と重複しないポート)"
        log_warn "  - CADDY_HTTP_PORT, CADDY_HTTPS_PORT (例: 8080/8443など)"
        log_warn "  - PREVIEW_DOMAIN (例: preview-${SANITIZED_BRANCH}.your-domain.com)"
        log_warn "  - DATABASE_URLは本番と同じでOK（共通DB使用）"
        exit 1
    fi

    # 環境変数をロード
    export $(cat .env.preview | grep -v '^#' | xargs)
    export BRANCH_NAME=${SANITIZED_BRANCH}

    # プロジェクト名を設定
    PROJECT_NAME="chat-preview-${SANITIZED_BRANCH}"

    # Dockerコンテナを停止
    log_info "既存のコンテナを停止中..."
    docker compose -f docker-compose.preview.yml -p ${PROJECT_NAME} down

    # Dockerイメージをビルドして起動
    log_info "Dockerイメージをビルド中..."
    docker compose -f docker-compose.preview.yml -p ${PROJECT_NAME} build --no-cache

    log_info "コンテナを起動中..."
    docker compose -f docker-compose.preview.yml -p ${PROJECT_NAME} up -d

    # ヘルスチェック
    log_info "ヘルスチェックを実行中..."
    sleep 10

    if docker compose -f docker-compose.preview.yml -p ${PROJECT_NAME} ps | grep -q "Up"; then
        log_info "✅ プレビュー環境のデプロイが完了しました"
        log_info "ブランチ: ${BRANCH}"
        log_info "プロジェクト名: ${PROJECT_NAME}"
        log_info "📍 注意: データベースは本番環境と共通です"
        docker compose -f docker-compose.preview.yml -p ${PROJECT_NAME} ps
    else
        log_error "❌ デプロイに失敗しました"
        docker compose -f docker-compose.preview.yml -p ${PROJECT_NAME} logs --tail=50
        exit 1
    fi
else
    log_error "不正な環境が指定されました: ${ENVIRONMENT}"
    log_error "使用可能な環境: production, preview"
    exit 1
fi
