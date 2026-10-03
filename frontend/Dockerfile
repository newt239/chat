FROM node:26-alpine AS base

WORKDIR /app

COPY pnpm-workspace.yaml package.json ./
# pnpm が自分で版を切り替えると minimumReleaseAge に阻まれるため packageManager の版を入れる
RUN npm install -g "$(node -p "require('./package.json').packageManager")"
COPY pnpm-lock.yaml ./
COPY frontend/package.json ./frontend/
COPY packages/design-tokens/package.json ./packages/design-tokens/
COPY packages/i18n/package.json ./packages/i18n/
RUN pnpm install --frozen-lockfile

COPY packages ./packages
COPY frontend ./frontend

WORKDIR /app/frontend

FROM base AS dev

EXPOSE 5173

# ボリュームの node_modules が空のときに備えて起動前にも入れる
CMD ["sh", "-c", "pnpm install --frozen-lockfile && pnpm run dev"]

FROM base AS builder

# Vite はビルド時に import.meta.env へ埋め込むため、環境ごとの値はビルド引数で渡す
ARG VITE_API_BASE_URL
ARG VITE_WS_URL
ARG VITE_GOOGLE_OAUTH_CLIENT_ID
ARG VITE_FIREBASE_API_KEY
ARG VITE_FIREBASE_PROJECT_ID
ARG VITE_FIREBASE_MESSAGING_SENDER_ID
ARG VITE_FIREBASE_APP_ID
ARG VITE_FIREBASE_VAPID_KEY

RUN pnpm run build

FROM nginx:alpine

COPY frontend/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/frontend/dist /usr/share/nginx/html
