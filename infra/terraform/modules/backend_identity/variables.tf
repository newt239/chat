variable "project_id" {
  type = string
}

variable "service_account_id" {
  type = string
}

variable "workload_identity_pool" {
  type = string
}

variable "kubernetes_namespace" {
  type = string
}

variable "kubernetes_service_account" {
  type = string
}

# 環境変数名 => 値。External Secrets が k8s の Secret に同期する
variable "secrets" {
  type      = map(string)
  sensitive = true
}
