import { create } from "@bufbuild/protobuf";
import { render, screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { DirectMessageSchema, DirectMessageType } from "#/gen/chat/v1/direct_message_service_pb";

import { DMAvatar } from "./DMAvatar";

describe("DMAvatar", () => {
  test("グループ DM は自分を含めた人数を表示し、読み上げ用の名前を付ける", () => {
    const dm = create(DirectMessageSchema, {
      members: [{ displayName: "B" }, { displayName: "C" }, { displayName: "D" }],
      type: DirectMessageType.GROUP_DM,
    });
    render(<DMAvatar dm={dm} size={24} />);

    const avatar = screen.getByRole("img", { name: "4 人のグループ" });
    expect(avatar).toHaveTextContent("4");
    expect(avatar).toHaveStyle({ width: "24px" });
  });
});
