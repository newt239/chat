variable "account_id" {
  type = string
}

variable "zone_id" {
  type = string
}

variable "name" {
  type = string
}

# 振り分け先の Service がある namespace
variable "kubernetes_namespace" {
  type = string
}

variable "frontend_domain" {
  type = string
}

variable "api_domain" {
  type = string
}
