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

variable "name" {
  type    = string
  default = "chat"
}

# owner/name 形式
variable "repository" {
  type    = string
  default = "newt239/chat"
}

# prod のデータも載るため既定で有効にする
variable "deletion_protection" {
  type    = bool
  default = true
}

variable "cloudsql_tier" {
  type    = string
  default = "db-g1-small"
}

variable "node_machine_type" {
  type    = string
  default = "e2-medium"
}

variable "node_min_count" {
  type    = number
  default = 2
}

variable "node_max_count" {
  type    = number
  default = 4
}
