# Slack ライク・コミュニケーションアプリ 実装計画

## 技術スタック

- フロントエンド: React 19, TypeScript, Vite, Mantine 8, Tailwind CSS, TanStack Router, TanStack Query, Vitest, Storybook, PWA（vite-plugin-pwa）
- バックエンド: Go 1.22+, Gin, Clean Architecture, WebSocket, GORM + Gen（ORM/コード生成）, Atlas（宣言的マイグレーション）
- データベース: PostgreSQL
- オブジェクトストレージ: Wasabi（S3 互換, aws-sdk-go-v2）
- デプロイ: Docker（compose）, リバースプロキシ（Caddy または Nginx）を想定, VPS 運用

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
  - `lib/`（apiClient, wsClient, queryClient, store）
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

## マイルストーン

1. ✅ **スケルトン/起動** - CA構成、GORM初期化、全Domain層定義完了
2. ✅ **Atlas導入** - 宣言的スキーマ(schema.hcl)、全テーブル定義完了
3. ✅ **認証/セッション** - Repository実装、Auth UseCase/Handler実装、JWT/Refresh完了
4. 🚧 **Workspace/Channel** - Repository完了、UseCase/Handler実装中
5. ⏳ **Message** - Repository完了、UseCase/Handler未実装（CRUD + Thread + 添付presign）
6. ⏳ **未読管理** - Repository完了、API未実装（最終既読upsert/集計API）
7. ⏳ **WebSocket** - Hub/Connection骨組み完了、イベントハンドラ未実装
8. ⏳ **フロント統合** - 基盤未着手（Router/Query/WS、未読UI）
9. ⏳ **PWA** - manifest作成済み、SW実装未着手
10. ⏳ **デプロイ/可観測性** - Docker/Caddy構成未着手

## 実装状況サマリー

### Backend 実装進捗: 約40%
- ✅ **Domain層**: 100% - 全エンティティ & Repository IF 定義完了
- ✅ **Infrastructure層**: 100% - Config, Logger, Auth Services, DB Models, Repository実装完了
- 🟡 **UseCase層**: 20% - Auth完了、Workspace/Channel/Message/ReadState未実装
- 🟡 **Interface層**: 30% - Auth Handler完了、その他未実装、WebSocket骨組みのみ
- ✅ **DB Schema**: 100% - Atlas schema.hcl全テーブル定義完了

### Frontend 実装進捗: 約60%
- ✅ **基盤**: Vite + PWA plugin設定完了、依存関係インストール完了、TypeScript/Tailwind/PostCSS/ESLint/Prettier/Vitest 設定完了
- ✅ **OpenAPI型生成**: openapi-typescript でスキーマ生成完了
- ✅ **APIクライアント**: openapi-fetch ベースのクライアント実装完了、自動認証リフレッシュ実装
- ✅ **状態管理**: Zustand で Auth/Workspace ストア実装完了
- ✅ **データフェッチ**: TanStack Query セットアップ完了
- ✅ **認証機能**: Login/Register フォーム実装完了、Auth hooks 実装完了
- ✅ **ワークスペース**: 一覧/作成 UI 実装完了、hooks 実装完了
- ✅ **WebSocket**: クライアント骨組み実装完了（接続/再接続/イベント管理）
- ✅ **ビルド**: 本番ビルド成功確認済み
- ✅ **テスト**: Vitest設定完了、12個のテストケース実装済み（Login/Register/Workspace UI）、全テスト通過
- 🟡 **チャネル/メッセージ**: UI未実装

### 次の優先タスク（バックエンド）
1. Workspace UseCase & Handler 実装
2. Channel UseCase & Handler 実装
3. Message UseCase & Handler 実装
4. ReadState API 実装
5. WebSocket イベントハンドラ実装

### 次の優先タスク（フロントエンド）
1. Channel 機能実装（hooks + UI）
2. Message 機能実装（hooks + UI + 仮想スクロール）
3. WebSocket統合（新着メッセージ/未読カウント）
4. 添付ファイル機能実装
5. Storybook セットアップ

---

## To-dos

### 完了済み ✅
- [x] **モノレポ構成** - frontend/backend/docker/schema
- [x] **クリーンアーキテクチャ構成**
  - ✅ Domain層: 全エンティティ & Repository IF (7種類)
  - ✅ Infrastructure層: Config, Logger, Auth Services, DB, Repository実装 (7種類)
  - ✅ Interface層: Router, Middleware (CORS/Auth/RateLimit), WebSocket骨組み
- [x] **OpenAPI 3.1 スキーマ** - auth/workspace/channel/message/reads/attachments 完全定義
- [x] **Atlas導入** - atlas.hcl, schema/schema.hcl 全テーブル定義完了
- [x] **GORM導入** - 接続/モデル定義完了、ビルド成功
- [x] **Repository層実装** (100%) - User, Session, Workspace, Channel, Message, ReadState, Attachment
- [x] **Auth UseCase実装** (100%) - Register/Login/Refresh/Logout
- [x] **Auth Handler実装** (100%) - Register/Login/Refresh/Logout エンドポイント + バリデーション
- [x] **DI/統合** - main.go で DB初期化、Repository/UseCase/Handler ワイヤリング完了
- [x] **ビルド検証** - `go build` 成功、実行可能バイナリ生成確認

### 進行中 🚧
- [ ] workspace/channel UseCase 実装
- [ ] workspace/channel/message HTTP ハンドラ実装

### 完了済み（フロントエンド） ✅
- [x] **フロント初期化** - Vite+React19+TS+Mantine8+Tailwind+ESLint/Prettier 完了
- [x] **OpenAPI クライアント生成** - openapi-typescript+openapi-fetch セットアップ完了
- [x] **TanStack Query 基盤** - セットアップ完了、認証/ワークスペース hooks 実装
- [x] **セッション管理** - Zustand で実装、localStorage連携
- [x] **認証 UI** - Login/Register フォーム実装完了
- [x] **ワークスペース UI** - 一覧/作成モーダル実装完了
- [x] **WebSocket クライアント** - 基本実装完了（接続/再接続/イベント管理）
- [x] **Vitest 導入** - 設定完了、テスト基盤構築
- [x] **PWA 基盤** - Vite PWA plugin 設定、manifest 定義、Workbox キャッシュ戦略

### 未着手 📋
- [ ] AuthProvider 抽象と OIDC 下地
- [ ] 未読管理 API 実装（upsert/集計/最適化）
- [ ] WebSocket イベントハンドラ実装（join_channel, post_message, typing, etc.）
- [ ] Wasabi S3 クライアント実装（presign/upload/download）
- [ ] チャネル UI（一覧/作成/詳細）
- [ ] メッセージ UI（一覧/送信/スレッド/仮想スクロール）
- [ ] 添付ファイル UI: presign/アップロード/表示
- [ ] WS 統合と Query 部分更新（未読/新着）
- [ ] 未読バッジ UI 実装
- [ ] Storybook 導入・ストーリー作成
- [ ] テスト拡充（jest-dom型定義修正、E2Eテスト）
- [ ] Docker/Caddy 構成と VPS デプロイ準備
- [ ] 可観測性実装（ログ/メトリクス/pprof/レート制限統合）
- [ ] Atlas マイグレーション適用（初回 migrate apply）

## 実装済みファイル一覧

### Backend
```
backend/
├── cmd/server/main.go                              ✅ サーバー起動 + DI/ワイヤリング完了
├── internal/
│   ├── domain/                                     ✅ 全エンティティ & Repository IF 完了
│   │   ├── user.go                                 (User, UserRepository)
│   │   ├── workspace.go                            (Workspace, WorkspaceMember, WorkspaceRepository)
│   │   ├── channel.go                              (Channel, ChannelMember, ChannelRepository)
│   │   ├── message.go                              (Message, MessageReaction, MessageRepository)
│   │   ├── read_state.go                           (ChannelReadState, ReadStateRepository)
│   │   ├── attachment.go                           (Attachment, AttachmentRepository)
│   │   └── session.go                              (Session, SessionRepository)
│   ├── usecase/                                    ✅ Auth UseCase 実装完了
│   │   └── auth/
│   │       ├── dto.go                              (RegisterInput/Output, LoginInput/Output, etc.)
│   │       └── interactor.go                       (Register/Login/Refresh/Logout ビジネスロジック)
│   ├── infrastructure/
│   │   ├── config/config.go                        ✅ 環境変数管理
│   │   ├── logger/logger.go                        ✅ Zap logger
│   │   ├── auth/
│   │   │   ├── jwt.go                              ✅ JWTService + 旧JWTManager
│   │   │   └── password.go                         ✅ PasswordService + 旧関数
│   │   ├── db/
│   │   │   ├── db.go                               ✅ GORM 接続 + InitDB
│   │   │   └── models.go                           ✅ 全 GORM モデル（User, Session, Workspace, Channel, Message, etc.）
│   │   └── repository/                             ✅ 全Repository実装完了（7つ）
│   │       ├── user_repository.go                  (UserRepository 実装)
│   │       ├── session_repository.go               (SessionRepository 実装)
│   │       ├── workspace_repository.go             (WorkspaceRepository 実装)
│   │       ├── channel_repository.go               (ChannelRepository 実装)
│   │       ├── message_repository.go               (MessageRepository 実装)
│   │       ├── read_state_repository.go            (ReadStateRepository 実装)
│   │       └── attachment_repository.go            (AttachmentRepository 実装)
│   ├── interface/
│   │   ├── http/
│   │   │   ├── router.go                           ✅ Auth エンドポイント登録完了
│   │   │   ├── handler/
│   │   │   │   ├── auth_handler.go                 ✅ Register/Login/Refresh/Logout ハンドラ
│   │   │   │   └── dto.go                          ✅ リクエスト/レスポンスDTO + バリデーション
│   │   │   └── middleware/
│   │   │       ├── auth.go                         ✅ JWT 認証ミドルウェア
│   │   │       ├── cors.go                         ✅ CORS
│   │   │       └── ratelimit.go                    ✅ レート制限
│   │   └── ws/
│   │       ├── hub.go                              ✅ WebSocket ハブ
│   │       └── connection.go                       ✅ WebSocket コネクション管理
│   └── openapi/openapi.yaml                        ✅ OpenAPI 3.1 完全定義
├── schema/schema.hcl                               ✅ Atlas 宣言的スキーマ（全テーブル）
├── atlas.hcl                                       ✅ Atlas 設定
└── bin/server                                      ✅ ビルド済みバイナリ（認証機能動作可能）
```

### Frontend
```
frontend/
├── vite.config.ts                                  ✅ Vite + PWA + alias 設定完了
├── tsconfig.json                                   ✅ TypeScript設定完了
├── tailwind.config.js                              ✅ Tailwind CSS設定完了
├── postcss.config.js                               ✅ PostCSS + @tailwindcss/postcss 設定完了
├── .eslintrc.json                                  ✅ ESLint設定完了
├── .prettierrc                                     ✅ Prettier設定完了
├── vitest.config.ts                                ✅ Vitest設定完了
├── package.json                                    ✅ 依存関係インストール完了
├── src/
│   ├── main.tsx                                    ✅ エントリーポイント（MantineProvider + QueryClient + App）
│   ├── App.tsx                                     ✅ ルーティング + 認証ガード実装
│   ├── vite-env.d.ts                               ✅ 環境変数型定義
│   ├── styles/globals.css                          ✅ Tailwind + グローバルスタイル
│   ├── lib/
│   │   ├── api/
│   │   │   ├── schema.ts                           ✅ OpenAPI型定義（生成済み）
│   │   │   └── client.ts                           ✅ APIクライアント + 認証インターセプター
│   │   ├── query.ts                                ✅ TanStack Query設定
│   │   ├── store/
│   │   │   ├── auth.ts                             ✅ 認証ストア（Zustand + persist）
│   │   │   └── workspace.ts                        ✅ ワークスペースストア
│   │   └── ws/
│   │       └── client.ts                           ✅ WebSocketクライアント（再接続機能付き）
│   ├── features/
│   │   ├── auth/
│   │   │   ├── hooks/useAuth.ts                    ✅ Login/Register/Logout hooks
│   │   │   └── components/
│   │   │       ├── LoginForm.tsx                   ✅ ログインフォーム + テスト
│   │   │       └── RegisterForm.tsx                ✅ 登録フォーム + テスト
│   │   └── workspace/
│   │       ├── hooks/useWorkspace.ts               ✅ Workspace CRUD hooks
│   │       └── components/
│   │           ├── WorkspaceList.tsx               ✅ ワークスペース一覧
│   │           └── CreateWorkspaceModal.tsx        ✅ 作成モーダル
│   └── test/setup.ts                               ✅ Vitest + Testing Library セットアップ
└── dist/                                           ✅ ビルド成功（本番用アセット生成済み）
```

### Root
```
.
├── .gitignore                                      ✅ 更新済み
├── package.json                                    ✅ Turbo スクリプト
├── pnpm-workspace.yaml                             ✅ ワークスペース定義
├── turbo.json                                      ✅ Turbo 設定
├── README.md                                       ✅ プロジェクト概要
└── plan.md                                         ✅ 本ドキュメント
```
