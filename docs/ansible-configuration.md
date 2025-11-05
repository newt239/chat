# Ansible 構成ガイド

## 概要

このプロジェクトでは Ansible を使用して、本番環境とプレビュー環境のデプロイを自動化しています。
ConoHa VPS 上の Ubuntu サーバーに対して、Docker Compose を使用したアプリケーションのデプロイが行われます。

## ディレクトリ構成

```
ansible/
├── group_vars/
│   └── production.yml        # 環境変数・設定値
├── playbooks/
│   ├── site.yml              # 本番環境デプロイ用プレイブック
│   └── preview.yml           # プレビュー環境デプロイ用プレイブック
├── roles/
│   ├── base/                 # 基本セットアップ
│   ├── docker/               # Docker インストール
│   ├── app/                  # 本番環境アプリケーション
│   └── app_preview/          # プレビュー環境アプリケーション
└── requirements.yml          # Ansible コレクション依存関係
```

## 環境変数設定

### group_vars/production.yml

本番環境とプレビュー環境の両方で使用される設定値を定義しています。

#### ドメイン設定

- `domain`: 本番環境のドメイン名 (`chat.newt239.dev`)
- `caddy_email`: SSL 証明書発行用のメールアドレス

#### データベース設定

- `postgres.user`: PostgreSQL のユーザー名
- `postgres.password`: PostgreSQL のパスワード (要変更)
- `postgres.db`: データベース名 (`chat`)

#### セキュリティ設定

- `jwt_secret`: JWT トークン署名用のシークレットキー (要変更)
- `cors_allowed_origins`: CORS 許可オリジン

#### オブジェクトストレージ設定 (Wasabi)

- `wasabi.access_key_id`: アクセスキー ID
- `wasabi.secret_access_key`: シークレットアクセスキー
- `wasabi.bucket`: バケット名
- `wasabi.region`: リージョン (`ap-northeast-1`)
- `wasabi.endpoint`: エンドポイント URL

**注意**: `<CHANGE_ME_*>` となっている値は実際の環境に合わせて変更する必要があります。

## プレイブック

### site.yml (本番環境)

本番環境へのデプロイを行うプレイブックです。

```yaml
- hosts: chat-prod
  become: true
  vars_files:
    - ../group_vars/production.yml
  roles:
    - base
    - docker
    - app
```

**実行される順序**:

1. `base`: 基本パッケージのインストールとサーバーセットアップ
2. `docker`: Docker のインストールと設定
3. `app`: 本番アプリケーションのデプロイ

### preview.yml (プレビュー環境)

プレビュー環境へのデプロイを行うプレイブックです。

```yaml
- hosts: chat-preview
  become: true
  vars_files:
    - ../group_vars/production.yml
  vars:
    branch: "" # デプロイするブランチ名
    preview_port: 8081 # プレビュー環境のポート
    caddy_http_port: 18080
    caddy_https_port: 18443
    preview_domain: "" # プレビュー環境のドメイン
  roles:
    - app_preview
```

**変数**:

- `branch`: GitHub のブランチ名を指定
- `preview_port`: プレビュー環境の内部ポート
- `preview_domain`: プレビュー環境のドメイン名

## ロール詳細

### base ロール

基本的なサーバーセットアップを行います。

**タスク**:

1. 必要パッケージのインストール
   - git, curl, ca-certificates, gnupg, lsb-release, awscli, ufw
2. `deploy` ユーザーを `sudo` グループに追加
3. UFW ファイアウォール設定
   - ポート 22 (SSH), 80 (HTTP), 443 (HTTPS) を許可
4. 作業ディレクトリの作成
   - `/opt/chat`: 本番環境用
   - `/opt/chat-preview`: プレビュー環境用
   - `/opt/backups`: バックアップ用

### docker ロール

Docker をインストールし、実行環境を整えます。

**タスク**:

1. Docker 公式 GPG キーの配置
2. Docker リポジトリの設定
3. Docker 関連パッケージのインストール
   - docker-ce
   - docker-ce-cli
   - containerd.io
   - docker-buildx-plugin
   - docker-compose-plugin
4. `deploy` ユーザーを `docker` グループに追加

### app ロール (本番環境)

本番環境のアプリケーションをデプロイします。

**タスク**:

1. GitHub リポジトリのクローン/更新
   - リポジトリ: `https://github.com/newt239/chat.git`
   - ブランチ: `main`
   - デプロイ先: `/opt/chat`
2. `.env.production` ファイルの生成
   - テンプレート: `env.production.j2`
   - 環境変数を自動設定
3. Docker Compose の起動
   - ファイル: `docker-compose.production.yml`
   - ビルドも実行
4. データベースバックアップ cron の設定
   - 実行タイミング: 毎日 3:00
   - スクリプト: `./scripts/backup.sh`
   - ログ: `/var/log/chat_backup.log`

**環境変数テンプレート (env.production.j2)**:

- データベース接続情報
- JWT シークレット
- CORS 設定
- Wasabi (S3) 設定
- Caddy (リバースプロキシ) 設定

### app_preview ロール (プレビュー環境)

プレビュー環境のアプリケーションをデプロイします。

**タスク**:

1. ブランチ名のサニタイズ
   - スラッシュ (`/`) をハイフン (`-`) に変換
   - 例: `feature/new-feature` → `feature-new-feature`
2. プレビュー用ディレクトリの作成
   - パス: `/opt/chat-preview/{{ sanitized_branch }}`
3. GitHub リポジトリのクローン/更新
   - 指定されたブランチを取得
4. `.env.preview` ファイルの生成
   - テンプレート: `env.preview.j2`
   - プレビュー環境用の設定
5. Docker ネットワークの作成
   - ネットワーク名: `chat_chat-network`
   - 本番環境のデータベースと共有するため
6. Docker Compose の起動
   - ファイル: `docker-compose.preview.yml`
   - プロジェクト名: `chat-preview-{{ sanitized_branch }}`

**環境変数テンプレート (env.preview.j2)**:

- プレビュー専用のデータベース (`chat_preview`)
- プレビューポート設定
- ブランチ名情報
- 本番環境と同じ Wasabi 設定

## デプロイ方法

### 環境変数の設定

Ansible のテンプレートファイルは、実行時の環境変数から機密情報を取得します。

#### GitHub Actions での実行

GitHub Actions では、リポジトリの Secrets に登録された環境変数が自動的に利用されます。特別な設定は不要です。

#### ローカルでの実行

ローカルから Ansible を実行する場合は、以下の手順で環境変数を設定します。

1. `.env.example`を`.env`にコピー

   ```bash
   cp .env.example .env
   ```

2. `.env`ファイルを編集して実際の値を設定

   ```bash
   # PostgreSQL設定
   POSTGRES_PASSWORD=your_secure_password_here

   # JWT設定
   JWT_SECRET=your_jwt_secret_key_here_change_this_in_production

   # Wasabi S3設定
   WASABI_ACCESS_KEY_ID=your_wasabi_access_key
   WASABI_SECRET_ACCESS_KEY=your_wasabi_secret_key
   WASABI_BUCKET_NAME=your_bucket_name
   WASABI_REGION=ap-northeast-1
   WASABI_ENDPOINT=https://ap-northeast-1.s3.wasabisys.com
   ```

3. 環境変数を読み込んで Ansible を実行

   ```bash
   # .envファイルから環境変数を読み込み
   set -a
   source .env
   set +a
   ```

### 本番環境へのデプロイ

```bash
cd ansible
ansible-playbook \
  -i 'chat-prod,' \
  -e "ansible_host=YOUR_SERVER_IP ansible_user=deploy ansible_port=22" \
  playbooks/site.yml
```

**注意**:

- 初回実行時は `ansible_user=root` に変更してください
- 2 回目以降は `ansible_user=deploy` を使用してください
- `YOUR_SERVER_IP` を実際のサーバーの IP アドレスに置き換えてください

### プレビュー環境へのデプロイ

```bash
cd ansible
ansible-playbook \
  -i 'chat-preview,' \
  -e "ansible_host=YOUR_SERVER_IP ansible_user=deploy ansible_port=22" \
  playbooks/preview.yml \
  -e "branch=feature/new-feature" \
  -e "preview_domain=preview.chat.newt239.dev"
```

**注意**:

- `YOUR_SERVER_IP` を実際のサーバーの IP アドレスに置き換えてください

## 依存パッケージ

### requirements.yml

Ansible コレクションの依存関係を定義しています。

```yaml
collections:
  - name: community.docker
```

**インストール方法**:

```bash
ansible-galaxy collection install -r requirements.yml
```

## トラブルシューティング

### よくある問題

1. **Docker Compose が起動しない**

   - `community.docker` コレクションがインストールされているか確認
   - Docker サービスが起動しているか確認: `systemctl status docker`

2. **環境変数が反映されない**

   - `ansible-playbook` 実行前に環境変数がエクスポートされているか確認
   - テンプレートファイルの変数名が正しいか確認

3. **プレビュー環境のポート競合**
   - 既存のプレビュー環境と異なるポート番号を指定する
   - `-e "preview_port=8082"` のように明示的に指定

## 参考情報

- Ansible 公式ドキュメント: https://docs.ansible.com/
- Docker Compose Ansible モジュール: https://docs.ansible.com/ansible/latest/collections/community/docker/
- UFW (Uncomplicated Firewall): https://help.ubuntu.com/community/UFW

```

```
