import { create } from "@bufbuild/protobuf";
import { timestampFromMs } from "@bufbuild/protobuf/wkt";
import { describe, expect, test } from "vite-plus/test";

import { ChannelSchema } from "#/gen/chat/v1/channel_service_pb";

import {
  buildChannelTree,
  categoryOfChannel,
  relativePath,
  sortChannelsByActivity,
  sumUnread,
} from "./channelTree";

const dev = create(ChannelSchema, { id: "dev", name: "dev", unreadCount: 1 });
const frontend = create(ChannelSchema, {
  hasMention: true,
  id: "fe",
  mentionCount: 1,
  name: "dev/frontend",
  parentId: "dev",
  unreadCount: 2,
});
const backend = create(ChannelSchema, {
  hasMention: true,
  id: "be",
  isMuted: true,
  mentionCount: 2,
  name: "dev/backend",
  parentId: "dev",
  unreadCount: 5,
});
const orphan = create(ChannelSchema, { id: "o", name: "secret/child", parentId: "secret" });
const general = create(ChannelSchema, { id: "g", name: "general" });
const at = (ms: number) => timestampFromMs(ms);

describe("buildChannelTree", () => {
  test("parent_id でつなぎ、各階層を名前順に並べてミュート中は最後に回す", () => {
    const tree = buildChannelTree([general, frontend, dev, backend]);
    expect(tree.map((node) => node.channel.name)).toEqual(["dev", "general"]);
    expect(tree[0]?.children.map((node) => node.channel.name)).toEqual([
      "dev/frontend",
      "dev/backend",
    ]);
  });

  test("親が一覧にないチャンネルは最上位に置く", () => {
    expect(buildChannelTree([orphan]).map((node) => node.channel.id)).toEqual(["o"]);
  });
});

describe("sortChannelsByActivity", () => {
  test("階層を無視して新しいメッセージ順に並べ、ミュート中と未参加は除くか最後に回す", () => {
    const channels = [
      create(ChannelSchema, { id: "a", isMember: true, lastMessageAt: at(1000), name: "a" }),
      create(ChannelSchema, { id: "b", isMember: true, lastMessageAt: at(3000), name: "a/b" }),
      create(ChannelSchema, { createdAt: at(2000), id: "c", isMember: true, name: "c" }),
      create(ChannelSchema, {
        id: "m",
        isMember: true,
        isMuted: true,
        lastMessageAt: at(9000),
        name: "m",
      }),
      create(ChannelSchema, { id: "x", lastMessageAt: at(9000), name: "x" }),
    ];
    expect(sortChannelsByActivity(channels).map((channel) => channel.id)).toEqual([
      "b",
      "c",
      "a",
      "m",
    ]);
  });
});

describe("categoryOfChannel", () => {
  test("自分の割り当て、なければ最も近い祖先の割り当てに従う", () => {
    const channels = [dev, frontend, backend, general];
    const assigned = new Map([
      ["dev", "work"],
      ["be", "later"],
    ]);
    expect(categoryOfChannel(frontend, channels, assigned)).toBe("work");
    expect(categoryOfChannel(backend, channels, assigned)).toBe("later");
    expect(categoryOfChannel(general, channels, assigned)).toBeNull();
  });
});

describe("sumUnread", () => {
  test("自分と子孫の未読を合計し、未読はミュート中のチャンネルを数えないがメンションは数える", () => {
    const [root] = buildChannelTree([dev, frontend, backend]);
    expect(root && sumUnread(root)).toEqual({ mentionCount: 3, unreadCount: 3 });
  });
});

describe("relativePath", () => {
  test("親から見た相対パスを返す", () => {
    expect(relativePath("dev", "dev/frontend/web")).toBe("frontend/web");
    expect(relativePath("dev", "devops")).toBe("devops");
  });
});
