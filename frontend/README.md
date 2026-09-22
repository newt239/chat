# Chat Frontend

Slack風コミュニケーションアプリのフロントエンドアプリケーション

## 技術スタック

- **React 19** - UIフレームワーク
- **TypeScript** - 型安全性
- **Vite** - ビルドツール
- **Mantine 8** - UIコンポーネントライブラリ
- **Tailwind CSS** - ユーティリティファーストCSS
- **TanStack Query** - サーバー状態管理
- **Jotai** - クライアント状態管理
- **openapi-fetch** - 型安全なAPIクライアント
- **PWA** - プログレッシブウェブアプリ対応

## 開発開始

### インストール

```bash
pnpm install
```

### 開発サーバー起動

```bash
pnpm dev
```

ブラウザで http://localhost:5173 を開く

### ビルド

```bash
pnpm build
```

### プレビュー

```bash
pnpm preview
```

### テスト

```bash
# テスト実行
pnpm test

# テストUI
pnpm test:ui
```

### Lint & Format

```bash
# Lint
pnpm lint

# Format
pnpm format

# 型チェック
pnpm typecheck
```

## API型定義の生成

バックエンドのOpenAPIスキーマから型定義を生成:

```bash
pnpm run generate:api
```

## プロジェクト構成

```
src/
├── main.tsx                 # エントリーポイント
├── App.tsx                  # ルートコンポーネント
├── routes/                  # ルート定義（React Router Data モード）
├── styles/                  # グローバルスタイル
├── lib/                     # 共通ライブラリ
│   ├── api/                 # APIクライアント（OpenAPI 生成型）
│   ├── paths.ts             # パスビルダー
│   ├── routeParams.ts       # ルートパラメータ取得
│   └── ws.ts                # WebSocketクライアント
├── providers/               # Jotai ストア / TanStack Query / WebSocket
├── features/                # 機能別モジュール
│   ├── attachment/ auth/ bookmark/ channel/ dm/ layout/ link/
│   ├── member/ message/ notification/ pin/ reaction/ search/
│   └── settings/ thread/ userGroup/ workspace/
└── types/                   # WebSocket イベントなどの型定義
```

## 環境変数

`.env`ファイルを作成して以下を設定:

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080
```

## 実装済み機能

- 認証（ログイン / 登録 / ログアウト / パスワード変更 / アカウント削除 / 自動リフレッシュ）
- ワークスペース（一覧 / 作成 / 設定 / メンバー管理 / 公開ワークスペースへの参加）
- チャンネル（一覧 / 作成 / 設定 / メンバー管理 / 参加 / 退出）
- DM・グループ DM の作成
- メッセージ（投稿 / 編集 / 削除 / Markdown / メンション / リンクプレビュー / 添付ファイル / 過去の読み込み）
- スレッド（返信 / 参加中スレッド一覧 / 既読）
- リアクション・ピン留め・ブックマーク
- 未読バッジと通知、入力中インジケータ（WebSocket 経由でリアルタイム更新）
- 検索（メッセージ / チャンネル / ユーザー / ユーザーグループ）
- ユーザーグループの作成とメンバー管理
- PWA 対応

## ライセンス

MIT
