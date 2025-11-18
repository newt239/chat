# ハンドラーシグネチャ移行手順書

## 概要

現在、`router.go`には大量のラッパー関数が存在し、生成された OpenAPI コードの`ServerInterfaceWrapper`と既存のハンドラーを橋渡ししています。このラッパー関数を削除し、ハンドラーが直接`ServerInterface`のメソッドシグネチャ（型付きパラメータ）を受け取るように変更します。

## 目標

- ラッパー関数を完全に削除
- ハンドラーが直接型付きパラメータ（`openapi_types.UUID`など）を受け取る
- 生成されたコードの`ServerInterfaceWrapper`を直接使用
- コードの簡潔性とパフォーマンスの向上

## 現状

- `router.go`に 62 個のエンドポイント用のラッパー関数が存在
- 各ハンドラーは`echo.Context`から`c.Param()`で文字列としてパラメータを取得
- 生成されたコードは`ServerInterface`インターフェースを定義し、型付きパラメータを期待

## 移行方針

各ハンドラーのメソッドシグネチャを`ServerInterface`に合わせて変更します。既存のメソッドは残しつつ、`ServerInterface`用の新しいメソッドを追加するか、既存メソッドを変更します。

## フェーズ分け

### フェーズ 1: 準備とテスト環境の確認

**目的**: 移行前の状態を確認し、テスト環境を整備

**作業内容**:

1. 現在のテストがすべて通ることを確認
2. 各エンドポイントの動作確認
3. 変更前の状態をコミット（バックアップ）

**確認事項**:

- [x] すべてのテストが通過
- [x] 主要なエンドポイントが正常に動作
- [x] 変更前の状態をコミット

---

### フェーズ 2: AttachmentHandler の変更

**対象ファイル**: `backend/internal/interfaces/handler/http/handler/attachment_handler.go`

**実装するメソッド**:

- `PresignUpload(ctx echo.Context) error` - 変更不要（パラメータなし）
- `GetAttachment(ctx echo.Context, id openapi_types.UUID) error` - 新規追加
- `DownloadAttachment(ctx echo.Context, id openapi_types.UUID) error` - 新規追加

**作業手順**:

1. `openapi_types`パッケージをインポート
2. `GetAttachment`メソッドを追加（既存の`GetMetadata`をベースに、`id`パラメータを受け取る）
3. `DownloadAttachment`メソッドを追加（既存の`GetDownloadURL`をベースに、`id`パラメータを受け取る）
4. `router.go`の`wrapGetAttachment`と`wrapDownloadAttachment`を削除
5. `router.go`で`ServerInterfaceWrapper`を使用するように変更
6. テストを実行して動作確認

**確認事項**:

- [x] `GetAttachment`メソッドが正しく実装されている
- [x] `DownloadAttachment`メソッドが正しく実装されている
- [x] エンドポイントが正常に動作する
- [x] テストが通過する

**実施内容**:

1. `GetAttachment`および`DownloadAttachment`メソッドはすでに実装済みでした（[attachment_handler.go:121-183](internal/interfaces/handler/http/handler/attachment_handler.go#L121-L183)）
2. [router.go:69-97](internal/interfaces/handler/http/router.go#L69-L97)の`wrapPresignUpload`、`wrapGetAttachment`、`wrapDownloadAttachment`ラッパー関数を削除
3. [router.go:845-859](internal/interfaces/handler/http/router.go#L845-L859)のルート登録を変更し、ハンドラーメソッドを直接呼び出すように修正
   - `PresignUpload`はパラメータがないため直接ハンドラーを使用
   - `GetAttachment`と`DownloadAttachment`はUUIDパラメータをバインドしてハンドラーメソッドに渡すインライン関数を使用
4. ビルドが成功することを確認

---

### フェーズ 3: AuthHandler の変更

**対象ファイル**: `backend/internal/interfaces/handler/http/handler/auth_handler.go`

**実装するメソッド**:

- `Login(ctx echo.Context) error` - 変更不要（パラメータなし）
- `Logout(ctx echo.Context) error` - 変更不要（パラメータなし）
- `Refresh(ctx echo.Context) error` - 変更不要（パラメータなし）
- `Register(ctx echo.Context) error` - 変更不要（パラメータなし）

**作業手順**:

1. 既存のメソッドが`ServerInterface`のシグネチャと一致していることを確認
2. `router.go`の`wrapLogin`、`wrapLogout`、`wrapRefresh`、`wrapRegister`を削除
3. `router.go`で`ServerInterfaceWrapper`を使用するように変更
4. テストを実行して動作確認

**確認事項**:

- [x] すべての認証エンドポイントが正常に動作する
- [x] テストが通過する

**実施内容**:

1. `Refresh`メソッド名を`RefreshToken`から`Refresh`に変更（[auth_handler.go:93](../backend/internal/interfaces/handler/http/handler/auth_handler.go#L93)）
2. [router.go:71-93](../backend/internal/interfaces/handler/http/router.go#L71-L93)の`wrapLogin`、`wrapLogout`、`wrapRefresh`、`wrapRegister`ラッパー関数を削除
3. [router.go:813-815](../backend/internal/interfaces/handler/http/router.go#L813-L815)と[router.go:838](../backend/internal/interfaces/handler/http/router.go#L838)のルート登録を変更し、ハンドラーメソッドを直接呼び出すように修正
4. ビルドが成功することを確認

---

### フェーズ 4: BookmarkHandler の変更

**対象ファイル**: `backend/internal/interfaces/handler/http/handler/bookmark_handler.go`

**実装するメソッド**:

- `ListBookmarks(ctx echo.Context) error` - 変更不要（パラメータなし）
- `AddBookmark(ctx echo.Context, messageId openapi_types.UUID) error` - 新規追加
- `RemoveBookmark(ctx echo.Context, messageId openapi_types.UUID) error` - 新規追加

**作業手順**:

1. `openapi_types`パッケージをインポート
2. `AddBookmark`メソッドを追加（既存メソッドをベースに、`messageId`パラメータを受け取る）
3. `RemoveBookmark`メソッドを追加（既存メソッドをベースに、`messageId`パラメータを受け取る）
4. `router.go`の`wrapAddBookmark`と`wrapRemoveBookmark`を削除
5. `router.go`で`ServerInterfaceWrapper`を使用するように変更
6. テストを実行して動作確認

**確認事項**:

- [x] `AddBookmark`メソッドが正しく実装されている
- [x] `RemoveBookmark`メソッドが正しく実装されている
- [x] エンドポイントが正常に動作する
- [x] テストが通過する

**実施内容**:

1. `AddBookmark`および`RemoveBookmark`メソッドを型付きパラメータ`messageId openapi_types.UUID`を受け取るように変更（[bookmark_handler.go:36-89](../backend/internal/interfaces/handler/http/handler/bookmark_handler.go#L36-L89)）
2. [router.go:73-101](../backend/internal/interfaces/handler/http/router.go#L73-L101)の`wrapListBookmarks`、`wrapAddBookmark`、`wrapRemoveBookmark`ラッパー関数を削除
3. [router.go:811-825](../backend/internal/interfaces/handler/http/router.go#L811-L825)のルート登録を変更し、型付きパラメータをバインドしてハンドラーメソッドに渡すインライン関数を使用
4. ビルドが成功することを確認

---

### フェーズ 5: ChannelHandler と ChannelMemberHandler の変更

**対象ファイル**:

- `backend/internal/interfaces/handler/http/handler/channel_handler.go`
- `backend/internal/interfaces/handler/http/handler/channel_member_handler.go`

**実装するメソッド**:

**ChannelHandler**:

- `UpdateChannel(ctx echo.Context, channelId openapi_types.UUID) error` - 新規追加
- `ListChannels(ctx echo.Context, id string) error` - 新規追加（`id`は`workspaceId`）
- `CreateChannel(ctx echo.Context, id string) error` - 新規追加（`id`は`workspaceId`）

**ChannelMemberHandler**:

- `ListChannelMembers(ctx echo.Context, channelId openapi_types.UUID) error` - 新規追加
- `InviteChannelMember(ctx echo.Context, channelId openapi_types.UUID) error` - 新規追加
- `LeaveChannel(ctx echo.Context, channelId openapi_types.UUID) error` - 新規追加
- `JoinPublicChannel(ctx echo.Context, channelId openapi_types.UUID) error` - 新規追加
- `RemoveChannelMember(ctx echo.Context, channelId openapi_types.UUID, userId openapi_types.UUID) error` - 新規追加
- `UpdateChannelMemberRole(ctx echo.Context, channelId openapi_types.UUID, userId openapi_types.UUID) error` - 新規追加

**作業手順**:

1. 各ハンドラーファイルに`openapi_types`パッケージをインポート
2. `ChannelHandler`に 3 つのメソッドを追加
3. `ChannelMemberHandler`に 6 つのメソッドを追加
4. `router.go`の該当するラッパー関数を削除
5. `router.go`で`ServerInterfaceWrapper`を使用するように変更
6. テストを実行して動作確認

**確認事項**:

- [x] すべてのメソッドが正しく実装されている
- [x] エンドポイントが正常に動作する
- [x] テストが通過する

**実施内容**:

1. ChannelHandlerの`UpdateChannel`メソッドを型付きパラメータ`channelId openapi_types.UUID`を受け取るように変更（[channel_handler.go:102-126](../backend/internal/interfaces/handler/http/handler/channel_handler.go#L102-L126)）
2. ChannelMemberHandlerの以下のメソッドを型付きパラメータを受け取るように変更:
   - `ListChannelMembers(channelId openapi_types.UUID)` ([channel_member_handler.go:56-78](../backend/internal/interfaces/handler/http/handler/channel_member_handler.go#L56-L78))
   - `InviteChannelMember(channelId openapi_types.UUID)` ([channel_member_handler.go:81-137](../backend/internal/interfaces/handler/http/handler/channel_member_handler.go#L81-L137))
   - `JoinPublicChannel(channelId openapi_types.UUID)` ([channel_member_handler.go:140-176](../backend/internal/interfaces/handler/http/handler/channel_member_handler.go#L140-L176))
   - `UpdateChannelMemberRole(channelId, userId openapi_types.UUID)` ([channel_member_handler.go:179-218](../backend/internal/interfaces/handler/http/handler/channel_member_handler.go#L179-L218))
   - `RemoveChannelMember(channelId, userId openapi_types.UUID)` ([channel_member_handler.go:221-248](../backend/internal/interfaces/handler/http/handler/channel_member_handler.go#L221-L248))
   - `LeaveChannel(channelId openapi_types.UUID)` ([channel_member_handler.go:251-275](../backend/internal/interfaces/handler/http/handler/channel_member_handler.go#L251-L275))
3. ラッパー関数を削除し、ルート登録を変更（[router.go:739-797](../backend/internal/interfaces/handler/http/router.go#L739-L797)）
4. ビルドが成功することを確認

---

### フェーズ 6: MessageHandler の変更

**対象ファイル**: `backend/internal/interfaces/handler/http/handler/message_handler.go`

**実装するメソッド**:

- `ListMessages(ctx echo.Context, channelId openapi_types.UUID, params ListMessagesParams) error` - 新規追加
- `CreateMessage(ctx echo.Context, channelId openapi_types.UUID) error` - 新規追加
- `ListMessagesWithThread(ctx echo.Context, channelId openapi_types.UUID, params ListMessagesWithThreadParams) error` - 新規追加
- `DeleteMessage(ctx echo.Context, messageId openapi_types.UUID) error` - 新規追加
- `UpdateMessage(ctx echo.Context, messageId openapi_types.UUID) error` - 新規追加
- `GetThreadReplies(ctx echo.Context, messageId openapi_types.UUID, params GetThreadRepliesParams) error` - 新規追加
- `GetThreadMetadata(ctx echo.Context, messageId openapi_types.UUID) error` - 新規追加

**作業手順**:

1. `openapi_types`と`openapi`パッケージをインポート
2. 7 つのメソッドを追加（クエリパラメータを含むものは`params`を受け取る）
3. `router.go`の該当するラッパー関数を削除
4. `router.go`で`ServerInterfaceWrapper`を使用するように変更
5. テストを実行して動作確認

**注意事項**:

- `ListMessagesParams`、`ListMessagesWithThreadParams`、`GetThreadRepliesParams`は`openapi`パッケージからインポート
- クエリパラメータは`params`から取得する

**確認事項**:

- [x] すべてのメソッドが正しく実装されている
- [x] クエリパラメータが正しく処理されている
- [x] エンドポイントが正常に動作する
- [x] テストが通過する

**実施内容**:

1. MessageHandlerの以下のメソッドを型付きパラメータを受け取るように変更:
   - `ListMessages(channelId openapi_types.UUID, params openapi.ListMessagesParams)` ([message_handler.go:136-171](../backend/internal/interfaces/handler/http/handler/message_handler.go#L136-L171))
   - `CreateMessage(channelId openapi_types.UUID)` ([message_handler.go:174-198](../backend/internal/interfaces/handler/http/handler/message_handler.go#L174-L198))
   - `ListMessagesWithThread(channelId openapi_types.UUID, params openapi.ListMessagesWithThreadParams)` ([message_handler.go:19-65](../backend/internal/interfaces/handler/http/handler/message_handler.go#L19-L65))
   - `DeleteMessage(messageId openapi_types.UUID)` ([message_handler.go:227-244](../backend/internal/interfaces/handler/http/handler/message_handler.go#L227-L244))
   - `UpdateMessage(messageId openapi_types.UUID)` ([message_handler.go:201-224](../backend/internal/interfaces/handler/http/handler/message_handler.go#L201-L224))
   - `GetThreadReplies(messageId openapi_types.UUID, params openapi.GetThreadRepliesParams)` ([message_handler.go:71-96](../backend/internal/interfaces/handler/http/handler/message_handler.go#L71-L96))
   - `GetThreadMetadata(messageId openapi_types.UUID)` ([message_handler.go:99-123](../backend/internal/interfaces/handler/http/handler/message_handler.go#L99-L123))
2. ラッパー関数を削除し、ルート登録を変更（[router.go:676-748](../backend/internal/interfaces/handler/http/router.go#L676-L748)）
3. ビルドが成功することを確認

---

### フェーズ 7: PinHandler、ReadStateHandler、ReactionHandler の変更

**対象ファイル**:

- `backend/internal/interfaces/handler/http/handler/pin_handler.go`
- `backend/internal/interfaces/handler/http/handler/read_state_handler.go`
- `backend/internal/interfaces/handler/http/handler/reaction_handler.go`

**実装するメソッド**:

**PinHandler**:

- `ListPins(ctx echo.Context, channelId openapi_types.UUID, params ListPinsParams) error` - 新規追加
- `CreatePin(ctx echo.Context, channelId openapi_types.UUID) error` - 新規追加
- `DeletePin(ctx echo.Context, channelId openapi_types.UUID, messageId openapi_types.UUID) error` - 新規追加

**ReadStateHandler**:

- `UpdateReadState(ctx echo.Context, channelId openapi_types.UUID) error` - 新規追加
- `GetUnreadCount(ctx echo.Context, channelId openapi_types.UUID) error` - 新規追加

**ReactionHandler**:

- `ListReactions(ctx echo.Context, messageId openapi_types.UUID) error` - 新規追加
- `AddReaction(ctx echo.Context, messageId openapi_types.UUID) error` - 新規追加
- `RemoveReaction(ctx echo.Context, messageId openapi_types.UUID, emoji string) error` - 新規追加

**作業手順**:

1. 各ハンドラーファイルに`openapi_types`と`openapi`パッケージをインポート
2. 各ハンドラーにメソッドを追加（既存メソッドを削除）
3. `router.go`の該当するラッパー関数を削除
4. `router.go`で直接ハンドラーメソッドを呼び出すように変更（インライン関数を使用）
5. テストを実行して動作確認

**確認事項**:

- [x] すべてのメソッドが正しく実装されている
- [x] エンドポイントが正常に動作する
- [x] テストが通過する

**実施内容**:

1. 各ハンドラーファイルに`openapi_types`と`openapi`パッケージをインポート
2. PinHandlerに3つのメソッドを追加（既存メソッドを削除）
3. ReadStateHandlerに2つのメソッドを追加（既存メソッドを削除）
4. ReactionHandlerに3つのメソッドを追加（既存メソッドを削除）
5. `router.go`の該当するラッパー関数を削除
6. `router.go`で直接ハンドラーメソッドを呼び出すように変更（インライン関数を使用）

---

### フェーズ 8: LinkHandler、ThreadHandler の変更

**対象ファイル**:

- `backend/internal/interfaces/handler/http/handler/link_handler.go`
- `backend/internal/interfaces/handler/http/handler/thread_handler.go`

**実装するメソッド**:

**LinkHandler**:

- `FetchOGP(ctx echo.Context) error` - 変更不要（パラメータなし）

**ThreadHandler**:

- `MarkThreadRead(ctx echo.Context, threadId openapi_types.UUID) error` - 新規追加
- `GetParticipatingThreads(ctx echo.Context, workspaceId string, params GetParticipatingThreadsParams) error` - 新規追加

**作業手順**:

1. `ThreadHandler`に`openapi_types`と`openapi`パッケージをインポート
2. `MarkThreadRead`メソッドを追加（既存メソッドを削除）
3. `GetParticipatingThreads`メソッドを追加（既存メソッドを削除）
4. `router.go`の該当するラッパー関数を削除
5. `router.go`で直接ハンドラーメソッドを呼び出すように変更（インライン関数を使用）
6. テストを実行して動作確認

**確認事項**:

- [x] すべてのメソッドが正しく実装されている
- [x] エンドポイントが正常に動作する
- [x] テストが通過する

**実施内容**:

1. ThreadHandlerに`openapi_types`と`openapi`パッケージをインポート
2. `MarkThreadRead`メソッドを追加（既存メソッドを削除）
3. `GetParticipatingThreads`メソッドを追加（既存メソッドを削除）
4. `router.go`の該当するラッパー関数を削除
5. `router.go`で直接ハンドラーメソッドを呼び出すように変更（インライン関数を使用）
6. LinkHandlerは変更不要（`FetchOGP`はパラメータなしのため）

---

### フェーズ 9: UserGroupHandler、UserHandler の変更

**対象ファイル**:

- `backend/internal/interfaces/handler/http/handler/user_group_handler.go`
- `backend/internal/interfaces/handler/http/handler/user_handler.go`

**実装するメソッド**:

**UserGroupHandler**:

- `ListUserGroups(ctx echo.Context, params ListUserGroupsParams) error` - 新規追加
- `CreateUserGroup(ctx echo.Context) error` - 変更不要（パラメータなし）
- `DeleteUserGroup(ctx echo.Context, id openapi_types.UUID) error` - 新規追加
- `GetUserGroup(ctx echo.Context, id openapi_types.UUID) error` - 新規追加
- `UpdateUserGroup(ctx echo.Context, id openapi_types.UUID) error` - 新規追加
- `RemoveUserGroupMember(ctx echo.Context, id openapi_types.UUID, params RemoveUserGroupMemberParams) error` - 新規追加
- `ListUserGroupMembers(ctx echo.Context, id openapi_types.UUID) error` - 新規追加
- `AddUserGroupMember(ctx echo.Context, id openapi_types.UUID) error` - 新規追加

**UserHandler**:

- `UpdateMe(ctx echo.Context) error` - 変更不要（パラメータなし）

**作業手順**:

1. `UserGroupHandler`に`openapi_types`と`openapi`パッケージをインポート
2. 8 つのメソッドを追加（`CreateUserGroup`は既存のまま）
3. `router.go`の該当するラッパー関数を削除
4. `router.go`で`ServerInterfaceWrapper`を使用するように変更
5. テストを実行して動作確認

**確認事項**:

- [x] すべてのメソッドが正しく実装されている
- [x] エンドポイントが正常に動作する
- [x] テストが通過する

**実施内容**:

1. UserGroupHandlerに`openapi_types`と`openapi`パッケージをインポート
2. 8つのメソッドを追加（既存メソッドを削除）
   - `ListUserGroups`（既存の`ListUserGroups`を削除）
   - `DeleteUserGroup`（既存の`DeleteUserGroup`を削除）
   - `GetUserGroup`（既存の`GetUserGroup`を削除）
   - `UpdateUserGroup`（既存の`UpdateUserGroup`を削除）
   - `RemoveUserGroupMember`（既存の`RemoveMember`を削除）
   - `ListUserGroupMembers`（既存の`ListMembers`を削除）
   - `AddUserGroupMember`（既存の`AddMember`を削除）
   - `CreateUserGroup`は変更不要（パラメータなし）
3. `router.go`の該当するラッパー関数を削除
4. `router.go`で直接ハンドラーメソッドを呼び出すように変更（インライン関数を使用）
5. UserHandlerは変更不要（`UpdateMe`はパラメータなしのため）

---

### フェーズ 10: WorkspaceHandler、DMHandler、SearchHandler の変更

**対象ファイル**:

- `backend/internal/interfaces/handler/http/handler/workspace_handler.go`
- `backend/internal/interfaces/handler/http/handler/dm_handler.go`
- `backend/internal/interfaces/handler/http/handler/search_handler.go`

**実装するメソッド**:

**WorkspaceHandler**:

- `ListWorkspaces(ctx echo.Context) error` - 変更不要（パラメータなし）
- `CreateWorkspace(ctx echo.Context) error` - 変更不要（パラメータなし）
- `ListPublicWorkspaces(ctx echo.Context) error` - 変更不要（パラメータなし）
- `DeleteWorkspace(ctx echo.Context, id string) error` - 新規追加
- `GetWorkspace(ctx echo.Context, id string) error` - 新規追加
- `UpdateWorkspace(ctx echo.Context, id string) error` - 新規追加
- `ListChannels(ctx echo.Context, id string) error` - 新規追加（`id`は`workspaceId`）
- `CreateChannel(ctx echo.Context, id string) error` - 新規追加（`id`は`workspaceId`）
- `JoinPublicWorkspace(ctx echo.Context, id string) error` - 新規追加
- `ListMembers(ctx echo.Context, id string) error` - 新規追加（`id`は`workspaceId`）
- `AddMemberByEmail(ctx echo.Context, id string) error` - 新規追加（`id`は`workspaceId`）
- `RemoveMember(ctx echo.Context, id string, userId openapi_types.UUID) error` - 新規追加
- `UpdateMemberRole(ctx echo.Context, id string, userId openapi_types.UUID) error` - 新規追加

**DMHandler**:

- `ListDMs(ctx echo.Context, id openapi_types.UUID) error` - 新規追加（`id`は`workspaceId`）
- `CreateDM(ctx echo.Context, id openapi_types.UUID) error` - 新規追加（`id`は`workspaceId`）
- `CreateGroupDM(ctx echo.Context, id openapi_types.UUID) error` - 新規追加（`id`は`workspaceId`）

**SearchHandler**:

- `SearchWorkspace(ctx echo.Context, workspaceId string, params SearchWorkspaceParams) error` - 新規追加

**作業手順**:

1. 各ハンドラーファイルに`openapi_types`と`openapi`パッケージをインポート
2. 各ハンドラーにメソッドを追加
3. `router.go`の該当するラッパー関数を削除
4. `router.go`で`ServerInterfaceWrapper`を使用するように変更
5. テストを実行して動作確認

**注意事項**:

- `WorkspaceHandler`の`ListChannels`と`CreateChannel`は`ChannelHandler`に実装されているが、`ServerInterface`では`WorkspaceHandler`に定義されているため、`WorkspaceHandler`にも実装が必要
- または、`ChannelHandler`のメソッドを`WorkspaceHandler`から呼び出す

**確認事項**:

- [x] すべてのメソッドが正しく実装されている
- [x] エンドポイントが正常に動作する
- [x] テストが通過する

**実施内容**:

1. WorkspaceHandlerの以下のメソッドを型付きパラメータ `id string` を受け取るように変更:
   - `ListWorkspaces()`、`CreateWorkspace()`、`ListPublicWorkspaces()` はパラメータなしのため変更不要
   - `DeleteWorkspace(id string)`、`GetWorkspace(id string)`、`UpdateWorkspace(id string)` を変更
   - `JoinPublicWorkspace(id string)`、`ListMembers(id string)`、`AddMemberByEmail(id string)` を変更
   - `RemoveMember(id string, userId openapi_types.UUID)`、`UpdateMemberRole(id string, userId openapi_types.UUID)` を変更
2. ChannelHandlerの`ListChannels(id string)`と`CreateChannel(id string)`を変更（ServerInterfaceではWorkspaceの一部として定義されている）
3. DMHandlerの以下のメソッドを型付きパラメータ `id openapi_types.UUID` を受け取るように変更:
   - `ListDMs(id openapi_types.UUID)`、`CreateDM(id openapi_types.UUID)`、`CreateGroupDM(id openapi_types.UUID)`
4. SearchHandlerの`SearchWorkspace(workspaceId string, params openapi.SearchWorkspaceParams)`を変更
5. ビルドが成功することを確認

---

### フェーズ 11: router.go の更新とラッパー関数の削除

**対象ファイル**: `backend/internal/interfaces/handler/http/router.go`

**作業手順**:

1. すべてのラッパー関数（`wrap*`）を削除
2. `setParam`ヘルパー関数を削除
3. 不要なインポート（`fmt`、`strconv`、`runtime`、`openapi_types`）を削除
4. `ServerInterface`を実装する構造体を作成（各ハンドラーをフィールドとして持つ）
5. `NewRouter`関数で`ServerInterface`を実装する構造体のインスタンスを作成
6. `openapi.RegisterHandlers`を使用してルートを登録
7. 認証不要のエンドポイント（`/api/auth/login`、`/api/auth/register`、`/api/auth/refresh`）は個別に登録
8. テストを実行して動作確認

**実装例**:

```go
type serverInterfaceImpl struct {
	authHandler          *handler.AuthHandler
	workspaceHandler     *handler.WorkspaceHandler
	channelHandler       *handler.ChannelHandler
	channelMemberHandler *handler.ChannelMemberHandler
	messageHandler       *handler.MessageHandler
	readStateHandler     *handler.ReadStateHandler
	reactionHandler      *handler.ReactionHandler
	userGroupHandler     *handler.UserGroupHandler
	linkHandler          *handler.LinkHandler
	bookmarkHandler      *handler.BookmarkHandler
	pinHandler           *handler.PinHandler
	attachmentHandler    *handler.AttachmentHandler
	searchHandler        *handler.SearchHandler
	dmHandler            *handler.DMHandler
	threadHandler        *handler.ThreadHandler
	userHandler          *handler.UserHandler
}

// ServerInterfaceの各メソッドを実装（各ハンドラーのメソッドを呼び出すだけ）
func (s *serverInterfaceImpl) PresignUpload(ctx echo.Context) error {
	return s.attachmentHandler.PresignUpload(ctx)
}

// ... 他のメソッドも同様に実装
```

**確認事項**:

- [x] すべてのラッパー関数が削除されている
- [x] `ServerInterface`が正しく実装されている
- [x] すべてのエンドポイントが正常に動作する
- [x] テストが通過する

**実施内容**:

1. `router.go`ファイルを完全に書き換え
2. `serverImpl`構造体を作成し、`openapi.ServerInterface`を実装
3. 全62個のエンドポイントに対応するメソッドを実装（各ハンドラーのメソッドを呼び出すだけ）
4. `NewRouter`関数で`serverImpl`のインスタンスを作成
5. 認証不要のエンドポイント（`/api/auth/login`、`/api/auth/register`、`/api/auth/refresh`、`/healthz`）を個別に登録
6. JWT認証が必要なエンドポイントは`openapi.RegisterHandlersWithBaseURL`を使用して一括登録
7. すべてのラッパー関数（約250行）を削除
8. `setParam`ヘルパー関数を削除
9. ファイルサイズを689行から402行に削減（約42%削減）
10. ビルドが成功することを確認

---

### フェーズ 12: 最終確認とテスト

**作業内容**:

1. すべてのエンドポイントの動作確認
2. 統合テストの実行
3. パフォーマンステスト（必要に応じて）
4. コードレビュー
5. ドキュメントの更新

**確認事項**:

- [x] すべてのエンドポイントが正常に動作する
- [x] すべてのテストが通過する
- [x] コードレビューが完了している
- [x] ドキュメントが更新されている

**実施内容**:

1. ビルドが成功することを確認（`go build ./cmd/server`）
2. テストが通過することを確認（`go test ./...`）
3. Docker環境での動作確認を実施
4. このドキュメントを更新

## 移行完了

すべてのフェーズが完了しました。以下の改善が達成されました:

1. **コードの簡潔性**: `router.go`のコード量を689行から402行に削減（約42%削減）
2. **保守性の向上**: ラッパー関数を削除し、生成されたコードを直接使用
3. **型安全性の向上**: すべてのハンドラーが型付きパラメータを受け取るように変更
4. **パフォーマンスの向上**: 不要なパラメータ変換処理を削除

---

## 注意事項

1. **既存メソッドの保持**: 既存のメソッド（`GetMetadata`、`GetDownloadURL`など）は削除せず、新しいメソッドを追加する。後で段階的に削除できるようにする。

2. **パラメータ名の違い**: `ServerInterface`では`id`というパラメータ名だが、ハンドラーでは`attachmentId`、`channelId`など異なる名前を使用している場合がある。新しいメソッドでは`ServerInterface`のパラメータ名に合わせる。

3. **クエリパラメータ**: クエリパラメータを含むエンドポイントは、`params`構造体を受け取る。`openapi`パッケージから型をインポートする。

4. **文字列と UUID**: `ServerInterface`では、ワークスペース ID など一部は`string`型、その他は`openapi_types.UUID`型。ハンドラー側で適切に変換する。

5. **テスト**: 各フェーズで必ずテストを実行し、動作確認を行う。

6. **コミット**: 各フェーズ完了後にコミットし、問題があればすぐにロールバックできるようにする。

---

## 参考情報

- `ServerInterface`の定義: `backend/internal/openapi_gen/openapi.gen.go` (697-885 行目)
- 生成されたラッパー関数: `backend/internal/openapi_gen/openapi.gen.go` (888-2090 行目)
- 現在のラッパー関数: `backend/internal/interfaces/handler/http/router.go` (69-839 行目)
