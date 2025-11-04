#!/bin/bash

set -e

# 色付きログ
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 作業ディレクトリをリポジトリルートに固定（/opt/chat を想定）
REPO_DIR="/opt/chat"
cd "${REPO_DIR}" 2>/dev/null || true

# 環境変数のロード (.env.production を想定)
if [ -f .env.production ]; then
    export $(grep -v '^#' .env.production | xargs)
else
    log_warn ".env.production が見つかりません。環境変数は事前にエクスポートされている前提で続行します。"
fi

# 必須環境変数チェック
REQUIRED_VARS=(POSTGRES_USER POSTGRES_DB WASABI_ACCESS_KEY_ID WASABI_SECRET_ACCESS_KEY WASABI_BUCKET_NAME)
for v in "${REQUIRED_VARS[@]}"; do
    if [ -z "${!v}" ]; then
        log_error "必須環境変数 ${v} が未設定です"
        exit 1
    fi
done

# Wasabi 設定（デフォルト値）
WASABI_REGION=${WASABI_REGION:-us-east-1}
WASABI_ENDPOINT=${WASABI_ENDPOINT:-https://s3.wasabisys.com}

# バックアップ出力先
BACKUP_DIR="/opt/backups"
mkdir -p "${BACKUP_DIR}"

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
DB_NAME="${POSTGRES_DB}"
BACKUP_BASENAME="${DB_NAME}_${TIMESTAMP}.sql"
BACKUP_PATH="${BACKUP_DIR}/${BACKUP_BASENAME}"
BACKUP_GZ_PATH="${BACKUP_PATH}.gz"
S3_KEY="db-backups/${BACKUP_BASENAME}.gz"

log_info "データベースダンプを作成します (${DB_NAME})"

# pg_dump 実行（db サービスに対して）
docker compose -f docker-compose.production.yml exec -T db \
  pg_dump -U "${POSTGRES_USER}" "${DB_NAME}" > "${BACKUP_PATH}"

log_info "圧縮中: ${BACKUP_PATH} -> ${BACKUP_GZ_PATH}"
gzip -f "${BACKUP_PATH}"

# aws CLI の確認
if ! command -v aws >/dev/null 2>&1; then
    log_error "aws CLI が見つかりません。サーバーに awscli をインストールしてください (例: apt install awscli)"
    exit 1
fi

log_info "Wasabi へアップロードします: s3://${WASABI_BUCKET_NAME}/${S3_KEY}"
AWS_ACCESS_KEY_ID="${WASABI_ACCESS_KEY_ID}" \
AWS_SECRET_ACCESS_KEY="${WASABI_SECRET_ACCESS_KEY}" \
AWS_DEFAULT_REGION="${WASABI_REGION}" \
  aws --endpoint-url="${WASABI_ENDPOINT}" s3 cp "${BACKUP_GZ_PATH}" "s3://${WASABI_BUCKET_NAME}/${S3_KEY}"

log_info "アップロード完了: ${BACKUP_GZ_PATH}"

# ローカルのバックアップ保持ポリシー（任意: 30日超を削除）
find "${BACKUP_DIR}" -name "${DB_NAME}_*.sql.gz" -mtime +30 -print -delete || true

log_info "バックアップ処理が完了しました"


