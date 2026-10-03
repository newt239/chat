import { create } from "@bufbuild/protobuf";
import { expect, test } from "vite-plus/test";

import { DirectMessageSchema, DirectMessageType } from "#/gen/chat/v1/direct_message_service_pb";

import { dmName } from "./dmName";

const members = [
  { displayName: "Bob", userId: "b" },
  { displayName: "Carol", userId: "c" },
];
const displayName = (_userId: string, name: string) => name;

test("1 対 1 は相手の名前にする", () => {
  const dm = create(DirectMessageSchema, {
    members: members.slice(0, 1),
    name: "dm-xxx",
    type: DirectMessageType.DM,
  });
  expect(dmName(dm, displayName)).toBe("Bob");
});

test("グループは付けた名前があっても参加者を並べる", () => {
  const dm = create(DirectMessageSchema, {
    members,
    name: "設計",
    type: DirectMessageType.GROUP_DM,
  });
  expect(dmName(dm, displayName)).toBe("Bob, Carol");
});

test("ニックネームを付けた相手はニックネームで並べる", () => {
  const dm = create(DirectMessageSchema, { members, type: DirectMessageType.GROUP_DM });
  expect(dmName(dm, (userId, name) => (userId === "b" ? "ボブさん" : name))).toBe(
    "ボブさん, Carol",
  );
});
