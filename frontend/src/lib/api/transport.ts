import { Code, ConnectError } from "@connectrpc/connect";

import { AuthService } from "#/gen/chat/v1/auth_service_pb";
import { refreshOrSignOut } from "#/lib/session";
import { sessionAtom } from "#/providers/store/auth";
import { store } from "#/providers/store/store";

import { createTransport } from "./createTransport";

import type { Interceptor } from "@connectrpc/connect";

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
  const token = store.get(sessionAtom)?.accessToken;
  if (token !== undefined) {
    req.header.set("Authorization", `Bearer ${token}`);
  }
  try {
    return await next(req);
  } catch (error) {
    // AuthService は資格情報の誤りでも Unauthenticated を返すため再試行しない
    if (
      req.service.typeName === AuthService.typeName ||
      ConnectError.from(error).code !== Code.Unauthenticated
    ) {
      throw error;
    }
    req.header.set("Authorization", `Bearer ${await refreshOrSignOut()}`);
    return next(req);
  }
};

export const transport = createTransport([displayErrorInterceptor, authInterceptor]);
