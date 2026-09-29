variable "project_id" {
  type = string
}

variable "region" {
  type    = string
  default = "asia-northeast1"
}

# shared の state を読むため、-backend-config と同じバケットを渡す
variable "state_bucket" {
  type = string
}

variable "frontend_domain" {
  type = string
}

variable "api_domain" {
  type = string
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
