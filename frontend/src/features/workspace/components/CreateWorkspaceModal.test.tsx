import { MantineProvider } from "@mantine/core";
import { QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi } from "vitest";

import { CreateWorkspaceModal } from "./CreateWorkspaceModal";

import { queryClient } from "@/lib/query";

// Mock the hook
vi.mock("../hooks/useWorkspace", () => ({
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

describe("CreateWorkspaceModal", () => {
  it("renders modal when opened", () => {
    render(<CreateWorkspaceModal opened={true} onClose={vi.fn()} />, { wrapper: Wrapper });

    expect(screen.getByText("新規ワークスペース作成")).toBeInTheDocument();
    expect(screen.getByLabelText(/ワークスペース名/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/説明/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /作成/i })).toBeInTheDocument();
  });

  it("does not render when closed", () => {
    render(<CreateWorkspaceModal opened={false} onClose={vi.fn()} />, { wrapper: Wrapper });

    expect(screen.queryByText("新規ワークスペース作成")).not.toBeInTheDocument();
  });

  it("allows input in form fields", async () => {
    const user = userEvent.setup();
    render(<CreateWorkspaceModal opened={true} onClose={vi.fn()} />, { wrapper: Wrapper });

    const nameInput = screen.getByLabelText(/ワークスペース名/i);
    const descriptionInput = screen.getByLabelText(/説明/i);

    await user.type(nameInput, "新しいワークスペース");
    await user.type(descriptionInput, "これはテストです");

    expect(nameInput).toHaveValue("新しいワークスペース");
    expect(descriptionInput).toHaveValue("これはテストです");
  });
});
