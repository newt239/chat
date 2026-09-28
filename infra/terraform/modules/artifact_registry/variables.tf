variable "project_id" {
  type = string
}

variable "region" {
  type = string
}

variable "name" {
  type = string
}

variable "keep_count" {
  type    = number
  default = 20
}

variable "reader_members" {
  type = map(string)
}

variable "writer_members" {
  type = map(string)
}
