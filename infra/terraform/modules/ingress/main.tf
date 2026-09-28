resource "google_compute_global_address" "this" {
  project = var.project_id
  name    = "${var.name}-ingress"
}

# ドメインを変えると作り直しになるため、名前にハッシュを含めて先に新しい証明書を作る
resource "google_compute_managed_ssl_certificate" "this" {
  project = var.project_id
  name    = "${var.name}-${substr(sha1(join(",", var.domains)), 0, 8)}"

  managed {
    domains = var.domains
  }

  lifecycle {
    create_before_destroy = true
  }
}
