# 認証と招待

Google アカウント（OpenID Connect）でのログインを主とし、メールアドレスとパスワードによる認証は管理用・e2e・開発用の補助とする。新しいアカウントは管理者からの招待でだけ作れる。誰でも登録できる `Register` と `/register` 画面は廃止した。

## Google ログイン

- フロントは Google Identity Services（`https://accounts.google.com/gsi/client`）のボタンで ID トークンを受け取り、`AuthService.LoginWithGoogle` に渡す。`VITE_GOOGLE_OAUTH_CLIENT_ID` が未設定ならスクリプトを読み込まず、ボタンも出さない。
- バックエンドは `google.golang.org/api/idtoken` で署名・有効期限・audience（`GOOGLE_OAUTH_CLIENT_ID`）を検証し、発行者が Google であることと `email_verified` を確かめる。`GOOGLE_OAUTH_CLIENT_ID` が未設定なら Google ログインは失敗する。
- ユーザーの照合は次の順で行う。
  1. `user.google_sub` が一致するユーザー。Google 側でメールアドレスが変わってもログインできる
  2. メールアドレスが一致し、まだ Google アカウントが紐付いていないユーザー。初回のログインで `google_sub` を保存する（別の sub が紐付いたユーザーやボットには紐付けない）
  3. どちらもなければ、そのメールアドレスへの未受諾・期限内の招待があるときだけユーザーを作り、招待されたワークスペースに参加させる。招待がなければ `permission_denied` で拒否する
- Google だけで作ったユーザーはパスワードを持たない（`password_hash` に照合できない値 `!` を入れる。Webhook のボットと同じ）。

## 招待

- `invitation` テーブルにワークスペース・メールアドレス・ロール・トークンのハッシュ・招待した人・有効期限（7 日）・受諾日時を保存する。メールアドレスは小文字にそろえて照合する。
- ワークスペースの `invite_members` 権限を持つメンバーが、管理画面の「招待」タブ（とワークスペース設定のメンバー欄）から作成・一覧・取り消しできる。管理者として招待できるのは管理者だけで、オーナーとしては招待できない。
- 既存ユーザーのメールアドレスを招待すると、招待は作らずに直ちにワークスペースへ追加する（旧 `AddMemberByEmail` を置き換えた）。
- 未登録のメールアドレスには招待リンク `/invite/{token}` を発行する。トークンは 32 バイトの乱数で、DB には SHA-256 のハッシュだけを保存するため、リンクは作成直後の応答でしか表示できない。
- 招待メールの送信は `invitation.Sender` として抽象化だけ行い、当面は何も送らない実装を使う。リンクは管理画面でコピーして共有する。
- 招待リンクを開くと招待先を表示し、Google でログインするか、パスワード認証が有効ならパスワードを設定して（`SignUpWithInvitation`）参加できる。どちらの場合も、同じメールアドレスへの他のワークスペースからの招待もまとめて受諾する。Google でログインする場合は、招待されたメールアドレスの Google アカウントである必要がある。
- 取り消した招待は削除する。

## パスワード認証

- `PASSWORD_AUTH_ENABLED` で有効にする。既定値は `ENV=production` のとき無効、それ以外は有効。無効のときは `Login` と `SignUpWithInvitation` が `failed_precondition` を返す。
- フロントは `AuthService.GetAuthConfig` で有効かを問い合わせ、有効なときだけログイン画面と招待画面に Google ボタンの下の補助としてフォームを出す。
- Google もパスワードも使えない設定ではサーバーを起動しない。
- 開発用シードの `alice@example.com` / `password123` などは、開発環境ではそのままパスワードでログインできる。

## トークンとセッション

- アクセストークンは JWT で、有効期間は `JWT_ACCESS_TOKEN_TTL`（分、既定 15）。
- リフレッシュトークンは JWT ではなく 32 バイトの乱数で、有効期間は `JWT_REFRESH_TOKEN_TTL`（日、既定 30）。`session.refresh_token_hash` に SHA-256 のハッシュを保存し、リフレッシュ時はハッシュでセッションを引いてトークンを差し替える。以前はリフレッシュトークンを JWT として検証していたため、リフレッシュが常に失敗していた。
- ハッシュの方式を bcrypt から SHA-256 に変えたため、変更前に発行されたリフレッシュトークンは使えない（再ログインが必要）。
