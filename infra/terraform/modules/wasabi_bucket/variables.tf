# Wasabi のバケット名は全アカウントで一意にする
variable "bucket_name" {
  type = string
}

variable "cors_origins" {
  type = list(string)
}
