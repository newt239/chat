# OpenAPI 整合性チェック導入・移行計画（実装手順）

本書は、既存実装から OpenAPI をソース・オブ・トゥルースとして運用し、バックエンド実装との整合性を高い確度で維持するための導入・移行手順です。以下の 1〜4 を順に導入します。

- 1) 実行時バリデーション（Echo ミドルウェア）
- 2) コンパイル時の齟齬低減（oapi-codegen による型/スタブ生成）
- 3) コントラクトテスト（Schemathesis / Dredd）
- 4) CI チェック（Spectral Lint + コントラクトテスト + フロント型の再生成差分チェック）

前提:
- バックエンドエントリポイント: `backend/cmd/server/main.go`
- OpenAPI スキーマ: `backend/internal/openapi/openapi.yaml`
- リバースプロキシ無しで `http://localhost:8080`（Docker コンテナ構成に合わせて後述の URL を調整してください）
- 本プロジェクトは Docker での動作が前提です

注意:
- 既存の `swaggo/swag` によるコードコメント→Swagger 生成は廃止し、手書きの OpenAPI を唯一のソースとします（移行手順に記載）。
- フロントは `openapi-typescript` を継続利用（`pnpm run generate:api`）。


## 0. 現状整理と移行方針

現状:
- `backend/internal/openapi/openapi.yaml` を中心に `paths.yaml` / `components/schemas.yaml` に分割定義
- `frontend` は `openapi-typescript` で型生成（`pnpm run generate:api`）
- `go.mod` に `github.com/swaggo/swag` が残存、`backend/docs` の Swagger 生成物は削除済み

移行方針:
- Swagger コメント起点を撤去（`swag` 依存を削除）。
- 実行時バリデーション（kin-openapi + oapi-codegen/echo-middleware）を導入。
- oapi-codegen で Go の型/サーバースタブを生成し、コンパイル時に齟齬検知。
- Schemathesis を使い、実サーバーに対するコントラクトテストを実施。
- CI に Spectral Lint と Schemathesis を組込み、フロント型の再生成差分もチェック。


## 1) 実行時バリデーション（Echo ミドルウェア）導入

目的:
- リクエスト/レスポンスが OpenAPI に準拠しているか、実行時に自動検証する。

主要ライブラリ:
- `github.com/getkin/kin-openapi/openapi3`
- `github.com/oapi-codegen/echo-middleware`

手順:
1. バックエンドに依存を追加
   - `go get github.com/getkin/kin-openapi/openapi3`
   - `go get github.com/oapi-codegen/echo-middleware`

2. サーバー起動時に `openapi.yaml` を読み込んで検証
   - `backend/cmd/server/main.go` にて、以下の概略を組み込みます（擬似コード）。

   ```go
   loader := &openapi3.Loader{IsExternalRefsAllowed: true}
   doc, err := loader.LoadFromFile("backend/internal/openapi/openapi.yaml")
   if err != nil { /* 致命 */ }
   if err := doc.Validate(loader.Context); err != nil { /* 致命 */ }
   e.Use(oapimw.OapiRequestValidator(doc))
   // レスポンス検証をしたい場合は openapi3filter の Options で有効化を検討
   ```

3. ルーティングのパス/メソッド/ステータスコード/レスポンスを OpenAPI と一致させる
   - ここでズレがあるとミドルウェアが 400/500 を返すため、齟齬の早期検知が可能。

4. Docker での動作
   - API サーバーコンテナ内部のワークディレクトリから見た `openapi.yaml` の相対/絶対パスに注意。ビルドコンテキストへ同梱するか、ボリュームでマウントしてください。


## 2) oapi-codegen による型/スタブ生成

目的:
- コンパイル時に仕様ズレを顕在化させ、実装の手戻りを減らす。

インストール（開発者ローカル）:

```bash
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
```

生成コマンド例:

```bash
oapi-codegen \
  -generate types,server \
  -package openapi \
  -o backend/internal/openapi_gen/openapi.gen.go \
  backend/internal/openapi/openapi.yaml
```

運用指針:
- 生成された `ServerInterface` を実装し、Echo へバインドします。
- `openapi.yaml` の更新時は再生成を必須化（Makefile/スクリプトを用意すると良い）。
- 生成物は基本コミット推奨（CI の再現性・差分検出のため）。


## 3) コントラクトテスト（実サーバー検証）

目的:
- 実サーバーが OpenAPI に準拠しているか E2E で自動検証する。

選択肢と推奨:
- 推奨: Schemathesis（Python 製、Fuzz 的な入力生成が強力）
- 代替: Dredd（Node 製、記述の忠実性が高い）

Schemathesis の例:

```bash
# コンテナ/ホストの URL は環境に合わせて変更
schemathesis run --checks all \
  --base-url http://localhost:8080 \
  backend/internal/openapi/openapi.yaml
```

よくある調整点:
- 認証が必要なエンドポイントは、`--auth` もしくはヘッダフック（`--header`）やカスタムローダで対処。
- データ依存のエンドポイントはデータシードを先に投入（Docker Compose の初期化スクリプトで対応）。

Dredd の例:

```bash
dredd backend/internal/openapi/openapi.yaml http://localhost:8080
```


## 4) CI 組み込み（Spectral Lint / Schemathesis / フロント型差分）

目的:
- スキーマの品質・API 実装の準拠・クライアント型の一貫性を継続的に担保する。

推奨チェック:
1. Spectral による OpenAPI Lint
   ```bash
   npx @stoplight/spectral lint backend/internal/openapi/openapi.yaml
   ```

2. Schemathesis によるコントラクトテスト
   ```bash
   schemathesis run --checks all \
     --base-url http://api:8080 \# コンテナ名/ネットワークに合わせる
     backend/internal/openapi/openapi.yaml
   ```

3. フロント型の再生成差分チェック
   ```bash
   pnpm --filter frontend run generate:api
   git diff --exit-code frontend/src/lib/api/schema.ts
   ```

GitHub Actions 例（概略・新規ワークフロー案）:

```yaml
name: openapi-checks
on:
  pull_request:
  push:
    branches: [ main ]

jobs:
  spectral:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: '20' }
      - run: npm i -g @stoplight/spectral
      - run: spectral lint backend/internal/openapi/openapi.yaml

  contract-test:
    runs-on: ubuntu-latest
    services:
      api:
        image: ghcr.io/your-org/your-api:latest # もしくは compose でビルド
        ports:
          - 8080:8080
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with: { python-version: '3.11' }
      - run: pip install schemathesis
      - run: schemathesis run --checks all --base-url http://localhost:8080 backend/internal/openapi/openapi.yaml

  frontend-types:
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: frontend
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
        with: { version: 9 }
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'pnpm'
          cache-dependency-path: frontend/pnpm-lock.yaml
      - run: pnpm install --frozen-lockfile
      - run: pnpm run generate:api
      - run: git diff --exit-code src/lib/api/schema.ts
```


## 5) 既存からの移行手順まとめ

1. Swagger コメント運用の撤退
   - `go.mod` から `github.com/swaggo/swag` を削除
   - CI/Makefile/スクリプトの `swag init` などの呼び出しを削除
   - `backend/docs` の生成物は既に削除済み（再生成禁止）

2. 実行時バリデーションの導入
   - `kin-openapi` と `oapi-codegen/echo-middleware` を追加し、`main.go` に読み込み+ミドルウェアを組み込み
   - 最低限、リクエスト検証を有効化。必要に応じてレスポンス検証も段階的に有効化

3. oapi-codegen による型/スタブ生成
   - `backend/internal/openapi_gen/openapi.gen.go` を生成
   - 生成 `ServerInterface` に合わせてハンドラを実装/バインド
   - 生成物をコミットし、スキーマ変更時に再生成する運用を徹底

4. コントラクトテストの整備
   - Schemathesis をローカル/Docker 上で実行できるようにする
   - 認証・データ依存のエンドポイントにはテスト前のシーディング/トークン取得を用意

5. CI の整備
   - Spectral Lint を追加
   - API サーバー立ち上げ後に Schemathesis を実行
   - フロントの `generate:api` → 差分チェックを追加

6. フロントエンドの運用（既存継続）
   - OpenAPI スキーマ変更時は `pnpm run generate:api`
   - フロントの型破壊があればビルド/型チェックで検出


## 6) トラブルシューティング

- 相対パス問題で `openapi.yaml` が見つからない
  - コンテナ内の作業ディレクトリからの相対位置を見直すか、絶対パス/環境変数でパスを固定化。

- レスポンス検証で大量に 500 になる
  - まずはリクエスト検証のみを導入し、レスポンス検証はエンドポイント単位で段階的に有効化。

- Schemathesis が認証で失敗
  - 事前にテストユーザーでトークンを取得し、`Authorization: Bearer <token>` を付与するフック/前処理を組み込む。


## 7) 今後の発展

- `oapi-codegen` の `-generate spec,skip-prune-tags` 等を活用し、より厳密な型導出/運用方針の自動化。
- Redoc などでのドキュメントサイト自動公開（CI で `redoc-cli` を使い静的サイトを生成）。
- 変更検知用の Bot（PR でスキーマ→コード差分のサマリ生成）。


---

以上。1→2→3→4 の順で導入すると、短期にズレ検知、長期に運用の安定化が見込めます。


