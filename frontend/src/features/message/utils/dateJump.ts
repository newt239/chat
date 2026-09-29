import { z } from "zod";

import { toDate } from "#/lib/timestamp";

import type { TimelineItem } from "#/gen/chat/v1/message_pb";

export const FIRST_MESSAGE = "first";

// ?date= の値。YYYY-MM-DD か、最初のメッセージを表す "first"
export const jumpDateSchema = z.union([z.literal(FIRST_MESSAGE), z.iso.date()]);

const pad = (value: number) => String(value).padStart(2, "0");

// 端末のタイムゾーンでの日付。日付区切りと日付ジャンプの単位
export const toDateKey = (date: Date) =>
  `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`;

// オフセットのない日時の文字列はローカル時刻として解釈される
export const startOfDateKey = (dateKey: string) =>
  dateKey === FIRST_MESSAGE ? new Date(0) : new Date(`${dateKey}T00:00:00`);

export const jumpPresets = (now: Date) => {
  const daysAgo = (days: number) =>
    toDateKey(new Date(now.getFullYear(), now.getMonth(), now.getDate() - days));
  // 3/31 の先月は 2/28 にする
  const lastDayOfLastMonth = new Date(now.getFullYear(), now.getMonth(), 0).getDate();
  return {
    lastMonth: toDateKey(
      new Date(now.getFullYear(), now.getMonth() - 1, Math.min(now.getDate(), lastDayOfLastMonth)),
    ),
    lastWeek: daysAgo(7),
    today: daysAgo(0),
    yesterday: daysAgo(1),
  };
};

// 古い順に並んだ項目を日付ごとにまとめる
export const groupByDate = (items: readonly TimelineItem[]) => {
  const groups: { dateKey: string; items: TimelineItem[] }[] = [];
  for (const item of items) {
    const dateKey = toDateKey(toDate(item.createdAt));
    const last = groups.at(-1);
    if (last?.dateKey === dateKey) {
      last.items.push(item);
    } else {
      groups.push({ dateKey, items: [item] });
    }
  }
  return groups;
};
