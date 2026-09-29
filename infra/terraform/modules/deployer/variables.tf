variable "project_id" {
  type = string
}

variable "service_account_id" {
  type = string
}

# GitHub Environment の名前
variable "environment" {
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

# デプロイ先を操作するための権限。GKE なら container.developer
variable "project_roles" {
  type    = list(string)
  default = ["roles/container.developer"]
}
