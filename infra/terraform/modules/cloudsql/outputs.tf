output "connection_name" {
  value = google_sql_database_instance.this.connection_name
}

# Pod 内の Cloud SQL Auth Proxy が TLS を張るため、アプリからの接続は平文でよい
output "database_url" {
  value     = "postgresql://${google_sql_user.app.name}:${random_password.app.result}@127.0.0.1:5432/${google_sql_database.app.name}?sslmode=disable"
  sensitive = true
}
