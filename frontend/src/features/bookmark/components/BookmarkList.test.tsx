import { render, screen } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MantineProvider } from "@mantine/core";
import { vi } from "vitest";

import { BookmarkList } from "./BookmarkList";

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });

  return ({ children }: { children: React.ReactNode }) => (
    <MantineProvider>
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    </MantineProvider>
  );
};

// Mock the API client
vi.mock("@/lib/api/client", () => ({
  api: {
    GET: vi.fn(),
  },
}));

// Mock the navigate function
vi.mock("@tanstack/react-router", () => ({
  useNavigate: () => vi.fn(),
}));

describe("BookmarkList", () => {
  it("renders loading state", () => {
    render(<BookmarkList />, { wrapper: createWrapper() });
    expect(screen.getByText("読み込み中...")).toBeInTheDocument();
  });

  it("renders empty state", async () => {
    // Mock empty response
    const { api } = await import("@/lib/api/client");
    vi.mocked(api.GET).mockResolvedValueOnce({
      data: { bookmarks: [] },
      error: undefined,
    });

    render(<BookmarkList />, { wrapper: createWrapper() });
    expect(screen.getByText("ブックマークされたメッセージはありません")).toBeInTheDocument();
  });

  it("renders error state", async () => {
    // Mock error response
    const { api } = await import("@/lib/api/client");
    vi.mocked(api.GET).mockRejectedValueOnce(new Error("API Error"));

    render(<BookmarkList />, { wrapper: createWrapper() });
    expect(screen.getByText("エラーが発生しました")).toBeInTheDocument();
  });
});
