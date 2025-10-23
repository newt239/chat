import { MantineProvider } from "@mantine/core";
import { QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";

import { Header } from "./Header";

import { queryClient } from "@/lib/query";

// Mock the hooks
vi.mock("@/features/workspace/hooks/useWorkspace", () => ({
  useWorkspaces: vi.fn(() => ({
    data: [
      { id: "1", name: "開発チーム", description: "開発用ワークスペース" },
      { id: "2", name: "マーケティング", description: null },
    ],
    isLoading: false,
    error: null,
  })),
}));

vi.mock("@/lib/store/auth", () => ({
  useAuthStore: vi.fn(() => ({
    clearAuth: vi.fn(),
  })),
}));

vi.mock("@/lib/store/workspace", () => ({
  useWorkspaceStore: vi.fn(() => ({
    currentWorkspaceId: "1",
    setCurrentWorkspace: vi.fn(),
  })),
}));

const Wrapper = ({ children }: { children: React.ReactNode }) => {
  return (
    <QueryClientProvider client={queryClient}>
      <MantineProvider>{children}</MantineProvider>
    </QueryClientProvider>
  );
};

describe("Header", () => {
  it("renders app title", () => {
    render(<Header />, { wrapper: Wrapper });

    expect(screen.getByText("Chat App")).toBeInTheDocument();
  });

  it("renders workspace selection button", () => {
    render(<Header />, { wrapper: Wrapper });

    expect(screen.getByText("ワークスペースを選択")).toBeInTheDocument();
  });

  it("renders logout button", () => {
    render(<Header />, { wrapper: Wrapper });

    expect(screen.getByRole("button", { name: /ログアウト/i })).toBeInTheDocument();
  });
});
