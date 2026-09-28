variable "project_id" {
  type = string
}

variable "region" {
  type    = string
  default = "asia-northeast1"
}

variable "name" {
  type    = string
  default = "chat-dev"
}

variable "frontend_domain" {
  type = string
}

variable "api_domain" {
  type = string
}

# owner/name 形式
variable "repository" {
  type    = string
  default = "newt239/chat"
}

variable "deletion_protection" {
  type    = bool
  default = false
}

variable "cloudsql_tier" {
  type    = string
  default = "db-custom-1-3840"
}
