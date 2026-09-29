resource "google_service_account" "node" {
  project      = var.project_id
  account_id   = "${var.name}-node"
  display_name = "GKE node (${var.name})"
}

resource "google_project_iam_member" "node" {
  project = var.project_id
  role    = "roles/container.defaultNodeServiceAccount"
  member  = google_service_account.node.member
}

# ゾーンクラスタは管理料が無料枠で相殺される
resource "google_container_cluster" "this" {
  project             = var.project_id
  name                = var.name
  location            = var.zone
  network             = var.network_id
  subnetwork          = var.subnetwork_id
  deletion_protection = var.deletion_protection

  # ノードは下の Spot のノードプールだけで動かす
  remove_default_node_pool = true
  initial_node_count       = 1

  node_config {
    service_account = google_service_account.node.email
  }

  ip_allocation_policy {
    cluster_secondary_range_name  = var.pods_range_name
    services_secondary_range_name = var.services_range_name
  }

  private_cluster_config {
    enable_private_nodes    = true
    enable_private_endpoint = false
  }

  workload_identity_config {
    workload_pool = "${var.project_id}.svc.id.goog"
  }

  release_channel {
    channel = "REGULAR"
  }
}

resource "google_container_node_pool" "spot" {
  project  = var.project_id
  name     = "spot"
  cluster  = google_container_cluster.this.id
  location = var.zone

  autoscaling {
    min_node_count = var.min_node_count
    max_node_count = var.max_node_count
  }

  management {
    auto_repair  = true
    auto_upgrade = true
  }

  node_config {
    machine_type    = var.machine_type
    spot            = true
    disk_type       = "pd-standard"
    disk_size_gb    = 30
    service_account = google_service_account.node.email
    oauth_scopes    = ["https://www.googleapis.com/auth/cloud-platform"]

    workload_metadata_config {
      mode = "GKE_METADATA"
    }

    shielded_instance_config {
      enable_secure_boot = true
    }
  }
}
