# API 仕様書

## 概要

本 API は Echo フレームワークベースの RESTful API で、クリーンアーキテクチャに従って実装されています。

## 認証

### JWT 認証

すべての保護されたエンドポイントは JWT 認証が必要です。

```http
Authorization: Bearer <access_token>
```

### トークン管理

- **アクセストークン**: 15 分間有効
- **リフレッシュトークン**: 7 日間有効
- **自動更新**: リフレッシュトークンによる自動更新

## エンドポイント一覧

### 認証 (Authentication)

#### POST /api/auth/register

ユーザー登録

**リクエスト:**

```json
{
  "email": "user@example.com",
  "password": "password123",
  "displayName": "User Name"
}
```

**レスポンス:**

```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expiresAt": "2024-01-01T12:00:00Z",
  "user": {
    "id": "user-id",
    "email": "user@example.com",
    "displayName": "User Name",
    "avatarURL": null
  }
}
```

#### POST /api/auth/login

ユーザーログイン

**リクエスト:**

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**レスポンス:**

```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expiresAt": "2024-01-01T12:00:00Z",
  "user": {
    "id": "user-id",
    "email": "user@example.com",
    "displayName": "User Name",
    "avatarURL": null
  }
}
```

#### POST /api/auth/refresh

トークンリフレッシュ

**リクエスト:**

```json
{
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**レスポンス:**

```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expiresAt": "2024-01-01T12:00:00Z",
  "user": {
    "id": "user-id",
    "email": "user@example.com",
    "displayName": "User Name",
    "avatarURL": null
  }
}
```

#### POST /api/auth/logout

ユーザーログアウト

**リクエスト:**

```json
{
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**レスポンス:**

```json
{
  "message": "Logged out successfully"
}
```

### ワークスペース (Workspaces)

#### GET /api/workspaces

ワークスペース一覧取得

**認証:** 必要

**レスポンス:**

```json
{
  "workspaces": [
    {
      "id": "workspace-id",
      "name": "My Workspace",
      "description": "Workspace description",
      "iconURL": null,
      "createdAt": "2024-01-01T00:00:00Z",
      "updatedAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### POST /api/workspaces

ワークスペース作成

**認証:** 必要

**リクエスト:**

```json
{
  "name": "New Workspace",
  "description": "Workspace description"
}
```

**レスポンス:**

```json
{
  "id": "workspace-id",
  "name": "New Workspace",
  "description": "Workspace description",
  "iconURL": null,
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z"
}
```

#### GET /api/workspaces/:id

ワークスペース詳細取得

**認証:** 必要

**レスポンス:**

```json
{
  "id": "workspace-id",
  "name": "My Workspace",
  "description": "Workspace description",
  "iconURL": null,
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z"
}
```

#### PATCH /api/workspaces/:id

ワークスペース更新

**認証:** 必要

**リクエスト:**

```json
{
  "name": "Updated Workspace",
  "description": "Updated description"
}
```

**レスポンス:**

```json
{
  "id": "workspace-id",
  "name": "Updated Workspace",
  "description": "Updated description",
  "iconURL": null,
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z"
}
```

#### DELETE /api/workspaces/:id

ワークスペース削除

**認証:** 必要

**レスポンス:**

```json
{
  "message": "Workspace deleted successfully"
}
```

#### GET /api/workspaces/:id/members

ワークスペースメンバー一覧

**認証:** 必要

**レスポンス:**

```json
{
  "members": [
    {
      "id": "user-id",
      "email": "user@example.com",
      "displayName": "User Name",
      "avatarURL": null,
      "role": "admin"
    }
  ]
}
```

#### POST /api/workspaces/:id/members

ワークスペースメンバー追加

**認証:** 必要

**リクエスト:**

```json
{
  "userID": "user-id",
  "role": "member"
}
```

**レスポンス:**

```json
{
  "message": "Member added successfully"
}
```

#### PATCH /api/workspaces/:id/members/:userId

メンバーロール更新

**認証:** 必要

**リクエスト:**

```json
{
  "role": "admin"
}
```

**レスポンス:**

```json
{
  "message": "Member role updated successfully"
}
```

#### DELETE /api/workspaces/:id/members/:userId

メンバー削除

**認証:** 必要

**レスポンス:**

```json
{
  "message": "Member removed successfully"
}
```

### チャンネル (Channels)

#### GET /api/workspaces/:id/channels

チャンネル一覧取得

**認証:** 必要

**レスポンス:**

```json
{
  "channels": [
    {
      "id": "channel-id",
      "name": "general",
      "description": "General channel",
      "isPrivate": false,
      "createdAt": "2024-01-01T00:00:00Z",
      "updatedAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### POST /api/workspaces/:id/channels

チャンネル作成

**認証:** 必要

**リクエスト:**

```json
{
  "name": "new-channel",
  "description": "New channel description",
  "isPrivate": false
}
```

**レスポンス:**

```json
{
  "id": "channel-id",
  "name": "new-channel",
  "description": "New channel description",
  "isPrivate": false,
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z"
}
```

### メッセージ (Messages)

#### GET /api/channels/:channelId/messages

メッセージ一覧取得

**認証:** 必要

**クエリパラメータ:**

- `limit`: 取得件数 (デフォルト: 50)
- `since`: 開始日時 (ISO 8601)
- `until`: 終了日時 (ISO 8601)

**レスポンス:**

```json
{
  "messages": [
    {
      "id": "message-id",
      "content": "Hello, World!",
      "type": "text",
      "userID": "user-id",
      "channelID": "channel-id",
      "parentID": null,
      "createdAt": "2024-01-01T00:00:00Z",
      "updatedAt": "2024-01-01T00:00:00Z",
      "user": {
        "id": "user-id",
        "displayName": "User Name",
        "avatarURL": null
      }
    }
  ]
}
```

#### POST /api/channels/:channelId/messages

メッセージ作成

**認証:** 必要

**リクエスト:**

```json
{
  "content": "Hello, World!",
  "type": "text",
  "parentID": null
}
```

**レスポンス:**

```json
{
  "id": "message-id",
  "content": "Hello, World!",
  "type": "text",
  "userID": "user-id",
  "channelID": "channel-id",
  "parentID": null,
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z",
  "user": {
    "id": "user-id",
    "displayName": "User Name",
    "avatarURL": null
  }
}
```

### 既読状態 (Read States)

#### GET /api/channels/:channelId/unread_count

未読数取得

**認証:** 必要

**レスポンス:**

```json
{
  "unreadCount": 5
}
```

#### POST /api/channels/:channelId/reads

既読状態更新

**認証:** 必要

**リクエスト:**

```json
{
  "messageID": "message-id"
}
```

**レスポンス:**

```json
{
  "message": "Read state updated successfully"
}
```

### リアクション (Reactions)

#### GET /api/messages/:messageId/reactions

リアクション一覧取得

**認証:** 必要

**レスポンス:**

```json
{
  "reactions": [
    {
      "emoji": "👍",
      "count": 3,
      "users": [
        {
          "id": "user-id",
          "displayName": "User Name",
          "avatarURL": null
        }
      ]
    }
  ]
}
```

#### POST /api/messages/:messageId/reactions

リアクション追加

**認証:** 必要

**リクエスト:**

```json
{
  "emoji": "👍"
}
```

**レスポンス:**

```json
{
  "message": "Reaction added successfully"
}
```

#### DELETE /api/messages/:messageId/reactions/:emoji

リアクション削除

**認証:** 必要

**レスポンス:**

```json
{
  "message": "Reaction removed successfully"
}
```

### ユーザーグループ (User Groups)

#### POST /api/user-groups

ユーザーグループ作成

**認証:** 必要

**リクエスト:**

```json
{
  "name": "Development Team",
  "description": "Development team group"
}
```

**レスポンス:**

```json
{
  "id": "group-id",
  "name": "Development Team",
  "description": "Development team group",
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z"
}
```

#### GET /api/user-groups

ユーザーグループ一覧取得

**認証:** 必要

**レスポンス:**

```json
{
  "groups": [
    {
      "id": "group-id",
      "name": "Development Team",
      "description": "Development team group",
      "createdAt": "2024-01-01T00:00:00Z",
      "updatedAt": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### GET /api/user-groups/:id

ユーザーグループ詳細取得

**認証:** 必要

**レスポンス:**

```json
{
  "id": "group-id",
  "name": "Development Team",
  "description": "Development team group",
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z"
}
```

#### PATCH /api/user-groups/:id

ユーザーグループ更新

**認証:** 必要

**リクエスト:**

```json
{
  "name": "Updated Group Name",
  "description": "Updated description"
}
```

**レスポンス:**

```json
{
  "id": "group-id",
  "name": "Updated Group Name",
  "description": "Updated description",
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z"
}
```

#### DELETE /api/user-groups/:id

ユーザーグループ削除

**認証:** 必要

**レスポンス:**

```json
{
  "message": "User group deleted successfully"
}
```

#### POST /api/user-groups/:id/members

グループメンバー追加

**認証:** 必要

**リクエスト:**

```json
{
  "userID": "user-id"
}
```

**レスポンス:**

```json
{
  "message": "Member added to group successfully"
}
```

#### DELETE /api/user-groups/:id/members

グループメンバー削除

**認証:** 必要

**リクエスト:**

```json
{
  "userID": "user-id"
}
```

**レスポンス:**

```json
{
  "message": "Member removed from group successfully"
}
```

#### GET /api/user-groups/:id/members

グループメンバー一覧

**認証:** 必要

**レスポンス:**

```json
{
  "members": [
    {
      "id": "user-id",
      "email": "user@example.com",
      "displayName": "User Name",
      "avatarURL": null
    }
  ]
}
```

### リンク (Links)

#### POST /api/links/fetch-ogp

OGP 情報取得

**認証:** 必要

**リクエスト:**

```json
{
  "url": "https://example.com"
}
```

**レスポンス:**

```json
{
  "title": "Example Site",
  "description": "Example description",
  "image": "https://example.com/image.jpg",
  "url": "https://example.com"
}
```

### WebSocket

#### GET /ws

WebSocket 接続

**認証:** 必要 (クエリパラメータまたはヘッダー)

**クエリパラメータ:**

- `workspaceId`: ワークスペース ID
- `token`: JWT トークン

**接続例:**

```
ws://localhost:8080/ws?workspaceId=workspace-id&token=jwt-token
```

**メッセージ形式:**

```json
{
  "type": "message",
  "content": "Hello, World!",
  "channelID": "channel-id"
}
```

### ヘルスチェック

#### GET /healthz

ヘルスチェック

**認証:** 不要

**レスポンス:**

```
ok
```

## エラーレスポンス

### 標準エラーレスポンス

```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": "Additional error details"
}
```

### HTTP ステータスコード

- `200 OK`: 成功
- `201 Created`: 作成成功
- `400 Bad Request`: リクエストエラー
- `401 Unauthorized`: 認証エラー
- `403 Forbidden`: 認可エラー
- `404 Not Found`: リソースが見つからない
- `409 Conflict`: 競合エラー
- `500 Internal Server Error`: サーバーエラー

### エラーコード一覧

- `INVALID_CREDENTIALS`: 認証情報が無効
- `USER_ALREADY_EXISTS`: ユーザーが既に存在
- `INVALID_TOKEN`: トークンが無効
- `NOT_FOUND`: リソースが見つからない
- `UNAUTHORIZED`: 認証が必要
- `FORBIDDEN`: アクセス権限なし
- `VALIDATION_ERROR`: バリデーションエラー
- `INTERNAL_ERROR`: 内部エラー

## レート制限

- **認証エンドポイント**: 5 回/分
- **一般エンドポイント**: 100 回/分
- **WebSocket**: 制限なし

## バージョニング

現在の API バージョン: v1

将来のバージョンアップ時は、後方互換性を保ちながら段階的に移行します。
