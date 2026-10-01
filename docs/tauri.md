# デスクトップ・モバイルアプリ（Tauri）

`frontend/` をそのまま Tauri v2 の WebView に載せて、macOS・Windows・iOS・Android のアプリにする。Rust 側は `frontend/src-tauri/`、Android と iOS のプロジェクトは `frontend/src-tauri/gen/` にある。

## Web 版との違い

`vp build --mode tauri` でビルドし、コードでは `#/lib/platform/platform` の `isTauri`（`import.meta.env.MODE === "tauri"`）で分ける。Tauri のプラグインは `src/lib/platform/tauri/` の中だけで使い、動的 import するため Web 版のバンドルには入らない。

| 機能 | Web 版 | アプリ |
| --- | --- | --- |
| Service Worker・PWA のインストール | あり | なし（`VitePWA({ disable })`） |
| 外部リンク・添付ファイルのダウンロード | 新しいタブ | システムのブラウザ（opener） |
| 共有用リンクのコピー | 表示中の origin | `VITE_PUBLIC_APP_URL` |
| 新着の通知・バッジ | Notification / Badging API | ネイティブ（notification、`setBadgeCount`）。デスクトップの通知は押してもアプリが前に出るだけ |
| WebSocket | 隠れると切断 | デスクトップは隠れても切断しない |
| Google ログイン | Google Identity Services | ブラウザで認可コードフロー（PKCE）を行い、`dev.newt239.chat://auth/callback` で戻る |
| 多重起動 | - | デスクトップは 2 つ目を起動せず既存のウィンドウを前に出す |

## 必要なもの

- Rust（stable）
- デスクトップ: macOS は Xcode、Windows は WebView2 と MSVC
- Android: JDK 17、Android SDK、NDK 27（`NDK_HOME` を設定）
- iOS: Xcode と iOS Simulator の SDK（Xcode の Settings > Components から入れる）

## 開発

```sh
cd frontend
cp .env.tauri.local.example .env.tauri.local   # API の URL などを書く
pnpm dev:tauri                                  # デスクトップ
pnpm tauri android dev                          # Android（エミュレータか実機）
pnpm tauri ios dev                              # iOS
```

- エミュレータから Mac の backend へは `10.0.2.2` で届く。実機では Mac の LAN の IP を `.env.tauri.local` に書く。
- Android で gradle が `ERR_PNPM_BAD_PM_VERSION` で失敗したら、古い PATH のまま動いている gradle のデーモンを止める（`src-tauri/gen/android/gradlew --stop -p src-tauri/gen/android`）。

## ビルド

```sh
pnpm --filter chat-frontend run build:tauri                       # デスクトップ
pnpm --filter chat-frontend exec tauri android build --apk        # Android
pnpm --filter chat-frontend exec tauri ios build                  # iOS
```

PR では `.github/workflows/tauri.yml` が 4 つのプラットフォームの署名なしデバッグビルドを確かめる。署名とストアへの配布はまだ用意していない。

npm の `@tauri-apps/*` と Rust の crate は同じ minor に揃える（`tauri build` が不一致を拒否する）。npm は `minimumReleaseAge` で 7 日経った版しか入らないため、Rust 側も `Cargo.toml` の `~2.x` で合わせている。

## Google ログインの設定

アプリは Web 版と同じ Google の「ウェブ アプリケーション」の OAuth クライアントを使い、backend が Google とアプリの間を取り持つ。

1. Google Cloud Console で、そのクライアントの「承認済みのリダイレクト URI」に `https://<API のドメイン>/oauth/google/callback` を追加する。
2. backend に次の環境変数を設定する（未設定ならアプリでは Google ログインを使えず、`/oauth/google/*` は 404 を返す）。

| 変数 | 値 |
| --- | --- |
| `GOOGLE_OAUTH_CLIENT_SECRET` | 同じクライアントのシークレット |
| `GOOGLE_OAUTH_REDIRECT_URL` | 1 で登録した URI |
| `NATIVE_APP_REDIRECT_URL` | 省略時 `dev.newt239.chat://auth/callback` |

本番では `GOOGLE_OAUTH_CLIENT_SECRET` を Secret Manager に登録し、`infra/k8s/base/identity.yaml` の `backend-secrets` と Terraform の `manual_secrets` に足す。

## origin

アプリの origin は macOS・iOS が `tauri://localhost`、Windows・Android が `https://tauri.localhost`。backend の `CORS_ALLOWED_ORIGINS` と添付ファイルのバケットの CORS に両方を入れてある。
