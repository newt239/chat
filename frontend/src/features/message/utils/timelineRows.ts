import { SystemMessageKind } from "#/gen/chat/v1/message_pb";
import { toDate } from "#/lib/timestamp";

import { toDateKey } from "./dateJump";

import type { Message, SystemMessage, TimelineItem } from "#/gen/chat/v1/message_pb";

export type TimelineRow =
  // スレッドの親メッセージなど、一覧の先頭に置く行
  | { kind: "header"; key: string; dateKey: string }
  | { kind: "date"; key: string; dateKey: string }
  | { kind: "user"; key: string; dateKey: string; message: Message }
  | { kind: "system"; key: string; dateKey: string; message: SystemMessage };

const joinKinds = new Set([SystemMessageKind.MEMBER_JOINED, SystemMessageKind.MEMBER_ADDED]);

// 古い順の項目を日付の区切りとメッセージの行にする。参加のお知らせだけの日は区切りも出さない
export const buildTimelineRows = (
  items: readonly TimelineItem[],
  hideJoinMessages: boolean,
  timeZone: string,
) => {
  const visible = hideJoinMessages
    ? items.filter(
        ({ content }) => content.case !== "systemMessage" || !joinKinds.has(content.value.kind),
      )
    : items;
  const rows: TimelineRow[] = [];
  for (const item of visible) {
    const dateKey = toDateKey(toDate(item.createdAt), timeZone);
    if (rows.at(-1)?.dateKey !== dateKey) {
      rows.push({ dateKey, key: `d-${dateKey}`, kind: "date" });
    }
    if (item.content.case === "userMessage") {
      const message = item.content.value;
      rows.push({ dateKey, key: `u-${message.id}`, kind: "user", message });
    } else if (item.content.case === "systemMessage") {
      const message = item.content.value;
      rows.push({ dateKey, key: `s-${message.id}`, kind: "system", message });
    }
  }
  return rows;
};
