# インフラストラクチャ（Google Cloud）

Google Cloud 上の dev 環境の構成と、構築・デプロイの手順をまとめる。Terraform は `infra/terraform/`、Kubernetes のマニフェストは `infra/k8s/` にある。

## 構成

```mermaid
flowchart LR
  user[ブラウザ] -->|HTTPS| lb[外部 HTTP(S) LB<br/>静的 IP + managed 証明書]
  lb -->|FRONTEND_DOMAIN| fe[frontend<br/>nginx]
  lb -->|API_DOMAIN<br/>Connect RPC / WebSocket| be[backend]
  user -->|署名付き URL| gcs[(GCS<br/>添付ファイル)]

  subgraph gke[GKE Autopilot / namespace chat]
    fe
    subgraph pod[backend Pod]
      be --> proxy[Cloud SQL Auth Proxy<br/>sidecar]
    end
    be --> meili[Meilisearch<br/>StatefulSet + PVC]
    eso[External Secrets] -.同期.-> secret[Secret backend-secrets]
  end

  proxy -->|Private IP| sql[(Cloud SQL<br/>PostgreSQL 18)]
  be -->|S3 互換 HMAC| gcs
  be -->|ADC| fcm[FCM]
  sm[Secret Manager] -.-> eso
  gha[GitHub Actions] -->|Workload Identity Federation| ar[Artifact Registry]
  gha -->|kubectl apply| gke
```

### 設計上の判断

- **ホストで振り分ける**: Connect RPC のパス（`/chat.v1.XxxService/Method`）はパス要素単位の前方一致で切り出せないため、frontend と backend を別ホスト（例: `chat-dev.example.com` / `api.chat-dev.example.com`）にし、1 つの GCE Ingress でホストごとに振り分ける。証明書は両ドメインを含む Google managed 証明書を Terraform で作り、`pre-shared-cert` で Ingress に渡す。
- **backend は 1 レプリカ**: WebSocket のハブがプロセス内メモリにあるため `replicas: 1` かつ `strategy: Recreate` にしている。水平スケールするにはハブを Redis などに外出しする必要がある。
- **WebSocket のタイムアウト**: GCE LB の既定のバックエンドタイムアウト（30 秒）では接続が切れるため、`BackendConfig` で 3600 秒にしている。
- **DB 接続**: Cloud SQL は Private IP のみ。backend Pod のネイティブサイドカー（`initContainers` + `restartPolicy: Always`）で Cloud SQL Auth Proxy を動かし、アプリは `127.0.0.1:5432` に平文で繋ぐ。スキーマは backend の起動時に自動マイグレーションされる。
- **添付ファイル**: GCS の S3 互換 API を HMAC キーで使い、既存の `wasabi` ドライバをそのまま使う（`WASABI_ENDPOINT=https://storage.googleapis.com`）。ブラウザから署名付き URL で直接 PUT するため、バケットに frontend のオリジンを CORS で許可している。
- **シークレット**: Terraform が生成した値（DB の接続文字列、JWT、Meilisearch のキー、HMAC キー）を Secret Manager に入れ、External Secrets Operator が `backend-secrets` に同期する。Secret Manager のシークレット ID は環境変数名と同じにしている（**環境ごとに GCP プロジェクトを分ける前提**）。値は Terraform の state にも残るため、state バケットの権限は絞ること。
- **Workload Identity**: k8s の ServiceAccount `chat/chat-backend` が GSA `<name>-backend` として振る舞い、Cloud SQL / Secret Manager / FCM（ADC）を使う。GitHub Actions は Workload Identity Federation で GSA `<name>-deployer` になり、`dev` Environment のジョブだけが使える。
- **環境の追加**: `infra/terraform/envs/prod/` と `infra/k8s/overlays/prod/` を dev からコピーし、変数と `dev-params` を差し替える。

## 環境変数

backend（`infra/k8s/base/kustomization.yaml` と `overlays/dev/kustomization.yaml` の `backend-env`、および Secret `backend-secrets`）

| 変数 | 由来 |
| --- | --- |
| `ENV` / `PORT` / `STORAGE_DRIVER` / `WASABI_ENDPOINT` / `WASABI_REGION` / `MEILISEARCH_URL` | base の ConfigMap |
| `CORS_ALLOWED_ORIGINS` / `WASABI_BUCKET` / `GOOGLE_OAUTH_CLIENT_ID` / `FIREBASE_PROJECT_ID` / `PASSWORD_AUTH_ENABLED` | overlay の ConfigMap |
| `DATABASE_URL` / `JWT_SECRET` / `MEILISEARCH_API_KEY` / `WASABI_ACCESS_KEY_ID` / `WASABI_SECRET_ACCESS_KEY` | Secret Manager（External Secrets） |

frontend はビルド時に埋め込む。`deploy-dev` ワークフローが GitHub の `dev` Environment の変数からビルド引数を渡す。

| 変数 | 値 |
| --- | --- |
| `VITE_API_BASE_URL` / `VITE_WS_URL` | `https://${API_DOMAIN}` / `wss://${API_DOMAIN}` |
| `VITE_GOOGLE_OAUTH_CLIENT_ID` / `VITE_FIREBASE_*` | Environment 変数をそのまま渡す |

## 初回セットアップ

前提: `gcloud`・`terraform`（1.11 以上）・`kubectl`・`helm`・`gh` が使えること。以下の `PROJECT_ID` などは適宜置き換える。

### 1. プロジェクトと state バケット

```sh
gcloud auth login
gcloud auth application-default login
gcloud config set project PROJECT_ID
gcloud services enable storage.googleapis.com cloudresourcemanager.googleapis.com serviceusage.googleapis.com

gcloud storage buckets create gs://PROJECT_ID-tfstate --location=asia-northeast1 --uniform-bucket-level-access
gcloud storage buckets update gs://PROJECT_ID-tfstate --versioning
```

Firebase（FCM）を使う場合は Firebase コンソールでこのプロジェクトに Firebase を追加し、ウェブアプリを登録して `apiKey` などと VAPID キーを控えておく。Google ログインを使う場合は「API とサービス → 認証情報」で OAuth クライアント ID（ウェブアプリ、承認済みの JavaScript 生成元に `https://FRONTEND_DOMAIN`）を作る。

### 2. Terraform

```sh
cd infra/terraform/envs/dev
cp terraform.tfvars.example terraform.tfvars   # project_id とドメインを書く
terraform init -backend-config="bucket=PROJECT_ID-tfstate"
terraform apply
terraform output
```

### 3. DNS

`terraform output ingress_ip_address` の IP を、`frontend_domain` と `api_domain` の A レコードに設定する。managed 証明書は DNS が向いて Ingress に紐づいてから発行される（数十分かかる）。

### 4. External Secrets Operator

```sh
gcloud container clusters get-credentials chat-dev --location=asia-northeast1
helm repo add external-secrets https://charts.external-secrets.io
helm install external-secrets external-secrets/external-secrets \
  --namespace external-secrets --create-namespace
```

### 5. マニフェストに値を反映

`infra/k8s/overlays/dev/kustomization.yaml` を `terraform output` の値で書き換えてコミットする。

- `images` の `newName`: `artifact_registry_url` + `/backend`・`/frontend`
- `backend-env`: `CORS_ALLOWED_ORIGINS`（`https://FRONTEND_DOMAIN`）、`WASABI_BUCKET`（`attachments_bucket`）、`GOOGLE_OAUTH_CLIENT_ID`、`FIREBASE_PROJECT_ID`、`PASSWORD_AUTH_ENABLED`
- `dev-params`: `PROJECT_ID`、`CLUSTER_NAME`・`CLUSTER_LOCATION`（`gke_cluster_*`）、`BACKEND_SERVICE_ACCOUNT`（`backend_service_account_email`）、`CLOUDSQL_CONNECTION_NAME`（`cloudsql_connection_name`）、`INGRESS_IP_NAME`（`ingress_ip_name`）、`CERTIFICATE_NAME`（`certificate_name`）、`FRONTEND_DOMAIN`、`API_DOMAIN`

`kustomize build infra/k8s/overlays/dev` で展開結果を確認できる。

### 6. GitHub の Environment

リポジトリの Settings → Environments で `dev` を作り、次の変数（Variables）を登録する。

| 変数 | 値 |
| --- | --- |
| `GCP_WORKLOAD_IDENTITY_PROVIDER` | `workload_identity_provider` |
| `GCP_DEPLOY_SERVICE_ACCOUNT` | `deployer_service_account_email` |
| `GCP_REGION` | `asia-northeast1` |
| `GKE_CLUSTER` | `gke_cluster_name` |
| `ARTIFACT_REGISTRY` | `artifact_registry_url` |
| `API_DOMAIN` | `api_domain` |
| `VITE_GOOGLE_OAUTH_CLIENT_ID` | OAuth クライアント ID |
| `VITE_FIREBASE_API_KEY` / `VITE_FIREBASE_PROJECT_ID` / `VITE_FIREBASE_MESSAGING_SENDER_ID` / `VITE_FIREBASE_APP_ID` / `VITE_FIREBASE_VAPID_KEY` | Firebase のウェブアプリ設定 |

### 7. 初回デプロイ

初回は両方の ref を指定する（クラスタに既存のイメージがないため）。

```sh
gh workflow run deploy-dev.yml -R newt239/chat -f backend_ref=main -f frontend_ref=main
```

`ENV=production` では自動シードしないため、テスト用データが必要なら `kubectl -n chat exec deploy/backend -c backend -- ./seed` を実行する。

### （任意）CI で terraform plan する

`terraform.yml` の `plan` ジョブはリポジトリ変数 `GCP_TERRAFORM_SERVICE_ACCOUNT` があるときだけ動く。plan 用の GSA を作って閲覧権限と state バケット・シークレットの読み取り権限を与え、GitHub のリポジトリから借用できるようにしたうえで、リポジトリ変数 `GCP_WORKLOAD_IDENTITY_PROVIDER`・`GCP_TERRAFORM_SERVICE_ACCOUNT`・`TF_STATE_BUCKET`・`GCP_PROJECT_ID`・`FRONTEND_DOMAIN`・`API_DOMAIN` を登録する。

```sh
gcloud iam service-accounts create chat-dev-terraform-plan
SA=chat-dev-terraform-plan@PROJECT_ID.iam.gserviceaccount.com
for role in roles/viewer roles/iam.securityReviewer roles/secretmanager.secretAccessor; do
  gcloud projects add-iam-policy-binding PROJECT_ID --member="serviceAccount:$SA" --role="$role"
done
gcloud storage buckets add-iam-policy-binding gs://PROJECT_ID-tfstate --member="serviceAccount:$SA" --role=roles/storage.objectViewer
gcloud iam service-accounts add-iam-policy-binding "$SA" --role=roles/iam.workloadIdentityUser \
  --member="principalSet://iam.googleapis.com/$(terraform output -raw workload_identity_provider | sed 's|/providers/.*||')/attribute.repository/newt239/chat"
```

## dev へのデプロイ

Actions の「Deploy dev」を `workflow_dispatch` で実行する。

- `backend_ref` / `frontend_ref` にブランチ・タグ・SHA を指定したものだけビルドし、Artifact Registry に `<commit SHA 12 桁>` のタグで push する
- 空にしたほうはクラスタで動いているイメージをそのまま使う
- マニフェストはワークフローを起動したブランチ（通常は `main`）の `infra/k8s/overlays/dev` を使う

```sh
# backend だけ feat/foo に差し替える
gh workflow run deploy-dev.yml -R newt239/chat -f backend_ref=feat/foo
```

frontend の Docker イメージは指定した ref の `frontend/Dockerfile` でビルドするため、ビルド引数（`ARG VITE_*`）に対応していない古いブランチでは環境ごとの値が埋め込まれない。
