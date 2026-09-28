# 着信 Webhook

外部サービスから HTTP の POST 1 回でチャンネルに投稿するための仕組み。管理は `WebhookService`（`proto/chat/v1/webhook_service.proto`）、投稿は Connect RPC とは別の素の HTTP エンドポイントで受け付ける。

## 投稿 API

```sh
curl -X POST -H 'Content-Type: application/json' \
  -d '{"text": "デプロイが完了しました", "username": "CI", "icon_url": "https://example.com/ci.png"}' \
  https://<API のオリジン>/webhooks/<webhook_id>/<token>
```

| フィールド | 必須 | 内容 |
| --- | --- | --- |
| `text` | ○ | 本文。Markdown・メンション（`@name`、`@channel`）・リンク展開は通常の投稿と同じ。前後の空白を除いて 1〜4000 文字 |
| `username` | | この投稿だけの表示名（80 文字で切り詰め） |
| `icon_url` / `avatar_url` | | この投稿だけのアイコン。http(s) の URL のみ。`avatar_url` は Discord 互換の別名 |

- Slack の Incoming Webhook の最小形式（`text` / `username` / `icon_url`）と互換にした。`blocks` や `attachments`、スレッドへの返信（`thread_ts`）は未対応で、未知のフィールドは無視する。
- Content-Type は見ずに本文を JSON として読む（`curl -d` の既定の form 形式でも通る）。本文は 64KB まで。
- 応答は `text/plain`。成功は `200 ok`。ID かトークンの誤りは区別せず `404`、本文の検証エラーは `400`、発行者がチャンネルを閲覧できない・アーカイブ済みは `403`、レート制限は `429`（`Retry-After` 付き）。

## 設計判断

- **投稿の名義は Webhook ごとのボットユーザー**。`Message.user` は必須のまま、`user.is_bot = true` のユーザーを Webhook の作成時に作る。既存のメッセージ一覧・検索・WebSocket 配信・通知が `user` を前提にしているため、user を null にするより影響が小さい。ボットユーザーはワークスペースのメンバーにせず、照合できないパスワードと `.invalid` ドメインのメールアドレスを持たせてログインできないようにしている。Webhook を削除してもボットユーザーは消さず、過去の投稿の名義として残す。
- **表示**: `UserSummary.is_bot` が true の投稿者はプロフィールを開けないようにし、名前の横に「アプリ」タグ（North Star の `.tag`、`Badge tone="tag"`）を付ける。投稿ごとの `username` / `icon_url` は `message.sender_name` / `sender_avatar_url` に保存し、出力時に投稿者の表示名・アイコンを上書きする。Webhook の名前やアイコンを変えるとボットユーザーも更新するため、上書きのない過去の投稿も新しい名前で表示される。
- **権限**: チャンネルを閲覧できるメンバーなら誰でも発行できる（DM は不可）。編集・URL の再発行・削除は発行者とワークスペースの owner / admin だけ。投稿時にも発行者がチャンネルを閲覧できるかを確認し、非公開チャンネルから抜けた・停止されたメンバーの Webhook は使えなくする。
- **トークン**: 32 バイトの乱数を base64url にして URL に埋め込み、DB には SHA-256 のハッシュだけを保存する（十分な長さの乱数なので低速なハッシュは不要）。平文は発行・再発行の応答でだけ返し、フロントはそのダイアログでだけ表示する。URL はフロントが API のオリジン（`VITE_API_BASE_URL`）から組み立てる。
- **レート制限**: Webhook ごとのトークンバケット（毎秒 1 回、瞬間的に 10 回まで）をプロセス内に持つ。複数台で動かすと台数分まで通るため、厳密にするなら Redis などに移す。
- **監査ログ**: 作成と削除を `webhook_created` / `webhook_deleted` として記録する。URL の再発行と名前の変更は記録しない。
- **既知の制約**: トークンが URL のパスに入るため、アクセスログ（Echo の RequestLogger やロードバランサー）に残る。本番ではログのマスクかヘッダーでのトークン受け渡しを検討する。
