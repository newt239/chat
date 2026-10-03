import { createConnectTransport } from "@connectrpc/connect-web";

import { isTauri } from "#/lib/platform/platform";

import type { Interceptor } from "@connectrpc/connect";

export const apiBaseUrl = import.meta.env.VITE_API_BASE_URL;

// リフレッシュトークンの Cookie を送受信するため
const fetchWithCredentials: typeof fetch = (input, init) =>
  fetch(input, { ...init, credentials: "include" });

// ネイティブアプリは Cookie を持ち回れないため、リフレッシュトークンを本文で受け渡すよう伝える
const nativeClientInterceptor: Interceptor = (next) => (req) => {
  req.header.set("X-Chat-Client", "native");
  return next(req);
};

export const createTransport = (interceptors: Interceptor[]) =>
  createConnectTransport({
    baseUrl: apiBaseUrl,
    fetch: fetchWithCredentials,
    interceptors: isTauri ? [nativeClientInterceptor, ...interceptors] : interceptors,
  });
