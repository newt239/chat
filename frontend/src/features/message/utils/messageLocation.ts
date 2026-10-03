import { linkOptions } from "@tanstack/react-router";

type MessageLocationArgs = {
  workspaceId: string;
  channelId: string;
  messageId: string | undefined;
  parentId: string | undefined;
};

// メッセージを開くリンク。返信はスレッドの中で開く
export const messageLocation = ({
  workspaceId,
  channelId,
  messageId,
  parentId,
}: MessageLocationArgs) =>
  parentId === undefined
    ? linkOptions({
        params: { channelId, workspaceId },
        search: { message: messageId },
        to: "/app/$workspaceId/$channelId",
      })
    : linkOptions({
        params: { channelId, messageId: parentId, workspaceId },
        search: { message: messageId },
        to: "/app/$workspaceId/$channelId/thread/$messageId",
      });
