import createClient from "openapi-fetch";

import { refreshAccessToken } from "#/lib/api/transport";
import { navigateTo } from "#/lib/navigation";
import { store } from "#/providers/store";
import { accessTokenAtom, clearAuthAtom } from "#/providers/store/auth";

import type { paths } from "./schema";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "";

const retryableRequestMap = new WeakMap<Request, Request>();

const getAccessToken = () => store.get(accessTokenAtom);

export const api = createClient<paths>({
  baseUrl: API_BASE_URL,
});

// 401時の再試行用リクエスト生成
const buildRetriedRequest = (source: Request, token: string) => {
  const headers = new Headers(source.headers);
  headers.set("Authorization", `Bearer ${token}`);
  headers.set("X-Auth-Retry", "1");
  return new Request(source, { headers });
};

// リクエストインターセプター: アクセストークンを自動付与
api.use({
  onRequest({ request }) {
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

    const hasRetried = request.headers.get("X-Auth-Retry") === "1";
    if (response.status !== 401 || hasRetried) {
      return response;
    }

    const newToken = await refreshAccessToken();
    if (newToken) {
      const sourceRequest = retrySource ?? request;
      const retryRequest = buildRetriedRequest(sourceRequest, newToken);
      return fetch(retryRequest);
    }

    // リフレッシュ失敗時は認証情報をクリアしてログイン画面へ
    store.set(clearAuthAtom);
    navigateTo({ to: "/login" });
    return response;
  },
});
