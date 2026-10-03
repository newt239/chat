import { z } from "zod";

import { apiBaseUrl } from "#/lib/api/createTransport";

const STORAGE_KEY = "google-oauth-pending";
// ブラウザでの操作に時間がかかっても間に合い、古い要求は使わせない長さ
const PENDING_TTL_MS = 10 * 60_000;

// モバイルではブラウザにいる間にアプリが終了することがあるため、メモリではなく保存しておく
const pendingSchema = z.object({
  createdAt: z.number(),
  nonce: z.string(),
  state: z.string(),
  verifier: z.string(),
  workspaceId: z.string().nullable(),
});

const toBase64Url = (bytes: Uint8Array) =>
  btoa(String.fromCodePoint(...bytes))
    .replaceAll("+", "-")
    .replaceAll("/", "_")
    .replace(/=+$/, "");

const randomString = () => toBase64Url(crypto.getRandomValues(new Uint8Array(32)));

const sha256 = async (value: string) =>
  new Uint8Array(await crypto.subtle.digest("SHA-256", new TextEncoder().encode(value)));

/** 認可リクエストを作って保存し、ブラウザで開く URL を返す */
export const startGoogleOAuth = async (workspaceId: string | null) => {
  const pending = {
    createdAt: Date.now(),
    nonce: randomString(),
    state: randomString(),
    verifier: randomString(),
    workspaceId,
  };
  localStorage.setItem(STORAGE_KEY, JSON.stringify(pending));
  const params = new URLSearchParams({
    code_challenge: toBase64Url(await sha256(pending.verifier)),
    nonce: pending.nonce,
    state: pending.state,
  });
  return `${apiBaseUrl}/oauth/google/start?${params.toString()}`;
};

type GoogleOAuthResult =
  | {
      type: "success";
      code: string;
      codeVerifier: string;
      nonce: string;
      workspaceId: string | null;
    }
  | { type: "error" };

/** アプリに戻ってきた URL を保存した要求と照合する。ログインの戻りでないか処理済みなら null */
export const takeGoogleOAuthResult = (url: string): GoogleOAuthResult | null => {
  const { pathname, host, searchParams } = new URL(url);
  // カスタムスキームの URL はプラットフォームによって auth が host にも path にもなる
  if (`${host}${pathname}`.replace(/^\/+/, "") !== "auth/callback") {
    return null;
  }
  const raw = localStorage.getItem(STORAGE_KEY);
  // 起動のきっかけになったリンクは何度も届くため、要求がなければ処理済みとして無視する
  if (raw === null) {
    return null;
  }
  localStorage.removeItem(STORAGE_KEY);
  const pending = pendingSchema.safeParse(JSON.parse(raw)).data;
  const code = searchParams.get("code");
  if (
    pending === undefined ||
    code === null ||
    searchParams.get("state") !== pending.state ||
    Date.now() - pending.createdAt > PENDING_TTL_MS
  ) {
    return { type: "error" };
  }
  return {
    code,
    codeVerifier: pending.verifier,
    nonce: pending.nonce,
    type: "success",
    workspaceId: pending.workspaceId,
  };
};
