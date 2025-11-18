# Handler コンストラクタ関数削除手順書

## 概要

現在、`backend/internal/interfaces/handler/http/handler/` 配下の各ハンドラーには `New〇〇Handler` という形式のコンストラクタ関数が存在しています。これらは単に構造体を初期化して返すだけであり、将来的な拡張予定もないため、直接構造体リテラルで初期化する方式に変更します。

## 目標

- すべての `New〇〇Handler` コンストラクタ関数を削除
- `InterfaceRegistry` で直接構造体リテラルを使用して初期化
- コードの簡潔性を向上させる
- YAGNI 原則（You Aren't Gonna Need It）に従った設計

## 現状

### 現在のパターン

```go
// handler/auth_handler.go
func NewAuthHandler(authUC authuc.AuthUseCase) *AuthHandler {
    return &AuthHandler{authUC: authUC}
}

// registry/interface_registry.go
func (r *InterfaceRegistry) NewAuthHandler() *handler.AuthHandler {
    return handler.NewAuthHandler(r.usecaseRegistry.NewAuthUseCase())
}
```

### 変更後のパターン

```go
// handler/auth_handler.go
// NewAuthHandler 関数を削除

// registry/interface_registry.go
func (r *InterfaceRegistry) NewAuthHandler() *handler.AuthHandler {
    return &handler.AuthHandler{
        authUC: r.usecaseRegistry.NewAuthUseCase(),
    }
}
```

## 対象ファイルと変更内容

### 1. AttachmentHandler

**ファイル**: `backend/internal/interfaces/handler/http/handler/attachment_handler.go:16-20`

**削除する関数**:
```go
func NewAttachmentHandler(attachmentUseCase *attachment.Interactor) *AttachmentHandler {
    return &AttachmentHandler{
        attachmentUseCase: attachmentUseCase,
    }
}
```

**Registry での変更後**:
```go
func (r *InterfaceRegistry) NewAttachmentHandler() *handler.AttachmentHandler {
    return &handler.AttachmentHandler{
        attachmentUseCase: r.usecaseRegistry.NewAttachmentUseCase(),
    }
}
```

---

### 2. AuthHandler

**ファイル**: `backend/internal/interfaces/handler/http/handler/auth_handler.go:16-18`

**削除する関数**:
```go
func NewAuthHandler(authUC authuc.AuthUseCase) *AuthHandler {
    return &AuthHandler{authUC: authUC}
}
```

**Registry での変更後**:
```go
func (r *InterfaceRegistry) NewAuthHandler() *handler.AuthHandler {
    return &handler.AuthHandler{
        authUC: r.usecaseRegistry.NewAuthUseCase(),
    }
}
```

---

### 3. BookmarkHandler

**ファイル**: `backend/internal/interfaces/handler/http/handler/bookmark_handler.go:15-17`

**削除する関数**:
```go
func NewBookmarkHandler(bookmarkUC bookmark.BookmarkUseCase) *BookmarkHandler {
    return &BookmarkHandler{bookmarkUC: bookmarkUC}
}
```

**Registry での変更後**:
```go
func (r *InterfaceRegistry) NewBookmarkHandler() *handler.BookmarkHandler {
    return &handler.BookmarkHandler{
        bookmarkUC: r.usecaseRegistry.NewBookmarkUseCase(),
    }
}
```

---

### 4. ChannelHandler

**ファイル**: `backend/internal/interfaces/handler/http/handler/channel_handler.go:16-18`

**削除する関数**:
```go
func NewChannelHandler(channelUC channeluc.ChannelUseCase) *ChannelHandler {
    return &ChannelHandler{channelUC: channelUC}
}
```

**Registry での変更後**:
```go
func (r *InterfaceRegistry) NewChannelHandler() *handler.ChannelHandler {
    return &handler.ChannelHandler{
        channelUC: r.usecaseRegistry.NewChannelUseCase(),
    }
}
```

---

### 5. ChannelMemberHandler

**ファイル**: `backend/internal/interfaces/handler/http/handler/channel_member_handler.go:21-23`

**削除する関数**:
```go
func NewChannelMemberHandler(channelMemberUseCase channelmember.ChannelMemberUseCase, systemMessageUC systemmessage.UseCase) *ChannelMemberHandler {
    return &ChannelMemberHandler{channelMemberUseCase: channelMemberUseCase, systemMessageUC: systemMessageUC}
}
```

**Registry での変更後**:
```go
func (r *InterfaceRegistry) NewChannelMemberHandler() *handler.ChannelMemberHandler {
    return &handler.ChannelMemberHandler{
        channelMemberUseCase: r.usecaseRegistry.NewChannelMemberUseCase(),
        systemMessageUC:      r.usecaseRegistry.NewSystemMessageUseCase(),
    }
}
```

**注意**: この Handler は複数の依存関係を持っているため、構造体リテラルを複数行で記述します。

---

### 6. DMHandler

**ファイル**: `backend/internal/interfaces/handler/http/handler/dm_handler.go:17-19`

**削除する関数**:
```go
func NewDMHandler(dmInteractor *dmuc.Interactor) *DMHandler {
    return &DMHandler{dmInteractor: dmInteractor}
}
```

**Registry での変更後**:
```go
func (r *InterfaceRegistry) NewDMHandler() *handler.DMHandler {
    return &handler.DMHandler{
        dmInteractor: r.usecaseRegistry.NewDMInteractor(),
    }
}
```

---

### 7. LinkHandler

**ファイル**: `backend/internal/interfaces/handler/http/handler/link_handler.go:15-17`

**削除する関数**:
```go
func NewLinkHandler(linkUC linkuc.LinkUseCase) *LinkHandler {
    return &LinkHandler{linkUC: linkUC}
}
```

**Registry での変更後**:
```go
func (r *InterfaceRegistry) NewLinkHandler() *handler.LinkHandler {
    return &handler.LinkHandler{
        linkUC: r.usecaseRegistry.NewLinkUseCase(),
    }
}
```

---

### 8. MessageHandler

**ファイル**: `backend/internal/interfaces/handler/http/handler/message_handler.go:66-68`

**削除する関数**:
```go
func NewMessageHandler(messageUC messageuc.MessageUseCase) *MessageHandler {
    return &MessageHandler{messageUC: messageUC}
}
```

**Registry での変更後**:
```go
func (r *InterfaceRegistry) NewMessageHandler() *handler.MessageHandler {
    return &handler.MessageHandler{
        messageUC: r.usecaseRegistry.NewMessageUseCase(),
    }
}
```

---

### 9. PinHandler

**ファイル**: `backend/internal/interfaces/handler/http/handler/pin_handler.go:17`

**削除する関数**:
```go
func NewPinHandler(uc pin.PinUseCase) *PinHandler { return &PinHandler{uc: uc} }
```

**Registry での変更後**:
```go
func (r *InterfaceRegistry) NewPinHandler() *handler.PinHandler {
    return &handler.PinHandler{
        uc: r.usecaseRegistry.NewPinUseCase(),
    }
}
```

---

### 10. ReadStateHandler

**ファイル**: `backend/internal/interfaces/handler/http/handler/read_state_handler.go:18-20`

**削除する関数**:
```go
func NewReadStateHandler(readStateUC readstateuc.ReadStateUseCase) *ReadStateHandler {
    return &ReadStateHandler{readStateUC: readStateUC}
}
```

**Registry での変更後**:
```go
func (r *InterfaceRegistry) NewReadStateHandler() *handler.ReadStateHandler {
    return &handler.ReadStateHandler{
        readStateUC: r.usecaseRegistry.NewReadStateUseCase(),
    }
}
```

---

### 11. ReactionHandler

**ファイル**: `backend/internal/interfaces/handler/http/handler/reaction_handler.go:17-19`

**削除する関数**:
```go
func NewReactionHandler(reactionUC reactionuc.ReactionUseCase) *ReactionHandler {
    return &ReactionHandler{reactionUC: reactionUC}
}
```

**Registry での変更後**:
```go
func (r *InterfaceRegistry) NewReactionHandler() *handler.ReactionHandler {
    return &handler.ReactionHandler{
        reactionUC: r.usecaseRegistry.NewReactionUseCase(),
    }
}
```

---

### 12. SearchHandler

**ファイル**: `backend/internal/interfaces/handler/http/handler/search_handler.go:17-19`

**削除する関数**:
```go
func NewSearchHandler(searchUC searchuc.SearchUseCase) *SearchHandler {
    return &SearchHandler{searchUC: searchUC}
}
```

**Registry での変更後**:
```go
func (r *InterfaceRegistry) NewSearchHandler() *handler.SearchHandler {
    return &handler.SearchHandler{
        searchUC: r.usecaseRegistry.NewSearchUseCase(),
    }
}
```

---

### 13. ThreadHandler

**ファイル**: `backend/internal/interfaces/handler/http/handler/thread_handler.go:20-28`

**削除する関数**:
```go
func NewThreadHandler(
    threadLister *threaduc.ThreadLister,
    threadReader *threaduc.ThreadReader,
) *ThreadHandler {
    return &ThreadHandler{
        threadLister: threadLister,
        threadReader: threadReader,
    }
}
```

**Registry での変更後**:
```go
func (r *InterfaceRegistry) NewThreadHandler() *handler.ThreadHandler {
    return &handler.ThreadHandler{
        threadLister: r.usecaseRegistry.NewThreadLister(),
        threadReader: r.usecaseRegistry.NewThreadReader(),
    }
}
```

**注意**: この Handler は複数の依存関係を持っているため、構造体リテラルを複数行で記述します。

---

### 14. UserHandler

**ファイル**: `backend/internal/interfaces/handler/http/handler/user_handler.go:15-17`

**削除する関数**:
```go
func NewUserHandler(uc useruc.UseCase) *UserHandler {
    return &UserHandler{uc: uc}
}
```

**Registry での変更後**:
```go
func (r *InterfaceRegistry) NewUserHandler() *handler.UserHandler {
    return &handler.UserHandler{
        uc: r.usecaseRegistry.NewUserUseCase(),
    }
}
```

---

### 15. UserGroupHandler

**ファイル**: `backend/internal/interfaces/handler/http/handler/user_group_handler.go:17-19`

**削除する関数**:
```go
func NewUserGroupHandler(userGroupUC usergroupuc.UserGroupUseCase) *UserGroupHandler {
    return &UserGroupHandler{userGroupUC: userGroupUC}
}
```

**Registry での変更後**:
```go
func (r *InterfaceRegistry) NewUserGroupHandler() *handler.UserGroupHandler {
    return &handler.UserGroupHandler{
        userGroupUC: r.usecaseRegistry.NewUserGroupUseCase(),
    }
}
```

---

### 16. WorkspaceHandler

**ファイル**: `backend/internal/interfaces/handler/http/handler/workspace_handler.go:17-19`

**削除する関数**:
```go
func NewWorkspaceHandler(workspaceUC workspaceuc.WorkspaceUseCase) *WorkspaceHandler {
    return &WorkspaceHandler{workspaceUC: workspaceUC}
}
```

**Registry での変更後**:
```go
func (r *InterfaceRegistry) NewWorkspaceHandler() *handler.WorkspaceHandler {
    return &handler.WorkspaceHandler{
        workspaceUC: r.usecaseRegistry.NewWorkspaceUseCase(),
    }
}
```

---

## 実装手順

### フェーズ 1: InterfaceRegistry の更新

**対象ファイル**: `backend/internal/registry/interface_registry.go`

**作業内容**:

1. `NewAuthHandler` から `NewWorkspaceHandler` までの 16 個の関数を一括で更新
2. 各関数内で `handler.New〇〇Handler()` の呼び出しを構造体リテラルに置き換え
3. 複数の依存関係を持つ Handler（ChannelMemberHandler、ThreadHandler）は複数行で記述

**変更前の例**:
```go
func (r *InterfaceRegistry) NewAuthHandler() *handler.AuthHandler {
    return handler.NewAuthHandler(r.usecaseRegistry.NewAuthUseCase())
}
```

**変更後の例**:
```go
func (r *InterfaceRegistry) NewAuthHandler() *handler.AuthHandler {
    return &handler.AuthHandler{
        authUC: r.usecaseRegistry.NewAuthUseCase(),
    }
}
```

**確認事項**:
- [ ] 16 個すべての Registry 関数を更新
- [ ] フィールド名が正しいことを確認（構造体定義と一致）
- [ ] ビルドが成功することを確認

---

### フェーズ 2: Handler ファイルからコンストラクタ関数を削除

**対象ファイル**: `backend/internal/interfaces/handler/http/handler/*_handler.go` (16 ファイル)

**作業内容**:

1. 各ハンドラーファイルから `New〇〇Handler` 関数を削除
2. 削除する関数は以下の 16 個:
   - `NewAttachmentHandler`
   - `NewAuthHandler`
   - `NewBookmarkHandler`
   - `NewChannelHandler`
   - `NewChannelMemberHandler`
   - `NewDMHandler`
   - `NewLinkHandler`
   - `NewMessageHandler`
   - `NewPinHandler`
   - `NewReadStateHandler`
   - `NewReactionHandler`
   - `NewSearchHandler`
   - `NewThreadHandler`
   - `NewUserHandler`
   - `NewUserGroupHandler`
   - `NewWorkspaceHandler`

**確認事項**:
- [ ] 16 個すべてのコンストラクタ関数を削除
- [ ] ビルドが成功することを確認
- [ ] 未使用のインポートが残っていないことを確認（goimports で自動整理）

---

### フェーズ 3: 最終確認とテスト

**作業内容**:

1. ビルドの確認
   ```bash
   cd backend
   go build ./cmd/server
   ```

2. テストの実行
   ```bash
   go test ./...
   ```

3. Docker 環境での動作確認
   ```bash
   docker-compose up -d --build
   # API エンドポイントの動作確認
   ```

**確認事項**:
- [ ] ビルドが成功する
- [ ] すべてのテストが通過する
- [ ] Docker 環境で正常に起動する
- [ ] 主要なエンドポイントが正常に動作する

---

## 期待される効果

### コード削減量

- 各ハンドラーファイル: 平均 3〜9 行の削減（合計約 80 行）
- `interface_registry.go`: 可読性の向上（行数は変わらないが、より直接的な記述になる）

### 保守性の向上

1. **コードの簡潔性**: 不要な関数を削除し、直接構造体を初期化
2. **依存関係の明確化**: Registry で依存関係が一目でわかる
3. **YAGNI 原則の遵守**: 将来使うかもしれない拡張ポイントを削除

---

## 注意事項

1. **エクスポートされたシンボルの削除**: `New〇〇Handler` 関数は公開関数ですが、プロジェクト内でのみ使用されているため削除しても問題ありません。

2. **後方互換性**: この変更は内部実装のリファクタリングであり、外部 API には影響しません。

3. **一括変更**: すべての Handler を同時に変更するため、フェーズ 1 完了後にビルドエラーが発生します。フェーズ 2 完了後に解消されます。

4. **コミット**: フェーズ 1 と 2 をまとめて 1 つのコミットにすることを推奨します。

---

## 参考情報

- Handler 定義: `backend/internal/interfaces/handler/http/handler/*_handler.go`
- Registry: `backend/internal/registry/interface_registry.go`
- 使用箇所: `backend/internal/registry/interface_registry.go` のみ（外部からは呼ばれていない）
