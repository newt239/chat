import { create } from "@bufbuild/protobuf";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { UserSummarySchema } from "#/gen/chat/v1/user_pb";
import { QueryWrapper } from "#/test/QueryWrapper";

import { ReactionButton } from "./ReactionButton";

const user = (name: string) => create(UserSummarySchema, { displayName: name, id: name });

describe("ReactionButton", () => {
  test("件数と付けた人を読み上げ、自分が付けたかを押下状態で表す", async () => {
    const onPress = vi.fn<() => void>();
    render(
      <ReactionButton
        group={{ count: 2, emoji: "👍", hasUserReacted: true, users: [user("Alice"), user("Bob")] }}
        onPress={onPress}
      />,
      { wrapper: QueryWrapper },
    );

    const button = screen.getByRole("button", { name: "👍 2 件。Alice、Bob" });
    expect(button).toHaveAttribute("aria-pressed", "true");
    expect(button).toHaveTextContent("2");

    await userEvent.click(button);
    expect(onPress).toHaveBeenCalledOnce();
  });
});
