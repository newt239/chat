output "url" {
  value = "${google_artifact_registry_repository.this.location}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.this.repository_id}"
}

output "location" {
  value = google_artifact_registry_repository.this.location
}

output "repository" {
  value = google_artifact_registry_repository.this.name
}
