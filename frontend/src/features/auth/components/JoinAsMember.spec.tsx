import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { JoinAsMember } from "./JoinAsMember";

describe("JoinAsMember", () => {
  test("未参加ならログイン中のアカウントのまま参加する", async () => {
    const join = vi.fn(() => ({}));
    await renderWithProviders(<JoinAsMember workspaceId="ws1" />, "/app/other", (routes) => {
      routes.rpc(WorkspaceService.method.listWorkspaces, () => ({ workspaces: [] }));
      routes.rpc(WorkspaceService.method.joinPublicWorkspace, join);
    });

    expect(await screen.findByText("Alice としてログインしています。")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "参加する" }));
    await waitFor(() => {
      expect(join).toHaveBeenCalledWith(
        expect.objectContaining({ workspaceId: "ws1" }),
        expect.anything(),
      );
    });
  });

  test("参加済みならワークスペースを開くリンクを出す", async () => {
    await renderWithProviders(<JoinAsMember workspaceId="ws1" />, "/app/other", (routes) => {
      routes.rpc(WorkspaceService.method.listWorkspaces, () => ({ workspaces: [{ id: "ws1" }] }));
    });

    expect(await screen.findByText("すでに参加しています")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "ワークスペースを開く" })).toBeInTheDocument();
  });
});
