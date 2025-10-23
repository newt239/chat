# Chat Application

Slack ライクなリアルタイムチャットアプリケーション。ワークスペース、チャンネル、ファイル添付機能を備えています。

## クイックスタート

**Docker Compose で全て起動（推奨）:**

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

**ローカル開発（ホットリロード重視）:**

```bash
# 1. データベースのみDockerで起動
docker-compose up -d postgres

# 2. アプリケーションをローカルで起動
pnpm run dev
```

詳細なセットアップ手順は [ローカル環境のセットアップ](#ローカル環境のセットアップ) を参照してください。

## 技術スタック

### バックエンド

- Go 1.22+
- Gin (HTTP ルーター)
- WebSocket (gorilla/websocket)
- GORM + Gen (ORM & コード生成)
- Atlas (宣言的スキーママイグレーション)
- PostgreSQL
- JWT 認証
- Wasabi S3 互換ストレージ

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

### 起動方法の選択

以下の 2 つの起動方法があります：

#### 方法 A: Docker Compose で全て起動（推奨）

全てのコンポーネント（PostgreSQL、バックエンド、フロントエンド）を Docker で起動します。

- **メリット**: 環境構築が簡単、依存関係の問題がない
- **デメリット**: コード変更時の反映が若干遅い

#### 方法 B: ローカル開発環境

データベースのみ Docker で起動し、アプリケーションはローカルで起動します。

- **メリット**: コード変更が即座に反映（高速ホットリロード）
- **デメリット**: Go、Node.js、pnpm のインストールが必要

---

### 方法 A: Docker Compose で全て起動

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

---

### 方法 B: ローカル開発環境

#### 必要な環境

- **Go** 1.22 以上
- **Node.js** 20 以上
- **pnpm** 10 以上
- **Docker Desktop** (PostgreSQL 用)

#### 1. リポジトリのクローン

```bash
git clone <repository-url>
cd chat
```

#### 2. 依存関係のインストール

```bash
pnpm install
```

#### 3. データベースのみ Docker で起動

```bash
# PostgreSQLのみ起動
docker-compose up -d postgres

# 起動確認
docker-compose ps
```

#### 4. 開発サーバーの起動

プロジェクトルートで実行：

```bash
pnpm run dev
```

これで以下が起動します：

- **フロントエンド**: http://localhost:5173
- **バックエンド API**: http://localhost:8080
- **データベース**: すでに起動済み（docker-compose）

#### 個別に起動する場合

**ターミナル 1 - バックエンド:**

```bash
cd backend
go run ./cmd/server
```

**ターミナル 2 - フロントエンド:**

```bash
cd frontend
pnpm dev
```

---

### アプリケーションへアクセス

ブラウザで http://localhost:5173 にアクセスしてください。

1. 初回は「新規登録」からアカウントを作成
2. ログイン後、ワークスペースを作成して利用開始

---

### トラブルシューティング

#### Docker Compose で起動時にエラーが出る

```bash
# コンテナの状態確認
docker-compose ps

# 各サービスのログ確認
docker-compose logs postgres
docker-compose logs backend
docker-compose logs frontend

# 再ビルドして起動
docker-compose up -d --build
```

#### フロントエンドで "Registration failed" が表示される

→ バックエンドサーバーが起動していません。

**Docker Compose 使用時:**

```bash
docker-compose logs backend
```

**ローカル開発時:**
バックエンドサーバーを起動してください。

#### バックエンドで "failed to connect to database" エラー

→ PostgreSQL が起動していません。

```bash
# PostgreSQLコンテナの状態確認
docker-compose ps postgres

# ログ確認
docker-compose logs postgres

# 再起動
docker-compose restart postgres
```

#### ポート 5173、8080、5432 が既に使用されている

→ 他のプロセスまたはコンテナがポートを使用しています：

```bash
# 使用中のプロセスを確認
lsof -i :5173
lsof -i :8080
lsof -i :5432

# 既存のコンテナを停止
docker-compose down
```

### 環境変数の設定（オプション）

#### Docker Compose 使用時

環境変数は [docker-compose.yml](docker-compose.yml) で設定済みです。変更する場合は直接編集してください。

#### ローカル開発時

デフォルト設定で動作しますが、カスタマイズする場合：

**backend/.env**:

```env
PORT=8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/chat?sslmode=disable
JWT_SECRET=your-secret-key
CORS_ALLOWED_ORIGINS=http://localhost:5173
```

**frontend/.env.local**:

```env
VITE_API_BASE_URL=http://localhost:8080
```

---

### 便利なコマンド

```bash
# 全てのコンテナを起動
docker-compose up -d

# 特定のサービスのみ起動
docker-compose up -d postgres
docker-compose up -d backend
docker-compose up -d frontend

# ログをリアルタイムで確認
docker-compose logs -f
docker-compose logs -f backend

# コンテナの状態確認
docker-compose ps

# コンテナを停止
docker-compose stop

# コンテナを削除（データは保持）
docker-compose down

# データも含めて完全削除
docker-compose down -v

# 再ビルドして起動
docker-compose up -d --build
```

## API ドキュメント

API は OpenAPI 3.1 で文書化されています。詳細は [backend/internal/openapi/openapi.yaml](backend/internal/openapi/openapi.yaml) を参照してください。

### 主要なエンドポイント

- `POST /api/auth/register` - ユーザー新規登録
- `POST /api/auth/login` - ログイン
- `POST /api/auth/refresh` - アクセストークン更新
- `GET /api/workspaces` - ワークスペース一覧
- `GET /api/workspaces/{id}/channels` - チャンネル一覧
- `GET /api/channels/{id}/messages` - メッセージ一覧
- `POST /api/channels/{id}/messages` - メッセージ送信
- `GET /ws?workspaceId={id}` - WebSocket 接続

## データベーススキーマ

データベーススキーマは Atlas の宣言的 HCL ファイルで管理されています。完全なスキーマ定義は [backend/schema/schema.hcl](backend/schema/schema.hcl) を参照してください。

### 主要なテーブル

- `users` - ユーザーアカウント
- `sessions` - JWT リフレッシュトークン
- `workspaces` - ワークスペースコンテナ
- `workspace_members` - ワークスペースメンバーシップとロール
- `channels` - コミュニケーションチャンネル
- `channel_members` - プライベートチャンネルメンバーシップ
- `messages` - チャットメッセージ（スレッド対応）
- `message_reactions` - メッセージリアクション（絵文字）
- `channel_read_states` - 未読メッセージ追跡
- `attachments` - ファイル添付メタデータ

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
