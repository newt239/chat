import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { ChannelSortOrder } from "#/gen/chat/v1/user_pb";
import { UserService } from "#/gen/chat/v1/user_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { ChannelSectionMenu } from "./ChannelSectionMenu";

import type { UserPreferences } from "#/gen/chat/v1/user_pb";

describe("ChannelSectionMenu", () => {
  test("並び順を新しいメッセージ順に切り替えて保存する", async () => {
    const saved = vi.fn<(preferences: UserPreferences | undefined) => void>();
    await renderWithProviders(<ChannelSectionMenu workspaceId="ws1" />, "/app/ws1", (routes) => {
      routes.rpc(UserService.method.updatePreferences, ({ preferences }) => {
        saved(preferences);
        return { preferences };
      });
    });

    await userEvent.click(await screen.findByRole("button", { name: "チャンネルの表示" }));
    expect(screen.getByRole("menuitem", { name: "チャンネルに参加" })).toBeInTheDocument();
    await userEvent.click(screen.getByRole("menuitem", { name: "新しいメッセージ順" }));

    await waitFor(() => {
      expect(saved).toHaveBeenCalledWith(
        expect.objectContaining({ channelSortOrder: ChannelSortOrder.RECENT_ACTIVITY }),
      );
    });
  });
});
