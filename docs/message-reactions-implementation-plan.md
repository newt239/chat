# メッセージリアクション機能 実装計画

## 概要

メッセージに絵文字でリアクションを追加する機能を実装します。ユーザーは絵文字ピッカーから好きな絵文字を選択してメッセージにリアクションでき、同じリアクションをクリックすることで取り消すこともできます。リアクション一覧では誰がどのリアクションをつけたか確認できるUIを提供します。また、将来的なカスタム絵文字対応を見据えた拡張可能な設計とします。

## 現状分析

### 既存の実装状況

#### バックエンド
- ✅ データモデル: `MessageReaction` (`backend/internal/domain/message.go`)
  - MessageID、UserID、Emoji、CreatedAtの情報を保持
  - 複合主キー: (message_id, user_id, emoji)
- ✅ リポジトリ層: `MessageRepository` にリアクション関連のメソッドが実装済み
  - `AddReaction(reaction *MessageReaction) error`
  - `RemoveReaction(messageID, userID, emoji string) error`
  - `FindReactions(messageID string) ([]*MessageReaction, error)`
- ✅ データベーステーブル: `message_reactions` テーブルが定義済み
- ❌ API層: リアクション用のHTTPエンドポイントが未実装
- ❌ OpenAPI定義: リアクション用のスキーマとパスが未定義

#### フロントエンド
- ✅ UI: `MessageActions` にリアクションボタンが存在（現在は動作なし）
- ❌ リアクション表示コンポーネント: 未実装
- ❌ 絵文字ピッカー: 未実装
- ❌ APIクライアント: リアクション用の型定義とAPI呼び出しが未実装
- ❌ 状態管理: リアクションの状態管理が未実装

## 実装計画

### Phase 1: バックエンドAPI実装

#### 1.1 OpenAPI定義の追加

**ファイル**: `backend/internal/openapi/openapi.yaml`

追加するスキーマ:
```yaml
MessageReaction:
  type: object
  properties:
    messageId:
      type: string
      format: uuid
    userId:
      type: string
      format: uuid
    emoji:
      type: string
      description: "Unicode絵文字または将来的にカスタム絵文字ID"
    createdAt:
      type: string
      format: date-time
  required:
    - messageId
    - userId
    - emoji
    - createdAt

# ユーザー情報を含むリアクション（表示用）
ReactionWithUser:
  type: object
  properties:
    messageId:
      type: string
      format: uuid
    user:
      type: object
      properties:
        id:
          type: string
          format: uuid
        displayName:
          type: string
        avatarUrl:
          type: string
          nullable: true
      required:
        - id
        - displayName
    emoji:
      type: string
    createdAt:
      type: string
      format: date-time
  required:
    - messageId
    - user
    - emoji
    - createdAt

AddReactionRequest:
  type: object
  properties:
    emoji:
      type: string
      minLength: 1
  required:
    - emoji

ListReactionsResponse:
  type: object
  properties:
    reactions:
      type: array
      items:
        $ref: '#/components/schemas/ReactionWithUser'
  required:
    - reactions
```

追加するエンドポイント:
```yaml
/api/messages/{messageId}/reactions:
  get:
    operationId: listReactions
    summary: List reactions for a message
    parameters:
      - name: messageId
        in: path
        required: true
        schema:
          type: string
          format: uuid
    responses:
      "200":
        description: List of reactions
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ListReactionsResponse'

  post:
    operationId: addReaction
    summary: Add a reaction to a message
    parameters:
      - name: messageId
        in: path
        required: true
        schema:
          type: string
          format: uuid
    requestBody:
      required: true
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/AddReactionRequest'
    responses:
      "201":
        description: Reaction added
      "400":
        description: Bad request

/api/messages/{messageId}/reactions/{emoji}:
  delete:
    operationId: removeReaction
    summary: Remove a reaction from a message
    parameters:
      - name: messageId
        in: path
        required: true
        schema:
          type: string
          format: uuid
      - name: emoji
        in: path
        required: true
        schema:
          type: string
    responses:
      "200":
        description: Reaction removed
```

#### 1.2 ユースケース層の実装

**新規ファイル**: `backend/internal/usecase/reaction/dto.go`
```go
type AddReactionInput struct {
    MessageID string
    UserID    string
    Emoji     string
}

type UserInfo struct {
    ID          string
    DisplayName string
    AvatarURL   *string
}

type ReactionOutput struct {
    MessageID string
    User      UserInfo
    Emoji     string
    CreatedAt time.Time
}
```

**新規ファイル**: `backend/internal/usecase/reaction/interactor.go`
- `AddReaction(input AddReactionInput) error`
- `RemoveReaction(messageID, userID, emoji string) error`
- `ListReactions(messageID string) ([]*ReactionOutput, error)`
  - ユーザー情報をJOINして取得し、ReactionOutputに含める

#### 1.3 HTTPハンドラーの実装

**新規ファイル**: `backend/internal/interface/http/handler/reaction_handler.go`
- `HandleListReactions(c *gin.Context) error`
- `HandleAddReaction(c *gin.Context) error`
- `HandleRemoveReaction(c *gin.Context) error`

**更新ファイル**: `backend/internal/interface/http/router.go`
- リアクション用のルートを追加

### Phase 2: フロントエンド基盤実装

#### 2.1 API型定義の生成

```bash
pnpm run generate:api
```

#### 2.2 絵文字ピッカーライブラリの追加

`@emoji-mart/react` と `@emoji-mart/data` を使用:
```bash
pnpm add @emoji-mart/react @emoji-mart/data
```

#### 2.3 型定義の追加

**新規ファイル**: `frontend/src/features/reaction/types.ts`
```typescript
import type { components } from "@/lib/api/schema";

export type ReactionWithUser = components["schemas"]["ReactionWithUser"];

export interface UserInfo {
  id: string;
  displayName: string;
  avatarUrl?: string | null;
}

export interface ReactionGroup {
  emoji: string;
  count: number;
  users: UserInfo[]; // ユーザー情報の配列
  hasUserReacted: boolean;
}

// 将来的なカスタム絵文字対応のための型
export interface EmojiData {
  id: string; // Unicode絵文字の場合は絵文字自体、カスタムの場合はID
  native?: string; // Unicode絵文字
  imageUrl?: string; // カスタム絵文字の画像URL
  name: string; // 絵文字の名前
  isCustom: boolean; // カスタム絵文字かどうか
}
```

#### 2.4 API呼び出しフックの実装

**新規ファイル**: `frontend/src/features/reaction/hooks/useReactions.ts`
- `useReactions(messageId: string)`: リアクション一覧を取得
- `useAddReaction()`: リアクションを追加
- `useRemoveReaction()`: リアクションを削除

```typescript
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient } from "@/lib/api/client";

export const useReactions = (messageId: string) => {
  return useQuery({
    queryKey: ["reactions", messageId],
    queryFn: async () => {
      const { data, error } = await apiClient.GET(
        "/api/messages/{messageId}/reactions",
        { params: { path: { messageId } } }
      );
      if (error) throw error;
      return data;
    },
  });
};

export const useAddReaction = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ messageId, emoji }: { messageId: string; emoji: string }) => {
      const { error } = await apiClient.POST(
        "/api/messages/{messageId}/reactions",
        {
          params: { path: { messageId } },
          body: { emoji },
        }
      );
      if (error) throw error;
    },
    onSuccess: (_, { messageId }) => {
      queryClient.invalidateQueries({ queryKey: ["reactions", messageId] });
    },
  });
};

export const useRemoveReaction = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ messageId, emoji }: { messageId: string; emoji: string }) => {
      const { error } = await apiClient.DELETE(
        "/api/messages/{messageId}/reactions/{emoji}",
        { params: { path: { messageId, emoji } } }
      );
      if (error) throw error;
    },
    onSuccess: (_, { messageId }) => {
      queryClient.invalidateQueries({ queryKey: ["reactions", messageId] });
    },
  });
};
```

### Phase 3: フロントエンドUI実装

#### 3.1 リアクション表示コンポーネント

**新規ファイル**: `frontend/src/features/reaction/components/ReactionList.tsx`

機能:
- リアクションをグループ化して表示（同じ絵文字をまとめる）
- 各リアクションに絵文字と数を表示
- ユーザーが既にリアクションしている場合は視覚的に区別（背景色など）
- クリックで追加/削除をトグル
- **ホバー時にツールチップでリアクションしたユーザー一覧を表示**
  - ユーザー名を改行区切りで表示
  - 例: 「太郎さん、花子さん、次郎さん」

デザイン:
```
[😀 3] [👍 5] [🎉 2]
   ↑ホバーで「太郎さん、花子さん、次郎さん」と表示
```

**新規ファイル**: `frontend/src/features/reaction/components/ReactionButton.tsx`

機能:
- 個別のリアクションボタンコンポーネント
- 絵文字、カウント、アクティブ状態を表示
- Mantineの`Tooltip`を使用してユーザー一覧を表示
- カスタム絵文字対応を考慮した設計（画像表示対応）

#### 3.2 絵文字ピッカーコンポーネント

**新規ファイル**: `frontend/src/features/reaction/components/EmojiPicker.tsx`

機能:
- `@emoji-mart/react` の `Picker` コンポーネントをラップ
- 絵文字選択時のコールバック
- Popover/Modal での表示対応
- 将来的なカスタム絵文字対応のための拡張ポイント
  - カスタム絵文字データを受け取れるprops設計
  - カスタムカテゴリを追加できる構造

#### 3.3 MessageItemコンポーネントの更新

**更新ファイル**: `frontend/src/features/message/components/MessageItem.tsx`

変更内容:
- `ReactionList` コンポーネントをメッセージ本文の下に配置
- リアクション表示領域を追加

#### 3.4 MessageActionsの更新

**更新ファイル**: `frontend/src/features/message/components/MessageActions.tsx`

変更内容:
- リアクションボタンクリック時に絵文字ピッカーを表示
- Popoverを使用してピッカーを配置

### Phase 4: テスト実装

#### 4.1 フロントエンドテスト

**新規ファイル**: `frontend/src/features/reaction/components/ReactionList.test.tsx`
- リアクションが正しく表示されること
- ユーザー自身のリアクションが強調表示されること
- クリックで追加/削除がトグルされること
- ツールチップにユーザー一覧が表示されること

**新規ファイル**: `frontend/src/features/reaction/components/ReactionButton.test.tsx`
- リアクションボタンが正しくレンダリングされること
- ホバー時にツールチップが表示されること
- クリック時にコールバックが呼ばれること

**新規ファイル**: `frontend/src/features/reaction/components/EmojiPicker.test.tsx`
- 絵文字ピッカーが正しく表示されること
- 絵文字選択時にコールバックが呼ばれること

**新規ファイル**: `frontend/src/features/reaction/hooks/useReactions.test.ts`
- リアクション取得のテスト
- リアクション追加のテスト
- リアクション削除のテスト

#### 4.2 バックエンドテスト

**新規ファイル**: `backend/internal/usecase/reaction/interactor_test.go`
- リアクション追加のテスト
- リアクション削除のテスト
- リアクション取得のテスト
- 重複リアクション追加のエラーハンドリング

### Phase 5: 統合とリファインメント

#### 5.1 型チェックとLint
```bash
cd frontend
pnpm run type-check
pnpm run lint
```

#### 5.2 E2Eテストシナリオ（手動）
1. メッセージにカーソルを合わせてリアクションボタンが表示されることを確認
2. リアクションボタンをクリックして絵文字ピッカーが表示されることを確認
3. 絵文字を選択してメッセージにリアクションが追加されることを確認
4. 追加されたリアクションがメッセージ下部に表示されることを確認
5. **リアクションボタンにホバーして、リアクションしたユーザー一覧が表示されることを確認**
6. 同じリアクションをもう一度クリックして削除されることを確認
7. 他のユーザーのリアクションと自分のリアクションが区別されることを確認
8. **複数ユーザーが同じリアクションをした場合、カウントが正しく増加し、ツールチップに全ユーザーが表示されることを確認**

## データフロー

```
ユーザー操作
    ↓
[EmojiPicker] 絵文字選択
    ↓
[useAddReaction] API呼び出し
    ↓
[Backend] POST /api/messages/{messageId}/reactions
    ↓
[ReactionInteractor] ビジネスロジック
    ↓
[MessageRepository] データベース保存
    ↓
[Frontend] React Query キャッシュ更新
    ↓
[ReactionList] UI再レンダリング
```

## 考慮事項

### セキュリティ
- リアクション追加時は認証必須（JWT）
- ユーザーは自分のリアクションのみ削除可能
- SQLインジェクション対策（prepared statement使用）
- XSS対策（絵文字のサニタイゼーション）

### パフォーマンス
- リアクション取得はメッセージ読み込み時に一括取得を検討
- TanStack Query によるキャッシング活用
- リアクション数が多い場合の表示制限（例: 上位5個まで表示）
- ユーザー情報のJOIN処理の最適化（N+1問題の回避）

### UX
- リアクション追加時のローディング状態表示
- 楽観的更新（Optimistic Update）の検討
- エラー時のトースト通知
- アニメーション効果（追加/削除時）

### アクセシビリティ
- キーボード操作対応（Tab、Enter）
- スクリーンリーダー対応（aria-label）
- フォーカス管理

## タスクリスト

### バックエンド
- [ ] OpenAPI定義の追加
- [ ] API型の再生成確認
- [ ] Usecaseインターフェース実装
- [ ] Usecaseインターフェースのテスト実装
- [ ] HTTPハンドラー実装
- [ ] ルーティング設定
- [ ] 統合テスト

### フロントエンド
- [ ] 絵文字ピッカーライブラリのインストール
- [ ] 型定義の作成（カスタム絵文字対応を考慮）
- [ ] useReactions フックの実装
- [ ] useAddReaction フックの実装
- [ ] useRemoveReaction フックの実装
- [ ] ReactionButton コンポーネントの実装
- [ ] ReactionButton のテスト実装
- [ ] ReactionList コンポーネントの実装
- [ ] ReactionList のテスト実装
- [ ] EmojiPicker コンポーネントの実装（カスタム絵文字拡張ポイント）
- [ ] EmojiPicker のテスト実装
- [ ] MessageItem の更新
- [ ] MessageActions の更新
- [ ] 型チェック実行
- [ ] Lint実行
- [ ] E2E手動テスト（ユーザー一覧表示を含む）

## 実装順序

1. **バックエンドAPI実装** (Phase 1)
   - OpenAPI定義 → Usecase → Handler → Router の順で実装

2. **フロントエンド基盤実装** (Phase 2)
   - API型生成 → ライブラリ追加 → フック実装

3. **フロントエンドUI実装** (Phase 3)
   - ReactionButton → ReactionList → EmojiPicker → 既存コンポーネント更新

4. **テスト実装** (Phase 4)
   - フロントエンド → バックエンドの順でテスト実装

5. **統合確認** (Phase 5)
   - 型チェック → Lint → E2Eテスト

## 見積もり

- Phase 1: 5-7時間（ユーザー情報JOIN処理追加のため）
- Phase 2: 2-3時間
- Phase 3: 5-6時間（ReactionButton、ツールチップ実装追加のため）
- Phase 4: 4-5時間（テストケース追加のため）
- Phase 5: 1-2時間

**合計**: 17-23時間

## カスタム絵文字対応の設計方針

将来的なカスタム絵文字機能追加を見据えて、以下の拡張ポイントを設けます。

### データベース設計
現在のスキーマでは `emoji` カラムに文字列を保存していますが、将来的には以下の対応が可能です：
- Unicode絵文字の場合: 絵文字文字列をそのまま保存（例: "😀"）
- カスタム絵文字の場合: カスタム絵文字IDを保存（例: "custom:team_logo"）
- プレフィックスで種別を判定できる構造

### 将来的な拡張テーブル（参考）
```sql
-- カスタム絵文字マスターテーブル（将来実装）
CREATE TABLE custom_emojis (
    id UUID PRIMARY KEY,
    workspace_id UUID NOT NULL REFERENCES workspaces(id),
    name VARCHAR(255) NOT NULL, -- 絵文字名（例: "team_logo"）
    image_url TEXT NOT NULL, -- 画像URL
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(workspace_id, name)
);
```

### フロントエンド設計
- `EmojiData` 型で Unicode/カスタムを区別
- `ReactionButton` コンポーネントで画像表示に対応
- `EmojiPicker` でカスタム絵文字カテゴリを追加可能な構造

### 実装時の考慮事項
1. **絵文字の表示**: Unicode絵文字と画像の両方をサポート
2. **ピッカーの拡張**: カスタム絵文字タブの追加
3. **APIレスポンス**: カスタム絵文字情報（画像URL等）を含める
4. **権限管理**: ワークスペース単位でカスタム絵文字を管理

## リスクと対策

| リスク | 対策 |
|--------|------|
| 絵文字ピッカーライブラリの互換性問題 | 事前に動作確認、代替ライブラリの選定 |
| リアクション数の増加によるパフォーマンス低下 | ページネーションまたは表示制限の実装 |
| 複数ユーザーの同時リアクションによる競合 | データベースの制約で重複防止 |
| WebSocket未実装によるリアルタイム更新なし | 将来的な拡張として記録、現状はポーリングまたは手動更新 |
| ツールチップのユーザー数が多い場合の表示問題 | 一定数以上は「他○名」と省略表示 |
| ユーザー情報JOIN処理のパフォーマンス | 適切なインデックス設定とクエリ最適化 |
