output "kubernetes_namespace" {
  value = local.kubernetes_namespace
}

output "backend_service_account_email" {
  value = module.backend_identity.service_account_email
}

output "deployer_service_account_email" {
  value = module.deployer.service_account_email
}

output "attachments_bucket" {
  value = module.attachments.bucket_name
}

output "tunnel_id" {
  value = module.tunnel.tunnel_id
}
