import { fromDate, parseDate, toCalendarDate } from "@internationalized/date";
import { z } from "zod";

import { toDate } from "#/lib/timestamp";

import type { TimelineItem } from "#/gen/chat/v1/message_pb";

export const FIRST_MESSAGE = "first";

// ?date= の値。YYYY-MM-DD か、最初のメッセージを表す "first"
export const jumpDateSchema = z.union([z.literal(FIRST_MESSAGE), z.iso.date()]);

// プロフィールのタイムゾーンでの日付。日付区切りと日付ジャンプの単位
export const toDateKey = (date: Date, timeZone: string) =>
  toCalendarDate(fromDate(date, timeZone)).toString();

export const startOfDateKey = (dateKey: string, timeZone: string) =>
  dateKey === FIRST_MESSAGE ? new Date(0) : parseDate(dateKey).toDate(timeZone);

// 3/31 の先月は 2/28 にする
export const jumpPresets = (now: Date, timeZone: string) => {
  const today = toCalendarDate(fromDate(now, timeZone));
  return {
    lastMonth: today.subtract({ months: 1 }).toString(),
    lastWeek: today.subtract({ days: 7 }).toString(),
    today: today.toString(),
    yesterday: today.subtract({ days: 1 }).toString(),
  };
};

// 古い順に並んだ項目を日付ごとにまとめる
export const groupByDate = (items: readonly TimelineItem[], timeZone: string) => {
  const groups: { dateKey: string; items: TimelineItem[] }[] = [];
  for (const item of items) {
    const dateKey = toDateKey(toDate(item.createdAt), timeZone);
    const last = groups.at(-1);
    if (last?.dateKey === dateKey) {
      last.items.push(item);
    } else {
      groups.push({ dateKey, items: [item] });
    }
  }
  return groups;
};
