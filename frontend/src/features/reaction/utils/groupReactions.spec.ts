import { create } from "@bufbuild/protobuf";
import { describe, expect, test } from "vite-plus/test";

import { ReactionSchema } from "#/gen/chat/v1/message_pb";
import { UserSummarySchema } from "#/gen/chat/v1/user_pb";

import { groupReactions } from "./groupReactions";

const reaction = (emoji: string, userId: string) =>
  create(ReactionSchema, {
    emoji,
    user: create(UserSummarySchema, { displayName: userId, id: userId }),
  });

describe("groupReactions", () => {
  test("絵文字ごとに件数とユーザーをまとめ、自分が付けたかを判定する", () => {
    const groups = groupReactions(
      [reaction("👍", "alice"), reaction("🎉", "bob"), reaction("👍", "bob")],
      "bob",
    );

    expect(
      groups.map(({ emoji, count, hasUserReacted }) => [emoji, count, hasUserReacted]),
    ).toEqual([
      ["👍", 2, true],
      ["🎉", 1, true],
    ]);
    expect(groups[0]?.users.map((user) => user.id)).toEqual(["alice", "bob"]);
  });

  test("ログインしていなければ自分のリアクションはない", () => {
    expect(groupReactions([reaction("👍", "alice")], null)[0]?.hasUserReacted).toBe(false);
  });
});
