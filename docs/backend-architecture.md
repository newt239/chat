# バックエンドアーキテクチャドキュメント

## 概要

このドキュメントでは、チャットアプリケーションのバックエンドのアーキテクチャとディレクトリ構成について説明します。バックエンドは Go 言語で実装されており、クリーンアーキテクチャの原則に従って設計されています。

## 技術スタック

- **言語**: Go 1.24.6
- **Web フレームワーク**: Echo v4.15.4 (Connect RPC のハンドラーと WebSocket をマウント)
- **API**: Connect RPC (connect-go v1.21) + Protocol Buffers (buf で `proto/` からコード生成)
- **ORM**: Ent v0.14.5 (スキーマ駆動型 ORM)
- **データベース**: PostgreSQL
- **認証**: JWT (golang-jwt/jwt/v5 v5.2.0)
- **WebSocket**: Gorilla WebSocket v1.5.3
- **ストレージ**: Wasabi (S3 互換、AWS SDK v2)
- **ログ**: Zap v1.26.0
- **バリデーション**: protovalidate (proto に書いた制約を Connect の interceptor で検証)

## ディレクトリ構成

```bash
backend/
├── cmd/                          # アプリケーションエントリーポイント
│   ├── server/                   # メインサーバー
│   ├── seed/                     # データベースシード
│   ├── reset/                    # データベースリセット
│   └── migrate/                  # マイグレーション
├── ent/                          # Ent ORM 生成コード (20スキーマ)
│   └── schema/                   # Entスキーマ定義
└── internal/                     # 内部パッケージ (総行数: 約20,209行)
    ├── domain/                   # ドメイン層 (976行)
    │   ├── entity/               # エンティティ (15ファイル)
    │   ├── repository/           # リポジトリインターフェース (16ファイル)
    │   ├── service/              # ドメインサービス (ChannelAccessService)
    │   ├── errors/               # ドメインエラー定義
    │   └── transaction/          # トランザクションインターフェース
    ├── usecase/                  # ユースケース層 (6,226行)
    │   ├── auth/                 # 認証ユースケース
    │   ├── user/                 # ユーザーユースケース
    │   ├── workspace/            # ワークスペースユースケース
    │   ├── channel/              # チャンネルユースケース
    │   ├── channelmember/        # チャンネルメンバーユースケース
    │   ├── dm/                   # DMユースケース
    │   ├── message/              # メッセージユースケース
    │   │   ├── interactor.go    # メインインターフェース
    │   │   ├── creator.go       # メッセージ作成
    │   │   ├── updater.go       # メッセージ更新
    │   │   ├── deleter.go       # メッセージ削除
    │   │   ├── lister.go        # メッセージ一覧・スレッド取得
    │   │   └── output_builder.go # 出力データ組み立て
    │   ├── thread/               # スレッドユースケース
    │   ├── systemmessage/        # システムメッセージユースケース
    │   ├── attachment/           # 添付ファイルユースケース
    │   ├── bookmark/             # ブックマークユースケース
    │   ├── pin/                  # ピン留めユースケース
    │   ├── link/                 # リンクユースケース
    │   ├── reaction/             # リアクションユースケース
    │   ├── readstate/            # 既読状態ユースケース
    │   ├── user_group/           # ユーザーグループユースケース
    │   └── search/               # 検索ユースケース
    ├── infrastructure/           # インフラストラクチャ層 (6,806行)
    │   ├── auth/                 # JWT認証実装
    │   ├── config/               # 設定管理
    │   ├── repository/           # リポジトリ実装 (16ファイル、254メソッド)
    │   ├── transaction/          # トランザクション実装 (Entベース)
    │   ├── storage/              # Wasabi S3ストレージ
    │   ├── logger/               # Zapロガー
    │   ├── mention/              # メンション処理サービス
    │   ├── link/                 # リンク処理サービス
    │   ├── ogp/                  # OGPメタデータ取得
    │   ├── utils/                # ユーティリティ (Ent変換など)
    │   └── seed/                 # データシード
    ├── interfaces/               # インターフェース層
    │   ├── handler/              # 外部インターフェース
    │   │   ├── http/             # Echo のルーター (Connect RPC・WebSocket・ヘルスチェックのマウント)
    │   │   ├── rpc/              # Connect RPC のサービス実装と interceptor (認証・入力検証・エラー変換)
    │   │   └── websocket/        # WebSocket の Hub と、ユースケースの変更通知の配信
    │   └── presenter/            # ユースケースの出力を proto のメッセージに変換
    ├── gen/                      # buf が proto から生成したコード (編集しない)
    ├── registry/                 # 依存性注入コンテナ (Registryパターン)
    │   ├── registry.go           # メインレジストリ
    │   ├── domain_registry.go    # ドメイン層の依存解決
    │   ├── infrastructure_registry.go  # インフラ層の依存解決
    │   ├── usecase_registry.go   # ユースケース層の依存解決
    │   └── interface_registry.go # インターフェース層の依存解決
```

## アーキテクチャパターン

### クリーンアーキテクチャ

このプロジェクトはクリーンアーキテクチャの原則に従って設計されています：

1. **ドメイン層 (Domain Layer)**

   - ビジネスロジックの中核
   - 外部依存を持たない
   - エンティティ、リポジトリインターフェース、ドメインサービスを含む

2. **ユースケース層 (Use Case Layer)**

   - アプリケーション固有のビジネスロジック
   - ドメイン層のインターフェースに依存
   - 各機能ごとに分離されたユースケース

3. **インターフェース層 (Interface Layer)**

   - 外部システムとの接続
   - HTTP ハンドラー、WebSocket ハンドラー
   - 外部からの入力をユースケースに変換
   - ミドルウェアによる共通処理

4. **インフラストラクチャ層 (Infrastructure Layer)**
   - 外部システムの実装
   - データベース、ストレージ、認証などの具体的な実装

## 主要コンポーネント

### 1. エンティティ (Domain Entities)

```go
// ユーザーエンティティ
type User struct {
    ID           string
    Email        string
    PasswordHash string
    DisplayName  string
    AvatarURL    *string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// メッセージエンティティ
type Message struct {
    ID        string
    ChannelID string
    UserID    string
    ParentID  *string
    Body      string
    CreatedAt time.Time
    EditedAt  *time.Time
    DeletedAt *time.Time
    DeletedBy *string
}
```

### 2. リポジトリパターン

各エンティティに対応するリポジトリインターフェースを定義：

```go
type UserRepository interface {
    FindByID(ctx context.Context, id string) (*entity.User, error)
    FindByEmail(ctx context.Context, email string) (*entity.User, error)
    Create(ctx context.Context, user *entity.User) error
    Update(ctx context.Context, user *entity.User) error
    Delete(ctx context.Context, id string) error
}
```

### 3. ユースケース層

ビジネスロジックを実装するインタラクター：

```go
type AuthUseCase interface {
    Register(ctx context.Context, input RegisterInput) (*AuthOutput, error)
    Login(ctx context.Context, input LoginInput) (*AuthOutput, error)
    RefreshToken(ctx context.Context, input RefreshTokenInput) (*AuthOutput, error)
    Logout(ctx context.Context, input LogoutInput) (*LogoutOutput, error)
}
```

#### 責務別の分割設計

大規模なユースケース（例: Message）は責務ごとにファイルを分割：

```go
// usecase/message/
// - interactor.go: メインインターフェース（各実装へ委譲）
// - creator.go: メッセージ作成ロジック
// - updater.go: メッセージ更新ロジック
// - deleter.go: メッセージ削除ロジック
// - lister.go: メッセージ一覧・スレッド取得
// - output_builder.go: 出力データ組み立て

type MessageUseCase interface {
    Create(ctx context.Context, input CreateInput) (*MessageOutput, error)
    Update(ctx context.Context, input UpdateInput) (*MessageOutput, error)
    Delete(ctx context.Context, input DeleteInput) error
    List(ctx context.Context, input ListInput) (*ListOutput, error)
    GetThreadReplies(ctx context.Context, input GetThreadRepliesInput) (*ThreadRepliesOutput, error)
}
```

### 4. 依存性注入 (Registry Pattern)

`registry`パッケージで依存性注入を 4 層に分けて管理：

```go
type Registry struct {
    domainRegistry         *DomainRegistry
    infrastructureRegistry *InfrastructureRegistry
    usecaseRegistry        *UseCaseRegistry
    interfaceRegistry      *InterfaceRegistry
}

// DomainRegistry: リポジトリとドメインサービスのインスタンス化
func (r *DomainRegistry) NewUserRepository() repository.UserRepository {
    return infrarepository.NewUserRepository(r.client, r.logger)
}

// InfrastructureRegistry: 外部サービス（JWT, OGP, Storage等）の初期化
func (r *InfrastructureRegistry) NewJWTService() *auth.JWTService {
    return auth.NewJWTService(r.config.JWT)
}

// UseCaseRegistry: 各ユースケースの依存解決
func (r *UseCaseRegistry) NewAuthUseCase() auth.AuthUseCase {
    return auth.NewAuthInteractor(
        r.domainRegistry.NewUserRepository(),
        r.domainRegistry.NewSessionRepository(),
        r.infrastructureRegistry.NewJWTService(),
        r.infrastructureRegistry.NewPasswordService(),
        r.domainRegistry.NewTransactionManager(),
        r.infrastructureRegistry.NewLogger(),
    )
}

// InterfaceRegistry: Connect RPC のサービスとルーターの構築
func (r *InterfaceRegistry) NewRPCHandler() http.Handler {
    uc := r.usecaseRegistry
    return rpc.NewHandler(r.infrastructureRegistry.NewJWTService(),
        rpc.Register(chatv1connect.NewAuthServiceHandler, chatv1connect.AuthServiceHandler(&rpc.AuthServer{UC: uc.NewAuthUseCase()})),
        // ...
    )
}
```

## 主要機能

### 1. 認証・認可

- JWT ベースの認証
- アクセストークン（既定 15 分）とリフレッシュトークン（既定 30 日）。それぞれ `JWT_ACCESS_TOKEN_TTL` / `JWT_REFRESH_TOKEN_TTL` で変更できます
- セッション管理
- パスワードハッシュ化（bcrypt）

### 2. ワークスペース管理

- マルチテナント対応
- ワークスペースメンバー管理
- ロールベースアクセス制御

### 3. チャンネル管理

- パブリック・プライベートチャンネル
- 階層チャンネル（スラッシュ区切りのパス。親子関係と集約表示は [channel-hierarchy.md](./channel-hierarchy.md)）
- DM・グループ DM（自分を含めて最大 10 人。超える場合はメンバー指定の非公開チャンネルを作る）
- チャンネル・DM へのスター、チャンネルの関連リンク
- チャンネルメンバー管理
- ロールベースアクセス制御（owner, admin, member, guest）
- システムメッセージによる変更履歴記録

### 4. メッセージング

- リアルタイムメッセージング（WebSocket）
- メッセージの編集・削除（論理削除）
- スレッド機能（親子関係）
- メンション機能（@user, @group）
- リアクション機能（絵文字）
- ピン留め機能
- メッセージ内リンクの OGP プレビュー（YouTube の動画情報を含む）
- 同じワークスペースのメッセージリンクの引用カード
- 出力は `MessageOutputBuilder` で閲覧者ごとに組み立てる（詳細は [message-display.md](./message-display.md)）

### 5. ファイル管理

- ファイルアップロード（Wasabi S3 互換ストレージ）
- プリサインド URL 生成
- メタデータ管理（画像・動画・音声の寸法と再生時間はクライアントが計測して送る）

### 6. ブックマーク機能

- メッセージのブックマーク
- ブックマーク一覧表示
- ブックマーク削除

### 7. リンク機能

- URL の自動検出
- OGP メタデータの取得
- リンクプレビュー表示

### 8. リアクション機能

- メッセージへのリアクション
- リアクション一覧表示
- リアクション削除

### 9. 既読状態管理

- チャンネル別の既読状態
- スレッド別の既読状態
- 未読メッセージカウント
- メンション数カウント
- 既読状態の更新（WebSocket による通知）

### 10. ユーザーグループ機能

- ユーザーグループの作成・管理
- グループメンバー管理
- グループ権限管理

### 11. 通知システム

- WebSocket ベースのリアルタイム通知。各ユースケースが必要な通知だけを `Notifier` インターフェースとして定義し、`interfaces/handler/websocket` の `Notifier` がまとめて実装する
- 新規メッセージの即時配信
- 未読メッセージカウント更新
- メンション通知
- 既読状態更新通知
- Hub パターンによる接続管理

### 12. 検索機能

- メッセージの全文検索
- チャンネル検索
- ユーザー検索
- ユーザーグループ検索
- 横断検索機能

### 13. スレッド機能

- スレッド返信の取得
- スレッドメタデータ（返信数、参加者、未読数）
- スレッドフォロー機能
- スレッド既読状態管理

## API 設計

### Connect RPC

API は `proto/chat/v1/*_service.proto` で定義しています。パスは `/chat.v1.<Service>/<Method>` で、すべて POST です (Connect プロトコル、JSON)。`AuthService` の `Register` / `Login` / `Refresh` 以外は `Authorization: Bearer <アクセストークン>` が必要です。

| サービス | RPC | 定義 |
| --- | --- | --- |
| `AdminService` | ListAuditLogs / ExportAuditLogs / ListAdminMembers / SuspendMember / ResumeMember | `admin_service.proto` |
| `AttachmentService` | PresignUpload / GetAttachment / GetDownloadUrl / DeleteAttachment | `attachment_service.proto` |
| `AuthService` | Register / Login / Refresh / Logout | `auth_service.proto` |
| `BookmarkService` | ListBookmarks / AddBookmark / RemoveBookmark | `bookmark_service.proto` |
| `ChannelMemberService` | ListChannelMembers / InviteChannelMember / JoinChannel / LeaveChannel / RemoveChannelMember / UpdateChannelMemberRole | `channel_member_service.proto` |
| `ChannelService` | ListChannels / CreateChannel / GetChannel / UpdateChannel / DeleteChannel / ArchiveChannel / UnarchiveChannel | `channel_service.proto` |
| `DirectMessageService` | ListDirectMessages / CreateDirectMessage / CreateGroupDirectMessage | `direct_message_service.proto` |
| `InsightService` | GetInsights | `insight_service.proto` |
| `LinkService` | FetchOgp | `link_service.proto` |
| `MentionService` | ListMentions | `mention_service.proto` |
| `MessageService` | ListMessages / ListMessagesWithThread / CreateMessage / UpdateMessage / DeleteMessage | `message_service.proto` |
| `PermissionService` | GetPermissions / UpdatePermission | `permission_service.proto` |
| `PinService` | ListPins / CreatePin / DeletePin | `pin_service.proto` |
| `ReactionService` | ListReactions / AddReaction / RemoveReaction | `reaction_service.proto` |
| `ReadStateService` | UpdateReadState / GetUnreadCount | `read_state_service.proto` |
| `SearchService` | SearchWorkspace | `search_service.proto` |
| `ThreadService` | GetThreadReplies / GetThreadMetadata / ListParticipatingThreads / MarkThreadRead / FollowThread / UnfollowThread | `thread_service.proto` |
| `UserGroupService` | CreateUserGroup / ListUserGroups / GetUserGroup / UpdateUserGroup / DeleteUserGroup / ListUserGroupMembers / AddUserGroupMember / RemoveUserGroupMember | `user_group_service.proto` |
| `UserService` | GetMe / UpdateMe / UpdatePassword / DeleteMe | `user_service.proto` |
| `WorkspaceService` | ListWorkspaces / CreateWorkspace / GetWorkspace / UpdateWorkspace / DeleteWorkspace / ListPublicWorkspaces / JoinPublicWorkspace / ListMembers / AddMemberByEmail / UpdateMemberRole / RemoveMember | `workspace_service.proto` |

ヘルスチェックのみ `GET /healthz` で提供しています。

エラーは Connect のエラーコードで返します。ユースケースのエラーとの対応は `internal/interfaces/handler/rpc/error.go` にまとめています (例: 見つからない → `not_found`、権限なし → `permission_denied`、入力の制約違反 → `invalid_argument`)。

### WebSocket

- エンドポイント: `GET /ws?token=<JWT>&workspaceId=<id>`
- JWT 認証による接続、`CORS_ALLOWED_ORIGINS` による Origin 検証
- メッセージは `proto/chat/v1/event.proto` の `ClientEvent` / `ServerEvent` を protojson で JSON にしたもの（例: `{"joinChannel":{"channelId":"..."}}`）
- クライアント → サーバー: `joinChannel` / `leaveChannel` / `typing` / `stopTyping` / `viewChannel`
- サーバー → クライアント: `newMessage` / `messageUpdated` / `messageDeleted` / `unreadCount` / `pinCreated` / `pinDeleted` / `systemMessageCreated` / `reactionAdded` / `reactionRemoved` / `typing` / `stopTyping` / `channelViewers` / `ack` / `error`

## データベース設計

### 主要テーブル（Ent スキーマ）

- `user` - ユーザー情報
- `session` - セッション管理
- `workspace` - ワークスペース
- `workspace_member` - ワークスペースメンバー
- `channel` - チャンネル
- `channel_member` - チャンネルメンバー
- `channel_read_state` - チャンネル既読状態
- `channel_star` - チャンネル・DM へのスター
- `channel_link` - チャンネルの関連リンク
- `message` - メッセージ
- `message_reaction` - メッセージリアクション
- `message_pin` - ピン留めメッセージ
- `message_bookmark` - ブックマーク
- `message_user_mention` - ユーザーメンション
- `message_group_mention` - グループメンション
- `message_link` - メッセージ内リンク
- `attachment` - 添付ファイル
- `user_group` - ユーザーグループ
- `user_group_member` - ユーザーグループメンバー
- `thread_read_state` - スレッド既読状態
- `user_thread_follow` - スレッドフォロー
- `system_message` - システムメッセージ（チャンネルの変更履歴）
- `audit_log` - 管理操作の監査ログ（[admin-insights.md](./admin-insights.md)）
- `workspace_permission` - ロールごとの操作権限のうち既定値から変更されたもの
- `user_note` - 自分だけに見える相手ユーザーのニックネームとメモ

### リレーション

- ユーザー ↔ ワークスペース（多対多: workspace_member）
- ワークスペース → チャンネル（1 対多）
- チャンネル → チャンネル（1 対多: 階層の親子関係）
- チャンネル ↔ ユーザー（多対多: channel_member）
- チャンネル → メッセージ（1 対多）
- メッセージ → メッセージ（1 対多: スレッド親子関係）
- メッセージ → リアクション（1 対多）
- メッセージ → 添付ファイル（1 対多）
- メッセージ → ピン留め（1 対多）
- メッセージ → ブックマーク（1 対多）
- メッセージ → ユーザーメンション（1 対多）
- メッセージ → グループメンション（1 対多）
- メッセージ → リンク（1 対多）
- ユーザー + チャンネル → 既読状態（複合キー）
- ユーザー + メッセージ → スレッド既読状態（複合キー）
- ユーザー + メッセージ → スレッドフォロー（複合キー）
- ユーザー ↔ ユーザーグループ（多対多: user_group_member）
- チャンネル → システムメッセージ（1 対多）

## 設定管理

環境変数による設定管理：

```go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    JWT      JWTConfig
    Wasabi   WasabiConfig
    CORS     CORSConfig
    Logger   LoggerConfig
}
```

## テスト戦略

`go test ./...` で実行します。CI（`.github/workflows/codecheck.yml`）でもビルド・lint と併せて実行されます。

### 現在あるテスト

| 対象 | 内容 |
| --- | --- |
| `internal/interfaces/handler/http/router_test.go` | Connect RPC とヘルスチェックのマウント |
| `internal/interfaces/handler/rpc/handler_test.go` | 認証 interceptor・入力検証・エラーコード変換 |
| `internal/usecase/workspace/interactor_test.go` | ワークスペースのロール変更・招待の権限と監査ログ |
| `internal/usecase/admin/interactor_test.go` | メンバーの停止・再開、監査ログの閲覧・CSV、権限設定の変更 |
| `internal/usecase/insight/interactor_test.go` | インサイトの前期比較・タイムゾーン・非公開チャンネルの扱い |
| `internal/usecase/channel/interactor_test.go` | 権限設定に基づくチャンネル作成、アーカイブ |
| `internal/usecase/auth/interactor_test.go` | ログインの監査ログ、リフレッシュ時のセッションの差し替え |
| `internal/domain/entity/permission_test.go` | 権限マトリクスの既定値と上書き |
| `internal/usecase/dm/interactor_test.go` | DM 作成時のワークスペースメンバー検証 |
| `internal/usecase/thread/reader_test.go` | スレッド既読・フォローの認可 |
| `internal/usecase/message/creator_test.go` | 添付ファイルの所有者・チャンネル検証 |
| `internal/infrastructure/ogp/ogp_test.go` | OGP 取得の内部ネットワーク遮断 |

### 方針

1. **ユニットテスト** — ドメインリポジトリをスタブに差し替え、ユースケースの分岐（特に権限チェック）を検証する
2. **統合テスト** — データベースや WebSocket を含む結合の検証は今後追加する

## デプロイメント

### Docker 対応

- マルチステージビルド
- 本番環境用の最適化されたイメージ
- 環境変数による設定

### 環境

- **開発環境**: Docker Compose
- **本番環境**: 環境変数による設定

## セキュリティ

### 認証・認可

- JWT トークンベース認証
- アクセストークン（既定 15 分）とリフレッシュトークン（既定 30 日）。それぞれ `JWT_ACCESS_TOKEN_TTL` / `JWT_REFRESH_TOKEN_TTL` で変更できます
- リフレッシュトークンによるセッション管理
- パスワードのハッシュ化（bcrypt）
- Connect の interceptor によるトークン検証

### データ保護

- リクエストバリデーション（protovalidate）
- SQL インジェクション対策（Ent ORM）
- CORS 設定（環境変数による制御）
- ファイルアップロードの検証

### 通信

- HTTPS 対応
- WebSocket の JWT 認証
- プリサインド URL（S3）による安全なファイルアクセス

## 監視・ログ

### ログ

- Zap による構造化ログ
- ログレベル管理
- エラートラッキング
- リクエスト/レスポンスログ

### 観測性

- メトリクス収集
- トレーシング
- パフォーマンス監視

### ヘルスチェック

- `/healthz` エンドポイント
- データベース接続確認
- 外部サービス接続確認

## トランザクション管理

### Context ベースのトランザクション伝播

```go
// domain/transaction/manager.go
type Manager interface {
    Do(ctx context.Context, fn func(ctx context.Context) error) error
}

// infrastructure/transaction/manager.go
func (m *transactionManager) Do(ctx context.Context, fn func(context.Context) error) error {
    tx, err := m.client.Tx(ctx)
    if err != nil {
        return err
    }

    // Context にトランザクションを格納
    ctxWithTx := contextWithTx(ctx, tx)

    // ビジネスロジック実行
    if err := fn(ctxWithTx); err != nil {
        if rerr := tx.Rollback(); rerr != nil {
            return fmt.Errorf("rollback failed: %w (original error: %v)", rerr, err)
        }
        return err
    }

    return tx.Commit()
}
```

### リポジトリでのトランザクション解決

```go
// infrastructure/repository/channel_repository.go
func (r *channelRepository) FindByID(ctx context.Context, id string) (*entity.Channel, error) {
    // Context からトランザクションを自動検出
    client := transaction.ResolveClient(ctx, r.client)

    ch, err := client.Channel.Query().
        Where(channel.ID(parsedID)).
        First(ctx)
    // ...
}
```

**利点**:

- ユースケース層でトランザクション境界を明示
- リポジトリ層はトランザクションを意識しない
- テストでモックが容易

## パフォーマンス最適化

### N+1 問題の回避

```go
// 一括取得メソッドの実装
func (r *messageRepository) FindByIDs(ctx context.Context, ids []string) ([]*entity.Message, error)
func (r *reactionRepository) FindByMessageIDs(ctx context.Context, messageIDs []string) ([]*entity.MessageReaction, error)

// Ent の Eager Loading
messages, err := client.Message.Query().
    WithUser().           // ユーザー情報を一括取得
    WithAttachments().    // 添付ファイルを一括取得
    Where(...).
    All(ctx)
```

### ページネーション

```go
// メッセージ一覧取得でのページネーション
type ListInput struct {
    ChannelID string
    Limit     int       // デフォルト50
    Before    *string   // カーソルベースページング
    After     *string
}
```

## エラー設計

### 3 層のエラー管理

1. **ドメインエラー** (`domain/errors/`):

   ```go
   var (
       ErrNotFound      = errors.New("リソースが見つかりません")
       ErrUnauthorized  = errors.New("権限がありません")
       ErrValidation    = errors.New("入力値が不正です")
   )
   ```

2. **ユースケース固有エラー** (各 `usecase/*/dto.go`):

   ```go
   var (
       ErrChannelNotFound       = errors.New("チャンネルが見つかりません")
       ErrMessageAlreadyDeleted = errors.New("メッセージは既に削除されています")
   )
   ```

3. **HTTP エラーマッピング** (ハンドラー層):
   ```go
   func mapMessageError(err error) error {
       switch err {
       case messageuc.ErrMessageNotFound:
           return echo.NewHTTPError(http.StatusNotFound, err.Error())
       case messageuc.ErrUnauthorized:
           return echo.NewHTTPError(http.StatusForbidden, err.Error())
       default:
           return echo.NewHTTPError(http.StatusInternalServerError, "内部エラーが発生しました")
       }
   }
   ```

## 共通処理の抽象化

### ChannelAccessService

チャンネルアクセス権限チェックを一元化：

```go
// domain/service/channel_access_service.go
type ChannelAccessService interface {
    CanReadChannel(ctx context.Context, userID, channelID string) (bool, error)
    CanWriteChannel(ctx context.Context, userID, channelID string) (bool, error)
}
```

使用箇所:

- メッセージ作成・更新・削除
- ブックマーク・ピン留め・リアクション
- 既読状態更新

### MessageOutputAssembler

メッセージ出力の組み立てロジックを共通化：

```go
// usecase/message/output_builder.go
type MessageOutputAssembler struct {
    userRepo       repository.UserRepository
    attachmentRepo repository.AttachmentRepository
    reactionRepo   repository.MessageReactionRepository
    // ...
}

func (a *MessageOutputAssembler) AssembleMessageOutputs(
    ctx context.Context,
    messages []*entity.Message,
) ([]*MessageOutput, error) {
    // ユーザー、添付ファイル、リアクションなどを一括取得
    // 各メッセージに関連データを紐付けて出力
}
```

### Ent エンティティ変換

Ent モデル → ドメインエンティティの変換を集約：

```go
// infrastructure/utils/ent_converters.go (416行)
func ToUserEntity(u *ent.User) *entity.User
func ToChannelEntity(c *ent.Channel) *entity.Channel
func ToMessageEntity(m *ent.Message) *entity.Message
// ... 全エンティティの変換関数
```

## システムメッセージによる監査ログ

チャンネルの変更履歴を自動記録します。メッセージ一覧 API はユーザーの投稿とシステムメッセージを
`TimelineItem` として時系列に混ぜて返します。

```go
// チャンネル名変更時
systemMessage := &entity.SystemMessage{
    ChannelID: channelID,
    Kind:      entity.SystemMessageKindChannelNameChanged,
    Payload: map[string]any{
        "from": oldName,
        "to":   newName,
    },
    ActorID:   &userID,
    CreatedAt: time.Now(),
}
```

記録される変更（`entity.SystemMessageKind`）:

- `channel_name_changed` — チャンネル名変更
- `channel_description_changed` — チャンネル説明変更
- `channel_privacy_changed` — 公開設定の変更
- `member_added` / `member_joined` — メンバーの追加・参加
- `member_removed` / `member_left` — メンバーの追放・退出
- `message_pinned` — メッセージのピン留め
