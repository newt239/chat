# インフラの構成と初回セットアップ

dev と prod を 1 つの GCP プロジェクト・1 つの GKE クラスタに namespace で分けて載せる。Terraform は `infra/terraform/`、Kubernetes のマニフェストは `infra/k8s/` にある。

```mermaid
flowchart LR
  user[ブラウザ] -->|HTTPS / WSS| cf[Cloudflare<br/>DNS + Tunnel]
  user -->|署名付き URL| wasabi[(Wasabi 東京<br/>添付ファイル)]

  subgraph gke[GKE Standard ゾーンクラスタ / Spot ノード]
    subgraph ns[namespace chat-dev / chat-prod]
      cfd[cloudflared ×2] --> fe[frontend ×2]
      cfd --> be[backend ×2]
      be --> redis[Redis<br/>Pub/Sub]
      be --> meili[Meilisearch<br/>StatefulSet + PVC]
      be --> proxy[Cloud SQL Auth Proxy]
    end
    eso[External Secrets]
  end

  cf --- cfd
  proxy -->|Private IP| sql[(Cloud SQL g1-small<br/>DB: chat_dev / chat_prod)]
  be --> wasabi
  be -->|ADC| fcm[FCM]
  sm[Secret Manager] -.-> eso
```

| 部分 | 置き場所 | 内容 |
| --- | --- | --- |
| 共有 | `infra/terraform/shared` | API の有効化、VPC と Cloud NAT、GKE（`asia-northeast1-b`、Spot の e2-medium を 2〜4 台）、Cloud SQL のインスタンス、Artifact Registry、GitHub Actions 用の Workload Identity Federation のプール |
| 環境ごと | `infra/terraform/envs/{dev,prod}`（中身は `modules/environment`） | DB とユーザー、backend の GSA と Workload Identity、Secret Manager のシークレット（ID は `<環境>-<変数名>`）、デプロイ用の GSA、Cloudflare Tunnel と DNS レコード、Wasabi のバケット |
| マニフェスト | `infra/k8s/base` と `overlays/{dev,prod}` | backend・frontend・cloudflared は 2 レプリカ、Redis と Meilisearch は 1 つ。環境ごとの値は `overlays/<環境>/params.yaml` と `components/gke`（Cloud SQL Auth Proxy・External Secrets・Workload Identity も持つ）で埋め込む |

知っておくこと:

- **backend は複数レプリカで動く**。WebSocket の配信・チャンネルの閲覧者一覧・Webhook のレート制限は Redis で共有する。Redis は永続化しないが、落ちても閲覧者は 30 秒ごとの延長で戻り、配信は再接続で続く
- **Spot ノード**は予告なく回収される。backend は SIGTERM で readiness を落とし、WebSocket を 1001 で閉じてから止まり、クライアントは他のレプリカへつなぎ直して取りこぼしを取り直す。2 台が同時に回収されると全体が止まる
- **Cloud SQL は dev と prod で共有**する（`max_connections` は 50）。backend は 1 レプリカあたり最大 10 接続（`DB_MAX_OPEN_CONNS`）。PITR と削除保護はインスタンス単位なので dev にも効く。prod の DB とユーザーは `terraform destroy` しても消えない（`deletion_policy = ABANDON`）
- **検索インデックス**は DB から作り直せる。Meilisearch のデータが消えても、backend は起動時にインデックスが空なら全件を登録し直す（`kubectl -n chat-dev rollout restart deployment/backend`）

## 初回セットアップ

前提: `gcloud`・`terraform`（1.11 以上）・`kubectl`・`kustomize`・`helm`・`gh` が使えること。以下の `PROJECT_ID` などは適宜置き換える。

### 1. プロジェクトと state バケット

```sh
gcloud auth login
gcloud auth application-default login
gcloud config set project PROJECT_ID
gcloud services enable storage.googleapis.com cloudresourcemanager.googleapis.com serviceusage.googleapis.com

gcloud storage buckets create gs://PROJECT_ID-tfstate --location=asia-northeast1 --uniform-bucket-level-access
gcloud storage buckets update gs://PROJECT_ID-tfstate --versioning
```

### 2. Cloudflare と Wasabi の準備

- 使うドメインを Cloudflare（Free プラン）に追加し、レジストラのネームサーバーを Cloudflare に向ける。アカウント ID とゾーン ID を控える（`cf auth whoami`・`cf zones list`）
- `frontend_domain` と `api_domain` は `chat.example.com` と `api-chat.example.com` のように 1 階層のサブドメインにする。Free プランの証明書は `*.example.com` までしか出ないため、`api.chat.example.com` は HTTPS にならない
- Cloudflare の API トークンを「Account → Cloudflare Tunnel: 編集」「Zone → DNS: 編集」の権限で作る。[cf](https://developers.cloudflare.com/changelog/post/2026-09-28-cloudflare-cli-beta/) なら次で作れる（トークンは作ったときにしか表示されない）

  ```sh
  export CLOUDFLARE_API_TOKEN=$(cf -q user tokens create --body "$(jq -n -c --arg account ACCOUNT_ID --arg zone ZONE_ID '{name: "chat-terraform", policies: [
    {effect: "allow", permission_groups: [{id: "c07321b023e944ff818fec44d8203567"}], resources: {("com.cloudflare.api.account." + $account): "*"}},
    {effect: "allow", permission_groups: [{id: "4755a26eedb94da69e1066d98aa820be"}], resources: {("com.cloudflare.api.account.zone." + $zone): "*"}}]}')" | jq -r .value)
  ```
- Wasabi で Terraform 用のアクセスキー（バケットを作れる権限）を作る
- アプリ用には、環境ごとのバケットだけを読み書きできるポリシーを付けたサブユーザーを作り、アクセスキーを発行しておく（手順 4 で Secret Manager に入れる）

Wasabi は 1TB 分の最低料金がかかり、削除したデータも 90 日分は課金される。契約前に最新の条件を確認する。

### 3. Terraform

shared を先に作り、次に各環境を作る。

```sh
cd infra/terraform/shared
cp terraform.tfvars.example terraform.tfvars   # project_id を書く
terraform init -backend-config="bucket=PROJECT_ID-tfstate"
terraform apply
terraform output

export CLOUDFLARE_API_TOKEN=...          # 手順 2 のトークン
export AWS_ACCESS_KEY_ID=...             # 手順 2 の Wasabi の Terraform 用キー
export AWS_SECRET_ACCESS_KEY=...
for env in dev prod; do
  cd ../envs/$env
  cp terraform.tfvars.example terraform.tfvars   # ドメイン、バケット名、Cloudflare の ID を書く
  terraform init -backend-config="bucket=PROJECT_ID-tfstate"
  terraform apply
  terraform output
done
```

Cloudflare Tunnel は Terraform が作り、`frontend_domain` と `api_domain` の CNAME も Cloudflare に登録される。LB・静的 IP・証明書は不要。

### 4. Wasabi のキーを Secret Manager に入れる

Terraform は箱だけを作るので、手順 2 のサブユーザーのキーを環境ごとに登録する。

```sh
for env in dev prod; do
  printf '%s' 'ACCESS_KEY_ID' | gcloud secrets versions add "$env-WASABI_ACCESS_KEY_ID" --data-file=-
  printf '%s' 'SECRET_ACCESS_KEY' | gcloud secrets versions add "$env-WASABI_SECRET_ACCESS_KEY" --data-file=-
done
```

### 5. External Secrets Operator

クラスタに 1 つだけ入れる。requests などは `infra/k8s/external-secrets/values.yaml` で管理する。

```sh
gcloud container clusters get-credentials chat --location=asia-northeast1-b
helm repo add external-secrets https://charts.external-secrets.io
helm upgrade --install external-secrets external-secrets/external-secrets \
  --namespace external-secrets --create-namespace \
  -f infra/k8s/external-secrets/values.yaml
```

### 6. マニフェストに値を反映

`infra/k8s/overlays/<環境>/` の 2 ファイルを `terraform output` の値で書き換えてコミットする。

- `params.yaml`: `PROJECT_ID`、`CLUSTER_NAME`・`CLUSTER_LOCATION`（shared の `gke_cluster_*`）、`BACKEND_SERVICE_ACCOUNT`（envs の `backend_service_account_email`）、`CLOUDSQL_CONNECTION_NAME`（shared の `cloudsql_connection_name`）
- `kustomization.yaml`: `images` の `newName`（shared の `artifact_registry_url` + `/backend`・`/frontend`）、`backend-env` の `CORS_ALLOWED_ORIGINS`（`https://FRONTEND_DOMAIN`）・`WASABI_BUCKET`（envs の `attachments_bucket`）・`GOOGLE_OAUTH_CLIENT_ID`・`FIREBASE_PROJECT_ID`・`PASSWORD_AUTH_ENABLED`

`kustomize build infra/k8s/overlays/dev` で展開結果を確認できる。

### 7. Google ログインと Firebase

- 「API とサービス → 認証情報」で OAuth クライアント ID（ウェブアプリ）を作り、承認済みの JavaScript 生成元に dev と prod の両方の `https://FRONTEND_DOMAIN` を登録する
- Firebase コンソールでこのプロジェクトに Firebase を追加し、ウェブアプリを登録して `apiKey` などと VAPID キーを控える。dev と prod は同じ Firebase プロジェクトを使う（通知先のトークンは環境ごとの DB に保存されるので混ざらない）。backend は Workload Identity の ADC で FCM に送る

### 8. GitHub の Environment

リポジトリの Settings → Environments で `dev` と `prod` を作る。`prod` には「Required reviewers」を付け、承認がないとデプロイが動かないようにする。それぞれに次の変数（Variables）を登録する。

| 変数 | 値 |
| --- | --- |
| `GCP_WORKLOAD_IDENTITY_PROVIDER` | shared の `workload_identity_provider` |
| `GCP_DEPLOY_SERVICE_ACCOUNT` | envs の `deployer_service_account_email` |
| `GCP_REGION` | `asia-northeast1` |
| `GKE_CLUSTER` | shared の `gke_cluster_name` |
| `GKE_LOCATION` | shared の `gke_cluster_location` |
| `ARTIFACT_REGISTRY` | shared の `artifact_registry_url` |
| `API_DOMAIN` | その環境の `api_domain` |
| `VITE_GOOGLE_OAUTH_CLIENT_ID` | OAuth クライアント ID |
| `VITE_FIREBASE_API_KEY` / `VITE_FIREBASE_PROJECT_ID` / `VITE_FIREBASE_MESSAGING_SENDER_ID` / `VITE_FIREBASE_APP_ID` / `VITE_FIREBASE_VAPID_KEY` | Firebase のウェブアプリ設定 |

### 9. 初回デプロイ

初回は両方の ref を指定する（クラスタに既存のイメージがないため）。

```sh
gh workflow run deploy.yml -R newt239/chat -f environment=dev -f backend_ref=main -f frontend_ref=main
```

`ENV=production` ではテスト用データを自動で作らない。検索インデックスを作り直すときは `kubectl -n chat-dev exec deploy/backend -c backend -- ./reindex` を実行する。

### 10. 動作確認

dev で次を確かめてから prod を作る。

- [ ] ログイン（Google とパスワード）
- [ ] 投稿が backend のレプリカをまたいで届く（2 つのブラウザで、`kubectl -n chat-dev logs` で別々の Pod につながっていることを見る）
- [ ] 閲覧者一覧、既読と未読数
- [ ] 添付ファイルのアップロードとダウンロード（Wasabi への署名付き URL）
- [ ] 検索（Meilisearch の Pod を消して PVC ごと作り直したあと、backend の再起動で検索できるようになる）
- [ ] 予約投稿が一度だけ送られる
- [ ] プッシュ通知
- [ ] Webhook（連続で送ると全レプリカ合わせて毎秒 1 回・瞬間 10 回で 429 になる）
- [ ] `kubectl -n chat-dev delete pod <backend の Pod>` で、クライアントがつなぎ直して配信が続く

### 11. 旧構成の片付け

Autopilot の構成を作っていた場合は、動作確認のあとで旧 Autopilot クラスタ、外部 HTTP(S) LB、静的 IP、managed 証明書、GCS の添付ファイル用バケット、HMAC キーとそのサービスアカウント、旧シークレット（接頭辞のない `DATABASE_URL` など）を削除する。旧 `envs/dev` の state は `envs/dev` の prefix に残っているので、新しい `envs/dev` を apply する前に `gsutil rm -r gs://PROJECT_ID-tfstate/envs/dev` で消すか、旧構成を `terraform destroy` しておく。

## デプロイ

GitHub Actions の「Deploy」（`.github/workflows/deploy.yml`）で行う。`environment` は `dev`・`prod`・`mini`（[mini 構成](#mini-構成k3s-の-vm-1-台)）から選ぶ。ref を空にした方はクラスタで動いているイメージをそのまま使う。

```sh
# dev の backend だけ feat/foo に差し替える
gh workflow run deploy.yml -R newt239/chat -f environment=dev -f backend_ref=feat/foo

# dev で動いている版を prod に出す（backend は同じイメージ、frontend は同じコミットを prod の URL でビルドし直す）
gh workflow run deploy.yml -R newt239/chat -f environment=prod -f promote_from_dev=true
```

## mini 構成（k3s の VM 1 台）

GKE を作る前に、同じマニフェストとアプリを安く確かめるための構成（月 $21 程度）。GCE の Spot VM（e2-medium）1 台に k3s を入れ、DB は Neon の無料枠を使う。shared には依存せず、`infra/terraform/envs/mini` だけで完結する（同じプロジェクトに shared を作っても名前が衝突しないよう、リソース名に `mini` を付けている）。

```mermaid
flowchart LR
  user[ブラウザ] -->|HTTPS / WSS| cf[Cloudflare<br/>DNS + Tunnel]
  user -->|署名付き URL| wasabi[(Wasabi 東京<br/>mini 用バケット)]

  subgraph vm[GCE VM e2-medium Spot / k3s]
    cfd[cloudflared] --> fe[frontend]
    cfd --> be[backend ×2]
    be --> redis[Redis]
    be --> meili[Meilisearch<br/>local-path PVC]
  end

  cf --- cfd
  be -->|TLS| neon[(Neon 無料枠<br/>AWS 東京)]
  be --> wasabi
  be -->|VM の SA で ADC| fcm[FCM]
```

| GKE との違い | mini |
| --- | --- |
| DB | Neon に TLS で直接つなぐ（Cloud SQL Auth Proxy なし） |
| シークレット | gitignore した `overlays/mini/*.env` から `secretGenerator` で作る（External Secrets なし） |
| GCP の権限 | Pod はメタデータサーバー経由で VM の SA を使う（Workload Identity なし） |
| イメージの取得 | CronJob が 30 分ごとに VM の SA のトークンで imagePullSecret を作り直す |
| デプロイ | k3s の API は公開せず、IAP 経由の SSH で VM に入って `sudo k3s kubectl` を実行する |
| 予約投稿 | Neon を止められるよう 15 分ごとにしか確かめないので、最大 15 分遅れる（「今すぐ送信」はすぐ送られる） |

mini で確かめられないのは Cloud SQL Auth Proxy・External Secrets・Workload Identity で、これらは GKE の構成で確かめる。

### 1. 準備

- 初回セットアップの手順 1（state バケット）と手順 2（Cloudflare・Wasabi）を済ませる。Wasabi のサブユーザーは mini のバケットだけを読み書きできるようにする
- Neon でプロジェクトを作る。リージョンは AWS 東京（`aws-ap-northeast-1`）、PostgreSQL のバージョンはローカルと同じ 18 にする。接続文字列は **Connection pooling をオフにした直接のエンドポイント**（ホスト名に `-pooler` が付かないほう）を使い、`sslmode=require` を付ける。`lib/pq` とトランザクションモードのプーラーを組み合わせたときの問題を避けるため

Neon の無料枠はストレージと稼働時間に上限がある。アクセスがないと 5 分ほどで停止し、最初のアクセスで数百ミリ秒〜数秒かけて起動する。mini の backend はアイドルの接続を 1 分で閉じる（`DB_CONN_MAX_IDLE_TIME`）。

### 2. Terraform

```sh
cd infra/terraform/envs/mini
cp terraform.tfvars.example terraform.tfvars   # ドメイン、バケット名、Cloudflare の ID を書く
export CLOUDFLARE_API_TOKEN=... AWS_ACCESS_KEY_ID=... AWS_SECRET_ACCESS_KEY=...
terraform init -backend-config="bucket=PROJECT_ID-tfstate"
terraform apply
terraform output
```

VM は起動スクリプトで k3s（Traefik なし）を入れる。数分後に次で確かめられる。

```sh
gcloud compute ssh chat-mini --zone asia-northeast1-b --tunnel-through-iap --command 'sudo k3s kubectl get nodes'
```

### 3. マニフェストとシークレット

- `infra/k8s/overlays/mini/kustomization.yaml` の `images` の `newName`（`artifact_registry_url` + `/backend`・`/frontend`）と `backend-env` の値を書き換えてコミットする
- `overlays/mini` の `*.env.example` をコピーして `backend-secrets.env` と `cloudflared.env` を作る（gitignore 済み）
  - `DATABASE_URL`: 手順 1 の Neon の接続文字列
  - `JWT_SECRET`・`MEILISEARCH_API_KEY`: `openssl rand -hex 32` などで作る
  - `WASABI_*`: 手順 1 のサブユーザーのキー
  - `TUNNEL_TOKEN`: `terraform output -raw tunnel_token`

### 4. GitHub の Environment

Settings → Environments で `mini` を作り、次を登録する。

| 種類 | 名前 | 値 |
| --- | --- | --- |
| 変数 | `GCP_WORKLOAD_IDENTITY_PROVIDER` | `workload_identity_provider` |
| 変数 | `GCP_DEPLOY_SERVICE_ACCOUNT` | `deployer_service_account_email` |
| 変数 | `GCP_REGION` | `asia-northeast1` |
| 変数 | `MINI_VM` / `MINI_ZONE` | `vm_name` / `vm_zone` |
| 変数 | `ARTIFACT_REGISTRY` | `artifact_registry_url` |
| 変数 | `API_DOMAIN`、`VITE_*` | 初回セットアップの手順 8 と同じ |
| シークレット | `BACKEND_SECRETS_ENV` | `backend-secrets.env` の中身 |
| シークレット | `CLOUDFLARED_ENV` | `cloudflared.env` の中身 |

Google ログインを使うなら、OAuth クライアントの承認済み JavaScript 生成元に mini の `https://FRONTEND_DOMAIN` を足す。

### 5. デプロイ

```sh
gh workflow run deploy.yml -R newt239/chat -f environment=mini -f backend_ref=main -f frontend_ref=main
```

初回は両方の ref を指定する。ワークフローは apply のあとに imagePullSecret を作る Job を動かすので、Pod が一時的に `ImagePullBackOff` になっても数分で起動する。シークレットを変えたときは Environment の Secrets を更新してデプロイし直す（`secretGenerator` が名前を変えるので Pod が作り直される）。

検索インデックスを作り直すときは次を実行する。

```sh
gcloud compute ssh chat-mini --zone asia-northeast1-b --tunnel-through-iap \
  --command 'sudo k3s kubectl -n chat-mini exec deploy/backend -- ./reindex'
```

### 6. 運用

- **Spot で回収されたら**、VM は停止する。`gcloud compute instances start chat-mini --zone asia-northeast1-b` で起動すると k3s と Pod も戻る
- **k3s の更新**は手で行う（起動スクリプトは k3s がないときだけ入れる）: VM に入って `curl -sfL https://get.k3s.io | INSTALL_K3S_VERSION=v1.xx.x+k3s1 sh -s - server --disable traefik`
- **kubectl を手元から使う**ときも `gcloud compute ssh ... --command 'sudo k3s kubectl ...'` を使う。ワークフローの `Connect to k3s` と同じ方法

### 7. 動作確認

- [ ] ログイン（パスワードと Google）
- [ ] 投稿が backend のレプリカをまたいで届く（2 つのブラウザで、別々の Pod につながっていることをログで見る）
- [ ] 閲覧者一覧、既読と未読数
- [ ] 添付ファイルのアップロードとダウンロード（Wasabi への署名付き URL）
- [ ] 検索（Meilisearch）と、Meilisearch の PVC を消したあとのインデックスの作り直し
- [ ] 予約投稿が一度だけ送られる（最大 15 分遅れる）
- [ ] プッシュ通知（VM の SA の ADC で FCM に送る）
- [ ] Webhook
- [ ] backend の Pod を 1 つ削除したときに、クライアントがつなぎ直して配信が続く
- [ ] Neon が停止した状態から最初にアクセスしたときの遅延が許容範囲か

## CI で terraform plan する（任意）

`terraform.yml` の `plan` ジョブはリポジトリ変数 `GCP_TERRAFORM_SERVICE_ACCOUNT` があるときだけ動き、`shared`・`envs/dev`・`envs/prod`・`envs/mini` を plan する。plan 用の GSA を作って閲覧権限と state バケット・シークレットの読み取り権限を与え、GitHub のリポジトリから借用できるようにする。

```sh
gcloud iam service-accounts create chat-terraform-plan
SA=chat-terraform-plan@PROJECT_ID.iam.gserviceaccount.com
for role in roles/viewer roles/iam.securityReviewer roles/secretmanager.secretAccessor; do
  gcloud projects add-iam-policy-binding PROJECT_ID --member="serviceAccount:$SA" --role="$role"
done
gcloud storage buckets add-iam-policy-binding gs://PROJECT_ID-tfstate --member="serviceAccount:$SA" --role=roles/storage.objectViewer
gcloud iam service-accounts add-iam-policy-binding "$SA" --role=roles/iam.workloadIdentityUser \
  --member="principalSet://iam.googleapis.com/$(terraform -chdir=infra/terraform/shared output -raw github_pool_name)/attribute.repository/newt239/chat"
```

リポジトリに次を登録する。

- 変数: `GCP_WORKLOAD_IDENTITY_PROVIDER`、`GCP_TERRAFORM_SERVICE_ACCOUNT`、`TF_STATE_BUCKET`、`TFVARS_SHARED`・`TFVARS_DEV`・`TFVARS_PROD`・`TFVARS_MINI`（各ディレクトリの `terraform.tfvars` の中身）
- シークレット: `CLOUDFLARE_API_TOKEN`、`WASABI_TERRAFORM_ACCESS_KEY_ID`、`WASABI_TERRAFORM_SECRET_ACCESS_KEY`（読み取りだけのキーでよい）。ないときは envs の plan を省く
