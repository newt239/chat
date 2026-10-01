import { apiBaseUrl } from "#/lib/api/baseUrl";

export const webhookUrl = (webhookId: string, token: string) =>
  new URL(`/webhooks/${webhookId}/${token}`, apiBaseUrl).href;
