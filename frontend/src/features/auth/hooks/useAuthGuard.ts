import { useEffect } from "react";

import { useLocation } from "@tanstack/react-router";
import { useAtomValue, useSetAtom } from "jotai";

import { navigateToLogin } from "@/lib/navigation";
import {
  accessTokenAtom,
  clearAuthAtom,
  isAuthInitializedAtom,
  isAuthenticatedAtom,
  userAtom,
} from "@/lib/store/auth";

const publicPaths = new Set(["/login", "/register"]);

/**
 * 認証状態をチェックし、未認証の場合はログイン画面にリダイレクトするフック
 */
export function useAuthGuard() {
  const isInitialized = useAtomValue(isAuthInitializedAtom);
  const isAuthenticated = useAtomValue(isAuthenticatedAtom);
  const user = useAtomValue(userAtom);
  const accessToken = useAtomValue(accessTokenAtom);
  const clearAuth = useSetAtom(clearAuthAtom);
  const location = useLocation();

  useEffect(() => {
    if (!isInitialized) {
      return;
    }

    if (isAuthenticated) {
      return;
    }

    if (publicPaths.has(location.pathname)) {
      return;
    }

    clearAuth();
    navigateToLogin();
  }, [isInitialized, isAuthenticated, location.pathname, clearAuth]);

  return {
    isAuthenticated: isInitialized && isAuthenticated && Boolean(user) && Boolean(accessToken),
    user,
    isLoading: !isInitialized,
  };
}
