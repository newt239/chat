import type { Message, TimelineItem } from "#/gen/chat/v1/message_pb";
import type { ListMessagesResponse } from "#/gen/chat/v1/message_service_pb";

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
