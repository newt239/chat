import { screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, test, vi } from "vite-plus/test";

import { loadGoogleIdentity } from "#/features/auth/utils/googleIdentity";
import { AuthService } from "#/gen/chat/v1/auth_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { GoogleSignInButton } from "./GoogleSignInButton";

vi.mock("#/features/auth/utils/googleIdentity", () => ({ loadGoogleIdentity: vi.fn() }));

const setupGoogle = () => {
  const configs: google.accounts.id.IdConfiguration[] = [];
  const renderButton = vi.fn();
  vi.stubGlobal("google", {
    accounts: {
      id: {
        initialize: (config: google.accounts.id.IdConfiguration) => {
          configs.push(config);
        },
        renderButton,
      },
    },
  });
  return { configs, renderButton };
};

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("GoogleSignInButton", () => {
  test("Google のボタンを描画し、受け取った ID トークンでログインする", async () => {
    vi.mocked(loadGoogleIdentity).mockResolvedValue();
    const { configs, renderButton } = setupGoogle();
    const loginWithGoogle = vi.fn(() => ({ accessToken: "a", user: { id: "u1" } }));
    await renderWithProviders(
      <GoogleSignInButton clientId="client-1" workspaceId={null} />,
      "/app/ws1",
      (routes) => {
        routes.rpc(AuthService.method.loginWithGoogle, loginWithGoogle);
      },
    );

    await waitFor(() => {
      expect(renderButton).toHaveBeenCalledTimes(1);
    });
    expect(configs[0]?.client_id).toBe("client-1");

    configs[0]?.callback?.({ credential: "id-token", select_by: "btn" });
    await waitFor(() => {
      expect(loginWithGoogle).toHaveBeenCalledWith(
        expect.objectContaining({ idToken: "id-token" }),
        expect.anything(),
      );
    });
  });

  test("スクリプトを読み込めなければその旨を表示する", async () => {
    vi.mocked(loadGoogleIdentity).mockRejectedValue(new Error("offline"));
    await renderWithProviders(
      <GoogleSignInButton clientId="client-1" workspaceId={null} />,
      "/app/ws1",
      () => {},
    );

    expect(await screen.findByText("Google ログインを読み込めませんでした")).toBeInTheDocument();
  });
});
