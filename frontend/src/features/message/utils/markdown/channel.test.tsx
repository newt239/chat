import { MantineProvider } from "@mantine/core";
import {
  RouterProvider,
  createMemoryHistory,
  createRootRoute,
  createRouter,
  Outlet,
} from "@tanstack/react-router";
import { render } from "@testing-library/react";
import { describe, it, expect } from "vitest";

import { renderMarkdown } from "@/features/message/utils/markdown/renderer";

const rootRoute = createRootRoute({
  component: () => <Outlet />,
});

const renderWithProviders = (element: React.ReactElement) => {
  const memoryHistory = createMemoryHistory({
    initialEntries: ["/"],
  });

  const router = createRouter({
    routeTree: rootRoute,
    history: memoryHistory,
  });

  return render(
    <MantineProvider>
      <RouterProvider router={router} />
      {element}
    </MantineProvider>
  );
};

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
