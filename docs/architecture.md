# アーキテクチャ設計書

## 概要

本アプリケーションはクリーンアーキテクチャの原則に従って設計されており、Gin から Echo への移行とアーキテクチャの改善が完了しています。

## アーキテクチャ図

```
┌─────────────────────────────────────────────────────────────────┐
│                    Presentation Layer                          │
│  (Echo Handlers, Middleware, WebSocket)                       │
│  - HTTP request/response handling                              │
│  - Input validation                                           │
│  - Error handling and formatting                              │
│  - WebSocket connection management                            │
└─────────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────────┐
│                     Use Case Layer                              │
│  (Application Business Rules)                                  │
│  - Orchestrate domain entities                                 │
│  - Implement application-specific logic                        │
│  - Define input/output DTOs                                    │
│  - Authentication, Workspace, Channel, Message, etc.          │
└─────────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────────┐
│                      Domain Layer                               │
│  (Enterprise Business Rules)                                   │
│  - Entities (User, Workspace, Channel, Message, etc.)          │
│  - Repository interfaces                                        │
│  - Domain services                                              │
│  - Business logic                                               │
└─────────────────────────────────────────────────────────────────┘
                           ↑
┌─────────────────────────────────────────────────────────────────┐
│                 Infrastructure Layer                            │
│  (Frameworks & Drivers)                                         │
│  - Database (GORM + PostgreSQL)                                │
│  - External services (OGP, S3, etc.)                            │
│  - Repository implementations                                   │
│  - Config, Logger, etc.                                         │
└─────────────────────────────────────────────────────────────────┘
```

## ディレクトリ構造

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # アプリケーションエントリーポイント
├── internal/
│   ├── domain/                  # Domain Layer（最内層）
│   │   ├── entity/             # エンティティ
│   │   │   ├── user.go
│   │   │   ├── workspace.go
│   │   │   ├── channel.go
│   │   │   ├── message.go
│   │   │   └── ...
│   │   ├── repository/         # リポジトリインターフェース
│   │   │   ├── user.go
│   │   │   ├── workspace.go
│   │   │   └── ...
│   │   └── errors/             # ドメインエラー
│   │       └── errors.go
│   │
│   ├── usecase/                 # Use Case Layer
│   │   ├── auth/
│   │   │   ├── interface.go    # UseCase インターフェース
│   │   │   ├── interactor.go   # UseCase 実装
│   │   │   └── dto.go          # Input/Output DTO
│   │   ├── workspace/
│   │   ├── channel/
│   │   ├── message/
│   │   └── ...
│   │
│   ├── adapter/                 # Adapter Layer（インターフェース実装）
│   │   ├── controller/         # Presentation Layer
│   │   │   ├── http/
│   │   │   │   ├── handler/   # Echoハンドラー
│   │   │   │   ├── middleware/
│   │   │   │   ├── router.go
│   │   │   │   └── validator.go
│   │   │   └── websocket/
│   │   │       ├── hub.go
│   │   │       └── connection.go
│   │   │
│   │   └── gateway/            # Infrastructure実装
│   │       ├── persistence/    # リポジトリ実装
│   │       │   ├── user.go
│   │       │   ├── workspace.go
│   │       │   └── ...
│   │       └── external/       # 外部サービス
│   │           └── ogp.go
│   │
│   ├── infrastructure/          # Infrastructure Layer
│   │   ├── database/
│   │   │   ├── models.go      # GORMモデル
│   │   │   └── connection.go
│   │   ├── config/
│   │   │   └── config.go
│   │   ├── logger/
│   │   │   └── logger.go
│   │   ├── auth/
│   │   │   ├── jwt.go
│   │   │   └── password.go
│   │   └── seed/
│   │       └── seed.go
│   │
│   ├── registry/                # DI Container
│   │   └── registry.go
│   │
│   └── test/                    # テスト用
│       ├── mocks/              # モック実装
│       └── integration/        # 統合テスト
│
└── docs/
    └── architecture.md
```

## 依存関係ルール

```
Domain Layer（最内層）
  ↑
  │ 依存
  │
Use Case Layer
  ↑
  │ 依存
  │
Adapter Layer (Controller + Gateway)
  ↑
  │ 依存
  │
Infrastructure Layer（最外層）
```

**重要な原則:**

- 内側の層は外側の層に依存してはいけない
- 依存性逆転の原則（DIP）を適用
- インターフェースを使って依存を抽象化

## 主要コンポーネント

### 1. Domain Layer

- **Entities**: ビジネスエンティティ（User, Workspace, Channel, Message 等）
- **Repository Interfaces**: データアクセスの抽象化
- **Domain Errors**: ビジネスルールに関するエラー

### 2. Use Case Layer

- **Interactors**: ビジネスロジックの実装
- **DTOs**: 入力・出力データの定義
- **Interfaces**: ユースケースの抽象化

### 3. Adapter Layer

#### Controller (Presentation)

- **HTTP Handlers**: Echo ベースの HTTP ハンドラー
- **WebSocket Handlers**: リアルタイム通信
- **Middleware**: 認証、CORS、バリデーション
- **Router**: ルーティング設定

#### Gateway (Infrastructure)

- **Persistence**: データベースアクセス実装
- **External Services**: 外部 API 連携

### 4. Infrastructure Layer

- **Database**: GORM + PostgreSQL
- **Authentication**: JWT 認証
- **Configuration**: 設定管理
- **Logging**: ログ出力

### 5. Registry (DI Container)

- **Dependency Injection**: 依存関係の管理
- **Service Locator**: サービスの取得

## 移行の成果

### Gin から Echo への移行

- **パフォーマンス向上**: Echo の軽量性を活用
- **標準互換性**: net/http との互換性向上
- **ミドルウェア改善**: より柔軟なミドルウェア設計

### クリーンアーキテクチャの実装

- **依存関係の整理**: 各層の責任を明確化
- **テスタビリティ向上**: モック化が容易
- **保守性向上**: 変更の影響範囲を限定

### テスト戦略

- **ユニットテスト**: 各層の独立したテスト
- **統合テスト**: エンドツーエンドのテスト
- **モック**: 外部依存の抽象化

## セキュリティ考慮事項

### 認証・認可

- **JWT 認証**: ステートレスな認証
- **トークン管理**: アクセストークンとリフレッシュトークン
- **セッション管理**: セキュアなセッション管理

### 入力検証

- **バリデーション**: 全入力の検証
- **SQL インジェクション対策**: GORM の活用
- **XSS 対策**: 適切なエスケープ処理

### CORS 設定

- **オリジン制限**: 本番環境での厳格な設定
- **認証情報**: 適切なクレデンシャル管理

## パフォーマンス考慮事項

### データベース

- **コネクションプール**: 適切な接続管理
- **N+1 問題回避**: 効率的なクエリ設計
- **インデックス**: 適切なインデックス設計

### キャッシュ

- **セッションキャッシュ**: 認証情報のキャッシュ
- **データキャッシュ**: 頻繁にアクセスされるデータのキャッシュ

### 非同期処理

- **WebSocket**: リアルタイム通信
- **バックグラウンド処理**: 重い処理の非同期化

## 今後の拡張性

### マイクロサービス化

- **サービス分割**: 機能ごとのサービス分割
- **API Gateway**: 統一された API 管理
- **サービス間通信**: gRPC や HTTP での通信

### スケーラビリティ

- **水平スケーリング**: 複数インスタンスでの運用
- **データベース分割**: 読み書きの分離
- **CDN**: 静的コンテンツの配信

### 監視・運用

- **メトリクス**: パフォーマンス監視
- **ログ**: 構造化ログの実装
- **トレーシング**: 分散トレーシング
