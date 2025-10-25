import { render, screen } from "@testing-library/react";
import { vi } from "vitest";

import { ChannelMembersModal } from "./ChannelMembersModal";

import { createAppWrapper, createTestQueryClient } from "@/test/utils";

const createWrapper = () => createAppWrapper(createTestQueryClient());

vi.mock("@/lib/api/client", () => ({
  api: {
    GET: vi.fn(),
    POST: vi.fn(),
    PATCH: vi.fn(),
    DELETE: vi.fn(),
  },
}));

describe("ChannelMembersModal", () => {
  it("renders when closed", () => {
    render(
      <ChannelMembersModal
        channelId="channel-1"
        workspaceId="workspace-1"
        opened={false}
        onClose={vi.fn()}
      />,
      { wrapper: createWrapper() }
    );

    expect(screen.queryByText("チャンネルメンバー管理")).not.toBeInTheDocument();
  });

  it("renders when opened", async () => {
    const { api } = await import("@/lib/api/client");
    vi.mocked(api.GET).mockResolvedValue({
      data: { members: [] },
      error: undefined,
    });

    render(
      <ChannelMembersModal
        channelId="channel-1"
        workspaceId="workspace-1"
        opened={true}
        onClose={vi.fn()}
      />,
      { wrapper: createWrapper() }
    );

    expect(screen.getByText("チャンネルメンバー管理")).toBeInTheDocument();
  });
});
