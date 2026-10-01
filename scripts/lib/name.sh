# shellcheck shell=bash disable=SC2034
# worktree ではブランチ名を前に付け、メインの環境と URL を分ける
name=chat
if [ "$(git rev-parse --absolute-git-dir)" != "$(cd "$(git rev-parse --git-common-dir)" && pwd)" ]; then
  branch=$(git branch --show-current | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9]+/-/g; s/^-+|-+$//g')
  name="$branch.chat"
fi

# Free プランの証明書は 1 階層のサブドメインにしか出ないため、ドットをハイフンにする
domain=$(git config --get chat.tunnelDomain || true)
public_app_host="${name//./-}-local.$domain"
public_api_host="api-$public_app_host"
tunnel_marker=.tunnel-public
