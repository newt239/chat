# バックエンド実装レビュー（2025-10-26）

以下は、現状のバックエンド実装をクリーンアーキテクチャの観点と技術的負債の観点から確認した結果です。優先度が高い順に記載しています。

## 1. UseCase層がインフラ実装（logger）に直接依存している
- 該当箇所: `backend/internal/usecase/message/creator.go:12`, `backend/internal/usecase/message/updater.go:12`, `backend/internal/usecase/message/deleter.go:10`
- 内容: UseCase内で `github.com/newt239/chat/internal/infrastructure/logger` を直接インポートし、`logger.Get().Warn(...)` を呼び出しています。インターフェース層より下の層がインフラ実装へ依存しており、依存関係逆転の原則に反します。
- 影響: UseCaseの単体テストが行いづらくなり、将来的にロガー実装を差し替える際にUseCase層の変更が必須になります。
- 推奨対応: ロギング用のポート（インターフェース）をドメインまたはUseCase層に定義し、インフラ側でアダプタを提供する。短期的にはロギングを呼び出す役割をNotificationService等の外部サービス側に寄せる検討も有効です。

## 2. WebSocket通知用のマップ変換処理が重複・不整合
- 該当箇所: `backend/internal/usecase/message/utils.go:11-52`, `backend/internal/usecase/reaction/interactor.go:212-222`, `backend/internal/infrastructure/notification/websocket_notification_service.go:110-161`
- 内容: 同様の `convertStructToMap` ロジックがUseCaseとインフラの複数箇所で重複しています。実装方法も Reflect ベース、JSON マーシャルベースと統一されておらず、ゼロ値の扱いが異なります。
- 影響: イベントペイロードのフィールド欠落や型揺れを招きやすく、保守コストも高くなります。
- 推奨対応: 変換処理をNotificationService内部に集約するか、専用のDTOとマーシャラを定義して、UseCase側は明示的な構造体を返却する形に改めます。

## 3. GORMモデル変換でUUIDパースエラーを握りつぶしている
- 該当箇所: `backend/internal/infrastructure/models/models.go:33`, `backend/internal/infrastructure/models/models.go:81`, `backend/internal/infrastructure/models/models.go:129` ほか `FromEntity` 系のメソッド全体
- 内容: `utils.ParseUUID` の戻り値エラーを無視しており、パース失敗時はゼロUUIDがそのまま利用されます。
- 影響: IDが空文字列や不正フォーマットでもエラーにならずにDBへ保存され、整合性問題や意図しないレコード更新を引き起こすリスクがあります。
- 推奨対応: `FromEntity` でエラーを返す形に変更する、あるいは呼び出し元でIDを確実に検証する仕組みを導入します。

## 4. ストレージサービス初期化失敗を黙殺している
- 該当箇所: `backend/internal/registry/infrastructure_registry.go:57-64`, `backend/internal/usecase/attachment/interactor.go:65`
- 内容: Wasabiクライアントの生成に失敗すると `NewStorageService` は `nil` を返却しますが、ログ出力もなく、UseCase側では `storageService` を即座に利用しています。
- 影響: 起動時の設定ミスや一時的な接続障害があった場合に、ランタイムで `nil` ポインタアクセスが発生し、アップロード機能が即座にパニックになります。
- 推奨対応: 初期化失敗時はエラーを返してアプリケーション起動を止める、もしくは呼び出し側でエラー処理できるようにインターフェースを見直します。また、最低限ログ出力は必須です。

## 5. WebSocketハンドラーがユースケースを経由せずリポジトリに依存
- 該当箇所: `backend/internal/interfaces/handler/websocket/handler.go:12-75`, `backend/internal/interfaces/handler/websocket/hub.go:116`
- 内容: ハンドラーが `repository.WorkspaceRepository` を直接受け取り、認可ロジックを自身で実装しています。また、ブロードキャスト時にチャンネル単位のフィルタリングが未実装（TODO）です。
- 影響: 認可ロジックがUseCase層と重複し、変更漏れが発生しやすい構造になっています。イベント送信対象の限定も未実装のため、将来的に不要な通知が大量に流れる恐れがあります。
- 推奨対応: Workspace所属確認や通知対象決定を担うUseCase/サービスを用意し、ハンドラーは依存注入を受けて呼び出すだけの薄い層に整理する。TODOのチャンネルフィルタリングも合わせて実装を進めます。

## 付記
- UseCase層の一部で `fmt.Printf` を用いたロギングが残っている（例: `backend/internal/usecase/reaction/interactor.go:91`, `backend/internal/usecase/readstate/interactor.go:83`）。ロギング経路を統一する際に併せて修正を検討してください。
- WebSocketハンドラーの同一ファイル内で `log.Printf` を直接使用している箇所（例: `backend/internal/interfaces/handler/websocket/hub.go:82` など）も、横断的なロギングポリシーに沿って置き換えると保守性が向上します。
