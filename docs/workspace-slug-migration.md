# ワークスペース Slug 移行・公開設定実装計画書

## 概要

ワークスペース識別子を UUID からユーザー定義の slug 形式に変更し、公開/非公開の管理機能を追加する。また、公開ワークスペースへの自発的な参加機能と、管理者によるユーザー追加機能を実装する。

## 変更内容サマリー

1. **ワークスペース ID を UUID から slug に変更**

   - slug 形式: 英小文字・数字・ハイフン、3〜12 文字、システム全体で一意
   - 既存データは破棄（新規構築）

2. **is_public カラムの追加**

   - デフォルト: false（非公開）
   - 公開ワークスペースは誰でも参加可能（承認不要、ロールは member）

3. **公開ワークスペース一覧・参加機能**

   - すべての公開ワークスペースを表示
   - 表示情報: 名前、説明、アイコン、メンバー数、参加済みフラグ

4. **管理者によるユーザー追加機能**

   - メールアドレスで検索・追加
   - 既存のメンバー追加 API を活用

5. **UI の変更**
   - ワークスペース選択画面を削除
   - 未参加ユーザーのみ `/app` で公開ワークスペース一覧を表示

---

## 1. データベーススキーマ変更

### 1.1 Workspace テーブル

**変更内容:**

```go
// backend/ent/schema/workspace.go

type Workspace struct {
    ent.Schema
}

func (Workspace) Fields() []ent.Field {
    return []ent.Field{
        field.String("id").
            MaxLen(12).
            MinLen(3).
            NotEmpty().
            Unique().
            Immutable().
            Match(regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)).
            Comment("ワークスペースのslug識別子"),
        field.String("name").
            NotEmpty().
            MaxLen(100).
            Comment("ワークスペース名"),
        field.String("description").
            Optional().
            MaxLen(500).
            Comment("ワークスペースの説明"),
        field.String("icon_url").
            Optional().
            MaxLen(2048).
            Comment("アイコン画像URL"),
        field.Bool("is_public").
            Default(false).
            Comment("公開ワークスペースかどうか"),
        field.Time("created_at").
            Default(time.Now).
            Immutable(),
        field.Time("updated_at").
            Default(time.Now).
            UpdateDefault(time.Now),
    }
}

func (Workspace) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("created_by", User.Type).
            Ref("created_workspaces").
            Unique().
            Required(),
        edge.To("members", WorkspaceMember.Type),
        edge.To("channels", Channel.Type),
        edge.To("user_groups", UserGroup.Type),
    }
}

func (Workspace) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("id").Unique(),
        index.Fields("is_public"),
    }
}
```

**主な変更点:**

- `id` フィールドを `uuid.UUID` から `string` に変更
- slug のバリデーション（正規表現、長さ制限）を追加
- `is_public` フィールドを追加（デフォルト: false）
- slug 検索用のユニークインデックス追加
- 公開ワークスペース検索用のインデックス追加

### 1.2 WorkspaceMember テーブル

**変更内容:**

```go
// backend/ent/schema/workspace_member.go

func (WorkspaceMember) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).
            Default(uuid.New).
            Unique().
            Immutable(),
        field.String("role").
            NotEmpty().
            Default(string(entity.WorkspaceRoleMember)),
        field.Time("joined_at").
            Default(time.Now).
            Immutable(),
    }
}

func (WorkspaceMember) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("workspace", Workspace.Type).
            Ref("members").
            Unique().
            Required(),
        edge.From("user", User.Type).
            Ref("workspace_memberships").
            Unique().
            Required(),
    }
}

func (WorkspaceMember) Indexes() []ent.Index {
    return []ent.Index{
        // workspace_id と user_id の組み合わせで一意制約
        index.Edges("workspace", "user").Unique(),
    }
}
```

**主な変更点:**

- workspace エッジの参照先が string 型の id になる（Ent が自動処理）

### 1.3 関連テーブル

**Channel, UserGroup など:**

- `workspace` エッジで参照するワークスペース ID が string 型になる
- Ent が自動的に処理するため、スキーマ定義の変更は不要

### 1.4 マイグレーション戦略

**データ破棄・再構築:**

1. 開発環境のデータベースを削除
2. Ent スキーマを更新
3. `go generate ./...` で Ent コードを再生成
4. Seed データを更新して投入

---

## 2. バックエンド実装

### 2.1 ドメイン層

#### 2.1.1 エンティティ更新

```go
// backend/internal/domain/entity/workspace.go

type Workspace struct {
    ID          string    // UUID string → slug string
    Name        string
    Description *string
    IconURL     *string
    IsPublic    bool      // 新規追加
    CreatedBy   string    // User ID (UUID string)
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// slug のバリデーション
func ValidateWorkspaceSlug(slug string) error {
    if len(slug) < 3 || len(slug) > 12 {
        return errors.New("ワークスペースIDは3〜12文字である必要があります")
    }

    matched, _ := regexp.MatchString(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`, slug)
    if !matched {
        return errors.New("ワークスペースIDは英小文字、数字、ハイフンのみ使用できます")
    }

    return nil
}
```

#### 2.1.2 リポジトリインターフェース更新

```go
// backend/internal/domain/repository/workspace_repository.go

type WorkspaceRepository interface {
    // 既存メソッド（引数の型は string のまま）
    FindByID(ctx context.Context, id string) (*entity.Workspace, error)
    FindByUserID(ctx context.Context, userID string) ([]*entity.Workspace, error)
    Create(ctx context.Context, workspace *entity.Workspace) error
    Update(ctx context.Context, workspace *entity.Workspace) error
    Delete(ctx context.Context, id string) error

    // メンバー管理（既存）
    AddMember(ctx context.Context, member *entity.WorkspaceMember) error
    UpdateMemberRole(ctx context.Context, workspaceID, userID string, role entity.WorkspaceRole) error
    RemoveMember(ctx context.Context, workspaceID, userID string) error
    FindMembersByWorkspaceID(ctx context.Context, workspaceID string) ([]*entity.WorkspaceMember, error)
    FindMember(ctx context.Context, workspaceID, userID string) (*entity.WorkspaceMember, error)
    SearchMembers(ctx context.Context, workspaceID, query string, limit, offset int) ([]*entity.WorkspaceMember, int, error)

    // 新規追加メソッド
    FindAllPublic(ctx context.Context) ([]*entity.Workspace, error)
    CountMembers(ctx context.Context, workspaceID string) (int, error)
    ExistsByID(ctx context.Context, id string) (bool, error)
}
```

**新規メソッド:**

- `FindAllPublic`: すべての公開ワークスペースを取得
- `CountMembers`: ワークスペースのメンバー数をカウント
- `ExistsByID`: slug の重複チェック用

#### 2.1.3 ユーザーリポジトリインターフェース更新

```go
// backend/internal/domain/repository/user_repository.go

type UserRepository interface {
    // 既存メソッド
    FindByID(ctx context.Context, id string) (*entity.User, error)
    FindByEmail(ctx context.Context, email string) (*entity.User, error)
    // ... その他既存メソッド

    // 新規追加（既に存在する可能性あり、確認必要）
    FindByEmail(ctx context.Context, email string) (*entity.User, error)
}
```

### 2.2 インフラストラクチャ層

#### 2.2.1 リポジトリ実装更新

```go
// backend/internal/infrastructure/repository/workspace_repository.go

type workspaceRepository struct {
    client *ent.Client
}

func (r *workspaceRepository) FindByID(ctx context.Context, id string) (*entity.Workspace, error) {
    // UUID.Parse の削除、そのまま string として使用
    w, err := r.client.Workspace.Query().
        Where(workspace.ID(id)).  // string 型
        WithCreatedBy().
        Only(ctx)

    if err != nil {
        if ent.IsNotFound(err) {
            return nil, errors.New("ワークスペースが見つかりません")
        }
        return nil, err
    }

    return utils.WorkspaceToEntity(w), nil
}

func (r *workspaceRepository) Create(ctx context.Context, ws *entity.Workspace) error {
    // slug のバリデーション
    if err := entity.ValidateWorkspaceSlug(ws.ID); err != nil {
        return err
    }

    // 重複チェック
    exists, err := r.ExistsByID(ctx, ws.ID)
    if err != nil {
        return err
    }
    if exists {
        return errors.New("このワークスペースIDは既に使用されています")
    }

    createdByUUID, err := utils.ParseUUID(ws.CreatedBy, "created by user ID")
    if err != nil {
        return err
    }

    _, err = r.client.Workspace.Create().
        SetID(ws.ID).  // slug を直接設定
        SetName(ws.Name).
        SetNillableDescription(ws.Description).
        SetNillableIconURL(ws.IconURL).
        SetIsPublic(ws.IsPublic).
        SetCreatedByID(createdByUUID).
        Save(ctx)

    return err
}

func (r *workspaceRepository) FindAllPublic(ctx context.Context) ([]*entity.Workspace, error) {
    workspaces, err := r.client.Workspace.Query().
        Where(workspace.IsPublic(true)).
        WithCreatedBy().
        Order(ent.Desc(workspace.FieldCreatedAt)).
        All(ctx)

    if err != nil {
        return nil, err
    }

    result := make([]*entity.Workspace, len(workspaces))
    for i, w := range workspaces {
        result[i] = utils.WorkspaceToEntity(w)
    }

    return result, nil
}

func (r *workspaceRepository) CountMembers(ctx context.Context, workspaceID string) (int, error) {
    count, err := r.client.WorkspaceMember.Query().
        Where(workspacemember.HasWorkspaceWith(workspace.ID(workspaceID))).
        Count(ctx)

    return count, err
}

func (r *workspaceRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
    count, err := r.client.Workspace.Query().
        Where(workspace.ID(id)).
        Count(ctx)

    if err != nil {
        return false, err
    }

    return count > 0, nil
}
```

**主な変更点:**

- UUID 変換処理の削除（workspace ID 関連）
- slug バリデーションの追加
- 公開ワークスペース取得の実装
- メンバー数カウント機能の実装

#### 2.2.2 ユーティリティ更新

```go
// backend/internal/infrastructure/utils/converter.go

func WorkspaceToEntity(w *ent.Workspace) *entity.Workspace {
    if w == nil {
        return nil
    }

    var createdBy string
    if w.Edges.CreatedBy != nil {
        createdBy = w.Edges.CreatedBy.ID.String()  // User ID は UUID のまま
    }

    return &entity.Workspace{
        ID:          w.ID,  // string 型（slug）
        Name:        w.Name,
        Description: nullableString(w.Description),
        IconURL:     nullableString(w.IconURL),
        IsPublic:    w.IsPublic,
        CreatedBy:   createdBy,
        CreatedAt:   w.CreatedAt,
        UpdatedAt:   w.UpdatedAt,
    }
}
```

### 2.3 ユースケース層

#### 2.3.1 既存ユースケースの更新

```go
// backend/internal/usecase/workspace/create_workspace.go

type CreateWorkspaceInput struct {
    ID          string   // slug
    Name        string
    Description *string
    IconURL     *string
    IsPublic    bool
    UserID      string   // UUID string
}

type CreateWorkspaceUseCase struct {
    workspaceRepo repository.WorkspaceRepository
}

func (uc *CreateWorkspaceUseCase) Execute(ctx context.Context, input CreateWorkspaceInput) (*WorkspaceOutput, error) {
    // slug のバリデーション
    if err := entity.ValidateWorkspaceSlug(input.ID); err != nil {
        return nil, err
    }

    // 重複チェック
    exists, err := uc.workspaceRepo.ExistsByID(ctx, input.ID)
    if err != nil {
        return nil, err
    }
    if exists {
        return nil, errors.New("このワークスペースIDは既に使用されています")
    }

    // ワークスペース作成
    workspace := &entity.Workspace{
        ID:          input.ID,
        Name:        input.Name,
        Description: input.Description,
        IconURL:     input.IconURL,
        IsPublic:    input.IsPublic,
        CreatedBy:   input.UserID,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }

    if err := uc.workspaceRepo.Create(ctx, workspace); err != nil {
        return nil, err
    }

    // 作成者をオーナーとして追加
    member := &entity.WorkspaceMember{
        WorkspaceID: workspace.ID,
        UserID:      input.UserID,
        Role:        entity.WorkspaceRoleOwner,
        JoinedAt:    time.Now(),
    }

    if err := uc.workspaceRepo.AddMember(ctx, member); err != nil {
        return nil, err
    }

    return &WorkspaceOutput{
        ID:          workspace.ID,
        Name:        workspace.Name,
        Description: workspace.Description,
        IconURL:     workspace.IconURL,
        IsPublic:    workspace.IsPublic,
        Role:        string(entity.WorkspaceRoleOwner),
        CreatedBy:   workspace.CreatedBy,
        CreatedAt:   workspace.CreatedAt,
        UpdatedAt:   workspace.UpdatedAt,
    }, nil
}
```

**UpdateWorkspaceInput の更新:**

```go
// backend/internal/usecase/workspace/update_workspace.go

type UpdateWorkspaceInput struct {
    ID          string   // slug (変更不可)
    Name        *string
    Description *string
    IconURL     *string
    IsPublic    *bool    // 新規追加
    UserID      string
}
```

#### 2.3.2 新規ユースケース: 公開ワークスペース一覧取得

```go
// backend/internal/usecase/workspace/list_public_workspaces.go

type PublicWorkspaceOutput struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description *string   `json:"description"`
    IconURL     *string   `json:"iconUrl"`
    MemberCount int       `json:"memberCount"`
    IsJoined    bool      `json:"isJoined"`
    CreatedAt   time.Time `json:"createdAt"`
}

type ListPublicWorkspacesUseCase struct {
    workspaceRepo repository.WorkspaceRepository
}

func NewListPublicWorkspacesUseCase(workspaceRepo repository.WorkspaceRepository) *ListPublicWorkspacesUseCase {
    return &ListPublicWorkspacesUseCase{
        workspaceRepo: workspaceRepo,
    }
}

func (uc *ListPublicWorkspacesUseCase) Execute(ctx context.Context, userID string) ([]*PublicWorkspaceOutput, error) {
    // すべての公開ワークスペースを取得
    workspaces, err := uc.workspaceRepo.FindAllPublic(ctx)
    if err != nil {
        return nil, err
    }

    // ユーザーが参加しているワークスペースを取得
    joinedWorkspaces, err := uc.workspaceRepo.FindByUserID(ctx, userID)
    if err != nil {
        return nil, err
    }

    // 参加済みワークスペースのマップを作成
    joinedMap := make(map[string]bool)
    for _, w := range joinedWorkspaces {
        joinedMap[w.ID] = true
    }

    // 出力用のデータ構造に変換
    result := make([]*PublicWorkspaceOutput, len(workspaces))
    for i, w := range workspaces {
        memberCount, err := uc.workspaceRepo.CountMembers(ctx, w.ID)
        if err != nil {
            return nil, err
        }

        result[i] = &PublicWorkspaceOutput{
            ID:          w.ID,
            Name:        w.Name,
            Description: w.Description,
            IconURL:     w.IconURL,
            MemberCount: memberCount,
            IsJoined:    joinedMap[w.ID],
            CreatedAt:   w.CreatedAt,
        }
    }

    return result, nil
}
```

#### 2.3.3 新規ユースケース: 公開ワークスペースに参加

```go
// backend/internal/usecase/workspace/join_public_workspace.go

type JoinPublicWorkspaceInput struct {
    WorkspaceID string  // slug
    UserID      string  // UUID string
}

type JoinPublicWorkspaceUseCase struct {
    workspaceRepo repository.WorkspaceRepository
}

func NewJoinPublicWorkspaceUseCase(workspaceRepo repository.WorkspaceRepository) *JoinPublicWorkspaceUseCase {
    return &JoinPublicWorkspaceUseCase{
        workspaceRepo: workspaceRepo,
    }
}

func (uc *JoinPublicWorkspaceUseCase) Execute(ctx context.Context, input JoinPublicWorkspaceInput) error {
    // ワークスペースの存在確認と公開状態の確認
    workspace, err := uc.workspaceRepo.FindByID(ctx, input.WorkspaceID)
    if err != nil {
        return err
    }

    if !workspace.IsPublic {
        return errors.New("このワークスペースは公開されていません")
    }

    // 既に参加済みかチェック
    existingMember, _ := uc.workspaceRepo.FindMember(ctx, input.WorkspaceID, input.UserID)
    if existingMember != nil {
        return errors.New("既にこのワークスペースに参加しています")
    }

    // メンバーとして追加
    member := &entity.WorkspaceMember{
        WorkspaceID: input.WorkspaceID,
        UserID:      input.UserID,
        Role:        entity.WorkspaceRoleMember,
        JoinedAt:    time.Now(),
    }

    return uc.workspaceRepo.AddMember(ctx, member)
}
```

#### 2.3.4 新規ユースケース: メールアドレスでユーザーを追加

```go
// backend/internal/usecase/workspace/add_member_by_email.go

type AddMemberByEmailInput struct {
    WorkspaceID string  // slug
    Email       string
    Role        entity.WorkspaceRole
    RequestedBy string  // 追加を実行するユーザーのID
}

type AddMemberByEmailUseCase struct {
    workspaceRepo repository.WorkspaceRepository
    userRepo      repository.UserRepository
}

func NewAddMemberByEmailUseCase(
    workspaceRepo repository.WorkspaceRepository,
    userRepo repository.UserRepository,
) *AddMemberByEmailUseCase {
    return &AddMemberByEmailUseCase{
        workspaceRepo: workspaceRepo,
        userRepo:      userRepo,
    }
}

func (uc *AddMemberByEmailUseCase) Execute(ctx context.Context, input AddMemberByEmailInput) error {
    // リクエスト元ユーザーの権限チェック
    requestedByMember, err := uc.workspaceRepo.FindMember(ctx, input.WorkspaceID, input.RequestedBy)
    if err != nil {
        return errors.New("ワークスペースへのアクセス権限がありません")
    }

    if requestedByMember.Role != entity.WorkspaceRoleOwner &&
       requestedByMember.Role != entity.WorkspaceRoleAdmin {
        return errors.New("メンバーを追加する権限がありません")
    }

    // メールアドレスからユーザーを検索
    user, err := uc.userRepo.FindByEmail(ctx, input.Email)
    if err != nil {
        return errors.New("指定されたメールアドレスのユーザーが見つかりません")
    }

    // 既に参加済みかチェック
    existingMember, _ := uc.workspaceRepo.FindMember(ctx, input.WorkspaceID, user.ID)
    if existingMember != nil {
        return errors.New("このユーザーは既にワークスペースに参加しています")
    }

    // メンバーとして追加
    member := &entity.WorkspaceMember{
        WorkspaceID: input.WorkspaceID,
        UserID:      user.ID,
        Role:        input.Role,
        JoinedAt:    time.Now(),
    }

    return uc.workspaceRepo.AddMember(ctx, member)
}
```

#### 2.3.5 DTO 更新

```go
// backend/internal/usecase/workspace/dto.go

type WorkspaceOutput struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description *string   `json:"description"`
    IconURL     *string   `json:"iconUrl"`
    IsPublic    bool      `json:"isPublic"`     // 新規追加
    Role        string    `json:"role"`
    CreatedBy   string    `json:"createdBy"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}
```

### 2.4 ハンドラー層

#### 2.4.1 既存ハンドラーの更新

```go
// backend/internal/interfaces/handler/http/handler/workspace_handler.go

type CreateWorkspaceRequest struct {
    ID          string  `json:"id" validate:"required,min=3,max=12"`  // slug
    Name        string  `json:"name" validate:"required,max=100"`
    Description *string `json:"description" validate:"omitempty,max=500"`
    IconURL     *string `json:"iconUrl" validate:"omitempty,url,max=2048"`
    IsPublic    bool    `json:"isPublic"`  // 新規追加、デフォルト false
}

func (h *WorkspaceHandler) CreateWorkspace(c echo.Context) error {
    var req CreateWorkspaceRequest
    if err := c.Bind(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }

    if err := c.Validate(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }

    userID := c.Get("user_id").(string)

    output, err := h.createWorkspaceUC.Execute(c.Request().Context(), workspace.CreateWorkspaceInput{
        ID:          req.ID,
        Name:        req.Name,
        Description: req.Description,
        IconURL:     req.IconURL,
        IsPublic:    req.IsPublic,
        UserID:      userID,
    })

    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
    }

    return c.JSON(http.StatusCreated, output)
}

type UpdateWorkspaceRequest struct {
    Name        *string `json:"name" validate:"omitempty,max=100"`
    Description *string `json:"description" validate:"omitempty,max=500"`
    IconURL     *string `json:"iconUrl" validate:"omitempty,url,max=2048"`
    IsPublic    *bool   `json:"isPublic"`  // 新規追加
}

func (h *WorkspaceHandler) UpdateWorkspace(c echo.Context) error {
    workspaceID := c.Param("id")  // slug
    userID := c.Get("user_id").(string)

    var req UpdateWorkspaceRequest
    if err := c.Bind(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }

    if err := c.Validate(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }

    output, err := h.updateWorkspaceUC.Execute(c.Request().Context(), workspace.UpdateWorkspaceInput{
        ID:          workspaceID,
        Name:        req.Name,
        Description: req.Description,
        IconURL:     req.IconURL,
        IsPublic:    req.IsPublic,
        UserID:      userID,
    })

    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
    }

    return c.JSON(http.StatusOK, output)
}
```

#### 2.4.2 新規ハンドラー

```go
// backend/internal/interfaces/handler/http/handler/workspace_handler.go に追加

// 公開ワークスペース一覧取得
func (h *WorkspaceHandler) ListPublicWorkspaces(c echo.Context) error {
    userID := c.Get("user_id").(string)

    workspaces, err := h.listPublicWorkspacesUC.Execute(c.Request().Context(), userID)
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
    }

    return c.JSON(http.StatusOK, map[string]interface{}{
        "workspaces": workspaces,
    })
}

// 公開ワークスペースに参加
func (h *WorkspaceHandler) JoinPublicWorkspace(c echo.Context) error {
    workspaceID := c.Param("id")  // slug
    userID := c.Get("user_id").(string)

    err := h.joinPublicWorkspaceUC.Execute(c.Request().Context(), workspace.JoinPublicWorkspaceInput{
        WorkspaceID: workspaceID,
        UserID:      userID,
    })

    if err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }

    return c.JSON(http.StatusOK, map[string]string{
        "message": "ワークスペースに参加しました",
    })
}

// メールアドレスでメンバー追加
type AddMemberByEmailRequest struct {
    Email string `json:"email" validate:"required,email"`
    Role  string `json:"role" validate:"required,oneof=admin member guest"`
}

func (h *WorkspaceHandler) AddMemberByEmail(c echo.Context) error {
    workspaceID := c.Param("id")  // slug
    userID := c.Get("user_id").(string)

    var req AddMemberByEmailRequest
    if err := c.Bind(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }

    if err := c.Validate(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }

    err := h.addMemberByEmailUC.Execute(c.Request().Context(), workspace.AddMemberByEmailInput{
        WorkspaceID: workspaceID,
        Email:       req.Email,
        Role:        entity.WorkspaceRole(req.Role),
        RequestedBy: userID,
    })

    if err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }

    return c.JSON(http.StatusCreated, map[string]string{
        "message": "メンバーを追加しました",
    })
}
```

#### 2.4.3 ルーティング更新

```go
// backend/internal/interfaces/handler/http/router.go

func SetupRoutes(e *echo.Echo, cfg RouteConfig, authMw echo.MiddlewareFunc) {
    api := e.Group("/api")

    // ワークスペース関連
    api.GET("/workspaces", cfg.WorkspaceHandler.GetWorkspaces, authMw)
    api.GET("/workspaces/public", cfg.WorkspaceHandler.ListPublicWorkspaces, authMw)  // 新規
    api.POST("/workspaces", cfg.WorkspaceHandler.CreateWorkspace, authMw)
    api.GET("/workspaces/:id", cfg.WorkspaceHandler.GetWorkspace, authMw)
    api.PATCH("/workspaces/:id", cfg.WorkspaceHandler.UpdateWorkspace, authMw)
    api.DELETE("/workspaces/:id", cfg.WorkspaceHandler.DeleteWorkspace, authMw)

    // ワークスペースメンバー管理
    api.GET("/workspaces/:id/members", cfg.WorkspaceHandler.ListMembers, authMw)
    api.POST("/workspaces/:id/members", cfg.WorkspaceHandler.AddMemberByEmail, authMw)  // 更新
    api.POST("/workspaces/:id/join", cfg.WorkspaceHandler.JoinPublicWorkspace, authMw)  // 新規
    api.PATCH("/workspaces/:id/members/:userId", cfg.WorkspaceHandler.UpdateMemberRole, authMw)
    api.DELETE("/workspaces/:id/members/:userId", cfg.WorkspaceHandler.RemoveMember, authMw)

    // ... その他のルート
}
```

### 2.5 依存性注入の更新

```go
// backend/cmd/server/main.go または DI設定ファイル

// 新規ユースケースの追加
listPublicWorkspacesUC := workspace.NewListPublicWorkspacesUseCase(workspaceRepo)
joinPublicWorkspaceUC := workspace.NewJoinPublicWorkspaceUseCase(workspaceRepo)
addMemberByEmailUC := workspace.NewAddMemberByEmailUseCase(workspaceRepo, userRepo)

// ハンドラーに注入
workspaceHandler := handler.NewWorkspaceHandler(
    getWorkspacesByUserIDUC,
    getWorkspaceUC,
    createWorkspaceUC,
    updateWorkspaceUC,
    deleteWorkspaceUC,
    listMembersUC,
    addMemberUC,
    updateMemberRoleUC,
    removeMemberUC,
    listPublicWorkspacesUC,      // 新規
    joinPublicWorkspaceUC,        // 新規
    addMemberByEmailUC,           // 新規
)
```

---

## 3. フロントエンド実装

### 3.1 型定義の更新

```typescript
// frontend/src/features/workspace/types.ts

import type { components } from "@/lib/api/schema";

export type WorkspaceSummary = components["schemas"]["Workspace"];
export type PublicWorkspace = components["schemas"]["PublicWorkspace"];

// 手動定義（OpenAPI生成後は不要）
export type WorkspaceDetail = {
  id: string; // slug
  name: string;
  description?: string | null;
  iconUrl?: string | null;
  isPublic: boolean;
  role: "owner" | "admin" | "member" | "guest";
  createdBy: string;
  createdAt: string;
  updatedAt: string;
};

export type PublicWorkspaceItem = {
  id: string; // slug
  name: string;
  description?: string | null;
  iconUrl?: string | null;
  memberCount: number;
  isJoined: boolean;
  createdAt: string;
};
```

### 3.2 API フック

#### 3.2.1 既存フックの更新

```typescript
// frontend/src/features/workspace/hooks/useWorkspace.ts

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api/client";

// ワークスペース作成
export function useCreateWorkspace() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: {
      id: string; // slug
      name: string;
      description?: string;
      iconUrl?: string;
      isPublic?: boolean;
    }) => {
      const { data: response, error } = await api.POST("/api/workspaces", {
        body: {
          id: data.id,
          name: data.name,
          description: data.description,
          iconUrl: data.iconUrl,
          isPublic: data.isPublic ?? false,
        },
      });

      if (error || !response) {
        throw new Error(error?.error ?? "ワークスペースの作成に失敗しました");
      }

      return response;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["workspaces"] });
    },
  });
}

// ワークスペース更新
export function useUpdateWorkspace(workspaceId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: {
      name?: string;
      description?: string;
      iconUrl?: string;
      isPublic?: boolean;
    }) => {
      const { data: response, error } = await api.PATCH(
        "/api/workspaces/{id}",
        {
          params: { path: { id: workspaceId } },
          body: data,
        }
      );

      if (error || !response) {
        throw new Error(error?.error ?? "ワークスペースの更新に失敗しました");
      }

      return response;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["workspaces"] });
      queryClient.invalidateQueries({ queryKey: ["workspace", workspaceId] });
    },
  });
}
```

#### 3.2.2 新規フック

```typescript
// frontend/src/features/workspace/hooks/usePublicWorkspaces.ts

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api/client";
import type { PublicWorkspaceItem } from "../types";

// 公開ワークスペース一覧取得
export function usePublicWorkspaces() {
  return useQuery({
    queryKey: ["public-workspaces"],
    queryFn: async () => {
      const { data, error } = await api.GET("/api/workspaces/public", {});

      if (error || !data) {
        throw new Error(
          error?.error ?? "公開ワークスペースの取得に失敗しました"
        );
      }

      return data.workspaces as PublicWorkspaceItem[];
    },
  });
}

// 公開ワークスペースに参加
export function useJoinWorkspace() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (workspaceId: string) => {
      const { data, error } = await api.POST("/api/workspaces/{id}/join", {
        params: { path: { id: workspaceId } },
      });

      if (error || !data) {
        throw new Error(error?.error ?? "ワークスペースへの参加に失敗しました");
      }

      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["workspaces"] });
      queryClient.invalidateQueries({ queryKey: ["public-workspaces"] });
    },
  });
}

// メールアドレスでメンバー追加
export function useAddMemberByEmail(workspaceId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: { email: string; role: string }) => {
      const { data: response, error } = await api.POST(
        "/api/workspaces/{id}/members",
        {
          params: { path: { id: workspaceId } },
          body: data,
        }
      );

      if (error || !response) {
        throw new Error(error?.error ?? "メンバーの追加に失敗しました");
      }

      return response;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["workspace-members", workspaceId],
      });
    },
  });
}
```

### 3.3 状態管理の更新

```typescript
// frontend/src/providers/store/workspace.ts

import { atom } from "jotai";
import { atomWithStorage } from "jotai/utils";

type WorkspaceStorage = {
  currentWorkspaceId: string | null; // slug
};

const workspaceStorageAtom = atomWithStorage<WorkspaceStorage>(
  "workspace-storage",
  { currentWorkspaceId: null }
);

export const currentWorkspaceIdAtom = atom<string | null>(
  (get) => get(workspaceStorageAtom).currentWorkspaceId
);

export const currentChannelIdAtom = atom<string | null>(null);

export const setCurrentWorkspaceAtom = atom(
  null,
  (_get, set, workspaceId: string) => {
    // slug
    set(workspaceStorageAtom, { currentWorkspaceId: workspaceId });
    set(currentChannelIdAtom, null);
  }
);

export const setCurrentChannelAtom = atom(
  null,
  (_get, set, channelId: string) => {
    set(currentChannelIdAtom, channelId);
  }
);
```

### 3.4 コンポーネント実装

#### 3.4.1 ワークスペース作成フォームの更新

```typescript
// frontend/src/features/workspace/components/CreateWorkspaceForm.tsx

import { useState } from "react";
import { useCreateWorkspace } from "../hooks/useWorkspace";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";

type CreateWorkspaceFormProps = {
  onSuccess?: () => void;
  onCancel?: () => void;
};

export const CreateWorkspaceForm = ({
  onSuccess,
  onCancel,
}: CreateWorkspaceFormProps) => {
  const [id, setId] = useState("");
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [isPublic, setIsPublic] = useState(false);
  const [slugError, setSlugError] = useState("");

  const createWorkspace = useCreateWorkspace();

  // slug のバリデーション
  const validateSlug = (value: string) => {
    if (value.length < 3 || value.length > 12) {
      setSlugError("3〜12文字で入力してください");
      return false;
    }

    if (!/^[a-z0-9][a-z0-9-]*[a-z0-9]$/.test(value)) {
      setSlugError("英小文字、数字、ハイフンのみ使用できます");
      return false;
    }

    setSlugError("");
    return true;
  };

  const handleSlugChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value.toLowerCase();
    setId(value);
    if (value) {
      validateSlug(value);
    } else {
      setSlugError("");
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateSlug(id)) {
      return;
    }

    if (!name.trim()) {
      return;
    }

    try {
      await createWorkspace.mutateAsync({
        id,
        name,
        description: description || undefined,
        isPublic,
      });

      onSuccess?.();
    } catch (error) {
      console.error("ワークスペースの作成に失敗しました:", error);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div>
        <Label htmlFor="workspace-id">ワークスペースID</Label>
        <Input
          id="workspace-id"
          type="text"
          value={id}
          onChange={handleSlugChange}
          placeholder="my-workspace"
          required
          maxLength={12}
          pattern="[a-z0-9][a-z0-9-]*[a-z0-9]"
        />
        {slugError && <p className="text-sm text-red-500 mt-1">{slugError}</p>}
        <p className="text-sm text-gray-500 mt-1">
          英小文字、数字、ハイフンのみ使用可能（3〜12文字）
        </p>
      </div>

      <div>
        <Label htmlFor="workspace-name">ワークスペース名</Label>
        <Input
          id="workspace-name"
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="My Workspace"
          required
          maxLength={100}
        />
      </div>

      <div>
        <Label htmlFor="workspace-description">説明（任意）</Label>
        <Textarea
          id="workspace-description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="ワークスペースの説明を入力..."
          maxLength={500}
        />
      </div>

      <div className="flex items-center space-x-2">
        <Checkbox
          id="workspace-public"
          checked={isPublic}
          onCheckedChange={(checked) => setIsPublic(checked === true)}
        />
        <Label htmlFor="workspace-public" className="cursor-pointer">
          公開ワークスペースとして作成
        </Label>
      </div>

      <div className="flex justify-end space-x-2">
        {onCancel && (
          <Button type="button" variant="outline" onClick={onCancel}>
            キャンセル
          </Button>
        )}
        <Button
          type="submit"
          disabled={createWorkspace.isPending || !!slugError}
        >
          作成
        </Button>
      </div>
    </form>
  );
};
```

#### 3.4.2 公開ワークスペース一覧コンポーネント

```typescript
// frontend/src/features/workspace/components/PublicWorkspaceList.tsx

import {
  usePublicWorkspaces,
  useJoinWorkspace,
} from "../hooks/usePublicWorkspaces";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Users } from "lucide-react";

export const PublicWorkspaceList = () => {
  const { data: workspaces, isLoading, error } = usePublicWorkspaces();
  const joinWorkspace = useJoinWorkspace();

  const handleJoin = async (workspaceId: string) => {
    try {
      await joinWorkspace.mutateAsync(workspaceId);
    } catch (error) {
      console.error("ワークスペースへの参加に失敗しました:", error);
    }
  };

  if (isLoading) {
    return <div className="text-center py-8">読み込み中...</div>;
  }

  if (error) {
    return (
      <div className="text-center py-8 text-red-500">
        公開ワークスペースの取得に失敗しました
      </div>
    );
  }

  if (!workspaces || workspaces.length === 0) {
    return (
      <div className="text-center py-8 text-gray-500">
        公開ワークスペースがありません
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {workspaces.map((workspace) => (
        <Card key={workspace.id} className="flex flex-col">
          <CardHeader>
            <div className="flex items-start justify-between">
              <div className="flex items-center space-x-3">
                {workspace.iconUrl ? (
                  <img
                    src={workspace.iconUrl}
                    alt={workspace.name}
                    className="w-12 h-12 rounded-lg object-cover"
                  />
                ) : (
                  <div className="w-12 h-12 rounded-lg bg-gray-200 flex items-center justify-center">
                    <span className="text-xl font-bold text-gray-600">
                      {workspace.name.charAt(0).toUpperCase()}
                    </span>
                  </div>
                )}
                <div>
                  <CardTitle className="text-lg">{workspace.name}</CardTitle>
                  <p className="text-sm text-gray-500">@{workspace.id}</p>
                </div>
              </div>
              {workspace.isJoined && (
                <Badge variant="secondary">参加済み</Badge>
              )}
            </div>
          </CardHeader>

          <CardContent className="flex-1 flex flex-col justify-between">
            <div>
              {workspace.description && (
                <p className="text-sm text-gray-600 mb-4">
                  {workspace.description}
                </p>
              )}

              <div className="flex items-center text-sm text-gray-500">
                <Users className="w-4 h-4 mr-1" />
                {workspace.memberCount} メンバー
              </div>
            </div>

            {!workspace.isJoined && (
              <Button
                onClick={() => handleJoin(workspace.id)}
                disabled={joinWorkspace.isPending}
                className="w-full mt-4"
              >
                参加する
              </Button>
            )}
          </CardContent>
        </Card>
      ))}
    </div>
  );
};
```

#### 3.4.3 メールアドレスでメンバー追加コンポーネント

```typescript
// frontend/src/features/workspace/components/AddMemberByEmailForm.tsx

import { useState } from "react";
import { useAddMemberByEmail } from "../hooks/usePublicWorkspaces";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

type AddMemberByEmailFormProps = {
  workspaceId: string;
  onSuccess?: () => void;
  onCancel?: () => void;
};

export const AddMemberByEmailForm = ({
  workspaceId,
  onSuccess,
  onCancel,
}: AddMemberByEmailFormProps) => {
  const [email, setEmail] = useState("");
  const [role, setRole] = useState("member");

  const addMember = useAddMemberByEmail(workspaceId);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!email.trim()) {
      return;
    }

    try {
      await addMember.mutateAsync({ email, role });
      setEmail("");
      setRole("member");
      onSuccess?.();
    } catch (error) {
      console.error("メンバーの追加に失敗しました:", error);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div>
        <Label htmlFor="member-email">メールアドレス</Label>
        <Input
          id="member-email"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="user@example.com"
          required
        />
      </div>

      <div>
        <Label htmlFor="member-role">ロール</Label>
        <Select value={role} onValueChange={setRole}>
          <SelectTrigger id="member-role">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="admin">管理者</SelectItem>
            <SelectItem value="member">メンバー</SelectItem>
            <SelectItem value="guest">ゲスト</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="flex justify-end space-x-2">
        {onCancel && (
          <Button type="button" variant="outline" onClick={onCancel}>
            キャンセル
          </Button>
        )}
        <Button type="submit" disabled={addMember.isPending}>
          追加
        </Button>
      </div>
    </form>
  );
};
```

### 3.5 ルーティングの更新

#### 3.5.1 ルート定義の更新

```typescript
// frontend/src/routes/app.tsx

import { createFileRoute, redirect } from "@tanstack/react-router";
import { PublicWorkspaceList } from "@/features/workspace/components/PublicWorkspaceList";
import { CreateWorkspaceForm } from "@/features/workspace/components/CreateWorkspaceForm";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";

export const Route = createFileRoute("/app")({
  beforeLoad: async ({ context }) => {
    // 認証チェック
    if (!context.auth.isAuthenticated) {
      throw redirect({ to: "/login" });
    }

    // ワークスペース一覧を取得
    const workspaces = await context.queryClient.ensureQueryData({
      queryKey: ["workspaces"],
      queryFn: async () => {
        const { data } = await api.GET("/api/workspaces", {});
        return data?.workspaces ?? [];
      },
    });

    // 参加中のワークスペースがある場合は最初のものへリダイレクト
    if (workspaces.length > 0) {
      const firstWorkspace = workspaces[0];
      throw redirect({
        to: "/app/$workspaceId",
        params: { workspaceId: firstWorkspace.id },
      });
    }
  },
  component: PublicWorkspaceListPage,
});

function PublicWorkspaceListPage() {
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);

  return (
    <div className="container mx-auto py-8">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-3xl font-bold">公開ワークスペース</h1>
          <p className="text-gray-600 mt-1">
            参加するワークスペースを選択するか、新しく作成してください
          </p>
        </div>

        <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
          <DialogTrigger asChild>
            <Button>ワークスペースを作成</Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>ワークスペースを作成</DialogTitle>
            </DialogHeader>
            <CreateWorkspaceForm
              onSuccess={() => setIsCreateDialogOpen(false)}
              onCancel={() => setIsCreateDialogOpen(false)}
            />
          </DialogContent>
        </Dialog>
      </div>

      <PublicWorkspaceList />
    </div>
  );
}
```

#### 3.5.2 ワークスペースルートの更新

```typescript
// frontend/src/routes/app.$workspaceId.tsx

import { createFileRoute, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/app/$workspaceId")({
  beforeLoad: async ({ params, context }) => {
    // 認証チェック
    if (!context.auth.isAuthenticated) {
      throw redirect({ to: "/login" });
    }

    // ワークスペースIDがslug形式であることを検証
    const { workspaceId } = params;
    if (!/^[a-z0-9][a-z0-9-]*[a-z0-9]$/.test(workspaceId)) {
      throw redirect({ to: "/app" });
    }

    // ワークスペースへのアクセス権限をチェック
    try {
      await context.queryClient.ensureQueryData({
        queryKey: ["workspace", workspaceId],
        queryFn: async () => {
          const { data, error } = await api.GET("/api/workspaces/{id}", {
            params: { path: { id: workspaceId } },
          });

          if (error || !data) {
            throw new Error("ワークスペースが見つかりません");
          }

          return data;
        },
      });
    } catch {
      // アクセス権限がない場合は /app へリダイレクト
      throw redirect({ to: "/app" });
    }
  },
  component: WorkspaceLayout,
});

// ... WorkspaceLayout コンポーネント
```

### 3.6 WorkspaceSelection コンポーネントの削除

```bash
# 削除するファイル
frontend/src/features/workspace/components/WorkspaceSelection.tsx
```

**影響箇所の修正:**

- このコンポーネントを使用している箇所を削除または修正
- ワークスペース選択ロジックは `/app` ルートの `beforeLoad` で処理

---

## 4. Seed データの更新

### 4.1 Seed スクリプトの更新

```go
// backend/cmd/seed/main.go

func seedWorkspaces(ctx context.Context, client *ent.Client, users []*ent.User) ([]*ent.Workspace, error) {
    workspaces := []struct {
        ID          string
        Name        string
        Description string
        IsPublic    bool
        CreatedBy   *ent.User
    }{
        {
            ID:          "general",
            Name:        "General",
            Description: "一般的なディスカッション用のワークスペース",
            IsPublic:    true,
            CreatedBy:   users[0], // alice
        },
        {
            ID:          "engineering",
            Name:        "Engineering",
            Description: "エンジニアリングチーム用のワークスペース",
            IsPublic:    false,
            CreatedBy:   users[1], // bob
        },
        {
            ID:          "marketing",
            Name:        "Marketing",
            Description: "マーケティングチーム用のワークスペース",
            IsPublic:    true,
            CreatedBy:   users[2], // charlie
        },
    }

    created := make([]*ent.Workspace, 0, len(workspaces))

    for _, ws := range workspaces {
        w, err := client.Workspace.Create().
            SetID(ws.ID).
            SetName(ws.Name).
            SetDescription(ws.Description).
            SetIsPublic(ws.IsPublic).
            SetCreatedBy(ws.CreatedBy).
            Save(ctx)

        if err != nil {
            return nil, fmt.Errorf("failed to create workspace %s: %w", ws.Name, err)
        }

        // 作成者をオーナーとして追加
        _, err = client.WorkspaceMember.Create().
            SetWorkspace(w).
            SetUser(ws.CreatedBy).
            SetRole(string(entity.WorkspaceRoleOwner)).
            Save(ctx)

        if err != nil {
            return nil, fmt.Errorf("failed to add owner to workspace %s: %w", ws.Name, err)
        }

        created = append(created, w)
        fmt.Printf("Created workspace: %s (%s)\n", w.Name, w.ID)
    }

    return created, nil
}

func seedWorkspaceMembers(ctx context.Context, client *ent.Client, workspaces []*ent.Workspace, users []*ent.User) error {
    // General ワークスペース（公開）に全員参加
    generalWS := workspaces[0]
    for i, user := range users {
        if i == 0 {
            continue // alice は既にオーナー
        }

        _, err := client.WorkspaceMember.Create().
            SetWorkspace(generalWS).
            SetUser(user).
            SetRole(string(entity.WorkspaceRoleMember)).
            Save(ctx)

        if err != nil {
            return fmt.Errorf("failed to add member to general workspace: %w", err)
        }
    }

    // Engineering ワークスペース（非公開）に alice と bob のみ
    engineeringWS := workspaces[1]
    _, err := client.WorkspaceMember.Create().
        SetWorkspace(engineeringWS).
        SetUser(users[0]). // alice
        SetRole(string(entity.WorkspaceRoleAdmin)).
        Save(ctx)

    if err != nil {
        return fmt.Errorf("failed to add alice to engineering workspace: %w", err)
    }

    // Marketing ワークスペース（公開）に charlie と alice
    marketingWS := workspaces[2]
    _, err = client.WorkspaceMember.Create().
        SetWorkspace(marketingWS).
        SetUser(users[0]). // alice
        SetRole(string(entity.WorkspaceRoleMember)).
        Save(ctx)

    if err != nil {
        return fmt.Errorf("failed to add alice to marketing workspace: %w", err)
    }

    fmt.Println("Seeded workspace members")
    return nil
}
```

---

## 5. 実装順序

### Phase 1: バックエンド - スキーマ変更とマイグレーション

1. Ent スキーマの更新（Workspace, WorkspaceMember）
2. `go generate ./...` でコード生成
3. データベースのリセットと再構築
4. Seed データの更新と投入
5. 動作確認

### Phase 2: バックエンド - ドメイン・リポジトリ層

1. エンティティの更新（Workspace, バリデーション関数）
2. リポジトリインターフェースの更新
3. リポジトリ実装の更新
4. ユーティリティ関数の更新
5. ユニットテスト

### Phase 3: バックエンド - ユースケース層

1. 既存ユースケースの更新（Create, Update）
2. 新規ユースケースの実装
   - ListPublicWorkspaces
   - JoinPublicWorkspace
   - AddMemberByEmail
3. ユニットテスト

### Phase 4: バックエンド - ハンドラー・ルーティング

1. 既存ハンドラーの更新
2. 新規ハンドラーの実装
3. ルーティングの更新
4. 依存性注入の更新
5. 統合テスト

### Phase 5: フロントエンド - API スキーマ生成

1. `pnpm run generate:api` でスキーマ更新
2. 型定義の確認

### Phase 6: フロントエンド - フック・状態管理

1. 既存フックの更新
2. 新規フックの実装
3. 状態管理の更新（必要に応じて）

### Phase 7: フロントエンド - コンポーネント

1. CreateWorkspaceForm の更新
2. PublicWorkspaceList の実装
3. AddMemberByEmailForm の実装
4. WorkspaceSelection の削除

### Phase 8: フロントエンド - ルーティング

1. `/app` ルートの更新
2. `/app/$workspaceId` ルートの更新
3. その他影響を受けるルートの修正

### Phase 9: テストと動作確認

1. バックエンドのテスト実行
2. フロントエンドの型チェック・Lint
3. E2E テスト（手動）
   - ワークスペース作成（slug 入力）
   - 公開ワークスペース一覧表示
   - ワークスペース参加
   - メールアドレスでメンバー追加
4. エラーハンドリングの確認

### Phase 10: ドキュメント更新

1. API ドキュメントの更新
2. README の更新（必要に応じて）

---

## 6. 考慮事項とリスク

### 6.1 破壊的変更

- **既存データの完全削除**: UUID 形式のワークスペース ID がすべて失われる
- **URL の変更**: `/app/{uuid}` → `/app/{slug}`
- **ローカルストレージ**: 保存されているワークスペース ID が無効になる

### 6.2 バリデーション

- slug の重複チェック（リポジトリ層とデータベース層の両方で）
- slug の形式チェック（正規表現）
- 予約語のチェック（必要に応じて）

### 6.3 パフォーマンス

- 公開ワークスペース一覧のクエリ最適化
- メンバー数のカウントクエリ最適化（N+1 問題の回避）
- インデックスの適切な設定

### 6.4 セキュリティ

- slug の推測可能性（公開ワークスペースのみ問題）
- 非公開ワークスペースへのアクセス制御
- メンバー追加時の権限チェック

### 6.5 UX

- slug 入力時のリアルタイムバリデーション
- 重複チェックのフィードバック
- エラーメッセージの分かりやすさ

---

## 7. テストケース

### 7.1 バックエンド

#### ユニットテスト

- slug のバリデーション（正常系・異常系）
- 公開ワークスペース一覧取得
- ワークスペース参加（公開・非公開）
- メンバー追加（権限チェック）
- slug 重複チェック

#### 統合テスト

- ワークスペース作成 API（slug 指定）
- 公開ワークスペース一覧 API
- ワークスペース参加 API
- メールアドレスでメンバー追加 API

### 7.2 フロントエンド

#### コンポーネントテスト

- CreateWorkspaceForm（slug バリデーション）
- PublicWorkspaceList（表示・参加）
- AddMemberByEmailForm（入力・送信）

#### E2E テスト

- 未参加ユーザーのワークスペース一覧表示
- ワークスペース作成（slug 入力）
- 公開ワークスペースへの参加
- メンバー追加フロー

---

## 8. API 仕様書

### 8.1 新規・変更エンドポイント

#### `GET /api/workspaces/public`

**説明**: すべての公開ワークスペースを取得

**レスポンス:**

```json
{
  "workspaces": [
    {
      "id": "general",
      "name": "General",
      "description": "一般的なディスカッション用のワークスペース",
      "iconUrl": null,
      "memberCount": 5,
      "isJoined": true,
      "createdAt": "2025-01-01T00:00:00Z"
    }
  ]
}
```

#### `POST /api/workspaces/:id/join`

**説明**: 公開ワークスペースに参加

**パラメータ:**

- `id` (path): ワークスペース slug

**レスポンス:**

```json
{
  "message": "ワークスペースに参加しました"
}
```

**エラー:**

- `400`: ワークスペースが公開されていない、または既に参加済み
- `404`: ワークスペースが見つからない

#### `POST /api/workspaces/:id/members`

**説明**: メールアドレスでメンバーを追加（既存 API の仕様変更）

**リクエストボディ:**

```json
{
  "email": "user@example.com",
  "role": "member"
}
```

**レスポンス:**

```json
{
  "message": "メンバーを追加しました"
}
```

**エラー:**

- `400`: ユーザーが見つからない、または既に参加済み
- `403`: メンバー追加の権限がない

#### `POST /api/workspaces`

**説明**: ワークスペースを作成（リクエスト仕様変更）

**リクエストボディ:**

```json
{
  "id": "my-workspace",
  "name": "My Workspace",
  "description": "説明文",
  "iconUrl": "https://example.com/icon.png",
  "isPublic": false
}
```

**レスポンス:**

```json
{
  "id": "my-workspace",
  "name": "My Workspace",
  "description": "説明文",
  "iconUrl": "https://example.com/icon.png",
  "isPublic": false,
  "role": "owner",
  "createdBy": "user-uuid",
  "createdAt": "2025-01-01T00:00:00Z",
  "updatedAt": "2025-01-01T00:00:00Z"
}
```

#### `PATCH /api/workspaces/:id`

**説明**: ワークスペースを更新（リクエスト仕様変更）

**リクエストボディ:**

```json
{
  "name": "Updated Name",
  "description": "Updated description",
  "isPublic": true
}
```

---

## 9. まとめ

この実装計画書では、ワークスペース ID を UUID から slug 形式に変更し、公開/非公開の管理機能を追加します。主な変更点は以下の通りです:

1. **データベーススキーマ**: `id` を string 型に変更、`is_public` カラムを追加
2. **バックエンド**: slug バリデーション、公開ワークスペース関連のユースケース追加
3. **フロントエンド**: slug 入力 UI、公開ワークスペース一覧・参加機能の実装
4. **Seed データ**: slug を使用したサンプルデータの更新

実装は 10 フェーズに分けて段階的に進め、各フェーズでテストと動作確認を行います。
