import { Code, ConnectError, createClient } from "@connectrpc/connect";

import { AuthService } from "#/gen/chat/v1/auth_service_pb";
import { createTransport } from "#/lib/api/createTransport";
import { navigateTo } from "#/lib/navigation";
import { queryClient } from "#/providers/query/query";
import { sessionAtom } from "#/providers/store/auth";
import { store } from "#/providers/store/store";

import type { User } from "#/gen/chat/v1/user_pb";

// ネイティブアプリだけが本文でリフレッシュトークンを受け取り、ここに保存する
const refreshTokenKey = "refresh-token";

// 認証 interceptor を通すと refresh が再帰するため、refresh 専用の client を分ける
const authClient = createClient(AuthService, createTransport([]));

const authChannel = new BroadcastChannel("auth");

type AuthResponse = {
  accessToken: string;
  refreshToken?: string;
  user?: User;
};

/** ログイン系と Refresh の応答からセッションを始める */
export const startSession = ({ accessToken, refreshToken, user }: AuthResponse) => {
  if (user === undefined) {
    throw new ConnectError("ユーザー情報のない認証応答です", Code.Internal);
  }
  if (refreshToken !== undefined) {
    localStorage.setItem(refreshTokenKey, refreshToken);
  }
  store.set(sessionAtom, { accessToken, userId: user.id });
};

let refreshing: Promise<string> | null = null;

/** アクセストークンを取り直す。ローテーションが競合しないよう、タブをまたいで直列にする */
const refreshSession = () => {
  refreshing ??= navigator.locks
    .request("chat-refresh", async () => {
      const response = await authClient.refresh({
        refreshToken: localStorage.getItem(refreshTokenKey) ?? undefined,
      });
      startSession(response);
      return response.accessToken;
    })
    .finally(() => {
      refreshing = null;
    });
  return refreshing;
};

const clearSession = () => {
  localStorage.removeItem(refreshTokenKey);
  // ログイン画面で読み込み中のクエリまで捨てるとフォームが出なくなるため、ログアウト済みなら何もしない
  if (store.get(sessionAtom) === null) {
    return;
  }
  store.set(sessionAtom, null);
  queryClient.clear();
  navigateTo({ to: "/login" });
};

authChannel.addEventListener("message", clearSession);

/** 認証情報とキャッシュを捨ててログイン画面へ戻す。ほかのタブもログアウトさせる */
export const signOut = () => {
  clearSession();
  authChannel.postMessage("signOut");
};

/** Refresh を試し、認証が切れていたときだけログアウトする。ほかの失敗は投げ直す */
export const refreshOrSignOut = async () => {
  try {
    return await refreshSession();
  } catch (error) {
    if (ConnectError.from(error).code === Code.Unauthenticated) {
      signOut();
    }
    throw error;
  }
};

/** 起動直後はメモリにトークンがないため、Cookie で取り直せるか確かめる */
export const ensureSession = async () => {
  if (store.get(sessionAtom) !== null) {
    return true;
  }
  try {
    await refreshSession();
    return true;
  } catch (error) {
    if (ConnectError.from(error).code === Code.Unauthenticated) {
      return false;
    }
    throw error;
  }
};
