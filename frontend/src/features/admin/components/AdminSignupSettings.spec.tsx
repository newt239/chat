import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { AuthService } from "#/gen/chat/v1/auth_service_pb";
import { WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { AdminSignupSettings } from "./AdminSignupSettings";

import type { UpdateWorkspaceRequest } from "#/gen/chat/v1/workspace_service_pb";

const setup = async (signupEnabled: boolean, passwordAuthEnabled: boolean) => {
  const update = vi.fn<(req: UpdateWorkspaceRequest) => void>();
  await renderWithProviders(
    <AdminSignupSettings workspaceId="ws1" />,
    "/app/ws1/admin",
    (routes) => {
      routes.rpc(WorkspaceService.method.getWorkspace, () => ({
        workspace: { emailSignupEnabled: false, id: "ws1", signupEnabled },
      }));
      routes.rpc(AuthService.method.getAuthConfig, () => ({ passwordAuthEnabled }));
      routes.rpc(WorkspaceService.method.updateWorkspace, (req) => {
        update(req);
        return {};
      });
    },
  );
  return { update };
};

describe("AdminSignupSettings", () => {
  test("新規登録を許可すると更新し、許可していなければメールでの登録は切り替えられない", async () => {
    const { update } = await setup(false, true);

    const email = await screen.findByRole("switch", { name: /メールアドレスとパスワード/u });
    expect(email).toBeDisabled();
    expect(screen.queryByText("参加リンク")).not.toBeInTheDocument();

    await userEvent.click(screen.getByRole("switch", { name: "参加リンクからの新規登録を許可" }));
    await waitFor(() => {
      expect(update).toHaveBeenCalledWith(
        expect.objectContaining({ signupEnabled: true, workspaceId: "ws1" }),
      );
    });
  });

  test("許可していれば参加リンクを表示し、メールでの登録を切り替えられる", async () => {
    const { update } = await setup(true, true);

    expect(await screen.findByText(/\/join\/ws1$/u)).toBeInTheDocument();
    await userEvent.click(screen.getByRole("switch", { name: /メールアドレスとパスワード/u }));
    await waitFor(() => {
      expect(update).toHaveBeenCalledWith(expect.objectContaining({ emailSignupEnabled: true }));
    });
  });

  test("パスワード認証が無効ならメールでの登録は切り替えられない", async () => {
    await setup(true, false);

    expect(await screen.findByText(/パスワード認証が無効なため/u)).toBeInTheDocument();
    expect(screen.getByRole("switch", { name: /メールアドレスとパスワード/u })).toBeDisabled();
  });
});
