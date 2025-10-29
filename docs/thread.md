<!-- e237226a-9700-448d-b8e7-b52b646f759a 2761b91e-dbe1-49f9-b9ce-05380d40f572 -->
# 参加中スレッド一覧 実装計画

## 要件要約

- 対象: 選択中ワークスペース内の全チャンネル＋DMを横断し、「自分が参加中」のスレッドのみ
  - 参加中定義: 明示フォロー、メンションされた、返信した
- 並び順: lastActivityAt（最新返信時刻）降順
- ページネーション: Cursor-based（page size=20）
- 表示: CenterPanelにスレッドカード（スレッド最初のメッセージ＋返信件数＋新着数）
- 導線: LeftSidePanelのチャンネル一覧の最上部に「スレッド」リンク

## バックエンド（Go, Clean Architecture）

- データモデル（ent／ドメイン）
  - スレッドの定義: 親メッセージ（`parent_id == null`）をスレッド起点とし、子メッセージ群で構成
  - 参加中の判定用
    - 既存: 返信した（message.author_id == me）
    - メンション: メッセージのメンション抽出結果があればそれを参照（なければ`mentions`中間テーブルを追加）
    - 明示フォロー: `user_thread_follow(user_id, thread_id, created_at)` テーブルを追加
  - 既読管理（新着数算出）
    - `thread_read_state(user_id, thread_id, last_read_at)` を追加
  - パフォーマンス
    - `message(thread_id, created_at)`、`thread(last_activity_at)`（更新時にスレッドへロールアップ）へインデックス

- ユースケース
  - `internal/usecase/thread/lister.go`
    - 入力: `workspaceID, userID, cursor(lastActivityAt, threadId), limit=20`
    - 参加中フィルタ（UNION/OR）
      - フォロー中
      - 返信した（同スレッドで author_id==userID のメッセージ存在）
      - メンションされた（同スレッドのメッセージで mentions に userID）
    - 取得: スレッド起点メッセージ、返信件数、`lastActivityAt`
    - 新着数: `count(messages.created_at > read_state.last_read_at)`
    - カーソル: `(lastActivityAt DESC, threadId DESC)` タイブレーク

- リポジトリ
  - `internal/domain/repository/thread_repository.go`（取得API定義）
  - `internal/infrastructure/repository/thread_repository.go`（ent実装）

- ハンドラ / ルーティング
  - `GET /api/workspaces/{workspaceId}/threads/participating`
    - クエリ: `cursorLastActivityAt?, cursorThreadId?, limit?`
    - レスポンス: `items: [{threadId, channelId|dmId, firstMessage, replyCount, lastActivityAt, unreadCount}], nextCursor?`
  - 既読更新: `POST /api/threads/{threadId}/read`（`lastReadAt`をサーバ時刻に）

- OpenAPI
  - `backend/internal/openapi/openapi.yaml`に上記スキーマ追加
  - フロントのスキーマ再生成前提

## フロントエンド（React/TS, TanStack Router）

- 画面
  - ルート: `frontend/src/routes/app/$workspaceId/threads.tsx`
  - コンポーネント
    - `ThreadListPage.tsx`（CenterPanel格納）
    - `ThreadCard.tsx`（最初のメッセージ、返信数、新着数、最終更新時刻）
  - スタイルは既存カードに準拠

- 導線
  - `frontend/src/features/layout/components/LeftSidePanel.tsx`
    - チャンネル一覧の一番上に `スレッド` エントリを追加（選択時に上記ルートへ）

- データ取得 / 型
  - `frontend/src/features/thread/schemas.ts`（zod）
  - `frontend/src/features/thread/hooks/useParticipatingThreads.ts`（cursor取得）
  - APIクライアント: `pnpm run generate:api`（OpenAPIから）

- ページネーションUI
  - 「さらに読み込む」ボタン or ビューポート下での無限スクロール（まずはボタン）
  - カーソルは`{ lastActivityAt, threadId }`を保持

- 既読
  - カードクリック時に `POST /api/threads/{threadId}/read`
  - カードの新着数を0に更新

- 型/Lint/テスト
  - ルール順守（`type`のみ、any禁止、絶対パスimport、1ファイル1コンポーネント）
  - 新規コンポーネントは Vitest で `*.test.tsx` を同階層に
  - 実装後に `npx tsc --noEmit && pnpm run lint:fix`

## 代表的な変更ファイル

- バックエンド
  - `backend/internal/domain/repository/thread_repository.go`
  - `backend/internal/infrastructure/repository/thread_repository.go`
  - `backend/internal/usecase/thread/lister.go`
  - `backend/internal/interfaces/handler/http/handler/thread_handler.go`
  - `backend/internal/interfaces/handler/http/router.go`
  - `backend/internal/openapi/openapi.yaml`
  - entスキーマ（必要に応じ `thread`, `user_thread_follow`, `thread_read_state`, `mentions`）

- フロントエンド
  - `frontend/src/routes/app/$workspaceId/threads.tsx`
  - `frontend/src/features/thread/components/ThreadListPage.tsx`
  - `frontend/src/features/thread/components/ThreadCard.tsx`
  - `frontend/src/features/thread/hooks/useParticipatingThreads.ts`
  - `frontend/src/features/thread/schemas.ts`
  - `frontend/src/features/layout/components/LeftSidePanel.tsx`

## API レスポンス例（簡略）

```json
{
  "items": [
    {
      "threadId": "t_123",
      "channelId": "c_1",
      "dmId": null,
      "firstMessage": { "id": "m_1", "text": "仕様相談です", "author": {"id":"u1","name":"Alice"}, "createdAt": "..." },
      "replyCount": 4,
      "lastActivityAt": "...",
      "unreadCount": 2
    }
  ],
  "nextCursor": { "lastActivityAt": "...", "threadId": "t_120" }
}
```

## 留意点

- DM/チャンネル横断時の権限: 所属していないプライベートチャンネルのスレッドは除外
- メンション抽出が未実装なら暫定で`@userId`文字列検知→後で置換可能な実装に
- タイブレーク: `lastActivityAt DESC, threadId DESC`で安定ソート
- 既存検索・DM実装と重複するDAOは安易に共通化しない

### To-dos

- [ ] スキーマ追加（follow, read_state, mentions or equivalent）
- [ ] 参加中スレッドListerユースケース実装（カーソル対応）
- [ ] ThreadRepository実装（参加中フィルタ＆集計）
- [ ] HTTPハンドラとルーティング追加（一覧・既読）
- [ ] OpenAPIにエンドポイント定義を追加
- [ ] threadsルートとページ骨組み作成
- [ ] ThreadCard／ThreadListPage実装（UI・既読送信）
- [ ] useParticipatingThreadsフック（カーソルページネーション）
- [ ] zodスキーマ定義とAPI生成反映
- [ ] BEユースケース・リポジトリ・ハンドラの統合テスト
- [ ] FEコンポーネントとフックのVitest追加
- [ ] 型チェックとlint実行（tsc noEmit, lint:fix）