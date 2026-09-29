import { create } from "@bufbuild/protobuf";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { InvitationSchema, InvitationService } from "#/gen/chat/v1/invitation_service_pb";
import { WorkspaceRole } from "#/gen/chat/v1/workspace_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { AdminInvitationsTab } from "./AdminInvitationsTab";

describe("AdminInvitationsTab", () => {
  test("保留中の招待を一覧し、取り消せる", async () => {
    const revoke = vi.fn(() => ({}));
    await renderWithProviders(<AdminInvitationsTab workspaceId="ws1" />, "/app/ws1", (routes) => {
      routes.rpc(InvitationService.method.listInvitations, () => ({
        invitations: [
          create(InvitationSchema, {
            email: "new@example.com",
            id: "inv1",
            invitedByName: "Alice",
            role: WorkspaceRole.GUEST,
          }),
        ],
      }));
      routes.rpc(InvitationService.method.revokeInvitation, revoke);
    });

    expect(await screen.findByText("new@example.com")).toBeInTheDocument();
    expect(screen.getByRole("cell", { name: "ゲスト" })).toBeInTheDocument();
    expect(screen.getByText("Alice")).toBeInTheDocument();

    await userEvent.click(
      screen.getByRole("button", { name: "new@example.com への招待を取り消す" }),
    );
    await waitFor(() => {
      expect(revoke).toHaveBeenCalledWith(
        expect.objectContaining({ invitationId: "inv1", workspaceId: "ws1" }),
        expect.anything(),
      );
    });
  });

  test("招待がなければその旨を表示する", async () => {
    await renderWithProviders(<AdminInvitationsTab workspaceId="ws1" />, "/app/ws1", (routes) => {
      routes.rpc(InvitationService.method.listInvitations, () => ({ invitations: [] }));
    });

    expect(await screen.findByText("保留中の招待はありません")).toBeInTheDocument();
  });
});
