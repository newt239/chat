# バックエンド移行計画：Gin → Echo & クリーンアーキテクチャ導入

## 目次

1. [現状分析](#現状分析)
2. [移行目標](#移行目標)
3. [アーキテクチャ設計](#アーキテクチャ設計)
4. [実装手順](#実装手順)
5. [注意事項](#注意事項)

## 現状分析

### プロジェクト構造

```
backend/
├── cmd/
│   ├── server/main.go       # アプリケーションエントリーポイント
│   ├── seed/main.go         # データシーディング
│   ├── seed-manual/main.go
│   └── reset/main.go
├── internal/
│   ├── domain/              # ドメインモデル（一部クリーンアーキテクチャに準拠）
│   │   ├── user.go
│   │   ├── workspace.go
│   │   ├── channel.go
│   │   ├── message.go
│   │   ├── session.go
│   │   ├── read_state.go
│   │   ├── attachment.go
│   │   ├── mention.go
│   │   ├── link.go
│   │   └── user_group.go
│   ├── usecase/             # ビジネスロジック層
│   │   ├── auth/
│   │   ├── workspace/
│   │   ├── channel/
│   │   ├── message/
│   │   ├── reaction/
│   │   ├── readstate/
│   │   ├── user_group/
│   │   └── link/
│   ├── interface/           # インターフェース層
│   │   ├── http/
│   │   │   ├── handler/    # Ginハンドラー
│   │   │   ├── middleware/ # 認証、CORS、Rate Limit
│   │   │   └── router.go
│   │   └── ws/             # WebSocket
│   └── infrastructure/      # インフラストラクチャ層
│       ├── db/             # DB接続とモデル
│       ├── repository/     # データアクセス層
│       ├── auth/           # JWT、パスワードハッシュ
│       ├── config/         # 設定管理
│       ├── logger/         # ロギング
│       ├── ogp/            # OGP取得
│       └── seed/           # シーディング
```

### 主要な機能とエンドポイント

#### 認証（/api/auth）
- `POST /register` - ユーザー登録
- `POST /login` - ログイン
- `POST /refresh` - トークンリフレッシュ
- `POST /logout` - ログアウト

#### ワークスペース（/api/workspaces）
- `GET /workspaces` - ワークスペース一覧取得
- `POST /workspaces` - ワークスペース作成
- `GET /workspaces/:id` - ワークスペース詳細取得
- `PATCH /workspaces/:id` - ワークスペース更新
- `DELETE /workspaces/:id` - ワークスペース削除
- `GET /workspaces/:id/members` - メンバー一覧
- `POST /workspaces/:id/members` - メンバー追加
- `PATCH /workspaces/:id/members/:userId` - メンバーロール更新
- `DELETE /workspaces/:id/members/:userId` - メンバー削除

#### チャンネル（/api/channels）
- `GET /workspaces/:id/channels` - チャンネル一覧
- `POST /workspaces/:id/channels` - チャンネル作成

#### メッセージ（/api/messages）
- `GET /channels/:channelId/messages` - メッセージ一覧
- `POST /channels/:channelId/messages` - メッセージ作成

#### リアクション（/api/messages/:messageId/reactions）
- `GET /reactions` - リアクション一覧
- `POST /reactions` - リアクション追加
- `DELETE /reactions/:emoji` - リアクション削除

#### 既読管理（/api/channels/:channelId）
- `GET /unread_count` - 未読数取得
- `POST /reads` - 既読状態更新

#### ユーザーグループ（/api/user-groups）
- `POST /user-groups` - グループ作成
- `GET /user-groups` - グループ一覧
- `GET /user-groups/:id` - グループ詳細
- `PATCH /user-groups/:id` - グループ更新
- `DELETE /user-groups/:id` - グループ削除
- `POST /user-groups/:id/members` - メンバー追加
- `DELETE /user-groups/:id/members` - メンバー削除
- `GET /user-groups/:id/members` - メンバー一覧

#### リンク（/api/links）
- `POST /links/fetch-ogp` - OGP情報取得

#### WebSocket
- `GET /ws?workspaceId={id}&token={jwt}` - WebSocket接続

#### その他
- `GET /healthz` - ヘルスチェック

### 現在使用している主要なライブラリ

- **Webフレームワーク**: Gin (github.com/gin-gonic/gin v1.10.0)
- **ORM**: GORM (gorm.io/gorm v1.31.0)
- **DB**: PostgreSQL (gorm.io/driver/postgres v1.6.0)
- **JWT**: golang-jwt/jwt/v5
- **WebSocket**: gorilla/websocket
- **ロガー**: uber/zap
- **パスワードハッシュ**: golang.org/x/crypto

### 現在のミドルウェア

1. **CORS**: 指定されたオリジンからのアクセスを許可
2. **認証**: JWT検証（Bearerトークン）
3. **Rate Limiting**: （実装済みだが未使用）

### 現在の課題

1. **Ginに強く依存**
   - ハンドラーが `gin.Context` に依存
   - テストがフレームワークに依存
   - ビジネスロジックとHTTP層の分離が不十分

2. **クリーンアーキテクチャの不完全な実装**
   - Domainレイヤーにリポジトリインターフェースが定義されている（良い点）
   - しかしUseCaseレイヤーが具体的な実装に依存している箇所がある
   - インフラ層の詳細がドメイン層に漏れ出している可能性

3. **依存性注入の手動管理**
   - `main.go` で全ての依存関係を手動で組み立て
   - DIコンテナ未使用

## 移行目標

### 主要目標

1. **Echoフレームワークへの移行**
   - パフォーマンス向上
   - 標準的なnet/httpとの互換性向上
   - ミドルウェアの改善

2. **完全なクリーンアーキテクチャの実装**
   - 依存性の方向を正しく保つ
   - ビジネスロジックをフレームワークから完全に分離
   - テスタビリティの向上

3. **保守性とスケーラビリティの向上**
   - レイヤー間の責任を明確化
   - 拡張性の向上
   - コードの再利用性向上

## アーキテクチャ設計

### クリーンアーキテクチャの各層

```
┌──────────────────────────────────────────────────────┐
│                   Presentation Layer                  │
│  (Echo Handlers, Middleware, WebSocket)              │
│  - HTTP request/response handling                     │
│  - Input validation                                   │
│  - Error handling and formatting                      │
└──────────────────────────────────────────────────────┘
                           ↓
┌──────────────────────────────────────────────────────┐
│                    Use Case Layer                     │
│  (Application Business Rules)                         │
│  - Orchestrate domain entities                        │
│  - Implement application-specific logic               │
│  - Define input/output DTOs                           │
└──────────────────────────────────────────────────────┘
                           ↓
┌──────────────────────────────────────────────────────┐
│                     Domain Layer                      │
│  (Enterprise Business Rules)                          │
│  - Entities                                           │
│  - Repository interfaces                              │
│  - Domain services                                    │
│  - Business logic                                     │
└──────────────────────────────────────────────────────┘
                           ↑
┌──────────────────────────────────────────────────────┐
│                 Infrastructure Layer                  │
│  (Frameworks & Drivers)                               │
│  - Database (GORM)                                    │
│  - External services (OGP, etc.)                      │
│  - Repository implementations                         │
│  - Config, Logger, etc.                               │
└──────────────────────────────────────────────────────┘
```

### 新しいディレクトリ構造

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # アプリケーションエントリーポイント
├── internal/
│   ├── domain/                  # Domain Layer（最内層）
│   │   ├── entity/             # エンティティ
│   │   │   ├── user.go
│   │   │   ├── workspace.go
│   │   │   ├── channel.go
│   │   │   ├── message.go
│   │   │   └── ...
│   │   ├── repository/         # リポジトリインターフェース
│   │   │   ├── user.go
│   │   │   ├── workspace.go
│   │   │   └── ...
│   │   ├── service/            # ドメインサービス
│   │   │   └── ...
│   │   └── errors/             # ドメインエラー
│   │       └── errors.go
│   │
│   ├── usecase/                 # Use Case Layer
│   │   ├── auth/
│   │   │   ├── interface.go    # UseCase インターフェース
│   │   │   ├── interactor.go   # UseCase 実装
│   │   │   └── dto.go          # Input/Output DTO
│   │   ├── workspace/
│   │   ├── channel/
│   │   ├── message/
│   │   └── ...
│   │
│   ├── adapter/                 # Adapter Layer（インターフェース実装）
│   │   ├── controller/         # Presentation Layer
│   │   │   ├── http/
│   │   │   │   ├── handler/   # Echoハンドラー
│   │   │   │   ├── middleware/
│   │   │   │   ├── router.go
│   │   │   │   └── presenter/ # レスポンス整形
│   │   │   └── websocket/
│   │   │       ├── hub.go
│   │   │       └── connection.go
│   │   │
│   │   └── gateway/            # Infrastructure実装
│   │       ├── persistence/    # リポジトリ実装
│   │       │   ├── user.go
│   │       │   ├── workspace.go
│   │       │   └── ...
│   │       └── external/       # 外部サービス
│   │           └── ogp.go
│   │
│   ├── infrastructure/          # Infrastructure Layer
│   │   ├── database/
│   │   │   ├── connection.go
│   │   │   ├── models.go       # GORMモデル
│   │   │   └── migration.go
│   │   ├── config/
│   │   │   └── config.go
│   │   ├── logger/
│   │   │   └── logger.go
│   │   ├── auth/
│   │   │   ├── jwt.go
│   │   │   └── password.go
│   │   └── seed/
│   │       └── seed.go
│   │
│   └── registry/                # DI Container
│       └── registry.go
│
├── pkg/                         # 共通パッケージ（プロジェクト外でも使える）
│   ├── validator/
│   └── utils/
│
└── docs/
    └── backend-migration-plan.md
```

### 依存関係ルール

```
Domain Layer（最内層）
  ↑
  │ 依存
  │
Use Case Layer
  ↑
  │ 依存
  │
Adapter Layer (Controller + Gateway)
  ↑
  │ 依存
  │
Infrastructure Layer（最外層）
```

**重要な原則:**
- 内側の層は外側の層に依存してはいけない
- 依存性逆転の原則（DIP）を適用
- インターフェースを使って依存を抽象化

## 実装手順

### Phase 0: 準備（1-2時間）

#### 0.1 Echoパッケージの追加

```bash
cd backend
go get github.com/labstack/echo/v4
go get github.com/labstack/echo/v4/middleware
```

#### 0.2 ブランチ作成

```bash
git checkout -b feature/migrate-to-echo-clean-arch
```

#### 0.3 既存コードのバックアップ確認

現在のコミットを確認し、いつでも戻れることを確認

### Phase 1: Domain Layer の再構築（3-4時間）

#### 1.1 エンティティの整理

**タスク:**
- `internal/domain/entity/` ディレクトリを作成
- 既存の `internal/domain/*.go` からエンティティを抽出
- エンティティは純粋なビジネスロジックとデータ構造のみを持つ
- ORMタグやJSONタグは削除（Infrastructure層で扱う）

**実装例:**

```go
// internal/domain/entity/user.go
package entity

import "time"

type User struct {
    ID           string
    Email        string
    PasswordHash string
    DisplayName  string
    AvatarURL    *string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// ビジネスロジックメソッド
func (u *User) IsActive() bool {
    // ビジネスルール
    return true
}
```

#### 1.2 リポジトリインターフェースの整理

**タスク:**
- `internal/domain/repository/` ディレクトリを作成
- 既存の `domain/*.go` からリポジトリインターフェースを抽出
- エンティティのみを扱うインターフェースに変更

**実装例:**

```go
// internal/domain/repository/user.go
package repository

import (
    "context"
    "github.com/example/chat/internal/domain/entity"
)

type UserRepository interface {
    FindByID(ctx context.Context, id string) (*entity.User, error)
    FindByEmail(ctx context.Context, email string) (*entity.User, error)
    FindByIDs(ctx context.Context, ids []string) ([]*entity.User, error)
    Create(ctx context.Context, user *entity.User) error
    Update(ctx context.Context, user *entity.User) error
    Delete(ctx context.Context, id string) error
}
```

**注意:**
- すべてのメソッドに `context.Context` を追加（キャンセル、タイムアウト対応）
- Domain entityのみを扱う

#### 1.3 ドメインエラーの定義

**タスク:**
- `internal/domain/errors/errors.go` を作成
- ビジネスルールに関するエラーを定義

**実装例:**

```go
// internal/domain/errors/errors.go
package errors

import "errors"

var (
    // Auth errors
    ErrInvalidCredentials = errors.New("invalid email or password")
    ErrUserAlreadyExists  = errors.New("user already exists")
    ErrInvalidToken       = errors.New("invalid or expired token")

    // Resource errors
    ErrNotFound       = errors.New("resource not found")
    ErrUnauthorized   = errors.New("unauthorized access")
    ErrForbidden      = errors.New("forbidden")

    // Validation errors
    ErrInvalidInput   = errors.New("invalid input")
)
```

### Phase 2: Use Case Layer の改善（4-5時間）

#### 2.1 UseCase インターフェースの明確化

各UseCaseディレクトリに `interface.go` を作���

**実装例:**

```go
// internal/usecase/auth/interface.go
package auth

import "context"

type UseCase interface {
    Register(ctx context.Context, input RegisterInput) (*AuthOutput, error)
    Login(ctx context.Context, input LoginInput) (*AuthOutput, error)
    RefreshToken(ctx context.Context, input RefreshTokenInput) (*AuthOutput, error)
    Logout(ctx context.Context, input LogoutInput) (*LogoutOutput, error)
}
```

#### 2.2 DTO の整理

Input/Output構造体を明確に分離

**実装例:**

```go
// internal/usecase/auth/dto.go
package auth

import "time"

// Input DTOs
type RegisterInput struct {
    Email       string
    Password    string
    DisplayName string
}

type LoginInput struct {
    Email    string
    Password string
}

// Output DTOs
type AuthOutput struct {
    AccessToken  string
    RefreshToken string
    ExpiresAt    time.Time
    User         UserInfo
}

type UserInfo struct {
    ID          string
    Email       string
    DisplayName string
    AvatarURL   *string
}
```

#### 2.3 Interactor の更新

- `context.Context` を全メソッドに追加
- Domainリポジトリインターフェースを使用
- フレームワーク依存を完全に排除

**実装例:**

```go
// internal/usecase/auth/interactor.go
package auth

import (
    "context"
    "github.com/example/chat/internal/domain/entity"
    "github.com/example/chat/internal/domain/repository"
    "github.com/example/chat/internal/domain/errors"
)

type interactor struct {
    userRepo    repository.UserRepository
    sessionRepo repository.SessionRepository
    jwtService  JWTService      // インターフェース
    passwordSvc PasswordService // インターフェース
}

func NewInteractor(
    userRepo repository.UserRepository,
    sessionRepo repository.SessionRepository,
    jwtService JWTService,
    passwordSvc PasswordService,
) UseCase {
    return &interactor{
        userRepo:    userRepo,
        sessionRepo: sessionRepo,
        jwtService:  jwtService,
        passwordSvc: passwordSvc,
    }
}

func (i *interactor) Register(ctx context.Context, input RegisterInput) (*AuthOutput, error) {
    // 実装...
}
```

#### 2.4 インフラサービスのインターフェース化

**タスク:**
- JWTService、PasswordServiceなどをインターフェース化
- UseCaseディレクトリに定義

**実装例:**

```go
// internal/usecase/auth/service.go
package auth

import "time"

type JWTService interface {
    GenerateToken(userID string, duration time.Duration) (string, error)
    VerifyToken(token string) (*TokenClaims, error)
}

type PasswordService interface {
    HashPassword(password string) (string, error)
    VerifyPassword(password, hash string) error
}

type TokenClaims struct {
    UserID string
    Email  string
}
```

### Phase 3: Infrastructure Layer の整備（3-4時間）

#### 3.1 Database層の整理

**タスク:**
- GORMモデルを `internal/infrastructure/database/models.go` に集約
- エンティティとGORMモデルの変換関数を実装

**実装例:**

```go
// internal/infrastructure/database/models.go
package database

import (
    "time"
    "github.com/example/chat/internal/domain/entity"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type User struct {
    ID           string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    Email        string    `gorm:"uniqueIndex;not null"`
    PasswordHash string    `gorm:"not null"`
    DisplayName  string    `gorm:"not null"`
    AvatarURL    *string   `gorm:"type:text"`
    CreatedAt    time.Time `gorm:"not null;autoCreateTime"`
    UpdatedAt    time.Time `gorm:"not null;autoUpdateTime"`
}

func (User) TableName() string {
    return "users"
}

// Entity → Model
func (m *User) FromEntity(e *entity.User) {
    m.ID = e.ID
    m.Email = e.Email
    m.PasswordHash = e.PasswordHash
    m.DisplayName = e.DisplayName
    m.AvatarURL = e.AvatarURL
    m.CreatedAt = e.CreatedAt
    m.UpdatedAt = e.UpdatedAt
}

// Model → Entity
func (m *User) ToEntity() *entity.User {
    return &entity.User{
        ID:           m.ID,
        Email:        m.Email,
        PasswordHash: m.PasswordHash,
        DisplayName:  m.DisplayName,
        AvatarURL:    m.AvatarURL,
        CreatedAt:    m.CreatedAt,
        UpdatedAt:    m.UpdatedAt,
    }
}
```

#### 3.2 認証サービスの実装更新

**タスク:**
- `internal/infrastructure/auth/` の各サービスを UseCase で定義したインターフェースに適合させる

**実装例:**

```go
// internal/infrastructure/auth/jwt.go
package auth

import (
    "time"
    "github.com/golang-jwt/jwt/v5"
    authuc "github.com/example/chat/internal/usecase/auth"
)

type jwtService struct {
    secret string
}

func NewJWTService(secret string) authuc.JWTService {
    return &jwtService{secret: secret}
}

func (s *jwtService) GenerateToken(userID string, duration time.Duration) (string, error) {
    // 実装
}

func (s *jwtService) VerifyToken(tokenString string) (*authuc.TokenClaims, error) {
    // 実装
}
```

### Phase 4: Adapter Layer - Gateway の実装（4-5時間）

#### 4.1 Persistence層（リポジトリ実装）の作成

**タスク:**
- `internal/adapter/gateway/persistence/` ディレクトリを作成
- 既存の `internal/infrastructure/repository/` を移行
- GORMモデルとエンティティの変換を実装

**実装例:**

```go
// internal/adapter/gateway/persistence/user.go
package persistence

import (
    "context"
    "errors"
    "github.com/example/chat/internal/domain/entity"
    "github.com/example/chat/internal/domain/repository"
    domerr "github.com/example/chat/internal/domain/errors"
    "github.com/example/chat/internal/infrastructure/database"
    "gorm.io/gorm"
)

type userRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repository.UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
    var model database.User
    if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, domerr.ErrNotFound
        }
        return nil, err
    }
    return model.ToEntity(), nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
    var model database.User
    if err := r.db.WithContext(ctx).First(&model, "email = ?", email).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil // メールが見つからない場合はnilを返す
        }
        return nil, err
    }
    return model.ToEntity(), nil
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
    model := &database.User{}
    model.FromEntity(user)

    if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
        return err
    }

    // IDなど自動生成された値を反映
    *user = *model.ToEntity()
    return nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
    model := &database.User{}
    model.FromEntity(user)
    return r.db.WithContext(ctx).Save(model).Error
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).Delete(&database.User{}, "id = ?", id).Error
}

func (r *userRepository) FindByIDs(ctx context.Context, ids []string) ([]*entity.User, error) {
    var models []database.User
    if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&models).Error; err != nil {
        return nil, err
    }

    users := make([]*entity.User, len(models))
    for i, model := range models {
        users[i] = model.ToEntity()
    }
    return users, nil
}
```

**繰り��し作業:**
すべてのリポジトリについて同様の実装を行う:
- WorkspaceRepository
- ChannelRepository
- MessageRepository
- SessionRepository
- ReadStateRepository
- UserGroupRepository
- MentionRepository
- LinkRepository

### Phase 5: Adapter Layer - Controller (Echo) の実装（5-6時間）

#### 5.1 Echoハンドラーの実装

**タスク:**
- `internal/adapter/controller/http/handler/` ディレクトリを作成
- Ginの `gin.Context` を Echoの `echo.Context` に置き換え
- ハンドラーはUseCaseインターフェースのみに依存

**実装例:**

```go
// internal/adapter/controller/http/handler/auth_handler.go
package handler

import (
    "net/http"
    "github.com/labstack/echo/v4"
    "github.com/example/chat/internal/usecase/auth"
)

type AuthHandler struct {
    authUC auth.UseCase
}

func NewAuthHandler(authUC auth.UseCase) *AuthHandler {
    return &AuthHandler{authUC: authUC}
}

// Register はユーザー登録を処理します
func (h *AuthHandler) Register(c echo.Context) error {
    var req RegisterRequest
    if err := c.Bind(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
    }

    // Validation
    if err := c.Validate(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }

    input := auth.RegisterInput{
        Email:       req.Email,
        Password:    req.Password,
        DisplayName: req.DisplayName,
    }

    output, err := h.authUC.Register(c.Request().Context(), input)
    if err != nil {
        return handleUseCaseError(err)
    }

    return c.JSON(http.StatusCreated, output)
}

// Login はユーザー認証を処理します
func (h *AuthHandler) Login(c echo.Context) error {
    var req LoginRequest
    if err := c.Bind(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
    }

    if err := c.Validate(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }

    input := auth.LoginInput{
        Email:    req.Email,
        Password: req.Password,
    }

    output, err := h.authUC.Login(c.Request().Context(), input)
    if err != nil {
        return handleUseCaseError(err)
    }

    return c.JSON(http.StatusOK, output)
}

// RefreshToken はトークンのリフレッシュを処理します
func (h *AuthHandler) RefreshToken(c echo.Context) error {
    var req RefreshTokenRequest
    if err := c.Bind(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
    }

    if err := c.Validate(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }

    input := auth.RefreshTokenInput{
        RefreshToken: req.RefreshToken,
    }

    output, err := h.authUC.RefreshToken(c.Request().Context(), input)
    if err != nil {
        return handleUseCaseError(err)
    }

    return c.JSON(http.StatusOK, output)
}

// Logout はログアウトを処理します
func (h *AuthHandler) Logout(c echo.Context) error {
    var req LogoutRequest
    if err := c.Bind(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
    }

    if err := c.Validate(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }

    // コンテキストからユーザーIDを取得
    userID := c.Get("userID").(string)

    input := auth.LogoutInput{
        UserID:       userID,
        RefreshToken: req.RefreshToken,
    }

    output, err := h.authUC.Logout(c.Request().Context(), input)
    if err != nil {
        return handleUseCaseError(err)
    }

    return c.JSON(http.StatusOK, output)
}
```

#### 5.2 エラーハンドリングヘルパー

**実装例:**

```go
// internal/adapter/controller/http/handler/error.go
package handler

import (
    "errors"
    "net/http"
    "github.com/labstack/echo/v4"
    domerr "github.com/example/chat/internal/domain/errors"
    authuc "github.com/example/chat/internal/usecase/auth"
)

func handleUseCaseError(err error) error {
    switch {
    case errors.Is(err, domerr.ErrNotFound):
        return echo.NewHTTPError(http.StatusNotFound, err.Error())
    case errors.Is(err, domerr.ErrUnauthorized):
        return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
    case errors.Is(err, domerr.ErrForbidden):
        return echo.NewHTTPError(http.StatusForbidden, err.Error())
    case errors.Is(err, domerr.ErrInvalidInput):
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    case errors.Is(err, authuc.ErrInvalidCredentials):
        return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
    case errors.Is(err, authuc.ErrUserAlreadyExists):
        return echo.NewHTTPError(http.StatusConflict, err.Error())
    case errors.Is(err, authuc.ErrInvalidToken):
        return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
    default:
        return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
    }
}
```

#### 5.3 リクエスト/レスポンスDTO

**実装例:**

```go
// internal/adapter/controller/http/handler/dto.go
package handler

// Auth DTOs
type RegisterRequest struct {
    Email       string `json:"email" validate:"required,email"`
    Password    string `json:"password" validate:"required,min=8"`
    DisplayName string `json:"displayName" validate:"required"`
}

type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

type RefreshTokenRequest struct {
    RefreshToken string `json:"refreshToken" validate:"required"`
}

type LogoutRequest struct {
    RefreshToken string `json:"refreshToken" validate:"required"`
}

// Error response
type ErrorResponse struct {
    Error string `json:"error"`
}
```

**繰り返し作業:**
すべてのハンドラーについて同様の実装を行う:
- WorkspaceHandler
- ChannelHandler
- MessageHandler
- ReadStateHandler
- ReactionHandler
- UserGroupHandler
- LinkHandler

#### 5.4 Echoミドルウェアの実装

**タスク:**
- `internal/adapter/controller/http/middleware/` ディレクトリを作成
- Gin用ミドルウェアをEcho用に書き換え

**実装例:**

```go
// internal/adapter/controller/http/middleware/auth.go
package middleware

import (
    "net/http"
    "strings"
    "github.com/labstack/echo/v4"
    authuc "github.com/example/chat/internal/usecase/auth"
)

const (
    authorizationHeader = "Authorization"
    bearerPrefix        = "Bearer "
    userIDKey           = "userID"
    userEmailKey        = "userEmail"
)

func Auth(jwtService authuc.JWTService) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            authHeader := c.Request().Header.Get(authorizationHeader)
            if authHeader == "" {
                return echo.NewHTTPError(http.StatusUnauthorized, "authorization header required")
            }

            if !strings.HasPrefix(authHeader, bearerPrefix) {
                return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
            }

            token := strings.TrimPrefix(authHeader, bearerPrefix)
            claims, err := jwtService.VerifyToken(token)
            if err != nil {
                return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
            }

            c.Set(userIDKey, claims.UserID)
            c.Set(userEmailKey, claims.Email)

            return next(c)
        }
    }
}

func GetUserID(c echo.Context) (string, bool) {
    userID, ok := c.Get(userIDKey).(string)
    return userID, ok
}

func GetUserEmail(c echo.Context) (string, bool) {
    email, ok := c.Get(userEmailKey).(string)
    return email, ok
}
```

```go
// internal/adapter/controller/http/middleware/cors.go
package middleware

import (
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
)

func CORS(allowedOrigins []string) echo.MiddlewareFunc {
    return middleware.CORSWithConfig(middleware.CORSConfig{
        AllowOrigins: allowedOrigins,
        AllowMethods: []string{
            echo.GET,
            echo.POST,
            echo.PUT,
            echo.PATCH,
            echo.DELETE,
            echo.OPTIONS,
        },
        AllowHeaders: []string{
            echo.HeaderOrigin,
            echo.HeaderContentType,
            echo.HeaderAccept,
            echo.HeaderAuthorization,
        },
        AllowCredentials: true,
    })
}
```

#### 5.5 Echoルーターの実装

**実装例:**

```go
// internal/adapter/controller/http/router.go
package http

import (
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"

    "github.com/example/chat/internal/adapter/controller/http/handler"
    custommw "github.com/example/chat/internal/adapter/controller/http/middleware"
    authuc "github.com/example/chat/internal/usecase/auth"
)

type RouterConfig struct {
    JWTService       authuc.JWTService
    AllowedOrigins   []string

    // Handlers
    AuthHandler      *handler.AuthHandler
    WorkspaceHandler *handler.WorkspaceHandler
    ChannelHandler   *handler.ChannelHandler
    MessageHandler   *handler.MessageHandler
    ReadStateHandler *handler.ReadStateHandler
    ReactionHandler  *handler.ReactionHandler
    UserGroupHandler *handler.UserGroupHandler
    LinkHandler      *handler.LinkHandler
}

func NewRouter(cfg RouterConfig) *echo.Echo {
    e := echo.New()

    // Middleware
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    e.Use(custommw.CORS(cfg.AllowedOrigins))

    // Validator
    e.Validator = NewValidator()

    // Health check
    e.GET("/healthz", func(c echo.Context) error {
        return c.String(200, "ok")
    })

    // API routes
    api := e.Group("/api")

    // Auth routes (public)
    auth := api.Group("/auth")
    {
        auth.POST("/register", cfg.AuthHandler.Register)
        auth.POST("/login", cfg.AuthHandler.Login)
        auth.POST("/refresh", cfg.AuthHandler.RefreshToken)
        auth.POST("/logout", cfg.AuthHandler.Logout)
    }

    // Protected routes
    authMw := custommw.Auth(cfg.JWTService)

    // Workspace routes
    api.GET("/workspaces", cfg.WorkspaceHandler.GetWorkspaces, authMw)
    api.POST("/workspaces", cfg.WorkspaceHandler.CreateWorkspace, authMw)
    api.GET("/workspaces/:id", cfg.WorkspaceHandler.GetWorkspace, authMw)
    api.PATCH("/workspaces/:id", cfg.WorkspaceHandler.UpdateWorkspace, authMw)
    api.DELETE("/workspaces/:id", cfg.WorkspaceHandler.DeleteWorkspace, authMw)
    api.GET("/workspaces/:id/members", cfg.WorkspaceHandler.ListMembers, authMw)
    api.POST("/workspaces/:id/members", cfg.WorkspaceHandler.AddMember, authMw)
    api.PATCH("/workspaces/:id/members/:userId", cfg.WorkspaceHandler.UpdateMemberRole, authMw)
    api.DELETE("/workspaces/:id/members/:userId", cfg.WorkspaceHandler.RemoveMember, authMw)

    // Channel routes
    api.GET("/workspaces/:id/channels", cfg.ChannelHandler.ListChannels, authMw)
    api.POST("/workspaces/:id/channels", cfg.ChannelHandler.CreateChannel, authMw)

    // Message routes
    api.GET("/channels/:channelId/messages", cfg.MessageHandler.ListMessages, authMw)
    api.POST("/channels/:channelId/messages", cfg.MessageHandler.CreateMessage, authMw)

    // Read state routes
    api.GET("/channels/:channelId/unread_count", cfg.ReadStateHandler.GetUnreadCount, authMw)
    api.POST("/channels/:channelId/reads", cfg.ReadStateHandler.UpdateReadState, authMw)

    // Reaction routes
    api.GET("/messages/:messageId/reactions", cfg.ReactionHandler.ListReactions, authMw)
    api.POST("/messages/:messageId/reactions", cfg.ReactionHandler.AddReaction, authMw)
    api.DELETE("/messages/:messageId/reactions/:emoji", cfg.ReactionHandler.RemoveReaction, authMw)

    // User group routes
    groups := api.Group("/user-groups", authMw)
    {
        groups.POST("", cfg.UserGroupHandler.CreateUserGroup)
        groups.GET("", cfg.UserGroupHandler.ListUserGroups)
        groups.GET("/:id", cfg.UserGroupHandler.GetUserGroup)
        groups.PATCH("/:id", cfg.UserGroupHandler.UpdateUserGroup)
        groups.DELETE("/:id", cfg.UserGroupHandler.DeleteUserGroup)
        groups.POST("/:id/members", cfg.UserGroupHandler.AddMember)
        groups.DELETE("/:id/members", cfg.UserGroupHandler.RemoveMember)
        groups.GET("/:id/members", cfg.UserGroupHandler.ListMembers)
    }

    // Link routes
    links := api.Group("/links", authMw)
    {
        links.POST("/fetch-ogp", cfg.LinkHandler.FetchOGP)
    }

    return e
}
```

#### 5.6 バリデーターの実装

**実装例:**

```go
// internal/adapter/controller/http/validator.go
package http

import (
    "github.com/go-playground/validator/v10"
    "github.com/labstack/echo/v4"
)

type CustomValidator struct {
    validator *validator.Validate
}

func NewValidator() *CustomValidator {
    return &CustomValidator{validator: validator.New()}
}

func (cv *CustomValidator) Validate(i interface{}) error {
    if err := cv.validator.Struct(i); err != nil {
        return echo.NewHTTPError(400, err.Error())
    }
    return nil
}
```

### Phase 6: DI Container（Registry）の実装（2-3時間）

#### 6.1 Registryパターンの実装

**タスク:**
- `internal/registry/registry.go` を作成
- すべての依存関係を一箇所で管理

**実装例:**

```go
// internal/registry/registry.go
package registry

import (
    "gorm.io/gorm"

    "github.com/example/chat/internal/adapter/controller/http"
    "github.com/example/chat/internal/adapter/controller/http/handler"
    "github.com/example/chat/internal/adapter/gateway/persistence"
    "github.com/example/chat/internal/infrastructure/auth"
    "github.com/example/chat/internal/infrastructure/config"
    authuc "github.com/example/chat/internal/usecase/auth"
    channeluc "github.com/example/chat/internal/usecase/channel"
    linkuc "github.com/example/chat/internal/usecase/link"
    messageuc "github.com/example/chat/internal/usecase/message"
    reactionuc "github.com/example/chat/internal/usecase/reaction"
    readstateuc "github.com/example/chat/internal/usecase/readstate"
    usergroupuc "github.com/example/chat/internal/usecase/user_group"
    workspaceuc "github.com/example/chat/internal/usecase/workspace"
)

type Registry struct {
    db     *gorm.DB
    config *config.Config
}

func NewRegistry(db *gorm.DB, cfg *config.Config) *Registry {
    return &Registry{
        db:     db,
        config: cfg,
    }
}

// Infrastructure Services
func (r *Registry) NewJWTService() authuc.JWTService {
    return auth.NewJWTService(r.config.JWT.Secret)
}

func (r *Registry) NewPasswordService() authuc.PasswordService {
    return auth.NewPasswordService()
}

// Repositories
func (r *Registry) NewUserRepository() persistence.UserRepository {
    return persistence.NewUserRepository(r.db)
}

func (r *Registry) NewSessionRepository() persistence.SessionRepository {
    return persistence.NewSessionRepository(r.db)
}

func (r *Registry) NewWorkspaceRepository() persistence.WorkspaceRepository {
    return persistence.NewWorkspaceRepository(r.db)
}

func (r *Registry) NewChannelRepository() persistence.ChannelRepository {
    return persistence.NewChannelRepository(r.db)
}

func (r *Registry) NewMessageRepository() persistence.MessageRepository {
    return persistence.NewMessageRepository(r.db)
}

func (r *Registry) NewReadStateRepository() persistence.ReadStateRepository {
    return persistence.NewReadStateRepository(r.db)
}

func (r *Registry) NewUserGroupRepository() persistence.UserGroupRepository {
    return persistence.NewUserGroupRepository(r.db)
}

func (r *Registry) NewMessageUserMentionRepository() persistence.MessageUserMentionRepository {
    return persistence.NewMessageUserMentionRepository(r.db)
}

func (r *Registry) NewMessageGroupMentionRepository() persistence.MessageGroupMentionRepository {
    return persistence.NewMessageGroupMentionRepository(r.db)
}

func (r *Registry) NewMessageLinkRepository() persistence.MessageLinkRepository {
    return persistence.NewMessageLinkRepository(r.db)
}

// Use Cases
func (r *Registry) NewAuthUseCase() authuc.UseCase {
    return authuc.NewInteractor(
        r.NewUserRepository(),
        r.NewSessionRepository(),
        r.NewJWTService(),
        r.NewPasswordService(),
    )
}

func (r *Registry) NewWorkspaceUseCase() workspaceuc.UseCase {
    return workspaceuc.NewInteractor(
        r.NewWorkspaceRepository(),
        r.NewUserRepository(),
    )
}

func (r *Registry) NewChannelUseCase() channeluc.UseCase {
    return channeluc.NewInteractor(
        r.NewChannelRepository(),
        r.NewWorkspaceRepository(),
    )
}

func (r *Registry) NewMessageUseCase() messageuc.UseCase {
    return messageuc.NewInteractor(
        r.NewMessageRepository(),
        r.NewChannelRepository(),
        r.NewWorkspaceRepository(),
        r.NewUserRepository(),
        r.NewUserGroupRepository(),
        r.NewMessageUserMentionRepository(),
        r.NewMessageGroupMentionRepository(),
        r.NewMessageLinkRepository(),
    )
}

func (r *Registry) NewReadStateUseCase() readstateuc.UseCase {
    return readstateuc.NewInteractor(
        r.NewReadStateRepository(),
        r.NewChannelRepository(),
        r.NewWorkspaceRepository(),
    )
}

func (r *Registry) NewReactionUseCase() reactionuc.UseCase {
    return reactionuc.NewInteractor(
        r.NewMessageRepository(),
        r.NewChannelRepository(),
        r.NewWorkspaceRepository(),
        r.NewUserRepository(),
    )
}

func (r *Registry) NewUserGroupUseCase() usergroupuc.UseCase {
    return usergroupuc.NewInteractor(
        r.NewUserGroupRepository(),
        r.NewWorkspaceRepository(),
        r.NewUserRepository(),
    )
}

func (r *Registry) NewLinkUseCase() linkuc.UseCase {
    return linkuc.NewInteractor()
}

// Handlers
func (r *Registry) NewAuthHandler() *handler.AuthHandler {
    return handler.NewAuthHandler(r.NewAuthUseCase())
}

func (r *Registry) NewWorkspaceHandler() *handler.WorkspaceHandler {
    return handler.NewWorkspaceHandler(r.NewWorkspaceUseCase())
}

func (r *Registry) NewChannelHandler() *handler.ChannelHandler {
    return handler.NewChannelHandler(r.NewChannelUseCase())
}

func (r *Registry) NewMessageHandler() *handler.MessageHandler {
    return handler.NewMessageHandler(r.NewMessageUseCase())
}

func (r *Registry) NewReadStateHandler() *handler.ReadStateHandler {
    return handler.NewReadStateHandler(r.NewReadStateUseCase())
}

func (r *Registry) NewReactionHandler() *handler.ReactionHandler {
    return handler.NewReactionHandler(r.NewReactionUseCase())
}

func (r *Registry) NewUserGroupHandler() *handler.UserGroupHandler {
    return handler.NewUserGroupHandler(r.NewUserGroupUseCase())
}

func (r *Registry) NewLinkHandler() *handler.LinkHandler {
    return handler.NewLinkHandler(r.NewLinkUseCase())
}

// Router
func (r *Registry) NewRouter() *echo.Echo {
    routerConfig := http.RouterConfig{
        JWTService:       r.NewJWTService(),
        AllowedOrigins:   r.config.CORS.AllowedOrigins,
        AuthHandler:      r.NewAuthHandler(),
        WorkspaceHandler: r.NewWorkspaceHandler(),
        ChannelHandler:   r.NewChannelHandler(),
        MessageHandler:   r.NewMessageHandler(),
        ReadStateHandler: r.NewReadStateHandler(),
        ReactionHandler:  r.NewReactionHandler(),
        UserGroupHandler: r.NewUserGroupHandler(),
        LinkHandler:      r.NewLinkHandler(),
    }

    return http.NewRouter(routerConfig)
}
```

### Phase 7: main.go の書き換えとWebSocket対応（2-3時間）

#### 7.1 main.go の更新

**実装例:**

```go
// cmd/server/main.go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "time"

    "github.com/example/chat/internal/infrastructure/config"
    "github.com/example/chat/internal/infrastructure/database"
    "github.com/example/chat/internal/infrastructure/logger"
    "github.com/example/chat/internal/infrastructure/seed"
    "github.com/example/chat/internal/registry"
    "github.com/example/chat/internal/adapter/controller/websocket"
)

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }

    // Initialize logger
    if err := logger.Init(cfg.Server.Env); err != nil {
        log.Fatalf("failed to initialize logger: %v", err)
    }
    defer logger.Sync()

    // Initialize database
    db, err := database.InitDB(cfg.Database.URL)
    if err != nil {
        log.Fatalf("failed to initialize database: %v", err)
    }

    // Auto-seed database if empty
    if err := seed.AutoSeed(db); err != nil {
        log.Fatalf("failed to auto-seed database: %v", err)
    }

    // Initialize registry (DI container)
    reg := registry.NewRegistry(db, cfg)

    // Initialize WebSocket hub
    hub := websocket.NewHub()
    go hub.Run()

    // Setup Echo router
    e := reg.NewRouter()

    // WebSocket endpoint
    jwtService := reg.NewJWTService()
    e.GET("/ws", websocket.NewHandler(hub, jwtService))

    // Start server
    addr := ":" + cfg.Server.Port
    log.Printf("Starting server on %s", addr)

    // Graceful shutdown
    go func() {
        if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
            log.Fatalf("server error: %v", err)
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt)
    <-quit

    // Graceful shutdown with 10 second timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := e.Shutdown(ctx); err != nil {
        log.Fatalf("server shutdown error: %v", err)
    }

    log.Println("Server gracefully stopped")
}
```

#### 7.2 WebSocketハンドラーのEcho対応

**実装例:**

```go
// internal/adapter/controller/websocket/handler.go
package websocket

import (
    "net/http"
    "github.com/gorilla/websocket"
    "github.com/labstack/echo/v4"
    authuc "github.com/example/chat/internal/usecase/auth"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // TODO: validate origin based on config
    },
}

func NewHandler(hub *Hub, jwtService authuc.JWTService) echo.HandlerFunc {
    return func(c echo.Context) error {
        workspaceID := c.QueryParam("workspaceId")
        if workspaceID == "" {
            return echo.NewHTTPError(http.StatusBadRequest, "workspaceId required")
        }

        // Extract and validate JWT
        token := c.QueryParam("token")
        if token == "" {
            token = c.Request().Header.Get("Sec-WebSocket-Protocol")
        }
        if token == "" {
            return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
        }

        claims, err := jwtService.VerifyToken(token)
        if err != nil {
            return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
        }

        // Upgrade connection
        conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
        if err != nil {
            return err
        }

        wsConn := NewConnection(hub, conn, claims.UserID, workspaceID)
        hub.Register(wsConn)

        go wsConn.WritePump()
        go wsConn.ReadPump()

        return nil
    }
}
```

### Phase 8: テストとデバッグ（4-6時間）

#### 8.1 ユニットテストの更新

**タスク:**
- モックの作成
- 既存テストの更新
- 新規テストの追加

**実装例:**

```go
// internal/usecase/auth/interactor_test.go
package auth_test

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"

    "github.com/example/chat/internal/domain/entity"
    "github.com/example/chat/internal/usecase/auth"
)

// Mock implementations
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
    args := m.Called(ctx, email)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*entity.User), args.Error(1)
}

// ... 他のメソッド

type MockJWTService struct {
    mock.Mock
}

func (m *MockJWTService) GenerateToken(userID string, duration time.Duration) (string, error) {
    args := m.Called(userID, duration)
    return args.String(0), args.Error(1)
}

// ... テストケース
```

#### 8.2 統合テスト

**タスク:**
- エンドポイントのテスト
- データベース統合テスト
- WebSocketテスト

#### 8.3 手動テスト

- すべてのエンドポイントの動作確認
- WebSocket接続の確認
- エラーハンドリングの確認

### Phase 9: 最終調整とドキュメント（2-3時間）

#### 9.1 go.mod の整理

```bash
go mod tidy
```

#### 9.2 不要なコードの削除

- 古いGin関連のコード削除
- 未使用のimport削除

#### 9.3 ドキュメント更新

- README.md の更新
- API仕様書の更新
- アーキテクチャ図の作成

#### 9.4 コードレビューと最適化

- コードの整形
- 命名の統一
- パフォーマンスチェック

## 注意事項

### 移行中の注意点

1. **段階的な移行**
   - 一度にすべてを書き換えない
   - フェーズごとに動作確認を行う
   - コミットを細かく分ける

2. **テストの継続実行**
   - 各フェーズ後にテストを実行
   - リグレッションを早期に発見

3. **データベースマイグレーション**
   - スキーマ変更が必要な場合は慎重に
   - バックアップを必ず取る

4. **後方互換性**
   - APIの互換性を保つ
   - フロントエンドとの連携に注意

### パフォーマンス考慮事項

1. **Context伝播**
   - すべてのレイヤーで `context.Context` を適切に扱う
   - タイムアウトとキャンセルを実装

2. **データベース接続**
   - コネクションプールの適切な設定
   - N+1問題の回避

3. **エラーハンドリング**
   - 適切なログ出力
   - ユーザーフレンドリーなエラーメッセージ

### セキュリティ考慮事項

1. **認証・認可**
   - JWTの適切な検証
   - トークンの有効期限管理

2. **入力検証**
   - すべての入力をバリデーション
   - SQLインジェクション対策

3. **CORS設定**
   - 適切なオリジン設定
   - 本番環境では厳格に設定

## 見積もり時間

| フェーズ | 作業内容 | 見積もり時間 |
|---------|---------|-------------|
| Phase 0 | 準備 | 1-2時間 |
| Phase 1 | Domain Layer 再構築 | 3-4時間 |
| Phase 2 | Use Case Layer 改善 | 4-5時間 |
| Phase 3 | Infrastructure Layer 整備 | 3-4時間 |
| Phase 4 | Adapter/Gateway 実装 | 4-5時間 |
| Phase 5 | Adapter/Controller (Echo) 実装 | 5-6時間 |
| Phase 6 | DI Container 実装 | 2-3時間 |
| Phase 7 | main.go 書き換え・WebSocket対応 | 2-3時間 |
| Phase 8 | テスト・デバッグ | 4-6時間 |
| Phase 9 | 最終調整・ドキュメント | 2-3時間 |
| **合計** | | **30-41時間** |

実際の作業では、予期しない問題や調整が発生する可能性があるため、**40-50時間**を見込むことをお勧めします。

## 次のステップ

このドキュメントを確認した後、以下の手順で進めてください:

1. ✅ このドキュメントをレビュー
2. ✅ 不明点があれば質問
3. ✅ 実装開始の承認を得る
4. 🚀 Phase 0 から順次実装開始

---

**作成日**: 2025-10-24
**バージョン**: 1.0
