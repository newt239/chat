# Pod 内の Cloud SQL Auth Proxy が TLS を張るため、アプリからの接続は平文でよい
output "database_url" {
  value     = "postgresql://${google_sql_user.this.name}:${random_password.this.result}@127.0.0.1:5432/${google_sql_database.this.name}?sslmode=disable"
  sensitive = true
}
