import createClient from "openapi-fetch";

import type { paths } from "./schema";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "";

export const apiClient = createClient<paths>({
  baseUrl: API_BASE_URL,
});

// リクエストインターセプター: アクセストークンを自動付与
apiClient.use({
  async onRequest({ request }) {
    const token = localStorage.getItem("accessToken");
    if (token) {
      request.headers.set("Authorization", `Bearer ${token}`);
    }
    return request;
  },
  async onResponse({ response }) {
    // 401エラーの場合はリフレッシュトークンで再認証を試みる
    if (response.status === 401) {
      const refreshToken = localStorage.getItem("refreshToken");
      if (refreshToken) {
        try {
          const { data, error } = await apiClient.POST("/api/auth/refresh", {
            body: { refreshToken },
          });
          if (data && !error) {
            localStorage.setItem("accessToken", data.accessToken);
            // 元のリクエストを再試行
            const retryRequest = new Request(response.url, {
              method: response.type,
              headers: response.headers,
            });
            retryRequest.headers.set("Authorization", `Bearer ${data.accessToken}`);
            return fetch(retryRequest);
          }
        } catch {
          // リフレッシュ失敗時はログアウト
          localStorage.removeItem("accessToken");
          localStorage.removeItem("refreshToken");
          window.location.href = "/login";
        }
      }
    }
    return response;
  },
});
