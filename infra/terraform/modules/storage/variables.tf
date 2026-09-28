variable "project_id" {
  type = string
}

variable "region" {
  type = string
}

variable "bucket_name" {
  type = string
}

variable "service_account_id" {
  type = string
}

variable "cors_origins" {
  type = list(string)
}
