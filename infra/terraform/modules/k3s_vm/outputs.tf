output "instance_name" {
  value = google_compute_instance.this.name
}

output "zone" {
  value = google_compute_instance.this.zone
}

output "service_account_email" {
  value = google_service_account.this.email
}

output "service_account_member" {
  value = google_service_account.this.member
}

output "service_account_name" {
  value = google_service_account.this.name
}
