output "gke_cluster_name" {
  value = module.gke.name
}

output "gke_cluster_location" {
  value = module.gke.location
}

output "artifact_registry_url" {
  value = module.artifact_registry.url
}

output "cloudsql_connection_name" {
  value = module.cloudsql.connection_name
}

output "attachments_bucket" {
  value = module.storage.bucket_name
}

output "backend_service_account_email" {
  value = module.backend_identity.service_account_email
}

output "ingress_ip_name" {
  value = module.ingress.ip_name
}

output "ingress_ip_address" {
  value = module.ingress.ip_address
}

output "certificate_name" {
  value = module.ingress.certificate_name
}

output "workload_identity_provider" {
  value = module.ci_oidc.provider_name
}

output "deployer_service_account_email" {
  value = module.ci_oidc.service_account_email
}
