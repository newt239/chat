output "pool_name" {
  value = google_iam_workload_identity_pool.github.name
}

output "provider_name" {
  value = google_iam_workload_identity_pool_provider.github.name
}
