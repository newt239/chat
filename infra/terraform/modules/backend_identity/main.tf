resource "google_service_account" "backend" {
  project      = var.project_id
  account_id   = var.service_account_id
  display_name = "chat backend"
}

# k8s の ServiceAccount がこの GSA として振る舞う（Cloud SQL Auth Proxy / FCM / External Secrets）
resource "google_service_account_iam_member" "workload_identity" {
  service_account_id = google_service_account.backend.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "serviceAccount:${var.workload_identity_pool}[${var.kubernetes_namespace}/${var.kubernetes_service_account}]"
}

resource "google_project_iam_member" "backend" {
  for_each = toset(["roles/cloudsql.client", "roles/firebasecloudmessaging.admin"])

  project = var.project_id
  role    = each.value
  member  = google_service_account.backend.member
}

# シークレット ID は環境変数名と同じにする（環境ごとにプロジェクトを分ける前提）
resource "google_secret_manager_secret" "this" {
  for_each = toset(nonsensitive(keys(var.secrets)))

  project   = var.project_id
  secret_id = each.value

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "this" {
  for_each = google_secret_manager_secret.this

  secret      = each.value.id
  secret_data = var.secrets[each.key]
}

resource "google_secret_manager_secret_iam_member" "backend" {
  for_each = google_secret_manager_secret.this

  project   = var.project_id
  secret_id = each.value.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = google_service_account.backend.member
}
