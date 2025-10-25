import type { ReactElement } from "react";

import { render } from "@testing-library/react";
import { describe, it, expect, vi, afterEach } from "vitest";

import { renderMarkdown } from "@/features/message/utils/markdown/renderer";
import { createAppWrapper } from "@/test/utils";

const mockNavigate = vi.fn();

vi.mock("@tanstack/react-router", () => ({
  useNavigate: () => mockNavigate,
  useParams: () => ({ workspaceId: "workspace-1" }),
}));

const renderWithProviders = (element: ReactElement) =>
  render(element, { wrapper: createAppWrapper() });

afterEach(() => {
  mockNavigate.mockClear();
});

describe("Channel link rendering", () => {
  it("チャンネルリンクをレンダリングできる", () => {
    const content = "Check out #general channel.";
    const result = renderMarkdown(content);
    const { container } = renderWithProviders(<>{result}</>);

    const channelLink = container.querySelector(".channel-link");
    expect(channelLink).toBeDefined();
    expect(channelLink).toHaveAttribute("data-channel", "general");
    expect(channelLink).toHaveTextContent("#general");
  });

  it("複数のチャンネルリンクをレンダリングできる", () => {
    const content = "See #general and #random channels.";
    const result = renderMarkdown(content);
    const { container } = renderWithProviders(<>{result}</>);

    const channelLinks = container.querySelectorAll(".channel-link");
    expect(channelLinks).toHaveLength(2);
    expect(channelLinks[0]).toHaveAttribute("data-channel", "general");
    expect(channelLinks[1]).toHaveAttribute("data-channel", "random");
  });
});
