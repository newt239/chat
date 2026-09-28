output "ip_name" {
  value = google_compute_global_address.this.name
}

output "ip_address" {
  value = google_compute_global_address.this.address
}

output "certificate_name" {
  value = google_compute_managed_ssl_certificate.this.name
}
