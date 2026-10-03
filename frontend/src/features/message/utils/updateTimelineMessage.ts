import { create } from "@bufbuild/protobuf";
import { timestampNow } from "@bufbuild/protobuf/wkt";

import { MessagePinSchema, ReactionSchema } from "#/gen/chat/v1/message_pb";

import type { PinEvent, ReactionEvent } from "#/gen/chat/v1/event_pb";
import type { Message, TimelineItem } from "#/gen/chat/v1/message_pb";
import type { ListMessagesResponse } from "#/gen/chat/v1/message_service_pb";
import type { WsClient } from "#/lib/ws";

import type { InfiniteData, QueryClient, QueryKey } from "@tanstack/react-query";

/** ユーザーメッセージの項目だけを更新する */
export const updateUserMessages = (
  items: TimelineItem[],
  shouldUpdate: (message: Message) => boolean,
  update: (message: Message) => Message,
) =>
  items.map((item) =>
    item.content.case === "userMessage" && shouldUpdate(item.content.value)
      ? { ...item, content: { case: "userMessage" as const, value: update(item.content.value) } }
      : item,
  );

export const updateTimelineMessage = (
  items: TimelineItem[],
  messageId: string,
  update: (message: Message) => Message,
) => updateUserMessages(items, (message) => message.id === messageId, update);

/** 読み込み済みのメッセージのページを書き換える。新しい順に並ぶので先頭のページが最新 */
export const updateMessagePages = (
  queryClient: QueryClient,
  queryKey: QueryKey,
  update: (items: TimelineItem[], isLatestPage: boolean) => TimelineItem[],
) => {
  queryClient.setQueriesData<InfiniteData<ListMessagesResponse>>(
    { queryKey },
    (data) =>
      data && {
        ...data,
        pages: data.pages.map((page, index) => ({
          ...page,
          messages: update(page.messages, index === 0 && !page.hasNewer),
        })),
      },
  );
};

export const addReaction = (message: Message, { createdAt, emoji, user, userId }: ReactionEvent) =>
  message.reactions.some((r) => r.emoji === emoji && r.user?.id === userId)
    ? message
    : {
        ...message,
        reactions: [
          ...message.reactions,
          create(ReactionSchema, {
            createdAt: createdAt ?? timestampNow(),
            emoji,
            user: user ?? { id: userId },
          }),
        ],
      };

export const removeReaction = (message: Message, { emoji, userId }: ReactionEvent) => ({
  ...message,
  reactions: message.reactions.filter((r) => !(r.emoji === emoji && r.user?.id === userId)),
});

export const addPin = (message: Message, { pinnedAt, pinnedByUser }: PinEvent) => ({
  ...message,
  pin: create(MessagePinSchema, { pinnedAt, pinnedBy: pinnedByUser }),
});

/** メッセージの更新・削除・リアクション・ピンの差分を購読し、patch で読み込み済みのデータへ当てる。戻り値を呼ぶと解除する */
export const subscribeMessagePatches = (
  wsClient: WsClient,
  patch: (
    channelId: string,
    messageIds: ReadonlySet<string>,
    update: (message: Message) => Message,
  ) => void,
) => {
  const unsubscribes = [
    wsClient.on("messageUpdated", ({ channelId, message }) => {
      if (message !== undefined) {
        // スレッドの情報は取得時にだけ付くため引き継ぐ
        patch(channelId, new Set([message.id]), (prev) => ({
          ...message,
          threadMetadata: prev.threadMetadata,
        }));
      }
    }),
    wsClient.on("messageDeleted", ({ channelId, deletedMessageIds }) => {
      patch(channelId, new Set(deletedMessageIds), (message) => ({ ...message, isDeleted: true }));
    }),
    wsClient.on("reactionAdded", (event) => {
      patch(event.channelId, new Set([event.messageId]), (message) => addReaction(message, event));
    }),
    wsClient.on("reactionRemoved", (event) => {
      patch(event.channelId, new Set([event.messageId]), (message) =>
        removeReaction(message, event),
      );
    }),
    wsClient.on("pinCreated", (event) => {
      patch(event.channelId, new Set([event.messageId]), (message) => addPin(message, event));
    }),
    wsClient.on("pinDeleted", ({ channelId, messageId }) => {
      patch(channelId, new Set([messageId]), (message) => ({ ...message, pin: undefined }));
    }),
  ];
  return () => {
    for (const unsubscribe of unsubscribes) {
      unsubscribe();
    }
  };
};
