# データベースマイグレーション手順

このドキュメントでは、データベースのマイグレーション手順と、本番環境・プレビュー環境でのデータベース管理方法について説明します。

## 目次

1. [データベース構成](#データベース構成)
2. [マイグレーション戦略](#マイグレーション戦略)
3. [マイグレーションの実行](#マイグレーションの実行)
4. [デプロイ時のマイグレーション手順](#デプロイ時のマイグレーション手順)
5. [トラブルシューティング](#トラブルシューティング)

---

## データベース構成

このプロジェクトでは、本番環境とプレビュー環境で**独立したデータベース**を使用します。

### 構成図

```
PostgreSQL (chat-postgres-prod)
  └─ データベース: chat (本番環境専用)

PostgreSQL (chat-postgres-preview)
  └─ データベース: chat_preview (全プレビューブランチ共通)
```

### 特徴

**本番環境:**
- 専用のPostgreSQLインスタンス
- データベース名: `chat`
- 本番データのみ保持

**プレビュー環境:**
- 専用のPostgreSQLインスタンス
- データベース名: `chat_preview`
- **全てのプレビューブランチで同じデータを共有**
- 開発用のテストデータを保持

---

## マイグレーション戦略

このプロジェクトでは、[ent](https://entgo.io/)の自動マイグレーション機能を使用してデータベーススキーマを管理しています。

### entの自動マイグレーション

entは、Goのコードで定義されたスキーマから自動的にデータベーステーブルを作成・更新します。

- **`migrate`コマンド**: 既存のテーブルを保持しながら、新しいカラムやテーブルを追加
- **`reset`コマンド**: すべてのテーブルを削除して再作成（開発環境のみ）

### 重要な原則

- **本番環境では `reset` を実行しない** - データが全削除されます
- **マイグレーションは手動で適用** - デプロイ後に手動実行します
- **バックアップを必ず取る** - マイグレーション前に必ずバックアップ

---

## マイグレーションの実行

### マイグレーションコマンド

#### 1. マイグレーションの適用

```bash
# 本番環境
./scripts/migrate.sh apply production

# プレビュー環境
./scripts/migrate.sh apply preview
```

このコマンドは、entの`client.Schema.Create()`を実行して、既存のテーブルを保持しながら新しいカラムやテーブルを追加します。

#### 2. シードデータの投入

```bash
# 本番環境（確認プロンプトあり）
./scripts/migrate.sh seed production

# プレビュー環境（確認プロンプトあり）
./scripts/migrate.sh seed preview
```

---

## デプロイ時のマイグレーション手順

### 本番環境へのデプロイ

#### ステップ1: バックアップ

```bash
ssh deploy@<server-ip>
cd /opt/chat
docker compose -f docker-compose.production.yml exec -T db pg_dump -U postgres chat > /opt/backups/chat_$(date +%Y%m%d_%H%M%S).sql
```

#### ステップ2: デプロイ

```bash
git push origin main  # GitHub Actionsが自動デプロイ
```

#### ステップ3: マイグレーション適用

```bash
./scripts/migrate.sh apply production
```

#### ステップ4: 動作確認

```bash
docker compose -f docker-compose.production.yml logs -f backend
docker compose -f docker-compose.production.yml exec db psql -U postgres chat -c "\dt"
```

### プレビュー環境へのデプロイ

#### ステップ1: ブランチをプッシュ

```bash
git push origin feature/new-feature
# GitHub Actionsが自動デプロイ
```

#### ステップ2: マイグレーション適用（必要な場合）

```bash
ssh deploy@<server-ip>
cd /opt/chat
./scripts/migrate.sh apply preview
```

**注意:** プレビュー環境は全ブランチで同じデータベースを共有するため、マイグレーションは一度だけ実行すればOKです。

#### ステップ3: シードデータ投入（初回のみ）

```bash
./scripts/migrate.sh seed preview
```

---

## スキーマ変更のワークフロー

### 1. ローカル開発環境での変更

```bash
# entスキーマを編集
vi backend/ent/schema/user.go

# entコードを生成
cd backend
go generate ./ent

# ローカルで動作確認
docker-compose up -d --build
```

### 2. プレビュー環境でのテスト

```bash
# ブランチにプッシュ
git add .
git commit -m "feat: add new column to user table"
git push origin feature/add-user-column

# プレビュー環境が自動作成される
# マイグレーションを適用
ssh deploy@<server-ip>
cd /opt/chat
./scripts/migrate.sh apply preview
```

### 3. 本番環境へのデプロイ

```bash
# mainブランチにマージ
git checkout main
git merge feature/add-user-column
git push origin main

# 本番環境が自動デプロイされる
# マイグレーションを手動で適用
ssh deploy@<server-ip>
cd /opt/chat

# バックアップ
docker compose -f docker-compose.production.yml exec -T db pg_dump -U postgres chat > /opt/backups/chat_before_migration_$(date +%Y%m%d_%H%M%S).sql

# マイグレーション適用
./scripts/migrate.sh apply production
```

---

## プレビュー環境のデータベース

### データの共有

- 全てのプレビューブランチは同じデータベース（`chat_preview`）を使用します
- ブランチAで作成したデータは、ブランチBでも参照できます
- テストデータは全ブランチで共通です

### データのリセット（必要な場合）

プレビュー環境のデータをリセットする場合は、以下のコマンドを実行します：

```bash
# プレビュー環境のデータベースをリセット
ssh deploy@<server-ip>
cd /opt/chat
docker compose -f docker-compose.production.yml exec db-preview psql -U postgres -c "DROP DATABASE IF EXISTS chat_preview;"
docker compose -f docker-compose.production.yml exec db-preview psql -U postgres -c "CREATE DATABASE chat_preview;"

# マイグレーション適用
./scripts/migrate.sh apply preview

# シードデータ投入
./scripts/migrate.sh seed preview
```

---

## トラブルシューティング

### マイグレーションが失敗する

```bash
# データベースログを確認
docker compose -f docker-compose.production.yml logs db

# 本番環境の場合
docker compose -f docker-compose.production.yml logs backend

# プレビュー環境の場合
docker ps --filter "name=chat-backend-preview-" --format "{{.Names}}" | xargs docker logs
```

### ロールバック

```bash
# バックアップからリストア
cd /opt/chat
docker compose -f docker-compose.production.yml exec -T db psql -U postgres chat < /opt/backups/chat_20250101_030000.sql
docker compose -f docker-compose.production.yml restart backend
```

### プレビュー環境でデータが見えない

プレビュー環境は全ブランチで同じデータベースを共有しているため、以下を確認してください：

1. データベースが起動しているか確認
```bash
docker compose -f docker-compose.production.yml ps db-preview
```

2. プレビュー環境の接続設定を確認
```bash
cat /opt/chat-preview/<branch-name>/.env.preview | grep DATABASE_URL
```

正しい設定:
```
PREVIEW_DATABASE_URL=postgresql://postgres:<password>@chat-postgres-preview:5432/chat_preview?sslmode=disable
```

---

## ベストプラクティス

### 1. マイグレーション前のチェックリスト

- [ ] バックアップを取得済み
- [ ] プレビュー環境でテスト済み
- [ ] マイグレーション内容を確認済み
- [ ] ロールバック手順を準備済み（バックアップから）

### 2. 安全なマイグレーション

- **段階的な変更**: 大きな変更は複数のマイグレーションに分割
- **後方互換性**: 既存のコードが動作し続けるようにする
- **テストデータ**: プレビュー環境で十分にテスト
- **ピークタイムを避ける**: 利用者が少ない時間帯に実行

### 3. プレビュー環境の管理

- **データは共有される**: 全ブランチで同じデータを使用
- **マイグレーションは一度だけ**: 複数ブランチで重複実行不要
- **定期的なリセット**: 不要なデータが溜まったらリセット

---

## まとめ

このドキュメントでは、entの自動マイグレーションを使用したデータベース管理方法について説明しました。

### 重要なポイント

- **独立したデータベース**: 本番とプレビューは完全に分離
- **プレビューは共有**: 全プレビューブランチが同じデータベースを使用
- **entの自動マイグレーション**: 新しいテーブル・カラムを自動追加
- **手動実行**: デプロイ後にマイグレーションを手動で適用
- **必ずバックアップ**: 本番環境では必ずバックアップを取得

### 関連ドキュメント

- [デプロイ手順書](./deployment.md)
- [バックエンドアーキテクチャ](./backend-architecture.md)
- [ent公式ドキュメント](https://entgo.io/)
- [entマイグレーションガイド](https://entgo.io/docs/migrate)
