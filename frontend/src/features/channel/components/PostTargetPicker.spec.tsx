import { create } from "@bufbuild/protobuf";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { ChannelSchema } from "#/gen/chat/v1/channel_service_pb";

import { PostTargetPicker } from "./PostTargetPicker";

const parent = create(ChannelSchema, { id: "dev", name: "dev" });
const child = create(ChannelSchema, { id: "web", name: "dev/frontend/web" });

describe("PostTargetPicker", () => {
  test("親と子孫から投稿先を選ぶ", async () => {
    const onChange = vi.fn<(channelId: string) => void>();
    render(
      <PostTargetPicker parent={parent} descendants={[child]} value="dev" onChange={onChange} />,
    );

    await userEvent.click(screen.getByRole("button", { name: "投稿先: #dev" }));
    expect(
      await screen.findByRole("menuitem", { name: "# dev（このチャンネル）" }),
    ).toBeInTheDocument();
    await userEvent.click(screen.getByRole("menuitem", { name: "# frontend/web" }));
    expect(onChange).toHaveBeenCalledWith("web");
  });
});
