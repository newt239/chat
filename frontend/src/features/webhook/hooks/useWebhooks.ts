import { createConnectQueryKey, useMutation, useQuery } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { WebhookService } from "#/gen/chat/v1/webhook_service_pb";

export const useWebhooks = (channelId: string) =>
  useQuery(WebhookService.method.listWebhooks, { channelId });

/** Webhook の発行・編集・URL の再発行・削除。成功したら一覧を取り直す */
export const useWebhookActions = (channelId: string) => {
  const queryClient = useQueryClient();
  const onSuccess = async () => {
    await queryClient.invalidateQueries({
      queryKey: createConnectQueryKey({
        cardinality: "finite",
        input: { channelId },
        schema: WebhookService.method.listWebhooks,
      }),
    });
  };
  return {
    create: useMutation(WebhookService.method.createWebhook, { onSuccess }),
    regenerate: useMutation(WebhookService.method.regenerateWebhookToken),
    remove: useMutation(WebhookService.method.deleteWebhook, { onSuccess }),
    update: useMutation(WebhookService.method.updateWebhook, { onSuccess }),
  };
};
