# dev と prod で共有する部分。環境ごとの部分は envs/<環境> で作る
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
  source = "../modules/network"

  project_id = var.project_id
  region     = var.region
  name       = var.name

  depends_on = [google_project_service.this]
}

module "gke" {
  source = "../modules/gke"

  project_id          = var.project_id
  zone                = var.zone
  name                = var.name
  network_id          = module.network.network_id
  subnetwork_id       = module.network.subnetwork_id
  pods_range_name     = module.network.pods_range_name
  services_range_name = module.network.services_range_name
  deletion_protection = var.deletion_protection
  machine_type        = var.node_machine_type
  min_node_count      = var.node_min_count
  max_node_count      = var.node_max_count
}

# PITR はインスタンス単位なので dev の DB にも効く
module "cloudsql" {
  source = "../modules/cloudsql"

  project_id             = var.project_id
  region                 = var.region
  name                   = var.name
  network_id             = module.network.network_id
  tier                   = var.cloudsql_tier
  availability_type      = "ZONAL"
  point_in_time_recovery = true
  deletion_protection    = var.deletion_protection

  # Private IP はサービスネットワーキングのピアリング完了後にしか割り当てられない
  depends_on = [module.network]
}

module "ci_oidc" {
  source = "../modules/ci_oidc"

  project_id = var.project_id
  pool_id    = "github"
  repository = var.repository

  depends_on = [google_project_service.this]
}

# 書き込みは環境ごとの deployer に envs で付ける
module "artifact_registry" {
  source = "../modules/artifact_registry"

  project_id     = var.project_id
  region         = var.region
  name           = var.name
  reader_members = { gke_node = module.gke.node_service_account_member }
  writer_members = {}
}
