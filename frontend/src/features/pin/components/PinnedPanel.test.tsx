import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { PinnedPanel } from "./PinnedPanel";

describe("PinnedPanel", () => {
  it("channelId が未指定の場合は何も表示しない", () => {
    const client = new QueryClient();
    render(
      <QueryClientProvider client={client}>
        <PinnedPanel channelId={null} />
      </QueryClientProvider>
    );

    expect(screen.queryByText("ピン留めされたメッセージはありません")).toBeNull();
  });
});
