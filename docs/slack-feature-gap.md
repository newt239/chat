# Slackライク機能の未実装まとめ

現状のコードベースを確認した結果、Slackライクなコミュニケーション体験に必須となるが、まだ実装が完了していない（もしくは未接続の）主要機能を以下に整理する。

## 1. リアルタイム更新の未接続
- WebSocket クライアントは実装済みだが、他コンポーネントから利用されておらずリアルタイム通知が前段に届かない。`rg` でも `WebSocketClient` の参照は宣言元のみである。`frontend/src/lib/ws/client.ts:21`
- ハブ側もチャンネル単位のブロードキャストが未完成で、`TODO` が残置されている。`backend/internal/interfaces/handler/websocket/hub.go:70`

**影響:** 新着メッセージ・リアクション・既読更新が即時反映されず、Slack のリアルタイム性を満たせない。

## 2. メッセージ編集 / 削除エンドポイント未公開
- ルーターはメッセージ関連で `GET` / `POST` のみ定義し、`PATCH` / `DELETE` が存在しない。`backend/internal/interfaces/handler/http/router.go:93` `backend/internal/interfaces/handler/http/router.go:95`
- ハンドラーも更新・削除用メソッドが未実装のまま終端している。`backend/internal/interfaces/handler/http/handler/message_handler.go:44`

**影響:** フロントの編集・削除 UI (`frontend/src/features/message/components/MessagePanel.tsx:162` など) からのリクエストが必ず失敗し、Slack の基本操作に欠かせない編集 / 削除機能が利用できない。

## 3. スレッド閲覧体験の不足
- 右サイドバーのスレッド表示はプレースホルダで止まっている。`frontend/src/features/workspace/components/ThreadPanel.tsx:9`
- スレッドアクションも `console.log` のみで未接続。`frontend/src/features/message/components/ThreadSidePanel.tsx:69`
- サーバー側ルーターに `/api/messages/{messageId}/thread` 系エンドポイントがなく、フロントが呼び出す API が存在しない。`backend/internal/interfaces/handler/http/router.go:93`

**影響:** Slack の核心機能であるスレッド会話を開いても中身が表示されず、ユーザーは返信のコンテキストを追えない。

## 4. 1対1 / グループ DM の欠如
- ドメイン層にはワークスペース内チャンネルのみが定義され、ユーザー間 DM を表すエンティティが存在しない。`backend/internal/domain/entity/channel.go:18`
- ルーターにも DM 作成・取得用のエンドポイントがない。`backend/internal/interfaces/handler/http/router.go:81`

**影響:** Slack で求められるダイレクトメッセージ体験が提供できず、個人や小規模グループでの会話が不可能。

## 5. 未読カウント / 通知表示の未実装
- WebSocket には `unread_count` イベントが定義されているが、フロント側で購読・表示されていない。`frontend/src/lib/ws/client.ts:10`
- チャンネルリストも未読数を表示しておらず、単に名称を羅列するのみ。`frontend/src/features/channel/components/ChannelList.tsx:82`

**影響:** どのチャンネルに未読が残っているか判別できず、Slack の通知体験が崩れる。

## 6. ブックマーク一覧からの遷移が未完成
- ブックマーク一覧でワークスペース ID がハードコード (`"current"`) されており、実際のチャンネルへ遷移できない。`frontend/src/features/bookmark/components/BookmarkList.tsx:11`
- メッセージ面のハンドラーも TODO のまま。`frontend/src/features/message/components/MessagePanel.tsx:157`

**影響:** ブックマークから対象メッセージへジャンプするという Slack 標準の導線が機能しない。

## 7. サーバーサイド検索の未実装
- 現状の検索はワークスペース内の全チャンネル・全メッセージをフロントで逐次取得してローカル検索しており、大規模データに耐えない。`frontend/src/features/search/hooks/useWorkspaceSearchIndex.ts:30`

**影響:** データ量の増加に伴い UI が極端に重くなり、Slack のような高速全文検索体験を提供できない。
