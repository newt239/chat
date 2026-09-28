import { useState } from "react";

import { formatRelativeTime } from "@chat/i18n";
import { IconEdit } from "@tabler/icons-react";
import { useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { Avatar } from "#/components/ui/Avatar";
import { Button } from "#/components/ui/Button";
import { IconButton } from "#/components/ui/IconButton";
import { toDate } from "#/lib/timestamp";
import { preferencesAtom } from "#/providers/store/preferences";

import { useWebhooks } from "../hooks/useWebhooks";
import { WebhookDialog } from "./WebhookDialog";

type WebhooksSectionProps = {
  channelId: string;
};

const NEW_WEBHOOK = "new";

// チャンネル情報の Webhook 一覧。誰でも追加でき、作成者と管理者は編集・削除できる
export const WebhooksSection = ({ channelId }: WebhooksSectionProps) => {
  const { t } = useTranslation();
  const { locale } = useAtomValue(preferencesAtom);
  const { data } = useWebhooks(channelId);
  // 開いているダイアログ。NEW_WEBHOOK は新規発行、それ以外は編集する Webhook の ID
  const [dialog, setDialog] = useState<string | null>(null);
  const webhooks = data?.webhooks ?? [];
  const editing = webhooks.find((webhook) => webhook.id === dialog);

  return (
    <section className="flex flex-col gap-1.5 border-b border-border px-4 py-3">
      <h4 className="m-0 flex items-center justify-between text-xs font-semibold text-muted">
        {t("webhook.heading")}
        <Button
          size="sm"
          variant="ghost"
          onPress={() => {
            setDialog(NEW_WEBHOOK);
          }}
        >
          {t("webhook.add")}
        </Button>
      </h4>
      {webhooks.length === 0 && (
        <p className="m-0 text-[12.5px] text-muted">{t("webhook.empty")}</p>
      )}
      <ul className="m-0 -mx-2 flex list-none flex-col p-0">
        {webhooks.map((webhook) => (
          <li key={webhook.id} className="flex items-center gap-2 rounded-md px-2 py-1">
            <Avatar name={webhook.name} src={webhook.avatarUrl} size={28} />
            <span className="flex min-w-0 flex-1 flex-col leading-[1.35]">
              <span className="truncate text-[13.5px]">{webhook.name}</span>
              <small className="truncate text-[11.5px] text-subtle">
                {t("webhook.createdBy", { name: webhook.createdBy?.displayName ?? "" })}
                {" · "}
                {webhook.lastUsedAt
                  ? t("webhook.lastUsed", {
                      time: formatRelativeTime(toDate(webhook.lastUsedAt), new Date(), locale),
                    })
                  : t("webhook.neverUsed")}
              </small>
            </span>
            {webhook.canManage && (
              <span className="flex shrink-0 [&_button]:size-7 [&_svg]:size-4!">
                <IconButton
                  label={t("webhook.editOf", { name: webhook.name })}
                  onPress={() => {
                    setDialog(webhook.id);
                  }}
                >
                  <IconEdit />
                </IconButton>
              </span>
            )}
          </li>
        ))}
      </ul>
      {(dialog === NEW_WEBHOOK || editing) && (
        <WebhookDialog
          key={dialog}
          channelId={channelId}
          webhook={editing ?? null}
          onClose={() => {
            setDialog(null);
          }}
        />
      )}
    </section>
  );
};
