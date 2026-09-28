# メッセージ表示の拡張（メディア・リンク・ピン・リアクション）

issue #14 のバックエンド側の設計判断をまとめる。

## メッセージの組み立て

- メッセージを返すユースケース（一覧・スレッド・作成・編集・検索・ピン・ブックマーク）は、すべて `usecase/message` の `MessageOutputBuilder.Build(ctx, viewerID, messages)` で出力を作る。メンション・リンク・リアクション・添付・ピンを一括で取得するため、どの API でも同じ内容が返る。
- 出力は閲覧者（viewerID）ごとに決まる。いまはメッセージリンクの引用カードだけが閲覧者によって変わる。

## 添付ファイルのメタデータ

- 幅・高さ・再生時間は、アップロード前にクライアントが計測して `PresignUpload` で送る。サーバーは値の範囲だけを検証して保存し、`MessageAttachment.media` で返す。
  - アップロードは署名付き URL でストレージへ直接送るため、サーバーはファイルの中身を持たない。寸法を読むには保存後にストレージから読み直す必要があり、投稿のたびに余計な往復が増える。
  - スマートフォンの写真は EXIF の回転情報でブラウザの表示寸法とファイル上の寸法が入れ替わる。ブラウザが計測した値のほうがレイアウトの予約に合う。
  - 値はレイアウトシフトを防ぐためのヒントであり、誤っていても表示が崩れるだけで安全性には影響しない。
  - ストレージの抽象（`StorageService`）には手を入れていないため、GCS へ移っても影響しない。
- MIME タイプはパラメータを除いて小文字に正規化する。`application/octet-stream` や空のときは拡張子から推定する（推定できる種類は実行環境の MIME テーブルに依存する）。

## メッセージリンク

- 同じワークスペースのメッセージを指す URL を「メッセージリンク」として扱う。形式は次のとおりで、ホストは問わない（フロントエンドの配信元が複数あっても動くようにするため）。
  - チャンネルのメッセージ: `/app/{workspaceId}/{channelId}?message={messageId}`
  - スレッド: `/app/{workspaceId}/{channelId}/thread/{threadId}`（親メッセージを指す）。`?message={replyId}` を付けると返信を指す。
- 投稿・編集時に `LinkProcessingService` が URL を判定する。同じワークスペースに実在し、URL のチャンネルと一致するメッセージなら `message_link.linked_message_id` に記録し、OGP は取得しない（自前の SPA を取りに行っても意味がないため）。
- 引用カード（投稿者・チャンネル名・本文の先頭 200 文字・日時）は読み出しのたびに組み立てる。削除済みのメッセージと、閲覧者がチャンネルを参照できないメッセージは展開しない。書き込み時に内容を保存しないのは、閲覧者ごとに権限が違い、引用元の編集・削除も反映したいため。
- WebSocket の `newMessage` / `messageUpdated` は投稿者の権限で組み立てた出力を購読者全員へ送るため、引用カード（`message_preview`）を取り除いて配信する。フロントエンドは `linked_message_id` があって `message_preview` がないリンクについて `MessageService.GetMessagePreview` を呼ぶ。参照できない場合は `NotFound` が返る。

## OGP と YouTube

- 保存済みの OGP は URL ごとに再利用する。タイトルが取れていないものは再取得する。
- `og:image:width` / `og:image:height` があれば画像の寸法も返す。
- YouTube の動画 URL（`watch?v=`・`youtu.be`・`shorts`・`embed`・`live`）は動画 ID を取り出し、動画ページの OGP と microdata（`itemprop="duration"`、投稿者の `itemprop="name"`）からタイトル・サムネイル・チャンネル名・再生時間を読む。外部 API キーは使わない。
  - 同意画面などでタイトルが取れないときは oEmbed（`/oembed`）で補う。oEmbed には再生時間がないため、その場合は再生時間なしになる。
  - サムネイルが取れないときは `i.ytimg.com/vi/{id}/hqdefault.jpg` を使う。
  - 動画ページはメタデータが 700KB 付近にあるため、HTML の読み込み上限を 2MB にしている。

## ピン留めとリアクション

- メッセージには、ピン留めされている場合に `pin`（ピン留めした人と日時）を含める。一覧を取得した時点でラベルを表示できる。
- WebSocket の `pinCreated` には、ラベル表示用に `pinned_by_user` を含める。
- リアクションは押した順（`created_at` 昇順）で返す。メッセージの `reactions` と `ReactionService.ListReactions` の両方にユーザーと日時が含まれるため、ツールチップと一覧モーダルはこれで表示できる。
- WebSocket の `reactionAdded` には `user` と `created_at` を含める。削除のイベントは `user_id` だけを持つ。

## シードデータ

YouTube のリンク、メッセージリンク（参照できるもの・private チャンネルのもの）、ピン留め、多数のリアクションのサンプルを作る。添付ファイルはストレージに実体が必要なため作らない。
