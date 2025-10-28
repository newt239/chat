# Chat Application

Slack ライクなリアルタイムチャットアプリケーション。ワークスペース、チャンネル、ファイル添付機能を備えています。

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
- ent (ORM)
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
│   │   ├── routes/   # TanStack Router routes
│   │   ├── features/ # Feature-based modules
│   │   ├── components/ # Reusable UI components
│   │   └── lib/      # API client, WS client, etc.
│   └── public/       # Static assets & PWA manifest
└── docker/           # Docker configurations

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

### 主要な環境変数

- `POSTGRES_USER`: PostgreSQL のユーザー名（デフォルト: postgres）
- `POSTGRES_PASSWORD`: PostgreSQL のパスワード（デフォルト: postgres）
- `POSTGRES_DB`: PostgreSQL のデータベース名（デフォルト: chat）
- `POSTGRES_HOST`: PostgreSQL のホスト名（デフォルト: db）
- `POSTGRES_PORT`: PostgreSQL のポート番号（デフォルト: 5432）
- `POSTGRES_URL`: PostgreSQL の接続 URL（上記の変数から自動生成）

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
# entのコード生成（スキーマ変更時）
docker-compose exec backend go generate ./ent

# フロントエンド用にOpenAPI型定義を生成
docker-compose exec frontend pnpm run generate:api
```
