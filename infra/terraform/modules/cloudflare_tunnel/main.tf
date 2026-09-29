resource "random_bytes" "tunnel_secret" {
  length = 32
}

# 設定はダッシュボードではなく Terraform で持つ（config_src = cloudflare でリモート管理）
resource "cloudflare_zero_trust_tunnel_cloudflared" "this" {
  account_id    = var.account_id
  name          = var.name
  config_src    = "cloudflare"
  tunnel_secret = random_bytes.tunnel_secret.base64
}

resource "cloudflare_zero_trust_tunnel_cloudflared_config" "this" {
  account_id = var.account_id
  tunnel_id  = cloudflare_zero_trust_tunnel_cloudflared.this.id

  config = {
    ingress = [
      {
        hostname = var.frontend_domain
        service  = "http://frontend.${var.kubernetes_namespace}.svc.cluster.local:80"
      },
      {
        hostname = var.api_domain
        service  = "http://backend.${var.kubernetes_namespace}.svc.cluster.local:8080"
      },
      {
        service = "http_status:404"
      },
    ]
  }
}

resource "cloudflare_dns_record" "this" {
  for_each = toset([var.frontend_domain, var.api_domain])

  zone_id = var.zone_id
  name    = each.value
  type    = "CNAME"
  content = "${cloudflare_zero_trust_tunnel_cloudflared.this.id}.cfargotunnel.com"
  proxied = true
  ttl     = 1
}

data "cloudflare_zero_trust_tunnel_cloudflared_token" "this" {
  account_id = var.account_id
  tunnel_id  = cloudflare_zero_trust_tunnel_cloudflared.this.id
}
