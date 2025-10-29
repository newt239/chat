<!-- 4b579cd8-a168-43ef-9234-84f68f2f2ec2 87e9fde9-2996-4206-8417-ee1f16f82477 -->
# チャンネルのメッセージをピン留め機能 追加計画

## 決定事項（仕様）

- 権限: チャンネル参加者なら誰でもピン/解除可能（監査用に実行者は記録）
- 対象範囲: チャンネル内の全メッセージ（ルート/返信問わず）
- UI: チャンネルヘッダーに件数表示の導線、クリックで右サイドパネルに一覧表示
- 並び/上限: ピン日時の新しい順、最大100件
- 通知: リアルタイム更新（WebSocket/サブスクリプション）でヘッダー件数とパネルを更新（トーストは任意、初期はなし）

## バックエンド（Go, クリーンアーキテクチャ）

1) スキーマ/モデル

- 新規テーブル `pin`（ユニーク: `channel_id` + `message_id`）
- `id`(uuid), `channel_id`, `message_id`, `pinned_by`, `pinned_at`(created_at)
- ドメインに `Pin` 型を追加（backend/internal/domain/pin.go）

2) リポジトリ/ユースケース/ハンドラ

- Repository: `CreatePin(channelID, messageID, userID)`, `DeletePin(channelID, messageID)`, `ListPins(channelID, limit, cursor)`
- UseCase: 権限チェック（チャンネル参加者か）、メッセージの存在チェック
- Handler（REST）:
- POST `/channels/{channelId}/pins` body: `{ messageId }`
- DELETE `/channels/{channelId}/pins/{messageId}`
- GET `/channels/{channelId}/pins?limit=100&cursor=...`
- WebSocket/イベント: `pin.created`, `pin.deleted` をチャンネル購読者にブロードキャスト（件数・最新ピンを反映可能なペイロード）

3) OpenAPI

- 上記エンドポイントを OpenAPI に追加。レスポンスは message サマリー + `pinnedAt`/`pinnedBy` を含む。成功時 200/204、競合時 409（重複 pin）、権限なし 403。

## フロントエンド（React/TypeScript）

1) API クライアント

- スキーマ更新後に `pnpm run generate:api` を実行

2) ストア/UI 状態

- `frontend/src/providers/store/ui.ts` に右サイドパネルの `mode: 'thread' | 'pins' | null` を追加
- ピン件数カウンタを保持（チャンネル別）

3) Hooks

- `usePinnedMessages.ts`（features/pin/hooks）: 一覧取得・カーソルページング・WebSocket購読での差分反映
- `usePinActions.ts`: pin/unpin の API 呼び出しと楽観的更新

4) UI コンポーネント

- ヘッダー導線: 現行のチャンネルヘッダー（例: `ThreadPanel.tsx` 近辺のヘッダー実装箇所）にピンアイコン＋件数表示、クリックで `mode='pins'`
- 右サイドパネル: `PinnedPanel.tsx` 新規（features/pin/components）。一覧（最大100件、ピン日時の新しい順）、メッセージカードは既存 `MessageItem.tsx` の縮小表示 or 共通カードを抽出
- メッセージアクション: `MessageActions.tsx` に「ピン留め/ピン解除」を追加（状態に応じて切替）

5) 型/テスト/品質

- 型は`type`で定義、any/unknown禁止、型アサーション回避
- Vitest による単体テストを各ファイル同階層に `*.test.ts(x)` で追加
- 実装後 `npx tsc --noEmit && pnpm run lint:fix`

## データフロー（要点）

- ユーザーがメッセージから Pin: POST 実行→成功でイベント `pin.created` 受信→ヘッダー件数と `PinnedPanel` を更新
- Unpin: DELETE 実行→`pin.deleted` 受信→同様に更新
- 初回表示や再接続時は GET で再同期

## 影響範囲/移行

- 既存メッセージ/チャンネルスキーマは不変。新規 `pins` 追加のみ
- 既存サイドパネル UI に `mode='pins'` を統合（既存の `ThreadPanel.tsx` とは排他表示）
- ロールバック: `pins` を空にするだけで機能停止可能。エンドポイントを behind feature flag にすることも可

### To-dos

- [ ] pins テーブルとドメイン型を追加（ユニーク制約含む）
- [ ] Pin Repository/UseCase 実装（参加者/存在チェック含む）
- [ ] POST/DELETE/GET pins ハンドラ実装とルーティング追加
- [ ] pin.created / pin.deleted のチャンネル配信を追加
- [ ] OpenAPI に pins エンドポイントと型を追加
- [ ] フロントの API クライアント生成を更新
- [ ] 右サイドパネルに pins モードと件数状態を追加
- [ ] usePinnedMessages / usePinActions hooks を実装
- [ ] PinnedPanel と ヘッダー導線を実装
- [ ] MessageActions に ピン/解除 アクション追加
- [ ] Vitest で hooks/コンポーネント/UseCase の単体テスト追加
- [ ] 型/ビルド/リンターの最終確認（tsc, lint:fix）