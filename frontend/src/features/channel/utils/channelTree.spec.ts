import { create } from "@bufbuild/protobuf";
import { describe, expect, test } from "vite-plus/test";

import { ChannelSchema } from "#/gen/chat/v1/channel_service_pb";

import { buildChannelTree, relativePath, sumUnread } from "./channelTree";

const dev = create(ChannelSchema, { id: "dev", name: "dev", unreadCount: 1 });
const frontend = create(ChannelSchema, {
  hasMention: true,
  id: "fe",
  name: "dev/frontend",
  parentId: "dev",
  unreadCount: 2,
});
const backend = create(ChannelSchema, {
  id: "be",
  isMuted: true,
  name: "dev/backend",
  parentId: "dev",
  unreadCount: 5,
});
const orphan = create(ChannelSchema, { id: "o", name: "secret/child", parentId: "secret" });
const general = create(ChannelSchema, { id: "g", name: "general" });

describe("buildChannelTree", () => {
  test("parent_id でつなぎ、各階層を名前順に並べる", () => {
    const tree = buildChannelTree([general, frontend, dev, backend]);
    expect(tree.map((node) => node.channel.name)).toEqual(["dev", "general"]);
    expect(tree[0]?.children.map((node) => node.channel.name)).toEqual([
      "dev/backend",
      "dev/frontend",
    ]);
  });

  test("親が一覧にないチャンネルは最上位に置く", () => {
    expect(buildChannelTree([orphan]).map((node) => node.channel.id)).toEqual(["o"]);
  });
});

describe("sumUnread", () => {
  test("自分と子孫の未読を合計し、ミュート中のチャンネルは数えない", () => {
    const [root] = buildChannelTree([dev, frontend, backend]);
    expect(root && sumUnread(root)).toEqual({ hasMention: true, unreadCount: 3 });
  });
});

describe("relativePath", () => {
  test("親から見た相対パスを返す", () => {
    expect(relativePath("dev", "dev/frontend/web")).toBe("frontend/web");
    expect(relativePath("dev", "devops")).toBe("devops");
  });
});
