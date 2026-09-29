# ビルドステージ
FROM node:26.10-alpine AS builder

WORKDIR /app

# Node 25 以降 corepack は同梱されない
RUN npm install -g pnpm@12.5.1

# ワークスペース設定をコピー
COPY pnpm-workspace.yaml package.json ./
COPY pnpm-lock.yaml ./

# 全ワークスペースのpackage.jsonをコピー
COPY frontend/package.json ./frontend/
COPY backend/package.json ./backend/
COPY packages/design-tokens/package.json ./packages/design-tokens/
COPY packages/i18n/package.json ./packages/i18n/
COPY packages/search-query/package.json ./packages/search-query/

# 依存関係をインストール
RUN pnpm install --frozen-lockfile

# フロントエンドのソースコードをコピー
COPY packages ./packages
COPY frontend ./frontend

WORKDIR /app/frontend

# Vite はビルド時に import.meta.env へ埋め込むため、環境ごとの値はビルド引数で渡す
ARG VITE_API_BASE_URL
ARG VITE_WS_URL
ARG VITE_GOOGLE_OAUTH_CLIENT_ID
ARG VITE_FIREBASE_API_KEY
ARG VITE_FIREBASE_PROJECT_ID
ARG VITE_FIREBASE_MESSAGING_SENDER_ID
ARG VITE_FIREBASE_APP_ID
ARG VITE_FIREBASE_VAPID_KEY

# ビルド実行
RUN pnpm run build

# 実行ステージ
FROM nginx:alpine

# Nginxの設定をコピー
COPY frontend/nginx.conf /etc/nginx/conf.d/default.conf

# ビルドステージから生成物をコピー
COPY --from=builder /app/frontend/dist /usr/share/nginx/html

# ポートを公開
EXPOSE 80

# Nginxを起動
CMD ["nginx", "-g", "daemon off;"]
