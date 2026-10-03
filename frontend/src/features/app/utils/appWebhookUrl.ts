import { apiBaseUrl } from "#/lib/api/createTransport";

export const appWebhookUrl = (appId: string, token: string) =>
  new URL(`/webhooks/${appId}/${token}`, apiBaseUrl).href;
