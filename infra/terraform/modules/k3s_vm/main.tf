resource "google_compute_network" "this" {
  project                 = var.project_id
  name                    = var.name
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "this" {
  project       = var.project_id
  name          = var.name
  region        = var.region
  network       = google_compute_network.this.id
  ip_cidr_range = var.subnet_cidr
}

# 外からは IAP 経由の SSH だけを受ける。アプリへの通信は cloudflared が中から張る
resource "google_compute_firewall" "iap_ssh" {
  project       = var.project_id
  name          = "${var.name}-allow-iap-ssh"
  network       = google_compute_network.this.id
  source_ranges = ["35.235.240.0/20"]

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }
}

resource "google_service_account" "this" {
  project      = var.project_id
  account_id   = var.name
  display_name = "k3s node (${var.name})"
}

# Pod はメタデータサーバー経由で VM の SA を ADC として使う（FCM）
resource "google_project_iam_member" "this" {
  for_each = toset(["roles/firebasecloudmessaging.admin", "roles/logging.logWriter"])

  project = var.project_id
  role    = each.value
  member  = google_service_account.this.member
}

resource "google_compute_instance" "this" {
  project      = var.project_id
  name         = var.name
  zone         = var.zone
  machine_type = var.machine_type

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-13"
      size  = var.disk_size_gb
      type  = "pd-balanced"
    }
  }

  # NAT を使わず、一時的な外部 IP で外へ出る
  network_interface {
    subnetwork = google_compute_subnetwork.this.id

    access_config {}
  }

  # 回収されたら停止する。起動し直すのは手動
  scheduling {
    provisioning_model          = "SPOT"
    preemptible                 = true
    automatic_restart           = false
    instance_termination_action = "STOP"
  }

  service_account {
    email  = google_service_account.this.email
    scopes = ["cloud-platform"]
  }

  metadata = {
    enable-oslogin = "TRUE"
  }

  # 起動のたびに動くので、入っていなければだけ入れる
  metadata_startup_script = <<-EOT
    #!/bin/bash
    set -euo pipefail
    if ! command -v k3s >/dev/null; then
      curl -sfL https://get.k3s.io | INSTALL_K3S_VERSION='${var.k3s_version}' sh -s - server --disable traefik
    fi
  EOT

  # k3s を入れ直すためだけに VM を作り直さない
  lifecycle {
    ignore_changes = [metadata_startup_script]
  }
}
