import { Code, ConnectError } from "@connectrpc/connect";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, test, vi } from "vite-plus/test";

import { AuthService } from "#/gen/chat/v1/auth_service_pb";
import { InvitationService } from "#/gen/chat/v1/invitation_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { InvitationAccept } from "./InvitationAccept";

afterEach(() => {
  vi.unstubAllEnvs();
});

describe("InvitationAccept", () => {
  test("招待先を表示し、パスワードを設定して参加できる", async () => {
    vi.stubEnv("VITE_GOOGLE_OAUTH_CLIENT_ID", "");
    const signUp = vi.fn(() => ({ accessToken: "a", user: { id: "u1" } }));
    await renderWithProviders(<InvitationAccept token="t1" />, "/app/ws1", (routes) => {
      routes.rpc(InvitationService.method.getInvitation, ({ token }) => {
        expect(token).toBe("t1");
        return { email: "new@example.com", workspaceName: "Acme" };
      });
      routes.rpc(AuthService.method.getAuthConfig, () => ({ passwordAuthEnabled: true }));
      routes.rpc(AuthService.method.signUpWithInvitation, signUp);
    });

    expect(await screen.findByRole("heading", { name: "Acme への招待" })).toBeInTheDocument();
    expect(screen.getByText(/new@example.com 宛ての招待です/u)).toBeInTheDocument();

    await userEvent.type(await screen.findByLabelText(/表示名/u), "New User");
    await userEvent.type(screen.getByLabelText(/パスワード/u), "password123");
    await userEvent.click(screen.getByRole("button", { name: "パスワードを設定して参加" }));
    await waitFor(() => {
      expect(signUp).toHaveBeenCalledWith(
        expect.objectContaining({ displayName: "New User", password: "password123", token: "t1" }),
        expect.anything(),
      );
    });
  });

  test("無効な招待はその旨を表示する", async () => {
    await renderWithProviders(<InvitationAccept token="bad" />, "/app/ws1", (routes) => {
      routes.rpc(InvitationService.method.getInvitation, () => {
        throw new ConnectError("not found", Code.NotFound);
      });
    });

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "招待が見つからないか、有効期限が切れています",
    );
  });
});
