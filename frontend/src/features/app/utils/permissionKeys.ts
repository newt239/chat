import { AppPermission } from "#/gen/chat/v1/app_service_pb";

// 画面に並べる順と、辞書の app.permissions.* のキー
export const appPermissionKeys = [
  [AppPermission.POST_JOINED_CHANNELS, "postJoinedChannels"],
  [AppPermission.POST_PUBLIC_CHANNELS, "postPublicChannels"],
  [AppPermission.POST_THREAD_REPLIES, "postThreadReplies"],
  [AppPermission.OUTGOING_WEBHOOK, "outgoingWebhook"],
] as const;
