import { screen, waitFor } from "@testing-library/react";
import { userEvent } from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { AuthService } from "#/gen/chat/v1/auth_service_pb";
import { openExternal } from "#/lib/platform/openExternal";
import { renderWithProviders } from "#/test/renderWithProviders";

import { GoogleSignInButtonNative } from "./GoogleSignInButtonNative";

const deepLink = vi.hoisted(() => ({ handler: (_url: string) => {} }));

vi.mock("#/lib/platform/openExternal", () => ({ openExternal: vi.fn() }));
vi.mock("#/features/auth/utils/deepLink", () => ({
  listenDeepLinks: (handler: (url: string) => void) => {
    deepLink.handler = handler;
    return () => {};
  },
}));

describe("GoogleSignInButtonNative", () => {
  test("ブラウザでログインを始め、戻ってきた認可コードでログインする", async () => {
    const loginWithGoogleCode = vi.fn(() => ({ accessToken: "a", user: { id: "u1" } }));
    await renderWithProviders(
      <GoogleSignInButtonNative workspaceId={null} />,
      "/login",
      (routes) => {
        routes.rpc(AuthService.method.loginWithGoogleCode, loginWithGoogleCode);
      },
    );

    await userEvent.click(screen.getByRole("button", { name: "ブラウザで Google にログイン" }));
    await waitFor(() => {
      expect(openExternal).toHaveBeenCalledTimes(1);
    });
    const started = new URL(vi.mocked(openExternal).mock.calls[0]?.[0] ?? "");
    const state = started.searchParams.get("state") ?? "";

    deepLink.handler(`dev.newt239.chat://auth/callback?code=c1&state=${state}`);

    await waitFor(() => {
      expect(loginWithGoogleCode).toHaveBeenCalledWith(
        expect.objectContaining({ code: "c1", nonce: started.searchParams.get("nonce") }),
        expect.anything(),
      );
    });
  });

  test("state が合わない戻りでは失敗を表示する", async () => {
    await renderWithProviders(<GoogleSignInButtonNative workspaceId={null} />, "/login", () => {});

    await userEvent.click(screen.getByRole("button", { name: "ブラウザで Google にログイン" }));
    await waitFor(() => {
      expect(openExternal).toHaveBeenCalled();
    });
    deepLink.handler("dev.newt239.chat://auth/callback?code=c1&state=forged");

    expect(
      await screen.findByText("Google ログインを完了できませんでした。もう一度お試しください"),
    ).toBeInTheDocument();
  });
});
