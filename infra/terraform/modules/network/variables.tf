variable "project_id" {
  type = string
}

variable "region" {
  type = string
}

variable "name" {
  type = string
}

variable "subnet_cidr" {
  type    = string
  default = "10.0.0.0/20"
}

variable "pods_cidr" {
  type    = string
  default = "10.16.0.0/14"
}

variable "services_cidr" {
  type    = string
  default = "10.20.0.0/20"
}
