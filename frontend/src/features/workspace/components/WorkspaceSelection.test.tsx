import { MantineProvider } from "@mantine/core";
import { QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";

import { WorkspaceSelection } from "./WorkspaceSelection";

import { queryClient } from "@/lib/query";

// Mock WorkspaceList
vi.mock("./WorkspaceList", () => ({
  WorkspaceList: () => <div data-testid="workspace-list">Workspace List</div>,
}));

// Mock console.log to test it's called
const consoleSpy = vi.spyOn(console, "log").mockImplementation(() => {});

const Wrapper = ({ children }: { children: React.ReactNode }) => {
  return (
    <QueryClientProvider client={queryClient}>
      <MantineProvider>{children}</MantineProvider>
    </QueryClientProvider>
  );
};

describe("WorkspaceSelection", () => {
  it("renders title and description", () => {
    render(<WorkspaceSelection />, { wrapper: Wrapper });

    expect(screen.getByText("ワークスペースを選択してください")).toBeInTheDocument();
    expect(
      screen.getByText(
        "参加しているワークスペースから選択するか、新しいワークスペースを作成してください。"
      )
    ).toBeInTheDocument();
  });

  it("renders WorkspaceList component", () => {
    render(<WorkspaceSelection />, { wrapper: Wrapper });

    expect(screen.getByTestId("workspace-list")).toBeInTheDocument();
  });

  it("logs component render", () => {
    render(<WorkspaceSelection />, { wrapper: Wrapper });

    expect(consoleSpy).toHaveBeenCalledWith(
      "WorkspaceSelection - コンポーネントがレンダリングされました"
    );
  });

  it("renders with correct structure", () => {
    render(<WorkspaceSelection />, { wrapper: Wrapper });

    expect(screen.getByTestId("workspace-list")).toBeInTheDocument();
    expect(screen.getByRole("heading", { level: 1 })).toBeInTheDocument();
  });

  it("has correct heading structure", () => {
    render(<WorkspaceSelection />, { wrapper: Wrapper });

    const heading = screen.getByRole("heading", { level: 1 });
    expect(heading).toHaveTextContent("ワークスペースを選択してください");
    expect(heading).toHaveClass("text-2xl", "font-bold", "text-gray-900", "mb-4");
  });
});
