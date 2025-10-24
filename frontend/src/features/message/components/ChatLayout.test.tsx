import { MantineProvider } from "@mantine/core";
import { QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";

import { ChatLayout } from "./ChatLayout";

import { queryClient } from "@/lib/query";

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
  it("メッセージパネルを表示する", () => {
    render(<ChatLayout workspaceId="workspace-1" channelId="channel-1" />, { wrapper: Wrapper });

    expect(screen.getByTestId("message-panel")).toBeInTheDocument();
  });

  it("メッセージパネルに正しいプロパティを渡す", () => {
    render(<ChatLayout workspaceId="workspace-1" channelId="channel-2" />, { wrapper: Wrapper });

    const messagePanel = screen.getByTestId("message-panel");
    expect(messagePanel).toHaveAttribute("data-workspace-id", "workspace-1");
    expect(messagePanel).toHaveAttribute("data-channel-id", "channel-2");
  });

  it("チャンネルIDが null の場合でもメッセージパネルを表示する", () => {
    render(<ChatLayout workspaceId="workspace-1" channelId={null} />, { wrapper: Wrapper });

    const messagePanel = screen.getByTestId("message-panel");
    expect(messagePanel).toBeInTheDocument();
    expect(messagePanel).not.toHaveAttribute("data-channel-id");
  });
});
