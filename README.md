# Chat Application

Slack ライクなリアルタイムチャットアプリケーション。ワークスペース、チャンネル、ファイル添付機能を備えています。

## クイックスタート

```bash
# Docker Desktopを起動してから実行
docker-compose up -d

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
- Echo (HTTP ルーター) - Gin から移行完了
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
│   │   ├── interface/
│   │   │   ├── http/       # HTTP handlers & routes
│   │   │   └── ws/         # WebSocket hub & connections
│   │   └── infrastructure/
│   │       ├── auth/       # JWT & password hashing
│   │       ├── config/     # Configuration management
│   │       ├── db/         # GORM models & connection
│   │       ├── logger/     # Zap logger setup
│   │       └── storage/    # Wasabi S3 client (to be implemented)
│   ├── schema/       # Atlas declarative schema (HCL)
│   └── atlas.hcl     # Atlas configuration
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

## 実装状況

### 完了

- [x] モノレポ構造 (pnpm workspaces + Turbo)
- [x] バックエンド Clean Architecture スケルトン
- [x] ドメインエンティティ (User, Workspace, Channel, Message 等)
- [x] リポジトリインターフェース
- [x] JWT 認証インフラ
- [x] パスワードハッシュ化 (bcrypt)
- [x] 設定管理
- [x] ロギング (zap)
- [x] HTTP ミドルウェア (CORS, 認証, レート制限)
- [x] WebSocket hub & コネクション管理
- [x] 全エンティティの GORM モデル
- [x] OpenAPI 3.1 仕様
- [x] Atlas スキーマ定義 (PostgreSQL)
- [x] フロントエンド初期化
- [x] 認証 UI (ログイン・新規登録)
- [x] TanStack Query セットアップ
- [x] Tailwind CSS & Mantine UI

### 進行中

- [ ] GORM を使ったリポジトリ実装
- [ ] ユースケース層 (ビジネスロジック)
- [ ] 全エンドポイントの HTTP ハンドラー
- [ ] Wasabi S3 クライアント実装
- [ ] チャット UI コンポーネント
- [ ] WebSocket 統合

### 予定

- [ ] TanStack Router セットアップ
- [ ] PWA 機能
- [ ] テスト (Vitest)
- [ ] Storybook
- [ ] 本番デプロイ設定

## ローカル環境のセットアップ

### 起動方法

#### 必要な環境

- **Docker Desktop**

#### 手順

```bash
# 1. リポジトリのクローン
git clone <repository-url>
cd chat

# 2. Docker Composeで全て起動
docker-compose up -d

# 3. ログ確認（オプション）
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

## データベース管理

### シードデータの管理

```bash
# データベースをリセット（全データ削除）
cd backend
go run cmd/reset/main.go

# 手動でシードデータを投入
go run cmd/seed/main.go

# 強制的にシードデータを再投入（既存データを無視）
go run cmd/seed-manual/main.go
```

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
# バックエンド
cd backend
go test ./...

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
