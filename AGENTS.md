# Coding Agent Guidelines

## 基本原則

- 常に日本語でコミュニケーションを行ってください。すべてのコミットメッセージ、コメント、エラーメッセージ、ユーザーとのやり取りは日本語で行ってください。
- 不要になったファイルは承認なしで削除して構いません。削除したファイルは PR 本文に明記してください。
- 不明な点は妥当な推測で進め、推測で決めたことは PR 本文と最終報告に論点としてまとめてください。
- 常に DRY なコードを心がけてください。無駄に引数を多くせず、必要な引数のみを渡してください。

## デバッグ

- フロントエンド、バックエンドいずれも Docker 上で実行しています。デバッグは原則 Docker コンテナ内で行いますが、worktree で並列に作業する場合は、Docker の Postgres に専用の DB を作り、ホストで backend / frontend を起動して構いません。
- 起動は `docker-compose up -d --build` で行ってください。ただし、すでにユーザーが起動している場合もあるので、すでに起動されている場合はスキップしてください。
- 依存関係に変更が生まれた場合は `docker-compose down -v` でコンテナを停止・削除し、`docker-compose up -d --build` で再起動してください。
- ローカルで動作させるための実装は不要です。
- テスト用アカウントとしてユーザー名`alice@example.com`、パスワード`password123`を使用できます。
- コミット時に lefthook の pre-commit フックが lint・format を実行します。ホスト側に`pnpm install`済みであることが前提のため、Docker のみで開発している場合は`LEFTHOOK=0 git commit`で回避できます。

## リファクタリング

- コードのリファクタリングを指示された場合は、テストを除く総コード行数がなるべく少なくなるよう意識してください。

## フロントエンド

ツールチェーンは **Vite+ (`vite-plus`)** に統合されています。lint (Oxlint) / format (Oxfmt) / test (Vitest) の設定はすべて `frontend/vite.config.ts` に集約されており、`.eslintrc` や `.prettierrc` は存在しません。

- フロントエンドの実装をした際は、必ず最後に`pnpm --filter chat-frontend run codecheck`を実行してください（typecheck / lint / format / knip / test を一括で実行します）。
  - 個別に実行する場合は `pnpm run typecheck`、`pnpm run lint:fix`、`pnpm run format:fix` を使ってください。
  - 修正にあたり、any/unknown などの型を使用することや、型アサーション・型ガードを使用することを禁止します。その実装にふさわしい型を書くか、ライブラリから提供されているものをインポートして使用してください。どうしても型アサーションを使用する必要がある場合は最後にまとめて確認を取ってください。
- 新しいコンポーネントを実装した際は必ず Vitest でテストを書いてください。
  - ユニットテストは対象のファイルと同階層に`filename.spec.{ts,tsx}`という名前で実装してください。
  - テストユーティリティは`vitest`ではなく`vite-plus/test`からインポートしてください（lint ルールで強制されます）。
- UI は Tailwind CSS v4 + React Aria Components で作ります。
  - 基本部品は`src/components/ui/`にあります。まずここの部品を使い、足りなければ React Aria Components で部品を追加してください。
  - 色・角丸・文字・影はトークンのユーティリティ（`bg-surface`、`text-muted`、`border-border`、`rounded-md`、`text-caption`、`shadow-lg`など）だけを使い、`bg-white`や`gray-*`、`#fff`などの色を直書きしないでください。
  - 状態によるスタイルは React Aria の data 属性（`data-hovered:`、`data-pressed:`、`data-selected:`、`data-focus-visible:`など）で書いてください。
  - 画面の文言は直接書かず、`useTranslation()`の`t("機能.文脈.項目")`で`packages/i18n`の辞書から取ってください。日本語と英語の両方の辞書に追加します。日時は`@chat/i18n`のフォーマッタを使ってください。
  - アニメーションは`motion/react`を使い、値は`#/lib/motion`の`transitions`から選んでください。
  - 通知は`#/components/ui/ToastRegion/toast`の`toast()`を使ってください。
- DOM に依存しない共有ロジック（トークン・テーマ生成・i18n 辞書など）は`packages/`に置きます。変更したら`pnpm --filter "./packages/*" run codecheck`を実行してください。
- 型定義に`interface`を使用せず、必ず`type`を使用してください（ライブラリの型拡張で`interface`が必須な`src/tanstack-router.d.ts`のような`.d.ts`は例外です）。
- 関数は関数宣言ではなくアロー関数式で定義してください（lint ルール `func-style` で強制されます）。
- 安易に`window`オブジェクトを使用しないでください。
  - ページ遷移には TanStack Router の`Link`や`useNavigate`を使い、`to`にはルート ID（`"/app/$workspaceId/$channelId"`など）、パラメータは`params` / `search`で渡してください。URL 文字列を手で組み立てないでください。
  - リンクは`#/components/ui/Link/Link`（見た目がボタンなら`#/components/ui/LinkButton/LinkButton`）を使ってください。TanStack Router の`createLink`で React Aria の`Link`を包んだもので、`to` / `params`が型検査されます。`@tanstack/react-router`の`Link`に見た目のクラスを直接書かないでください。
  - React のツリー外（fetch インターセプタや WebSocket クライアント）から遷移する場合は`src/lib/navigation.ts`の`navigateTo`を使ってください。
  - ルートパラメータは`useParams({ from: "/app/$workspaceId" })`のように`from`を指定して取得してください。ルートの外からも使うコンポーネントでは`useParams({ strict: false })`を使います。
- ルーティングは TanStack Router のファイルベースルーティングです。
  - `src/routes/`はルート定義専用です。`createFileRoute`で`Route`をエクスポートするだけにし、画面コンポーネントは`src/features/<機能>/components/`に定義してください（ルートファイルは 1 ファイル 1 コンポーネント規約の対象外です）。
  - search params は各ルートの`validateSearch`に zod スキーマを渡して検証し、`getRouteApi(...).useSearch()`などで読んでください。
  - `src/routeTree.gen.ts`は Vite プラグインが自動生成します。手で編集しないでください。
- 使用しない引数は削除してください。また、極力引数は Optional にしないようにしてください。
- インポート文は原則として絶対パスで書いてください。パスエイリアスは`#/`です（`#/lib/router`のように書きます）。ただし、同階層や一つ上の階層に限って相対パスでの記述を許可します。
- ディレクトリの置き場所は次のとおりです。
  - 機能に閉じるコンポーネント・hooks・ユーティリティは`src/features/<機能>/`の`components/`・`hooks/`・`utils/`に置きます。1 つの機能でしか使わないものを`src/lib/`や`src/hooks/`に置かないでください。
  - ドメインを持たない汎用コンポーネントは`src/components/<グループ>/<コンポーネント名>/<コンポーネント名>.tsx`に置きます。グループは基本部品の`ui/`と、それを組み合わせた`block/`です。spec や付属のヘルパーも同じディレクトリに置きます。
  - 複数の機能で使う hooks は`src/hooks/`、それ以外の横断的な処理は`src/lib/`に置きます。
- バレルファイル（`index.ts`や、他のモジュールを再エクスポートするだけのファイル）は作らないでください。`export ... from`での再エクスポートも禁止です。import は定義元のファイルから直接行ってください（`export *`は lint ルール`oxc/no-barrel-file`で検出されます）。
- 1 つのファイルにつき 1 つのコンポーネントを定義してください。コンポーネント名とファイル名は一致させ、Named Export でコンポーネントをエクスポートしてください。
  - ファイル名はコンポーネントを PascalCase、それ以外（hooks・ユーティリティ）を camelCase にしてください。
- 関数の返り値の型は明示しないでください。
- API は Connect RPC です。`proto/chat/v1/`の定義を変更したらリポジトリルートで`pnpm run proto:format && pnpm run proto:lint && pnpm run generate:proto`を実行し、生成物（`backend/internal/gen/`・`frontend/src/gen/`）もコミットしてください。生成物は手で編集しないでください。
  - API の呼び出しは`@connectrpc/connect-query`の`useQuery(Service.method.xxx, input)` / `useMutation(Service.method.xxx)`を使ってください。キャッシュの無効化には`createConnectQueryKey`で作ったキーを使います（`useQuery`のキーには transport も含まれるため、`setQueryData`ではなく部分一致の`setQueriesData`を使ってください）。
  - 日時は`google.protobuf.Timestamp`で届くため、`#/lib/timestamp`の`toDate`で`Date`に変換してください。

## バックエンド

- エントリーポイントは`backend/cmd/server/main.go`です。
- 常にクリーンアーキテクチャを意識してください。
- 冗長なコードは避けてください。
- 過度に共通化しないでください。同様の処理が 2, 3 個しかないのに共通化してしまうと保守性が低下します。
- データベースのテーブル名は単数形で命名してください。
- API は`internal/interfaces/handler/rpc/`のサービスに実装し、ユースケースの出力から proto のメッセージへの変換は`internal/interfaces/presenter/`に置いてください。
  - ユースケースのエラーは`rpc/error.go`の対応表で Connect のエラーコードに変換されるため、サービスではそのまま返してください。新しいエラーを追加したら対応表にも追加してください。
  - 入力の制約は proto に protovalidate のルールとして書いてください。
