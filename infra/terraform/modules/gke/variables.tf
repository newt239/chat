variable "project_id" {
  type = string
}

variable "zone" {
  type = string
}

variable "name" {
  type = string
}

variable "network_id" {
  type = string
}

variable "subnetwork_id" {
  type = string
}

variable "pods_range_name" {
  type = string
}

variable "services_range_name" {
  type = string
}

variable "deletion_protection" {
  type = bool
}

variable "machine_type" {
  type    = string
  default = "e2-medium"
}

variable "min_node_count" {
  type    = number
  default = 2
}

variable "max_node_count" {
  type    = number
  default = 4
}
