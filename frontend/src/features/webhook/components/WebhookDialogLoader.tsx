import { useWebhooks } from "../hooks/useWebhooks";
import { WebhookDialog } from "./WebhookDialog";

type WebhookDialogLoaderProps = {
  channelId: string;
  // null なら新規発行
  webhookId: string | null;
  onClose: () => void;
};

// URL の ID から編集する Webhook を引いてダイアログを開く。入力欄の初期値にするため、読み込めてから描く
export const WebhookDialogLoader = ({
  channelId,
  webhookId,
  onClose,
}: WebhookDialogLoaderProps) => {
  const { data } = useWebhooks(channelId);
  const webhook = data?.webhooks.find((candidate) => candidate.id === webhookId) ?? null;

  if (data === undefined || (webhookId !== null && (webhook === null || !webhook.canManage))) {
    return null;
  }
  return (
    <WebhookDialog key={webhookId} channelId={channelId} webhook={webhook} onClose={onClose} />
  );
};
