import { create } from "@bufbuild/protobuf";
import { describe, expect, test } from "vite-plus/test";

import { ChannelCategorySchema } from "#/gen/chat/v1/channel_category_service_pb";
import { ChannelSchema } from "#/gen/chat/v1/channel_service_pb";
import { DirectMessageSchema } from "#/gen/chat/v1/direct_message_service_pb";

import { nextUnreadId, sidebarOrder } from "./sidebarOrder";

const channel = (id: string, name: string, extra: object = {}) =>
  create(ChannelSchema, { id, isMember: true, name, ...extra });

const channels = [
  channel("dev", "dev"),
  channel("fe", "dev/frontend", { parentId: "dev", unreadCount: 1 }),
  channel("g", "general", { isStarred: true }),
  channel("muted", "muted", { isMuted: true, unreadCount: 5 }),
  channel("work", "work", { unreadCount: 2 }),
];
const dms = [create(DirectMessageSchema, { id: "dm", unreadCount: 1 })];
const categories = [create(ChannelCategorySchema, { channelIds: ["work"], id: "c1" })];

describe("sidebarOrder", () => {
  test("スター、カテゴリ、チャンネルのツリー、DM の順に重複なく並べる", () => {
    expect(
      sidebarOrder({ categories, channels, dms, order: "default" }).map(({ id }) => id),
    ).toEqual(["g", "work", "dev", "fe", "muted", "dm"]);
  });
});

describe("nextUnreadId", () => {
  const items = sidebarOrder({ categories, channels, dms, order: "default" });

  test("現在の会話より後ろの未読を返し、ミュート中は飛ばす", () => {
    expect(nextUnreadId(items, "fe")).toBe("dm");
    expect(nextUnreadId(items, "work")).toBe("fe");
  });

  test("後ろに無ければ先頭に戻って探す", () => {
    expect(nextUnreadId(items, "dm")).toBe("work");
  });

  test("会話を開いていなければ先頭から探す", () => {
    expect(nextUnreadId(items, undefined)).toBe("work");
  });
});
