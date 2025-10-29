import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach } from "vitest";

import { usePinnedMessages } from "./usePinnedMessages";

describe("usePinnedMessages", () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it("channelId が null の場合はクエリが実行されない", async () => {
    const client = new QueryClient();
    const wrapper = ({ children }: { children: React.ReactNode }) => (
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    );

    const { result } = renderHook(() => usePinnedMessages(null), { wrapper });

    await waitFor(() => expect(result.current.isFetching).toBe(false));
    expect(result.current.data).toBeUndefined();
  });
});
