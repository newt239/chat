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

resource "google_container_cluster" "this" {
  project             = var.project_id
  name                = var.name
  location            = var.region
  enable_autopilot    = true
  network             = var.network_id
  subnetwork          = var.subnetwork_id
  deletion_protection = var.deletion_protection

  ip_allocation_policy {
    cluster_secondary_range_name  = var.pods_range_name
    services_secondary_range_name = var.services_range_name
  }

  private_cluster_config {
    enable_private_nodes    = true
    enable_private_endpoint = false
  }

  cluster_autoscaling {
    auto_provisioning_defaults {
      service_account = google_service_account.node.email
    }
  }

  release_channel {
    channel = "REGULAR"
  }
}
