resource "google_service_account" "this" {
  project      = var.project_id
  account_id   = var.service_account_id
  display_name = "GitHub Actions deployer (${var.environment})"
}

resource "google_project_iam_member" "this" {
  project = var.project_id
  role    = "roles/container.developer"
  member  = google_service_account.this.member
}

resource "google_artifact_registry_repository_iam_member" "this" {
  project    = var.project_id
  location   = var.artifact_registry_location
  repository = var.artifact_registry_repository
  role       = "roles/artifactregistry.writer"
  member     = google_service_account.this.member
}

# 指定した GitHub Environment のジョブだけがこのサービスアカウントになれる
resource "google_service_account_iam_member" "this" {
  service_account_id = google_service_account.this.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${var.github_pool_name}/attribute.environment/${var.environment}"
}
