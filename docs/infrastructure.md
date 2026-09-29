# dev 環境の初回セットアップ（Google Cloud）

未実施のデプロイ手順。Terraform は `infra/terraform/`、Kubernetes のマニフェストは `infra/k8s/` にある。すべて終えたらこのファイルは削除してよい。

前提: `gcloud`・`terraform`（1.11 以上）・`kubectl`・`helm`・`gh` が使えること。以下の `PROJECT_ID` などは適宜置き換える。

## 1. プロジェクトと state バケット

```sh
gcloud auth login
gcloud auth application-default login
gcloud config set project PROJECT_ID
gcloud services enable storage.googleapis.com cloudresourcemanager.googleapis.com serviceusage.googleapis.com

gcloud storage buckets create gs://PROJECT_ID-tfstate --location=asia-northeast1 --uniform-bucket-level-access
gcloud storage buckets update gs://PROJECT_ID-tfstate --versioning
```

Firebase（FCM）を使う場合は Firebase コンソールでこのプロジェクトに Firebase を追加し、ウェブアプリを登録して `apiKey` などと VAPID キーを控えておく。Google ログインを使う場合は「API とサービス → 認証情報」で OAuth クライアント ID（ウェブアプリ、承認済みの JavaScript 生成元に `https://FRONTEND_DOMAIN`）を作る。

## 2. Terraform

```sh
cd infra/terraform/envs/dev
cp terraform.tfvars.example terraform.tfvars   # project_id とドメインを書く
terraform init -backend-config="bucket=PROJECT_ID-tfstate"
terraform apply
terraform output
```

## 3. DNS

`terraform output ingress_ip_address` の IP を、`frontend_domain` と `api_domain` の A レコードに設定する。managed 証明書は DNS が向いて Ingress に紐づいてから発行される（数十分かかる）。

## 4. External Secrets Operator

```sh
gcloud container clusters get-credentials chat-dev --location=asia-northeast1
helm repo add external-secrets https://charts.external-secrets.io
helm install external-secrets external-secrets/external-secrets \
  --namespace external-secrets --create-namespace
```

## 5. マニフェストに値を反映

`infra/k8s/overlays/dev/kustomization.yaml` を `terraform output` の値で書き換えてコミットする。

- `images` の `newName`: `artifact_registry_url` + `/backend`・`/frontend`
- `backend-env`: `CORS_ALLOWED_ORIGINS`（`https://FRONTEND_DOMAIN`）、`WASABI_BUCKET`（`attachments_bucket`）、`GOOGLE_OAUTH_CLIENT_ID`、`FIREBASE_PROJECT_ID`、`PASSWORD_AUTH_ENABLED`
- `dev-params`: `PROJECT_ID`、`CLUSTER_NAME`・`CLUSTER_LOCATION`（`gke_cluster_*`）、`BACKEND_SERVICE_ACCOUNT`（`backend_service_account_email`）、`CLOUDSQL_CONNECTION_NAME`（`cloudsql_connection_name`）、`INGRESS_IP_NAME`（`ingress_ip_name`）、`CERTIFICATE_NAME`（`certificate_name`）、`FRONTEND_DOMAIN`、`API_DOMAIN`

`kustomize build infra/k8s/overlays/dev` で展開結果を確認できる。

## 6. GitHub の Environment

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

## 7. 初回デプロイ

初回は両方の ref を指定する（クラスタに既存のイメージがないため）。

```sh
gh workflow run deploy-dev.yml -R newt239/chat -f backend_ref=main -f frontend_ref=main
```

`ENV=production` では自動シードしないため、テスト用データが必要なら `kubectl -n chat exec deploy/backend -c backend -- ./seed` を実行する。

## （任意）CI で terraform plan する

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
