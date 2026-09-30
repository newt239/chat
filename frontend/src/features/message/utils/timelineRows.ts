import { groupByDate } from "./dateJump";

import type { Message, SystemMessage, TimelineItem } from "#/gen/chat/v1/message_pb";

export type TimelineRow =
  | { kind: "date"; key: string; dateKey: string }
  | { kind: "user"; key: string; dateKey: string; message: Message }
  | { kind: "system"; key: string; dateKey: string; message: SystemMessage };

// 古い順の項目を、日付の区切りとメッセージを 1 行ずつ並べた仮想リストの行にする
export const buildTimelineRows = (items: readonly TimelineItem[]) => {
  const rows: TimelineRow[] = [];
  for (const { dateKey, items: dayItems } of groupByDate(items)) {
    rows.push({ dateKey, key: `d-${dateKey}`, kind: "date" });
    for (const item of dayItems) {
      if (item.content.case === "userMessage") {
        const message = item.content.value;
        rows.push({ dateKey, key: `u-${message.id}`, kind: "user", message });
      } else if (item.content.case === "systemMessage") {
        const message = item.content.value;
        rows.push({ dateKey, key: `s-${message.id}`, kind: "system", message });
      }
    }
  }
  return rows;
};

export const findRowIndex = (rows: readonly TimelineRow[], messageId: string) =>
  rows.findIndex((row) => row.kind !== "date" && row.message.id === messageId);
