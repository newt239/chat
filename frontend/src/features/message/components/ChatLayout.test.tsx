import { MantineProvider } from "@mantine/core";
import { QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";

import { ChatLayout } from "./ChatLayout";

import { queryClient } from "@/lib/query";

// Mock the components
vi.mock("@/features/channel/components/ChannelList", () => ({
  ChannelList: ({ workspaceId }: { workspaceId: string }) => (
    <div data-testid="channel-list" data-workspace-id={workspaceId}>
      Channel List
    </div>
  ),
}));

vi.mock("@/features/message/components/MessagePanel", () => ({
  MessagePanel: ({ workspaceId, channelId }: { workspaceId: string; channelId: string | null }) => (
    <div data-testid="message-panel" data-workspace-id={workspaceId} data-channel-id={channelId}>
      Message Panel
    </div>
  ),
}));

const Wrapper = ({ children }: { children: React.ReactNode }) => {
  return (
    <QueryClientProvider client={queryClient}>
      <MantineProvider>{children}</MantineProvider>
    </QueryClientProvider>
  );
};

describe("ChatLayout", () => {
  it("renders ChannelList and MessagePanel", () => {
    render(<ChatLayout workspaceId="workspace-1" channelId="channel-1" />, { wrapper: Wrapper });

    expect(screen.getByTestId("channel-list")).toBeInTheDocument();
    expect(screen.getByTestId("message-panel")).toBeInTheDocument();
  });

  it("passes correct props to ChannelList", () => {
    render(<ChatLayout workspaceId="workspace-1" channelId="channel-1" />, { wrapper: Wrapper });

    const channelList = screen.getByTestId("channel-list");
    expect(channelList).toHaveAttribute("data-workspace-id", "workspace-1");
  });

  it("passes correct props to MessagePanel", () => {
    render(<ChatLayout workspaceId="workspace-1" channelId="channel-1" />, { wrapper: Wrapper });

    const messagePanel = screen.getByTestId("message-panel");
    expect(messagePanel).toHaveAttribute("data-workspace-id", "workspace-1");
    expect(messagePanel).toHaveAttribute("data-channel-id", "channel-1");
  });

  it("handles null channelId", () => {
    render(<ChatLayout workspaceId="workspace-1" channelId={null} />, { wrapper: Wrapper });

    expect(screen.getByTestId("channel-list")).toBeInTheDocument();
    expect(screen.getByTestId("message-panel")).toBeInTheDocument();
  });

  it("renders with correct structure", () => {
    render(<ChatLayout workspaceId="workspace-1" channelId="channel-1" />, { wrapper: Wrapper });

    expect(screen.getByTestId("channel-list")).toBeInTheDocument();
    expect(screen.getByTestId("message-panel")).toBeInTheDocument();
  });
});
