import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { ChannelMemberService } from "#/gen/chat/v1/channel_member_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { JoinChannelBar } from "./JoinChannelBar";

import type { JoinChannelRequest } from "#/gen/chat/v1/channel_member_service_pb";

describe("JoinChannelBar", () => {
  test("プレビュー中であることを示し、参加ボタンで参加する", async () => {
    const joined = vi.fn<(req: JoinChannelRequest) => void>();
    await renderWithProviders(
      <JoinChannelBar workspaceId="ws1" channelId="c1" channelName="design" />,
      "/app/ws1",
      (routes) => {
        routes.rpc(ChannelMemberService.method.joinChannel, (req) => {
          joined(req);
          return {};
        });
      },
    );

    expect(await screen.findByText("#design をプレビューしています")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "チャンネルに参加" }));
    await waitFor(() => {
      expect(joined).toHaveBeenCalledWith(expect.objectContaining({ channelId: "c1" }));
    });
  });
});
