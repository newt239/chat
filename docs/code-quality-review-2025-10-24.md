# コード品質・ディレクトリ構成レビュー (2025-10-24)

## フロントエンド

### コード品質上の懸念
- `frontend/src/features/search/components/SearchPage.tsx:27-33` ワークスペース内のチャンネルID配列から先頭要素だけを選んで `useMessages` を呼び出しており、他のチャンネルのメッセージが検索対象にならないため検索結果が著しく欠落する。
- `frontend/src/features/message/hooks/useMessage.ts:15-31` APIレスポンスを `data as MessagesResponse` で強制キャストしているが、構造検証を行っていないためスキーマ変更時に型安全性が失われ、実行時エラーや描画崩れにつながる。
- `frontend/src/lib/api/client.ts:64-70` 401 リトライ時に元の `Request` オブジェクトの `body` を再利用しているが、`ReadableStream` を含むリクエストでは既に消費済みのストリームを再送しようとして失敗する恐れがあり、副作用系エンドポイントがリトライ不能になる。

### ディレクトリ／責務構成上の懸念
- `frontend/src/components/layout/Header.tsx:15-19` と `frontend/src/features/workspace/components/WorkspaceList.tsx:16-20` でワークスペース型が重複定義されており、共通の `features/workspace/types.ts` のような置き場が存在しないため型定義が散逸している。
- `frontend/src/features/search/components/SearchPage.tsx` が検索用データ取得を `features/message/hooks` や `features/workspace/hooks` に直接依存しており、`features/search` 配下にクエリロジックが集約されていない。結果として検索機能の責務境界が曖昧になり、将来的なAPI実装・切り替えが難しい。

## バックエンド

### コード品質上の懸念
- `backend/internal/interface/http/router.go:23-35` で `/api/auth/logout` が認証ミドルウェアの外に配置されているため、`backend/internal/interface/http/handler/auth_handler.go:162-197` で `c.Get("userID")` が常に失敗し、ログアウト処理が機能しない。
- `backend/cmd/server/main.go:26-90` WebSocketアップグレード時に `CheckOrigin` が常に `true` を返し、さらにワークスペース所属確認を行っていないため、外部オリジンや未所属ユーザーが任意ワークスペースへ接続できるセキュリティリスクがある。
- `backend/internal/interface/ws/connection.go:118-126` 送信キューが満杯の場合にメッセージを黙殺しており、障害調査が困難。ログ出力や接続切断など明示的なハンドリングが必要。

### ディレクトリ／責務構成上の懸念
- `backend/internal/interface/http/handler/dto.go` に全ドメインのリクエストDTOが一括で置かれており、機能単位でのモジュール分割ができていない。各ハンドラ配下にDTOを切り出すか、ドメイン別パッケージを設けると保守性が向上する。
- WebSocketのハンドシェイク処理が `backend/cmd/server/main.go` に直書きされ、`internal/interface/ws` と分離されていないため、コマンド層が輸送レイヤの実装詳細を抱え込んでいる。WebSocketエンドポイントを `internal/interface/ws` 側に移し、`cmd/server` からは初期化のみ行う構成が望ましい。
