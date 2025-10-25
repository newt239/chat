import { MantineProvider } from "@mantine/core";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import { describe, it, beforeEach, vi, expect } from "vitest";

import { WorkspaceSelection } from "./WorkspaceSelection";

import type { WorkspaceSummary } from "@/features/workspace/types";

const mockNavigateToWorkspace = vi.fn();

vi.mock("@/lib/navigation", () => ({
  navigateToWorkspace: mockNavigateToWorkspace,
}));

const mockUseWorkspaces = vi.fn();

vi.mock("@/features/workspace/hooks/useWorkspace", () => ({
  useWorkspaces: () => mockUseWorkspaces(),
}));

vi.mock("./WorkspaceList", () => ({
  WorkspaceList: () => <div data-testid="workspace-list">Workspace List</div>,
}));

const createWorkspace = (id: string): WorkspaceSummary => ({
  id,
  name: `Workspace ${id}`,
  description: null,
  iconUrl: null,
  createdBy: "user-1",
  createdAt: "2024-01-01T00:00:00.000Z",
});

const renderComponent = (client: QueryClient) => {
  return render(
    <QueryClientProvider client={client}>
      <MantineProvider>
        <WorkspaceSelection />
      </MantineProvider>
    </QueryClientProvider>
  );
};

describe("WorkspaceSelection", () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
  });

  it("既定のワークスペースがなく一覧も空の場合はリダイレクトしない", async () => {
    const queryClient = new QueryClient();
    mockUseWorkspaces.mockReturnValue({
      data: [],
      isLoading: false,
      error: null,
    });

    renderComponent(queryClient);

    expect(
      screen.getByText("ワークスペースを選択してください")
    ).toBeInTheDocument();

    await waitFor(() => {
      expect(mockNavigateToWorkspace).not.toHaveBeenCalled();
    });
  });

  it("ローカルストレージに既定のワークスペースがある場合はそのワークスペースへリダイレクトする", async () => {
    const queryClient = new QueryClient();
    const storedWorkspaceId = "workspace-123";
    localStorage.setItem(
      "workspace-storage",
      JSON.stringify({ state: { currentWorkspaceId: storedWorkspaceId } })
    );

    mockUseWorkspaces.mockReturnValue({
      data: [createWorkspace(storedWorkspaceId)],
      isLoading: false,
      error: null,
    });

    renderComponent(queryClient);

    await waitFor(() => {
      expect(mockNavigateToWorkspace).toHaveBeenCalledWith(storedWorkspaceId);
    });
  });

  it("既定のワークスペースが無く配列の1件目へリダイレクトする", async () => {
    const queryClient = new QueryClient();
    const firstWorkspaceId = "workspace-001";
    mockUseWorkspaces.mockReturnValue({
      data: [createWorkspace(firstWorkspaceId), createWorkspace("workspace-002")],
      isLoading: false,
      error: null,
    });

    renderComponent(queryClient);

    await waitFor(() => {
      expect(mockNavigateToWorkspace).toHaveBeenCalledWith(firstWorkspaceId);
    });
  });
});
