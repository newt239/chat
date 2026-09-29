locals {
  name                       = "chat-${var.environment}"
  kubernetes_namespace       = "chat-${var.environment}"
  kubernetes_service_account = "chat-backend"
}

module "database" {
  source = "../cloudsql_database"

  project_id      = var.project_id
  instance_name   = var.cloudsql_instance_name
  name            = "chat_${var.environment}"
  deletion_policy = var.protect_database ? "ABANDON" : "DELETE"
}

module "deployer" {
  source = "../deployer"

  project_id                   = var.project_id
  service_account_id           = "${local.name}-deployer"
  environment                  = var.environment
  github_pool_name             = var.github_pool_name
  artifact_registry_location   = var.artifact_registry_location
  artifact_registry_repository = var.artifact_registry_repository
}

module "attachments" {
  source = "../wasabi_bucket"

  bucket_name = var.wasabi_bucket_name
  # Tauri のアプリは macOS/iOS と Windows/Android で origin が異なる
  cors_origins = ["https://${var.frontend_domain}", "tauri://localhost", "https://tauri.localhost"]
}

module "tunnel" {
  source = "../cloudflare_tunnel"

  account_id           = var.cloudflare_account_id
  zone_id              = var.cloudflare_zone_id
  name                 = local.name
  kubernetes_namespace = local.kubernetes_namespace
  frontend_domain      = var.frontend_domain
  api_domain           = var.api_domain
}

resource "random_password" "jwt" {
  length  = 64
  special = false
}

resource "random_password" "meilisearch" {
  length  = 32
  special = false
}

module "backend_identity" {
  source = "../backend_identity"

  project_id                 = var.project_id
  environment                = var.environment
  service_account_id         = "${local.name}-backend"
  workload_identity_pool     = var.workload_identity_pool
  kubernetes_namespace       = local.kubernetes_namespace
  kubernetes_service_account = local.kubernetes_service_account

  secrets = {
    DATABASE_URL            = module.database.database_url
    JWT_SECRET              = random_password.jwt.result
    MEILISEARCH_API_KEY     = random_password.meilisearch.result
    CLOUDFLARE_TUNNEL_TOKEN = module.tunnel.token
  }

  # バケットに絞った Wasabi のサブユーザーで発行し、手で登録する
  manual_secrets = ["WASABI_ACCESS_KEY_ID", "WASABI_SECRET_ACCESS_KEY"]
}
