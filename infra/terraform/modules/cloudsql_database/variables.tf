variable "project_id" {
  type = string
}

variable "instance_name" {
  type = string
}

# DB 名とユーザー名に使う
variable "name" {
  type = string
}

# ABANDON にすると terraform destroy でも DB とユーザーを消さない
variable "deletion_policy" {
  type    = string
  default = "DELETE"
}
