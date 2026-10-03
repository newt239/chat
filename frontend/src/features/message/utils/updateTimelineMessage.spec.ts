import { create } from "@bufbuild/protobuf";
import { describe, expect, test } from "vite-plus/test";

import { PinEventSchema, ReactionEventSchema } from "#/gen/chat/v1/event_pb";
import { MessageSchema } from "#/gen/chat/v1/message_pb";

import { addPin, addReaction, removeReaction } from "./updateTimelineMessage";

const message = create(MessageSchema, { id: "m1" });
const reaction = create(ReactionEventSchema, {
  emoji: "👍",
  messageId: "m1",
  user: { displayName: "Alice", id: "u1" },
  userId: "u1",
});

describe("リアクションの差分", () => {
  test("同じ人の同じ絵文字は二重に足さない", () => {
    const added = addReaction(message, reaction);

    expect(added.reactions).toHaveLength(1);
    expect(addReaction(added, reaction).reactions).toHaveLength(1);
  });

  test("外すとその人の絵文字だけが消える", () => {
    const other = create(ReactionEventSchema, {
      emoji: "👍",
      messageId: "m1",
      user: { displayName: "Bob", id: "u2" },
      userId: "u2",
    });
    const both = addReaction(addReaction(message, reaction), other);

    expect(removeReaction(both, reaction).reactions.map((r) => r.user?.id)).toEqual(["u2"]);
  });
});

test("ピン留めした人を付ける", () => {
  const pinned = addPin(
    message,
    create(PinEventSchema, { messageId: "m1", pinnedByUser: { displayName: "Bob", id: "u2" } }),
  );

  expect(pinned.pin?.pinnedBy?.id).toBe("u2");
});
