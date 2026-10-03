import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { InvitationService } from "#/gen/chat/v1/invitation_service_pb";
import { WorkspaceRole } from "#/gen/chat/v1/workspace_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { InviteMemberForm } from "./InviteMemberForm";

const setup = async (addedDirectly: boolean) => {
  const user = userEvent.setup();
  const create = vi.fn(() => ({ addedDirectly, token: addedDirectly ? "" : "secret-token" }));
  await renderWithProviders(<InviteMemberForm workspaceId="ws1" />, "/app/ws1", (routes) => {
    routes.rpc(InvitationService.method.createInvitation, create);
  });
  await user.type(screen.getByLabelText(/メールアドレスで招待/u), "new@example.com");
  await user.click(screen.getByRole("button", { name: "招待" }));
  return { create, user };
};

describe("InviteMemberForm", () => {
  test("未登録のメールアドレスには招待リンクを発行してコピーできる", async () => {
    const { create, user } = await setup(false);

    const link = await screen.findByText(/\/invite\/secret-token$/u);
    expect(create).toHaveBeenCalledWith(
      expect.objectContaining({
        email: "new@example.com",
        role: WorkspaceRole.MEMBER,
        workspaceId: "ws1",
      }),
      expect.anything(),
    );
    expect(screen.getByText("new@example.com への招待リンク")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "コピー" }));
    await waitFor(async () => {
      expect(await navigator.clipboard.readText()).toBe(link.textContent);
    });
  });

  test("登録済みのメールアドレスはリンクを出さずに追加する", async () => {
    await setup(true);

    await waitFor(() => {
      expect(screen.getByLabelText(/メールアドレスで招待/u)).toHaveValue("");
    });
    expect(screen.queryByText(/\/invite\//u)).not.toBeInTheDocument();
  });
});
