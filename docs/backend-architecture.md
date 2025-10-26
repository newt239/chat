# バックエンドアーキテクチャドキュメント

## 概要

このドキュメントでは、チャットアプリケーションのバックエンドのアーキテクチャとディレクトリ構成について説明します。バックエンドは Go 言語で実装されており、クリーンアーキテクチャの原則に従って設計されています。

## 技術スタック

- **言語**: Go 1.23.0
- **Web フレームワーク**: Echo v4
- **ORM**: GORM v1.31.0
- **データベース**: PostgreSQL
- **認証**: JWT (golang-jwt/jwt/v5)
- **WebSocket**: Gorilla WebSocket
- **ストレージ**: Wasabi (S3 互換)
- **ログ**: Zap
- **テスト**: Testify

## ディレクトリ構成

```bash
backend/
├── cmd/                          # アプリケーションエントリーポイント
│   ├── server/                   # メインサーバー
│   ├── seed/                     # データベースシード
│   └── reset/                    # データベースリセット
├── internal/                     # 内部パッケージ
│   ├── interfaces/               # インターフェース層
│   │   └── handler/              # 外部インターフェース
│   │       ├── http/             # HTTPハンドラー
│   │       │   ├── handler/       # HTTPハンドラー実装
│   │       │   └── middleware/   # ミドルウェア
│   │       └── websocket/        # WebSocketハンドラー
│   ├── domain/                   # ドメイン層
│   │   ├── entity/               # エンティティ
│   │   ├── repository/              # リポジトリインターフェース
│   │   ├── service/              # ドメインサービス
│   │   ├── errors/               # ドメインエラー
│   │   └── transaction/           # トランザクション管理
│   ├── infrastructure/           # インフラストラクチャ層
│   │   ├── auth/                 # 認証関連
│   │   ├── config/               # 設定管理
│   │   ├── database/             # データベースモデル
│   │   ├── repository/           # リポジトリ実装
│   │   ├── storage/              # ストレージ関連
│   │   ├── logger/               # ログ機能
│   │   ├── observability/        # 監視・観測性
│   │   ├── notification/         # 通知機能
│   │   ├── link/                 # リンク処理
│   │   ├── mention/              # メンション機能
│   │   ├── ogp/                  # OGP処理
│   │   └── seed/                 # データシード
│   ├── registry/                 # 依存性注入コンテナ
│   ├── usecase/                  # ユースケース層
│   │   ├── auth/                 # 認証ユースケース
│   │   ├── workspace/            # ワークスペースユースケース
│   │   ├── channel/              # チャンネルユースケース
│   │   ├── channelmember/        # チャンネルメンバーユースケース
│   │   ├── message/              # メッセージユースケース
│   │   ├── attachment/           # 添付ファイルユースケース
│   │   ├── bookmark/            # ブックマークユースケース
│   │   ├── link/                 # リンクユースケース
│   │   ├── reaction/             # リアクションユースケース
│   │   ├── readstate/            # 既読状態ユースケース
│   │   └── user_group/            # ユーザーグループユースケース
│   ├── openapi/                  # OpenAPI仕様
│   └── test/                     # テスト関連
│       ├── integration/          # 統合テスト
│       └── mocks/               # モック実装
└── schema/                       # データベーススキーマ
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

### 4. 依存性注入 (Registry Pattern)

`registry`パッケージで依存性注入を管理：

```go
type Registry struct {
    db     *gorm.DB
    config *config.Config
    hub    *websocket.Hub
}

func (r *Registry) NewAuthUseCase() authuc.AuthUseCase {
    return authuc.NewAuthInteractor(
        r.NewUserRepository(),
        r.NewSessionRepository(),
        r.NewJWTService(),
        r.NewPasswordService(),
    )
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
- チャンネルメンバー管理
- 権限管理

### 4. メッセージング

- リアルタイムメッセージング（WebSocket）
- メッセージの編集・削除
- スレッド機能
- メンション機能
- リアクション機能

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
- 未読メッセージカウント
- 既読状態の更新

### 10. ユーザーグループ機能

- ユーザーグループの作成・管理
- グループメンバー管理
- グループ権限管理

### 11. 通知システム

- WebSocket ベースのリアルタイム通知
- 未読メッセージカウント
- メンション通知

## API 設計

### RESTful API

```
# ワークスペース
GET    /api/workspaces                    # ワークスペース一覧
POST   /api/workspaces                    # ワークスペース作成
GET    /api/workspaces/:id                # ワークスペース詳細
PATCH  /api/workspaces/:id                # ワークスペース更新
DELETE /api/workspaces/:id                # ワークスペース削除

# チャンネル
GET    /api/workspaces/:id/channels       # チャンネル一覧
POST   /api/workspaces/:id/channels       # チャンネル作成
GET    /api/channels/:id                  # チャンネル詳細
PATCH  /api/channels/:id                  # チャンネル更新
DELETE /api/channels/:id                  # チャンネル削除

# チャンネルメンバー
GET    /api/channels/:id/members          # チャンネルメンバー一覧
POST   /api/channels/:id/members          # チャンネルメンバー追加
DELETE /api/channels/:id/members/:userId  # チャンネルメンバー削除

# メッセージ
GET    /api/channels/:id/messages         # メッセージ一覧
POST   /api/channels/:id/messages         # メッセージ送信
PATCH  /api/messages/:id                  # メッセージ更新
DELETE /api/messages/:id                  # メッセージ削除

# 添付ファイル
POST   /api/attachments                   # ファイルアップロード
GET    /api/attachments/:id               # ファイル情報取得
DELETE /api/attachments/:id               # ファイル削除

# ブックマーク
GET    /api/bookmarks                     # ブックマーク一覧
POST   /api/bookmarks                     # ブックマーク作成
DELETE /api/bookmarks/:id                 # ブックマーク削除

# リンク
GET    /api/links                         # リンク一覧
POST   /api/links                         # リンク作成

# リアクション
GET    /api/messages/:id/reactions        # リアクション一覧
POST   /api/messages/:id/reactions        # リアクション作成
DELETE /api/messages/:id/reactions/:emoji  # リアクション削除

# 既読状態
GET    /api/read-states                   # 既読状態一覧
POST   /api/read-states                   # 既読状態更新

# ユーザーグループ
GET    /api/user-groups                   # ユーザーグループ一覧
POST   /api/user-groups                   # ユーザーグループ作成
GET    /api/user-groups/:id               # ユーザーグループ詳細
PATCH  /api/user-groups/:id               # ユーザーグループ更新
DELETE /api/user-groups/:id               # ユーザーグループ削除
```

### WebSocket

- エンドポイント: `/ws`
- JWT 認証による接続
- リアルタイムメッセージング
- 通知配信

## データベース設計

### 主要テーブル

- `users` - ユーザー情報
- `workspaces` - ワークスペース
- `workspace_members` - ワークスペースメンバー
- `channels` - チャンネル
- `channel_members` - チャンネルメンバー
- `messages` - メッセージ
- `message_reactions` - メッセージリアクション
- `sessions` - セッション管理
- `attachments` - 添付ファイル
- `bookmarks` - ブックマーク
- `links` - リンク情報
- `read_states` - 既読状態
- `user_groups` - ユーザーグループ
- `mentions` - メンション情報

### リレーション

- ユーザー ↔ ワークスペース（多対多）
- ワークスペース → チャンネル（1 対多）
- チャンネル → メッセージ（1 対多）
- メッセージ → リアクション（1 対多）
- メッセージ → 添付ファイル（1 対多）
- ユーザー → ブックマーク（1 対多）
- メッセージ → ブックマーク（1 対多）
- メッセージ → リンク（1 対多）
- ユーザー → 既読状態（1 対多）
- チャンネル → 既読状態（1 対多）
- ユーザー → ユーザーグループ（多対多）
- メッセージ → メンション（1 対多）

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
- リフレッシュトークンによるセッション管理
- パスワードのハッシュ化

### データ保護

- 入力値検証
- SQL インジェクション対策（GORM）
- CORS 設定

### 通信

- HTTPS 対応
- WebSocket の認証

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
