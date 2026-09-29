output "kubernetes_namespace" {
  value = module.environment.kubernetes_namespace
}

output "backend_service_account_email" {
  value = module.environment.backend_service_account_email
}

output "deployer_service_account_email" {
  value = module.environment.deployer_service_account_email
}

output "attachments_bucket" {
  value = module.environment.attachments_bucket
}

output "tunnel_id" {
  value = module.environment.tunnel_id
}
