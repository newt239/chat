# チャット UI 機能拡張 実装計画

## 概要

Slack ライクなチャット機能を実装するため、以下の機能を追加します:

1. メッセージに投稿者のアイコンと名前を表示
2. メッセージホバー時のアクションメニュー表示
3. メッセージへのリンクコピー、リアクション追加、スレッド作成、ブックマーク機能

## 現状分析

### 既存の実装

- **フロントエンド**

  - `MessagePanel.tsx`: メッセージ一覧表示とメッセージ送信 UI
  - 現在のメッセージ表示: 日時とメッセージ本文のみ
  - ユーザー情報の表示なし

- **バックエンド**
  - `Message` ドメインモデル: `UserID` フィールドあり
  - `User` ドメインモデル: `DisplayName`, `AvatarURL` フィールドあり
  - `MessageReaction` モデル: リアクション機能の基盤あり
  - スレッド機能: `ParentID` フィールドで対応可能

### 必要な変更

メッセージレスポンスにユーザー情報が含まれていない可能性があるため、バックエンドとフロントエンドの両方で対応が必要です。

## 実装計画

### フェーズ 1: バックエンド - メッセージレスポンスの拡張

#### 1.1 DTO の拡張

**ファイル**: `backend/internal/usecase/message/dto.go`

```go
type MessageOutput struct {
    ID        string     `json:"id"`
    ChannelID string     `json:"channelId"`
    UserID    string     `json:"userId"`
    User      UserInfo   `json:"user"`  // 追加
    ParentID  *string    `json:"parentId,omitempty"`
    Body      string     `json:"body"`
    CreatedAt time.Time  `json:"createdAt"`
    EditedAt  *time.Time `json:"editedAt,omitempty"`
}

type UserInfo struct {
    ID          string  `json:"id"`
    DisplayName string  `json:"displayName"`
    AvatarURL   *string `json:"avatarUrl,omitempty"`
}
```

#### 1.2 UseCase の修正

**ファイル**: `backend/internal/usecase/message/interactor.go`

- `ListMessages` メソッドを修正して、各メッセージのユーザー情報を取得
- UserRepository を使用してユーザー情報を取得
- 効率化のため、まとめて取得する実装を検討

#### 1.3 Repository の拡張 (必要に応じて)

**ファイル**: `backend/internal/domain/user.go`

```go
// UserRepository に追加
FindByIDs(ids []string) ([]*User, error)
```

複数ユーザーを一度に取得するメソッドを追加することで、N+1 問題を回避します。

### フェーズ 2: フロントエンド - メッセージ表示コンポーネントの作成

#### 2.1 Message コンポーネントの分離

**新規ファイル**: `frontend/src/features/message/components/MessageItem.tsx`

```tsx
interface MessageItemProps {
  message: Message;
  dateTimeFormatter: Intl.DateTimeFormat;
  onCopyLink: (messageId: string) => void;
  onAddReaction: (messageId: string) => void;
  onCreateThread: (messageId: string) => void;
  onBookmark: (messageId: string) => void;
}

export const MessageItem = ({ ... }: MessageItemProps) => {
  const [isHovered, setIsHovered] = useState(false);

  return (
    <div
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      className="relative group rounded-md px-4 py-2 transition-colors hover:bg-gray-50"
    >
      {/* アバター */}
      <div className="flex gap-3">
        <Avatar
          src={message.user.avatarUrl}
          alt={message.user.displayName}
          size="md"
        />

        {/* メッセージコンテンツ */}
        <div className="flex-1 min-w-0">
          {/* ヘッダー: 名前と日時 */}
          <div className="flex items-baseline gap-2">
            <Text fw={600} size="sm">
              {message.user.displayName}
            </Text>
            <Text size="xs" c="dimmed">
              {dateTimeFormatter.format(new Date(message.createdAt))}
            </Text>
          </div>

          {/* メッセージ本文 */}
          <Text className="mt-1 whitespace-pre-wrap wrap-break-word text-sm">
            {message.body}
          </Text>
        </div>
      </div>

      {/* ホバー時のアクションメニュー */}
      {isHovered && (
        <MessageActions
          messageId={message.id}
          onCopyLink={onCopyLink}
          onAddReaction={onAddReaction}
          onCreateThread={onCreateThread}
          onBookmark={onBookmark}
        />
      )}
    </div>
  );
};
```

#### 2.2 アクションメニューコンポーネント

**新規ファイル**: `frontend/src/features/message/components/MessageActions.tsx`

```tsx
interface MessageActionsProps {
  messageId: string;
  onCopyLink: (messageId: string) => void;
  onAddReaction: (messageId: string) => void;
  onCreateThread: (messageId: string) => void;
  onBookmark: (messageId: string) => void;
}

export const MessageActions = ({ ... }: MessageActionsProps) => {
  return (
    <div className="absolute right-4 top-2 flex gap-1 rounded-md border bg-white p-1 shadow-sm">
      <ActionIcon
        variant="subtle"
        size="sm"
        onClick={() => onAddReaction(messageId)}
        title="リアクションを追加"
      >
        <IconMoodSmile size={16} />
      </ActionIcon>

      <ActionIcon
        variant="subtle"
        size="sm"
        onClick={() => onCreateThread(messageId)}
        title="スレッドで返信"
      >
        <IconMessage size={16} />
      </ActionIcon>

      <ActionIcon
        variant="subtle"
        size="sm"
        onClick={() => onBookmark(messageId)}
        title="ブックマークに追加"
      >
        <IconBookmark size={16} />
      </ActionIcon>

      <Menu position="bottom-end">
        <Menu.Target>
          <ActionIcon variant="subtle" size="sm">
            <IconDots size={16} />
          </ActionIcon>
        </Menu.Target>
        <Menu.Dropdown>
          <Menu.Item
            leftSection={<IconLink size={14} />}
            onClick={() => onCopyLink(messageId)}
          >
            リンクをコピー
          </Menu.Item>
          <Menu.Item
            leftSection={<IconEdit size={14} />}
          >
            メッセージを編集
          </Menu.Item>
          <Menu.Item
            leftSection={<IconTrash size={14} />}
            c="red"
          >
            メッセージを削除
          </Menu.Item>
        </Menu.Dropdown>
      </Menu>
    </div>
  );
};
```

#### 2.3 MessagePanel の修正

**ファイル**: `frontend/src/features/message/components/MessagePanel.tsx`

- `MessageItem` コンポーネントを使用するように変更
- アクションハンドラーを実装:
  - `handleCopyLink`: メッセージ URL をクリップボードにコピー
  - `handleAddReaction`: リアクション追加モーダル表示
  - `handleCreateThread`: スレッドビューへ遷移
  - `handleBookmark`: ブックマーク機能 (将来実装)

### フェーズ 3: フロントエンド - アクション機能の実装

#### 3.1 メッセージリンクコピー機能

```tsx
const handleCopyLink = useCallback(
  (messageId: string) => {
    const url = `${window.location.origin}/app/${workspaceId}/${channelId}?message=${messageId}`;
    navigator.clipboard.writeText(url);
    notifications.show({
      title: "コピーしました",
      message: "メッセージリンクをクリップボードにコピーしました",
    });
  },
  [workspaceId, channelId]
);
```

#### 3.2 リアクション追加モーダル

**新規ファイル**: `frontend/src/features/message/components/ReactionPicker.tsx`

- Emoji Picker ライブラリの選定と統合 (例: `@emoji-mart/react`)
- リアクション追加 API の呼び出し

**新規ファイル**: `frontend/src/features/message/hooks/useReaction.ts`

```tsx
export function useAddReaction() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({
      messageId,
      emoji,
    }: {
      messageId: string;
      emoji: string;
    }) => {
      const { data, error } = await apiClient.POST(
        "/api/messages/{messageId}/reactions",
        {
          params: { path: { messageId } },
          body: { emoji },
        }
      );

      if (error || !data) {
        throw new Error(error?.error ?? "リアクションの追加に失敗しました");
      }

      return data;
    },
    onSuccess: () => {
      // メッセージ一覧を再取得
      queryClient.invalidateQueries({ queryKey: ["channels"] });
    },
  });
}
```

#### 3.3 スレッド機能

**新規ファイル**: `frontend/src/features/message/components/ThreadPanel.tsx`

- スレッドのメッセージ一覧表示
- スレッド内でのメッセージ送信

**新規ファイル**: `frontend/src/features/message/hooks/useThread.ts`

```tsx
export function useThreadMessages(parentMessageId: string | null) {
  return useQuery({
    queryKey: ["messages", parentMessageId, "thread"],
    queryFn: async () => {
      if (!parentMessageId) return [];

      const { data, error } = await apiClient.GET(
        "/api/messages/{messageId}/thread",
        {
          params: { path: { messageId: parentMessageId } },
        }
      );

      if (error || !data) {
        throw new Error(error?.error ?? "スレッドの取得に失敗しました");
      }

      return data;
    },
    enabled: parentMessageId !== null,
  });
}
```

### フェーズ 4: バックエンド - 新規 API エンドポイントの追加

#### 4.1 リアクション関連 API

**ファイル**: `backend/internal/interface/http/handler/message_handler.go`

```go
// POST /api/messages/{messageId}/reactions
func (h *MessageHandler) AddReaction(c echo.Context) error

// DELETE /api/messages/{messageId}/reactions/{emoji}
func (h *MessageHandler) RemoveReaction(c echo.Context) error

// GET /api/messages/{messageId}/reactions
func (h *MessageHandler) GetReactions(c echo.Context) error
```

#### 4.2 スレッド関連 API

```go
// GET /api/messages/{messageId}/thread
func (h *MessageHandler) GetThreadReplies(c echo.Context) error
```

#### 4.3 ブックマーク関連 (将来実装)

ブックマーク機能は優先度が低いため、後回しにします。

### フェーズ 5: UI/UX の調整

#### 5.1 アバター表示の最適化

- 連続する同一ユーザーのメッセージは、アバターを省略して日時のみ表示 (Slack 風)
- 時間の経過が大きい場合は、日付セパレーターを表示

#### 5.2 レスポンシブ対応

- モバイル表示時のアクションメニューの調整
- タッチデバイスでのホバー動作の代替 (長押しなど)

#### 5.3 アニメーション

- アクションメニューのフェードイン/アウト
- リアクション追加時のアニメーション

## 実装順序

1. **バックエンド - メッセージレスポンスの拡張** (フェーズ 1)

   - DTO の拡張
   - UseCase の修正
   - Repository の拡張

2. **フロントエンド - 基本的な UI 構築** (フェーズ 2)

   - MessageItem コンポーネント作成
   - MessageActions コンポーネント作成
   - MessagePanel の修正

3. **メッセージリンクコピー機能** (フェーズ 3.1)

   - クリップボード API の実装
   - 通知機能の統合

4. **リアクション機能** (フェーズ 4.1 + フェーズ 3.2)

   - バックエンド API 実装
   - フロントエンド UI 実装
   - Emoji Picker の統合

5. **スレッド機能** (フェーズ 4.2 + フェーズ 3.3)

   - バックエンド API 実装
   - ThreadPanel コンポーネント実装
   - ルーティングの調整

6. **UI/UX の最終調整** (フェーズ 5)
   - アバター表示の最適化
   - レスポンシブ対応
   - アニメーション追加

## 技術的考慮事項

### 依存ライブラリ

- **@mantine/core**: 既存 UI コンポーネント
- **@tabler/icons-react**: アイコン
- **@emoji-mart/react** (新規追加): Emoji Picker
- **@tanstack/react-query**: データフェッチング (既存)

### パフォーマンス最適化

- ユーザー情報の取得: N+1 問題を避けるため、一括取得を実装
- メッセージリストの仮想化: 大量のメッセージがある場合、`react-virtual` などの検討
- メモ化: `useMemo`, `useCallback` を適切に使用

### アクセシビリティ

- キーボードナビゲーション対応
- ARIA 属性の適切な設定
- スクリーンリーダー対応

## テスト計画

### バックエンド

- ユニットテスト: 各 UseCase, Repository のテスト
- 統合テスト: API エンドポイントのテスト

### フロントエンド

- コンポーネントテスト: MessageItem, MessageActions のテスト
- 統合テスト: MessagePanel 全体の動作テスト
- E2E テスト: ユーザーフローのテスト (Playwright など)

## 今後の拡張

- メッセージ編集機能
- メッセージ削除機能
- ブックマーク機能
- メンション機能 (@ユーザー名)
- ファイル添付機能
- リッチテキストエディタ
- メッセージ検索機能
- 既読管理機能

## まとめ

この実装計画に従って、段階的に Slack ライクなチャット機能を実装します。各フェーズを完了後、動作確認とテストを行い、次のフェーズに進みます。
