output "bucket_name" {
  value = google_storage_bucket.attachments.name
}

output "hmac_access_id" {
  value     = google_storage_hmac_key.this.access_id
  sensitive = true
}

output "hmac_secret" {
  value     = google_storage_hmac_key.this.secret
  sensitive = true
}
