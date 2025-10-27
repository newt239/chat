# Chat Application

Slack ライクなリアルタイムチャットアプリケーション。ワークスペース、チャンネル、ファイル添付機能を備えています。

## クイックスタート

```bash
# 1. Docker Desktopを起動
# 2. アプリケーションを起動
docker-compose up -d --build

# 3. シードデータを作成（初回のみ）
docker-compose exec backend go run cmd/seed/main.go
```

→ http://localhost:5173 にアクセス

### 利用可能なコマンド

```bash
# アプリケーションを起動
docker-compose up -d --build

# アプリケーションを停止
docker-compose down

# データベースをリセット
docker-compose exec backend go run cmd/reset/main.go

# シードデータを作成
docker-compose exec backend go run cmd/seed/main.go

# ログを表示
docker-compose logs -f

# コンテナの状態を確認
docker-compose ps
```

### テストアカウント

- **メールアドレス**: alice@example.com
- **パスワード**: password123

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

# 2. アプリケーションを起動
docker-compose up -d --build

# 3. シードデータを作成（初回のみ）
docker-compose exec backend go run cmd/seed/main.go

# 4. 起動完了後、http://localhost:5173 にアクセス
```

#### 停止方法

```bash
# アプリケーションを停止
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

# シードデータを作成
docker-compose exec backend go run cmd/seed/main.go
```

### マイグレーション管理

```bash
# 例: カラムを追加するマイグレーションを生成（Dockerコンテナ内で実行）
pnpm run migrate:generate add_email_column

# 例: マイグレーションを適用（Dockerコンテナ内で実行）
docker-compose exec backend atlas migrate apply --env docker
```

**マイグレーション生成の流れ:**

1. `backend/schema/schema.hcl` ファイルを編集してスキーマを変更
2. `pnpm run migrate:generate [マイグレーション名]` でマイグレーションファイルを生成（Docker コンテナ内で実行）
3. 生成されたマイグレーションファイルを確認・編集（必要に応じて）
4. `docker-compose exec backend atlas migrate apply --env docker` でマイグレーションを適用

**注意:** マイグレーションは Docker 環境でのみサポートされています。

### データベースの状態確認

```bash
# PostgreSQLに接続（Dockerコンテナ内で実行）
docker-compose exec db psql -U postgres -d chat

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
# バックエンド
docker-compose exec backend go test ./...

# フロントエンド
docker-compose exec frontend pnpm test
```

### コード生成

```bash
# フロントエンド用にOpenAPI型定義を生成
docker-compose exec frontend pnpm run generate:api
```
