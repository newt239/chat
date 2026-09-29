import { beforeEach, describe, expect, test } from "vite-plus/test";

import { startGoogleOAuth, takeGoogleOAuthResult } from "#/features/auth/utils/googleOAuthPkce";

const CALLBACK = "dev.newt239.chat://auth/callback";

const start = async (workspaceId: string | null = null) => {
  const url = new URL(await startGoogleOAuth(workspaceId));
  return url.searchParams.get("state") ?? "";
};

beforeEach(() => {
  localStorage.clear();
});

describe("startGoogleOAuth", () => {
  test("S256 の code_challenge・state・nonce を付けた開始 URL を返す", async () => {
    const url = new URL(await startGoogleOAuth(null));

    expect(url.pathname).toBe("/oauth/google/start");
    expect(url.searchParams.get("code_challenge")).toMatch(/^[\w-]{43}$/);
    expect(url.searchParams.get("state")).toMatch(/^[\w-]{43}$/);
    expect(url.searchParams.get("nonce")).toMatch(/^[\w-]{43}$/);
  });
});

describe("takeGoogleOAuthResult", () => {
  test("state が一致すれば code_verifier と nonce を返し、要求を消す", async () => {
    const state = await start("ws1");

    const result = takeGoogleOAuthResult(`${CALLBACK}?code=c1&state=${state}`);

    expect(result).toMatchObject({ code: "c1", type: "success", workspaceId: "ws1" });
    expect(takeGoogleOAuthResult(`${CALLBACK}?code=c1&state=${state}`)).toBeNull();
  });

  test("state が一致しなければ失敗にする", async () => {
    await start();

    expect(takeGoogleOAuthResult(`${CALLBACK}?code=c1&state=forged`)).toEqual({ type: "error" });
  });

  test("Google でキャンセルされたら失敗にする", async () => {
    const state = await start();

    expect(takeGoogleOAuthResult(`${CALLBACK}?error=access_denied&state=${state}`)).toEqual({
      type: "error",
    });
  });

  test("ログインの戻り以外の URL は無視する", async () => {
    await start();

    expect(takeGoogleOAuthResult("dev.newt239.chat://app/ws1")).toBeNull();
  });
});
