output "tunnel_id" {
  value = cloudflare_zero_trust_tunnel_cloudflared.this.id
}

# cloudflared の TUNNEL_TOKEN に渡す
output "token" {
  value     = data.cloudflare_zero_trust_tunnel_cloudflared_token.this.token
  sensitive = true
}
