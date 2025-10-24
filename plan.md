# Slack ライク・コミュニケーションアプリ 実装計画

**最終更新: 2025-10-23**

**プロジェクト進捗: 約 60% 完了 (MVP 基準)**

- Backend: 約 80% (MVP 機能完了、添付ファイル・WebSocket イベント処理が未完)
- Frontend: 約 50% (基本 UI 完了、リアクション・Markdown・メンション・DM・検索など多数未実装)
- DevOps: 約 40% (開発環境完了、本番環境未完)

## 技術スタック

- フロントエンド: React 19, TypeScript, Vite, Mantine 8, Tailwind CSS, TanStack Router, TanStack Query, Vitest, PWA（vite-plugin-pwa）
- バックエンド: Go 1.22+, Gin, Clean Architecture, WebSocket, GORM, Atlas（宣言的マイグレーション）
- データベース: PostgreSQL 16
- オブジェクトストレージ: Wasabi（S3 互換, aws-sdk-go-v2）
- デプロイ: Docker Compose（開発環境完了）, リバースプロキシ（Caddy 予定）, VPS 運用予定

## ディレクトリ構成（モノレポ）

- `docker/`（compose, Caddy/Nginx 設定, Dockerfiles）
- `backend/`
- `cmd/server/main.go`
- `internal/`
  - `domain/`（Entity/ValueObject, Repository IF, Domain Service）
  - `usecase/`（Input/Output DTO, Interactor, Tx 境界）
  - `interface/`
  - `http/`（Gin ルータ/ハンドラ/ミドルウェア）
  - `ws/`（ハブ, コネクション, イベント仲介）
  - `infrastructure/`
  - `db/`（GORM 初期化, Gen 生成 `gen/Query`）
  - `auth/`（local, oidc）
  - `storage/wasabi/`（S3 互換クライアント/プリサイン）
  - `logger/`, `config/`, `observability/`
  - `openapi/openapi.yaml`（OpenAPI 3.1）
- `schema/`（Atlas declarative schema: HCL/SQL）
- `atlas.hcl`
- `air.toml`
- `frontend/`
- `src/`
  - `routes/`（TanStack Router: `/login`, `/app/workspaces/:wsId/channels/:chId`）
  - `features/`（auth, workspace, channel, message, attachment, unread）
  - `components/`（再利用 UI）
  - `lib/`（api, wsClient, queryClient, store）
  - `styles/`（tailwind.css）
- `public/manifest.webmanifest`
- 設定: `vite.config.ts`, `eslint`, `prettier`, `tailwind`, `postcss`, `vitest`, `storybook`

## データモデル（PostgreSQL ｜ Atlas 管理）

- 拡張: `pgcrypto`, `btree_gin`
- 主なテーブル
- `users`（email[uniq], password_hash, display_name, avatar_url, timestamps）
- `sessions`（user_id, refresh_token_hash, expires_at, revoked_at）
- `workspaces` / `workspace_members(role: owner|admin|member|guest)`
- `channels`（workspace_id, name[uniq in ws], is_private, created_by）
- `channel_members`（private 用メンバーシップ）
- `messages`（channel_id, user_id, parent_id[thread], body, created/edited/deleted_at）
- `message_reactions`（message_id, user_id, emoji, created_at）
- `channel_read_states`（channel_id, user_id, last_read_at）
- `attachments`（message_id, file_name, mime_type, size_bytes, storage_key）
- `oidc_accounts`（provider, subject, user_id, linked_at）
- インデックス例
- `messages(channel_id, created_at desc)` / `messages(parent_id, created_at)`
- `channels(workspace_id)` / `channels(workspace_id, is_private)`
- `channel_read_states(user_id, last_read_at desc)`
- 運用: `schema/` を真実源として `atlas migrate diff/apply/lint/validate` を利用

## 認証/認可

- ローカル認証（メール/パスワード + bcrypt）
- JWT（アクセス: 短寿命 ~15 分）+ リフレッシュ（httpOnly Secure Cookie もしくはボディ返却）
- `AuthProvider` 抽象により将来 OIDC 追加（独自プロバイダー連携）
- RBAC: workspace/channel 単位の権限チェック（ミドルウェア）

## API（OpenAPI 3.1）

- ファイル: `backend/internal/openapi/openapi.yaml`
- クライアント: `openapi-typescript` で型生成、`openapi-fetch` で RPC 的呼び出し
- 代表エンドポイント
- Auth: `POST /api/auth/register`, `POST /api/auth/login`, `POST /api/auth/refresh`, `POST /api/auth/logout`
- Workspaces/Channels: `GET/POST /api/workspaces`, `GET/POST /api/workspaces/{id}/channels`
- Messages: `GET/POST /api/channels/{channelId}/messages`（`since/until/limit` サポート）
- Reads: `GET /api/channels/{channelId}/unread_count`, `POST /api/channels/{channelId}/reads`（`lastReadAt` 更新）
- Attachments: `POST /api/attachments/presign`, `GET /api/attachments/{id}`, `GET /api/attachments/{id}/download`
- 健康/監視: `GET /healthz`, `GET /metrics`

## WebSocket

- エンドポイント: `GET /ws?workspaceId=...`（JWT 検証, 参加権限チェック）
- イベント
- クライアント → サーバ: `join_channel`, `leave_channel`, `post_message`, `typing`, `update_read_state`
- サーバ → クライアント: `new_message`, `unread_count`, `ack`, `error`
- スケール: 単一ノードはメモリハブ、将来 Redis Pub/Sub アダプタで水平分散

## ストレージ（Wasabi）

- `aws-sdk-go-v2` を S3 互換設定で利用（endpoint/region/credentials）
- フロー: `presign` 取得 → クライアント直アップロード → メタを `attachments` 登録
- ダウンロードも presign を発行

## フロントエンド

- ルーティング: TanStack Router（認証ゲート, AppShell レイアウト）
- データ取得: TanStack Query（OpenAPI クライアント, キャッシュキー=operationId+params）
- UI: Mantine 8 + Tailwind 併用（アクセシビリティ優先, テーマ統一）
- リアルタイム: WebSocket 受信で Query 部分更新（`queryClient.setQueryData`）
- テスト: Vitest + Testing Library
- Storybook: Mantine/Tailwind を読み込むプレビュー設定でコンポーネントカタログ化

## PWA

- `vite-plugin-pwa`（`registerType: 'autoUpdate'`）
- `public/manifest.webmanifest`（name, short_name, icons[512/192/maskable], theme_color）
- Workbox キャッシュ戦略
- API GET: Stale-While-Revalidate
- 静的アセット: Cache First
- POST 等の送信は IndexedDB の送信キュー＋再接続時フラッシュ（必要に応じて Background Sync）
- IndexedDB に最近メッセージを保持して簡易オフライン閲覧

## モバイル最適化

- レイアウト: Mantine `AppShell` + Tailwind ブレークポイント（`sm`, `md`）
- ナビ: モバイルではタブ/ドロワー切替
- 入力: iOS safe-area/`100dvh` 対応、送信バー固定、ファイルピッカー
- パフォーマンス: `@tanstack/react-virtual` でメッセージ仮想リスト、画像の遅延読込

## セキュリティ/可観測性

- セキュリティ: bcrypt cost, JWT 鍵管理、レート制限（ログイン）、CORS、ヘッダ強化、入力バリデーション
- 可観測性: zap ログ, OpenTelemetry（トレーシング/メトリクス）, pprof

## デプロイ（VPS, Docker）

- `docker-compose.yml`: `db`（Postgres）, `backend`, `frontend`, `caddy`
- 起動時: `atlas migrate apply` を backend の entrypoint に組込
- 環境変数: `DATABASE_URL`, `JWT_SECRET`, `WASABI_*`, `CORS_ALLOWED_ORIGINS`
- TLS/圧縮/HTTP2: Caddy で終端

## 現在の実装状況（マイルストーン）

### ✅ 完了済み

1. **スケルトン/起動** - CA 構成、GORM 初期化、全 Domain 層定義
2. **Atlas 導入** - 宣言的スキーマ(schema.hcl)、全テーブル定義
3. **認証/セッション** - Repository 実装、Auth UseCase/Handler 実装、JWT/Refresh
4. **Workspace/Channel** - Repository + UseCase/Handler (CRUD + メンバー管理)
5. **Message** - Repository + UseCase/Handler (取得・投稿・スレッド対応)
6. **未読管理** - Repository + ReadState API (既読更新・未読数取得)
7. **フロント基盤** - Router, Query, Auth, Workspace, Channel, Message UI
8. **開発環境** - Docker Compose (Postgres + Backend + Frontend)
9. **テスト基盤** - Vitest 設定、27 テスト実装（100%パス）

### 🚧 進行中・未完了

10. **WebSocket** - Hub/Connection 完了、イベントハンドラ未実装
11. **添付ファイル** - バックエンド Repository 完了、UseCase/S3 統合未完
12. **フロント統合** - WebSocket と Query 連携未完、未読バッジ UI 未完
13. **PWA** - manifest 完了、IndexedDB/オフライン機能未完
14. **本番デプロイ** - Dockerfile.prod 未完、Caddy 設定未完

## 実装状況詳細

### Backend 実装進捗: 約 80%

#### ✅ 完全実装 (100%)

- **Domain 層**: 全 7 エンティティ & Repository IF (User, Session, Workspace, Channel, Message, ReadState, Attachment)
- **Infrastructure 層**:
  - Config, Logger (Zap), Auth Services (JWT, Password), DB (GORM), 全 7 Repository 実装
  - 約 700 行のインフラコード
- **DB Schema**: Atlas declarative schema - 全 10+テーブル定義、インデックス、制約
- **HTTP Handlers**:
  - Auth (Register/Login/Refresh/Logout) - 4 エンドポイント
  - Workspace CRUD + メンバー管理 - 8 エンドポイント
  - Channel 一覧/作成 - 2 エンドポイント
  - Message 一覧/投稿 - 2 エンドポイント
  - ReadState 未読取得/更新 - 2 エンドポイント
  - 合計 18 エンドポイント実装済み (約 1,250 行)
- **UseCase 層**:
  - Auth (216 行), Workspace (379 行), Channel (124 行), Message (154 行), ReadState (100 行)
  - 合計約 1,200 行のビジネスロジック

#### 🚧 部分実装 (20-70%)

- **WebSocket (60%)**:

  - ✅ Hub 実装 (Register/Unregister/Broadcast)
  - ✅ Connection 管理 (ReadPump/WritePump, 140 行)
  - ✅ main.go でのエンドポイント登録 (JWT 検証)
  - ❌ イベントハンドラ未実装 (join_channel, leave_channel, post_message, typing, update_read_state)
  - **課題**: connection.go:76-77 でプレースホルダーのみ

- **Attachment (10%)**:
  - ✅ Domain Entity & Repository 完了
  - ❌ UseCase 未実装
  - ❌ Handler は 501 Not Implemented stub
  - ❌ S3/Wasabi 統合なし (aws-sdk-go-v2 未使用)

#### ❌ 未実装 (0%)

- **Observability**: OpenTelemetry, Metrics, pprof
- **Backend Tests**: Go テストファイルなし
- **OIDC**: 認証プロバイダー抽象化のみ
- **Reactions API**: Schema 完備、UseCase/Handler 未実装
- **Mentions API**: Schema 未実装、機能設計未着手
- **DM 機能**: Channel.isDM フラグ追加必要、API 未実装
- **Search API**: PostgreSQL FTS 未実装

### Frontend 実装進捗: 約 75%

#### ✅ 完全実装 (100%)

- **ビルド環境**: Vite, TypeScript, ESLint, Prettier, Tailwind, Mantine 8
- **ルーティング**: TanStack Router - 7 ルート (Login, Register, App, Workspace, Channel)
- **認証システム**:
  - Login/Register フォーム
  - Auth hooks (useAuth, useAuthGuard)
  - Zustand store (localStorage persist)
  - JWT refresh 機能
- **API クライアント**:
  - openapi-typescript 型生成
  - openapi-fetch クライアント
  - 自動認証リフレッシュ
- **Workspace 機能**:
  - 一覧/作成 UI
  - useWorkspace hooks
  - WorkspaceSelection component
- **Channel 機能 (90%)**:
  - 一覧/作成 UI
  - useChannel hooks
  - ChannelList component
  - **未完**: 詳細表示、設定 UI
- **Message 機能 (80%)**:
  - メッセージ表示 (MessagePanel)
  - 送信フォーム
  - useMessage hooks
  - 自動スクロール
  - **未完**: スレッド UI、仮想スクロール、編集/削除、リアクション
- **WebSocket クライアント (70%)**:
  - 接続/切断管理
  - 再接続ロジック (exponential backoff)
  - イベント送受信
  - **未完**: TanStack Query 統合、join_channel イベント送信
- **テスト**:
  - Vitest 設定
  - 8 ファイル、27 テスト (100% pass)
  - Auth/Workspace/Layout/Header コンポーネントカバー済み
  - **課題**: Header.test.tsx に 2 つの ESLint エラー (unused imports)

#### 🚧 部分実装 (30%)

- **PWA (30%)**:
  - ✅ vite-plugin-pwa 設定
  - ✅ manifest.webmanifest
  - ❌ Service Worker カスタマイズ

#### ❌ 未実装 (0%)

- **Storybook**: 設定ファイル、ストーリー作成
- **Attachment UI**: アップロード/ダウンロード/プレビュー
- **Unread UI**: バッジ、未読カウント表示
- **Virtual Scrolling**: メッセージリスト最適化
- **Message Threads**: スレッド表示 UI
- **Typing Indicators**: 入力中表示
- **Message Reactions**: リアクション選択/表示 UI
- **Markdown Support**: メッセージ Markdown 表示/編集
- **Mentions**: @ユーザーメンション機能
- **Direct Messages**: 1 対 1 DM 機能
- **Channel Search**: チャンネル名検索/フィルタ
- **Message Search**: メッセージ全文検索
- **User Profile**: プロフィール表示/編集
- **Channel Settings**: チャンネル設定/権限管理
- **Member List**: メンバー一覧/オンライン状態
- **Notification Settings**: 通知設定 UI
- **Theme Support**: ダークモード切り替え

### DevOps 実装進捗: 約 40%

#### ✅ 完全実装 (100%)

- **Docker Compose 開発環境**:
  - PostgreSQL 16 Alpine
  - Backend service (Dockerfile.dev)
  - Frontend service (Dockerfile.dev)
  - ネットワーク分離、ボリューム管理
  - ヘルスチェック

#### ❌ 未実装 (0%)

- **本番デプロイ**:
  - Dockerfile.prod (backend/frontend)
  - Docker Compose production.yml
  - Caddy/Nginx 設定
  - TLS 証明書管理
- **CI/CD**: GitHub Actions ワークフロー
- **監視**: ログ集約、メトリクス収集、アラート

---

## 優先タスク (MVP 向け)

### 🔴 Critical (MVP 必須)

#### Backend

1. **WebSocket イベントハンドラ実装** (優先度: 最高)

   - 実装内容:
     - `join_channel`: チャンネル参加通知
     - `post_message`: リアルタイムメッセージ配信
     - `update_read_state`: 未読状態同期
     - エラーハンドリング & ack 応答
   - 影響範囲: `backend/internal/interface/ws/connection.go`
   - 前提: Message/ReadState UseCase 既存のため依存少ない

2. **Attachment UseCase & Handler 実装** (優先度: 高)

   - 実装内容:
     - Presign URL 生成 UseCase
     - Attachment metadata 登録/取得
     - Download presign URL 生成
   - 影響範囲:
     - `backend/internal/usecase/attachment/`
     - `backend/internal/interface/http/handler/attachment_handler.go`
   - 依存: S3 クライアント統合 (次項)

3. **Wasabi S3 統合** (優先度: 高)
   - 実装内容:
     - aws-sdk-go-v2 初期化
     - S3 Presigner 設定
     - エンドポイント/リージョン/認証設定
   - 影響範囲:
     - `backend/internal/infrastructure/storage/wasabi/client.go`
     - `backend/cmd/server/main.go` (DI)
   - 環境変数: `WASABI_ENDPOINT`, `WASABI_REGION`, `WASABI_ACCESS_KEY`, `WASABI_SECRET_KEY`, `WASABI_BUCKET`

#### Frontend

4. **WebSocket & Query 統合** (優先度: 最高)

   - 実装内容:
     - `new_message`イベント受信 → queryClient.setQueryData
     - `unread_count`イベント受信 → 未読カウント更新
     - チャンネル参加時に`join_channel`イベント送信
     - 楽観的 UI 更新 (メッセージ送信時)
   - 影響範囲:
     - `frontend/src/lib/ws/client.ts`
     - `frontend/src/features/message/hooks/useMessage.ts`
   - 技術的課題: queryKey 構造と WebSocket イベント対応

5. **未読バッジ UI 実装** (優先度: 中)

   - 実装内容:
     - チャンネルリストに未読カウント表示
     - 未読があるチャンネルをハイライト
     - メッセージ閲覧時に既読 API 呼び出し
   - 影響範囲:
     - `frontend/src/features/channel/components/ChannelList.tsx`
     - `frontend/src/features/message/components/MessagePanel.tsx`
   - 依存: ReadState API (既存)

6. **ESLint エラー修正** (優先度: 最高、工数小)
   - 実装内容: `Header.test.tsx` から未使用 import 削除 (waitFor, userEvent)
   - 影響範囲: `frontend/src/components/layout/Header.test.tsx`
   - 工数: 5 分

#### DevOps

7. **本番 Docker 環境構築** (優先度: 中)
   - 実装内容:
     - Backend Dockerfile.prod (multi-stage build)
     - Frontend Dockerfile.prod (nginx serve)
     - docker-compose.prod.yml
     - Caddyfile (TLS, リバースプロキシ, 圧縮)
   - 影響範囲: `docker/`
   - 環境変数管理: .env.production

### 🟡 Medium (MVP 推奨)

8. **Attachment UI 実装** (優先度: 中)

   - 前提: Backend S3 統合完了後
   - 実装内容: ファイルピッカー, アップロード進捗, プレビュー, ダウンロード

9. **メッセージ仮想スクロール** (優先度: 中)

   - ライブラリ: `@tanstack/react-virtual`
   - パフォーマンス改善: 1000+ メッセージ対応

10. **Message Thread UI** (優先度: 中)

    - スレッド表示/返信 UI
    - parent_id 活用 (Backend 対応済み)

11. **Typing Indicators** (優先度: 中)

    - "○○ が入力中..." UI
    - WebSocket typing イベント連携

12. **Backend テスト整備** (優先度: 中)
    - UseCase 単体テスト
    - Repository 統合テスト (testcontainers)
    - テストカバレッジ目標: 60%+

### 🟢 Low (Post-MVP 機能拡張)

13. **Message Reactions** (優先度: 低)

    - リアクション追加/削除 API (**Schema 完備: message_reactions table**)
    - リアクション選択 UI (絵文字ピッカー)
    - WebSocket 同期

14. **Markdown Support** (優先度: 低)

    - メッセージ Markdown 表示 (react-markdown)
    - Markdown 編集プレビュー
    - コードブロックシンタックスハイライト
    - **Backend 変更不要**: body フィールドそのまま使用

15. **Mentions 機能** (優先度: 低)

    - Backend: Mentions テーブル設計・実装
    - @ユーザーメンション入力 (autocomplete)
    - メンション通知 API
    - メンション一覧表示

16. **Direct Messages** (優先度: 低)

    - Backend: Channel.isDM フラグ追加
    - 1 対 1 DM 用チャンネル作成 API
    - DM 一覧 UI
    - DM 専用通知

17. **検索機能** (優先度: 低)

    - チャンネル名検索/フィルタ
    - メッセージ全文検索 API (PostgreSQL FTS)
    - 検索 UI (モーダル, Ctrl+K)

18. **User Profile** (優先度: 低)

    - プロフィール表示/編集 UI
    - アバター画像アップロード
    - ステータスメッセージ

19. **Channel Settings** (優先度: 低)

    - チャンネル設定画面
    - 権限管理 (owner/admin/member)
    - チャンネル削除/アーカイブ

20. **Member List & Presence** (優先度: 低)

    - メンバー一覧 UI
    - オンライン状態表示
    - WebSocket presence イベント

21. **Notification Settings** (優先度: 低)

    - 通知設定 UI
    - チャンネル別通知 ON/OFF
    - メンション専用通知

22. **Theme Support** (優先度: 低)

    - ダークモード実装
    - Mantine ColorSchemeProvider 統合
    - localStorage 保存

23. **Storybook** (優先度: 低)

    - .storybook 設定
    - Mantine/Tailwind 統合
    - 主要コンポーネントのストーリー作成

24. **Observability 強化** (優先度: 低)

    - OpenTelemetry 統合
    - Prometheus metrics
    - pprof 有効化

25. **OIDC 認証** (優先度: 低)
    - Google/GitHub OAuth
    - AuthProvider 抽象化活用

---

## 既知の技術的課題

### Backend

1. **main.go:28** - CORS origin validation TODO (現在は全許可)
2. **Attachment handlers** - 501 Not Implemented
3. **WebSocket** - イベント処理プレースホルダー (connection.go:76-77)
4. **ログ統合** - Zap ロガー定義済みだが Handler 層で未使用
5. **レート制限** - Middleware あるが適用不十分
6. **エラーレスポンス** - 統一されたエラー構造なし

### Frontend

1. **Header.test.tsx:3,4** - ESLint unused imports エラー
2. **WebSocket 再接続** - 最大 5 回で停止、手動再接続 UI なし
3. **型安全性** - 一部 inferred だが明示的型推奨箇所あり
4. **アクセシビリティ** - ARIA 属性・キーボードナビ未検証
5. **エラーバウンダリ** - グローバルエラーハンドリング未実装

### DevOps

1. **環境変数管理** - .env ファイル分離未完 (dev/prod)
2. **シークレット管理** - JWT_SECRET 等ハードコード禁止ルール未設定
3. **ヘルスチェック** - `/healthz`エンドポイント未実装
4. **ログローテーション** - 設定なし
5. **バックアップ戦略** - DB/添付ファイルバックアップ未計画

---

## 実装ロードマップ

### Phase 1: MVP 完成 (現在 → 1-2 週間)

**目標**: 基本的なチャット機能が動作する最小限のプロダクト

1. **ESLint エラー修正** (30 分)

   - [ ] Header.test.tsx の未使用 import 削除

2. **WebSocket イベントハンドラ** (2-3 日)

   - [ ] join_channel ハンドラ
   - [ ] post_message ハンドラ (MessageUseCase と連携)
   - [ ] update_read_state ハンドラ (ReadStateUseCase と連携)
   - [ ] エラーハンドリング & ack 応答
   - [ ] 単体テスト作成

3. **WebSocket & TanStack Query 統合** (1-2 日)

   - [ ] new_message イベント → queryClient 更新
   - [ ] unread_count イベント → 未読カウント更新
   - [ ] join_channel イベント送信 (チャンネル参加時)
   - [ ] 楽観的 UI 更新 (メッセージ送信)

4. **未読バッジ UI** (1 日)

   - [ ] ChannelList に未読カウント表示
   - [ ] 未読チャンネルのハイライト
   - [ ] メッセージ閲覧時の既読 API 呼び出し

5. **本番 Docker 環境** (1-2 日)

   - [ ] Backend Dockerfile.prod (multi-stage)
   - [ ] Frontend Dockerfile.prod (nginx)
   - [ ] docker-compose.prod.yml
   - [ ] Caddyfile (TLS/proxy/compress)
   - [ ] 環境変数管理 (.env.production)

6. **ヘルスチェックエンドポイント** (1 時間)
   - [ ] GET /healthz (DB 接続確認)
   - [ ] Docker ヘルスチェック統合

### Phase 2: ファイル共有機能 (1-2 週間)

**目標**: 添付ファイルのアップロード・ダウンロード

7. **Wasabi S3 クライアント** (1 日)

   - [ ] aws-sdk-go-v2 初期化
   - [ ] Presigner 設定
   - [ ] 環境変数読み込み

8. **Attachment UseCase & Handler** (2 日)

   - [ ] Presign URL 生成 UseCase
   - [ ] Metadata 登録/取得 UseCase
   - [ ] Download presign UseCase
   - [ ] Handler 実装 (3 エンドポイント)
   - [ ] OpenAPI 動作確認

9. **Attachment UI** (2-3 日)
   - [ ] ファイルピッカー統合
   - [ ] アップロード進捗表示
   - [ ] ファイルプレビュー (画像/PDF)
   - [ ] ダウンロードボタン
   - [ ] エラーハンドリング

### Phase 3: パフォーマンス & テスト (1 週間)

**目標**: 安定性・パフォーマンス向上

10. **仮想スクロール** (1 日)

    - [ ] @tanstack/react-virtual 導入
    - [ ] MessagePanel に適用
    - [ ] 1000+メッセージでの動作確認

11. **Backend テスト** (2-3 日)

    - [ ] UseCase 単体テスト (Auth/Workspace/Channel/Message)
    - [ ] Repository 統合テスト (testcontainers)
    - [ ] WebSocket ハンドラテスト
    - [ ] カバレッジ 60%+達成

12. **Frontend E2E テスト** (1-2 日)

    - [ ] Playwright 導入
    - [ ] ログイン → メッセージ送信フロー
    - [ ] ワークスペース/チャンネル作成フロー

13. **エラーハンドリング改善** (1 日)
    - [ ] 統一エラーレスポンス構造 (backend)
    - [ ] グローバルエラーバウンダリ (frontend)
    - [ ] Toast 通知統合

### Phase 4: UX 向上 (1-2 週間)

**目標**: ユーザー体験の洗練

14. **メッセージスレッド** (2-3 日)

    - [ ] スレッド表示 UI
    - [ ] 返信フォーム
    - [ ] parent_id 連携 (backend 対応済み)

15. **入力中表示** (1 日)

    - [ ] typing イベント送信 (WebSocket)
    - [ ] "○○ が入力中..." UI
    - [ ] デバウンス処理

16. **Markdown Support** (1-2 日)

    - [ ] react-markdown 導入
    - [ ] メッセージ Markdown 表示
    - [ ] コードブロックシンタックスハイライト

17. **Message Reactions** (2 日)
    - [ ] リアクション追加/削除 API (backend UseCase/Handler - **Schema 完備**)
    - [ ] 絵文字ピッカー UI
    - [ ] リアクション表示 UI
    - [ ] WebSocket 同期

### Phase 5: 運用準備 (1 週間)

**目標**: 本番運用に向けた監視・セキュリティ

18. **Observability** (2-3 日)

    - [ ] OpenTelemetry 統合
    - [ ] Prometheus metrics (/metrics)
    - [ ] pprof 有効化 (/debug/pprof)
    - [ ] 構造化ログ (Zap) の全 Handler 適用

19. **セキュリティ強化** (1-2 日)

    - [ ] CORS origin validation (main.go:28 TODO 解消)
    - [ ] レート制限の全エンドポイント適用
    - [ ] シークレット管理 (環境変数検証)
    - [ ] CSP/X-Frame-Options ヘッダ (Caddy)

20. **CI/CD** (1-2 日)
    - [ ] GitHub Actions ワークフロー
    - [ ] Lint/Test 自動実行
    - [ ] Docker image ビルド & push
    - [ ] VPS デプロイスクリプト

### Phase 6: 機能拡張 (Post-MVP)

**目標**: より高度なコミュニケーション機能

21. **検索機能** (2-3 日)

    - [ ] Backend: チャンネル名検索 API
    - [ ] Backend: メッセージ全文検索 API (PostgreSQL FTS 追加)
    - [ ] Frontend: 検索 UI (モーダル, Ctrl+K)
    - [ ] Frontend: 検索結果ハイライト

22. **Mentions 機能** (3-4 日)

    - [ ] Backend: Mentions テーブル設計 (user_id, mentioned_by, message_id)
    - [ ] Backend: メンション通知 API
    - [ ] Frontend: @ユーザーメンション入力 (autocomplete)
    - [ ] Frontend: メンション一覧表示
    - [ ] Frontend: 未読メンション管理

23. **Direct Messages** (2-3 日)

    - [ ] Backend: Channel.isDM フラグ追加 (schema migration)
    - [ ] Backend: 1 対 1 DM 用チャンネル作成 API
    - [ ] Frontend: DM 一覧 UI
    - [ ] Frontend: DM 通知設定

24. **User Profile & Settings** (2 日)

    - [ ] プロフィール表示/編集 UI
    - [ ] アバター画像アップロード
    - [ ] ステータスメッセージ
    - [ ] 通知設定 UI

25. **Channel Management** (2-3 日)

    - [ ] チャンネル設定画面
    - [ ] 権限管理 (owner/admin/member)
    - [ ] チャンネル削除/アーカイブ
    - [ ] メンバー一覧/招待 UI

26. **Member Presence** (1-2 日)

    - [ ] オンライン状態管理 API
    - [ ] WebSocket presence イベント
    - [ ] メンバー一覧にオンライン表示
    - [ ] "最終ログイン" 表示

27. **Theme Support** (1 日)

    - [ ] ダークモード実装
    - [ ] Mantine ColorSchemeProvider 統合
    - [ ] localStorage 保存
    - [ ] システム設定連動

28. **OIDC 認証** (2-3 日)

    - [ ] AuthProvider 抽象化活用
    - [ ] Google OAuth 統合
    - [ ] GitHub OAuth 統合

29. **Storybook** (1-2 日)

    - [ ] .storybook 設定
    - [ ] Mantine/Tailwind 統合
    - [ ] 主要コンポーネントのストーリー作成

30. **モバイル最適化** (2-3 日)
    - [ ] レスポンシブナビゲーション
    - [ ] ドロワーメニュー (スマホ)
    - [ ] タッチジェスチャー対応
    - [ ] iOS/Android PWA インストール促進

## 実装済みコンポーネント一覧

### Backend (約 2,450 行)

```
backend/
├── cmd/server/main.go                              ✅ DI/ワイヤリング/エンドポイント登録
├── internal/
│   ├── domain/                                     ✅ 7 Entities + 7 Repository IF
│   │   ├── user.go, session.go, workspace.go
│   │   ├── channel.go, message.go, read_state.go, attachment.go
│   ├── usecase/                                    ✅ 5機能実装 (約1,200行)
│   │   ├── auth/                                   (Register/Login/Refresh/Logout)
│   │   ├── workspace/                              (CRUD + メンバー管理)
│   │   ├── channel/                                (List/Create)
│   │   ├── message/                                (List/Create + スレッド)
│   │   └── read_state/                             (GetUnreadCount/Update)
│   ├── infrastructure/                             ✅ (約700行)
│   │   ├── config/, logger/                        (Zap初期化)
│   │   ├── auth/                                   (JWT, Password bcrypt)
│   │   ├── db/                                     (GORM, Models)
│   │   └── repository/                             (7 Repository実装)
│   ├── interface/                                  ✅ HTTP 18EP + WS基盤 (約1,250行)
│   │   ├── http/
│   │   │   ├── router.go                           (ルート登録)
│   │   │   ├── handler/                            (Auth/Workspace/Channel/Message/ReadState)
│   │   │   └── middleware/                         (CORS/Auth/RateLimit)
│   │   └── ws/                                     (Hub/Connection, ⚠️ イベントハンドラ未実装)
│   │       ├── hub.go                              (Register/Unregister/Broadcast)
│   │       └── connection.go                       (ReadPump/WritePump, 140行)
│   └── openapi/openapi.yaml                        ✅ 791行 OpenAPI 3.1
├── schema/schema.hcl                               ✅ Atlas declarative schema
└── atlas.hcl                                       ✅ Atlas config
```

### Frontend (約 2,000+行)

```
frontend/
├── vite.config.ts, tsconfig.json                   ✅ ビルド設定
├── tailwind.config.js, postcss.config.js           ✅ スタイル設定
├── .eslintrc.json, .prettierrc                     ✅ リント設定 (⚠️ Header.test.tsx 2エラー)
├── vitest.config.ts                                ✅ テスト設定
├── src/
│   ├── main.tsx, App.tsx                           ✅ エントリーポイント
│   ├── routes/                                     ✅ 7ルート (TanStack Router)
│   │   ├── __root.tsx                              (Root layout)
│   │   ├── login.tsx, register.tsx                 (認証)
│   │   ├── app.tsx, app/index.tsx                  (App shell)
│   │   ├── app/$workspaceId.tsx                    (Workspace)
│   │   └── app/$workspaceId/$channelId.tsx         (Channel + Messages)
│   ├── lib/
│   │   ├── api/                                    ✅ OpenAPI型生成 + Client
│   │   ├── query.ts                                ✅ TanStack Query設定
│   │   ├── store/                                  ✅ Zustand (auth, workspace)
│   │   └── ws/client.ts                            ✅ WebSocket (⚠️ Query統合未完)
│   ├── features/
│   │   ├── auth/                                   ✅ Login/Register (hooks + UI + tests)
│   │   ├── workspace/                              ✅ List/Create (hooks + UI + tests)
│   │   ├── channel/                                🟡 List/Create (hooks + UI, 詳細未完)
│   │   └── message/                                🟡 List/Send (hooks + UI, スレッド/仮想スクロール未完)
│   ├── components/
│   │   └── layout/                                 ✅ AppLayout, Header (+ tests)
│   └── test/                                       ✅ 8ファイル, 27テスト (100% pass)
└── dist/                                           ✅ 本番ビルド成功
```

### DevOps

```
docker/
├── docker-compose.yml                              ✅ 開発環境 (Postgres/Backend/Frontend)
├── backend/Dockerfile.dev                          ✅
├── frontend/Dockerfile.dev                         ✅
└── .dockerignore                                   ✅
```

### Documentation

```
.
├── README.md                                       ✅ プロジェクト概要
├── plan.md                                         ✅ 本ドキュメント (更新済)
└── CLAUDE.md                                       ✅ AI Agent ガイドライン
```
