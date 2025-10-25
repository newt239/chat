# チャンネルメンバー管理（招待・削除）実装計画

## 実装日時
2025-10-25

## 目的
- プライベートチャンネルのアクセス制御を強化し、必要なユーザーのみが閲覧・投稿できる状態を保つ。
- チャンネル参加者の増減を UI から安全かつ迅速に行えるようにし、運用コストを下げる。
- 将来的なチャンネル権限拡張（モデレーター機能など）に備えた土台を整備する。

## 現状整理
- リポジトリ層 (`backend/internal/adapter/gateway/persistence/channel_repository.go`) には `AddMember` / `RemoveMember` / `FindMembers` / `IsMember` が実装済みで、`channel_members` テーブルも存在するが、ロール情報は保持していない。
- チャンネル用ユースケース (`backend/internal/usecase/channel`) にはメンバー操作が実装されておらず、HTTP ハンドラーやルーティングも未提供。
- プライベートチャンネルへのアクセス判定はメッセージ・リアクション等のユースケースが `IsMember` を呼び出す形で行っている。
- フロントエンド (`frontend/src/features/channel`, `frontend/src/features/workspace`) にはワークスペース単位のメンバー一覧表示のみ存在し、チャンネルメンバー管理 UI / API 呼び出しは未実装。
- OpenAPI (`backend/internal/openapi/openapi.yaml`) でもチャンネルメンバー用エンドポイントは定義されていない。

## 想定ユースケース
1. ワークスペース管理者またはチャンネル作成者がユーザーを検索し、チャンネルへ招待する。
2. 不要になったメンバーをチャンネルから削除する。
3. パブリックチャンネルの利用者が自発的に参加・離脱する。
4. チャンネル管理者ロールを付与・剥奪し、最低 1 名の管理者を常に維持する。
5. これらの操作結果をリアルタイム通知やメールで関係者へ共有する。

## 要件・前提
- チャンネルへ招待できる対象は同一ワークスペース内の既存ユーザーに限定する。
- 招待・削除・ロール変更を実行できるのはワークスペースの `owner/admin` とチャンネル作成者。
- 重複参加は 409 応答で弾き、削除対象がメンバーでない場合は 404 を返す。
- `channel_members.joined_at` は UTC 現在時刻で設定する。
- 招待・削除・ロール変更を行うオペレーターは、対象チャンネルへアクセス可能である必要がある。
- パブリックチャンネルは自由参加／離脱を許可し、招待 API からメンバー追加することも可能とする。
- チャンネル管理者（`role = admin`）が最低 1 名存在することを保証する。管理者削除やロール剥奪時には検証を行う。
- 招待・削除・ロール変更・自己参加/離脱の各イベントは WebSocket 通知およびメール通知へ連携する。

## データモデル／スキーマ変更
- `channel_members` テーブルに `role text not null default 'member'` 列を追加し、`('member','admin')` のチェック制約を設定する。
- 既存データに対しては `role = 'admin'` をチャンネル作成者、その他を `member` として初期化する移行スクリプトを作成。
- Domain 層 `entity.ChannelMember` に `Role string` プロパティを追加し、GORM モデル／Repository を更新。
- `ChannelRepository` インターフェースに `UpdateMemberRole` と `FindAdmins` 等が必要であれば追加検討。

## API 設計（ドラフト）
| HTTP | パス | 内容 | 認可 |
|------|------|------|-----------|
| GET | `/api/channels/{channelId}/members` | チャンネルメンバー一覧取得 | チャンネル参加者（パブリックはワークスペース参加者） |
| POST | `/api/channels/{channelId}/members` | メンバー招待（ロール指定可） | ワークスペース `owner/admin` またはチャンネル作成者 |
| POST | `/api/channels/{channelId}/members/self` | パブリックチャンネルへの自己参加 | 対象チャンネルがパブリックであること |
| PATCH | `/api/channels/{channelId}/members/{userId}/role` | チャンネル管理者ロール付与/剥奪 | ワークスペース `owner/admin` またはチャンネル作成者 |
| DELETE | `/api/channels/{channelId}/members/{userId}` | チャンネルメンバー削除 | ワークスペース `owner/admin` またはチャンネル作成者 |
| DELETE | `/api/channels/{channelId}/members/self` | 自己離脱（パブリック/プライベート問わず） | 当該ユーザーがメンバーであること |

### リクエスト/レスポンス（案）
- GET `/members`
  - Response: `{ "members": ChannelMemberInfo[] }`
  - `ChannelMemberInfo`: `{ userId, role, joinedAt, displayName, email, avatarUrl }`。
- POST `/members`
  - Request Body: `{ "userId": "uuid", "role": "admin" | "member" }`
  - Response: `{ "success": true }`
- POST `/members/self`
  - Response: `{ "success": true }`
- PATCH `/members/{userId}/role`
  - Request Body: `{ "role": "admin" | "member" }`
  - Response: `{ "success": true }`
- DELETE `/members/{userId}` / DELETE `/members/self`
  - Response: `{ "success": true }`

## バックエンド実装計画
- ユースケース新設案: `internal/usecase/channelmember`（または `channel` ユースケース拡張）を追加し、以下のメソッドを提供。
  - `ListMembers`
  - `InviteMember`
  - `JoinPublicChannel`
  - `UpdateMemberRole`
  - `RemoveMember`
  - `LeaveChannel`
- DTO 定義: `InviteMemberInput{ ChannelID, OperatorID, TargetUserID, Role }`, `UpdateMemberRoleInput{ ChannelID, OperatorID, TargetUserID, Role }`, `RemoveMemberInput{ ChannelID, OperatorID, TargetUserID }`, `JoinChannelInput{ ChannelID, UserID }`, `LeaveChannelInput{ ChannelID, UserID }`, `MemberListOutput`。
- 認可ロジック:
  - チャンネル存在確認 (`channelRepo.FindByID`)。
  - オペレーターがワークスペースメンバーか (`workspaceRepo.FindMember`)。
  - プライベートチャンネルでは `channelRepo.IsMember` でアクセス権を確認。
  - 招待・削除・ロール変更ではオペレーターがワークスペース `owner/admin` もしくはチャンネル作成者かを検証。
- 招待処理:
  - ターゲットユーザーがワークスペースメンバーであるか検証。
  - 既存メンバーなら `ErrAlreadyMember` を返却。
  - `channelRepo.AddMember` を拡張し、`Role` と `JoinedAt` を設定。
- 自己参加処理:
  - 対象チャンネルがパブリックであることを確認。
  - 既存メンバーの場合は冪等に成功応答。
- ロール変更処理:
  - 対象ユーザーがチャンネルメンバーであることを確認。
  - `channelRepo.UpdateMemberRole` を追加し、更新後に管理者が 1 名以上残るか検証。
- 削除処理:
  - 対象ユーザーがメンバーであることを確認。
  - 削除対象が管理者の場合、残りの管理者数を確認し 0 になる場合は 409 応答。
  - `channelRepo.RemoveMember` を呼び出し。
- 自己離脱処理:
  - 削除と同様に最低 1 名の管理者が残るかを検証（離脱者が管理者かつ最後の 1 人なら離脱不可）。
- 一覧処理:
  - `channelRepo.FindMembers` でメンバーリストを取得し、`userRepo.FindByIDs` でプロフィールを補完。DTO へマッピング。
- 通知・メール送信:
  - `NotificationService` にチャンネルメンバーイベントを追加し、WebSocket でクライアントへ配信。
  - メール送信要件に応じて `EmailService`（既存なければスタブ）へイベントを発行。
- ルーティング/ハンドラー:
  - `internal/interface/http/handler/channel_member_handler.go` を新設し、上記エンドポイントを実装。
  - `internal/interface/http/router.go` で `/channels/:channelId/members` 配下のルートを登録。
- OpenAPI 更新:
  - `backend/internal/openapi/openapi.yaml` に新エンドポイントとスキーマを追記。
  - `pnpm run generate:api` をフロントエンドで実行。
- Repository 拡張:
  - `ChannelRepository` に `UpdateMemberRole`, `CountAdmins`, `FindAdmins` 等を追加。
  - GORM 実装に対応するクエリを実装。
- テスト:
  - ユースケース単体テストで認可・ロール制約・各種エラーケースをカバー。
  - HTTP ハンドラテストでステータスコードとレスポンスを確認。
  - リポジトリ統合テストでロール更新と管理者数検証を確認。
  - 通知サービスのモックを用意し、イベント発火を検証。

## フロントエンド実装計画
- OpenAPI 更新後に `frontend` で `pnpm run generate:api` を実行し、クライアントコードを再生成。
- TanStack Query hooks を追加。
  - `useChannelMembers(channelId)`
  - `useInviteChannelMember(channelId)`
  - `useJoinChannel(channelId)` / `useLeaveChannel(channelId)`
  - `useUpdateChannelMemberRole(channelId)`
  - いずれも成功時は `invalidateQueries(["channels", channelId, "members"])` を実行。
- UI 実装:
  - `ChannelInfoPanel` もしくは新規モーダルでメンバー一覧と各操作を提供。
  - ワークスペースメンバー一覧（`useMembers`）から未参加者を抽出して招待候補に表示。
  - パブリックチャンネルでは「参加する」「離脱する」ボタンを追加。
  - 管理者ロールのトグル UI（ドロップダウンやトグルスイッチ）を提供し、ラスト管理者の削除は警告ダイアログでブロック。
  - 操作結果はトースト通知で表示。
- 通知対応:
  - WebSocket クライアント (`frontend/src/lib/ws`) にチャンネルメンバー更新イベントを追加し、受信時に Query キャッシュを更新。
  - 通知内容を UI に反映（例: トースト、バッジ更新）。
- テスト:
  - hooks の単体テスト（Mock Service Worker を用いた API 成功/失敗ケース）。
  - UI コンポーネントの Vitest + Testing Library テストで招待・ロール変更・離脱フローを検証。
  - WebSocket イベント受信時の状態更新テスト。

## テスト戦略
- バックエンド:
  - ユースケース単体テストで全ロジック分岐を網羅。
  - HTTP ハンドラテストで認可・入力バリデーション・レスポンス整合性を確認。
  - リポジトリ統合テストでロール更新／管理者数検証を確認。
  - 通知・メールイベントが期待通り呼び出されることをモック検証。
- フロントエンド:
  - TanStack Query hooks と UI コンポーネントの単体テスト。
  - WebSocket イベントによるリアルタイム更新のテスト。
  - 必要に応じて Playwright などで E2E シナリオ（招待→通知確認→ロール変更→削除→離脱）を追加。
- 手動確認:
  - `docker-compose up -d --build` で環境を起動。
  - `alice@example.com`（管理者）でログインし、招待→通知確認→メッセージ閲覧可否→管理者ロール変更→削除→離脱を確認。
  - パブリックチャンネルで自己参加／離脱が行えることを確認。

## 運用・移行
- Atlas でスキーマ変更を適用し、既存データについてロール初期化を行う。
- 新しい通知イベントに合わせて監視・ログ出力を追加。
- メール通知に必要なテンプレートや送信設定を整備。
- ドキュメント（API 仕様、運用手順）を更新。

## 未解決事項
- 現時点なし。
