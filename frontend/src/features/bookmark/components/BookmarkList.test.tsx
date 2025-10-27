import { render, screen } from "@testing-library/react";
import { vi } from "vitest";

import { BookmarkList } from "./BookmarkList";

import { api } from "@/lib/api/client";
import { createAppWrapper, createTestQueryClient } from "@/test/utils";

const createWrapper = () => createAppWrapper(createTestQueryClient());

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
    vi.mocked(api.GET).mockResolvedValueOnce({
      data: { bookmarks: [] },
      error: undefined,
    });

    render(<BookmarkList />, { wrapper: createWrapper() });
    expect(screen.getByText("ブックマークされたメッセージはありません")).toBeInTheDocument();
  });

  it("renders error state", async () => {
    // Mock error response
    vi.mocked(api.GET).mockRejectedValueOnce(new Error("API Error"));

    render(<BookmarkList />, { wrapper: createWrapper() });
    expect(screen.getByText("エラーが発生しました")).toBeInTheDocument();
  });
});
