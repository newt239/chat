#!/bin/bash
# ローカルの開発環境を Cloudflare Tunnel で公開・非公開にする。使い方: tunnel.sh on|off|status
set -euo pipefail

cd "$(dirname "$0")/.."
source scripts/lib/name.sh

if [ -z "$domain" ]; then
  echo "公開に使うドメインを git config chat.tunnelDomain <ドメイン> で設定してください" >&2
  exit 1
fi

# メインと各 worktree で 1 つの Tunnel とコネクタを共有し、ホスト名ごとのルールで振り分ける
tunnel_name=chat-local
container=chat-tunnel
image=$(grep -o 'cloudflare/cloudflared:[^ ]*' infra/k8s/base/cloudflared.yaml)

tunnel_id() {
  local id
  id=$(cf -q tunnels list --name "$tunnel_name" --is-deleted false | jq -r '.[0].id // empty')
  if [ -z "$id" ]; then
    id=$(cf -q tunnels create --name "$tunnel_name" --config-src cloudflare | jq -r '.id')
  fi
  echo "$id"
}

# この環境のルールを除いた ingress に、追加分と末尾の 404 を足して置き換える
update_ingress() {
  local id=$1 rules=$2 body
  body=$(ingress "$id" | jq -c --arg app "$public_app_host" --arg api "$public_api_host" --argjson rules "$rules" \
    '{config: {ingress: ([.[] | select(.hostname != $app and .hostname != $api)] + $rules + [{service: "http_status:404"}])}}')
  cf -q tunnels config update "$id" --body "$body" >/dev/null
}

# 作ったばかりの Tunnel は設定がなく 404 になる
ingress() {
  local config
  config=$(cf -q tunnels config get "$1" 2>/dev/null) || config='{}'
  jq -c '[.config.ingress[]? | select(.hostname)]' <<<"$config"
}

dns_record_id() {
  cf -q dns records list -z "$domain" --name-exact "$1" | jq -r '.[0].id // empty'
}

# portless に Host と SNI を合わせて渡す
rule() {
  jq -n --arg hostname "$1" --arg origin "$2" \
    '{hostname: $hostname, service: "https://host.docker.internal", originRequest: {httpHostHeader: $origin, originServerName: $origin, noTLSVerify: true}}'
}

on() {
  local id
  id=$(tunnel_id)
  update_ingress "$id" "$(jq -s -c . <(rule "$public_app_host" "$name.localhost") <(rule "$public_api_host" "api.$name.localhost"))"

  for host in "$public_app_host" "$public_api_host"; do
    if [ -z "$(dns_record_id "$host")" ]; then
      cf -q dns records create -z "$domain" --body "$(jq -n -c --arg name "$host" --arg content "$id.cfargotunnel.com" \
        '{type: "CNAME", name: $name, content: $content, proxied: true, ttl: 1}')" >/dev/null
    fi
  done

  if [ -z "$(docker ps -q -f "name=^$container$")" ]; then
    docker rm -f "$container" >/dev/null 2>&1 || true
    local token
    token=$(cf -q tunnels token get "$id")
    TUNNEL_TOKEN=${token//\"/} docker run -d --name "$container" --restart unless-stopped -e TUNNEL_TOKEN \
      "$image" tunnel --no-autoupdate run >/dev/null
  fi

  touch "$tunnel_marker"
  ./scripts/dev.sh
}

off() {
  local id
  id=$(tunnel_id)
  update_ingress "$id" "[]"

  local record
  for host in "$public_app_host" "$public_api_host"; do
    record=$(dns_record_id "$host")
    if [ -n "$record" ]; then
      cf -q dns records delete "$record" -z "$domain" -f >/dev/null
    fi
  done

  if [ "$(ingress "$id" | jq length)" -eq 0 ]; then
    docker rm -f "$container" >/dev/null 2>&1 || true
  fi

  rm -f "$tunnel_marker"
  ./scripts/dev.sh
}

status() {
  if [ -f "$tunnel_marker" ]; then
    echo "公開中: https://$public_app_host (API: https://$public_api_host)"
  else
    echo "非公開: https://$name.localhost"
  fi
}

case "${1:-}" in
  on | off | status) "$1" ;;
  *)
    echo "使い方: $0 on|off|status" >&2
    exit 1
    ;;
esac
