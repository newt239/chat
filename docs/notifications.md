# 通知と PWA

アプリ内の通知（WebSocket）、プッシュ通知（FCM）、PWA としてのインストールと更新の設計です。

## 通知の種類

| 種類 | 届くとき | 仕組み |
| --- | --- | --- |
| アプリ内のデスクトップ通知 | アプリを開いている間 | WebSocket の `newMessage` を見て `new Notification()` を出す（`useDesktopNotifications`） |
| プッシュ通知 | アプリを閉じている・裏にある間 | サーバーが FCM へ送り、Service Worker（`frontend/sw/sw.ts`）が表示する |

- 二重に出さないため、プッシュ通知が届くのは表示中のタブがないときだけにした。Firebase の SDK は表示中のタブがあると Service Worker ではなくタブへ渡すので、タブ側では何もしない（`onMessage` を使わない）。
- デスクトップ通知は、この端末でプッシュ通知を登録していればタブが裏にある間は出さない。両方が出てもよいよう、どちらも `tag` をメッセージ ID にして OS 側で 1 件にまとめる。

## 通知の設定

- 通知する範囲（すべて / メンションと DM / なし）は `UserPreferences.notification_level` に移し、アカウントに保存する。サーバーがプッシュ通知の宛先を決めるのに使うため。以前は端末ごとの localStorage に持っていた。
- デスクトップ通知の有無と、この端末で登録したプッシュ通知の ID は端末ごとに localStorage に持つ（`notificationPreferencesAtom`）。通知の許可がブラウザごとに違うため。
- ミュートしたチャンネル・DM は、メンションでも通知しない。

## プッシュ通知の宛先

`usecase/notification.Dispatcher` がメッセージ作成の後段（`MessageCreator.publish`）で呼ばれ、非同期に送る。予約送信や Webhook の投稿も同じ経路を通る。

| 理由 | 対象 | 「すべて」 | 「メンションと DM」 |
| --- | --- | --- | --- |
| DM | DM・グループ DM の参加者 | 送る | 送る |
| メンション | メンションされたユーザーと、メンションされたグループのメンバー | 送る | 送る |
| スレッド | フォロー中のスレッドへの返信 | 送る | 送らない |

- 送信者本人、ボット、そのチャンネルを閲覧できないユーザー（非公開チャンネルの非メンバー・停止中）、ミュート中のユーザーは除く。
- チャンネルへの通常の投稿は「すべて」でもプッシュ通知しない（アプリ内のデスクトップ通知では「すべて」ならすべて出す）。スマートフォンに全投稿が届くと多すぎるため。
- 本文は 200 文字で切る。通知を押したときの遷移先は、サーバーがメッセージのパス（`entity.MessagePermalinkPath`。スレッドの返信はスレッドの中）を `link` として載せる。フロントでは URL を組み立てない。

## FCM

- 送信は `firebase.google.com/go/v4` の messaging（`infrastructure/fcm`）。認証は ADC（GKE では Workload Identity）。`FIREBASE_PROJECT_ID` が未設定なら送信役を作らず、何も送らない。
- 送信先は Firebase Installation ID（FID）。登録トークンによる送信は JS SDK・Admin SDK とも非推奨になったため、フロントは `register` / `onRegistered` で FID を得る。
- ウェブにはデータだけのメッセージを送り、表示は Service Worker が決める。Android / iOS には表示内容（notification / aps）を付ける。
- FCM が無効（unregistered / sender ID の不一致）と返した ID は削除する。

## 端末の登録 API

`NotificationService.RegisterPushToken` / `UnregisterPushToken`。テーブルは `push_token`（user, token, platform, user_agent, last_seen_at）。

- `platform` は web / ios / android。将来のネイティブアプリも同じ API で FID を登録する。
- 同じ ID を別のユーザーが登録したら付け替える（端末を共有してログインし直した場合）。登録のたびに `last_seen_at` を更新する。
- フロントは設定の「プッシュ通知」をオンにしたときに許可を求めて登録し、ワークスペースを開くたびに登録し直す（ID が変わっていたら古いものを外す）。ログアウトの前に解除する。
- `VITE_FIREBASE_*` がそろっていない、またはブラウザが対応していなければ、プッシュ通知の項目を出さない。

## PWA

- `vite-plugin-pwa` を `injectManifest` にして、Service Worker を自前で書いた（プッシュ通知の受信と通知のクリックを扱うため）。以前の `generateSW` のキャッシュ設定（フォントの CacheFirst、`api.` の NetworkFirst）と SPA のナビゲーションのフォールバックを移した。
- 型検査は Service Worker だけ `webworker` の lib で別に行う（`sw/tsconfig.json`）。
- manifest: アイコンは `public/logo.svg` から通常・maskable・apple-touch-icon を生成する。`theme_color` / `background_color` は既定テーマのトークン（`src/lib/pwaColors.ts`。設定ファイルからは `@chat/design-tokens` を読めないため値で持ち、spec で一致を確かめる）。表示中の `theme-color` は `ThemeProvider` がテーマの surface に書き換える。
- `shortcuts`（DM・通知・検索）はワークスペース ID を含められないため `/app/?open=dms` のようにし、前回のワークスペースの画面へ移る。
- 更新は `registerType: "prompt"`。新しい版を見つけたら操作付きのトーストを出し、「再読み込み」で切り替える。
- インストール: `beforeinstallprompt` を起動時から受け取っておき、「自分」タブに「アプリをインストール」を出す。iOS の Safari はこのイベントがないため、同じ行でホーム画面への追加の手順を案内する。
- アプリアイコンのバッジ（`setAppBadge`）は、モバイルのタブと同じ未読（DM の未読の合計＋未読のメンションがあるチャンネル数）を出す。

## モバイルの表示

- `viewport-fit=cover` を指定しないと `env(safe-area-inset-*)` が 0 になるため指定した。`interactive-widget=resizes-content` で Android はキーボードの分だけ画面が縮む。
- iOS はキーボードでページごとずらすため、`visualViewport` の高さに `MobileShell` の高さを合わせる（`useVisualViewport`）。キーボードが出ている間はボトムタブを隠し、下端のセーフエリアの余白も外す。
