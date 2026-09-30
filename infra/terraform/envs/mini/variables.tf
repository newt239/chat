variable "project_id" {
  type = string
}

variable "region" {
  type    = string
  default = "asia-northeast1"
}

variable "zone" {
  type    = string
  default = "asia-northeast1-b"
}

# owner/name 形式
variable "repository" {
  type    = string
  default = "newt239/chat"
}

variable "machine_type" {
  type    = string
  default = "e2-medium"
}

# VM を作るときだけ使う。入れたあとの更新は手で行う
variable "k3s_version" {
  type    = string
  default = "v1.36.4+k3s1"
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
