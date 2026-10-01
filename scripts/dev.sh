#!/bin/bash
set -euo pipefail

cd "$(dirname "$0")/.."

# worktree ではブランチ名を前に付け、メインの環境と URL を分ける
name=chat
if [ "$(git rev-parse --absolute-git-dir)" != "$(cd "$(git rev-parse --git-common-dir)" && pwd)" ]; then
  branch=$(git branch --show-current | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9]+/-/g; s/^-+|-+$//g')
  name="$branch.chat"
fi

export CHAT_APP_URL="https://$name.localhost"
export CHAT_API_URL="https://api.$name.localhost"
export CHAT_WS_URL="wss://api.$name.localhost"

docker compose up -d --build --wait "$@"

host_port() {
  docker compose port "$1" "$2" | sed 's/.*://'
}

pnpm exec portless alias "$name" "$(host_port frontend 5173)" --force
pnpm exec portless alias "api.$name" "$(host_port backend 8080)" --force

echo
echo "App: $CHAT_APP_URL"
echo "API: $CHAT_API_URL"
echo "DB:  postgresql://postgres:postgres@$(docker compose port db 5432)/chat"
