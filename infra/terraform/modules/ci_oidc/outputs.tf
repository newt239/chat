output "provider_name" {
  value = google_iam_workload_identity_pool_provider.github.name
}

output "service_account_email" {
  value = google_service_account.deployer.email
}

output "service_account_member" {
  value = google_service_account.deployer.member
}
