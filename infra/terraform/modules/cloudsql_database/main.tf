resource "google_sql_database" "this" {
  project         = var.project_id
  instance        = var.instance_name
  name            = var.name
  deletion_policy = var.deletion_policy
}

resource "random_password" "this" {
  length  = 32
  special = false
}

resource "google_sql_user" "this" {
  project         = var.project_id
  instance        = var.instance_name
  name            = var.name
  password        = random_password.this.result
  deletion_policy = var.deletion_policy == "ABANDON" ? "ABANDON" : null
}
