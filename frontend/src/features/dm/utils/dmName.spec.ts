import { create } from "@bufbuild/protobuf";
import { expect, test } from "vite-plus/test";

import { DirectMessageSchema, DirectMessageType } from "#/gen/chat/v1/direct_message_service_pb";

import { dmName } from "./dmName";

const members = [
  { displayName: "Bob", userId: "b" },
  { displayName: "Carol", userId: "c" },
];

test("1 対 1 は相手の名前にする", () => {
  const dm = create(DirectMessageSchema, {
    members: members.slice(0, 1),
    name: "dm-xxx",
    type: DirectMessageType.DM,
  });
  expect(dmName(dm)).toBe("Bob");
});

test("グループは名前があれば名前、なければ参加者を並べる", () => {
  expect(
    dmName(
      create(DirectMessageSchema, { members, name: "設計", type: DirectMessageType.GROUP_DM }),
    ),
  ).toBe("設計");
  expect(dmName(create(DirectMessageSchema, { members, type: DirectMessageType.GROUP_DM }))).toBe(
    "Bob, Carol",
  );
});
