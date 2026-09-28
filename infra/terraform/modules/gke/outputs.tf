output "name" {
  value = google_container_cluster.this.name
}

output "location" {
  value = google_container_cluster.this.location
}

output "node_service_account_member" {
  value = google_service_account.node.member
}

output "workload_identity_pool" {
  value = google_container_cluster.this.workload_identity_config[0].workload_pool
}
