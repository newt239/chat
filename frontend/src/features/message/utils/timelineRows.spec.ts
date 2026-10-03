import { create } from "@bufbuild/protobuf";
import { timestampFromDate } from "@bufbuild/protobuf/wkt";
import { getLocalTimeZone } from "@internationalized/date";
import { describe, expect, test } from "vite-plus/test";

import {
  MessageSchema,
  SystemMessageKind,
  SystemMessageSchema,
  TimelineItemSchema,
} from "#/gen/chat/v1/message_pb";

import { buildTimelineRows } from "./timelineRows";

const at = (iso: string) => timestampFromDate(new Date(iso));

const userItem = (id: string, iso: string) =>
  create(TimelineItemSchema, {
    content: { case: "userMessage", value: create(MessageSchema, { id }) },
    createdAt: at(iso),
  });

const systemItem = (id: string, iso: string, kind = SystemMessageKind.CHANNEL_NAME_CHANGED) =>
  create(TimelineItemSchema, {
    content: { case: "systemMessage", value: create(SystemMessageSchema, { id, kind }) },
    createdAt: at(iso),
  });

describe("buildTimelineRows", () => {
  test("日付が変わるところに区切りの行を入れる", () => {
    const rows = buildTimelineRows(
      [
        userItem("m1", "2026-09-28T10:00:00"),
        systemItem("s1", "2026-09-28T11:00:00"),
        userItem("m2", "2026-09-29T09:00:00"),
      ],
      false,
      getLocalTimeZone(),
    );
    expect(rows.map((row) => row.key)).toEqual([
      "d-2026-09-28",
      "u-m1",
      "s-s1",
      "d-2026-09-29",
      "u-m2",
    ]);
  });

  test("参加のお知らせを隠すと、それしかない日の区切りも出さない", () => {
    const items = [
      systemItem("s1", "2026-09-27T10:00:00", SystemMessageKind.MEMBER_JOINED),
      systemItem("s2", "2026-09-28T10:00:00", SystemMessageKind.MEMBER_ADDED),
      systemItem("s3", "2026-09-28T11:00:00"),
    ];
    expect(buildTimelineRows(items, true, getLocalTimeZone()).map((row) => row.key)).toEqual([
      "d-2026-09-28",
      "s-s3",
    ]);
    expect(buildTimelineRows(items, false, getLocalTimeZone())).toHaveLength(5);
  });
});
