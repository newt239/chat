# GitHub Environment mini の vars に入れる値
output "vm_name" {
  value = module.vm.instance_name
}

output "vm_zone" {
  value = module.vm.zone
}

output "artifact_registry_url" {
  value = module.artifact_registry.url
}

output "workload_identity_provider" {
  value = module.ci_oidc.provider_name
}

output "deployer_service_account_email" {
  value = module.deployer.service_account_email
}

output "attachments_bucket" {
  value = module.attachments.bucket_name
}

# overlays/mini/cloudflared.env の TUNNEL_TOKEN に入れる
output "tunnel_token" {
  value     = module.tunnel.token
  sensitive = true
}
