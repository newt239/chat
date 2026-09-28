# Coding Agent Guidelines

## 基本原則

- 常に日本語でコミュニケーションを行ってください。すべてのコミットメッセージ、コメント、エラーメッセージ、ユーザーとのやり取りは日本語で行ってください。
- ファイルの削除を行う場合は、必ず実行前に以下を報告し、明示的なユーザー承認を得てください。
  - 対象ファイルのリスト
  - 実行する変更の詳細説明
  - 影響範囲の説明
- 不明な点がある場合は常に質問し、推測で進めてはなりません。
- 常に DRY なコードを心がけてください。無駄に引数を多くせず、必要な引数のみを渡してください。

## デバッグ

- フロントエンド、バックエンドいずれも Docker 上で実行しています。デバッグを行う場合は、Docker コンテナ内で行ってください。
- 起動は `docker-compose up -d --build` で行ってください。ただし、すでにユーザーが起動している場合もあるので、すでに起動されている場合はスキップしてください。
- 依存関係に変更が生まれた場合は `docker-compose down -v` でコンテナを停止・削除し、`docker-compose up -d --build` で再起動してください。
- ローカルで動作させるための実装は不要です。
- テスト用アカウントとしてユーザー名`alice@example.com`、パスワード`password123`を使用できます。
- コミット時に lefthook の pre-commit フックが lint・format・ファイル名チェックを実行します。ホスト側に`pnpm install`済みであることが前提のため、Docker のみで開発している場合は`LEFTHOOK=0 git commit`で回避できます。

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
- 型定義に`interface`を使用せず、必ず`type`を使用してください（ライブラリの型拡張で`interface`が必須な`src/tanstack-router.d.ts`のような`.d.ts`は例外です）。
- 関数は関数宣言ではなくアロー関数式で定義してください（lint ルール `func-style` で強制されます）。
- 安易に`window`オブジェクトを使用しないでください。
  - ページ遷移には TanStack Router の`Link`や`useNavigate`を使い、`to`にはルート ID（`"/app/$workspaceId/$channelId"`など）、パラメータは`params` / `search`で渡してください。URL 文字列を手で組み立てないでください。
  - Mantine のコンポーネントをリンクにする場合は`component={Link}`ではなく`renderRoot={(props) => <Link {...props} to="..." />}`を使ってください（`to`の型検査を効かせるため）。
  - React のツリー外（fetch インターセプタや WebSocket クライアント）から遷移する場合は`src/lib/navigation.ts`の`navigateTo`を使ってください。
  - ルートパラメータは`useParams({ from: "/app/$workspaceId" })`のように`from`を指定して取得してください。ルートの外からも使うコンポーネントでは`useParams({ strict: false })`を使います。
- ルーティングは TanStack Router のファイルベースルーティングです。
  - `src/routes/`はルート定義専用です。`createFileRoute`で`Route`をエクスポートするだけにし、コンポーネントは`src/pages/`や`src/features/`に定義してください（ルートファイルは 1 ファイル 1 コンポーネント規約の対象外です）。
  - search params は各ルートの`validateSearch`に zod スキーマを渡して検証し、`getRouteApi(...).useSearch()`などで読んでください。
  - `src/routeTree.gen.ts`は Vite プラグインが自動生成します。手で編集しないでください。
- 使用しない引数は削除してください。また、極力引数は Optional にしないようにしてください。
- インポート文は原則として絶対パスで書いてください。パスエイリアスは`#/`です（`#/lib/router`のように書きます）。ただし、同階層や一つ上の階層に限って相対パスでの記述を許可します。
- 1 つのファイルにつき 1 つのコンポーネントを定義してください。コンポーネント名とファイル名は一致させ、Named Export でコンポーネントをエクスポートしてください。
  - ファイル名はコンポーネントを PascalCase、それ以外（hooks・ユーティリティ）を camelCase にしてください。
- 関数の返り値の型は明示しないでください。
- バックエンドのレスポンススキーマを変更した場合はリポジトリルートで`pnpm run openapi:bundle && pnpm run generate:api`を実行して、バンドル済みスキーマと API クライアントの型を更新してください。
  - `openapi-typescript`は TypeScript 5 系にしか対応していないため、フロントエンド（TypeScript 7）ではなくルートワークスペースに配置しています。

## バックエンド

- エントリーポイントは`backend/cmd/server/main.go`です。
- 常にクリーンアーキテクチャを意識してください。
- 冗長なコードは避けてください。
- 過度に共通化しないでください。同様の処理が 2, 3 個しかないのに共通化してしまうと保守性が低下します。
- データベースのテーブル名は単数形で命名してください。
