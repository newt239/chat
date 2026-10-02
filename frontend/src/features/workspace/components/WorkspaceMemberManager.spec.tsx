import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { WorkspaceRole, WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { WorkspaceMemberManager } from "./WorkspaceMemberManager";

import type { RemoveMemberRequest } from "#/gen/chat/v1/workspace_service_pb";

describe("WorkspaceMemberManager", () => {
  test("メンバーを外す前に確認し、確定したときだけ外す", async () => {
    const removeMember = vi.fn<(req: RemoveMemberRequest) => void>();
    await renderWithProviders(
      <WorkspaceMemberManager workspaceId="ws1" canManage />,
      "/app/ws1",
      (routes) => {
        routes.rpc(WorkspaceService.method.listMembers, () => ({
          members: [{ displayName: "Bob", role: WorkspaceRole.MEMBER, userId: "u-bob" }],
        }));
        routes.rpc(WorkspaceService.method.removeMember, (req) => {
          removeMember(req);
          return {};
        });
      },
    );

    await userEvent.click(
      await screen.findByRole("button", { name: "Bob をワークスペースから外す" }),
    );
    const dialog = screen.getByRole("alertdialog", {
      name: "Bob をワークスペースから外しますか？",
    });
    expect(removeMember).not.toHaveBeenCalled();

    await userEvent.click(within(dialog).getByRole("button", { name: "外す" }));
    await waitFor(() => {
      expect(removeMember).toHaveBeenCalledWith(
        expect.objectContaining({ userId: "u-bob", workspaceId: "ws1" }),
      );
    });
  });
});
