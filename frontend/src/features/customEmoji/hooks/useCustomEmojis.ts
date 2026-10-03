import { createConnectQueryKey, skipToken, useQuery } from "@connectrpc/connect-query";
import { useParams } from "@tanstack/react-router";

import { CustomEmojiService } from "#/gen/chat/v1/custom_emoji_service_pb";

import type { CustomEmoji } from "#/gen/chat/v1/custom_emoji_service_pb";

// 画像 URL の有効期限（12 時間）の半分で取り直す
const IMAGE_URL_REFRESH_MS = 6 * 60 * 60 * 1000;

const emptyMap: ReadonlyMap<string, CustomEmoji> = new Map();

export const customEmojiListKey = (workspaceId: string) =>
  createConnectQueryKey({
    cardinality: "finite",
    input: { workspaceId },
    schema: CustomEmojiService.method.listCustomEmojis,
  });

export const useCustomEmojis = (workspaceId: string | undefined) =>
  useQuery(
    CustomEmojiService.method.listCustomEmojis,
    workspaceId === undefined ? skipToken : { workspaceId },
    {
      refetchInterval: IMAGE_URL_REFRESH_MS,
      refetchOnWindowFocus: false,
      select: (res) => res.emojis,
      staleTime: IMAGE_URL_REFRESH_MS,
    },
  );

// 表示中のワークスペースの絵文字を名前で引く
export const useCustomEmojiMap = () => {
  const { workspaceId } = useParams({ strict: false });
  const { data } = useCustomEmojis(workspaceId);
  return data === undefined ? emptyMap : new Map(data.map((emoji) => [emoji.name, emoji]));
};
