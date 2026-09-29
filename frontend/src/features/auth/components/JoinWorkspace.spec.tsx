import { Code, ConnectError } from "@connectrpc/connect";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, test, vi } from "vite-plus/test";

import { AuthService } from "#/gen/chat/v1/auth_service_pb";
import { WorkspaceService } from "#/gen/chat/v1/workspace_service_pb";
import { authAtom } from "#/providers/store/auth";
import { renderWithProviders } from "#/test/renderWithProviders";

import { JoinWorkspace } from "./JoinWorkspace";

import type { ConnectRouter } from "@connectrpc/connect";

afterEach(() => {
  vi.unstubAllEnvs();
});

const renderLoggedOut = async (routes: (router: ConnectRouter) => void) => {
  vi.stubEnv("VITE_GOOGLE_OAUTH_CLIENT_ID", "");
  const { store } = await renderWithProviders(
    <JoinWorkspace workspaceId="ws1" />,
    "/app/ws1",
    routes,
  );
  store.set(authAtom, { accessToken: null, refreshToken: null, user: null });
};

describe("JoinWorkspace", () => {
  test("未ログインならメールアドレスでアカウントを作って参加できる", async () => {
    const signUp = vi.fn(() => ({ accessToken: "a", refreshToken: "r" }));
    await renderLoggedOut((routes) => {
      routes.rpc(WorkspaceService.method.getWorkspaceSignupInfo, () => ({
        emailSignupEnabled: true,
        id: "ws1",
        name: "Acme",
      }));
      routes.rpc(AuthService.method.getAuthConfig, () => ({ passwordAuthEnabled: true }));
      routes.rpc(AuthService.method.signUp, signUp);
    });

    expect(await screen.findByRole("heading", { name: "Acme に参加" })).toBeInTheDocument();
    await userEvent.type(await screen.findByLabelText(/メールアドレス/u), "new@example.com");
    await userEvent.type(screen.getByLabelText(/表示名/u), "New User");
    await userEvent.type(screen.getByLabelText(/パスワード/u), "password123");
    await userEvent.click(screen.getByRole("button", { name: "アカウントを作って参加" }));
    await waitFor(() => {
      expect(signUp).toHaveBeenCalledWith(
        expect.objectContaining({
          displayName: "New User",
          email: "new@example.com",
          password: "password123",
          workspaceId: "ws1",
        }),
        expect.anything(),
      );
    });
  });

  test("メールでの登録を許可していなければフォームを出さない", async () => {
    await renderLoggedOut((routes) => {
      routes.rpc(WorkspaceService.method.getWorkspaceSignupInfo, () => ({
        emailSignupEnabled: false,
        id: "ws1",
        name: "Acme",
      }));
      routes.rpc(AuthService.method.getAuthConfig, () => ({ passwordAuthEnabled: true }));
    });

    expect(await screen.findByText(/Google アカウントでアカウントを作って/u)).toBeInTheDocument();
    expect(screen.queryByLabelText(/パスワード/u)).not.toBeInTheDocument();
  });

  test("新規登録を受け付けていなければその旨を表示する", async () => {
    await renderLoggedOut((routes) => {
      routes.rpc(WorkspaceService.method.getWorkspaceSignupInfo, () => {
        throw new ConnectError("not found", Code.NotFound);
      });
    });

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "参加リンクが見つからないか、新規登録を受け付けていません",
    );
  });
});
