variable "project_id" {
  type = string
}

# dev / prod。namespace・シークレット ID・GitHub Environment の名前に使う
variable "environment" {
  type = string
}

variable "frontend_domain" {
  type = string
}

variable "api_domain" {
  type = string
}

# true なら terraform destroy でも DB とユーザーを消さない
variable "protect_database" {
  type = bool
}

variable "wasabi_bucket_name" {
  type = string
}

variable "cloudflare_account_id" {
  type = string
}

variable "cloudflare_zone_id" {
  type = string
}

# 以下は shared の出力
variable "cloudsql_instance_name" {
  type = string
}

variable "workload_identity_pool" {
  type = string
}

variable "github_pool_name" {
  type = string
}

variable "artifact_registry_location" {
  type = string
}

variable "artifact_registry_repository" {
  type = string
}
