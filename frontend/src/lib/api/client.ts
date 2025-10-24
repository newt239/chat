import createClient from "openapi-fetch";

import type { paths } from "./schema";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "";

// リフレッシュ処理中かどうかを追跡するフラグ
let isRefreshing = false;
let refreshPromise: Promise<string | null> | null = null;
const retryableRequestMap = new WeakMap<Request, Request>();

export const apiClient = createClient<paths>({
  baseUrl: API_BASE_URL,
});

// リフレッシュトークンを使用してアクセストークンを更新する関数
async function refreshAccessToken(): Promise<string | null> {
  if (isRefreshing && refreshPromise) {
    return refreshPromise;
  }

  isRefreshing = true;
  refreshPromise = (async () => {
    try {
      const refreshToken = localStorage.getItem("refreshToken");
      if (!refreshToken) {
        return null;
      }

      const { data, error } = await apiClient.POST("/api/auth/refresh", {
        body: { refreshToken },
      });

      if (data && !error) {
        localStorage.setItem("accessToken", data.accessToken);
        return data.accessToken;
      }
      return null;
    } catch {
      return null;
    } finally {
      isRefreshing = false;
      refreshPromise = null;
    }
  })();

  return refreshPromise;
}

// リクエストインターセプター: アクセストークンを自動付与
apiClient.use({
  async onRequest({ request }) {
    const token = localStorage.getItem("accessToken");
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

    // 401エラーの場合のみリフレッシュを試みる
    if (
      response.status === 401 &&
      !request.url.includes("/api/auth/refresh") &&
      request.headers.get("X-Auth-Retry") !== "1"
    ) {
      const newToken = await refreshAccessToken();
      if (newToken) {
        // 新しいトークンで元のリクエストを再試行
        const sourceRequest = retrySource ?? request;
        const headers = new Headers(sourceRequest.headers);
        headers.set("Authorization", `Bearer ${newToken}`);
        headers.set("X-Auth-Retry", "1");

        const retryRequest = new Request(sourceRequest, { headers });

        return fetch(retryRequest);
      } else {
        // リフレッシュ失敗時は認証情報をクリアしてログイン画面へ
        localStorage.removeItem("accessToken");
        localStorage.removeItem("refreshToken");
        window.location.href = "/login";
      }
    }
    return response;
  },
});
