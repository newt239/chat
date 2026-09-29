import { screen } from "@testing-library/react";
import { afterEach, describe, expect, test, vi } from "vite-plus/test";

import { AuthService } from "#/gen/chat/v1/auth_service_pb";
import { renderWithProviders } from "#/test/renderWithProviders";

import { AuthMethods } from "./AuthMethods";

vi.mock("./GoogleSignInButton", () => ({
  GoogleSignInButton: ({ clientId }: { clientId: string }) => <p>Google: {clientId}</p>,
}));

const setup = (passwordAuthEnabled: boolean) =>
  renderWithProviders(
    <AuthMethods passwordForm={<p>パスワードのフォーム</p>} workspaceId={null} />,
    "/app/ws1",
    (routes) => {
      routes.rpc(AuthService.method.getAuthConfig, () => ({ passwordAuthEnabled }));
    },
  );

afterEach(() => {
  vi.unstubAllEnvs();
});

describe("AuthMethods", () => {
  test("Google とパスワードの両方が使えるときは区切りを挟んで並べる", async () => {
    vi.stubEnv("VITE_GOOGLE_OAUTH_CLIENT_ID", "client-1");
    await setup(true);

    expect(await screen.findByText("パスワードのフォーム")).toBeInTheDocument();
    expect(screen.getByText("Google: client-1")).toBeInTheDocument();
    expect(screen.getByText("または")).toBeInTheDocument();
  });

  test("パスワード認証が無効ならフォームを出さない", async () => {
    vi.stubEnv("VITE_GOOGLE_OAUTH_CLIENT_ID", "client-1");
    await setup(false);

    expect(await screen.findByText("Google: client-1")).toBeInTheDocument();
    expect(screen.queryByText("パスワードのフォーム")).not.toBeInTheDocument();
    expect(screen.queryByText("または")).not.toBeInTheDocument();
  });

  test("クライアント ID が未設定なら Google のボタンを出さない", async () => {
    vi.stubEnv("VITE_GOOGLE_OAUTH_CLIENT_ID", "");
    await setup(true);

    expect(await screen.findByText("パスワードのフォーム")).toBeInTheDocument();
    expect(screen.queryByText(/^Google:/u)).not.toBeInTheDocument();
    expect(screen.queryByText("または")).not.toBeInTheDocument();
  });
});
