import { apiBaseUrl } from "#/lib/api/baseUrl";

export const appWebhookUrl = (appId: string, token: string) =>
  new URL(`/webhooks/${appId}/${token}`, apiBaseUrl).href;
