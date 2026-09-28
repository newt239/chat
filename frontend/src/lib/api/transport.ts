import { Code, ConnectError } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";

import { AuthService } from "#/gen/chat/v1/auth_service_pb";
import { refreshAccessToken } from "#/lib/api/client";
import { navigateTo } from "#/lib/navigation";
import { store } from "#/providers/store";
import { accessTokenAtom, clearAuthAtom } from "#/providers/store/auth";

import type { Interceptor } from "@connectrpc/connect";

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
  baseUrl: import.meta.env.VITE_API_BASE_URL || window.location.origin,
  interceptors: [authInterceptor],
});
