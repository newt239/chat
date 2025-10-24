import { atom } from "jotai";
import { atomWithStorage } from "jotai/utils";

import { navigateToAppWithWorkspace, navigateToLogin } from "../navigation";

import type { components } from "@/lib/api/schema";

type User = components["schemas"]["User"];

interface AuthStorage {
  user: User | null;
  isAuthenticated: boolean;
}

// LocalStorageからトークンを読み込む関数
const loadTokensFromStorage = () => {
  return {
    accessToken: localStorage.getItem("accessToken"),
    refreshToken: localStorage.getItem("refreshToken"),
  };
};

// 認証情報をストレージに保存（userとisAuthenticatedのみ）
export const authStorageAtom = atomWithStorage<AuthStorage>(
  "auth-storage",
  {
    user: null,
    isAuthenticated: false,
  },
  undefined,
  { getOnInit: true }
);

// アクセストークン（LocalStorageから動的に取得）
export const accessTokenAtom = atom<string | null>((get) => {
  const storage = get(authStorageAtom);
  if (!storage.isAuthenticated) {
    return null;
  }
  return localStorage.getItem("accessToken");
});

// リフレッシュトークン（LocalStorageから動的に取得）
export const refreshTokenAtom = atom<string | null>((get) => {
  const storage = get(authStorageAtom);
  if (!storage.isAuthenticated) {
    return null;
  }
  return localStorage.getItem("refreshToken");
});

// ユーザー情報
export const userAtom = atom<User | null>((get) => get(authStorageAtom).user);

// 認証状態
export const isAuthenticatedAtom = atom<boolean>(
  (get) => get(authStorageAtom).isAuthenticated
);

// 認証情報を設定
export const setAuthAtom = atom(
  null,
  (
    _get,
    set,
    args: { user: User; accessToken: string; refreshToken: string }
  ) => {
    const { user, accessToken, refreshToken } = args;
    localStorage.setItem("accessToken", accessToken);
    localStorage.setItem("refreshToken", refreshToken);
    set(authStorageAtom, { user, isAuthenticated: true });
    navigateToAppWithWorkspace();
  }
);

// 認証情報をクリア
export const clearAuthAtom = atom(null, (_get, set) => {
  localStorage.removeItem("accessToken");
  localStorage.removeItem("refreshToken");
  set(authStorageAtom, { user: null, isAuthenticated: false });
  navigateToLogin();
});

// 初期化時のトークン検証と復元
export const initializeAuthAtom = atom(null, (get, set) => {
  const storage = get(authStorageAtom);
  const { accessToken, refreshToken } = loadTokensFromStorage();

  console.log("AuthStore - ストレージ復元:", {
    hasAccessToken: !!accessToken,
    hasRefreshToken: !!refreshToken,
    currentUser: storage.user,
    currentIsAuthenticated: storage.isAuthenticated,
  });

  if (accessToken && refreshToken) {
    console.log("AuthStore - トークン復元成功");
    // トークンが存在する場合は認証状態を維持
    if (!storage.isAuthenticated && storage.user) {
      set(authStorageAtom, { ...storage, isAuthenticated: true });
    }
  } else {
    // トークンが存在しない場合は認証状態をリセット
    console.log("AuthStore - トークンなし、認証状態をリセット");
    set(authStorageAtom, { user: null, isAuthenticated: false });
  }
});
