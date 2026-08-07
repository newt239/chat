# Chat Application

## クイックスタート

```bash
# 1. Docker Desktopを起動
# 2. アプリケーションを起動（スキーマのリセットとシードデータは自動実行されます）
docker-compose up -d --build
```

→ http://localhost:5173 にアクセス

### 利用可能なコマンド

```bash
# アプリケーションを起動
docker-compose up -d --build

# アプリケーションを停止
docker-compose down

# データベーススキーマをリセット
docker-compose exec backend go run cmd/reset/main.go

# シードデータを投入（通常は自動実行されます）
docker-compose exec backend go run cmd/seed/main.go

# バックエンドコードのリント
docker-compose exec backend golangci-lint run

# ログを表示
docker-compose logs -f

# コンテナの状態を確認
docker-compose ps
```

### フロントエンドの開発コマンド

lint / format / test の設定は `frontend/vite.config.ts` に集約されています（Vite+ による統合）。

```bash
# 型チェック・lint・format・ファイル名チェック・未使用コード検出・テストを一括実行
pnpm --filter chat-frontend run codecheck

# 個別に実行
pnpm --filter chat-frontend run typecheck
pnpm --filter chat-frontend run lint:fix
pnpm --filter chat-frontend run format:fix
pnpm --filter chat-frontend run test

# OpenAPI スキーマを変更したとき（リポジトリルートで実行）
pnpm run openapi:bundle && pnpm run generate:api
```

コミット時は lefthook の pre-commit フックが lint・format・ファイル名チェックを実行します。ホスト側に `pnpm install` 済みであることが前提のため、Docker のみで開発している場合は `LEFTHOOK=0 git commit` で回避できます。

### テストアカウント

- **メールアドレス**: alice@example.com
- **パスワード**: password123

詳細なセットアップ手順は [ローカル環境のセットアップ](#ローカル環境のセットアップ) を参照してください。

## 技術スタック

### バックエンド

- Go 1.24
- Echo
- WebSocket (gorilla/websocket)
- ent (ORM)
- PostgreSQL 18
- Wasabi

### フロントエンド

- React 19
- TypeScript 7
- Vite+ (`vite-plus`) — Vite 8 / Vitest / Oxlint / Oxfmt を統合したツールチェーン
- Mantine 8
- Tailwind CSS 4
- React Router 8 (Data モード)
- TanStack Query
- PWA (vite-plugin-pwa)

### 開発ツール

- pnpm 11 (workspace) + Turborepo
- lefthook (pre-commit フック)
- ls-lint (ファイル名規約) / knip (未使用コード検出)
- OpenAPI (Redocly でバンドル、oapi-codegen と openapi-typescript でコード生成)

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
│   │   └── seed/    # Seed data tool
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
│   │       ├── repository/ # Repository implementation
│   │       ├── storage/    # Wasabi S3 client
│   │       └── utils/      # Utility functions
│   └── ent/              # ent schema definitions & generated code
├── frontend/         # React frontend
│   ├── src/
│   │   ├── routes/   # React Router のルート定義（routeTree.ts）とページコンポーネント
│   │   ├── features/ # Feature-based modules
│   │   ├── providers/ # Jotai ストア・TanStack Query・WebSocket の Provider
│   │   └── lib/      # API client, WS client, paths, routeParams など
│   ├── tests/        # Vitest のセットアップ
│   └── public/       # Static assets（PWA アイコンの元になる logo.svg）
├── openapi/          # OpenAPI スキーマ（分割定義と bundled.yaml）
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
docker-compose up -d --build

# 3. 起動完了後、http://localhost:5173 にアクセス
```

#### 停止方法

```bash
# アプリケーションを停止
docker-compose down

# データベースも含めて完全削除
docker-compose down -v
```

### アプリケーションへアクセス

ブラウザで http://localhost:5173 にアクセスしてください。

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
docker-compose exec backend go run cmd/reset/main.go

# シードデータを投入（通常は自動実行されます）
docker-compose exec backend go run cmd/seed/main.go
```

### スキーマの変更

スキーマを変更する場合は、以下の手順で行います：

1. `backend/ent/schema/` ディレクトリ内のスキーマファイルを編集
2. ent のコード生成を実行:
   ```bash
   docker-compose exec backend go generate ./ent
   ```
3. アプリケーションを再起動すると、自動的にスキーマが適用されます

**注意:** ent はコードファーストのアプローチを採用しており、SQL マイグレーションファイルを使用しません。スキーマの変更は全て Go コードで管理されます。

### ER 図の生成と確認

このプロジェクトでは [entviz](https://github.com/hedwigz/entviz) を使用して ER 図を自動生成できます。

#### ER 図の更新手順

スキーマを変更した際は、以下のコマンドで ER 図を更新します：

```bash
# ER図を生成（entのコード生成と同時に実行されます）
docker-compose exec backend go generate ./ent
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
| frontend | typecheck / Oxlint / Oxfmt / ls-lint / knip / Vitest / ビルド |
| backend | `go build` と golangci-lint |
| openapi | `openapi/bundled.yaml` と `frontend/src/lib/api/schema.ts` が最新かを再生成して差分検証 |

依存関係の更新は Dependabot が週次でまとめて PR を作成し、`dependabot-auto-merge.yml` が自動マージします。

## デプロイ

現在、自動デプロイの基盤（GitHub Actions のデプロイワークフローと Ansible の構成）はリポジトリから削除されています。本番環境向けの成果物は次の方法で用意できます。

```bash
# フロントエンド: nginx で静的配信するイメージをビルド
docker build -f frontend/Dockerfile -t chat-frontend .

# バックエンド: server / migrate / reset / seed のバイナリを含むイメージをビルド
docker build -f backend/Dockerfile -t chat-backend ./backend
```

バックエンドはリクエストバリデーションのために実行時に `/app/openapi/openapi.yaml` を読み込みます。本番イメージには `openapi/` が含まれていないため、コンテナ実行時にマウントするか Dockerfile を調整してください。
