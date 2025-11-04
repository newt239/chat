# Cloudflare DNS (optional)

## 事前準備

- Cloudflare API Token を用意（DNS 編集権限）
- `CLOUDFLARE_API_TOKEN` を環境変数に設定

```bash
export CLOUDFLARE_API_TOKEN=xxxxx
```

## 使い方

```bash
cd terraform/cloudflare
terraform init
terraform apply \
  -var "zone_id=<YOUR_ZONE_ID>" \
  -var "conoha_ip=<VPS_PUBLIC_IP>"
```

生成レコード:

- `chat.<zone>` → A レコード（本番）
- `*.preview.<zone>` → A レコード（プレビュー共通）
