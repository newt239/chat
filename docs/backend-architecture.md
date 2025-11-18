# バックエンドアーキテクチャドキュメント

## 概要

このドキュメントでは、チャットアプリケーションのバックエンドのアーキテクチャとディレクトリ構成について説明します。バックエンドは Go 言語で実装されており、クリーンアーキテクチャの原則に従って設計されています。

## 技術スタック

- **言語**: Go 1.23.0
- **Web フレームワーク**: Echo v4.13.4
- **ORM**: Ent v0.14.5 (スキーマ駆動型 ORM)
- **データベース**: PostgreSQL
- **認証**: JWT (golang-jwt/jwt/v5 v5.2.0)
- **WebSocket**: Gorilla WebSocket v1.5.3
- **ストレージ**: Wasabi (S3 互換、AWS SDK v2)
- **ログ**: Zap v1.26.0
- **バリデーション**: validator/v10 v10.27.0
- **OpenAPI**: oapi-codegen (スキーマ生成・検証)

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
    │   ├── notification/         # WebSocket通知サービス
    │   ├── mention/              # メンション処理サービス
    │   ├── link/                 # リンク処理サービス
    │   ├── ogp/                  # OGPメタデータ取得
    │   ├── utils/                # ユーティリティ (Ent変換など)
    │   └── seed/                 # データシード
    ├── interfaces/               # インターフェース層 (3,379行)
    │   ├── handler/              # 外部インターフェース
    │   │   ├── http/             # HTTPハンドラー (17ファイル)
    │   │   │   ├── handler/      # ハンドラー実装
    │   │   │   └── middleware/   # 認証・検証ミドルウェア
    │   │   └── websocket/        # WebSocketハンドラー
    │   └── openapi/              # OpenAPI生成コード
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

// InterfaceRegistry: ハンドラーとルーターの構築
func (r *InterfaceRegistry) NewAuthHandler() *handler.AuthHandler {
    return handler.NewAuthHandler(r.usecaseRegistry.NewAuthUseCase())
}
```

## 主要機能

### 1. 認証・認可

- JWT ベースの認証
- アクセストークン（15 分）とリフレッシュトークン（7 日）
- セッション管理
- パスワードハッシュ化（bcrypt）

### 2. ワークスペース管理

- マルチテナント対応
- ワークスペースメンバー管理
- ロールベースアクセス制御

### 3. チャンネル管理

- パブリック・プライベートチャンネル
- DM・グループ DM（最大 9 人）
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
- メッセージ内リンクの OGP プレビュー

### 5. ファイル管理

- ファイルアップロード（Wasabi S3 互換ストレージ）
- プリサインド URL 生成
- メタデータ管理

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

- WebSocket ベースのリアルタイム通知
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

### RESTful API

```
# 認証
POST   /api/auth/register                 # ユーザー登録
POST   /api/auth/login                    # ログイン
POST   /api/auth/refresh                  # トークンリフレッシュ
POST   /api/auth/logout                   # ログアウト

# ユーザー
GET    /api/users/me                      # 現在のユーザー情報
PATCH  /api/users/me                      # ユーザー情報更新
GET    /api/users/:id                     # ユーザー詳細

# ワークスペース
GET    /api/workspaces                    # ワークスペース一覧
POST   /api/workspaces                    # ワークスペース作成
GET    /api/workspaces/:id                # ワークスペース詳細
PATCH  /api/workspaces/:id                # ワークスペース更新
DELETE /api/workspaces/:id                # ワークスペース削除
GET    /api/workspaces/:id/members        # メンバー一覧
POST   /api/workspaces/:id/members        # メンバー追加
DELETE /api/workspaces/:id/members/:userId # メンバー削除

# チャンネル
GET    /api/workspaces/:id/channels       # チャンネル一覧
POST   /api/workspaces/:id/channels       # チャンネル作成
GET    /api/channels/:id                  # チャンネル詳細
PATCH  /api/channels/:id                  # チャンネル更新
DELETE /api/channels/:id                  # チャンネル削除
GET    /api/channels/:id/members          # チャンネルメンバー一覧
POST   /api/channels/:id/members          # チャンネルメンバー追加
DELETE /api/channels/:id/members/:userId  # チャンネルメンバー削除

# DM
POST   /api/workspaces/:id/dms            # DM作成
POST   /api/workspaces/:id/group-dms      # グループDM作成

# メッセージ
GET    /api/channels/:id/messages         # メッセージ一覧
POST   /api/channels/:id/messages         # メッセージ送信
GET    /api/messages/:id                  # メッセージ詳細
PATCH  /api/messages/:id                  # メッセージ更新
DELETE /api/messages/:id                  # メッセージ削除

# スレッド
GET    /api/messages/:id/thread           # スレッド返信一覧
POST   /api/messages/:id/follow           # スレッドフォロー
DELETE /api/messages/:id/follow           # スレッドフォロー解除

# 添付ファイル
POST   /api/attachments                   # ファイルアップロード
GET    /api/attachments/:id               # ファイル情報取得
DELETE /api/attachments/:id               # ファイル削除

# ブックマーク
GET    /api/channels/:id/bookmarks        # ブックマーク一覧
POST   /api/messages/:id/bookmark         # ブックマーク作成
DELETE /api/messages/:id/bookmark         # ブックマーク削除

# ピン留め
GET    /api/channels/:id/pins             # ピン留め一覧
POST   /api/messages/:id/pin              # ピン留め作成
DELETE /api/messages/:id/pin              # ピン留め削除

# リアクション
POST   /api/messages/:id/reactions        # リアクション作成
DELETE /api/messages/:id/reactions/:emoji # リアクション削除

# 既読状態
GET    /api/channels/:id/read-state       # チャンネル既読状態取得
PUT    /api/channels/:id/read-state       # チャンネル既読状態更新
GET    /api/messages/:id/thread-read-state # スレッド既読状態取得
PUT    /api/messages/:id/thread-read-state # スレッド既読状態更新

# ユーザーグループ
GET    /api/workspaces/:id/user-groups    # ユーザーグループ一覧
POST   /api/workspaces/:id/user-groups    # ユーザーグループ作成
GET    /api/user-groups/:id               # ユーザーグループ詳細
PATCH  /api/user-groups/:id               # ユーザーグループ更新
DELETE /api/user-groups/:id               # ユーザーグループ削除
POST   /api/user-groups/:id/members       # メンバー追加
DELETE /api/user-groups/:id/members/:userId # メンバー削除

# 検索
GET    /api/search                        # 横断検索
```

### WebSocket

- エンドポイント: `/ws`
- JWT 認証による接続
- リアルタイムメッセージング
- 通知配信

## データベース設計

### 主要テーブル（Ent スキーマ: 20 テーブル）

- `user` - ユーザー情報
- `session` - セッション管理
- `workspace` - ワークスペース
- `workspace_member` - ワークスペースメンバー
- `channel` - チャンネル
- `channel_member` - チャンネルメンバー
- `channel_read_state` - チャンネル既読状態
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
- `system_message` - システムメッセージ（監査ログ）

### リレーション

- ユーザー ↔ ワークスペース（多対多: workspace_member）
- ワークスペース → チャンネル（1 対多）
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

### テストの種類

1. **ユニットテスト**

   - 各ユースケースのテスト
   - モックを使用した依存関係の分離

2. **統合テスト**

   - データベースとの統合テスト
   - WebSocket 通信のテスト

3. **モック**
   - リポジトリのモック実装
   - 外部サービスのモック

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
- アクセストークン（15 分）とリフレッシュトークン（7 日）
- リフレッシュトークンによるセッション管理
- パスワードのハッシュ化（bcrypt）
- ミドルウェアによるトークン検証

### データ保護

- リクエストバリデーション（validator/v10）
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

チャンネルの変更履歴を自動記録：

```go
// チャンネル名変更時
systemMessage := &entity.SystemMessage{
    ChannelID: channelID,
    Type:      "channel_name_changed",
    Metadata: map[string]interface{}{
        "old_name": oldName,
        "new_name": newName,
        "changed_by": userID,
    },
    CreatedAt: time.Now(),
}
```

記録される変更:

- チャンネル名変更
- チャンネル説明変更
- メンバーの追加・削除
- 権限変更
