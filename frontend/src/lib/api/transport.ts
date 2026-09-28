import { Code, ConnectError, createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";

import { AuthService } from "#/gen/chat/v1/auth_service_pb";
import { navigateTo } from "#/lib/navigation";
import { store } from "#/providers/store";
import { accessTokenAtom, authAtom, clearAuthAtom } from "#/providers/store/auth";

import { apiBaseUrl as baseUrl } from "./baseUrl";

import type { Interceptor } from "@connectrpc/connect";

// 認証 interceptor を通すと 401 時に refresh が再帰するため、refresh 専用の client を分ける
const authClient = createClient(AuthService, createConnectTransport({ baseUrl }));

let refreshPromise: Promise<string | null> | null = null;

/** リフレッシュトークンでアクセストークンを更新する。同時に呼ばれても通信は 1 回にまとめる */
const refreshAccessToken = () => {
  if (refreshPromise) {
    return refreshPromise;
  }
  refreshPromise = (async () => {
    const { refreshToken, user } = store.get(authAtom);
    if (!refreshToken) {
      return null;
    }
    try {
      const response = await authClient.refresh({ refreshToken });
      store.set(authAtom, {
        accessToken: response.accessToken,
        refreshToken: response.refreshToken,
        user,
      });
      return response.accessToken;
    } catch {
      return null;
    } finally {
      refreshPromise = null;
    }
  })();
  return refreshPromise;
};

// ConnectError の message には "[not_found]" のようなコードが前置されるため、画面表示用にサーバーのメッセージだけを残す
const displayErrorInterceptor: Interceptor = (next) => async (req) => {
  try {
    return await next(req);
  } catch (error) {
    const connectError = ConnectError.from(error);
    connectError.message = connectError.rawMessage;
    throw connectError;
  }
};

const authInterceptor: Interceptor = (next) => async (req) => {
  const token = store.get(accessTokenAtom);
  if (token) {
    req.header.set("Authorization", `Bearer ${token}`);
  }
  try {
    return await next(req);
  } catch (error) {
    // AuthService は資格情報の誤りでも Unauthenticated を返すため再試行しない
    if (req.service.typeName === AuthService.typeName) {
      throw error;
    }
    if (ConnectError.from(error).code !== Code.Unauthenticated) {
      throw error;
    }
    const newToken = await refreshAccessToken();
    if (newToken === null) {
      store.set(clearAuthAtom);
      navigateTo({ to: "/login" });
      throw error;
    }
    req.header.set("Authorization", `Bearer ${newToken}`);
    return next(req);
  }
};

export const transport = createConnectTransport({
  baseUrl,
  interceptors: [displayErrorInterceptor, authInterceptor],
});
