import createClient from "openapi-fetch";

import type { paths } from "./schema";

import { router } from "@/lib/router";
import { store } from "@/providers/store";
import { accessTokenAtom, authAtom, clearAuthAtom, refreshTokenAtom } from "@/providers/store/auth";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "";

// リフレッシュ処理の単一フライト制御
let refreshPromise: Promise<string | null> | null = null;
const retryableRequestMap = new WeakMap<Request, Request>();

const getAccessToken = () => store.get(accessTokenAtom);
const getRefreshToken = () => store.get(refreshTokenAtom);

const updateAuthTokens = (accessToken: string, refreshToken?: string) => {
  const current = store.get(authAtom);
  store.set(authAtom, {
    user: current.user,
    accessToken,
    refreshToken: refreshToken ?? current.refreshToken,
  });
};

const resetAuthState = () => {
  store.set(clearAuthAtom);
};

export const api = createClient<paths>({
  baseUrl: API_BASE_URL,
});

// リフレッシュトークンを使用してアクセストークンを更新する関数
async function refreshAccessToken(): Promise<string | null> {
  if (refreshPromise) return refreshPromise;

  refreshPromise = (async () => {
    const refreshToken = getRefreshToken();
    if (!refreshToken) return null;
    try {
      const { data, error } = await api.POST("/api/auth/refresh", {
        body: { refreshToken },
      });
      if (data && !error) {
        updateAuthTokens(data.accessToken);
        return data.accessToken;
      }
      return null;
    } catch {
      return null;
    } finally {
      refreshPromise = null;
    }
  })();

  return refreshPromise;
}

// 401時の再試行用リクエスト生成
const buildRetriedRequest = (source: Request, token: string) => {
  const headers = new Headers(source.headers);
  headers.set("Authorization", `Bearer ${token}`);
  headers.set("X-Auth-Retry", "1");
  return new Request(source, { headers });
};

// リクエストインターセプター: アクセストークンを自動付与
api.use({
  async onRequest({ request }) {
    const token = getAccessToken();
    if (token) {
      request.headers.set("Authorization", `Bearer ${token}`);
    }

    try {
      const cloned = request.clone();
      retryableRequestMap.set(request, cloned);
    } catch {
      retryableRequestMap.delete(request);
    }
    return request;
  },
  async onResponse({ response, request }) {
    const retrySource = retryableRequestMap.get(request);
    retryableRequestMap.delete(request);

    const isRefreshEndpoint = request.url.includes("/api/auth/refresh");
    const hasRetried = request.headers.get("X-Auth-Retry") === "1";
    if (response.status !== 401 || isRefreshEndpoint || hasRetried) {
      return response;
    }

    const newToken = await refreshAccessToken();
    if (newToken) {
      const sourceRequest = retrySource ?? request;
      const retryRequest = buildRetriedRequest(sourceRequest, newToken);
      return fetch(retryRequest);
    }

    // リフレッシュ失敗時は認証情報をクリアしてログイン画面へ
    resetAuthState();
    router.navigate({ to: "/login" });
    return response;
  },
});
