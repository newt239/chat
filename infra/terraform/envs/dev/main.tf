locals {
  kubernetes_namespace       = "chat"
  kubernetes_service_account = "chat-backend"
  github_environment         = "dev"
}

resource "google_project_service" "this" {
  for_each = toset([
    "artifactregistry.googleapis.com",
    "compute.googleapis.com",
    "container.googleapis.com",
    "fcm.googleapis.com",
    "iamcredentials.googleapis.com",
    "secretmanager.googleapis.com",
    "servicenetworking.googleapis.com",
    "sqladmin.googleapis.com",
    "sts.googleapis.com",
  ])

  project            = var.project_id
  service            = each.value
  disable_on_destroy = false
}

module "network" {
  source = "../../modules/network"

  project_id = var.project_id
  region     = var.region
  name       = var.name

  depends_on = [google_project_service.this]
}

module "gke" {
  source = "../../modules/gke"

  project_id          = var.project_id
  region              = var.region
  name                = var.name
  network_id          = module.network.network_id
  subnetwork_id       = module.network.subnetwork_id
  pods_range_name     = module.network.pods_range_name
  services_range_name = module.network.services_range_name
  deletion_protection = var.deletion_protection
}

module "cloudsql" {
  source = "../../modules/cloudsql"

  project_id             = var.project_id
  region                 = var.region
  name                   = var.name
  network_id             = module.network.network_id
  tier                   = var.cloudsql_tier
  availability_type      = "ZONAL"
  point_in_time_recovery = false
  deletion_protection    = var.deletion_protection

  # Private IP はサービスネットワーキングのピアリング完了後にしか割り当てられない
  depends_on = [module.network]
}

module "ci_oidc" {
  source = "../../modules/ci_oidc"

  project_id         = var.project_id
  pool_id            = "github"
  service_account_id = "${var.name}-deployer"
  repository         = var.repository
  environment        = local.github_environment

  depends_on = [google_project_service.this]
}

module "artifact_registry" {
  source = "../../modules/artifact_registry"

  project_id     = var.project_id
  region         = var.region
  name           = "chat"
  reader_members = { gke_node = module.gke.node_service_account_member }
  writer_members = { deployer = module.ci_oidc.service_account_member }
}

module "storage" {
  source = "../../modules/storage"

  project_id         = var.project_id
  region             = var.region
  bucket_name        = "${var.project_id}-attachments"
  service_account_id = "${var.name}-storage"
  cors_origins       = ["https://${var.frontend_domain}"]
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
  source = "../../modules/backend_identity"

  project_id                 = var.project_id
  service_account_id         = "${var.name}-backend"
  workload_identity_pool     = module.gke.workload_identity_pool
  kubernetes_namespace       = local.kubernetes_namespace
  kubernetes_service_account = local.kubernetes_service_account

  secrets = {
    DATABASE_URL             = module.cloudsql.database_url
    JWT_SECRET               = random_password.jwt.result
    MEILISEARCH_API_KEY      = random_password.meilisearch.result
    WASABI_ACCESS_KEY_ID     = module.storage.hmac_access_id
    WASABI_SECRET_ACCESS_KEY = module.storage.hmac_secret
  }
}

module "ingress" {
  source = "../../modules/ingress"

  project_id = var.project_id
  name       = var.name
  domains    = [var.frontend_domain, var.api_domain]

  depends_on = [google_project_service.this]
}
