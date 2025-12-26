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
├── vite-env.d.ts            # Vite環境変数型定義
├── styles/                  # グローバルスタイル
├── lib/                     # 共通ライブラリ
│   ├── api/                 # APIクライアント
│   ├── query.ts             # TanStack Query設定
│   ├── store/               # Jotai状態管理
│   └── ws/                  # WebSocketクライアント
├── features/                # 機能別モジュール
│   ├── auth/                # 認証機能
│   └── workspace/           # ワークスペース機能
├── components/              # 共有コンポーネント
└── test/                    # テスト設定
```

## 環境変数

`.env`ファイルを作成して以下を設定:

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080
```

## 実装済み機能

- ✅ 認証（ログイン/登録/ログアウト）
- ✅ ワークスペース一覧/作成
- ✅ WebSocketクライアント（基本実装）
- ✅ PWA対応
- ✅ 自動認証リフレッシュ

## 未実装機能

- チャネル機能
- メッセージ機能
- 添付ファイル
- 未読管理
- リアルタイム通知

## ライセンス

MIT
