variable "project_id" {
  type = string
}

variable "region" {
  type = string
}

variable "zone" {
  type = string
}

variable "name" {
  type = string
}

variable "machine_type" {
  type = string
}

variable "disk_size_gb" {
  type = number
}

# 例: v1.36.4+k3s1
variable "k3s_version" {
  type = string
}

variable "subnet_cidr" {
  type    = string
  default = "10.10.0.0/24"
}
