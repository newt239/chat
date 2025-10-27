# Chat Application

Slack ライクなリアルタイムチャットアプリケーション。ワークスペース、チャンネル、ファイル添付機能を備えています。

## クイックスタート

```bash
# Docker Desktopを起動してから実行
docker-compose up -d --build

# ログを確認する場合
docker-compose logs -f
```

→ http://localhost:5173 にアクセス

これで PostgreSQL、バックエンド、フロントエンドの全てが起動します。

**初回起動時は自動的にシードデータが投入されます:**

- テスト用のワークスペース「Test Workspace」
- 4 つのテストアカウント（alice@example.com, bob@example.com, charlie@example.com, diana@example.com）
- パスワード: `password123`
- サンプルチャンネル（general, random, development, private-team）
- サンプルメッセージとリアクション

詳細なセットアップ手順は [ローカル環境のセットアップ](#ローカル環境のセットアップ) を参照してください。

## 技術スタック

### バックエンド

- Go 1.23+
- Echo (HTTP ルーター)
- WebSocket (gorilla/websocket)
- GORM (ORM)
- Atlas (宣言的スキーママイグレーション)
- PostgreSQL
- JWT 認証
- Wasabi S3 互換ストレージ
- クリーンアーキテクチャ実装

### フロントエンド

- React 19
- TypeScript
- Vite
- Mantine 8 (UI コンポーネント)
- Tailwind CSS
- TanStack Router
- TanStack Query
- PWA 対応 (vite-plugin-pwa)
- Vitest + Storybook

### インフラ

- Docker Compose
- Caddy (リバースプロキシ)
- VPS デプロイ対応

## プロジェクト構造

```
chat/
├── backend/          # Go backend
│   ├── cmd/
│   │   └── server/  # Main application entry point
│   ├── internal/
│   │   ├── domain/         # Domain entities & repository interfaces
│   │   ├── usecase/        # Business logic (to be implemented)
│   │   ├── interfaces/handler/
│   │   │   ├── http/       # HTTP handlers & routes
│   │   │   └── websocket/ # WebSocket hub & connections
│   │   └── infrastructure/
│   │       ├── auth/       # JWT & password hashing
│   │       ├── config/     # Configuration management
│   │       ├── database/         # GORM models & connection
│   │       ├── logger/     # Zap logger setup
│   │       ├── repository/   # Repository implementation
│   │       ├── storage/      # Wasabi S3 client (to be implemented)
│   │       ├── utils/        # Utility functions
│   │       └── logger/       # Zap logger setup
│   ├── schema/             # Atlas declarative schema (HCL)
│   └── atlas.hcl           # Atlas configuration
├── frontend/         # React frontend
│   ├── src/
│   │   ├── routes/   # TanStack Router routes
│   │   ├── features/ # Feature-based modules
│   │   ├── components/ # Reusable UI components
│   │   └── lib/      # API client, WS client, etc.
│   └── public/       # Static assets & PWA manifest
├── docker/           # Docker configurations
└── schema/           # Shared schema files

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

# 2. 環境変数ファイルの設定
cp .env.example .env
cp backend/.env.example backend/.env
# .envファイルを編集して必要に応じて設定を変更

# 3. Docker Composeで全て起動（初回は自動的にシードデータが投入されます）
docker-compose up -d --build

# 4. ログ確認（オプション）
docker-compose logs -f

# 起動完了後、http://localhost:5173 にアクセス
```

#### 停止方法

```bash
# コンテナを停止（データは保持）
docker-compose stop

# コンテナを削除（データは保持）
docker-compose down

# データも含めて完全削除
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

### 主要な環境変数

- `POSTGRES_USER`: PostgreSQL のユーザー名（デフォルト: postgres）
- `POSTGRES_PASSWORD`: PostgreSQL のパスワード（デフォルト: postgres）
- `POSTGRES_DB`: PostgreSQL のデータベース名（デフォルト: chat）
- `POSTGRES_HOST`: PostgreSQL のホスト名（デフォルト: db）
- `POSTGRES_PORT`: PostgreSQL のポート番号（デフォルト: 5432）
- `POSTGRES_URL`: PostgreSQL の接続 URL（上記の変数から自動生成）

## データベース管理

### シードデータの管理

```bash
# データベースをリセット
docker-compose exec backend go run cmd/reset/main.go

# 手動でシードデータを投入
docker-compose exec backend go run cmd/seed/main.go
```

### マイグレーション管理

```bash
# 例: Docker環境でカラムを追加するマイグレーションを生成
pnpm run migrate:generate docker add_email_column

# 例: Docker環境にマイグレーションを適用
docker-compose exec backend atlas migrate apply --env docker
```

**利用可能な環境:**

- `dev` - ローカル開発環境 (postgres://postgres:postgres@localhost:5432/chat)
- `docker` - Docker 環境 (postgres://postgres:postgres@db:5432/chat)

**マイグレーション生成の流れ:**

1. `schema/schema.hcl` ファイルを編集してスキーマを変更
2. `pnpm run migrate:generate [環境名] [マイグレーション名]` でマイグレーションファイルを生成
3. 生成されたマイグレーションファイルを確認・編集（必要に応じて）
4. `docker-compose exec backend atlas migrate apply --env [環境名]` でマイグレーションを適用

### データベースの状態確認

```bash
# PostgreSQLに接続
docker-compose exec postgres psql -U chat_user -d chat_db

# テーブル一覧
\dt

# ユーザー一覧
SELECT email, display_name FROM users;

# ワークスペース一覧
SELECT name, description FROM workspaces;

# チャンネル一覧
SELECT c.name, w.name as workspace_name FROM channels c
JOIN workspaces w ON c.workspace_id = w.id;
```

## 開発

### テストの実行

```bash
# バックエンド（Dockerコンテナ内で実行）
docker-compose exec backend go test ./...

# フロントエンド
cd frontend
pnpm test
```

### コード生成

```bash
# フロントエンド用にOpenAPI型定義を生成
cd frontend
pnpm run generate:api
```

## ライセンス

MIT
