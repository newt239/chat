## クライアントサイド WebSocket 実装

### 1. 構成

| ファイル | 役割 |
| --- | --- |
| `frontend/src/lib/ws.ts` | `WsClient`。接続・再接続・タブ間調停・イベント配信 |
| `frontend/src/gen/chat/v1/event_pb.ts` | `proto/chat/v1/event.proto` から生成したイベントの型 |
| `frontend/src/providers/ws/` | `WsClient` を React ツリーへ供給する Provider |

### 2. 接続管理

- エンドポイントは `ws(s)://{サーバー}/ws?token={JWT}&workspaceId={id}`。JWT はクエリパラメータで付与する
- `BroadcastChannel` でタブ間を調停し、アクティブなタブ 1 つだけが接続を保持する
- 切断時は指数バックオフで再接続する（最大 5 回）。正常終了（1000）と認証エラー（1008）では再接続しない
- サーバーから `error`（code=401）を受けた場合はログイン画面へ遷移する

### 3. イベントの購読

`WsClient` は型付きの `on(type, handler)` / `off(type, handler)` を提供する。`on` の戻り値を呼ぶと購読を解除できる。

```ts
const unsubscribe = wsClient.on("newMessage", ({ channelId, message }) => {
  // payload の型はイベント名 (ServerEvent の oneof の case) から推論される
});
```

受信したイベントは `fromJsonString(ServerEventSchema, ...)` で解析してから配信するため、想定外の形式のイベントは無視される。

### 4. 購読しているフック

| フック | 購読イベント | 反映先 |
| --- | --- | --- |
| `features/message/hooks/useChannelTimeline` | `newMessage` / `messageUpdated` / `messageDeleted` / `systemMessageCreated` / `reactionAdded` / `reactionRemoved` / `typing` / `stopTyping` | 表示中チャンネルのタイムラインと入力中インジケータ |
| `features/channel/hooks/useChannelRealtimeSync` | `newMessage` / `unreadCount` / `pinCreated` / `pinDeleted` | チャンネル一覧の未読バッジとピン件数 |
| `features/notification/hooks/useNotificationSync` | `newMessage` | 通知パネルとベルバッジ |

チャンネルの購読は `useChannelTimeline` が `joinChannel` / `leaveChannel` を送信して管理する。

### 5. 送信 API

`joinChannel` / `leaveChannel` / `typing` / `stopTyping` を提供する。
入力中の通知は `features/message/hooks/useTypingNotifier` が 2 秒間隔に間引き、5 秒操作がなければ `stopTyping` を送る。

### 6. イベントを追加するとき

1. `proto/chat/v1/event.proto` の `ServerEvent` の oneof にイベントを追加し、`pnpm run generate:proto` を実行する
2. バックエンドの `notifier.go`（または `hub.go`）でイベントを組み立てて配信する
3. `frontend/src/lib/ws.ts` の `handlers` と `eventDispatcher` に追加する（型検査で漏れが分かる）
4. 反映したいフックで `wsClient.on(...)` を購読する
