#!/bin/bash
set -euo pipefail

cd "$(dirname "$0")/.."
source scripts/lib/name.sh

app_url="https://$name.localhost"
api_url="https://api.$name.localhost"
origins="$app_url"
if [ -f "$tunnel_marker" ]; then
  app_url="https://$public_app_host"
  api_url="https://$public_api_host"
  origins="$app_url,$origins"
fi
export CHAT_API_URL="$api_url"
export CHAT_WS_URL="${api_url/https:/wss:}"
export CHAT_CORS_ORIGINS="$origins"

docker compose up -d --build --wait "$@"

host_port() {
  docker compose port "$1" "$2" | sed 's/.*://'
}

pnpm exec portless alias "$name" "$(host_port frontend 5173)" --force
pnpm exec portless alias "api.$name" "$(host_port backend 8080)" --force

echo
echo "App: $app_url"
echo "API: $api_url"
echo "DB:  postgresql://postgres:postgres@$(docker compose port db 5432)/chat"
