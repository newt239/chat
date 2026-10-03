# Chat Application

## クイックスタート

```bash
# 1. Docker Desktopを起動
# 2. 依存関係をインストール
pnpm install
# 3. アプリケーションを起動（スキーマのリセットとシードデータは自動実行されます）
pnpm start
```

→ https://chat.localhost にアクセス

コンテナのホストポートは自動で割り当て、[portless](https://github.com/vercel-labs/portless) で名前付きの URL に振り分けるため、他のプロジェクトとポートやコンテナ名がぶつかりません。git worktree で起動すると `https://<ブランチ名>.chat.localhost` になり、メインの環境と並べて動かせます。初回の `pnpm start` では portless のプロキシ（443 番）の起動とローカル CA の信頼登録のために sudo のパスワードを求められます。

### 外部に公開する

Cloudflare Tunnel で、手元の環境をスマホなどの外部の端末から開けるようにできます。

```bash
# 初回だけ: Cloudflare の CLI を入れてログインし、公開に使うドメイン（Cloudflare のゾーン）を設定する
npm i -g cf && cf auth login
git config chat.tunnelDomain newt239.dev

pnpm public          # 公開する → https://chat-local.newt239.dev
pnpm private         # 非公開に戻す
pnpm tunnel:status   # 今の状態を表示する
```

- 公開中は、`https://chat.localhost` から開いても API は公開用の URL を通ります。
- worktree では `https://<ブランチ名>-chat-local.<ドメイン>` になります。Tunnel（`chat-local`）とコネクタのコンテナ（`chat-tunnel`）はメインと各 worktree で共有し、公開・非公開はホスト名ごとの振り分けルールと DNS レコードの追加・削除で切り替えます。
- 公開中は、URL を知っていれば誰でもシードのアカウント（`alice@example.com`）でログインできます。使い終わったら `pnpm private` で戻してください。

### 利用可能なコマンド

```bash
# アプリケーションを起動（ポートが変わったときも再実行すれば URL の振り分けを登録し直す）
pnpm start

# アプリケーションを停止
pnpm stop

# データベーススキーマをリセット
pnpm db:reset

# 性能検証用に大量のメッセージを投入（初期データは起動時に自動で作られます）
pnpm db:seed:bulk

# バックエンドコードのリント
docker compose exec backend golangci-lint run

# ログを表示
pnpm logs

# コンテナの状態を確認
docker compose ps
```

### フロントエンドの開発コマンド

lint / format / test の設定は `frontend/vite.config.ts` に集約されています（Vite+ による統合）。

```bash
# 型チェック・lint・format・未使用コード検出・テストを一括実行
pnpm --filter chat-frontend run codecheck

# 個別に実行
pnpm --filter chat-frontend run typecheck
pnpm --filter chat-frontend run lint:fix
pnpm --filter chat-frontend run format:fix
pnpm --filter chat-frontend run test

# proto を変更したとき（リポジトリルートで実行）
pnpm run proto:format && pnpm run proto:lint && pnpm run generate:proto
```

コミット時は lefthook の pre-commit フックが lint・format を実行します。ホスト側に `pnpm install` 済みであることが前提のため、Docker のみで開発している場合は `LEFTHOOK=0 git commit` で回避できます。

### テストアカウント

- **メールアドレス**: alice@example.com
- **パスワード**: password123

詳細なセットアップ手順は [ローカル環境のセットアップ](#ローカル環境のセットアップ) を参照してください。

## 技術スタック

### バックエンド

- Go 1.27
- Echo
- Connect RPC (connect-go) + Protocol Buffers
- WebSocket (gorilla/websocket)
- ent (ORM)
- PostgreSQL 18
- Redis（WebSocket の配信などをレプリカ間で共有）
- Wasabi

### フロントエンド

- React 19
- TypeScript 7
- Vite+ (`vite-plus`) — Vite 8 / Vitest / Oxlint / Oxfmt を統合したツールチェーン
- Mantine 8
- Tailwind CSS 4
- TanStack Router (ファイルベースルーティング / SPA)
- TanStack Query + connect-query
- PWA (vite-plugin-pwa)

### 開発ツール

- pnpm 11 (workspace) + Turborepo
- lefthook (pre-commit フック)
- knip (未使用コード検出)
- buf (`proto/` から Go と TypeScript のコードを生成)

### インフラ

- Docker Compose
- nginx (本番の静的配信)

## プロジェクト構造

```bash
chat/
├── backend/          # Go backend
│   ├── cmd/
│   │   ├── server/  # Main application entry point
│   │   ├── reset/   # Database schema reset tool
│   │   └── seed/    # 大量データの投入
│   ├── internal/
│   │   ├── domain/         # Domain entities & repository interfaces
│   │   ├── usecase/        # Business logic
│   │   ├── interfaces/handler/
│   │   │   ├── http/       # HTTP handlers & routes
│   │   │   └── websocket/ # WebSocket hub & connections
│   │   └── infrastructure/
│   │       ├── auth/       # JWT & password hashing
│   │       ├── config/     # Configuration management
│   │       ├── database/   # ent client connection
│   │       ├── logger/     # Zap logger setup
│   │       ├── redis/      # Redis Pub/Sub・閲覧者・レート制限
│   │       ├── repository/ # Repository implementation
│   │       ├── storage/    # Wasabi S3 client
│   │       └── utils/      # Utility functions
│   └── ent/              # ent schema definitions & generated code
├── frontend/         # React frontend
│   ├── src/
│   │   ├── routes/   # TanStack Router のファイルベースルート定義
│   │   ├── components/ # 汎用コンポーネント（ui/・block/）
│   │   ├── features/ # Feature-based modules
│   │   ├── hooks/    # 複数の機能で使う hooks
│   │   ├── providers/ # Jotai ストア・TanStack Query・WebSocket の Provider
│   │   └── lib/      # API client, WS client, router など
│   ├── tests/        # Vitest のセットアップ
│   └── public/       # Static assets（PWA アイコンの元になる logo.svg）
├── proto/            # Protocol Buffers の API 定義（buf で Go / TypeScript を生成）
└── scripts/          # 開発用スクリプト

```

## ローカル環境のセットアップ

### 起動方法

#### 必要な環境

- **Docker Desktop**

#### 手順

```bash
# 1. リポジトリのクローン
git clone <repository-url>
cd chat

# 2. アプリケーションを起動（スキーマのリセットとシードデータは自動実行されます）
pnpm install
pnpm start

# 3. 起動完了後、https://chat.localhost にアクセス
```

#### 停止方法

```bash
# アプリケーションを停止
pnpm stop

# データベースも含めて完全削除
docker compose down -v
```

### アプリケーションへアクセス

ブラウザで https://chat.localhost にアクセスしてください。

1. 初回は「新規登録」からアカウントを作成
2. ログイン後、ワークスペースを作成して利用開始

## 環境変数の設定

### 環境変数ファイル

バックエンドディレクトリの`.env.example`ファイルをコピーして`.env`ファイルを作成し、必要に応じて設定を変更してください。

```bash
cp backend/.env.example backend/.env
```

## データベース管理

### スキーマ管理

このプロジェクトでは [ent](https://entgo.io/) を使用してデータベーススキーマを管理しています。

```bash
# データベーススキーマをリセット（全テーブルを再作成）
docker compose exec backend go run cmd/reset/main.go

# 性能検証用に大量のメッセージを投入（初期データは起動時に自動で作られます）
docker compose exec backend go run cmd/seed/main.go -messages 1000
```

### スキーマの変更

スキーマを変更する場合は、以下の手順で行います：

1. `backend/ent/schema/` ディレクトリ内のスキーマファイルを編集
2. ent のコード生成を実行:
   ```bash
   docker compose exec backend go generate ./ent
   ```
3. アプリケーションを再起動すると、自動的にスキーマが適用されます

**注意:** ent はコードファーストのアプローチを採用しており、SQL マイグレーションファイルを使用しません。スキーマの変更は全て Go コードで管理されます。

### ER 図の生成と確認

このプロジェクトでは [entviz](https://github.com/hedwigz/entviz) を使用して ER 図を自動生成できます。

#### ER 図の更新手順

スキーマを変更した際は、以下のコマンドで ER 図を更新します：

```bash
# ER図を生成（entのコード生成と同時に実行されます）
docker compose exec backend go generate ./ent
```

#### ER 図の確認手順

生成された ER 図を確認するには、`backend/ent/schema-viz.html` をブラウザで開いてください：

```bash
# Macの場合
open backend/ent/schema-viz.html

# Windowsの場合
start backend/ent/schema-viz.html

# Linuxの場合
xdg-open backend/ent/schema-viz.html
```

## CI

プルリクエストに対して `.github/workflows/codecheck.yml` が以下を実行します。

| ジョブ | 内容 |
| --- | --- |
| frontend | typecheck / Oxlint / Oxfmt / knip / Vitest / ビルド |
| backend | `go build` と golangci-lint |
| proto | buf lint と format の検査、生成物が最新かを再生成して差分検証 |

依存関係の更新は Dependabot が週次でまとめて PR を作成し、`dependabot-auto-merge.yml` が自動マージします。

## デプロイ

Google Cloud の GKE に dev と prod を namespace で分けて載せ、GitHub Actions の「Deploy」（`.github/workflows/deploy.yml`）でデプロイします。Terraform は `infra/terraform/`、Kubernetes のマニフェストは `infra/k8s/` にあります。GKE の前に安く検証するための mini 構成（GCE の Spot VM 1 台に k3s、DB は Neon）もあります。構成と初回セットアップの手順は `docs/infrastructure.md` にまとめています。

```sh
# dev の backend だけ feat/foo に差し替える（空にした方はクラスタで動いているイメージを使う）
gh workflow run deploy.yml -R newt239/chat -f environment=dev -f backend_ref=feat/foo

# mini 構成（k3s の VM 1 台）の frontend を feat/bar にする
gh workflow run deploy.yml -R newt239/chat -f environment=mini -f frontend_ref=feat/bar

# dev で動いている版を prod に出す（prod は承認が必要）
gh workflow run deploy.yml -R newt239/chat -f environment=prod -f promote_from_dev=true
```
