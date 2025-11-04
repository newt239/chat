# ConoHa VPS デプロイ手順書

このドキュメントでは、チャットアプリケーションを ConoHa VPS にデプロイし、GitHub Actions による自動デプロイを設定する手順を説明します。

## 目次

1. [必要なもの](#必要なもの)
2. [ConoHa VPS の初期設定](#conoha-vpsの初期設定)
3. [サーバーの準備](#サーバーの準備)
4. [GitHub Actions の設定](#github-actionsの設定)
5. [本番環境のデプロイ](#本番環境のデプロイ)
6. [プレビュー環境のデプロイ](#プレビュー環境のデプロイ)
7. [トラブルシューティング](#トラブルシューティング)

---

## 必要なもの

### サービス・アカウント

- ConoHa VPS アカウント
- GitHub アカウント
- ドメイン名（本番環境用）
- Wasabi アカウント（S3 互換ストレージ）

### ローカル環境

- SSH クライアント
- Git

---

## ConoHa VPS の初期設定

### 1. VPS インスタンスの作成

1. ConoHa コントロールパネルにログイン
2. 「サーバー追加」をクリック
3. 以下の設定を推奨：

```
プラン: VPS (2GB以上推奨)
イメージタイプ: Ubuntu 24.04 LTS
rootパスワード: 強固なパスワードを設定
SSH Key: 公開鍵を登録（推奨）
```

4. サーバーを起動し、IP アドレスを確認

### 2. ドメインの設定

1. DNS レコードを設定

```
# 本番環境
A レコード: your-domain.com → ConoHa VPSのIPアドレス

# プレビュー環境（任意）
A レコード: *.preview.your-domain.com → ConoHa VPSのIPアドレス
```

2. DNS 設定が反映されるまで待機（最大 48 時間）

---

## サーバーの準備

### 1. サーバーに SSH 接続

```bash
ssh root@<ConoHa VPSのIPアドレス>
```

### 2. 作業用ユーザーの作成

```bash
# ユーザーを作成
adduser deploy

# sudo権限を付与
usermod -aG sudo deploy

# deployユーザーに切り替え
su - deploy
```

### 3. 必要なパッケージのインストール

```bash
# システムアップデート
sudo apt update && sudo apt upgrade -y

# 必要なパッケージをインストール
sudo apt install -y \
    git \
    curl \
    ca-certificates \
    gnupg \
    lsb-release \
    awscli
```

### 4. Docker のインストール

```bash
# Dockerの公式GPGキーを追加
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

# Dockerリポジトリを追加
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Dockerをインストール
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# deployユーザーをdockerグループに追加
sudo usermod -aG docker deploy

# 設定を反映するため一度ログアウトして再ログイン
exit
su - deploy

# Dockerの動作確認
docker --version
docker compose version
```

### 5. アプリケーションディレクトリの作成

```bash
# 本番環境用ディレクトリ
sudo mkdir -p /opt/chat
sudo chown deploy:deploy /opt/chat

# プレビュー環境用ディレクトリ
sudo mkdir -p /opt/chat-preview
sudo chown deploy:deploy /opt/chat-preview
```

### 6. リポジトリのクローン

```bash
cd /opt/chat
git clone https://github.com/<your-username>/<your-repo>.git .
```

### 7. 環境変数ファイルの作成

```bash
# 本番環境の環境変数を設定
cd /opt/chat
cp .env.production.example .env.production

# エディタで環境変数を編集
nano .env.production
```

以下の環境変数を設定してください：

```bash
# データベース設定
POSTGRES_DB=chat
POSTGRES_USER=postgres
POSTGRES_PASSWORD=<強固なパスワードを設定>
DATABASE_URL=postgresql://postgres:<パスワード>@db:5432/chat?sslmode=disable

# バックエンド設定
PORT=8080
JWT_SECRET=<ランダムな文字列を生成して設定>
CORS_ALLOWED_ORIGINS=https://your-domain.com

# Wasabi S3設定（バックアップアップロードに使用）
WASABI_ACCESS_KEY_ID=<WasabiアクセスキーID>
WASABI_SECRET_ACCESS_KEY=<Wasabiシークレットアクセスキー>
WASABI_BUCKET_NAME=<バケット名>
# 任意: リージョン/エンドポイント（未設定時は us-east-1 / https://s3.wasabisys.com）
# WASABI_REGION=us-east-1
# WASABI_ENDPOINT=https://s3.wasabisys.com

# Caddy設定
DOMAIN=your-domain.com
CADDY_EMAIL=your-email@example.com
```

**JWT_SECRET の生成例：**

```bash
openssl rand -base64 32
```

### 8. ファイアウォールの設定

```bash
# UFWをインストール（未インストールの場合）
sudo apt install -y ufw

# SSH接続を許可
sudo ufw allow 22/tcp

# HTTP/HTTPSを許可
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# ファイアウォールを有効化
sudo ufw enable

# 設定確認
sudo ufw status
```

### 9. 初回デプロイの実行

```bash
cd /opt/chat

# 環境変数をロード
export $(cat .env.production | grep -v '^#' | xargs)

# Dockerコンテナを起動
docker compose -f docker-compose.production.yml up -d --build

# ログを確認
docker compose -f docker-compose.production.yml logs -f
```

### 10. 動作確認

ブラウザで `https://your-domain.com` にアクセスし、アプリケーションが正常に動作することを確認します。

---

## GitHub Actions の設定

### 1. SSH 鍵の作成（ローカル環境）

```bash
# SSH鍵ペアを生成
ssh-keygen -t ed25519 -C "github-actions" -f ~/.ssh/github_actions_chat

# 公開鍵の内容を確認
cat ~/.ssh/github_actions_chat.pub

# 秘密鍵の内容を確認（GitHubに設定する）
cat ~/.ssh/github_actions_chat
```

### 2. サーバーに公開鍵を追加

```bash
# ConoHa VPSにSSH接続
ssh deploy@<ConoHa VPSのIPアドレス>

# authorized_keysに公開鍵を追加
mkdir -p ~/.ssh
chmod 700 ~/.ssh
echo "<公開鍵の内容>" >> ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys
```

### 3. GitHub リポジトリに Secrets を設定

GitHub リポジトリの「Settings」→「Secrets and variables」→「Actions」から、以下の Secrets を追加します：

#### 本番環境用

| Secret 名             | 値                        | 説明                                            |
| --------------------- | ------------------------- | ----------------------------------------------- |
| `SSH_PRIVATE_KEY`     | SSH 秘密鍵の内容          | GitHub Actions がサーバーに接続するための秘密鍵 |
| `SSH_USER`            | `deploy`                  | SSH 接続ユーザー名                              |
| `PRODUCTION_HOST`     | ConoHa VPS の IP アドレス | 本番環境のホスト                                |
| `PRODUCTION_SSH_PORT` | `22`                      | SSH 接続ポート                                  |
| `PRODUCTION_DOMAIN`   | `your-domain.com`         | 本番環境のドメイン                              |

#### プレビュー環境用（任意）

| Secret 名          | 値                                            | 説明                   |
| ------------------ | --------------------------------------------- | ---------------------- |
| `PREVIEW_HOST`     | ConoHa VPS の IP アドレス（本番と同じでも可） | プレビュー環境のホスト |
| `PREVIEW_SSH_PORT` | `22`                                          | SSH 接続ポート         |

### 4. GitHub Actions の動作確認

1. main ブランチにプッシュして、自動デプロイが実行されることを確認

```bash
git add .
git commit -m "feat: GitHub Actionsのセットアップ"
git push origin main
```

2. GitHub の「Actions」タブでワークフローの実行状況を確認

---

## 本番環境のデプロイ

### 自動デプロイ

main ブランチにプッシュすると、GitHub Actions が自動的にデプロイを実行します。

```bash
git push origin main
```

### 手動デプロイ

サーバーに直接 SSH 接続して手動デプロイすることも可能です。

```bash
# サーバーにSSH接続
ssh deploy@<ConoHa VPSのIPアドレス>

# デプロイスクリプトを実行
cd /opt/chat
./scripts/deploy.sh production main

# デプロイ前自動バックアップについて
# 上記のデプロイスクリプトおよび GitHub Actions は、デプロイ直前に
# `scripts/backup.sh` を実行し、PostgreSQL のダンプを Wasabi にアップロードします。
# バケット内の保存先: s3://$WASABI_BUCKET_NAME/db-backups/<DB名>_YYYYmmdd_HHMMSS.sql.gz
```

### デプロイの確認

```bash
# コンテナの状態を確認
docker compose -f docker-compose.production.yml ps

# ログを確認
docker compose -f docker-compose.production.yml logs -f
```

---

## プレビュー環境のデプロイ

### 自動デプロイ

main 以外のブランチにプッシュすると、プレビュー環境が自動的に作成されます。

```bash
# 新しいブランチを作成
git checkout -b feature/new-feature

# 変更をコミット
git add .
git commit -m "feat: 新機能の実装"

# プッシュ（自動デプロイが実行される）
git push origin feature/new-feature
```

### 手動デプロイ

```bash
# サーバーにSSH接続
ssh deploy@<ConoHa VPSのIPアドレス>

# デプロイスクリプトを実行
cd /opt/chat
./scripts/deploy.sh preview feature/new-feature
```

### プレビュー環境のクリーンアップ

不要になったプレビュー環境は、以下のコマンドで削除できます。

```bash
# サーバーにSSH接続
ssh deploy@<ConoHa VPSのIPアドレス>

# クリーンアップスクリプトを実行
cd /opt/chat
./scripts/cleanup-preview.sh feature/new-feature
```

---

## トラブルシューティング

### デプロイが失敗する場合

#### 1. ログを確認

```bash
# GitHub Actionsのログを確認
# GitHubの「Actions」タブから該当のワークフローを開く

# サーバー側のログを確認
ssh deploy@<ConoHa VPSのIPアドレス>
cd /opt/chat
docker compose -f docker-compose.production.yml logs -f
```

#### 2. コンテナの状態を確認

```bash
docker compose -f docker-compose.production.yml ps
```

#### 3. 環境変数を確認

```bash
cat .env.production
```

### データベース接続エラー

```bash
# データベースコンテナのログを確認
docker compose -f docker-compose.production.yml logs db

# データベースコンテナに接続してテスト
docker compose -f docker-compose.production.yml exec db psql -U postgres -d chat
```

### SSL 証明書の問題

Caddy は自動的に Let's Encrypt から SSL 証明書を取得しますが、以下の条件が必要です：

1. ドメインの DNS 設定が正しいこと
2. ポート 80 と 443 が開いていること
3. ドメインが正しく解決されること

```bash
# Caddyのログを確認
docker compose -f docker-compose.production.yml logs caddy
```

### ディスク容量不足

```bash
# ディスク使用量を確認
df -h

# 未使用のDockerイメージを削除
docker system prune -a

# 古いDockerボリュームを削除
docker volume prune
```

### メモリ不足

```bash
# メモリ使用量を確認
free -h

# 不要なプレビュー環境を削除
./scripts/cleanup-preview.sh <branch-name>
```

---

## セキュリティのベストプラクティス

### 1. 定期的なアップデート

```bash
# システムパッケージのアップデート
sudo apt update && sudo apt upgrade -y

# Dockerイメージのアップデート
cd /opt/chat
docker compose -f docker-compose.production.yml pull
docker compose -f docker-compose.production.yml up -d
```

### 2. SSH 接続の強化

```bash
# SSH設定を編集
sudo nano /etc/ssh/sshd_config

# 以下の設定を変更
PermitRootLogin no
PasswordAuthentication no
PubkeyAuthentication yes

# SSH サービスを再起動
sudo systemctl restart sshd
```

### 3. ファイアウォールの見直し

```bash
# 必要最小限のポートのみ開放
sudo ufw status numbered
```

### 4. ログの監視

```bash
# アプリケーションログの定期確認
docker compose -f docker-compose.production.yml logs --tail=100

# システムログの確認
sudo journalctl -u docker -n 100
```

---

## バックアップ

### データベースのバックアップ（手動）

```bash
# バックアップディレクトリを作成
mkdir -p /opt/backups

# データベースをバックアップ（手動実行）
/opt/chat/scripts/backup.sh

# 古いバックアップを削除（30日以上前）
find /opt/backups -name "chat_*.sql" -mtime +30 -delete
```

### 自動バックアップの設定（任意）

```bash
# cronで毎日午前3時にバックアップを実行
crontab -e

# 以下を追加（毎日3時にWasabiへアップロードも行う）
0 3 * * * /opt/chat/scripts/backup.sh >> /var/log/chat_backup.log 2>&1
```

### バックアップからのリストア

```bash
# データベースコンテナに接続
docker compose -f docker-compose.production.yml exec -T db psql -U postgres chat < /opt/backups/chat_20250101_030000.sql
```

---

## パフォーマンス最適化

### 1. Docker のログローテーション

```bash
# /etc/docker/daemon.json を編集
sudo nano /etc/docker/daemon.json

# 以下を追加
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}

# Dockerを再起動
sudo systemctl restart docker
```

### 2. リソース監視

```bash
# Dockerコンテナのリソース使用状況を確認
docker stats

# システムリソースを確認
htop
```

---

## まとめ

このドキュメントでは、ConoHa VPS へのデプロイと GitHub Actions による自動デプロイの設定方法を説明しました。

### デプロイフロー

```
開発者がコードをプッシュ
    ↓
GitHub Actionsが自動実行
    ↓
SSH経由でConoHa VPSに接続
    ↓
最新コードを取得
    ↓
Dockerイメージをビルド
    ↓
コンテナを再起動
    ↓
デプロイ完了
```

### 重要なポイント

- main ブランチへのプッシュで本番環境が自動更新されます
- 他のブランチへのプッシュでプレビュー環境が作成されます
- SSL 証明書は Caddy が自動で管理します
- 環境変数は適切に設定してください
- 定期的なバックアップとセキュリティアップデートを忘れずに

### サポート

問題が発生した場合は、以下を確認してください：

1. GitHub Actions のログ
2. Docker コンテナのログ
3. サーバーのシステムログ
4. このドキュメントのトラブルシューティングセクション
