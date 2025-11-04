terraform {
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 4.0"
    }
  }
}

provider "cloudflare" {}

resource "cloudflare_record" "prod" {
  zone_id = var.zone_id
  name    = "chat"
  value   = var.conoha_ip
  type    = "A"
  proxied = false
}

resource "cloudflare_record" "preview_wildcard" {
  zone_id = var.zone_id
  name    = "*.preview"
  value   = var.conoha_ip
  type    = "A"
  proxied = false
}


