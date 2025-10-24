import { useEffect, useState } from "react";

import { useAtomValue, useSetAtom } from "jotai";

import {
  isAuthenticatedAtom,
  userAtom,
  accessTokenAtom,
  clearAuthAtom,
} from "@/lib/store/auth";

/**
 * 認証状態をチェックし、未認証の場合はログイン画面にリダイレクトするフック
 */
export function useAuthGuard() {
  const isAuthenticated = useAtomValue(isAuthenticatedAtom);
  const user = useAtomValue(userAtom);
  const accessToken = useAtomValue(accessTokenAtom);
  const clearAuth = useSetAtom(clearAuthAtom);
  const [isInitialized, setIsInitialized] = useState(false);

  useEffect(() => {
    // ストアの初期化を待つ
    const timer = setTimeout(() => {
      setIsInitialized(true);
    }, 100);

    return () => clearTimeout(timer);
  }, []);

  useEffect(() => {
    if (!isInitialized) return;

    // デバッグログを追加
    console.log("useAuthGuard - 認証状態チェック:", {
      isAuthenticated,
      hasUser: !!user,
      hasAccessToken: !!accessToken,
      isInitialized,
      currentPath: window.location.pathname,
    });

    // 認証状態をより厳密にチェック
    const hasValidAuth = isAuthenticated && user && accessToken;

    if (!hasValidAuth) {
      const currentPath = window.location.pathname;
      console.log("useAuthGuard - 認証失敗、リダイレクト実行:", { currentPath, hasValidAuth });
      if (currentPath !== "/login" && currentPath !== "/register") {
        // 認証情報をクリアしてからリダイレクト
        clearAuth();
        window.location.href = "/login";
      }
    }
  }, [isAuthenticated, user, accessToken, isInitialized, clearAuth]);

  return {
    isAuthenticated: isAuthenticated && !!user && !!accessToken,
    user,
    isLoading: !isInitialized,
  };
}
