import type { Message, TimelineItem } from "#/gen/chat/v1/message_pb";

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
