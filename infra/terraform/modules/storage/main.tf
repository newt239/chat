resource "google_storage_bucket" "attachments" {
  project                     = var.project_id
  name                        = var.bucket_name
  location                    = var.region
  uniform_bucket_level_access = true
  public_access_prevention    = "enforced"

  # 署名付き URL でブラウザから直接 PUT / GET する
  cors {
    origin          = var.cors_origins
    method          = ["GET", "HEAD", "PUT"]
    response_header = ["Content-Type", "Content-Length", "ETag"]
    max_age_seconds = 3600
  }
}

# 既存の wasabi ドライバ（S3 互換 API）から使うため HMAC キーを発行する
resource "google_service_account" "hmac" {
  project      = var.project_id
  account_id   = var.service_account_id
  display_name = "Attachments storage (S3 compatible)"
}

resource "google_storage_bucket_iam_member" "hmac" {
  bucket = google_storage_bucket.attachments.name
  role   = "roles/storage.objectUser"
  member = google_service_account.hmac.member
}

resource "google_storage_hmac_key" "this" {
  project               = var.project_id
  service_account_email = google_service_account.hmac.email
}
