# GKE を作る前に同じマニフェストを確かめる mini 構成。shared には依存せず、これだけで完結する
# 名前に mini を付け、同じプロジェクトに shared を作っても衝突しないようにする
locals {
  name = "chat-mini"
}

resource "google_project_service" "this" {
  for_each = toset([
    "artifactregistry.googleapis.com",
    "compute.googleapis.com",
    "fcm.googleapis.com",
    "iamcredentials.googleapis.com",
    "iap.googleapis.com",
    "oslogin.googleapis.com",
    "sts.googleapis.com",
  ])

  project            = var.project_id
  service            = each.value
  disable_on_destroy = false
}

module "vm" {
  source = "../../modules/k3s_vm"

  project_id   = var.project_id
  region       = var.region
  zone         = var.zone
  name         = local.name
  machine_type = var.machine_type
  disk_size_gb = 20
  k3s_version  = var.k3s_version

  depends_on = [google_project_service.this]
}

module "artifact_registry" {
  source = "../../modules/artifact_registry"

  project_id     = var.project_id
  region         = var.region
  name           = local.name
  keep_count     = 10
  reader_members = { vm = module.vm.service_account_member }
  writer_members = {}

  depends_on = [google_project_service.this]
}

module "ci_oidc" {
  source = "../../modules/ci_oidc"

  project_id = var.project_id
  pool_id    = "github-mini"
  repository = var.repository

  depends_on = [google_project_service.this]
}

# GitHub Environment mini のジョブが、IAP 経由の SSH で VM に入って k3s に apply する
module "deployer" {
  source = "../../modules/deployer"

  project_id                   = var.project_id
  service_account_id           = "${local.name}-deployer"
  environment                  = "mini"
  github_pool_name             = module.ci_oidc.pool_name
  artifact_registry_location   = module.artifact_registry.location
  artifact_registry_repository = module.artifact_registry.repository
  project_roles = [
    "roles/compute.osAdminLogin",
    "roles/compute.viewer",
    "roles/iap.tunnelResourceAccessor",
  ]
}

# SA の付いた VM に OS Login で入るのに要る
resource "google_service_account_iam_member" "deployer_vm_user" {
  service_account_id = module.vm.service_account_name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:${module.deployer.service_account_email}"
}

module "attachments" {
  source = "../../modules/wasabi_bucket"

  bucket_name  = var.wasabi_bucket_name
  cors_origins = ["https://${var.frontend_domain}"]
}

module "tunnel" {
  source = "../../modules/cloudflare_tunnel"

  account_id           = var.cloudflare_account_id
  zone_id              = var.cloudflare_zone_id
  name                 = local.name
  kubernetes_namespace = local.name
  frontend_domain      = var.frontend_domain
  api_domain           = var.api_domain
}
