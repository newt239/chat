import { MantineProvider } from "@mantine/core";
import { QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi } from "vitest";

import { WorkspaceList } from "./WorkspaceList";

import { queryClient } from "@/lib/query";

// Mock the hooks
vi.mock("../hooks/useWorkspace", () => ({
  useWorkspaces: vi.fn(() => ({
    data: [
      { id: "1", name: "開発チーム", description: "開発用ワークスペース" },
      { id: "2", name: "マーケティング", description: null },
    ],
    isLoading: false,
    error: null,
  })),
  useCreateWorkspace: vi.fn(() => ({
    mutate: vi.fn(),
    isPending: false,
    isError: false,
    error: null,
  })),
}));

const Wrapper = ({ children }: { children: React.ReactNode }) => {
  return (
    <QueryClientProvider client={queryClient}>
      <MantineProvider>{children}</MantineProvider>
    </QueryClientProvider>
  );
}

describe("WorkspaceList", () => {
  it("renders workspace list", () => {
    render(<WorkspaceList />, { wrapper: Wrapper });

    expect(screen.getByText("あなたのワークスペース")).toBeInTheDocument();
    expect(screen.getByText("開発チーム")).toBeInTheDocument();
    expect(screen.getByText("開発用ワークスペース")).toBeInTheDocument();
    expect(screen.getByText("マーケティング")).toBeInTheDocument();
  });

  it("shows create button", () => {
    render(<WorkspaceList />, { wrapper: Wrapper });

    const createButton = screen.getByRole("button", { name: /新規作成/i });
    expect(createButton).toBeInTheDocument();
  });

  it("opens create modal when button clicked", async () => {
    const user = userEvent.setup();
    render(<WorkspaceList />, { wrapper: Wrapper });

    const createButton = screen.getByRole("button", { name: /新規作成/i });
    await user.click(createButton);

    await waitFor(() => {
      expect(screen.getByText("新規ワークスペース作成")).toBeInTheDocument();
    });
  });
});
