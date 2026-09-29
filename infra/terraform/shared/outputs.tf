output "gke_cluster_name" {
  value = module.gke.name
}

output "gke_cluster_location" {
  value = module.gke.location
}

output "workload_identity_pool" {
  value = module.gke.workload_identity_pool
}

output "cloudsql_instance_name" {
  value = module.cloudsql.instance_name
}

output "cloudsql_connection_name" {
  value = module.cloudsql.connection_name
}

output "artifact_registry_url" {
  value = module.artifact_registry.url
}

output "artifact_registry_location" {
  value = module.artifact_registry.location
}

output "artifact_registry_repository" {
  value = module.artifact_registry.repository
}

output "github_pool_name" {
  value = module.ci_oidc.pool_name
}

output "workload_identity_provider" {
  value = module.ci_oidc.provider_name
}
