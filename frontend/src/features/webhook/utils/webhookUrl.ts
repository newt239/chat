import { apiBaseUrl } from "#/lib/api/baseUrl";

export const webhookUrl = (webhookId: string, token: string) =>
  new URL(`/webhooks/${webhookId}/${token}`, apiBaseUrl).href;

export const isHttpUrl = (value: string) => {
  try {
    const { protocol } = new URL(value);
    return protocol === "http:" || protocol === "https:";
  } catch {
    return false;
  }
};
