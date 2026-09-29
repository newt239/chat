data "terraform_remote_state" "shared" {
  backend = "gcs"

  config = {
    bucket = var.state_bucket
    prefix = "shared"
  }
}

module "environment" {
  source = "../../modules/environment"

  project_id            = var.project_id
  environment           = "dev"
  frontend_domain       = var.frontend_domain
  api_domain            = var.api_domain
  protect_database      = false
  wasabi_bucket_name    = var.wasabi_bucket_name
  cloudflare_account_id = var.cloudflare_account_id
  cloudflare_zone_id    = var.cloudflare_zone_id

  cloudsql_instance_name       = data.terraform_remote_state.shared.outputs.cloudsql_instance_name
  workload_identity_pool       = data.terraform_remote_state.shared.outputs.workload_identity_pool
  github_pool_name             = data.terraform_remote_state.shared.outputs.github_pool_name
  artifact_registry_location   = data.terraform_remote_state.shared.outputs.artifact_registry_location
  artifact_registry_repository = data.terraform_remote_state.shared.outputs.artifact_registry_repository
}
