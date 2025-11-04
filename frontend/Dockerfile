# ビルドステージ
FROM node:20-alpine AS builder

WORKDIR /app

# pnpmをインストール
RUN corepack enable && corepack prepare pnpm@10.19.0 --activate

# ワークスペース設定をコピー
COPY pnpm-workspace.yaml package.json ./
COPY pnpm-lock.yaml ./

# 全ワークスペースのpackage.jsonをコピー
COPY frontend/package.json ./frontend/
COPY backend/package.json ./backend/

# 依存関係をインストール
RUN pnpm install --frozen-lockfile

# フロントエンドのソースコードをコピー
COPY frontend ./frontend

WORKDIR /app/frontend

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
